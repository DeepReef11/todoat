package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"todoat/internal/testutil"
)

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
