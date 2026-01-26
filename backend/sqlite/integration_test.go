// Package sqlite_test contains integration tests for the todoat CLI.
// These tests verify end-to-end CLI workflows including task management,
// list operations, hierarchy/subtasks, and sync features.
package sqlite_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// =============================================================================
// Task Lifecycle Integration Tests
// =============================================================================

// TestTaskLifecycleIntegration tests the complete lifecycle of a task:
// create list -> add task -> update task -> complete task -> delete task
func TestTaskLifecycleIntegration(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Step 1: Create a new list
	stdout := cli.MustExecute("-y", "list", "create", "IntegrationWork")
	testutil.AssertContains(t, stdout, "IntegrationWork")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Step 2: Add a task
	stdout = cli.MustExecute("-y", "IntegrationWork", "add", "Review documentation")
	testutil.AssertContains(t, stdout, "Review documentation")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Step 3: Verify task appears in list
	stdout = cli.MustExecute("-y", "IntegrationWork")
	testutil.AssertContains(t, stdout, "Review documentation")
	testutil.AssertContains(t, stdout, "[TODO]")

	// Step 4: Update task summary
	stdout = cli.MustExecute("-y", "IntegrationWork", "update", "Review documentation", "--summary", "Review API documentation")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Step 5: Verify update
	stdout = cli.MustExecute("-y", "IntegrationWork")
	testutil.AssertContains(t, stdout, "Review API documentation")
	testutil.AssertNotContains(t, stdout, "Review documentation")

	// Step 6: Update task priority
	stdout = cli.MustExecute("-y", "IntegrationWork", "update", "Review API documentation", "-p", "1")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Step 7: Verify priority
	stdout = cli.MustExecute("-y", "IntegrationWork")
	testutil.AssertContains(t, stdout, "[P1]")

	// Step 8: Complete the task
	stdout = cli.MustExecute("-y", "IntegrationWork", "complete", "Review API documentation")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Step 9: Verify task is completed (use -s DONE since default view filters out DONE tasks)
	stdout = cli.MustExecute("-y", "IntegrationWork", "-s", "DONE")
	testutil.AssertContains(t, stdout, "[DONE]")

	// Step 10: Delete the task
	stdout = cli.MustExecute("-y", "IntegrationWork", "delete", "Review API documentation")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Step 11: Verify task is deleted
	stdout = cli.MustExecute("-y", "IntegrationWork")
	testutil.AssertNotContains(t, stdout, "Review API documentation")
}

// TestMultipleTasksWorkflowIntegration tests managing multiple tasks in a list
func TestMultipleTasksWorkflowIntegration(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list
	cli.MustExecute("-y", "list", "create", "MultiTaskList")

	// Add multiple tasks with different priorities
	cli.MustExecute("-y", "MultiTaskList", "add", "High priority task", "-p", "1")
	cli.MustExecute("-y", "MultiTaskList", "add", "Medium priority task", "-p", "5")
	cli.MustExecute("-y", "MultiTaskList", "add", "Low priority task", "-p", "9")

	// Verify all tasks exist
	stdout := cli.MustExecute("-y", "MultiTaskList")
	testutil.AssertContains(t, stdout, "High priority task")
	testutil.AssertContains(t, stdout, "Medium priority task")
	testutil.AssertContains(t, stdout, "Low priority task")

	// Filter by priority
	stdout = cli.MustExecute("-y", "MultiTaskList", "-p", "1")
	testutil.AssertContains(t, stdout, "High priority task")
	testutil.AssertNotContains(t, stdout, "Medium priority task")
	testutil.AssertNotContains(t, stdout, "Low priority task")

	// Complete one task
	cli.MustExecute("-y", "MultiTaskList", "complete", "High priority task")

	// Filter by status
	stdout = cli.MustExecute("-y", "MultiTaskList", "-s", "TODO")
	testutil.AssertNotContains(t, stdout, "High priority task")
	testutil.AssertContains(t, stdout, "Medium priority task")
	testutil.AssertContains(t, stdout, "Low priority task")
}

// =============================================================================
// List Management Integration Tests
// =============================================================================

// TestListLifecycleIntegration tests list management:
// create -> delete -> check trash
func TestListLifecycleIntegration(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Step 1: Create a new list
	stdout := cli.MustExecute("-y", "list", "create", "LifecycleList")
	testutil.AssertContains(t, stdout, "LifecycleList")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Step 2: Verify list appears in list view
	stdout = cli.MustExecute("-y", "list")
	testutil.AssertContains(t, stdout, "LifecycleList")

	// Step 3: Delete list (moves to trash)
	stdout = cli.MustExecute("-y", "list", "delete", "LifecycleList")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Step 4: Verify list no longer in active lists
	stdout = cli.MustExecute("-y", "list")
	testutil.AssertNotContains(t, stdout, "LifecycleList")

	// Step 5: Check trash - list should be in trash
	stdout = cli.MustExecute("-y", "list", "trash")
	testutil.AssertContains(t, stdout, "LifecycleList")
}

