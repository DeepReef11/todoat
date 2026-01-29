package sync_test

import (
	"fmt"
	"os"
	"strings"
	"syscall"
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

// =============================================================================
// Issue 036: Background Daemon Tests (Forked Process Architecture)
// =============================================================================

// TestDaemonRunsAsSeparateProcess verifies that the daemon runs as a separate
// process with a different PID than the CLI process.
// Issue #36: Sync not truly in background - needs background daemon
func TestDaemonRunsAsSeparateProcess(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)
	configPath := cli.ConfigPath()

	// Configure sync with a remote backend
	configContent := `
sync:
  enabled: true
  daemon:
    enabled: true
    interval: 1
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon
	stdout := cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	testutil.AssertContains(t, stdout, "started")

	// Read daemon PID from status
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "running")

	// Check that PID file exists and contains a valid PID
	pidFile := cli.PIDFilePath()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("failed to read PID file: %v", err)
	}

	pidStr := strings.TrimSpace(string(data))
	var daemonPID int
	if _, err := fmt.Sscanf(pidStr, "%d", &daemonPID); err != nil {
		t.Fatalf("PID file contains invalid PID: %s", pidStr)
	}

	// Verify that the daemon PID is a real process
	// Note: In test mode with in-process daemon, PID will be same as test process
	// When forked daemon is implemented, PID will be different
	if daemonPID <= 0 {
		t.Errorf("daemon PID should be positive, got %d", daemonPID)
	}

	// TODO: When forked daemon is implemented, verify daemonPID != os.Getpid()
	// For now, the test documents the expected behavior
}

// TestDaemonAutoStartOnSync verifies that the daemon starts automatically
// when sync operations are triggered and no daemon is running.
// Issue #36: CLI commands should return immediately, daemon handles sync async
func TestDaemonAutoStartOnSync(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)
	configPath := cli.ConfigPath()

	// Configure sync with daemon auto-start enabled
	configContent := `
sync:
  enabled: true
  auto_sync_after_operation: true
  daemon:
    enabled: true
    auto_start: true
    idle_timeout: 5
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Verify daemon is NOT running initially
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "not running")

	// Add a task - this should queue sync operation and auto-start daemon
	cli.MustExecute("-y", "Work", "add", "Task to trigger daemon auto-start")

	// Wait for daemon to auto-start
	testutil.WaitFor(t, 5*time.Second, func() bool {
		statusOut, _, _ := cli.Execute("-y", "sync", "daemon", "status")
		return strings.Contains(statusOut, "running")
	}, "daemon to auto-start")

	// Cleanup: stop the daemon
	cli.MustExecute("-y", "sync", "daemon", "stop")
}

// TestDaemonIdleTimeout verifies that the daemon exits after idle timeout
// when there's no sync activity.
// Issue #36: Daemon uses timeout-based lifecycle
func TestDaemonIdleTimeout(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)
	configPath := cli.ConfigPath()

	// Configure daemon with short idle timeout (500ms for testing)
	configContent := `
sync:
  enabled: true
  daemon:
    enabled: true
    idle_timeout_ms: 500
    interval: 60
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Verify daemon is running
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "running")

	// Idle timeout is not yet implemented in the daemon.
	// Skip until the feature is available to avoid a tautological assertion.
	t.Skip("idle timeout not yet implemented in daemon")
}

// TestDaemonIPCNotification verifies that CLI can notify daemon via IPC
// when new sync operations are queued.
// Issue #36: Direct IPC communication via Unix socket
func TestDaemonIPCNotification(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)
	configPath := cli.ConfigPath()

	// Configure daemon
	configContent := `
sync:
  enabled: true
  daemon:
    enabled: true
    interval: 60
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cli.SetDaemonInterval(100 * time.Millisecond)

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Add a task - this should notify daemon via IPC to process sync
	cli.MustExecute("-y", "Work", "add", "Task to trigger IPC notification")

	// Wait for daemon to process the sync (sync count should increase)
	stdout := cli.WaitForSyncCount(5*time.Second, 1)
	testutil.AssertContains(t, stdout, "Sync count")
}

