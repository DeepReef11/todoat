package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// =============================================================================
// Test Helper Functions (006-cli-tests)
// =============================================================================

// assertContains is a helper that fails the test if output doesn't contain expected
func assertContains(t *testing.T, output, expected string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, output)
	}
}

// assertNotContains is a helper that fails the test if output contains unexpected
func assertNotContains(t *testing.T, output, unexpected string) {
	t.Helper()
	if strings.Contains(output, unexpected) {
		t.Errorf("expected output NOT to contain %q, got:\n%s", unexpected, output)
	}
}

// assertExitCode is a helper that fails the test if exit code doesn't match
func assertExitCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("expected exit code %d, got %d", want, got)
	}
}

// assertResultCode verifies that the output ends with the expected result code
func assertResultCode(t *testing.T, output, expectedCode string) {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Errorf("expected result code %q but output is empty", expectedCode)
		return
	}
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	if lastLine != expectedCode {
		t.Errorf("expected result code %q, got %q\nFull output:\n%s", expectedCode, lastLine, output)
	}
}

// TestHelpFlag verifies that --help displays usage information
func TestHelpFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"--help"}, &stdout, &stderr, nil)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "todoat") {
		t.Errorf("help output should contain 'todoat', got: %s", output)
	}
	if !strings.Contains(output, "Usage:") {
		t.Errorf("help output should contain 'Usage:', got: %s", output)
	}
}

// TestVersionFlag verifies that --version displays version string
func TestVersionFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"--version"}, &stdout, &stderr, nil)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "todoat") {
		t.Errorf("version output should contain 'todoat', got: %s", output)
	}
}

// TestNoPromptFlag verifies that -y / --no-prompt flag is recognized
func TestNoPromptFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"short flag", []string{"-y", "--help"}},
		{"long flag", []string{"--no-prompt", "--help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			exitCode := Execute(tt.args, &stdout, &stderr, nil)

			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
			}
		})
	}
}

// TestVerboseFlag verifies that -V / --verbose flag is recognized
func TestVerboseFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"short flag", []string{"-V", "--help"}},
		{"long flag", []string{"--verbose", "--help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			exitCode := Execute(tt.args, &stdout, &stderr, nil)

			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
			}
		})
	}
}

// TestExitCodeSuccess verifies exit code 0 for successful operations
func TestExitCodeSuccess(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"--help"}, &stdout, &stderr, nil)

	if exitCode != 0 {
		t.Errorf("expected exit code 0 for help, got %d", exitCode)
	}
}

// TestExitCodeError verifies exit code 1 for errors (unknown flag)
func TestExitCodeError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"--unknown-flag-xyz"}, &stdout, &stderr, nil)

	if exitCode != 1 {
		t.Errorf("expected exit code 1 for unknown flag, got %d", exitCode)
	}
}

// TestMaxThreePositionalArgs verifies that more than 3 positional args fails
func TestMaxThreePositionalArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// 4 positional arguments should fail
	exitCode := Execute([]string{"list", "action", "task", "extra"}, &stdout, &stderr, nil)

	if exitCode != 1 {
		t.Errorf("expected exit code 1 for 4 positional args, got %d", exitCode)
	}

	combinedOutput := stderr.String() + stdout.String()
	if !strings.Contains(combinedOutput, "at most 3") && !strings.Contains(combinedOutput, "accepts at most 3") {
		t.Errorf("error should mention 'at most 3', got: %s", combinedOutput)
	}
}

// TestThreePositionalArgsAllowed verifies that exactly 3 positional args is allowed
func TestThreePositionalArgsAllowed(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// 3 positional arguments should be accepted (even if the command doesn't do anything useful yet)
	exitCode := Execute([]string{"list", "get", "task"}, &stdout, &stderr, nil)

	// Should not fail due to arg count (might fail for other reasons, but not arg count)
	if exitCode == 1 {
		combinedOutput := stderr.String() + stdout.String()
		// Fail only if it's an argument count error
		if strings.Contains(combinedOutput, "accepts at most") {
			t.Errorf("3 positional args should be allowed, but got: %s", combinedOutput)
		}
	}
}

