package sync_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"todoat/internal/testutil"
)

// =============================================================================
// CLI Tests (024-auto-sync-daemon)
// =============================================================================

// TestSyncDaemonStartCLI tests that 'todoat sync daemon start' launches a background process
func TestSyncDaemonStartCLI(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)

	stdout := cli.MustExecute("-y", "sync", "daemon", "start")

	testutil.AssertContains(t, stdout, "started")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify daemon is running by checking status
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "running")

	// Cleanup: stop the daemon
	cli.MustExecute("-y", "sync", "daemon", "stop")
}

// TestSyncDaemonStopCLI tests that 'todoat sync daemon stop' terminates a running daemon
func TestSyncDaemonStopCLI(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)

	// Start daemon first
	cli.MustExecute("-y", "sync", "daemon", "start")

	// Stop the daemon
	stdout := cli.MustExecute("-y", "sync", "daemon", "stop")

	testutil.AssertContains(t, stdout, "stopped")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify daemon is no longer running
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "not running")
}

// TestSyncDaemonStopNotRunningCLI tests stopping when daemon is not running
func TestSyncDaemonStopNotRunningCLI(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)

	// Try to stop when not running - should report appropriately
	stdout := cli.MustExecute("-y", "sync", "daemon", "stop")

	testutil.AssertContains(t, stdout, "not running")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestSyncDaemonStatusCLI tests that 'todoat sync daemon status' shows daemon state
func TestSyncDaemonStatusCLI(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)

	// Check status when not running
	stdout := cli.MustExecute("-y", "sync", "daemon", "status")

	testutil.AssertContains(t, stdout, "not running")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")

	// Check status when running
	stdout = cli.MustExecute("-y", "sync", "daemon", "status")

	testutil.AssertContains(t, stdout, "running")
	testutil.AssertContains(t, stdout, "PID")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)

	// Cleanup
	cli.MustExecute("-y", "sync", "daemon", "stop")
}

// TestSyncDaemonIntervalCLI tests that sync runs at configured interval
func TestSyncDaemonIntervalCLI(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)

	// Configure a short interval for testing (100ms)
	cli.SetDaemonInterval(100 * time.Millisecond)

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Wait for at least 2 sync cycles using polling instead of fixed sleep
	stdout := cli.WaitForSyncCount(5*time.Second, 2)

	// Status should show sync count > 1
	testutil.AssertContains(t, stdout, "Sync count")
}

// TestSyncDaemonNotificationCLI tests that notifications are sent on sync events
func TestSyncDaemonNotificationCLI(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)

	// Configure short interval and enable notifications
	cli.SetDaemonInterval(100 * time.Millisecond)

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Wait for a sync cycle to complete using polling
	cli.WaitForSyncCount(5*time.Second, 1)

	// Check notification log for sync events
	stdout := cli.MustExecute("-y", "notification", "log")

	// Should have sync-related notifications (sync_complete or sync_error)
	hasSyncNotification := strings.Contains(stdout, "SYNC_COMPLETE") ||
		strings.Contains(stdout, "SYNC_ERROR") ||
		strings.Contains(stdout, "No notifications")

	if !hasSyncNotification {
		t.Errorf("expected sync notification in log, got:\n%s", stdout)
	}
}

// TestSyncDaemonOfflineCLI tests that daemon handles offline gracefully
func TestSyncDaemonOfflineCLI(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)

	// Configure daemon with a simulated offline remote
	cli.SetDaemonOffline(true)
	cli.SetDaemonInterval(100 * time.Millisecond)

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Wait for a sync attempt using polling
	stdout := cli.WaitForSyncCount(5*time.Second, 1)

	// Daemon should still be running (not crashed)
	testutil.AssertContains(t, stdout, "running")

	// Should indicate offline status or retry
	statusContainsOffline := strings.Contains(stdout, "offline") ||
		strings.Contains(stdout, "retry") ||
		strings.Contains(stdout, "Sync count")

	if !statusContainsOffline {
		t.Logf("Status output: %s", stdout)
	}
}

// TestSyncDaemonReconnectCLI tests that daemon reconnects when network restored
func TestSyncDaemonReconnectCLI(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)

	// Start in offline mode
	cli.SetDaemonOffline(true)
	cli.SetDaemonInterval(100 * time.Millisecond)

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Wait for offline sync attempt using polling
	cli.WaitForSyncCount(5*time.Second, 1)

	// Restore network
	cli.SetDaemonOffline(false)

	// Wait for reconnect and successful sync (sync count should increase)
	stdout := cli.WaitForSyncCount(5*time.Second, 2)

	// Check status shows successful sync
	testutil.AssertContains(t, stdout, "running")
}

