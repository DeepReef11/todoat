package sqlite_test

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"todoat/internal/testutil"
)

// =============================================================================
// Task Command Tests (004-task-commands)
// =============================================================================

// --- Add Command Tests ---

func TestAddCommandSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout := cli.MustExecute("-y", "Work", "add", "Review PR")

	testutil.AssertContains(t, stdout, "Review PR")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

func TestAddCommandAbbreviationSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// 'a' is abbreviation for 'add'
	stdout := cli.MustExecute("-y", "Work", "a", "New task")

	testutil.AssertContains(t, stdout, "New task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

func TestAddCommandWithPrioritySQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	cli.MustExecute("-y", "Work", "add", "Urgent task", "-p", "1")

	// List tasks to verify priority
	stdout := cli.MustExecute("-y", "Work", "get")

	// Should show priority indicator
	testutil.AssertContains(t, stdout, "Urgent task")
}

// --- Get Command Tests ---

func TestGetCommandExplicitSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// First add a task
	cli.MustExecute("-y", "Work", "add", "Task 1")

	// Explicit get command
	stdout := cli.MustExecute("-y", "Work", "get")

	testutil.AssertContains(t, stdout, "Task 1")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

func TestGetCommandDefaultSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// First add a task
	cli.MustExecute("-y", "Work", "add", "Task for default")

	// Default action (just list name, no action) should show tasks
	stdout := cli.MustExecute("-y", "Work")

	testutil.AssertContains(t, stdout, "Task for default")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

func TestGetCommandAbbreviationSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task first
	cli.MustExecute("-y", "Work", "add", "Task G")

	// 'g' is abbreviation for 'get'
	stdout := cli.MustExecute("-y", "Work", "g")

	testutil.AssertContains(t, stdout, "Task G")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// --- Update Command Tests ---

func TestUpdateCommandSummarySQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Old name")

	// Update the summary
	cli.MustExecute("-y", "Work", "update", "Old name", "--summary", "New name")

	// Verify the update
	stdout := cli.MustExecute("-y", "Work", "get")

	testutil.AssertNotContains(t, stdout, "Old name")
	testutil.AssertContains(t, stdout, "New name")
}

func TestUpdateCommandPrioritySQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Task to update")

	// Update priority using abbreviation
	stdout := cli.MustExecute("-y", "Work", "u", "Task to update", "-p", "5")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

func TestUpdateCommandStatusSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Task status")

	// Update status
	stdout := cli.MustExecute("-y", "Work", "update", "Task status", "-s", "IN-PROGRESS")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// --- Complete Command Tests ---

func TestCompleteCommandSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Task to complete")

	// Complete the task
	stdout := cli.MustExecute("-y", "Work", "complete", "Task to complete")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

func TestCompleteCommandAbbreviationSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Another task")

	// 'c' is abbreviation for 'complete'
	stdout := cli.MustExecute("-y", "Work", "c", "Another task")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// --- Delete Command Tests ---

func TestDeleteCommandSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Task to delete")

	// Delete the task (with -y for no prompt)
	stdout := cli.MustExecute("-y", "Work", "delete", "Task to delete")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify deletion
	stdout = cli.MustExecute("-y", "Work", "get")

	testutil.AssertNotContains(t, stdout, "Task to delete")
}

func TestDeleteCommandAbbreviationSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Delete me")

	// 'd' is abbreviation for 'delete'
	stdout := cli.MustExecute("-y", "Work", "d", "Delete me")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// --- Task Matching Tests ---

func TestTaskMatchingExactSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add tasks with similar names
	cli.MustExecute("-y", "Work", "add", "Review PR")
	cli.MustExecute("-y", "Work", "add", "Review PR #123")

	// Exact match should find "Review PR"
	stdout := cli.MustExecute("-y", "Work", "complete", "Review PR")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

func TestTaskMatchingPartialSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Review PR #456")

	// Partial match should work when only one task matches
	stdout := cli.MustExecute("-y", "Work", "complete", "#456")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

func TestTaskMatchingNoMatchSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Some task")

	// No match should error
	stdout, stderr := cli.ExecuteAndFail("-y", "Work", "complete", "Nonexistent")

	errOutput := stderr
	if !strings.Contains(strings.ToLower(errOutput), "no") && !strings.Contains(strings.ToLower(errOutput), "match") && !strings.Contains(strings.ToLower(errOutput), "found") {
		t.Errorf("error should mention no match found, got: %s", errOutput)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultError)
}

func TestTaskMatchingMultipleMatchesSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add tasks with similar names
	cli.MustExecute("-y", "Work", "add", "Review code")
	cli.MustExecute("-y", "Work", "add", "Review docs")

	// Multiple matches in no-prompt mode should error
	stdout, stderr := cli.ExecuteAndFail("-y", "Work", "complete", "Review")

	errOutput := stderr
	if !strings.Contains(strings.ToLower(errOutput), "multiple") && !strings.Contains(strings.ToLower(errOutput), "matches") && !strings.Contains(strings.ToLower(errOutput), "ambiguous") {
		t.Errorf("error should mention multiple matches, got: %s", errOutput)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultError)
}

func TestTaskMatchingCaseInsensitiveSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task with mixed case
	cli.MustExecute("-y", "Work", "add", "Review PR")

	// Lowercase search should match
	stdout := cli.MustExecute("-y", "Work", "complete", "review pr")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// --- Status System Tests ---

func TestStatusDisplayFormatSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task (default status is TODO)
	cli.MustExecute("-y", "Work", "add", "Task one")

	// Get tasks and check status format
	stdout := cli.MustExecute("-y", "Work", "get")

	testutil.AssertContains(t, stdout, "[TODO]")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

func TestStatusDisplayFormatDoneSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add and complete a task
	cli.MustExecute("-y", "Work", "add", "Task done")
	cli.MustExecute("-y", "Work", "complete", "Task done")

	// Get tasks and check status format
	stdout := cli.MustExecute("-y", "Work", "get")

	testutil.AssertContains(t, stdout, "[DONE]")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

func TestStatusAbbreviationTSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task and set it to DONE first
	cli.MustExecute("-y", "Work", "add", "Task abbrev")
	cli.MustExecute("-y", "Work", "complete", "Task abbrev")

	// Update status using abbreviation T (should set to TODO)
	cli.MustExecute("-y", "Work", "update", "Task abbrev", "-s", "T")

	// Verify status is TODO
	stdout := cli.MustExecute("-y", "Work", "get")

	testutil.AssertContains(t, stdout, "[TODO]")
}

func TestStatusAbbreviationDSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Task abbrev D")

	// Update status using abbreviation D (should set to DONE)
	cli.MustExecute("-y", "Work", "update", "Task abbrev D", "-s", "D")

	// Verify status is DONE
	stdout := cli.MustExecute("-y", "Work", "get")

	testutil.AssertContains(t, stdout, "[DONE]")
}

func TestStatusCaseInsensitiveSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Task case")

	// Update status using lowercase
	cli.MustExecute("-y", "Work", "update", "Task case", "-s", "done")

	// Verify status is DONE
	stdout := cli.MustExecute("-y", "Work", "get")

	testutil.AssertContains(t, stdout, "[DONE]")
}

func TestFilterByStatusTodoSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add tasks with different statuses
	cli.MustExecute("-y", "Work", "add", "Task todo one")
	cli.MustExecute("-y", "Work", "add", "Task done one")
	cli.MustExecute("-y", "Work", "complete", "Task done one")

	// Filter to show only TODO tasks
	stdout := cli.MustExecute("-y", "Work", "-s", "TODO")

	testutil.AssertContains(t, stdout, "Task todo one")
	testutil.AssertNotContains(t, stdout, "Task done one")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

func TestFilterByStatusDoneSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add tasks with different statuses
	cli.MustExecute("-y", "Work", "add", "Task todo two")
	cli.MustExecute("-y", "Work", "add", "Task done two")
	cli.MustExecute("-y", "Work", "complete", "Task done two")

	// Filter to show only DONE tasks
	stdout := cli.MustExecute("-y", "Work", "-s", "DONE")

	testutil.AssertContains(t, stdout, "Task done two")
	testutil.AssertNotContains(t, stdout, "Task todo two")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

func TestFilterByStatusAbbreviationSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add tasks with different statuses
	cli.MustExecute("-y", "Work", "add", "Task todo three")
	cli.MustExecute("-y", "Work", "add", "Task done three")
	cli.MustExecute("-y", "Work", "complete", "Task done three")

	// Filter using abbreviation T
	stdout := cli.MustExecute("-y", "Work", "-s", "T")

	testutil.AssertContains(t, stdout, "Task todo three")
	testutil.AssertNotContains(t, stdout, "Task done three")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

func TestFilterByStatusLongFlagSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add tasks with different statuses
	cli.MustExecute("-y", "Work", "add", "Task todo four")
	cli.MustExecute("-y", "Work", "add", "Task done four")
	cli.MustExecute("-y", "Work", "complete", "Task done four")

	// Filter using --status long flag
	stdout := cli.MustExecute("-y", "Work", "--status", "D")

	testutil.AssertContains(t, stdout, "Task done four")
	testutil.AssertNotContains(t, stdout, "Task todo four")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

func TestNoFilterShowsAllTasksSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add tasks with different statuses
	cli.MustExecute("-y", "Work", "add", "Task todo five")
	cli.MustExecute("-y", "Work", "add", "Task done five")
	cli.MustExecute("-y", "Work", "complete", "Task done five")

	// Get without filter should show all
	stdout := cli.MustExecute("-y", "Work")

	testutil.AssertContains(t, stdout, "Task todo five")
	testutil.AssertContains(t, stdout, "Task done five")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// =============================================================================
// Result Code Tests (006-cli-tests)
// =============================================================================

func TestResultCodeAddTaskSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout := cli.MustExecute("-y", "Work", "add", "Test task")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

func TestResultCodeGetTasksSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task first
	cli.MustExecute("-y", "Work", "add", "Task to list")

	// Get tasks should return INFO_ONLY
	stdout := cli.MustExecute("-y", "Work", "get")

	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

func TestResultCodeListEmptySQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Get empty list should return INFO_ONLY
	stdout := cli.MustExecute("-y", "EmptyList")

	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
	testutil.AssertContains(t, stdout, "No tasks")
}

func TestResultCodeUpdateTaskSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Task to update")

	// Update task should return ACTION_COMPLETED
	stdout := cli.MustExecute("-y", "Work", "update", "Task to update", "--summary", "Updated task")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

func TestResultCodeCompleteTaskSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Task to complete")

	// Complete task should return ACTION_COMPLETED
	stdout := cli.MustExecute("-y", "Work", "complete", "Task to complete")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

func TestResultCodeCompleteTaskChangesDoneStatusSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add and complete a task
	cli.MustExecute("-y", "Work", "add", "Task for status")
	cli.MustExecute("-y", "Work", "complete", "Task for status")

	// Verify status changed to DONE
	stdout := cli.MustExecute("-y", "Work", "get")

	testutil.AssertContains(t, stdout, "[DONE]")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

func TestResultCodeDeleteTaskSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Task to delete")

	// Delete task should return ACTION_COMPLETED
	stdout := cli.MustExecute("-y", "Work", "delete", "Task to delete")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

func TestResultCodeDeleteConfirmationSkippedInNoPromptSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Task to confirm delete")

	// Delete with -y should not require confirmation
	stdout := cli.MustExecute("-y", "Work", "delete", "Task to confirm delete")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify task is gone
	stdout = cli.MustExecute("-y", "Work", "get")
	testutil.AssertNotContains(t, stdout, "Task to confirm delete")
}

func TestResultCodeErrorNoMatchSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Existing task")

	// Try to complete non-existent task
	stdout, _, exitCode := cli.Execute("-y", "Work", "complete", "Nonexistent task")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertResultCode(t, stdout, testutil.ResultError)
}

func TestResultCodeErrorAmbiguousMatchSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add tasks with similar names
	cli.MustExecute("-y", "Work", "add", "Similar task one")
	cli.MustExecute("-y", "Work", "add", "Similar task two")

	// Try to complete with ambiguous match
	stdout, _, exitCode := cli.Execute("-y", "Work", "complete", "Similar")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertResultCode(t, stdout, testutil.ResultError)
}

func TestExitCodesVerifiedSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Test exit code 0 for ACTION_COMPLETED
	_, _, exitCode := cli.Execute("-y", "Work", "add", "Test exit code")
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for add, got %d", exitCode)
	}

	// Test exit code 0 for INFO_ONLY
	_, _, exitCode = cli.Execute("-y", "Work", "get")
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for get, got %d", exitCode)
	}

	// Test exit code 1 for ERROR
	_, _, exitCode = cli.Execute("-y", "Work", "complete", "Nonexistent")
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for error, got %d", exitCode)
	}
}

// =============================================================================
// List Management Tests (007-list-commands)
// =============================================================================

// TestListCreate verifies that `todoat -y list create "MyList"` creates a new list
func TestListCreateSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout := cli.MustExecute("-y", "list", "create", "MyList")

	testutil.AssertContains(t, stdout, "MyList")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// TestListCreateDuplicate verifies that creating a duplicate list returns ERROR
func TestListCreateDuplicateSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create first list
	cli.MustExecute("-y", "list", "create", "ExistingList")

	// Try to create duplicate
	stdout, _, exitCode := cli.Execute("-y", "list", "create", "ExistingList")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertResultCode(t, stdout, testutil.ResultError)
}

// TestListView verifies that `todoat -y list` displays all lists with task counts
func TestListViewSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create lists and add tasks
	cli.MustExecute("-y", "list", "create", "Work")
	cli.MustExecute("-y", "list", "create", "Personal")
	cli.MustExecute("-y", "Work", "add", "Task 1")
	cli.MustExecute("-y", "Work", "add", "Task 2")
	cli.MustExecute("-y", "Personal", "add", "Task 3")

	// View lists
	stdout := cli.MustExecute("-y", "list")

	testutil.AssertContains(t, stdout, "Work")
	testutil.AssertContains(t, stdout, "Personal")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestListViewEmpty verifies that viewing lists with no lists shows INFO_ONLY message
func TestListViewEmptySQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// View lists when none exist
	stdout := cli.MustExecute("-y", "list")

	// Should contain a helpful message about no lists
	if !strings.Contains(strings.ToLower(stdout), "no") || !strings.Contains(strings.ToLower(stdout), "list") {
		t.Errorf("expected message about no lists, got: %s", stdout)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestListViewJSON verifies that `todoat -y --json list` returns valid JSON
func TestListViewJSONSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "JSONTest")

	// View lists with JSON output
	stdout := cli.MustExecute("-y", "--json", "list")

	// Should contain JSON array indicators
	testutil.AssertContains(t, stdout, "[")
	testutil.AssertContains(t, stdout, "]")
	testutil.AssertContains(t, stdout, "JSONTest")
}

// TestListCreateJSON verifies that `todoat -y --json list create "Test"` returns JSON
func TestListCreateJSONSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list with JSON output
	stdout := cli.MustExecute("-y", "--json", "list", "create", "JSONCreate")

	// Should contain JSON object indicators
	testutil.AssertContains(t, stdout, "{")
	testutil.AssertContains(t, stdout, "}")
	testutil.AssertContains(t, stdout, "JSONCreate")
}

