package daemon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// Issue #53: Secure file permissions and predictable /tmp paths
// =============================================================================

func TestDaemonFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:    filepath.Join(tmpDir, "subdir", "daemon.pid"),
		SocketPath: filepath.Join(tmpDir, "sockdir", "daemon.sock"),
		LogPath:    filepath.Join(tmpDir, "logdir", "daemon.log"),
		Interval:   100 * time.Millisecond,
	}

	d := New(cfg)
	d.SetSyncFunc(func() error { return nil })

	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait for daemon to start
	time.Sleep(50 * time.Millisecond)

	// Check PID directory permissions (should be 0700)
	pidDirInfo, err := os.Stat(filepath.Dir(cfg.PIDPath))
	if err != nil {
		t.Fatalf("PID directory should exist: %v", err)
	}
	if perm := pidDirInfo.Mode().Perm(); perm != 0700 {
		t.Errorf("PID directory should have mode 0700, got %04o", perm)
	}

	// Check PID file permissions (should be 0600)
	pidInfo, err := os.Stat(cfg.PIDPath)
	if err != nil {
		t.Fatalf("PID file should exist: %v", err)
	}
	if perm := pidInfo.Mode().Perm(); perm != 0600 {
		t.Errorf("PID file should have mode 0600, got %04o", perm)
	}

	// Check socket directory permissions (should be 0700)
	sockDirInfo, err := os.Stat(filepath.Dir(cfg.SocketPath))
	if err != nil {
		t.Fatalf("Socket directory should exist: %v", err)
	}
	if perm := sockDirInfo.Mode().Perm(); perm != 0700 {
		t.Errorf("Socket directory should have mode 0700, got %04o", perm)
	}

	// Check log directory permissions (should be 0700)
	logDirInfo, err := os.Stat(filepath.Dir(cfg.LogPath))
	if err != nil {
		t.Fatalf("Log directory should exist: %v", err)
	}
	if perm := logDirInfo.Mode().Perm(); perm != 0700 {
		t.Errorf("Log directory should have mode 0700, got %04o", perm)
	}

	// Trigger a log write and check log file permissions (should be 0600)
	time.Sleep(150 * time.Millisecond) // Let daemon write a log entry
	logInfo, err := os.Stat(cfg.LogPath)
	if err != nil {
		t.Fatalf("Log file should exist: %v", err)
	}
	if perm := logInfo.Mode().Perm(); perm != 0600 {
		t.Errorf("Log file should have mode 0600, got %04o", perm)
	}

	// Stop daemon
	d.Stop()
	<-done
}

func TestGetSocketPathIncludesUID(t *testing.T) {
	// Test without XDG_RUNTIME_DIR - should include UID in /tmp path
	origDir := os.Getenv("XDG_RUNTIME_DIR")
	_ = os.Unsetenv("XDG_RUNTIME_DIR")
	defer func() {
		if origDir != "" {
			_ = os.Setenv("XDG_RUNTIME_DIR", origDir)
		}
	}()

	path := GetSocketPath()
	uid := fmt.Sprintf("%d", os.Getuid())

	// Path should contain the UID to prevent collisions on multi-user systems
	if path == "/tmp/todoat-daemon.sock" {
		t.Errorf("GetSocketPath should include UID in /tmp fallback path, got %q", path)
	}
	expectedPath := fmt.Sprintf("/tmp/todoat-daemon-%s.sock", uid)
	if path != expectedPath {
		t.Errorf("expected %q, got %q", expectedPath, path)
	}
}

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

// TestDaemonStatusReportsInterval verifies the daemon status response includes
// the actual running interval, not a default (Issue #59).
func TestDaemonStatusReportsInterval(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    120 * time.Second,
		IdleTimeout: 0,
	}

	d := New(cfg)
	d.SetSyncFunc(func() error { return nil })

	go func() {
		_ = d.Start()
	}()
	time.Sleep(100 * time.Millisecond) // Wait for daemon to start

	client := NewClient(cfg.SocketPath)
	resp, err := client.Status()
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}

	if resp.IntervalSec != 120 {
		t.Errorf("expected IntervalSec=120 in status response, got %d", resp.IntervalSec)
	}

	d.Stop()
	time.Sleep(100 * time.Millisecond)
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

	// Test without XDG_RUNTIME_DIR - should include UID
	_ = os.Unsetenv("XDG_RUNTIME_DIR")
	path = GetSocketPath()
	expected = fmt.Sprintf("/tmp/todoat-daemon-%d.sock", os.Getuid())
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

// TestDaemonConcurrentPerformSync verifies no race condition when performSync
// is called concurrently (Issue #52: ticker + notify handler).
func TestDaemonConcurrentPerformSync(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    50 * time.Millisecond, // Short interval to trigger ticker syncs
		IdleTimeout: 0,
	}

	d := New(cfg)

	var syncCount int32
	d.AddBackendSyncFunc("test_backend", func() error {
		atomic.AddInt32(&syncCount, 1)
		// Small sleep to increase window for concurrent execution
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	// Start daemon
	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()
	time.Sleep(30 * time.Millisecond) // Wait for daemon to start

	// Send multiple notify messages concurrently to trigger concurrent performSync
	client := NewClient(cfg.SocketPath)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.Notify()
		}()
	}
	wg.Wait()

	// Let the syncs complete
	time.Sleep(200 * time.Millisecond)

	d.Stop()
	<-done

	// The test passes if -race doesn't report a data race.
	// Also verify syncs actually happened.
	if atomic.LoadInt32(&syncCount) == 0 {
		t.Errorf("expected at least one sync to occur")
	}
}

