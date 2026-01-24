package views_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// =============================================================================
// Pagination Tests (076-task-pagination)
// =============================================================================

// TestPaginationDefault verifies that `todoat MyList` shows first page with default page size
// For small task sets (<100 tasks), all tasks should be shown (default behavior unchanged)
func TestPaginationDefault(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add 10 tasks
	cli.MustExecute("-y", "list", "create", "PaginationDefaultTest")
	for i := 1; i <= 10; i++ {
		cli.MustExecute("-y", "PaginationDefaultTest", "add", "Task "+string(rune('A'+i-1)))
	}

	// List tasks without pagination flags - should show all tasks (default behavior)
	stdout, _, exitCode := cli.Execute("-y", "PaginationDefaultTest")

	testutil.AssertExitCode(t, exitCode, 0)
	// All 10 tasks should be visible
	testutil.AssertContains(t, stdout, "Task A")
	testutil.AssertContains(t, stdout, "Task J")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPaginationWithLimit verifies that `todoat MyList --limit 20` limits output to 20 tasks
func TestPaginationWithLimit(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add 30 tasks
	cli.MustExecute("-y", "list", "create", "PaginationLimitTest")
	for i := 1; i <= 30; i++ {
		cli.MustExecute("-y", "PaginationLimitTest", "add", "Task "+padNumber(i))
	}

	// List tasks with limit 10
	stdout, _, exitCode := cli.Execute("-y", "PaginationLimitTest", "--limit", "10")

	testutil.AssertExitCode(t, exitCode, 0)
	// Should show pagination info
	testutil.AssertContains(t, stdout, "Showing 1-10 of 30 tasks")
	// Should not show task beyond limit (tasks are shown in order)
	// Count how many "Task " occurrences there are in the output
	taskCount := strings.Count(stdout, "Task ")
	if taskCount > 10 {
		t.Errorf("expected at most 10 tasks, got %d", taskCount)
	}
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPaginationWithOffset verifies that `todoat MyList --offset 20 --limit 20` shows second page
func TestPaginationWithOffset(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add 50 tasks
	cli.MustExecute("-y", "list", "create", "PaginationOffsetTest")
	for i := 1; i <= 50; i++ {
		cli.MustExecute("-y", "PaginationOffsetTest", "add", "Task "+padNumber(i))
	}

	// List tasks with offset 20 and limit 20
	stdout, _, exitCode := cli.Execute("-y", "PaginationOffsetTest", "--offset", "20", "--limit", "20")

	testutil.AssertExitCode(t, exitCode, 0)
	// Should show pagination info for second page
	testutil.AssertContains(t, stdout, "Showing 21-40 of 50 tasks")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPaginationPageFlag verifies that `todoat MyList --page 2` shows second page
func TestPaginationPageFlag(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add 100 tasks
	cli.MustExecute("-y", "list", "create", "PaginationPageTest")
	for i := 1; i <= 100; i++ {
		cli.MustExecute("-y", "PaginationPageTest", "add", "Task "+padNumber(i))
	}

	// List tasks page 2 with default page size (50)
	stdout, _, exitCode := cli.Execute("-y", "PaginationPageTest", "--page", "2")

	testutil.AssertExitCode(t, exitCode, 0)
	// Should show pagination info for second page
	testutil.AssertContains(t, stdout, "Showing 51-100 of 100 tasks")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPaginationWithFilters verifies that pagination works correctly with view filters
func TestPaginationWithFilters(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add tasks with different statuses
	cli.MustExecute("-y", "list", "create", "PaginationFilterTest")
	for i := 1; i <= 20; i++ {
		cli.MustExecute("-y", "PaginationFilterTest", "add", "TODO Task "+padNumber(i))
	}
	for i := 1; i <= 10; i++ {
		cli.MustExecute("-y", "PaginationFilterTest", "add", "Done Task "+padNumber(i))
		cli.MustExecute("-y", "PaginationFilterTest", "complete", "Done Task "+padNumber(i))
	}

	// List only TODO tasks with limit
	stdout, _, exitCode := cli.Execute("-y", "PaginationFilterTest", "--status", "TODO", "--limit", "5")

	testutil.AssertExitCode(t, exitCode, 0)
	// Should show pagination info reflecting filtered count (20 TODO tasks)
	testutil.AssertContains(t, stdout, "Showing 1-5 of 20 tasks")
	// Should not contain completed tasks
	testutil.AssertNotContains(t, stdout, "Done Task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPaginationWithSort verifies that pagination preserves sort order
func TestPaginationWithSort(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViews(t)

	// Create a custom view with priority sort (ascending - lower priority number = higher priority)
	viewYAML := `name: priority-sorted
fields:
  - name: priority
  - name: summary
sort:
  - field: priority
    direction: asc
`
	if err := os.WriteFile(viewsDir+"/priority-sorted.yaml", []byte(viewYAML), 0644); err != nil {
		t.Fatalf("failed to write view file: %v", err)
	}

	// Create a list and add tasks with different priorities
	cli.MustExecute("-y", "list", "create", "PaginationSortTest")
	cli.MustExecute("-y", "PaginationSortTest", "add", "Low priority task", "-p", "9")
	cli.MustExecute("-y", "PaginationSortTest", "add", "High priority task", "-p", "1")
	cli.MustExecute("-y", "PaginationSortTest", "add", "Medium priority task", "-p", "5")

	// List tasks with view sort and limit
	stdout, stderr, exitCode := cli.Execute("-y", "PaginationSortTest", "-v", "priority-sorted", "--limit", "2")
	if exitCode != 0 {
		t.Logf("stderr: %s", stderr)
	}
	testutil.AssertExitCode(t, exitCode, 0)
	// Should show pagination info
	testutil.AssertContains(t, stdout, "Showing 1-2 of 3 tasks")
	// High priority (p1) should appear first (top 2 should not include low priority p9)
	testutil.AssertContains(t, stdout, "High priority task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPaginationTotalCount verifies that output includes total task count for UI navigation
func TestPaginationTotalCount(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add tasks
	cli.MustExecute("-y", "list", "create", "PaginationCountTest")
	for i := 1; i <= 75; i++ {
		cli.MustExecute("-y", "PaginationCountTest", "add", "Task "+padNumber(i))
	}

	// List tasks with pagination
	stdout, _, exitCode := cli.Execute("-y", "PaginationCountTest", "--limit", "25")

	testutil.AssertExitCode(t, exitCode, 0)
	// Should show total count in footer
	testutil.AssertContains(t, stdout, "of 75 tasks")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestPaginationJSONMetadata verifies that JSON output includes pagination metadata
func TestPaginationJSONMetadata(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add tasks
	cli.MustExecute("-y", "list", "create", "PaginationJSONTest")
	for i := 1; i <= 60; i++ {
		cli.MustExecute("-y", "PaginationJSONTest", "add", "Task "+padNumber(i))
	}

	// List tasks with pagination in JSON format
	stdout, _, exitCode := cli.Execute("-y", "PaginationJSONTest", "--limit", "20", "--json")

	testutil.AssertExitCode(t, exitCode, 0)

	// Parse JSON response
	var response struct {
		Tasks    []interface{} `json:"tasks"`
		List     string        `json:"list"`
		Count    int           `json:"count"`
		Total    int           `json:"total"`
		Page     int           `json:"page"`
		PageSize int           `json:"page_size"`
		HasMore  bool          `json:"has_more"`
		Result   string        `json:"result"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &response); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, stdout)
	}

	// Verify pagination metadata
	if response.Total != 60 {
		t.Errorf("expected total=60, got %d", response.Total)
	}
	if response.Page != 1 {
		t.Errorf("expected page=1, got %d", response.Page)
	}
	if response.PageSize != 20 {
		t.Errorf("expected page_size=20, got %d", response.PageSize)
	}
	if !response.HasMore {
		t.Errorf("expected has_more=true, got false")
	}
	if len(response.Tasks) != 20 {
		t.Errorf("expected 20 tasks in response, got %d", len(response.Tasks))
	}
}

// TestPaginationPageSize verifies that `--page-size N` configures page size
func TestPaginationPageSize(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list and add tasks
	cli.MustExecute("-y", "list", "create", "PaginationPageSizeTest")
	for i := 1; i <= 50; i++ {
		cli.MustExecute("-y", "PaginationPageSizeTest", "add", "Task "+padNumber(i))
	}

	// List page 2 with custom page size of 15
	stdout, _, exitCode := cli.Execute("-y", "PaginationPageSizeTest", "--page", "2", "--page-size", "15")

	testutil.AssertExitCode(t, exitCode, 0)
	// Should show pagination info for page 2 with 15 items per page
	testutil.AssertContains(t, stdout, "Showing 16-30 of 50 tasks")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// padNumber pads a number to 3 digits for consistent sorting (e.g., 001, 002, ...)
func padNumber(n int) string {
	return string([]byte{
		byte('0' + n/100%10),
		byte('0' + n/10%10),
		byte('0' + n%10),
	})
}