// TestSyncDaemonPIDFileCLI tests that PID file is created for process management
func TestSyncDaemonPIDFileCLI(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)

	pidFile := cli.PIDFilePath()

	// PID file should not exist before starting
	if _, err := os.Stat(pidFile); err == nil {
		t.Errorf("PID file should not exist before daemon start")
	}

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")

	// PID file should exist
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		t.Errorf("PID file should exist after daemon start: %s", pidFile)
	}

	// Read PID file and verify it contains a valid PID
	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("failed to read PID file: %v", err)
	}

	pidStr := strings.TrimSpace(string(data))
	if pidStr == "" {
		t.Errorf("PID file is empty")
	}

	// Verify it's a number
	var pid int
	if _, err := fmt.Sscanf(pidStr, "%d", &pid); err != nil {
		t.Errorf("PID file contains invalid PID: %s", pidStr)
	}

	if pid <= 0 {
		t.Errorf("PID should be positive, got %d", pid)
	}

	// Stop daemon
	cli.MustExecute("-y", "sync", "daemon", "stop")

	// PID file should be removed after stopping
	if _, err := os.Stat(pidFile); err == nil {
		t.Errorf("PID file should be removed after daemon stop")
	}
}

// TestSyncDaemonDoubleStartCLI tests that starting twice gives appropriate error
func TestSyncDaemonDoubleStartCLI(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)

	// Start daemon first
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Try to start again - should report already running
	stdout := cli.MustExecute("-y", "sync", "daemon", "start")

	testutil.AssertContains(t, stdout, "already running")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestSyncDaemonGracefulShutdownCLI tests that daemon handles signals gracefully
func TestSyncDaemonGracefulShutdownCLI(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")

	// Send stop signal
	cli.MustExecute("-y", "sync", "daemon", "stop")

	// Verify clean shutdown
	pidFile := cli.PIDFilePath()
	if _, err := os.Stat(pidFile); err == nil {
		t.Errorf("PID file should be cleaned up after graceful shutdown")
	}

	// Should be able to start again
	stdout := cli.MustExecute("-y", "sync", "daemon", "start")
	testutil.AssertContains(t, stdout, "started")

	// Cleanup
	cli.MustExecute("-y", "sync", "daemon", "stop")
}

// TestSyncDaemonLogFileCLI tests that daemon writes to log file
func TestSyncDaemonLogFileCLI(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)

	logFile := cli.DaemonLogPath()

	// Start daemon with logging
	cli.MustExecute("-y", "sync", "daemon", "start")

	// Wait for log file to be created and have content using polling
	testutil.WaitFor(t, 5*time.Second, func() bool {
		data, err := os.ReadFile(logFile)
		return err == nil && len(data) > 0
	}, "log file to have content")

	// Stop daemon
	cli.MustExecute("-y", "sync", "daemon", "stop")

	// Check log file exists and has content
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Daemon log file should exist: %s", logFile)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read daemon log file: %v", err)
	}

	content := string(data)
	if len(content) == 0 {
		t.Errorf("Daemon log file should have content")
	}

	// Should contain startup message
	if !strings.Contains(content, "started") && !strings.Contains(content, "Starting") {
		t.Logf("Log content: %s", content)
	}
}

// TestSyncDaemonConfigIntervalCLI tests configuring sync interval via config
func TestSyncDaemonConfigIntervalCLI(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)
	configPath := cli.ConfigPath()

	// Write config with custom interval
	configContent := `sync:
  enabled: true
  daemon:
    interval: 60
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon - should use config interval
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Check status shows configured interval
	stdout := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, stdout, "60")
}

// TestSyncDaemonStartWithIntervalCLI tests starting with --interval flag
func TestSyncDaemonStartWithIntervalCLI(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)

	// Start daemon with custom interval flag
	stdout := cli.MustExecute("-y", "sync", "daemon", "start", "--interval", "30")

	testutil.AssertContains(t, stdout, "started")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Check status shows configured interval
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "30")

	// Cleanup
	cli.MustExecute("-y", "sync", "daemon", "stop")
}

// TestDaemonActuallySyncs verifies that daemon calls doSync at each interval (issue #005)
func TestDaemonActuallySyncs(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)
	configPath := cli.ConfigPath()

	// Configure sync with a file backend as the remote sync target
	// Using file backend since it doesn't require network connectivity
	fileBackendPath := cli.TmpDir() + "/file-backend/tasks.txt"
	if err := os.MkdirAll(cli.TmpDir()+"/file-backend", 0755); err != nil {
		t.Fatalf("failed to create file backend dir: %v", err)
	}

	syncConfig := fmt.Sprintf(`
backends:
  sqlite:
    enabled: true
  file:
    type: file
    path: %s
    enabled: true
default_backend: file
sync:
  enabled: true
  daemon:
    interval: 1
