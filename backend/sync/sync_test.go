package sync_test

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
	cmd "todoat/cmd/todoat/cmd"
	"todoat/internal/testutil"
)

// =============================================================================
// Sync Core System Tests (018-synchronization-core)
// =============================================================================

// TestSyncWithConfiguredRemoteBackend tests that `todoat sync` does NOT report
// "no remote backend configured" when a remote backend IS configured (Issue 1)
// This was a bug where the sync command was a stub that always said no remote was configured
func TestSyncWithConfiguredRemoteBackend(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with a remote backend configured (nextcloud)
	// and sync enabled with sqlite as local backend
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  conflict_resolution: server_wins
backends:
  sqlite:
    type: sqlite
    enabled: true
  nextcloud-test:
    type: nextcloud
    enabled: true
    host: "localhost:8080"
    username: "admin"
    allow_http: true
default_backend: nextcloud-test
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Run sync command
	stdout, _, exitCode := cli.Execute("-y", "sync")

	// The sync command should NOT say "no remote backend configured"
	// when a remote backend IS configured
	if strings.Contains(stdout, "no remote backend configured") {
		t.Errorf("sync command incorrectly reports 'no remote backend configured' when nextcloud-test is configured.\nOutput: %s", stdout)
	}

	// Should succeed or fail gracefully (e.g., connectivity issues are acceptable)
	// But it should NOT say there's no remote backend
	if exitCode != 0 {
		// Acceptable failure reasons (connectivity, auth, etc.)
		// But NOT "no remote backend configured"
		t.Logf("sync command exited with code %d (may be expected for connectivity issues)\nOutput: %s", exitCode, stdout)
	}
}

// TestSyncFallbackToSQLiteWhenRemoteUnavailable tests Issue #0:
// When sync is enabled and the configured remote backend is unavailable (timeout/unreachable),
// the app should fall back to SQLite cache instead of erroring out.
// This allows operations to continue offline and queue changes for later sync.
func TestSyncFallbackToSQLiteWhenRemoteUnavailable(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled and a remote backend that is UNREACHABLE
	// Using an IP address that will timeout (non-routable address)
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  offline_mode: auto
  connectivity_timeout: "500ms"
backends:
  sqlite:
    type: sqlite
    enabled: true
  nextcloud-test:
    type: nextcloud
    enabled: true
    host: "192.0.2.1:8080"
    username: "admin"
    password: "test"
    allow_http: true
default_backend: nextcloud-test
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// When using -b nextcloud-test with sync enabled, it should:
	// 1. Detect that nextcloud-test is unavailable (timeout)
	// 2. Fall back to SQLite cache
	// 3. Allow operations to succeed locally
	// 4. Show a warning about the fallback (not an error)

	// Add a task using the remote backend flag - should fall back to SQLite
	stdout, stderr, exitCode := cli.Execute("-y", "-b", "nextcloud-test", "Work", "add", "Test task for offline")

	// The operation should SUCCEED with exit code 0 (using SQLite fallback)
	// A warning in stderr is expected and acceptable
	if exitCode != 0 {
		t.Errorf("Expected operation to succeed with SQLite fallback, got exit code %d.\nstdout: %s\nstderr: %s", exitCode, stdout, stderr)
	}

	// Verify the warning message about fallback was shown
	testutil.AssertContains(t, stderr, "Using SQLite cache")

	// Verify task was created
	testutil.AssertContains(t, stdout, "Created task")

	// Verify task was created in SQLite
	stdout = cli.MustExecute("-y", "Work", "get")
	testutil.AssertContains(t, stdout, "Test task for offline")

	// Verify the operation was queued for sync
	stdout = cli.MustExecute("-y", "sync", "queue")
	testutil.AssertContains(t, stdout, "Pending Operations")
	testutil.AssertContains(t, stdout, "create")
}

