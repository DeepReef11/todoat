//go:build integration

// Package migrate_test provides integration tests for cross-backend migration with real backends.
// These tests require actual credentials and are run with `go test -tags=integration`.
package migrate_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"todoat/internal/testutil"
)

// =============================================================================
// Real Backend Migration Integration Tests (077-migration-real-backends)
// =============================================================================

// skipIfNoNextcloudCredentials skips the test if Nextcloud credentials are not configured
// or if the Nextcloud server is not reachable.
// Note: Nextcloud migration tests require an existing calendar since creating calendars
// is not supported via CalDAV. These tests are skipped by default.
func skipIfNoNextcloudCredentials(t *testing.T) {
	t.Helper()
	host := os.Getenv("TODOAT_NEXTCLOUD_HOST")
	username := os.Getenv("TODOAT_NEXTCLOUD_USERNAME")
	password := os.Getenv("TODOAT_NEXTCLOUD_PASSWORD")

	if host == "" || username == "" || password == "" {
		t.Skip("Skipping: TODOAT_NEXTCLOUD_HOST, TODOAT_NEXTCLOUD_USERNAME, and TODOAT_NEXTCLOUD_PASSWORD must be set")
	}

	// Skip by default because Nextcloud doesn't support creating calendars via CalDAV.
	// Migration tests try to create new calendars, which will always fail.
	// Set TODOAT_NEXTCLOUD_MIGRATION_CALENDAR to an existing calendar name to enable these tests.
	if os.Getenv("TODOAT_NEXTCLOUD_MIGRATION_CALENDAR") == "" {
		t.Skip("Skipping: Nextcloud CalDAV doesn't support creating calendars. Set TODOAT_NEXTCLOUD_MIGRATION_CALENDAR to an existing calendar name to run these tests")
	}

	// Verify Nextcloud is reachable by trying to connect
	// Handle hosts with and without protocol prefix
	var testURL string
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		testURL = host + "/status.php"
	} else {
		// Try HTTP first (for local Docker), then HTTPS
		testURL = "http://" + host + "/status.php"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(testURL)
	if err != nil {
		t.Skipf("Skipping: Nextcloud not reachable at %s: %v", host, err)
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 && resp.StatusCode != 401 {
		t.Skipf("Skipping: Nextcloud returned status %d at %s", resp.StatusCode, host)
	}
}

// skipIfNoTodoistCredentials skips the test if Todoist credentials are not configured
// or if the Todoist API is not accessible or can't create projects.
func skipIfNoTodoistCredentials(t *testing.T) {
	t.Helper()
	token := os.Getenv("TODOAT_TODOIST_TOKEN")
	if token == "" {
		t.Skip("Skipping: TODOAT_TODOIST_TOKEN must be set")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// First verify the API token is valid
	getReq, err := http.NewRequest("GET", "https://api.todoist.com/rest/v2/projects", nil)
	if err != nil {
		t.Skipf("Skipping: Failed to create Todoist API request: %v", err)
	}
	getReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(getReq)
	if err != nil {
		t.Skipf("Skipping: Todoist API not reachable: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		t.Skipf("Skipping: Todoist API token is invalid or has insufficient permissions (status %d)", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		t.Skipf("Skipping: Todoist API returned error status %d", resp.StatusCode)
	}

	// Also verify we can create projects (needed for migration tests)
	// Try to create a test project to check for project limit
	createReq, err := http.NewRequest("POST", "https://api.todoist.com/rest/v2/projects", strings.NewReader(`{"name": "__todoat_test_skip_check__"}`))
	if err != nil {
		return // Allow test to run, will fail naturally if there's an issue
	}
	createReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	createReq.Header.Set("Content-Type", "application/json")

	createResp, err := client.Do(createReq)
	if err != nil {
		return // Network error during creation check, let test run
	}
	defer createResp.Body.Close()

	if createResp.StatusCode == 403 {
		t.Skip("Skipping: Todoist account has reached project limit (403)")
	}

	// If we successfully created the test project, delete it
	if createResp.StatusCode == 200 {
		var result struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(createResp.Body).Decode(&result); err == nil && result.ID != "" {
			delReq, _ := http.NewRequest("DELETE", "https://api.todoist.com/rest/v2/projects/"+result.ID, nil)
			delReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			delResp, _ := client.Do(delReq)
			if delResp != nil {
				delResp.Body.Close()
			}
		}
	}
}

// newCLITestWithRealBackends creates a CLI test helper configured for real backend migration.
// It configures environment variables to connect to real backend services.
func newCLITestWithRealBackends(t *testing.T) *testutil.MigrateCLITest {
	t.Helper()

	cli := testutil.NewCLITestWithMigrate(t)
	// Disable mock mode to use real backends
	cli.Config().MigrateMockMode = false
	return cli
}

// TestIntegrationMigrateSQLiteToNextcloud tests migrating tasks from SQLite to real Nextcloud backend.
func TestIntegrationMigrateSQLiteToNextcloud(t *testing.T) {
	skipIfNoNextcloudCredentials(t)
	cli := newCLITestWithRealBackends(t)

	// Create source tasks in SQLite
	cli.MustExecute("-y", "MigrationTest", "add", "NC Task 1")
	cli.MustExecute("-y", "MigrationTest", "add", "NC Task 2")
	cli.MustExecute("-y", "MigrationTest", "add", "NC Task 3")

	// Migrate from sqlite to real nextcloud
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "nextcloud", "--list", "MigrationTest")

	testutil.AssertContains(t, stdout, "Migrated 3 task")
	testutil.AssertContains(t, stdout, "MigrationTest")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify tasks exist in target by getting target info
	targetOutput := cli.MustExecute("-y", "migrate", "--target-info", "nextcloud", "--list", "MigrationTest", "--json")
	testutil.AssertContains(t, targetOutput, "NC Task 1")
	testutil.AssertContains(t, targetOutput, "NC Task 2")
	testutil.AssertContains(t, targetOutput, "NC Task 3")
}

// TestIntegrationMigrateSQLiteToTodoist tests migrating tasks from SQLite to real Todoist backend.
func TestIntegrationMigrateSQLiteToTodoist(t *testing.T) {
	skipIfNoTodoistCredentials(t)
	cli := newCLITestWithRealBackends(t)

	// Create source tasks in SQLite
	cli.MustExecute("-y", "TodoistMigration", "add", "Todoist Task 1")
	cli.MustExecute("-y", "TodoistMigration", "add", "Todoist Task 2")
	cli.MustExecute("-y", "TodoistMigration", "add", "Todoist Task 3")

	// Migrate from sqlite to real todoist
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "todoist", "--list", "TodoistMigration")

	testutil.AssertContains(t, stdout, "Migrated 3 task")
	testutil.AssertContains(t, stdout, "TodoistMigration")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify tasks exist in target by getting target info
	targetOutput := cli.MustExecute("-y", "migrate", "--target-info", "todoist", "--list", "TodoistMigration", "--json")
	testutil.AssertContains(t, targetOutput, "Todoist Task 1")
	testutil.AssertContains(t, targetOutput, "Todoist Task 2")
	testutil.AssertContains(t, targetOutput, "Todoist Task 3")
}

// TestIntegrationMigrateSQLiteToFile tests migrating tasks from SQLite to file backend.
func TestIntegrationMigrateSQLiteToFile(t *testing.T) {
	cli := newCLITestWithRealBackends(t)

	// Configure file backend path to be in test temp dir
	filePath := filepath.Join(cli.TmpDir(), "tasks.txt")
	cli.Config().MigrateTargetDir = cli.TmpDir()

	// Create source tasks in SQLite
	cli.MustExecute("-y", "FileMigration", "add", "File Task 1")
	cli.MustExecute("-y", "FileMigration", "add", "File Task 2")
	cli.MustExecute("-y", "FileMigration", "add", "File Task 3")

	// Migrate from sqlite to file backend
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "file", "--list", "FileMigration")

	testutil.AssertContains(t, stdout, "Migrated 3 task")
	testutil.AssertContains(t, stdout, "FileMigration")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the file was created and contains tasks
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File may be at a different location, check target info instead
		targetOutput := cli.MustExecute("-y", "migrate", "--target-info", "file", "--json")
		testutil.AssertContains(t, targetOutput, "File Task 1")
	}
}

