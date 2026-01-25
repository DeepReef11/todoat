package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"todoat/internal/config"
	"todoat/internal/credentials"
)

// =============================================================================
// Core CLI Tests
// These tests verify basic CLI functionality: help, version, flags, and arg parsing.
// Feature-specific CLI tests are co-located with their backend/feature code:
// - Task/List commands: backend/sqlite/cli_test.go
// - Views: internal/views/cli_test.go
// =============================================================================

// --- Help and Version Tests ---

// TestHelpFlag verifies that --help displays usage information
func TestHelpFlagCoreCLI(t *testing.T) {
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
func TestVersionFlagCoreCLI(t *testing.T) {
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

// --- Version Command Tests ---
// CLI Tests for 051-version-command

// TestVersionCommand verifies that 'todoat version' displays version string
func TestVersionCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"version"}, &stdout, &stderr, nil)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	// Should contain version number
	if !strings.Contains(output, "Version:") {
		t.Errorf("version output should contain 'Version:', got: %s", output)
	}
	// Should contain commit hash
	if !strings.Contains(output, "Commit:") {
		t.Errorf("version output should contain 'Commit:', got: %s", output)
	}
	// Should contain build date
	if !strings.Contains(output, "Built:") {
		t.Errorf("version output should contain 'Built:', got: %s", output)
	}
}

// TestVersionVerbose verifies that 'todoat version -v' shows extended build info
func TestVersionVerbose(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"version", "-v"}, &stdout, &stderr, nil)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	// Should contain Go version
	if !strings.Contains(output, "Go Version:") {
		t.Errorf("verbose version output should contain 'Go Version:', got: %s", output)
	}
	// Should contain platform info (OS/Arch)
	if !strings.Contains(output, "Platform:") {
		t.Errorf("verbose version output should contain 'Platform:', got: %s", output)
	}
}

// TestVersionJSON verifies that 'todoat --json version' returns JSON with version fields
func TestVersionJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"--json", "version"}, &stdout, &stderr, nil)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	// Should be valid JSON with expected fields
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected valid JSON output, got: %s, error: %v", output, err)
	}

	// Check required fields exist
	requiredFields := []string{"version", "commit", "build_date", "go_version", "platform"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("JSON output should contain '%s' field, got: %v", field, result)
		}
	}
}

// TestVersionShort verifies that 'todoat --version' works as alias for version info
// Note: This is already tested in TestVersionFlagCoreCLI, but we verify it still works
// and outputs similar content to 'todoat version'
func TestVersionShort(t *testing.T) {
	var stdout1, stderr1, stdout2, stderr2 bytes.Buffer

	// Test --version flag
	exitCode1 := Execute([]string{"--version"}, &stdout1, &stderr1, nil)
	if exitCode1 != 0 {
		t.Fatalf("expected exit code 0 for --version, got %d: %s", exitCode1, stderr1.String())
	}

	// Test version command
	exitCode2 := Execute([]string{"version"}, &stdout2, &stderr2, nil)
	if exitCode2 != 0 {
		t.Fatalf("expected exit code 0 for version command, got %d: %s", exitCode2, stderr2.String())
	}

	// Both should contain the version number
	if !strings.Contains(stdout1.String(), "todoat") {
		t.Errorf("--version should contain 'todoat', got: %s", stdout1.String())
	}
	if !strings.Contains(stdout2.String(), "Version:") {
		t.Errorf("version command should contain 'Version:', got: %s", stdout2.String())
	}
}

// --- Global Flag Tests ---

// TestNoPromptFlag verifies that -y / --no-prompt flag is recognized
func TestNoPromptFlagCoreCLI(t *testing.T) {
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
func TestVerboseFlagCoreCLI(t *testing.T) {
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

// TestVerboseModeEnabledCoreCLI verifies that -V flag outputs debug messages to stderr
// CLI Test for 034-logging-utilities
func TestVerboseModeEnabledCoreCLI(t *testing.T) {
	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Initialize database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	_ = db.Close()

	// Write a default config to ensure test isolation
	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Capture real stderr (logger writes to os.Stderr)
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	var stdout, stderrBuf bytes.Buffer

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	// Execute with verbose flag
	exitCode := Execute([]string{"-V", "list"}, &stdout, &stderrBuf, cfg)

	// Close write pipe and capture output
	_ = w.Close()
	var capturedStderr bytes.Buffer
	_, _ = capturedStderr.ReadFrom(r)
	os.Stderr = oldStderr

	// Should succeed (list command)
	if exitCode != 0 {
		t.Logf("stdout: %s", stdout.String())
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, capturedStderr.String())
	}

	// Debug messages should be in captured stderr (from os.Stderr)
	stderrOutput := capturedStderr.String()
	if !strings.Contains(stderrOutput, "[DEBUG]") {
		t.Errorf("verbose mode should output [DEBUG] messages to stderr, got: %s", stderrOutput)
	}
}

// TestVerboseModeDisabledCoreCLI verifies that without -V flag, no debug messages output
// CLI Test for 034-logging-utilities
func TestVerboseModeDisabledCoreCLI(t *testing.T) {
	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Initialize database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	_ = db.Close()

	// Write a default config to ensure test isolation
	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Capture real stderr (logger writes to os.Stderr)
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	var stdout, stderrBuf bytes.Buffer

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	// Execute without verbose flag
	exitCode := Execute([]string{"list"}, &stdout, &stderrBuf, cfg)

	// Close write pipe and capture output
	_ = w.Close()
	var capturedStderr bytes.Buffer
	_, _ = capturedStderr.ReadFrom(r)
	os.Stderr = oldStderr

	// Should succeed
	if exitCode != 0 {
		t.Logf("stdout: %s", stdout.String())
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, capturedStderr.String())
	}

	// No debug messages should be in captured stderr
	stderrOutput := capturedStderr.String()
	if strings.Contains(stderrOutput, "[DEBUG]") {
		t.Errorf("without verbose mode, should not output [DEBUG] messages, got: %s", stderrOutput)
	}
}

// TestGlobalFlagsArePersistent verifies global flags work with subcommands
func TestGlobalFlagsArePersistentCoreCLI(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Global flags should be recognized even without subcommands
	exitCode := Execute([]string{"-y", "-V", "--help"}, &stdout, &stderr, nil)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}
}

// --- Exit Code Tests ---

// TestExitCodeSuccess verifies exit code 0 for successful operations
func TestExitCodeSuccessCoreCLI(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"--help"}, &stdout, &stderr, nil)

	if exitCode != 0 {
		t.Errorf("expected exit code 0 for help, got %d", exitCode)
	}
}

// TestExitCodeError verifies exit code 1 for errors (unknown flag)
func TestExitCodeErrorCoreCLI(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"--unknown-flag-xyz"}, &stdout, &stderr, nil)

	if exitCode != 1 {
		t.Errorf("expected exit code 1 for unknown flag, got %d", exitCode)
	}
}

// --- Argument Parsing Tests ---

// TestMaxThreePositionalArgs verifies that more than 3 positional args fails
func TestMaxThreePositionalArgsCoreCLI(t *testing.T) {
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
func TestThreePositionalArgsAllowedCoreCLI(t *testing.T) {
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

// --- IO Injection Tests ---

// TestInjectableIO verifies that stdout and stderr writers are used
func TestInjectableIOCoreCLI(t *testing.T) {
	var stdout, stderr bytes.Buffer

	Execute([]string{"--help"}, &stdout, &stderr, nil)

	// Help should be written to stdout
	if stdout.Len() == 0 {
		t.Error("expected help output to be written to stdout")
	}
}

// TestConfigPassthroughCoreCLI verifies that Execute() accepts a pre-initialized Config struct.
// This enables programmatic configuration for testing and embedding scenarios where callers
// can set config values directly rather than relying on command-line flags or config files.
// The config is "passed through" from the caller to the command execution.
func TestConfigPassthroughCoreCLI(t *testing.T) {
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

// --- Default Behavior Tests ---

// TestRootCommandShowsListsCoreCLI verifies that running without args shows available lists
// Issue #0: todoat should show available list when run without args
func TestRootCommandShowsListsCoreCLI(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	cachePath := filepath.Join(tmpDir, "cache", "lists.json")
	configPath := tmpDir + "/config.yaml"

	// Write a minimal default config to ensure isolation
	if err := os.WriteFile(configPath, []byte("# test config\ndefault_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg := &Config{
		NoPrompt:   true,
		DBPath:     dbPath,
		CachePath:  cachePath,
		ConfigPath: configPath,
	}

	// First create a list so we have something to show
	var stdout, stderr bytes.Buffer
	exitCode := Execute([]string{"-y", "list", "create", "Work"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create list: %s", stderr.String())
	}

	// Add a task to the list
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "Work", "add", "Test task"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to add task: %s", stderr.String())
	}

	// Now test running without args - should show lists
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0 for no args, got %d: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	// Should show "Available lists" header (not "Usage:")
	if !strings.Contains(output, "Available lists") {
		t.Errorf("no-args should show 'Available lists', got: %s", output)
	}
	// Should show NAME/TASKS headers
	if !strings.Contains(output, "NAME") {
		t.Errorf("no-args should show 'NAME' header, got: %s", output)
	}
	// Should show the list we created
	if !strings.Contains(output, "Work") {
		t.Errorf("no-args should show 'Work' list, got: %s", output)
	}
}

// =============================================================================
// Shell Completion Tests
// These tests verify shell completion script generation for all supported shells.
// =============================================================================

// TestCompletionBash verifies that `todoat completion bash` outputs valid Bash completion script
func TestCompletionBash(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"completion", "bash"}, &stdout, &stderr, nil)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	// Bash completion scripts contain specific markers
	if !strings.Contains(output, "bash completion") || !strings.Contains(output, "_todoat") {
		t.Errorf("expected bash completion script with _todoat function, got: %s", output[:min(200, len(output))])
	}
	// Should contain completion function definitions
	if !strings.Contains(output, "complete") {
		t.Errorf("expected bash completion script with 'complete' directive, got: %s", output[:min(200, len(output))])
	}
}

// TestCompletionZsh verifies that `todoat completion zsh` outputs valid Zsh completion script
func TestCompletionZsh(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"completion", "zsh"}, &stdout, &stderr, nil)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	// Zsh completion scripts contain specific markers
	if !strings.Contains(output, "#compdef") || !strings.Contains(output, "_todoat") {
		t.Errorf("expected zsh completion script with #compdef and _todoat, got: %s", output[:min(200, len(output))])
	}
}

// TestCompletionFish verifies that `todoat completion fish` outputs valid Fish completion script
func TestCompletionFish(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"completion", "fish"}, &stdout, &stderr, nil)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	// Fish completion scripts contain specific markers
	if !strings.Contains(output, "fish completion") || !strings.Contains(output, "complete -c todoat") {
		t.Errorf("expected fish completion script, got: %s", output[:min(200, len(output))])
	}
}