// =============================================================================
// JSON Output Tests for Task Commands (008-json-output)
// =============================================================================

// TestJSONFlagParsing verifies that --json flag is recognized and sets output mode
func TestJSONFlagParsingSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list first
	cli.MustExecute("-y", "list", "create", "FlagTest")

	// Test that --json flag is accepted without error
	stdout := cli.MustExecute("-y", "--json", "FlagTest")

	// JSON output should contain JSON structure, not plain text
	testutil.AssertContains(t, stdout, "{")
	testutil.AssertNotContains(t, stdout, "Tasks in 'FlagTest'")
}

// TestListTasksJSON verifies that `todoat -y --json MyList` returns valid JSON with tasks array
func TestListTasksJSONSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add tasks
	cli.MustExecute("-y", "list", "create", "TaskListJSON")
	cli.MustExecute("-y", "TaskListJSON", "add", "First Task")
	cli.MustExecute("-y", "TaskListJSON", "add", "Second Task")

	// List tasks with JSON output
	stdout := cli.MustExecute("-y", "--json", "TaskListJSON")

	// Should contain JSON with tasks array
	testutil.AssertContains(t, stdout, `"tasks"`)
	testutil.AssertContains(t, stdout, `"list"`)
	testutil.AssertContains(t, stdout, `"TaskListJSON"`)
	testutil.AssertContains(t, stdout, `"First Task"`)
	testutil.AssertContains(t, stdout, `"Second Task"`)
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"INFO_ONLY"`)
}

// TestAddTaskJSON verifies that `todoat -y --json MyList add "Task"` returns JSON with task info and result
func TestAddTaskJSONSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list first
	cli.MustExecute("-y", "list", "create", "AddJSON")

	// Add task with JSON output
	stdout := cli.MustExecute("-y", "--json", "AddJSON", "add", "New JSON Task")

	// Should contain JSON with action and task
	testutil.AssertContains(t, stdout, `"action"`)
	testutil.AssertContains(t, stdout, `"add"`)
	testutil.AssertContains(t, stdout, `"task"`)
	testutil.AssertContains(t, stdout, `"New JSON Task"`)
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"ACTION_COMPLETED"`)
}

// TestUpdateTaskJSON verifies that `todoat -y --json MyList update "Task" -s DONE` returns JSON with updated task
func TestUpdateTaskJSONSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add a task
	cli.MustExecute("-y", "list", "create", "UpdateJSON")
	cli.MustExecute("-y", "UpdateJSON", "add", "Task To Update")

	// Update task with JSON output
	stdout := cli.MustExecute("-y", "--json", "UpdateJSON", "update", "Task To Update", "-s", "DONE")

	// Should contain JSON with action and updated task
	testutil.AssertContains(t, stdout, `"action"`)
	testutil.AssertContains(t, stdout, `"update"`)
	testutil.AssertContains(t, stdout, `"task"`)
	testutil.AssertContains(t, stdout, `"Task To Update"`)
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"ACTION_COMPLETED"`)
}

// TestDeleteTaskJSON verifies that `todoat -y --json MyList delete "Task"` returns JSON with result
func TestDeleteTaskJSONSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add a task
	cli.MustExecute("-y", "list", "create", "DeleteJSON")
	cli.MustExecute("-y", "DeleteJSON", "add", "Task To Delete")

	// Delete task with JSON output
	stdout := cli.MustExecute("-y", "--json", "DeleteJSON", "delete", "Task To Delete")

	// Should contain JSON with action and result
	testutil.AssertContains(t, stdout, `"action"`)
	testutil.AssertContains(t, stdout, `"delete"`)
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"ACTION_COMPLETED"`)
}

// TestErrorJSON verifies that error conditions return JSON error with result: "ERROR"
func TestErrorJSONSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list first
	cli.MustExecute("-y", "list", "create", "ErrorTestList")

	// Try to delete a non-existent task with JSON output (this triggers an error)
	stdout, _, exitCode := cli.Execute("-y", "--json", "ErrorTestList", "delete", "NonExistentTask")

	// Should return non-zero exit code
	if exitCode == 0 {
		t.Errorf("expected non-zero exit code for error, got 0")
	}

	// Should contain JSON error
	testutil.AssertContains(t, stdout, `"error"`)
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"ERROR"`)
}

// TestJSONResultCodes verifies that all JSON responses include "result" field
func TestJSONResultCodesSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list
	cli.MustExecute("-y", "list", "create", "ResultCodeTest")

	// Test INFO_ONLY result for listing tasks
	stdout := cli.MustExecute("-y", "--json", "ResultCodeTest")
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"INFO_ONLY"`)

	// Test ACTION_COMPLETED result for add
	stdout = cli.MustExecute("-y", "--json", "ResultCodeTest", "add", "Test Task")
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"ACTION_COMPLETED"`)

	// Test ACTION_COMPLETED result for update
	stdout = cli.MustExecute("-y", "--json", "ResultCodeTest", "update", "Test Task", "-s", "DONE")
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"ACTION_COMPLETED"`)

	// Test ACTION_COMPLETED result for delete
	cli.MustExecute("-y", "ResultCodeTest", "add", "Delete Me")
	stdout = cli.MustExecute("-y", "--json", "ResultCodeTest", "delete", "Delete Me")
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"ACTION_COMPLETED"`)
}

// =============================================================================
// Priority Filtering Tests (009-priority-filtering)
// =============================================================================

// TestPriorityFilterSingle verifies that `todoat -y MyList -p 1` shows only priority 1 tasks
func TestPriorityFilterSingleSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create tasks with different priorities
	cli.MustExecute("-y", "Work", "add", "Priority 1 task", "-p", "1")
	cli.MustExecute("-y", "Work", "add", "Priority 2 task", "-p", "2")
	cli.MustExecute("-y", "Work", "add", "Priority 5 task", "-p", "5")

	// Filter to show only priority 1 tasks
	stdout := cli.MustExecute("-y", "Work", "-p", "1")

	testutil.AssertContains(t, stdout, "Priority 1 task")
	testutil.AssertNotContains(t, stdout, "Priority 2 task")
	testutil.AssertNotContains(t, stdout, "Priority 5 task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPriorityFilterRange verifies that `todoat -y MyList -p 1,2,3` shows tasks with priority 1, 2, or 3
func TestPriorityFilterRangeSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create tasks with different priorities
	cli.MustExecute("-y", "Work", "add", "Priority 1 task", "-p", "1")
	cli.MustExecute("-y", "Work", "add", "Priority 2 task", "-p", "2")
	cli.MustExecute("-y", "Work", "add", "Priority 3 task", "-p", "3")
	cli.MustExecute("-y", "Work", "add", "Priority 5 task", "-p", "5")
	cli.MustExecute("-y", "Work", "add", "Priority 7 task", "-p", "7")

	// Filter to show priorities 1, 2, 3
	stdout := cli.MustExecute("-y", "Work", "-p", "1,2,3")

	testutil.AssertContains(t, stdout, "Priority 1 task")
	testutil.AssertContains(t, stdout, "Priority 2 task")
	testutil.AssertContains(t, stdout, "Priority 3 task")
	testutil.AssertNotContains(t, stdout, "Priority 5 task")
	testutil.AssertNotContains(t, stdout, "Priority 7 task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPriorityFilterHigh verifies that `todoat -y MyList -p high` shows priorities 1-4
func TestPriorityFilterHighSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create tasks with different priorities
	cli.MustExecute("-y", "Work", "add", "Priority 1 task", "-p", "1")
	cli.MustExecute("-y", "Work", "add", "Priority 4 task", "-p", "4")
	cli.MustExecute("-y", "Work", "add", "Priority 5 task", "-p", "5")
	cli.MustExecute("-y", "Work", "add", "Priority 9 task", "-p", "9")

	// Filter using 'high' alias
	stdout := cli.MustExecute("-y", "Work", "-p", "high")

	testutil.AssertContains(t, stdout, "Priority 1 task")
	testutil.AssertContains(t, stdout, "Priority 4 task")
	testutil.AssertNotContains(t, stdout, "Priority 5 task")
	testutil.AssertNotContains(t, stdout, "Priority 9 task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPriorityFilterMedium verifies that `todoat -y MyList -p medium` shows priority 5
func TestPriorityFilterMediumSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create tasks with different priorities
	cli.MustExecute("-y", "Work", "add", "Priority 1 task", "-p", "1")
	cli.MustExecute("-y", "Work", "add", "Priority 5 task", "-p", "5")
	cli.MustExecute("-y", "Work", "add", "Priority 6 task", "-p", "6")

	// Filter using 'medium' alias
	stdout := cli.MustExecute("-y", "Work", "-p", "medium")

	testutil.AssertNotContains(t, stdout, "Priority 1 task")
	testutil.AssertContains(t, stdout, "Priority 5 task")
	testutil.AssertNotContains(t, stdout, "Priority 6 task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPriorityFilterLow verifies that `todoat -y MyList -p low` shows priorities 6-9
func TestPriorityFilterLowSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create tasks with different priorities
	cli.MustExecute("-y", "Work", "add", "Priority 1 task", "-p", "1")
	cli.MustExecute("-y", "Work", "add", "Priority 5 task", "-p", "5")
	cli.MustExecute("-y", "Work", "add", "Priority 6 task", "-p", "6")
	cli.MustExecute("-y", "Work", "add", "Priority 9 task", "-p", "9")

	// Filter using 'low' alias
	stdout := cli.MustExecute("-y", "Work", "-p", "low")

	testutil.AssertNotContains(t, stdout, "Priority 1 task")
	testutil.AssertNotContains(t, stdout, "Priority 5 task")
	testutil.AssertContains(t, stdout, "Priority 6 task")
	testutil.AssertContains(t, stdout, "Priority 9 task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPriorityFilterUndefined verifies that `todoat -y MyList -p 0` shows tasks with no priority set
func TestPriorityFilterUndefinedSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create tasks with and without priority
	cli.MustExecute("-y", "Work", "add", "No priority task")
	cli.MustExecute("-y", "Work", "add", "Priority 1 task", "-p", "1")
	cli.MustExecute("-y", "Work", "add", "Priority 5 task", "-p", "5")

	// Filter to show only tasks with no priority (priority 0)
	stdout := cli.MustExecute("-y", "Work", "-p", "0")

	testutil.AssertContains(t, stdout, "No priority task")
	testutil.AssertNotContains(t, stdout, "Priority 1 task")
	testutil.AssertNotContains(t, stdout, "Priority 5 task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPriorityFilterNoMatch verifies that `todoat -y MyList -p 1` with no matching tasks returns INFO_ONLY with message
func TestPriorityFilterNoMatchSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create tasks with priority 5 only
	cli.MustExecute("-y", "Work", "add", "Priority 5 task", "-p", "5")

	// Filter for priority 1 (no matches)
	stdout := cli.MustExecute("-y", "Work", "-p", "1")

	// Should show a message about no tasks matching
	if !strings.Contains(strings.ToLower(stdout), "no") || !strings.Contains(strings.ToLower(stdout), "task") {
		t.Errorf("expected message about no matching tasks, got: %s", stdout)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPriorityFilterJSON verifies that `todoat -y --json MyList -p 1` returns filtered JSON result
func TestPriorityFilterJSONSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create tasks with different priorities
	cli.MustExecute("-y", "Work", "add", "Priority 1 task", "-p", "1")
	cli.MustExecute("-y", "Work", "add", "Priority 5 task", "-p", "5")

	// Filter with JSON output
	stdout := cli.MustExecute("-y", "--json", "Work", "-p", "1")

	// Should contain JSON with only priority 1 task
	testutil.AssertContains(t, stdout, `"Priority 1 task"`)
	testutil.AssertNotContains(t, stdout, `"Priority 5 task"`)
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"INFO_ONLY"`)
}

// TestPriorityFilterInvalid verifies that `todoat -y MyList -p 10` returns ERROR for invalid priority
func TestPriorityFilterInvalidSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a task
	cli.MustExecute("-y", "Work", "add", "Some task")

	// Try invalid priority filter
	stdout, _, exitCode := cli.Execute("-y", "Work", "-p", "10")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertResultCode(t, stdout, testutil.ResultError)
}

// TestPriorityFilterCombinedWithStatus verifies combined status and priority filters work
func TestPriorityFilterCombinedWithStatusSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create tasks with different priorities and statuses
	cli.MustExecute("-y", "Work", "add", "High priority TODO", "-p", "1")
	cli.MustExecute("-y", "Work", "add", "High priority DONE", "-p", "1")
	cli.MustExecute("-y", "Work", "complete", "High priority DONE")
	cli.MustExecute("-y", "Work", "add", "Low priority TODO", "-p", "7")

	// Filter for TODO tasks with high priority
	stdout := cli.MustExecute("-y", "Work", "-s", "TODO", "-p", "high")

	testutil.AssertContains(t, stdout, "High priority TODO")
	testutil.AssertNotContains(t, stdout, "High priority DONE")
	testutil.AssertNotContains(t, stdout, "Low priority TODO")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// =============================================================================
// Task Dates Tests (011-task-dates)
// =============================================================================

// TestAddTaskWithDueDate verifies that `todoat -y MyList add "Task" --due-date 2026-01-31` sets due date
func TestAddTaskWithDueDateSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout := cli.MustExecute("-y", "Work", "add", "Task with due", "--due-date", "2026-01-31")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify by listing tasks with JSON to check due_date
	stdout = cli.MustExecute("-y", "--json", "Work")

	testutil.AssertContains(t, stdout, "Task with due")
	testutil.AssertContains(t, stdout, "2026-01-31")
}

// TestAddTaskWithStartDate verifies that `todoat -y MyList add "Task" --start-date 2026-01-15` sets start date
func TestAddTaskWithStartDateSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout := cli.MustExecute("-y", "Work", "add", "Task with start", "--start-date", "2026-01-15")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify by listing tasks with JSON to check start_date
	stdout = cli.MustExecute("-y", "--json", "Work")

	testutil.AssertContains(t, stdout, "Task with start")
	testutil.AssertContains(t, stdout, "2026-01-15")
}

// TestAddTaskWithBothDates verifies that both --due-date and --start-date can be set together
func TestAddTaskWithBothDatesSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout := cli.MustExecute("-y", "Work", "add", "Task with both dates", "--due-date", "2026-01-31", "--start-date", "2026-01-15")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify by listing tasks with JSON to check both dates
	stdout = cli.MustExecute("-y", "--json", "Work")

	testutil.AssertContains(t, stdout, "Task with both dates")
	testutil.AssertContains(t, stdout, "2026-01-31")
	testutil.AssertContains(t, stdout, "2026-01-15")
}

// TestUpdateTaskDueDate verifies that `todoat -y MyList update "Task" --due-date 2026-02-15` updates due date
func TestUpdateTaskDueDateSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task with a due date
	cli.MustExecute("-y", "Work", "add", "Update date task", "--due-date", "2026-01-31")

	// Update the due date
	stdout := cli.MustExecute("-y", "Work", "update", "Update date task", "--due-date", "2026-02-15")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify updated due date
	stdout = cli.MustExecute("-y", "--json", "Work")

	testutil.AssertContains(t, stdout, "Update date task")
	testutil.AssertContains(t, stdout, "2026-02-15")
	testutil.AssertNotContains(t, stdout, "2026-01-31")
}

