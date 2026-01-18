package notification_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"todoat/internal/notification"
	"todoat/internal/testutil"
)

// =============================================================================
// CLI Tests (022-notification-system)
// =============================================================================

// TestNotificationTest tests that 'todoat notification test' sends a test notification and exits 0
func TestNotificationTest(t *testing.T) {
	cli := testutil.NewCLITestWithNotification(t)

	stdout := cli.MustExecute("-y", "notification", "test")

	testutil.AssertContains(t, stdout, "Test notification sent")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// TestNotificationLog tests that 'todoat notification log' displays notification history
func TestNotificationLog(t *testing.T) {
	cli := testutil.NewCLITestWithNotification(t)

	// First send a test notification to create log entry
	cli.MustExecute("-y", "notification", "test")

	// View the log
	stdout := cli.MustExecute("-y", "notification", "log")

	// Should contain the test notification
	testutil.AssertContains(t, stdout, "TEST")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestNotificationLogClear tests that 'todoat notification log clear' clears the log file
func TestNotificationLogClear(t *testing.T) {
	cli := testutil.NewCLITestWithNotification(t)

	// First send a test notification to create log entry
	cli.MustExecute("-y", "notification", "test")

	// Clear the log
	stdout := cli.MustExecute("-y", "notification", "log", "clear")

	testutil.AssertContains(t, stdout, "cleared")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify log is empty
	stdout = cli.MustExecute("-y", "notification", "log")
	testutil.AssertContains(t, stdout, "No notifications")
}

// =============================================================================
// Unit Tests - OS Notification (with mock command executor)
// =============================================================================

// TestOSNotificationLinux tests that OS notification is sent via notify-send on Linux
func TestOSNotificationLinux(t *testing.T) {
	var executedCmd string
	var executedArgs []string

	mock := &notification.MockCommandExecutor{
		ExecuteFunc: func(cmd string, args ...string) error {
			executedCmd = cmd
			executedArgs = args
			return nil
		},
	}

	channel := notification.NewOSNotificationChannel(
		&notification.OSNotificationConfig{
			Enabled:        true,
			OnSyncComplete: true,
			OnSyncError:    true,
			OnConflict:     true,
		},
		notification.WithCommandExecutor(mock),
		notification.WithPlatform("linux"),
	)

	n := notification.Notification{
		Type:      notification.NotifySyncComplete,
		Title:     "Sync Complete",
		Message:   "Synced 5 tasks",
		Timestamp: time.Now(),
	}

	err := channel.Send(n)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if executedCmd != "notify-send" {
		t.Errorf("expected notify-send command, got %q", executedCmd)
	}

	// Check args contain title and message
	argsStr := strings.Join(executedArgs, " ")
	if !strings.Contains(argsStr, "Sync Complete") {
		t.Errorf("expected args to contain title, got %v", executedArgs)
	}
	if !strings.Contains(argsStr, "Synced 5 tasks") {
		t.Errorf("expected args to contain message, got %v", executedArgs)
	}
}

// TestOSNotificationDarwin tests that OS notification is sent via osascript on macOS
func TestOSNotificationDarwin(t *testing.T) {
	var executedCmd string
	var executedArgs []string

	mock := &notification.MockCommandExecutor{
		ExecuteFunc: func(cmd string, args ...string) error {
			executedCmd = cmd
			executedArgs = args
			return nil
		},
	}

	channel := notification.NewOSNotificationChannel(
		&notification.OSNotificationConfig{
			Enabled:        true,
			OnSyncComplete: true,
			OnSyncError:    true,
			OnConflict:     true,
		},
		notification.WithCommandExecutor(mock),
		notification.WithPlatform("darwin"),
	)

	n := notification.Notification{
		Type:      notification.NotifySyncError,
		Title:     "Sync Error",
		Message:   "Connection failed",
		Timestamp: time.Now(),
	}

	err := channel.Send(n)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if executedCmd != "osascript" {
		t.Errorf("expected osascript command, got %q", executedCmd)
	}

	// Check args contain -e flag with display notification
	argsStr := strings.Join(executedArgs, " ")
	if !strings.Contains(argsStr, "-e") {
		t.Errorf("expected args to contain -e flag, got %v", executedArgs)
	}
	if !strings.Contains(argsStr, "display notification") {
		t.Errorf("expected args to contain 'display notification', got %v", executedArgs)
	}
}

// =============================================================================
// Unit Tests - Log Notification
// =============================================================================

// TestLogNotification tests that notifications are written to log file with correct format
func TestLogNotification(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "notifications.log")

	channel := notification.NewLogNotificationChannel(&notification.LogNotificationConfig{
		Enabled:       true,
		Path:          logPath,
		MaxSizeMB:     10,
		RetentionDays: 30,
	})
	defer func() { _ = channel.Close() }()

	n := notification.Notification{
		Type:      notification.NotifySyncComplete,
		Title:     "Sync Complete",
		Message:   "Synced 5 tasks with nextcloud",
		Timestamp: time.Date(2026, 1, 16, 10, 30, 0, 0, time.UTC),
	}

	err := channel.Send(n)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Read the log file
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	content := string(data)

	// Check format: 2026-01-16T10:30:00Z [SYNC_COMPLETE] Message
	if !strings.Contains(content, "2026-01-16T10:30:00Z") {
		t.Errorf("expected log to contain timestamp, got:\n%s", content)
	}
	if !strings.Contains(content, "[SYNC_COMPLETE]") {
		t.Errorf("expected log to contain [SYNC_COMPLETE], got:\n%s", content)
	}
	if !strings.Contains(content, "Synced 5 tasks with nextcloud") {
		t.Errorf("expected log to contain message, got:\n%s", content)
	}
}

// =============================================================================
// Unit Tests - Configuration
// =============================================================================

// TestNotificationConfig tests that configuration enables/disables notification channels
func TestNotificationConfig(t *testing.T) {
	tests := []struct {
		name             string
		osEnabled        bool
		logEnabled       bool
		expectedChannels int
	}{
		{"both enabled", true, true, 2},
		{"only os enabled", true, false, 1},
		{"only log enabled", false, true, 1},
		{"both disabled", false, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "notifications.log")

			cfg := &notification.Config{
				Enabled: true,
				OSNotification: notification.OSNotificationConfig{
					Enabled:        tt.osEnabled,
					OnSyncComplete: true,
					OnSyncError:    true,
					OnConflict:     true,
				},
				LogNotification: notification.LogNotificationConfig{
					Enabled:       tt.logEnabled,
					Path:          logPath,
					MaxSizeMB:     10,
					RetentionDays: 30,
				},
			}

			// Use mock executor to avoid actual OS notifications
			mock := &notification.MockCommandExecutor{
				ExecuteFunc: func(cmd string, args ...string) error {
					return nil
				},
			}

			manager, err := notification.NewManager(cfg, notification.WithCommandExecutor(mock))
			if err != nil {
				t.Fatalf("failed to create manager: %v", err)
			}
			defer func() { _ = manager.Close() }()

			// Test that the manager was created successfully
			if manager == nil {
				t.Fatal("expected manager to be non-nil")
			}

			// Verify channel count (internal implementation detail - may need adjustment)
			channelCount := manager.ChannelCount()
			if channelCount != tt.expectedChannels {
				t.Errorf("expected %d channels, got %d", tt.expectedChannels, channelCount)
			}
		})
	}
}