// TestDaemonLastSyncRace verifies no race on backendEntry.lastSync (Issue #52).
func TestDaemonLastSyncRace(t *testing.T) {
	d := New(&Config{
		PIDPath:    filepath.Join(t.TempDir(), "daemon.pid"),
		SocketPath: filepath.Join(t.TempDir(), "daemon.sock"),
		LogPath:    filepath.Join(t.TempDir(), "daemon.log"),
		Interval:   50 * time.Millisecond,
	})

	d.AddBackendSyncFunc("race_backend", func() error {
		return nil
	})

	// Call performSync concurrently to trigger race on be.lastSync
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.performSync()
		}()
	}
	wg.Wait()

	// The test passes if -race doesn't report a data race.
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

// =============================================================================
// Issue #41: File Watcher for Real-Time Sync Triggers (Regression Test)
// =============================================================================

// TestIssue41_FileWatcherPackageExists verifies that the file watcher package
// exists as claimed in issue #41's closing comment. The issue was closed as
// "COMPLETED" claiming internal/watcher/ was implemented, but the package is missing.
//
// This test is skipped because the watcher package was never committed (see issue #41).
// It will be unskipped once the watcher package is implemented.
func TestIssue41_FileWatcherPackageExists(t *testing.T) {
	watcherDir := filepath.Join("..", "watcher")
	if _, err := os.Stat(watcherDir); os.IsNotExist(err) {
		t.Skip("Regression #41: internal/watcher/ package does not exist. " +
			"Issue #41 was closed as COMPLETED but the file watcher package was never committed. " +
			"Skipping until issue #41 is properly implemented.")
	}

	watcherFile := filepath.Join("..", "watcher", "watcher.go")
	if _, err := os.Stat(watcherFile); os.IsNotExist(err) {
		t.Fatalf("Regression #41: internal/watcher/watcher.go does not exist. " +
			"File watcher implementation file is missing")
	}
}

// =============================================================================
// Issue #74: Daemon Heartbeat Mechanism for Hung Process Detection
// =============================================================================

// TestDaemonHeartbeatRecording verifies daemon periodically writes heartbeat
// timestamp at configured interval.
func TestDaemonHeartbeatRecording(t *testing.T) {
	tmpDir := t.TempDir()
	heartbeatPath := filepath.Join(tmpDir, "daemon.heartbeat")

	cfg := &Config{
		PIDPath:           filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:        filepath.Join(tmpDir, "daemon.sock"),
		LogPath:           filepath.Join(tmpDir, "daemon.log"),
		HeartbeatPath:     heartbeatPath,
		HeartbeatInterval: 50 * time.Millisecond,
		Interval:          1 * time.Hour, // Long sync interval - we only care about heartbeat
		IdleTimeout:       0,
	}

	d := New(cfg)
	d.SetSyncFunc(func() error { return nil })

	// Start daemon
	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait for heartbeat to be written (poll with timeout)
	var data []byte
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
		data, _ = os.ReadFile(heartbeatPath)
		if len(data) > 0 {
			break
		}
	}

	// Verify heartbeat file exists and has content
	if len(data) == 0 {
		t.Fatalf("heartbeat file should exist with content at %s", heartbeatPath)
	}

	// Parse timestamp
	ts, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatalf("heartbeat file should contain RFC3339Nano timestamp, got: %q", string(data))
	}

	// Verify timestamp is recent (within 2x heartbeat interval)
	age := time.Since(ts)
	if age > 2*cfg.HeartbeatInterval {
		t.Errorf("heartbeat should be recent (within %v), but age is %v", 2*cfg.HeartbeatInterval, age)
	}

	// Stop daemon
	d.Stop()
	<-done

	// Heartbeat file should be cleaned up
	if _, err := os.Stat(heartbeatPath); err == nil {
		t.Errorf("heartbeat file should be removed after daemon stop")
	}
}