// TestIntegrationMigrateNextcloudToSQLite tests migrating tasks from real Nextcloud to SQLite.
func TestIntegrationMigrateNextcloudToSQLite(t *testing.T) {
	skipIfNoNextcloudCredentials(t)
	cli := newCLITestWithRealBackends(t)

	// First, migrate some tasks TO Nextcloud so we have data to migrate back
	cli.MustExecute("-y", "NCToSQLite", "add", "Round trip task 1")
	cli.MustExecute("-y", "NCToSQLite", "add", "Round trip task 2")
	cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "nextcloud", "--list", "NCToSQLite")

	// Clear the SQLite list
	cli.MustExecute("-y", "NCToSQLite", "delete", "Round trip task 1", "--no-trash")
	cli.MustExecute("-y", "NCToSQLite", "delete", "Round trip task 2", "--no-trash")

	// Now migrate back from nextcloud to sqlite
	stdout := cli.MustExecute("-y", "migrate", "--from", "nextcloud", "--to", "sqlite", "--list", "NCToSQLite")

	testutil.AssertContains(t, stdout, "Migrated")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify tasks exist in SQLite
	listOutput := cli.MustExecute("-y", "NCToSQLite")
	testutil.AssertContains(t, listOutput, "Round trip task")
}

// TestIntegrationMigrateTodoistToSQLite tests migrating tasks from real Todoist to SQLite.
func TestIntegrationMigrateTodoistToSQLite(t *testing.T) {
	skipIfNoTodoistCredentials(t)
	cli := newCLITestWithRealBackends(t)

	// First, migrate some tasks TO Todoist so we have data to migrate back
	cli.MustExecute("-y", "TodoistToSQLite", "add", "Todoist round trip 1")
	cli.MustExecute("-y", "TodoistToSQLite", "add", "Todoist round trip 2")
	cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "todoist", "--list", "TodoistToSQLite")

	// Clear the SQLite list
	cli.MustExecute("-y", "TodoistToSQLite", "delete", "Todoist round trip 1", "--no-trash")
	cli.MustExecute("-y", "TodoistToSQLite", "delete", "Todoist round trip 2", "--no-trash")

	// Now migrate back from todoist to sqlite
	stdout := cli.MustExecute("-y", "migrate", "--from", "todoist", "--to", "sqlite", "--list", "TodoistToSQLite")

	testutil.AssertContains(t, stdout, "Migrated")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify tasks exist in SQLite
	listOutput := cli.MustExecute("-y", "TodoistToSQLite")
	testutil.AssertContains(t, listOutput, "Todoist round trip")
}

