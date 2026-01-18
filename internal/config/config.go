// Package config handles application configuration
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Backends       BackendsConfig `yaml:"backends"`
	DefaultBackend string         `yaml:"default_backend"`
	NoPrompt       bool           `yaml:"no_prompt"`
	OutputFormat   string         `yaml:"output_format"`
	Sync           SyncConfig     `yaml:"sync"`
}

// SyncConfig holds synchronization settings
type SyncConfig struct {
	Enabled            bool   `yaml:"enabled"`
	LocalBackend       string `yaml:"local_backend"`
	ConflictResolution string `yaml:"conflict_resolution"`
}

// BackendsConfig holds configuration for all backends
type BackendsConfig struct {
	SQLite SQLiteConfig `yaml:"sqlite"`
}

// SQLiteConfig holds SQLite backend configuration
type SQLiteConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Backends: BackendsConfig{
			SQLite: SQLiteConfig{
				Enabled: true,
				Path:    filepath.Join(GetDataDir(), "tasks.db"),
			},
		},
		DefaultBackend: "sqlite",
		NoPrompt:       false,
		OutputFormat:   "text",
	}
}

// Load loads configuration from the specified path, or the default XDG path if empty.
// If the config file doesn't exist, it creates one with defaults.
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = filepath.Join(GetConfigDir(), "config.yaml")
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create config with defaults
		cfg := DefaultConfig()
		if err := cfg.save(configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return cfg, nil
	}

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the file without defaults first
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("invalid YAML in config file: %w", err)
	}

	// Apply defaults for unset fields (but not backends - those must be explicit)
	if cfg.DefaultBackend == "" {
		cfg.DefaultBackend = "sqlite"
	}
	if cfg.OutputFormat == "" {
		cfg.OutputFormat = "text"
	}

	// Expand paths if backend is configured
	if cfg.Backends.SQLite.Path != "" {
		cfg.Backends.SQLite.Path = ExpandPath(cfg.Backends.SQLite.Path)
	}

	return cfg, nil
}

// save writes the configuration to the specified path
func (c *Config) save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add a header comment
	content := "# todoat configuration\n" + string(data)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Check output format
	if c.OutputFormat != "text" && c.OutputFormat != "json" {
		return fmt.Errorf("invalid output_format: %q (must be 'text' or 'json')", c.OutputFormat)
	}

	// Check default backend
	validBackends := map[string]bool{"sqlite": true}
	if !validBackends[c.DefaultBackend] {
		return fmt.Errorf("unknown default_backend: %q", c.DefaultBackend)
	}

	// Check that the default backend is enabled
	switch c.DefaultBackend {
	case "sqlite":
		if !c.Backends.SQLite.Enabled {
			return errors.New("default backend 'sqlite' is not enabled in backends configuration")
		}
	}

	return nil
}

// ApplyFlags applies CLI flag overrides to the configuration
func (c *Config) ApplyFlags(noPrompt bool, outputFormat string) {
	if noPrompt {
		c.NoPrompt = true
	}
	if outputFormat != "" {
		c.OutputFormat = outputFormat
	}
}

// GetDatabasePath returns the path to the SQLite database
func (c *Config) GetDatabasePath() string {
	return c.Backends.SQLite.Path
}

// IsSyncEnabled returns true if synchronization is enabled
func (c *Config) IsSyncEnabled() bool {
	return c.Sync.Enabled
}

// LoadFromPath loads configuration from a specific path without creating defaults
func LoadFromPath(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("config path is required")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // File doesn't exist, return nil config
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("invalid YAML in config file: %w", err)
	}

	return cfg, nil
}

// getXDGDir returns a directory path following XDG spec.
// envVar is the XDG environment variable (e.g., "XDG_CONFIG_HOME").
// fallbackPath is the relative path from home (e.g., ".config").
func getXDGDir(envVar, fallbackPath string) string {
	if xdgDir := os.Getenv(envVar); xdgDir != "" {
		return filepath.Join(xdgDir, "todoat")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", fallbackPath, "todoat")
	}
	return filepath.Join(home, fallbackPath, "todoat")
}

// GetConfigDir returns the configuration directory following XDG spec
func GetConfigDir() string {
	return getXDGDir("XDG_CONFIG_HOME", ".config")
}

// GetDataDir returns the data directory following XDG spec
func GetDataDir() string {
	return getXDGDir("XDG_DATA_HOME", filepath.Join(".local", "share"))
}

// GetCacheDir returns the cache directory following XDG spec
func GetCacheDir() string {
	return getXDGDir("XDG_CACHE_HOME", ".cache")
}

// ExpandPath expands ~ and environment variables in a path
func ExpandPath(path string) string {
	if path == "" {
		return path
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	return path
}
