package config

import (
	"os"
	"path/filepath"
	"testing"
)

// =============================================================================
// Configuration System Tests (010-configuration)
// =============================================================================

// TestConfigAutoCreate verifies first run creates config file at XDG path with defaults
func TestConfigAutoCreate(t *testing.T) {
	// Create temporary XDG directories
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	dataDir := filepath.Join(tmpDir, "data")

	// Set XDG environment variables
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("XDG_DATA_HOME", dataDir)
	t.Setenv("HOME", tmpDir)

	// Load config (should auto-create)
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify config file was created
	configPath := filepath.Join(configDir, "todoat", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config file not created at %s", configPath)
	}

	// Verify defaults were set
	if cfg.DefaultBackend != "sqlite" {
		t.Errorf("expected DefaultBackend = 'sqlite', got %q", cfg.DefaultBackend)
	}
	if cfg.NoPrompt != false {
		t.Errorf("expected NoPrompt = false, got %v", cfg.NoPrompt)
	}
	if cfg.OutputFormat != "text" {
		t.Errorf("expected OutputFormat = 'text', got %q", cfg.OutputFormat)
	}
}

// TestConfigCustomPath verifies --config /path/to/config.yaml uses specified config
func TestConfigCustomPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a custom config file
	customConfigPath := filepath.Join(tmpDir, "custom-config.yaml")
	customConfig := `
backends:
  sqlite:
    enabled: true
    path: "/custom/path/tasks.db"
default_backend: sqlite
no_prompt: true
output_format: json
`
	if err := os.WriteFile(customConfigPath, []byte(customConfig), 0644); err != nil {
		t.Fatalf("failed to write custom config: %v", err)
	}

	// Load from custom path
	cfg, err := Load(customConfigPath)
	if err != nil {
		t.Fatalf("Load(%q) error = %v", customConfigPath, err)
	}

	// Verify custom values
	if cfg.NoPrompt != true {
		t.Errorf("expected NoPrompt = true, got %v", cfg.NoPrompt)
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("expected OutputFormat = 'json', got %q", cfg.OutputFormat)
	}
	if cfg.Backends.SQLite.Path != "/custom/path/tasks.db" {
		t.Errorf("expected SQLite path = '/custom/path/tasks.db', got %q", cfg.Backends.SQLite.Path)
	}
}

// TestConfigDatabasePath verifies database path from config is used for SQLite backend
func TestConfigDatabasePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config with custom database path
	dbPath := filepath.Join(tmpDir, "custom.db")
	customConfig := `
backends:
  sqlite:
    enabled: true
    path: "` + dbPath + `"
default_backend: sqlite
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(customConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify database path is correctly read
	if cfg.Backends.SQLite.Path != dbPath {
		t.Errorf("expected SQLite path = %q, got %q", dbPath, cfg.Backends.SQLite.Path)
	}

	// Verify GetDatabasePath returns the correct path
	gotPath := cfg.GetDatabasePath()
	if gotPath != dbPath {
		t.Errorf("GetDatabasePath() = %q, want %q", gotPath, dbPath)
	}
}

// TestConfigNoPromptDefault verifies no_prompt: true in config enables no-prompt mode globally
func TestConfigNoPromptDefault(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config with no_prompt: true
	customConfig := `
backends:
  sqlite:
    enabled: true
default_backend: sqlite
no_prompt: true
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(customConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.NoPrompt != true {
		t.Errorf("expected NoPrompt = true, got %v", cfg.NoPrompt)
	}
}

