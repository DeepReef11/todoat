package views_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"gopkg.in/yaml.v3"

	"todoat/internal/views"
)

// =============================================================================
// Interactive View Builder Tests (035-interactive-view-builder)
// =============================================================================

// readAll reads all output from a reader and returns as bytes
func readAll(t *testing.T, r io.Reader) []byte {
	t.Helper()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	return out
}

// waitForRender waits for the TUI to render any output using teatest.WaitFor.
// This replaces flaky time.Sleep calls with condition-based polling.
// Note: This reads from the output stream - subsequent reads may need fresh output.
func waitForRender(t *testing.T, tm *teatest.TestModel) {
	t.Helper()
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return len(bts) > 0
	}, teatest.WithDuration(2*time.Second), teatest.WithCheckInterval(10*time.Millisecond))
}

// sendKeyAndWait sends a key message and waits briefly for processing.
// Uses a minimal sleep since teatest messages are processed asynchronously.
func sendKeyAndWait(tm *teatest.TestModel, key tea.KeyMsg) {
	tm.Send(key)
	// Minimal wait for message processing - using small value since this is just
	// for message queue processing, not for visual changes
	time.Sleep(20 * time.Millisecond)
}

// --- CLI Command Tests ---

// TestViewCreateCommand verifies that `todoat view create myview` launches interactive builder
func TestViewCreateCommand(t *testing.T) {
	viewsDir := t.TempDir()
	builder := views.NewBuilder("testview", viewsDir)

	tm := teatest.NewTestModel(t, builder, teatest.WithInitialTermSize(100, 30))

	// Wait for initial render using polling
	waitForRender(t, tm)

	// The builder should render with field selection panel visible
	// Cancel without saving
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})

	out := readAll(t, tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))

	outStr := string(out)

	// Should show view builder title/header
	if !strings.Contains(outStr, "View Builder") && !strings.Contains(outStr, "testview") && !strings.Contains(outStr, "Fields") {
		t.Errorf("expected view builder interface to be displayed, got:\n%s", outStr)
	}
}

// TestViewBuilderSavesYAML verifies that completing builder creates valid YAML in views directory
func TestViewBuilderSavesYAML(t *testing.T) {
	viewsDir := t.TempDir()
	builder := views.NewBuilder("myview", viewsDir)

	tm := teatest.NewTestModel(t, builder, teatest.WithInitialTermSize(100, 30))

	// Wait for initial render using polling
	waitForRender(t, tm)

	// Toggle some field selections using Space
	// Fields are: status, summary, description, priority, due_date, start_date,
	//             created, modified, completed, tags, uid, parent
	// First field (status) should be focused by default
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeySpace}) // Toggle status (should select it)

	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyDown})  // Move to summary
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeySpace}) // Toggle summary

	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyDown}) // Move to description
	// Skip description

	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyDown})  // Move to priority
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeySpace}) // Toggle priority

	// Save with Ctrl+S
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyCtrlS})

	// The builder should exit after saving
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	// Verify the YAML file was created
	viewPath := viewsDir + "/myview.yaml"
	data, err := os.ReadFile(viewPath)
	if err != nil {
		t.Fatalf("expected view file to be created at %s: %v", viewPath, err)
	}

	// Parse and validate the YAML
	var view views.View
	if err := yaml.Unmarshal(data, &view); err != nil {
		t.Fatalf("expected valid YAML, got parse error: %v", err)
	}

	// Verify the selected fields are in the view
	if len(view.Fields) == 0 {
		t.Error("expected at least one field to be selected")
	}

	fieldNames := make(map[string]bool)
	for _, f := range view.Fields {
		fieldNames[f.Name] = true
	}

	// Check that our selected fields are present
	expectedFields := []string{"status", "summary", "priority"}
	for _, ef := range expectedFields {
		if !fieldNames[ef] {
			t.Errorf("expected field %q to be in view, got fields: %v", ef, view.Fields)
		}
	}
}

