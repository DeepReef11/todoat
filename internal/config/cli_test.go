package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// TestConfigSetPreservesComments verifies that 'config set' preserves YAML comments in the config file (issue #57)
func TestConfigSetPreservesComments(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	configWithComments := `backends:
  # SQLite backend - local database storage (recommended default)
  sqlite:
    type: sqlite
    enabled: true
    # path: "~/.local/share/todoat/tasks.db"  # Optional: custom database path

  # Nextcloud backend - sync with Nextcloud Tasks via CalDAV
  # nextcloud:
  #   type: nextcloud
  #   enabled: false

# Default backend when multiple are enabled
default_backend: sqlite

# Disable interactive prompts (for scripting)
no_prompt: false

sync:
  enabled: false
  # offline_mode: auto  # Options: auto | online | offline
`
	cli.SetFullConfig(configWithComments)

	// Run config set
	stdout := cli.MustExecute("-y", "config", "set", "default_backend", "sqlite")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Read the config file after set
	content, err := os.ReadFile(cli.ConfigPath())
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	result := string(content)

	// Comments must be preserved
	if !strings.Contains(result, "# SQLite backend - local database storage (recommended default)") {
		t.Errorf("config set destroyed SQLite backend comment.\nFile contents after set:\n%s", result)
	}
	if !strings.Contains(result, "# Nextcloud backend - sync with Nextcloud Tasks via CalDAV") {
		t.Errorf("config set destroyed Nextcloud backend comment.\nFile contents after set:\n%s", result)
	}
	if !strings.Contains(result, "# Default backend when multiple are enabled") {
		t.Errorf("config set destroyed default_backend comment.\nFile contents after set:\n%s", result)
	}
	if !strings.Contains(result, "# Disable interactive prompts (for scripting)") {
		t.Errorf("config set destroyed no_prompt comment.\nFile contents after set:\n%s", result)
	}
	if !strings.Contains(result, "# offline_mode: auto") {
		t.Errorf("config set destroyed commented-out offline_mode sample.\nFile contents after set:\n%s", result)
	}
}

// =============================================================================
// Config CLI Command Tests (049-config-cli-commands)
// =============================================================================

// --- Config Get Tests ---

// TestConfigGetCLI verifies 'todoat config get default_backend' returns current value
func TestConfigGetCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set a known config value
	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
no_prompt: true
`)

	stdout := cli.MustExecute("-y", "config", "get", "default_backend")

	testutil.AssertContains(t, stdout, "sqlite")
}

// TestConfigGetNestedCLI verifies 'todoat config get sync.enabled' returns nested value
func TestConfigGetNestedCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: true
  offline_mode: auto
`)

	stdout := cli.MustExecute("-y", "config", "get", "sync.enabled")

	testutil.AssertContains(t, stdout, "true")
}

// TestConfigGetAllCLI verifies 'todoat config get' returns all config as YAML
func TestConfigGetAllCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
no_prompt: false
output_format: text
`)

	stdout := cli.MustExecute("-y", "config", "get")

	// Should contain YAML structure indicators
	testutil.AssertContains(t, stdout, "default_backend")
	testutil.AssertContains(t, stdout, "sqlite")
	testutil.AssertContains(t, stdout, "backends")
}

// --- Config Set Tests ---

// TestConfigSetCLI verifies 'todoat config set no_prompt true' updates config file
func TestConfigSetCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
no_prompt: false
`)

	stdout := cli.MustExecute("-y", "config", "set", "no_prompt", "true")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the value was changed
	stdout = cli.MustExecute("-y", "config", "get", "no_prompt")
	testutil.AssertContains(t, stdout, "true")
}

// TestConfigSetNestedCLI verifies 'todoat config set sync.offline_mode auto' updates nested value
func TestConfigSetNestedCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: false
  offline_mode: online