// TestInjectableIO verifies that stdout and stderr writers are used
func TestInjectableIO(t *testing.T) {
	var stdout, stderr bytes.Buffer

	Execute([]string{"--help"}, &stdout, &stderr, nil)

	// Help should be written to stdout
	if stdout.Len() == 0 {
		t.Error("expected help output to be written to stdout")
	}
}

// TestConfigPassthrough verifies that config is accessible
func TestConfigPassthrough(t *testing.T) {
	var stdout, stderr bytes.Buffer

	cfg := &Config{
		NoPrompt:     true,
		OutputFormat: "json",
	}

	// Should not panic with config passed
	exitCode := Execute([]string{"--help"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0 with config, got %d", exitCode)
	}
}

// TestRootCommandShowsHelp verifies that running without args shows help
func TestRootCommandShowsHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{}, &stdout, &stderr, nil)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0 for no args, got %d: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Usage:") {
		t.Errorf("no-args should show help with Usage:, got: %s", output)
	}
}

// TestGlobalFlagsArePersistent verifies global flags work with subcommands
func TestGlobalFlagsArePersistent(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Global flags should be recognized even without subcommands
	exitCode := Execute([]string{"-y", "-V", "--help"}, &stdout, &stderr, nil)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}
}

// =============================================================================
// Task Command Tests (004-task-commands)
// =============================================================================

// testWithDB creates an in-memory database and config for testing
// Note: Each call returns a NEW in-memory database (isolated per test)
func testWithDB(t *testing.T) (*Config, func()) {
	t.Helper()

	// Use in-memory SQLite for fast, isolated tests
	// Note: We use a temp file path since each in-memory DB is separate per connection
	// For true isolation, use t.TempDir() or ":memory:" but track connections
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	cfg := &Config{
		NoPrompt: true,
		DBPath:   dbPath,
	}

	return cfg, func() {
		// Cleanup is automatic with t.TempDir()
	}
}

// --- Add Command Tests ---

func TestAddCommand(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"-y", "Work", "add", "Review PR"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Review PR") {
		t.Errorf("output should confirm task creation, got: %s", output)
	}
}

func TestAddCommandAbbreviation(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// 'a' is abbreviation for 'add'
	exitCode := Execute([]string{"-y", "Work", "a", "New task"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}
}

func TestAddCommandWithPriority(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"-y", "Work", "add", "Urgent task", "-p", "1"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	// List tasks to verify priority
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "Work", "get"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0 for get, got %d: stderr=%s", exitCode, stderr.String())
	}

	// Should show priority indicator
	output := stdout.String()
	if !strings.Contains(output, "Urgent task") {
		t.Errorf("task list should contain 'Urgent task', got: %s", output)
	}
}

// --- Get Command Tests ---

