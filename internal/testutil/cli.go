// Package testutil provides shared test utilities for CLI testing across packages.
// This enables co-located CLI tests while maintaining consistent test infrastructure.
package testutil

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"todoat/cmd/todoat/cmd"
)

// CLITest provides a test helper for running CLI commands in isolation.
type CLITest struct {
	t      *testing.T
	cfg    *cmd.Config
	tmpDir string
}

// NewCLITest creates a new CLI test helper with an isolated in-memory database.
func NewCLITest(t *testing.T) *CLITest {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	cfg := &cmd.Config{
		NoPrompt: true,
		DBPath:   dbPath,
	}

	return &CLITest{
		t:      t,
		cfg:    cfg,
		tmpDir: tmpDir,
	}
}

// NewCLITestWithViews creates a new CLI test helper with views directory support.
func NewCLITestWithViews(t *testing.T) (*CLITest, string) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	viewsDir := tmpDir + "/views"

	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		t.Fatalf("failed to create views directory: %v", err)
	}

	cfg := &cmd.Config{
		NoPrompt:  true,
		DBPath:    dbPath,
		ViewsPath: viewsDir,
	}

	return &CLITest{
		t:      t,
		cfg:    cfg,
		tmpDir: tmpDir,
	}, viewsDir
}

// NewCLITestWithNotification creates a new CLI test helper with notification support.
func NewCLITestWithNotification(t *testing.T) *CLITest {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	notificationLogPath := tmpDir + "/notifications.log"

	cfg := &cmd.Config{
		NoPrompt:            true,
		DBPath:              dbPath,
		NotificationLogPath: notificationLogPath,
		NotificationMock:    true, // Use mock executor for OS notifications
	}

	return &CLITest{
		t:      t,
		cfg:    cfg,
		tmpDir: tmpDir,
	}
}

// Config returns the test configuration.
func (c *CLITest) Config() *cmd.Config {
	return c.cfg
}

// TmpDir returns the temporary directory for the test.
func (c *CLITest) TmpDir() string {
	return c.tmpDir
}

// Execute runs a CLI command with the given arguments and returns stdout, stderr, and exit code.
func (c *CLITest) Execute(args ...string) (stdout, stderr string, exitCode int) {
	c.t.Helper()

	var stdoutBuf, stderrBuf bytes.Buffer
	exitCode = cmd.Execute(args, &stdoutBuf, &stderrBuf, c.cfg)
	return stdoutBuf.String(), stderrBuf.String(), exitCode
}

// MustExecute runs a CLI command and fails the test if exit code is non-zero.
func (c *CLITest) MustExecute(args ...string) string {
	c.t.Helper()

	stdout, stderr, exitCode := c.Execute(args...)
	if exitCode != 0 {
		c.t.Fatalf("expected exit code 0, got %d: stdout=%s stderr=%s", exitCode, stdout, stderr)
	}
	return stdout
}

// ExecuteAndFail runs a CLI command and fails the test if exit code is zero.
func (c *CLITest) ExecuteAndFail(args ...string) (stdout, stderr string) {
	c.t.Helper()

	stdout, stderr, exitCode := c.Execute(args...)
	if exitCode == 0 {
		c.t.Fatalf("expected non-zero exit code, got 0: stdout=%s", stdout)
	}
	return stdout, stderr
}

// AssertContains fails the test if output doesn't contain expected string.
func AssertContains(t *testing.T, output, expected string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, output)
	}
}

// AssertNotContains fails the test if output contains unexpected string.
func AssertNotContains(t *testing.T, output, unexpected string) {
	t.Helper()
	if strings.Contains(output, unexpected) {
		t.Errorf("expected output NOT to contain %q, got:\n%s", unexpected, output)
	}
}

// AssertExitCode fails the test if exit code doesn't match expected.
func AssertExitCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("expected exit code %d, got %d", want, got)
	}
}

// AssertResultCode verifies that the output ends with the expected result code.
func AssertResultCode(t *testing.T, output, expectedCode string) {
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

// Result code constants for convenience.
const (
	ResultActionCompleted = cmd.ResultActionCompleted
	ResultInfoOnly        = cmd.ResultInfoOnly
	ResultError           = cmd.ResultError
)