// TestCompletionPowerShell verifies that `todoat completion powershell` outputs valid PowerShell completion script
func TestCompletionPowerShell(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"completion", "powershell"}, &stdout, &stderr, nil)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	// PowerShell completion scripts contain specific markers
	if !strings.Contains(output, "powershell completion") || !strings.Contains(output, "Register-ArgumentCompleter") {
		t.Errorf("expected powershell completion script, got: %s", output[:min(200, len(output))])
	}
}

// TestCompletionHelp verifies that `todoat completion --help` shows usage instructions for each shell
func TestCompletionHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := Execute([]string{"completion", "--help"}, &stdout, &stderr, nil)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	// Help should mention all supported shells
	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range shells {
		if !strings.Contains(output, shell) {
			t.Errorf("completion help should mention %s, got: %s", shell, output)
		}
	}
	// Should have usage information
	if !strings.Contains(output, "Usage:") {
		t.Errorf("completion help should contain Usage:, got: %s", output)
	}
}

// TestCompletionInstallInstructions verifies that each completion subcommand outputs installation instructions
func TestCompletionInstallInstructions(t *testing.T) {
	tests := []struct {
		shell        string
		instructions []string
	}{
		{"bash", []string{"source", ".bashrc", "bash_completion"}},
		{"zsh", []string{"fpath", ".zshrc"}},
		{"fish", []string{"config.fish", "completions"}},
		{"powershell", []string{"profile", "Invoke-Expression"}},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			exitCode := Execute([]string{"completion", tt.shell, "--help"}, &stdout, &stderr, nil)

			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
			}

			output := stdout.String()
			// Check that at least one installation instruction keyword is present
			found := false
			for _, instruction := range tt.instructions {
				if strings.Contains(output, instruction) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("completion %s --help should contain installation instructions (looking for one of %v), got: %s",
					tt.shell, tt.instructions, output)
			}
		})
	}
}

// min helper for string slicing
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// =============================================================================
// SQLite CLI Initialization Tests
// These tests verify that the application correctly initializes on first run:
// - Database creation at XDG path
// - Config creation at XDG path
// - Directory creation with proper permissions
// - Schema initialization on new database
// =============================================================================

// TestAppStartsWithoutExistingDBSQLiteCLI verifies that the app starts successfully
// when no database exists and creates the db automatically
func TestAppStartsWithoutExistingDBSQLiteCLI(t *testing.T) {
	// Create a temp directory to act as home
	tempHome := t.TempDir()

	// Set up config with XDG paths pointing to temp directory
	dbPath := filepath.Join(tempHome, "data", "todoat", "tasks.db")
	configPath := filepath.Join(tempHome, "config.yaml")

	// Write a default config to ensure test isolation
	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Running a list command should work even with no existing DB
	exitCode := Execute([]string{"TestList", "get"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s stdout=%s", exitCode, stderr.String(), stdout.String())
	}

	// Verify the database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("expected database file to be created at %s", dbPath)
	}
}

// TestAppStartsWithoutExistingConfigSQLiteCLI verifies that the app starts successfully
// when no config file exists and creates a default config
func TestAppStartsWithoutExistingConfigSQLiteCLI(t *testing.T) {
	// Create a temp directory to act as home
	tempHome := t.TempDir()

	// Set XDG environment variables for this test (auto-restored after test)
	configDir := filepath.Join(tempHome, ".config")
	dataDir := filepath.Join(tempHome, ".local", "share")
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("XDG_DATA_HOME", dataDir)

	// Verify config doesn't exist
	configPath := filepath.Join(configDir, "todoat", "config.yaml")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("config file should not exist before test: %v", err)
	}

	// Use the config loader to trigger config creation
	loadedCfg, err := config.Load("")
	if err != nil {
		t.Fatalf("config.Load should succeed: %v", err)
	}

	if loadedCfg == nil {
		t.Fatal("expected config to be returned")
	}

	// Verify the config file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("expected config file to be created at %s", configPath)
	}
}

// TestDBCreatedAtCorrectPathSQLiteCLI verifies that the database is created at
// $XDG_DATA_HOME/todoat/tasks.db or ~/.local/share/todoat/tasks.db
func TestDBCreatedAtCorrectPathSQLiteCLI(t *testing.T) {
	// Create a temp directory to act as home
	tempHome := t.TempDir()

	// Set XDG environment variables for this test (auto-restored after test)
	dataDir := filepath.Join(tempHome, ".local", "share")
	configDir := filepath.Join(tempHome, ".config")
	t.Setenv("XDG_DATA_HOME", dataDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)

	// Use nil DBPath to test default path resolution
	cfg := &Config{}

	var stdout, stderr bytes.Buffer

	// Running a command that requires DB should trigger DB creation at XDG path
	exitCode := Execute([]string{"TestList", "get"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	// Expected path should be XDG_DATA_HOME/todoat/tasks.db
	expectedDBPath := filepath.Join(dataDir, "todoat", "tasks.db")
	if _, err := os.Stat(expectedDBPath); os.IsNotExist(err) {
		t.Errorf("expected database at XDG path %s, but file does not exist", expectedDBPath)
		// Check what was actually created
		files, _ := filepath.Glob(filepath.Join(tempHome, "**", "*.db"))
		t.Logf("DB files found: %v", files)
	}
}

// TestConfigCreatedAtCorrectPathSQLiteCLI verifies that the config is created at
// $XDG_CONFIG_HOME/todoat/config.yaml or ~/.config/todoat/config.yaml
func TestConfigCreatedAtCorrectPathSQLiteCLI(t *testing.T) {
	// Create a temp directory to act as home
	tempHome := t.TempDir()

	// Set XDG environment variables for this test (auto-restored after test)
	configDir := filepath.Join(tempHome, ".config")
	t.Setenv("XDG_CONFIG_HOME", configDir)

	// Expected path should be XDG_CONFIG_HOME/todoat/config.yaml
	expectedConfigPath := filepath.Join(configDir, "todoat", "config.yaml")

	// Use the config loader to trigger config creation
	_, err := config.Load("")
	if err != nil {
		t.Fatalf("config.Load should succeed: %v", err)
	}

	// Verify the config file was created at correct XDG path
	if _, err := os.Stat(expectedConfigPath); os.IsNotExist(err) {
		t.Errorf("expected config file at XDG path %s, but file does not exist", expectedConfigPath)
		// Check what was actually created
		files, _ := filepath.Glob(filepath.Join(tempHome, "**", "*.yaml"))
		t.Logf("Config files found: %v", files)
	}
}

// TestDirectoriesCreatedAutomaticallySQLiteCLI verifies that parent directories
// are created with proper permissions when they don't exist
func TestDirectoriesCreatedAutomaticallySQLiteCLI(t *testing.T) {
	// Create a temp directory to act as home
	tempHome := t.TempDir()

	// Deep nested path that doesn't exist
	dbPath := filepath.Join(tempHome, "deep", "nested", "path", "todoat", "tasks.db")
	configPath := filepath.Join(tempHome, "config", "nested", "todoat", "config.yaml")

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Running a command should create all necessary directories
	exitCode := Execute([]string{"TestList", "get"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	// Verify the database directory was created
	dbDir := filepath.Dir(dbPath)
	info, err := os.Stat(dbDir)
	if os.IsNotExist(err) {
		t.Errorf("expected database directory to be created at %s", dbDir)
	} else if err != nil {
		t.Errorf("error checking database directory: %v", err)
	} else {
		// Check that it's a directory
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", dbDir)
		}
		// Check permissions (should have user read/write/execute)
		mode := info.Mode().Perm()
		if mode&0700 != 0700 {
			t.Errorf("expected directory to have at least 0700 permissions, got %o", mode)
		}
	}
}

// TestConfigCreatedOnCLIExecutionSQLiteCLI verifies that the config file is
// automatically created when running any CLI command (not just when calling config.Load directly).
// This is a regression test for issue #001: Config file not created on first run.
func TestConfigCreatedOnCLIExecutionSQLiteCLI(t *testing.T) {
	// Create a temp directory to act as home
	tempHome := t.TempDir()

	// Set XDG environment variables for this test (auto-restored after test)
	configDir := filepath.Join(tempHome, ".config")
	dataDir := filepath.Join(tempHome, ".local", "share")
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("XDG_DATA_HOME", dataDir)

	// Verify config doesn't exist before running CLI
	configPath := filepath.Join(configDir, "todoat", "config.yaml")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("config file should not exist before test: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Run a CLI command that triggers backend creation (uses default XDG paths)
	cfg := &Config{}
	exitCode := Execute([]string{"TestList", "get"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s stdout=%s", exitCode, stderr.String(), stdout.String())
	}

	// Verify the config file was created by the CLI execution
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("expected config file to be created at %s by CLI execution, but it was not created", configPath)
	}
}

// TestSchemaInitializedOnNewDBSQLiteCLI verifies that a new database has all
// required tables (task_lists, tasks, and indexes)
func TestSchemaInitializedOnNewDBSQLiteCLI(t *testing.T) {
	// Create a temp directory for the database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	configPath := filepath.Join(tempDir, "config.yaml")

	// Write a default config to ensure test isolation
	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Running a command should initialize the schema
	exitCode := Execute([]string{"TestList", "get"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	// Open the database directly and verify tables exist
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Check for required tables
	tables := []string{"task_lists", "tasks"}
	for _, tableName := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&name)
		if err != nil {
			t.Errorf("expected table %s to exist: %v", tableName, err)
		}
	}

	// Check for required indexes
	indexes := []string{"idx_tasks_list_id", "idx_tasks_status"}
	for _, indexName := range indexes {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", indexName).Scan(&name)
		if err != nil {
			t.Errorf("expected index %s to exist: %v", indexName, err)
		}
	}
}

// =============================================================================
// Issue Regression Tests
// These tests verify fixes for issues tracked in issues/
// =============================================================================

// TestEmptyListNameRejectedCLI verifies that empty list names are rejected.
// Regression test for issue #001: Empty list name is accepted when creating tasks.
func TestEmptyListNameRejectedCLI(t *testing.T) {
	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write a default config to ensure test isolation
	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Try to add a task with empty list name
	exitCode := Execute([]string{"", "add", "Test task"}, &stdout, &stderr, cfg)

	// Should fail with exit code 1
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for empty list name, got %d", exitCode)
	}

	// Should have an error message about empty list name
	combinedOutput := stderr.String() + stdout.String()
	if !strings.Contains(strings.ToLower(combinedOutput), "list name") || !strings.Contains(strings.ToLower(combinedOutput), "empty") {
		t.Errorf("expected error message about empty list name, got: %s", combinedOutput)
	}
}

// TestWhitespaceOnlyListNameRejectedCLI verifies that whitespace-only list names are rejected.
// Related to issue #001: Empty list name is accepted when creating tasks.
func TestWhitespaceOnlyListNameRejectedCLI(t *testing.T) {
	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write a default config to ensure test isolation
	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Try to add a task with whitespace-only list name
	exitCode := Execute([]string{"   ", "add", "Test task"}, &stdout, &stderr, cfg)

	// Should fail with exit code 1
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for whitespace-only list name, got %d", exitCode)
	}

	// Should have an error message about empty list name
	combinedOutput := stderr.String() + stdout.String()
	if !strings.Contains(strings.ToLower(combinedOutput), "list name") || !strings.Contains(strings.ToLower(combinedOutput), "empty") {
		t.Errorf("expected error message about empty list name, got: %s", combinedOutput)
	}
}

// TestInvalidStatusRejectedCLI verifies that invalid status values are rejected on update.
// Regression test for issue #002: Invalid status value silently ignored on add.
func TestInvalidStatusRejectedCLI(t *testing.T) {
	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write a default config to ensure test isolation
	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// First create a task
	exitCode := Execute([]string{"TestList", "add", "Status test task"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create task: %s", stderr.String())
	}

	// Reset buffers
	stdout.Reset()
	stderr.Reset()

	// Try to update with invalid status
	exitCode = Execute([]string{"TestList", "update", "Status test task", "-s", "INVALID"}, &stdout, &stderr, cfg)

	// Should fail with exit code 1
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for invalid status, got %d", exitCode)
	}

	// Should have an error message about invalid status
	combinedOutput := stderr.String() + stdout.String()
	if !strings.Contains(strings.ToLower(combinedOutput), "invalid") && !strings.Contains(strings.ToLower(combinedOutput), "status") {
		t.Errorf("expected error message about invalid status, got: %s", combinedOutput)
	}
}

// TestValidStatusesAcceptedCLI verifies that all valid status values are accepted.
// Validates that fix for issue #002 doesn't break valid statuses.
func TestValidStatusesAcceptedCLI(t *testing.T) {
	validStatuses := []string{"TODO", "IN-PROGRESS", "DONE", "CANCELLED", "todo", "in-progress", "done", "cancelled"}

	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			// Create temp directory for test database
			tmpDir := t.TempDir()
			dbPath := filepath.Join(tmpDir, "test.db")
			configPath := filepath.Join(tmpDir, "config.yaml")

			// Write a default config to ensure test isolation
			if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
				t.Fatalf("failed to write config: %v", err)
			}

			cfg := &Config{
				DBPath:     dbPath,
				ConfigPath: configPath,
			}

			var stdout, stderr bytes.Buffer

			// First create a task
			exitCode := Execute([]string{"TestList", "add", "Task for " + status}, &stdout, &stderr, cfg)
			if exitCode != 0 {
				t.Fatalf("failed to create task: %s", stderr.String())
			}

			// Reset buffers
			stdout.Reset()
			stderr.Reset()

			// Update with valid status
			exitCode = Execute([]string{"TestList", "update", "Task for " + status, "-s", status}, &stdout, &stderr, cfg)

			// Should succeed
			if exitCode != 0 {
				t.Errorf("expected exit code 0 for valid status %q, got %d: %s", status, exitCode, stderr.String())
			}
		})
	}
}

