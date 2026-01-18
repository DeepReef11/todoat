package utils

import (
	"errors"
	"testing"
	"time"
)

// =============================================================================
// Validation Tests (034-logging-utilities)
// =============================================================================

// TestValidatePriorityValid verifies valid priorities pass
func TestValidatePriorityValid(t *testing.T) {
	validPriorities := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	for _, p := range validPriorities {
		t.Run("Priority"+string(rune('0'+p)), func(t *testing.T) {
			err := ValidatePriority(p)
			if err != nil {
				t.Errorf("ValidatePriority(%d) = %v, want nil", p, err)
			}
		})
	}
}

// TestValidatePriorityInvalid verifies invalid priorities return error
func TestValidatePriorityInvalid(t *testing.T) {
	invalidPriorities := []int{-1, -10, 10, 15, 100}

	for _, p := range invalidPriorities {
		t.Run("Priority"+string(rune('0'+p)), func(t *testing.T) {
			err := ValidatePriority(p)
			if err == nil {
				t.Errorf("ValidatePriority(%d) = nil, want error", p)
			}

			// Should return ErrInvalidPriority
			var errWithSuggestion *ErrorWithSuggestion
			if !errors.As(err, &errWithSuggestion) {
				t.Errorf("ValidatePriority(%d) should return *ErrorWithSuggestion", p)
			}
		})
	}
}

// TestParseDateFlagValid verifies valid date strings parse correctly
func TestParseDateFlagValid(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{"2026-01-15", time.Date(2026, 1, 15, 0, 0, 0, 0, time.Local)},
		{"2025-12-31", time.Date(2025, 12, 31, 0, 0, 0, 0, time.Local)},
		{"2024-06-01", time.Date(2024, 6, 1, 0, 0, 0, 0, time.Local)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseDateFlag(tt.input)
			if err != nil {
				t.Errorf("ParseDateFlag(%q) error = %v", tt.input, err)
				return
			}
			if result == nil {
				t.Errorf("ParseDateFlag(%q) = nil, want %v", tt.input, tt.expected)
				return
			}

			// Compare year, month, day
			if result.Year() != tt.expected.Year() ||
				result.Month() != tt.expected.Month() ||
				result.Day() != tt.expected.Day() {
				t.Errorf("ParseDateFlag(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseDateFlagEmpty verifies empty string returns nil (clear date)
func TestParseDateFlagEmpty(t *testing.T) {
	result, err := ParseDateFlag("")
	if err != nil {
		t.Errorf("ParseDateFlag(\"\") error = %v, want nil", err)
	}
	if result != nil {
		t.Errorf("ParseDateFlag(\"\") = %v, want nil", result)
	}
}

// TestParseDateFlagInvalid verifies invalid dates return error
func TestParseDateFlagInvalid(t *testing.T) {
	invalidDates := []string{
		"invalid",
		"2026/01/15", // Wrong separator
		"01-15-2026", // Wrong order
		"2026-1-15",  // Missing leading zeros (may be valid depending on parser)
		"not-a-date",
		"tomorrow",
		"2026-13-01", // Invalid month
		"2026-01-32", // Invalid day
	}

	for _, input := range invalidDates {
		t.Run(input, func(t *testing.T) {
			_, err := ParseDateFlag(input)
			if err == nil {
				t.Errorf("ParseDateFlag(%q) = nil error, want error", input)
			}
		})
	}
}

// TestParseDateFlagReturnsErrorWithSuggestion verifies error has format hint
func TestParseDateFlagReturnsErrorWithSuggestion(t *testing.T) {
	_, err := ParseDateFlag("invalid-date")
	if err == nil {
		t.Fatal("Expected error for invalid date")
	}

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(err, &errWithSuggestion) {
		t.Error("ParseDateFlag should return *ErrorWithSuggestion for invalid dates")
	}
}

// TestValidateDateRangeValid verifies valid date ranges pass
func TestValidateDateRangeValid(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	due := time.Date(2026, 1, 15, 0, 0, 0, 0, time.Local)

	err := ValidateDateRange(&start, &due)
	if err != nil {
		t.Errorf("ValidateDateRange(start, due) = %v, want nil", err)
	}
}

// TestValidateDateRangeSameDay verifies same day is valid
func TestValidateDateRangeSameDay(t *testing.T) {
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.Local)

	err := ValidateDateRange(&date, &date)
	if err != nil {
		t.Errorf("ValidateDateRange(same, same) = %v, want nil", err)
	}
}

// TestValidateDateRangeStartAfterDue verifies start after due returns error
func TestValidateDateRangeStartAfterDue(t *testing.T) {
	start := time.Date(2026, 1, 20, 0, 0, 0, 0, time.Local)
	due := time.Date(2026, 1, 15, 0, 0, 0, 0, time.Local)

	err := ValidateDateRange(&start, &due)
	if err == nil {
		t.Error("ValidateDateRange(start > due) should return error")
	}
}

// TestValidateDateRangeNilValues verifies nil dates are valid
func TestValidateDateRangeNilValues(t *testing.T) {
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name  string
		start *time.Time
		due   *time.Time
	}{
		{"nil start", nil, &date},
		{"nil due", &date, nil},
		{"both nil", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDateRange(tt.start, tt.due)
			if err != nil {
				t.Errorf("ValidateDateRange(%v, %v) = %v, want nil", tt.start, tt.due, err)
			}
		})
	}
}
