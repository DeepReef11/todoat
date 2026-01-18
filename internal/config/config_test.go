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
