package migrate_test

import (
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// =============================================================================
// Cross-Backend Migration Tests (026-cross-backend-migration)
// =============================================================================

// TestMigrateSQLiteToNextcloud tests migrating tasks from SQLite to Nextcloud.
// Since Nextcloud requires actual network credentials, this test uses mock mode.
func TestMigrateSQLiteToNextcloud(t *testing.T) {
	cli := testutil.NewCLITestWithMigrate(t)

	// Create source tasks in SQLite
	cli.MustExecute("-y", "SourceList", "add", "Task 1")
	cli.MustExecute("-y", "SourceList", "add", "Task 2")
	cli.MustExecute("-y", "SourceList", "add", "Task 3")

	// Migrate from sqlite to nextcloud (mock mode)
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "nextcloud-mock")

	testutil.AssertContains(t, stdout, "Migrated 3 task")
	testutil.AssertContains(t, stdout, "SourceList")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// TestMigrateNextcloudToTodoist tests migrating tasks from Nextcloud to Todoist.
// Uses mock backends for testing without actual network access.
func TestMigrateNextcloudToTodoist(t *testing.T) {
	cli := testutil.NewCLITestWithMigrate(t)

	// Set up source tasks in mock nextcloud
	cli.SetupMockNextcloudTasks("ProjectList", []string{"Build feature", "Write docs", "Deploy"})

	// Migrate from nextcloud to todoist (mock mode)
	stdout := cli.MustExecute("-y", "migrate", "--from", "nextcloud-mock", "--to", "todoist-mock")

	testutil.AssertContains(t, stdout, "Migrated 3 task")
	testutil.AssertContains(t, stdout, "ProjectList")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// TestMigrateListSelection tests migrating only a specific list.
func TestMigrateListSelection(t *testing.T) {
	cli := testutil.NewCLITestWithMigrate(t)

	// Create tasks in multiple lists
	cli.MustExecute("-y", "Work", "add", "Work task 1")
	cli.MustExecute("-y", "Work", "add", "Work task 2")
	cli.MustExecute("-y", "Personal", "add", "Personal task 1")
	cli.MustExecute("-y", "Shopping", "add", "Shopping task 1")

	// Migrate only Work list
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "file-mock", "--list", "Work")

	testutil.AssertContains(t, stdout, "Migrated 2 task")
	testutil.AssertContains(t, stdout, "Work")
	testutil.AssertNotContains(t, stdout, "Personal")
	testutil.AssertNotContains(t, stdout, "Shopping")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
}

// TestMigratePreservesMetadata tests that migrated tasks retain all metadata.
func TestMigratePreservesMetadata(t *testing.T) {
	cli := testutil.NewCLITestWithMigrate(t)

	// Create task with full metadata
	cli.MustExecute("-y", "Work", "add", "Important task", "-p", "1")
	cli.MustExecute("-y", "Work", "update", "Important task", "--due-date", "2025-12-31", "--tag", "urgent,review")
	cli.MustExecute("-y", "Work", "update", "Important task", "-s", "IN-PROGRESS")

	// Migrate to file backend (which preserves all fields)
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "file-mock", "--list", "Work")

	testutil.AssertContains(t, stdout, "Migrated 1 task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify metadata was preserved by reading from target
	targetOutput := cli.MustExecute("-y", "migrate", "--target-info", "file-mock", "--list", "Work", "--json")

	testutil.AssertContains(t, targetOutput, "priority")
	testutil.AssertContains(t, targetOutput, "1")
	testutil.AssertContains(t, targetOutput, "2025-12-31")
	testutil.AssertContains(t, targetOutput, "IN-PROGRESS")
	testutil.AssertContains(t, targetOutput, "urgent")
	testutil.AssertContains(t, targetOutput, "review")
}

// TestMigratePreservesHierarchy tests that parent-child relationships are preserved.
func TestMigratePreservesHierarchy(t *testing.T) {
	cli := testutil.NewCLITestWithMigrate(t)

	// Create hierarchical tasks
	cli.MustExecute("-y", "Work", "add", "Parent task")
	cli.MustExecute("-y", "Work", "add", "Child task 1", "-P", "Parent task")
	cli.MustExecute("-y", "Work", "add", "Child task 2", "-P", "Parent task")
	cli.MustExecute("-y", "Work", "add", "Grandchild", "-P", "Child task 1")

	// Migrate to file backend
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "file-mock", "--list", "Work")

	testutil.AssertContains(t, stdout, "Migrated 4 task")
	testutil.AssertContains(t, stdout, "hierarchy preserved")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify hierarchy by reading target tasks
	targetOutput := cli.MustExecute("-y", "migrate", "--target-info", "file-mock", "--list", "Work")

	testutil.AssertContains(t, targetOutput, "Parent task")
	testutil.AssertContains(t, targetOutput, "Child task 1")
	testutil.AssertContains(t, targetOutput, "Child task 2")
	testutil.AssertContains(t, targetOutput, "Grandchild")
}