`)

	stdout := cli.MustExecute("-y", "config", "set", "sync.offline_mode", "auto")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the value was changed
	stdout = cli.MustExecute("-y", "config", "get", "sync.offline_mode")
	testutil.AssertContains(t, stdout, "auto")
}

// TestConfigSetValidationCLI verifies 'todoat config set no_prompt invalid' returns ERROR with valid values
func TestConfigSetValidationCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
no_prompt: false
`)

	stdout, stderr := cli.ExecuteAndFail("-y", "config", "set", "no_prompt", "invalid")

	// Should show error with valid values
	combined := stdout + stderr
	if !strings.Contains(combined, "true") || !strings.Contains(combined, "false") {
		t.Errorf("expected error message to mention valid values (true/false), got: %s", combined)
	}
}

// --- Config Path Test ---

// TestConfigPathCLI verifies 'todoat config path' shows config file location
func TestConfigPathCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	stdout := cli.MustExecute("-y", "config", "path")

	// Should contain path to config file
	testutil.AssertContains(t, stdout, "config.yaml")
}

// --- Config Edit Test ---

// TestConfigEditCLI verifies 'todoat config edit' opens config in $EDITOR
func TestConfigEditCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set EDITOR to a simple command that just exits
	tmpDir := cli.TmpDir()
	touchFile := filepath.Join(tmpDir, "editor_ran")

	// Create a fake editor script that touches a file to prove it ran
	editorScript := filepath.Join(tmpDir, "fake_editor.sh")
	scriptContent := "#!/bin/sh\ntouch " + touchFile + "\n"
	if err := os.WriteFile(editorScript, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create editor script: %v", err)
	}

	// Set EDITOR environment variable
	t.Setenv("EDITOR", editorScript)

	stdout := cli.MustExecute("-y", "config", "edit")

	// Check that the fake editor was invoked
	if _, err := os.Stat(touchFile); os.IsNotExist(err) {
		t.Errorf("expected editor to be invoked, but marker file not created")
	}

	_ = stdout // May contain additional output
}

// --- Config Reset Test ---

// TestConfigResetCLI verifies 'todoat config reset' restores default config (with confirmation)
func TestConfigResetCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set non-default values
	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
no_prompt: true
output_format: json
`)

	// Verify non-default value before reset
	stdout := cli.MustExecute("-y", "config", "get", "output_format")
	testutil.AssertContains(t, stdout, "json")

	// Reset config (with -y flag to skip confirmation)
	stdout = cli.MustExecute("-y", "config", "reset")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify defaults were restored
	stdout = cli.MustExecute("-y", "config", "get", "output_format")
	testutil.AssertContains(t, stdout, "text")
}

// --- Config Set Analytics Tests (078-fix-config-set-analytics-key) ---

// TestConfigSetAnalyticsEnabledCLI verifies 'todoat config set analytics.enabled true' works
func TestConfigSetAnalyticsEnabledCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
analytics:
  enabled: false
  retention_days: 30
`)

	stdout := cli.MustExecute("-y", "config", "set", "analytics.enabled", "true")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the value was changed
	stdout = cli.MustExecute("-y", "config", "get", "analytics.enabled")
	testutil.AssertContains(t, stdout, "true")
}

// TestConfigSetAnalyticsRetentionDaysCLI verifies 'todoat config set analytics.retention_days 365' works
func TestConfigSetAnalyticsRetentionDaysCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
analytics:
  enabled: true
  retention_days: 30
`)

	stdout := cli.MustExecute("-y", "config", "set", "analytics.retention_days", "365")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the value was changed
	stdout = cli.MustExecute("-y", "config", "get", "analytics.retention_days")
	testutil.AssertContains(t, stdout, "365")
}

// TestConfigSetAnalyticsValidationCLI verifies invalid values are rejected
func TestConfigSetAnalyticsValidationCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
analytics:
  enabled: false
  retention_days: 30
`)

	// Test invalid boolean for analytics.enabled
	stdout, stderr := cli.ExecuteAndFail("-y", "config", "set", "analytics.enabled", "invalid")
	combined := stdout + stderr
	if !strings.Contains(combined, "true") || !strings.Contains(combined, "false") {
		t.Errorf("expected error message to mention valid boolean values, got: %s", combined)
	}

	// Test negative value for analytics.retention_days
	// Note: Use "--" to separate flags from arguments so "-1" is not interpreted as a flag
	stdout, stderr = cli.ExecuteAndFail("-y", "config", "set", "analytics.retention_days", "--", "-1")
	combined = stdout + stderr
	if !strings.Contains(combined, "non-negative") && !strings.Contains(combined, "invalid") {
		t.Errorf("expected error message about invalid value for retention_days, got: %s", combined)
	}

	// Test non-integer value for analytics.retention_days
	stdout, stderr = cli.ExecuteAndFail("-y", "config", "set", "analytics.retention_days", "abc")
	combined = stdout + stderr
	if !strings.Contains(combined, "invalid") && !strings.Contains(combined, "integer") {
		t.Errorf("expected error message about invalid value for retention_days, got: %s", combined)
	}
}

