package utils

import (
	"errors"
	"strings"
	"testing"
)

// =============================================================================
// Error Tests (034-logging-utilities)
// =============================================================================

// TestErrorWithSuggestionImplementsError verifies interface compliance
func TestErrorWithSuggestionImplementsError(t *testing.T) {
	var _ error = &ErrorWithSuggestion{}
}

// TestErrorWithSuggestionError verifies Error() method output
func TestErrorWithSuggestionError(t *testing.T) {
	err := &ErrorWithSuggestion{
		Err:        errors.New("something went wrong"),
		Suggestion: "Try doing X",
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "something went wrong") {
		t.Errorf("Error() should contain error message, got: %s", errStr)
	}
	if !strings.Contains(errStr, "Suggestion:") {
		t.Errorf("Error() should contain 'Suggestion:', got: %s", errStr)
	}
	if !strings.Contains(errStr, "Try doing X") {
		t.Errorf("Error() should contain suggestion text, got: %s", errStr)
	}
}

// TestErrorWithSuggestionGetSuggestion verifies Suggestion() method
func TestErrorWithSuggestionGetSuggestion(t *testing.T) {
	err := &ErrorWithSuggestion{
		Err:        errors.New("error"),
		Suggestion: "helpful suggestion",
	}

	if err.GetSuggestion() != "helpful suggestion" {
		t.Errorf("GetSuggestion() = %s, want 'helpful suggestion'", err.GetSuggestion())
	}
}

// TestErrorWithSuggestionUnwrap verifies Unwrap() for error chain
func TestErrorWithSuggestionUnwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &ErrorWithSuggestion{
		Err:        underlying,
		Suggestion: "suggestion",
	}

	unwrapped := errors.Unwrap(err)
	if unwrapped != underlying {
		t.Errorf("Unwrap() should return underlying error")
	}
}

// TestWrapWithSuggestion verifies WrapWithSuggestion function
func TestWrapWithSuggestion(t *testing.T) {
	underlying := errors.New("original error")
	wrapped := WrapWithSuggestion(underlying, "custom suggestion")

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(wrapped, &errWithSuggestion) {
		t.Fatal("WrapWithSuggestion should return *ErrorWithSuggestion")
	}

	if errWithSuggestion.GetSuggestion() != "custom suggestion" {
		t.Errorf("Suggestion = %s, want 'custom suggestion'", errWithSuggestion.GetSuggestion())
	}
}

// =============================================================================
// Pre-built Error Constructor Tests
// =============================================================================

// TestErrTaskNotFound verifies task not found error with search suggestion
func TestErrTaskNotFound(t *testing.T) {
	err := ErrTaskNotFound("my-task")

	errStr := err.Error()
	if !strings.Contains(errStr, "my-task") {
		t.Errorf("Error should contain search term, got: %s", errStr)
	}
	if !strings.Contains(strings.ToLower(errStr), "not found") || !strings.Contains(strings.ToLower(errStr), "task") {
		t.Errorf("Error should indicate task not found, got: %s", errStr)
	}

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(err, &errWithSuggestion) {
		t.Fatal("Should return *ErrorWithSuggestion")
	}

	suggestion := errWithSuggestion.GetSuggestion()
	if suggestion == "" {
		t.Error("Should have a suggestion")
	}
	// Suggestion should mention searching
	if !strings.Contains(strings.ToLower(suggestion), "search") && !strings.Contains(strings.ToLower(suggestion), "check") {
		t.Errorf("Suggestion should mention search/check, got: %s", suggestion)
	}
}

// TestErrListNotFound verifies list not found error with creation suggestion
func TestErrListNotFound(t *testing.T) {
	err := ErrListNotFound("MyList")

	errStr := err.Error()
	if !strings.Contains(errStr, "MyList") {
		t.Errorf("Error should contain list name, got: %s", errStr)
	}

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(err, &errWithSuggestion) {
		t.Fatal("Should return *ErrorWithSuggestion")
	}

	suggestion := errWithSuggestion.GetSuggestion()
	// Suggestion should mention creating the list
	if !strings.Contains(strings.ToLower(suggestion), "create") {
		t.Errorf("Suggestion should mention creating list, got: %s", suggestion)
	}
}