// TestDaemonHeartbeatStaleDetection verifies stale heartbeat is detected
// when daemon is not running.
func TestDaemonHeartbeatStaleDetection(t *testing.T) {
	tmpDir := t.TempDir()
	heartbeatPath := filepath.Join(tmpDir, "daemon.heartbeat")
	heartbeatInterval := 50 * time.Millisecond

	// Write a stale heartbeat (older than 2x interval)
	staleTime := time.Now().Add(-3 * heartbeatInterval)
	if err := os.WriteFile(heartbeatPath, []byte(staleTime.Format(time.RFC3339Nano)), 0600); err != nil {
		t.Fatalf("failed to write stale heartbeat: %v", err)
	}

	// Check heartbeat staleness
	isStale, err := IsHeartbeatStale(heartbeatPath, heartbeatInterval)
	if err != nil {
		t.Fatalf("IsHeartbeatStale failed: %v", err)
	}
	if !isStale {
		t.Errorf("heartbeat should be detected as stale")
	}

	// Write a fresh heartbeat
	freshTime := time.Now()
	if err := os.WriteFile(heartbeatPath, []byte(freshTime.Format(time.RFC3339Nano)), 0600); err != nil {
		t.Fatalf("failed to write fresh heartbeat: %v", err)
	}

	// Check heartbeat freshness
	isStale, err = IsHeartbeatStale(heartbeatPath, heartbeatInterval)
	if err != nil {
		t.Fatalf("IsHeartbeatStale failed: %v", err)
	}
	if isStale {
		t.Errorf("heartbeat should NOT be detected as stale")
	}
}

// TestDaemonHeartbeatConfigInterval verifies HeartbeatInterval config field
// controls the recording interval.
func TestDaemonHeartbeatConfigInterval(t *testing.T) {
	tmpDir := t.TempDir()
	heartbeatPath := filepath.Join(tmpDir, "daemon.heartbeat")

	// Use a specific heartbeat interval
	heartbeatInterval := 100 * time.Millisecond

	cfg := &Config{
		PIDPath:           filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:        filepath.Join(tmpDir, "daemon.sock"),
		LogPath:           filepath.Join(tmpDir, "daemon.log"),
		HeartbeatPath:     heartbeatPath,
		HeartbeatInterval: heartbeatInterval,
		Interval:          1 * time.Hour,
		IdleTimeout:       0,
	}

	d := New(cfg)
	d.SetSyncFunc(func() error { return nil })

	// Start daemon
	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait for first heartbeat to be written (poll with timeout)
	var data1 []byte
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
		data1, _ = os.ReadFile(heartbeatPath)
		if len(data1) > 0 {
			break
		}
	}
	if len(data1) == 0 {
		t.Fatalf("heartbeat file should have content")
	}

	// Record first timestamp
	ts1, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(string(data1)))
	if err != nil {
		t.Fatalf("failed to parse timestamp: %v", err)
	}

	// Wait for another heartbeat interval
	time.Sleep(heartbeatInterval + 50*time.Millisecond)

	// Record second timestamp
	data2, err := os.ReadFile(heartbeatPath)
	if err != nil {
		t.Fatalf("failed to read heartbeat file: %v", err)
	}
	ts2, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(string(data2)))
	if err != nil {
		t.Fatalf("failed to parse timestamp: %v", err)
	}

	// Verify heartbeat was updated
	if !ts2.After(ts1) {
		t.Errorf("heartbeat should be updated over time: ts1=%v, ts2=%v", ts1, ts2)
	}

	// Verify update interval is approximately correct (within 2x tolerance for timing variance)
	diff := ts2.Sub(ts1)
	if diff > 2*heartbeatInterval {
		t.Errorf("heartbeat update interval should be close to %v, but was %v", heartbeatInterval, diff)
	}

	// Stop daemon
	d.Stop()
	<-done
}

// TestDaemonHealthCheckBeforeIPC verifies CLI can check heartbeat freshness
// before attempting socket communication.
func TestDaemonHealthCheckBeforeIPC(t *testing.T) {
	tmpDir := t.TempDir()
	pidPath := filepath.Join(tmpDir, "daemon.pid")
	socketPath := filepath.Join(tmpDir, "daemon.sock")
	heartbeatPath := filepath.Join(tmpDir, "daemon.heartbeat")
	heartbeatInterval := 50 * time.Millisecond

	// Test 1: No heartbeat file - daemon not started
	healthy, reason := CheckDaemonHealth(pidPath, socketPath, heartbeatPath, heartbeatInterval)
	if healthy {
		t.Errorf("daemon should be unhealthy when heartbeat file doesn't exist")
	}
	if reason != "heartbeat file not found" {
		t.Errorf("expected reason 'heartbeat file not found', got: %q", reason)
	}

	// Test 2: Stale heartbeat - daemon potentially hung
	staleTime := time.Now().Add(-3 * heartbeatInterval)
	if err := os.WriteFile(heartbeatPath, []byte(staleTime.Format(time.RFC3339Nano)), 0600); err != nil {
		t.Fatalf("failed to write stale heartbeat: %v", err)
	}

	healthy, reason = CheckDaemonHealth(pidPath, socketPath, heartbeatPath, heartbeatInterval)
	if healthy {
		t.Errorf("daemon should be unhealthy when heartbeat is stale")
	}
	if !strings.Contains(reason, "stale") {
		t.Errorf("expected reason to contain 'stale', got: %q", reason)
	}

	// Test 3: Fresh heartbeat - daemon healthy
	freshTime := time.Now()
	if err := os.WriteFile(heartbeatPath, []byte(freshTime.Format(time.RFC3339Nano)), 0600); err != nil {
		t.Fatalf("failed to write fresh heartbeat: %v", err)
	}

	healthy, reason = CheckDaemonHealth(pidPath, socketPath, heartbeatPath, heartbeatInterval)
	if !healthy {
		t.Errorf("daemon should be healthy when heartbeat is fresh, reason: %s", reason)
	}
	if reason != "healthy" {
		t.Errorf("expected reason 'healthy', got: %q", reason)
	}
}

