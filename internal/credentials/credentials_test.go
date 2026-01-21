// Package credentials provides secure credential storage and retrieval
package credentials

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestCredentialsSetKeyring tests that credentials can be stored in the keyring
// CLI: todoat credentials set nextcloud myuser --prompt
func TestCredentialsSetKeyring(t *testing.T) {
	// Create a mock keyring for testing
	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	// Simulate setting credentials (in real usage, password comes from prompt)
	err := manager.Set(context.Background(), "nextcloud", "myuser", "testpassword123")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify the credentials were stored
	stored, err := mockKeyring.Get("todoat-nextcloud", "myuser")
	if err != nil {
		t.Fatalf("Keyring Get failed: %v", err)
	}
	if stored != "testpassword123" {
		t.Errorf("Expected password 'testpassword123', got '%s'", stored)
	}
}

// TestCredentialsGetKeyring tests that credentials can be retrieved from keyring
// CLI: todoat credentials get nextcloud myuser
func TestCredentialsGetKeyring(t *testing.T) {
	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	// Pre-store credentials in the mock keyring
	err := mockKeyring.Set("todoat-nextcloud", "myuser", "secretpass")
	if err != nil {
		t.Fatalf("Failed to pre-store credentials: %v", err)
	}

	// Retrieve credentials
	info, err := manager.Get(context.Background(), "nextcloud", "myuser")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if info.Source != SourceKeyring {
		t.Errorf("Expected source %s, got %s", SourceKeyring, info.Source)
	}
	if info.Username != "myuser" {
		t.Errorf("Expected username 'myuser', got '%s'", info.Username)
	}
	if info.Password != "secretpass" {
		t.Errorf("Expected password 'secretpass', got '%s'", info.Password)
	}
	if info.Backend != "nextcloud" {
		t.Errorf("Expected backend 'nextcloud', got '%s'", info.Backend)
	}
	if !info.Found {
		t.Error("Expected Found to be true")
	}
}

// TestCredentialsGetEnvVar tests credential retrieval from environment variables
// Environment: TODOAT_NEXTCLOUD_USERNAME, TODOAT_NEXTCLOUD_PASSWORD
func TestCredentialsGetEnvVar(t *testing.T) {
	// Save original env vars and restore after test
	origUser := os.Getenv("TODOAT_NEXTCLOUD_USERNAME")
	origPass := os.Getenv("TODOAT_NEXTCLOUD_PASSWORD")
	defer func() {
		_ = os.Setenv("TODOAT_NEXTCLOUD_USERNAME", origUser)
		_ = os.Setenv("TODOAT_NEXTCLOUD_PASSWORD", origPass)
	}()

	// Set environment variables
	_ = os.Setenv("TODOAT_NEXTCLOUD_USERNAME", "envuser")
	_ = os.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "envpass123")

	// Create manager with empty keyring
	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	// Retrieve credentials - should come from environment
	info, err := manager.Get(context.Background(), "nextcloud", "envuser")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if info.Source != SourceEnvironment {
		t.Errorf("Expected source %s, got %s", SourceEnvironment, info.Source)
	}
	if info.Username != "envuser" {
		t.Errorf("Expected username 'envuser', got '%s'", info.Username)
	}
	if info.Password != "envpass123" {
		t.Errorf("Expected password 'envpass123', got '%s'", info.Password)
	}
	if !info.Found {
		t.Error("Expected Found to be true")
	}
}

// TestCredentialsPriority tests that keyring takes precedence over env vars
func TestCredentialsPriority(t *testing.T) {
	// Save original env vars and restore after test
	origUser := os.Getenv("TODOAT_NEXTCLOUD_USERNAME")
	origPass := os.Getenv("TODOAT_NEXTCLOUD_PASSWORD")
	defer func() {
		_ = os.Setenv("TODOAT_NEXTCLOUD_USERNAME", origUser)
		_ = os.Setenv("TODOAT_NEXTCLOUD_PASSWORD", origPass)
	}()

	// Set environment variables
	_ = os.Setenv("TODOAT_NEXTCLOUD_USERNAME", "envuser")
	_ = os.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "envpass")

	// Create manager with keyring containing different credentials
	mockKeyring := NewMockKeyring()
	_ = mockKeyring.Set("todoat-nextcloud", "myuser", "keyringpass")
	manager := NewManager(WithKeyring(mockKeyring))

	// Retrieve credentials - keyring should take precedence
	info, err := manager.Get(context.Background(), "nextcloud", "myuser")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if info.Source != SourceKeyring {
		t.Errorf("Expected source %s, got %s", SourceKeyring, info.Source)
	}
	if info.Password != "keyringpass" {
		t.Errorf("Expected keyring password 'keyringpass', got '%s'", info.Password)
	}
}

