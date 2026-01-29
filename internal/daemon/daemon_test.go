package daemon

import (
	"os"
	"path/filepath"
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
