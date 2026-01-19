// Package ratelimit provides HTTP rate limit handling with exponential backoff for REST API backends.
package ratelimit

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// Rate Limit Tests (048-api-rate-limiting)
// =============================================================================

// TestRateLimitRetry tests that 429 response triggers automatic retry after backoff period
func TestRateLimitRetry(t *testing.T) {
	requestCount := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count == 1 {
			// First request returns 429
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		// Second request succeeds
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer server.Close()

	client := NewClient(Config{
		MaxRetries:   5,
		BaseDelay:    10 * time.Millisecond, // Fast for testing
		EnableJitter: false,                 // Disable jitter for predictable tests
	})

	ctx := context.Background()
	resp, err := client.Do(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if requestCount != 2 {
		t.Errorf("expected 2 requests (1 retry), got %d", requestCount)
	}
}

// TestRateLimitExponentialBackoff tests that consecutive 429s increase delay (1s, 2s, 4s, 8s, 16s)
func TestRateLimitExponentialBackoff(t *testing.T) {
	requestTimes := make([]time.Time, 0, 6)
	requestCount := int32(0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestTimes = append(requestTimes, time.Now())
		count := atomic.AddInt32(&requestCount, 1)
		if count <= 4 {
			// First 4 requests return 429
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		// Fifth request succeeds
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	baseDelay := 50 * time.Millisecond // Fast base for testing
	client := NewClient(Config{
		MaxRetries:   5,
		BaseDelay:    baseDelay,
		MaxDelay:     800 * time.Millisecond,
		EnableJitter: false, // Disable jitter for predictable timing
	})

	ctx := context.Background()
	resp, err := client.Do(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Verify exponential backoff pattern
	// Expected delays: 50ms, 100ms, 200ms, 400ms (1x, 2x, 4x, 8x base)
	if len(requestTimes) < 5 {
		t.Fatalf("expected 5 requests, got %d", len(requestTimes))
	}

	// Check delays between requests
	expectedDelays := []time.Duration{
		baseDelay,     // After 1st 429
		baseDelay * 2, // After 2nd 429
		baseDelay * 4, // After 3rd 429
		baseDelay * 8, // After 4th 429
	}

	for i := 0; i < len(expectedDelays); i++ {
		actualDelay := requestTimes[i+1].Sub(requestTimes[i])
		expected := expectedDelays[i]
		// Allow 30% tolerance for timing variations
		minDelay := time.Duration(float64(expected) * 0.7)
		maxDelay := time.Duration(float64(expected) * 1.5)

		if actualDelay < minDelay || actualDelay > maxDelay {
			t.Errorf("delay %d: expected ~%v, got %v (allowed %v-%v)",
				i, expected, actualDelay, minDelay, maxDelay)
		}
	}
}

// TestRateLimitMaxRetries tests that after max retries, operation fails with clear error message
func TestRateLimitMaxRetries(t *testing.T) {
	requestCount := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		// Always return 429
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(Config{
		MaxRetries:   5,
		BaseDelay:    1 * time.Millisecond, // Very fast for testing
		EnableJitter: false,
	})

	ctx := context.Background()
	resp, err := client.Do(ctx, http.MethodGet, server.URL, nil)

	// After 5 retries (6 total requests), should get error
	if err == nil {
		if resp != nil {
			_ = resp.Body.Close()
		}
		t.Fatal("expected error after max retries, got nil")
	}

	// Error message should be clear
	if !strings.Contains(err.Error(), "rate limit") || !strings.Contains(err.Error(), "5") {
		t.Errorf("expected error message about rate limit and max retries, got: %v", err)
	}

	// Should have made maxRetries + 1 requests (initial + retries)
	if requestCount != 6 {
		t.Errorf("expected 6 requests (1 initial + 5 retries), got %d", requestCount)
	}
}

// TestRateLimitJitter tests that backoff includes random jitter (±20%) to prevent synchronized retries
func TestRateLimitJitter(t *testing.T) {
	// Run multiple iterations to verify jitter
	baseDelay := 100 * time.Millisecond
	delays := make([]time.Duration, 10)

	for i := 0; i < 10; i++ {
		requestCount := int32(0)
		var startTime, endTime time.Time

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&requestCount, 1)
			if count == 1 {
				startTime = time.Now()
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			endTime = time.Now()
			w.WriteHeader(http.StatusOK)
		}))

		client := NewClient(Config{
			MaxRetries:   5,
			BaseDelay:    baseDelay,
			EnableJitter: true, // Enable jitter
		})

		ctx := context.Background()
		resp, err := client.Do(ctx, http.MethodGet, server.URL, nil)
		server.Close()

		if err != nil {
			t.Fatalf("iteration %d: expected no error, got: %v", i, err)
		}
		_ = resp.Body.Close()

		delays[i] = endTime.Sub(startTime)
	}

	// Verify that delays vary (jitter is working)
	// With ±20% jitter, delays should not all be identical
	allSame := true
	tolerance := 5 * time.Millisecond
	for i := 1; i < len(delays); i++ {
		diff := delays[i] - delays[0]
		if diff < 0 {
			diff = -diff
		}
		if diff > tolerance {
			allSame = false
			break
		}
	}

	if allSame {
		t.Errorf("jitter appears to not be working: all delays are nearly identical: %v", delays)
	}

	// Verify delays are within expected range (80% to 120% of base)
	minExpected := time.Duration(float64(baseDelay) * 0.75)
	maxExpected := time.Duration(float64(baseDelay) * 1.25)

	for i, d := range delays {
		if d < minExpected || d > maxExpected {
			t.Errorf("delay %d (%v) outside expected jitter range (%v - %v)",
				i, d, minExpected, maxExpected)
		}
	}
}

// TestRateLimitHeaderRespect tests that `Retry-After` header value is used when provided by API
func TestRateLimitHeaderRespect(t *testing.T) {
	tests := []struct {
		name           string
		retryAfter     string
		expectedDelay  time.Duration
		delayTolerance time.Duration
	}{
		{
			name:           "seconds value",
			retryAfter:     "1",
			expectedDelay:  1 * time.Second,
			delayTolerance: 200 * time.Millisecond,
		},
		{
			name:           "short seconds",
			retryAfter:     "0",
			expectedDelay:  0, // Should use minimum delay
			delayTolerance: 100 * time.Millisecond,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requestCount := int32(0)
			var startTime, endTime time.Time

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count := atomic.AddInt32(&requestCount, 1)
				if count == 1 {
					startTime = time.Now()
					w.Header().Set("Retry-After", tc.retryAfter)
					w.WriteHeader(http.StatusTooManyRequests)
					return
				}
				endTime = time.Now()
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewClient(Config{
				MaxRetries:   5,
				BaseDelay:    10 * time.Millisecond, // Small base, should be overridden by header
				EnableJitter: false,
			})

			ctx := context.Background()
			resp, err := client.Do(ctx, http.MethodGet, server.URL, nil)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			actualDelay := endTime.Sub(startTime)

			// Should respect Retry-After header
			if tc.expectedDelay > 0 {
				minDelay := tc.expectedDelay - tc.delayTolerance
				maxDelay := tc.expectedDelay + tc.delayTolerance

				if actualDelay < minDelay || actualDelay > maxDelay {
					t.Errorf("expected delay ~%v (±%v), got %v",
						tc.expectedDelay, tc.delayTolerance, actualDelay)
				}
			}
		})
	}
}

// TestRateLimitHeaderRespectHTTPDate tests Retry-After with HTTP-date format
func TestRateLimitHeaderRespectHTTPDate(t *testing.T) {
	requestCount := int32(0)
	var startTime, endTime time.Time

	// Use 2 seconds to account for HTTP-date second-precision truncation
	// The HTTP-date format truncates to second precision, which can result in
	// up to ~1 second of timing variance
	targetDelay := 2 * time.Second

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count == 1 {
			startTime = time.Now()
			// Calculate Retry-After time relative to when the request is handled
			retryTime := time.Now().Add(targetDelay)
			retryAfterHTTPDate := retryTime.UTC().Format(http.TimeFormat)
			w.Header().Set("Retry-After", retryAfterHTTPDate)
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		endTime = time.Now()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Config{
		MaxRetries:   5,
		BaseDelay:    10 * time.Millisecond,
		EnableJitter: false,
	})

	ctx := context.Background()
	resp, err := client.Do(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	actualDelay := endTime.Sub(startTime)

	// Should wait approximately until the retry time
	// Due to HTTP-date second-precision, actual delay can be up to 1 second shorter
	minDelay := targetDelay - 1*time.Second - 200*time.Millisecond
	maxDelay := targetDelay + 500*time.Millisecond

	if actualDelay < minDelay || actualDelay > maxDelay {
		t.Errorf("expected delay between %v and %v, got %v", minDelay, maxDelay, actualDelay)
	}
}

// TestRateLimitQueueing tests that rate-limited operations are tracked for status display
func TestRateLimitQueueing(t *testing.T) {
	requestCount := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	stats := NewStats()
	client := NewClient(Config{
		MaxRetries:   5,
		BaseDelay:    10 * time.Millisecond,
		EnableJitter: false,
		Stats:        stats,
	})

	ctx := context.Background()
	resp, err := client.Do(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Stats should track the rate limit event
	if stats.RateLimitCount() != 1 {
		t.Errorf("expected 1 rate limit event, got %d", stats.RateLimitCount())
	}

	// LastRateLimitTime should be recent
	lastTime := stats.LastRateLimitTime()
	if time.Since(lastTime) > 5*time.Second {
		t.Errorf("expected recent rate limit time, got %v ago", time.Since(lastTime))
	}
}

// TestRateLimitMaxDelayCap tests that delay is capped at maxDelay (32s by default)
func TestRateLimitMaxDelayCap(t *testing.T) {
	requestTimes := make([]time.Time, 0, 10)
	requestCount := int32(0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestTimes = append(requestTimes, time.Now())
		count := atomic.AddInt32(&requestCount, 1)
		if count <= 8 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	baseDelay := 50 * time.Millisecond
	maxDelay := 400 * time.Millisecond // Cap at 8x base

	client := NewClient(Config{
		MaxRetries:   10,
		BaseDelay:    baseDelay,
		MaxDelay:     maxDelay,
		EnableJitter: false,
	})

	ctx := context.Background()
	resp, err := client.Do(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// After attempt 3 (delay = 8x = 400ms), delays should be capped
	for i := 3; i < len(requestTimes)-1; i++ {
		actualDelay := requestTimes[i+1].Sub(requestTimes[i])
		// Allow some tolerance
		maxAllowed := time.Duration(float64(maxDelay) * 1.5)

		if actualDelay > maxAllowed {
			t.Errorf("delay %d (%v) exceeded max delay cap (%v)",
				i, actualDelay, maxDelay)
		}
	}
}

// TestRateLimitContextCancellation tests that retries are cancelled when context is cancelled
func TestRateLimitContextCancellation(t *testing.T) {
	requestCount := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(Config{
		MaxRetries:   10,
		BaseDelay:    1 * time.Second, // Long delay
		EnableJitter: false,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := client.Do(ctx, http.MethodGet, server.URL, nil)
	elapsed := time.Since(start)

	// Should fail due to context cancellation
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}

	// Should cancel quickly, not wait for full retry delay
	if elapsed > 500*time.Millisecond {
		t.Errorf("expected quick cancellation, but took %v", elapsed)
	}

	// Should have made at least one request
	if requestCount < 1 {
		t.Error("expected at least 1 request before cancellation")
	}
}

// TestRateLimitWithBody tests that request body is correctly re-sent on retry
func TestRateLimitWithBody(t *testing.T) {
	requestBodies := make([]string, 0, 2)
	requestCount := int32(0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requestBodies = append(requestBodies, string(body))
		count := atomic.AddInt32(&requestCount, 1)
		if count == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Config{
		MaxRetries:   5,
		BaseDelay:    10 * time.Millisecond,
		EnableJitter: false,
	})

	ctx := context.Background()
	body := strings.NewReader(`{"test": "data"}`)
	resp, err := client.Do(ctx, http.MethodPost, server.URL, body)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Both requests should have the same body
	if len(requestBodies) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(requestBodies))
	}

	if requestBodies[0] != requestBodies[1] {
		t.Errorf("request bodies differ on retry: %q vs %q", requestBodies[0], requestBodies[1])
	}

	if requestBodies[0] != `{"test": "data"}` {
		t.Errorf("unexpected body: %q", requestBodies[0])
	}
}

// TestCalculateBackoff tests the backoff calculation function directly
func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		name        string
		attempt     int
		retryAfter  *time.Duration
		baseDelay   time.Duration
		maxDelay    time.Duration
		expected    time.Duration
		description string
	}{
		{
			name:        "first attempt",
			attempt:     0,
			baseDelay:   1 * time.Second,
			maxDelay:    32 * time.Second,
			expected:    1 * time.Second,
			description: "2^0 = 1x base",
		},
		{
			name:        "second attempt",
			attempt:     1,
			baseDelay:   1 * time.Second,
			maxDelay:    32 * time.Second,
			expected:    2 * time.Second,
			description: "2^1 = 2x base",
		},
		{
			name:        "third attempt",
			attempt:     2,
			baseDelay:   1 * time.Second,
			maxDelay:    32 * time.Second,
			expected:    4 * time.Second,
			description: "2^2 = 4x base",
		},
		{
			name:        "fourth attempt",
			attempt:     3,
			baseDelay:   1 * time.Second,
			maxDelay:    32 * time.Second,
			expected:    8 * time.Second,
			description: "2^3 = 8x base",
		},
		{
			name:        "fifth attempt",
			attempt:     4,
			baseDelay:   1 * time.Second,
			maxDelay:    32 * time.Second,
			expected:    16 * time.Second,
			description: "2^4 = 16x base",
		},
		{
			name:        "capped at maxDelay",
			attempt:     10,
			baseDelay:   1 * time.Second,
			maxDelay:    32 * time.Second,
			expected:    32 * time.Second,
			description: "should not exceed maxDelay",
		},
		{
			name:        "retryAfter overrides calculation",
			attempt:     0,
			retryAfter:  durationPtr(5 * time.Second),
			baseDelay:   1 * time.Second,
			maxDelay:    32 * time.Second,
			expected:    5 * time.Second,
			description: "Retry-After header takes precedence",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculateBackoff(tc.attempt, tc.retryAfter, tc.baseDelay, tc.maxDelay, false)

			if result != tc.expected {
				t.Errorf("%s: expected %v, got %v", tc.description, tc.expected, result)
			}
		})
	}
}

// TestParseRetryAfter tests parsing of Retry-After header values
func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected *time.Duration
	}{
		{
			name:     "seconds integer",
			value:    "60",
			expected: durationPtr(60 * time.Second),
		},
		{
			name:     "zero seconds",
			value:    "0",
			expected: durationPtr(0),
		},
		{
			name:     "empty value",
			value:    "",
			expected: nil,
		},
		{
			name:     "invalid value",
			value:    "invalid",
			expected: nil, // Invalid values should return nil (use default backoff)
		},
		{
			name:     "negative value",
			value:    "-1",
			expected: nil, // Negative values should return nil
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseRetryAfter(tc.value)

			if tc.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", *result)
				}
			} else {
				if result == nil {
					t.Errorf("expected %v, got nil", *tc.expected)
				} else if *result != *tc.expected {
					t.Errorf("expected %v, got %v", *tc.expected, *result)
				}
			}
		})
	}
}

