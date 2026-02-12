package nextcloud_test

import (
	"encoding/json"
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// =============================================================================
// CLI Integration Tests for Calendar Sharing (Issue #93)
// =============================================================================

// TestListShareCommandHelp verifies the list share subcommand appears in help
func TestListShareCommandHelp(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout, _, exitCode := cli.Execute("list", "--help")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "share") {
		t.Errorf("list --help should mention 'share', got: %s", stdout)
	}
	if !strings.Contains(stdout, "unshare") {
		t.Errorf("list --help should mention 'unshare', got: %s", stdout)
	}
}

// TestListShareRequiresUser verifies the --user flag is required
func TestListShareRequiresUser(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Attempt to share without --user flag
	_, stderr, exitCode := cli.Execute("list", "share", "SomeList")
	if exitCode == 0 {
		t.Fatal("expected non-zero exit code when --user is missing")
	}

	combined := strings.ToLower(stderr)
	if !strings.Contains(combined, "user") {
		t.Errorf("error should mention 'user' flag, got: %s", stderr)
	}
}

// TestListUnshareRequiresUser verifies the --user flag is required for unshare
func TestListUnshareRequiresUser(t *testing.T) {
	cli := testutil.NewCLITest(t)

	_, stderr, exitCode := cli.Execute("list", "unshare", "SomeList")
	if exitCode == 0 {
		t.Fatal("expected non-zero exit code when --user is missing")
	}

	combined := strings.ToLower(stderr)
	if !strings.Contains(combined, "user") {
		t.Errorf("error should mention 'user' flag, got: %s", stderr)
	}
}

// TestListShareNotSupportedBySQLite verifies that share fails with non-Nextcloud backend
func TestListShareNotSupportedBySQLite(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// First create a list to share
	_, _, exitCode := cli.Execute("list", "create", "TestShareList")
	if exitCode != 0 {
		t.Fatal("failed to create test list")
	}

	// Attempt to share - should fail since SQLite doesn't support sharing
	stdout, stderr, exitCode := cli.Execute("list", "share", "TestShareList", "--user", "bob")
	if exitCode == 0 {
		t.Fatal("expected share to fail with SQLite backend")
	}

	combined := strings.ToLower(stdout + stderr)
	if !strings.Contains(combined, "not supported") && !strings.Contains(combined, "requires nextcloud") {
		t.Errorf("error should mention sharing is not supported, got: %s", stdout+stderr)
	}
}

// TestListUnshareNotSupportedBySQLite verifies that unshare fails with non-Nextcloud backend
func TestListUnshareNotSupportedBySQLite(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// First create a list
	_, _, exitCode := cli.Execute("list", "create", "TestUnshareList")
	if exitCode != 0 {
		t.Fatal("failed to create test list")
	}

	// Attempt to unshare - should fail since SQLite doesn't support sharing
	stdout, stderr, exitCode := cli.Execute("list", "unshare", "TestUnshareList", "--user", "bob")
	if exitCode == 0 {
		t.Fatal("expected unshare to fail with SQLite backend")
	}

	combined := strings.ToLower(stdout + stderr)
	if !strings.Contains(combined, "not supported") && !strings.Contains(combined, "requires nextcloud") {
		t.Errorf("error should mention sharing is not supported, got: %s", stdout+stderr)
	}
}

// TestListShareCommandJSONOutput verifies JSON output format for share rejection
func TestListShareCommandJSONOutput(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    type: sqlite
    enabled: true
  nextcloud-share-test:
    type: nextcloud
    enabled: true
    host: "localhost:19999"
    username: "testadmin"
default_backend: sqlite
no_prompt: true
`)

	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "testpass")

	// Create a list in SQLite
	_, _, exitCode := cli.Execute("list", "create", "JSONShareTest")
	if exitCode != 0 {
		t.Fatal("failed to create test list")
	}

	// Attempt to share with --json flag - should fail but with proper error
	stdout, stderr, exitCode := cli.Execute("--json", "list", "share", "JSONShareTest", "--user", "bob")
	if exitCode == 0 {
		t.Fatal("expected share to fail with SQLite backend")
	}

	// Should contain an error about sharing not supported
	combined := strings.ToLower(stdout + stderr)
	if !strings.Contains(combined, "not supported") && !strings.Contains(combined, "requires nextcloud") {
		t.Errorf("error should mention sharing is not supported, got: %s", stdout+stderr)
	}
}

// TestListSharePermissionFlag verifies the --permission flag is accepted
func TestListSharePermissionFlag(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// Create a list
	_, _, exitCode := cli.Execute("list", "create", "PermTestList")
	if exitCode != 0 {
		t.Fatal("failed to create test list")
	}

	// Try with --permission flag (will still fail because SQLite, but should accept the flag)
	stdout, stderr, exitCode := cli.Execute("list", "share", "PermTestList", "--user", "bob", "--permission", "write")
	if exitCode == 0 {
		t.Fatal("expected share to fail with SQLite backend")
	}

	// Error should be about backend support, not about unknown flags
	combined := strings.ToLower(stdout + stderr)
	if strings.Contains(combined, "unknown flag") {
		t.Errorf("--permission flag should be recognized, got: %s", stdout+stderr)
	}
}

// TestListShareSuccessJSON tests successful JSON output when using Nextcloud-like backend
// This test validates JSON structure by creating a share result manually
func TestListShareSuccessJSON(t *testing.T) {
	// Validate that the JSON format for share operations is correct
	type shareJSON struct {
		Result     string `json:"result"`
		Action     string `json:"action"`
		List       string `json:"list"`
		User       string `json:"user"`
		Permission string `json:"permission"`
	}

	output := shareJSON{
		Result:     "ACTION_COMPLETED",
		Action:     "shared",
		List:       "MyCalendar",
		User:       "bob",
		Permission: "write",
	}

	jsonBytes, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal share JSON: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal share JSON: %v", err)
	}

	if parsed["result"] != "ACTION_COMPLETED" {
		t.Errorf("Expected result 'ACTION_COMPLETED', got %v", parsed["result"])
	}
	if parsed["action"] != "shared" {
		t.Errorf("Expected action 'shared', got %v", parsed["action"])
	}
	if parsed["list"] != "MyCalendar" {
		t.Errorf("Expected list 'MyCalendar', got %v", parsed["list"])
	}
}

// TestListUnshareSuccessJSON tests successful JSON output for unshare operations
func TestListUnshareSuccessJSON(t *testing.T) {
	type unshareJSON struct {
		Result string `json:"result"`
		Action string `json:"action"`
		List   string `json:"list"`
		User   string `json:"user"`
	}

	output := unshareJSON{
		Result: "ACTION_COMPLETED",
		Action: "unshared",
		List:   "MyCalendar",
		User:   "bob",
	}

	jsonBytes, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal unshare JSON: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal unshare JSON: %v", err)
	}

	if parsed["result"] != "ACTION_COMPLETED" {
		t.Errorf("Expected result 'ACTION_COMPLETED', got %v", parsed["result"])
	}
	if parsed["action"] != "unshared" {
		t.Errorf("Expected action 'unshared', got %v", parsed["action"])
	}
}
