// Package daemon - Circuit breaker pattern for backend-specific errors.
// Issue #114: Per-backend circuit breaker to prevent one failing backend
// from blocking sync for all backends.
package daemon

import (
	"sync"
	"time"
)

// DefaultCircuitBreakerThreshold is the number of consecutive failures before
// the circuit opens.
const DefaultCircuitBreakerThreshold = 3

// DefaultCircuitBreakerCooldown is how long the circuit stays open before
// transitioning to half-open.
const DefaultCircuitBreakerCooldown = 30 * time.Second

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	// CircuitClosed is the normal state - requests are allowed.
	CircuitClosed CircuitState = iota
	// CircuitOpen means the backend is failing - requests are blocked.
	CircuitOpen
	// CircuitHalfOpen means the cooldown expired - one probe request is allowed.
	CircuitHalfOpen
)

// String returns the string representation of the circuit state.
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern for a single backend.
type CircuitBreaker struct {
	mu           sync.Mutex
	threshold    int           // consecutive failures to open circuit
	cooldown     time.Duration // time to wait before half-open probe
	failureCount int           // current consecutive failures
	state        CircuitState  // current state
	openedAt     time.Time     // when the circuit was opened
}

// NewCircuitBreaker creates a new CircuitBreaker with the given threshold and cooldown.
func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		cooldown:  cooldown,
		state:     CircuitClosed,
	}
}

// Allow checks if a request should be allowed through the circuit breaker.
// Returns true if the request should proceed, false if it should be skipped.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if cooldown has elapsed → transition to half-open
		if time.Since(cb.openedAt) >= cb.cooldown {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		// Already in half-open, allow the probe
		return true
	default:
		return true
	}
}

// RecordSuccess records a successful operation, resetting the circuit breaker.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0
	cb.state = CircuitClosed
}

// RecordFailure records a failed operation.
// If the failure count reaches the threshold, the circuit opens.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	if cb.failureCount >= cb.threshold {
		cb.state = CircuitOpen
		cb.openedAt = time.Now()
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check for transition from open → half-open
	if cb.state == CircuitOpen && time.Since(cb.openedAt) >= cb.cooldown {
		cb.state = CircuitHalfOpen
	}
	return cb.state
}

// FailureCount returns the current consecutive failure count.
func (cb *CircuitBreaker) FailureCount() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.failureCount
}
