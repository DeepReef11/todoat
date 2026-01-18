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

	// 4 positional arguments should fail (use "mylist" instead of "list" which is now a subcommand)
	exitCode := Execute([]string{"mylist", "action", "task", "extra"}, &stdout, &stderr, nil)

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

	// 3 positional arguments should be accepted (use "mylist" instead of "list" which is now a subcommand)
	exitCode := Execute([]string{"mylist", "get", "task"}, &stdout, &stderr, nil)

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

// =============================================================================
// List Management Tests (007-list-commands)
// =============================================================================

// TestListCreate verifies that `todoat -y list create "MyList"` creates a new list
func TestListCreate(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"-y", "list", "create", "MyList"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout.String(), "MyList")
	assertResultCode(t, stdout.String(), ResultActionCompleted)
}

// TestListCreateDuplicate verifies that creating a duplicate list returns ERROR
func TestListCreateDuplicate(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create first list
	Execute([]string{"-y", "list", "create", "ExistingList"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Try to create duplicate
	exitCode := Execute([]string{"-y", "list", "create", "ExistingList"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 1)
	assertResultCode(t, stdout.String(), ResultError)
}

// TestListView verifies that `todoat -y list` displays all lists with task counts
func TestListView(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create lists and add tasks
	Execute([]string{"-y", "list", "create", "Work"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "list", "create", "Personal"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Task 1"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Task 2"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Personal", "add", "Task 3"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// View lists
	exitCode := Execute([]string{"-y", "list"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Work")
	assertContains(t, output, "Personal")
	assertResultCode(t, output, ResultInfoOnly)
}

// TestListViewEmpty verifies that viewing lists with no lists shows INFO_ONLY message
func TestListViewEmpty(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// View lists when none exist
	exitCode := Execute([]string{"-y", "list"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	// Should contain a helpful message about no lists
	if !strings.Contains(strings.ToLower(output), "no") || !strings.Contains(strings.ToLower(output), "list") {
		t.Errorf("expected message about no lists, got: %s", output)
	}
	assertResultCode(t, output, ResultInfoOnly)
}

// TestListViewJSON verifies that `todoat -y --json list` returns valid JSON
func TestListViewJSON(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create a list
	Execute([]string{"-y", "list", "create", "JSONTest"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// View lists with JSON output
	exitCode := Execute([]string{"-y", "--json", "list"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	// Should contain JSON array indicators
	assertContains(t, output, "[")
	assertContains(t, output, "]")
	assertContains(t, output, "JSONTest")
}

// TestListCreateJSON verifies that `todoat -y --json list create "Test"` returns JSON
func TestListCreateJSON(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create list with JSON output
	exitCode := Execute([]string{"-y", "--json", "list", "create", "JSONCreate"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	// Should contain JSON object indicators
	assertContains(t, output, "{")
	assertContains(t, output, "}")
	assertContains(t, output, "JSONCreate")
}

// =============================================================================
// JSON Output Tests for Task Commands (008-json-output)
// =============================================================================

// TestJSONFlagParsing verifies that --json flag is recognized and sets output mode
func TestJSONFlagParsing(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create a list first
	Execute([]string{"-y", "list", "create", "FlagTest"}, &stdout, &stderr, cfg)
	stdout.Reset()
	stderr.Reset()

	// Test that --json flag is accepted without error
	exitCode := Execute([]string{"-y", "--json", "FlagTest"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	// JSON output should contain JSON structure, not plain text
	assertContains(t, output, "{")
	assertNotContains(t, output, "Tasks in 'FlagTest'")
}

// TestListTasksJSON verifies that `todoat -y --json MyList` returns valid JSON with tasks array
func TestListTasksJSON(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create a list and add tasks
	Execute([]string{"-y", "list", "create", "TaskListJSON"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "TaskListJSON", "add", "First Task"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "TaskListJSON", "add", "Second Task"}, &stdout, &stderr, cfg)
	stdout.Reset()
	stderr.Reset()

	// List tasks with JSON output
	exitCode := Execute([]string{"-y", "--json", "TaskListJSON"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	// Should contain JSON with tasks array
	assertContains(t, output, `"tasks"`)
	assertContains(t, output, `"list"`)
	assertContains(t, output, `"TaskListJSON"`)
	assertContains(t, output, `"First Task"`)
	assertContains(t, output, `"Second Task"`)
	assertContains(t, output, `"result"`)
	assertContains(t, output, `"INFO_ONLY"`)
}

// TestAddTaskJSON verifies that `todoat -y --json MyList add "Task"` returns JSON with task info and result
func TestAddTaskJSON(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create a list first
	Execute([]string{"-y", "list", "create", "AddJSON"}, &stdout, &stderr, cfg)
	stdout.Reset()
	stderr.Reset()

	// Add task with JSON output
	exitCode := Execute([]string{"-y", "--json", "AddJSON", "add", "New JSON Task"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	// Should contain JSON with action and task
	assertContains(t, output, `"action"`)
	assertContains(t, output, `"add"`)
	assertContains(t, output, `"task"`)
	assertContains(t, output, `"New JSON Task"`)
	assertContains(t, output, `"result"`)
	assertContains(t, output, `"ACTION_COMPLETED"`)
}

// TestUpdateTaskJSON verifies that `todoat -y --json MyList update "Task" -s DONE` returns JSON with updated task
func TestUpdateTaskJSON(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create a list and add a task
	Execute([]string{"-y", "list", "create", "UpdateJSON"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "UpdateJSON", "add", "Task To Update"}, &stdout, &stderr, cfg)
	stdout.Reset()
	stderr.Reset()

	// Update task with JSON output
	exitCode := Execute([]string{"-y", "--json", "UpdateJSON", "update", "Task To Update", "-s", "DONE"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	// Should contain JSON with action and updated task
	assertContains(t, output, `"action"`)
	assertContains(t, output, `"update"`)
	assertContains(t, output, `"task"`)
	assertContains(t, output, `"Task To Update"`)
	assertContains(t, output, `"result"`)
	assertContains(t, output, `"ACTION_COMPLETED"`)
}

// TestDeleteTaskJSON verifies that `todoat -y --json MyList delete "Task"` returns JSON with result
func TestDeleteTaskJSON(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create a list and add a task
	Execute([]string{"-y", "list", "create", "DeleteJSON"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "DeleteJSON", "add", "Task To Delete"}, &stdout, &stderr, cfg)
	stdout.Reset()
	stderr.Reset()

	// Delete task with JSON output
	exitCode := Execute([]string{"-y", "--json", "DeleteJSON", "delete", "Task To Delete"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	// Should contain JSON with action and result
	assertContains(t, output, `"action"`)
	assertContains(t, output, `"delete"`)
	assertContains(t, output, `"result"`)
	assertContains(t, output, `"ACTION_COMPLETED"`)
}

// TestErrorJSON verifies that error conditions return JSON error with result: "ERROR"
func TestErrorJSON(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create a list first
	Execute([]string{"-y", "list", "create", "ErrorTestList"}, &stdout, &stderr, cfg)
	stdout.Reset()
	stderr.Reset()

	// Try to delete a non-existent task with JSON output (this triggers an error)
	exitCode := Execute([]string{"-y", "--json", "ErrorTestList", "delete", "NonExistentTask"}, &stdout, &stderr, cfg)

	// Should return non-zero exit code
	if exitCode == 0 {
		t.Errorf("expected non-zero exit code for error, got 0")
	}

	output := stdout.String()
	// Should contain JSON error
	assertContains(t, output, `"error"`)
	assertContains(t, output, `"result"`)
	assertContains(t, output, `"ERROR"`)
}

// TestJSONResultCodes verifies that all JSON responses include "result" field
func TestJSONResultCodes(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create list
	Execute([]string{"-y", "list", "create", "ResultCodeTest"}, &stdout, &stderr, cfg)
	stdout.Reset()
	stderr.Reset()

	// Test INFO_ONLY result for listing tasks
	exitCode := Execute([]string{"-y", "--json", "ResultCodeTest"}, &stdout, &stderr, cfg)
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout.String(), `"result"`)
	assertContains(t, stdout.String(), `"INFO_ONLY"`)
	stdout.Reset()
	stderr.Reset()

	// Test ACTION_COMPLETED result for add
	exitCode = Execute([]string{"-y", "--json", "ResultCodeTest", "add", "Test Task"}, &stdout, &stderr, cfg)
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout.String(), `"result"`)
	assertContains(t, stdout.String(), `"ACTION_COMPLETED"`)
	stdout.Reset()
	stderr.Reset()

	// Test ACTION_COMPLETED result for update
	exitCode = Execute([]string{"-y", "--json", "ResultCodeTest", "update", "Test Task", "-s", "DONE"}, &stdout, &stderr, cfg)
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout.String(), `"result"`)
	assertContains(t, stdout.String(), `"ACTION_COMPLETED"`)
	stdout.Reset()
	stderr.Reset()

	// Test ACTION_COMPLETED result for delete
	Execute([]string{"-y", "ResultCodeTest", "add", "Delete Me"}, &stdout, &stderr, cfg)
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "--json", "ResultCodeTest", "delete", "Delete Me"}, &stdout, &stderr, cfg)
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout.String(), `"result"`)
	assertContains(t, stdout.String(), `"ACTION_COMPLETED"`)
}

// =============================================================================
// Priority Filtering Tests (009-priority-filtering)
// =============================================================================

// TestPriorityFilterSingle verifies that `todoat -y MyList -p 1` shows only priority 1 tasks
func TestPriorityFilterSingle(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create tasks with different priorities
	Execute([]string{"-y", "Work", "add", "Priority 1 task", "-p", "1"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 2 task", "-p", "2"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 5 task", "-p", "5"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter to show only priority 1 tasks
	exitCode := Execute([]string{"-y", "Work", "-p", "1"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Priority 1 task")
	assertNotContains(t, output, "Priority 2 task")
	assertNotContains(t, output, "Priority 5 task")
	assertResultCode(t, output, ResultInfoOnly)
}

// TestPriorityFilterRange verifies that `todoat -y MyList -p 1,2,3` shows tasks with priority 1, 2, or 3
func TestPriorityFilterRange(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create tasks with different priorities
	Execute([]string{"-y", "Work", "add", "Priority 1 task", "-p", "1"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 2 task", "-p", "2"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 3 task", "-p", "3"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 5 task", "-p", "5"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 7 task", "-p", "7"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter to show priorities 1, 2, 3
	exitCode := Execute([]string{"-y", "Work", "-p", "1,2,3"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Priority 1 task")
	assertContains(t, output, "Priority 2 task")
	assertContains(t, output, "Priority 3 task")
	assertNotContains(t, output, "Priority 5 task")
	assertNotContains(t, output, "Priority 7 task")
	assertResultCode(t, output, ResultInfoOnly)
}

// TestPriorityFilterHigh verifies that `todoat -y MyList -p high` shows priorities 1-4
func TestPriorityFilterHigh(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create tasks with different priorities
	Execute([]string{"-y", "Work", "add", "Priority 1 task", "-p", "1"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 4 task", "-p", "4"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 5 task", "-p", "5"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 9 task", "-p", "9"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter using 'high' alias
	exitCode := Execute([]string{"-y", "Work", "-p", "high"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Priority 1 task")
	assertContains(t, output, "Priority 4 task")
	assertNotContains(t, output, "Priority 5 task")
	assertNotContains(t, output, "Priority 9 task")
	assertResultCode(t, output, ResultInfoOnly)
}

// TestPriorityFilterMedium verifies that `todoat -y MyList -p medium` shows priority 5
func TestPriorityFilterMedium(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create tasks with different priorities
	Execute([]string{"-y", "Work", "add", "Priority 1 task", "-p", "1"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 5 task", "-p", "5"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 6 task", "-p", "6"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter using 'medium' alias
	exitCode := Execute([]string{"-y", "Work", "-p", "medium"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertNotContains(t, output, "Priority 1 task")
	assertContains(t, output, "Priority 5 task")
	assertNotContains(t, output, "Priority 6 task")
	assertResultCode(t, output, ResultInfoOnly)
}

// TestPriorityFilterLow verifies that `todoat -y MyList -p low` shows priorities 6-9
func TestPriorityFilterLow(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create tasks with different priorities
	Execute([]string{"-y", "Work", "add", "Priority 1 task", "-p", "1"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 5 task", "-p", "5"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 6 task", "-p", "6"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 9 task", "-p", "9"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter using 'low' alias
	exitCode := Execute([]string{"-y", "Work", "-p", "low"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertNotContains(t, output, "Priority 1 task")
	assertNotContains(t, output, "Priority 5 task")
	assertContains(t, output, "Priority 6 task")
	assertContains(t, output, "Priority 9 task")
	assertResultCode(t, output, ResultInfoOnly)
}

// TestPriorityFilterUndefined verifies that `todoat -y MyList -p 0` shows tasks with no priority set
func TestPriorityFilterUndefined(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create tasks with and without priority
	Execute([]string{"-y", "Work", "add", "No priority task"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 1 task", "-p", "1"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 5 task", "-p", "5"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter to show only tasks with no priority (priority 0)
	exitCode := Execute([]string{"-y", "Work", "-p", "0"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "No priority task")
	assertNotContains(t, output, "Priority 1 task")
	assertNotContains(t, output, "Priority 5 task")
	assertResultCode(t, output, ResultInfoOnly)
}

// TestPriorityFilterNoMatch verifies that `todoat -y MyList -p 1` with no matching tasks returns INFO_ONLY with message
func TestPriorityFilterNoMatch(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create tasks with priority 5 only
	Execute([]string{"-y", "Work", "add", "Priority 5 task", "-p", "5"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter for priority 1 (no matches)
	exitCode := Execute([]string{"-y", "Work", "-p", "1"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	// Should show a message about no tasks matching
	if !strings.Contains(strings.ToLower(output), "no") || !strings.Contains(strings.ToLower(output), "task") {
		t.Errorf("expected message about no matching tasks, got: %s", output)
	}
	assertResultCode(t, output, ResultInfoOnly)
}

// TestPriorityFilterJSON verifies that `todoat -y --json MyList -p 1` returns filtered JSON result
func TestPriorityFilterJSON(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create tasks with different priorities
	Execute([]string{"-y", "Work", "add", "Priority 1 task", "-p", "1"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Priority 5 task", "-p", "5"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter with JSON output
	exitCode := Execute([]string{"-y", "--json", "Work", "-p", "1"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	// Should contain JSON with only priority 1 task
	assertContains(t, output, `"Priority 1 task"`)
	assertNotContains(t, output, `"Priority 5 task"`)
	assertContains(t, output, `"result"`)
	assertContains(t, output, `"INFO_ONLY"`)
}

// TestPriorityFilterInvalid verifies that `todoat -y MyList -p 10` returns ERROR for invalid priority
func TestPriorityFilterInvalid(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create a task
	Execute([]string{"-y", "Work", "add", "Some task"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Try invalid priority filter
	exitCode := Execute([]string{"-y", "Work", "-p", "10"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 1)
	assertResultCode(t, stdout.String(), ResultError)
}

// TestPriorityFilterCombinedWithStatus verifies combined status and priority filters work
func TestPriorityFilterCombinedWithStatus(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create tasks with different priorities and statuses
	Execute([]string{"-y", "Work", "add", "High priority TODO", "-p", "1"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "High priority DONE", "-p", "1"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "complete", "High priority DONE"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Low priority TODO", "-p", "7"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter for TODO tasks with high priority
	exitCode := Execute([]string{"-y", "Work", "-s", "TODO", "-p", "high"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "High priority TODO")
	assertNotContains(t, output, "High priority DONE")
	assertNotContains(t, output, "Low priority TODO")
	assertResultCode(t, output, ResultInfoOnly)
}

// =============================================================================
// Task Dates Tests (011-task-dates)
// =============================================================================

// TestAddTaskWithDueDate verifies that `todoat -y MyList add "Task" --due-date 2026-01-31` sets due date
func TestAddTaskWithDueDate(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"-y", "Work", "add", "Task with due", "--due-date", "2026-01-31"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)

	// Verify by listing tasks with JSON to check due_date
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "--json", "Work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Task with due")
	assertContains(t, output, "2026-01-31")
}

// TestAddTaskWithStartDate verifies that `todoat -y MyList add "Task" --start-date 2026-01-15` sets start date
func TestAddTaskWithStartDate(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"-y", "Work", "add", "Task with start", "--start-date", "2026-01-15"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)

	// Verify by listing tasks with JSON to check start_date
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "--json", "Work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Task with start")
	assertContains(t, output, "2026-01-15")
}

// TestAddTaskWithBothDates verifies that both --due-date and --start-date can be set together
func TestAddTaskWithBothDates(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"-y", "Work", "add", "Task with both dates", "--due-date", "2026-01-31", "--start-date", "2026-01-15"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)

	// Verify by listing tasks with JSON to check both dates
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "--json", "Work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Task with both dates")
	assertContains(t, output, "2026-01-31")
	assertContains(t, output, "2026-01-15")
}

// TestUpdateTaskDueDate verifies that `todoat -y MyList update "Task" --due-date 2026-02-15` updates due date
func TestUpdateTaskDueDate(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task with a due date
	Execute([]string{"-y", "Work", "add", "Update date task", "--due-date", "2026-01-31"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Update the due date
	exitCode := Execute([]string{"-y", "Work", "update", "Update date task", "--due-date", "2026-02-15"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)

	// Verify updated due date
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "--json", "Work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Update date task")
	assertContains(t, output, "2026-02-15")
	assertNotContains(t, output, "2026-01-31")
}

// TestClearTaskDueDate verifies that `todoat -y MyList update "Task" --due-date ""` clears due date
func TestClearTaskDueDate(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task with a due date
	Execute([]string{"-y", "Work", "add", "Clear date task", "--due-date", "2026-01-31"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Clear the due date
	exitCode := Execute([]string{"-y", "Work", "update", "Clear date task", "--due-date", ""}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)

	// Verify due date is cleared - JSON output should not have the date
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "--json", "Work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Clear date task")
	// Due date should be empty or null in JSON
	assertNotContains(t, output, "2026-01-31")
}

// TestInvalidDateFormat verifies that `todoat -y MyList add "Task" --due-date "invalid"` returns ERROR
func TestInvalidDateFormat(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"-y", "Work", "add", "Invalid date task", "--due-date", "invalid"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 1)
	assertResultCode(t, stdout.String(), ResultError)
}

// TestDateFormatValidation verifies that `todoat -y MyList add "Task" --due-date "01-31-2026"` returns ERROR (wrong format)
func TestDateFormatValidation(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Wrong format: MM-DD-YYYY instead of YYYY-MM-DD
	exitCode := Execute([]string{"-y", "Work", "add", "Wrong format task", "--due-date", "01-31-2026"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 1)
	assertResultCode(t, stdout.String(), ResultError)
}

// TestTaskDatesInJSON verifies that `todoat -y --json MyList` includes due_date and start_date fields
func TestTaskDatesInJSON(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task with both dates
	Execute([]string{"-y", "Work", "add", "JSON date task", "--due-date", "2026-01-31", "--start-date", "2026-01-15"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Get tasks as JSON
	exitCode := Execute([]string{"-y", "--json", "Work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, `"due_date"`)
	assertContains(t, output, `"start_date"`)
	assertContains(t, output, "2026-01-31")
	assertContains(t, output, "2026-01-15")
}

// TestCompletedTimestamp verifies that `todoat -y MyList complete "Task"` sets completed timestamp automatically
func TestCompletedTimestamp(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task
	Execute([]string{"-y", "Work", "add", "Task to complete"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Complete the task
	exitCode := Execute([]string{"-y", "Work", "complete", "Task to complete"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)

	// Get tasks as JSON and verify completed timestamp is set
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "--json", "Work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Task to complete")
	assertContains(t, output, `"completed"`)
}

// =============================================================================
// Tag Filtering Tests (012-tag-filtering)
// =============================================================================

// TestAddTaskWithTag verifies that `todoat -y MyList add "Task" --tag work` adds task with tag
func TestAddTaskWithTag(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"-y", "Work", "add", "Tagged task", "--tag", "work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)

	// Verify by listing tasks with JSON to check tags
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "--json", "Work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Tagged task")
	assertContains(t, output, "work")
}

// TestAddTaskMultipleTags verifies that `todoat -y MyList add "Task" --tag work --tag urgent` adds task with multiple tags
func TestAddTaskMultipleTags(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"-y", "Work", "add", "Multi-tagged task", "--tag", "work", "--tag", "urgent"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)

	// Verify by listing tasks with JSON to check tags
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "--json", "Work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Multi-tagged task")
	assertContains(t, output, "work")
	assertContains(t, output, "urgent")
}

// TestAddTaskCommaSeparatedTags verifies that `todoat -y MyList add "Task" --tag "work,urgent"` adds task with comma-separated tags
func TestAddTaskCommaSeparatedTags(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"-y", "Work", "add", "Comma-tagged task", "--tag", "work,urgent"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)

	// Verify by listing tasks with JSON to check tags
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "--json", "Work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Comma-tagged task")
	assertContains(t, output, "work")
	assertContains(t, output, "urgent")
}

// TestUpdateTaskTags verifies that `todoat -y MyList update "Task" --tag home` updates task tags
func TestUpdateTaskTags(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task with a tag
	Execute([]string{"-y", "Work", "add", "Update tag task", "--tag", "work"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Update the tag
	exitCode := Execute([]string{"-y", "Work", "update", "Update tag task", "--tag", "home"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)

	// Verify updated tag
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "--json", "Work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Update tag task")
	assertContains(t, output, "home")
	assertNotContains(t, output, `"work"`)
}

// TestClearTaskTags verifies that `todoat -y MyList update "Task" --tag ""` clears task tags
func TestClearTaskTags(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task with a tag
	Execute([]string{"-y", "Work", "add", "Clear tag task", "--tag", "work"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Clear the tag
	exitCode := Execute([]string{"-y", "Work", "update", "Clear tag task", "--tag", ""}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)

	// Verify tag is cleared
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "--json", "Work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Clear tag task")
	// The tags field should be empty or not contain "work"
	// We check that the specific tag value is no longer present
	if strings.Contains(output, `"tags":["work"]`) {
		t.Errorf("expected tags to be cleared, but still found work tag in output: %s", output)
	}
}

// TestFilterByTag verifies that `todoat -y MyList --tag work` shows only tasks with "work" tag
func TestFilterByTag(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add tasks with different tags
	Execute([]string{"-y", "Work", "add", "Work task", "--tag", "work"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Home task", "--tag", "home"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "No tag task"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter by work tag
	exitCode := Execute([]string{"-y", "Work", "--tag", "work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Work task")
	assertNotContains(t, output, "Home task")
	assertNotContains(t, output, "No tag task")
	assertResultCode(t, output, ResultInfoOnly)
}

// TestFilterByMultipleTags verifies that `todoat -y MyList --tag work --tag urgent` shows tasks with ANY of the tags (OR logic)
func TestFilterByMultipleTags(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add tasks with different tags
	Execute([]string{"-y", "Work", "add", "Work task", "--tag", "work"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Urgent task", "--tag", "urgent"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Home task", "--tag", "home"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter by work OR urgent tag
	exitCode := Execute([]string{"-y", "Work", "--tag", "work", "--tag", "urgent"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Work task")
	assertContains(t, output, "Urgent task")
	assertNotContains(t, output, "Home task")
	assertResultCode(t, output, ResultInfoOnly)
}

// TestFilterTagNoMatch verifies that `todoat -y MyList --tag nonexistent` returns INFO_ONLY with message
func TestFilterTagNoMatch(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add a task with a different tag
	Execute([]string{"-y", "Work", "add", "Some task", "--tag", "work"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter by non-existent tag
	exitCode := Execute([]string{"-y", "Work", "--tag", "nonexistent"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	// Should show a message about no matching tasks
	if !strings.Contains(strings.ToLower(output), "no") || !strings.Contains(strings.ToLower(output), "task") {
		t.Errorf("expected message about no matching tasks, got: %s", output)
	}
	assertResultCode(t, output, ResultInfoOnly)
}

// TestFilterTagJSON verifies that `todoat -y --json MyList --tag work` returns filtered JSON result with tags array
func TestFilterTagJSON(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add tasks with different tags
	Execute([]string{"-y", "Work", "add", "Work task", "--tag", "work"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Home task", "--tag", "home"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter with JSON output
	exitCode := Execute([]string{"-y", "--json", "Work", "--tag", "work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	// Should contain JSON with only work-tagged task
	assertContains(t, output, `"Work task"`)
	assertNotContains(t, output, `"Home task"`)
	assertContains(t, output, `"tags"`)
	assertContains(t, output, `"work"`)
	assertContains(t, output, `"result"`)
	assertContains(t, output, `"INFO_ONLY"`)
}

// TestFilterTagCombined verifies that `todoat -y MyList -s TODO --tag work` combined with status filter works
func TestFilterTagCombined(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Add tasks with different statuses and tags
	Execute([]string{"-y", "Work", "add", "Work TODO task", "--tag", "work"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Work DONE task", "--tag", "work"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "complete", "Work DONE task"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "Work", "add", "Home TODO task", "--tag", "home"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Filter by TODO status AND work tag
	exitCode := Execute([]string{"-y", "Work", "-s", "TODO", "--tag", "work"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "Work TODO task")
	assertNotContains(t, output, "Work DONE task")
	assertNotContains(t, output, "Home TODO task")
	assertResultCode(t, output, ResultInfoOnly)
}

// =============================================================================
// List Management Tests (013-list-management)
// =============================================================================

// TestListDelete verifies that `todoat -y list delete "ListName"` soft-deletes a list
func TestListDelete(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create a list
	Execute([]string{"-y", "list", "create", "ToDelete"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Delete the list
	exitCode := Execute([]string{"-y", "list", "delete", "ToDelete"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout.String(), "ToDelete")
	assertResultCode(t, stdout.String(), ResultActionCompleted)

	// Verify list is no longer visible in normal list view
	stdout.Reset()
	stderr.Reset()
	Execute([]string{"-y", "list"}, &stdout, &stderr, cfg)
	assertNotContains(t, stdout.String(), "ToDelete")
}

// TestListDeleteNotFound verifies that deleting a non-existent list returns ERROR
func TestListDeleteNotFound(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Try to delete non-existent list
	exitCode := Execute([]string{"-y", "list", "delete", "NonExistent"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 1)
	assertResultCode(t, stdout.String(), ResultError)
}

// TestListTrash verifies that `todoat -y list trash` displays deleted lists
func TestListTrash(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create and delete a list
	Execute([]string{"-y", "list", "create", "TrashTest"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "list", "delete", "TrashTest"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// View trash
	exitCode := Execute([]string{"-y", "list", "trash"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout.String(), "TrashTest")
	assertResultCode(t, stdout.String(), ResultInfoOnly)
}

// TestListTrashEmpty verifies that viewing empty trash returns INFO_ONLY
func TestListTrashEmpty(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// View trash with no deleted lists
	exitCode := Execute([]string{"-y", "list", "trash"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultInfoOnly)
}

// TestListRestore verifies that `todoat -y list trash restore "Name"` restores a deleted list
func TestListRestore(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create, delete, then restore a list
	Execute([]string{"-y", "list", "create", "RestoreTest"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "list", "delete", "RestoreTest"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Restore the list
	exitCode := Execute([]string{"-y", "list", "trash", "restore", "RestoreTest"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout.String(), "RestoreTest")
	assertResultCode(t, stdout.String(), ResultActionCompleted)

	// Verify list is visible in normal list view again
	stdout.Reset()
	stderr.Reset()
	Execute([]string{"-y", "list"}, &stdout, &stderr, cfg)
	assertContains(t, stdout.String(), "RestoreTest")
}

// TestListRestoreNotInTrash verifies that restoring an active list returns ERROR
func TestListRestoreNotInTrash(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create a list but don't delete it
	Execute([]string{"-y", "list", "create", "ActiveList"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Try to restore an active list
	exitCode := Execute([]string{"-y", "list", "trash", "restore", "ActiveList"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 1)
	assertResultCode(t, stdout.String(), ResultError)
}

// TestListPurge verifies that `todoat -y list trash purge "Name"` permanently deletes
func TestListPurge(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create, delete, then purge a list
	Execute([]string{"-y", "list", "create", "PurgeTest"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "list", "delete", "PurgeTest"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Purge the list
	exitCode := Execute([]string{"-y", "list", "trash", "purge", "PurgeTest"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	assertResultCode(t, stdout.String(), ResultActionCompleted)

	// Verify list is not in trash anymore
	stdout.Reset()
	stderr.Reset()
	Execute([]string{"-y", "list", "trash"}, &stdout, &stderr, cfg)
	assertNotContains(t, stdout.String(), "PurgeTest")
}

// TestListInfo verifies that `todoat -y list info "Name"` shows list details
func TestListInfo(t *testing.T) {
	cfg, cleanup := testWithDB(t)
	defer cleanup()

	var stdout, stderr bytes.Buffer

	// Create a list and add some tasks
	Execute([]string{"-y", "list", "create", "InfoTest"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "InfoTest", "add", "Task 1"}, &stdout, &stderr, cfg)
	Execute([]string{"-y", "InfoTest", "add", "Task 2"}, &stdout, &stderr, cfg)

	stdout.Reset()
	stderr.Reset()

	// Get list info
	exitCode := Execute([]string{"-y", "list", "info", "InfoTest"}, &stdout, &stderr, cfg)

	assertExitCode(t, exitCode, 0)
	output := stdout.String()
	assertContains(t, output, "InfoTest")
	// Should show task count (2 tasks)
	assertContains(t, output, "2")
	assertResultCode(t, output, ResultInfoOnly)
}