// TestConfigFlagOverride verifies CLI flags override config values
func TestConfigFlagOverride(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config with no_prompt: false
	customConfig := `
backends:
  sqlite:
    enabled: true
default_backend: sqlite
no_prompt: false
output_format: text
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(customConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify config values
	if cfg.NoPrompt != false {
		t.Errorf("expected NoPrompt = false from config, got %v", cfg.NoPrompt)
	}

	// Apply flag override
	cfg.ApplyFlags(true, "json")

	// Verify flag overrides config
	if cfg.NoPrompt != true {
		t.Errorf("expected NoPrompt = true after flag override, got %v", cfg.NoPrompt)
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("expected OutputFormat = 'json' after flag override, got %q", cfg.OutputFormat)
	}
}

// TestConfigInvalid verifies invalid YAML returns clear error message
func TestConfigInvalid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid YAML config
	invalidConfig := `
backends:
  sqlite:
    enabled: [invalid yaml structure
default_backend: sqlite
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

// TestConfigMissingBackend verifies missing backend config returns helpful error
func TestConfigMissingBackend(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config with no backends configured
	noBackendConfig := `
default_backend: sqlite
no_prompt: false
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(noBackendConfig), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Validate should return error for missing backend
	err = cfg.Validate()
	if err == nil {
		t.Error("expected validation error for missing backend, got nil")
	}
}

// =============================================================================
// Unit Tests for XDG Path Handling
// =============================================================================

// TestXDGPathExpansion verifies XDG path expansion works on Linux/macOS/Windows
func TestXDGPathExpansion(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		envVar     string
		envValue   string
		wantPrefix string
	}{
		{
			name:       "XDG_CONFIG_HOME set",
			envVar:     "XDG_CONFIG_HOME",
			envValue:   filepath.Join(tmpDir, "xdg-config"),
			wantPrefix: filepath.Join(tmpDir, "xdg-config"),
		},
		{
			name:       "XDG_DATA_HOME set",
			envVar:     "XDG_DATA_HOME",
			envValue:   filepath.Join(tmpDir, "xdg-data"),
			wantPrefix: filepath.Join(tmpDir, "xdg-data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envVar, tt.envValue)

			var got string
			if tt.envVar == "XDG_CONFIG_HOME" {
				got = GetConfigDir()
			} else {
				got = GetDataDir()
			}

			if got != filepath.Join(tt.wantPrefix, "todoat") {
				t.Errorf("got %q, want prefix %q/todoat", got, tt.wantPrefix)
			}
		})
	}
}

// TestPathExpansionTilde verifies ~ expansion to home directory
func TestPathExpansionTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("could not get home directory: %v", err)
	}

	tests := []struct {
		input string
		want  string
	}{
		{"~/foo/bar", filepath.Join(home, "foo", "bar")},
		{"~/.config/todoat", filepath.Join(home, ".config", "todoat")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ExpandPath(tt.input)
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestPathExpansionEnvVars verifies $HOME and $XDG_* expansion
func TestPathExpansionEnvVars(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "xdg-data"))

	tests := []struct {
		input string
		want  string
	}{
		{"$HOME/foo", filepath.Join(tmpDir, "foo")},
		{"$XDG_DATA_HOME/todoat", filepath.Join(tmpDir, "xdg-data", "todoat")},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ExpandPath(tt.input)
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestConfigValidation verifies config validation catches invalid values
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Backends: BackendsConfig{
					SQLite: SQLiteConfig{
						Enabled: true,
						Path:    "/path/to/db",
					},
				},
				DefaultBackend: "sqlite",
				OutputFormat:   "text",
			},
			wantErr: false,
		},
		{
			name: "invalid output format",
			config: &Config{
				Backends: BackendsConfig{
					SQLite: SQLiteConfig{
						Enabled: true,
						Path:    "/path/to/db",
					},
				},
				DefaultBackend: "sqlite",
				OutputFormat:   "invalid",
			},
			wantErr: true,
		},
		{
			name: "unknown default backend",
			config: &Config{
				Backends: BackendsConfig{
					SQLite: SQLiteConfig{
						Enabled: true,
						Path:    "/path/to/db",
					},
				},
				DefaultBackend: "unknown",
				OutputFormat:   "text",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDefaultConfig verifies default config values
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DefaultBackend != "sqlite" {
		t.Errorf("DefaultBackend = %q, want 'sqlite'", cfg.DefaultBackend)
	}
	if cfg.NoPrompt != false {
		t.Errorf("NoPrompt = %v, want false", cfg.NoPrompt)
	}
	if cfg.OutputFormat != "text" {
		t.Errorf("OutputFormat = %q, want 'text'", cfg.OutputFormat)
	}
	if !cfg.Backends.SQLite.Enabled {
		t.Error("SQLite backend should be enabled by default")
	}
}

// TestAnalyticsEnabledByDefault verifies that analytics is enabled by default for new installations
// as per decision FEAT-008 in docs/decisions/question-log.md
func TestAnalyticsEnabledByDefault(t *testing.T) {
	cfg := DefaultConfig()

	// Per FEAT-008: Analytics should be enabled by default with clear notice
	if !cfg.Analytics.Enabled {
		t.Error("Analytics.Enabled should be true by default (FEAT-008)")
	}
}

// =============================================================================
// Tests for Issue 007: Additional Coverage
// =============================================================================

// TestIsSyncEnabled verifies sync status detection
func TestIsSyncEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{
			name:     "sync disabled by default",
			config:   DefaultConfig(),
			expected: false,
		},
		{
			name: "sync enabled",
			config: &Config{
				Sync: SyncConfig{Enabled: true},
			},
			expected: true,
		},
		{
			name: "sync explicitly disabled",
			config: &Config{
				Sync: SyncConfig{Enabled: false},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsSyncEnabled()
			if got != tt.expected {
				t.Errorf("IsSyncEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsAutoDetectEnabled verifies auto-detect backend setting
func TestIsAutoDetectEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{
			name:     "auto-detect disabled by default",
			config:   DefaultConfig(),
			expected: false,
		},
		{
			name: "auto-detect enabled",
			config: &Config{
				AutoDetectBackend: true,
			},
			expected: true,
		},
		{
			name: "auto-detect explicitly disabled",
			config: &Config{
				AutoDetectBackend: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsAutoDetectEnabled()
			if got != tt.expected {
				t.Errorf("IsAutoDetectEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestGetCacheDir verifies cache directory XDG path
func TestGetCacheDir(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("XDG_CACHE_HOME set", func(t *testing.T) {
		cacheDir := filepath.Join(tmpDir, "xdg-cache")
		t.Setenv("XDG_CACHE_HOME", cacheDir)

		got := GetCacheDir()
		want := filepath.Join(cacheDir, "todoat")
		if got != want {
			t.Errorf("GetCacheDir() = %q, want %q", got, want)
		}
	})

	t.Run("XDG_CACHE_HOME unset falls back to home", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", "")
		t.Setenv("HOME", tmpDir)

		got := GetCacheDir()
		want := filepath.Join(tmpDir, ".cache", "todoat")
		if got != want {
			t.Errorf("GetCacheDir() = %q, want %q", got, want)
		}
	})
}

// TestLoadFromPath verifies loading config from specific path
func TestLoadFromPath(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("empty path returns error", func(t *testing.T) {
		_, err := LoadFromPath("")
		if err == nil {
			t.Error("expected error for empty path, got nil")
		}
	})

	t.Run("non-existent file returns nil config", func(t *testing.T) {
		cfg, err := LoadFromPath(filepath.Join(tmpDir, "nonexistent.yaml"))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg != nil {
			t.Error("expected nil config for non-existent file")
		}
	})

	t.Run("valid config file", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "valid.yaml")
		content := `
backends:
  sqlite:
    enabled: true
    path: "/tmp/test.db"
default_backend: sqlite
output_format: json
`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		cfg, err := LoadFromPath(configPath)
		if err != nil {
			t.Fatalf("LoadFromPath() error = %v", err)
		}
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if cfg.OutputFormat != "json" {
			t.Errorf("OutputFormat = %q, want 'json'", cfg.OutputFormat)
		}
	})

	t.Run("invalid YAML returns error", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "invalid.yaml")
		content := `invalid: [yaml structure`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		_, err := LoadFromPath(configPath)
		if err == nil {
			t.Error("expected error for invalid YAML")
		}
	})
}

// TestIsBackendConfigured verifies backend configuration check
func TestIsBackendConfigured(t *testing.T) {
	raw := map[string]interface{}{
		"backends": map[string]interface{}{
			"sqlite": map[string]interface{}{
				"enabled": true,
				"path":    "/tmp/test.db",
			},
			"todoist": map[string]interface{}{
				"enabled": false,
			},
		},
	}

	tests := []struct {
		name     string
		backend  string
		expected bool
	}{
		{"sqlite configured", "sqlite", true},
		{"todoist configured", "todoist", true},
		{"nextcloud not configured", "nextcloud", false},
		{"unknown backend", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBackendConfigured(raw, tt.backend)
			if got != tt.expected {
				t.Errorf("IsBackendConfigured(raw, %q) = %v, want %v", tt.backend, got, tt.expected)
			}
		})
	}

	t.Run("nil backends map", func(t *testing.T) {
		emptyRaw := map[string]interface{}{}
		got := IsBackendConfigured(emptyRaw, "sqlite")
		if got != false {
			t.Error("expected false for empty raw map")
		}
	})
}

// TestLoadWithRaw verifies loading config with raw map
func TestLoadWithRaw(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("creates default config if not exists", func(t *testing.T) {
		configDir := filepath.Join(tmpDir, "config-loadwithraw")
		t.Setenv("XDG_CONFIG_HOME", configDir)
		dataDir := filepath.Join(tmpDir, "data-loadwithraw")
		t.Setenv("XDG_DATA_HOME", dataDir)
		t.Setenv("HOME", tmpDir)

		cfg, raw, err := LoadWithRaw("")
		if err != nil {
			t.Fatalf("LoadWithRaw() error = %v", err)
		}
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		// Raw is nil for newly created default config
		if raw != nil {
			t.Error("expected nil raw for newly created default config")
		}
		if cfg.DefaultBackend != "sqlite" {
			t.Errorf("DefaultBackend = %q, want 'sqlite'", cfg.DefaultBackend)
		}
	})

	t.Run("loads existing config with raw", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "existing.yaml")
		content := `
backends:
  sqlite:
    enabled: true
    path: "/tmp/test.db"
  custom_backend:
    type: sqlite
    enabled: true
    path: "/tmp/custom.db"
default_backend: sqlite
output_format: text
`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		cfg, raw, err := LoadWithRaw(configPath)
		if err != nil {
			t.Fatalf("LoadWithRaw() error = %v", err)
		}
		if cfg == nil {
			t.Error("expected non-nil config")
		}
		if raw == nil {
			t.Error("expected non-nil raw map")
		}

		// Verify custom backend is accessible via raw
		if backends, ok := raw["backends"].(map[string]interface{}); ok {
			if _, ok := backends["custom_backend"]; !ok {
				t.Error("expected custom_backend in raw map")
			}
		} else {
			t.Error("expected backends in raw map")
		}
	})
}

// TestGetTrashRetentionDays verifies trash retention configuration
func TestGetTrashRetentionDays(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected int
	}{
		{
			name:     "nil retention returns default 30",
			config:   &Config{},
			expected: 30,
		},
		{
			name: "zero retention (disabled)",
			config: &Config{
				Trash: TrashConfig{RetentionDays: intPtr(0)},
			},
			expected: 0,
		},
		{
			name: "custom retention",
			config: &Config{
				Trash: TrashConfig{RetentionDays: intPtr(7)},
			},
			expected: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetTrashRetentionDays()
			if got != tt.expected {
				t.Errorf("GetTrashRetentionDays() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

// TestValidateBackendNotEnabled verifies validation for disabled backends
func TestValidateBackendNotEnabled(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "todoist default but not enabled",
			config: &Config{
				DefaultBackend: "todoist",
				OutputFormat:   "text",
				Backends: BackendsConfig{
					Todoist: TodoistConfig{Enabled: false},
				},
			},
			wantErr: true,
		},
		{
			name: "nextcloud default but not enabled",
			config: &Config{
				DefaultBackend: "nextcloud",
				OutputFormat:   "text",
				Backends: BackendsConfig{
					Nextcloud: NextcloudConfig{Enabled: false},
				},
			},
			wantErr: true,
		},
		{
			name: "todoist default and enabled",
			config: &Config{
				DefaultBackend: "todoist",
				OutputFormat:   "text",
				Backends: BackendsConfig{
					Todoist: TodoistConfig{Enabled: true},
				},
			},
			wantErr: false,
		},
		{
			name: "nextcloud default and enabled",
			config: &Config{
				DefaultBackend: "nextcloud",
				OutputFormat:   "text",
				Backends: BackendsConfig{
					Nextcloud: NextcloudConfig{Enabled: true},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetConnectivityTimeout verifies connectivity timeout configuration
func TestGetConnectivityTimeout(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name:     "empty timeout returns default",
			config:   &Config{},
			expected: "5s",
		},
		{
			name: "custom timeout",
			config: &Config{
				Sync: SyncConfig{ConnectivityTimeout: "10s"},
			},
			expected: "10s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetConnectivityTimeout()
			if got != tt.expected {
				t.Errorf("GetConnectivityTimeout() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestExpandPathEmpty verifies empty path handling
func TestExpandPathEmpty(t *testing.T) {
	got := ExpandPath("")
	if got != "" {
		t.Errorf("ExpandPath(\"\") = %q, want empty string", got)
	}
}

// =============================================================================
// Tests for Issue 009: Auto-Sync Default When Sync Enabled
// =============================================================================

// TestAutoSyncDefaultsToTrueWhenSyncEnabled verifies that when sync.enabled: true
// and auto_sync_after_operation is not explicitly set, auto-sync should default to true.
// This is Issue #009: Auto-Sync Not Triggering After Task Operations
func TestAutoSyncDefaultsToTrueWhenSyncEnabled(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name                    string
		config                  *Config
		expectedAutoSyncAfterOp bool
	}{
		{
			name: "sync enabled, auto_sync not set - should default to true",
			config: &Config{
				Sync: SyncConfig{
					Enabled:                true,
					AutoSyncAfterOperation: nil, // not set
				},
			},
			expectedAutoSyncAfterOp: true,
		},
		{
			name: "sync enabled, auto_sync explicitly true",
			config: &Config{
				Sync: SyncConfig{
					Enabled:                true,
					AutoSyncAfterOperation: &trueVal,
				},
			},
			expectedAutoSyncAfterOp: true,
		},
		{
			name: "sync enabled, auto_sync explicitly false",
			config: &Config{
				Sync: SyncConfig{
					Enabled:                true,
					AutoSyncAfterOperation: &falseVal,
				},
			},
			// When explicitly set to false, should remain false
			expectedAutoSyncAfterOp: false,
		},
		{
			name: "sync disabled, auto_sync not set - should be false",
			config: &Config{
				Sync: SyncConfig{
					Enabled:                false,
					AutoSyncAfterOperation: nil,
				},
			},
			expectedAutoSyncAfterOp: false,
		},
		{
			name: "sync disabled, auto_sync explicitly true - should still be false",
			config: &Config{
				Sync: SyncConfig{
					Enabled:                false,
					AutoSyncAfterOperation: &trueVal,
				},
			},
			// When sync is disabled, auto-sync should be disabled regardless
			expectedAutoSyncAfterOp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsAutoSyncAfterOperationEnabled()
			if got != tt.expectedAutoSyncAfterOp {
				t.Errorf("IsAutoSyncAfterOperationEnabled() = %v, want %v", got, tt.expectedAutoSyncAfterOp)
			}
		})
	}
}

// TestAutoSyncDefaultWithYAMLNull verifies that when config file has auto_sync_after_operation: null
// (explicitly set to null), it is treated as "not set" and defaults to true when sync is enabled.
// This is the end-to-end test for Issue #30.
func TestAutoSyncDefaultWithYAMLNull(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		configYAML     string
		expectAutoSync bool
	}{
		{
			name: "sync enabled, auto_sync_after_operation: null - should default to true",
			configYAML: `
sync:
  enabled: true
  auto_sync_after_operation: null
backends:
  sqlite:
    path: test.db
`,
			expectAutoSync: true,
		},
		{
			name: "sync enabled, auto_sync_after_operation not set - should default to true",
			configYAML: `
sync:
  enabled: true
backends:
  sqlite:
    path: test.db
`,
			expectAutoSync: true,
		},
		{
			name: "sync enabled, auto_sync_after_operation: false - should be false",
			configYAML: `
sync:
  enabled: true
  auto_sync_after_operation: false
backends:
  sqlite:
    path: test.db
`,
			expectAutoSync: false,
		},
		{
			name: "sync disabled, auto_sync_after_operation: null - should be false",
			configYAML: `
sync:
  enabled: false
  auto_sync_after_operation: null
backends:
  sqlite:
    path: test.db
`,
			expectAutoSync: false,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a unique config file for each test
			configPath := filepath.Join(tmpDir, "config_"+string(rune('a'+i))+".yaml")
			if err := os.WriteFile(configPath, []byte(tt.configYAML), 0644); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			// Load the config
			cfg, _, err := LoadWithRaw(configPath)
			if err != nil {
				t.Fatalf("LoadWithRaw() error = %v", err)
			}

			// Check the result
			got := cfg.IsAutoSyncAfterOperationEnabled()
			if got != tt.expectAutoSync {
				t.Errorf("IsAutoSyncAfterOperationEnabled() = %v, want %v", got, tt.expectAutoSync)
			}
		})
	}
}

// =============================================================================
// Tests for Issue 082: Background Pull Sync Cooldown Configuration
// =============================================================================

// TestBackgroundPullCooldownConfig verifies that the config value is parsed and applied correctly
func TestBackgroundPullCooldownConfig(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		config   string
		expected string
	}{
		{
			name: "30 seconds",
			config: `
sync:
  enabled: true
  background_pull_cooldown: "30s"
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`,
			expected: "30s",
		},
		{
			name: "1 minute",
			config: `
sync:
  enabled: true
  background_pull_cooldown: "1m"
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`,
			expected: "1m",
		},
		{
			name: "5 seconds (minimum valid)",
			config: `
sync:
  enabled: true
  background_pull_cooldown: "5s"
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`,
			expected: "5s",
		},
		{
			name: "2 minutes",
			config: `
sync:
  enabled: true
  background_pull_cooldown: "2m"
backends:
  sqlite:
    enabled: true
default_backend: sqlite
`,
			expected: "2m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tmpDir, tt.name+"-config.yaml")
			if err := os.WriteFile(configPath, []byte(tt.config), 0644); err != nil {
				t.Fatalf("failed to write config: %v", err)
			}

			cfg, err := Load(configPath)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			got := cfg.GetBackgroundPullCooldown()
			if got != tt.expected {
				t.Errorf("GetBackgroundPullCooldown() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestBackgroundPullCooldownDefault verifies default value is 30 seconds when not specified
func TestBackgroundPullCooldownDefault(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name:     "nil config sync - returns default 30s",
			config:   &Config{},
			expected: "30s",
		},
		{
			name: "sync enabled but cooldown not set - returns default 30s",
			config: &Config{
				Sync: SyncConfig{
					Enabled: true,
				},
			},
			expected: "30s",
		},
		{
			name: "sync disabled and cooldown not set - returns default 30s",
			config: &Config{
				Sync: SyncConfig{
					Enabled: false,
				},
			},
			expected: "30s",
		},
		{
			name: "empty string cooldown - returns default 30s",
			config: &Config{
				Sync: SyncConfig{
					Enabled:                true,
					BackgroundPullCooldown: "",
				},
			},
			expected: "30s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetBackgroundPullCooldown()
			if got != tt.expected {
				t.Errorf("GetBackgroundPullCooldown() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestBackgroundPullCooldownValidation verifies invalid values are rejected
func TestBackgroundPullCooldownValidation(t *testing.T) {
	tests := []struct {
		name      string
		cooldown  string
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "valid 30s",
			cooldown: "30s",
			wantErr:  false,
		},
		{
			name:     "valid 1m",
			cooldown: "1m",
			wantErr:  false,
		},
		{
			name:     "valid 5s (minimum)",
			cooldown: "5s",
			wantErr:  false,
		},
		{
			name:      "invalid - 4s (below minimum)",
			cooldown:  "4s",
			wantErr:   true,
			errSubstr: "at least 5s",
		},
		{
			name:      "invalid - negative",
			cooldown:  "-10s",
			wantErr:   true,
			errSubstr: "at least 5s",
		},
		{
			name:      "invalid - zero",
			cooldown:  "0s",
			wantErr:   true,
			errSubstr: "at least 5s",
		},
		{
			name:      "invalid - not a duration",
			cooldown:  "notaduration",
			wantErr:   true,
			errSubstr: "invalid duration",
		},
		{
			name:     "valid empty string (uses default)",
			cooldown: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Backends: BackendsConfig{
					SQLite: SQLiteConfig{
						Enabled: true,
						Path:    "/path/to/db",
					},
				},
				DefaultBackend: "sqlite",
				OutputFormat:   "text",
				Sync: SyncConfig{
					Enabled:                true,
					BackgroundPullCooldown: tt.cooldown,
				},
			}

			err := cfg.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for cooldown %q, got nil", tt.cooldown)
				} else if tt.errSubstr != "" && !containsSubstring(err.Error(), tt.errSubstr) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errSubstr)
				}
			} else if err != nil {
				t.Errorf("unexpected error for cooldown %q: %v", tt.cooldown, err)
			}
		})
	}
}

// containsSubstring checks if s contains substr
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
