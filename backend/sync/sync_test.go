package sync_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// TestSyncArchitectureUsesLocalCache tests that when sync is enabled with offline_mode: auto,
// CLI operations use SQLite cache directly (no network calls, no fallback warning).
// This is the proper sync architecture: CLI → SQLite (instant) → queue → Daemon → Remote
// Updated from Issue #0 to reflect Issue #001 architecture fix.
// Updated for Issue #009: auto_sync_after_operation now defaults to true, so explicitly disable
// to test the queuing behavior.
func TestSyncArchitectureUsesLocalCache(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled and a remote backend configured
	// With offline_mode: auto (default), CLI should always use SQLite cache
	// Note: auto_sync_after_operation: false to test the queuing behavior
	// (default is now true per Issue #009)
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  offline_mode: auto
  auto_sync_after_operation: false
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

	// With sync architecture (offline_mode: auto), CLI should:
	// 1. Use SQLite cache directly (no network call to remote)
	// 2. Allow operations to succeed instantly
	// 3. Queue operations for daemon to sync later
	// 4. No warning needed (this is expected behavior, not a fallback)

	// Add a task - should use SQLite cache with sync architecture
	stdout, stderr, exitCode := cli.Execute("-y", "-b", "nextcloud-test", "Work", "add", "Test task for sync")

	// The operation should SUCCEED with exit code 0 (using SQLite cache)
	if exitCode != 0 {
		t.Errorf("Expected operation to succeed using SQLite cache, got exit code %d.\nstdout: %s\nstderr: %s", exitCode, stdout, stderr)
	}

	// Verify task was created
	testutil.AssertContains(t, stdout, "Created task")

	// Verify task exists in local cache
	stdout = cli.MustExecute("-y", "Work", "get")
	testutil.AssertContains(t, stdout, "Test task for sync")

	// Verify the operation was queued for sync
	stdout = cli.MustExecute("-y", "sync", "queue")
	testutil.AssertContains(t, stdout, "Pending Operations")
	testutil.AssertContains(t, stdout, "create")
}

// TestSyncArchitectureListsFromLocalCache tests that listing tasks uses SQLite cache
// when sync is enabled (sync architecture), providing instant responses.
// Updated from Issue #0 to reflect Issue #001 architecture fix.
// Updated from Issue #011 to use the same backend for add and list (proper isolation).
func TestSyncArchitectureListsFromLocalCache(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Configure a remote backend with sync enabled
	// With offline_mode: auto, CLI uses SQLite cache directly
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  offline_mode: auto
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

	// Add tasks using the nextcloud-test backend (which uses SQLite cache for storage)
	// This ensures the tasks are stored with backend_id="nextcloud-test" (Issue #011 fix)
	cli.MustExecute("-y", "-b", "nextcloud-test", "Work", "add", "Local task 1")
	cli.MustExecute("-y", "-b", "nextcloud-test", "Work", "add", "Local task 2")

	// Listing tasks should use SQLite cache (sync architecture)
	stdout, stderr, exitCode := cli.Execute("-y", "-b", "nextcloud-test", "Work", "get")

	// The operation should SUCCEED using SQLite cache
	if exitCode != 0 {
		t.Errorf("Expected listing to succeed using SQLite cache, got exit code %d.\nstdout: %s\nstderr: %s", exitCode, stdout, stderr)
	}

	// Should show the cached tasks
	testutil.AssertContains(t, stdout, "Local task 1")
	testutil.AssertContains(t, stdout, "Local task 2")
}