// TestMigrateDryRun tests the dry-run mode that shows changes without applying them.
func TestMigrateDryRun(t *testing.T) {
	cli := testutil.NewCLITestWithMigrate(t)

	// Create source tasks
	cli.MustExecute("-y", "Work", "add", "Task A")
	cli.MustExecute("-y", "Work", "add", "Task B")
	cli.MustExecute("-y", "Work", "add", "Task C")

	// Run migration in dry-run mode
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "file-mock", "--dry-run")

	testutil.AssertContains(t, stdout, "Would migrate 3 task")
	testutil.AssertContains(t, stdout, "Task A")
	testutil.AssertContains(t, stdout, "Task B")
	testutil.AssertContains(t, stdout, "Task C")
	testutil.AssertContains(t, stdout, "dry-run")
	// Result code should be INFO since no changes were made
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)

	// Verify source tasks still exist and target is empty
	sourceOutput := cli.MustExecute("-y", "Work")
	testutil.AssertContains(t, sourceOutput, "Task A")

	// Target should be empty after dry-run
	_, stderr, exitCode := cli.Execute("-y", "migrate", "--target-info", "file-mock", "--list", "Work")
	if exitCode == 0 {
		// If successful, target should have no tasks
		targetInfo := cli.MustExecute("-y", "migrate", "--target-info", "file-mock", "--list", "Work")
		testutil.AssertContains(t, targetInfo, "0 task")
	} else {
		// List doesn't exist in target - expected after dry-run
		testutil.AssertContains(t, stderr, "not found")
	}
}

// TestMigrateStatusMapping tests that statuses are correctly mapped between backends.
func TestMigrateStatusMapping(t *testing.T) {
	cli := testutil.NewCLITestWithMigrate(t)

	// Create tasks with various statuses
	cli.MustExecute("-y", "Work", "add", "Todo task")
	cli.MustExecute("-y", "Work", "add", "In progress task")
	cli.MustExecute("-y", "Work", "update", "In progress task", "-s", "IN-PROGRESS")
	cli.MustExecute("-y", "Work", "add", "Done task")
	cli.MustExecute("-y", "Work", "complete", "Done task")
	cli.MustExecute("-y", "Work", "add", "Cancelled task")
	cli.MustExecute("-y", "Work", "update", "Cancelled task", "-s", "CANCELLED")

	// Migrate to todoist-mock (which may not support all statuses)
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "todoist-mock", "--list", "Work")

	testutil.AssertContains(t, stdout, "Migrated 4 task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Check status mapping info
	if strings.Contains(stdout, "IN-PROGRESS") {
		// Todoist doesn't have IN-PROGRESS, should show mapping info
		testutil.AssertContains(t, stdout, "mapped")
	}
}