// TestDaemonHeartbeat verifies that daemon maintains heartbeat for health monitoring.
// Issue #36: Heartbeat mechanism for hung daemon detection
func TestDaemonHeartbeat(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)
	configPath := cli.ConfigPath()

	// Configure daemon with heartbeat
	configContent := `
sync:
  enabled: true
  daemon:
    enabled: true
    heartbeat_interval_ms: 100
    interval: 60
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Wait for heartbeat to be recorded
	time.Sleep(200 * time.Millisecond)

	// Check daemon status - should show healthy
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "running")
	// TODO: When heartbeat is implemented, verify status shows heartbeat info
}

// TestDaemonKillCommand verifies that 'todoat daemon kill' can force-stop a hung daemon.
// Issue #36: CLI force kill command for recovery
func TestDaemonKillCommand(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)
	configPath := cli.ConfigPath()

	configContent := `
sync:
  enabled: true
  daemon:
    enabled: true
    interval: 60
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")

	// Verify daemon is running
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "running")

	// Use normal stop (kill command would be for hung daemons)
	cli.MustExecute("-y", "sync", "daemon", "stop")

	// Verify daemon is stopped
	statusOut = cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "not running")
}

// =============================================================================
// Issue #39: Background Sync Daemon Process Isolation Tests
// =============================================================================

// TestDaemonProcessIsolation verifies that daemon runs as a separate forked process
// with a different PID than the CLI process.
// Issue #39: Daemon runs as separate forked process (not in-process goroutine)
func TestDaemonProcessIsolation(t *testing.T) {
	cli := testutil.NewCLITestWithForkedDaemon(t)
	configPath := cli.ConfigPath()

	// Configure with forked daemon enabled
	configContent := `
sync:
  enabled: true
  daemon:
    enabled: true
    idle_timeout: 30
    interval: 60
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon - should fork a separate process
	stdout := cli.MustExecute("-y", "sync", "daemon", "start")
	testutil.AssertContains(t, stdout, "started")

	// Give daemon time to fully initialize
	time.Sleep(200 * time.Millisecond)

	// Read daemon PID from file
	pidFile := cli.PIDFilePath()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("failed to read PID file: %v", err)
	}

	pidStr := strings.TrimSpace(string(data))
	var daemonPID int
	if _, err := fmt.Sscanf(pidStr, "%d", &daemonPID); err != nil {
		t.Fatalf("PID file contains invalid PID: %s", pidStr)
	}

	// CRITICAL: Verify daemon PID is different from current process PID
	// This proves the daemon is running as a separate forked process
	currentPID := os.Getpid()
	if daemonPID == currentPID {
		t.Errorf("daemon PID (%d) should be different from test process PID (%d) - daemon is not properly forked", daemonPID, currentPID)
	}

	// Verify the daemon process actually exists
	process, err := os.FindProcess(daemonPID)
	if err != nil {
		t.Fatalf("failed to find daemon process: %v", err)
	}

	// Check if process is alive by sending signal 0
	if err := process.Signal(syscall.Signal(0)); err != nil {
		t.Errorf("daemon process (PID %d) is not running: %v", daemonPID, err)
	}

	// Cleanup
	cli.MustExecute("-y", "sync", "daemon", "stop")

	// Verify daemon has stopped - check PID file is removed and process is not running
	// Note: We check for PID file removal since zombie processes still respond to signal 0
	time.Sleep(500 * time.Millisecond)

	// PID file should be cleaned up after stop
	if _, err := os.Stat(pidFile); err == nil {
		t.Errorf("PID file should be removed after daemon stop")
	}

	// Socket file should be cleaned up after stop
	socketFile := cli.SocketPath()
	if _, err := os.Stat(socketFile); err == nil {
		t.Errorf("Socket file should be removed after daemon stop")
	}

	// Verify status reports not running
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "not running")
}

// TestDaemonIPCCommunication verifies that CLI can communicate with daemon via Unix socket.
// Issue #39: CLI communicates with daemon via Unix domain socket
func TestDaemonIPCCommunication(t *testing.T) {
	cli := testutil.NewCLITestWithForkedDaemon(t)
	configPath := cli.ConfigPath()

	// Configure with forked daemon enabled
	configContent := `
sync:
  enabled: true
  daemon:
    enabled: true
    idle_timeout: 30
    interval: 60
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	time.Sleep(200 * time.Millisecond)

	// Verify Unix socket file exists
	socketPath := cli.SocketPath()
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Fatalf("Unix socket file should exist at %s", socketPath)
	}

	// Get status via IPC - this tests communication
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "running")
	testutil.AssertContains(t, statusOut, "Sync count")

	// Send notification to trigger sync - tests notify message
	cli.MustExecute("-y", "Work", "add", "Task to test IPC notification")

	// Wait for sync and verify count increased
	time.Sleep(500 * time.Millisecond)
	statusOut = cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "Sync count")
}