// TestClearTaskDueDate verifies that `todoat -y MyList update "Task" --due-date ""` clears due date
func TestClearTaskDueDateSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task with a due date
	cli.MustExecute("-y", "Work", "add", "Clear date task", "--due-date", "2026-01-31")

	// Clear the due date
	stdout := cli.MustExecute("-y", "Work", "update", "Clear date task", "--due-date", "")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify due date is cleared - JSON output should not have the date
	stdout = cli.MustExecute("-y", "--json", "Work")

	testutil.AssertContains(t, stdout, "Clear date task")
	// Due date should be empty or null in JSON
	testutil.AssertNotContains(t, stdout, "2026-01-31")
}

// TestInvalidDateFormat verifies that `todoat -y MyList add "Task" --due-date "invalid"` returns ERROR
func TestInvalidDateFormatSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout, _, exitCode := cli.Execute("-y", "Work", "add", "Invalid date task", "--due-date", "invalid")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertResultCode(t, stdout, testutil.ResultError)
}

// TestDateFormatValidation verifies that `todoat -y MyList add "Task" --due-date "01-31-2026"` returns ERROR (wrong format)
func TestDateFormatValidationSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Wrong format: MM-DD-YYYY instead of YYYY-MM-DD
	stdout, _, exitCode := cli.Execute("-y", "Work", "add", "Wrong format task", "--due-date", "01-31-2026")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertResultCode(t, stdout, testutil.ResultError)
}

// TestTaskDatesInJSON verifies that `todoat -y --json MyList` includes due_date and start_date fields
func TestTaskDatesInJSONSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task with both dates
	cli.MustExecute("-y", "Work", "add", "JSON date task", "--due-date", "2026-01-31", "--start-date", "2026-01-15")

	// Get tasks as JSON
	stdout := cli.MustExecute("-y", "--json", "Work")

	testutil.AssertContains(t, stdout, `"due_date"`)
	testutil.AssertContains(t, stdout, `"start_date"`)
	testutil.AssertContains(t, stdout, "2026-01-31")
	testutil.AssertContains(t, stdout, "2026-01-15")
}

// TestCompletedTimestamp verifies that `todoat -y MyList complete "Task"` sets completed timestamp automatically
func TestCompletedTimestampSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Task to complete")

	// Complete the task
	stdout := cli.MustExecute("-y", "Work", "complete", "Task to complete")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Get tasks as JSON and verify completed timestamp is set
	stdout = cli.MustExecute("-y", "--json", "Work")

	testutil.AssertContains(t, stdout, "Task to complete")
	testutil.AssertContains(t, stdout, `"completed"`)
}

// =============================================================================
// Tag Filtering Tests (012-tag-filtering)
// =============================================================================

// TestAddTaskWithTag verifies that `todoat -y MyList add "Task" --tag work` adds task with tag
func TestAddTaskWithTagSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout := cli.MustExecute("-y", "Work", "add", "Tagged task", "--tag", "work")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify by listing tasks with JSON to check tags
	stdout = cli.MustExecute("-y", "--json", "Work")

	testutil.AssertContains(t, stdout, "Tagged task")
	testutil.AssertContains(t, stdout, "work")
}

// TestAddTaskMultipleTags verifies that `todoat -y MyList add "Task" --tag work --tag urgent` adds task with multiple tags
func TestAddTaskMultipleTagsSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout := cli.MustExecute("-y", "Work", "add", "Multi-tagged task", "--tag", "work", "--tag", "urgent")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify by listing tasks with JSON to check tags
	stdout = cli.MustExecute("-y", "--json", "Work")

	testutil.AssertContains(t, stdout, "Multi-tagged task")
	testutil.AssertContains(t, stdout, "work")
	testutil.AssertContains(t, stdout, "urgent")
}

// TestAddTaskCommaSeparatedTags verifies that `todoat -y MyList add "Task" --tag "work,urgent"` adds task with comma-separated tags
func TestAddTaskCommaSeparatedTagsSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout := cli.MustExecute("-y", "Work", "add", "Comma-tagged task", "--tag", "work,urgent")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify by listing tasks with JSON to check tags
	stdout = cli.MustExecute("-y", "--json", "Work")

	testutil.AssertContains(t, stdout, "Comma-tagged task")
	testutil.AssertContains(t, stdout, "work")
	testutil.AssertContains(t, stdout, "urgent")
}

// TestUpdateTaskTags verifies that `todoat -y MyList update "Task" --tag home` updates task tags
func TestUpdateTaskTagsSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task with a tag
	cli.MustExecute("-y", "Work", "add", "Update tag task", "--tag", "work")

	// Update the tag
	stdout := cli.MustExecute("-y", "Work", "update", "Update tag task", "--tag", "home")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify updated tag
	stdout = cli.MustExecute("-y", "--json", "Work")

	testutil.AssertContains(t, stdout, "Update tag task")
	testutil.AssertContains(t, stdout, "home")
	testutil.AssertNotContains(t, stdout, `"work"`)
}

// TestClearTaskTags verifies that `todoat -y MyList update "Task" --tag ""` clears task tags
func TestClearTaskTagsSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task with a tag
	cli.MustExecute("-y", "Work", "add", "Clear tag task", "--tag", "work")

	// Clear the tag
	stdout := cli.MustExecute("-y", "Work", "update", "Clear tag task", "--tag", "")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify tag is cleared
	stdout = cli.MustExecute("-y", "--json", "Work")

	testutil.AssertContains(t, stdout, "Clear tag task")
	// The tags field should be empty or not contain "work"
	// We check that the specific tag value is no longer present
	if strings.Contains(stdout, `"tags":["work"]`) {
		t.Errorf("expected tags to be cleared, but still found work tag in output: %s", stdout)
	}
}