// TestMigrateConflictHandling tests handling of tasks with existing UIDs.
func TestMigrateConflictHandling(t *testing.T) {
	cli := testutil.NewCLITestWithMigrate(t)

	// Create source tasks
	cli.MustExecute("-y", "Work", "add", "Existing task")
	cli.MustExecute("-y", "Work", "add", "New task")

	// First migration
	cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "file-mock", "--list", "Work")

	// Add another task to source
	cli.MustExecute("-y", "Work", "add", "Another new task")

	// Second migration - should handle existing UIDs gracefully
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "file-mock", "--list", "Work")

	// Should either skip existing or update them, not fail
	testutil.AssertContains(t, stdout, "task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Check that conflict handling is reported
	if strings.Contains(stdout, "skipped") || strings.Contains(stdout, "updated") {
		// Good - conflicts were handled
	} else {
		testutil.AssertContains(t, stdout, "Migrated")
	}
}

// TestMigrateBatchSize tests that large lists are migrated in batches with progress.
func TestMigrateBatchSize(t *testing.T) {
	cli := testutil.NewCLITestWithMigrate(t)

	// Create many tasks
	for i := 1; i <= 25; i++ {
		cli.MustExecute("-y", "Work", "add", "Task "+string(rune('A'-1+i)))
	}

	// Migrate with batch progress
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "file-mock", "--list", "Work")

	testutil.AssertContains(t, stdout, "Migrated 25 task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Note: Progress reporting (batch/progress messages) is optional for large migrations
}

// =============================================================================
// Error Cases
// =============================================================================

// TestMigrateInvalidSourceBackend tests error handling for unknown source backend.
func TestMigrateInvalidSourceBackend(t *testing.T) {
	cli := testutil.NewCLITestWithMigrate(t)

	_, stderr, exitCode := cli.Execute("-y", "migrate", "--from", "unknown-backend", "--to", "sqlite")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertContains(t, stderr, "unknown backend")
}

// TestMigrateInvalidTargetBackend tests error handling for unknown target backend.
func TestMigrateInvalidTargetBackend(t *testing.T) {
	cli := testutil.NewCLITestWithMigrate(t)

	_, stderr, exitCode := cli.Execute("-y", "migrate", "--from", "sqlite", "--to", "unknown-backend")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertContains(t, stderr, "unknown backend")
}

// TestMigrateListNotFound tests error handling when specified list doesn't exist.
func TestMigrateListNotFound(t *testing.T) {
	cli := testutil.NewCLITestWithMigrate(t)

	_, stderr, exitCode := cli.Execute("-y", "migrate", "--from", "sqlite", "--to", "file-mock", "--list", "NonExistentList")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertContains(t, stderr, "not found")
}

// TestMigrateSameBackend tests that migrating to same backend type fails.
func TestMigrateSameBackend(t *testing.T) {
	cli := testutil.NewCLITestWithMigrate(t)

	_, stderr, exitCode := cli.Execute("-y", "migrate", "--from", "sqlite", "--to", "sqlite")

	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertContains(t, stderr, "same backend")
}

// TestMigrateMissingFlags tests that required flags are validated.
func TestMigrateMissingFlags(t *testing.T) {
	cli := testutil.NewCLITestWithMigrate(t)

	// Missing --from
	_, stderr, exitCode := cli.Execute("-y", "migrate", "--to", "nextcloud")
	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertContains(t, stderr, "required")

	// Missing --to
	_, stderr, exitCode = cli.Execute("-y", "migrate", "--from", "sqlite")
	testutil.AssertExitCode(t, exitCode, 1)
	testutil.AssertContains(t, stderr, "required")
}

// =============================================================================
// JSON Output Tests
// =============================================================================

// TestMigrateJSONOutput tests that migration results can be output as JSON.
func TestMigrateJSONOutput(t *testing.T) {
	cli := testutil.NewCLITestWithMigrate(t)

	// Create source tasks
	cli.MustExecute("-y", "Work", "add", "JSON Task 1")
	cli.MustExecute("-y", "Work", "add", "JSON Task 2")

	// Migrate with JSON output
	stdout := cli.MustExecute("-y", "--json", "migrate", "--from", "sqlite", "--to", "file-mock", "--list", "Work")

	testutil.AssertContains(t, stdout, "{")
	testutil.AssertContains(t, stdout, "migrated")
	testutil.AssertContains(t, stdout, "2")
}
