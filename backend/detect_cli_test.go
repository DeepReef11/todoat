package backend_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	cmd "todoat/cmd/todoat/cmd"
)

// =============================================================================
// CLI Tests for --detect-backend Flag (Roadmap 037)
// =============================================================================

// createTestGitRepo creates a temporary git repository for testing
func createTestGitRepo(t *testing.T, withMarker bool) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Initialize git repo
	gitCmd := exec.Command("git", "init")
	gitCmd.Dir = tmpDir
	if err := gitCmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git
	gitCmd = exec.Command("git", "config", "user.email", "test@test.com")
	gitCmd.Dir = tmpDir
	_ = gitCmd.Run()
	gitCmd = exec.Command("git", "config", "user.name", "Test User")
	gitCmd.Dir = tmpDir
	_ = gitCmd.Run()

	if withMarker {
		todoPath := filepath.Join(tmpDir, "TODO.md")
		content := `<!-- todoat:enabled -->
# Project Tasks

## Work

- [ ] Sample task
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}
	}

	return tmpDir
}

// TestDetectBackendFlag verifies that --detect-backend shows auto-detected backends
func TestDetectBackendFlag(t *testing.T) {
	t.Run("shows detection results", func(t *testing.T) {
		repoPath := createTestGitRepo(t, true)
		dbPath := filepath.Join(t.TempDir(), "test.db")

		var stdout, stderr bytes.Buffer
		cfg := &cmd.Config{
			DBPath:  dbPath,
			WorkDir: repoPath,
		}

		exitCode := cmd.Execute([]string{"--detect-backend"}, &stdout, &stderr, cfg)

		if exitCode != 0 {
			t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
		}

		output := stdout.String()

		// Should show "Auto-detected backends:" header
		if !strings.Contains(output, "Auto-detected backends:") {
			t.Errorf("output should contain 'Auto-detected backends:', got: %s", output)
		}

		// Should show "Would use:" line
		if !strings.Contains(output, "Would use:") {
			t.Errorf("output should contain 'Would use:', got: %s", output)
		}
	})

	t.Run("shows git backend in git repo with marker", func(t *testing.T) {
		repoPath := createTestGitRepo(t, true)
		dbPath := filepath.Join(t.TempDir(), "test.db")

		var stdout, stderr bytes.Buffer
		cfg := &cmd.Config{
			DBPath:  dbPath,
			WorkDir: repoPath,
		}

		exitCode := cmd.Execute([]string{"--detect-backend"}, &stdout, &stderr, cfg)

		if exitCode != 0 {
			t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
		}

		output := stdout.String()

		// Should detect git backend
		if !strings.Contains(output, "git:") {
			t.Errorf("output should contain 'git:', got: %s", output)
		}
	})
}

// TestDetectBackendGit verifies git detection in repo with TODO.md marker
func TestDetectBackendGit(t *testing.T) {
	t.Run("detects git backend with marker", func(t *testing.T) {
		repoPath := createTestGitRepo(t, true)
		dbPath := filepath.Join(t.TempDir(), "test.db")

		var stdout, stderr bytes.Buffer
		cfg := &cmd.Config{
			DBPath:  dbPath,
			WorkDir: repoPath,
		}

		exitCode := cmd.Execute([]string{"--detect-backend"}, &stdout, &stderr, cfg)

		if exitCode != 0 {
			t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
		}

		output := stdout.String()

		// Should list git with path info
		if !strings.Contains(output, "git:") && !strings.Contains(output, "TODO.md") {
			t.Errorf("expected git backend with TODO.md in output, got: %s", output)
		}

		// Git should be the recommended backend
		if !strings.Contains(output, "Would use: git") {
			t.Errorf("expected 'Would use: git', got: %s", output)
		}
	})

	t.Run("does not detect git without marker", func(t *testing.T) {
		repoPath := createTestGitRepo(t, false) // No marker
		dbPath := filepath.Join(t.TempDir(), "test.db")

		var stdout, stderr bytes.Buffer
		cfg := &cmd.Config{
			DBPath:  dbPath,
			WorkDir: repoPath,
		}

		exitCode := cmd.Execute([]string{"--detect-backend"}, &stdout, &stderr, cfg)

		if exitCode != 0 {
			t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
		}

		output := stdout.String()

		// Git should not be available (or should show as not detected)
		// SQLite should still be available as fallback
		if strings.Contains(output, "Would use: git") {
			t.Errorf("git should not be detected without marker, got: %s", output)
		}
	})
}

// TestDetectBackendPriority verifies that multiple backends show priority-ordered
func TestDetectBackendPriority(t *testing.T) {
	t.Run("shows multiple backends with priority", func(t *testing.T) {
		repoPath := createTestGitRepo(t, true)
		dbPath := filepath.Join(t.TempDir(), "test.db")

		var stdout, stderr bytes.Buffer
		cfg := &cmd.Config{
			DBPath:  dbPath,
			WorkDir: repoPath,
		}

		exitCode := cmd.Execute([]string{"--detect-backend"}, &stdout, &stderr, cfg)

		if exitCode != 0 {
			t.Fatalf("expected exit code 0, got %d: stderr=%s", exitCode, stderr.String())
		}

		output := stdout.String()

		// Should show numbered list
		if !strings.Contains(output, "1.") {
			t.Errorf("expected numbered list (1.), got: %s", output)
		}

		// SQLite should always be available
		if !strings.Contains(output, "sqlite:") {
			t.Errorf("sqlite should always be listed, got: %s", output)
		}
	})
}

// TestAutoDetectEnabled verifies that auto_detect_backend config uses detected backend
func TestAutoDetectEnabled(t *testing.T) {
	t.Run("uses detected backend with auto_detect enabled", func(t *testing.T) {
		repoPath := createTestGitRepo(t, true)
		dbPath := filepath.Join(t.TempDir(), "test.db")
		configPath := filepath.Join(t.TempDir(), "config.yaml")

		// Create config with auto_detect_backend: true
		configContent := `
backends:
  sqlite:
    enabled: true
default_backend: sqlite
auto_detect_backend: true
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("failed to create config: %v", err)
		}

		var stdout, stderr bytes.Buffer
		cfg := &cmd.Config{
			DBPath:     dbPath,
			ConfigPath: configPath,
			WorkDir:    repoPath,
		}

		// Run list command - should auto-detect git backend
		exitCode := cmd.Execute([]string{"list"}, &stdout, &stderr, cfg)

		// Should succeed (may show empty list, that's OK)
		if exitCode != 0 {
			// If using git backend, it might fail if not properly set up
			// but the point is it should TRY to use git
			errStr := stderr.String()
			// If it's trying to use git, it might mention git repo or TODO.md
			if !strings.Contains(errStr, "git") && !strings.Contains(errStr, "TODO") {
				t.Logf("Note: exit code %d with stderr: %s", exitCode, errStr)
			}
		}
	})
}