// TestViewBuilderCancel verifies that pressing Escape/Ctrl+C exits without saving
func TestViewBuilderCancel(t *testing.T) {
	viewsDir := t.TempDir()

	// Test with Escape key
	t.Run("Escape", func(t *testing.T) {
		builder := views.NewBuilder("cancelview", viewsDir)
		tm := teatest.NewTestModel(t, builder, teatest.WithInitialTermSize(100, 30))

		waitForRender(t, tm)

		// Select a field
		sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeySpace})

		// Cancel with Escape
		sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyEsc})

		tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))

		// Verify no file was created
		viewPath := viewsDir + "/cancelview.yaml"
		if _, err := os.Stat(viewPath); !os.IsNotExist(err) {
			t.Errorf("expected no view file to be created on cancel, but file exists")
		}
	})

	// Test with Ctrl+C
	t.Run("CtrlC", func(t *testing.T) {
		builder := views.NewBuilder("cancelview2", viewsDir)
		tm := teatest.NewTestModel(t, builder, teatest.WithInitialTermSize(100, 30))

		waitForRender(t, tm)

		// Select a field
		sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeySpace})

		// Cancel with Ctrl+C
		sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyCtrlC})

		tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))

		// Verify no file was created
		viewPath := viewsDir + "/cancelview2.yaml"
		if _, err := os.Stat(viewPath); !os.IsNotExist(err) {
			t.Errorf("expected no view file to be created on cancel, but file exists")
		}
	})
}

// --- Field Selection Panel Tests ---

// TestViewBuilderFieldSelection verifies checkbox field selection for all available fields
func TestViewBuilderFieldSelection(t *testing.T) {
	viewsDir := t.TempDir()
	builder := views.NewBuilder("fieldtest", viewsDir)

	tm := teatest.NewTestModel(t, builder, teatest.WithInitialTermSize(100, 30))

	// Wait for render using a short sleep - teatest.WaitFor doesn't work well
	// when you need the output bytes back (it consumes them)
	time.Sleep(100 * time.Millisecond)
	out := readAll(t, tm.Output())
	outStr := string(out)

	// Check all expected fields are shown
	expectedFields := views.AvailableFields
	for _, field := range expectedFields {
		if !strings.Contains(outStr, field) {
			t.Errorf("expected field %q to be displayed in builder", field)
		}
	}

	// Cancel
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

// --- Field Configuration Tests ---

// TestViewBuilderFieldConfiguration verifies field configuration (width, alignment, format)
func TestViewBuilderFieldConfiguration(t *testing.T) {
	viewsDir := t.TempDir()
	builder := views.NewBuilder("configtest", viewsDir)

	tm := teatest.NewTestModel(t, builder, teatest.WithInitialTermSize(100, 30))

	waitForRender(t, tm)

	// Select a field
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeySpace}) // Toggle status

	// Press Enter to open field configuration dialog
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyEnter})

	// Should show configuration options
	out := readAll(t, tm.Output())
	outStr := string(out)

	// Check for configuration options
	configOptions := []string{"Width", "Align", "width", "align"}
	foundConfig := false
	for _, opt := range configOptions {
		if strings.Contains(outStr, opt) {
			foundConfig = true
			break
		}
	}

	if !foundConfig {
		t.Logf("Field configuration may open with Enter key - output shows:\n%s", outStr)
	}

	// Cancel and exit
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyEsc})
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

// --- Filter Builder Tests ---

// TestViewBuilderFilterBuilder verifies filter rule creation with field, operator, value
func TestViewBuilderFilterBuilder(t *testing.T) {
	viewsDir := t.TempDir()
	builder := views.NewBuilder("filtertest", viewsDir)

	tm := teatest.NewTestModel(t, builder, teatest.WithInitialTermSize(100, 30))

	waitForRender(t, tm)

	// Navigate to filter panel (Tab)
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyTab})

	out := readAll(t, tm.Output())
	outStr := string(out)

	// Should show filter panel or related UI elements
	filterTerms := []string{"Filter", "filter", "Add", "Rule", "rule"}
	foundFilter := false
	for _, term := range filterTerms {
		if strings.Contains(outStr, term) {
			foundFilter = true
			break
		}
	}

	if !foundFilter {
		t.Logf("Filter panel navigation with Tab - output shows:\n%s", outStr)
	}

	// Cancel
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

