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

// timePattern matches time components like 14:30 or 14:30:00
var timePattern = regexp.MustCompile(`^(\d{1,2}):(\d{2})(?::(\d{2}))?$`)

// parseTimeComponent parses a time string like "14:30" or "14:30:00" and returns hour, minute, second.
// Returns -1, -1, -1 if the string is not a valid time format.
func parseTimeComponent(timeStr string) (hour, minute, second int) {
	matches := timePattern.FindStringSubmatch(timeStr)
	if matches == nil {
		return -1, -1, -1
	}

	hour, _ = strconv.Atoi(matches[1])
	minute, _ = strconv.Atoi(matches[2])
	second = 0
	if matches[3] != "" {
		second, _ = strconv.Atoi(matches[3])
	}

	// Validate ranges
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 || second < 0 || second > 59 {
		return -1, -1, -1
	}

	return hour, minute, second
}

// parseRelativeDate parses relative date strings like "today", "tomorrow", "yesterday", "+7d", "-3d", "+2w", "+1m".
// Also supports optional time component: "tomorrow 14:30", "today 09:00", "+7d 10:00".
// Returns nil if the string is not a relative date format.
// Returns the calculated time and nil for valid relative dates.
// Returns nil and error for invalid relative date values.
func parseRelativeDate(dateStr string) (*time.Time, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	// Check for time component (space-separated)
	parts := strings.SplitN(dateStr, " ", 2)
	datePart := parts[0]
	var timePart string
	if len(parts) == 2 {
		timePart = parts[1]
	}

	lower := strings.ToLower(datePart)

	var baseDate time.Time
	matched := false

	switch lower {
	case "today":
		baseDate = today
		matched = true
	case "tomorrow":
		baseDate = today.AddDate(0, 0, 1)
		matched = true
	case "yesterday":
		baseDate = today.AddDate(0, 0, -1)
		matched = true
	default:
		// Check for relative format (+/-Nd, +/-Nw, +/-Nm)
		matches := relativePattern.FindStringSubmatch(lower)
		if matches != nil {
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

			switch unit {
			case "d":
				baseDate = today.AddDate(0, 0, num)
			case "w":
				baseDate = today.AddDate(0, 0, num*7)
			case "m":
				baseDate = today.AddDate(0, num, 0)
			}
			matched = true
		}
	}

	if !matched {
		return nil, nil // Not a relative format
	}

	// Apply time component if present
	if timePart != "" {
		hour, minute, second := parseTimeComponent(timePart)
		if hour < 0 {
			return nil, ErrInvalidDate(dateStr)
		}
		baseDate = time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), hour, minute, second, 0, time.Local)
	}

	return &baseDate, nil
}

// ParseDateFlag parses a date string supporting both relative and absolute formats.
// Supported relative formats: today, tomorrow, yesterday, +Nd, -Nd, +Nw, +Nm
// Relative formats also support time: tomorrow 14:30, +7d 09:00
// Supported absolute formats:
//   - YYYY-MM-DD (date only, midnight assumed)
//   - YYYY-MM-DDTHH:MM or YYYY-MM-DDTHH:MM:SS (datetime)
//   - YYYY-MM-DDTHH:MM±HH:MM or YYYY-MM-DDTHH:MMZ (datetime with timezone)
//
// Returns nil, nil for empty string (clear date).
// Returns parsed time and nil for valid date.
// Returns nil and error for invalid date.
func ParseDateFlag(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	// Try relative date first (handles time component via space separator)
	t, err := parseRelativeDate(dateStr)
	if err != nil {
		return nil, err
	}
	if t != nil {
		return t, nil
	}

	// Try various datetime formats
	// Order matters: try most specific first

	// RFC3339 / ISO8601 with timezone (2026-01-20T14:30:00-05:00 or 2026-01-20T14:30:00Z)
	if parsed, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return &parsed, nil
	}

	// ISO8601 datetime without seconds and with timezone offset (2026-01-20T14:30-05:00)
	if parsed, err := time.Parse("2006-01-02T15:04-07:00", dateStr); err == nil {
		return &parsed, nil
	}

	// ISO8601 datetime without seconds, UTC (2026-01-20T14:30Z)
	if parsed, err := time.Parse("2006-01-02T15:04Z", dateStr); err == nil {
		return &parsed, nil
	}

	// ISO8601 datetime with seconds, local timezone (2026-01-20T14:30:00)
	if parsed, err := time.ParseInLocation("2006-01-02T15:04:05", dateStr, time.Local); err == nil {
		return &parsed, nil
	}

	// ISO8601 datetime without seconds, local timezone (2026-01-20T14:30)
	if parsed, err := time.ParseInLocation("2006-01-02T15:04", dateStr, time.Local); err == nil {
		return &parsed, nil
	}

	// Date only (2026-01-20)
	if parsed, err := time.ParseInLocation("2006-01-02", dateStr, time.Local); err == nil {
		return &parsed, nil
	}

	return nil, ErrInvalidDate(dateStr)
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
