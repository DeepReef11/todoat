package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDaemonStartStop(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    100 * time.Millisecond,
		IdleTimeout: 0, // Disable idle timeout for this test
	}

	d := New(cfg)

	var syncCount int32
	d.SetSyncFunc(func() error {
		atomic.AddInt32(&syncCount, 1)
		return nil
	})

	// Start daemon in goroutine
	started := make(chan struct{})
	go func() {
		close(started)
		_ = d.Start()
	}()

	<-started
	time.Sleep(50 * time.Millisecond) // Wait for daemon to start

	// Check PID file exists
	if _, err := os.Stat(cfg.PIDPath); os.IsNotExist(err) {
		t.Errorf("PID file should exist after daemon start")
	}

	// Check socket exists
	if _, err := os.Stat(cfg.SocketPath); os.IsNotExist(err) {
		t.Errorf("Socket file should exist after daemon start")
	}

	// Wait for at least one sync
	time.Sleep(150 * time.Millisecond)

	// Stop daemon
	d.Stop()
	time.Sleep(50 * time.Millisecond)

	// Verify sync was called
	if atomic.LoadInt32(&syncCount) == 0 {
		t.Errorf("expected sync to be called at least once, got 0")
	}
}

func TestDaemonClientNotify(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    1 * time.Hour, // Long interval - we'll trigger via notify
		IdleTimeout: 0,
	}

	d := New(cfg)

	var syncCount int32
	d.SetSyncFunc(func() error {
		atomic.AddInt32(&syncCount, 1)
		return nil
	})

	// Start daemon
	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()
	time.Sleep(50 * time.Millisecond)

	// Create client
	client := NewClient(cfg.SocketPath)

	// Initial sync count should be 0
	if atomic.LoadInt32(&syncCount) != 0 {
		t.Errorf("expected initial sync count 0, got %d", syncCount)
	}

	// Notify to trigger sync
	if err := client.Notify(); err != nil {
		t.Errorf("notify failed: %v", err)
	}

	// Wait for sync to complete
	time.Sleep(100 * time.Millisecond)

	// Sync should have been triggered
	if atomic.LoadInt32(&syncCount) != 1 {
		t.Errorf("expected sync count 1 after notify, got %d", syncCount)
	}

	// Stop daemon and wait for it to finish
	d.Stop()
	<-done
}

func TestDaemonClientStatus(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    100 * time.Millisecond,
		IdleTimeout: 0,
	}

	d := New(cfg)
	d.SetSyncFunc(func() error { return nil })

	// Start daemon
	go func() {
		_ = d.Start()
	}()
	time.Sleep(150 * time.Millisecond) // Wait for at least one sync

	// Create client and get status
	client := NewClient(cfg.SocketPath)
	resp, err := client.Status()
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}

	if !resp.Running {
		t.Errorf("expected daemon to be running")
	}
	if resp.SyncCount == 0 {
		t.Errorf("expected sync count > 0")
	}

	// Stop daemon
	d.Stop()
	time.Sleep(100 * time.Millisecond) // Wait for cleanup
}

func TestDaemonIdleTimeout(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    1 * time.Hour,          // Long interval
		IdleTimeout: 100 * time.Millisecond, // Short idle timeout
	}

	d := New(cfg)
	d.SetSyncFunc(func() error { return nil })

	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait for daemon to auto-exit due to idle timeout
	select {
	case <-done:
		// Good, daemon exited
	case <-time.After(1 * time.Second):
		t.Errorf("daemon should have exited due to idle timeout")
		d.Stop()
	}

	// PID file should be cleaned up
	if _, err := os.Stat(cfg.PIDPath); err == nil {
		t.Errorf("PID file should be removed after daemon exit")
	}
}