// TestNotificationDisabled tests that when notification.enabled is false, no notifications are sent
func TestNotificationDisabled(t *testing.T) {
	cfg := &notification.Config{
		Enabled: false,
	}

	manager, err := notification.NewManager(cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer func() { _ = manager.Close() }()

	n := notification.Notification{
		Type:      notification.NotifySyncComplete,
		Title:     "Test",
		Message:   "Test message",
		Timestamp: time.Now(),
	}

	// Should not error when notifications are disabled
	err = manager.Send(n)
	if err != nil {
		t.Errorf("expected no error for disabled notifications, got %v", err)
	}
}

// TestNotificationTypeFiltering tests that notification types are filtered based on config
func TestNotificationTypeFiltering(t *testing.T) {
	var sentNotifications []notification.Notification

	mock := &notification.MockCommandExecutor{
		ExecuteFunc: func(cmd string, args ...string) error {
			return nil
		},
	}

	// Create channel with OnSyncComplete disabled
	channel := notification.NewOSNotificationChannel(
		&notification.OSNotificationConfig{
			Enabled:        true,
			OnSyncComplete: false, // Disabled
			OnSyncError:    true,
			OnConflict:     true,
		},
		notification.WithCommandExecutor(mock),
		notification.WithPlatform("linux"),
		notification.WithSendCallback(func(n notification.Notification) {
			sentNotifications = append(sentNotifications, n)
		}),
	)

	// Try to send sync_complete - should be filtered
	err := channel.Send(notification.Notification{
		Type:      notification.NotifySyncComplete,
		Title:     "Complete",
		Message:   "Test",
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Try to send sync_error - should pass through
	err = channel.Send(notification.Notification{
		Type:      notification.NotifySyncError,
		Title:     "Error",
		Message:   "Test",
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Only sync_error should have been sent
	if len(sentNotifications) != 1 {
		t.Errorf("expected 1 notification sent, got %d", len(sentNotifications))
	}
	if len(sentNotifications) > 0 && sentNotifications[0].Type != notification.NotifySyncError {
		t.Errorf("expected sync_error to be sent, got %s", sentNotifications[0].Type)
	}
}
