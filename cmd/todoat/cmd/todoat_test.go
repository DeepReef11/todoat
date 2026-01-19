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

	// Initialize database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	_ = db.Close()

	// Capture real stderr (logger writes to os.Stderr)
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	var stdout, stderrBuf bytes.Buffer

	cfg := &Config{
		DBPath: dbPath,
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

	// Initialize database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	_ = db.Close()

	// Capture real stderr (logger writes to os.Stderr)
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	var stdout, stderrBuf bytes.Buffer

	cfg := &Config{
		DBPath: dbPath,
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

// TestConfigPassthrough verifies that config is accessible
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

// TestRootCommandShowsHelp verifies that running without args shows help
func TestRootCommandShowsHelpCoreCLI(t *testing.T) {
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
	cfg := &Config{
		DBPath: dbPath,
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

	// Set XDG environment variables for this test
	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	defer func() {
		_ = os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		_ = os.Setenv("XDG_DATA_HOME", oldDataHome)
	}()

	configDir := filepath.Join(tempHome, ".config")
	dataDir := filepath.Join(tempHome, ".local", "share")
	if err := os.Setenv("XDG_CONFIG_HOME", configDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}
	if err := os.Setenv("XDG_DATA_HOME", dataDir); err != nil {
		t.Fatalf("failed to set XDG_DATA_HOME: %v", err)
	}

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

	// Set XDG environment variables for this test
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	defer func() {
		_ = os.Setenv("XDG_DATA_HOME", oldDataHome)
	}()

	dataDir := filepath.Join(tempHome, ".local", "share")
	if err := os.Setenv("XDG_DATA_HOME", dataDir); err != nil {
		t.Fatalf("failed to set XDG_DATA_HOME: %v", err)
	}

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

	// Set XDG environment variables for this test
	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		_ = os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
	}()

	configDir := filepath.Join(tempHome, ".config")
	if err := os.Setenv("XDG_CONFIG_HOME", configDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}

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

	// Set XDG environment variables for this test
	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	defer func() {
		_ = os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		_ = os.Setenv("XDG_DATA_HOME", oldDataHome)
	}()

	configDir := filepath.Join(tempHome, ".config")
	dataDir := filepath.Join(tempHome, ".local", "share")
	if err := os.Setenv("XDG_CONFIG_HOME", configDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}
	if err := os.Setenv("XDG_DATA_HOME", dataDir); err != nil {
		t.Fatalf("failed to set XDG_DATA_HOME: %v", err)
	}

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

	cfg := &Config{
		DBPath: dbPath,
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

	cfg := &Config{
		DBPath: dbPath,
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

	cfg := &Config{
		DBPath: dbPath,
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

	cfg := &Config{
		DBPath: dbPath,
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

			cfg := &Config{
				DBPath: dbPath,
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
