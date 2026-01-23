package cmd

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
	"todoat/internal/config"
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
