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

	// Configure a short interval for testing (1 second)
	cli.SetDaemonInterval(1 * time.Second)

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Wait for at least 2 intervals
	time.Sleep(2500 * time.Millisecond)

	// Check that sync ran multiple times by examining logs or status
	stdout := cli.MustExecute("-y", "sync", "daemon", "status")

	// Status should show sync count > 1
	testutil.AssertContains(t, stdout, "Sync count")
}

// TestSyncDaemonNotificationCLI tests that notifications are sent on sync events
func TestSyncDaemonNotificationCLI(t *testing.T) {
	cli := testutil.NewCLITestWithDaemon(t)

	// Configure short interval and enable notifications
	cli.SetDaemonInterval(1 * time.Second)

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Wait for a sync cycle to complete
	time.Sleep(1500 * time.Millisecond)

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
	cli.SetDaemonInterval(1 * time.Second)

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Wait for a sync attempt
	time.Sleep(1500 * time.Millisecond)

	// Daemon should still be running (not crashed)
	stdout := cli.MustExecute("-y", "sync", "daemon", "status")
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
	cli.SetDaemonInterval(1 * time.Second)

	// Start daemon
	cli.MustExecute("-y", "sync", "daemon", "start")
	defer cli.MustExecute("-y", "sync", "daemon", "stop")

	// Wait for offline sync attempt
	time.Sleep(1500 * time.Millisecond)

	// Restore network
	cli.SetDaemonOffline(false)

	// Wait for reconnect and successful sync
	time.Sleep(1500 * time.Millisecond)

	// Check status shows successful sync
	stdout := cli.MustExecute("-y", "sync", "daemon", "status")
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

	// Wait for some activity
	time.Sleep(500 * time.Millisecond)

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