func TestIsRunning(t *testing.T) {
	tmpDir := t.TempDir()
	pidPath := filepath.Join(tmpDir, "daemon.pid")
	socketPath := filepath.Join(tmpDir, "daemon.sock")

	// Initially should not be running
	if IsRunning(pidPath, socketPath) {
		t.Errorf("expected daemon to not be running initially")
	}

	// Start a daemon
	cfg := &Config{
		PIDPath:     pidPath,
		SocketPath:  socketPath,
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    100 * time.Millisecond,
		IdleTimeout: 0,
	}

	d := New(cfg)
	d.SetSyncFunc(func() error { return nil })

	go func() {
		_ = d.Start()
	}()
	time.Sleep(50 * time.Millisecond)

	// Now should be running
	if !IsRunning(pidPath, socketPath) {
		t.Errorf("expected daemon to be running")
	}

	// Stop daemon
	d.Stop()
	time.Sleep(50 * time.Millisecond)

	// Should not be running anymore
	if IsRunning(pidPath, socketPath) {
		t.Errorf("expected daemon to not be running after stop")
	}
}

func TestGetSocketPath(t *testing.T) {
	// Test with XDG_RUNTIME_DIR set
	origRuntime := os.Getenv("XDG_RUNTIME_DIR")
	defer func() { _ = os.Setenv("XDG_RUNTIME_DIR", origRuntime) }()

	_ = os.Setenv("XDG_RUNTIME_DIR", "/run/user/1000")
	path := GetSocketPath()
	expected := "/run/user/1000/todoat/daemon.sock"
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}

	// Test without XDG_RUNTIME_DIR
	_ = os.Unsetenv("XDG_RUNTIME_DIR")
	path = GetSocketPath()
	expected = "/tmp/todoat-daemon.sock"
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

// =============================================================================
// Issue #40: Multi-Backend Sync Support Tests
// =============================================================================

// TestDaemonMultiBackendIteration verifies daemon processes all enabled backends
func TestDaemonMultiBackendIteration(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    100 * time.Millisecond,
		IdleTimeout: 0,
	}

	d := New(cfg)

	// Track which backends were synced
	var backendsSynced []string
	var mu sync.Mutex

	// Register multiple backend sync functions
	d.AddBackendSyncFunc("backend1", func() error {
		mu.Lock()
		backendsSynced = append(backendsSynced, "backend1")
		mu.Unlock()
		return nil
	})
	d.AddBackendSyncFunc("backend2", func() error {
		mu.Lock()
		backendsSynced = append(backendsSynced, "backend2")
		mu.Unlock()
		return nil
	})
	d.AddBackendSyncFunc("backend3", func() error {
		mu.Lock()
		backendsSynced = append(backendsSynced, "backend3")
		mu.Unlock()
		return nil
	})

	// Start daemon
	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait for at least one sync cycle
	time.Sleep(200 * time.Millisecond)

	// Stop daemon
	d.Stop()
	<-done

	// Verify all backends were synced
	mu.Lock()
	defer mu.Unlock()

	if len(backendsSynced) < 3 {
		t.Errorf("expected at least 3 backends to be synced, got %d: %v", len(backendsSynced), backendsSynced)
	}

	// Check that each backend was synced at least once
	backendCounts := make(map[string]int)
	for _, b := range backendsSynced {
		backendCounts[b]++
	}

	for _, name := range []string{"backend1", "backend2", "backend3"} {
		if backendCounts[name] == 0 {
			t.Errorf("backend %q was not synced", name)
		}
	}
}

// TestDaemonBackendFailureIsolation verifies failure in one backend doesn't affect others
func TestDaemonBackendFailureIsolation(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    100 * time.Millisecond,
		IdleTimeout: 0,
	}

	d := New(cfg)

	// Track successful syncs
	var successfulSyncs []string
	var mu sync.Mutex

	// Backend 1: succeeds
	d.AddBackendSyncFunc("backend1", func() error {
		mu.Lock()
		successfulSyncs = append(successfulSyncs, "backend1")
		mu.Unlock()
		return nil
	})

	// Backend 2: fails
	d.AddBackendSyncFunc("backend2", func() error {
		return fmt.Errorf("simulated failure")
	})

	// Backend 3: succeeds
	d.AddBackendSyncFunc("backend3", func() error {
		mu.Lock()
		successfulSyncs = append(successfulSyncs, "backend3")
		mu.Unlock()
		return nil
	})

	// Start daemon
	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait for at least one sync cycle
	time.Sleep(200 * time.Millisecond)

	// Stop daemon
	d.Stop()
	<-done

	// Verify that backend1 and backend3 were synced despite backend2 failing
	mu.Lock()
	defer mu.Unlock()

	backendCounts := make(map[string]int)
	for _, b := range successfulSyncs {
		backendCounts[b]++
	}

	if backendCounts["backend1"] == 0 {
		t.Errorf("backend1 should have been synced successfully")
	}
	if backendCounts["backend3"] == 0 {
		t.Errorf("backend3 should have been synced successfully (failure in backend2 should not affect it)")
	}

	// Verify backend state shows error for backend2
	state := d.GetBackendState("backend2")
	if state == nil {
		t.Fatalf("expected state for backend2")
	}
	if state.ErrorCount == 0 {
		t.Errorf("expected error count > 0 for failing backend, got 0")
	}
	if state.LastError == "" {
		t.Errorf("expected last error to be recorded for failing backend")
	}
}

