// Package ratelimit provides HTTP rate limit handling with exponential backoff for REST API backends.
package ratelimit

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Config holds configuration for the rate-limiting HTTP client.
type Config struct {
	// MaxRetries is the maximum number of retry attempts after receiving 429.
	// Default: 5
	MaxRetries int

	// BaseDelay is the initial delay before the first retry.
	// Default: 1 second
	BaseDelay time.Duration

	// MaxDelay is the maximum delay between retries.
	// Default: 32 seconds
	MaxDelay time.Duration

	// EnableJitter adds random jitter (±20%) to prevent thundering herd.
	// Default: true
	EnableJitter bool

	// Stats is an optional stats tracker for recording rate limit events.
	Stats *Stats

	// Backend name for error messages and logging.
	Backend string
}

// Client is an HTTP client that handles rate limiting with exponential backoff.
type Client struct {
	httpClient   *http.Client
	maxRetries   int
	baseDelay    time.Duration
	maxDelay     time.Duration
	enableJitter bool
	stats        *Stats
	backend      string
}

// NewClient creates a new rate-limiting HTTP client with the given configuration.
func NewClient(cfg Config) *Client {
	// Apply defaults
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 5
	}

	baseDelay := cfg.BaseDelay
	if baseDelay <= 0 {
		baseDelay = 1 * time.Second
	}

	maxDelay := cfg.MaxDelay
	if maxDelay <= 0 {
		maxDelay = 32 * time.Second
	}

	// EnableJitter defaults to true unless explicitly configured
	enableJitter := cfg.EnableJitter

	return &Client{
		httpClient:   &http.Client{},
		maxRetries:   maxRetries,
		baseDelay:    baseDelay,
		maxDelay:     maxDelay,
		enableJitter: enableJitter,
		stats:        cfg.Stats,
		backend:      cfg.Backend,
	}
}

// Do performs an HTTP request with automatic retry on rate limiting (429 responses).
// It handles the Retry-After header and implements exponential backoff with optional jitter.
func (c *Client) Do(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	// Read body into buffer so we can re-send on retry
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
	}

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Create request with fresh body reader
		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Perform request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		// Check if rate limited
		if resp.StatusCode != http.StatusTooManyRequests {
			// Not rate limited - return response
			return resp, nil
		}

		// Close body from rate-limited response (we'll retry)
		_ = resp.Body.Close()

		// Record rate limit event in stats
		if c.stats != nil {
			c.stats.RecordRateLimit()
		}

		// Check if we've exhausted retries
		if attempt >= c.maxRetries {
			break
		}

		// Parse Retry-After header if present
		retryAfter := ParseRetryAfter(resp.Header.Get("Retry-After"))

		// Calculate backoff delay
		delay := c.calculateBackoff(attempt, retryAfter)

		// Wait for backoff delay or context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			// Continue to next retry
		}
	}

	// Exhausted all retries
	return nil, &RateLimitError{
		Backend:     c.backend,
		RetryAfter:  c.baseDelay,
		Attempt:     c.maxRetries,
		MaxAttempts: c.maxRetries,
	}
}

// calculateBackoff computes the backoff duration for a given attempt.
func (c *Client) calculateBackoff(attempt int, retryAfter *time.Duration) time.Duration {
	if retryAfter != nil {
		return *retryAfter
	}

	// Exponential backoff: base * 2^attempt
	delay := c.baseDelay * time.Duration(math.Pow(2, float64(attempt)))

	// Cap at maxDelay
	if delay > c.maxDelay {
		delay = c.maxDelay
	}

	// Add jitter if enabled (±20%)
	if c.enableJitter {
		jitterFactor := 0.8 + rand.Float64()*0.4 // 0.8 to 1.2
		delay = time.Duration(float64(delay) * jitterFactor)
	}

	return delay
}

// RateLimitError represents an error when rate limit retries are exhausted.
type RateLimitError struct {
	Backend     string
	RetryAfter  time.Duration
	Attempt     int
	MaxAttempts int
}

// Error implements the error interface.
func (e *RateLimitError) Error() string {
	backend := e.Backend
	if backend == "" {
		backend = "API"
	}
	return fmt.Sprintf("%s rate limit exceeded after %d retries (max %d)", backend, e.Attempt, e.MaxAttempts)
}

// ParseRetryAfter parses the Retry-After header value.
// It supports both seconds format (integer) and HTTP-date format.
// Returns nil if the value is invalid or empty.
func ParseRetryAfter(value string) *time.Duration {
	if value == "" {
		return nil
	}

	// Try parsing as seconds (integer)
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		if seconds < 0 {
			return nil
		}
		d := time.Duration(seconds) * time.Second
		return &d
	}

	// Try parsing as HTTP-date
	if t, err := http.ParseTime(value); err == nil {
		d := time.Until(t)
		if d < 0 {
			d = 0
		}
		return &d
	}

	return nil
}

// Stats tracks rate limit statistics for a backend.
type Stats struct {
	mu              sync.RWMutex
	rateLimitCount  int64
	lastRateLimitAt time.Time
}

// NewStats creates a new Stats instance.
func NewStats() *Stats {
	return &Stats{}
}

// RecordRateLimit records a rate limit event.
func (s *Stats) RecordRateLimit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rateLimitCount++
	s.lastRateLimitAt = time.Now()
}

// RateLimitCount returns the total number of rate limit events.
func (s *Stats) RateLimitCount() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rateLimitCount
}

// LastRateLimitTime returns the time of the last rate limit event.
func (s *Stats) LastRateLimitTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastRateLimitAt
}