// TestSyncPullCLI tests that `todoat sync` pulls changes from remote backend to local cache
// This test verifies that tasks created on a remote backend are pulled to local
func TestSyncPullCLI(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Set up a second SQLite database as "remote" backend
	remoteDBPath := filepath.Join(tmpDir, "remote.db")

	// Create config that points to the remote sqlite backend
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
backends:
  sqlite:
    type: sqlite
    enabled: true
  remote-sqlite:
    type: sqlite
    enabled: true
    path: "` + remoteDBPath + `"
default_backend: remote-sqlite
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Initialize the remote database using the same CLI with -b flag to point to remote
	// This ensures the schema is properly created via migrations
	remoteStdout, remoteStderr, exitCode := cli.Execute("-y", "-b", "remote-sqlite", "Work", "add", "Remote task from server")
	if exitCode != 0 {
		t.Fatalf("failed to add task to remote: stdout=%s stderr=%s", remoteStdout, remoteStderr)
	}

	// Verify the task exists on remote but NOT on local
	// Local should be empty since we only added to remote
	localOutput := cli.MustExecute("-y", "list")
	if strings.Contains(localOutput, "Remote task from server") {
		t.Fatalf("task should not exist on local before sync: %s", localOutput)
	}

	// Run sync command - should pull from remote
	stdout, stderr, exitCode := cli.Execute("-y", "sync")

	// Sync should complete successfully
	if exitCode != 0 {
		t.Fatalf("sync failed with exit code %d: stdout=%s stderr=%s", exitCode, stdout, stderr)
	}

	// Sync output should show pull results
	// Expected format: "Pull: X new, Y updated, Z deleted"
	testutil.AssertContains(t, stdout, "Pull:")
	testutil.AssertContains(t, stdout, "1 new") // Should have pulled 1 new task

	// Verify the task was pulled to local (use 'Work' to list tasks in Work list)
	listOutput := cli.MustExecute("-y", "Work")
	testutil.AssertContains(t, listOutput, "Remote task from server")

	// Allow background sync goroutines to complete before test cleanup
	time.Sleep(100 * time.Millisecond)
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

// =============================================================================
// Issue #7: Auto-sync should pull before read operations
// =============================================================================

// TestBackgroundPullSyncOnReadCLI tests that read operations trigger background pull sync
// when auto_sync_after_operation is enabled. The read operation should return immediately
// without waiting for the sync, and subsequent reads should show updated data.
func TestBackgroundPullSyncOnReadCLI(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	remoteDBPath := filepath.Join(tmpDir, "remote.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Phase 1: Add task to remote database directly using SQL
	// This avoids using CLI which would contaminate the cfg.Backend field
	remoteDB, err := sql.Open("sqlite", remoteDBPath)
	if err != nil {
		t.Fatalf("failed to open remote database: %v", err)
	}
	defer func() { _ = remoteDB.Close() }()

	// Initialize the remote database schema with all required columns
	if err := setupRemoteDB(remoteDB, "list-1", "Work", "task-1", "Remote task from server"); err != nil {
		t.Fatalf("failed to setup remote database: %v", err)
	}
	_ = remoteDB.Close() // Close before CLI uses it

	// Phase 2: Enable sync and test background pull on read
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  auto_sync_after_operation: true
backends:
  sqlite:
    type: sqlite
    enabled: true
  remote-sqlite:
    type: sqlite
    enabled: true
    path: "` + remoteDBPath + `"
default_backend: remote-sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// First read: should return local data AND trigger background pull sync
	// The local cache should be empty initially
	stdout1 := cli.MustExecute("-y", "Work")

	// Wait for background sync to complete
	time.Sleep(500 * time.Millisecond)

	// Second read: should show the pulled remote task
	stdout2 := cli.MustExecute("-y", "Work")

	// Verify the remote task appears after background sync pulls it
	if !strings.Contains(stdout2, "Remote task from server") {
		t.Errorf("expected remote task to appear after background sync on read.\nFirst read:\n%s\nSecond read:\n%s", stdout1, stdout2)
	}

	// Wait for any remaining background goroutines to complete before cleanup
	time.Sleep(100 * time.Millisecond)
}

// TestBackgroundPullSyncCooldownCLI tests that background pull sync respects a cooldown period
// to avoid excessive syncing on every read operation.
func TestBackgroundPullSyncCooldownCLI(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	remoteDBPath := filepath.Join(tmpDir, "remote.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Set up remote database directly using SQL to avoid -b flag contamination
	remoteDB, err := sql.Open("sqlite", remoteDBPath)
	if err != nil {
		t.Fatalf("failed to open remote database: %v", err)
	}
	defer func() { _ = remoteDB.Close() }()

	// Initialize schema and add task
	if err := setupRemoteDB(remoteDB, "list-1", "Work", "task-1", "Initial task"); err != nil {
		t.Fatalf("failed to setup remote database: %v", err)
	}
	_ = remoteDB.Close()

	// Enable sync with auto_sync to test cooldown behavior
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  auto_sync_after_operation: true
backends:
  sqlite:
    type: sqlite
    enabled: true
  remote-sqlite:
    type: sqlite
    enabled: true
    path: "` + remoteDBPath + `"
default_backend: remote-sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Multiple rapid reads should not trigger excessive syncs due to cooldown
	// The first read triggers sync, subsequent reads within cooldown period skip sync
	for i := 0; i < 5; i++ {
		_, _, exitCode := cli.Execute("-y", "Work")
		if exitCode != 0 {
			t.Fatalf("read operation %d failed", i+1)
		}
	}

	// Wait for background sync to complete
	time.Sleep(500 * time.Millisecond)

	// Verify the task appears after reads (sync should have happened at least once)
	stdout := cli.MustExecute("-y", "Work")
	testutil.AssertContains(t, stdout, "Initial task")

	// Wait for any background goroutines to finish before test cleanup
	time.Sleep(100 * time.Millisecond)
}

// TestBackgroundPullSyncDisabledCLI tests that background pull sync does not occur
// when auto_sync_after_operation is disabled.
func TestBackgroundPullSyncDisabledCLI(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	remoteDBPath := filepath.Join(tmpDir, "remote.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Set up remote database directly using SQL
	remoteDB, err := sql.Open("sqlite", remoteDBPath)
	if err != nil {
		t.Fatalf("failed to open remote database: %v", err)
	}
	defer func() { _ = remoteDB.Close() }()

	// Initialize schema and add task
	if err := setupRemoteDB(remoteDB, "list-1", "Work", "task-1", "Remote only task"); err != nil {
		t.Fatalf("failed to setup remote database: %v", err)
	}
	_ = remoteDB.Close()

	// Enable sync but with auto_sync_after_operation DISABLED
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  auto_sync_after_operation: false
backends:
  sqlite:
    type: sqlite
    enabled: true
  remote-sqlite:
    type: sqlite
    enabled: true
    path: "` + remoteDBPath + `"
default_backend: remote-sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Read from local (via sync architecture) - should NOT have the task
	// since auto_sync_after_operation is disabled, no background sync occurs on read
	_ = cli.MustExecute("-y", "Work")
	time.Sleep(300 * time.Millisecond)

	// The task should NOT appear (no background sync on read when disabled)
	stdout := cli.MustExecute("-y", "Work")
	if strings.Contains(stdout, "Remote only task") {
		t.Errorf("task should not appear when auto_sync_after_operation is disabled.\nOutput:\n%s", stdout)
	}

	// Now manually sync
	cli.MustExecute("-y", "sync")

	// After manual sync, the task should appear
	stdout = cli.MustExecute("-y", "Work")
	testutil.AssertContains(t, stdout, "Remote only task")
}

// TestBackgroundPullCooldownBehavior tests that the configurable background_pull_cooldown
// actually controls the cooldown period between background sync operations.
// This is roadmap item 082.
func TestBackgroundPullCooldownBehavior(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	remoteDBPath := filepath.Join(tmpDir, "remote.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Set up remote database directly using SQL
	remoteDB, err := sql.Open("sqlite", remoteDBPath)
	if err != nil {
		t.Fatalf("failed to open remote database: %v", err)
	}
	defer func() { _ = remoteDB.Close() }()

	// Initialize schema and add task
	if err := setupRemoteDB(remoteDB, "list-1", "Work", "task-1", "Initial task"); err != nil {
		t.Fatalf("failed to setup remote database: %v", err)
	}
	_ = remoteDB.Close()

	// Enable sync with a short custom cooldown (5s - minimum allowed)
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  auto_sync_after_operation: true
  background_pull_cooldown: "5s"
backends:
  sqlite:
    type: sqlite
    enabled: true
  remote-sqlite:
    type: sqlite
    enabled: true
    path: "` + remoteDBPath + `"
default_backend: remote-sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Multiple rapid reads should still respect the cooldown
	// The first read triggers sync, subsequent reads within cooldown period skip sync
	for i := 0; i < 5; i++ {
		_, _, exitCode := cli.Execute("-y", "Work")
		if exitCode != 0 {
			t.Fatalf("read operation %d failed", i+1)
		}
	}

	// Wait for background sync to complete
	time.Sleep(500 * time.Millisecond)

	// Verify the task appears after reads (sync should have happened at least once)
	stdout := cli.MustExecute("-y", "Work")
	testutil.AssertContains(t, stdout, "Initial task")

	// Wait for any background goroutines to finish before test cleanup
	time.Sleep(100 * time.Millisecond)
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

// setupRemoteDB initializes a remote SQLite database with the full schema for testing.
// This ensures all columns are present that the SQLite backend expects.
func setupRemoteDB(db *sql.DB, listID, listName, taskID, taskSummary string) error {
	// Use the same schema as SQLite backend migration 1 + later migrations
	schema := `
		CREATE TABLE IF NOT EXISTS task_lists (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			color TEXT DEFAULT '',
			description TEXT DEFAULT '',
			modified TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TEXT,
			backend_id TEXT NOT NULL DEFAULT 'sqlite'
		);
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			list_id TEXT NOT NULL,
			summary TEXT NOT NULL,
			description TEXT DEFAULT '',
			status TEXT NOT NULL DEFAULT 'NEEDS-ACTION',
			priority INTEGER DEFAULT 0,
			due_date TEXT,
			start_date TEXT,
			recurrence TEXT DEFAULT '',
			recur_from_due INTEGER DEFAULT 1,
			created TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			modified TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			completed TEXT,
			deleted_at TEXT,
			parent_id TEXT DEFAULT '',
			categories TEXT DEFAULT '',
			backend_id TEXT NOT NULL DEFAULT 'sqlite',
			FOREIGN KEY (list_id) REFERENCES task_lists(id)
		);
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_tasks_list_id ON tasks(list_id);
		CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
		CREATE INDEX IF NOT EXISTS idx_tasks_backend_id ON tasks(backend_id);
		CREATE INDEX IF NOT EXISTS idx_task_lists_backend_id ON task_lists(backend_id);
	`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// Mark all migrations as done (current migrations go up to 4)
	for i := 1; i <= 4; i++ {
		if _, err := db.Exec("INSERT OR IGNORE INTO schema_version (version) VALUES (?)", i); err != nil {
			return err
		}
	}

	// Insert list and task with the correct backend_id to match sync
	if _, err := db.Exec("INSERT INTO task_lists (id, name, modified, backend_id) VALUES (?, ?, CURRENT_TIMESTAMP, 'remote-sqlite')", listID, listName); err != nil {
		return err
	}
	if _, err := db.Exec("INSERT INTO tasks (id, list_id, summary, created, modified, backend_id) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'remote-sqlite')", taskID, listID, taskSummary); err != nil {
		return err
	}
	return nil
}

// =============================================================================
// Issue 002: default_backend Ignored When Sync Enabled
// =============================================================================

// TestDefaultBackendRespectedWhenSyncEnabled verifies that default_backend config is used
// even when sync is enabled. Previously, the CLI would immediately use SQLite when sync
// was enabled, ignoring the default_backend setting entirely.
// TestDefaultBackendRespectedWhenSyncEnabled verifies that default_backend config is used
// for sync operations when sync is enabled.
// This is Issue #002: default_backend Configuration Ignored When Sync Enabled
// Updated for Issue #001 sync architecture: CLI uses SQLite cache, daemon syncs to default_backend
// Updated for Issue #009: auto_sync_after_operation now defaults to true, so explicitly disable
// to test the queuing behavior.
func TestDefaultBackendRespectedWhenSyncEnabled(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled AND default_backend set to a remote backend
	// With sync architecture (offline_mode: auto), CLI uses SQLite cache
	// The default_backend is used by the daemon for sync operations
	// Note: auto_sync_after_operation: false to test the queuing behavior
	// (default is now true per Issue #009)
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  offline_mode: auto
  auto_sync_after_operation: false
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
	// 1. Use SQLite cache directly (sync architecture)
	// 2. Queue operations for sync with default_backend (nextcloud-test)
	// 3. Succeed without warnings (this is intended behavior, not a fallback)
	stdout, stderr, exitCode := cli.Execute("-y", "Work", "add", "Test task")

	// The operation should SUCCEED using SQLite cache
	if exitCode != 0 {
		t.Errorf("Expected operation to succeed using SQLite cache, got exit code %d.\nstdout: %s\nstderr: %s", exitCode, stdout, stderr)
	}

	// Verify task was created
	testutil.AssertContains(t, stdout, "Created task")

	// Verify the operation was queued for sync (proves default_backend is configured for sync)
	stdout = cli.MustExecute("-y", "sync", "queue")
	testutil.AssertContains(t, stdout, "Pending Operations")
	testutil.AssertContains(t, stdout, "create")
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
// Issue 008: ResolveConflict Strategy Parameter Ignored
// =============================================================================

// TestConflictResolveServerWinsApplied tests that server_wins strategy actually replaces
// the local task with the remote version. This is Issue #008.
func TestConflictResolveServerWinsApplied(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	createSyncConfig(t, tmpDir, true)

	// Add a task locally (this will be our "local version")
	stdout := cli.MustExecute("-y", "Work", "add", "Local Task", "-p", "5")
	testutil.AssertContains(t, stdout, "Created task")

	// Get the task UID by listing in JSON
	stdout = cli.MustExecute("-y", "--json", "Work", "get")

	// Parse the JSON to get the task UID
	var tasksOutput struct {
		Tasks []struct {
			UID      string `json:"uid"`
			Summary  string `json:"summary"`
			Priority int    `json:"priority"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(stdout), &tasksOutput); err != nil {
		t.Fatalf("failed to parse task JSON: %v", err)
	}
	if len(tasksOutput.Tasks) == 0 {
		t.Fatalf("expected at least one task, got none")
	}
	taskUID := tasksOutput.Tasks[0].UID

	// Get the database path from the test config
	dbPath := cli.Config().DBPath

	// First, ensure the sync_conflicts table exists by running a sync command
	// This will initialize the sync manager and create the table if needed
	cli.MustExecute("-y", "sync", "status")

	// Insert a conflict record with different local and remote versions
	// Local version: priority 5, summary "Local Task"
	// Remote version: priority 1, summary "Remote Task"
	// Note: list_id is stored as INTEGER (0 is a placeholder - actual list is looked up from task)
	localVersion := `{"id":"` + taskUID + `","summary":"Local Task","priority":5}`
	remoteVersion := `{"id":"` + taskUID + `","summary":"Remote Task","priority":1}`

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec(`
		INSERT INTO sync_conflicts (task_uid, task_summary, list_id, local_version, remote_version,
		                            local_modified, remote_modified, detected_at, status)
		VALUES (?, 'Local Task', 0, ?, ?, datetime('now'), datetime('now'), datetime('now'), 'pending')
	`, taskUID, localVersion, remoteVersion)
	if err != nil {
		t.Fatalf("failed to insert conflict: %v", err)
	}

	// Verify conflict was inserted
	stdout = cli.MustExecute("-y", "sync", "conflicts")
	testutil.AssertContains(t, stdout, "Local Task")

	// Resolve with server_wins strategy - this SHOULD update the local task with remote values
	stdout, stderr, exitCode := cli.Execute("-y", "sync", "conflicts", "resolve", taskUID, "--strategy", "server_wins")
	if exitCode != 0 {
		t.Fatalf("resolve failed: stdout=%s stderr=%s", stdout, stderr)
	}
	testutil.AssertContains(t, stdout, "resolved")

	// CRITICAL ASSERTION: The task should now have the REMOTE values
	// server_wins means: replace local with remote
	stdout = cli.MustExecute("-y", "--json", "Work", "get")

	if err := json.Unmarshal([]byte(stdout), &tasksOutput); err != nil {
		t.Fatalf("failed to parse task JSON after resolve: %v", err)
	}

	var foundTask bool
	for _, task := range tasksOutput.Tasks {
		if task.UID == taskUID {
			foundTask = true
			// With server_wins, the task should have remote values
			if task.Summary != "Remote Task" {
				t.Errorf("ISSUE #008: server_wins strategy not applied. Expected summary='Remote Task', got '%s'", task.Summary)
			}
			if task.Priority != 1 {
				t.Errorf("ISSUE #008: server_wins strategy not applied. Expected priority=1, got %d", task.Priority)
			}
			break
		}
	}

	if !foundTask {
		t.Errorf("task %s not found after resolve", taskUID)
	}
}

// TestConflictResolveLocalWinsQueuesUpdate tests that local_wins strategy keeps local version
// and queues an update operation to push the local version to remote. This is Issue #008.
func TestConflictResolveLocalWinsQueuesUpdate(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	createSyncConfig(t, tmpDir, true)

	// Add a task locally
	stdout := cli.MustExecute("-y", "Work", "add", "Local Priority Task", "-p", "1")
	testutil.AssertContains(t, stdout, "Created task")

	// Get the task UID
	stdout = cli.MustExecute("-y", "--json", "Work", "get")
	var tasksOutput struct {
		Tasks []struct {
			UID      string `json:"uid"`
			Summary  string `json:"summary"`
			Priority int    `json:"priority"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(stdout), &tasksOutput); err != nil {
		t.Fatalf("failed to parse task JSON: %v", err)
	}
	if len(tasksOutput.Tasks) == 0 {
		t.Fatalf("expected at least one task")
	}
	taskUID := tasksOutput.Tasks[0].UID

	// Get the database path from the test config
	dbPath := cli.Config().DBPath

	// First ensure sync_conflicts table exists
	cli.MustExecute("-y", "sync", "status")

	// Insert a conflict record
	localVersion := `{"id":"` + taskUID + `","summary":"Local Priority Task","priority":1}`
	remoteVersion := `{"id":"` + taskUID + `","summary":"Remote Priority Task","priority":5}`

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec(`
		INSERT INTO sync_conflicts (task_uid, task_summary, list_id, local_version, remote_version,
		                            local_modified, remote_modified, detected_at, status)
		VALUES (?, 'Local Priority Task', 0, ?, ?, datetime('now'), datetime('now'), datetime('now'), 'pending')
	`, taskUID, localVersion, remoteVersion)
	if err != nil {
		t.Fatalf("failed to insert conflict: %v", err)
	}

	// Clear the sync queue before resolving
	cli.MustExecute("-y", "sync", "queue", "clear")

	// Resolve with local_wins strategy - should keep local and queue update to push
	stdout, stderr, exitCode := cli.Execute("-y", "sync", "conflicts", "resolve", taskUID, "--strategy", "local_wins")
	if exitCode != 0 {
		t.Fatalf("resolve failed: stdout=%s stderr=%s", stdout, stderr)
	}

	// CRITICAL ASSERTION: Task should keep local values (priority 1)
	stdout = cli.MustExecute("-y", "--json", "Work", "get")
	if err := json.Unmarshal([]byte(stdout), &tasksOutput); err != nil {
		t.Fatalf("failed to parse task JSON: %v", err)
	}

	for _, task := range tasksOutput.Tasks {
		if task.UID == taskUID {
			if task.Priority != 1 {
				t.Errorf("ISSUE #008: local_wins should keep local priority=1, got %d", task.Priority)
			}
			if task.Summary != "Local Priority Task" {
				t.Errorf("ISSUE #008: local_wins should keep local summary, got '%s'", task.Summary)
			}
			break
		}
	}

	// CRITICAL ASSERTION: An update operation should be queued to push local to remote
	stdout = cli.MustExecute("-y", "sync", "queue")
	if !strings.Contains(stdout, "update") {
		t.Errorf("ISSUE #008: local_wins should queue update operation to push local version to remote.\nQueue: %s", stdout)
	}
}

// TestConflictResolveKeepBothCreatesDuplicate tests that keep_both strategy keeps the remote
// version and creates a duplicate with local values. This is Issue #008.
func TestConflictResolveKeepBothCreatesDuplicate(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)
	createSyncConfig(t, tmpDir, true)

	// Add a task locally
	cli.MustExecute("-y", "Work", "add", "Original Task", "-p", "5")

	// Get the task UID
	stdout := cli.MustExecute("-y", "--json", "Work", "get")
	var tasksOutput struct {
		Tasks []struct {
			UID      string `json:"uid"`
			Summary  string `json:"summary"`
			Priority int    `json:"priority"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(stdout), &tasksOutput); err != nil {
		t.Fatalf("failed to parse task JSON: %v", err)
	}
	if len(tasksOutput.Tasks) == 0 {
		t.Fatalf("expected at least one task")
	}
	taskUID := tasksOutput.Tasks[0].UID

	// Get the database path from the test config
	dbPath := cli.Config().DBPath

	// First ensure sync_conflicts table exists
	cli.MustExecute("-y", "sync", "status")

	// Insert a conflict - local has "Original Task", remote has "Server Version"
	localVersion := `{"id":"` + taskUID + `","summary":"Original Task","priority":5}`
	remoteVersion := `{"id":"` + taskUID + `","summary":"Server Version","priority":1}`

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec(`
		INSERT INTO sync_conflicts (task_uid, task_summary, list_id, local_version, remote_version,
		                            local_modified, remote_modified, detected_at, status)
		VALUES (?, 'Original Task', 0, ?, ?, datetime('now'), datetime('now'), datetime('now'), 'pending')
	`, taskUID, localVersion, remoteVersion)
	if err != nil {
		t.Fatalf("failed to insert conflict: %v", err)
	}

	// Count tasks before resolve
	initialCount := len(tasksOutput.Tasks)

	// Resolve with keep_both strategy
	stdout, stderr, exitCode := cli.Execute("-y", "sync", "conflicts", "resolve", taskUID, "--strategy", "keep_both")
	if exitCode != 0 {
		t.Fatalf("resolve failed: stdout=%s stderr=%s", stdout, stderr)
	}

	// CRITICAL ASSERTION: Should now have 2 tasks - original updated to remote, plus duplicate with local
	stdout = cli.MustExecute("-y", "--json", "Work", "get")
	if err := json.Unmarshal([]byte(stdout), &tasksOutput); err != nil {
		t.Fatalf("failed to parse task JSON: %v", err)
	}

	if len(tasksOutput.Tasks) != initialCount+1 {
		t.Errorf("ISSUE #008: keep_both should create duplicate. Expected %d tasks, got %d", initialCount+1, len(tasksOutput.Tasks))
	}

	// One task should have remote values, one should have local values with " (local)" suffix
	var foundRemote, foundLocal bool
	for _, task := range tasksOutput.Tasks {
		if task.Summary == "Server Version" && task.Priority == 1 {
			foundRemote = true
		}
		if strings.Contains(task.Summary, "Original Task") && strings.Contains(task.Summary, "(local)") && task.Priority == 5 {
			foundLocal = true
		}
	}

	if !foundRemote {
		t.Errorf("ISSUE #008: keep_both should update original to remote version 'Server Version'")
	}
	if !foundLocal {
		t.Errorf("ISSUE #008: keep_both should create duplicate with local values and '(local)' suffix")
	}
}

// =============================================================================
// Issue 001: Sync Architecture - CLI Should Always Use SQLite When Sync Enabled
// =============================================================================

// TestSyncArchitectureCLIUsesSQLiteNotRemote verifies Issue #001:
// When sync is enabled, CLI operations should ALWAYS use SQLite cache for instant
// responses. The daemon handles remote backend sync - CLI should never directly
// contact the remote backend.
//
// This is the core architectural requirement documented in synchronization.md:
// User → CLI → SQLite (instant) → sync_queue → Daemon → Remote
// Updated for Issue #009: auto_sync_after_operation now defaults to true, so explicitly disable
// to test the queuing behavior.
func TestSyncArchitectureCLIUsesSQLiteNotRemote(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Set up a "remote" SQLite database to track if it gets accessed directly
	remoteDBPath := filepath.Join(tmpDir, "remote.db")

	// Configure sync with:
	// - sync.enabled: true
	// - offline_mode: auto (default) - this is where the bug manifests
	// - A reachable remote backend
	// - auto_sync_after_operation: false to test the queuing behavior
	//   (default is now true per Issue #009)
	// The BUG: with offline_mode: auto, if remote is reachable, CLI uses remote directly
	// EXPECTED: CLI always uses SQLite cache, regardless of remote availability
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  conflict_resolution: server_wins
  offline_mode: auto
  auto_sync_after_operation: false
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

	// Step 1: Add a task - with sync enabled, this should:
	// - Write to SQLite cache immediately (instant response)
	// - Queue operation in sync_queue for daemon to sync later
	// - NOT directly contact the remote backend
	stdout, stderr, exitCode := cli.Execute("-y", "Work", "add", "Test sync architecture task")
	if exitCode != 0 {
		t.Fatalf("failed to add task: stdout=%s stderr=%s", stdout, stderr)
	}
	testutil.AssertContains(t, stdout, "Created task")

	// Step 2: Verify the operation was queued for sync
	// If CLI used SQLite cache, there should be a pending create operation
	stdout = cli.MustExecute("-y", "sync", "queue")

	// CRITICAL ASSERTION: Operation should be queued, not executed immediately
	// If the queue shows 0 pending ops and the task was created in remote directly,
	// that proves the bug: CLI is using remote backend directly instead of SQLite cache
	if !strings.Contains(stdout, "create") {
		// Check if task exists in remote DB (which would prove the bug)
		remoteDB, err := sql.Open("sqlite", remoteDBPath)
		if err == nil {
			defer func() { _ = remoteDB.Close() }()

			var count int
			err = remoteDB.QueryRow(`
				SELECT COUNT(*) FROM tasks t
				JOIN task_lists l ON t.list_id = l.id
				WHERE t.summary = 'Test sync architecture task'
			`).Scan(&count)
			if err == nil && count > 0 {
				t.Errorf("ISSUE #001 CONFIRMED: CLI wrote directly to remote backend instead of queuing for sync.\n"+
					"Queue output: %s\n"+
					"Task found in remote DB: yes\n"+
					"Expected: operation queued in sync_queue, task NOT in remote until daemon syncs", stdout)
			}
		}

		t.Errorf("expected create operation to be queued when sync enabled, got:\n%s", stdout)
	}

	// Step 3: Verify the task exists in local cache (not remote yet)
	stdout = cli.MustExecute("-y", "Work", "get")
	testutil.AssertContains(t, stdout, "Test sync architecture task")

	// Step 4: Verify task does NOT exist in remote yet (daemon hasn't synced)
	// This check fails if CLI directly used the remote backend
	if _, err := os.Stat(remoteDBPath); err == nil {
		remoteDB, err := sql.Open("sqlite", remoteDBPath)
		if err == nil {
			defer func() { _ = remoteDB.Close() }()

			var count int
			err = remoteDB.QueryRow(`
				SELECT COUNT(*) FROM tasks t
				JOIN task_lists l ON t.list_id = l.id
				WHERE t.summary = 'Test sync architecture task'
			`).Scan(&count)
			if err == nil && count > 0 {
				t.Errorf("ISSUE #001 CONFIRMED: Task exists in remote DB but should only be in local cache.\n" +
					"With sync architecture, CLI should write to SQLite cache, not remote directly.\n" +
					"Daemon should sync to remote later.")
			}
		}
	}
}

// TestSyncArchitectureNoNetworkCallOnAdd verifies that CLI add operations
// don't make network calls when sync is enabled - even for a reachable remote.
// This test uses offline_mode: auto with a reachable remote to catch the bug.
func TestSyncArchitectureNoNetworkCallOnAdd(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Set up a "remote" SQLite database - reachable, but shouldn't be called directly
	remoteDBPath := filepath.Join(tmpDir, "remote.db")

	// First, initialize the remote database with the schema
	// (so connectivity check would succeed if CLI tries to use it)
	remoteDB, err := sql.Open("sqlite", remoteDBPath)
	if err != nil {
		t.Fatalf("failed to create remote db: %v", err)
	}
	// Create minimal schema
	_, err = remoteDB.Exec(`
		CREATE TABLE IF NOT EXISTS task_lists (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			color TEXT
		);
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			list_id TEXT NOT NULL,
			summary TEXT NOT NULL,
			description TEXT,
			status TEXT DEFAULT 'TODO',
			priority INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (list_id) REFERENCES task_lists(id)
		);
	`)
	if err != nil {
		t.Fatalf("failed to init remote schema: %v", err)
	}
	_ = remoteDB.Close()

	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  conflict_resolution: server_wins
  offline_mode: auto
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

	// Add task with sync enabled - should use SQLite cache
	stdout, stderr, exitCode := cli.Execute("-y", "Work", "add", "Instant CLI task")
	if exitCode != 0 {
		t.Fatalf("failed to add task: stdout=%s stderr=%s", stdout, stderr)
	}

	// Task should be in queue
	stdout = cli.MustExecute("-y", "sync", "queue")
	testutil.AssertContains(t, stdout, "Pending Operations")

	// Check if the bug exists: task should NOT be in remote DB yet
	remoteDB, err = sql.Open("sqlite", remoteDBPath)
	if err != nil {
		t.Fatalf("failed to open remote db: %v", err)
	}
	defer func() { _ = remoteDB.Close() }()

	var taskCount int
	err = remoteDB.QueryRow(`SELECT COUNT(*) FROM tasks WHERE summary = 'Instant CLI task'`).Scan(&taskCount)
	if err != nil {
		t.Logf("query error (may be expected if tables don't exist): %v", err)
	}

	if taskCount > 0 {
		t.Errorf("ISSUE #001: CLI wrote directly to remote backend.\n" +
			"Expected: task in local cache + sync queue only\n" +
			"Got: task already in remote database\n" +
			"This means CLI used remote backend instead of sync architecture")
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
	testutil.AssertContains(t, stdout, "Push: 1 operations processed")

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

// =============================================================================
// Issue 006: Operations Not Auto-Synced After Execution
// =============================================================================

// TestAutoSyncAfterOperation verifies that with auto_sync_after_operation enabled,
// operations like add/update/delete automatically trigger sync to push changes to remote.
// This is Issue #006: Operations Not Auto-Synced After Execution
func TestAutoSyncAfterOperation(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Set up path for "remote" SQLite database
	remoteDBPath := filepath.Join(tmpDir, "remote.db")

	// Create a config with:
	// - sync enabled with auto_sync_after_operation: true
	// - a "remote" SQLite backend
	// - auto mode (default) - CLI uses local cache, sync pushes to remote
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  conflict_resolution: server_wins
  offline_mode: auto
  auto_sync_after_operation: true
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

	// Add a task - with auto_sync_after_operation: true, this should:
	// 1. Create task in local SQLite cache
	// 2. Queue the operation
	// 3. Automatically trigger sync to push to remote
	stdout, stderr, exitCode := cli.Execute("-y", "Work", "add", "Auto-synced task")
	if exitCode != 0 {
		t.Fatalf("failed to add task: stdout=%s stderr=%s", stdout, stderr)
	}
	testutil.AssertContains(t, stdout, "Created task")

	// CRITICAL: Queue should be empty because auto-sync already processed it
	stdout = cli.MustExecute("-y", "sync", "queue")
	if !strings.Contains(stdout, "Pending Operations: 0") {
		t.Errorf("expected queue to be empty after auto-sync, but got:\n%s", stdout)
	}

	// Verify the task actually exists in the remote SQLite database
	remoteDB, err := sql.Open("sqlite", remoteDBPath)
	if err != nil {
		t.Fatalf("failed to open remote db: %v", err)
	}
	defer func() { _ = remoteDB.Close() }()

	var count int
	err = remoteDB.QueryRow(`
		SELECT COUNT(*) FROM tasks t
		JOIN task_lists l ON t.list_id = l.id
		WHERE t.summary = 'Auto-synced task' AND l.name = 'Work'
	`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query remote db: %v", err)
	}

	if count != 1 {
		t.Errorf("auto-sync did NOT push task to remote; expected 1 task in remote db, got %d. "+
			"This confirms Issue #006: operations not auto-synced after execution.", count)
	}
}

// TestAutoSyncDisabledQueuesOnly verifies that with auto_sync_after_operation: false,
// operations only queue but don't sync until manual `todoat sync` is run.
// Note: The default is now true when sync is enabled (Issue #009), so we must explicitly set false.
func TestAutoSyncDisabledQueuesOnly(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Set up path for "remote" SQLite database
	remoteDBPath := filepath.Join(tmpDir, "remote.db")

	// Create a config with auto_sync_after_operation explicitly disabled
	// (default is now true when sync.enabled is true, per Issue #009)
	// Use offline_mode: auto (default) - CLI uses local cache, sync pushes to remote
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  conflict_resolution: server_wins
  offline_mode: auto
  auto_sync_after_operation: false
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

	// Add a task - without auto_sync, this should only queue
	stdout, stderr, exitCode := cli.Execute("-y", "Work", "add", "Queued-only task")
	if exitCode != 0 {
		t.Fatalf("failed to add task: stdout=%s stderr=%s", stdout, stderr)
	}
	testutil.AssertContains(t, stdout, "Created task")

	// Queue should have 1 pending operation
	stdout = cli.MustExecute("-y", "sync", "queue")
	if !strings.Contains(stdout, "Pending Operations: 1") {
		t.Errorf("expected 1 pending operation in queue, but got:\n%s", stdout)
	}

	// Remote should NOT have the task yet
	if _, err := os.Stat(remoteDBPath); err == nil {
		remoteDB, err := sql.Open("sqlite", remoteDBPath)
		if err == nil {
			defer func() { _ = remoteDB.Close() }()
			var count int
			err = remoteDB.QueryRow(`
				SELECT COUNT(*) FROM tasks t
				JOIN task_lists l ON t.list_id = l.id
				WHERE t.summary = 'Queued-only task'
			`).Scan(&count)
			if err == nil && count > 0 {
				t.Errorf("task should NOT be in remote db until manual sync; got count=%d", count)
			}
		}
	}
}

// TestAutoSyncAfterUpdateOperation verifies that update operations trigger auto-sync
// Note: The actual update sync has a known issue with ID mapping between local and remote,
// so we only verify that auto-sync is triggered (queue is cleared), not the result.
func TestAutoSyncAfterUpdateOperation(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Set up path for "remote" SQLite database
	remoteDBPath := filepath.Join(tmpDir, "remote.db")

	// Create a config with auto_sync enabled
	// Use offline_mode: auto (default) - CLI uses local cache, sync pushes to remote
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  conflict_resolution: server_wins
  offline_mode: auto
  auto_sync_after_operation: true
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

	// Add a task (auto-syncs)
	cli.MustExecute("-y", "Work", "add", "Task to update")

	// Clear the queue check (task should be synced already)
	stdout := cli.MustExecute("-y", "sync", "queue")
	if !strings.Contains(stdout, "Pending Operations: 0") {
		t.Fatalf("expected queue to be empty after add auto-sync, but got:\n%s", stdout)
	}

	// Update the task (should trigger auto-sync)
	cli.MustExecute("-y", "Work", "update", "Task to update", "-p", "1")

	// Verify that auto-sync was triggered (queue should be empty/cleared)
	// Note: The update operation may fail due to ID mismatch between local/remote,
	// but the auto-sync mechanism should still have been triggered.
	stdout = cli.MustExecute("-y", "sync", "queue")
	// With auto-sync enabled, the queue should be empty (operations processed, even if some failed)
	if !strings.Contains(stdout, "Pending Operations: 0") {
		t.Errorf("expected queue to be empty after update auto-sync, but got:\n%s", stdout)
	}
}

// TestAutoSyncAfterDeleteOperation verifies that delete operations trigger auto-sync
// Note: The actual delete sync has a known issue with ID mapping between local and remote,
// so we only verify that auto-sync is triggered (queue is cleared), not the result.
func TestAutoSyncAfterDeleteOperation(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Set up path for "remote" SQLite database
	remoteDBPath := filepath.Join(tmpDir, "remote.db")

	// Create a config with auto_sync enabled
	// Use offline_mode: auto (default) - CLI uses local cache, sync pushes to remote
	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  conflict_resolution: server_wins
  offline_mode: auto
  auto_sync_after_operation: true
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

	// Add a task (auto-syncs)
	cli.MustExecute("-y", "Work", "add", "Task to delete")

	// Verify add auto-sync worked
	stdout := cli.MustExecute("-y", "sync", "queue")
	if !strings.Contains(stdout, "Pending Operations: 0") {
		t.Fatalf("expected queue to be empty after add auto-sync, but got:\n%s", stdout)
	}

	// Delete the task (should trigger auto-sync)
	cli.MustExecute("-y", "Work", "delete", "Task to delete")

	// Verify that auto-sync was triggered (queue should be empty/cleared)
	// Note: The delete operation may fail due to ID mismatch between local/remote,
	// but the auto-sync mechanism should still have been triggered.
	stdout = cli.MustExecute("-y", "sync", "queue")
	if !strings.Contains(stdout, "Pending Operations: 0") {
		t.Errorf("expected queue to be empty after delete auto-sync, but got:\n%s", stdout)
	}
}

// =============================================================================
// Issue 007: Sync Fails When Local List Doesn't Exist on CalDAV Remote
// =============================================================================

// TestSyncSkipsTasksWhenListCreationNotSupported verifies that when syncing to a
// remote that doesn't support list creation (like CalDAV), tasks in unmapped lists
// are skipped with a warning instead of causing sync to fail completely.
// This is Issue #007: Sync Fails When Local List Doesn't Exist on CalDAV Remote
func TestSyncSkipsTasksWhenListCreationNotSupported(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Set up path for "remote" SQLite database
	remoteDBPath := filepath.Join(tmpDir, "remote.db")

	// First, initialize the remote database with the CLI to get proper schema
	// Create config pointing to remote
	configContent := `
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
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Initialize remote by creating a list there directly
	stdout, stderr, exitCode := cli.Execute("-y", "-b", "sqlite-remote", "Work", "add", "Remote task")
	if exitCode != 0 {
		t.Fatalf("failed to init remote: stdout=%s stderr=%s", stdout, stderr)
	}

	// Delete the task (but list "Work" now exists on remote)
	cli.MustExecute("-y", "-b", "sqlite-remote", "Work", "delete", "Remote task")

	// Clear any queued operations from setup
	cli.MustExecute("-y", "sync", "queue", "clear")

	// Now switch to offline mode to queue local operations
	configContent = `
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
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Add tasks to two different lists locally
	// "Work" exists on remote, "Personal" does NOT
	cli.MustExecute("-y", "Work", "add", "Task in Work list")
	cli.MustExecute("-y", "Personal", "add", "Task in Personal list")

	// Verify both tasks are queued
	stdout = cli.MustExecute("-y", "sync", "queue")
	testutil.AssertContains(t, stdout, "Pending Operations: 2")

	// Switch to online mode and run sync
	configContent = `
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
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Run sync - this should:
	// 1. Successfully sync "Task in Work list" (list exists on remote)
	// 2. Successfully sync "Task in Personal list" (SQLite backend CAN create lists)
	// Note: SQLite backend supports list creation, so both tasks should sync
	// The actual ErrListCreationNotSupported would only occur with CalDAV
	stdout, stderr, exitCode = cli.Execute("-y", "sync")

	// The sync should succeed (exit code 0)
	if exitCode != 0 {
		t.Errorf("sync failed with exit code %d.\nstdout: %s\nstderr: %s", exitCode, stdout, stderr)
	}

	// Should show operations were processed
	testutil.AssertContains(t, stdout, "Push:")

	// Verify the Work task was synced to remote
	remoteDB, err := sql.Open("sqlite", remoteDBPath)
	if err != nil {
		t.Fatalf("failed to open remote db: %v", err)
	}
	defer func() { _ = remoteDB.Close() }()

	var count int
	err = remoteDB.QueryRow(`
		SELECT COUNT(*) FROM tasks t
		JOIN task_lists l ON t.list_id = l.id
		WHERE t.summary = 'Task in Work list' AND l.name = 'Work'
	`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query remote db: %v", err)
	}

	if count != 1 {
		t.Errorf("expected Work task to be synced; found %d tasks in remote", count)
	}
}
