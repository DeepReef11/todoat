package config

import (
	"testing"
)

// TestCustomBackendNaming tests that custom backend names (e.g., nextcloud-test) are supported.
// This test reproduces issue 002 - Config Should Support Custom Backend Naming.
// Users should be able to have multiple instances of the same backend type with custom names
// like "nextcloud-test", "nextcloud-prod", etc.
func TestCustomBackendNaming(t *testing.T) {
	// Create config YAML with custom-named backends
	configYAML := `
backends:
  sqlite:
    type: sqlite
    enabled: true
  nextcloud:
    type: nextcloud
    enabled: true
    host: "prod.example.com"
    username: "admin"
  nextcloud-test:
    type: nextcloud
    enabled: true
    host: "localhost:8080"
    username: "testadmin"
default_backend: nextcloud-test
no_prompt: true
`

	// Test that LoadRaw can parse the config with custom backend names
	cfg, raw, err := LoadRaw([]byte(configYAML))
	if err != nil {
		t.Fatalf("LoadRaw failed: %v", err)
	}
	if cfg == nil || raw == nil {
		t.Fatal("LoadRaw returned nil config or raw map")
	}

	// Check that custom backend is accessible in raw map
	backends, ok := raw["backends"].(map[string]interface{})
	if !ok {
		t.Fatal("backends not found in raw map")
	}

	// Check that nextcloud-test exists in the backends
	_, hasCustomBackend := backends["nextcloud-test"]
	if !hasCustomBackend {
		t.Error("custom backend 'nextcloud-test' not found in raw config")
	}

	// Check that the default_backend can be set to custom name
	if defaultBackend, ok := raw["default_backend"].(string); !ok || defaultBackend != "nextcloud-test" {
		t.Errorf("expected default_backend to be 'nextcloud-test', got %v", raw["default_backend"])
	}
}

// TestGetBackendConfig tests retrieval of backend configuration by name.
func TestGetBackendConfig(t *testing.T) {
	configYAML := `
backends:
  sqlite:
    type: sqlite
    enabled: true
  nextcloud-test:
    type: nextcloud
    enabled: true
    host: "localhost:8080"
    username: "testadmin"
default_backend: nextcloud-test
`

	_, raw, err := LoadRaw([]byte(configYAML))
	if err != nil {
		t.Fatalf("LoadRaw failed: %v", err)
	}

	// Retrieve the backend config for nextcloud-test
	backendCfg, backendType, err := GetBackendConfig(raw, "nextcloud-test")
	if err != nil {
		t.Fatalf("GetBackendConfig failed: %v", err)
	}

	// Verify type is "nextcloud"
	if backendType != "nextcloud" {
		t.Errorf("expected backend type 'nextcloud', got '%s'", backendType)
	}

	// Verify host setting
	host, _ := backendCfg["host"].(string)
	if host != "localhost:8080" {
		t.Errorf("expected host 'localhost:8080', got '%s'", host)
	}

	// Verify username setting
	username, _ := backendCfg["username"].(string)
	if username != "testadmin" {
		t.Errorf("expected username 'testadmin', got '%s'", username)
	}
}

// TestGetBackendConfigImplicitType tests that backends without explicit type field
// use the backend name as the type (for backward compatibility).
func TestGetBackendConfigImplicitType(t *testing.T) {
	configYAML := `
backends:
  sqlite:
    enabled: true
  nextcloud:
    enabled: true
    host: "example.com"
`

	_, raw, err := LoadRaw([]byte(configYAML))
	if err != nil {
		t.Fatalf("LoadRaw failed: %v", err)
	}

	// For "sqlite" without explicit type, type should default to "sqlite"
	_, sqliteType, err := GetBackendConfig(raw, "sqlite")
	if err != nil {
		t.Fatalf("GetBackendConfig for sqlite failed: %v", err)
	}
	if sqliteType != "sqlite" {
		t.Errorf("expected sqlite type to default to 'sqlite', got '%s'", sqliteType)
	}

	// For "nextcloud" without explicit type, type should default to "nextcloud"
	_, nextcloudType, err := GetBackendConfig(raw, "nextcloud")
	if err != nil {
		t.Fatalf("GetBackendConfig for nextcloud failed: %v", err)
	}
	if nextcloudType != "nextcloud" {
		t.Errorf("expected nextcloud type to default to 'nextcloud', got '%s'", nextcloudType)
	}
}

// TestGetBackendConfigNotFound tests error handling for unknown backend names.
func TestGetBackendConfigNotFound(t *testing.T) {
	configYAML := `
backends:
  sqlite:
    enabled: true
`

	_, raw, err := LoadRaw([]byte(configYAML))
	if err != nil {
		t.Fatalf("LoadRaw failed: %v", err)
	}

	_, _, err = GetBackendConfig(raw, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent backend, got nil")
	}
}
