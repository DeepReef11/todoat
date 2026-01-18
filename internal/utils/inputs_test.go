package utils

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// =============================================================================
// Input Tests (034-logging-utilities)
// =============================================================================

// TestPromptYesNoYes verifies yes responses
func TestPromptYesNoYes(t *testing.T) {
	yesInputs := []string{"y\n", "Y\n", "yes\n", "Yes\n", "YES\n"}

	for _, input := range yesInputs {
		t.Run(input[:len(input)-1], func(t *testing.T) {
			reader := strings.NewReader(input)
			result := PromptYesNoWithReader("Test?", reader, io.Discard)
			if !result {
				t.Errorf("PromptYesNo with input %q = false, want true", input)
			}
		})
	}
}

// TestPromptYesNoNo verifies no responses
func TestPromptYesNoNo(t *testing.T) {
	noInputs := []string{"n\n", "N\n", "no\n", "No\n", "NO\n"}

	for _, input := range noInputs {
		t.Run(input[:len(input)-1], func(t *testing.T) {
			reader := strings.NewReader(input)
			result := PromptYesNoWithReader("Test?", reader, io.Discard)
			if result {
				t.Errorf("PromptYesNo with input %q = true, want false", input)
			}
		})
	}
}

// TestPromptYesNoRetryOnInvalid verifies loop until valid input
func TestPromptYesNoRetryOnInvalid(t *testing.T) {
	// Invalid input followed by valid
	input := "invalid\nmaybe\ny\n"
	reader := strings.NewReader(input)
	var output bytes.Buffer

	result := PromptYesNoWithReader("Test?", reader, &output)
	if !result {
		t.Error("PromptYesNo should return true after valid 'y' input")
	}

	// Should have prompted multiple times
	outputStr := output.String()
	if strings.Count(outputStr, "Test?") < 2 {
		t.Error("PromptYesNo should re-prompt on invalid input")
	}
}

// TestPromptYesNoTrimInput verifies whitespace trimming
func TestPromptYesNoTrimInput(t *testing.T) {
	inputs := []string{"  y  \n", "\ty\t\n", " yes \n"}

	for _, input := range inputs {
		t.Run("trimmed", func(t *testing.T) {
			reader := strings.NewReader(input)
			result := PromptYesNoWithReader("Test?", reader, io.Discard)
			if !result {
				t.Errorf("PromptYesNo with whitespace input %q = false, want true", input)
			}
		})
	}
}

// TestPromptSelectionValid verifies valid selection returns index
func TestPromptSelectionValid(t *testing.T) {
	items := []string{"Item A", "Item B", "Item C"}
	input := "2\n" // Select second item
	reader := strings.NewReader(input)

	idx, err := PromptSelectionWithReader(items, "Select:", reader, io.Discard, func(i int, item string) {
		// Display function
	})

	if err != nil {
		t.Errorf("PromptSelection error = %v", err)
	}
	if idx != 1 { // 0-based index
		t.Errorf("PromptSelection = %d, want 1", idx)
	}
}

// TestPromptSelectionFirstItem verifies selecting first item
func TestPromptSelectionFirstItem(t *testing.T) {
	items := []string{"First", "Second"}
	input := "1\n"
	reader := strings.NewReader(input)

	idx, err := PromptSelectionWithReader(items, "Select:", reader, io.Discard, func(i int, item string) {})

	if err != nil {
		t.Errorf("PromptSelection error = %v", err)
	}
	if idx != 0 {
		t.Errorf("PromptSelection = %d, want 0", idx)
	}
}

// TestPromptSelectionLastItem verifies selecting last item
func TestPromptSelectionLastItem(t *testing.T) {
	items := []string{"A", "B", "C", "D"}
	input := "4\n"
	reader := strings.NewReader(input)

	idx, err := PromptSelectionWithReader(items, "Select:", reader, io.Discard, func(i int, item string) {})

	if err != nil {
		t.Errorf("PromptSelection error = %v", err)
	}
	if idx != 3 {
		t.Errorf("PromptSelection = %d, want 3", idx)
	}
}