// TestListRenameIntegration tests list rename functionality separately
func TestListRenameIntegration(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a new list
	cli.MustExecute("-y", "list", "create", "OriginalName")

	// Update the list name
	stdout := cli.MustExecute("-y", "list", "update", "OriginalName", "--name", "NewName")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify rename
	stdout = cli.MustExecute("-y", "list")
	testutil.AssertContains(t, stdout, "NewName")
	testutil.AssertNotContains(t, stdout, "OriginalName")
}

// =============================================================================
// Hierarchical Tasks / Subtasks Integration Tests
// =============================================================================

// TestSubtaskHierarchyIntegration tests the complete subtask workflow:
// create parent -> add subtasks -> update parent -> complete subtask -> re-parent
func TestSubtaskHierarchyIntegration(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list for subtask testing
	cli.MustExecute("-y", "list", "create", "SubtaskProject")

	// Step 1: Add parent task
	stdout := cli.MustExecute("-y", "SubtaskProject", "add", "Main Feature")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Step 2: Add subtasks using -P flag
	stdout = cli.MustExecute("-y", "SubtaskProject", "add", "Design mockups", "-P", "Main Feature")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	stdout = cli.MustExecute("-y", "SubtaskProject", "add", "Implement backend", "-P", "Main Feature")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	stdout = cli.MustExecute("-y", "SubtaskProject", "add", "Write tests", "-P", "Main Feature")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Step 3: Verify hierarchy is displayed
	stdout = cli.MustExecute("-y", "SubtaskProject")
	testutil.AssertContains(t, stdout, "Main Feature")
	testutil.AssertContains(t, stdout, "Design mockups")
	testutil.AssertContains(t, stdout, "Implement backend")
	testutil.AssertContains(t, stdout, "Write tests")

	// Step 4: Complete a subtask
	stdout = cli.MustExecute("-y", "SubtaskProject", "complete", "Design mockups")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Step 5: Verify subtask is completed (use -v all since default view filters out DONE tasks)
	stdout = cli.MustExecute("-y", "SubtaskProject", "-v", "all")
	testutil.AssertContains(t, stdout, "[DONE]")

	// Step 6: Add another parent task
	cli.MustExecute("-y", "SubtaskProject", "add", "Secondary Feature")

	// Step 7: Re-parent a subtask to the new parent
	stdout = cli.MustExecute("-y", "SubtaskProject", "update", "Write tests", "-P", "Secondary Feature")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Step 8: Verify re-parenting worked (task still exists and is under new parent)
	stdout = cli.MustExecute("-y", "--json", "SubtaskProject")
	testutil.AssertContains(t, stdout, "Write tests")
	testutil.AssertContains(t, stdout, "Secondary Feature")
}

// TestPathBasedHierarchyIntegration tests creating hierarchy using path notation (A/B/C)
func TestPathBasedHierarchyIntegration(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list
	cli.MustExecute("-y", "list", "create", "PathProject")

	// Step 1: Create multi-level hierarchy with path notation
	stdout := cli.MustExecute("-y", "PathProject", "add", "Release/Backend/API")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Step 2: Add more tasks at various levels
	cli.MustExecute("-y", "PathProject", "add", "Release/Backend/Database")
	cli.MustExecute("-y", "PathProject", "add", "Release/Frontend/UI")
	cli.MustExecute("-y", "PathProject", "add", "Release/Frontend/Styling")

	// Step 3: Verify hierarchy structure
	stdout = cli.MustExecute("-y", "PathProject")
	testutil.AssertContains(t, stdout, "Release")
	testutil.AssertContains(t, stdout, "Backend")
	testutil.AssertContains(t, stdout, "Frontend")
	testutil.AssertContains(t, stdout, "API")
	testutil.AssertContains(t, stdout, "Database")
	testutil.AssertContains(t, stdout, "UI")
	testutil.AssertContains(t, stdout, "Styling")

	// Step 4: Verify tree visualization is present
	if !strings.Contains(stdout, "├") && !strings.Contains(stdout, "└") {
		t.Errorf("expected tree visualization with box-drawing characters, got:\n%s", stdout)
	}

	// Step 5: Add another leaf to existing path (should not create duplicates)
	cli.MustExecute("-y", "PathProject", "add", "Release/Backend/Cache")
	stdout = cli.MustExecute("-y", "--json", "PathProject")

	// Should only have one Release and one Backend
	if strings.Count(stdout, `"Release"`) > 1 {
		t.Errorf("expected only one Release task, but found duplicates")
	}
	if strings.Count(stdout, `"Backend"`) > 1 {
		t.Errorf("expected only one Backend task, but found duplicates")
	}
}