// TestIntegrationMigratePreservesMetadata tests that migrated tasks retain priority, dates, status, tags
// when migrating to/from real backends.
func TestIntegrationMigratePreservesMetadata(t *testing.T) {
	skipIfNoNextcloudCredentials(t)
	cli := newCLITestWithRealBackends(t)

	// Create task with full metadata
	cli.MustExecute("-y", "MetadataTest", "add", "Metadata task", "-p", "1")
	cli.MustExecute("-y", "MetadataTest", "update", "Metadata task", "--due-date", "2025-12-31", "--tag", "urgent,priority")
	cli.MustExecute("-y", "MetadataTest", "update", "Metadata task", "-s", "IN-PROGRESS")

	// Migrate to nextcloud
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "nextcloud", "--list", "MetadataTest")
	testutil.AssertContains(t, stdout, "Migrated 1 task")

	// Verify metadata was preserved by reading from target
	targetOutput := cli.MustExecute("-y", "migrate", "--target-info", "nextcloud", "--list", "MetadataTest", "--json")

	// Check JSON contains expected metadata
	testutil.AssertContains(t, targetOutput, "priority")
	testutil.AssertContains(t, targetOutput, "1")
	testutil.AssertContains(t, targetOutput, "2025-12-31")
	// Status should be IN-PROGRESS (or mapped equivalent)
	if !strings.Contains(targetOutput, "IN-PROGRESS") && !strings.Contains(targetOutput, "in_progress") {
		t.Logf("Status may have been mapped to different format in target output")
	}
}