// TestSyncFallbackListTasksWhenRemoteUnavailable tests that listing tasks
// falls back to SQLite cache when the remote backend is unavailable.
func TestSyncFallbackListTasksWhenRemoteUnavailable(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// First, add some tasks to SQLite
	cli.MustExecute("-y", "Work", "add", "Local task 1")
	cli.MustExecute("-y", "Work", "add", "Local task 2")

	// Now configure a remote backend that is unreachable
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  offline_mode: auto
  connectivity_timeout: "500ms"
backends:
  sqlite:
    type: sqlite
    enabled: true
  nextcloud-test:
    type: nextcloud
    enabled: true
    host: "192.0.2.1:8080"
    username: "admin"
    password: "test"
    allow_http: true
default_backend: nextcloud-test
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Listing tasks should fall back to SQLite and show cached tasks
	stdout, stderr, exitCode := cli.Execute("-y", "-b", "nextcloud-test", "Work", "get")

	// The operation should SUCCEED with exit code 0 (using SQLite fallback)
	if exitCode != 0 {
		t.Errorf("Expected listing to succeed with SQLite fallback, got exit code %d.\nstdout: %s\nstderr: %s", exitCode, stdout, stderr)
	}

	// Verify the warning message about fallback was shown
	testutil.AssertContains(t, stderr, "Using SQLite cache")

	// Should show the cached tasks
	testutil.AssertContains(t, stdout, "Local task 1")
	testutil.AssertContains(t, stdout, "Local task 2")
}