// --- Config Set Sync Auto-Sync Tests (002-config-set-sync-auto-sync-key-not-recognized) ---

// TestConfigSetSyncAutoSyncAfterOperationCLI verifies 'todoat config set sync.auto_sync_after_operation true' works
func TestConfigSetSyncAutoSyncAfterOperationCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: false
  auto_sync_after_operation: false
`)

	stdout := cli.MustExecute("-y", "config", "set", "sync.auto_sync_after_operation", "true")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the value was changed
	stdout = cli.MustExecute("-y", "config", "get", "sync.auto_sync_after_operation")
	testutil.AssertContains(t, stdout, "true")
}

// TestConfigSetSyncAutoSyncAfterOperationValidationCLI verifies invalid values are rejected
func TestConfigSetSyncAutoSyncAfterOperationValidationCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: false
  auto_sync_after_operation: false
`)

	// Test invalid boolean for sync.auto_sync_after_operation
	stdout, stderr := cli.ExecuteAndFail("-y", "config", "set", "sync.auto_sync_after_operation", "invalid")
	combined := stdout + stderr
	if !strings.Contains(combined, "true") || !strings.Contains(combined, "false") {
		t.Errorf("expected error message to mention valid boolean values, got: %s", combined)
	}
}

// --- Config Set Sync Conflict Resolution Tests (080-fix-conflict-resolution-values-mismatch) ---

// TestConfigSetSyncConflictResolutionServerWinsCLI verifies 'todoat config set sync.conflict_resolution server_wins' works
func TestConfigSetSyncConflictResolutionServerWinsCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: false
  conflict_resolution: local_wins
`)

	stdout := cli.MustExecute("-y", "config", "set", "sync.conflict_resolution", "server_wins")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the value was changed
	stdout = cli.MustExecute("-y", "config", "get", "sync.conflict_resolution")
	testutil.AssertContains(t, stdout, "server_wins")
}

// TestConfigSetSyncConflictResolutionLocalWinsCLI verifies 'todoat config set sync.conflict_resolution local_wins' works
func TestConfigSetSyncConflictResolutionLocalWinsCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: false
  conflict_resolution: server_wins
`)

	stdout := cli.MustExecute("-y", "config", "set", "sync.conflict_resolution", "local_wins")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the value was changed
	stdout = cli.MustExecute("-y", "config", "get", "sync.conflict_resolution")
	testutil.AssertContains(t, stdout, "local_wins")
}

// TestConfigSetSyncConflictResolutionMergeCLI verifies 'todoat config set sync.conflict_resolution merge' works
func TestConfigSetSyncConflictResolutionMergeCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: false
  conflict_resolution: server_wins
`)

	stdout := cli.MustExecute("-y", "config", "set", "sync.conflict_resolution", "merge")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the value was changed
	stdout = cli.MustExecute("-y", "config", "get", "sync.conflict_resolution")
	testutil.AssertContains(t, stdout, "merge")
}

// TestConfigSetSyncConflictResolutionKeepBothCLI verifies 'todoat config set sync.conflict_resolution keep_both' works
func TestConfigSetSyncConflictResolutionKeepBothCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: false
  conflict_resolution: server_wins
`)

	stdout := cli.MustExecute("-y", "config", "set", "sync.conflict_resolution", "keep_both")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the value was changed
	stdout = cli.MustExecute("-y", "config", "get", "sync.conflict_resolution")
	testutil.AssertContains(t, stdout, "keep_both")
}