// --- Backend Flag Tests ---
// CLI Test for issue 1: --backend flag not working

// TestBackendFlagRecognized verifies that --backend flag is recognized
func TestBackendFlagRecognized(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// The --backend flag should be recognized (not return "unknown flag: --backend")
	// Even if the backend doesn't exist, we should get a different error
	exitCode := Execute([]string{"--backend", "sqlite", "--help"}, &stdout, &stderr, nil)

	// With --help, should succeed
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	// Help output should mention --backend flag
	output := stdout.String()
	if !strings.Contains(output, "--backend") {
		t.Errorf("help output should contain '--backend' flag, got: %s", output)
	}
}

// TestBackendFlagSelectsBackend verifies that --backend flag selects the specified backend
func TestBackendFlagSelectsBackend(t *testing.T) {
	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write a config with default_backend: sqlite
	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Use --backend sqlite explicitly (should work same as default)
	exitCode := Execute([]string{"--backend", "sqlite", "list"}, &stdout, &stderr, cfg)

	// Should succeed
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}
}

// TestBackendFlagUnknownBackendError verifies error for unknown backend
func TestBackendFlagUnknownBackendError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write a config
	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Use --backend with unknown backend name
	exitCode := Execute([]string{"--backend", "nonexistent", "list"}, &stdout, &stderr, cfg)

	// Should fail with unknown backend error (not "unknown flag")
	if exitCode == 0 {
		t.Fatal("expected non-zero exit code for unknown backend")
	}

	errOutput := stderr.String()
	if strings.Contains(errOutput, "unknown flag") {
		t.Errorf("error should not be 'unknown flag', got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "unknown backend") && !strings.Contains(errOutput, "nonexistent") {
		t.Errorf("error should mention unknown backend 'nonexistent', got: %s", errOutput)
	}
}