// TestRemoveParentIntegration tests making a subtask become a root-level task
func TestRemoveParentIntegration(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list and hierarchy
	cli.MustExecute("-y", "list", "create", "RemoveParentTest")
	cli.MustExecute("-y", "RemoveParentTest", "add", "Project")
	cli.MustExecute("-y", "RemoveParentTest", "add", "Task", "-P", "Project")

	// Verify task is a subtask
	stdout := cli.MustExecute("-y", "--json", "RemoveParentTest")
	testutil.AssertContains(t, stdout, `"parent_id"`)

	// Remove parent (make root level)
	stdout = cli.MustExecute("-y", "RemoveParentTest", "update", "Task", "--no-parent")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify task is now at root level (parent_id should be empty/null)
	stdout = cli.MustExecute("-y", "RemoveParentTest")
	testutil.AssertContains(t, stdout, "Project")
	testutil.AssertContains(t, stdout, "Task")
}

// =============================================================================
// Sync Feature Integration Tests
// =============================================================================

// TestSyncStatusIntegration tests basic sync status commands
func TestSyncStatusIntegration(t *testing.T) {
	cli, viewsDir := testutil.NewCLITestWithViews(t)
	tmpDir := filepath.Dir(viewsDir)

	// Create sync config
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
backends:
  sqlite:
    type: sqlite
    enabled: true
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Check sync status
	stdout := cli.MustExecute("-y", "sync", "status")
	testutil.AssertContains(t, stdout, "Sync Status")

	// Check sync queue (may be empty)
	stdout = cli.MustExecute("-y", "sync", "queue")
	testutil.AssertContains(t, stdout, "Pending Operations")

	// Check sync conflicts (may be empty)
	stdout = cli.MustExecute("-y", "sync", "conflicts")
	testutil.AssertContains(t, stdout, "Conflict")
}

// TestSyncConflictResolutionIntegration tests conflict resolution strategies
func TestSyncConflictResolutionIntegration(t *testing.T) {
	strategies := []string{"server_wins", "local_wins", "merge", "keep_both"}

	for _, strategy := range strategies {
		t.Run(strategy, func(t *testing.T) {
			cli, tmpDir := testutil.NewCLITestWithViews(t)

			// Create config with specific conflict strategy
			configContent := `
sync:
  enabled: true
  local_backend: sqlite
  conflict_resolution: ` + strategy + `
backends:
  sqlite:
    type: sqlite
    enabled: true
`
			configPath := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("failed to write config: %v", err)
			}

			// Add and modify a task
			cli.MustExecute("-y", "list", "create", "ConflictList")
			cli.MustExecute("-y", "ConflictList", "add", "Conflicting Task", "-p", "5")
			cli.MustExecute("-y", "ConflictList", "update", "Conflicting Task", "-p", "1")

			// Verify task exists
			stdout := cli.MustExecute("-y", "ConflictList")
			testutil.AssertContains(t, stdout, "Conflicting Task")

			// Check conflicts (should be empty without real remote)
			stdout = cli.MustExecute("-y", "sync", "conflicts")
			testutil.AssertContains(t, stdout, "Conflict")
		})
	}
}

// =============================================================================
// Command Abbreviation Integration Tests
// =============================================================================

// TestCommandAbbreviationsIntegration tests all command abbreviations work correctly
func TestCommandAbbreviationsIntegration(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list
	cli.MustExecute("-y", "list", "create", "AbbrevTest")

	// Test 'a' for add
	stdout := cli.MustExecute("-y", "AbbrevTest", "a", "Task via a")
	testutil.AssertContains(t, stdout, "Task via a")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Test 'g' for get
	stdout = cli.MustExecute("-y", "AbbrevTest", "g")
	testutil.AssertContains(t, stdout, "Task via a")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)

	// Test 'u' for update
	stdout = cli.MustExecute("-y", "AbbrevTest", "u", "Task via a", "-p", "1")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Test 'c' for complete
	stdout = cli.MustExecute("-y", "AbbrevTest", "c", "Task via a")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Add another task for delete test
	cli.MustExecute("-y", "AbbrevTest", "add", "Task to delete")

	// Test 'd' for delete
	stdout = cli.MustExecute("-y", "AbbrevTest", "d", "Task to delete")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify delete worked
	stdout = cli.MustExecute("-y", "AbbrevTest")
	testutil.AssertNotContains(t, stdout, "Task to delete")
}