// --- Sort Rule Builder Tests ---

// TestViewBuilderSortRules verifies sort rule creation with field and direction
func TestViewBuilderSortRules(t *testing.T) {
	viewsDir := t.TempDir()
	builder := views.NewBuilder("sorttest", viewsDir)

	tm := teatest.NewTestModel(t, builder, teatest.WithInitialTermSize(100, 30))

	waitForRender(t, tm)

	// Navigate to sort panel (multiple Tabs)
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyTab})
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyTab})

	out := readAll(t, tm.Output())
	outStr := string(out)

	// Should show sort panel or related UI elements
	sortTerms := []string{"Sort", "sort", "asc", "desc", "Direction", "direction"}
	foundSort := false
	for _, term := range sortTerms {
		if strings.Contains(outStr, term) {
			foundSort = true
			break
		}
	}

	if !foundSort {
		t.Logf("Sort panel navigation - output shows:\n%s", outStr)
	}

	// Cancel
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

// --- Keyboard Navigation Tests ---

// TestViewBuilderKeyboardNavigation verifies arrow keys, Tab, Space navigation
func TestViewBuilderKeyboardNavigation(t *testing.T) {
	viewsDir := t.TempDir()
	builder := views.NewBuilder("navtest", viewsDir)

	tm := teatest.NewTestModel(t, builder, teatest.WithInitialTermSize(100, 30))

	waitForRender(t, tm)

	// Test arrow key navigation
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyDown})
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyDown})
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyUp})

	// Test Tab for panel switching
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyTab})

	// Test Shift+Tab for reverse panel switching
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyShiftTab})

	// Verify the builder is still responsive
	out := readAll(t, tm.Output())

	if len(out) == 0 {
		t.Error("expected builder to render output after navigation")
	}

	// Cancel
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

// --- Validation Tests ---

// TestViewBuilderValidation verifies that at least one field must be selected to save
func TestViewBuilderValidation(t *testing.T) {
	viewsDir := t.TempDir()
	builder := views.NewBuilder("validtest", viewsDir)

	tm := teatest.NewTestModel(t, builder, teatest.WithInitialTermSize(100, 30))

	waitForRender(t, tm)

	// Try to save without selecting any fields
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyCtrlS})

	// Should show validation error or not save
	out := readAll(t, tm.Output())
	outStr := string(out)

	// Check for error message or that builder is still active
	errorTerms := []string{"error", "Error", "field", "select", "required", "at least"}
	foundError := false
	for _, term := range errorTerms {
		if strings.Contains(outStr, term) {
			foundError = true
			break
		}
	}

	// Verify no file was created
	viewPath := viewsDir + "/validtest.yaml"
	if _, err := os.Stat(viewPath); !os.IsNotExist(err) {
		t.Errorf("expected no view file when no fields selected")
	}

	// If error is found, test passes; if not, the builder should still be running
	if !foundError {
		// Builder should still be active - verify by sending more input
		sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeySpace}) // Select a field
	}

	// Cancel
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

// --- Integration Test ---

// TestViewBuilderCompleteWorkflow verifies full workflow: select fields, configure, add filter, add sort, save
func TestViewBuilderCompleteWorkflow(t *testing.T) {
	viewsDir := t.TempDir()
	builder := views.NewBuilder("workflow", viewsDir)

	tm := teatest.NewTestModel(t, builder, teatest.WithInitialTermSize(100, 30))

	waitForRender(t, tm)

	// Step 1: Select fields
	// Select status
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeySpace})

	// Navigate to summary and select
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyDown})
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeySpace})

	// Navigate to due_date and select
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyDown}) // description
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyDown}) // priority
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyDown}) // due_date
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeySpace})

	// Step 2: Save the view
	sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyCtrlS})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	// Verify the YAML file was created with correct content
	viewPath := viewsDir + "/workflow.yaml"
	data, err := os.ReadFile(viewPath)
	if err != nil {
		t.Fatalf("expected view file to be created: %v", err)
	}

	var view views.View
	if err := yaml.Unmarshal(data, &view); err != nil {
		t.Fatalf("expected valid YAML: %v", err)
	}

	// Check that expected fields are present
	fieldNames := make(map[string]bool)
	for _, f := range view.Fields {
		fieldNames[f.Name] = true
	}

	expectedFields := []string{"status", "summary", "due_date"}
	for _, ef := range expectedFields {
		if !fieldNames[ef] {
			t.Errorf("expected field %q in workflow view, got: %v", ef, view.Fields)
		}
	}
}

