package utils

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorWithSuggestion wraps an error with a user-friendly suggestion.
type ErrorWithSuggestion struct {
	Err        error
	Suggestion string
}

// Error implements the error interface.
func (e *ErrorWithSuggestion) Error() string {
	return fmt.Sprintf("%s\n\nSuggestion: %s", e.Err.Error(), e.Suggestion)
}

// GetSuggestion returns the suggestion text.
func (e *ErrorWithSuggestion) GetSuggestion() string {
	return e.Suggestion
}

// Unwrap returns the underlying error for error chain support.
func (e *ErrorWithSuggestion) Unwrap() error {
	return e.Err
}

// WrapWithSuggestion wraps an existing error with a suggestion.
func WrapWithSuggestion(err error, suggestion string) error {
	return &ErrorWithSuggestion{
		Err:        err,
		Suggestion: suggestion,
	}
}

// ErrTaskNotFound returns an error for when a task is not found.
func ErrTaskNotFound(searchTerm string) error {
	return &ErrorWithSuggestion{
		Err:        fmt.Errorf("task not found: %s", searchTerm),
		Suggestion: "Check the search term or use 'todoat list' to see all tasks",
	}
}

// ErrListNotFound returns an error for when a list is not found.
func ErrListNotFound(listName string) error {
	return &ErrorWithSuggestion{
		Err:        fmt.Errorf("list not found: %s", listName),
		Suggestion: fmt.Sprintf("Create the list with 'todoat list create %s'", listName),
	}
}

// ErrNoListsAvailable returns an error when no lists exist.
func ErrNoListsAvailable() error {
	return &ErrorWithSuggestion{
		Err:        errors.New("no lists available"),
		Suggestion: "Create a list with 'todoat list create <name>'",
	}
}

// ErrSyncNotEnabled returns an error when sync is not configured.
func ErrSyncNotEnabled() error {
	return &ErrorWithSuggestion{
		Err:        errors.New("sync is not enabled"),
		Suggestion: "Enable sync in your config file or run 'todoat config edit'",
	}
}

// ErrBackendNotConfigured returns an error when a backend is not configured.
func ErrBackendNotConfigured(name string) error {
	return &ErrorWithSuggestion{
		Err:        fmt.Errorf("backend not configured: %s", name),
		Suggestion: fmt.Sprintf("Add %s configuration to your config file", name),
	}
}

// ErrBackendOffline returns an error when a backend is unreachable with smart suggestions.
func ErrBackendOffline(name, reason string) error {
	suggestion := getSmartSuggestion(reason)
	return &ErrorWithSuggestion{
		Err:        fmt.Errorf("backend %s is offline: %s", name, reason),
		Suggestion: suggestion,
	}
}

// getSmartSuggestion returns a context-aware suggestion based on the error reason.
func getSmartSuggestion(reason string) string {
	lowerReason := strings.ToLower(reason)

	if strings.Contains(lowerReason, "no such host") || strings.Contains(lowerReason, "dns") {
		return "Check your DNS settings and internet connection"
	}

	if strings.Contains(lowerReason, "connection refused") {
		return "Check if the server is running and accessible"
	}

	if strings.Contains(lowerReason, "timeout") || strings.Contains(lowerReason, "i/o timeout") {
		return "The server may be slow or unreachable. Try again later"
	}

	return "Check your internet connection and try again"
}

// ErrInvalidPriority returns an error for an invalid priority value.
func ErrInvalidPriority(priority int) error {
	return &ErrorWithSuggestion{
		Err:        fmt.Errorf("invalid priority: %d", priority),
		Suggestion: "Priority must be between 0 and 9",
	}
}

// ErrInvalidDate returns an error for an invalid date string.
func ErrInvalidDate(dateStr string) error {
	return &ErrorWithSuggestion{
		Err:        fmt.Errorf("invalid date: %s", dateStr),
		Suggestion: "Use date format YYYY-MM-DD (e.g., 2026-01-15)",
	}
}

// ErrInvalidStatus returns an error for an invalid status with valid options.
func ErrInvalidStatus(status string, valid []string) error {
	return &ErrorWithSuggestion{
		Err:        fmt.Errorf("invalid status: %s", status),
		Suggestion: fmt.Sprintf("Valid options: %s", strings.Join(valid, ", ")),
	}
}

// ErrCredentialsNotFound returns an error when credentials are missing.
func ErrCredentialsNotFound(backend, user string) error {
	return &ErrorWithSuggestion{
		Err:        fmt.Errorf("credentials not found for %s user %s", backend, user),
		Suggestion: fmt.Sprintf("Run 'todoat setup %s' to configure credentials", backend),
	}
}

// ErrAuthenticationFailed returns an error when authentication fails.
func ErrAuthenticationFailed(backend string) error {
	return &ErrorWithSuggestion{
		Err:        fmt.Errorf("authentication failed for %s", backend),
		Suggestion: "Verify your credentials are correct and have not expired",
	}
}
