package views_test

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// =============================================================================
// Plugin Formatters Tests (057-plugin-formatters)
// =============================================================================

// TestPluginFormatterStatus verifies that view with status plugin displays custom formatted status (e.g., emoji)
func TestPluginFormatterStatus(t *testing.T) {
	cli, viewsDir, tmpDir := testutil.NewCLITestWithViewsAndTmpDir(t)

	// Create plugin directory
	pluginsDir := tmpDir + "/todoat/plugins"
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatalf("failed to create plugins directory: %v", err)
	}

	// Create status emoji plugin script
	pluginScript := `#!/bin/bash
read -r task
status=$(echo "$task" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
case "$status" in
  "TODO") echo "📋";;
  "DONE") echo "✅";;
  "IN-PROGRESS") echo "🔄";;
  "CANCELLED") echo "❌";;
  *) echo "$status";;
esac
`
	pluginPath := pluginsDir + "/status-emoji.sh"
	if err := os.WriteFile(pluginPath, []byte(pluginScript), 0755); err != nil {
		t.Fatalf("failed to write plugin script: %v", err)
	}

	// Create view with plugin configuration
	viewYAML := `name: emoji_status
fields:
  - name: status
    width: 5
    plugin:
      command: ` + pluginPath + `
      timeout: 1000
  - name: summary
    width: 40
`
	if err := os.WriteFile(viewsDir+"/emoji_status.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create list and add tasks
	cli.MustExecute("-y", "list", "create", "PluginStatusTest")
	cli.MustExecute("-y", "PluginStatusTest", "add", "Todo task")
	cli.MustExecute("-y", "PluginStatusTest", "add", "Done task")
	cli.MustExecute("-y", "PluginStatusTest", "complete", "Done task")

	// List tasks with plugin view
	stdout, _, exitCode := cli.Execute("-y", "PluginStatusTest", "-v", "emoji_status")

	testutil.AssertExitCode(t, exitCode, 0)
	// Plugin should format status as emoji
	testutil.AssertContains(t, stdout, "📋")
	testutil.AssertContains(t, stdout, "✅")
	testutil.AssertContains(t, stdout, "Todo task")
	testutil.AssertContains(t, stdout, "Done task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPluginFormatterPriority verifies that view with priority plugin displays custom priority
func TestPluginFormatterPriority(t *testing.T) {
	cli, viewsDir, tmpDir := testutil.NewCLITestWithViewsAndTmpDir(t)

	// Create plugin directory
	pluginsDir := tmpDir + "/todoat/plugins"
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatalf("failed to create plugins directory: %v", err)
	}

	// Create priority formatter plugin script - maps priority to text
	pluginScript := `#!/bin/bash
read -r task
priority=$(echo "$task" | grep -o '"priority":[0-9]*' | cut -d':' -f2)
if [ -z "$priority" ] || [ "$priority" = "0" ]; then
  echo "none"
elif [ "$priority" -le 3 ]; then
  echo "HIGH"
elif [ "$priority" -le 6 ]; then
  echo "MEDIUM"
else
  echo "LOW"
fi
`
	pluginPath := pluginsDir + "/priority-text.sh"
	if err := os.WriteFile(pluginPath, []byte(pluginScript), 0755); err != nil {
		t.Fatalf("failed to write plugin script: %v", err)
	}

	// Create view with plugin configuration
	viewYAML := `name: priority_text
fields:
  - name: summary
    width: 40
  - name: priority
    width: 10
    plugin:
      command: ` + pluginPath + `
      timeout: 1000
`
	if err := os.WriteFile(viewsDir+"/priority_text.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create list and add tasks with different priorities
	cli.MustExecute("-y", "list", "create", "PluginPriorityTest")
	cli.MustExecute("-y", "PluginPriorityTest", "add", "High priority task", "-p", "1")
	cli.MustExecute("-y", "PluginPriorityTest", "add", "Medium priority task", "-p", "5")
	cli.MustExecute("-y", "PluginPriorityTest", "add", "Low priority task", "-p", "9")

	// List tasks with plugin view
	stdout, _, exitCode := cli.Execute("-y", "PluginPriorityTest", "-v", "priority_text")

	testutil.AssertExitCode(t, exitCode, 0)
	// Plugin should format priority as text
	testutil.AssertContains(t, stdout, "HIGH")
	testutil.AssertContains(t, stdout, "MEDIUM")
	testutil.AssertContains(t, stdout, "LOW")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPluginFormatterDate verifies that view with date plugin displays relative dates
func TestPluginFormatterDate(t *testing.T) {
	cli, viewsDir, tmpDir := testutil.NewCLITestWithViewsAndTmpDir(t)

	// Create plugin directory
	pluginsDir := tmpDir + "/todoat/plugins"
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatalf("failed to create plugins directory: %v", err)
	}

	// Create date formatter plugin script that outputs "relative" for any date
	pluginScript := `#!/bin/bash
read -r task
due_date=$(echo "$task" | grep -o '"due_date":"[^"]*"' | cut -d'"' -f4)
if [ -z "$due_date" ] || [ "$due_date" = "null" ]; then
  echo ""
else
  # Output a recognizable marker for testing
  echo "due:relative"
fi
`
	pluginPath := pluginsDir + "/relative-date.sh"
	if err := os.WriteFile(pluginPath, []byte(pluginScript), 0755); err != nil {
		t.Fatalf("failed to write plugin script: %v", err)
	}

	// Create view with plugin configuration
	viewYAML := `name: relative_date
fields:
  - name: summary
    width: 40
  - name: due_date
    width: 15
    plugin:
      command: ` + pluginPath + `
      timeout: 1000
`
	if err := os.WriteFile(viewsDir+"/relative_date.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create list and add task with due date
	cli.MustExecute("-y", "list", "create", "PluginDateTest")
	cli.MustExecute("-y", "PluginDateTest", "add", "Task with due date", "--due-date", "2026-01-31")

	// List tasks with plugin view
	stdout, _, exitCode := cli.Execute("-y", "PluginDateTest", "-v", "relative_date")

	testutil.AssertExitCode(t, exitCode, 0)
	// Plugin should format date as relative
	testutil.AssertContains(t, stdout, "due:relative")
	testutil.AssertContains(t, stdout, "Task with due date")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPluginTimeout verifies that plugin exceeding timeout falls back to raw value
func TestPluginTimeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping timeout test on Windows")
	}

	cli, viewsDir, tmpDir := testutil.NewCLITestWithViewsAndTmpDir(t)

	// Create plugin directory
	pluginsDir := tmpDir + "/todoat/plugins"
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatalf("failed to create plugins directory: %v", err)
	}

	// Create slow plugin that sleeps longer than timeout
	pluginScript := `#!/bin/bash
sleep 5
echo "should not appear"
`
	pluginPath := pluginsDir + "/slow-plugin.sh"
	if err := os.WriteFile(pluginPath, []byte(pluginScript), 0755); err != nil {
		t.Fatalf("failed to write plugin script: %v", err)
	}

	// Create view with short timeout (100ms)
	viewYAML := `name: timeout_test
fields:
  - name: status
    width: 15
    plugin:
      command: ` + pluginPath + `
      timeout: 100
  - name: summary
    width: 40
`
	if err := os.WriteFile(viewsDir+"/timeout_test.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create list and add task
	cli.MustExecute("-y", "list", "create", "PluginTimeoutTest")
	cli.MustExecute("-y", "PluginTimeoutTest", "add", "Timeout test task")

	// List tasks with plugin view
	stdout, _, exitCode := cli.Execute("-y", "PluginTimeoutTest", "-v", "timeout_test")

	testutil.AssertExitCode(t, exitCode, 0)
	// Plugin should timeout and fall back to raw value
	testutil.AssertNotContains(t, stdout, "should not appear")
	testutil.AssertContains(t, stdout, "Timeout test task")
	// Should show fallback value (raw status like [TODO] or TODO)
	if !strings.Contains(stdout, "TODO") {
		t.Errorf("expected fallback to raw status value, got:\n%s", stdout)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPluginError verifies that plugin returning non-zero exit shows fallback value gracefully
func TestPluginError(t *testing.T) {
	cli, viewsDir, tmpDir := testutil.NewCLITestWithViewsAndTmpDir(t)

	// Create plugin directory
	pluginsDir := tmpDir + "/todoat/plugins"
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatalf("failed to create plugins directory: %v", err)
	}

	// Create failing plugin
	pluginScript := `#!/bin/bash
exit 1
`
	pluginPath := pluginsDir + "/failing-plugin.sh"
	if err := os.WriteFile(pluginPath, []byte(pluginScript), 0755); err != nil {
		t.Fatalf("failed to write plugin script: %v", err)
	}

	// Create view with failing plugin
	viewYAML := `name: error_test
fields:
  - name: status
    width: 15
    plugin:
      command: ` + pluginPath + `
      timeout: 1000
  - name: summary
    width: 40
`
	if err := os.WriteFile(viewsDir+"/error_test.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create list and add task
	cli.MustExecute("-y", "list", "create", "PluginErrorTest")
	cli.MustExecute("-y", "PluginErrorTest", "add", "Error test task")

	// List tasks with plugin view - should not fail
	stdout, _, exitCode := cli.Execute("-y", "PluginErrorTest", "-v", "error_test")

	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Error test task")
	// Should show fallback value (raw status)
	if !strings.Contains(stdout, "TODO") {
		t.Errorf("expected fallback to raw status value on error, got:\n%s", stdout)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPluginInvalidOutput verifies that plugin returning invalid JSON falls back to raw value
func TestPluginInvalidOutput(t *testing.T) {
	cli, viewsDir, tmpDir := testutil.NewCLITestWithViewsAndTmpDir(t)

	// Create plugin directory
	pluginsDir := tmpDir + "/todoat/plugins"
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatalf("failed to create plugins directory: %v", err)
	}

	// Create plugin that outputs nothing (empty output)
	pluginScript := `#!/bin/bash
# Output nothing - empty stdout
`
	pluginPath := pluginsDir + "/empty-plugin.sh"
	if err := os.WriteFile(pluginPath, []byte(pluginScript), 0755); err != nil {
		t.Fatalf("failed to write plugin script: %v", err)
	}

	// Create view with empty output plugin
	viewYAML := `name: invalid_output
fields:
  - name: status
    width: 15
    plugin:
      command: ` + pluginPath + `
      timeout: 1000
  - name: summary
    width: 40
`
	if err := os.WriteFile(viewsDir+"/invalid_output.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create list and add task
	cli.MustExecute("-y", "list", "create", "PluginInvalidTest")
	cli.MustExecute("-y", "PluginInvalidTest", "add", "Invalid output test task")

	// List tasks with plugin view - should not fail
	stdout, _, exitCode := cli.Execute("-y", "PluginInvalidTest", "-v", "invalid_output")

	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Invalid output test task")
	// Should show fallback value (raw status)
	if !strings.Contains(stdout, "TODO") {
		t.Errorf("expected fallback to raw status value on empty output, got:\n%s", stdout)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPluginNotFound verifies that non-existent plugin path shows warning and uses raw value
func TestPluginNotFound(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViews(t)

	// Create view with non-existent plugin path
	viewYAML := `name: missing_plugin
fields:
  - name: status
    width: 15
    plugin:
      command: /nonexistent/path/to/plugin.sh
      timeout: 1000
  - name: summary
    width: 40
`
	if err := os.WriteFile(viewsDir+"/missing_plugin.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create list and add task
	cli.MustExecute("-y", "list", "create", "PluginNotFoundTest")
	cli.MustExecute("-y", "PluginNotFoundTest", "add", "Missing plugin test task")

	// List tasks with plugin view - should not fail
	stdout, _, exitCode := cli.Execute("-y", "PluginNotFoundTest", "-v", "missing_plugin")

	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Missing plugin test task")
	// Should show fallback value (raw status)
	if !strings.Contains(stdout, "TODO") {
		t.Errorf("expected fallback to raw status value when plugin not found, got:\n%s", stdout)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPluginEnvironmentVariables verifies that plugin receives configured environment variables
func TestPluginEnvironmentVariables(t *testing.T) {
	cli, viewsDir, tmpDir := testutil.NewCLITestWithViewsAndTmpDir(t)

	// Create plugin directory
	pluginsDir := tmpDir + "/todoat/plugins"
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatalf("failed to create plugins directory: %v", err)
	}

	// Create plugin that echoes environment variable
	pluginScript := `#!/bin/bash
read -r task
echo "$CUSTOM_THEME"
`
	pluginPath := pluginsDir + "/env-plugin.sh"
	if err := os.WriteFile(pluginPath, []byte(pluginScript), 0755); err != nil {
		t.Fatalf("failed to write plugin script: %v", err)
	}

	// Create view with plugin and env configuration
	viewYAML := `name: env_test
fields:
  - name: status
    width: 15
    plugin:
      command: ` + pluginPath + `
      timeout: 1000
      env:
        CUSTOM_THEME: dark_mode
  - name: summary
    width: 40
`
	if err := os.WriteFile(viewsDir+"/env_test.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create list and add task
	cli.MustExecute("-y", "list", "create", "PluginEnvTest")
	cli.MustExecute("-y", "PluginEnvTest", "add", "Env test task")

	// List tasks with plugin view
	stdout, _, exitCode := cli.Execute("-y", "PluginEnvTest", "-v", "env_test")

	testutil.AssertExitCode(t, exitCode, 0)
	// Plugin should output the env variable value
	testutil.AssertContains(t, stdout, "dark_mode")
	testutil.AssertContains(t, stdout, "Env test task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPluginReceivesTaskJSON verifies that plugins receive complete task data as JSON
func TestPluginReceivesTaskJSON(t *testing.T) {
	cli, viewsDir, tmpDir := testutil.NewCLITestWithViewsAndTmpDir(t)

	// Create plugin directory
	pluginsDir := tmpDir + "/todoat/plugins"
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatalf("failed to create plugins directory: %v", err)
	}

	// Create plugin that extracts and displays task summary from JSON
	pluginScript := `#!/bin/bash
read -r task
summary=$(echo "$task" | grep -o '"summary":"[^"]*"' | cut -d'"' -f4)
echo "GOT:$summary"
`
	pluginPath := pluginsDir + "/json-test.sh"
	if err := os.WriteFile(pluginPath, []byte(pluginScript), 0755); err != nil {
		t.Fatalf("failed to write plugin script: %v", err)
	}

	// Create view with plugin
	viewYAML := `name: json_test
fields:
  - name: status
    width: 30
    plugin:
      command: ` + pluginPath + `
      timeout: 1000
  - name: summary
    width: 40
`
	if err := os.WriteFile(viewsDir+"/json_test.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create list and add task with recognizable name
	cli.MustExecute("-y", "list", "create", "PluginJSONTest")
	cli.MustExecute("-y", "PluginJSONTest", "add", "MyUniqueTaskName123")

	// List tasks with plugin view
	stdout, _, exitCode := cli.Execute("-y", "PluginJSONTest", "-v", "json_test")

	testutil.AssertExitCode(t, exitCode, 0)
	// Plugin should have received and parsed the task JSON
	testutil.AssertContains(t, stdout, "GOT:MyUniqueTaskName123")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// =============================================================================
// Plugin Security Tests (Issue #73)
// =============================================================================

// TestPluginCommandMustBeInPluginDir verifies that plugin commands outside the plugin directory are rejected
func TestPluginCommandMustBeInPluginDir(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViews(t)

	// Create view that tries to use /usr/bin/id (arbitrary system command)
	viewYAML := `name: malicious
fields:
  - name: status
    width: 15
    plugin:
      command: /usr/bin/id
      timeout: 5000
  - name: summary
    width: 40
`
	if err := os.WriteFile(viewsDir+"/malicious.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create list and add task
	cli.MustExecute("-y", "list", "create", "SecurityTest")
	cli.MustExecute("-y", "SecurityTest", "add", "Security test task")

	// List tasks with malicious view - should NOT execute /usr/bin/id
	// Instead, should fall back to raw value since plugin command is rejected
	stdout, _, exitCode := cli.Execute("-y", "SecurityTest", "-v", "malicious")

	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Security test task")
	// Should show fallback value (raw status) since /usr/bin/id is outside plugin dir
	if !strings.Contains(stdout, "TODO") {
		t.Errorf("expected fallback to raw status value when plugin outside plugin dir, got:\n%s", stdout)
	}
	// Verify that the id command output is NOT in the output
	// (uid=xxx, gid=xxx patterns from /usr/bin/id)
	if strings.Contains(stdout, "uid=") || strings.Contains(stdout, "gid=") {
		t.Errorf("SECURITY: arbitrary command /usr/bin/id was executed! output:\n%s", stdout)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPluginCommandInPluginDirWorks verifies that plugin commands inside the plugin directory work
func TestPluginCommandInPluginDirWorks(t *testing.T) {
	cli, viewsDir, tmpDir := testutil.NewCLITestWithViewsAndTmpDir(t)

	// Create plugin directory - must be in XDG_CONFIG_HOME/todoat/plugins
	// Tests set XDG_CONFIG_HOME to tmpDir/xdg-config
	pluginsDir := tmpDir + "/todoat/plugins"
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatalf("failed to create plugins directory: %v", err)
	}

	// Create a valid plugin in the plugin directory
	pluginScript := `#!/bin/bash
echo "PLUGIN_OUTPUT"
`
	pluginPath := pluginsDir + "/valid-plugin.sh"
	if err := os.WriteFile(pluginPath, []byte(pluginScript), 0755); err != nil {
		t.Fatalf("failed to write plugin script: %v", err)
	}

	// Create view with plugin in the plugin directory
	viewYAML := `name: valid_plugin
fields:
  - name: status
    width: 20
    plugin:
      command: ` + pluginPath + `
      timeout: 1000
  - name: summary
    width: 40
`
	if err := os.WriteFile(viewsDir+"/valid_plugin.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create list and add task
	cli.MustExecute("-y", "list", "create", "ValidPluginTest")
	cli.MustExecute("-y", "ValidPluginTest", "add", "Valid plugin test task")

	// List tasks with valid plugin view - should execute the plugin
	stdout, _, exitCode := cli.Execute("-y", "ValidPluginTest", "-v", "valid_plugin")

	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Valid plugin test task")
	// Plugin should have executed and returned output
	testutil.AssertContains(t, stdout, "PLUGIN_OUTPUT")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPluginPathTraversalRejected verifies that path traversal in plugin commands is rejected
func TestPluginPathTraversalRejected(t *testing.T) {
	cli, viewsDir, tmpDir := testutil.NewCLITestWithViewsAndTmpDir(t)

	// Create plugin directory - must be in XDG_CONFIG_HOME/todoat/plugins
	// Tests set XDG_CONFIG_HOME to tmpDir/xdg-config
	pluginsDir := tmpDir + "/todoat/plugins"
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatalf("failed to create plugins directory: %v", err)
	}

	// Create a script outside the plugins directory
	outsideScript := `#!/bin/bash
echo "OUTSIDE_PLUGIN_DIR"
`
	outsidePath := tmpDir + "/outside-plugin.sh"
	if err := os.WriteFile(outsidePath, []byte(outsideScript), 0755); err != nil {
		t.Fatalf("failed to write outside script: %v", err)
	}

	// Create view with path traversal attempt
	viewYAML := `name: traversal
fields:
  - name: status
    width: 25
    plugin:
      command: ` + pluginsDir + `/../outside-plugin.sh
      timeout: 1000
  - name: summary
    width: 40
`
	if err := os.WriteFile(viewsDir+"/traversal.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create list and add task
	cli.MustExecute("-y", "list", "create", "TraversalTest")
	cli.MustExecute("-y", "TraversalTest", "add", "Traversal test task")

	// List tasks with traversal view - should NOT execute the outside script
	stdout, _, exitCode := cli.Execute("-y", "TraversalTest", "-v", "traversal")

	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Traversal test task")
	// Should show fallback value since path traversal is detected
	if !strings.Contains(stdout, "TODO") {
		t.Errorf("expected fallback to raw status value when path traversal detected, got:\n%s", stdout)
	}
	// Verify that the outside script output is NOT in the output
	if strings.Contains(stdout, "OUTSIDE_PLUGIN_DIR") {
		t.Errorf("SECURITY: path traversal allowed execution of script outside plugin dir! output:\n%s", stdout)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}