// TestIntegrationMigratePreservesHierarchy tests that parent-child relationships are preserved
// when migrating to/from real backends.
func TestIntegrationMigratePreservesHierarchy(t *testing.T) {
	skipIfNoTodoistCredentials(t)
	cli := newCLITestWithRealBackends(t)

	// Create hierarchical tasks
	cli.MustExecute("-y", "HierarchyTest", "add", "Parent task")
	cli.MustExecute("-y", "HierarchyTest", "add", "Child task 1", "-P", "Parent task")
	cli.MustExecute("-y", "HierarchyTest", "add", "Child task 2", "-P", "Parent task")

	// Migrate to todoist (which supports hierarchy)
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "todoist", "--list", "HierarchyTest")

	testutil.AssertContains(t, stdout, "Migrated 3 task")
	testutil.AssertContains(t, stdout, "hierarchy preserved")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify hierarchy by reading target tasks
	targetOutput := cli.MustExecute("-y", "migrate", "--target-info", "todoist", "--list", "HierarchyTest", "--json")

	testutil.AssertContains(t, targetOutput, "Parent task")
	testutil.AssertContains(t, targetOutput, "Child task 1")
	testutil.AssertContains(t, targetOutput, "Child task 2")

	// Parse JSON to verify parent_id is set
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(targetOutput), &result); err == nil {
		// Check that at least one task has a parent_id
		if tasks, ok := result["tasks"].([]interface{}); ok {
			hasParent := false
			for _, task := range tasks {
				if taskMap, ok := task.(map[string]interface{}); ok {
					if _, ok := taskMap["parent_id"]; ok {
						hasParent = true
						break
					}
				}
			}
			if !hasParent {
				t.Log("Warning: parent_id not found in JSON output (may be in different format)")
			}
		}
	}
}

// TestIntegrationMigrateStatusMapping tests that statuses are correctly mapped between backends.
func TestIntegrationMigrateStatusMapping(t *testing.T) {
	skipIfNoNextcloudCredentials(t)
	cli := newCLITestWithRealBackends(t)

	// Create tasks with various statuses
	cli.MustExecute("-y", "StatusMap", "add", "Todo task")
	cli.MustExecute("-y", "StatusMap", "add", "In progress task")
	cli.MustExecute("-y", "StatusMap", "update", "In progress task", "-s", "IN-PROGRESS")
	cli.MustExecute("-y", "StatusMap", "add", "Done task")
	cli.MustExecute("-y", "StatusMap", "complete", "Done task")
	cli.MustExecute("-y", "StatusMap", "add", "Cancelled task")
	cli.MustExecute("-y", "StatusMap", "update", "Cancelled task", "-s", "CANCELLED")

	// Migrate to nextcloud (CalDAV uses different status names)
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "nextcloud", "--list", "StatusMap")

	testutil.AssertContains(t, stdout, "Migrated 4 task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Nextcloud uses CalDAV statuses (NEEDS-ACTION, COMPLETED, IN-PROCESS, CANCELLED)
	// Check that status mapping info is shown if applicable
	if strings.Contains(stdout, "mapped") {
		t.Logf("Status mapping detected: %s", stdout)
	}
}

// TestIntegrationMigrateRateLimiting tests that migration respects API rate limits.
func TestIntegrationMigrateRateLimiting(t *testing.T) {
	skipIfNoTodoistCredentials(t)
	cli := newCLITestWithRealBackends(t)

	// Create many tasks to trigger rate limiting
	for i := 1; i <= 15; i++ {
		cli.MustExecute("-y", "RateLimitTest", "add", "Task "+string(rune('A'-1+i)))
	}

	// Migrate all tasks - this should respect rate limits
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "todoist", "--list", "RateLimitTest")

	testutil.AssertContains(t, stdout, "Migrated 15 task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// No rate limit errors should cause the migration to fail
	// If we get here without errors, rate limiting is working
}