// TestPromptSelectionCancel verifies 0 cancels and returns error
func TestPromptSelectionCancel(t *testing.T) {
	items := []string{"Item A", "Item B"}
	input := "0\n"
	reader := strings.NewReader(input)

	_, err := PromptSelectionWithReader(items, "Select:", reader, io.Discard, func(i int, item string) {})

	if err == nil {
		t.Error("PromptSelection with 0 should return error (cancelled)")
	}
}

// TestPromptSelectionOutOfRange verifies out of range retries
func TestPromptSelectionOutOfRange(t *testing.T) {
	items := []string{"A", "B"}
	input := "5\n2\n" // Invalid then valid
	reader := strings.NewReader(input)
	var output bytes.Buffer

	idx, err := PromptSelectionWithReader(items, "Select:", reader, &output, func(i int, item string) {})

	if err != nil {
		t.Errorf("PromptSelection error = %v", err)
	}
	if idx != 1 {
		t.Errorf("PromptSelection = %d, want 1", idx)
	}
}

// TestPromptSelectionNonNumeric verifies non-numeric input retries
func TestPromptSelectionNonNumeric(t *testing.T) {
	items := []string{"A", "B"}
	input := "abc\n1\n" // Invalid then valid
	reader := strings.NewReader(input)

	idx, err := PromptSelectionWithReader(items, "Select:", reader, io.Discard, func(i int, item string) {})

	if err != nil {
		t.Errorf("PromptSelection error = %v", err)
	}
	if idx != 0 {
		t.Errorf("PromptSelection = %d, want 0", idx)
	}
}

// TestPromptSelectionDisplayFunction verifies display function is called
func TestPromptSelectionDisplayFunction(t *testing.T) {
	items := []string{"First", "Second", "Third"}
	input := "1\n"
	reader := strings.NewReader(input)
	var output bytes.Buffer

	displayCalls := 0
	_, err := PromptSelectionWithReader(items, "Select:", reader, &output, func(i int, item string) {
		displayCalls++
		output.WriteString(item + "\n")
	})

	if err != nil {
		t.Errorf("PromptSelection error = %v", err)
	}
	if displayCalls != len(items) {
		t.Errorf("Display function called %d times, want %d", displayCalls, len(items))
	}

	// Verify items were displayed
	outputStr := output.String()
	for _, item := range items {
		if !strings.Contains(outputStr, item) {
			t.Errorf("Output should contain %q", item)
		}
	}
}

// TestPromptConfirmationYes verifies confirmation returns true for yes
func TestPromptConfirmationYes(t *testing.T) {
	input := "y\n"
	reader := strings.NewReader(input)

	confirmed, err := PromptConfirmationWithReader("Confirm?", reader, io.Discard)
	if err != nil {
		t.Errorf("PromptConfirmation error = %v", err)
	}
	if !confirmed {
		t.Error("PromptConfirmation with 'y' should return true")
	}
}

// TestPromptConfirmationNo verifies confirmation returns false for no
func TestPromptConfirmationNo(t *testing.T) {
	input := "n\n"
	reader := strings.NewReader(input)

	confirmed, err := PromptConfirmationWithReader("Confirm?", reader, io.Discard)
	if err != nil {
		t.Errorf("PromptConfirmation error = %v", err)
	}
	if confirmed {
		t.Error("PromptConfirmation with 'n' should return false")
	}
}

// TestReadInt verifies integer reading
func TestReadInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"42\n", 42, false},
		{"0\n", 0, false},
		{"-5\n", -5, false},
		{"  100  \n", 100, false},
		{"abc\n", 0, true},
		{"\n", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input[:len(tt.input)-1], func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result, err := ReadIntWithReader(reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ReadInt(%q) should return error", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ReadInt(%q) error = %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("ReadInt(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestReadString verifies string reading with trimming
func TestReadString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"hello\n", "hello", false},
		{"  trimmed  \n", "trimmed", false},
		{"\ttabs\t\n", "tabs", false},
		{"multi word string\n", "multi word string", false},
		{"\n", "", false}, // Empty is valid
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result, err := ReadStringWithReader(reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ReadString(%q) should return error", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ReadString(%q) error = %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("ReadString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