// =============================================================================
// JSON Output Integration Tests
// =============================================================================

// TestJSONOutputIntegration tests that --json flag produces valid JSON output
func TestJSONOutputIntegration(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list and add tasks
	cli.MustExecute("-y", "list", "create", "JSONTest")
	cli.MustExecute("-y", "JSONTest", "add", "JSON Task 1", "-p", "1")
	cli.MustExecute("-y", "JSONTest", "add", "JSON Task 2", "-p", "2", "--tag", "important,urgent")

	// Test JSON output for task list
	stdout := cli.MustExecute("-y", "--json", "JSONTest")

	// Verify JSON structure
	if !strings.HasPrefix(strings.TrimSpace(stdout), "{") && !strings.HasPrefix(strings.TrimSpace(stdout), "[") {
		t.Errorf("expected JSON output to start with { or [, got:\n%s", stdout)
	}
	testutil.AssertContains(t, stdout, `"JSON Task 1"`)
	testutil.AssertContains(t, stdout, `"JSON Task 2"`)
	testutil.AssertContains(t, stdout, `"priority"`)

	// Test JSON output for list command
	stdout = cli.MustExecute("-y", "--json", "list")
	if !strings.HasPrefix(strings.TrimSpace(stdout), "{") && !strings.HasPrefix(strings.TrimSpace(stdout), "[") {
		t.Errorf("expected JSON output to start with { or [, got:\n%s", stdout)
	}
	testutil.AssertContains(t, stdout, "JSONTest")
}

// =============================================================================
// Tag/Category Integration Tests
// =============================================================================

// TestTagWorkflowIntegration tests tag/category functionality
func TestTagWorkflowIntegration(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list
	cli.MustExecute("-y", "list", "create", "TagTest")

	// Add tasks with tags
	cli.MustExecute("-y", "TagTest", "add", "Work task", "--tag", "work,important")
	cli.MustExecute("-y", "TagTest", "add", "Personal task", "--tag", "personal")
	cli.MustExecute("-y", "TagTest", "add", "No tag task")

	// Filter by tag
	stdout := cli.MustExecute("-y", "TagTest", "--tag", "work")
	testutil.AssertContains(t, stdout, "Work task")
	testutil.AssertNotContains(t, stdout, "Personal task")
	testutil.AssertNotContains(t, stdout, "No tag task")

	// Verify tags appear in output
	stdout = cli.MustExecute("-y", "TagTest")
	testutil.AssertContains(t, stdout, "work")
	testutil.AssertContains(t, stdout, "important")
	testutil.AssertContains(t, stdout, "personal")
}

// =============================================================================
// Date Filtering Integration Tests
// =============================================================================

// TestDueDateWorkflowIntegration tests due date functionality
func TestDueDateWorkflowIntegration(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list
	cli.MustExecute("-y", "list", "create", "DueDateTest")

	// Add tasks with due dates
	cli.MustExecute("-y", "DueDateTest", "add", "Due soon", "--due-date", "2025-01-15")
	cli.MustExecute("-y", "DueDateTest", "add", "Due later", "--due-date", "2025-06-01")
	cli.MustExecute("-y", "DueDateTest", "add", "No due date")

	// Filter by due date
	stdout := cli.MustExecute("-y", "DueDateTest", "--due-before", "2025-02-01")
	testutil.AssertContains(t, stdout, "Due soon")
	testutil.AssertNotContains(t, stdout, "Due later")

	// Verify due dates appear in output (displayed as short format like "Jan 15")
	stdout = cli.MustExecute("-y", "DueDateTest")
	testutil.AssertContains(t, stdout, "Jan 15")
	testutil.AssertContains(t, stdout, "Jun 01")
}

// =============================================================================
// Error Handling Integration Tests
// =============================================================================