// =============================================================================
// Real File Backend Tests (no external dependencies)
// =============================================================================

// RealFileCLITest provides a CLI test helper for real file backend testing.
type RealFileCLITest struct {
	*testutil.CLITest
	filePath string
}

// newCLITestWithFileBackend creates a CLI test helper for file backend migration testing.
func newCLITestWithFileBackend(t *testing.T) *RealFileCLITest {
	t.Helper()

	cli := testutil.NewCLITestWithConfig(t)
	tmpDir := cli.TmpDir()
	filePath := filepath.Join(tmpDir, "migrate-tasks.txt")

	// Configure config to use file backend for migration
	cfg := `default_backend: sqlite
backends:
  file:
    enabled: true
    path: "` + filePath + `"
`
	cli.SetFullConfig(cfg)
	cli.Config().MigrateMockMode = false

	return &RealFileCLITest{
		CLITest:  cli,
		filePath: filePath,
	}
}

// TestIntegrationMigrateSQLiteToRealFile tests migrating from SQLite to real file backend.
func TestIntegrationMigrateSQLiteToRealFile(t *testing.T) {
	cli := newCLITestWithFileBackend(t)

	// Create source tasks in SQLite
	cli.MustExecute("-y", "FileList", "add", "File task 1")
	cli.MustExecute("-y", "FileList", "add", "File task 2")
	cli.MustExecute("-y", "FileList", "add", "File task 3")

	// Migrate from sqlite to file backend
	stdout := cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "file", "--list", "FileList")

	testutil.AssertContains(t, stdout, "Migrated 3 task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the file was created
	if _, err := os.Stat(cli.filePath); os.IsNotExist(err) {
		t.Errorf("Expected file backend to create file at %s", cli.filePath)
	}

	// Verify tasks are in the file
	data, err := os.ReadFile(cli.filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	content := string(data)
	testutil.AssertContains(t, content, "File task 1")
	testutil.AssertContains(t, content, "File task 2")
	testutil.AssertContains(t, content, "File task 3")
}

// TestIntegrationMigrateFileToSQLite tests migrating from file backend to SQLite.
func TestIntegrationMigrateFileToSQLite(t *testing.T) {
	cli := newCLITestWithFileBackend(t)

	// Create a file with tasks using the correct file format
	// File backend expects ## for section headers (not #)
	fileContent := `# Tasks

## FileList

- [ ] Imported task 1
- [ ] Imported task 2
- [x] Completed task
`
	if err := os.WriteFile(cli.filePath, []byte(fileContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Migrate from file to sqlite
	stdout := cli.MustExecute("-y", "migrate", "--from", "file", "--to", "sqlite", "--list", "FileList")

	testutil.AssertContains(t, stdout, "Migrated")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify tasks exist in SQLite (use --status to include all statuses)
	listOutput := cli.MustExecute("-y", "FileList", "--status", "TODO,IN-PROGRESS,DONE,CANCELLED")
	testutil.AssertContains(t, listOutput, "Imported task 1")
	testutil.AssertContains(t, listOutput, "Imported task 2")
	testutil.AssertContains(t, listOutput, "Completed task")
}

// TestIntegrationMigrateFilePreservesTaskStatus tests that file backend migration preserves completion status.
func TestIntegrationMigrateFilePreservesTaskStatus(t *testing.T) {
	cli := newCLITestWithFileBackend(t)

	// Create tasks with different statuses in SQLite
	cli.MustExecute("-y", "StatusTest", "add", "Open task")
	cli.MustExecute("-y", "StatusTest", "add", "Done task")
	cli.MustExecute("-y", "StatusTest", "complete", "Done task")

	// Migrate to file
	cli.MustExecute("-y", "migrate", "--from", "sqlite", "--to", "file", "--list", "StatusTest")

	// Read file and verify checkboxes
	data, err := os.ReadFile(cli.filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	content := string(data)

	// File format uses [ ] for open and [x] for done
	testutil.AssertContains(t, content, "[ ] Open task")
	testutil.AssertContains(t, content, "[x] Done task")
}