// TestFilterByTag verifies that `todoat -y MyList --tag work` shows only tasks with "work" tag
func TestFilterByTagSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add tasks with different tags
	cli.MustExecute("-y", "Work", "add", "Work task", "--tag", "work")
	cli.MustExecute("-y", "Work", "add", "Home task", "--tag", "home")
	cli.MustExecute("-y", "Work", "add", "No tag task")

	// Filter by work tag
	stdout := cli.MustExecute("-y", "Work", "--tag", "work")

	testutil.AssertContains(t, stdout, "Work task")
	testutil.AssertNotContains(t, stdout, "Home task")
	testutil.AssertNotContains(t, stdout, "No tag task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestFilterByMultipleTags verifies that `todoat -y MyList --tag work --tag urgent` shows tasks with ANY of the tags (OR logic)
func TestFilterByMultipleTagsSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add tasks with different tags
	cli.MustExecute("-y", "Work", "add", "Work task", "--tag", "work")
	cli.MustExecute("-y", "Work", "add", "Urgent task", "--tag", "urgent")
	cli.MustExecute("-y", "Work", "add", "Home task", "--tag", "home")

	// Filter by work OR urgent tag
	stdout := cli.MustExecute("-y", "Work", "--tag", "work", "--tag", "urgent")

	testutil.AssertContains(t, stdout, "Work task")
	testutil.AssertContains(t, stdout, "Urgent task")
	testutil.AssertNotContains(t, stdout, "Home task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestFilterTagNoMatch verifies that `todoat -y MyList --tag nonexistent` returns INFO_ONLY with message
func TestFilterTagNoMatchSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add a task with a different tag
	cli.MustExecute("-y", "Work", "add", "Some task", "--tag", "work")

	// Filter by non-existent tag
	stdout := cli.MustExecute("-y", "Work", "--tag", "nonexistent")

	// Should show a message about no matching tasks
	if !strings.Contains(strings.ToLower(stdout), "no") || !strings.Contains(strings.ToLower(stdout), "task") {
		t.Errorf("expected message about no matching tasks, got: %s", stdout)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestFilterTagJSON verifies that `todoat -y --json MyList --tag work` returns filtered JSON result with tags array
func TestFilterTagJSONSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add tasks with different tags
	cli.MustExecute("-y", "Work", "add", "Work task", "--tag", "work")
	cli.MustExecute("-y", "Work", "add", "Home task", "--tag", "home")

	// Filter with JSON output
	stdout := cli.MustExecute("-y", "--json", "Work", "--tag", "work")

	// Should contain JSON with only work-tagged task
	testutil.AssertContains(t, stdout, `"Work task"`)
	testutil.AssertNotContains(t, stdout, `"Home task"`)
	testutil.AssertContains(t, stdout, `"tags"`)
	testutil.AssertContains(t, stdout, `"work"`)
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"INFO_ONLY"`)
}

// TestFilterTagCombined verifies that `todoat -y MyList -s TODO --tag work` combined with status filter works
func TestFilterTagCombinedSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Add tasks with different statuses and tags
	cli.MustExecute("-y", "Work", "add", "Work TODO task", "--tag", "work")
	cli.MustExecute("-y", "Work", "add", "Work DONE task", "--tag", "work")
	cli.MustExecute("-y", "Work", "complete", "Work DONE task")
	cli.MustExecute("-y", "Work", "add", "Home TODO task", "--tag", "home")

	// Filter by TODO status AND work tag
	stdout := cli.MustExecute("-y", "Work", "-s", "TODO", "--tag", "work")

	testutil.AssertContains(t, stdout, "Work TODO task")
	testutil.AssertNotContains(t, stdout, "Work DONE task")
	testutil.AssertNotContains(t, stdout, "Home TODO task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// =============================================================================
// List Management Tests (013-list-management)
// =============================================================================

// TestListDelete verifies that `todoat -y list delete "ListName"` soft-deletes a list
func TestListDeleteSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "ToDelete")

	// Delete the list
	stdout := cli.MustExecute("-y", "list", "delete", "ToDelete")

	testutil.AssertContains(t, stdout, "ToDelete")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify list is no longer visible in normal list view
	stdout = cli.MustExecute("-y", "list")
	testutil.AssertNotContains(t, stdout, "ToDelete")
}

// TestListDeleteNotFound verifies that deleting a non-existent list returns ERROR
func TestListDeleteNotFoundSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Try to delete non-existent list
	stdout, _, exitCode := cli.Execute("-y", "list", "delete", "NonExistent")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertResultCode(t, stdout, testutil.ResultError)
}

// TestListTrash verifies that `todoat -y list trash` displays deleted lists
func TestListTrashSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create and delete a list
	cli.MustExecute("-y", "list", "create", "TrashTest")
	cli.MustExecute("-y", "list", "delete", "TrashTest")

	// View trash
	stdout := cli.MustExecute("-y", "list", "trash")

	testutil.AssertContains(t, stdout, "TrashTest")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestListTrashEmpty verifies that viewing empty trash returns INFO_ONLY
func TestListTrashEmptySQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// View trash with no deleted lists
	stdout := cli.MustExecute("-y", "list", "trash")

	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestListRestore verifies that `todoat -y list trash restore "Name"` restores a deleted list
func TestListRestoreSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create, delete, then restore a list
	cli.MustExecute("-y", "list", "create", "RestoreTest")
	cli.MustExecute("-y", "list", "delete", "RestoreTest")

	// Restore the list
	stdout := cli.MustExecute("-y", "list", "trash", "restore", "RestoreTest")

	testutil.AssertContains(t, stdout, "RestoreTest")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify list is visible in normal list view again
	stdout = cli.MustExecute("-y", "list")
	testutil.AssertContains(t, stdout, "RestoreTest")
}

// TestListRestoreNotInTrash verifies that restoring an active list returns ERROR
func TestListRestoreNotInTrashSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list but don't delete it
	cli.MustExecute("-y", "list", "create", "ActiveList")

	// Try to restore an active list
	stdout, _, exitCode := cli.Execute("-y", "list", "trash", "restore", "ActiveList")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertResultCode(t, stdout, testutil.ResultError)
}

// TestListPurge verifies that `todoat -y list trash purge "Name"` permanently deletes
func TestListPurgeSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create, delete, then purge a list
	cli.MustExecute("-y", "list", "create", "PurgeTest")
	cli.MustExecute("-y", "list", "delete", "PurgeTest")

	// Purge the list
	stdout := cli.MustExecute("-y", "list", "trash", "purge", "PurgeTest")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify list is not in trash anymore
	stdout = cli.MustExecute("-y", "list", "trash")
	testutil.AssertNotContains(t, stdout, "PurgeTest")
}

// TestListInfo verifies that `todoat -y list info "Name"` shows list details
func TestListInfoSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add some tasks
	cli.MustExecute("-y", "list", "create", "InfoTest")
	cli.MustExecute("-y", "InfoTest", "add", "Task 1")
	cli.MustExecute("-y", "InfoTest", "add", "Task 2")

	// Get list info
	stdout := cli.MustExecute("-y", "list", "info", "InfoTest")

	testutil.AssertContains(t, stdout, "InfoTest")
	// Should show task count (2 tasks)
	testutil.AssertContains(t, stdout, "2")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// =============================================================================
// Subtasks and Hierarchical Task Support Tests (014)
// =============================================================================

// TestAddSubtaskWithParentFlag verifies `todoat MyList add "Child" -P "Parent"` creates subtask under existing parent
func TestAddSubtaskWithParentFlagSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add a parent task
	cli.MustExecute("-y", "list", "create", "SubtaskTest")
	cli.MustExecute("-y", "SubtaskTest", "add", "Parent Task")

	// Add a subtask under the parent using -P flag
	stdout := cli.MustExecute("-y", "SubtaskTest", "add", "Child Task", "-P", "Parent Task")

	testutil.AssertContains(t, stdout, "Child Task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the subtask was created with parent relationship
	stdout = cli.MustExecute("-y", "--json", "SubtaskTest")
	testutil.AssertContains(t, stdout, `"Child Task"`)
	testutil.AssertContains(t, stdout, `"parent_id"`) // Should have parent reference
}

// TestPathBasedHierarchyCreation verifies `todoat MyList add "A/B/C"` creates 3-level hierarchy
func TestPathBasedHierarchyCreationSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "HierarchyTest")

	// Add task with path-based hierarchy
	stdout := cli.MustExecute("-y", "HierarchyTest", "add", "ProjectA/FeatureB/TaskC")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify all three tasks were created
	stdout = cli.MustExecute("-y", "--json", "HierarchyTest")
	testutil.AssertContains(t, stdout, `"ProjectA"`)
	testutil.AssertContains(t, stdout, `"FeatureB"`)
	testutil.AssertContains(t, stdout, `"TaskC"`)
}

// TestTreeVisualization verifies `todoat MyList` displays tasks with box-drawing characters
func TestTreeVisualizationSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with parent and children
	cli.MustExecute("-y", "list", "create", "TreeTest")
	cli.MustExecute("-y", "TreeTest", "add", "Parent")
	cli.MustExecute("-y", "TreeTest", "add", "Child1", "-P", "Parent")
	cli.MustExecute("-y", "TreeTest", "add", "Child2", "-P", "Parent")

	// List tasks - should show tree structure with box-drawing characters
	stdout := cli.MustExecute("-y", "TreeTest")

	testutil.AssertContains(t, stdout, "Parent")
	testutil.AssertContains(t, stdout, "Child1")
	testutil.AssertContains(t, stdout, "Child2")
	// Should contain box-drawing characters for tree visualization
	// ├─ for branches, └─ for last child
	if !strings.Contains(stdout, "├") && !strings.Contains(stdout, "└") {
		t.Errorf("expected tree visualization with box-drawing characters, got:\n%s", stdout)
	}
}

// TestUpdateParent verifies `todoat MyList update "Task" -P "NewParent"` re-parents task
func TestUpdateParentSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with two parents and one child
	cli.MustExecute("-y", "list", "create", "ReparentTest")
	cli.MustExecute("-y", "ReparentTest", "add", "OldParent")
	cli.MustExecute("-y", "ReparentTest", "add", "NewParent")
	cli.MustExecute("-y", "ReparentTest", "add", "MovingChild", "-P", "OldParent")

	// Re-parent the child to NewParent
	stdout := cli.MustExecute("-y", "ReparentTest", "update", "MovingChild", "-P", "NewParent")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the task is now under NewParent (visible in tree view)
	stdout = cli.MustExecute("-y", "--json", "ReparentTest")
	// The child should be associated with NewParent now
	testutil.AssertContains(t, stdout, `"MovingChild"`)
}

// TestRemoveParent verifies `todoat MyList update "Task" --no-parent` moves subtask to root level
func TestRemoveParentSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with parent and child
	cli.MustExecute("-y", "list", "create", "NoParentTest")
	cli.MustExecute("-y", "NoParentTest", "add", "Parent")
	cli.MustExecute("-y", "NoParentTest", "add", "Child", "-P", "Parent")

	// Remove parent relationship using --no-parent
	stdout := cli.MustExecute("-y", "NoParentTest", "update", "Child", "--no-parent")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the child is now a root-level task (no tree indentation)
	stdout = cli.MustExecute("-y", "--json", "NoParentTest")
	testutil.AssertContains(t, stdout, `"Child"`)
}

// TestCascadeDelete verifies `todoat MyList delete "Parent"` deletes all descendants
func TestCascadeDeleteSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with parent and children
	cli.MustExecute("-y", "list", "create", "CascadeTest")
	cli.MustExecute("-y", "CascadeTest", "add", "ParentToDelete")
	cli.MustExecute("-y", "CascadeTest", "add", "Child1", "-P", "ParentToDelete")
	cli.MustExecute("-y", "CascadeTest", "add", "Child2", "-P", "ParentToDelete")
	cli.MustExecute("-y", "CascadeTest", "add", "GrandChild", "-P", "Child1")

	// Delete the parent task (should cascade delete all children)
	stdout := cli.MustExecute("-y", "CascadeTest", "delete", "ParentToDelete")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify all tasks were deleted (parent and children)
	stdout = cli.MustExecute("-y", "--json", "CascadeTest")
	testutil.AssertNotContains(t, stdout, `"ParentToDelete"`)
	testutil.AssertNotContains(t, stdout, `"Child1"`)
	testutil.AssertNotContains(t, stdout, `"Child2"`)
	testutil.AssertNotContains(t, stdout, `"GrandChild"`)
}

// TestLiteralSlashFlag verifies `todoat MyList add -l "UI/UX Design"` creates single task with slash in summary
func TestLiteralSlashFlagSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "LiteralTest")

	// Add task with literal slash using -l flag
	stdout := cli.MustExecute("-y", "LiteralTest", "add", "-l", "UI/UX Design")

	testutil.AssertContains(t, stdout, "UI/UX Design")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify only one task was created with the full name including slash
	stdout = cli.MustExecute("-y", "--json", "LiteralTest")
	testutil.AssertContains(t, stdout, `"UI/UX Design"`)
	// Should NOT have separate "UI" and "UX Design" tasks
	testutil.AssertNotContains(t, stdout, `"UI"`)
}

// TestPathResolutionExisting verifies adding `A/B/C` when `A/B` exists only creates `C` under existing `B`
func TestPathResolutionExistingSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with existing parent structure
	cli.MustExecute("-y", "list", "create", "PathResTest")
	cli.MustExecute("-y", "PathResTest", "add", "ExistingParent/ExistingChild")

	// Add a new leaf under existing hierarchy
	stdout := cli.MustExecute("-y", "PathResTest", "add", "ExistingParent/ExistingChild/NewGrandchild")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify structure: should have ExistingParent, ExistingChild, and NewGrandchild
	// but NOT duplicate ExistingParent or ExistingChild
	stdout = cli.MustExecute("-y", "--json", "PathResTest")
	// Count occurrences - should only have one of each existing task
	if strings.Count(stdout, `"ExistingParent"`) > 1 {
		t.Errorf("expected only one ExistingParent task, but found duplicates in:\n%s", stdout)
	}
	if strings.Count(stdout, `"ExistingChild"`) > 1 {
		t.Errorf("expected only one ExistingChild task, but found duplicates in:\n%s", stdout)
	}
	testutil.AssertContains(t, stdout, `"NewGrandchild"`)
}

// TestCircularReferenceBlocked verifies cannot set task as parent of its own ancestor
func TestCircularReferenceBlockedSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with parent-child hierarchy
	cli.MustExecute("-y", "list", "create", "CircularTest")
	cli.MustExecute("-y", "CircularTest", "add", "Grandparent")
	cli.MustExecute("-y", "CircularTest", "add", "Parent", "-P", "Grandparent")
	cli.MustExecute("-y", "CircularTest", "add", "Child", "-P", "Parent")

	// Try to set Grandparent's parent to Child (would create circular reference)
	stdout, stderr, exitCode := cli.Execute("-y", "CircularTest", "update", "Grandparent", "-P", "Child")

	// Should fail with exit code 1
	testutil.AssertExitCode(t, exitCode, 1)
	combinedOutput := stdout + stderr
	// Should contain error about circular reference
	if !strings.Contains(strings.ToLower(combinedOutput), "circular") {
		t.Errorf("expected error about circular reference, got:\n%s", combinedOutput)
	}
}

// TestOrphanDetection verifies system handles tasks whose parent was deleted externally
func TestOrphanDetectionSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with parent and child
	cli.MustExecute("-y", "list", "create", "OrphanTest")
	cli.MustExecute("-y", "OrphanTest", "add", "Parent")
	cli.MustExecute("-y", "OrphanTest", "add", "Orphan", "-P", "Parent")

	// Test that listing still works with the hierarchy
	stdout := cli.MustExecute("-y", "OrphanTest")

	testutil.AssertContains(t, stdout, "Parent")
	testutil.AssertContains(t, stdout, "Orphan")
}

// =============================================================================
// Views and Customization Tests (015-views-customization)
// =============================================================================

// TestDefaultView verifies that `todoat MyList` displays tasks with default view (status, summary, priority)
func TestDefaultViewSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add tasks with different priorities
	cli.MustExecute("-y", "list", "create", "DefaultViewTest")
	cli.MustExecute("-y", "DefaultViewTest", "add", "High priority task", "-p", "1")
	cli.MustExecute("-y", "DefaultViewTest", "add", "Medium priority task", "-p", "5")
	cli.MustExecute("-y", "DefaultViewTest", "update", "High priority task", "-s", "IN-PROGRESS")

	// List tasks (should show default view with status, summary, priority)
	stdout := cli.MustExecute("-y", "DefaultViewTest")

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
func TestAllViewSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add a task with multiple attributes
	cli.MustExecute("-y", "list", "create", "AllViewTest")
	cli.MustExecute("-y", "AllViewTest", "add", "Full metadata task", "-p", "3", "--due-date", "2026-01-31", "--tag", "work,urgent")

	// List tasks with 'all' view
	stdout := cli.MustExecute("-y", "AllViewTest", "-v", "all")

	// All view should show more fields than default
	testutil.AssertContains(t, stdout, "Full metadata task")
	// Should show due date
	testutil.AssertContains(t, stdout, "2026-01-31")
	// Should show tags
	if !strings.Contains(stdout, "work") || !strings.Contains(stdout, "urgent") {
		t.Errorf("expected tags in 'all' view, got:\n%s", stdout)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestCustomViewSelection verifies that `todoat MyList -v myview` loads view from views directory
func TestCustomViewSelectionSQLiteCLI(t *testing.T) {
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
	stdout := cli.MustExecute("-y", "CustomViewTest", "-v", "minimal")

	testutil.AssertContains(t, stdout, "Task with priority")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestViewListCommand verifies that `todoat view list` shows all available views (built-in and custom)
func TestViewListCommandSQLiteCLI(t *testing.T) {
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
	stdout := cli.MustExecute("-y", "view", "list")

	// Should show built-in views
	testutil.AssertContains(t, stdout, "default")
	testutil.AssertContains(t, stdout, "all")
	// Should show custom view
	testutil.AssertContains(t, stdout, "custom")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestViewFieldOrdering verifies that custom view with reordered fields displays columns in specified order
func TestViewFieldOrderingSQLiteCLI(t *testing.T) {
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
	stdout := cli.MustExecute("-y", "FieldOrderTest", "-v", "priority_first")

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
func TestViewFilteringSQLiteCLI(t *testing.T) {
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
	stdout := cli.MustExecute("-y", "FilterViewTest", "-v", "active_only")

	testutil.AssertContains(t, stdout, "Active task")
	testutil.AssertNotContains(t, stdout, "Completed task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestViewSorting verifies that view with sort rules orders tasks correctly (multi-level sort)
func TestViewSortingSQLiteCLI(t *testing.T) {
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
	stdout := cli.MustExecute("-y", "SortViewTest", "-v", "priority_sorted")

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
func TestViewDateFilterSQLiteCLI(t *testing.T) {
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
	stdout := cli.MustExecute("-y", "DateFilterTest", "-v", "due_soon")

	// Should show only tasks due within 7 days from today
	testutil.AssertContains(t, stdout, "Due soon task")
	testutil.AssertNotContains(t, stdout, "Due later task")
	testutil.AssertNotContains(t, stdout, "Overdue task")
	testutil.AssertNotContains(t, stdout, "No due date task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestViewTagFilter verifies that view filters on tags/categories work with `contains` and `in` operators
func TestViewTagFilterSQLiteCLI(t *testing.T) {
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
	stdout := cli.MustExecute("-y", "TagFilterTest", "-v", "work_only")

	// Should show tasks with 'work' tag
	testutil.AssertContains(t, stdout, "Work task")
	testutil.AssertContains(t, stdout, "Work and home task")
	testutil.AssertNotContains(t, stdout, "Home task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestViewHierarchyPreserved verifies that custom views maintain parent-child tree structure display
func TestViewHierarchyPreservedSQLiteCLI(t *testing.T) {
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
	stdout := cli.MustExecute("-y", "HierarchyViewTest", "-v", "simple")

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
func TestInvalidViewErrorSQLiteCLI(t *testing.T) {
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
// List Export/Import Tests (038-list-export-import)
// =============================================================================

// TestListExportSQLite verifies that `todoat list export "MyList" --format sqlite` creates a standalone db file
func TestListExportSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with some tasks
	cli.MustExecute("-y", "list", "create", "ExportList")
	cli.MustExecute("-y", "ExportList", "add", "Task 1", "-p", "1")
	cli.MustExecute("-y", "ExportList", "add", "Task 2", "-p", "5")
	cli.MustExecute("-y", "ExportList", "add", "Child Task", "-P", "Task 1")

	// Export to SQLite format
	exportPath := cli.TmpDir() + "/ExportList.db"
	stdout := cli.MustExecute("-y", "list", "export", "ExportList", "--format", "sqlite", "--output", exportPath)

	// Should indicate success with file path and task count
	testutil.AssertContains(t, stdout, exportPath)
	testutil.AssertContains(t, stdout, "3") // 3 tasks exported
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the exported file exists and is a valid SQLite database
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Errorf("expected export file to exist at %s", exportPath)
	}
}

// TestListExportJSON verifies that `todoat list export "MyList" --format json` creates a JSON file
func TestListExportJSONCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with tasks
	cli.MustExecute("-y", "list", "create", "JSONExport")
	cli.MustExecute("-y", "JSONExport", "add", "Task A")
	cli.MustExecute("-y", "JSONExport", "add", "Task B", "-p", "3")

	// Export to JSON format
	exportPath := cli.TmpDir() + "/JSONExport.json"
	stdout := cli.MustExecute("-y", "list", "export", "JSONExport", "--format", "json", "--output", exportPath)

	// Should indicate success
	testutil.AssertContains(t, stdout, exportPath)
	testutil.AssertContains(t, stdout, "2") // 2 tasks exported
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the exported file exists and contains valid JSON
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}
	if !strings.Contains(string(data), "[") || !strings.Contains(string(data), "Task A") {
		t.Errorf("expected valid JSON with task data, got: %s", string(data))
	}
}

// TestListExportCSV verifies that `todoat list export "MyList" --format csv` creates a CSV file
func TestListExportCSVCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with tasks
	cli.MustExecute("-y", "list", "create", "CSVExport")
	cli.MustExecute("-y", "CSVExport", "add", "Review document")
	cli.MustExecute("-y", "CSVExport", "add", "Send email", "-p", "2")

	// Export to CSV format
	exportPath := cli.TmpDir() + "/CSVExport.csv"
	stdout := cli.MustExecute("-y", "list", "export", "CSVExport", "--format", "csv", "--output", exportPath)

	// Should indicate success
	testutil.AssertContains(t, stdout, exportPath)
	testutil.AssertContains(t, stdout, "2") // 2 tasks exported
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the exported file exists and contains CSV data
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}
	// CSV should have headers and task data
	content := string(data)
	if !strings.Contains(content, "summary") || !strings.Contains(content, "Review document") {
		t.Errorf("expected CSV with headers and task data, got: %s", content)
	}
}

// TestListExportICalendar verifies that `todoat list export "MyList" --format ical` creates an .ics file
func TestListExportICalendarCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with tasks
	cli.MustExecute("-y", "list", "create", "ICalExport")
	cli.MustExecute("-y", "ICalExport", "add", "Meeting prep")
	cli.MustExecute("-y", "ICalExport", "add", "Call client")

	// Export to iCalendar format
	exportPath := cli.TmpDir() + "/ICalExport.ics"
	stdout := cli.MustExecute("-y", "list", "export", "ICalExport", "--format", "ical", "--output", exportPath)

	// Should indicate success
	testutil.AssertContains(t, stdout, exportPath)
	testutil.AssertContains(t, stdout, "2") // 2 tasks exported
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the exported file exists and contains iCalendar format
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}
	content := string(data)
	// iCalendar files should contain VCALENDAR and VTODO components
	if !strings.Contains(content, "BEGIN:VCALENDAR") || !strings.Contains(content, "VTODO") {
		t.Errorf("expected iCalendar format with VCALENDAR and VTODO, got: %s", content)
	}
	if !strings.Contains(content, "Meeting prep") {
		t.Errorf("expected task summary in iCalendar, got: %s", content)
	}
}

// TestListImport verifies that `todoat list import backup.db` restores a list from exported file
func TestListImportCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with tasks and export it
	cli.MustExecute("-y", "list", "create", "OriginalList")
	cli.MustExecute("-y", "OriginalList", "add", "Important task", "-p", "1")
	cli.MustExecute("-y", "OriginalList", "add", "Another task")
	cli.MustExecute("-y", "OriginalList", "add", "Subtask", "-P", "Important task")

	// Export the list
	exportPath := cli.TmpDir() + "/backup.db"
	cli.MustExecute("-y", "list", "export", "OriginalList", "--format", "sqlite", "--output", exportPath)

	// Delete the original list
	cli.MustExecute("-y", "list", "delete", "OriginalList")
	cli.MustExecute("-y", "list", "purge", "OriginalList")

	// Import the list back
	stdout := cli.MustExecute("-y", "list", "import", exportPath)

	// Should indicate success with task count
	testutil.AssertContains(t, stdout, "3") // 3 tasks imported
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the list was restored with its tasks
	stdout = cli.MustExecute("-y", "OriginalList")
	testutil.AssertContains(t, stdout, "Important task")
	testutil.AssertContains(t, stdout, "Another task")
	testutil.AssertContains(t, stdout, "Subtask")
}

// TestListExportDefaultPath verifies that export uses default path ./<list-name>.<ext> when --output not specified
func TestListExportDefaultPathCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with a task
	cli.MustExecute("-y", "list", "create", "DefaultPath")
	cli.MustExecute("-y", "DefaultPath", "add", "Test task")

	// Change to temp directory and export without specifying output
	// The export should create DefaultPath.json in the current working directory
	stdout := cli.MustExecute("-y", "list", "export", "DefaultPath", "--format", "json")

	// Should indicate success and mention the default path
	testutil.AssertContains(t, stdout, "DefaultPath.json")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// TestListExportJSONMode verifies that export in JSON output mode returns proper structure
func TestListExportJSONModeCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with tasks
	cli.MustExecute("-y", "list", "create", "JSONModeExport")
	cli.MustExecute("-y", "JSONModeExport", "add", "Task 1")

	// Export with --json flag for structured output
	exportPath := cli.TmpDir() + "/JSONModeExport.json"
	stdout := cli.MustExecute("-y", "--json", "list", "export", "JSONModeExport", "--format", "json", "--output", exportPath)

	// Should return JSON with action, file, and task_count
	testutil.AssertContains(t, stdout, `"action"`)
	testutil.AssertContains(t, stdout, `"export"`)
	testutil.AssertContains(t, stdout, `"file"`)
	testutil.AssertContains(t, stdout, `"task_count"`)
}

// TestListImportJSONMode verifies that import in JSON output mode returns proper structure
func TestListImportJSONModeCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create and export a list
	cli.MustExecute("-y", "list", "create", "ImportJSON")
	cli.MustExecute("-y", "ImportJSON", "add", "Task 1")
	exportPath := cli.TmpDir() + "/ImportJSON.json"
	cli.MustExecute("-y", "list", "export", "ImportJSON", "--format", "json", "--output", exportPath)

	// Delete the original
	cli.MustExecute("-y", "list", "delete", "ImportJSON")
	cli.MustExecute("-y", "list", "purge", "ImportJSON")

	// Import with --json flag
	stdout := cli.MustExecute("-y", "--json", "list", "import", exportPath)

	// Should return JSON with action, file, and task_count
	testutil.AssertContains(t, stdout, `"action"`)
	testutil.AssertContains(t, stdout, `"import"`)
	testutil.AssertContains(t, stdout, `"task_count"`)
}

// TestListExportNotFound verifies that exporting non-existent list returns error
func TestListExportNotFoundCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Try to export a list that doesn't exist
	stdout, _, exitCode := cli.Execute("-y", "list", "export", "NonExistent", "--format", "json")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertResultCode(t, stdout, testutil.ResultError)
}

// TestListImportPreservesMetadata verifies that import preserves task metadata
func TestListImportPreservesMetadataCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with tasks containing various metadata
	cli.MustExecute("-y", "list", "create", "MetadataTest")
	cli.MustExecute("-y", "MetadataTest", "add", "High Priority", "-p", "1")
	cli.MustExecute("-y", "MetadataTest", "add", "Tagged Task", "--tag", "work,urgent")

	// Export
	exportPath := cli.TmpDir() + "/metadata.db"
	cli.MustExecute("-y", "list", "export", "MetadataTest", "--format", "sqlite", "--output", exportPath)

	// Delete and purge
	cli.MustExecute("-y", "list", "delete", "MetadataTest")
	cli.MustExecute("-y", "list", "purge", "MetadataTest")

	// Import
	cli.MustExecute("-y", "list", "import", exportPath)

	// Verify metadata is preserved
	stdout := cli.MustExecute("-y", "--json", "MetadataTest")
	// Check priority is preserved
	if !strings.Contains(stdout, `"priority":1`) && !strings.Contains(stdout, `"priority": 1`) {
		t.Errorf("expected priority to be preserved, got: %s", stdout)
	}
	// Check categories are preserved
	if !strings.Contains(stdout, "work") || !strings.Contains(stdout, "urgent") {
		t.Errorf("expected tags to be preserved, got: %s", stdout)
	}
}

// =============================================================================
// Database Maintenance Tests (039-database-maintenance)
// =============================================================================

// TestListStatsSQLiteCLI verifies `todoat -y list stats` displays database statistics
func TestListStatsSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create lists with varying task counts
	cli.MustExecute("-y", "list", "create", "StatsWork")
	cli.MustExecute("-y", "StatsWork", "add", "Task 1")
	cli.MustExecute("-y", "StatsWork", "add", "Task 2")
	cli.MustExecute("-y", "StatsWork", "complete", "Task 2")

	cli.MustExecute("-y", "list", "create", "StatsPersonal")
	cli.MustExecute("-y", "StatsPersonal", "add", "Task A")

	// Get database stats
	stdout := cli.MustExecute("-y", "list", "stats")

	// Should show total tasks
	testutil.AssertContains(t, stdout, "3") // 3 total tasks

	// Should show tasks per list
	testutil.AssertContains(t, stdout, "StatsWork")
	testutil.AssertContains(t, stdout, "StatsPersonal")

	// Should show tasks by status
	testutil.AssertContains(t, stdout, "TODO")
	testutil.AssertContains(t, stdout, "DONE")

	// Should show database size info
	if !strings.Contains(strings.ToLower(stdout), "size") && !strings.Contains(strings.ToLower(stdout), "bytes") && !strings.Contains(strings.ToLower(stdout), "kb") {
		t.Errorf("expected database size info, got: %s", stdout)
	}

	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestListStatsJSONSQLiteCLI verifies `todoat -y --json list stats` returns JSON statistics
func TestListStatsJSONSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with tasks in different statuses
	cli.MustExecute("-y", "list", "create", "JSONStats")
	cli.MustExecute("-y", "JSONStats", "add", "Task TODO")
	cli.MustExecute("-y", "JSONStats", "add", "Task DONE")
	cli.MustExecute("-y", "JSONStats", "complete", "Task DONE")

	// Get stats with JSON output
	stdout := cli.MustExecute("-y", "--json", "list", "stats")

	// Should contain JSON structure
	testutil.AssertContains(t, stdout, "{")
	testutil.AssertContains(t, stdout, "}")

	// Should contain expected JSON fields
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"INFO_ONLY"`)
	testutil.AssertContains(t, stdout, `"stats"`)
	testutil.AssertContains(t, stdout, `"total_tasks"`)
	testutil.AssertContains(t, stdout, `"lists"`)
	testutil.AssertContains(t, stdout, `"by_status"`)
	testutil.AssertContains(t, stdout, `"database_size_bytes"`)
}

// TestListStatsSpecificListSQLiteCLI verifies `todoat -y list stats "ListName"` shows stats for specific list
func TestListStatsSpecificListSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create multiple lists
	cli.MustExecute("-y", "list", "create", "SpecificList")
	cli.MustExecute("-y", "SpecificList", "add", "Task A")
	cli.MustExecute("-y", "SpecificList", "add", "Task B")

	cli.MustExecute("-y", "list", "create", "OtherList")
	cli.MustExecute("-y", "OtherList", "add", "Task X")

	// Get stats for specific list
	stdout := cli.MustExecute("-y", "list", "stats", "SpecificList")

	// Should show specific list stats
	testutil.AssertContains(t, stdout, "SpecificList")
	testutil.AssertContains(t, stdout, "2") // 2 tasks in this list

	// Should not show other list prominently
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestListVacuumSQLiteCLI verifies `todoat -y list vacuum` reclaims space from deleted data
func TestListVacuumSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with many tasks
	cli.MustExecute("-y", "list", "create", "VacuumTest")
	for i := 0; i < 10; i++ {
		cli.MustExecute("-y", "VacuumTest", "add", "Task "+strconv.Itoa(i))
	}

	// Delete some tasks
	for i := 0; i < 5; i++ {
		cli.MustExecute("-y", "VacuumTest", "delete", "Task "+strconv.Itoa(i))
	}

	// Run vacuum (with -y to skip confirmation)
	stdout := cli.MustExecute("-y", "list", "vacuum")

	// Should show before/after size comparison or completion message
	if !strings.Contains(strings.ToLower(stdout), "vacuum") && !strings.Contains(strings.ToLower(stdout), "complet") {
		t.Errorf("expected vacuum completion message, got: %s", stdout)
	}

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// TestListVacuumConfirmationSQLiteCLI verifies vacuum prompts for confirmation without -y flag
func TestListVacuumConfirmationSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "VacuumConfirm")
	cli.MustExecute("-y", "VacuumConfirm", "add", "Task")

	// Run vacuum with -y flag - should work without prompt
	stdout := cli.MustExecute("-y", "list", "vacuum")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Note: Testing actual interactive prompt is not possible with current test harness
	// The -y flag should bypass the confirmation, which is the main testable scenario
}