// TestCredentialsDelete tests that credentials can be removed from keyring
// CLI: todoat credentials delete nextcloud myuser
func TestCredentialsDelete(t *testing.T) {
	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	// Pre-store credentials
	err := mockKeyring.Set("todoat-nextcloud", "myuser", "todelete")
	if err != nil {
		t.Fatalf("Failed to pre-store credentials: %v", err)
	}

	// Delete credentials
	err = manager.Delete(context.Background(), "nextcloud", "myuser")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify credentials are gone from keyring
	_, err = mockKeyring.Get("todoat-nextcloud", "myuser")
	if err == nil {
		t.Error("Expected error after deletion, credentials should be gone")
	}
}

// TestCredentialsNotFound tests that appropriate error is returned when credentials not found
// CLI: todoat credentials get nonexistent user
func TestCredentialsNotFound(t *testing.T) {
	// Clear any environment variables that could match
	origUser := os.Getenv("TODOAT_NONEXISTENT_USERNAME")
	origPass := os.Getenv("TODOAT_NONEXISTENT_PASSWORD")
	defer func() {
		_ = os.Setenv("TODOAT_NONEXISTENT_USERNAME", origUser)
		_ = os.Setenv("TODOAT_NONEXISTENT_PASSWORD", origPass)
	}()
	_ = os.Unsetenv("TODOAT_NONEXISTENT_USERNAME")
	_ = os.Unsetenv("TODOAT_NONEXISTENT_PASSWORD")

	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	// Try to get non-existent credentials
	info, err := manager.Get(context.Background(), "nonexistent", "user")
	if err != nil {
		t.Fatalf("Get should not fail, it should return info with Found=false: %v", err)
	}

	if info.Found {
		t.Error("Expected Found to be false for non-existent credentials")
	}
}

// TestCredentialsHiddenInput tests that password input is hidden during --prompt mode
// This verifies the interface for hidden input exists
func TestCredentialsHiddenInput(t *testing.T) {
	// Create a mock reader that simulates terminal input
	input := bytes.NewBufferString("mysecretpassword\n")
	output := &bytes.Buffer{}

	// The prompt function should hide input and return the password
	password, err := PromptPassword(input, output, "nextcloud", "myuser")
	if err != nil {
		t.Fatalf("PromptPassword failed: %v", err)
	}

	if password != "mysecretpassword" {
		t.Errorf("Expected password 'mysecretpassword', got '%s'", password)
	}

	// Verify the prompt was written
	promptOutput := output.String()
	if !strings.Contains(promptOutput, "nextcloud") {
		t.Errorf("Expected prompt to mention backend 'nextcloud', got '%s'", promptOutput)
	}
	if !strings.Contains(promptOutput, "myuser") {
		t.Errorf("Expected prompt to mention user 'myuser', got '%s'", promptOutput)
	}
}

// TestPromptPasswordWithTTY tests that PromptPasswordWithTTY uses the terminal reader
// when provided. This is a regression test for issue #003 - password displayed in plain text.
func TestPromptPasswordWithTTY(t *testing.T) {
	output := &bytes.Buffer{}

	// Create a mock terminal reader that returns password without echo
	mockTermReader := &mockTerminalReader{
		password: "hiddenpassword",
	}

	// Use the TTY-aware prompt function
	password, err := PromptPasswordWithTTY(nil, output, "nextcloud", "myuser", mockTermReader)
	if err != nil {
		t.Fatalf("PromptPasswordWithTTY failed: %v", err)
	}

	if password != "hiddenpassword" {
		t.Errorf("Expected password 'hiddenpassword', got '%s'", password)
	}

	// Verify the prompt was written
	promptOutput := output.String()
	if !strings.Contains(promptOutput, "nextcloud") {
		t.Errorf("Expected prompt to mention backend 'nextcloud', got '%s'", promptOutput)
	}

	// Verify mock was called (simulating masked input)
	if !mockTermReader.readCalled {
		t.Error("Expected terminal reader to be called for masked input")
	}
}