// =============================================================================
// Issue #82: Daemon Error Loop Prevention with Exponential Backoff
// =============================================================================

// TestDaemonExponentialBackoff verifies backoff delays: 1s, 2s, 4s, 8s, 16s (max 60s)
func TestDaemonExponentialBackoff(t *testing.T) {
	// Test the backoff calculation function
	testCases := []struct {
		consecutiveErrors int
		expectedSeconds   float64
	}{
		{1, 2},   // 2^1 = 2
		{2, 4},   // 2^2 = 4
		{3, 8},   // 2^3 = 8
		{4, 16},  // 2^4 = 16
		{5, 32},  // 2^5 = 32
		{6, 60},  // 2^6 = 64, but capped at 60
		{10, 60}, // 2^10 = 1024, but capped at 60
	}

	for _, tc := range testCases {
		backoff := CalculateBackoff(tc.consecutiveErrors)
		expectedDuration := time.Duration(tc.expectedSeconds) * time.Second
		if backoff != expectedDuration {
			t.Errorf("CalculateBackoff(%d) = %v, want %v", tc.consecutiveErrors, backoff, expectedDuration)
		}
	}
}

// TestDaemonMaxConsecutiveErrors verifies daemon shuts down after 5 consecutive errors
func TestDaemonMaxConsecutiveErrors(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    10 * time.Millisecond, // Fast interval for testing
		IdleTimeout: 0,
	}

	d := New(cfg)

	// Count how many times sync was attempted
	var syncAttempts int32

	// Always fail the sync
	d.SetSyncFunc(func() error {
		atomic.AddInt32(&syncAttempts, 1)
		return fmt.Errorf("simulated sync failure")
	})

	// Enable fast backoff for testing (skip actual delays)
	d.SetTestBackoffMultiplier(0) // 0 = instant, no actual sleep

	// Start daemon
	done := make(chan struct{})
	startTime := time.Now()
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait for daemon to exit due to max consecutive errors
	select {
	case <-done:
		// Good, daemon exited
	case <-time.After(5 * time.Second):
		t.Fatalf("daemon should have exited due to MaxConsecutiveErrors")
		d.Stop()
	}

	elapsed := time.Since(startTime)

	// Verify daemon exited quickly (not waiting for actual backoff)
	if elapsed > 3*time.Second {
		t.Errorf("daemon should exit quickly with test multiplier, took %v", elapsed)
	}

	// Verify sync was attempted MaxConsecutiveErrors times
	attempts := atomic.LoadInt32(&syncAttempts)
	if attempts != int32(MaxConsecutiveErrors) {
		t.Errorf("expected %d sync attempts before shutdown, got %d", MaxConsecutiveErrors, attempts)
	}

	// Note: We don't check the log file here because cleanup() removes it.
	// The critical verification is that:
	// 1. The daemon exited (not timed out)
	// 2. It exited after exactly MaxConsecutiveErrors attempts
}

// TestDaemonErrorCountReset verifies successful operation resets consecutive error count
func TestDaemonErrorCountReset(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    10 * time.Millisecond,
		IdleTimeout: 0,
	}

	d := New(cfg)

	// Pattern: fail 3 times, succeed, fail 3 times, succeed, ...
	// This should never trigger MaxConsecutiveErrors since we reset after each success
	var syncAttempts int32
	var successCount int32

	d.SetSyncFunc(func() error {
		attempt := atomic.AddInt32(&syncAttempts, 1)
		// Fail first 3, succeed on 4th, fail next 3, succeed on 8th, etc.
		if attempt%4 == 0 {
			atomic.AddInt32(&successCount, 1)
			return nil // Success - should reset error count
		}
		return fmt.Errorf("simulated failure %d", attempt)
	})

	// Enable fast backoff for testing
	d.SetTestBackoffMultiplier(0)

	// Start daemon
	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait for enough sync cycles to verify error count reset
	// We need at least 12 attempts (3 successes at attempts 4, 8, 12)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&successCount) >= 3 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Stop daemon
	d.Stop()
	<-done

	// Daemon should NOT have exited due to MaxConsecutiveErrors
	// (it should have been stopped by us)
	if atomic.LoadInt32(&successCount) < 3 {
		t.Errorf("expected at least 3 successful syncs to verify error reset, got %d", successCount)
	}

	// Note: We don't check the log file here because cleanup() removes it.
	// The critical verification is that the daemon:
	// 1. Continued running through multiple fail/success cycles
	// 2. Had to be stopped manually (not auto-shutdown due to errors)
	// 3. Achieved at least 3 successful syncs
}

// =============================================================================
// Issue #84: Per-Task Timeout Protection for Sync Operations
// =============================================================================

