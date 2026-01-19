package utils

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ValidatePriority validates that priority is within valid range (0-9).
func ValidatePriority(priority int) error {
	if priority < 0 || priority > 9 {
		return ErrInvalidPriority(priority)
	}
	return nil
}

// relativePattern matches relative date formats like +7d, -3d, +2w, +1m
var relativePattern = regexp.MustCompile(`^([+-])(\d+)([dwm])$`)

// parseRelativeDate parses relative date strings like "today", "tomorrow", "yesterday", "+7d", "-3d", "+2w", "+1m".
// Returns nil if the string is not a relative date format.
// Returns the calculated time and nil for valid relative dates.
// Returns nil and error for invalid relative date values.
func parseRelativeDate(dateStr string) (*time.Time, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	lower := strings.ToLower(dateStr)

	switch lower {
	case "today":
		return &today, nil
	case "tomorrow":
		t := today.AddDate(0, 0, 1)
		return &t, nil
	case "yesterday":
		t := today.AddDate(0, 0, -1)
		return &t, nil
	}

	// Check for relative format (+/-Nd, +/-Nw, +/-Nm)
	matches := relativePattern.FindStringSubmatch(lower)
	if matches == nil {
		return nil, nil // Not a relative format
	}

	sign := matches[1]
	numStr := matches[2]
	unit := matches[3]

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return nil, ErrInvalidDate(dateStr)
	}

	if sign == "-" {
		num = -num
	}

	var result time.Time
	switch unit {
	case "d":
		result = today.AddDate(0, 0, num)
	case "w":
		result = today.AddDate(0, 0, num*7)
	case "m":
		result = today.AddDate(0, num, 0)
	}

	return &result, nil
}

// ParseDateFlag parses a date string supporting both relative and absolute formats.
// Supported relative formats: today, tomorrow, yesterday, +Nd, -Nd, +Nw, +Nm
// Supported absolute format: YYYY-MM-DD
// Returns nil, nil for empty string (clear date).
// Returns parsed time and nil for valid date.
// Returns nil and error for invalid date.
func ParseDateFlag(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	// Try relative date first
	t, err := parseRelativeDate(dateStr)
	if err != nil {
		return nil, err
	}
	if t != nil {
		return t, nil
	}

	// Fall back to absolute date parsing in local timezone
	parsed, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return nil, ErrInvalidDate(dateStr)
	}

	return &parsed, nil
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
