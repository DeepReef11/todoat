package credentials

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// TestCredentialsSetCLI tests the CLI command: todoat credentials set nextcloud myuser --prompt
func TestCredentialsSetCLI(t *testing.T) {
	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	// Simulate CLI input with password
	stdin := bytes.NewBufferString("mysecretpass\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	handler := NewCLIHandler(manager, stdin, stdout, stderr)
	err := handler.Set("nextcloud", "myuser", true) // --prompt mode

	if err != nil {
		t.Fatalf("Set command failed: %v", err)
	}

	// Check output
	output := stdout.String()
	if !strings.Contains(output, "Credentials stored") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify credentials were stored
	info, _ := manager.Get(context.TODO(), "nextcloud", "myuser")
	if !info.Found {
		t.Error("Credentials should be stored")
	}
	if info.Password != "mysecretpass" {
		t.Errorf("Expected password 'mysecretpass', got '%s'", info.Password)
	}
}

// TestCredentialsGetCLI tests the CLI command: todoat credentials get nextcloud myuser
func TestCredentialsGetCLI(t *testing.T) {
	mockKeyring := NewMockKeyring()
	_ = mockKeyring.Set("todoat-nextcloud", "myuser", "storedpass")
	manager := NewManager(WithKeyring(mockKeyring))

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	handler := NewCLIHandler(manager, nil, stdout, stderr)
	err := handler.Get("nextcloud", "myuser", false) // not JSON

	if err != nil {
		t.Fatalf("Get command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Source: keyring") {
		t.Errorf("Expected source info, got: %s", output)
	}
	if !strings.Contains(output, "myuser") {
		t.Errorf("Expected username in output, got: %s", output)
	}
	// Password should be masked
	if strings.Contains(output, "storedpass") {
		t.Error("Password should not appear in output")
	}
	if !strings.Contains(output, "********") {
		t.Errorf("Expected masked password in output, got: %s", output)
	}
}

// TestCredentialsGetJSONCLI tests the CLI command: todoat --json credentials get nextcloud myuser
func TestCredentialsGetJSONCLI(t *testing.T) {
	mockKeyring := NewMockKeyring()
	_ = mockKeyring.Set("todoat-nextcloud", "myuser", "storedpass")
	manager := NewManager(WithKeyring(mockKeyring))

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	handler := NewCLIHandler(manager, nil, stdout, stderr)
	err := handler.Get("nextcloud", "myuser", true) // JSON output

	if err != nil {
		t.Fatalf("Get command failed: %v", err)
	}

	// Parse JSON output
	var response struct {
		Backend  string `json:"backend"`
		Username string `json:"username"`
		Source   string `json:"source"`
		Found    bool   `json:"found"`
	}
	err = json.Unmarshal(stdout.Bytes(), &response)
	if err != nil {
		t.Fatalf("JSON parse failed: %v, output was: %s", err, stdout.String())
	}

	if response.Backend != "nextcloud" {
		t.Errorf("Expected backend 'nextcloud', got '%s'", response.Backend)
	}
	if response.Source != "keyring" {
		t.Errorf("Expected source 'keyring', got '%s'", response.Source)
	}
	if !response.Found {
		t.Error("Expected found to be true")
	}
}

// TestCredentialsDeleteCLI tests the CLI command: todoat credentials delete nextcloud myuser
func TestCredentialsDeleteCLI(t *testing.T) {
	mockKeyring := NewMockKeyring()
	_ = mockKeyring.Set("todoat-nextcloud", "myuser", "toberemoved")
	manager := NewManager(WithKeyring(mockKeyring))

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	handler := NewCLIHandler(manager, nil, stdout, stderr)
	err := handler.Delete("nextcloud", "myuser")

	if err != nil {
		t.Fatalf("Delete command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Credentials removed") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify credentials are gone
	info, _ := manager.Get(context.TODO(), "nextcloud", "myuser")
	if info.Found && info.Source == SourceKeyring {
		t.Error("Credentials should be removed from keyring")
	}
}

// TestCredentialsNotFoundCLI tests error handling when credentials not found
func TestCredentialsNotFoundCLI(t *testing.T) {
	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	handler := NewCLIHandler(manager, nil, stdout, stderr)
	err := handler.Get("nonexistent", "user", false)

	// Get should not fail but show not found message
	if err != nil {
		t.Fatalf("Get should not fail for not found, got: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "No credentials found") {
		t.Errorf("Expected not found message, got: %s", output)
	}
}

// TestCredentialsListCLI tests the CLI command: todoat credentials list
func TestCredentialsListCLI(t *testing.T) {
	mockKeyring := NewMockKeyring()
	_ = mockKeyring.Set("todoat-nextcloud", "ncuser", "pass1")
	manager := NewManager(WithKeyring(mockKeyring))

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	backends := []BackendConfig{
		{Name: "nextcloud", Username: "ncuser"},
		{Name: "todoist", Username: "todoistuser"},
	}

	handler := NewCLIHandler(manager, nil, stdout, stderr)
	err := handler.List(backends, false)

	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "nextcloud") {
		t.Errorf("Expected nextcloud in output, got: %s", output)
	}
	if !strings.Contains(output, "todoist") {
		t.Errorf("Expected todoist in output, got: %s", output)
	}
}

// TestCredentialsListJSONCLI tests the CLI command: todoat --json credentials list
func TestCredentialsListJSONCLI(t *testing.T) {
	mockKeyring := NewMockKeyring()
	_ = mockKeyring.Set("todoat-nextcloud", "ncuser", "pass1")
	manager := NewManager(WithKeyring(mockKeyring))

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	backends := []BackendConfig{
		{Name: "nextcloud", Username: "ncuser"},
		{Name: "todoist", Username: "todoistuser"},
	}

	handler := NewCLIHandler(manager, nil, stdout, stderr)
	err := handler.List(backends, true) // JSON output

	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	// Parse JSON output
	var response []struct {
		Backend        string `json:"backend"`
		Username       string `json:"username"`
		HasCredentials bool   `json:"has_credentials"`
		Source         string `json:"source,omitempty"`
	}
	err = json.Unmarshal(stdout.Bytes(), &response)
	if err != nil {
		t.Fatalf("JSON parse failed: %v, output was: %s", err, stdout.String())
	}

	if len(response) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(response))
	}

	// Find nextcloud entry
	var ncEntry, todoistEntry *struct {
		Backend        string `json:"backend"`
		Username       string `json:"username"`
		HasCredentials bool   `json:"has_credentials"`
		Source         string `json:"source,omitempty"`
	}
	for i := range response {
		if response[i].Backend == "nextcloud" {
			ncEntry = &response[i]
		}
		if response[i].Backend == "todoist" {
			todoistEntry = &response[i]
		}
	}

	if ncEntry == nil {
		t.Fatal("Expected nextcloud in response")
	}
	if !ncEntry.HasCredentials {
		t.Error("Expected nextcloud to have credentials")
	}

	if todoistEntry == nil {
		t.Fatal("Expected todoist in response")
	}
	if todoistEntry.HasCredentials {
		t.Error("Expected todoist to NOT have credentials")
	}
}

// TestCredentialUpdate tests the CLI command: todoat credentials update nextcloud myuser --prompt
func TestCredentialUpdate(t *testing.T) {
	mockKeyring := NewMockKeyring()
	// Pre-set existing credential
	_ = mockKeyring.Set("todoat-nextcloud", "myuser", "oldpassword")
	manager := NewManager(WithKeyring(mockKeyring))

	// Simulate CLI input with new password
	stdin := bytes.NewBufferString("newpassword\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	handler := NewCLIHandler(manager, stdin, stdout, stderr)
	err := handler.Update("nextcloud", "myuser", true, false) // --prompt, no --verify

	if err != nil {
		t.Fatalf("Update command failed: %v", err)
	}

	// Check success message
	output := stdout.String()
	if !strings.Contains(output, "Credential updated") {
		t.Errorf("Expected success message, got: %s", output)
	}
	if !strings.Contains(output, "nextcloud/myuser") {
		t.Errorf("Expected backend/user in output, got: %s", output)
	}

	// Verify credential was updated
	info, _ := manager.Get(context.TODO(), "nextcloud", "myuser")
	if !info.Found {
		t.Error("Credentials should still exist")
	}
	if info.Password != "newpassword" {
		t.Errorf("Expected password 'newpassword', got '%s'", info.Password)
	}
}

// TestCredentialUpdateNonExistent tests that update on non-existent credential shows appropriate error
func TestCredentialUpdateNonExistent(t *testing.T) {
	mockKeyring := NewMockKeyring()
	// No pre-existing credentials
	manager := NewManager(WithKeyring(mockKeyring))

	stdin := bytes.NewBufferString("newpassword\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	handler := NewCLIHandler(manager, stdin, stdout, stderr)
	err := handler.Update("nextcloud", "nonexistent", true, false)

	// Should fail
	if err == nil {
		t.Fatal("Expected error for non-existent credential")
	}

	// Error message should be helpful
	errMsg := err.Error()
	if !strings.Contains(errMsg, "no credential found") {
		t.Errorf("Expected 'no credential found' in error, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "credentials set") {
		t.Errorf("Expected suggestion to use 'credentials set', got: %s", errMsg)
	}
}

// TestCredentialUpdateVerify tests that updated credential can be retrieved and verified
func TestCredentialUpdateVerify(t *testing.T) {
	mockKeyring := NewMockKeyring()
	// Pre-set existing credential
	_ = mockKeyring.Set("todoat-nextcloud", "myuser", "oldpassword")
	manager := NewManager(WithKeyring(mockKeyring))

	stdin := bytes.NewBufferString("verifiedpassword\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	handler := NewCLIHandler(manager, stdin, stdout, stderr)
	err := handler.Update("nextcloud", "myuser", true, false)

	if err != nil {
		t.Fatalf("Update command failed: %v", err)
	}

	// Now retrieve it to verify
	info, err := manager.Get(context.TODO(), "nextcloud", "myuser")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !info.Found {
		t.Error("Credential should be found after update")
	}
	if info.Password != "verifiedpassword" {
		t.Errorf("Expected password 'verifiedpassword', got '%s'", info.Password)
	}
	if info.Source != SourceKeyring {
		t.Errorf("Expected source 'keyring', got '%s'", info.Source)
	}
}

// TestCredentialUpdateNoChange tests that update with same password succeeds (idempotent)
func TestCredentialUpdateNoChange(t *testing.T) {
	mockKeyring := NewMockKeyring()
	// Pre-set existing credential
	_ = mockKeyring.Set("todoat-nextcloud", "myuser", "samepassword")
	manager := NewManager(WithKeyring(mockKeyring))

	// Update with same password
	stdin := bytes.NewBufferString("samepassword\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	handler := NewCLIHandler(manager, stdin, stdout, stderr)
	err := handler.Update("nextcloud", "myuser", true, false)

	if err != nil {
		t.Fatalf("Update with same password should succeed: %v", err)
	}

	// Verify credential still exists with same value
	info, _ := manager.Get(context.TODO(), "nextcloud", "myuser")
	if !info.Found {
		t.Error("Credential should exist")
	}
	if info.Password != "samepassword" {
		t.Errorf("Expected password 'samepassword', got '%s'", info.Password)
	}
}

// TestCredentialsSetKeyringNotAvailableCLI tests that helpful error message is shown
// when keyring is not available. This is a regression test for issue #003.
func TestCredentialsSetKeyringNotAvailableCLI(t *testing.T) {
	// Use the system keyring which always returns ErrKeyringNotAvailable
	manager := NewManager() // No mock, uses real systemKeyring

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	stdin := strings.NewReader("test-password\n")

	handler := NewCLIHandler(manager, stdin, stdout, stderr)
	err := handler.Set("todoist", "apiuser", true)

	// Should fail with keyring not available
	if err == nil {
		t.Fatal("Expected error when keyring not available")
	}

	// Error message should provide helpful guidance about environment variables
	errMsg := err.Error()
	if !strings.Contains(errMsg, "environment variable") {
		t.Errorf("Expected error to mention environment variables, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "TODOAT_TODOIST_TOKEN") {
		t.Errorf("Expected error to mention TODOAT_TODOIST_TOKEN, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "TODOAT_TODOIST_PASSWORD") {
		t.Errorf("Expected error to mention TODOAT_TODOIST_PASSWORD, got: %s", errMsg)
	}
}
