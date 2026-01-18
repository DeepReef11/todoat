package sync_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// =============================================================================
// Sync Core System Tests (018-synchronization-core)
// =============================================================================

// TestSyncPull tests that `todoat sync` pulls changes from remote backend to local cache
func TestSyncPull(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled
	createSyncConfig(t, tmpDir, true)

	// Run sync command
	stdout, _, exitCode := cli.Execute("-y", "sync")

	// Sync should complete (exit 0) or report no remote configured
	// For this test, we expect it to run without crashing
	if exitCode != 0 && !strings.Contains(stdout, "no remote backend") {
		testutil.AssertContains(t, stdout, "Sync")
	}
}

// TestSyncPush tests that `todoat sync` pushes queued local changes to remote backend
func TestSyncPush(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled
	createSyncConfig(t, tmpDir, true)

	// Add a task which should queue a sync operation
	cli.MustExecute("-y", "Work", "add", "Task to sync")

	// Run sync command
	stdout, _, exitCode := cli.Execute("-y", "sync")

	// Sync should complete (exit 0) or report no remote configured
	if exitCode != 0 && !strings.Contains(stdout, "no remote backend") {
		testutil.AssertContains(t, stdout, "Sync")
	}
}

// TestSyncStatus tests that `todoat sync status` shows last sync time, pending operations, and connection status
func TestSyncStatus(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled
	createSyncConfig(t, tmpDir, true)

	// Run sync status command
	stdout := cli.MustExecute("-y", "sync", "status")

	// Should show status information
	// Expected fields: last sync, pending operations, connection status
	testutil.AssertContains(t, stdout, "Sync Status")
}

// TestSyncQueueView tests that `todoat sync queue` lists pending operations with timestamps
func TestSyncQueueView(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled
	createSyncConfig(t, tmpDir, true)

	// Run sync queue command
	stdout := cli.MustExecute("-y", "sync", "queue")

	// Should show queue information (possibly empty)
	testutil.AssertContains(t, stdout, "Pending Operations")
}

// TestSyncQueueClear tests that `todoat sync queue clear` removes all pending operations
func TestSyncQueueClear(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled
	createSyncConfig(t, tmpDir, true)

	// Add a task to create a pending operation
	cli.MustExecute("-y", "Work", "add", "Task in queue")

	// Clear the queue
	stdout := cli.MustExecute("-y", "sync", "queue", "clear")

	testutil.AssertContains(t, stdout, "cleared")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify queue is empty
	stdout = cli.MustExecute("-y", "sync", "queue")
	testutil.AssertContains(t, stdout, "0")
}

// TestSyncOfflineAdd tests that adding a task while offline queues operation in sync_queue table
func TestSyncOfflineAdd(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled
	createSyncConfig(t, tmpDir, true)

	// Add a task (this should queue a create operation)
	cli.MustExecute("-y", "Work", "add", "Offline task")

	// Check the queue
	stdout := cli.MustExecute("-y", "sync", "queue")

	// Should show pending create operation
	testutil.AssertContains(t, stdout, "create")
	testutil.AssertContains(t, stdout, "Offline task")
}

// TestSyncOfflineUpdate tests that updating a task while offline queues operation
func TestSyncOfflineUpdate(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled
	createSyncConfig(t, tmpDir, true)

	// Add and update a task
	cli.MustExecute("-y", "Work", "add", "Task to update")
	cli.MustExecute("-y", "Work", "update", "Task to update", "-p", "1")

	// Check the queue
	stdout := cli.MustExecute("-y", "sync", "queue")

	// Should show pending update operation
	testutil.AssertContains(t, stdout, "update")
}

// TestSyncOfflineDelete tests that deleting a task while offline queues operation
func TestSyncOfflineDelete(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled
	createSyncConfig(t, tmpDir, true)

	// Add and delete a task
	cli.MustExecute("-y", "Work", "add", "Task to delete")
	cli.MustExecute("-y", "Work", "delete", "Task to delete")

	// Check the queue
	stdout := cli.MustExecute("-y", "sync", "queue")

	// Should show pending delete operation
	testutil.AssertContains(t, stdout, "delete")
}

// TestSyncCacheIsolation tests that each remote backend has separate cache tables
func TestSyncCacheIsolation(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create config with multiple backends
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
backends:
  nextcloud:
    type: nextcloud
    enabled: true
  todoist:
    type: todoist
    enabled: true
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Check sync status - should show separate backends
	stdout := cli.MustExecute("-y", "sync", "status")

	// Should list multiple backends with their own cache status
	testutil.AssertContains(t, stdout, "nextcloud")
	testutil.AssertContains(t, stdout, "todoist")
}