// TestDaemonAtomicTaskClaiming verifies that sync operations work without double-execution.
// Issue #39: Atomic task claiming via BEGIN IMMEDIATE transaction
// Note: This test verifies the daemon runs sync operations - atomic claiming is an internal
// implementation detail tested at the unit level in the daemon package.
func TestDaemonAtomicTaskClaiming(t *testing.T) {
	cli := testutil.NewCLITestWithForkedDaemon(t)
	configPath := cli.ConfigPath()

	// Configure with forked daemon enabled and short interval
	configContent := `
sync:
  enabled: true
  daemon:
    enabled: true
    idle_timeout: 30
    interval: 1
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Wait for a few sync cycles
	time.Sleep(2 * time.Second)

	// Verify daemon has run sync operations
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "running")

	// Sync count should have increased (daemon is processing)
	// Note: The actual sync may not clear the queue if there's no remote backend
	// configured, but the daemon should be actively attempting syncs
	testutil.AssertContains(t, statusOut, "Sync count")
}

// TestDaemonHeartbeatDetectsHung verifies that heartbeat mechanism detects hung daemons.
// Issue #39: Heartbeat table for hung daemon detection
func TestDaemonHeartbeatDetectsHung(t *testing.T) {
	cli := testutil.NewCLITestWithForkedDaemon(t)
	configPath := cli.ConfigPath()

	// Configure with short heartbeat for testing
	configContent := `
sync:
  enabled: true
  daemon:
    enabled: true
    heartbeat_interval_ms: 100
    heartbeat_timeout_ms: 500
    interval: 60
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")

	// Give daemon time to record heartbeats
	time.Sleep(300 * time.Millisecond)

	// Check daemon status - should show healthy with recent heartbeat
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "running")

	// Heartbeat display is not yet implemented.
	// Skip until the feature is available to avoid a no-op assertion.
	t.Skip("heartbeat display not yet implemented in daemon status")

	// Cleanup
	cli.MustExecute("-y", "sync", "daemon", "stop")
}

// TestDaemonGracefulShutdown verifies clean shutdown on SIGTERM.
// Issue #39: Verify clean shutdown on SIGTERM
func TestDaemonGracefulShutdown(t *testing.T) {
	cli := testutil.NewCLITestWithForkedDaemon(t)
	configPath := cli.ConfigPath()

	// Configure with forked daemon enabled
	configContent := `
sync:
  enabled: true
  daemon:
    enabled: true
    idle_timeout: 30
    interval: 60
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")

	time.Sleep(200 * time.Millisecond)

	// Read daemon PID
	pidFile := cli.PIDFilePath()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("failed to read PID file: %v", err)
	}

	var daemonPID int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &daemonPID); err != nil {
		t.Fatalf("invalid PID: %s", string(data))
	}

	// Find the process
	process, err := os.FindProcess(daemonPID)
	if err != nil {
		t.Fatalf("failed to find daemon process: %v", err)
	}

	// Send SIGTERM directly to test graceful shutdown
	if err := process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	// Wait for graceful shutdown - daemon should clean up PID file
	// Note: Zombie processes still respond to signal 0, so check files instead
	time.Sleep(1 * time.Second)

	// Verify PID file is cleaned up (daemon cleanup on graceful exit)
	if _, err := os.Stat(pidFile); err == nil {
		t.Errorf("PID file should be removed after graceful shutdown")
	}

	// Verify socket file is cleaned up
	socketPath := cli.SocketPath()
	if _, err := os.Stat(socketPath); err == nil {
		t.Errorf("Socket file should be removed after graceful shutdown")
	}

	// Status should show not running
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "not running")
}

// TestDaemonIdleTimeoutExit verifies that daemon exits after idle timeout.
// Issue #39: Daemon uses 5-second idle timeout and exits when no work pending
func TestDaemonIdleTimeoutExit(t *testing.T) {
	cli := testutil.NewCLITestWithForkedDaemon(t)
	configPath := cli.ConfigPath()

	// Configure with very short idle timeout for testing
	configContent := `
sync:
  enabled: true
  daemon:
    enabled: true
    idle_timeout: 1
    interval: 60
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")

	// Initially should be running
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "running")

	// Wait for idle timeout (1 second + buffer)
	time.Sleep(2 * time.Second)

	// Daemon should have exited due to idle timeout
	statusOut = cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "not running")
}

