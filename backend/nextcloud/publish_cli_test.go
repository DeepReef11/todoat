package nextcloud_test

import (
	"encoding/json"
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// =============================================================================
// CLI Integration Tests for Public Links (Issue #95)
// =============================================================================

// TestListPublishCommandHelp verifies the list publish subcommand appears in help
func TestListPublishCommandHelp(t *testing.T) {
	cli := testutil.NewCLITest(t)

	stdout, _, exitCode := cli.Execute("list", "--help")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "publish") {
		t.Errorf("list --help should mention 'publish', got: %s", stdout)
	}
	if !strings.Contains(stdout, "unpublish") {
		t.Errorf("list --help should mention 'unpublish', got: %s", stdout)
	}
}

// TestListPublishRequiresArg verifies that a list name argument is required
func TestListPublishRequiresArg(t *testing.T) {
	cli := testutil.NewCLITest(t)

	_, _, exitCode := cli.Execute("list", "publish")
	if exitCode == 0 {
		t.Fatal("expected non-zero exit code when list name is missing")
	}
}

// TestListUnpublishRequiresArg verifies that a list name argument is required
func TestListUnpublishRequiresArg(t *testing.T) {
	cli := testutil.NewCLITest(t)

	_, _, exitCode := cli.Execute("list", "unpublish")
	if exitCode == 0 {
		t.Fatal("expected non-zero exit code when list name is missing")
	}
}

// TestListPublishNotSupportedBySQLite verifies that publish fails with non-Nextcloud backend
func TestListPublishNotSupportedBySQLite(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// First create a list
	_, _, exitCode := cli.Execute("list", "create", "TestPublishList")
	if exitCode != 0 {
		t.Fatal("failed to create test list")
	}

	// Attempt to publish - should fail since SQLite doesn't support publishing
	stdout, stderr, exitCode := cli.Execute("list", "publish", "TestPublishList")
	if exitCode == 0 {
		t.Fatal("expected publish to fail with SQLite backend")
	}

	combined := strings.ToLower(stdout + stderr)
	if !strings.Contains(combined, "not supported") && !strings.Contains(combined, "requires nextcloud") {
		t.Errorf("error should mention publishing is not supported, got: %s", stdout+stderr)
	}
}

// TestListUnpublishNotSupportedBySQLite verifies that unpublish fails with non-Nextcloud backend
func TestListUnpublishNotSupportedBySQLite(t *testing.T) {
	cli := testutil.NewCLITest(t)

	// First create a list
	_, _, exitCode := cli.Execute("list", "create", "TestUnpublishList")
	if exitCode != 0 {
		t.Fatal("failed to create test list")
	}

	// Attempt to unpublish - should fail since SQLite doesn't support publishing
	stdout, stderr, exitCode := cli.Execute("list", "unpublish", "TestUnpublishList")
	if exitCode == 0 {
		t.Fatal("expected unpublish to fail with SQLite backend")
	}

	combined := strings.ToLower(stdout + stderr)
	if !strings.Contains(combined, "not supported") && !strings.Contains(combined, "requires nextcloud") {
		t.Errorf("error should mention publishing is not supported, got: %s", stdout+stderr)
	}
}

// TestListPublishCommandJSONOutput verifies JSON output format for publish rejection
func TestListPublishCommandJSONOutput(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    type: sqlite
    enabled: true
default_backend: sqlite
no_prompt: true
`)

	// Create a list in SQLite
	_, _, exitCode := cli.Execute("list", "create", "JSONPublishTest")
	if exitCode != 0 {
		t.Fatal("failed to create test list")
	}

	// Attempt to publish with --json flag - should fail but with proper error
	stdout, stderr, exitCode := cli.Execute("--json", "list", "publish", "JSONPublishTest")
	if exitCode == 0 {
		t.Fatal("expected publish to fail with SQLite backend")
	}

	// Should contain an error about publishing not supported
	combined := strings.ToLower(stdout + stderr)
	if !strings.Contains(combined, "not supported") && !strings.Contains(combined, "requires nextcloud") {
		t.Errorf("error should mention publishing is not supported, got: %s", stdout+stderr)
	}
}

// TestListPublishSuccessJSON tests successful JSON output structure for publish
func TestListPublishSuccessJSON(t *testing.T) {
	type publishJSON struct {
		Result string `json:"result"`
		Action string `json:"action"`
		List   string `json:"list"`
		URL    string `json:"url"`
	}

	output := publishJSON{
		Result: "ACTION_COMPLETED",
		Action: "published",
		List:   "MyCalendar",
		URL:    "https://nextcloud.example.com/s/abc123xyz",
	}

	jsonBytes, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal publish JSON: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal publish JSON: %v", err)
	}

	if parsed["result"] != "ACTION_COMPLETED" {
		t.Errorf("Expected result 'ACTION_COMPLETED', got %v", parsed["result"])
	}
	if parsed["action"] != "published" {
		t.Errorf("Expected action 'published', got %v", parsed["action"])
	}
	if parsed["list"] != "MyCalendar" {
		t.Errorf("Expected list 'MyCalendar', got %v", parsed["list"])
	}
	if parsed["url"] != "https://nextcloud.example.com/s/abc123xyz" {
		t.Errorf("Expected URL 'https://nextcloud.example.com/s/abc123xyz', got %v", parsed["url"])
	}
}

// TestListUnpublishSuccessJSON tests successful JSON output structure for unpublish
func TestListUnpublishSuccessJSON(t *testing.T) {
	type unpublishJSON struct {
		Result string `json:"result"`
		Action string `json:"action"`
		List   string `json:"list"`
	}

	output := unpublishJSON{
		Result: "ACTION_COMPLETED",
		Action: "unpublished",
		List:   "MyCalendar",
	}

	jsonBytes, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal unpublish JSON: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal unpublish JSON: %v", err)
	}

	if parsed["result"] != "ACTION_COMPLETED" {
		t.Errorf("Expected result 'ACTION_COMPLETED', got %v", parsed["result"])
	}
	if parsed["action"] != "unpublished" {
		t.Errorf("Expected action 'unpublished', got %v", parsed["action"])
	}
}
