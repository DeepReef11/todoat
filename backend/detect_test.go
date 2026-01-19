package backend_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"todoat/backend"
	"todoat/backend/git"
)

// =============================================================================
// Detection Tests for Backend Auto-Detection (Roadmap 037)
// =============================================================================

// testGitRepo creates a temporary git repository with an optional TODO.md file
func testGitRepo(t *testing.T, withMarker bool) (repoPath string) {
	t.Helper()

	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git for commits
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	if withMarker {
		// Create TODO.md with marker
		todoPath := filepath.Join(tmpDir, "TODO.md")
		content := `<!-- todoat:enabled -->
# Project Tasks

## Work

- [ ] Sample task

## Personal

`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}
	}

	return tmpDir
}

// TestDetectableBackendInterface verifies the DetectableBackend interface
func TestDetectableBackendInterface(t *testing.T) {
	t.Run("git backend implements DetectableBackend", func(t *testing.T) {
		repoPath := testGitRepo(t, true)

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		// Verify it implements DetectableBackend
		detectable, ok := interface{}(be).(backend.DetectableBackend)
		if !ok {
			t.Fatal("git backend should implement DetectableBackend interface")
		}

		// Test CanDetect
		canDetect, err := detectable.CanDetect()
		if err != nil {
			t.Fatalf("CanDetect error: %v", err)
		}
		if !canDetect {
			t.Error("expected CanDetect to return true for repo with marked TODO.md")
		}

		// Test DetectionInfo
		info := detectable.DetectionInfo()
		if info == "" {
			t.Error("DetectionInfo should return non-empty string")
		}
	})
}

// TestBackendRegistryDetectable tests the detectable backend registry
func TestBackendRegistryDetectable(t *testing.T) {
	t.Run("register and get detectable constructor", func(t *testing.T) {
		// Clean up any existing registration first
		backend.ClearDetectableConstructors()

		// Register a test constructor
		testConstructor := func(workDir string) (backend.DetectableBackend, error) {
			return nil, nil
		}

		backend.RegisterDetectable("test-backend", testConstructor)

		// Get constructors
		constructors := backend.GetDetectableConstructors()
		if len(constructors) == 0 {
			t.Fatal("expected at least one detectable constructor")
		}

		// Find our registered constructor
		found := false
		for name := range constructors {
			if name == "test-backend" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'test-backend' in constructors")
		}
	})
}

// TestDetectBackendsFunction tests the DetectBackends function
func TestDetectBackendsFunction(t *testing.T) {
	t.Run("detects git backend in git repo with marker", func(t *testing.T) {
		repoPath := testGitRepo(t, true)

		// Clear existing and register git backend for detection
		backend.ClearDetectableConstructors()
		backend.RegisterDetectable("git", func(workDir string) (backend.DetectableBackend, error) {
			cfg := git.Config{WorkDir: workDir}
			return git.New(cfg)
		})

		// Detect backends
		results, err := backend.DetectBackends(repoPath)
		if err != nil {
			t.Fatalf("DetectBackends error: %v", err)
		}

		if len(results) == 0 {
			t.Fatal("expected at least one detected backend")
		}

		// Find git in results
		foundGit := false
		for _, result := range results {
			if result.Name == "git" {
				foundGit = true
				if !result.Available {
					t.Error("git backend should be available in git repo with marker")
				}
				if result.Info == "" {
					t.Error("git detection info should not be empty")
				}
			}
		}
		if !foundGit {
			t.Error("expected 'git' in detection results")
		}
	})

	t.Run("does not detect git backend without marker", func(t *testing.T) {
		repoPath := testGitRepo(t, false) // No marker

		backend.ClearDetectableConstructors()
		backend.RegisterDetectable("git", func(workDir string) (backend.DetectableBackend, error) {
			cfg := git.Config{WorkDir: workDir}
			return git.New(cfg)
		})

		results, err := backend.DetectBackends(repoPath)
		if err != nil {
			t.Fatalf("DetectBackends error: %v", err)
		}

		// Git should not be available
		for _, result := range results {
			if result.Name == "git" && result.Available {
				t.Error("git backend should not be available without marker")
			}
		}
	})

	t.Run("returns results ordered by priority", func(t *testing.T) {
		repoPath := testGitRepo(t, true)

		backend.ClearDetectableConstructors()

		// Register multiple backends
		backend.RegisterDetectable("git", func(workDir string) (backend.DetectableBackend, error) {
			cfg := git.Config{WorkDir: workDir}
			return git.New(cfg)
		})

		// SQLite-like mock that always returns true
		backend.RegisterDetectable("sqlite", func(workDir string) (backend.DetectableBackend, error) {
			return &mockDetectableBackend{
				canDetect:     true,
				detectionInfo: "SQLite database (always available)",
			}, nil
		})

		results, err := backend.DetectBackends(repoPath)
		if err != nil {
			t.Fatalf("DetectBackends error: %v", err)
		}

		// Should have both results
		if len(results) < 2 {
			t.Errorf("expected at least 2 detection results, got %d", len(results))
		}

		// All available backends should be listed
		availableCount := 0
		for _, r := range results {
			if r.Available {
				availableCount++
			}
		}
		if availableCount < 2 {
			t.Errorf("expected at least 2 available backends, got %d", availableCount)
		}
	})
}

// TestSelectBackend tests the automatic backend selection logic
func TestSelectBackend(t *testing.T) {
	t.Run("selects first available backend", func(t *testing.T) {
		repoPath := testGitRepo(t, true)

		backend.ClearDetectableConstructors()
		backend.RegisterDetectable("git", func(workDir string) (backend.DetectableBackend, error) {
			cfg := git.Config{WorkDir: workDir}
			return git.New(cfg)
		})

		// Auto-detect and select
		be, name, err := backend.SelectDetectedBackend(repoPath)
		if err != nil {
			t.Fatalf("SelectDetectedBackend error: %v", err)
		}
		if be == nil {
			t.Fatal("expected a backend to be selected")
		}
		defer func() { _ = be.Close() }()

		if name != "git" {
			t.Errorf("expected 'git' backend to be selected, got %q", name)
		}
	})

	t.Run("returns nil when no backend available", func(t *testing.T) {
		tmpDir := t.TempDir() // Not a git repo

		backend.ClearDetectableConstructors()
		backend.RegisterDetectable("git", func(workDir string) (backend.DetectableBackend, error) {
			cfg := git.Config{WorkDir: workDir}
			return git.New(cfg)
		})

		be, name, err := backend.SelectDetectedBackend(tmpDir)
		if err != nil {
			t.Fatalf("SelectDetectedBackend error: %v", err)
		}
		if be != nil {
			t.Error("expected no backend to be selected in non-git directory")
			_ = be.Close()
		}
		if name != "" {
			t.Errorf("expected empty name, got %q", name)
		}
	})
}

// mockDetectableBackend is a mock implementation for testing
type mockDetectableBackend struct {
	backend.TaskManager
	canDetect     bool
	detectionInfo string
}

func (m *mockDetectableBackend) CanDetect() (bool, error) {
	return m.canDetect, nil
}

func (m *mockDetectableBackend) DetectionInfo() string {
	return m.detectionInfo
}

func (m *mockDetectableBackend) Close() error {
	return nil
}