// TestPromptPasswordWithTTYFallback tests that when no TTY is available,
// the function falls back to reading from stdin (for piped input).
func TestPromptPasswordWithTTYFallback(t *testing.T) {
	input := bytes.NewBufferString("pipedpassword\n")
	output := &bytes.Buffer{}

	// No terminal reader provided (nil), should fall back to stdin
	password, err := PromptPasswordWithTTY(input, output, "nextcloud", "myuser", nil)
	if err != nil {
		t.Fatalf("PromptPasswordWithTTY fallback failed: %v", err)
	}

	if password != "pipedpassword" {
		t.Errorf("Expected password 'pipedpassword', got '%s'", password)
	}
}

// mockTerminalReader is a mock implementation of TerminalReader for testing
type mockTerminalReader struct {
	password   string
	readCalled bool
	err        error
}

func (m *mockTerminalReader) ReadPassword() (string, error) {
	m.readCalled = true
	if m.err != nil {
		return "", m.err
	}
	return m.password, nil
}

// TestCredentialsJSON tests JSON output format for credentials get
// CLI: todoat --json credentials get nextcloud myuser
func TestCredentialsJSON(t *testing.T) {
	mockKeyring := NewMockKeyring()
	_ = mockKeyring.Set("todoat-nextcloud", "myuser", "jsonpass")
	manager := NewManager(WithKeyring(mockKeyring))

	info, err := manager.Get(context.Background(), "nextcloud", "myuser")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Serialize to JSON
	jsonOutput, err := info.JSON()
	if err != nil {
		t.Fatalf("JSON serialization failed: %v", err)
	}

	// Parse the JSON to verify structure
	var parsed struct {
		Backend  string `json:"backend"`
		Username string `json:"username"`
		Source   string `json:"source"`
		Found    bool   `json:"found"`
	}
	err = json.Unmarshal(jsonOutput, &parsed)
	if err != nil {
		t.Fatalf("JSON parsing failed: %v", err)
	}

	if parsed.Backend != "nextcloud" {
		t.Errorf("Expected backend 'nextcloud', got '%s'", parsed.Backend)
	}
	if parsed.Username != "myuser" {
		t.Errorf("Expected username 'myuser', got '%s'", parsed.Username)
	}
	if parsed.Source != string(SourceKeyring) {
		t.Errorf("Expected source '%s', got '%s'", SourceKeyring, parsed.Source)
	}
	if !parsed.Found {
		t.Error("Expected found to be true")
	}

	// Password should NOT be in JSON output (security)
	if strings.Contains(string(jsonOutput), "jsonpass") {
		t.Error("Password should not appear in JSON output")
	}
}

// TestCredentialsListBackends tests listing all backends with credential status
// CLI: todoat credentials list
func TestCredentialsListBackends(t *testing.T) {
	// Save and restore todoist env vars that could interfere
	origToken := os.Getenv("TODOAT_TODOIST_TOKEN")
	origUser := os.Getenv("TODOAT_TODOIST_USERNAME")
	origPass := os.Getenv("TODOAT_TODOIST_PASSWORD")
	defer func() {
		_ = os.Setenv("TODOAT_TODOIST_TOKEN", origToken)
		_ = os.Setenv("TODOAT_TODOIST_USERNAME", origUser)
		_ = os.Setenv("TODOAT_TODOIST_PASSWORD", origPass)
	}()
	_ = os.Unsetenv("TODOAT_TODOIST_TOKEN")
	_ = os.Unsetenv("TODOAT_TODOIST_USERNAME")
	_ = os.Unsetenv("TODOAT_TODOIST_PASSWORD")

	mockKeyring := NewMockKeyring()
	// Store credentials for one backend
	_ = mockKeyring.Set("todoat-nextcloud", "myuser", "pass1")
	manager := NewManager(WithKeyring(mockKeyring))

	// Configure backends to check
	backendConfigs := []BackendConfig{
		{Name: "nextcloud", Username: "myuser"},
		{Name: "todoist", Username: "apiuser"},
	}

	statuses, err := manager.ListBackends(context.Background(), backendConfigs)
	if err != nil {
		t.Fatalf("ListBackends failed: %v", err)
	}

	if len(statuses) != 2 {
		t.Fatalf("Expected 2 backend statuses, got %d", len(statuses))
	}

	// nextcloud should have credentials
	var ncStatus, todoistStatus *BackendStatus
	for i := range statuses {
		if statuses[i].Backend == "nextcloud" {
			ncStatus = &statuses[i]
		}
		if statuses[i].Backend == "todoist" {
			todoistStatus = &statuses[i]
		}
	}

	if ncStatus == nil {
		t.Fatal("Expected nextcloud in status list")
	}
	if !ncStatus.HasCredentials {
		t.Error("Expected nextcloud to have credentials")
	}
	if ncStatus.Source != SourceKeyring {
		t.Errorf("Expected source %s for nextcloud, got %s", SourceKeyring, ncStatus.Source)
	}

	if todoistStatus == nil {
		t.Fatal("Expected todoist in status list")
	}
	if todoistStatus.HasCredentials {
		t.Error("Expected todoist to NOT have credentials")
	}
}