// TestConfigSetSyncConflictResolutionValidationCLI verifies invalid values are rejected with correct error message
func TestConfigSetSyncConflictResolutionValidationCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: false
  conflict_resolution: server_wins
`)

	// Test invalid value for sync.conflict_resolution
	stdout, stderr := cli.ExecuteAndFail("-y", "config", "set", "sync.conflict_resolution", "invalid")
	combined := stdout + stderr

	// Error message should mention all valid values
	if !strings.Contains(combined, "server_wins") {
		t.Errorf("expected error message to mention 'server_wins', got: %s", combined)
	}
	if !strings.Contains(combined, "local_wins") {
		t.Errorf("expected error message to mention 'local_wins', got: %s", combined)
	}
	if !strings.Contains(combined, "merge") {
		t.Errorf("expected error message to mention 'merge', got: %s", combined)
	}
	if !strings.Contains(combined, "keep_both") {
		t.Errorf("expected error message to mention 'keep_both', got: %s", combined)
	}
}

// --- Config Get/Set Reminder and Daemon Keys Tests (Issue #61) ---

// TestConfigGetReminderEnabledCLI verifies 'todoat config get reminder.enabled' works
func TestConfigGetReminderEnabledCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
reminder:
  enabled: true
  os_notification: false
  log_notification: true
  intervals:
    - "1h"
    - "24h"
`)

	stdout := cli.MustExecute("-y", "config", "get", "reminder.enabled")
	testutil.AssertContains(t, stdout, "true")
}

// TestConfigSetReminderEnabledCLI verifies 'todoat config set reminder.enabled true' works
func TestConfigSetReminderEnabledCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
reminder:
  enabled: false
`)

	stdout := cli.MustExecute("-y", "config", "set", "reminder.enabled", "true")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	stdout = cli.MustExecute("-y", "config", "get", "reminder.enabled")
	testutil.AssertContains(t, stdout, "true")
}

// TestConfigGetReminderOSNotificationCLI verifies 'todoat config get reminder.os_notification' works
func TestConfigGetReminderOSNotificationCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
reminder:
  enabled: true
  os_notification: true
`)

	stdout := cli.MustExecute("-y", "config", "get", "reminder.os_notification")
	testutil.AssertContains(t, stdout, "true")
}

// TestConfigSetReminderOSNotificationCLI verifies 'todoat config set reminder.os_notification true' works
func TestConfigSetReminderOSNotificationCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
reminder:
  enabled: true
  os_notification: false
`)

	stdout := cli.MustExecute("-y", "config", "set", "reminder.os_notification", "true")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	stdout = cli.MustExecute("-y", "config", "get", "reminder.os_notification")
	testutil.AssertContains(t, stdout, "true")
}

// TestConfigGetReminderLogNotificationCLI verifies 'todoat config get reminder.log_notification' works
func TestConfigGetReminderLogNotificationCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
reminder:
  enabled: true
  log_notification: true
`)

	stdout := cli.MustExecute("-y", "config", "get", "reminder.log_notification")
	testutil.AssertContains(t, stdout, "true")
}

// TestConfigSetReminderLogNotificationCLI verifies 'todoat config set reminder.log_notification false' works
func TestConfigSetReminderLogNotificationCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
reminder:
  enabled: true
  log_notification: true
`)

	stdout := cli.MustExecute("-y", "config", "set", "reminder.log_notification", "false")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	stdout = cli.MustExecute("-y", "config", "get", "reminder.log_notification")
	testutil.AssertContains(t, stdout, "false")
}

// TestConfigGetSyncDaemonEnabledCLI verifies 'todoat config get sync.daemon.enabled' works
func TestConfigGetSyncDaemonEnabledCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: true
  daemon:
    enabled: true
    interval: 300
    idle_timeout: 1800
`)

	stdout := cli.MustExecute("-y", "config", "get", "sync.daemon.enabled")
	testutil.AssertContains(t, stdout, "true")
}