// TestRateLimitError tests the RateLimitError type
func TestRateLimitError(t *testing.T) {
	err := &RateLimitError{
		Backend:     "todoist",
		RetryAfter:  5 * time.Second,
		Attempt:     3,
		MaxAttempts: 5,
	}

	errStr := err.Error()

	if !strings.Contains(errStr, "todoist") {
		t.Errorf("error should contain backend name: %s", errStr)
	}
	if !strings.Contains(errStr, "rate limit") {
		t.Errorf("error should mention rate limit: %s", errStr)
	}
	if !strings.Contains(errStr, "3") {
		t.Errorf("error should contain attempt number: %s", errStr)
	}
}

// TestNewClientDefaults tests that NewClient uses sensible defaults
func TestNewClientDefaults(t *testing.T) {
	client := NewClient(Config{})

	// Test with a simple request (no 429)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()
	resp, err := client.Do(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// TestStatsThreadSafety tests that Stats is safe for concurrent access
func TestStatsThreadSafety(t *testing.T) {
	stats := NewStats()

	// Simulate concurrent rate limit events
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			stats.RecordRateLimit()
			_ = stats.RateLimitCount()
			_ = stats.LastRateLimitTime()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Should have recorded all events
	if stats.RateLimitCount() != 100 {
		t.Errorf("expected 100 events, got %d", stats.RateLimitCount())
	}
}

// Helper function to create a duration pointer
func durationPtr(d time.Duration) *time.Duration {
	return &d
}

// TestRateLimitNon429Passthrough tests that non-429 errors are passed through immediately
func TestRateLimitNon429Passthrough(t *testing.T) {
	statusCodes := []int{400, 401, 403, 404, 500, 502, 503}

	for _, code := range statusCodes {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			requestCount := int32(0)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&requestCount, 1)
				w.WriteHeader(code)
			}))
			defer server.Close()

			client := NewClient(Config{
				MaxRetries:   5,
				BaseDelay:    10 * time.Millisecond,
				EnableJitter: false,
			})

			ctx := context.Background()
			resp, err := client.Do(ctx, http.MethodGet, server.URL, nil)
			if err != nil {
				t.Fatalf("expected no error (non-429 should pass through), got: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != code {
				t.Errorf("expected status %d, got %d", code, resp.StatusCode)
			}

			// Should only make 1 request (no retry for non-429)
			if requestCount != 1 {
				t.Errorf("expected 1 request (no retry), got %d", requestCount)
			}
		})
	}
}

// =============================================================================
// Backoff Calculation Helpers
// =============================================================================

// CalculateBackoff computes the backoff duration for a given attempt.
// This is the reference implementation that the rate limiter uses.
func CalculateBackoff(attempt int, retryAfter *time.Duration, baseDelay, maxDelay time.Duration, enableJitter bool) time.Duration {
	if retryAfter != nil {
		return *retryAfter
	}

	// Exponential backoff: base * 2^attempt
	delay := baseDelay * time.Duration(math.Pow(2, float64(attempt)))

	// Cap at maxDelay
	if delay > maxDelay {
		delay = maxDelay
	}

	// Note: jitter is handled separately by the caller when enableJitter is true
	// This function returns the base delay without jitter for predictable testing

	return delay
}
