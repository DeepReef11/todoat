package views_test

import (
	"os"
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// =============================================================================
// Default View Configuration Tests (033-default-view-config)
// =============================================================================

// TestDefaultViewConfig verifies that `default_view: myview` in config is respected
func TestDefaultViewConfig(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViewsAndConfig(t)

	// Create a custom view YAML file that shows due_date
	// The default tree output does NOT show due_date, so if due_date appears,
	// we know the custom view from config is being used
	viewYAML := `name: myview
fields:
  - name: summary
    width: 50
  - name: due_date
    width: 12
`
	viewPath := viewsDir + "/myview.yaml"
	if err := os.WriteFile(viewPath, []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Set default_view in config
	cli.SetConfigValue("default_view", "myview")

	// Create a list and add tasks with due date
	cli.MustExecute("-y", "list", "create", "ConfigViewTest")
	cli.MustExecute("-y", "ConfigViewTest", "add", "Task with due date", "--due-date", "2026-03-15")

	// List tasks WITHOUT -v flag - should use default_view from config
	stdout, _, exitCode := cli.Execute("-y", "ConfigViewTest")

	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Task with due date")
	// The due date should appear ONLY if the custom view from config is used
	// Default tree output does NOT show due dates
	testutil.AssertContains(t, stdout, "2026-03-15")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestDefaultViewFallback verifies fallback to "default" if configured view not found
func TestDefaultViewFallback(t *testing.T) {
	cli, _ := testutil.NewCLITestWithViewsAndConfig(t)

	// Set default_view to a non-existent view (but not "missing-view" which should warn)
	// This tests that when the view file doesn't exist, we fallback gracefully
	cli.SetConfigValue("default_view", "nonexistent-but-fallback")

	// Create a list and add tasks
	cli.MustExecute("-y", "list", "create", "FallbackViewTest")
	cli.MustExecute("-y", "FallbackViewTest", "add", "Test task")

	// List tasks - should fallback to default view and still work
	stdout, stderr, exitCode := cli.Execute("-y", "FallbackViewTest")

	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Test task")
	// Should show a warning about the missing view
	combinedOutput := stdout + stderr
	if !strings.Contains(strings.ToLower(combinedOutput), "warning") || !strings.Contains(strings.ToLower(combinedOutput), "nonexistent-but-fallback") {
		t.Errorf("expected warning about missing default view, got stdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestDefaultViewOverride verifies that `-v` flag overrides config default
func TestDefaultViewOverride(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViewsAndConfig(t)

	// Create two custom views with different fields
	// view1 shows due_date, view2 shows start_date
	view1YAML := `name: view1
fields:
  - name: summary
    width: 30
  - name: due_date
    width: 12
`
	view2YAML := `name: view2
fields:
  - name: summary
    width: 50
  - name: start_date
    width: 12
`
	if err := os.WriteFile(viewsDir+"/view1.yaml", []byte(view1YAML), 0644); err != nil {
		t.Fatalf("failed to write view1 file: %v", err)
	}
	if err := os.WriteFile(viewsDir+"/view2.yaml", []byte(view2YAML), 0644); err != nil {
		t.Fatalf("failed to write view2 file: %v", err)
	}

	// Set default_view to view1 (shows due_date)
	cli.SetConfigValue("default_view", "view1")

	// Create a list and add tasks with both due_date and start_date
	cli.MustExecute("-y", "list", "create", "OverrideViewTest")
	cli.MustExecute("-y", "OverrideViewTest", "add", "Task with dates", "--due-date", "2026-03-20", "--start-date", "2026-03-10")

	// First, verify WITHOUT -v flag that config's default_view (view1) is used
	stdout1, _, exitCode1 := cli.Execute("-y", "OverrideViewTest")
	testutil.AssertExitCode(t, exitCode1, 0)
	testutil.AssertContains(t, stdout1, "2026-03-20")    // due_date from view1
	testutil.AssertNotContains(t, stdout1, "2026-03-10") // start_date NOT in view1

	// Now test -v flag override - use view2 which shows start_date instead of due_date
	stdout2, _, exitCode2 := cli.Execute("-y", "OverrideViewTest", "-v", "view2")

	testutil.AssertExitCode(t, exitCode2, 0)
	testutil.AssertContains(t, stdout2, "Task with dates")
	// view2 shows start_date, NOT due_date
	testutil.AssertContains(t, stdout2, "2026-03-10")    // start_date from view2
	testutil.AssertNotContains(t, stdout2, "2026-03-20") // due_date NOT in view2
	testutil.AssertResultCode(t, stdout2, testutil.ResultInfoOnly)
}

// TestDefaultViewBuiltin verifies that built-in views can be set as default
func TestDefaultViewBuiltin(t *testing.T) {
	cli, _ := testutil.NewCLITestWithViewsAndConfig(t)

	// Set default_view to built-in "all" view
	cli.SetConfigValue("default_view", "all")

	// Create a list and add a task with many fields
	cli.MustExecute("-y", "list", "create", "BuiltinViewTest")
	cli.MustExecute("-y", "BuiltinViewTest", "add", "Full task", "-p", "2", "--due-date", "2026-01-31", "--tag", "work")

	// List tasks without -v flag - should use "all" view from config
	stdout, _, exitCode := cli.Execute("-y", "BuiltinViewTest")

	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Full task")
	// "all" view should show due date
	testutil.AssertContains(t, stdout, "2026-01-31")
	// "all" view should show tags
	testutil.AssertContains(t, stdout, "work")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestDefaultViewCustom verifies that custom views from ~/.config/todoat/views/ can be set as default
func TestDefaultViewCustom(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViewsAndConfig(t)

	// Create a custom view that shows start_date (not shown in default tree output)
	viewYAML := `name: custom_with_start
fields:
  - name: summary
    width: 40
  - name: start_date
    width: 12
`
	if err := os.WriteFile(viewsDir+"/custom_with_start.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Set default_view to the custom view
	cli.SetConfigValue("default_view", "custom_with_start")

	// Create a list and add tasks with start_date
	cli.MustExecute("-y", "list", "create", "CustomDefaultTest")
	cli.MustExecute("-y", "CustomDefaultTest", "add", "Task with start", "--start-date", "2026-04-01")

	// List tasks - should use custom "custom_with_start" view
	stdout, _, exitCode := cli.Execute("-y", "CustomDefaultTest")

	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Task with start")
	// start_date should be visible (only if custom view is used; default tree output doesn't show it)
	testutil.AssertContains(t, stdout, "2026-04-01")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestDefaultViewMissing verifies that a warning is shown if configured default view doesn't exist
func TestDefaultViewMissing(t *testing.T) {
	cli, _ := testutil.NewCLITestWithViewsAndConfig(t)

	// Set default_view to a non-existent view
	cli.SetConfigValue("default_view", "missing-view")

	// Create a list and add tasks
	cli.MustExecute("-y", "list", "create", "MissingViewTest")
	cli.MustExecute("-y", "MissingViewTest", "add", "Test task")

	// List tasks - should show warning about missing view but still work
	stdout, stderr, exitCode := cli.Execute("-y", "MissingViewTest")

	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Test task")

	// Should show a warning about the missing view
	combinedOutput := stdout + stderr
	if !strings.Contains(strings.ToLower(combinedOutput), "warning") {
		t.Errorf("expected warning about missing default view, got stdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(combinedOutput, "missing-view") {
		t.Errorf("expected warning to mention 'missing-view', got stdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}