// TestConfigSetSyncDaemonEnabledCLI verifies 'todoat config set sync.daemon.enabled true' works
func TestConfigSetSyncDaemonEnabledCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: true
  daemon:
    enabled: false
`)

	stdout := cli.MustExecute("-y", "config", "set", "sync.daemon.enabled", "true")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	stdout = cli.MustExecute("-y", "config", "get", "sync.daemon.enabled")
	testutil.AssertContains(t, stdout, "true")
}

// TestConfigGetSyncDaemonIntervalCLI verifies 'todoat config get sync.daemon.interval' works
func TestConfigGetSyncDaemonIntervalCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: true
  daemon:
    enabled: true
    interval: 300
`)

	stdout := cli.MustExecute("-y", "config", "get", "sync.daemon.interval")
	testutil.AssertContains(t, stdout, "300")
}

// TestConfigSetSyncDaemonIntervalCLI verifies 'todoat config set sync.daemon.interval 60' works
func TestConfigSetSyncDaemonIntervalCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: true
  daemon:
    enabled: true
    interval: 300
`)

	stdout := cli.MustExecute("-y", "config", "set", "sync.daemon.interval", "60")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	stdout = cli.MustExecute("-y", "config", "get", "sync.daemon.interval")
	testutil.AssertContains(t, stdout, "60")
}

// TestConfigGetSyncDaemonIdleTimeoutCLI verifies 'todoat config get sync.daemon.idle_timeout' works
func TestConfigGetSyncDaemonIdleTimeoutCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: true
  daemon:
    enabled: true
    idle_timeout: 1800
`)

	stdout := cli.MustExecute("-y", "config", "get", "sync.daemon.idle_timeout")
	testutil.AssertContains(t, stdout, "1800")
}

// TestConfigSetSyncDaemonIdleTimeoutCLI verifies 'todoat config set sync.daemon.idle_timeout 3600' works
func TestConfigSetSyncDaemonIdleTimeoutCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: true
  daemon:
    enabled: true
    idle_timeout: 1800
`)

	stdout := cli.MustExecute("-y", "config", "set", "sync.daemon.idle_timeout", "3600")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	stdout = cli.MustExecute("-y", "config", "get", "sync.daemon.idle_timeout")
	testutil.AssertContains(t, stdout, "3600")
}

// TestConfigGetSyncDaemonStuckTimeoutCLI verifies 'todoat config get sync.daemon.stuck_timeout' works
func TestConfigGetSyncDaemonStuckTimeoutCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: true
  daemon:
    enabled: true
    stuck_timeout: 15
`)

	stdout := cli.MustExecute("-y", "config", "get", "sync.daemon.stuck_timeout")
	testutil.AssertContains(t, stdout, "15")
}

// TestConfigSetSyncDaemonStuckTimeoutCLI verifies 'todoat config set sync.daemon.stuck_timeout 20' works
func TestConfigSetSyncDaemonStuckTimeoutCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: true
  daemon:
    enabled: true
    stuck_timeout: 10
`)

	stdout := cli.MustExecute("-y", "config", "set", "sync.daemon.stuck_timeout", "20")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	stdout = cli.MustExecute("-y", "config", "get", "sync.daemon.stuck_timeout")
	testutil.AssertContains(t, stdout, "20")
}

// TestConfigGetSyncDaemonTaskTimeoutCLI verifies 'todoat config get sync.daemon.task_timeout' works
func TestConfigGetSyncDaemonTaskTimeoutCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: true
  daemon:
    enabled: true
    task_timeout: "10m"
`)

	stdout := cli.MustExecute("-y", "config", "get", "sync.daemon.task_timeout")
	testutil.AssertContains(t, stdout, "10m")
}

// TestConfigSetSyncDaemonTaskTimeoutCLI verifies 'todoat config set sync.daemon.task_timeout 15m' works
func TestConfigSetSyncDaemonTaskTimeoutCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: true
  daemon:
    enabled: true
    task_timeout: "5m"
