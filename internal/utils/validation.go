package utils

import (
	"errors"
	"time"
)

// ValidatePriority validates that priority is within valid range (0-9).
func ValidatePriority(priority int) error {
	if priority < 0 || priority > 9 {
		return ErrInvalidPriority(priority)
	}
	return nil
}

// ParseDateFlag parses a date string in YYYY-MM-DD format.
// Returns nil, nil for empty string (clear date).
// Returns parsed time and nil for valid date.
// Returns nil and error for invalid date.
func ParseDateFlag(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	// Parse in local timezone
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return nil, ErrInvalidDate(dateStr)
	}

	return &t, nil
}

// ValidateDateRange validates that start date is not after due date.
// Nil dates are considered valid.
func ValidateDateRange(start, due *time.Time) error {
	if start == nil || due == nil {
		return nil
	}

	if start.After(*due) {
		return errors.New("start date cannot be after due date")
	}

	return nil
}