// TestCredentialsDeleteIdempotent tests that deleting non-existent credentials succeeds
func TestCredentialsDeleteIdempotent(t *testing.T) {
	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	// Delete non-existent credentials should succeed (idempotent)
	err := manager.Delete(context.Background(), "nonexistent", "user")
	if err != nil {
		t.Errorf("Delete of non-existent credentials should succeed, got: %v", err)
	}
}

// TestCredentialsSetOverwrite tests that Set overwrites existing credentials
func TestCredentialsSetOverwrite(t *testing.T) {
	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	// Set initial credentials
	err := manager.Set(context.Background(), "nextcloud", "myuser", "oldpass")
	if err != nil {
		t.Fatalf("Initial Set failed: %v", err)
	}

	// Overwrite with new password
	err = manager.Set(context.Background(), "nextcloud", "myuser", "newpass")
	if err != nil {
		t.Fatalf("Overwrite Set failed: %v", err)
	}

	// Verify new password is stored
	info, err := manager.Get(context.Background(), "nextcloud", "myuser")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if info.Password != "newpass" {
		t.Errorf("Expected password 'newpass', got '%s'", info.Password)
	}
}

// TestCredentialsBackendNameNormalization tests that backend names are normalized
func TestCredentialsBackendNameNormalization(t *testing.T) {
	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	// Set with uppercase backend name
	err := manager.Set(context.Background(), "NEXTCLOUD", "myuser", "pass")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get with lowercase backend name
	info, err := manager.Get(context.Background(), "nextcloud", "myuser")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !info.Found {
		t.Error("Expected credentials to be found with normalized backend name")
	}
}

// TestCredentialsEnvVarPriority tests that env vars work when keyring is empty
func TestCredentialsEnvVarPriority(t *testing.T) {
	// Save and restore env vars (including token which takes priority)
	origUser := os.Getenv("TODOAT_TODOIST_USERNAME")
	origPass := os.Getenv("TODOAT_TODOIST_PASSWORD")
	origToken := os.Getenv("TODOAT_TODOIST_TOKEN")
	defer func() {
		_ = os.Setenv("TODOAT_TODOIST_USERNAME", origUser)
		_ = os.Setenv("TODOAT_TODOIST_PASSWORD", origPass)
		_ = os.Setenv("TODOAT_TODOIST_TOKEN", origToken)
	}()

	// Clear token so password env var is used instead
	_ = os.Unsetenv("TODOAT_TODOIST_TOKEN")

	// Set environment variables for a different backend
	_ = os.Setenv("TODOAT_TODOIST_USERNAME", "todoistuser")
	_ = os.Setenv("TODOAT_TODOIST_PASSWORD", "todoistpass")

	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	// Get should return from environment
	info, err := manager.Get(context.Background(), "todoist", "todoistuser")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if info.Source != SourceEnvironment {
		t.Errorf("Expected source %s, got %s", SourceEnvironment, info.Source)
	}
	if info.Password != "todoistpass" {
		t.Errorf("Expected password 'todoistpass', got '%s'", info.Password)
	}
}

// TestCredentialsTodoistTokenEnvVar tests that TODOAT_TODOIST_TOKEN env var is detected
// This is a regression test for issue #002 - TODOAT_TODOIST_TOKEN not detected
func TestCredentialsTodoistTokenEnvVar(t *testing.T) {
	// Save and restore env var
	origToken := os.Getenv("TODOAT_TODOIST_TOKEN")
	defer func() {
		_ = os.Setenv("TODOAT_TODOIST_TOKEN", origToken)
	}()

	// Set the token environment variable
	_ = os.Setenv("TODOAT_TODOIST_TOKEN", "test-api-token-12345")

	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	// Get should detect the token from environment
	// For Todoist, username can be empty since it uses a token
	info, err := manager.Get(context.Background(), "todoist", "")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !info.Found {
		t.Error("Expected credentials to be found from TODOAT_TODOIST_TOKEN")
	}
	if info.Source != SourceEnvironment {
		t.Errorf("Expected source %s, got %s", SourceEnvironment, info.Source)
	}
	if info.Password != "test-api-token-12345" {
		t.Errorf("Expected token 'test-api-token-12345', got '%s'", info.Password)
	}
}
