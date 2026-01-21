package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// Sample Config Tests (055-sample-config-file)
// =============================================================================

// TestSampleConfigEmbedded verifies config.sample.yaml is embedded in binary via go:embed
func TestSampleConfigEmbedded(t *testing.T) {
	// GetSampleConfig should return the embedded sample config content
	content := GetSampleConfig()

	if content == "" {
		t.Error("expected embedded sample config to have content, got empty string")
	}

	// Verify it's valid YAML by checking for expected structure
	if !strings.Contains(content, "backends:") {
		t.Error("expected sample config to contain 'backends:' section")
	}

	if !strings.Contains(content, "default_backend:") {
		t.Error("expected sample config to contain 'default_backend:' key")
	}
}

// TestSampleConfigCopyOnFirstRun verifies first run copies sample to ~/.config/todoat/config.yaml
func TestSampleConfigCopyOnFirstRun(t *testing.T) {
	// Create temporary XDG directories
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	dataDir := filepath.Join(tmpDir, "data")

	// Set XDG environment variables
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("XDG_DATA_HOME", dataDir)
	t.Setenv("HOME", tmpDir)

	// Load config (should auto-create from sample)
	_, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify config file was created
	configPath := filepath.Join(configDir, "todoat", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read created config file: %v", err)
	}

	content := string(data)

	// Verify it contains sample config content (comments and structure)
	// The created config should have YAML comments explaining options
	if !strings.Contains(content, "#") {
		t.Error("expected created config to contain YAML comments from sample")
	}

	// Verify it has the basic structure
	if !strings.Contains(content, "backends:") {
		t.Error("expected created config to contain 'backends:' section")
	}
}

// TestSampleConfigAllBackends verifies sample includes examples for all backend types
func TestSampleConfigAllBackends(t *testing.T) {
	content := GetSampleConfig()

	// Check for SQLite backend
	if !strings.Contains(content, "sqlite") {
		t.Error("expected sample config to include sqlite backend example")
	}

	// Check for Nextcloud backend
	if !strings.Contains(content, "nextcloud") {
		t.Error("expected sample config to include nextcloud backend example")
	}

	// Check for Todoist backend
	if !strings.Contains(content, "todoist") {
		t.Error("expected sample config to include todoist backend example")
	}

	// Check for Git backend
	if !strings.Contains(content, "git") {
		t.Error("expected sample config to include git backend example")
	}

	// Check for File backend
	if !strings.Contains(content, "file") {
		t.Error("expected sample config to include file backend example")
	}
}

// TestSampleConfigComments verifies sample contains inline YAML comments explaining each option
func TestSampleConfigComments(t *testing.T) {
	content := GetSampleConfig()

	// Count comment lines (lines starting with # after trimming)
	lines := strings.Split(content, "\n")
	commentCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			commentCount++
		}
	}

	// Sample config should have substantial comments (at least 10 comment lines)
	if commentCount < 10 {
		t.Errorf("expected sample config to have at least 10 comment lines for documentation, got %d", commentCount)
	}

	// Verify specific documentation comments exist
	requiredComments := []string{
		"todoat",    // Header comment mentioning the app
		"Backend",   // Backend section documentation
		"SQLite",    // SQLite documentation
		"Nextcloud", // Nextcloud documentation
		"Todoist",   // Todoist documentation
		"sync",      // Sync documentation
	}

	for _, keyword := range requiredComments {
		if !strings.Contains(strings.ToLower(content), strings.ToLower(keyword)) {
			t.Errorf("expected sample config to contain documentation about %q", keyword)
		}
	}
}

// TestSampleConfigNextcloudTLSOptions verifies sample includes Nextcloud TLS options
func TestSampleConfigNextcloudTLSOptions(t *testing.T) {
	content := GetSampleConfig()

	// Check for TLS options
	if !strings.Contains(content, "insecure_skip_verify") {
		t.Error("expected sample config to include insecure_skip_verify option for Nextcloud")
	}
}