// TestTaskProcessingTimeout verifies that task exceeding timeout is cancelled.
// Issue #84: Per-task timeout protection for sync operations.
func TestTaskProcessingTimeout(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    100 * time.Millisecond,
		IdleTimeout: 0,
		TaskTimeout: 50 * time.Millisecond, // Short timeout for testing
	}

	d := New(cfg)

	// Track whether sync was interrupted
	var syncStarted, syncCompleted int32

	// Backend sync that takes longer than the task timeout
	d.AddBackendSyncFunc("slow_backend", func() error {
		atomic.AddInt32(&syncStarted, 1)
		time.Sleep(200 * time.Millisecond) // Takes longer than 50ms timeout
		atomic.AddInt32(&syncCompleted, 1)
		return nil
	})

	// Start daemon
	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait for at least one sync attempt
	time.Sleep(300 * time.Millisecond)

	// Stop daemon
	d.Stop()
	<-done

	// Verify sync was started
	if atomic.LoadInt32(&syncStarted) == 0 {
		t.Errorf("expected sync to start at least once")
	}

	// Verify the backend state shows failure due to timeout
	state := d.GetBackendState("slow_backend")
	if state == nil {
		t.Fatalf("expected state for slow_backend")
	}
	if state.ErrorCount == 0 {
		t.Errorf("expected error count > 0 for timed out backend, got 0")
	}
	if state.LastError == "" || !strings.Contains(state.LastError, "timeout") {
		t.Errorf("expected timeout error in LastError, got: %q", state.LastError)
	}
}

// TestTaskTimeoutMarksFailure verifies that timed out task is marked as failed in queue.
// Issue #84: Per-task timeout protection for sync operations.
func TestTaskTimeoutMarksFailure(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    100 * time.Millisecond,
		IdleTimeout: 0,
		TaskTimeout: 30 * time.Millisecond, // Short timeout for testing
	}

	d := New(cfg)

	var timeoutCount int32

	// Backend sync that blocks indefinitely (until timeout)
	d.AddBackendSyncFunc("blocking_backend", func() error {
		time.Sleep(100 * time.Millisecond) // Will timeout
		return nil
	})

	// Register callback to track timeout events
	d.SetOnTaskTimeout(func(backendName string, duration time.Duration) {
		atomic.AddInt32(&timeoutCount, 1)
	})

	// Start daemon
	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait for timeout events
	time.Sleep(250 * time.Millisecond)

	// Stop daemon
	d.Stop()
	<-done

	// Verify timeout was detected
	if atomic.LoadInt32(&timeoutCount) == 0 {
		t.Errorf("expected at least one timeout event, got 0")
	}

	// Verify backend state shows error
	state := d.GetBackendState("blocking_backend")
	if state == nil {
		t.Fatalf("expected state for blocking_backend")
	}
	if state.ErrorCount == 0 {
		t.Errorf("expected error count > 0 for timed out backend")
	}
}

// TestTaskTimeoutContextCancellation verifies that Context.Done() is properly handled.
// Issue #84: Per-task timeout protection for sync operations.
func TestTaskTimeoutContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    100 * time.Millisecond,
		IdleTimeout: 0,
		TaskTimeout: 30 * time.Millisecond, // Short timeout for testing
	}

	d := New(cfg)

	var contextCancelled int32

	// Backend sync that respects context cancellation
	d.AddBackendSyncFuncWithContext("context_aware_backend", func(ctx context.Context) error {
		select {
		case <-time.After(100 * time.Millisecond):
			return nil // Would complete normally but timeout is shorter
		case <-ctx.Done():
			atomic.AddInt32(&contextCancelled, 1)
			return ctx.Err()
		}
	})

	// Start daemon
	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait for context cancellation
	time.Sleep(250 * time.Millisecond)

	// Stop daemon
	d.Stop()
	<-done

	// Verify context was cancelled due to timeout
	if atomic.LoadInt32(&contextCancelled) == 0 {
		t.Errorf("expected context to be cancelled at least once due to timeout")
	}

	// Verify backend state shows timeout error
	state := d.GetBackendState("context_aware_backend")
	if state == nil {
		t.Fatalf("expected state for context_aware_backend")
	}
	if state.ErrorCount == 0 {
		t.Errorf("expected error count > 0 for backend with cancelled context")
	}
}

// TestTaskTimeoutDefaultValue verifies default MaxTaskDuration is 5 minutes.
// Issue #84: Default MaxTaskDuration = 5 minutes.
func TestTaskTimeoutDefaultValue(t *testing.T) {
	if DefaultTaskTimeout != 5*time.Minute {
		t.Errorf("expected DefaultTaskTimeout to be 5 minutes, got %v", DefaultTaskTimeout)
	}
}

// TestTaskTimeoutConfigurable verifies TaskTimeout can be set via Config.
// Issue #84: Add task_timeout config option to daemon section.
func TestTaskTimeoutConfigurable(t *testing.T) {
	tmpDir := t.TempDir()

	customTimeout := 10 * time.Minute

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    100 * time.Millisecond,
		IdleTimeout: 0,
		TaskTimeout: customTimeout,
	}

	d := New(cfg)

	// Verify the configured timeout is used
	if d.cfg.TaskTimeout != customTimeout {
		t.Errorf("expected TaskTimeout %v, got %v", customTimeout, d.cfg.TaskTimeout)
	}
}