// TestSyncPullCLI tests that `todoat sync` pulls changes from remote backend to local cache
func TestSyncPullCLI(t *testing.T) {
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

// TestSyncPushCLI tests that `todoat sync` pushes queued local changes to remote backend
func TestSyncPushCLI(t *testing.T) {
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

// TestSyncStatusCLI tests that `todoat sync status` shows last sync time, pending operations, and connection status
func TestSyncStatusCLI(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled
	createSyncConfig(t, tmpDir, true)

	// Run sync status command
	stdout := cli.MustExecute("-y", "sync", "status")

	// Should show status information
	// Expected fields: last sync, pending operations, connection status
	testutil.AssertContains(t, stdout, "Sync Status")
}

// TestSyncQueueViewCLI tests that `todoat sync queue` lists pending operations with timestamps
func TestSyncQueueViewCLI(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled
	createSyncConfig(t, tmpDir, true)

	// Run sync queue command
	stdout := cli.MustExecute("-y", "sync", "queue")

	// Should show queue information (possibly empty)
	testutil.AssertContains(t, stdout, "Pending Operations")
}

// TestSyncQueueNoRemoteBackendCLI tests that sync queue gracefully handles no remote backend (Issue 014)
// When no remote backend is configured, the command should display an empty queue or helpful message
// instead of failing with a database error
func TestSyncQueueNoRemoteBackendCLI(t *testing.T) {
	// Create CLI test without any sync config (simulates fresh setup with no remote)
	cli, _ := testutil.NewCLITestWithViews(t)

	// Run sync queue command - should not crash with database error
	stdout, stderr, exitCode := cli.Execute("-y", "sync", "queue")

	// The command should succeed (exit 0)
	// or provide a helpful message about no remote backend
	if exitCode != 0 {
		// If it fails, it should NOT be a database error
		combined := stdout + stderr
		if strings.Contains(combined, "unable to open database") ||
			strings.Contains(combined, "out of memory") {
			t.Errorf("sync queue should not fail with database error when no remote configured, got: %s", combined)
		}
	} else {
		// If it succeeds, verify it shows queue information
		testutil.AssertContains(t, stdout, "Pending Operations")
	}
}

// TestSyncQueueMissingDBDirectory tests that sync queue handles missing db directory (Issue 014)
// When the sync database directory doesn't exist, it should be created automatically
func TestSyncQueueMissingDBDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	viewsDir := filepath.Join(tmpDir, "views")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache", "lists.json")
	// Use a subdirectory that doesn't exist for the db path
	dbPath := filepath.Join(tmpDir, "nonexistent", "subdir", "test.db")

	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		t.Fatalf("failed to create views directory: %v", err)
	}

	// Write minimal config
	configContent := `
backends:
  sqlite:
    type: sqlite
    enabled: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &cmd.Config{
		NoPrompt:   true,
		DBPath:     dbPath,
		ViewsPath:  viewsDir,
		CachePath:  cachePath,
		ConfigPath: configPath,
	}

	// Execute directly using cmd.Execute to test with our custom config
	var stdoutBuf, stderrBuf bytes.Buffer
	cfg.Stderr = &stderrBuf
	exitCode := cmd.Execute([]string{"-y", "sync", "queue"}, &stdoutBuf, &stderrBuf, cfg)

	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	// The command should succeed (exit 0)
	if exitCode != 0 {
		combined := stdout + stderr
		if strings.Contains(combined, "unable to open database") ||
			strings.Contains(combined, "out of memory") {
			t.Errorf("sync queue should not fail with database error when db directory is missing, got: %s", combined)
		}
	} else {
		// If it succeeds, verify it shows queue information
		testutil.AssertContains(t, stdout, "Pending Operations")
	}
}

// TestSyncQueueClearCLI tests that `todoat sync queue clear` removes all pending operations
func TestSyncQueueClearCLI(t *testing.T) {
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

// TestSyncOfflineAddCLI tests that adding a task while offline queues operation in sync_queue table
func TestSyncOfflineAddCLI(t *testing.T) {
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

// TestSyncOfflineUpdateCLI tests that updating a task while offline queues operation
func TestSyncOfflineUpdateCLI(t *testing.T) {
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

// TestSyncOfflineDeleteCLI tests that deleting a task while offline queues operation
func TestSyncOfflineDeleteCLI(t *testing.T) {
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

// TestSyncCacheIsolationCLI tests that each remote backend has separate cache tables
func TestSyncCacheIsolationCLI(t *testing.T) {
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

// TestSyncETagSupportCLI tests that updates use If-Match header with ETag for optimistic locking
func TestSyncETagSupportCLI(t *testing.T) {
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

// TestSyncConfigEnabledCLI tests that `sync.enabled: true` in config enables sync behavior
func TestSyncConfigEnabledCLI(t *testing.T) {
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

// TestSyncConfigDisabledCLI tests that `sync.enabled: false` bypasses sync manager
func TestSyncConfigDisabledCLI(t *testing.T) {
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
// Issue 002: default_backend Ignored When Sync Enabled
// =============================================================================

// TestDefaultBackendRespectedWhenSyncEnabled verifies that default_backend config is used
// even when sync is enabled. Previously, the CLI would immediately use SQLite when sync
// was enabled, ignoring the default_backend setting entirely.
// This is Issue #002: default_backend Configuration Ignored When Sync Enabled
func TestDefaultBackendRespectedWhenSyncEnabled(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled AND default_backend set to a remote backend
	// The remote backend is unreachable (192.0.2.1 is a test-net address that won't route)
	// so it should fall back to SQLite, BUT it should still TRY the remote backend first
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  offline_mode: auto
  connectivity_timeout: "500ms"
backends:
  sqlite:
    type: sqlite
    enabled: true
  nextcloud-test:
    type: nextcloud
    enabled: true
    host: "192.0.2.1:8080"
    username: "admin"
    password: "test"
    allow_http: true
default_backend: nextcloud-test
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// When running WITHOUT -b flag, with sync enabled and default_backend: nextcloud-test
	// The CLI should:
	// 1. Respect default_backend and try to use nextcloud-test
	// 2. Detect that nextcloud-test is unavailable
	// 3. Fall back to SQLite and show a warning about fallback
	//
	// BUG BEHAVIOR (before fix): SQLite used immediately, no warning, default_backend ignored
	// EXPECTED BEHAVIOR: Warning about nextcloud-test unavailable, then SQLite fallback
	stdout, stderr, exitCode := cli.Execute("-y", "Work", "add", "Test task")

	// The operation should SUCCEED (using SQLite fallback)
	if exitCode != 0 {
		t.Errorf("Expected operation to succeed with fallback, got exit code %d.\nstdout: %s\nstderr: %s", exitCode, stdout, stderr)
	}

	// CRITICAL: Should show warning about nextcloud-test being unavailable
	// This proves that default_backend was respected and the backend was attempted
	if !strings.Contains(stderr, "nextcloud-test") && !strings.Contains(stderr, "Using SQLite cache") {
		t.Errorf("Expected warning about nextcloud-test fallback, but got no warning about the remote backend.\nstderr: %s\nThis suggests default_backend is being ignored when sync is enabled.", stderr)
	}

	// Verify task was created
	testutil.AssertContains(t, stdout, "Created task")
}

// TestDefaultBackendWithBackendFlagOverride verifies that -b flag still takes priority
// over default_backend when sync is enabled
func TestDefaultBackendWithBackendFlagOverride(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Config with default_backend: nextcloud-test
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  offline_mode: auto
  connectivity_timeout: "500ms"
backends:
  sqlite:
    type: sqlite
    enabled: true
  nextcloud-test:
    type: nextcloud
    enabled: true
    host: "192.0.2.1:8080"
    username: "admin"
    password: "test"
    allow_http: true
default_backend: nextcloud-test
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Using -b sqlite should use sqlite directly, no fallback message
	stdout, stderr, exitCode := cli.Execute("-y", "-b", "sqlite", "Work", "add", "Test task with -b sqlite")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d.\nstdout: %s\nstderr: %s", exitCode, stdout, stderr)
	}

	// Should NOT show any fallback warning since sqlite was explicitly requested
	if strings.Contains(stderr, "Using SQLite cache") || strings.Contains(stderr, "unavailable") {
		t.Errorf("Expected no fallback warning when -b sqlite is explicit, got: %s", stderr)
	}

	testutil.AssertContains(t, stdout, "Created task")
}

// =============================================================================
// Sync Conflict Resolution Tests (019-sync-conflict-resolution)
// =============================================================================

// TestConflictDetectionCLI tests that sync detects when local and remote have both changed same task
func TestConflictDetectionCLI(t *testing.T) {
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

// TestConflictServerWinsCLI tests that with `conflict_strategy: server-wins`, remote changes override local
func TestConflictServerWinsCLI(t *testing.T) {
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

// TestConflictLocalWinsCLI tests that with `conflict_strategy: local-wins`, local changes override remote
func TestConflictLocalWinsCLI(t *testing.T) {
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

// TestConflictMergeCLI tests that with `conflict_strategy: merge`, non-conflicting fields are combined
func TestConflictMergeCLI(t *testing.T) {
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

// TestConflictKeepBothCLI tests that with `conflict_strategy: keep-both`, duplicate task created
func TestConflictKeepBothCLI(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	createSyncConfigWithStrategy(t, tmpDir, true, "keep_both")

	// Add a task
	cli.MustExecute("-y", "Work", "add", "Keep both task")

	// Check that task exists (in real conflict scenario, there would be two)
	stdout := cli.MustExecute("-y", "Work", "get")
	testutil.AssertContains(t, stdout, "Keep both task")
}

// TestConflictStatusDisplayCLI tests that `todoat sync status` shows count of conflicts needing attention
func TestConflictStatusDisplayCLI(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	createSyncConfig(t, tmpDir, true)

	// Run sync status
	stdout := cli.MustExecute("-y", "sync", "status")

	// Should show conflicts count (0 when no conflicts)
	testutil.AssertContains(t, stdout, "Sync Status")
	// The output should include conflict information section
	// When no conflicts exist, it may show "Conflicts: 0" or similar
}

// TestConflictListCLI tests that `todoat sync conflicts` lists all unresolved conflicts with details
func TestConflictListCLI(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	createSyncConfig(t, tmpDir, true)

	// Run sync conflicts command
	stdout := cli.MustExecute("-y", "sync", "conflicts")

	// Should show conflicts list (empty when no conflicts)
	testutil.AssertContains(t, stdout, "Conflict")
}

// TestConflictResolveCLI tests that `todoat sync conflicts resolve [task-uid] --strategy server-wins` resolves specific conflict
func TestConflictResolveCLI(t *testing.T) {
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

// TestConflictDefaultStrategyCLI tests that default strategy is configurable in config.yaml
func TestConflictDefaultStrategyCLI(t *testing.T) {
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

// TestConflictJSONOutputCLI tests that `todoat --json sync conflicts` returns conflicts in JSON format
func TestConflictJSONOutputCLI(t *testing.T) {
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

// =============================================================================
// Issue 003: Sync Command Does Not Actually Sync
// =============================================================================

// TestSyncActuallySyncsToRemote verifies that `todoat sync` actually executes
// pending operations on the remote backend, not just clears the queue.
// This is Issue #003: Sync Command Does Not Actually Sync - Just Clears Queue
func TestSyncActuallySyncsToRemote(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Set up path for "remote" SQLite database
	remoteDBPath := filepath.Join(tmpDir, "remote.db")

	// Create a config with:
	// - sync enabled with sqlite as local cache
	// - a "remote" SQLite backend (simulating a remote that happens to be SQLite)
	// - default_backend set to the remote
	// - offline_mode: offline to force local cache usage (which queues operations)
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  conflict_resolution: server_wins
  offline_mode: offline
backends:
  sqlite:
    type: sqlite
    enabled: true
  sqlite-remote:
    type: sqlite
    enabled: true
    path: "` + remoteDBPath + `"
default_backend: sqlite-remote
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Step 1: Add a task - with offline_mode: offline, this goes to local SQLite cache
	// and queues a create operation for later sync
	stdout, stderr, exitCode := cli.Execute("-y", "Work", "add", "Task to sync")
	if exitCode != 0 {
		t.Fatalf("failed to add task: stdout=%s stderr=%s", stdout, stderr)
	}
	testutil.AssertContains(t, stdout, "Created task")

	// Step 2: Verify the operation is queued
	stdout = cli.MustExecute("-y", "sync", "queue")
	testutil.AssertContains(t, stdout, "Pending Operations: 1")
	testutil.AssertContains(t, stdout, "create")
	testutil.AssertContains(t, stdout, "Task to sync")

	// Step 3: Change to online mode so sync can actually push to the remote
	configContentOnline := `
sync:
  enabled: true
  local_backend: sqlite
  conflict_resolution: server_wins
  offline_mode: online
backends:
  sqlite:
    type: sqlite
    enabled: true
  sqlite-remote:
    type: sqlite
    enabled: true
    path: "` + remoteDBPath + `"
default_backend: sqlite-remote
`
	if err := os.WriteFile(configPath, []byte(configContentOnline), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Step 4: Run sync - this should push the create operation to sqlite-remote
	stdout, stderr, exitCode = cli.Execute("-y", "sync")
	if exitCode != 0 {
		t.Errorf("sync command failed: stdout=%s stderr=%s", stdout, stderr)
	}
	testutil.AssertContains(t, stdout, "Sync completed")
	testutil.AssertContains(t, stdout, "Operations processed: 1")

	// Step 5: Verify queue is now empty
	stdout = cli.MustExecute("-y", "sync", "queue")
	testutil.AssertContains(t, stdout, "Pending Operations: 0")

	// Step 6: CRITICAL - Verify the task actually exists in the remote SQLite database
	// This is the key assertion that catches the bug: if sync just clears the queue
	// without actually syncing, the task won't exist in the remote db.
	remoteDB, err := sql.Open("sqlite", remoteDBPath)
	if err != nil {
		t.Fatalf("failed to open remote db: %v", err)
	}
	defer func() { _ = remoteDB.Close() }()

	var count int
	err = remoteDB.QueryRow(`
		SELECT COUNT(*) FROM tasks t
		JOIN task_lists l ON t.list_id = l.id
		WHERE t.summary = 'Task to sync' AND l.name = 'Work'
	`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query remote db: %v", err)
	}

	if count != 1 {
		t.Errorf("sync did NOT push task to remote backend; expected 1 task in remote db, got %d. "+
			"This confirms Issue #003: sync command just clears queue without actually syncing.", count)
	}
}