// TestErrNoListsAvailable verifies no lists error with list create suggestion
func TestErrNoListsAvailable(t *testing.T) {
	err := ErrNoListsAvailable()

	errStr := err.Error()
	if !strings.Contains(strings.ToLower(errStr), "no") || !strings.Contains(strings.ToLower(errStr), "list") {
		t.Errorf("Error should indicate no lists available, got: %s", errStr)
	}

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(err, &errWithSuggestion) {
		t.Fatal("Should return *ErrorWithSuggestion")
	}

	suggestion := errWithSuggestion.GetSuggestion()
	// Suggestion should mention creating a list
	if !strings.Contains(strings.ToLower(suggestion), "create") {
		t.Errorf("Suggestion should mention creating a list, got: %s", suggestion)
	}
}

// TestErrSyncNotEnabled verifies sync not enabled error with config suggestion
func TestErrSyncNotEnabled(t *testing.T) {
	err := ErrSyncNotEnabled()

	errStr := err.Error()
	if !strings.Contains(strings.ToLower(errStr), "sync") {
		t.Errorf("Error should mention sync, got: %s", errStr)
	}

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(err, &errWithSuggestion) {
		t.Fatal("Should return *ErrorWithSuggestion")
	}

	suggestion := errWithSuggestion.GetSuggestion()
	// Suggestion should mention configuration
	if !strings.Contains(strings.ToLower(suggestion), "config") {
		t.Errorf("Suggestion should mention configuration, got: %s", suggestion)
	}
}

// TestErrBackendNotConfigured verifies backend not configured error
func TestErrBackendNotConfigured(t *testing.T) {
	err := ErrBackendNotConfigured("nextcloud")

	errStr := err.Error()
	if !strings.Contains(errStr, "nextcloud") {
		t.Errorf("Error should contain backend name, got: %s", errStr)
	}
	if !strings.Contains(strings.ToLower(errStr), "not configured") && !strings.Contains(strings.ToLower(errStr), "not found") {
		t.Errorf("Error should indicate not configured, got: %s", errStr)
	}

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(err, &errWithSuggestion) {
		t.Fatal("Should return *ErrorWithSuggestion")
	}

	suggestion := errWithSuggestion.GetSuggestion()
	// Suggestion should provide config help
	if !strings.Contains(strings.ToLower(suggestion), "config") {
		t.Errorf("Suggestion should mention configuration, got: %s", suggestion)
	}
}

// TestErrBackendOfflineDNS verifies smart suggestion for DNS errors
func TestErrBackendOfflineDNS(t *testing.T) {
	err := ErrBackendOffline("nextcloud", "no such host")

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(err, &errWithSuggestion) {
		t.Fatal("Should return *ErrorWithSuggestion")
	}

	suggestion := errWithSuggestion.GetSuggestion()
	if !strings.Contains(strings.ToLower(suggestion), "dns") {
		t.Errorf("DNS error suggestion should mention DNS, got: %s", suggestion)
	}
}

// TestErrBackendOfflineConnectionRefused verifies smart suggestion for connection refused
func TestErrBackendOfflineConnectionRefused(t *testing.T) {
	err := ErrBackendOffline("sqlite", "connection refused")

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(err, &errWithSuggestion) {
		t.Fatal("Should return *ErrorWithSuggestion")
	}

	suggestion := errWithSuggestion.GetSuggestion()
	if !strings.Contains(strings.ToLower(suggestion), "server") || !strings.Contains(strings.ToLower(suggestion), "running") {
		t.Errorf("Connection refused suggestion should mention server running, got: %s", suggestion)
	}
}

// TestErrBackendOfflineTimeout verifies smart suggestion for timeout
func TestErrBackendOfflineTimeout(t *testing.T) {
	err := ErrBackendOffline("todoist", "i/o timeout")

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(err, &errWithSuggestion) {
		t.Fatal("Should return *ErrorWithSuggestion")
	}

	suggestion := errWithSuggestion.GetSuggestion()
	if !strings.Contains(strings.ToLower(suggestion), "slow") || !strings.Contains(strings.ToLower(suggestion), "try again") {
		t.Errorf("Timeout suggestion should mention slow/try again, got: %s", suggestion)
	}
}

// TestErrBackendOfflineDefault verifies default suggestion for unknown errors
func TestErrBackendOfflineDefault(t *testing.T) {
	err := ErrBackendOffline("backend", "unknown error xyz")

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(err, &errWithSuggestion) {
		t.Fatal("Should return *ErrorWithSuggestion")
	}

	suggestion := errWithSuggestion.GetSuggestion()
	if !strings.Contains(strings.ToLower(suggestion), "internet") || !strings.Contains(strings.ToLower(suggestion), "connection") {
		t.Errorf("Default suggestion should mention internet/connection, got: %s", suggestion)
	}
}