// TestTaskTimeoutLogging verifies timeout events are logged with task ID and duration.
// Issue #84: Log timeout events with task ID and duration.
func TestTaskTimeoutLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "daemon.log")

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     logPath,
		Interval:    100 * time.Millisecond,
		IdleTimeout: 0,
		TaskTimeout: 30 * time.Millisecond, // Short timeout for testing
	}

	d := New(cfg)

	// Backend sync that will timeout
	d.AddBackendSyncFunc("log_test_backend", func() error {
		time.Sleep(100 * time.Millisecond) // Will timeout
		return nil
	})

	// Start daemon
	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait for timeout event to be logged
	time.Sleep(200 * time.Millisecond)

	// Read log file BEFORE stopping daemon (cleanup() removes it)
	logContent, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Stop daemon
	d.Stop()
	<-done

	logStr := string(logContent)
	if !strings.Contains(logStr, "timed out") || !strings.Contains(logStr, "log_test_backend") {
		t.Errorf("expected log to contain timeout message with backend name, got: %s", logStr)
	}
}

// TestDaemonHeartbeatDisabled verifies daemon works without heartbeat when interval is 0.
func TestDaemonHeartbeatDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	heartbeatPath := filepath.Join(tmpDir, "daemon.heartbeat")

	cfg := &Config{
		PIDPath:           filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:        filepath.Join(tmpDir, "daemon.sock"),
		LogPath:           filepath.Join(tmpDir, "daemon.log"),
		HeartbeatPath:     heartbeatPath,
		HeartbeatInterval: 0, // Disabled
		Interval:          100 * time.Millisecond,
		IdleTimeout:       0,
	}

	d := New(cfg)
	d.SetSyncFunc(func() error { return nil })

	// Start daemon
	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait a bit
	time.Sleep(200 * time.Millisecond)

	// Heartbeat file should NOT be created when disabled
	if _, err := os.Stat(heartbeatPath); err == nil {
		t.Errorf("heartbeat file should NOT be created when HeartbeatInterval is 0")
	}

	// Stop daemon
	d.Stop()
	<-done
}

// TestForkPassesTaskTimeout verifies that Fork passes --daemon-task-timeout when TaskTimeout is set.
// Issue #98: Per-task timeout not passed to forked daemon process.
func TestForkPassesTaskTimeout(t *testing.T) {
	cfg := &Config{
		PIDPath:     "/tmp/test.pid",
		SocketPath:  "/tmp/test.sock",
		LogPath:     "/tmp/test.log",
		Interval:    5 * time.Minute,
		TaskTimeout: 10 * time.Minute,
	}

	args := buildForkArgs(cfg)

	// Verify --daemon-task-timeout is present with the correct value
	found := false
	for i, arg := range args {
		if arg == "--daemon-task-timeout" {
			if i+1 >= len(args) {
				t.Fatal("--daemon-task-timeout flag has no value")
			}
			if args[i+1] != "10" {
				t.Errorf("expected --daemon-task-timeout value '10', got '%s'", args[i+1])
			}
			found = true
			break
		}
	}
	if !found {
		t.Errorf("--daemon-task-timeout not found in fork args: %v", args)
	}
}

// TestMultiBackendConsecutiveErrors verifies daemon shuts down after MaxConsecutiveErrors
// when ALL backends fail consecutively in multi-backend mode.
// Issue #100: Multi-backend daemon bypasses error loop prevention.
func TestMultiBackendConsecutiveErrors(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    10 * time.Millisecond,
		IdleTimeout: 0,
		TaskTimeout: 1 * time.Second,
	}

	d := New(cfg)

	// Count sync attempts per backend
	var backendASyncs int32
	var backendBSyncs int32

	// Both backends always fail
	d.AddBackendSyncFunc("backendA", func() error {
		atomic.AddInt32(&backendASyncs, 1)
		return fmt.Errorf("backendA always fails")
	})
	d.AddBackendSyncFunc("backendB", func() error {
		atomic.AddInt32(&backendBSyncs, 1)
		return fmt.Errorf("backendB always fails")
	})

	d.SetTestBackoffMultiplier(0) // No actual sleep

	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	select {
	case <-done:
		// Good - daemon exited due to all-backends-failing error loop prevention
	case <-time.After(5 * time.Second):
		t.Fatalf("daemon should have exited due to MaxConsecutiveErrors in multi-backend mode")
		d.Stop()
	}

	// Issue #114: With circuit breaker, each backend is called DefaultCircuitBreakerThreshold times
	// before the circuit opens. After that, circuit-blocked ticks count as global failures
	// until MaxConsecutiveErrors is reached and the daemon shuts down.
	aSyncs := atomic.LoadInt32(&backendASyncs)
	bSyncs := atomic.LoadInt32(&backendBSyncs)
	if aSyncs != int32(DefaultCircuitBreakerThreshold) {
		t.Errorf("expected backendA to sync %d times (circuit breaker threshold), got %d", DefaultCircuitBreakerThreshold, aSyncs)
	}
	if bSyncs != int32(DefaultCircuitBreakerThreshold) {
		t.Errorf("expected backendB to sync %d times (circuit breaker threshold), got %d", DefaultCircuitBreakerThreshold, bSyncs)
	}
}

