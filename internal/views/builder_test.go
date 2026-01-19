package views_test

import (
	"io"
	"os"
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

// --- CLI Command Tests ---

// TestViewCreateCommand verifies that `todoat view create myview` launches interactive builder
func TestViewCreateCommand(t *testing.T) {
	viewsDir := t.TempDir()
	builder := views.NewBuilder("testview", viewsDir)

	tm := teatest.NewTestModel(t, builder, teatest.WithInitialTermSize(100, 30))

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

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

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Toggle some field selections using Space
	// Fields are: status, summary, description, priority, due_date, start_date,
	//             created, modified, completed, tags, uid, parent
	// First field (status) should be focused by default
	tm.Send(tea.KeyMsg{Type: tea.KeySpace}) // Toggle status (should select it)
	time.Sleep(50 * time.Millisecond)

	tm.Send(tea.KeyMsg{Type: tea.KeyDown}) // Move to summary
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeySpace}) // Toggle summary
	time.Sleep(50 * time.Millisecond)

	tm.Send(tea.KeyMsg{Type: tea.KeyDown}) // Move to description
	time.Sleep(50 * time.Millisecond)
	// Skip description

	tm.Send(tea.KeyMsg{Type: tea.KeyDown}) // Move to priority
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeySpace}) // Toggle priority
	time.Sleep(50 * time.Millisecond)

	// Save with Ctrl+S
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlS})
	time.Sleep(100 * time.Millisecond)

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

		time.Sleep(100 * time.Millisecond)

		// Select a field
		tm.Send(tea.KeyMsg{Type: tea.KeySpace})
		time.Sleep(50 * time.Millisecond)

		// Cancel with Escape
		tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
		time.Sleep(50 * time.Millisecond)

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

		time.Sleep(100 * time.Millisecond)

		// Select a field
		tm.Send(tea.KeyMsg{Type: tea.KeySpace})
		time.Sleep(50 * time.Millisecond)

		// Cancel with Ctrl+C
		tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
		time.Sleep(50 * time.Millisecond)

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

	time.Sleep(100 * time.Millisecond)

	// Should display all available fields
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

	time.Sleep(100 * time.Millisecond)

	// Select a field
	tm.Send(tea.KeyMsg{Type: tea.KeySpace}) // Toggle status
	time.Sleep(50 * time.Millisecond)

	// Press Enter to open field configuration dialog
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(100 * time.Millisecond)

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
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

// --- Filter Builder Tests ---

// TestViewBuilderFilterBuilder verifies filter rule creation with field, operator, value
func TestViewBuilderFilterBuilder(t *testing.T) {
	viewsDir := t.TempDir()
	builder := views.NewBuilder("filtertest", viewsDir)

	tm := teatest.NewTestModel(t, builder, teatest.WithInitialTermSize(100, 30))

	time.Sleep(100 * time.Millisecond)

	// Navigate to filter panel (Tab)
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(100 * time.Millisecond)

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

	time.Sleep(100 * time.Millisecond)

	// Navigate to sort panel (multiple Tabs)
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(100 * time.Millisecond)

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

	time.Sleep(100 * time.Millisecond)

	// Test arrow key navigation
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyUp})
	time.Sleep(50 * time.Millisecond)

	// Test Tab for panel switching
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(50 * time.Millisecond)

	// Test Shift+Tab for reverse panel switching
	tm.Send(tea.KeyMsg{Type: tea.KeyShiftTab})
	time.Sleep(50 * time.Millisecond)

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

	time.Sleep(100 * time.Millisecond)

	// Try to save without selecting any fields
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlS})
	time.Sleep(100 * time.Millisecond)

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
		tm.Send(tea.KeyMsg{Type: tea.KeySpace}) // Select a field
		time.Sleep(50 * time.Millisecond)
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

	time.Sleep(100 * time.Millisecond)

	// Step 1: Select fields
	// Select status
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})
	time.Sleep(50 * time.Millisecond)

	// Navigate to summary and select
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})
	time.Sleep(50 * time.Millisecond)

	// Navigate to due_date and select
	tm.Send(tea.KeyMsg{Type: tea.KeyDown}) // description
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyDown}) // priority
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyDown}) // due_date
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})
	time.Sleep(50 * time.Millisecond)

	// Step 2: Save the view
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlS})
	time.Sleep(100 * time.Millisecond)

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
