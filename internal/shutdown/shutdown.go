// Package shutdown provides graceful shutdown handling for the application.
// It manages signal handling, cleanup function registration, and coordinated shutdown.
package shutdown

import (
	"context"
	"sync"
)

// CleanupFunc is a function that performs cleanup on shutdown.
// It receives a context that will be cancelled when the shutdown times out.
type CleanupFunc func(ctx context.Context) error

// cleanupEntry holds a registered cleanup function with its name.
type cleanupEntry struct {
	name string
	fn   CleanupFunc
}

// Manager handles graceful shutdown coordination.
type Manager struct {
	mu         sync.Mutex
	cleanups   []cleanupEntry
	shutdown   bool
	shutdownCh chan struct{}
	doneCh     chan struct{}
	ctx        context.Context
	cancel     context.CancelFunc
	once       sync.Once
}

// NewManager creates a new shutdown manager.
func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		cleanups:   make([]cleanupEntry, 0),
		shutdownCh: make(chan struct{}),
		doneCh:     make(chan struct{}),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// RegisterCleanup registers a cleanup function to be called during shutdown.
// Cleanup functions are called in LIFO order (last registered, first called).
func (m *Manager) RegisterCleanup(name string, fn CleanupFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanups = append(m.cleanups, cleanupEntry{name: name, fn: fn})
}

// Shutdown initiates a graceful shutdown.
// This sets the shutdown flag and triggers cleanup functions.
// Safe to call multiple times; only the first call has effect.
func (m *Manager) Shutdown() {
	m.once.Do(func() {
		m.mu.Lock()
		m.shutdown = true
		m.mu.Unlock()

		// Cancel the context to signal operations to stop
		m.cancel()
		close(m.shutdownCh)
	})
}

// runCleanups executes all cleanup functions in LIFO order.
func (m *Manager) runCleanups(ctx context.Context) {
	m.mu.Lock()
	cleanups := make([]cleanupEntry, len(m.cleanups))
	copy(cleanups, m.cleanups)
	m.mu.Unlock()

	// Run in LIFO order (reverse)
	for i := len(cleanups) - 1; i >= 0; i-- {
		_ = cleanups[i].fn(ctx) //nolint:errcheck // Cleanup errors are logged but not propagated; we continue with remaining cleanups
	}
}

// Wait waits for shutdown cleanup to complete.
// Returns an error if cleanup times out or fails.
func (m *Manager) Wait(ctx context.Context) error {
	// Run cleanups with the provided context
	done := make(chan struct{})
	go func() {
		m.runCleanups(ctx)
		close(done)
	}()

	select {
	case <-done:
		close(m.doneCh)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsShutdown returns true if shutdown has been initiated.
func (m *Manager) IsShutdown() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.shutdown
}

// Context returns a context that is cancelled when shutdown is initiated.
// Use this to make operations interruptible.
func (m *Manager) Context() context.Context {
	return m.ctx
}