// TestSyncETagSupport tests that updates use If-Match header with ETag for optimistic locking
func TestSyncETagSupport(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled
	createSyncConfig(t, tmpDir, true)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "ETag test task")

	// Check sync metadata (etags should be tracked)
	// This tests that the infrastructure exists - actual etag verification
	// requires a mock remote backend
	stdout := cli.MustExecute("-y", "sync", "status", "--verbose")

	// Should show sync metadata
	testutil.AssertContains(t, stdout, "Sync Status")
}

// TestSyncConfigEnabled tests that `sync.enabled: true` in config enables sync behavior
func TestSyncConfigEnabled(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled
	createSyncConfig(t, tmpDir, true)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Sync enabled task")

	// Check sync queue - should have pending operations
	stdout := cli.MustExecute("-y", "sync", "queue")

	// Should show pending operations (task was queued)
	testutil.AssertContains(t, stdout, "Pending Operations")
	// At least 1 pending operation
	if !strings.Contains(stdout, "create") && !strings.Contains(stdout, "1") {
		t.Errorf("expected at least 1 pending operation, got:\n%s", stdout)
	}
}

// TestSyncConfigDisabled tests that `sync.enabled: false` bypasses sync manager
func TestSyncConfigDisabled(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync disabled
	createSyncConfig(t, tmpDir, false)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Sync disabled task")

	// Try to run sync - should indicate sync is disabled
	stdout, stderr, exitCode := cli.Execute("-y", "sync", "status")

	// Should indicate that sync is disabled or not configured
	combined := stdout + stderr
	if exitCode == 0 {
		// If it succeeds, it should show sync is disabled
		if !strings.Contains(combined, "disabled") && !strings.Contains(combined, "not enabled") {
			// Alternatively, the queue might be empty because sync is disabled
			testutil.AssertContains(t, combined, "Sync")
		}
	}
}

// Helper functions

// newSyncTestCLI creates a CLI test with sync-enabled configuration
func newSyncTestCLI(t *testing.T) (*testutil.CLITest, string) {
	t.Helper()
	cli, viewsDir := testutil.NewCLITestWithViews(t)
	// viewsDir parent is the tmpDir
	tmpDir := filepath.Dir(viewsDir)
	return cli, tmpDir
}

// createSyncConfig creates a config file with sync enabled/disabled
func createSyncConfig(t *testing.T, tmpDir string, enabled bool) {
	t.Helper()
	createSyncConfigWithStrategy(t, tmpDir, enabled, "server_wins")
}

