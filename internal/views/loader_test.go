package views

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoaderPathTraversal tests that view names with path traversal sequences are rejected
func TestLoaderPathTraversal(t *testing.T) {
	// Create a temp directory structure to test actual path traversal
	// Structure: /tmp/xxx/views/ and /tmp/xxx/secret/
	tmpDir := t.TempDir()
	viewsDir := filepath.Join(tmpDir, "views")
	secretDir := filepath.Join(tmpDir, "secret")

	err := os.MkdirAll(viewsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create views dir: %v", err)
	}
	err = os.MkdirAll(secretDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create secret dir: %v", err)
	}

	// Create a valid view file in the views directory
	validViewContent := `name: test
description: A test view
fields:
  - name: status
  - name: summary
`
	err = os.WriteFile(filepath.Join(viewsDir, "test.yaml"), []byte(validViewContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test view file: %v", err)
	}

	// Create a "secret" YAML file OUTSIDE the views directory that we'll try to access via traversal
	secretContent := `name: secret
description: This should NOT be accessible
fields:
  - name: status
`
	err = os.WriteFile(filepath.Join(secretDir, "secret.yaml"), []byte(secretContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create secret file: %v", err)
	}

	loader := NewLoader(viewsDir)

	// Test that valid view name works
	_, err = loader.LoadView("test")
	if err != nil {
		t.Errorf("LoadView('test') should succeed, got error: %v", err)
	}

	// CRITICAL TEST: Try to access the secret file via path traversal
	// The path "../secret/secret" from viewsDir should resolve to secretDir/secret.yaml
	t.Run("actual_traversal_to_existing_file", func(t *testing.T) {
		view, err := loader.LoadView("../secret/secret")
		if err == nil {
			t.Errorf("SECURITY VULNERABILITY: LoadView('../secret/secret') succeeded in reading file outside views directory! Got view: %+v", view)
		} else if !strings.Contains(err.Error(), "invalid view name") && !strings.Contains(err.Error(), "not allowed") {
			// If the error is just "not found" or parse error, the traversal might have worked but file wasn't valid YAML
			// We need to check if the error indicates the path was actually blocked
			t.Logf("LoadView('../secret/secret') error: %v - verify this is a security rejection, not just file format issue", err)
		}
	})

	// Test standard path traversal patterns that should be rejected
	traversalTests := []struct {
		name        string
		viewName    string
		description string
	}{
		{"parent directory", "../../../etc/passwd", "path traversal with ../"},
		{"parent with name", "../sibling/view", "path traversal to sibling directory"},
		{"double parent", "../../view", "double parent traversal"},
		{"hidden parent", "foo/../../../etc/passwd", "hidden traversal in path"},
		{"absolute path unix", "/etc/passwd", "absolute path on Unix"},
		{"forward slash", "foo/bar", "forward slash in name"},
		{"backslash", "foo\\bar", "backslash in name"},
		{"dot dot", "..", "just dot dot"},
		{"dot dot slash", "../", "dot dot slash"},
	}

	for _, tc := range traversalTests {
		t.Run(tc.name, func(t *testing.T) {
			view, err := loader.LoadView(tc.viewName)
			if err == nil {
				t.Errorf("LoadView(%q) should have failed (%s), but succeeded and returned view: %+v", tc.viewName, tc.description, view)
			}
			// Verify it's a security-related rejection with explicit error message
			if err != nil && !strings.Contains(err.Error(), "invalid view name") {
				t.Errorf("LoadView(%q) should fail with 'invalid view name' error, got: %v", tc.viewName, err)
			}
		})
	}
}

// TestLoaderViewExistsPathTraversal tests that ViewExists also rejects path traversal
func TestLoaderViewExistsPathTraversal(t *testing.T) {
	// Create a temp directory structure with a file outside views dir
	tmpDir := t.TempDir()
	viewsDir := filepath.Join(tmpDir, "views")
	secretDir := filepath.Join(tmpDir, "secret")

	os.MkdirAll(viewsDir, 0755)
	os.MkdirAll(secretDir, 0755)

	// Create a file outside the views directory
	os.WriteFile(filepath.Join(secretDir, "secret.yaml"), []byte("test"), 0644)

	loader := NewLoader(viewsDir)

	// Path traversal names should return false, even if the traversed file exists
	traversalNames := []string{
		"../../../etc/passwd",
		"../secret/secret", // This file actually exists!
		"foo/../bar",
		"/etc/passwd",
		"foo/bar",
	}

	for _, name := range traversalNames {
		t.Run(name, func(t *testing.T) {
			if loader.ViewExists(name) {
				t.Errorf("ViewExists(%q) should return false for path traversal attempt", name)
			}
		})
	}
}
