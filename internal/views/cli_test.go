package views_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"todoat/internal/testutil"
)

// =============================================================================
// Views and Customization Tests (015-views-customization)
// =============================================================================

// TestDefaultView verifies that `todoat MyList` displays tasks with default view (status, summary, priority)
func TestDefaultViewViewsCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add tasks with different priorities
	cli.MustExecute("-y", "list", "create", "DefaultViewTest")
	cli.MustExecute("-y", "DefaultViewTest", "add", "High priority task", "-p", "1")
	cli.MustExecute("-y", "DefaultViewTest", "add", "Medium priority task", "-p", "5")
	cli.MustExecute("-y", "DefaultViewTest", "update", "High priority task", "-s", "IN-PROGRESS")

	// List tasks (should show default view with status, summary, priority)
	stdout, _, exitCode := cli.Execute("-y", "DefaultViewTest")

	testutil.AssertExitCode(t, exitCode, 0)
	// Default view should include status, summary, and priority
	testutil.AssertContains(t, stdout, "High priority task")
	testutil.AssertContains(t, stdout, "Medium priority task")
	// Status should be visible
	if !strings.Contains(stdout, "TODO") && !strings.Contains(stdout, "IN-PROGRESS") {
		t.Errorf("expected status indicators in default view, got:\n%s", stdout)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestAllView verifies that `todoat MyList -v all` displays all task metadata fields
func TestAllViewViewsCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add a task with multiple attributes
	cli.MustExecute("-y", "list", "create", "AllViewTest")
	cli.MustExecute("-y", "AllViewTest", "add", "Full metadata task", "-p", "3", "--due-date", "2026-01-31", "--tag", "work,urgent")

	// List tasks with 'all' view
	stdout, _, exitCode := cli.Execute("-y", "AllViewTest", "-v", "all")

	testutil.AssertExitCode(t, exitCode, 0)
	// All view should show more fields than default
	testutil.AssertContains(t, stdout, "Full metadata task")
	// Should show due date (Jan 31 format for display)
	testutil.AssertContains(t, stdout, "Jan 31")
	// Should show tags
	if !strings.Contains(stdout, "work") || !strings.Contains(stdout, "urgent") {
		t.Errorf("expected tags in 'all' view, got:\n%s", stdout)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestCustomViewSelection verifies that `todoat MyList -v myview` loads view from views directory
func TestCustomViewSelectionViewsCLI(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViews(t)

	// Create a custom view YAML file that shows only summary and status
	viewYAML := `name: minimal
fields:
  - name: summary
    width: 40
  - name: status
    width: 12
`
	viewPath := viewsDir + "/minimal.yaml"
	if err := os.WriteFile(viewPath, []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create a list and add tasks
	cli.MustExecute("-y", "list", "create", "CustomViewTest")
	cli.MustExecute("-y", "CustomViewTest", "add", "Task with priority", "-p", "1")

	// List tasks with custom view
	stdout, _, exitCode := cli.Execute("-y", "CustomViewTest", "-v", "minimal")

	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Task with priority")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestViewListCommand verifies that `todoat view list` shows all available views (built-in and custom)
func TestViewListCommandViewsCLI(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViews(t)

	// Create a custom view
	viewYAML := `name: custom
fields:
  - name: summary
`
	if err := os.WriteFile(viewsDir+"/custom.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// List available views
	stdout, _, exitCode := cli.Execute("-y", "view", "list")

	testutil.AssertExitCode(t, exitCode, 0)
	// Should show built-in views
	testutil.AssertContains(t, stdout, "default")
	testutil.AssertContains(t, stdout, "all")
	// Should show custom view
	testutil.AssertContains(t, stdout, "custom")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestViewFieldOrdering verifies that custom view with reordered fields displays columns in specified order
func TestViewFieldOrderingViewsCLI(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViews(t)

	// Create a view with priority BEFORE summary (reversed from default)
	viewYAML := `name: priority_first
fields:
  - name: priority
    width: 8
  - name: summary
    width: 40
  - name: status
    width: 12
`
	if err := os.WriteFile(viewsDir+"/priority_first.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create a list and add a task
	cli.MustExecute("-y", "list", "create", "FieldOrderTest")
	cli.MustExecute("-y", "FieldOrderTest", "add", "Important task", "-p", "1")

	// List tasks with priority_first view
	stdout, _, exitCode := cli.Execute("-y", "FieldOrderTest", "-v", "priority_first")

	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Important task")
	// Priority should appear before the summary in output
	// The priority indicator (1 or P1) should appear before "Important task"
	prioIdx := strings.Index(stdout, "1")
	summIdx := strings.Index(stdout, "Important task")
	if prioIdx >= summIdx {
		t.Errorf("expected priority to appear before summary in output, got:\n%s", stdout)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestViewFiltering verifies that view with filters only shows matching tasks (e.g., status != DONE)
func TestViewFilteringViewsCLI(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViews(t)

	// Create a view that filters out completed tasks
	viewYAML := `name: active_only
fields:
  - name: status
  - name: summary
filters:
  - field: status
    operator: ne
    value: DONE
`
	if err := os.WriteFile(viewsDir+"/active_only.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create tasks with different statuses
	cli.MustExecute("-y", "list", "create", "FilterViewTest")
	cli.MustExecute("-y", "FilterViewTest", "add", "Active task")
	cli.MustExecute("-y", "FilterViewTest", "add", "Completed task")
	cli.MustExecute("-y", "FilterViewTest", "complete", "Completed task")

	// List tasks with active_only view
	stdout, _, exitCode := cli.Execute("-y", "FilterViewTest", "-v", "active_only")

	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Active task")
	testutil.AssertNotContains(t, stdout, "Completed task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestViewSorting verifies that view with sort rules orders tasks correctly (multi-level sort)
func TestViewSortingViewsCLI(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViews(t)

	// Create a view that sorts by priority ascending, then by summary
	viewYAML := `name: priority_sorted
fields:
  - name: priority
  - name: summary
sort:
  - field: priority
    direction: asc
  - field: summary
    direction: asc
`
	if err := os.WriteFile(viewsDir+"/priority_sorted.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create tasks with different priorities
	cli.MustExecute("-y", "list", "create", "SortViewTest")
	cli.MustExecute("-y", "SortViewTest", "add", "Low priority task", "-p", "9")
	cli.MustExecute("-y", "SortViewTest", "add", "High priority task", "-p", "1")
	cli.MustExecute("-y", "SortViewTest", "add", "Medium priority task", "-p", "5")

	// List tasks with priority_sorted view
	stdout, _, exitCode := cli.Execute("-y", "SortViewTest", "-v", "priority_sorted")

	testutil.AssertExitCode(t, exitCode, 0)
	// Tasks should appear in priority order: High (1), Medium (5), Low (9)
	highIdx := strings.Index(stdout, "High priority task")
	medIdx := strings.Index(stdout, "Medium priority task")
	lowIdx := strings.Index(stdout, "Low priority task")
	if highIdx >= medIdx || medIdx >= lowIdx {
		t.Errorf("expected tasks sorted by priority (high, medium, low), got:\n%s", stdout)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestViewDateFilter verifies that view filters with relative dates (`today`, `+7d`, `-3d`) work correctly
func TestViewDateFilterViewsCLI(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViews(t)

	// Create a view that shows tasks due within next 7 days
	viewYAML := `name: due_soon
fields:
  - name: summary
  - name: due_date
filters:
  - field: due_date
    operator: lte
    value: "+7d"
  - field: due_date
    operator: gte
    value: "today"
`
	if err := os.WriteFile(viewsDir+"/due_soon.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Calculate dates relative to today for test data
	today := time.Now()
	threeDays := today.AddDate(0, 0, 3).Format("2006-01-02")
	tenDays := today.AddDate(0, 0, 10).Format("2006-01-02")
	yesterday := today.AddDate(0, 0, -1).Format("2006-01-02")

	// Create tasks with different due dates
	cli.MustExecute("-y", "list", "create", "DateFilterTest")
	cli.MustExecute("-y", "DateFilterTest", "add", "Due soon task", "--due-date", threeDays)
	cli.MustExecute("-y", "DateFilterTest", "add", "Due later task", "--due-date", tenDays)
	cli.MustExecute("-y", "DateFilterTest", "add", "Overdue task", "--due-date", yesterday)
	cli.MustExecute("-y", "DateFilterTest", "add", "No due date task")

	// List tasks with due_soon view
	stdout, _, exitCode := cli.Execute("-y", "DateFilterTest", "-v", "due_soon")

	testutil.AssertExitCode(t, exitCode, 0)
	// Should show only tasks due within 7 days from today
	testutil.AssertContains(t, stdout, "Due soon task")
	testutil.AssertNotContains(t, stdout, "Due later task")
	testutil.AssertNotContains(t, stdout, "Overdue task")
	testutil.AssertNotContains(t, stdout, "No due date task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestViewTagFilter verifies that view filters on tags/categories work with `contains` and `in` operators
func TestViewTagFilterViewsCLI(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViews(t)

	// Create a view that filters by tags using 'contains' operator
	viewYAML := `name: work_only
fields:
  - name: summary
  - name: tags
filters:
  - field: tags
    operator: contains
    value: work
`
	if err := os.WriteFile(viewsDir+"/work_only.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create tasks with different tags
	cli.MustExecute("-y", "list", "create", "TagFilterTest")
	cli.MustExecute("-y", "TagFilterTest", "add", "Work task", "--tag", "work")
	cli.MustExecute("-y", "TagFilterTest", "add", "Home task", "--tag", "home")
	cli.MustExecute("-y", "TagFilterTest", "add", "Work and home task", "--tag", "work,home")

	// List tasks with work_only view
	stdout, _, exitCode := cli.Execute("-y", "TagFilterTest", "-v", "work_only")

	testutil.AssertExitCode(t, exitCode, 0)
	// Should show tasks with 'work' tag
	testutil.AssertContains(t, stdout, "Work task")
	testutil.AssertContains(t, stdout, "Work and home task")
	testutil.AssertNotContains(t, stdout, "Home task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestViewHierarchyPreserved verifies that custom views maintain parent-child tree structure display
func TestViewHierarchyPreservedViewsCLI(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViews(t)

	// Create a simple custom view
	viewYAML := `name: simple
fields:
  - name: summary
  - name: status
`
	if err := os.WriteFile(viewsDir+"/simple.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create a list with hierarchical tasks
	cli.MustExecute("-y", "list", "create", "HierarchyViewTest")
	cli.MustExecute("-y", "HierarchyViewTest", "add", "Parent Task")
	cli.MustExecute("-y", "HierarchyViewTest", "add", "Child Task 1", "-P", "Parent Task")
	cli.MustExecute("-y", "HierarchyViewTest", "add", "Child Task 2", "-P", "Parent Task")

	// List tasks with custom view
	stdout, _, exitCode := cli.Execute("-y", "HierarchyViewTest", "-v", "simple")

	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Parent Task")
	testutil.AssertContains(t, stdout, "Child Task 1")
	testutil.AssertContains(t, stdout, "Child Task 2")
	// Should contain box-drawing characters for hierarchy
	if !strings.Contains(stdout, "├") && !strings.Contains(stdout, "└") {
		t.Errorf("expected hierarchy preserved with tree characters in custom view, got:\n%s", stdout)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestInvalidViewError verifies that invalid view name shows helpful error message
func TestInvalidViewErrorViewsCLI(t *testing.T) {
	cli, _ := testutil.NewCLITestWithViews(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "InvalidViewTest")
	cli.MustExecute("-y", "InvalidViewTest", "add", "Test task")

	// Try to use a non-existent view
	stdout, stderr, exitCode := cli.Execute("-y", "InvalidViewTest", "-v", "nonexistent")

	// Should return error
	testutil.AssertExitCode(t, exitCode, 1)
	combinedOutput := stdout + stderr
	// Error should mention the view name or indicate it's not found
	if !strings.Contains(strings.ToLower(combinedOutput), "view") || !strings.Contains(strings.ToLower(combinedOutput), "nonexistent") {
		t.Errorf("expected error message about invalid view 'nonexistent', got:\n%s", combinedOutput)
	}
}

// =============================================================================
// View Create Filter Flags Tests (068-view-create-filter-flags)
// =============================================================================

// TestViewCreate_FilterStatus verifies --filter-status flag creates view with status filter
func TestViewCreate_FilterStatus(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViews(t)

	// Create a view with --filter-status flag
	stdout, stderr, exitCode := cli.Execute("-y", "view", "create", "statusview", "--filter-status", "TODO,IN-PROGRESS")

	testutil.AssertExitCode(t, exitCode, 0)
	if strings.Contains(stderr, "unknown flag") {
		t.Fatalf("--filter-status flag not recognized: %s", stderr)
	}

	// Verify the view YAML file was created with status filter
	viewPath := viewsDir + "/statusview.yaml"
	data, err := os.ReadFile(viewPath)
	if err != nil {
		t.Fatalf("expected view file to be created at %s: %v", viewPath, err)
	}

	viewContent := string(data)
	// Should have filters section with status filter
	if !strings.Contains(viewContent, "filters:") {
		t.Errorf("expected view to have filters section, got:\n%s", viewContent)
	}
	if !strings.Contains(viewContent, "status") {
		t.Errorf("expected view to have status filter, got:\n%s", viewContent)
	}
	// Should include the filter values
	if !strings.Contains(viewContent, "TODO") && !strings.Contains(viewContent, "IN-PROGRESS") {
		t.Errorf("expected view to have TODO or IN-PROGRESS status values, got:\n%s", viewContent)
	}

	// Verify the view output message
	testutil.AssertContains(t, stdout, "statusview")
}

// TestViewCreate_FilterPriority verifies --filter-priority flag creates view with priority filter
func TestViewCreate_FilterPriority(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViews(t)

	// Create a view with --filter-priority flag using keyword
	stdout, stderr, exitCode := cli.Execute("-y", "view", "create", "highprioview", "--filter-priority", "high")

	testutil.AssertExitCode(t, exitCode, 0)
	if strings.Contains(stderr, "unknown flag") {
		t.Fatalf("--filter-priority flag not recognized: %s", stderr)
	}

	// Verify the view YAML file was created with priority filter
	viewPath := viewsDir + "/highprioview.yaml"
	data, err := os.ReadFile(viewPath)
	if err != nil {
		t.Fatalf("expected view file to be created at %s: %v", viewPath, err)
	}

	viewContent := string(data)
	// Should have filters section with priority filter
	if !strings.Contains(viewContent, "filters:") {
		t.Errorf("expected view to have filters section, got:\n%s", viewContent)
	}
	if !strings.Contains(viewContent, "priority") {
		t.Errorf("expected view to have priority filter, got:\n%s", viewContent)
	}

	// Verify the view output message
	testutil.AssertContains(t, stdout, "highprioview")

	// Test with numeric range (e.g., "1-3")
	stdout2, stderr2, exitCode2 := cli.Execute("-y", "view", "create", "numprioview", "--filter-priority", "1-3")

	testutil.AssertExitCode(t, exitCode2, 0)
	if strings.Contains(stderr2, "unknown flag") {
		t.Fatalf("--filter-priority flag not recognized with numeric range: %s", stderr2)
	}

	viewPath2 := viewsDir + "/numprioview.yaml"
	data2, err := os.ReadFile(viewPath2)
	if err != nil {
		t.Fatalf("expected view file to be created at %s: %v", viewPath2, err)
	}

	viewContent2 := string(data2)
	if !strings.Contains(viewContent2, "priority") {
		t.Errorf("expected view to have priority filter, got:\n%s", viewContent2)
	}

	testutil.AssertContains(t, stdout2, "numprioview")
}

// TestViewCreate_CombinedFilters verifies multiple filter flags can be combined
func TestViewCreate_CombinedFilters(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViews(t)

	// Create a view with both --filter-status and --filter-priority flags
	stdout, stderr, exitCode := cli.Execute("-y", "view", "create", "combinedview",
		"--filter-status", "TODO",
		"--filter-priority", "high",
		"--fields", "status,summary,priority")

	testutil.AssertExitCode(t, exitCode, 0)
	if strings.Contains(stderr, "unknown flag") {
		t.Fatalf("filter flags not recognized: %s", stderr)
	}

	// Verify the view YAML file was created with both filters
	viewPath := viewsDir + "/combinedview.yaml"
	data, err := os.ReadFile(viewPath)
	if err != nil {
		t.Fatalf("expected view file to be created at %s: %v", viewPath, err)
	}

	viewContent := string(data)
	// Should have filters section with both filters
	if !strings.Contains(viewContent, "filters:") {
		t.Errorf("expected view to have filters section, got:\n%s", viewContent)
	}
	if !strings.Contains(viewContent, "status") {
		t.Errorf("expected view to have status filter, got:\n%s", viewContent)
	}
	if !strings.Contains(viewContent, "priority") {
		t.Errorf("expected view to have priority filter, got:\n%s", viewContent)
	}

	// Verify the view output message
	testutil.AssertContains(t, stdout, "combinedview")

	// Now test that the created view actually works for filtering tasks
	cli.MustExecute("-y", "list", "create", "CombinedFilterTest")
	cli.MustExecute("-y", "CombinedFilterTest", "add", "High priority TODO", "-p", "1")
	cli.MustExecute("-y", "CombinedFilterTest", "add", "Low priority TODO", "-p", "9")
	cli.MustExecute("-y", "CombinedFilterTest", "add", "High priority DONE", "-p", "1")
	cli.MustExecute("-y", "CombinedFilterTest", "complete", "High priority DONE")

	// List tasks with the combined view - should only show high priority TODO tasks
	listStdout, _, listExitCode := cli.Execute("-y", "CombinedFilterTest", "-v", "combinedview")
	testutil.AssertExitCode(t, listExitCode, 0)
	testutil.AssertContains(t, listStdout, "High priority TODO")
	testutil.AssertNotContains(t, listStdout, "Low priority TODO")
	testutil.AssertNotContains(t, listStdout, "High priority DONE")
}

// =============================================================================
// Issue 001: Command-line filters ignored when view is specified
// =============================================================================

// TestViewWithCLITagFilter verifies that --tag filter works in combination with -v all
func TestViewWithCLITagFilter(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add tasks with different tags
	cli.MustExecute("-y", "list", "create", "ViewTagFilterTest")
	cli.MustExecute("-y", "ViewTagFilterTest", "add", "Bug fix", "--tags", "urgent")
	cli.MustExecute("-y", "ViewTagFilterTest", "add", "Feature request", "--tags", "feature")

	// Filter by tag without view - works correctly
	stdout1, _, exitCode1 := cli.Execute("-y", "ViewTagFilterTest", "--tag", "urgent")
	testutil.AssertExitCode(t, exitCode1, 0)
	testutil.AssertContains(t, stdout1, "Bug fix")
	testutil.AssertNotContains(t, stdout1, "Feature request")

	// Filter by tag with -v all - should ALSO respect --tag filter
	stdout2, _, exitCode2 := cli.Execute("-y", "ViewTagFilterTest", "--tag", "urgent", "-v", "all")
	testutil.AssertExitCode(t, exitCode2, 0)
	testutil.AssertContains(t, stdout2, "Bug fix")
	testutil.AssertNotContains(t, stdout2, "Feature request") // This was failing before fix
}

// TestViewWithCLIStatusFilter verifies that -s filter works in combination with -v all
func TestViewWithCLIStatusFilter(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add tasks with different statuses
	cli.MustExecute("-y", "list", "create", "ViewStatusFilterTest")
	cli.MustExecute("-y", "ViewStatusFilterTest", "add", "Active task")
	cli.MustExecute("-y", "ViewStatusFilterTest", "add", "Completed task")
	cli.MustExecute("-y", "ViewStatusFilterTest", "complete", "Completed task")

	// Filter by status with -v all - should respect -s filter
	stdout, _, exitCode := cli.Execute("-y", "ViewStatusFilterTest", "-s", "TODO", "-v", "all")
	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Active task")
	testutil.AssertNotContains(t, stdout, "Completed task") // This was failing before fix
}

// TestViewWithCLIPriorityFilter verifies that -p filter works in combination with -v all
func TestViewWithCLIPriorityFilter(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add tasks with different priorities
	cli.MustExecute("-y", "list", "create", "ViewPrioFilterTest")
	cli.MustExecute("-y", "ViewPrioFilterTest", "add", "Urgent task", "-p", "1")
	cli.MustExecute("-y", "ViewPrioFilterTest", "add", "Low priority task", "-p", "9")

	// Filter by priority with -v all - should respect -p filter
	stdout, _, exitCode := cli.Execute("-y", "ViewPrioFilterTest", "-p", "1", "-v", "all")
	testutil.AssertExitCode(t, exitCode, 0)
	testutil.AssertContains(t, stdout, "Urgent task")
	testutil.AssertNotContains(t, stdout, "Low priority task") // This was failing before fix
}
