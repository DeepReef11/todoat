// Package config handles application configuration
package config

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed config.sample.yaml
var sampleConfig string

// GetSampleConfig returns the embedded sample configuration content
func GetSampleConfig() string {
	return sampleConfig
}

// UIConfig holds user interface settings
type UIConfig struct {
	InteractivePromptForAllTasks bool `yaml:"interactive_prompt_for_all_tasks"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	BackgroundEnabled *bool `yaml:"background_enabled"` // Controls background log file creation (default: true)
}

// Config represents the application configuration
type Config struct {
	Backends          BackendsConfig  `yaml:"backends"`
	DefaultBackend    string          `yaml:"default_backend"`
	DefaultView       string          `yaml:"default_view"`
	NoPrompt          bool            `yaml:"no_prompt"`
	OutputFormat      string          `yaml:"output_format"`
	Sync              SyncConfig      `yaml:"sync"`
	AutoDetectBackend bool            `yaml:"auto_detect_backend"`
	Trash             TrashConfig     `yaml:"trash"`
	Analytics         AnalyticsConfig `yaml:"analytics"`
	Reminder          ReminderConfig  `yaml:"reminder"`
	UI                UIConfig        `yaml:"ui"`
	Logging           LoggingConfig   `yaml:"logging"`
	CacheTTL          string          `yaml:"cache_ttl"` // List metadata cache TTL (e.g., "5m", "30s", "10m")
}

// ReminderConfig holds reminder settings
type ReminderConfig struct {
	Enabled         bool     `yaml:"enabled"`
	Intervals       []string `yaml:"intervals"`
	OSNotification  bool     `yaml:"os_notification"`
	LogNotification bool     `yaml:"log_notification"`
}

// AnalyticsConfig holds analytics settings
type AnalyticsConfig struct {
	Enabled       bool `yaml:"enabled"`
	RetentionDays int  `yaml:"retention_days"`
}

// TrashConfig holds trash management settings
type TrashConfig struct {
	RetentionDays *int `yaml:"retention_days"`
}

// SyncConfig holds synchronization settings
type SyncConfig struct {
	Enabled                bool         `yaml:"enabled"`
	LocalBackend           string       `yaml:"local_backend"`
	ConflictResolution     string       `yaml:"conflict_resolution"`
	OfflineMode            string       `yaml:"offline_mode"`              // auto, online, offline
	ConnectivityTimeout    string       `yaml:"connectivity_timeout"`      // e.g., "5s"
	AutoSyncAfterOperation *bool        `yaml:"auto_sync_after_operation"` // sync immediately after operations (default: true when sync enabled)
	BackgroundPullCooldown string       `yaml:"background_pull_cooldown"`  // cooldown between background pull syncs (default: "30s", minimum: "5s")
	Daemon                 DaemonConfig `yaml:"daemon"`
}

// DaemonConfig holds background daemon settings
type DaemonConfig struct {
	Enabled           bool `yaml:"enabled"`            // Enable forked daemon process (Issue #36)
	Interval          int  `yaml:"interval"`           // Sync interval in seconds
	IdleTimeout       int  `yaml:"idle_timeout"`       // Idle timeout in seconds before daemon exits
	HeartbeatInterval int  `yaml:"heartbeat_interval"` // Heartbeat recording interval in seconds (Issue #74)
	FileWatcher       bool `yaml:"file_watcher"`       // Enable file watcher for real-time sync triggers (Issue #41)
	SmartTiming       bool `yaml:"smart_timing"`       // Enable smart timing to avoid sync during active editing (Issue #41)
	DebounceMs        int  `yaml:"debounce_ms"`        // Debounce duration in milliseconds (Issue #41)
}

// BackendsConfig holds configuration for all backends
type BackendsConfig struct {
	SQLite    SQLiteConfig    `yaml:"sqlite"`
	Todoist   TodoistConfig   `yaml:"todoist"`
	Nextcloud NextcloudConfig `yaml:"nextcloud"`
}

// TodoistConfig holds Todoist backend configuration
type TodoistConfig struct {
	Enabled bool `yaml:"enabled"`
}

// NextcloudConfig holds Nextcloud backend configuration
type NextcloudConfig struct {
	Enabled bool `yaml:"enabled"`
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
		Analytics: AnalyticsConfig{
			Enabled: true, // Enabled by default per FEAT-008
		},
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

	// Use the embedded sample config which includes all documentation and comments
	content := sampleConfig

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
	validBackends := map[string]bool{"sqlite": true, "todoist": true, "nextcloud": true}
	if !validBackends[c.DefaultBackend] {
		return fmt.Errorf("unknown default_backend: %q", c.DefaultBackend)
	}

	// Check that the default backend is enabled
	switch c.DefaultBackend {
	case "sqlite":
		if !c.Backends.SQLite.Enabled {
			return errors.New("default backend 'sqlite' is not enabled in backends configuration")
		}
	case "todoist":
		if !c.Backends.Todoist.Enabled {
			return errors.New("default backend 'todoist' is not enabled in backends configuration")
		}
	case "nextcloud":
		if !c.Backends.Nextcloud.Enabled {
			return errors.New("default backend 'nextcloud' is not enabled in backends configuration")
		}
	}

	// Validate background_pull_cooldown if specified
	if c.Sync.BackgroundPullCooldown != "" {
		duration, err := time.ParseDuration(c.Sync.BackgroundPullCooldown)
		if err != nil {
			return fmt.Errorf("invalid duration for sync.background_pull_cooldown: %q", c.Sync.BackgroundPullCooldown)
		}
		if duration < 5*time.Second {
			return fmt.Errorf("sync.background_pull_cooldown must be at least 5s, got %q", c.Sync.BackgroundPullCooldown)
		}
	}

	// Validate cache_ttl if specified
	if c.CacheTTL != "" {
		_, err := time.ParseDuration(c.CacheTTL)
		if err != nil {
			return fmt.Errorf("invalid duration for cache_ttl: %q", c.CacheTTL)
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

// GetOfflineMode returns the offline mode setting.
// Returns "auto" as default if not configured.
func (c *Config) GetOfflineMode() string {
	mode := c.Sync.OfflineMode
	if mode == "" {
		return "auto"
	}
	return mode
}

// GetConnectivityTimeout returns the connectivity timeout setting.
// Returns "5s" as default if not configured.
func (c *Config) GetConnectivityTimeout() string {
	timeout := c.Sync.ConnectivityTimeout
	if timeout == "" {
		return "5s"
	}
	return timeout
}

// GetBackgroundPullCooldown returns the background pull sync cooldown setting.
// Returns "30s" as default if not configured.
func (c *Config) GetBackgroundPullCooldown() string {
	cooldown := c.Sync.BackgroundPullCooldown
	if cooldown == "" {
		return "30s"
	}
	return cooldown
}

// GetBackgroundPullCooldownDuration returns the background pull sync cooldown as a time.Duration.
// Returns 30 seconds as default if not configured or if parsing fails.
func (c *Config) GetBackgroundPullCooldownDuration() time.Duration {
	cooldownStr := c.GetBackgroundPullCooldown()
	duration, err := time.ParseDuration(cooldownStr)
	if err != nil {
		return 30 * time.Second // Default fallback
	}
	return duration
}

// IsAutoSyncAfterOperationEnabled returns true if auto-sync after operation is enabled.
// When sync is enabled and auto_sync_after_operation is not explicitly set, it defaults to true.
// When sync is disabled, auto-sync is always disabled regardless of the setting.
func (c *Config) IsAutoSyncAfterOperationEnabled() bool {
	// If sync is disabled, auto-sync is always disabled
	if !c.Sync.Enabled {
		return false
	}
	// If auto_sync_after_operation is not set (nil), default to true when sync is enabled
	if c.Sync.AutoSyncAfterOperation == nil {
		return true
	}
	// Return the explicitly set value
	return *c.Sync.AutoSyncAfterOperation
}

// GetAutoSyncAfterOperationConfigValue returns the configured value of auto_sync_after_operation.
// This returns the raw value stored in config, not the effective runtime value.
// If not explicitly set, it returns the default (true when sync is enabled).
func (c *Config) GetAutoSyncAfterOperationConfigValue() bool {
	if c.Sync.AutoSyncAfterOperation == nil {
		// Default to true when sync is enabled, false otherwise
		return c.Sync.Enabled
	}
	return *c.Sync.AutoSyncAfterOperation
}

// IsAutoDetectEnabled returns true if auto-detection is enabled
func (c *Config) IsAutoDetectEnabled() bool {
	return c.AutoDetectBackend
}

// IsInteractivePromptForAllTasks returns true if interactive prompts should show all tasks
// including completed and cancelled.
func (c *Config) IsInteractivePromptForAllTasks() bool {
	return c.UI.InteractivePromptForAllTasks
}

// GetTrashRetentionDays returns the trash retention period in days.
// Returns 30 (default) if not configured, or 0 if auto-purge is disabled.
func (c *Config) GetTrashRetentionDays() int {
	// If RetentionDays is nil (not set), return default of 30
	if c.Trash.RetentionDays == nil {
		return 30 // Default retention period
	}
	// Return the configured value (0 means disabled)
	return *c.Trash.RetentionDays
}

// IsAnalyticsEnabled returns true if analytics is enabled in config
func (c *Config) IsAnalyticsEnabled() bool {
	return c.Analytics.Enabled
}

// GetAnalyticsRetentionDays returns the analytics retention period in days.
// Returns 365 (default) if not configured.
func (c *Config) GetAnalyticsRetentionDays() int {
	if c.Analytics.RetentionDays <= 0 {
		return 365 // Default retention period
	}
	return c.Analytics.RetentionDays
}

// IsDaemonEnabled returns true if the forked daemon feature is enabled.
// This enables the Issue #36 background daemon architecture.
func (c *Config) IsDaemonEnabled() bool {
	return c.Sync.Daemon.Enabled
}

// GetDaemonInterval returns the daemon sync interval in seconds.
// Returns 300 (5 minutes) if not configured.
func (c *Config) GetDaemonInterval() int {
	if c.Sync.Daemon.Interval <= 0 {
		return 300 // Default: 5 minutes
	}
	return c.Sync.Daemon.Interval
}

// GetDaemonIdleTimeout returns the daemon idle timeout in seconds.
// Returns 300 (5 minutes) if not configured.
func (c *Config) GetDaemonIdleTimeout() int {
	if c.Sync.Daemon.IdleTimeout <= 0 {
		return 300 // Default: 5 minutes
	}
	return c.Sync.Daemon.IdleTimeout
}

// IsFileWatcherEnabled returns true if the file watcher is enabled for the daemon.
func (c *Config) IsFileWatcherEnabled() bool {
	return c.Sync.Daemon.FileWatcher
}

// IsSmartTimingEnabled returns true if smart timing is enabled for the daemon.
func (c *Config) IsSmartTimingEnabled() bool {
	return c.Sync.Daemon.SmartTiming
}

// GetDaemonDebounceMs returns the debounce duration in milliseconds.
// Returns 1000 (1 second) if not configured.
func (c *Config) GetDaemonDebounceMs() int {
	if c.Sync.Daemon.DebounceMs <= 0 {
		return 1000 // Default: 1 second
	}
	return c.Sync.Daemon.DebounceMs
}

// GetDaemonHeartbeatInterval returns the daemon heartbeat interval in seconds.
// Returns 5 (5 seconds) if not configured.
// Issue #74: Heartbeat mechanism for hung daemon detection.
func (c *Config) GetDaemonHeartbeatInterval() int {
	if c.Sync.Daemon.HeartbeatInterval <= 0 {
		return 5 // Default: 5 seconds
	}
	return c.Sync.Daemon.HeartbeatInterval
}

// IsBackgroundLoggingEnabled returns true if background logging is enabled.
// Background logging creates PID-specific log files in /tmp for background processes.
// Returns true (default) if not configured.
func (c *Config) IsBackgroundLoggingEnabled() bool {
	if c.Logging.BackgroundEnabled == nil {
		return true // Default: enabled
	}
	return *c.Logging.BackgroundEnabled
}

// GetCacheTTL returns the cache TTL setting as a string.
// Returns "5m" (default) if not configured.
func (c *Config) GetCacheTTL() string {
	if c.CacheTTL == "" {
		return "5m" // Default: 5 minutes
	}
	return c.CacheTTL
}

// GetCacheTTLDuration returns the cache TTL as a time.Duration.
// Returns 5 minutes as default if not configured or if parsing fails.
func (c *Config) GetCacheTTLDuration() time.Duration {
	ttlStr := c.GetCacheTTL()
	duration, err := time.ParseDuration(ttlStr)
	if err != nil {
		return 5 * time.Minute // Default fallback
	}
	return duration
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

// LoadRaw loads configuration from YAML bytes and returns both the structured config
// and the raw map for accessing custom backend configurations.
func LoadRaw(data []byte) (*Config, map[string]interface{}, error) {
	// Parse structured config
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, nil, fmt.Errorf("invalid YAML in config: %w", err)
	}

	// Parse raw map for custom backend access
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, nil, fmt.Errorf("invalid YAML in config: %w", err)
	}

	// Apply defaults for unset fields
	if cfg.DefaultBackend == "" {
		cfg.DefaultBackend = "sqlite"
	}
	if cfg.OutputFormat == "" {
		cfg.OutputFormat = "text"
	}

	return cfg, raw, nil
}

// GetBackendConfig retrieves the configuration for a backend by name.
// It returns the backend configuration map, the backend type, and any error.
// If the backend has no explicit "type" field, the name is used as the type
// for backward compatibility (e.g., "sqlite" backend defaults to type "sqlite").
//
// This function looks for the backend in two locations:
// 1. Under the "backends:" section (preferred format)
// 2. At the top level of the config (for backwards compatibility)
func GetBackendConfig(raw map[string]interface{}, name string) (map[string]interface{}, string, error) {
	var backendCfg map[string]interface{}
	var ok bool

	// First, try to find the backend under the "backends:" section
	if backends, backendsOk := raw["backends"].(map[string]interface{}); backendsOk {
		backendCfg, ok = backends[name].(map[string]interface{})
	}

	// If not found under backends, check top-level for backwards compatibility
	if !ok {
		backendCfg, ok = raw[name].(map[string]interface{})
	}

	if !ok {
		return nil, "", fmt.Errorf("backend '%s' not found in configuration", name)
	}

	// Determine the backend type
	backendType, _ := backendCfg["type"].(string)
	if backendType == "" {
		// For backward compatibility, use the backend name as the type
		// if no explicit type is specified
		backendType = name
	}

	return backendCfg, backendType, nil
}

// IsBackendConfigured checks if a backend with the given name exists in the configuration.
// This function checks both the "backends:" section and top-level for backwards compatibility.
func IsBackendConfigured(raw map[string]interface{}, name string) bool {
	// Check under backends: section first
	if backends, ok := raw["backends"].(map[string]interface{}); ok {
		if _, ok := backends[name].(map[string]interface{}); ok {
			return true
		}
	}
	// Check top-level for backwards compatibility
	_, ok := raw[name].(map[string]interface{})
	return ok
}

// LoadWithRaw loads configuration from the specified path and returns both the structured config
// and the raw map. If the config file doesn't exist, it returns nil for the raw map.
func LoadWithRaw(configPath string) (*Config, map[string]interface{}, error) {
	if configPath == "" {
		configPath = filepath.Join(GetConfigDir(), "config.yaml")
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create config with defaults
		cfg := DefaultConfig()
		if err := cfg.save(configPath); err != nil {
			return nil, nil, fmt.Errorf("failed to create default config: %w", err)
		}
		// Return default config with nil raw (no custom backends in default config)
		return cfg, nil, nil
	}

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return LoadRaw(data)
}
