package views_test

import (
	"testing"

	"todoat/internal/testutil"
)

// =============================================================================
// Issue #19: Default view does not filter DONE tasks as documented
// =============================================================================

// TestDefaultViewFiltersDoneTasks verifies that the default view excludes DONE/completed tasks
func TestDefaultViewFiltersDoneTasks(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add tasks
	cli.MustExecute("-y", "list", "create", "DefaultFilterTest")
	cli.MustExecute("-y", "DefaultFilterTest", "add", "Active task")
	cli.MustExecute("-y", "DefaultFilterTest", "add", "Completed task")
	cli.MustExecute("-y", "DefaultFilterTest", "complete", "Completed task")

	// List tasks with default view (no -v flag)
	stdout, _, exitCode := cli.Execute("-y", "DefaultFilterTest")

	testutil.AssertExitCode(t, exitCode, 0)
	// Default view should show active tasks
	testutil.AssertContains(t, stdout, "Active task")
	// Default view should NOT show completed tasks
	testutil.AssertNotContains(t, stdout, "Completed task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestDefaultViewFiltersMultipleStatuses verifies correct filtering of various statuses
func TestDefaultViewFiltersMultipleStatuses(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add tasks with different statuses
	cli.MustExecute("-y", "list", "create", "MultiStatusFilterTest")
	cli.MustExecute("-y", "MultiStatusFilterTest", "add", "TODO task")
	cli.MustExecute("-y", "MultiStatusFilterTest", "add", "In progress task")
	cli.MustExecute("-y", "MultiStatusFilterTest", "update", "In progress task", "-s", "IN-PROGRESS")
	cli.MustExecute("-y", "MultiStatusFilterTest", "add", "Done task")
	cli.MustExecute("-y", "MultiStatusFilterTest", "complete", "Done task")

	// List tasks with default view
	stdout, _, exitCode := cli.Execute("-y", "MultiStatusFilterTest")

	testutil.AssertExitCode(t, exitCode, 0)
	// Should show TODO tasks
	testutil.AssertContains(t, stdout, "TODO task")
	// Should show IN-PROGRESS tasks
	testutil.AssertContains(t, stdout, "In progress task")
	// Should NOT show DONE/COMPLETED tasks
	testutil.AssertNotContains(t, stdout, "Done task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestAllViewShowsCompletedTasks verifies that 'all' view still shows completed tasks
func TestAllViewShowsCompletedTasks(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add tasks
	cli.MustExecute("-y", "list", "create", "AllViewTest2")
	cli.MustExecute("-y", "AllViewTest2", "add", "Active task")
	cli.MustExecute("-y", "AllViewTest2", "add", "Completed task")
	cli.MustExecute("-y", "AllViewTest2", "complete", "Completed task")

	// List tasks with 'all' view
	stdout, _, exitCode := cli.Execute("-y", "AllViewTest2", "-v", "all")

	testutil.AssertExitCode(t, exitCode, 0)
	// 'all' view should show both active and completed tasks
	testutil.AssertContains(t, stdout, "Active task")
	testutil.AssertContains(t, stdout, "Completed task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}
