package cmd

import (
	"bytes"
	"strings"
	"testing"
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