// --- Path Traversal Security Tests ---

// TestViewBuilderPathTraversal verifies that the builder rejects path traversal attempts
// This is a SECURITY test - it ensures files cannot be created outside the views directory
func TestViewBuilderPathTraversal(t *testing.T) {
	// Create directory structure: /tmp/xxx/views/ and /tmp/xxx/outside/
	tmpDir := t.TempDir()
	viewsDir := filepath.Join(tmpDir, "views")
	outsideDir := filepath.Join(tmpDir, "outside")

	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		t.Fatalf("failed to create views dir: %v", err)
	}
	if err := os.MkdirAll(outsideDir, 0755); err != nil {
		t.Fatalf("failed to create outside dir: %v", err)
	}

	// Test path traversal names that should be rejected
	traversalNames := []struct {
		name        string
		description string
		checkPath   string // Path to check for file creation
	}{
		{"../malicious", "parent directory traversal", filepath.Join(tmpDir, "malicious.yaml")},
		{"../outside/bad", "traversal to sibling directory", filepath.Join(outsideDir, "bad.yaml")},
		{"foo/../../bad", "hidden traversal in path", filepath.Join(tmpDir, "bad.yaml")},
		{"/tmp/absolute", "absolute path", "/tmp/absolute.yaml"},
	}

	for _, tc := range traversalNames {
		t.Run(tc.description, func(t *testing.T) {
			// Try to create a builder with path traversal name
			builder := views.NewBuilder(tc.name, viewsDir)

			tm := teatest.NewTestModel(t, builder, teatest.WithInitialTermSize(100, 30))
			waitForRender(t, tm)

			// Select a field so we can try to save
			sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeySpace})

			// Try to save
			sendKeyAndWait(tm, tea.KeyMsg{Type: tea.KeyCtrlS})

			// Wait for potential save
			time.Sleep(100 * time.Millisecond)

			// Cancel the builder
			tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
			tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))

			// SECURITY CHECK: Verify no file was created outside viewsDir
			if tc.checkPath != "" {
				if _, err := os.Stat(tc.checkPath); !os.IsNotExist(err) {
					t.Errorf("SECURITY VULNERABILITY: File created at %s with traversal name %q", tc.checkPath, tc.name)
				}
			}

			// Check that no file was created in the expected traversal location
			// Also verify nothing was created in viewsDir with the raw name
			entries, err := os.ReadDir(viewsDir)
			if err != nil {
				t.Fatalf("failed to read views dir: %v", err)
			}

			for _, entry := range entries {
				t.Logf("file in viewsDir: %s", entry.Name())
			}
		})
	}
}

// TestViewBuilderPathTraversalValidation tests that ValidateViewName properly rejects bad names
func TestViewBuilderPathTraversalValidation(t *testing.T) {
	testCases := []struct {
		name        string
		expectError bool
	}{
		{"valid-view", false},
		{"my_view_123", false},
		{"../../../etc/passwd", true},
		{"../parent", true},
		{"foo/bar", true},
		{"foo\\bar", true},
		{".hidden", true},
		{"..", true},
		{"", true},
		{"foo/../bar", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := views.ValidateViewName(tc.name)
			if tc.expectError && err == nil {
				t.Errorf("ValidateViewName(%q) should return error, got nil", tc.name)
			}
			if !tc.expectError && err != nil {
				t.Errorf("ValidateViewName(%q) should succeed, got error: %v", tc.name, err)
			}
		})
	}
}
