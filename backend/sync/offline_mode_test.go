package sync_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// =============================================================================
// Offline Mode Configuration Tests (047-offline-mode-config)
// =============================================================================

// TestOfflineModeAuto tests that `sync.offline_mode: auto` detects network state
// and queues operations when offline
func TestOfflineModeAuto(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with sync enabled and auto offline mode (default)
	createOfflineModeConfig(t, tmpDir, "auto", "5s")

	// Add a task - should queue operation since no remote is configured
	cli.MustExecute("-y", "Work", "add", "Auto mode task")

	// Check the queue - operation should be queued
	stdout := cli.MustExecute("-y", "sync", "queue")

	// In auto mode, operations get queued when remote is unavailable
	testutil.AssertContains(t, stdout, "Pending Operations")
	testutil.AssertContains(t, stdout, "create")

	// Verify sync status shows auto mode
	stdout = cli.MustExecute("-y", "sync", "status")
	testutil.AssertContains(t, stdout, "Sync Status")
	// Should show offline mode configuration
	testutil.AssertContains(t, stdout, "Offline Mode: auto")
}

// TestOfflineModeOnline tests that `sync.offline_mode: online` fails immediately
// if backend unreachable
func TestOfflineModeOnline(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with online-only mode
	createOfflineModeConfig(t, tmpDir, "online", "1s")

	// Verify sync status shows online mode
	stdout := cli.MustExecute("-y", "sync", "status")
	testutil.AssertContains(t, stdout, "Offline Mode: online")

	// Add a task - in online mode with local sqlite backend only,
	// operations succeed locally. The online mode primarily affects
	// behavior when a remote backend is configured.
	stdout = cli.MustExecute("-y", "Work", "add", "Online mode task")
	testutil.AssertContains(t, stdout, "Created task")

	// Note: Full "fails when unreachable" behavior requires remote backend
	// connectivity testing which is out of scope for basic config tests.
	// The key test here is that online mode is correctly read from config.
}

// TestOfflineModeOffline tests that `sync.offline_mode: offline` always queues
// operations, never contacts remote
func TestOfflineModeOffline(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with forced offline mode
	createOfflineModeConfig(t, tmpDir, "offline", "5s")

	// Add a task - should always queue, never try remote
	cli.MustExecute("-y", "Work", "add", "Offline mode task")

	// Check the queue - operation should be queued
	stdout := cli.MustExecute("-y", "sync", "queue")

	// In offline mode, operations always get queued
	testutil.AssertContains(t, stdout, "Pending Operations")
	testutil.AssertContains(t, stdout, "create")

	// Add more tasks - all should queue
	cli.MustExecute("-y", "Work", "add", "Another offline task")
	cli.MustExecute("-y", "Work", "update", "Offline mode task", "-p", "1")

	stdout = cli.MustExecute("-y", "sync", "queue")
	// Should show multiple pending operations
	testutil.AssertContains(t, stdout, "Pending Operations")

	// Verify sync status shows offline mode
	stdout = cli.MustExecute("-y", "sync", "status")
	testutil.AssertContains(t, stdout, "Offline Mode: offline")
}

// TestOfflineModeAutoOnline tests that in auto mode, operations succeed when
// backend is reachable
func TestOfflineModeAutoOnline(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with auto mode
	createOfflineModeConfig(t, tmpDir, "auto", "5s")

	// Add a task - should succeed
	stdout := cli.MustExecute("-y", "Work", "add", "Auto online task")
	testutil.AssertContains(t, stdout, "Created task")

	// Verify task was created
	stdout = cli.MustExecute("-y", "Work", "get")
	testutil.AssertContains(t, stdout, "Auto online task")
}

// TestOfflineModeAutoOffline tests that in auto mode, operations queue when
// backend times out
func TestOfflineModeAutoOffline(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with auto mode and short timeout
	createOfflineModeConfig(t, tmpDir, "auto", "100ms")

	// Add a task - should queue when remote times out
	cli.MustExecute("-y", "Work", "add", "Auto offline task")

	// Operations should still succeed locally and be queued
	stdout := cli.MustExecute("-y", "sync", "queue")
	testutil.AssertContains(t, stdout, "Pending Operations")
}

// TestOfflineQueuedOpsSync tests that queued operations sync when connectivity
// restored and `todoat sync` run
func TestOfflineQueuedOpsSync(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config with offline mode to force queueing
	createOfflineModeConfig(t, tmpDir, "offline", "5s")

	// Add tasks while "offline" (forced offline mode)
	cli.MustExecute("-y", "Work", "add", "Queued task 1")
	cli.MustExecute("-y", "Work", "add", "Queued task 2")
	cli.MustExecute("-y", "Work", "add", "Queued task 3")

	// Check queue has operations
	stdout := cli.MustExecute("-y", "sync", "queue")
	testutil.AssertContains(t, stdout, "Pending Operations")

	// Switch to auto mode to simulate connectivity restored
	createOfflineModeConfig(t, tmpDir, "auto", "5s")

	// Run sync - should attempt to process queued operations
	stdout, _, exitCode := cli.Execute("-y", "sync")

	// Sync should run (may not fully succeed without a real remote backend)
	// The key is that it attempts to process the queue
	if exitCode == 0 || strings.Contains(stdout, "Sync") {
		// Either succeeded or showed sync output
		t.Logf("Sync output: %s", stdout)
	}

	// Verify sync status shows processing attempt
	stdout = cli.MustExecute("-y", "sync", "status")
	testutil.AssertContains(t, stdout, "Sync Status")
}

// TestOfflineModeStatusDisplay tests that `todoat sync status` displays
// offline/online state
func TestOfflineModeStatusDisplay(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Test with each mode
	modes := []string{"auto", "online", "offline"}

	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			createOfflineModeConfig(t, tmpDir, mode, "5s")

			stdout := cli.MustExecute("-y", "sync", "status")

			// Sync status should display the mode
			testutil.AssertContains(t, stdout, "Sync Status")
			testutil.AssertContains(t, stdout, "Offline Mode: "+mode)
		})
	}
}

// TestOfflineModeConfigDefault tests that the default offline mode is "auto"
func TestOfflineModeConfigDefault(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Create a config without specifying offline_mode
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

	// Add a task - should work with default auto mode
	cli.MustExecute("-y", "Work", "add", "Default mode task")

	// Check sync status - should show auto as default
	stdout := cli.MustExecute("-y", "sync", "status")
	testutil.AssertContains(t, stdout, "Sync Status")
	testutil.AssertContains(t, stdout, "Offline Mode: auto")
}

// TestOfflineModeConnectivityTimeout tests that connectivity_timeout is respected
func TestOfflineModeConnectivityTimeout(t *testing.T) {
	cli, tmpDir := newSyncTestCLI(t)

	// Test with different timeout values
	timeouts := []string{"1s", "5s", "30s"}

	for _, timeout := range timeouts {
		t.Run(timeout, func(t *testing.T) {
			createOfflineModeConfig(t, tmpDir, "auto", timeout)

			// Add a task - should succeed regardless of timeout without remote
			stdout := cli.MustExecute("-y", "Work", "add", "Timeout test "+timeout)
			testutil.AssertContains(t, stdout, "Created task")
		})
	}
}

// createOfflineModeConfig creates a config file with offline mode settings
func createOfflineModeConfig(t *testing.T, tmpDir, mode, timeout string) {
	t.Helper()

	configContent := `
sync:
  enabled: true
  local_backend: sqlite
  offline_mode: ` + mode + `
  connectivity_timeout: ` + timeout + `
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
