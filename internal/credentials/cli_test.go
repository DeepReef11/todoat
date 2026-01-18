package credentials

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// TestCLICredentialsSet tests the CLI command: todoat credentials set nextcloud myuser --prompt
func TestCLICredentialsSet(t *testing.T) {
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

// TestCLICredentialsGet tests the CLI command: todoat credentials get nextcloud myuser
func TestCLICredentialsGet(t *testing.T) {
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

// TestCLICredentialsGetJSON tests the CLI command: todoat --json credentials get nextcloud myuser
func TestCLICredentialsGetJSON(t *testing.T) {
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

// TestCLICredentialsDelete tests the CLI command: todoat credentials delete nextcloud myuser
func TestCLICredentialsDelete(t *testing.T) {
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

// TestCLICredentialsNotFound tests error handling when credentials not found
func TestCLICredentialsNotFound(t *testing.T) {
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

// TestCLICredentialsList tests the CLI command: todoat credentials list
func TestCLICredentialsList(t *testing.T) {
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

// TestCLICredentialsListJSON tests the CLI command: todoat --json credentials list
func TestCLICredentialsListJSON(t *testing.T) {
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