`)

	stdout := cli.MustExecute("-y", "config", "set", "sync.daemon.task_timeout", "15m")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	stdout = cli.MustExecute("-y", "config", "get", "sync.daemon.task_timeout")
	testutil.AssertContains(t, stdout, "15m")
}

// TestConfigGetSyncBackgroundPullCooldownCLI verifies 'todoat config get sync.background_pull_cooldown' works
func TestConfigGetSyncBackgroundPullCooldownCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: true
  background_pull_cooldown: "60s"
`)

	stdout := cli.MustExecute("-y", "config", "get", "sync.background_pull_cooldown")
	testutil.AssertContains(t, stdout, "60s")
}

// TestConfigGetAllIncludesReminderAndDaemonCLI verifies 'todoat config get' includes reminder and daemon sections
func TestConfigGetAllIncludesReminderAndDaemonCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
reminder:
  enabled: true
  os_notification: true
  log_notification: false
sync:
  enabled: true
  daemon:
    enabled: true
    interval: 300
`)

	stdout := cli.MustExecute("-y", "config", "get")

	testutil.AssertContains(t, stdout, "reminder")
	testutil.AssertContains(t, stdout, "daemon")
}

// --- Config Get Section-Level Keys Test (Issue #65) ---

// TestConfigGetSectionKeyYAMLFormatCLI verifies section-level keys output YAML, not raw Go map format
func TestConfigGetSectionKeyYAMLFormatCLI(t *testing.T) {
	t.Run("sync section", func(t *testing.T) {
		cli := testutil.NewCLITestWithConfig(t)
		cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  enabled: false
  offline_mode: auto
  daemon:
    enabled: true
    interval: 60
`)
		stdout := cli.MustExecute("-y", "config", "get", "sync")

		// Must NOT contain raw Go map format
		testutil.AssertNotContains(t, stdout, "map[")
		// Must contain YAML-formatted keys
		testutil.AssertContains(t, stdout, "enabled:")
		testutil.AssertContains(t, stdout, "daemon:")
	})

	t.Run("sync.daemon subsection", func(t *testing.T) {
		cli := testutil.NewCLITestWithConfig(t)
		cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
sync:
  daemon:
    enabled: true
    interval: 60
    idle_timeout: 300
`)
		stdout := cli.MustExecute("-y", "config", "get", "sync.daemon")

		testutil.AssertNotContains(t, stdout, "map[")
		testutil.AssertContains(t, stdout, "enabled:")
		testutil.AssertContains(t, stdout, "interval:")
	})

	t.Run("reminder section", func(t *testing.T) {
		cli := testutil.NewCLITestWithConfig(t)
		cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
reminder:
  enabled: true
  os_notification: false
  log_notification: true
`)
		stdout := cli.MustExecute("-y", "config", "get", "reminder")

		testutil.AssertNotContains(t, stdout, "map[")
		testutil.AssertContains(t, stdout, "enabled:")
	})

	t.Run("backends section", func(t *testing.T) {
		cli := testutil.NewCLITestWithConfig(t)
		cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`)
		stdout := cli.MustExecute("-y", "config", "get", "backends")

		testutil.AssertNotContains(t, stdout, "map[")
		testutil.AssertContains(t, stdout, "sqlite:")
	})

	t.Run("analytics section", func(t *testing.T) {
		cli := testutil.NewCLITestWithConfig(t)
		cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
analytics:
  enabled: true
  retention_days: 365
`)
		stdout := cli.MustExecute("-y", "config", "get", "analytics")

		testutil.AssertNotContains(t, stdout, "map[")
		testutil.AssertContains(t, stdout, "enabled:")
	})

	t.Run("trash section", func(t *testing.T) {
		cli := testutil.NewCLITestWithConfig(t)
		cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
trash:
  retention_days: 30
`)
		stdout := cli.MustExecute("-y", "config", "get", "trash")

		testutil.AssertNotContains(t, stdout, "map[")
		testutil.AssertContains(t, stdout, "retention_days:")
	})
}

// --- Config JSON Output Test ---

// TestConfigJSONCLI verifies 'todoat --json config get' returns JSON format
func TestConfigJSONCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
no_prompt: false
`)

	stdout := cli.MustExecute("--json", "config", "get")

	// Should be valid JSON with expected structure
	testutil.AssertContains(t, stdout, `"default_backend"`)
	testutil.AssertContains(t, stdout, `"sqlite"`)
	// JSON should have braces
	testutil.AssertContains(t, stdout, "{")
	testutil.AssertContains(t, stdout, "}")
}