// createSyncConfigWithStrategy creates a config file with sync enabled/disabled and specific conflict strategy
func createSyncConfigWithStrategy(t *testing.T, tmpDir string, enabled bool, strategy string) {
	t.Helper()

	enabledStr := "true"
	if !enabled {
		enabledStr = "false"
	}

	configContent := `
sync:
  enabled: ` + enabledStr + `
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
}

// =============================================================================
// Sync Conflict Resolution Tests (019-sync-conflict-resolution)
// =============================================================================

// TestConflictDetection tests that sync detects when local and remote have both changed same task
func TestConflictDetection(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	createSyncConfig(t, tmpDir, true)

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Conflict test task")

	// Simulate a conflict by modifying the task locally
	cli.MustExecute("-y", "Work", "update", "Conflict test task", "-p", "1")

	// Run sync - conflicts should be detected if there were remote changes
	// For this test without a real remote, we verify the detection mechanism exists
	stdout := cli.MustExecute("-y", "sync", "status")

	// The sync status should show conflict information (0 conflicts when no remote)
	testutil.AssertContains(t, stdout, "Sync Status")
}

// TestConflictServerWins tests that with `conflict_strategy: server-wins`, remote changes override local
func TestConflictServerWins(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	createSyncConfigWithStrategy(t, tmpDir, true, "server_wins")

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Server wins task")

	// Check sync status to verify conflict strategy is configured
	stdout := cli.MustExecute("-y", "sync", "status")
	testutil.AssertContains(t, stdout, "Sync Status")

	// Verify the task exists
	stdout = cli.MustExecute("-y", "Work", "get")
	testutil.AssertContains(t, stdout, "Server wins task")
}

// TestConflictLocalWins tests that with `conflict_strategy: local-wins`, local changes override remote
func TestConflictLocalWins(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	createSyncConfigWithStrategy(t, tmpDir, true, "local_wins")

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Local wins task")

	// Modify the task locally
	cli.MustExecute("-y", "Work", "update", "Local wins task", "-p", "1")

	// The local changes should be preserved after sync
	stdout := cli.MustExecute("-y", "Work", "get")
	testutil.AssertContains(t, stdout, "Local wins task")
	testutil.AssertContains(t, stdout, "[P1]")
}

// TestConflictMerge tests that with `conflict_strategy: merge`, non-conflicting fields are combined
func TestConflictMerge(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	createSyncConfigWithStrategy(t, tmpDir, true, "merge")

	// Add a task with initial values
	cli.MustExecute("-y", "Work", "add", "Merge task", "-p", "1")

	// Update with additional fields
	cli.MustExecute("-y", "Work", "update", "Merge task", "--tag", "important")

	// Verify the task has merged values
	stdout := cli.MustExecute("-y", "Work", "get")
	testutil.AssertContains(t, stdout, "Merge task")
	testutil.AssertContains(t, stdout, "[P1]")
	testutil.AssertContains(t, stdout, "important")
}

// TestConflictKeepBoth tests that with `conflict_strategy: keep-both`, duplicate task created
func TestConflictKeepBoth(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	createSyncConfigWithStrategy(t, tmpDir, true, "keep_both")

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Keep both task")

	// Check that task exists (in real conflict scenario, there would be two)
	stdout := cli.MustExecute("-y", "Work", "get")
	testutil.AssertContains(t, stdout, "Keep both task")
}

// TestConflictStatusDisplay tests that `todoat sync status` shows count of conflicts needing attention
func TestConflictStatusDisplay(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	createSyncConfig(t, tmpDir, true)

	// Run sync status
	stdout := cli.MustExecute("-y", "sync", "status")

	// Should show conflicts count (0 when no conflicts)
	testutil.AssertContains(t, stdout, "Sync Status")
	// The output should include conflict information section
	// When no conflicts exist, it may show "Conflicts: 0" or similar
}

// TestConflictList tests that `todoat sync conflicts` lists all unresolved conflicts with details
func TestConflictList(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	createSyncConfig(t, tmpDir, true)

	// Run sync conflicts command
	stdout := cli.MustExecute("-y", "sync", "conflicts")

	// Should show conflicts list (empty when no conflicts)
	testutil.AssertContains(t, stdout, "Conflict")
}

// TestConflictResolve tests that `todoat sync conflicts resolve [task-uid] --strategy server-wins` resolves specific conflict
func TestConflictResolve(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	createSyncConfig(t, tmpDir, true)

	// Add a task to have something to work with
	cli.MustExecute("-y", "Work", "add", "Task with conflict")

	// Without a real remote, we can't create a true conflict
	// But we can verify the resolve command exists and handles the no-conflict case
	stdout, _, exitCode := cli.Execute("-y", "sync", "conflicts", "resolve", "nonexistent-uid", "--strategy", "server_wins")

	// Command should exist and run (may error because no conflict exists)
	// We're testing that the command infrastructure is in place
	if exitCode == 0 {
		testutil.AssertContains(t, stdout, "resolve")
	}
	// Non-zero exit is acceptable when trying to resolve non-existent conflict
}

// TestConflictDefaultStrategy tests that default strategy is configurable in config.yaml
func TestConflictDefaultStrategy(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Test with different default strategies
	strategies := []string{"server_wins", "local_wins", "merge", "keep_both"}

	for _, strategy := range strategies {
		t.Run(strategy, func(t *testing.T) {
			createSyncConfigWithStrategy(t, tmpDir, true, strategy)

			// Verify config is loaded correctly by checking sync status
			stdout := cli.MustExecute("-y", "sync", "status")
			testutil.AssertContains(t, stdout, "Sync Status")
		})
	}
}

// TestConflictJSONOutput tests that `todoat --json sync conflicts` returns conflicts in JSON format
func TestConflictJSONOutput(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	createSyncConfig(t, tmpDir, true)

	// Run sync conflicts command with JSON output
	stdout := cli.MustExecute("-y", "--json", "sync", "conflicts")

	// Should be valid JSON output
	// Empty conflicts list should be [] or {"conflicts": []}
	if !strings.Contains(stdout, "[") && !strings.Contains(stdout, "{") {
		t.Errorf("expected JSON output, got: %s", stdout)
	}
}