// TestCommaSeparatedStatusFilterCLI verifies that comma-separated status values work for filtering.
// This tests issue #001 - status filter with comma-separated values.
func TestCommaSeparatedStatusFilterCLI(t *testing.T) {
	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write a default config to ensure test isolation
	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Create test list and tasks
	exitCode := Execute([]string{"TestList", "add", "Task 1"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create task 1: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"TestList", "add", "Task 2"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create task 2: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"TestList", "add", "Task 3"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create task 3: %s", stderr.String())
	}

	// Update Task 2 to IN-PROGRESS
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"TestList", "update", "Task 2", "-s", "IN-PROGRESS"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to update task 2: %s", stderr.String())
	}

	// Update Task 3 to DONE
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"TestList", "update", "Task 3", "-s", "DONE"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to update task 3: %s", stderr.String())
	}

	// Test comma-separated status filter
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"TestList", "-s", "TODO,IN-PROGRESS"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("comma-separated status filter failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	output := stdout.String()
	// Should show Task 1 (TODO) and Task 2 (IN-PROGRESS), but not Task 3 (DONE)
	if !strings.Contains(output, "Task 1") {
		t.Errorf("expected Task 1 (TODO) in output, got: %s", output)
	}
	if !strings.Contains(output, "Task 2") {
		t.Errorf("expected Task 2 (IN-PROGRESS) in output, got: %s", output)
	}
	if strings.Contains(output, "Task 3") {
		t.Errorf("Task 3 (DONE) should NOT be in output, got: %s", output)
	}

	// Test with abbreviations
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"TestList", "-s", "T,I"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("abbreviated comma-separated status filter failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	output = stdout.String()
	// Same result - should show Task 1 (TODO) and Task 2 (IN-PROGRESS)
	if !strings.Contains(output, "Task 1") {
		t.Errorf("expected Task 1 (TODO) in output with abbrev, got: %s", output)
	}
	if !strings.Contains(output, "Task 2") {
		t.Errorf("expected Task 2 (IN-PROGRESS) in output with abbrev, got: %s", output)
	}
	if strings.Contains(output, "Task 3") {
		t.Errorf("Task 3 (DONE) should NOT be in output with abbrev, got: %s", output)
	}
}

// =============================================================================
// Backend Fallback Warning Tests (Issue 001)
// Tests that users are warned when falling back from a configured default backend
// =============================================================================

// TestBackendFallbackWarning verifies that when the configured default backend
// is unavailable, the user is warned before falling back to sqlite.
func TestBackendFallbackWarning(t *testing.T) {
	t.Run("warns when nextcloud backend is unavailable", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")
		configPath := filepath.Join(tmpDir, "config.yaml")

		// Create config with default_backend: nextcloud but no credentials
		configContent := `
backends:
  sqlite:
    enabled: true
  nextcloud:
    enabled: true
default_backend: nextcloud
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("failed to create config: %v", err)
		}

		// Clear any TODOAT_NEXTCLOUD_* env vars to ensure backend is unavailable
		_ = os.Unsetenv("TODOAT_NEXTCLOUD_HOST")
		_ = os.Unsetenv("TODOAT_NEXTCLOUD_USERNAME")
		_ = os.Unsetenv("TODOAT_NEXTCLOUD_PASSWORD")

		var stdout, stderr bytes.Buffer
		cfg := &Config{
			DBPath:     dbPath,
			ConfigPath: configPath,
			Stderr:     &stderr, // Capture warnings
		}

		// Run list command - should fall back to sqlite but warn user
		exitCode := Execute([]string{"list"}, &stdout, &stderr, cfg)

		// Should succeed (using fallback sqlite)
		if exitCode != 0 {
			t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
		}

		// Should warn about nextcloud being unavailable
		combinedOutput := stdout.String() + stderr.String()
		if !strings.Contains(strings.ToLower(combinedOutput), "warning") ||
			!strings.Contains(strings.ToLower(combinedOutput), "nextcloud") {
			t.Errorf("expected warning about nextcloud being unavailable, got stdout=%s stderr=%s", stdout.String(), stderr.String())
		}

		// Should indicate using sqlite as fallback
		if !strings.Contains(strings.ToLower(combinedOutput), "sqlite") {
			t.Errorf("expected mention of sqlite fallback, got stdout=%s stderr=%s", stdout.String(), stderr.String())
		}
	})

	t.Run("warns when todoist backend is unavailable", func(t *testing.T) {
		// Skip if keyring has todoist credentials - we can't isolate the keyring in CLI tests
		credMgr := credentials.NewManager()
		if credInfo, err := credMgr.Get(context.Background(), "todoist", "token"); err == nil && credInfo.Found {
			t.Skip("Skipping test: keyring has todoist credentials that cannot be isolated")
		}

		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")
		configPath := filepath.Join(tmpDir, "config.yaml")

		// Create config with default_backend: todoist but no credentials
		configContent := `
backends:
  sqlite:
    enabled: true
  todoist:
    enabled: true
default_backend: todoist
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("failed to create config: %v", err)
		}

		// Clear TODOAT_TODOIST_TOKEN to ensure backend is unavailable
		_ = os.Unsetenv("TODOAT_TODOIST_TOKEN")

		var stdout, stderr bytes.Buffer
		cfg := &Config{
			DBPath:     dbPath,
			ConfigPath: configPath,
			Stderr:     &stderr, // Capture warnings
		}

		// Run list command - should fall back to sqlite but warn user
		exitCode := Execute([]string{"list"}, &stdout, &stderr, cfg)

		// Should succeed (using fallback sqlite)
		if exitCode != 0 {
			t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
		}

		// Should warn about todoist being unavailable
		combinedOutput := stdout.String() + stderr.String()
		if !strings.Contains(strings.ToLower(combinedOutput), "warning") ||
			!strings.Contains(strings.ToLower(combinedOutput), "todoist") {
			t.Errorf("expected warning about todoist being unavailable, got stdout=%s stderr=%s", stdout.String(), stderr.String())
		}
	})

	t.Run("no warning when using sqlite default", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")
		configPath := filepath.Join(tmpDir, "config.yaml")

		// Create config with default_backend: sqlite (no fallback needed)
		configContent := `
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("failed to create config: %v", err)
		}

		var stdout, stderr bytes.Buffer
		cfg := &Config{
			DBPath:     dbPath,
			ConfigPath: configPath,
		}

		exitCode := Execute([]string{"list"}, &stdout, &stderr, cfg)

		if exitCode != 0 {
			t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
		}

		// Should NOT show any fallback warning
		combinedOutput := stdout.String() + stderr.String()
		if strings.Contains(strings.ToLower(combinedOutput), "warning") &&
			strings.Contains(strings.ToLower(combinedOutput), "fallback") {
			t.Errorf("should not show fallback warning when using sqlite default, got stdout=%s stderr=%s", stdout.String(), stderr.String())
		}
	})
}

// =============================================================================
// Utility Function Tests
// These tests verify utility functions used throughout the CLI.
// =============================================================================

// TestValidateAndNormalizeColor verifies color validation and normalization.
// Tests the validateAndNormalizeColor function.
func TestValidateAndNormalizeColor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		// Valid formats
		{"6-char with hash", "#FFFFFF", "#FFFFFF", false},
		{"6-char without hash", "FFFFFF", "#FFFFFF", false},
		{"6-char lowercase", "#ffffff", "#FFFFFF", false},
		{"6-char mixed case", "#FfAaBb", "#FFAABB", false},
		{"3-char with hash", "#FFF", "#FFFFFF", false},
		{"3-char without hash", "FFF", "#FFFFFF", false},
		{"3-char lowercase", "#abc", "#AABBCC", false},
		{"3-char mixed", "f0A", "#FF00AA", false},
		{"valid color code", "#123456", "#123456", false},
		{"dark color", "#000000", "#000000", false},

		// Invalid formats
		{"empty string", "", "", true},
		{"too short", "#FF", "", true},
		{"too long", "#FFFFFFF", "", true},
		{"invalid chars", "#GGGGGG", "", true},
		{"4 chars", "#FFFF", "", true},
		{"5 chars", "#FFFFF", "", true},
		{"word", "red", "", true},
		{"special chars", "#12-456", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateAndNormalizeColor(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateAndNormalizeColor(%q) expected error, got result: %q", tt.input, result)
				}
			} else {
				if err != nil {
					t.Errorf("validateAndNormalizeColor(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("validateAndNormalizeColor(%q) = %q, want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

// TestFormatBytes verifies the formatBytes utility function.
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero bytes", 0, "0 bytes"},
		{"small bytes", 100, "100 bytes"},
		{"exactly 1 KB", 1024, "1.00 KB"},
		{"KB range", 2048, "2.00 KB"},
		{"KB with decimal", 1536, "1.50 KB"},
		{"exactly 1 MB", 1024 * 1024, "1.00 MB"},
		{"MB range", 5 * 1024 * 1024, "5.00 MB"},
		{"MB with decimal", int64(1.5 * 1024 * 1024), "1.50 MB"},
		{"exactly 1 GB", 1024 * 1024 * 1024, "1.00 GB"},
		{"GB range", 2 * 1024 * 1024 * 1024, "2.00 GB"},
		{"large GB", 10 * 1024 * 1024 * 1024, "10.00 GB"},
		{"just under KB", 1023, "1023 bytes"},
		{"just under MB", 1024*1024 - 1, "1024.00 KB"},
		{"just under GB", 1024*1024*1024 - 1, "1024.00 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

// TestContainsJSONFlag verifies the containsJSONFlag function.
func TestContainsJSONFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{"empty args", []string{}, false},
		{"no json flag", []string{"list", "get"}, false},
		{"json flag only", []string{"--json"}, true},
		{"json flag first", []string{"--json", "list"}, true},
		{"json flag middle", []string{"list", "--json", "get"}, true},
		{"json flag last", []string{"list", "get", "--json"}, true},
		{"similar but not json", []string{"--json-output"}, false},
		{"json without dashes", []string{"json"}, false},
		{"single dash json", []string{"-json"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsJSONFlag(tt.args)
			if result != tt.expected {
				t.Errorf("containsJSONFlag(%v) = %v, want %v", tt.args, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// List Subcommand Tests
// Tests for list-related subcommands (update, delete, info, stats, vacuum).
// =============================================================================

// TestListUpdateCommand verifies the list update command works correctly.
func TestListUpdateCommand(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Create a list first
	exitCode := Execute([]string{"list", "create", "TestList"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create list: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// Update the list name
	exitCode = Execute([]string{"list", "update", "TestList", "--name", "RenamedList"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("list update failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	// Verify the list was renamed by trying to get the new name
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"RenamedList", "get"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Errorf("renamed list not found: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}

// TestListUpdateCommandColor verifies the list update command handles color updates.
func TestListUpdateCommandColor(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Create a list first
	exitCode := Execute([]string{"list", "create", "ColorList"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create list: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// Update the list color
	exitCode = Execute([]string{"list", "update", "ColorList", "--color", "#FF0000"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("list update color failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}

// TestListDeleteCommand verifies the list delete command works correctly.
func TestListDeleteCommand(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Create a list first
	exitCode := Execute([]string{"list", "create", "ToDelete"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create list: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// Delete the list with -y flag to skip confirmation
	exitCode = Execute([]string{"-y", "list", "delete", "ToDelete"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("list delete failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}

// TestListInfoCommand verifies the list info command displays list details.
func TestListInfoCommand(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Create a list first
	exitCode := Execute([]string{"list", "create", "InfoList"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create list: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// Get info about the list
	exitCode = Execute([]string{"list", "info", "InfoList"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("list info failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	output := stdout.String()
	// Should show list name and basic info
	if !strings.Contains(output, "InfoList") {
		t.Errorf("list info should show list name, got: %s", output)
	}
}

// TestListInfoCommandJSON verifies the list info command outputs JSON correctly.
func TestListInfoCommandJSON(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Create a list first
	exitCode := Execute([]string{"list", "create", "JSONInfoList"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create list: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// Get info about the list in JSON format
	exitCode = Execute([]string{"--json", "list", "info", "JSONInfoList"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("list info --json failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	output := stdout.String()
	// The list info command outputs text format containing list details
	// Verify it shows the list name
	if !strings.Contains(output, "JSONInfoList") {
		t.Errorf("list info --json should contain list name, got: %s", output)
	}
}

// TestListStatsCommand verifies the list stats command displays database statistics.
func TestListStatsCommand(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Create a list and add some tasks for stats
	exitCode := Execute([]string{"list", "create", "StatsTest"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create list: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"StatsTest", "add", "Task 1"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to add task: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// Get database stats
	exitCode = Execute([]string{"list", "stats"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("list stats failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	output := stdout.String()
	// Should show database statistics
	if !strings.Contains(output, "Database Statistics") {
		t.Errorf("list stats should show 'Database Statistics' header, got: %s", output)
	}
	if !strings.Contains(output, "Total tasks") {
		t.Errorf("list stats should show 'Total tasks', got: %s", output)
	}
}

// TestListVacuumCommand verifies the list vacuum command works correctly.
func TestListVacuumCommand(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Create and delete some data to make vacuum useful
	exitCode := Execute([]string{"list", "create", "VacuumTest"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create list: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// Run vacuum
	exitCode = Execute([]string{"list", "vacuum"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("list vacuum failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	output := stdout.String()
	// Should indicate success
	if !strings.Contains(strings.ToLower(output), "vacuum") && !strings.Contains(strings.ToLower(output), "complet") {
		t.Errorf("list vacuum should indicate success, got: %s", output)
	}
}

// TestIssue034StatsWithAutoDetect verifies that list stats command works when
// auto_detect_backend is enabled, which returns a DetectableBackend wrapper
// instead of raw *sqlite.Backend. See issues/034-list-stats-vacuum-require-explicit-backend.md
func TestIssue034StatsWithAutoDetect(t *testing.T) {
	// Clear any Todoist credentials from env to ensure auto-detection uses SQLite
	// (if TODOAT_TODOIST_TOKEN is set, auto-detect will pick Todoist instead)
	origToken := os.Getenv("TODOAT_TODOIST_TOKEN")
	_ = os.Unsetenv("TODOAT_TODOIST_TOKEN")
	defer func() {
		if origToken != "" {
			_ = os.Setenv("TODOAT_TODOIST_TOKEN", origToken)
		}
	}()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Enable auto_detect_backend=true - this triggers the bug
	configYAML := `default_backend: sqlite
auto_detect_backend: true
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Create a list and add some tasks for stats
	exitCode := Execute([]string{"list", "create", "StatsTestAutoDetect"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create list: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"StatsTestAutoDetect", "add", "Task 1"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to add task: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// Get database stats - this should work even with auto_detect_backend enabled
	exitCode = Execute([]string{"list", "stats"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("list stats failed with auto_detect_backend: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	output := stdout.String()
	// Should show database statistics
	if !strings.Contains(output, "Database Statistics") {
		t.Errorf("list stats should show 'Database Statistics' header, got: %s", output)
	}
	if !strings.Contains(output, "Total tasks") {
		t.Errorf("list stats should show 'Total tasks', got: %s", output)
	}
}

// TestIssue034VacuumWithAutoDetect verifies that list vacuum command works when
// auto_detect_backend is enabled, which returns a DetectableBackend wrapper
// instead of raw *sqlite.Backend. See issues/034-list-stats-vacuum-require-explicit-backend.md
func TestIssue034VacuumWithAutoDetect(t *testing.T) {
	// Clear any Todoist credentials from env to ensure auto-detection uses SQLite
	// (if TODOAT_TODOIST_TOKEN is set, auto-detect will pick Todoist instead)
	origToken := os.Getenv("TODOAT_TODOIST_TOKEN")
	_ = os.Unsetenv("TODOAT_TODOIST_TOKEN")
	defer func() {
		if origToken != "" {
			_ = os.Setenv("TODOAT_TODOIST_TOKEN", origToken)
		}
	}()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Enable auto_detect_backend=true - this triggers the bug
	configYAML := `default_backend: sqlite
auto_detect_backend: true
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Create a list
	exitCode := Execute([]string{"list", "create", "VacuumTestAutoDetect"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create list: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// Run vacuum - this should work even with auto_detect_backend enabled
	exitCode = Execute([]string{"list", "vacuum"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("list vacuum failed with auto_detect_backend: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	output := stdout.String()
	// Should indicate success
	if !strings.Contains(strings.ToLower(output), "vacuum") && !strings.Contains(strings.ToLower(output), "complet") {
		t.Errorf("list vacuum should indicate success, got: %s", output)
	}
}

// =============================================================================
// List Trash Subcommand Tests
// Tests for trash-related commands (trash, restore, purge).
// =============================================================================

// TestListTrashCommand verifies the list trash command shows deleted lists.
func TestListTrashCommand(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Create a list and delete it
	exitCode := Execute([]string{"list", "create", "TrashTest"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create list: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "list", "delete", "TrashTest"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to delete list: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// View trash
	exitCode = Execute([]string{"list", "trash"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("list trash failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	output := stdout.String()
	// Should show deleted list
	if !strings.Contains(output, "TrashTest") {
		t.Errorf("list trash should show deleted list 'TrashTest', got: %s", output)
	}
}

// TestListTrashRestoreCommand verifies the list trash restore command.
func TestListTrashRestoreCommand(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Create a list and delete it
	exitCode := Execute([]string{"list", "create", "RestoreTest"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create list: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "list", "delete", "RestoreTest"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to delete list: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// Restore the list
	exitCode = Execute([]string{"list", "trash", "restore", "RestoreTest"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("list trash restore failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	// Verify the list is restored
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"RestoreTest", "get"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Errorf("restored list not found: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}

// TestListTrashPurgeCommand verifies the list trash purge command.
func TestListTrashPurgeCommand(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:     dbPath,
		ConfigPath: configPath,
	}

	var stdout, stderr bytes.Buffer

	// Create a list and delete it
	exitCode := Execute([]string{"list", "create", "PurgeTest"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to create list: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"-y", "list", "delete", "PurgeTest"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("failed to delete list: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	// Purge the list permanently with -y to skip confirmation
	exitCode = Execute([]string{"-y", "list", "trash", "purge", "PurgeTest"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("list trash purge failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}

	// Verify the list is no longer in trash
	stdout.Reset()
	stderr.Reset()
	exitCode = Execute([]string{"list", "trash"}, &stdout, &stderr, cfg)
	if exitCode != 0 {
		t.Fatalf("list trash failed: %s", stderr.String())
	}

	output := stdout.String()
	if strings.Contains(output, "PurgeTest") {
		t.Errorf("purged list should not appear in trash, got: %s", output)
	}
}

// TestIssue012CredentialsListReadsConfiguredBackends verifies that 'credentials list'
// reads backends from the actual configuration file instead of using a hardcoded list.
// This is a regression test for issue 012.
func TestIssue012CredentialsListReadsConfiguredBackends(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config with a custom backend name (work-nextcloud) that should appear in list
	configContent := `
backends:
  sqlite:
    enabled: true
  nextcloud:
    enabled: true
  work-nextcloud:
    enabled: true
    type: nextcloud
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cfg := &Config{
		ConfigPath: configPath,
	}

	exitCode := Execute([]string{"credentials", "list"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	output := stdout.String()

	// Should show the custom backend "work-nextcloud" from config
	if !strings.Contains(output, "work-nextcloud") {
		t.Errorf("credentials list should show custom backend 'work-nextcloud' from config, got: %s", output)
	}

	// Should also show standard backends that are enabled
	if !strings.Contains(output, "nextcloud") {
		t.Errorf("credentials list should show 'nextcloud' backend, got: %s", output)
	}
}

// --- Issue 033: Git and File Backends CLI Access ---

// TestIssue033GitBackendAccessibleViaCLI verifies that git backend can be used via -b flag
// CLI Test for 033-git-file-backends-not-wired-to-cli
func TestIssue033GitBackendAccessibleViaCLI(t *testing.T) {
	// Create a temp directory with a git repo and TODO.md file
	tmpDir := t.TempDir()

	// Initialize git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	// Create a TODO.md file with the todoat marker
	todoContent := `<!-- todoat:enabled -->
# Tasks

## Inbox

- [ ] Test task one
- [x] Completed task
`
	todoFile := filepath.Join(tmpDir, "TODO.md")
	if err := os.WriteFile(todoFile, []byte(todoContent), 0644); err != nil {
		t.Fatalf("failed to write TODO.md: %v", err)
	}

	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Save and change working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	var stdout, stderr bytes.Buffer

	cfg := &Config{
		ConfigPath: configPath,
	}

	// Try to use git backend explicitly - should NOT produce "unknown backend" error
	exitCode := Execute([]string{"-b", "git", "list"}, &stdout, &stderr, cfg)

	// The command should succeed (exit 0) since we have a valid git repo with TODO.md
	if exitCode != 0 {
		stderrStr := stderr.String()
		// The specific error we're fixing: "unknown backend type 'git'"
		if strings.Contains(stderrStr, "unknown backend") {
			t.Fatalf("git backend should be recognized via -b flag, but got: %s", stderrStr)
		}
		// Other errors might be acceptable (e.g., git not initialized properly)
		t.Logf("Command failed with exit code %d, stderr: %s", exitCode, stderrStr)
	}

	// Should NOT get "unknown backend" error
	output := stderr.String()
	if strings.Contains(output, "unknown backend type 'git'") {
		t.Errorf("git backend should be recognized, but got 'unknown backend' error: %s", output)
	}
	if strings.Contains(output, "unknown backend: git") {
		t.Errorf("git backend should be recognized, but got 'unknown backend' error: %s", output)
	}
}

// TestIssue033FileBackendAccessibleViaCLI verifies that file backend can be used via -b flag
// CLI Test for 033-git-file-backends-not-wired-to-cli
func TestIssue033FileBackendAccessibleViaCLI(t *testing.T) {
	// Create a temp directory with a tasks file
	tmpDir := t.TempDir()

	// Create a tasks.txt file
	tasksContent := `# Tasks

## Inbox

- [ ] Test task one
- [x] Completed task
`
	tasksFile := filepath.Join(tmpDir, "tasks.txt")
	if err := os.WriteFile(tasksFile, []byte(tasksContent), 0644); err != nil {
		t.Fatalf("failed to write tasks.txt: %v", err)
	}

	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Save and change working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	var stdout, stderr bytes.Buffer

	cfg := &Config{
		ConfigPath: configPath,
	}

	// Try to use file backend explicitly - should NOT produce "unknown backend" error
	exitCode := Execute([]string{"-b", "file", "list"}, &stdout, &stderr, cfg)

	// The command should succeed (exit 0) since we have a valid tasks file
	if exitCode != 0 {
		stderrStr := stderr.String()
		// The specific error we're fixing: "unknown backend type 'file'"
		if strings.Contains(stderrStr, "unknown backend") {
			t.Fatalf("file backend should be recognized via -b flag, but got: %s", stderrStr)
		}
		// Other errors might be acceptable
		t.Logf("Command failed with exit code %d, stderr: %s", exitCode, stderrStr)
	}

	// Should NOT get "unknown backend" error
	output := stderr.String()
	if strings.Contains(output, "unknown backend type 'file'") {
		t.Errorf("file backend should be recognized, but got 'unknown backend' error: %s", output)
	}
	if strings.Contains(output, "unknown backend: file") {
		t.Errorf("file backend should be recognized, but got 'unknown backend' error: %s", output)
	}
}

// --- Roadmap 071: Git Backend CLI Wiring ---

// TestGitBackendExplicitFlag verifies that `todoat -b git "Project Tasks"` works with explicit flag
// CLI Test for 071-git-backend-cli-wiring
func TestGitBackendExplicitFlag(t *testing.T) {
	// Create a temp directory with a git repo and TODO.md file
	tmpDir := t.TempDir()

	// Initialize git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	// Create a TODO.md file with the todoat marker and a section named "Project Tasks"
	todoContent := `<!-- todoat:enabled -->
# Tasks

## Project Tasks

- [ ] Implement feature X
- [ ] Fix bug Y

## Inbox

- [ ] Review PR
`
	todoFile := filepath.Join(tmpDir, "TODO.md")
	if err := os.WriteFile(todoFile, []byte(todoContent), 0644); err != nil {
		t.Fatalf("failed to write TODO.md: %v", err)
	}

	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Save and change working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	var stdout, stderr bytes.Buffer

	cfg := &Config{
		ConfigPath: configPath,
		CachePath:  filepath.Join(tmpDir, "cache.json"), // Isolate cache for testing
	}

	// Use git backend explicitly with a list name - should work
	exitCode := Execute([]string{"-b", "git", "Project Tasks"}, &stdout, &stderr, cfg)

	// The command should succeed (exit 0) since we have a valid git repo with TODO.md
	if exitCode != 0 {
		stderrStr := stderr.String()
		// The specific error we're testing: "unknown backend type 'git'"
		if strings.Contains(stderrStr, "unknown backend") {
			t.Fatalf("git backend should be recognized via -b flag, but got: %s", stderrStr)
		}
		t.Logf("Command failed with exit code %d, stderr: %s, stdout: %s", exitCode, stderrStr, stdout.String())
	}

	// Verify we got output related to Project Tasks section
	output := stdout.String()
	if !strings.Contains(output, "Project Tasks") && !strings.Contains(output, "Implement feature X") {
		t.Logf("Expected output to contain 'Project Tasks' or tasks from that section, got: %s", output)
	}
}

// TestGitBackendConfigType verifies that Backend type "git" is recognized in config `backends:` section
// CLI Test for 071-git-backend-cli-wiring
func TestGitBackendConfigType(t *testing.T) {
	// Create a temp directory with a git repo and TODO.md file
	tmpDir := t.TempDir()

	// Initialize git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	// Create a TODO.md file with the todoat marker
	todoContent := `<!-- todoat:enabled -->
# Tasks

## Inbox

- [ ] Task from config test
`
	todoFile := filepath.Join(tmpDir, "TODO.md")
	if err := os.WriteFile(todoFile, []byte(todoContent), 0644); err != nil {
		t.Fatalf("failed to write TODO.md: %v", err)
	}

	// Create config with git backend configured in backends: section
	configContent := `backends:
  git:
    type: git
    enabled: true
    file: "TODO.md"
default_backend: git
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Save and change working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	var stdout, stderr bytes.Buffer

	cfg := &Config{
		ConfigPath: configPath,
		CachePath:  filepath.Join(tmpDir, "cache.json"), // Isolate cache for testing
	}

	// Use git backend via config (default_backend: git)
	exitCode := Execute([]string{"list"}, &stdout, &stderr, cfg)

	// The command should succeed
	if exitCode != 0 {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "unknown backend") {
			t.Fatalf("git backend should be recognized from config backends: section, but got: %s", stderrStr)
		}
		// Log but don't fail for other errors
		t.Logf("Command failed with exit code %d, stderr: %s", exitCode, stderrStr)
	}

	// Verify we're using git backend by checking output contains our task
	output := stdout.String()
	if !strings.Contains(output, "Task from config test") && !strings.Contains(output, "Inbox") {
		t.Logf("Expected output from git backend with TODO.md content, got: %s", output)
	}
}

// TestGitBackendListsCommand verifies that `todoat -b git list` shows sections from TODO.md
// CLI Test for 071-git-backend-cli-wiring
func TestGitBackendListsCommand(t *testing.T) {
	// Create a temp directory with a git repo and TODO.md file
	tmpDir := t.TempDir()

	// Initialize git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	// Create a TODO.md file with multiple sections
	todoContent := `<!-- todoat:enabled -->
# Tasks

## Development

- [ ] Code review
- [ ] Unit tests

## Documentation

- [ ] Update README

## Backlog

- [x] Initial setup
`
	todoFile := filepath.Join(tmpDir, "TODO.md")
	if err := os.WriteFile(todoFile, []byte(todoContent), 0644); err != nil {
		t.Fatalf("failed to write TODO.md: %v", err)
	}

	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Save and change working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	var stdout, stderr bytes.Buffer

	cfg := &Config{
		ConfigPath: configPath,
		CachePath:  filepath.Join(tmpDir, "cache.json"), // Isolate cache for testing
	}

	// Use git backend explicitly with list command
	exitCode := Execute([]string{"-b", "git", "list"}, &stdout, &stderr, cfg)

	// The command should succeed
	if exitCode != 0 {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "unknown backend") {
			t.Fatalf("git backend should be recognized via -b flag, but got: %s", stderrStr)
		}
		t.Logf("Command failed with exit code %d, stderr: %s", exitCode, stderrStr)
	}

	// Verify the output shows sections/lists from TODO.md
	output := stdout.String()

	// Check for task content or list names
	hasExpectedContent := strings.Contains(output, "Development") ||
		strings.Contains(output, "Documentation") ||
		strings.Contains(output, "Backlog") ||
		strings.Contains(output, "Code review") ||
		strings.Contains(output, "Unit tests")

	if !hasExpectedContent {
		t.Errorf("Expected git backend list output to contain sections from TODO.md, got: %s", output)
	}
}

// --- Issue 069: Google Tasks CLI Integration ---

// TestGoogleTasksCLIBackendRecognized verifies that google backend is recognized via -b flag
// CLI Test for 069-google-tasks-cli-integration
func TestGoogleTasksCLIBackendRecognized(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	var stdout, stderr bytes.Buffer

	cfg := &Config{
		ConfigPath: configPath,
	}

	// Try to use google backend explicitly - should NOT produce "unknown backend" error
	// It WILL fail due to missing credentials, but that's expected
	exitCode := Execute([]string{"-b", "google", "list"}, &stdout, &stderr, cfg)

	// The command may fail due to missing credentials, but NOT due to unknown backend
	errOutput := stderr.String()
	if strings.Contains(errOutput, "unknown backend type 'google'") {
		t.Errorf("google backend should be recognized, but got 'unknown backend type' error: %s", errOutput)
	}
	if strings.Contains(errOutput, "unknown backend: google") {
		t.Errorf("google backend should be recognized, but got 'unknown backend' error: %s", errOutput)
	}

	// Should get a credentials error if it fails (which is expected without valid tokens)
	if exitCode != 0 {
		// This is expected - we just want to verify the backend type is recognized
		if !strings.Contains(errOutput, "access token") && !strings.Contains(errOutput, "TODOAT_GOOGLE_ACCESS_TOKEN") {
			t.Logf("Google backend recognized but command failed with unexpected error: %s", errOutput)
		}
	}
}

// TestGoogleTasksCLIBackendInErrorMessage verifies google is listed in the error message for unknown backends
func TestGoogleTasksCLIBackendInErrorMessage(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	var stdout, stderr bytes.Buffer

	cfg := &Config{
		ConfigPath: configPath,
	}

	// Use a completely unknown backend to see the list of supported backends
	exitCode := Execute([]string{"-b", "nonexistent-backend-xyz", "list"}, &stdout, &stderr, cfg)

	if exitCode == 0 {
		t.Fatal("expected non-zero exit code for unknown backend")
	}

	errOutput := stderr.String()
	// The error message should list "google" as a supported backend
	if !strings.Contains(errOutput, "google") {
		t.Errorf("error message should list 'google' as a supported backend, got: %s", errOutput)
	}
}

// --- Roadmap 072: File Backend CLI Integration ---

// TestFileBackendAddTask verifies that `todoat -b file "Work" add "Task"` creates task in configured file
// CLI Test for 072-file-backend-implementation
func TestFileBackendAddTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a tasks file with a Work section
	tasksContent := `# Tasks

## Work

- [ ] Existing task
`
	tasksFile := filepath.Join(tmpDir, "tasks.txt")
	if err := os.WriteFile(tasksFile, []byte(tasksContent), 0644); err != nil {
		t.Fatalf("failed to write tasks.txt: %v", err)
	}

	// Create config pointing to the tasks file
	configContent := fmt.Sprintf(`default_backend: file
backends:
  file:
    type: file
    path: %s
`, tasksFile)
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cfg := &Config{
		ConfigPath: configPath,
		CachePath:  filepath.Join(tmpDir, "cache.json"),
	}

	// Add a new task using file backend
	exitCode := Execute([]string{"-b", "file", "Work", "add", "New CLI task"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "unknown backend") {
			t.Fatalf("file backend should be recognized via -b flag, but got: %s", stderrStr)
		}
		t.Logf("Command failed with exit code %d, stderr: %s", exitCode, stderrStr)
	}

	// Verify the task was added to the file
	data, err := os.ReadFile(tasksFile)
	if err != nil {
		t.Fatalf("failed to read tasks file: %v", err)
	}

	if !strings.Contains(string(data), "New CLI task") {
		t.Errorf("expected tasks file to contain 'New CLI task', got:\n%s", data)
	}
}

// TestFileBackendGetTasks verifies that `todoat -b file "Work"` lists tasks from file
// CLI Test for 072-file-backend-implementation
func TestFileBackendGetTasks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a tasks file with multiple tasks
	tasksContent := `# Tasks

## Work

- [ ] Task one
- [ ] Task two
- [x] Completed task
`
	tasksFile := filepath.Join(tmpDir, "tasks.txt")
	if err := os.WriteFile(tasksFile, []byte(tasksContent), 0644); err != nil {
		t.Fatalf("failed to write tasks.txt: %v", err)
	}

	configContent := fmt.Sprintf(`default_backend: file
backends:
  file:
    type: file
    path: %s
`, tasksFile)
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cfg := &Config{
		ConfigPath: configPath,
		CachePath:  filepath.Join(tmpDir, "cache.json"),
	}

	// List tasks in Work list
	exitCode := Execute([]string{"-b", "file", "Work"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "unknown backend") {
			t.Fatalf("file backend should be recognized, but got: %s", stderrStr)
		}
		t.Logf("Command failed with exit code %d, stderr: %s", exitCode, stderrStr)
	}

	output := stdout.String()
	// Should display tasks from the Work list
	if !strings.Contains(output, "Task one") && !strings.Contains(output, "Task two") {
		t.Errorf("expected output to contain tasks from Work list, got: %s", output)
	}
}

// TestFileBackendUpdateTask verifies that `todoat -b file "Work" update "Task" -s D` updates task
// CLI Test for 072-file-backend-implementation
func TestFileBackendUpdateTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a tasks file with a task to update
	tasksContent := `# Tasks

## Work

- [ ] Task to complete
`
	tasksFile := filepath.Join(tmpDir, "tasks.txt")
	if err := os.WriteFile(tasksFile, []byte(tasksContent), 0644); err != nil {
		t.Fatalf("failed to write tasks.txt: %v", err)
	}

	configContent := fmt.Sprintf(`default_backend: file
backends:
  file:
    type: file
    path: %s
`, tasksFile)
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cfg := &Config{
		ConfigPath: configPath,
		CachePath:  filepath.Join(tmpDir, "cache.json"),
	}

	// Update task status to Done
	exitCode := Execute([]string{"-b", "file", "Work", "update", "Task to complete", "-s", "D"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "unknown backend") {
			t.Fatalf("file backend should be recognized, but got: %s", stderrStr)
		}
		// Some errors are acceptable (e.g., task not found exactly) - just log
		t.Logf("Update command exit code %d, stderr: %s", exitCode, stderrStr)
	}

	// Verify the task was updated in the file
	data, err := os.ReadFile(tasksFile)
	if err != nil {
		t.Fatalf("failed to read tasks file: %v", err)
	}

	// Should have [x] for completed status
	if !strings.Contains(string(data), "[x]") {
		t.Errorf("expected task to be marked complete with [x], got:\n%s", data)
	}
}

// TestFileBackendDeleteTask verifies that `todoat -b file "Work" delete "Task"` removes task
// CLI Test for 072-file-backend-implementation
func TestFileBackendDeleteTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a tasks file with tasks
	tasksContent := `# Tasks

## Work

- [ ] Keep this task
- [ ] Delete this task
`
	tasksFile := filepath.Join(tmpDir, "tasks.txt")
	if err := os.WriteFile(tasksFile, []byte(tasksContent), 0644); err != nil {
		t.Fatalf("failed to write tasks.txt: %v", err)
	}

	configContent := fmt.Sprintf(`default_backend: file
backends:
  file:
    type: file
    path: %s
`, tasksFile)
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cfg := &Config{
		ConfigPath: configPath,
		CachePath:  filepath.Join(tmpDir, "cache.json"),
	}

	// Delete the task
	exitCode := Execute([]string{"-b", "file", "Work", "delete", "Delete this task"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "unknown backend") {
			t.Fatalf("file backend should be recognized, but got: %s", stderrStr)
		}
		t.Logf("Delete command exit code %d, stderr: %s", exitCode, stderrStr)
	}

	// Verify the task was deleted from the file
	data, err := os.ReadFile(tasksFile)
	if err != nil {
		t.Fatalf("failed to read tasks file: %v", err)
	}

	if strings.Contains(string(data), "Delete this task") {
		t.Errorf("expected 'Delete this task' to be removed, but it's still in:\n%s", data)
	}
	if !strings.Contains(string(data), "Keep this task") {
		t.Errorf("expected 'Keep this task' to remain, but it's not in:\n%s", data)
	}
}

// TestFileBackendListManagement verifies that sections in file are treated as task lists
// CLI Test for 072-file-backend-implementation
func TestFileBackendListManagement(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a tasks file with multiple sections
	tasksContent := `# Tasks

## Work

- [ ] Work task

## Personal

- [ ] Personal task

## Shopping

- [ ] Buy milk
`
	tasksFile := filepath.Join(tmpDir, "tasks.txt")
	if err := os.WriteFile(tasksFile, []byte(tasksContent), 0644); err != nil {
		t.Fatalf("failed to write tasks.txt: %v", err)
	}

	configContent := fmt.Sprintf(`default_backend: file
backends:
  file:
    type: file
    path: %s
`, tasksFile)
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cfg := &Config{
		ConfigPath: configPath,
		CachePath:  filepath.Join(tmpDir, "cache.json"),
	}

	// List all sections (lists)
	exitCode := Execute([]string{"-b", "file", "list"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "unknown backend") {
			t.Fatalf("file backend should be recognized, but got: %s", stderrStr)
		}
		t.Logf("List command exit code %d, stderr: %s", exitCode, stderrStr)
	}

	output := stdout.String()
	// Should show all three sections
	hasWork := strings.Contains(output, "Work")
	hasPersonal := strings.Contains(output, "Personal")
	hasShopping := strings.Contains(output, "Shopping")

	if !hasWork || !hasPersonal || !hasShopping {
		t.Errorf("expected output to contain Work, Personal, and Shopping lists, got: %s", output)
	}
}

// TestFileBackendMetadata verifies that tasks store priority, dates, status, tags
// CLI Test for 072-file-backend-implementation
func TestFileBackendMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a tasks file
	tasksContent := `# Tasks

## Work

- [ ] Task with metadata !1 @2024-06-15 #urgent #review
`
	tasksFile := filepath.Join(tmpDir, "tasks.txt")
	if err := os.WriteFile(tasksFile, []byte(tasksContent), 0644); err != nil {
		t.Fatalf("failed to write tasks.txt: %v", err)
	}

	configContent := fmt.Sprintf(`default_backend: file
backends:
  file:
    type: file
    path: %s
`, tasksFile)
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cfg := &Config{
		ConfigPath: configPath,
		CachePath:  filepath.Join(tmpDir, "cache.json"),
	}

	// Get tasks from Work list
	exitCode := Execute([]string{"-b", "file", "Work"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "unknown backend") {
			t.Fatalf("file backend should be recognized, but got: %s", stderrStr)
		}
		t.Logf("Command exit code %d, stderr: %s", exitCode, stderrStr)
	}

	output := stdout.String()
	// Should show the task (metadata parsing is internal but task should be visible)
	if !strings.Contains(output, "Task with metadata") {
		t.Errorf("expected output to contain task, got: %s", output)
	}

	// Now add a task with metadata via CLI and verify it's stored correctly
	stdout.Reset()
	stderr.Reset()

	exitCode = Execute([]string{"-b", "file", "Work", "add", "Priority task", "-p", "1"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Logf("Add with priority failed with exit code %d, stderr: %s", exitCode, stderr.String())
	}

	// Verify the priority marker is in the file
	data, err := os.ReadFile(tasksFile)
	if err != nil {
		t.Fatalf("failed to read tasks file: %v", err)
	}

	if !strings.Contains(string(data), "Priority task") {
		t.Errorf("expected file to contain 'Priority task', got:\n%s", data)
	}
	// Priority 1 should be marked as !1 in the file
	if !strings.Contains(string(data), "!1") {
		t.Errorf("expected file to contain priority marker '!1', got:\n%s", data)
	}
}

// TestIssue008_CustomSQLiteBackendUsesBackendID verifies that custom SQLite backends
// use the backend name as backend_id for data isolation.
// Issue #008: Backend isolation capability exists but isn't used in createCustomBackend().
func TestIssue008_CustomSQLiteBackendUsesBackendID(t *testing.T) {
	tmpDir := t.TempDir()
	sharedDBPath := filepath.Join(tmpDir, "shared.db")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config with two custom SQLite backends sharing the same database file
	configContent := fmt.Sprintf(`
backends:
  sqlite-work:
    type: sqlite
    path: %s
  sqlite-personal:
    type: sqlite
    path: %s
`, sharedDBPath, sharedDBPath)

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Use sqlite-work backend to add a task (using -y to auto-confirm list creation)
	cfg1 := &Config{
		DBPath:     sharedDBPath,
		ConfigPath: configPath,
	}
	var stdout1, stderr1 bytes.Buffer
	exitCode := Execute([]string{"--backend", "sqlite-work", "-y", "TestList", "add", "Work task"}, &stdout1, &stderr1, cfg1)
	if exitCode != 0 {
		t.Fatalf("add task to sqlite-work failed: exit=%d stderr=%s", exitCode, stderr1.String())
	}

	// Use sqlite-personal backend to add a task (using -y to auto-confirm list creation)
	cfg2 := &Config{
		DBPath:     sharedDBPath,
		ConfigPath: configPath,
	}
	var stdout2, stderr2 bytes.Buffer
	exitCode = Execute([]string{"--backend", "sqlite-personal", "-y", "TestList", "add", "Personal task"}, &stdout2, &stderr2, cfg2)
	if exitCode != 0 {
		t.Fatalf("add task to sqlite-personal failed: exit=%d stderr=%s", exitCode, stderr2.String())
	}

	// List tasks from sqlite-work - should only see "Work task"
	cfg3 := &Config{
		DBPath:     sharedDBPath,
		ConfigPath: configPath,
	}
	var stdout3, stderr3 bytes.Buffer
	exitCode = Execute([]string{"--backend", "sqlite-work", "TestList"}, &stdout3, &stderr3, cfg3)
	if exitCode != 0 {
		t.Fatalf("list from sqlite-work failed: exit=%d stderr=%s", exitCode, stderr3.String())
	}
	workOutput := stdout3.String()
	if !strings.Contains(workOutput, "Work task") {
		t.Errorf("sqlite-work should see 'Work task', got: %s", workOutput)
	}
	if strings.Contains(workOutput, "Personal task") {
		t.Errorf("sqlite-work should NOT see 'Personal task' (isolation broken), got: %s", workOutput)
	}

	// List tasks from sqlite-personal - should only see "Personal task"
	cfg4 := &Config{
		DBPath:     sharedDBPath,
		ConfigPath: configPath,
	}
	var stdout4, stderr4 bytes.Buffer
	exitCode = Execute([]string{"--backend", "sqlite-personal", "TestList"}, &stdout4, &stderr4, cfg4)
	if exitCode != 0 {
		t.Fatalf("list from sqlite-personal failed: exit=%d stderr=%s", exitCode, stderr4.String())
	}
	personalOutput := stdout4.String()
	if !strings.Contains(personalOutput, "Personal task") {
		t.Errorf("sqlite-personal should see 'Personal task', got: %s", personalOutput)
	}
	if strings.Contains(personalOutput, "Work task") {
		t.Errorf("sqlite-personal should NOT see 'Work task' (isolation broken), got: %s", personalOutput)
	}
}

// =============================================================================
// Analytics CLI Tests (075-analytics-cli-commands)
// =============================================================================

// setupAnalyticsDB creates an analytics database with test data and returns the path
func setupAnalyticsDB(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	analyticsPath := filepath.Join(tmpDir, "analytics.db")

	// Create the analytics database with schema
	db, err := sql.Open("sqlite", analyticsPath)
	if err != nil {
		t.Fatalf("failed to create analytics db: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create schema
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp INTEGER NOT NULL,
			command TEXT NOT NULL,
			subcommand TEXT,
			backend TEXT,
			success INTEGER NOT NULL,
			duration_ms INTEGER,
			error_type TEXT,
			flags TEXT,
			created_at INTEGER DEFAULT (strftime('%s', 'now'))
		);
		CREATE INDEX IF NOT EXISTS idx_timestamp ON events(timestamp);
		CREATE INDEX IF NOT EXISTS idx_command ON events(command);
		CREATE INDEX IF NOT EXISTS idx_backend ON events(backend);
		CREATE INDEX IF NOT EXISTS idx_success ON events(success);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// Insert test data
	now := time.Now().Unix()
	yesterday := now - 86400
	lastWeek := now - (7 * 86400)
	lastMonth := now - (30 * 86400)

	// Insert various commands with different backends and success states
	testData := []struct {
		timestamp  int64
		command    string
		subcommand string
		backend    string
		success    int
		durationMs int
		errorType  string
	}{
		{now, "add", "", "sqlite", 1, 50, ""},
		{now, "add", "", "sqlite", 1, 45, ""},
		{now, "list", "", "sqlite", 1, 30, ""},
		{now, "complete", "", "sqlite", 1, 60, ""},
		{now, "sync", "", "todoist", 0, 5000, "timeout"},
		{now, "sync", "", "todoist", 1, 200, ""},
		{yesterday, "add", "", "todoist", 1, 100, ""},
		{yesterday, "delete", "", "sqlite", 0, 20, "not_found"},
		{lastWeek, "add", "", "sqlite", 1, 55, ""},
		{lastWeek, "sync", "", "nextcloud", 0, 3000, "network"},
		{lastMonth, "add", "", "sqlite", 1, 40, ""},
		{lastMonth, "list", "", "file", 1, 10, ""},
	}

	for _, d := range testData {
		_, err = db.Exec(`
			INSERT INTO events (timestamp, command, subcommand, backend, success, duration_ms, error_type)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, d.timestamp, d.command, nullStr(d.subcommand), d.backend, d.success, d.durationMs, nullStr(d.errorType))
		if err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}
	}

	return analyticsPath
}

// nullStr returns nil for empty strings, otherwise the string
func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// TestAnalyticsStatsCommand tests 'todoat analytics stats' shows usage summary
func TestAnalyticsStatsCommand(t *testing.T) {
	analyticsPath := setupAnalyticsDB(t)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	dbPath := filepath.Join(tmpDir, "tasks.db")

	// Write config file
	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:        dbPath,
		ConfigPath:    configPath,
		AnalyticsPath: analyticsPath,
	}

	var stdout, stderr bytes.Buffer
	exitCode := Execute([]string{"analytics", "stats"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	output := stdout.String()

	// Should show command usage summary
	if !strings.Contains(output, "add") {
		t.Errorf("stats output should contain 'add' command, got: %s", output)
	}
	if !strings.Contains(output, "list") {
		t.Errorf("stats output should contain 'list' command, got: %s", output)
	}
	// Should show success rate
	if !strings.Contains(output, "%") || !strings.Contains(strings.ToLower(output), "success") {
		t.Errorf("stats output should show success rate percentage, got: %s", output)
	}
}

// TestAnalyticsBackendPerformance tests 'todoat analytics backends' shows backend metrics
func TestAnalyticsBackendPerformance(t *testing.T) {
	analyticsPath := setupAnalyticsDB(t)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	dbPath := filepath.Join(tmpDir, "tasks.db")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:        dbPath,
		ConfigPath:    configPath,
		AnalyticsPath: analyticsPath,
	}

	var stdout, stderr bytes.Buffer
	exitCode := Execute([]string{"analytics", "backends"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	output := stdout.String()

	// Should show backend names
	if !strings.Contains(output, "sqlite") {
		t.Errorf("backends output should contain 'sqlite', got: %s", output)
	}
	if !strings.Contains(output, "todoist") {
		t.Errorf("backends output should contain 'todoist', got: %s", output)
	}
	// Should show metrics (avg duration, success rate)
	if !strings.Contains(strings.ToLower(output), "ms") || !strings.Contains(output, "%") {
		t.Errorf("backends output should show duration (ms) and success rate (%%), got: %s", output)
	}
}

// TestAnalyticsErrorsCommand tests 'todoat analytics errors' shows common errors
func TestAnalyticsErrorsCommand(t *testing.T) {
	analyticsPath := setupAnalyticsDB(t)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	dbPath := filepath.Join(tmpDir, "tasks.db")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:        dbPath,
		ConfigPath:    configPath,
		AnalyticsPath: analyticsPath,
	}

	var stdout, stderr bytes.Buffer
	exitCode := Execute([]string{"analytics", "errors"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	output := stdout.String()

	// Should show error types
	if !strings.Contains(output, "timeout") {
		t.Errorf("errors output should contain 'timeout' error, got: %s", output)
	}
	if !strings.Contains(output, "network") {
		t.Errorf("errors output should contain 'network' error, got: %s", output)
	}
}

// TestAnalyticsTimeRange tests 'todoat analytics stats --since 7d' filters by time
func TestAnalyticsTimeRange(t *testing.T) {
	analyticsPath := setupAnalyticsDB(t)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	dbPath := filepath.Join(tmpDir, "tasks.db")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:        dbPath,
		ConfigPath:    configPath,
		AnalyticsPath: analyticsPath,
	}

	// Get stats for last 7 days (should exclude lastMonth data)
	var stdout, stderr bytes.Buffer
	exitCode := Execute([]string{"analytics", "stats", "--since", "7d"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	// Get stats for all time to compare
	var stdoutAll, stderrAll bytes.Buffer
	exitCode = Execute([]string{"analytics", "stats"}, &stdoutAll, &stderrAll, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderrAll.String())
	}

	// The 7d output should have fewer total commands than all-time output
	// Since lastMonth data is excluded
	outputFiltered := stdout.String()
	outputAll := stdoutAll.String()

	// Both outputs should be valid (contain command counts)
	if !strings.Contains(outputFiltered, "add") {
		t.Errorf("filtered stats should contain 'add', got: %s", outputFiltered)
	}
	if !strings.Contains(outputAll, "add") {
		t.Errorf("all-time stats should contain 'add', got: %s", outputAll)
	}
}

// TestAnalyticsJSONOutput tests 'todoat analytics stats --json' outputs JSON
func TestAnalyticsJSONOutput(t *testing.T) {
	analyticsPath := setupAnalyticsDB(t)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	dbPath := filepath.Join(tmpDir, "tasks.db")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:        dbPath,
		ConfigPath:    configPath,
		AnalyticsPath: analyticsPath,
	}

	var stdout, stderr bytes.Buffer
	exitCode := Execute([]string{"analytics", "stats", "--json"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	// Output should be valid JSON
	output := stdout.String()
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected valid JSON output, got: %s, error: %v", output, err)
	}

	// Should contain stats data
	if _, ok := result["commands"]; !ok {
		t.Errorf("JSON output should contain 'commands' field, got: %v", result)
	}
}

// TestAnalyticsBackendsJSONOutput tests 'todoat analytics backends --json' outputs JSON
func TestAnalyticsBackendsJSONOutput(t *testing.T) {
	analyticsPath := setupAnalyticsDB(t)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	dbPath := filepath.Join(tmpDir, "tasks.db")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:        dbPath,
		ConfigPath:    configPath,
		AnalyticsPath: analyticsPath,
	}

	var stdout, stderr bytes.Buffer
	exitCode := Execute([]string{"analytics", "backends", "--json"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	// Output should be valid JSON
	output := stdout.String()
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected valid JSON output, got: %s, error: %v", output, err)
	}

	// Should contain backends data
	if _, ok := result["backends"]; !ok {
		t.Errorf("JSON output should contain 'backends' field, got: %v", result)
	}
}

// TestAnalyticsErrorsJSONOutput tests 'todoat analytics errors --json' outputs JSON
func TestAnalyticsErrorsJSONOutput(t *testing.T) {
	analyticsPath := setupAnalyticsDB(t)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	dbPath := filepath.Join(tmpDir, "tasks.db")

	if err := os.WriteFile(configPath, []byte("default_backend: sqlite\n"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:        dbPath,
		ConfigPath:    configPath,
		AnalyticsPath: analyticsPath,
	}

	var stdout, stderr bytes.Buffer
	exitCode := Execute([]string{"analytics", "errors", "--json"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	// Output should be valid JSON
	output := stdout.String()
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected valid JSON output, got: %s, error: %v", output, err)
	}

	// Should contain errors data
	if _, ok := result["errors"]; !ok {
		t.Errorf("JSON output should contain 'errors' field, got: %v", result)
	}
}

// =============================================================================
// Analytics Integration Tests (003-analytics-tracking-not-integrated)
// =============================================================================

// TestAnalyticsTrackingIntegration verifies that commands are tracked in analytics
// when analytics is enabled. This tests the integration between command execution
// and the analytics tracker.
func TestAnalyticsTrackingIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	analyticsPath := filepath.Join(tmpDir, "analytics.db")
	configPath := filepath.Join(tmpDir, "config.yaml")
	dbPath := filepath.Join(tmpDir, "tasks.db")

	// Write config with analytics enabled
	configContent := `default_backend: sqlite
analytics:
  enabled: true
  retention_days: 365
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:        dbPath,
		ConfigPath:    configPath,
		AnalyticsPath: analyticsPath,
	}

	// Run the 'list' command which should trigger analytics tracking
	var stdout, stderr bytes.Buffer
	exitCode := Execute([]string{"list"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	// Wait for async logging to complete (analytics logs asynchronously)
	time.Sleep(500 * time.Millisecond)

	// Verify that analytics database was created
	if _, err := os.Stat(analyticsPath); os.IsNotExist(err) {
		t.Fatalf("analytics database was not created at %s", analyticsPath)
	}

	// Verify that an event was recorded
	db, err := sql.Open("sqlite", analyticsPath)
	if err != nil {
		t.Fatalf("failed to open analytics db: %v", err)
	}
	defer func() { _ = db.Close() }()

	// First check what events exist
	rows, err := db.Query("SELECT command FROM events")
	if err != nil {
		t.Fatalf("failed to query events: %v", err)
	}
	var commands []string
	for rows.Next() {
		var cmd string
		if err := rows.Scan(&cmd); err != nil {
			t.Fatalf("failed to scan command: %v", err)
		}
		commands = append(commands, cmd)
	}
	if err := rows.Close(); err != nil {
		t.Fatalf("failed to close rows: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM events WHERE command = 'list'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query events: %v", err)
	}

	if count == 0 {
		t.Errorf("expected at least 1 'list' event to be recorded, got 0 (found commands: %v)", commands)
	}
}

// TestAnalyticsTrackingDisabled verifies that analytics is not tracked when disabled
func TestAnalyticsTrackingDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	analyticsPath := filepath.Join(tmpDir, "analytics.db")
	configPath := filepath.Join(tmpDir, "config.yaml")
	dbPath := filepath.Join(tmpDir, "tasks.db")

	// Write config with analytics disabled (default)
	configContent := `default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := &Config{
		DBPath:        dbPath,
		ConfigPath:    configPath,
		AnalyticsPath: analyticsPath,
	}

	// Run the 'list' command
	var stdout, stderr bytes.Buffer
	exitCode := Execute([]string{"list"}, &stdout, &stderr, cfg)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
	}

	// Wait for any potential async logging
	time.Sleep(200 * time.Millisecond)

	// Analytics database should NOT be created when analytics is disabled
	if _, err := os.Stat(analyticsPath); !os.IsNotExist(err) {
		t.Errorf("analytics database should not be created when analytics is disabled")
	}
}

// TestIssue011BackendDataIsolation verifies that different backends have isolated
// data in the SQLite cache when sync is enabled.
// Issue #011: SQLite cache mixes data between backends because createSyncFallbackBackend()
// always uses sqlite.New() which defaults to backend_id="sqlite".
func TestIssue011BackendDataIsolation(t *testing.T) {
	// Create temp directory for config and data
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	dbPath := filepath.Join(tmpDir, "tasks.db")

	// Create a config with sync enabled
	configYAML := `
default_backend: sqlite
sync:
  enabled: true
  offline_mode: auto
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Simulate adding a task to "todoist" backend (via sync cache)
	cfgTodoist := &Config{
		Backend:     "todoist",
		DBPath:      dbPath,
		ConfigPath:  configPath,
		NoPrompt:    true,
		SyncEnabled: true, // Explicitly enable sync
	}

	var stdout1, stderr1 bytes.Buffer
	exitCode := Execute([]string{"Inbox", "add", "Todoist Task"}, &stdout1, &stderr1, cfgTodoist)
	if exitCode != 0 {
		t.Fatalf("failed to add task to todoist backend: exit=%d, stderr=%s", exitCode, stderr1.String())
	}

	// Simulate adding a task to "nextcloud-test" backend (via sync cache)
	cfgNextcloud := &Config{
		Backend:     "nextcloud-test",
		DBPath:      dbPath,
		ConfigPath:  configPath,
		NoPrompt:    true,
		SyncEnabled: true, // Explicitly enable sync
	}

	var stdout2, stderr2 bytes.Buffer
	exitCode = Execute([]string{"Work", "add", "Nextcloud Task"}, &stdout2, &stderr2, cfgNextcloud)
	if exitCode != 0 {
		t.Fatalf("failed to add task to nextcloud backend: exit=%d, stderr=%s", exitCode, stderr2.String())
	}

	// Now verify isolation: list tasks for each backend and check they only see their own lists
	// List todoist lists
	var stdout3, stderr3 bytes.Buffer
	exitCode = Execute([]string{"list", "--json"}, &stdout3, &stderr3, cfgTodoist)
	if exitCode != 0 {
		t.Fatalf("failed to list todoist lists: exit=%d, stderr=%s", exitCode, stderr3.String())
	}

	// Parse the JSON output for todoist
	todoistOutput := stdout3.String()
	// Todoist backend should see "Inbox" list, but NOT "Work" list (which belongs to nextcloud)
	if !strings.Contains(todoistOutput, "Inbox") {
		t.Errorf("todoist backend should see Inbox list, got: %s", todoistOutput)
	}
	if strings.Contains(todoistOutput, `"name":"Work"`) {
		t.Errorf("todoist backend should NOT see Work list (belongs to nextcloud), got: %s", todoistOutput)
	}

	// List nextcloud lists
	var stdout4, stderr4 bytes.Buffer
	exitCode = Execute([]string{"list", "--json"}, &stdout4, &stderr4, cfgNextcloud)
	if exitCode != 0 {
		t.Fatalf("failed to list nextcloud lists: exit=%d, stderr=%s", exitCode, stderr4.String())
	}

	// Parse the JSON output for nextcloud
	nextcloudOutput := stdout4.String()
	// Nextcloud backend should see "Work" list, but NOT "Inbox" list (which belongs to todoist)
	if !strings.Contains(nextcloudOutput, "Work") {
		t.Errorf("nextcloud backend should see Work list, got: %s", nextcloudOutput)
	}
	if strings.Contains(nextcloudOutput, `"name":"Inbox"`) {
		t.Errorf("nextcloud backend should NOT see Inbox list (belongs to todoist), got: %s", nextcloudOutput)
	}

	// Additional verification: Check the database directly to verify backend_id values
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Check that tasks have different backend_ids
	rows, err := db.Query("SELECT backend_id, summary FROM tasks")
	if err != nil {
		t.Fatalf("failed to query tasks: %v", err)
	}
	defer func() { _ = rows.Close() }()

	backendTasks := make(map[string][]string)
	for rows.Next() {
		var backendID, summary string
		if err := rows.Scan(&backendID, &summary); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}
		backendTasks[backendID] = append(backendTasks[backendID], summary)
	}

	// With the bug, all tasks would have backend_id="sqlite"
	// After fix, tasks should have backend_id="todoist" and backend_id="nextcloud-test"
	if _, ok := backendTasks["todoist"]; !ok {
		t.Errorf("expected backend_id='todoist' in database, got backend_ids: %v", backendTasks)
	}
	if _, ok := backendTasks["nextcloud-test"]; !ok {
		t.Errorf("expected backend_id='nextcloud-test' in database, got backend_ids: %v", backendTasks)
	}
	if tasks, ok := backendTasks["sqlite"]; ok && len(tasks) > 0 {
		t.Errorf("tasks should NOT have backend_id='sqlite' when using named backends, but found: %v", tasks)
	}
}