// TestMultiBackendErrorResetOnPartialSuccess verifies that if at least one backend
// succeeds, the consecutive error counter resets (no shutdown).
// Issue #100: Multi-backend daemon bypasses error loop prevention.
func TestMultiBackendErrorResetOnPartialSuccess(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    10 * time.Millisecond,
		IdleTimeout: 0,
		TaskTimeout: 1 * time.Second,
	}

	d := New(cfg)

	var totalSyncs int32

	// backendA always fails, backendB always succeeds
	d.AddBackendSyncFunc("backendA", func() error {
		atomic.AddInt32(&totalSyncs, 1)
		return fmt.Errorf("backendA always fails")
	})
	d.AddBackendSyncFunc("backendB", func() error {
		atomic.AddInt32(&totalSyncs, 1)
		return nil
	})

	d.SetTestBackoffMultiplier(0)

	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait for enough syncs to prove daemon doesn't shut down
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&totalSyncs) >= int32(MaxConsecutiveErrors*2+4) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	d.Stop()
	<-done

	// Daemon should NOT have shut down on its own - it ran well past MaxConsecutiveErrors
	syncs := atomic.LoadInt32(&totalSyncs)
	if syncs < int32(MaxConsecutiveErrors*2+4) {
		t.Errorf("expected daemon to keep running with partial success, but only got %d total syncs", syncs)
	}
}

// =============================================================================
// Issue #114: Circuit Breaker Pattern for Backend-Specific Daemon Errors
// =============================================================================

// TestCircuitBreakerOpensOnConsecutiveFailures verifies the circuit opens after
// N consecutive failures for a specific backend.
func TestCircuitBreakerOpensOnConsecutiveFailures(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)

	// Initially closed - should allow requests
	if !cb.Allow() {
		t.Fatal("circuit should allow requests when closed")
	}
	if cb.State() != CircuitClosed {
		t.Fatalf("expected state Closed, got %v", cb.State())
	}

	// Record 2 failures - should still be closed
	cb.RecordFailure()
	cb.RecordFailure()
	if !cb.Allow() {
		t.Fatal("circuit should still allow after 2 failures (threshold=3)")
	}
	if cb.State() != CircuitClosed {
		t.Fatalf("expected state Closed after 2 failures, got %v", cb.State())
	}

	// Record 3rd failure - should open
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatalf("expected state Open after 3 failures, got %v", cb.State())
	}
	if cb.Allow() {
		t.Fatal("circuit should NOT allow requests when open")
	}
}

// TestCircuitBreakerHalfOpenRecovery verifies that after cooldown, the circuit
// enters half-open state and allows a probe request.
func TestCircuitBreakerHalfOpenRecovery(t *testing.T) {
	cooldown := 50 * time.Millisecond
	cb := NewCircuitBreaker(2, cooldown)

	// Trip the circuit
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatalf("expected Open, got %v", cb.State())
	}
	if cb.Allow() {
		t.Fatal("should not allow when open")
	}

	// Wait for cooldown
	time.Sleep(cooldown + 10*time.Millisecond)

	// Should now be half-open and allow one probe
	if cb.State() != CircuitHalfOpen {
		t.Fatalf("expected HalfOpen after cooldown, got %v", cb.State())
	}
	if !cb.Allow() {
		t.Fatal("half-open circuit should allow one probe request")
	}

	// Probe succeeds → circuit should close
	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Fatalf("expected Closed after successful probe, got %v", cb.State())
	}
	if cb.FailureCount() != 0 {
		t.Fatalf("failure count should be 0 after reset, got %d", cb.FailureCount())
	}

	// Now test: probe fails → circuit re-opens
	cb2 := NewCircuitBreaker(2, cooldown)
	cb2.RecordFailure()
	cb2.RecordFailure()
	time.Sleep(cooldown + 10*time.Millisecond)

	// Half-open, allow probe
	if !cb2.Allow() {
		t.Fatal("should allow probe in half-open")
	}
	// Probe fails
	cb2.RecordFailure()
	if cb2.State() != CircuitOpen {
		t.Fatalf("expected Open after failed probe, got %v", cb2.State())
	}
}