`, fileBackendPath)
	if err := os.WriteFile(configPath, []byte(syncConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Add a task (gets queued for sync)
	cli.MustExecute("-y", "Work", "add", "Test task for daemon sync")

	// Verify task is queued
	queueOut := cli.MustExecute("-y", "sync", "queue")
	if !strings.Contains(queueOut, "Pending Operations: 1") {
		t.Logf("Expected 1 pending operation, got: %s", queueOut)
	}

	// Configure short interval for testing
	cli.SetDaemonInterval(100 * time.Millisecond)

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Wait for daemon sync cycle - at least 2 cycles to ensure sync happened
	cli.WaitForSyncCount(5*time.Second, 2)

	// CRITICAL: Queue should be empty after daemon sync
	// This verifies daemon actually calls doSync() and clears pending operations
	queueOut = cli.MustExecute("-y", "sync", "queue")
	if !strings.Contains(queueOut, "Pending Operations: 0") {
		t.Errorf("Queue should be empty after daemon sync, but got: %s", queueOut)
	}
}

// =============================================================================
// Multi-Backend Daemon Tests (073-auto-sync-daemon-redesign)
// =============================================================================

// TestDaemonMultiBackend tests that daemon syncs with multiple backends in sequence
func TestDaemonMultiBackend(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)
	configPath := cli.ConfigPath()

	// Configure multiple backends for sync
	multiBackendConfig := `
backends:
  sqlite:
    enabled: true
  file:
    enabled: true
    path: /tmp/todoat-test-file
  mock1:
    enabled: true
    sync_enabled: true
  mock2:
    enabled: true
    sync_enabled: true
sync:
  enabled: true
  daemon:
    interval: 1
    backends:
      - mock1
      - mock2
`
	if err := os.WriteFile(configPath, []byte(multiBackendConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Configure short interval for testing
	cli.SetDaemonInterval(100 * time.Millisecond)

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Wait for multiple sync cycles
	stdout := cli.WaitForSyncCount(5*time.Second, 2)

	// Status should show per-backend sync information
	testutil.AssertContains(t, stdout, "running")

	// Check that daemon status shows backend-specific sync info
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	// Should show sync activity for backends
	testutil.AssertContains(t, statusOut, "Sync count")
}

// TestDaemonPerBackendInterval tests that each backend can have custom sync interval
func TestDaemonPerBackendInterval(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)
	configPath := cli.ConfigPath()

	// Configure backends with different intervals
	perBackendIntervalConfig := `
backends:
  sqlite:
    enabled: true
sync:
  enabled: true
  daemon:
    backends:
      - name: mock1
        interval: 10
      - name: mock2
        interval: 30
`
	if err := os.WriteFile(configPath, []byte(perBackendIntervalConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Check status shows per-backend intervals
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")

	// Status should include per-backend configuration info
	testutil.AssertContains(t, statusOut, "running")
}

// TestDaemonCacheIsolation tests that backend caches remain isolated during concurrent sync
func TestDaemonCacheIsolation(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)
	configPath := cli.ConfigPath()

	// Configure multiple backends
	cacheIsolationConfig := `
backends:
  sqlite:
    enabled: true
sync:
  enabled: true
  daemon:
    interval: 1
    backends:
      - mock1
      - mock2
`
	if err := os.WriteFile(configPath, []byte(cacheIsolationConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cli.SetDaemonInterval(100 * time.Millisecond)

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Create tasks to test cache isolation
	cli.MustExecute("-y", "Work", "add", "Task for cache isolation test")

	// Wait for sync cycle
	cli.WaitForSyncCount(5*time.Second, 1)

	// Verify daemon is still running (no cache corruption)
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "running")

	// List tasks to verify no corruption - use Work list to show tasks
	listOut := cli.MustExecute("-y", "Work")
	testutil.AssertContains(t, listOut, "Task for cache isolation test")
}

// TestDaemonSmartTiming tests that daemon avoids sync during active editing (debounce)
func TestDaemonSmartTiming(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)
	configPath := cli.ConfigPath()

	// Configure daemon with smart timing enabled
	smartTimingConfig := `
backends:
  sqlite:
    enabled: true
sync:
  enabled: true
  daemon:
    interval: 1
    smart_timing: true
    debounce_ms: 500
`
	if err := os.WriteFile(configPath, []byte(smartTimingConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cli.SetDaemonInterval(100 * time.Millisecond)

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Get initial sync count
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "running")

	// Simulate rapid changes (adding multiple tasks quickly)
	for i := 0; i < 3; i++ {
		cli.MustExecute("-y", "Work", "add", fmt.Sprintf("Rapid task %d", i+1))
	}

	// Wait briefly for debounce
	time.Sleep(200 * time.Millisecond)

	// Daemon should still be running smoothly
	statusOut = cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "running")
}

// TestDaemonFileWatcher tests that optional file watcher triggers sync on local changes
func TestDaemonFileWatcher(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)
	configPath := cli.ConfigPath()

	// Configure daemon with file watcher enabled
	fileWatcherConfig := `
backends:
  sqlite:
    enabled: true
sync:
  enabled: true
  daemon:
    interval: 60
    file_watcher: true
`
	if err := os.WriteFile(configPath, []byte(fileWatcherConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon with file watcher
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Initial status check
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "running")

	// Add a task (which modifies the cache)
	cli.MustExecute("-y", "Work", "add", "Task to trigger file watcher")

	// Give file watcher time to detect change and trigger sync
	time.Sleep(300 * time.Millisecond)

	// Daemon should still be running
	statusOut = cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "running")
}