// TestErrInvalidPriority verifies invalid priority error with valid range
func TestErrInvalidPriority(t *testing.T) {
	err := ErrInvalidPriority(15)

	errStr := err.Error()
	if !strings.Contains(errStr, "15") {
		t.Errorf("Error should contain invalid priority value, got: %s", errStr)
	}

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(err, &errWithSuggestion) {
		t.Fatal("Should return *ErrorWithSuggestion")
	}

	suggestion := errWithSuggestion.GetSuggestion()
	// Suggestion should mention valid range 0-9
	if !strings.Contains(suggestion, "0") || !strings.Contains(suggestion, "9") {
		t.Errorf("Suggestion should mention valid range 0-9, got: %s", suggestion)
	}
}

// TestErrInvalidDate verifies invalid date error with format hint
func TestErrInvalidDate(t *testing.T) {
	err := ErrInvalidDate("not-a-date")

	errStr := err.Error()
	if !strings.Contains(errStr, "not-a-date") {
		t.Errorf("Error should contain invalid date string, got: %s", errStr)
	}

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(err, &errWithSuggestion) {
		t.Fatal("Should return *ErrorWithSuggestion")
	}

	suggestion := errWithSuggestion.GetSuggestion()
	// Suggestion should mention YYYY-MM-DD format
	if !strings.Contains(suggestion, "YYYY-MM-DD") {
		t.Errorf("Suggestion should mention date format YYYY-MM-DD, got: %s", suggestion)
	}
}

// TestErrInvalidStatus verifies invalid status error with valid options
func TestErrInvalidStatus(t *testing.T) {
	validOptions := []string{"pending", "completed", "cancelled"}
	err := ErrInvalidStatus("unknown", validOptions)

	errStr := err.Error()
	if !strings.Contains(errStr, "unknown") {
		t.Errorf("Error should contain invalid status, got: %s", errStr)
	}

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(err, &errWithSuggestion) {
		t.Fatal("Should return *ErrorWithSuggestion")
	}

	suggestion := errWithSuggestion.GetSuggestion()
	// Suggestion should list valid options
	for _, opt := range validOptions {
		if !strings.Contains(suggestion, opt) {
			t.Errorf("Suggestion should contain valid option '%s', got: %s", opt, suggestion)
		}
	}
}

// TestErrCredentialsNotFound verifies credentials not found error with setup command
func TestErrCredentialsNotFound(t *testing.T) {
	err := ErrCredentialsNotFound("nextcloud", "admin")

	errStr := err.Error()
	if !strings.Contains(errStr, "nextcloud") {
		t.Errorf("Error should contain backend name, got: %s", errStr)
	}
	if !strings.Contains(errStr, "admin") {
		t.Errorf("Error should contain username, got: %s", errStr)
	}

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(err, &errWithSuggestion) {
		t.Fatal("Should return *ErrorWithSuggestion")
	}

	suggestion := errWithSuggestion.GetSuggestion()
	// Suggestion should mention setup or credential command
	if !strings.Contains(strings.ToLower(suggestion), "setup") && !strings.Contains(strings.ToLower(suggestion), "credential") {
		t.Errorf("Suggestion should mention setup/credential, got: %s", suggestion)
	}
}

// TestErrAuthenticationFailed verifies auth failed error with verification suggestion
func TestErrAuthenticationFailed(t *testing.T) {
	err := ErrAuthenticationFailed("todoist")

	errStr := err.Error()
	if !strings.Contains(errStr, "todoist") {
		t.Errorf("Error should contain backend name, got: %s", errStr)
	}
	if !strings.Contains(strings.ToLower(errStr), "auth") {
		t.Errorf("Error should mention authentication, got: %s", errStr)
	}

	var errWithSuggestion *ErrorWithSuggestion
	if !errors.As(err, &errWithSuggestion) {
		t.Fatal("Should return *ErrorWithSuggestion")
	}

	suggestion := errWithSuggestion.GetSuggestion()
	// Suggestion should mention verification
	if !strings.Contains(strings.ToLower(suggestion), "verif") && !strings.Contains(strings.ToLower(suggestion), "check") {
		t.Errorf("Suggestion should mention verification/check, got: %s", suggestion)
	}
}