// TestConfigSetBooleanNumericValues verifies 'config set' converts 1/0 to true/false for boolean fields (issue #63)
func TestConfigSetBooleanNumericValues(t *testing.T) {
	t.Run("set boolean with 1", func(t *testing.T) {
		cli := testutil.NewCLITestWithConfig(t)

		cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
no_prompt: false
`)

		// Set boolean field with "1" — should be accepted and written as true
		stdout := cli.MustExecute("-y", "config", "set", "no_prompt", "1")
		testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

		// Config must still be loadable after setting with "1"
		stdout = cli.MustExecute("-y", "config", "get", "no_prompt")
		testutil.AssertContains(t, stdout, "true")

		// Verify raw YAML file contains "true", not integer 1
		configPath := cli.ConfigPath()
		raw, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}
		content := string(raw)
		if strings.Contains(content, "no_prompt: 1") {
			t.Errorf("config file contains 'no_prompt: 1' (raw integer), expected 'no_prompt: true'")
		}
	})

	t.Run("set boolean with 0", func(t *testing.T) {
		cli := testutil.NewCLITestWithConfig(t)

		cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
no_prompt: true
`)

		// Set boolean field with "0" — should be accepted and written as false
		stdout := cli.MustExecute("-y", "config", "set", "no_prompt", "0")
		testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

		// Config must still be loadable after setting with "0"
		stdout = cli.MustExecute("-y", "config", "get", "no_prompt")
		testutil.AssertContains(t, stdout, "false")

		// Verify raw YAML file contains "false", not integer 0
		configPath := cli.ConfigPath()
		raw, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}
		content := string(raw)
		if strings.Contains(content, "no_prompt: 0") {
			t.Errorf("config file contains 'no_prompt: 0' (raw integer), expected 'no_prompt: false'")
		}
	})
}

// --- Config Get Missing Keys Tests (Issue #111) ---

// TestConfigGetCacheTTLCLI verifies 'todoat config get cache_ttl' works
func TestConfigGetCacheTTLCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
cache_ttl: "5m"
`)

	stdout := cli.MustExecute("-y", "config", "get", "cache_ttl")
	testutil.AssertContains(t, stdout, "5m")
}

// TestConfigGetUIInteractivePromptCLI verifies 'todoat config get ui.interactive_prompt_for_all_tasks' works
func TestConfigGetUIInteractivePromptCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
ui:
  interactive_prompt_for_all_tasks: true
`)

	stdout := cli.MustExecute("-y", "config", "get", "ui.interactive_prompt_for_all_tasks")
	testutil.AssertContains(t, stdout, "true")
}

// TestConfigGetUISectionCLI verifies 'todoat config get ui' returns the UI section
func TestConfigGetUISectionCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
ui:
  interactive_prompt_for_all_tasks: true
`)

	stdout := cli.MustExecute("-y", "config", "get", "ui")
	testutil.AssertNotContains(t, stdout, "map[")
	testutil.AssertContains(t, stdout, "interactive_prompt_for_all_tasks:")
}

// TestConfigGetLoggingBackgroundEnabledCLI verifies 'todoat config get logging.background_enabled' works
func TestConfigGetLoggingBackgroundEnabledCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
logging:
  background_enabled: true
`)

	stdout := cli.MustExecute("-y", "config", "get", "logging.background_enabled")
	testutil.AssertContains(t, stdout, "true")
}

// TestConfigGetLoggingSectionCLI verifies 'todoat config get logging' returns the logging section
func TestConfigGetLoggingSectionCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
default_backend: sqlite
logging:
  background_enabled: false
`)

	stdout := cli.MustExecute("-y", "config", "get", "logging")
	testutil.AssertNotContains(t, stdout, "map[")
	testutil.AssertContains(t, stdout, "background_enabled:")
}