// TestDaemonPerBackendSyncState verifies separate sync state per backend
func TestDaemonPerBackendSyncState(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    100 * time.Millisecond,
		IdleTimeout: 0,
	}

	d := New(cfg)

	// Track sync calls per backend
	var backend1Syncs, backend2Syncs int32

	d.AddBackendSyncFunc("backend1", func() error {
		atomic.AddInt32(&backend1Syncs, 1)
		return nil
	})
	d.AddBackendSyncFunc("backend2", func() error {
		atomic.AddInt32(&backend2Syncs, 1)
		return nil
	})

	// Start daemon
	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait for multiple sync cycles
	time.Sleep(350 * time.Millisecond)

	// Stop daemon
	d.Stop()
	<-done

	// Verify each backend has its own sync state
	state1 := d.GetBackendState("backend1")
	state2 := d.GetBackendState("backend2")

	if state1 == nil {
		t.Fatalf("expected state for backend1")
	}
	if state2 == nil {
		t.Fatalf("expected state for backend2")
	}

	// Each backend should have its own sync count
	if state1.SyncCount == 0 {
		t.Errorf("expected sync count > 0 for backend1")
	}
	if state2.SyncCount == 0 {
		t.Errorf("expected sync count > 0 for backend2")
	}

	// Verify last sync time is tracked per backend
	if state1.LastSync.IsZero() {
		t.Errorf("expected last sync time to be set for backend1")
	}
	if state2.LastSync.IsZero() {
		t.Errorf("expected last sync time to be set for backend2")
	}
}

// TestDaemonPerBackendIntervals verifies configurable sync intervals per backend
func TestDaemonPerBackendIntervals(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    1 * time.Hour, // Default interval (long, won't trigger)
		IdleTimeout: 0,
	}

	d := New(cfg)

	// Track sync calls per backend
	var fastBackendSyncs, slowBackendSyncs int32

	// Add backends with different intervals
	d.AddBackendSyncFuncWithInterval("fast_backend", 50*time.Millisecond, func() error {
		atomic.AddInt32(&fastBackendSyncs, 1)
		return nil
	})
	d.AddBackendSyncFuncWithInterval("slow_backend", 200*time.Millisecond, func() error {
		atomic.AddInt32(&slowBackendSyncs, 1)
		return nil
	})

	// Start daemon
	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait long enough for fast_backend to sync multiple times, slow_backend once or twice
	time.Sleep(350 * time.Millisecond)

	// Stop daemon
	d.Stop()
	<-done

	fast := atomic.LoadInt32(&fastBackendSyncs)
	slow := atomic.LoadInt32(&slowBackendSyncs)

	// Fast backend should have synced more times than slow backend
	if fast <= slow {
		t.Errorf("expected fast_backend (%d syncs) to sync more often than slow_backend (%d syncs)", fast, slow)
	}

	// Fast backend should have synced at least 3 times (350ms / 50ms = 7, but timing overhead)
	if fast < 3 {
		t.Errorf("expected fast_backend to sync at least 3 times, got %d", fast)
	}

	// Slow backend should have synced 1-2 times (350ms / 200ms = 1.75)
	if slow < 1 || slow > 3 {
		t.Errorf("expected slow_backend to sync 1-3 times, got %d", slow)
	}
}