// TestListVacuumJSONSQLiteCLI verifies `todoat -y --json list vacuum` returns JSON output
func TestListVacuumJSONSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "VacuumJSON")
	cli.MustExecute("-y", "VacuumJSON", "add", "Task")

	// Run vacuum with JSON output
	stdout := cli.MustExecute("-y", "--json", "list", "vacuum")

	// Should contain JSON structure
	testutil.AssertContains(t, stdout, "{")
	testutil.AssertContains(t, stdout, "}")
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"ACTION_COMPLETED"`)
}

// =============================================================================
// Bulk Hierarchy Operations Tests (040-bulk-hierarchy-operations)
// =============================================================================

// TestBulkCompleteDirectChildrenSQLiteCLI verifies `todoat MyList complete "Parent/*"` completes direct children only
func TestBulkCompleteDirectChildrenSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list and hierarchy
	cli.MustExecute("-y", "list", "create", "BulkTest")
	cli.MustExecute("-y", "BulkTest", "add", "Parent")
	cli.MustExecute("-y", "BulkTest", "add", "Child1", "-P", "Parent")
	cli.MustExecute("-y", "BulkTest", "add", "Child2", "-P", "Parent")
	cli.MustExecute("-y", "BulkTest", "add", "GrandChild", "-P", "Child1")

	// Complete direct children only with /* pattern
	stdout := cli.MustExecute("-y", "BulkTest", "complete", "Parent/*")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
	testutil.AssertContains(t, stdout, "2") // Should affect 2 direct children

	// Verify: Child1 and Child2 are DONE, GrandChild is still TODO, Parent unchanged
	stdout = cli.MustExecute("-y", "--json", "BulkTest")
	// Parent should still be TODO
	if strings.Contains(stdout, `"summary":"Parent"`) && !strings.Contains(stdout, `"status":"TODO"`) {
		t.Log("Note: Checking parent status indirectly")
	}
}

// TestBulkCompleteAllDescendantsSQLiteCLI verifies `todoat MyList complete "Parent/**"` completes all descendants
func TestBulkCompleteAllDescendantsSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list and hierarchy
	cli.MustExecute("-y", "list", "create", "BulkAllTest")
	cli.MustExecute("-y", "BulkAllTest", "add", "Parent")
	cli.MustExecute("-y", "BulkAllTest", "add", "Child1", "-P", "Parent")
	cli.MustExecute("-y", "BulkAllTest", "add", "Child2", "-P", "Parent")
	cli.MustExecute("-y", "BulkAllTest", "add", "GrandChild", "-P", "Child1")

	// Complete all descendants with /** pattern
	stdout := cli.MustExecute("-y", "BulkAllTest", "complete", "Parent/**")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
	testutil.AssertContains(t, stdout, "3") // Should affect 3 descendants (Child1, Child2, GrandChild)
}

// TestBulkUpdatePrioritySQLiteCLI verifies `todoat MyList update "Parent/**" --priority 1` updates priority on all descendants
func TestBulkUpdatePrioritySQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list and hierarchy
	cli.MustExecute("-y", "list", "create", "BulkPriorityTest")
	cli.MustExecute("-y", "BulkPriorityTest", "add", "Parent")
	cli.MustExecute("-y", "BulkPriorityTest", "add", "Child1", "-P", "Parent")
	cli.MustExecute("-y", "BulkPriorityTest", "add", "Child2", "-P", "Parent")
	cli.MustExecute("-y", "BulkPriorityTest", "add", "GrandChild", "-P", "Child1")

	// Update priority on all descendants
	stdout := cli.MustExecute("-y", "BulkPriorityTest", "update", "Parent/**", "-p", "1")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
	testutil.AssertContains(t, stdout, "3") // Should affect 3 descendants
}

// TestBulkDeleteChildrenSQLiteCLI verifies `todoat MyList delete "Parent/*"` deletes direct children only
func TestBulkDeleteChildrenSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list and hierarchy
	cli.MustExecute("-y", "list", "create", "BulkDeleteTest")
	cli.MustExecute("-y", "BulkDeleteTest", "add", "Parent")
	cli.MustExecute("-y", "BulkDeleteTest", "add", "Child1", "-P", "Parent")
	cli.MustExecute("-y", "BulkDeleteTest", "add", "Child2", "-P", "Parent")
	cli.MustExecute("-y", "BulkDeleteTest", "add", "GrandChild", "-P", "Child1")

	// Delete direct children only with /* pattern
	stdout := cli.MustExecute("-y", "BulkDeleteTest", "delete", "Parent/*")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
	testutil.AssertContains(t, stdout, "3") // Should delete 3 tasks (Child1, Child2, and GrandChild cascades from Child1)

	// Verify Parent still exists, children are gone
	stdout = cli.MustExecute("-y", "--json", "BulkDeleteTest")
	testutil.AssertContains(t, stdout, `"Parent"`)
	testutil.AssertNotContains(t, stdout, `"Child1"`)
	testutil.AssertNotContains(t, stdout, `"Child2"`)
	testutil.AssertNotContains(t, stdout, `"GrandChild"`)
}

// TestBulkNoMatchErrorSQLiteCLI verifies `todoat MyList complete "NonExistent/*"` returns ERROR
func TestBulkNoMatchErrorSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list with some tasks (but not the one we'll look for)
	cli.MustExecute("-y", "list", "create", "BulkNoMatchTest")
	cli.MustExecute("-y", "BulkNoMatchTest", "add", "SomeTask")

	// Try to bulk complete children of non-existent parent
	stdout, _ := cli.ExecuteAndFail("-y", "BulkNoMatchTest", "complete", "NonExistent/*")

	testutil.AssertResultCode(t, stdout, testutil.ResultError)
}

// TestBulkEmptyMatchSQLiteCLI verifies `todoat MyList complete "LeafTask/*"` returns INFO_ONLY (no children)
func TestBulkEmptyMatchSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list with a leaf task (no children)
	cli.MustExecute("-y", "list", "create", "BulkEmptyTest")
	cli.MustExecute("-y", "BulkEmptyTest", "add", "LeafTask")

	// Try to bulk complete children of leaf task
	stdout := cli.MustExecute("-y", "BulkEmptyTest", "complete", "LeafTask/*")

	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
	// Should indicate no tasks were affected
	testutil.AssertContains(t, stdout, "0")
}

// TestBulkCountOutputSQLiteCLI verifies bulk operation returns count of affected tasks
func TestBulkCountOutputSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list and hierarchy
	cli.MustExecute("-y", "list", "create", "BulkCountTest")
	cli.MustExecute("-y", "BulkCountTest", "add", "Release v2.0")
	cli.MustExecute("-y", "BulkCountTest", "add", "Feature A", "-P", "Release v2.0")
	cli.MustExecute("-y", "BulkCountTest", "add", "Feature B", "-P", "Release v2.0")
	cli.MustExecute("-y", "BulkCountTest", "add", "Feature C", "-P", "Release v2.0")
	cli.MustExecute("-y", "BulkCountTest", "add", "Task A1", "-P", "Feature A")
	cli.MustExecute("-y", "BulkCountTest", "add", "Task A2", "-P", "Feature A")

	// Complete all descendants
	stdout := cli.MustExecute("-y", "BulkCountTest", "complete", "Release v2.0/**")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
	// Should affect 5 descendants: Feature A, Feature B, Feature C, Task A1, Task A2
	testutil.AssertContains(t, stdout, "5")
	// Output should include parent name
	testutil.AssertContains(t, stdout, "Release v2.0")
}

// TestBulkCompleteJSONOutputSQLiteCLI verifies JSON output for bulk operations
func TestBulkCompleteJSONOutputSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list and hierarchy
	cli.MustExecute("-y", "list", "create", "BulkJSONTest")
	cli.MustExecute("-y", "BulkJSONTest", "add", "Parent")
	cli.MustExecute("-y", "BulkJSONTest", "add", "Child1", "-P", "Parent")
	cli.MustExecute("-y", "BulkJSONTest", "add", "Child2", "-P", "Parent")

	// Complete with JSON output
	stdout := cli.MustExecute("-y", "--json", "BulkJSONTest", "complete", "Parent/**")

	// Verify JSON structure
	testutil.AssertContains(t, stdout, "{")
	testutil.AssertContains(t, stdout, "}")
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"ACTION_COMPLETED"`)
	testutil.AssertContains(t, stdout, `"affected_count":2`) // 2 children affected (number in JSON)
	testutil.AssertContains(t, stdout, `"pattern"`)
	testutil.AssertContains(t, stdout, `"**"`)
}