// TestSampleConfigTodoistToken verifies sample includes Todoist token placeholder
func TestSampleConfigTodoistToken(t *testing.T) {
	content := GetSampleConfig()

	// Check for token reference
	if !strings.Contains(strings.ToLower(content), "token") {
		t.Error("expected sample config to include token placeholder/documentation for Todoist")
	}
}

// TestSampleConfigSyncSettings verifies sample includes sync configuration
func TestSampleConfigSyncSettings(t *testing.T) {
	content := GetSampleConfig()

	// Check for sync section
	if !strings.Contains(content, "sync:") {
		t.Error("expected sample config to include sync: section")
	}

	// Check for conflict resolution options
	if !strings.Contains(content, "conflict") {
		t.Error("expected sample config to document conflict resolution options")
	}

	// Check for offline mode options
	if !strings.Contains(content, "offline") {
		t.Error("expected sample config to document offline mode options")
	}
}

// TestSampleConfigViewDefaults verifies sample includes view defaults configuration
func TestSampleConfigViewDefaults(t *testing.T) {
	content := GetSampleConfig()

	// Check for default_view option
	if !strings.Contains(content, "default_view") {
		t.Error("expected sample config to include default_view option")
	}
}

// TestSampleConfigCredentialsPatterns verifies sample shows keyring/env var patterns
func TestSampleConfigCredentialsPatterns(t *testing.T) {
	content := GetSampleConfig()

	// Check for keyring mention
	if !strings.Contains(strings.ToLower(content), "keyring") {
		t.Error("expected sample config to mention keyring-based credential storage")
	}

	// Check for environment variable pattern
	if !strings.Contains(content, "TODOAT_") || !strings.Contains(strings.ToLower(content), "env") {
		t.Error("expected sample config to mention environment variable patterns (TODOAT_*)")
	}
}

// =============================================================================
// Sample Config Options Tests (065-sample-config-missing-options)
// =============================================================================

// TestSampleConfigContainsAllOptions verifies sample config includes all documented options
func TestSampleConfigContainsAllOptions(t *testing.T) {
	content := GetSampleConfig()

	// Test for allow_http option (065 requirement)
	// Must be present with explanation for HTTP connections
	t.Run("allow_http option", func(t *testing.T) {
		if !strings.Contains(content, "allow_http") {
			t.Error("expected sample config to include allow_http option")
		}
		// Check it has an explanation comment
		if !strings.Contains(strings.ToLower(content), "http") {
			t.Error("expected sample config to explain allow_http option")
		}
	})

	// Test for auto_detect_backend option (065 requirement)
	// Must be present with enabled: false default
	t.Run("auto_detect_backend option", func(t *testing.T) {
		if !strings.Contains(content, "auto_detect_backend") {
			t.Error("expected sample config to include auto_detect_backend option")
		}
		// Verify default is false (simple mode)
		if !strings.Contains(content, "auto_detect_backend: false") {
			t.Error("expected sample config to have auto_detect_backend: false as default")
		}
	})

	// Test for backend_priority configuration (065 requirement)
	t.Run("backend_priority configuration", func(t *testing.T) {
		if !strings.Contains(content, "backend_priority") {
			t.Error("expected sample config to include backend_priority configuration")
		}
	})

	// Test for suppress_http_warning option (related to allow_http)
	t.Run("suppress_http_warning option", func(t *testing.T) {
		if !strings.Contains(content, "suppress_http_warning") {
			t.Error("expected sample config to include suppress_http_warning option")
		}
	})

	// Verify extended auto_detect_backend mode is documented
	t.Run("extended auto_detect_backend mode", func(t *testing.T) {
		// Should show the object format with enabled: true
		if !strings.Contains(content, "enabled: true") {
			t.Error("expected sample config to show extended auto_detect_backend mode with enabled: true")
		}
	})
}