// TestCircuitBreakerPerBackendIsolation verifies one backend's circuit breaker
// state does not affect other backends.
func TestCircuitBreakerPerBackendIsolation(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    50 * time.Millisecond,
		IdleTimeout: 0,
		TaskTimeout: 1 * time.Second,
	}

	d := New(cfg)
	d.SetTestBackoffMultiplier(0)

	var successSyncs int32
	var failAttempts int32

	// backendA: always fails → circuit should open
	d.AddBackendSyncFunc("backendA", func() error {
		atomic.AddInt32(&failAttempts, 1)
		return fmt.Errorf("backendA down")
	})

	// backendB: always succeeds → circuit stays closed
	d.AddBackendSyncFunc("backendB", func() error {
		atomic.AddInt32(&successSyncs, 1)
		return nil
	})

	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Let it run for a while - enough for backendA's circuit to open
	time.Sleep(400 * time.Millisecond)

	d.Stop()
	<-done

	fails := atomic.LoadInt32(&failAttempts)
	successes := atomic.LoadInt32(&successSyncs)

	// backendA should have been called up to the threshold (3) then blocked by circuit
	// backendB should have been called many more times since its circuit stays closed
	if successes < 4 {
		t.Errorf("backendB should have synced many times, got %d", successes)
	}

	// backendA's circuit should have opened after DefaultCircuitBreakerThreshold failures,
	// so it should have fewer total attempts than backendB
	if fails >= successes {
		t.Errorf("backendA (failing, circuit should open) attempts (%d) should be less than backendB successes (%d)", fails, successes)
	}

	// Verify circuit breaker states
	stateA := d.GetBackendState("backendA")
	stateB := d.GetBackendState("backendB")
	if stateA == nil || stateB == nil {
		t.Fatal("expected backend states to exist")
	}
	// backendA should show circuit open
	if stateA.CircuitState != CircuitOpen.String() {
		t.Errorf("backendA circuit should be Open, got %s", stateA.CircuitState)
	}
	// backendB should show circuit closed
	if stateB.CircuitState != CircuitClosed.String() {
		t.Errorf("backendB circuit should be Closed, got %s", stateB.CircuitState)
	}
}

// TestCircuitBreakerResetOnSuccess verifies that a successful sync closes the
// circuit and resets the failure count.
func TestCircuitBreakerResetOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(3, 50*time.Millisecond)

	// Record 2 failures (below threshold)
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.FailureCount() != 2 {
		t.Fatalf("expected 2 failures, got %d", cb.FailureCount())
	}

	// Success resets failure count
	cb.RecordSuccess()
	if cb.FailureCount() != 0 {
		t.Fatalf("expected 0 failures after success, got %d", cb.FailureCount())
	}
	if cb.State() != CircuitClosed {
		t.Fatalf("expected Closed after success, got %v", cb.State())
	}

	// Trip it fully, then recover via half-open probe
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatalf("expected Open, got %v", cb.State())
	}

	// Wait for cooldown, then succeed
	time.Sleep(60 * time.Millisecond)
	if !cb.Allow() {
		t.Fatal("should allow probe in half-open")
	}
	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Fatalf("expected Closed after successful probe, got %v", cb.State())
	}
	if cb.FailureCount() != 0 {
		t.Fatalf("expected 0 failures, got %d", cb.FailureCount())
	}
	if !cb.Allow() {
		t.Fatal("should allow requests after reset")
	}
}

// TestCircuitBreakerStatusVisible verifies circuit breaker state appears in daemon status output.
func TestCircuitBreakerStatusVisible(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		PIDPath:     filepath.Join(tmpDir, "daemon.pid"),
		SocketPath:  filepath.Join(tmpDir, "daemon.sock"),
		LogPath:     filepath.Join(tmpDir, "daemon.log"),
		Interval:    50 * time.Millisecond,
		IdleTimeout: 0,
		TaskTimeout: 1 * time.Second,
	}

	d := New(cfg)
	d.SetTestBackoffMultiplier(0)

	// backendA always fails
	d.AddBackendSyncFunc("backendA", func() error {
		return fmt.Errorf("always fails")
	})
	// backendB always succeeds
	d.AddBackendSyncFunc("backendB", func() error {
		return nil
	})

	done := make(chan struct{})
	go func() {
		_ = d.Start()
		close(done)
	}()

	// Wait for circuit to open on backendA
	time.Sleep(300 * time.Millisecond)

	// Check status via getBackendStatuses
	statuses := d.getBackendStatuses()
	if statuses == nil {
		t.Fatal("expected backend statuses")
	}

	statusA := statuses["backendA"]
	statusB := statuses["backendB"]
	if statusA == nil || statusB == nil {
		t.Fatal("expected statuses for both backends")
	}

	if statusA.CircuitState != CircuitOpen.String() {
		t.Errorf("backendA status should show circuit Open, got %q", statusA.CircuitState)
	}
	if statusB.CircuitState != CircuitClosed.String() {
		t.Errorf("backendB status should show circuit Closed, got %q", statusB.CircuitState)
	}

	d.Stop()
	<-done
}

// TestForkOmitsTaskTimeoutWhenZero verifies that Fork omits --daemon-task-timeout when TaskTimeout is zero.
// Issue #98: Per-task timeout not passed to forked daemon process.
func TestForkOmitsTaskTimeoutWhenZero(t *testing.T) {
	cfg := &Config{
		PIDPath:    "/tmp/test.pid",
		SocketPath: "/tmp/test.sock",
		LogPath:    "/tmp/test.log",
		Interval:   5 * time.Minute,
	}

	args := buildForkArgs(cfg)

	for _, arg := range args {
		if arg == "--daemon-task-timeout" {
			t.Errorf("--daemon-task-timeout should not be in fork args when TaskTimeout is zero: %v", args)
		}
	}
}