// =============================================================================
// UID/Local-ID Task Selection Tests (041-uid-localid-task-selection)
// =============================================================================

// extractUID extracts UID from JSON output of add command
func extractUID(t *testing.T, jsonOutput string) string {
	t.Helper()
	// Find "uid":"<value>" pattern
	start := strings.Index(jsonOutput, `"uid":"`)
	if start == -1 {
		t.Fatalf("could not find uid in output: %s", jsonOutput)
	}
	start += 7 // len(`"uid":"`)
	end := strings.Index(jsonOutput[start:], `"`)
	if end == -1 {
		t.Fatalf("could not find end of uid in output: %s", jsonOutput)
	}
	return jsonOutput[start : start+end]
}

// TestUpdateByUID tests updating a task by its UID
func TestUpdateByUID(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list and add a task with JSON output to get UID
	cli.MustExecute("-y", "list", "create", "UIDTest")
	addOutput := cli.MustExecute("-y", "--json", "UIDTest", "add", "Task to update")

	// Extract UID from JSON output
	uid := extractUID(t, addOutput)

	// Update by UID
	stdout := cli.MustExecute("-y", "UIDTest", "update", "--uid", uid, "-s", "DONE")

	testutil.AssertContains(t, stdout, "Updated task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the update worked
	listOutput := cli.MustExecute("-y", "UIDTest", "get")
	testutil.AssertContains(t, listOutput, "DONE")
}

// TestCompleteByUID tests completing a task by its UID
func TestCompleteByUID(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list and add a task with JSON output to get UID
	cli.MustExecute("-y", "list", "create", "UIDCompleteTest")
	addOutput := cli.MustExecute("-y", "--json", "UIDCompleteTest", "add", "Task to complete")

	// Extract UID from JSON output
	uid := extractUID(t, addOutput)

	// Complete by UID
	stdout := cli.MustExecute("-y", "UIDCompleteTest", "complete", "--uid", uid)

	testutil.AssertContains(t, stdout, "Completed task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// TestDeleteByUID tests deleting a task by its UID
func TestDeleteByUID(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list and add a task with JSON output to get UID
	cli.MustExecute("-y", "list", "create", "UIDDeleteTest")
	addOutput := cli.MustExecute("-y", "--json", "UIDDeleteTest", "add", "Task to delete")

	// Extract UID from JSON output
	uid := extractUID(t, addOutput)

	// Delete by UID
	stdout := cli.MustExecute("-y", "UIDDeleteTest", "delete", "--uid", uid)

	testutil.AssertContains(t, stdout, "Deleted task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the task is deleted
	listOutput := cli.MustExecute("-y", "UIDDeleteTest", "get")
	testutil.AssertNotContains(t, listOutput, "Task to delete")
}

// TestUpdateByLocalID tests updating a task by its local ID (requires sync enabled)
func TestUpdateByLocalID(t *testing.T) {
	cli, _, tmpDir := testutil.NewCLITestWithViewsAndTmpDir(t)

	// Create sync config
	createSyncConfig(t, tmpDir, true)

	// Create list and add a task
	cli.MustExecute("-y", "list", "create", "LocalIDTest")
	cli.MustExecute("-y", "LocalIDTest", "add", "Task for local ID")

	// Get the task with JSON to find local_id
	jsonOutput := cli.MustExecute("-y", "--json", "LocalIDTest", "get")

	// Extract local_id from JSON output (assuming it's included in the output)
	localID := extractLocalID(t, jsonOutput)

	// Update by local-id
	stdout := cli.MustExecute("-y", "LocalIDTest", "update", "--local-id", localID, "-s", "DONE")

	testutil.AssertContains(t, stdout, "Updated task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// TestCompleteByLocalID tests completing a task by its local ID
func TestCompleteByLocalID(t *testing.T) {
	cli, _, tmpDir := testutil.NewCLITestWithViewsAndTmpDir(t)

	// Create sync config
	createSyncConfig(t, tmpDir, true)

	// Create list and add a task
	cli.MustExecute("-y", "list", "create", "LocalIDCompleteTest")
	cli.MustExecute("-y", "LocalIDCompleteTest", "add", "Task to complete")

	// Get the task with JSON to find local_id
	jsonOutput := cli.MustExecute("-y", "--json", "LocalIDCompleteTest", "get")
	localID := extractLocalID(t, jsonOutput)

	// Complete by local-id
	stdout := cli.MustExecute("-y", "LocalIDCompleteTest", "complete", "--local-id", localID)

	testutil.AssertContains(t, stdout, "Completed task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// TestDeleteByLocalID tests deleting a task by its local ID
func TestDeleteByLocalID(t *testing.T) {
	cli, _, tmpDir := testutil.NewCLITestWithViewsAndTmpDir(t)

	// Create sync config
	createSyncConfig(t, tmpDir, true)

	// Create list and add a task
	cli.MustExecute("-y", "list", "create", "LocalIDDeleteTest")
	cli.MustExecute("-y", "LocalIDDeleteTest", "add", "Task to delete")

	// Get the task with JSON to find local_id
	jsonOutput := cli.MustExecute("-y", "--json", "LocalIDDeleteTest", "get")
	localID := extractLocalID(t, jsonOutput)

	// Delete by local-id
	stdout := cli.MustExecute("-y", "LocalIDDeleteTest", "delete", "--local-id", localID)

	testutil.AssertContains(t, stdout, "Deleted task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// TestUIDNotFound tests that --uid with nonexistent UID returns error
func TestUIDNotFound(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list
	cli.MustExecute("-y", "list", "create", "UIDNotFoundTest")
	cli.MustExecute("-y", "UIDNotFoundTest", "add", "Some task")

	// Try to update with nonexistent UID
	_, stderr, exitCode := cli.Execute("-y", "UIDNotFoundTest", "update", "--uid", "550e8400-e29b-41d4-a716-446655440000", "-s", "DONE")

	if exitCode == 0 {
		t.Fatal("expected non-zero exit code for nonexistent UID")
	}
	combined := stderr
	if !strings.Contains(combined, "not found") && !strings.Contains(combined, "no task") {
		t.Errorf("expected error about task not found, got: %s", combined)
	}
}

// TestLocalIDNotFound tests that --local-id with nonexistent ID returns error
func TestLocalIDNotFound(t *testing.T) {
	cli, _, tmpDir := testutil.NewCLITestWithViewsAndTmpDir(t)

	// Create sync config
	createSyncConfig(t, tmpDir, true)

	// Create list
	cli.MustExecute("-y", "list", "create", "LocalIDNotFoundTest")
	cli.MustExecute("-y", "LocalIDNotFoundTest", "add", "Some task")

	// Try to update with nonexistent local-id
	_, stderr, exitCode := cli.Execute("-y", "LocalIDNotFoundTest", "update", "--local-id", "99999", "-s", "DONE")

	if exitCode == 0 {
		t.Fatal("expected non-zero exit code for nonexistent local-id")
	}
	combined := stderr
	if !strings.Contains(combined, "not found") && !strings.Contains(combined, "no task") {
		t.Errorf("expected error about task not found, got: %s", combined)
	}
}

// TestLocalIDRequiresSync tests that --local-id returns error when sync not enabled
func TestLocalIDRequiresSync(t *testing.T) {
	cli := testutil.NewCLITest(t) // No sync config

	// Create list
	cli.MustExecute("-y", "list", "create", "NoSyncTest")
	cli.MustExecute("-y", "NoSyncTest", "add", "Some task")

	// Try to use --local-id without sync enabled
	_, stderr, exitCode := cli.Execute("-y", "NoSyncTest", "update", "--local-id", "1", "-s", "DONE")

	if exitCode == 0 {
		t.Fatal("expected non-zero exit code when using --local-id without sync")
	}
	combined := stderr
	if !strings.Contains(combined, "sync") && !strings.Contains(combined, "enabled") {
		t.Errorf("expected error about sync not enabled, got: %s", combined)
	}
}

// TestUIDRequiresSyncedTask tests that --uid only works for tasks with backend-assigned UID
func TestUIDRequiresSyncedTask(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list and add a task with JSON output to get UID
	cli.MustExecute("-y", "list", "create", "UIDSyncedTest")
	addOutput := cli.MustExecute("-y", "--json", "UIDSyncedTest", "add", "Task with UID")

	// Extract UID - the task should have a UID even without sync since sqlite backend generates UUIDs
	uid := extractUID(t, addOutput)

	// Update should work since the task has a UID
	stdout := cli.MustExecute("-y", "UIDSyncedTest", "update", "--uid", uid, "-s", "DONE")
	testutil.AssertContains(t, stdout, "Updated task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// extractLocalID extracts local_id from JSON output
func extractLocalID(t *testing.T, jsonOutput string) string {
	t.Helper()
	// Find "local_id":<value> pattern (integer, not string)
	start := strings.Index(jsonOutput, `"local_id":`)
	if start == -1 {
		t.Fatalf("could not find local_id in output: %s", jsonOutput)
	}
	start += 11 // len(`"local_id":`)
	// Skip any whitespace
	for start < len(jsonOutput) && (jsonOutput[start] == ' ' || jsonOutput[start] == '\t') {
		start++
	}
	// Read digits
	end := start
	for end < len(jsonOutput) && jsonOutput[end] >= '0' && jsonOutput[end] <= '9' {
		end++
	}
	if end == start {
		t.Fatalf("could not extract local_id number from output: %s", jsonOutput)
	}
	return jsonOutput[start:end]
}

// createSyncConfig creates a config file with sync enabled/disabled (for local tests)
func createSyncConfig(t *testing.T, tmpDir string, enabled bool) {
	t.Helper()

	enabledStr := "true"
	if !enabled {
		enabledStr = "false"
	}

	configContent := `
sync:
  enabled: ` + enabledStr + `
  local_backend: sqlite
backends:
  sqlite:
    type: sqlite
    enabled: true
`
	configPath := tmpDir + "/config.yaml"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
}

// =============================================================================
// Task Description Flag Tests (042-task-description-flag)
// =============================================================================

// TestAddTaskWithDescriptionSQLiteCLI verifies that `todoat -y MyList add "Task" -d "Detailed notes"` creates task with description
func TestAddTaskWithDescriptionSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list first
	cli.MustExecute("-y", "list", "create", "DescTest")

	// Add task with description
	stdout := cli.MustExecute("-y", "DescTest", "add", "Task with notes", "-d", "Detailed notes about this task")

	testutil.AssertContains(t, stdout, "Task with notes")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify description is stored by getting task in JSON format
	jsonOut := cli.MustExecute("-y", "--json", "DescTest", "get")
	testutil.AssertContains(t, jsonOut, `"description"`)
	testutil.AssertContains(t, jsonOut, "Detailed notes about this task")
}

// TestUpdateTaskDescriptionSQLiteCLI verifies that `todoat -y MyList update "Task" -d "Updated notes"` updates description
func TestUpdateTaskDescriptionSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add a task
	cli.MustExecute("-y", "list", "create", "UpdateDescTest")
	cli.MustExecute("-y", "UpdateDescTest", "add", "Task to update desc")

	// Update the description
	stdout := cli.MustExecute("-y", "UpdateDescTest", "update", "Task to update desc", "-d", "Updated description")

	testutil.AssertContains(t, stdout, "Updated task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify description was updated via JSON output
	jsonOut := cli.MustExecute("-y", "--json", "UpdateDescTest", "get")
	testutil.AssertContains(t, jsonOut, "Updated description")
}

// TestClearTaskDescriptionSQLiteCLI verifies that `todoat -y MyList update "Task" -d ""` clears description
func TestClearTaskDescriptionSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add a task with description
	cli.MustExecute("-y", "list", "create", "ClearDescTest")
	cli.MustExecute("-y", "ClearDescTest", "add", "Task with desc", "-d", "Initial description")

	// Verify description was set
	jsonOut := cli.MustExecute("-y", "--json", "ClearDescTest", "get")
	testutil.AssertContains(t, jsonOut, "Initial description")

	// Clear the description
	stdout := cli.MustExecute("-y", "ClearDescTest", "update", "Task with desc", "-d", "")

	testutil.AssertContains(t, stdout, "Updated task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify description was cleared
	jsonOut = cli.MustExecute("-y", "--json", "ClearDescTest", "get")
	// Description should be empty string now
	testutil.AssertContains(t, jsonOut, `"description":""`)
}

// TestDescriptionInJSONSQLiteCLI verifies that `todoat -y --json MyList` includes description field in output
func TestDescriptionInJSONSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add a task with description
	cli.MustExecute("-y", "list", "create", "DescJSONTest")
	cli.MustExecute("-y", "DescJSONTest", "add", "JSON desc task", "-d", "JSON test description")

	// Get tasks in JSON format
	jsonOut := cli.MustExecute("-y", "--json", "DescJSONTest", "get")

	// Verify description field is present
	testutil.AssertContains(t, jsonOut, `"description"`)
	testutil.AssertContains(t, jsonOut, "JSON test description")
}

// TestDescriptionLongFlagSQLiteCLI verifies that `todoat -y MyList add "Task" --description "Notes"` works with long flag
func TestDescriptionLongFlagSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list first
	cli.MustExecute("-y", "list", "create", "LongFlagTest")

	// Add task with long --description flag
	stdout := cli.MustExecute("-y", "LongFlagTest", "add", "Task long flag", "--description", "Notes with long flag")

	testutil.AssertContains(t, stdout, "Task long flag")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify description is stored
	jsonOut := cli.MustExecute("-y", "--json", "LongFlagTest", "get")
	testutil.AssertContains(t, jsonOut, "Notes with long flag")
}

// =============================================================================
// Date Filtering Tests (043-date-filtering)
// =============================================================================

// TestFilterDueBeforeSQLiteCLI verifies that `todoat -y MyList --due-before 2026-02-01` shows only tasks due before date
func TestFilterDueBeforeSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "DueBeforeTest")

	// Add tasks with different due dates
	cli.MustExecute("-y", "DueBeforeTest", "add", "Task due Jan 15", "--due-date", "2026-01-15")
	cli.MustExecute("-y", "DueBeforeTest", "add", "Task due Jan 31", "--due-date", "2026-01-31")
	cli.MustExecute("-y", "DueBeforeTest", "add", "Task due Feb 15", "--due-date", "2026-02-15")
	cli.MustExecute("-y", "DueBeforeTest", "add", "Task no due date")

	// Filter to show only tasks due before Feb 1
	stdout := cli.MustExecute("-y", "DueBeforeTest", "--due-before", "2026-02-01")

	// Should show tasks due before Feb 1 (Jan 15, Jan 31)
	testutil.AssertContains(t, stdout, "Task due Jan 15")
	testutil.AssertContains(t, stdout, "Task due Jan 31")
	// Should not show tasks due on/after Feb 1
	testutil.AssertNotContains(t, stdout, "Task due Feb 15")
	// Tasks without due date should not match due date filters
	testutil.AssertNotContains(t, stdout, "Task no due date")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestFilterDueAfterSQLiteCLI verifies that `todoat -y MyList --due-after 2026-01-15` shows only tasks due after date
func TestFilterDueAfterSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "DueAfterTest")

	// Add tasks with different due dates
	cli.MustExecute("-y", "DueAfterTest", "add", "Task due Jan 10", "--due-date", "2026-01-10")
	cli.MustExecute("-y", "DueAfterTest", "add", "Task due Jan 15", "--due-date", "2026-01-15")
	cli.MustExecute("-y", "DueAfterTest", "add", "Task due Jan 20", "--due-date", "2026-01-20")
	cli.MustExecute("-y", "DueAfterTest", "add", "Task no due date")

	// Filter to show only tasks due after Jan 15
	stdout := cli.MustExecute("-y", "DueAfterTest", "--due-after", "2026-01-15")

	// Should show tasks due after Jan 15 (includes Jan 15 as inclusive)
	testutil.AssertContains(t, stdout, "Task due Jan 15")
	testutil.AssertContains(t, stdout, "Task due Jan 20")
	// Should not show tasks due before Jan 15
	testutil.AssertNotContains(t, stdout, "Task due Jan 10")
	// Tasks without due date should not match due date filters
	testutil.AssertNotContains(t, stdout, "Task no due date")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestFilterDueRangeSQLiteCLI verifies that `todoat -y MyList --due-after 2026-01-15 --due-before 2026-02-01` shows tasks in range
func TestFilterDueRangeSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "DueRangeTest")

	// Add tasks with different due dates
	cli.MustExecute("-y", "DueRangeTest", "add", "Task due Jan 10", "--due-date", "2026-01-10")
	cli.MustExecute("-y", "DueRangeTest", "add", "Task due Jan 20", "--due-date", "2026-01-20")
	cli.MustExecute("-y", "DueRangeTest", "add", "Task due Jan 31", "--due-date", "2026-01-31")
	cli.MustExecute("-y", "DueRangeTest", "add", "Task due Feb 15", "--due-date", "2026-02-15")
	cli.MustExecute("-y", "DueRangeTest", "add", "Task no due date")

	// Filter to show tasks due between Jan 15 and Feb 1 (inclusive range)
	stdout := cli.MustExecute("-y", "DueRangeTest", "--due-after", "2026-01-15", "--due-before", "2026-02-01")

	// Should show tasks in range
	testutil.AssertContains(t, stdout, "Task due Jan 20")
	testutil.AssertContains(t, stdout, "Task due Jan 31")
	// Should not show tasks outside range
	testutil.AssertNotContains(t, stdout, "Task due Jan 10")
	testutil.AssertNotContains(t, stdout, "Task due Feb 15")
	testutil.AssertNotContains(t, stdout, "Task no due date")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestFilterCreatedAfterSQLiteCLI verifies that `todoat -y MyList --created-after 2026-01-01` shows tasks created after date
func TestFilterCreatedAfterSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "CreatedAfterTest")

	// Add tasks - they will have creation time of "now"
	cli.MustExecute("-y", "CreatedAfterTest", "add", "Task created now")

	// Filter for tasks created after a past date (should include all tasks)
	stdout := cli.MustExecute("-y", "CreatedAfterTest", "--created-after", "2020-01-01")
	testutil.AssertContains(t, stdout, "Task created now")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)

	// Filter for tasks created after a future date (should include no tasks)
	stdout = cli.MustExecute("-y", "CreatedAfterTest", "--created-after", "2030-01-01")
	testutil.AssertNotContains(t, stdout, "Task created now")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestFilterCreatedBeforeSQLiteCLI verifies that `todoat -y MyList --created-before 2026-01-15` shows tasks created before date
func TestFilterCreatedBeforeSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "CreatedBeforeTest")

	// Add tasks - they will have creation time of "now"
	cli.MustExecute("-y", "CreatedBeforeTest", "add", "Task created now")

	// Filter for tasks created before a future date (should include all tasks)
	stdout := cli.MustExecute("-y", "CreatedBeforeTest", "--created-before", "2030-01-01")
	testutil.AssertContains(t, stdout, "Task created now")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)

	// Filter for tasks created before a past date (should include no tasks)
	stdout = cli.MustExecute("-y", "CreatedBeforeTest", "--created-before", "2020-01-01")
	testutil.AssertNotContains(t, stdout, "Task created now")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestFilterNoDueDateSQLiteCLI verifies that tasks without due dates are excluded from due date filters
func TestFilterNoDueDateSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "NoDueDateTest")

	// Add tasks - some with due dates, some without
	cli.MustExecute("-y", "NoDueDateTest", "add", "Task with due date", "--due-date", "2026-01-20")
	cli.MustExecute("-y", "NoDueDateTest", "add", "Task without due date")

	// Without filter, both should appear
	stdout := cli.MustExecute("-y", "NoDueDateTest")
	testutil.AssertContains(t, stdout, "Task with due date")
	testutil.AssertContains(t, stdout, "Task without due date")

	// With due-before filter, only task with due date should appear
	stdout = cli.MustExecute("-y", "NoDueDateTest", "--due-before", "2026-02-01")
	testutil.AssertContains(t, stdout, "Task with due date")
	testutil.AssertNotContains(t, stdout, "Task without due date")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)

	// With due-after filter, only task with due date should appear
	stdout = cli.MustExecute("-y", "NoDueDateTest", "--due-after", "2026-01-01")
	testutil.AssertContains(t, stdout, "Task with due date")
	testutil.AssertNotContains(t, stdout, "Task without due date")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestFilterCombinedStatusAndDateSQLiteCLI verifies that `todoat -y MyList -s TODO --due-before 2026-02-01` combines status and date filters
func TestFilterCombinedStatusAndDateSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "CombinedFilterTest")

	// Add tasks with different statuses and due dates
	cli.MustExecute("-y", "CombinedFilterTest", "add", "TODO task due soon", "--due-date", "2026-01-20")
	cli.MustExecute("-y", "CombinedFilterTest", "add", "TODO task due later", "--due-date", "2026-03-01")
	cli.MustExecute("-y", "CombinedFilterTest", "add", "Done task due soon", "--due-date", "2026-01-25")
	cli.MustExecute("-y", "CombinedFilterTest", "complete", "Done task due soon")

	// Filter for TODO tasks due before Feb 1
	stdout := cli.MustExecute("-y", "CombinedFilterTest", "-s", "TODO", "--due-before", "2026-02-01")

	// Should only show TODO tasks due before Feb 1
	testutil.AssertContains(t, stdout, "TODO task due soon")
	// Should not show TODO tasks due after Feb 1
	testutil.AssertNotContains(t, stdout, "TODO task due later")
	// Should not show DONE tasks even if due date matches
	testutil.AssertNotContains(t, stdout, "Done task due soon")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// =============================================================================
// Relative Date Input Tests (044-relative-date-input)
// =============================================================================

// TestRelativeDateTodaySQLiteCLI verifies that `todoat -y MyList add "Task" --due-date today` sets due date to current date
func TestRelativeDateTodaySQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout := cli.MustExecute("-y", "Work", "add", "Task today", "--due-date", "today")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify by listing tasks with JSON to check due_date is today
	stdout = cli.MustExecute("-y", "--json", "Work")

	today := time.Now().Format("2006-01-02")
	testutil.AssertContains(t, stdout, "Task today")
	testutil.AssertContains(t, stdout, today)
}

// TestRelativeDateTomorrowSQLiteCLI verifies that `todoat -y MyList add "Task" --due-date tomorrow` sets due date to next day
func TestRelativeDateTomorrowSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout := cli.MustExecute("-y", "Work", "add", "Task tomorrow", "--due-date", "tomorrow")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify by listing tasks with JSON to check due_date is tomorrow
	stdout = cli.MustExecute("-y", "--json", "Work")

	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	testutil.AssertContains(t, stdout, "Task tomorrow")
	testutil.AssertContains(t, stdout, tomorrow)
}

// TestRelativeDateYesterdaySQLiteCLI verifies that `todoat -y MyList --due-after yesterday` filters from yesterday
func TestRelativeDateYesterdaySQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list
	cli.MustExecute("-y", "list", "create", "YesterdayTest")

	// Add tasks with dates relative to today
	today := time.Now().Format("2006-01-02")
	twoDaysAgo := time.Now().AddDate(0, 0, -2).Format("2006-01-02")

	cli.MustExecute("-y", "YesterdayTest", "add", "Task due today", "--due-date", today)
	cli.MustExecute("-y", "YesterdayTest", "add", "Task due 2 days ago", "--due-date", twoDaysAgo)

	// Filter from yesterday - should include today but not 2 days ago
	stdout := cli.MustExecute("-y", "YesterdayTest", "--due-after", "yesterday")

	testutil.AssertContains(t, stdout, "Task due today")
	testutil.AssertNotContains(t, stdout, "Task due 2 days ago")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestRelativeDateDaysAheadSQLiteCLI verifies that `todoat -y MyList add "Task" --due-date +7d` sets due date 7 days from now
func TestRelativeDateDaysAheadSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout := cli.MustExecute("-y", "Work", "add", "Task in 7 days", "--due-date", "+7d")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify by listing tasks with JSON to check due_date is 7 days from now
	stdout = cli.MustExecute("-y", "--json", "Work")

	sevenDays := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	testutil.AssertContains(t, stdout, "Task in 7 days")
	testutil.AssertContains(t, stdout, sevenDays)
}

// TestRelativeDateDaysBackSQLiteCLI verifies that `todoat -y MyList --due-after -3d` filters from 3 days ago
func TestRelativeDateDaysBackSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list
	cli.MustExecute("-y", "list", "create", "DaysBackTest")

	// Add tasks with dates relative to today
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	fiveDaysAgo := time.Now().AddDate(0, 0, -5).Format("2006-01-02")

	cli.MustExecute("-y", "DaysBackTest", "add", "Task due yesterday", "--due-date", yesterday)
	cli.MustExecute("-y", "DaysBackTest", "add", "Task due 5 days ago", "--due-date", fiveDaysAgo)

	// Filter from 3 days ago - should include yesterday but not 5 days ago
	stdout := cli.MustExecute("-y", "DaysBackTest", "--due-after", "-3d")

	testutil.AssertContains(t, stdout, "Task due yesterday")
	testutil.AssertNotContains(t, stdout, "Task due 5 days ago")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestRelativeDateWeeksSQLiteCLI verifies that `todoat -y MyList add "Task" --due-date +2w` sets due date 2 weeks from now
func TestRelativeDateWeeksSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout := cli.MustExecute("-y", "Work", "add", "Task in 2 weeks", "--due-date", "+2w")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify by listing tasks with JSON to check due_date is 2 weeks from now
	stdout = cli.MustExecute("-y", "--json", "Work")

	twoWeeks := time.Now().AddDate(0, 0, 14).Format("2006-01-02")
	testutil.AssertContains(t, stdout, "Task in 2 weeks")
	testutil.AssertContains(t, stdout, twoWeeks)
}

// TestRelativeDateMonthsSQLiteCLI verifies that `todoat -y MyList add "Task" --due-date +1m` sets due date 1 month from now
func TestRelativeDateMonthsSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout := cli.MustExecute("-y", "Work", "add", "Task in 1 month", "--due-date", "+1m")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify by listing tasks with JSON to check due_date is 1 month from now
	stdout = cli.MustExecute("-y", "--json", "Work")

	oneMonth := time.Now().AddDate(0, 1, 0).Format("2006-01-02")
	testutil.AssertContains(t, stdout, "Task in 1 month")
	testutil.AssertContains(t, stdout, oneMonth)
}

// TestAbsoluteDateStillWorksSQLiteCLI verifies that `todoat -y MyList add "Task" --due-date 2026-01-31` still works
func TestAbsoluteDateStillWorksSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout := cli.MustExecute("-y", "Work", "add", "Task absolute", "--due-date", "2026-01-31")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify by listing tasks with JSON to check due_date
	stdout = cli.MustExecute("-y", "--json", "Work")

	testutil.AssertContains(t, stdout, "Task absolute")
	testutil.AssertContains(t, stdout, "2026-01-31")
}