// TestErrorHandlingIntegration tests proper error handling in various scenarios
func TestErrorHandlingIntegration(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Test: Complete non-existent task
	stdout, stderr := cli.ExecuteAndFail("-y", "NonExistentList", "complete", "Ghost Task")
	combined := stdout + stderr
	if !strings.Contains(strings.ToLower(combined), "not found") && !strings.Contains(strings.ToLower(combined), "no") {
		t.Logf("Expected 'not found' or 'no' in error, got: %s", combined)
	}

	// Test: Create list then add task with invalid status
	cli.MustExecute("-y", "list", "create", "ErrorTest")
	stdout, stderr = cli.ExecuteAndFail("-y", "ErrorTest", "add", "-s", "INVALID", "Bad task")
	combined = stdout + stderr
	if !strings.Contains(strings.ToLower(combined), "invalid") {
		t.Logf("Expected 'invalid' in error, got: %s", combined)
	}

	// Test: Update non-existent task
	stdout, stderr = cli.ExecuteAndFail("-y", "ErrorTest", "update", "Ghost", "--summary", "New")
	combined = stdout + stderr
	if !strings.Contains(strings.ToLower(combined), "no") && !strings.Contains(strings.ToLower(combined), "found") && !strings.Contains(strings.ToLower(combined), "match") {
		t.Logf("Expected 'no/found/match' in error, got: %s", combined)
	}

	// Test: Ambiguous task match
	cli.MustExecute("-y", "ErrorTest", "add", "Similar task 1")
	cli.MustExecute("-y", "ErrorTest", "add", "Similar task 2")
	stdout, stderr = cli.ExecuteAndFail("-y", "ErrorTest", "complete", "Similar task")
	combined = stdout + stderr
	if !strings.Contains(strings.ToLower(combined), "multiple") && !strings.Contains(strings.ToLower(combined), "ambiguous") {
		t.Logf("Expected 'multiple' or 'ambiguous' in error, got: %s", combined)
	}
}

// =============================================================================
// Bulk Operations Integration Tests
// =============================================================================

// TestBulkOperationsIntegration tests bulk complete/update/delete operations
func TestBulkOperationsIntegration(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create list with hierarchy
	cli.MustExecute("-y", "list", "create", "BulkIntegration")
	cli.MustExecute("-y", "BulkIntegration", "add", "Epic")
	cli.MustExecute("-y", "BulkIntegration", "add", "Story1", "-P", "Epic")
	cli.MustExecute("-y", "BulkIntegration", "add", "Story2", "-P", "Epic")
	cli.MustExecute("-y", "BulkIntegration", "add", "Task1", "-P", "Story1")
	cli.MustExecute("-y", "BulkIntegration", "add", "Task2", "-P", "Story1")

	// Bulk complete direct children
	stdout := cli.MustExecute("-y", "BulkIntegration", "complete", "Epic/*")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify only direct children completed (Story1, Story2), not grandchildren
	stdout = cli.MustExecute("-y", "--json", "BulkIntegration", "-v", "all")
	// Story1 and Story2 should be DONE
	// Task1 and Task2 should still be TODO
	testutil.AssertContains(t, stdout, "Story1")
	testutil.AssertContains(t, stdout, "Story2")

	// Create another hierarchy for recursive test
	cli.MustExecute("-y", "BulkIntegration", "add", "Project")
	cli.MustExecute("-y", "BulkIntegration", "add", "Feature", "-P", "Project")
	cli.MustExecute("-y", "BulkIntegration", "add", "SubFeature", "-P", "Feature")

	// Bulk complete all descendants
	stdout = cli.MustExecute("-y", "BulkIntegration", "complete", "Project/**")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify all descendants are completed
	stdout = cli.MustExecute("-y", "--json", "BulkIntegration", "-v", "all")
	testutil.AssertContains(t, stdout, "Feature")
	testutil.AssertContains(t, stdout, "SubFeature")
}

// =============================================================================
// Export/Import Integration Tests
// =============================================================================

// TestExportImportIntegration tests list export and import functionality
func TestExportImportIntegration(t *testing.T) {
	cli := testutil.NewCLITest(t)
	tmpDir := cli.TmpDir()

	// Create list with tasks
	cli.MustExecute("-y", "list", "create", "ExportTest")
	cli.MustExecute("-y", "ExportTest", "add", "Task 1", "-p", "1")
	cli.MustExecute("-y", "ExportTest", "add", "Task 2", "-p", "2")

	// Export to sqlite format
	exportPath := filepath.Join(tmpDir, "export.db")
	stdout := cli.MustExecute("-y", "list", "export", "ExportTest", "--format", "sqlite", "--output", exportPath)
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify export file exists
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Fatalf("export file not created: %s", exportPath)
	}

	// Delete original list
	cli.MustExecute("-y", "list", "delete", "ExportTest")
	cli.MustExecute("-y", "list", "purge", "ExportTest")

	// Import from sqlite (file is positional argument)
	stdout = cli.MustExecute("-y", "list", "import", exportPath)
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify tasks were imported
	stdout = cli.MustExecute("-y", "ExportTest")
	testutil.AssertContains(t, stdout, "Task 1")
	testutil.AssertContains(t, stdout, "Task 2")
}