func TestGetCommandExplicit(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// First add a task
	Execute([]string{"-y", "Work", "add", "Task 1"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Explicit get command
	exitCode := Execute([]string{"-y", "Work", "get"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Task 1") {
		t.Errorf("task list should contain 'Task 1', got: %s", output)
	}
}

func TestGetCommandDefault(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// First add a task
	Execute([]string{"-y", "Work", "add", "Task for default"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Default action (just list name, no action) should show tasks
	exitCode := Execute([]string{"-y", "Work"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Task for default") {
		t.Errorf("default action should show tasks, got: %s", output)
	}
}

func TestGetCommandAbbreviation(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task first
	Execute([]string{"-y", "Work", "add", "Task G"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// 'g' is abbreviation for 'get'
	exitCode := Execute([]string{"-y", "Work", "g"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Task G") {
		t.Errorf("task list should contain 'Task G', got: %s", output)
	}
}

// --- Update Command Tests ---

func TestUpdateCommandSummary(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Old name"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Update the summary
	exitCode := Execute([]string{"-y", "Work", "update", "Old name", "--summary", "New name"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	// Verify the update
	stdout.Reset()
	stderr.Reset()
	Execute([]string{"-y", "Work", "get"}, &stdout, &stderr, cfg)

	output := stdout.String()
	if strings.Contains(output, "Old name") {
		t.Errorf("old name should not appear, got: %s", output)
	}
	if !strings.Contains(output, "New name") {
		t.Errorf("new name should appear, got: %s", output)
	}
}

func TestUpdateCommandPriority(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Task to update"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Update priority using abbreviation
	exitCode := Execute([]string{"-y", "Work", "u", "Task to update", "-p", "5"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}
}

func TestUpdateCommandStatus(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Task status"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Update status
	exitCode := Execute([]string{"-y", "Work", "update", "Task status", "-s", "IN-PROGRESS"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}
}

// --- Complete Command Tests ---

func TestCompleteCommand(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Task to complete"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Complete the task
	exitCode := Execute([]string{"-y", "Work", "complete", "Task to complete"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}
}

func TestCompleteCommandAbbreviation(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Another task"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// 'c' is abbreviation for 'complete'
	exitCode := Execute([]string{"-y", "Work", "c", "Another task"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}
}

// --- Delete Command Tests ---

func TestDeleteCommand(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Task to delete"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Delete the task (with -y for no prompt)
	exitCode := Execute([]string{"-y", "Work", "delete", "Task to delete"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	// Verify deletion
	stdout.Reset()
	stderr.Reset()
	Execute([]string{"-y", "Work", "get"}, &stdout, &stderr, cfg)

	output := stdout.String()
	if strings.Contains(output, "Task to delete") {
		t.Errorf("deleted task should not appear, got: %s", output)
	}
}

func TestDeleteCommandAbbreviation(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Delete me"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// 'd' is abbreviation for 'delete'
	exitCode := Execute([]string{"-y", "Work", "d", "Delete me"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}
}

// --- Task Matching Tests ---

func TestTaskMatchingExact(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add tasks with similar names
	Execute([]string{"-y", "Work", "add", "Review PR"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Review PR #123"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Exact match should find "Review PR"
	exitCode := Execute([]string{"-y", "Work", "complete", "Review PR"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}
}

func TestTaskMatchingPartial(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Review PR #456"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Partial match should work when only one task matches
	exitCode := Execute([]string{"-y", "Work", "complete", "#456"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}
}

func TestTaskMatchingNoMatch(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Some task"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// No match should error
	exitCode := Execute([]string{"-y", "Work", "complete", "Nonexistent"}, &stdout, &stderr, cfg)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1 for no match, got %d", exitCode)
	}

	errOutput := stderr.String()
	if !strings.Contains(strings.ToLower(errOutput), "no") && !strings.Contains(strings.ToLower(errOutput), "match") && !strings.Contains(strings.ToLower(errOutput), "found") {
		t.Errorf("error should mention no match found, got: %s", errOutput)
	}
}

func TestTaskMatchingMultipleMatches(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add tasks with similar names
	Execute([]string{"-y", "Work", "add", "Review code"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Review docs"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Multiple matches in no-prompt mode should error
	exitCode := Execute([]string{"-y", "Work", "complete", "Review"}, &stdout, &stderr, cfg)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1 for multiple matches, got %d", exitCode)
	}

	errOutput := stderr.String()
	if !strings.Contains(strings.ToLower(errOutput), "multiple") && !strings.Contains(strings.ToLower(errOutput), "matches") && !strings.Contains(strings.ToLower(errOutput), "ambiguous") {
		t.Errorf("error should mention multiple matches, got: %s", errOutput)
	}
}

func TestTaskMatchingCaseInsensitive(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task with mixed case
	Execute([]string{"-y", "Work", "add", "Review PR"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Lowercase search should match
	exitCode := Execute([]string{"-y", "Work", "complete", "review pr"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0 for case-insensitive match, got %d: stderr=%s", exitCode, stderr.String())
	}
}

// --- Status System Tests ---

func TestStatusDisplayFormat(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task (default status is TODO)
	Execute([]string{"-y", "Work", "add", "Task one"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Get tasks and check status format
	exitCode := Execute([]string{"-y", "Work", "get"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "[TODO]") {
		t.Errorf("expected [TODO] status indicator, got: %s", output)
	}
}

func TestStatusDisplayFormatDone(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add and complete a task
	Execute([]string{"-y", "Work", "add", "Task done"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "complete", "Task done"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Get tasks and check status format
	exitCode := Execute([]string{"-y", "Work", "get"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "[DONE]") {
		t.Errorf("expected [DONE] status indicator, got: %s", output)
	}
}

func TestStatusAbbreviationT(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task and set it to DONE first
	Execute([]string{"-y", "Work", "add", "Task abbrev"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "complete", "Task abbrev"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Update status using abbreviation T (should set to TODO)
	exitCode := Execute([]string{"-y", "Work", "update", "Task abbrev", "-s", "T"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// Verify status is TODO
	Execute([]string{"-y", "Work", "get"}, &stdout, &stderr, cfg)
	output := stdout.String()
	if !strings.Contains(output, "[TODO]") {
		t.Errorf("expected [TODO] after using -s T abbreviation, got: %s", output)
	}
}

func TestStatusAbbreviationD(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Task abbrev D"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Update status using abbreviation D (should set to DONE)
	exitCode := Execute([]string{"-y", "Work", "update", "Task abbrev D", "-s", "D"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// Verify status is DONE
	Execute([]string{"-y", "Work", "get"}, &stdout, &stderr, cfg)
	output := stdout.String()
	if !strings.Contains(output, "[DONE]") {
		t.Errorf("expected [DONE] after using -s D abbreviation, got: %s", output)
	}
}

func TestStatusCaseInsensitive(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Task case"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Update status using lowercase
	exitCode := Execute([]string{"-y", "Work", "update", "Task case", "-s", "done"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// Verify status is DONE
	Execute([]string{"-y", "Work", "get"}, &stdout, &stderr, cfg)
	output := stdout.String()
	if !strings.Contains(output, "[DONE]") {
		t.Errorf("expected [DONE] after using lowercase status, got: %s", output)
	}
}

func TestFilterByStatusTodo(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add tasks with different statuses
	Execute([]string{"-y", "Work", "add", "Task todo one"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Task done one"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "complete", "Task done one"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter to show only TODO tasks
	exitCode := Execute([]string{"-y", "Work", "-s", "TODO"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Task todo one") {
		t.Errorf("expected to see TODO task, got: %s", output)
	}
	if strings.Contains(output, "Task done one") {
		t.Errorf("should NOT see DONE task when filtering for TODO, got: %s", output)
	}
}

func TestFilterByStatusDone(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add tasks with different statuses
	Execute([]string{"-y", "Work", "add", "Task todo two"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Task done two"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "complete", "Task done two"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter to show only DONE tasks
	exitCode := Execute([]string{"-y", "Work", "-s", "DONE"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Task done two") {
		t.Errorf("expected to see DONE task, got: %s", output)
	}
	if strings.Contains(output, "Task todo two") {
		t.Errorf("should NOT see TODO task when filtering for DONE, got: %s", output)
	}
}

func TestFilterByStatusAbbreviation(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add tasks with different statuses
	Execute([]string{"-y", "Work", "add", "Task todo three"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Task done three"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "complete", "Task done three"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter using abbreviation T
	exitCode := Execute([]string{"-y", "Work", "-s", "T"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Task todo three") {
		t.Errorf("expected to see TODO task with -s T, got: %s", output)
	}
	if strings.Contains(output, "Task done three") {
		t.Errorf("should NOT see DONE task with -s T, got: %s", output)
	}
}

func TestFilterByStatusLongFlag(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add tasks with different statuses
	Execute([]string{"-y", "Work", "add", "Task todo four"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Task done four"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "complete", "Task done four"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter using --status long flag
	exitCode := Execute([]string{"-y", "Work", "--status", "D"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Task done four") {
		t.Errorf("expected to see DONE task with --status D, got: %s", output)
	}
	if strings.Contains(output, "Task todo four") {
		t.Errorf("should NOT see TODO task with --status D, got: %s", output)
	}
}

func TestNoFilterShowsAllTasks(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add tasks with different statuses
	Execute([]string{"-y", "Work", "add", "Task todo five"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Task done five"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "complete", "Task done five"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Get without filter should show all
	exitCode := Execute([]string{"-y", "Work"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Task todo five") {
		t.Errorf("expected to see TODO task without filter, got: %s", output)
	}
	if !strings.Contains(output, "Task done five") {
		t.Errorf("expected to see DONE task without filter, got: %s", output)
	}
}

// =============================================================================
// Result Code Tests (006-cli-tests)
// =============================================================================

func TestResultCodeAddTask(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"-y", "Work", "add", "Test task"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)
}

func TestResultCodeGetTasks(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task first
	Execute([]string{"-y", "Work", "add", "Task to list"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Get tasks should return INFO_ONLY
	exitCode := Execute([]string{"-y", "Work", "get"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultInfoOnly)
}

func TestResultCodeListEmpty(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Get empty list should return INFO_ONLY
	exitCode := Execute([]string{"-y", "EmptyList"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultInfoOnly)
	assertContains(t, stdout.String(), "No tasks")
}

func TestResultCodeUpdateTask(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Task to update"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Update task should return ACTION_COMPLETED
	exitCode := Execute([]string{"-y", "Work", "update", "Task to update", "--summary", "Updated task"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)
}

func TestResultCodeCompleteTask(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Task to complete"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Complete task should return ACTION_COMPLETED
	exitCode := Execute([]string{"-y", "Work", "complete", "Task to complete"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)
}

func TestResultCodeCompleteTaskChangesDoneStatus(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add and complete a task
	Execute([]string{"-y", "Work", "add", "Task for status"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "complete", "Task for status"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Verify status changed to DONE
	Execute([]string{"-y", "Work", "get"}, &stdout, &stderr, cfg)

	assertContains(t, stdout.String(), "[DONE]")
	assertResultCode(t, stdout.String(), ResultInfoOnly)
}

func TestResultCodeDeleteTask(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Task to delete"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Delete task should return ACTION_COMPLETED
	exitCode := Execute([]string{"-y", "Work", "delete", "Task to delete"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)
}

func TestResultCodeDeleteConfirmationSkippedInNoPrompt(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Task to confirm delete"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Delete with -y should not require confirmation
	exitCode := Execute([]string{"-y", "Work", "delete", "Task to confirm delete"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)

	// Verify task is gone
	stdout.Reset()
	stderr.Reset()
	Execute([]string{"-y", "Work", "get"}, &stdout, &stderr, cfg)
	assertNotContains(t, stdout.String(), "Task to confirm delete")
}

func TestResultCodeErrorNoMatch(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Existing task"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Try to complete non-existent task
	exitCode := Execute([]string{"-y", "Work", "complete", "Nonexistent task"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 1)
	assertResultCode(t, stdout.String(), ResultError)
}

func TestResultCodeErrorAmbiguousMatch(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add tasks with similar names
	Execute([]string{"-y", "Work", "add", "Similar task one"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Similar task two"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Try to complete with ambiguous match
	exitCode := Execute([]string{"-y", "Work", "complete", "Similar"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 1)
	assertResultCode(t, stdout.String(), ResultError)
}

func TestExitCodesVerified(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Test exit code 0 for ACTION_COMPLETED
	exitCode := Execute([]string{"-y", "Work", "add", "Test exit code"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for add, got %d", exitCode)
	}

	stdout.Reset()
	stderr.Reset()

	// Test exit code 0 for INFO_ONLY
	exitCode = Execute([]string{"-y", "Work", "get"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for get, got %d", exitCode)
	}

	stdout.Reset()
	stderr.Reset()

	// Test exit code 1 for ERROR
	exitCode = Execute([]string{"-y", "Work", "complete", "Nonexistent"}, &stdout, &stderr, cfg)
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for error, got %d", exitCode)
	}
}