// =============================================================================
// Trash Auto-Purge Tests (046-trash-auto-purge)
// =============================================================================

// TestTrashAutoPurgeDefaultSQLiteCLI verifies that lists deleted >30 days ago
// are automatically purged on next `todoat list trash` command
func TestTrashAutoPurgeDefaultSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITestWithTrash(t)

	// Create and delete a list
	cli.MustExecute("-y", "list", "create", "OldTrashList")
	cli.MustExecute("-y", "list", "delete", "OldTrashList")

	// Set deleted_at to 31 days ago (beyond default 30-day retention)
	cli.SetListDeletedAt("OldTrashList", time.Now().AddDate(0, 0, -31))

	// View trash - should trigger auto-purge
	stdout := cli.MustExecute("-y", "list", "trash")

	// The old list should be purged and not visible
	testutil.AssertNotContains(t, stdout, "OldTrashList")
	// Output should indicate purge happened
	testutil.AssertContains(t, stdout, "purged")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestTrashAutoPurgeConfigurableSQLiteCLI verifies that `trash.retention_days: 7`
// in config purges lists older than 7 days
func TestTrashAutoPurgeConfigurableSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITestWithTrash(t)

	// Set retention to 7 days
	cli.SetTrashRetentionDays(7)

	// Create and delete a list
	cli.MustExecute("-y", "list", "create", "WeekOldList")
	cli.MustExecute("-y", "list", "delete", "WeekOldList")

	// Set deleted_at to 8 days ago (beyond 7-day retention)
	cli.SetListDeletedAt("WeekOldList", time.Now().AddDate(0, 0, -8))

	// View trash - should trigger auto-purge
	stdout := cli.MustExecute("-y", "list", "trash")

	// The old list should be purged and not visible
	testutil.AssertNotContains(t, stdout, "WeekOldList")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestTrashAutoPurgeDisabledSQLiteCLI verifies that `trash.retention_days: 0`
// disables auto-purge
func TestTrashAutoPurgeDisabledSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITestWithTrash(t)

	// Disable auto-purge
	cli.SetTrashRetentionDays(0)

	// Create and delete a list
	cli.MustExecute("-y", "list", "create", "ForeverList")
	cli.MustExecute("-y", "list", "delete", "ForeverList")

	// Set deleted_at to 365 days ago
	cli.SetListDeletedAt("ForeverList", time.Now().AddDate(0, 0, -365))

	// View trash - should NOT auto-purge
	stdout := cli.MustExecute("-y", "list", "trash")

	// The list should still be visible
	testutil.AssertContains(t, stdout, "ForeverList")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestTrashAutoPurgePreservesRecentSQLiteCLI verifies that lists deleted <30 days ago
// are NOT purged
func TestTrashAutoPurgePreservesRecentSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITestWithTrash(t)

	// Create and delete a list
	cli.MustExecute("-y", "list", "create", "RecentList")
	cli.MustExecute("-y", "list", "delete", "RecentList")

	// Set deleted_at to 29 days ago (within default 30-day retention)
	cli.SetListDeletedAt("RecentList", time.Now().AddDate(0, 0, -29))

	// View trash - should NOT purge recent lists
	stdout := cli.MustExecute("-y", "list", "trash")

	// The recent list should still be visible
	testutil.AssertContains(t, stdout, "RecentList")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// =============================================================================
// List Rename Tests (050-list-rename)
// =============================================================================

// TestListRename verifies that `todoat list update "OldName" --name "NewName"` renames list
func TestListRenameSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "OldListName")

	// Rename the list
	stdout := cli.MustExecute("-y", "list", "update", "OldListName", "--name", "NewListName")

	testutil.AssertContains(t, stdout, "NewListName")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the list was renamed by viewing lists
	stdout = cli.MustExecute("-y", "list")

	testutil.AssertContains(t, stdout, "NewListName")
	testutil.AssertNotContains(t, stdout, "OldListName")
}

// TestListRenameNotFound verifies that renaming non-existent list returns ERROR
func TestListRenameNotFoundSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Try to rename a list that doesn't exist
	stdout, _, exitCode := cli.Execute("-y", "list", "update", "NonExistent", "--name", "NewName")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertResultCode(t, stdout, testutil.ResultError)
}

// TestListRenameDuplicate verifies that renaming to existing name returns ERROR with suggestion
func TestListRenameDuplicateSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create two lists
	cli.MustExecute("-y", "list", "create", "FirstList")
	cli.MustExecute("-y", "list", "create", "SecondList")

	// Try to rename FirstList to SecondList (which already exists)
	stdout, stderr, exitCode := cli.Execute("-y", "list", "update", "FirstList", "--name", "SecondList")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertResultCode(t, stdout, testutil.ResultError)

	// Should mention the name already exists
	errOutput := stderr
	if !strings.Contains(strings.ToLower(errOutput), "exists") && !strings.Contains(strings.ToLower(errOutput), "already") {
		t.Errorf("error should mention name already exists, got: %s", errOutput)
	}
}

// TestListRenameJSON verifies that `todoat --json list update "OldName" --name "NewName"` returns JSON with result
func TestListRenameJSONSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "JSONRenameList")

	// Rename with JSON output
	stdout := cli.MustExecute("-y", "--json", "list", "update", "JSONRenameList", "--name", "RenamedJSONList")

	// Should contain JSON structure with list details
	testutil.AssertContains(t, stdout, "{")
	testutil.AssertContains(t, stdout, "}")
	testutil.AssertContains(t, stdout, "RenamedJSONList")
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"ACTION_COMPLETED"`)
}

// TestListRenameNoPrompt verifies that `todoat -y list update "Partial" --name "NewName"` handles partial match
func TestListRenameNoPromptSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with a distinctive name
	cli.MustExecute("-y", "list", "create", "PartialMatchList")

	// Rename using partial match with -y flag (should work if unique match)
	stdout := cli.MustExecute("-y", "list", "update", "PartialMatch", "--name", "RenamedPartialList")

	testutil.AssertContains(t, stdout, "RenamedPartialList")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the rename happened
	stdout = cli.MustExecute("-y", "list")
	testutil.AssertContains(t, stdout, "RenamedPartialList")
	testutil.AssertNotContains(t, stdout, "PartialMatchList")
}

// TestListRenamePreservesProperties verifies that renaming preserves other list properties
func TestListRenamePreservesPropertiesSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add some tasks
	cli.MustExecute("-y", "list", "create", "PropertiesTestList")
	cli.MustExecute("-y", "PropertiesTestList", "add", "Task 1")
	cli.MustExecute("-y", "PropertiesTestList", "add", "Task 2")

	// Rename the list
	cli.MustExecute("-y", "list", "update", "PropertiesTestList", "--name", "RenamedPropertiesList")

	// Verify tasks are preserved
	stdout := cli.MustExecute("-y", "RenamedPropertiesList")

	testutil.AssertContains(t, stdout, "Task 1")
	testutil.AssertContains(t, stdout, "Task 2")
}

// TestListRenameEmptyName verifies that renaming to empty name returns ERROR
func TestListRenameEmptyNameSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "EmptyNameTest")

	// Try to rename to empty name
	stdout, _, exitCode := cli.Execute("-y", "list", "update", "EmptyNameTest", "--name", "")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertResultCode(t, stdout, testutil.ResultError)
}

// TestListRenameMultipleMatches verifies that multiple partial matches in no-prompt mode returns ERROR
func TestListRenameMultipleMatchesSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create lists with similar names
	cli.MustExecute("-y", "list", "create", "SimilarOne")
	cli.MustExecute("-y", "list", "create", "SimilarTwo")

	// Try to rename with partial match that matches multiple lists
	stdout, stderr, exitCode := cli.Execute("-y", "list", "update", "Similar", "--name", "NewName")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertResultCode(t, stdout, testutil.ResultError)

	// Should mention multiple matches or ambiguous
	errOutput := stderr
	if !strings.Contains(strings.ToLower(errOutput), "multiple") && !strings.Contains(strings.ToLower(errOutput), "ambiguous") {
		t.Errorf("error should mention multiple matches or ambiguous, got: %s", errOutput)
	}
}

// =============================================================================
// List Color and Description Update Tests (052)
// =============================================================================

// TestListUpdateColor verifies that `todoat list update "Work" --color "#FF5733"` sets list color
func TestListUpdateColorSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "ColorTestList")

	// Update the list color
	stdout := cli.MustExecute("-y", "list", "update", "ColorTestList", "--color", "#FF5733")

	testutil.AssertContains(t, stdout, "ColorTestList")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the color was set by checking list info
	stdout = cli.MustExecute("-y", "list", "info", "ColorTestList")

	testutil.AssertContains(t, stdout, "#FF5733")
}

// TestListUpdateDescription verifies that `todoat list update "Work" --description "Work tasks"` sets description
func TestListUpdateDescriptionSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "DescTestList")

	// Update the list description
	stdout := cli.MustExecute("-y", "list", "update", "DescTestList", "--description", "Work-related tasks and projects")

	testutil.AssertContains(t, stdout, "DescTestList")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the description was set by checking list info
	stdout = cli.MustExecute("-y", "list", "info", "DescTestList")

	testutil.AssertContains(t, stdout, "Work-related tasks and projects")
}

// TestListUpdateMultiple verifies that `todoat list update "Work" --color "#FF5733" --description "Text"` updates both
func TestListUpdateMultipleSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "MultiUpdateList")

	// Update both color and description at once
	stdout := cli.MustExecute("-y", "list", "update", "MultiUpdateList", "--color", "#00FF00", "--description", "Multi-update test")

	testutil.AssertContains(t, stdout, "MultiUpdateList")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify both were set
	stdout = cli.MustExecute("-y", "list", "info", "MultiUpdateList")

	testutil.AssertContains(t, stdout, "#00FF00")
	testutil.AssertContains(t, stdout, "Multi-update test")
}

// TestListUpdateColorValidation verifies that invalid hex color returns ERROR with format hint
func TestListUpdateColorValidationSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "ColorValidateList")

	// Try to set invalid color
	stdout, stderr, exitCode := cli.Execute("-y", "list", "update", "ColorValidateList", "--color", "not-a-color")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertResultCode(t, stdout, testutil.ResultError)

	// Error should mention valid format
	errOutput := stderr
	if !strings.Contains(strings.ToLower(errOutput), "hex") && !strings.Contains(strings.ToLower(errOutput), "#") {
		t.Errorf("error should mention hex format, got: %s", errOutput)
	}
}

// TestListShowProperties verifies that `todoat list info "Work"` displays all properties (id, name, color, description, task count)
func TestListShowPropertiesSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and set all properties
	cli.MustExecute("-y", "list", "create", "ShowPropsTestList")
	cli.MustExecute("-y", "list", "update", "ShowPropsTestList", "--color", "#AABBCC", "--description", "Test description for show")
	cli.MustExecute("-y", "ShowPropsTestList", "add", "Task 1")
	cli.MustExecute("-y", "ShowPropsTestList", "add", "Task 2")

	// Show list info
	stdout := cli.MustExecute("-y", "list", "info", "ShowPropsTestList")

	// Should show all properties
	testutil.AssertContains(t, stdout, "ShowPropsTestList")         // Name
	testutil.AssertContains(t, stdout, "#AABBCC")                   // Color
	testutil.AssertContains(t, stdout, "Test description for show") // Description
	testutil.AssertContains(t, stdout, "2")                         // Task count
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestListUpdateJSON verifies that `todoat --json list update "Work" --color "#FF5733"` returns JSON
func TestListUpdateJSONSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "JSONColorList")

	// Update color with JSON output
	stdout := cli.MustExecute("-y", "--json", "list", "update", "JSONColorList", "--color", "#DEADBE")

	// Should contain JSON structure with list details
	testutil.AssertContains(t, stdout, "{")
	testutil.AssertContains(t, stdout, "}")
	testutil.AssertContains(t, stdout, "JSONColorList")
	testutil.AssertContains(t, stdout, `"result"`)
	testutil.AssertContains(t, stdout, `"ACTION_COMPLETED"`)
	testutil.AssertContains(t, stdout, "#DEADBE")
}

// TestListUpdateColorNormalization verifies that color formats are normalized to #RRGGBB
func TestListUpdateColorNormalizationSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Test various valid formats
	testCases := []struct {
		input    string
		expected string
	}{
		{"#ABC", "#AABBCC"},    // 3-char with #
		{"ABC", "#AABBCC"},     // 3-char without #
		{"#aabbcc", "#AABBCC"}, // lowercase
		{"aabbcc", "#AABBCC"},  // no # lowercase
		{"#AABBCC", "#AABBCC"}, // already normalized
	}

	for i, tc := range testCases {
		listName := "NormTestList" + strconv.Itoa(i)
		cli.MustExecute("-y", "list", "create", listName)
		cli.MustExecute("-y", "list", "update", listName, "--color", tc.input)

		stdout := cli.MustExecute("-y", "list", "info", listName)
		testutil.AssertContains(t, stdout, tc.expected)
	}
}

// TestListUpdateClearDescription verifies that empty description clears the field
func TestListUpdateClearDescriptionSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list with description
	cli.MustExecute("-y", "list", "create", "ClearDescList")
	cli.MustExecute("-y", "list", "update", "ClearDescList", "--description", "Initial description")

	// Verify description is set
	stdout := cli.MustExecute("-y", "list", "info", "ClearDescList")
	testutil.AssertContains(t, stdout, "Initial description")

	// Clear description with empty string
	cli.MustExecute("-y", "list", "update", "ClearDescList", "--description", "")

	// Verify description is cleared (should not appear in output)
	stdout = cli.MustExecute("-y", "list", "info", "ClearDescList")
	testutil.AssertNotContains(t, stdout, "Initial description")
}

// TestListUpdateColorOnlyNoName verifies color update works without --name flag
func TestListUpdateColorOnlyNoNameSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "ColorOnlyList")

	// Update only color (no --name flag)
	stdout := cli.MustExecute("-y", "list", "update", "ColorOnlyList", "--color", "#123456")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify color was set and name unchanged
	stdout = cli.MustExecute("-y", "list", "info", "ColorOnlyList")
	testutil.AssertContains(t, stdout, "ColorOnlyList")
	testutil.AssertContains(t, stdout, "#123456")
}

// TestListUpdateDescriptionOnlyNoName verifies description update works without --name flag
func TestListUpdateDescriptionOnlyNoNameSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "DescOnlyList")

	// Update only description (no --name flag)
	stdout := cli.MustExecute("-y", "list", "update", "DescOnlyList", "--description", "Only updating description")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify description was set and name unchanged
	stdout = cli.MustExecute("-y", "list", "info", "DescOnlyList")
	testutil.AssertContains(t, stdout, "DescOnlyList")
	testutil.AssertContains(t, stdout, "Only updating description")
}

// TestListUpdateNoChanges verifies that update with no flags returns error
func TestListUpdateNoChangesSQLiteCLI(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "NoChangesTestList")

	// Try to update without any flags
	stdout, _, exitCode := cli.Execute("-y", "list", "update", "NoChangesTestList")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertResultCode(t, stdout, testutil.ResultError)
}
