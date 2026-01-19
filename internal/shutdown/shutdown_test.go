package shutdown_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"todoat/internal/shutdown"
)

// =============================================================================
// CLI Tests (054-graceful-shutdown)
// =============================================================================

// TestGracefulShutdownSIGINT tests that the application handles SIGINT cleanly without data loss.
// This is simulated by triggering the shutdown handler directly since we can't
// send real signals in unit tests reliably.
func TestGracefulShutdownSIGINT(t *testing.T) {
	// Create a shutdown manager
	mgr := shutdown.NewManager()

	// Track if cleanup was called
	var cleanupCalled atomic.Bool

	// Register a cleanup function
	mgr.RegisterCleanup("test-cleanup", func(ctx context.Context) error {
		cleanupCalled.Store(true)
		return nil
	})

	// Trigger shutdown (simulating SIGINT)
	mgr.Shutdown()

	// Wait for shutdown to complete
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = mgr.Wait(ctx)

	// Verify cleanup was called
	if !cleanupCalled.Load() {
		t.Error("expected cleanup to be called on SIGINT shutdown")
	}
}

// TestGracefulShutdownSIGTERM tests that the application handles SIGTERM cleanly.
func TestGracefulShutdownSIGTERM(t *testing.T) {
	// SIGTERM and SIGINT should behave the same in the shutdown manager
	mgr := shutdown.NewManager()

	var cleanupCalled atomic.Bool

	mgr.RegisterCleanup("test-cleanup", func(ctx context.Context) error {
		cleanupCalled.Store(true)
		return nil
	})

	// Trigger shutdown (simulating SIGTERM)
	mgr.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = mgr.Wait(ctx)

	if !cleanupCalled.Load() {
		t.Error("expected cleanup to be called on SIGTERM shutdown")
	}
}

// TestShutdownDuringSync tests that sync in progress completes or rolls back safely.
func TestShutdownDuringSync(t *testing.T) {
	mgr := shutdown.NewManager()

	// Simulate a long-running sync operation
	syncStarted := make(chan struct{})
	syncCompleted := make(chan struct{})

	// Register a cleanup that simulates waiting for sync to complete
	mgr.RegisterCleanup("sync-cleanup", func(ctx context.Context) error {
		// Check if sync is running
		select {
		case <-syncStarted:
			// Wait for sync to complete or timeout
			select {
			case <-syncCompleted:
				return nil
			case <-ctx.Done():
				// Sync timed out, will be rolled back
				return ctx.Err()
			}
		default:
			return nil
		}
	})

	// Start a "sync" operation in the background
	go func() {
		close(syncStarted)
		time.Sleep(100 * time.Millisecond) // Simulate short sync
		close(syncCompleted)
	}()

	// Wait for sync to start
	<-syncStarted

	// Trigger shutdown while sync is in progress
	mgr.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := mgr.Wait(ctx)

	// Sync should have completed successfully
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Verify sync completed
	select {
	case <-syncCompleted:
		// Good, sync completed
	default:
		t.Error("expected sync to complete before shutdown finished")
	}
}

// TestShutdownDuringWrite tests that database write completes or rolls back safely.
func TestShutdownDuringWrite(t *testing.T) {
	mgr := shutdown.NewManager()

	writeStarted := make(chan struct{})
	writeCompleted := make(chan struct{})
	var writeCommitted atomic.Bool

	// Register a cleanup that waits for write to complete
	mgr.RegisterCleanup("db-cleanup", func(ctx context.Context) error {
		select {
		case <-writeStarted:
			// Write is in progress, wait for it
			select {
			case <-writeCompleted:
				return nil
			case <-ctx.Done():
				// Write will be rolled back by DB
				return ctx.Err()
			}
		default:
			return nil
		}
	})

	// Start a "write" operation
	go func() {
		close(writeStarted)
		time.Sleep(50 * time.Millisecond) // Simulate quick write
		writeCommitted.Store(true)
		close(writeCompleted)
	}()

	// Wait for write to start
	<-writeStarted

	// Trigger shutdown
	mgr.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := mgr.Wait(ctx)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if !writeCommitted.Load() {
		t.Error("expected write to be committed before shutdown")
	}
}