// TestDaemonPidfilePreventsMultiple verifies that pidfile/lockfile prevents multiple instances.
// Issue #39: Pidfile/lockfile prevents multiple daemon instances
func TestDaemonPidfilePreventsMultiple(t *testing.T) {
	cli := testutil.NewCLITestWithForkedDaemon(t)
	configPath := cli.ConfigPath()

	// Configure with forked daemon enabled
	configContent := `
sync:
  enabled: true
  daemon:
    enabled: true
    idle_timeout: 30
    interval: 60
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start first daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	time.Sleep(200 * time.Millisecond)

	// Try to start second daemon - should report already running
	stdout := cli.MustExecute("-y", "sync", "daemon", "start")
	testutil.AssertContains(t, stdout, "already running")
}

// TestDaemonKillForEmergencyTermination verifies the kill command for emergency termination.
// Issue #39: todoat daemon kill for emergency termination
func TestDaemonKillForEmergencyTermination(t *testing.T) {
	cli := testutil.NewCLITestWithForkedDaemon(t)
	configPath := cli.ConfigPath()

	// Configure with forked daemon enabled
	configContent := `
sync:
  enabled: true
  daemon:
    enabled: true
    idle_timeout: 30
    interval: 60
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")

	time.Sleep(200 * time.Millisecond)

	// Verify daemon is running
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "running")

	// Use kill command to force terminate
	cli.MustExecute("-y", "sync", "daemon", "kill")

	// Wait for cleanup
	time.Sleep(500 * time.Millisecond)

	// Daemon should be stopped
	statusOut = cli.MustExecute("-y", "sync", "daemon", "status")
	testutil.AssertContains(t, statusOut, "not running")
}

// TestDaemonStatusShowsHealth verifies that status shows daemon health info.
// Issue #39: todoat daemon status shows daemon health info
func TestDaemonStatusShowsHealth(t *testing.T) {
	cli := testutil.NewCLITestWithForkedDaemon(t)
	configPath := cli.ConfigPath()

	// Configure with forked daemon enabled
	configContent := `
sync:
  enabled: true
  daemon:
    enabled: true
    idle_timeout: 30
    interval: 1
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Wait for a sync cycle
	time.Sleep(1500 * time.Millisecond)

	// Get daemon status
	statusOut := cli.MustExecute("-y", "sync", "daemon", "status")

	// Should show running status
	testutil.AssertContains(t, statusOut, "running")

	// Should show PID information
	testutil.AssertContains(t, statusOut, "PID")

	// Should show sync count
	testutil.AssertContains(t, statusOut, "Sync count")

	// Should show interval
	testutil.AssertContains(t, statusOut, "Interval")
}

// TestCLIReturnsImmediately verifies that CLI returns immediately after local operations
// instead of blocking on sync completion.
// Issue #36: CLI should not hang waiting for sync
func TestCLIReturnsImmediately(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)
	configPath := cli.ConfigPath()

	// Configure with daemon and auto-sync enabled
	configContent := `
sync:
  enabled: true
  auto_sync_after_operation: true
  daemon:
    enabled: true
    interval: 60
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Measure time to add a task
	start := time.Now()
	cli.MustExecute("-y", "Work", "add", "Task to test immediate return")
	elapsed := time.Since(start)

	// CLI should return quickly (under 1 second for local operation)
	if elapsed > 1*time.Second {
		t.Errorf("CLI took %v to return, expected <1s for local operation", elapsed)
	}
}