// TestShutdownExitCode tests that clean shutdown returns exit code 0.
func TestShutdownExitCode(t *testing.T) {
	mgr := shutdown.NewManager()

	// Register a simple cleanup
	mgr.RegisterCleanup("test", func(ctx context.Context) error {
		return nil
	})

	mgr.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := mgr.Wait(ctx)

	// Clean shutdown should return nil error (exit code 0)
	if err != nil {
		t.Errorf("expected exit code 0 (nil error), got: %v", err)
	}

	// Verify shutdown state
	if !mgr.IsShutdown() {
		t.Error("expected shutdown flag to be set")
	}
}

// TestShutdownPreventsNewOperations tests that shutdown flag prevents new operations.
func TestShutdownPreventsNewOperations(t *testing.T) {
	mgr := shutdown.NewManager()

	// Trigger shutdown
	mgr.Shutdown()

	// Check that shutdown flag is set
	if !mgr.IsShutdown() {
		t.Error("expected IsShutdown to return true after Shutdown call")
	}

	// Trying to do operations after shutdown should check this flag
	ctx := mgr.Context()
	select {
	case <-ctx.Done():
		// Good, context is cancelled
	default:
		t.Error("expected context to be cancelled after shutdown")
	}
}

// TestShutdownMultipleCleanups tests that multiple cleanup functions are all called.
func TestShutdownMultipleCleanups(t *testing.T) {
	mgr := shutdown.NewManager()

	var cleanup1Called, cleanup2Called, cleanup3Called atomic.Bool

	mgr.RegisterCleanup("cleanup1", func(ctx context.Context) error {
		cleanup1Called.Store(true)
		return nil
	})
	mgr.RegisterCleanup("cleanup2", func(ctx context.Context) error {
		cleanup2Called.Store(true)
		return nil
	})
	mgr.RegisterCleanup("cleanup3", func(ctx context.Context) error {
		cleanup3Called.Store(true)
		return nil
	})

	mgr.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = mgr.Wait(ctx)

	if !cleanup1Called.Load() {
		t.Error("cleanup1 was not called")
	}
	if !cleanup2Called.Load() {
		t.Error("cleanup2 was not called")
	}
	if !cleanup3Called.Load() {
		t.Error("cleanup3 was not called")
	}
}

// TestShutdownTimeout tests that shutdown times out if cleanup takes too long.
func TestShutdownTimeout(t *testing.T) {
	mgr := shutdown.NewManager()

	// Register a cleanup that takes forever
	mgr.RegisterCleanup("slow-cleanup", func(ctx context.Context) error {
		<-ctx.Done() // Wait for context to be cancelled
		return ctx.Err()
	})

	mgr.Shutdown()

	// Wait with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err := mgr.Wait(ctx)

	// Should timeout
	if err == nil {
		t.Error("expected timeout error")
	}
}

// TestShutdownConcurrentSafety tests that shutdown is safe to call from multiple goroutines.
func TestShutdownConcurrentSafety(t *testing.T) {
	mgr := shutdown.NewManager()

	var cleanupCount atomic.Int32

	mgr.RegisterCleanup("test", func(ctx context.Context) error {
		cleanupCount.Add(1)
		return nil
	})

	// Call shutdown from multiple goroutines
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mgr.Shutdown()
		}()
	}
	wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = mgr.Wait(ctx)

	// Cleanup should only be called once
	if cleanupCount.Load() != 1 {
		t.Errorf("expected cleanup to be called exactly once, got %d", cleanupCount.Load())
	}
}

// TestShutdownOrder tests that cleanup functions run in LIFO order (last registered first).
func TestShutdownOrder(t *testing.T) {
	mgr := shutdown.NewManager()

	var order []string
	var mu sync.Mutex

	mgr.RegisterCleanup("first", func(ctx context.Context) error {
		mu.Lock()
		order = append(order, "first")
		mu.Unlock()
		return nil
	})
	mgr.RegisterCleanup("second", func(ctx context.Context) error {
		mu.Lock()
		order = append(order, "second")
		mu.Unlock()
		return nil
	})
	mgr.RegisterCleanup("third", func(ctx context.Context) error {
		mu.Lock()
		order = append(order, "third")
		mu.Unlock()
		return nil
	})

	mgr.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = mgr.Wait(ctx)

	// LIFO order: third, second, first
	expected := []string{"third", "second", "first"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d cleanups, got %d", len(expected), len(order))
	}
	for i, name := range expected {
		if order[i] != name {
			t.Errorf("expected cleanup %d to be %q, got %q", i, name, order[i])
		}
	}
}
