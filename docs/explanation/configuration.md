# Configuration System

This document provides comprehensive documentation for todoat' configuration system, covering all aspects of the YAML-based configuration, XDG compliance, path expansion, and configuration management.

## Table of Contents

- [Overview](#overview)
- [Configuration File Location](#configuration-file-location)
- [Configuration File Structure](#configuration-file-structure)
- [Feature Documentation](#feature-documentation)
  - [YAML Configuration Format](#yaml-configuration-format)
  - [XDG Base Directory Compliance](#xdg-base-directory-compliance)
  - [Multi-Backend Configuration](#multi-backend-configuration)
  - [Sync Configuration](#sync-configuration)
  - [Path Expansion](#path-expansion)
  - [Auto-Initialization](#auto-initialization)
  - [Custom Config Path](#custom-config-path)
  - [Config Validation](#config-validation)
  - [Default Backend](#default-backend)
  - [Backend Priority](#backend-priority)
  - [Conflict Resolution Configuration](#conflict-resolution-configuration)
  - [Sync Interval](#sync-interval)
  - [Offline Mode Configuration](#offline-mode-configuration)
  - [View Defaults](#view-defaults)
  - [Cache Configuration](#cache-configuration)
  - [No-Prompt Mode Configuration](#no-prompt-mode-configuration)
  - [Output Format Configuration](#output-format-configuration)
  - [Singleton Pattern](#singleton-pattern)
- [Notification Configuration](#notification-configuration)
- [Configuration Examples](#configuration-examples)
- [Environment Variables](#environment-variables)
- [Related Features](#related-features)

---

## Overview

**Purpose**: The configuration system provides flexible, standards-compliant configuration management for todoat, allowing users to customize backend connections, sync behavior, credentials, and application settings through a human-readable YAML file.

**Key Characteristics**:
- **Standards-Compliant**: Follows XDG Base Directory Specification
- **Flexible**: Supports multiple backends, credential sources, and configuration options
- **Safe**: Validation ensures configuration correctness before use
- **Portable**: Path expansion with `~` and `$HOME` for cross-platform compatibility
- **Secure**: Separates credentials from configuration (keyring support)

---

## Configuration File Location

The configuration file follows the XDG Base Directory Specification:

| Priority | Location | Description |
|----------|----------|-------------|
| **1. Custom Path** | `--config` flag | Explicit path specified via command-line |
| **2. XDG Config** | `$XDG_CONFIG_HOME/todoat/config.yaml` | User-specific configuration |
| **3. Default** | `~/.config/todoat/config.yaml` | Fallback if XDG_CONFIG_HOME not set |

**Related XDG Directories**:

| Type | Variable | Default | Purpose |
|------|----------|---------|---------|
| Config | `XDG_CONFIG_HOME` | `~/.config/todoat/` | Configuration files, custom views |
| Data | `XDG_DATA_HOME` | `~/.local/share/todoat/` | SQLite databases, cache files |
| Cache | `XDG_CACHE_HOME` | `~/.cache/todoat/` | Task list cache (`lists.json`) |

---

## Configuration File Structure

```yaml
# Backend configuration
backends:
  nextcloud-prod:
    type: nextcloud
    enabled: true
    host: "nextcloud.example.com"
    username: "myuser"
  nextcloud-test:
    type: nextcloud
    enabled: true
    host: "nextcloud://admin:admin123@localhost:8080/" # No keyring, no username needed, insecure login
    suppress_ssl_warning: true
    insecure_skip_verify: true
    allow_http: true  # Required for HTTP test server (localhost:8080)
    suppress_http_warning: true  # Suppress HTTP security warning for local testing

  sqlite:
    type: sqlite
    enabled: false
    db_path: ""  # Empty = XDG default

# Backend selection
default_backend: nextcloud-prod
auto_detect_backend: false
backend_priority:
  - nextcloud-prod
  - git

# Synchronization
sync:
  enabled: true
  auto_sync_after_operation: true  # Sync immediately after add/update/delete
  local_backend: sqlite
  conflict_resolution: server_wins
  offline_mode: auto

# UI & Display
canWriteConfig: true
ui: cli
date_format: "2006-01-02"

# Prompt & Output (for scripting/automation)
no_prompt: false           # Disable interactive prompts (-y flag)
output_format: text        # Output format: text | json (--json flag)
```

---

## Feature Documentation

### YAML Configuration Format

**Purpose**: Provides a human-readable, structured configuration format that's easy to edit and version control.

**How It Works**:

1. **Parsing**: Configuration is parsed using the `gopkg.in/yaml.v3` library
2. **Structure**: The YAML file maps to the `Config` struct in Go
3. **Comments**: Supports YAML comments for documentation
4. **Validation**: Parsed configuration is validated before use

**User Journey**:

1. User creates or edits `~/.config/todoat/config.yaml`
2. User structures configuration using YAML syntax (maps, lists, strings, booleans)
3. todoat loads and parses the YAML on startup
4. Validation errors are reported with helpful messages
5. Valid configuration is used throughout the application

**Prerequisites**:
- None - configuration is created from sample on first run

**Outputs/Results**:
- Parsed `Config` object used by application
- Clear error messages for YAML syntax errors

**Technical Details**:

```go
// Config struct definition
type Config struct {
    Backends          map[string]backend.BackendConfig `yaml:"backends,omitempty"`
    DefaultBackend    string                           `yaml:"default_backend,omitempty"`
    AutoDetectBackend bool                             `yaml:"auto_detect_backend,omitempty"`
    BackendPriority   []string                         `yaml:"backend_priority,omitempty"`
    UI                string                           `yaml:"ui" validate:"oneof=cli tui gui"`
    DateFormat        string                           `yaml:"date_format,omitempty"`
    Sync              *SyncConfig                      `yaml:"sync,omitempty"`
    NoPrompt          bool                             `yaml:"no_prompt,omitempty"`
    OutputFormat      string                           `yaml:"output_format,omitempty" validate:"oneof=text json"`
}
```

**Related Features**:
- [Config Validation](#config-validation) - Ensures YAML structure is correct
- [Auto-Initialization](#auto-initialization) - Creates sample config on first run

---

### XDG Base Directory Compliance

**Purpose**: Follows the XDG Base Directory Specification to organize configuration, data, and cache files in standard locations, ensuring compatibility with Linux/Unix systems and avoiding home directory clutter.

**How It Works**:

1. **Directory Detection**: Checks environment variables for XDG paths
2. **Fallback Logic**: Uses standard defaults if XDG variables not set
3. **Directory Creation**: Creates directories as needed with correct permissions
4. **Path Resolution**: All file operations use XDG-compliant paths

**User Journey**:

1. User installs todoat
2. On first run, todoat checks `XDG_CONFIG_HOME` environment variable
3. If set, uses `$XDG_CONFIG_HOME/todoat/`
4. If not set, uses `~/.config/todoat/`
5. Creates directory structure automatically
6. Places config, data, and cache files in appropriate locations

**Prerequisites**:
- None - XDG directories are created automatically

**Outputs/Results**:
- Standard directory structure:
  - Config: `~/.config/todoat/config.yaml`
  - Data: `~/.local/share/todoat/tasks.db`
  - Cache: `~/.cache/todoat/lists.json`
  - Views: `~/.config/todoat/views/`

**Technical Details**:

```go
// Get config directory
func GetConfigPath() (string, error) {
    dir, err := os.UserConfigDir()  // Uses XDG_CONFIG_HOME or ~/.config
    if err != nil {
        return "", fmt.Errorf("failed to get user config dir: %w", err)
    }
    return filepath.Join(dir, CONFIG_DIR_PATH, CONFIG_FILE_PATH), nil
}

// Directory permissions
const (
    CONFIG_DIR_PATH  = "todoat"
    CONFIG_FILE_PATH = "config.yaml"
    CONFIG_DIR_PERM  = 0755  // rwxr-xr-x
    CONFIG_FILE_PERM = 0644  // rw-r--r--
)
```

**Environment Variables**:
- `XDG_CONFIG_HOME` - Configuration directory (default: `~/.config`)
- `XDG_DATA_HOME` - Data directory (default: `~/.local/share`)
- `XDG_CACHE_HOME` - Cache directory (default: `~/.cache`)

**Related Features**:
- [Path Expansion](#path-expansion) - Expands `~` and `$HOME` in paths
- [Cache Configuration](#cache-configuration) - Uses XDG cache directory

---

### Multi-Backend Configuration

**Purpose**: Allows configuration of multiple storage backends in a single configuration file, enabling users to work with different task sources (Nextcloud, Todoist, SQLite, Git) simultaneously.

**How It Works**:

1. **Backend Map**: Each backend is a named entry in the `backends` map
2. **Type-Specific Settings**: Each backend has type-specific configuration
3. **Enable/Disable**: Individual backends can be toggled with `enabled` flag
4. **Independent Credentials**: Each backend has separate credential configuration
5. **Backend Selection**: Backends are selected via priority, default, or explicit flag

**User Journey**:

1. User edits config file to add new backend
2. User sets `type` field (nextcloud, todoist, sqlite, git, file)
3. User provides type-specific configuration (host, username, db_path, etc.)
4. User sets `enabled: true` to activate backend
5. User saves config file
6. todoat validates backend configuration on next run
7. User accesses backend via `--backend` flag or selection priority

**Prerequisites**:
- Valid backend type
- Required fields for backend type (varies by type)
- [Credential Management](credential-management.md) for remote backends

**Outputs/Results**:
- Multiple backends available for selection
- Each backend appears in `--detect-backend` output
- Backend-specific validation errors if configuration invalid

**Technical Details**:

**Backend Types and Required Fields**:

```yaml
# Nextcloud/CalDAV
nextcloud-prod:
  type: nextcloud
  enabled: true
  host: "nextcloud.example.com"  # OR url with credentials (legacy)
  username: "myuser"             # Required for keyring/env lookup

# Todoist
todoist:
  type: todoist
  enabled: true
  username: "token"              # "token" for API key lookup
  # api_token: "..."             # Alternative to keyring

# SQLite
sqlite:
  type: sqlite
  enabled: true
  db_path: ""                    # Empty = XDG default

# Git/Markdown
git:
  type: git
  enabled: true
  file: "TODO.md"
  auto_detect: true
  fallback_files:
    - "todo.md"
    - ".todoat.md"
  auto_commit: false

# File 
file:
  type: file
  enabled: false
  url: "file://~/.config/todoat/tasks.md"
```

**Validation Rules**:
- Backend `type` must be valid: nextcloud, todoist, sqlite, git, file
- Nextcloud: requires `url`, `host`, or `username`
- Git: `file` defaults to "TODO.md" if not specified
- SQLite: `db_path` is optional (empty = XDG default)

**Related Features**:
- [Backend System](backend-system.md) - Backend architecture and selection
- [Credential Management](credential-management.md) - Storing backend credentials
- [Default Backend](#default-backend) - Setting primary backend

---

### Sync Configuration

**Purpose**: Configures global synchronization behavior that applies to all remote backends, enabling offline-first task management with automatic caching and conflict resolution.

**How It Works**:

1. **Global Sync Setting**: `sync.enabled` controls automatic caching for all remote backends
2. **Shared Cache Database**: All remote backends share a single cache database
3. **Per-Backend Isolation**: Tasks from different backends are isolated via `backend_name` column
4. **Auto-Sync**: Background sync can be enabled/disabled. When disabled, user must launch the command `todoat sync` to sync with backend
5. **Conflict Resolution**: Global strategy applies to all backends (individual backends can have its own conflict resolution)
6. **Opt-Out**: Individual backends can disable caching with `sync: {enabled: false}`

**User Journey**:

1. User enables sync in config: `sync.enabled: true`
2. User configures sync settings (conflict resolution, interval, offline mode)
3. todoat creates shared cache database on startup
4. Each enabled remote backend is automatically cached
5. User performs task operations (local-first, instant response)
6. Sync happens manually (`todoat sync`) or automatically after operations (if `auto_sync_after_operation: true`)
7. Changes propagate between local cache and remote backends

**Prerequisites**:
- At least one remote backend configured and enabled
- See [Synchronization](synchronization.md) for detailed sync requirements

**Outputs/Results**:
- Shared cache database: `~/.local/share/todoat/cache.db`
- All task operations use local cache (fast)
- Sync operations propagate changes to/from remotes
- Conflicts resolved according to configured strategy

**Technical Details**:

**Configuration Structure**:

```yaml
sync:
  enabled: true                        # Enable automatic caching for all remote backends
  auto_sync_after_operation: true      # Sync immediately after add/update/delete operations
  local_backend: sqlite                # Cache backend type (default: sqlite)
  conflict_resolution: server_wins     # server_wins, local_wins, merge, keep_both
  offline_mode: auto                   # auto, online, offline
```

**Sync Configuration Fields**:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | boolean | false | Enable automatic caching for remote backends |
| `local_backend` | string | sqlite | Cache storage type |
| `conflict_resolution` | string | server_wins | Conflict strategy (server_wins, local_wins, merge, keep_both) |
| `auto_sync_after_operation` | boolean | false | Sync immediately after add/update/delete operations |
| `offline_mode` | string | auto | Offline behavior (auto, online, offline) |
| `connectivity_timeout` | string | 5s | Timeout for connectivity checks |

**Validation Rules**:
- `local_backend` must be: sqlite, file, or git
- `conflict_resolution` must be: server_wins, local_wins, merge, or keep_both
- `offline_mode` must be: auto, online, or offline
- `sync_interval` cannot be negative (0 = manual only)

**Cache Database Location**:
```go
// GetCacheDatabasePath returns the shared cache path
func (c *Config) GetCacheDatabasePath() (string, error) {
    dataDir := os.Getenv("XDG_DATA_HOME")
    if dataDir == "" {
        homeDir, _ := os.UserHomeDir()
        dataDir = filepath.Join(homeDir, ".local", "share")
    }
    return filepath.Join(dataDir, "todoat", "cache.db"), nil
}
```


**Related Features**:
- [Synchronization](synchronization.md) - Comprehensive sync documentation
- [Backend System](backend-system.md) - Backend architecture
- [Conflict Resolution Configuration](#conflict-resolution-configuration) - Conflict strategies
- [Offline Mode Configuration](#offline-mode-configuration) - Offline behavior

---

## Notification Configuration

The notification system provides alerts for background sync operations.

```yaml
notification:
  enabled: true
  os_notification:
    enabled: true
    on_sync_error: true
    on_conflict: true
  log_notification:
    enabled: true
```

See [Notification Manager](notification-manager.md) for full configuration options.

---

### Path Expansion

**Purpose**: Expands shell-style path shortcuts (`~`, `$HOME`) and environment variables in configuration file paths, making configs portable across different user accounts and systems.

**How It Works**:

1. **Tilde Expansion**: `~` at start of path expands to user home directory
2. **Environment Variables**: `$HOME` expands to home directory anywhere in path
3. **Escaped Sequences**: `\~` and `\$` preserve literal characters
4. **Cross-Platform**: Works on Linux, macOS, and Windows
5. **Automatic**: Expansion happens during config loading

**User Journey**:

1. User writes config with portable paths using `~` or `$HOME`
2. User saves config file
3. On load, todoat expands paths to absolute paths
4. Application uses fully-resolved paths for all operations
5. Config remains portable across different users/systems

**Prerequisites**:
- None - path expansion is automatic

**Outputs/Results**:
- Portable configuration files
- Absolute paths used internally
- No need to hardcode user-specific paths

**Technical Details**:

**Expansion Rules**:

```yaml
# Before expansion (in config file)
backends:
  sqlite:
    db_path: "~/tasks/mydb.db"      # Expands ~ at start

  file:
    url: "file://$HOME/data/tasks.json"  # Expands $HOME anywhere

  git:
    file: "~/projects/\~archive/TODO.md"  # \~ becomes literal ~
```

**After expansion** (on system with HOME=/home/user):

```yaml
backends:
  sqlite:
    db_path: "/home/user/tasks/mydb.db"

  file:
    url: "file:///home/user/data/tasks.json"

  git:
    file: "/home/user/projects/~archive/TODO.md"
```

**Implementation**:

```go
func expandPath(path string) string {
    homeDir, _ := os.UserHomeDir()

    // Protect escaped sequences with placeholders
    path = strings.ReplaceAll(path, `\~`, "\x00ESCAPED_TILDE\x00")
    path = strings.ReplaceAll(path, `\$`, "\x00ESCAPED_DOLLAR\x00")

    // Expand ~ at start
    if strings.HasPrefix(path, "~/") {
        path = filepath.Join(homeDir, path[2:])
    }

    // Expand $HOME anywhere
    path = strings.ReplaceAll(path, "$HOME", homeDir)

    // Restore escaped sequences
    path = strings.ReplaceAll(path, "\x00ESCAPED_TILDE\x00", "~")
    path = strings.ReplaceAll(path, "\x00ESCAPED_DOLLAR\x00", "$")

    return path
}
```

**Supported Paths**:
- `db_path` (sqlite backend)
- `file` (git backend)
- `fallback_files` (git backend)
- `url` with `file://` scheme (file backend)

**Escape Sequences**:
- `\~` → literal `~` (not expanded)
- `\$` → literal `$` (not expanded)
- Use for paths with actual tilde/dollar characters

**Related Features**:
- [XDG Base Directory Compliance](#xdg-base-directory-compliance) - Default paths
- [Multi-Backend Configuration](#multi-backend-configuration) - Backend-specific paths

---

### Auto-Initialization

**Purpose**: Automatically creates a configuration file from an embedded sample on first run, providing users with a documented starting point and eliminating manual setup.

**How It Works**:

1. **First Run Detection**: Checks if config file exists at expected location
2. **User Prompt**: Asks user if they want to create config from sample
3. **Sample Embedding**: Sample config is embedded in binary at compile time
4. **Directory Creation**: Creates config directory with proper permissions
5. **File Writing**: Writes sample config with comments and examples
6. **View Installation**: Copies built-in views to user config directory

**User Journey**:

1. User runs todoat for first time
2. todoat detects no config file exists
3. User sees prompt: "Do you want to copy config sample to ~/.config/todoat/config.yaml?"
4. User selects "yes"
5. Config directory is created: `~/.config/todoat/`
6. Sample config is written: `config.yaml`
7. Built-in views are copied: `views/default.yaml`, `views/all.yaml`
8. User sees: "Built-in views copied to user config directory"
9. User edits config file to customize backends and settings

**Prerequisites**:
- Write permissions to config directory (usually `~/.config/`)

**Outputs/Results**:
- New config file created: `~/.config/todoat/config.yaml`
- Built-in views installed: `~/.config/todoat/views/`
- Sample includes:
  - All backend types with examples
  - Commented-out credential options
  - Sync configuration examples
  - Usage examples and documentation

**Technical Details**:

**Embedded Sample Config**:

```go
//go:embed config.sample.yaml
var sampleConfig []byte

func createConfigFromSample(configPath string) []byte {
    // Create directory
    err := createConfigDir(configPath)
    if err != nil {
        log.Fatal(err)
    }

    // Write sample config
    configData := sampleConfig
    err = WriteConfigFile(configPath, configData)
    if err != nil {
        log.Fatal(err)
    }

    // Copy built-in views
    copied, err := views.CopyBuiltInViewsToUserConfig()
    if err != nil {
        log.Printf("Warning: Failed to copy built-in views: %v", err)
    } else if copied {
        fmt.Println("Built-in views copied to user config directory")
    }

    return configData
}
```

**Sample Config Contents**:
- Backend examples for all types (nextcloud, todoist, sqlite, git, file)
- Three credential storage options (keyring, environment, config)
- Sync configuration with explanations
- SSL/TLS and HTTP settings for development
- Backend selection and priority examples
- Usage examples and command references

**File Permissions**:
- Config directory: `0755` (rwxr-xr-x)
- Config file: `0644` (rw-r--r--)

**What Happens If User Declines**:
- Sample config is used in-memory (not saved to disk)
- User must manually create config file later
- Application runs with sample config for current session

**Related Features**:
- [Config Validation](#config-validation) - Validates sample config
- [Views & Customization](views-customization.md) - Built-in views installation

---

### Custom Config Path

**Purpose**: Allows users to specify a custom configuration file location using the `--config` flag, enabling per-project configs, testing configurations, or non-standard setups.

**How It Works**:

1. **Flag Parsing**: `--config` flag is parsed before config loading
2. **Path Resolution**: Determines if path is file or directory
3. **Priority Override**: Custom path takes priority over XDG defaults
4. **Singleton Reset**: Resets config singleton to force reload with new path
5. **Relative Paths**: Supports relative paths (resolved from current directory)

**User Journey**:

1. User wants to use different config for testing/project
2. User specifies config path: `todoat --config /path/to/config.yaml [command]`
3. todoat loads config from specified path instead of default
4. All operations use custom config for this session
5. Default config location remains unchanged

**Prerequisites**:
- Config file exists at specified path (or directory to create in)
- Valid YAML configuration in custom file

**Outputs/Results**:
- Custom config loaded and used for session
- XDG default config unaffected
- Useful for:
  - Testing configurations
  - Per-project settings
  - CI/CD environments
  - Development vs production configs

**Technical Details**:

**Usage Patterns**:

```bash
# Specify config file directly
todoat --config /path/to/config.yaml MyList

# Specify config directory (looks for config.yaml inside)
todoat --config /path/to/config-dir/ MyList

# Use current directory
todoat --config . MyList
# Looks for: ./todoat/config.yaml

# Relative path
todoat --config ../other-project/config.yaml MyList
```

**Path Resolution Logic**:

```go
func SetCustomConfigPath(path string) {
    if path == "" || path == "." {
        // Current directory: ./todoat/config.yaml
        customConfigPath = filepath.Join(".", CONFIG_DIR_PATH, CONFIG_FILE_PATH)
    } else {
        info, err := os.Stat(path)
        if err == nil && info.IsDir() {
            // Existing directory: append config.yaml
            customConfigPath = filepath.Join(path, CONFIG_FILE_PATH)
        } else if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
            // File path: use directly
            customConfigPath = path
        } else {
            // Unknown: assume directory, append config.yaml
            customConfigPath = filepath.Join(path, CONFIG_FILE_PATH)
        }
    }

    // Reset singleton to force reload
    configOnce = sync.Once{}
    globalConfig = nil
}
```

**Extension Detection**:
- `.yaml`, `.yml` → treated as file path
- `.YAML`, `.YML` → treated as file path
- No extension → treated as directory

**Testing Use Case**:

```bash
# Test config for Docker test server
todoat --config ./todoat/config MyList

# CI/CD config
todoat --config /etc/todoat/ci-config.yaml sync
```

**Related Features**:
- [XDG Base Directory Compliance](#xdg-base-directory-compliance) - Default config locations
- [Config Validation](#config-validation) - Custom configs are validated too

---

### Config Validation

**Purpose**: Validates configuration structure and values before use, catching errors early and providing helpful error messages to guide users in fixing configuration problems.

**How It Works**:

1. **Structural Validation**: Uses `go-playground/validator` for field validation
2. **Type-Specific Validation**: Custom checks for each backend type
3. **Relationship Validation**: Ensures references (default_backend, backend_priority) are valid
4. **Sync Validation**: Validates sync configuration and conflict strategies
5. **Early Exit**: Fatal error on validation failure with descriptive message

**User Journey**:

1. User edits config file with invalid values
2. User runs todoat command
3. Config is loaded and parsed
4. Validation detects error
5. Application exits with error message explaining problem
6. User fixes config based on error message
7. User runs command again successfully

**Prerequisites**:
- Valid YAML syntax (parse errors reported separately)

**Outputs/Results**:
- Clear error messages for configuration problems
- Specific field names and invalid values identified
- Suggestions for valid values
- Application prevents running with invalid config

**Technical Details**:

**Validation Stages**:

```go
func (c Config) Validate() error {
    validate := validator.New()

    // 1. Struct validation (field types, required fields)
    if err := validate.Struct(c); err != nil {
        return err
    }

    // 2. Check backends exist
    if len(c.Backends) == 0 {
        return fmt.Errorf("no backends configured")
    }

    // 3. Validate each backend
    for name, backendConfig := range c.Backends {
        if err := validate.Struct(backendConfig); err != nil {
            return fmt.Errorf("backend %q validation failed: %w", name, err)
        }

        // Type-specific validation
        switch backendConfig.Type {
        case "nextcloud", "file":
            // Validate URL/host/username requirements
        case "git":
            // Validate file paths
        case "sqlite":
            // db_path is optional
        }
    }

    // 4. Validate default backend
    if c.DefaultBackend != "" {
        backend, exists := c.Backends[c.DefaultBackend]
        if !exists {
            return fmt.Errorf("default backend %q not found", c.DefaultBackend)
        }
        if !backend.Enabled {
            return fmt.Errorf("default backend %q is disabled", c.DefaultBackend)
        }
    }

    // 5. Validate backend priority references
    for _, name := range c.BackendPriority {
        if _, exists := c.Backends[name]; !exists {
            return fmt.Errorf("backend_priority references unknown backend %q", name)
        }
    }

    // 6. Validate sync configuration
    if c.Sync != nil && c.Sync.Enabled {
        // Validate local_backend type
        // Validate conflict_resolution strategy
        // Validate offline_mode
        // Validate sync_interval
    }

    return nil
}
```

**Validation Rules**:

| Field | Validation |
|-------|------------|
| `ui` | Must be "cli" or "tui" |
| `backends` | Must not be empty |
| `default_backend` | Must exist in backends map and be enabled |
| `backend_priority` | All entries must exist in backends map |
| `sync.local_backend` | Must be sqlite, file, or git |
| `sync.conflict_resolution` | Must be server_wins, local_wins, merge, or keep_both |
| `sync.offline_mode` | Must be auto, online, or offline |
| `sync.sync_interval` | Cannot be negative |

**Backend-Specific Validation**:

**Nextcloud/File**:
- Must have `url`, `host`, or `username`
- If no URL/host, must have username for env var lookup

**Git**:
- `file` defaults to "TODO.md" if not specified
- Paths are validated during expansion

**SQLite**:
- `db_path` is optional (empty = XDG default)
- No additional validation needed

**Error Message Examples**:

```
Invalid YAML in config file ~/.config/todoat/config.yaml:
  yaml: line 5: mapping values are not allowed in this context

Missing field(s) in YAML config file ~/.config/todoat/config.yaml:
  backend "nextcloud": URL, host, or username is required for nextcloud backend

default backend "sqlite-cache" not found in configured backends

sync.conflict_resolution must be server_wins, local_wins, merge, or keep_both, got "always_local"
```

**Auto-Fix Features**:
- Missing defaults are applied (auto-sync: true, conflict_resolution: server_wins)

**Related Features**:
- [Multi-Backend Configuration](#multi-backend-configuration) - Backend validation rules
- [Sync Configuration](#sync-configuration) - Sync validation rules

---

### Default Backend

**Purpose**: Specifies which backend to use by default when no explicit backend is selected via command-line flag, simplifying commands and providing sensible defaults.

**How It Works**:

1. **Configuration Field**: `default_backend` specifies backend name
2. **Fallback Logic**: If not set, uses first enabled backend
3. **Validation**: Default backend must exist and be enabled
4. **Override**: Can be overridden with `--backend` flag
5. **Sync Integration**: When sync enabled, local backend used instead

**User Journey**:

1. User configures multiple backends in config file
2. User sets `default_backend: nextcloud-prod`
3. User runs commands without `--backend` flag
4. todoat uses nextcloud-prod backend automatically
5. User can override with `--backend sqlite` for specific commands

**Prerequisites**:
- Backend specified in `default_backend` must exist in `backends` map
- Backend must be enabled

**Outputs/Results**:
- Simplified commands (no need for `--backend` flag)
- Predictable behavior across commands
- Easy to change default without updating scripts

**Technical Details**:

**Configuration**:

```yaml
backends:
  nextcloud-prod:
    type: nextcloud
    enabled: true
    # ... config ...

  sqlite:
    type: sqlite
    enabled: true
    # ... config ...

# Set default backend
default_backend: nextcloud-prod  # Uses this when no --backend flag
```

**Backend Selection Priority** (when no `--backend` flag):
Note: When **Sync enabled**: Uses sync manager if remote backend is selected

1. If auto_detect_backend is enabled, this will decide which backend to use
2. **Default backend set**: Uses `default_backend` value 
3. **No default set**: Uses first enabled backend
4. **No enabled backends**: Error


**Usage Examples**:

```bash
# Uses default backend (nextcloud-prod)
todoat MyList

# Override with explicit backend
todoat --backend sqlite MyList

# Detect backend to see default
todoat --detect-backend
```

**Common Patterns**:

**Work/Personal Separation**:
```yaml
default_backend: work-nextcloud

backends:
  work-nextcloud:
    type: nextcloud
    enabled: true
    host: "nextcloud.work.com"

  personal-todoist:
    type: todoist
    enabled: true
```

**Sync with Remote Default**:
```yaml
default_backend: nextcloud-prod
sync:
  enabled: true  # Actually uses local cache for operations

backends:
  nextcloud-prod:
    type: nextcloud
    enabled: true
```

**Related Features**:
- [Backend Priority](#backend-priority) - Alternative selection mechanism
- [Backend System](backend-system.md) - Backend selection logic
- [Sync Configuration](#sync-configuration) - Sync affects default selection

---

### Backend Priority

**Purpose**: Defines an ordered list of backends for automatic selection, enabling auto-detection and fallback strategies when no explicit backend is specified.

**How It Works**:

1. **Priority List**: `backend_priority` is an array of backend names
2. **Order Matters**: First backend in list is tried first
3. **Auto-Detection**: Used when `auto_detect_backend: enabled: true`
4. **Fallback Chain**: If first backend unavailable, tries next in list
5. **Validation**: All names must reference valid backends

**User Journey**:

1. User wants automatic backend selection based on context
2. User enables auto-detection: `auto_detect_backend: enabled: true`
3. User sets priority order: `backend_priority: [git, nextcloud-prod, sqlite]`
4. User runs command in git repository
5. todoat detects git backend available (finds TODO.md)
6. Uses git backend automatically
7. If not in git repo, falls back to nextcloud-prod.
8. If nextcloud-prod not responding, warn user, fallback to sqlite.

**Prerequisites**:
- 
- auto_detect_backend
```yml
auto_detect_backend: 
  enabled: true
  backend_priority: [git, nextcloud-prod, sqlite]

``` 
 to enable priority-based selection
- All backends in priority list must exist in `backends` map
- Override default_backend when enabled 

**Outputs/Results**:
- Context-aware backend selection
- Automatic fallback if preferred backend unavailable
- No need to specify `--backend` for common workflows

**Technical Details**:

**Configuration**:

```yaml
# Enable automatic backend detection
auto_detect_backend: 
  enabled: true
  backend_priority:
    - git              # Try git backend first (if in repo with TODO.md)
    - nextcloud-prod   # Fall back to Nextcloud
    - sqlite           # Last resort: local database


backends:
  git:
    type: git
    enabled: true
    auto_detect: true  # Required for auto-detection
    file: "TODO.md"

  nextcloud-prod:
    type: nextcloud
    enabled: true

  sqlite:
    type: sqlite
    enabled: true
```

**Selection Algorithm**:

```
FOR each backend in backend_priority:
  IF backend is enabled:
    IF backend can auto-detect (git, file):
      IF detection succeeds:
        RETURN backend
    ELSE:
      RETURN backend  # Use first non-auto-detecting backend
END

// No backends in priority list worked
RETURN default_backend OR first enabled backend
```

**Auto-Detection Capabilities**:

| Backend Type | Auto-Detection Method |
|--------------|----------------------|
| **Git** | Checks for git repository and presence of TODO.md (or fallback files) with marker comment |
| **File** | Checks if file exists at specified path |
| **Nextcloud** | Check if reachable |
| **Todoist** | Check if reachable |
| **SQLite** | No auto-detection (always available if enabled) |



**Related Features**:
- [Default Backend](#default-backend) - Alternative selection method
- [Backend System](backend-system.md) - Backend auto-detection
- [Multi-Backend Configuration](#multi-backend-configuration) - Configuring multiple backends

---

### Conflict Resolution Configuration

**Purpose**: Defines how synchronization conflicts are resolved when both local and remote backends have changed the same task, ensuring data consistency and preventing data loss.

**How It Works**:

1. **Global Setting**: `sync.conflict_resolution` applies to all sync operations
2. **Detection**: Conflicts detected via ETag mismatches during sync
3. **Strategy Application**: Selected strategy determines which changes are kept
4. **Metadata Tracking**: Sync metadata records resolution decisions
5. **User Transparency**: Conflict details logged (verbose mode)

**User Journey**:

1. User edits task in todoat (local change)
2. Task is also edited in Nextcloud web interface (remote change)
3. User runs `todoat sync` or autosync is enabled so it sync in background
4. Sync manager detects conflict (ETags don't match)
5. Configured strategy is applied to resolve conflict
6. Result is synced to both local and remote
7. User sees result based on strategy chosen

**Prerequisites**:
- [Synchronization](synchronization.md) enabled
- Conflicting changes to same task on local and remote

**Outputs/Results**:
- Consistent task state across local and remote
- Predictable conflict behavior
- Data preserved according to strategy

**Technical Details**:

**Configuration**:

```yaml
sync:
  enabled: true
  conflict_resolution: server_wins  # Choose strategy
```

**Available Strategies**:

| Strategy | Behavior | Use Case | Data Loss Risk |
|----------|----------|----------|----------------|
| **server_wins** | Remote changes override local changes | Trust remote as source of truth, collaborative environments | Local changes lost |
| **local_wins** | Local changes override remote changes | Offline-first workflow, local edits more important | Remote changes lost |
| **merge** | Intelligently combine local and remote changes | Best-effort preservation of all changes | Minimal (but possible edge cases) |
| **keep_both** | Create duplicate tasks for both versions | Never lose data, manual resolution later | Duplicates require cleanup |

**Strategy Details**:

**1. server_wins (Default)**:
```yaml
conflict_resolution: server_wins
```
- **Behavior**: Remote version is kept, local changes discarded
- **Use case**:
  - Multi-user environments
  - Remote is authoritative source
  - Prefer consistency over local changes
- **Example**:
  - Local: Changed priority to 1
  - Remote: Changed status to DONE
  - **Result**: Remote wins → Status is DONE, priority unchanged

**2. local_wins**:
```yaml
conflict_resolution: local_wins
```
- **Behavior**: Local version is kept, remote changes discarded
- **Use case**:
  - Single-user offline workflow
  - Local device is primary
  - Remote is just backup
- **Example**:
  - Local: Changed priority to 1
  - Remote: Changed status to DONE
  - **Result**: Local wins → Priority is 1, status unchanged

**3. merge**:
```yaml
conflict_resolution: merge
```
- **Behavior**: Combine non-conflicting fields, prefer local for conflicts
- **Use case**:
  - Want to preserve all changes when possible
  - Non-overlapping field changes common
  - Best-effort data preservation
- **Example**:
  - Local: Changed priority to 1
  - Remote: Changed status to DONE
  - **Result**: Merged → Priority is 1 AND status is DONE

**Merge Field Priority**:
```
If both changed summary: local wins
If both changed description: local wins
If only one changed field: keep that change
If different fields changed: merge both changes
```

**4. keep_both**:
```yaml
conflict_resolution: keep_both
```
- **Behavior**: Create two tasks (original + duplicate with suffix)
- **Use case**:
  - Never want to lose any changes
  - Manual conflict resolution preferred
  - Data preservation critical
- **Example**:
  - Local: Changed priority to 1
  - Remote: Changed status to DONE
  - **Result**: Two tasks created:
    - "Task name" (remote version - status DONE)
    - "Task name (conflict)" (local version - priority 1)

**Validation**:

```go
// Validate conflict resolution strategy
validStrategies := map[string]bool{
    "server_wins": true,
    "local_wins":  true,
    "merge":       true,
    "keep_both":   true,
}
if !validStrategies[c.Sync.ConflictResolution] {
    return fmt.Errorf("sync.conflict_resolution must be server_wins, local_wins, merge, or keep_both, got %q", c.Sync.ConflictResolution)
}
```

**Choosing a Strategy**:

```
Use server_wins if:
- Working in team with shared remote
- Remote backend is source of truth
- Local is just a cache

Use local_wins if:
- Working alone
- Offline-first workflow
- Remote is backup only

Use merge if:
- Want best of both worlds
- Typically edit different fields on each device
- Accept occasional merge quirks

Use keep_both if:
- Cannot afford to lose any data
- Want to manually review conflicts
- Don't mind duplicate cleanup
```

**Related Features**:
- [Synchronization](synchronization.md) - Comprehensive sync documentation
- [Sync Configuration](#sync-configuration) - Global sync settings

---

### Offline Mode Configuration

**Purpose**: Controls how todoat behaves when remote backends are unavailable, enabling continued operation offline with automatic operation queueing and retry.

**How It Works**:

1. **Mode Selection**: `sync.online_mode` determines behavior
2. **Connectivity Detection**: Checks remote availability before operations
3. **Queue Management**: Operations queued when offline (if auto mode)
4. **Retry Logic**: Queued operations retried when connectivity returns
5. **Transparency**: User sees clear feedback about offline status

**User Journey**:

**Auto Mode** (Recommended):
1. User is online, performs task operations normally
2. Network connection is lost (plane, tunnel, etc.)
3. User continues performing operations
4. Operations are queued locally
5. User runs `todoat sync status` to see queue
6. Connection returns
7. User runs `todoat sync` or waits for auto-sync
8. Queued operations are pushed to remote

**Online Mode**:
1. User forces online-only: `offline_mode: online`
2. Network connection is lost
3. User attempts task operation
4. Operation fails with error: "Remote backend unavailable"
5. User must wait for connection to return

**Prerequisites**:
- [Synchronization](synchronization.md) enabled

**Outputs/Results**:
- Continued operation during network outages (auto mode)
- Clear feedback about connection status
- Operation queue for retry when online
- Configurable behavior for different workflows

**Technical Details**:

**Configuration**:

```yaml
sync:
  enabled: true
  offline_mode: auto  # auto, online
```

**Available Modes**:

| Mode | Behavior | Use Case | Network Required? |
|------|----------|----------|-------------------|
| **auto** | Automatically detect and handle offline situations | Normal usage, unpredictable connectivity | No (queues operations) |
| **online** | Require connection, fail if unavailable | Server-authoritative, never work offline | Yes (fails without network) |

**Mode Details**:

**1. auto (Default - Recommended)**:
```yaml
offline_mode: auto
```
- **Behavior**:
  - Check remote availability before operations
  - If available: sync normally
  - If unavailable: queue operations locally
  - Retry queued operations when connection returns
- **Use case**:
  - Normal usage patterns
  - Laptops moving between networks
  - Mobile devices with intermittent connectivity
- **Advantages**:
  - Seamless offline/online transitions
  - Never blocks user operations
  - Automatic recovery

**2. online (Force Online)**:
```yaml
offline_mode: online
```
- **Behavior**:
  - Always require remote connection
  - Fail operations if remote unavailable
  - No operation queueing
- **Use case**:
  - Server is source of truth
  - Never want local-only changes
  - Collaborative environments
- **Error example**:
  ```
  Error: Remote backend "nextcloud-prod" is unavailable
  Cannot perform operation in online-only mode
  ```


**Queue Management**:

When offline (auto mode):
1. Operation requested (add, update, delete)
2. Change saved to local cache immediately
3. Operation added to `sync_queue` table
4. User sees success (operation completed locally)
5. Sync attempt made when connection returns
6. Queue entry marked completed or retried
7. UID is assigned by remote backend, sqlite use id to avoid generating uid 

**Sync Queue Table**:
```sql
CREATE TABLE sync_queue (
    id INTEGER PRIMARY KEY,
    operation TEXT NOT NULL,      -- 'create', 'update', 'delete'
    task_uid TEXT,
    retry_count INTEGER DEFAULT 0,
    last_attempt DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
```


**Default**: "auto" if not specified

**Checking Status**:

```bash
# View sync status and queue
todoat sync status
# Output:
#   Offline mode: auto
#   Remote status: Unavailable
#   Queued operations: 5

# View queued operations
todoat sync queue
# Output:
#   1. create "New task" (2 retries)
#   2. update "Updated task" (0 retries)
#   3. delete "Old task" (1 retry)
```

**Related Features**:
- [Synchronization](synchronization.md) - Sync system documentation
- [Sync Configuration](#sync-configuration) - Global sync settings
- [Sync Interval](#sync-interval) - Auto-sync frequency

---

### View Defaults

**Purpose**: Allows configuration of the default view used for task display, enabling users to customize which fields and formatting are shown by default without specifying `-v` flag each time.

**How It Works**:

1. **Configuration Field**: (Currently not in Config struct - feature planned)
2. **Default View**: Uses "default" view if not specified
3. **Override**: Can be overridden with `-v` flag on command line
4. **View Resolution**: Loads view definition from user config `~/.config/todoat/views/`. If default and all are not defined, use builtin all and default view

**User Journey** (Planned):

1. User creates custom view: `todoat view create myview`
2. User configures as default: `default_view: myview` in config
3. User runs commands without `-v` flag
4. Tasks displayed using custom view automatically
5. User can still use `-v other-view` to override

**Prerequisites**:
- [Views & Customization](views-customization.md) - Custom views
- View file exists: `~/.config/todoat/views/{view-name}.yaml`

**Outputs/Results**:
- Consistent task display across commands
- No need to specify `-v` flag repeatedly
- Easy to change default display format

**Technical Details**:

**Current Behavior**:
- Default view is "default"
- Built-in views: "default", "all"
- Custom views selected with `-v` flag

**Planned Configuration**:

```yaml
# Future configuration option
default_view: myview  # Use custom view by default

# Or stay with built-in
default_view: all     # Show all fields by default
```

**View Resolution Order** (Planned):

1. **Command-line flag**: `-v custom-view` (highest priority)
2. **Config default**: `default_view: myview`
3. **Built-in default**: "default" view (fallback)

**Usage Examples**:

```bash
# Uses configured default view
todoat MyList

# Override with specific view
todoat MyList -v all

# Override with custom view
todoat MyList -v minimal
```

**View Storage Location**:
- Built-in views: Embedded in binary
- User views: `~/.config/todoat/views/`
- Custom default view must exist in user views directory

**Related Features**:
- [Views & Customization](views-customization.md) - Complete view system documentation
- [CLI Interface](cli-interface.md) - View selection via `-v` flag

---

### No-Prompt Mode Configuration

**Purpose**: Configures non-interactive mode for scripting and automation, allowing todoat to be used in scripts, CI/CD pipelines, and other automated environments without requiring user input.

**How It Works**:

1. **Config Setting**: Set `no_prompt: true` in config to disable prompts by default
2. **CLI Override**: `-y` or `--no-prompt` flag overrides config setting
3. **Test Default**: Application tests use `no_prompt: true` for deterministic behavior
4. **Behavior**: When enabled, all interactive prompts are bypassed with deterministic output

**Configuration**:

```yaml
# Disable prompts for scripting (default: false)
no_prompt: false
```

**Config Values**:

| Value | Meaning |
|-------|---------|
| `false` (default) | Interactive prompts enabled - user interaction required |
| `true` | Non-interactive mode - all prompts bypassed |

**Behavior When Enabled**:

| Scenario | No-Prompt Behavior |
|----------|-------------------|
| Delete confirmation | Deletes immediately (force mode) |
| Single task match | Uses match automatically |
| Multiple task matches | Outputs match table + `ACTION_INCOMPLETE` |
| No list specified | Outputs available lists + `INFO_ONLY` |
| Config auto-init | Creates config silently |

**User Journey**:

```yaml
# For scripting: enable no-prompt in config
no_prompt: true
```

```bash
# Or override per-command with flag
todoat -y MyList delete "Task"

# Both produce same non-interactive behavior
```

**Prerequisites**:
- None - setting is optional with sensible default

**Outputs/Results**:
- All operations include result codes (`ACTION_COMPLETED`, `ACTION_INCOMPLETE`, `INFO_ONLY`)
- Deterministic, parseable output
- No stdin reads required

**Technical Details**:

```go
// Config field
NoPrompt bool `yaml:"no_prompt,omitempty"`

// Default value
const defaultNoPrompt = false

// CLI flag override
rootCmd.PersistentFlags().BoolP("no-prompt", "y", false, "Disable interactive prompts")
```

**Testing Configuration**:
- Test harness sets `no_prompt: true` by default
- Enables automated testing without mock stdin
- Tests can verify exact output without prompt variability

**Related Features**:
- [CLI Interface - No-Prompt Mode](cli-interface.md#no-prompt-mode) - Full no-prompt mode documentation
- [CLI Interface - Result Codes](cli-interface.md#result-codes) - Output status indicators
- [Output Format Configuration](#output-format-configuration) - JSON output for scripting

---

### Output Format Configuration

**Purpose**: Configures the default output format for command results, enabling machine-parseable JSON output for scripting and automation tools.

**How It Works**:

1. **Config Setting**: Set `output_format: json` in config for JSON output by default
2. **CLI Override**: `--json` flag overrides config setting
3. **Applies Globally**: Affects all command output
4. **Structured Data**: JSON output includes consistent field names and result codes

**Configuration**:

```yaml
# Set default output format (default: text)
output_format: text  # text | json
```

**Config Values**:

| Value | Meaning |
|-------|---------|
| `text` (default) | Human-readable text output with formatting |
| `json` | Machine-parseable JSON output |

**Output Format Comparison**:

| Aspect | Text Output | JSON Output |
|--------|-------------|-------------|
| Parsing | Regex/awk (fragile) | Native JSON parsing |
| Structure | Flat, tabular | Hierarchical, typed |
| Parent hierarchy | Awkward encoding | Clean arrays |
| Extensibility | Breaking changes likely | Add fields safely |
| Error handling | Parse stderr | Structured error object |

**User Journey**:

```yaml
# For automation: enable JSON output in config
output_format: json
```

```bash
# Or override per-command with flag
todoat --json MyList

# Parse output with jq
todoat --json MyList | jq '.tasks[].summary'
```

**JSON Output Examples**:

**List Tasks**:
```json
{
  "list": {"id": "abc-123", "name": "MyList", "task_count": 3},
  "tasks": [
    {"uid": "550e8400...", "summary": "Task 1", "status": "TODO", "priority": 1}
  ],
  "result": "INFO_ONLY"
}
```

**Add Task**:
```json
{
  "action": "add",
  "task": {"uid": "770e8400...", "summary": "New task", "status": "TODO"},
  "result": "ACTION_COMPLETED"
}
```

**Multiple Matches**:
```json
{
  "matches": [
    {"uid": "550e8400...", "summary": "Review PR", "parents": ["Project"]},
    {"uid": "660e8400...", "summary": "Code review", "parents": []}
  ],
  "result": "ACTION_INCOMPLETE",
  "message": "Multiple tasks match 'review'. Use --uid to specify exact task."
}
```

**Prerequisites**:
- None - setting is optional with sensible default

**Outputs/Results**:
- All output in valid JSON format
- Single JSON object per command
- UTF-8 encoded
- Parseable with `jq`, Python `json` module, or any JSON parser

**Technical Details**:

```go
// Config field
OutputFormat string `yaml:"output_format,omitempty" validate:"oneof=text json"`

// Default value
const defaultOutputFormat = "text"

// CLI flag override
rootCmd.PersistentFlags().Bool("json", false, "Output in JSON format")
```

**JSON Field Reference**:

| Field | Type | Description |
|-------|------|-------------|
| `result` | string | Result code: `ACTION_COMPLETED`, `ACTION_INCOMPLETE`, `INFO_ONLY`, `ERROR` |
| `action` | string | Operation performed: `add`, `update`, `complete`, `delete` |
| `task` | object | Task object for single-task operations |
| `tasks` | array | Array of task objects for list operations |
| `matches` | array | Array of matching tasks when ambiguous |
| `list` | object | List metadata for list-specific operations |
| `lists` | array | Array of list objects |
| `message` | string | Human-readable message: info for `ACTION_INCOMPLETE`, error details for `ERROR` (format: `Error #: description`) |
| `synced` | boolean | Whether task has been synced to remote (in task objects) |

**Error Codes**:

| Error # | Description |
|---------|-------------|
| 1 | Resource not found (list, task) |
| 2 | Invalid input (bad flag value, malformed date) |
| 3 | Backend connection failed |
| 4 | Permission denied |
| 5 | Sync conflict |
| 6 | Validation error |

**UID Format**:

| Format | Description |
|--------|-------------|
| `<uuid>` | Backend-assigned UID (e.g., `550e8400-e29b-41d4-a716-446655440000`) |
| `NOT-SYNCED-<id>` | SQLite internal ID for unsynced tasks (e.g., `NOT-SYNCED-42`) |

Tasks created locally but not yet synced to the remote backend will have `"synced": false` and a UID in `NOT-SYNCED-<sqlite-id>` format. This allows scripts to operate on unsynced tasks using `--uid "NOT-SYNCED-42"`. After sync, the UID becomes the backend-assigned UUID.

**Related Features**:
- [CLI Interface - JSON Output Mode](cli-interface.md#json-output-mode) - Full JSON output documentation
- [CLI Interface - Result Codes](cli-interface.md#result-codes) - Output status indicators
- [No-Prompt Mode Configuration](#no-prompt-mode-configuration) - Pairs well for scripting

---

### Singleton Pattern

**Purpose**: Ensures only one configuration instance exists throughout the application lifetime, preventing inconsistent config states and enabling thread-safe config access.

**How It Works**:

1. **sync.Once**: Go's thread-safe initialization primitive
2. **Global Instance**: Single `globalConfig` variable
3. **Lazy Loading**: Config loaded on first `GetConfig()` call
4. **Thread Safety**: Concurrent calls wait for initialization
5. **Immutability**: Config loaded once, not reloaded during runtime

**User Journey**:

1. Application starts
2. Multiple goroutines may call `GetConfig()` simultaneously
3. First call triggers config loading (others wait)
4. Subsequent calls return cached config instance instantly
5. All parts of application use same config
6. No inconsistencies from multiple config loads

**Prerequisites**:
- None - singleton pattern is internal implementation detail

**Outputs/Results**:
- Consistent configuration across entire application
- Thread-safe config access
- Efficient - config loaded only once
- Predictable behavior

**Technical Details**:

**Implementation**:

```go
// Package-level variables
var configOnce sync.Once        // Ensures single initialization
var globalConfig *Config        // The singleton instance
var customConfigPath string     // Custom path (if set)

// GetConfig returns the singleton config instance
func GetConfig() *Config {
    configOnce.Do(func() {
        config, err := loadUserOrSampleConfig()
        if err != nil {
            log.Fatal(err)
        }
        globalConfig = config
    })
    return globalConfig
}
```

**Key Components**:

**1. sync.Once**:
- Go standard library synchronization primitive
- Guarantees exactly one execution of initialization function
- Thread-safe - multiple goroutines can call simultaneously
- First caller executes, others wait and receive result

**2. Global Instance**:
- `globalConfig` holds the loaded configuration
- Shared across entire application
- Read-only after initialization (no mutations)

**3. Lazy Loading**:
- Config not loaded until first `GetConfig()` call
- Allows setting custom path before loading
- Reduces startup time if config not needed

**Resetting Singleton** (Testing/Custom Path):

```go
// SetCustomConfigPath resets singleton for new path
func SetCustomConfigPath(path string) {
    customConfigPath = path

    // Reset singleton to force reload
    configOnce = sync.Once{}
    globalConfig = nil
}

// SetConfigForTest allows test-specific config
func SetConfigForTest(cfg *Config) {
    globalConfig = cfg
}
```

**Usage Pattern**:

```go
// Anywhere in application
import "todoat/internal/config"

func someFunction() {
    cfg := config.GetConfig()  // Always returns same instance
    backend := cfg.GetDefaultBackend()
    // ... use config ...
}

// In another goroutine simultaneously
func anotherFunction() {
    cfg := config.GetConfig()  // Same instance, thread-safe
    backends := cfg.GetEnabledBackends()
    // ... use config ...
}
```

**Benefits**:

1. **Consistency**: All code sees same configuration
2. **Performance**: Config loaded once, no repeated parsing
3. **Thread Safety**: Safe for concurrent access
4. **Simplicity**: No need to pass config around
5. **Predictability**: Initialization controlled, no surprises

**Trade-offs**:

1. **Testing**: Requires reset mechanism for tests
2. **Reloading**: Can't reload config without restart (by design)
3. **Global State**: Package-level state (acceptable for config)

**Comparison with Alternatives**:

| Approach | Pros | Cons | todoat Choice |
|----------|------|------|-------------------|
| **Singleton** | Simple, efficient, consistent | Global state | ✅ Used |
| **Dependency Injection** | Testable, flexible | Complex, verbose | Not needed |
| **Context Passing** | Explicit, testable | Tedious, repetitive | Overkill |

**Related Features**:
- [Config Validation](#config-validation) - Validated on singleton initialization
- [Custom Config Path](#custom-config-path) - Can reset singleton for new path

---

## Configuration Examples

### Minimal Configuration (SQLite Only)

```yaml
backends:
  sqlite:
    type: sqlite
    enabled: true
    db_path: ""  # XDG default

default_backend: sqlite

canWriteConfig: true
ui: cli
```

### Nextcloud with Keyring (Recommended)

```yaml
backends:
  nextcloud-prod:
    type: nextcloud
    enabled: true
    host: "nextcloud.example.com"
    username: "myuser"
    # Password in keyring: todoat credentials set nextcloud myuser --prompt

default_backend: nextcloud-prod

sync:
  enabled: false

canWriteConfig: true
ui: cli
```

### Multi-Backend with Sync

```yaml
backends:
  nextcloud-prod:
    type: nextcloud
    enabled: true
    host: "nextcloud.example.com"
    username: "myuser"

  todoist:
    type: todoist
    enabled: true
    username: "token"

  sqlite:
    type: sqlite
    enabled: true

default_backend: nextcloud-prod

sync:
  enabled: true
  auto_sync_after_operation: true
  local_backend: sqlite
  conflict_resolution: server_wins
  offline_mode: auto

canWriteConfig: true
ui: cli
```

### Developer Workflow (Git-First)

```yaml
backends:
  git:
    type: git
    enabled: true
    file: "TODO.md"
    auto_detect: true
    fallback_files:
      - "todo.md"
      - ".todoat.md"
    auto_commit: false

  nextcloud-work:
    type: nextcloud
    enabled: true
    host: "nextcloud.work.com"
    username: "dev-user"

auto_detect_backend: 
  enabled: true
  backend_priority:
      - git
      - nextcloud-work

sync:
  enabled: false

canWriteConfig: true
ui: cli
```

### Work/Personal Separation

```yaml
backends:
  work-nextcloud:
    type: nextcloud
    enabled: true
    host: "nextcloud.work.com"
    username: "work-user"

  personal-todoist:
    type: todoist
    enabled: true
    username: "token"

default_backend: work-nextcloud

# Use --backend personal-todoist for personal tasks

canWriteConfig: true
ui: cli
```

### Offline-First with Manual Sync

```yaml
backends:
  sqlite:
    type: sqlite
    enabled: true
    db_path: "~/tasks/offline.db"

  nextcloud-prod:
    type: nextcloud
    enabled: true
    host: "nextcloud.example.com"
    username: "myuser"

default_backend: nextcloud-prod

sync:
  enabled: true
  local_backend: sqlite
  conflict_resolution: local_wins        # Prefer local changes
  auto_sync_after_operation: false       # Manual sync only
  offline_mode: auto

canWriteConfig: true
ui: cli
```

---

## Environment Variables

Configuration can be supplemented or overridden with environment variables:

### XDG Directories

| Variable | Purpose | Default |
|----------|---------|---------|
| `XDG_CONFIG_HOME` | Configuration directory | `~/.config` |
| `XDG_DATA_HOME` | Data directory | `~/.local/share` |
| `XDG_CACHE_HOME` | Cache directory | `~/.cache` |

### Backend Credentials

Format: `TODOAT_{BACKEND}_{FIELD}`

**Nextcloud**:
```bash
export TODOAT_NEXTCLOUD_HOST="nextcloud.example.com"
export TODOAT_NEXTCLOUD_USERNAME="myuser"
export TODOAT_NEXTCLOUD_PASSWORD="secret"
```

**Todoist**:
```bash
export TODOAT_TODOIST_TOKEN="api-token-here"
```

**Environment Variable Priority**:
1. Keyring (highest priority)
2. Environment variables
3. Config file (lowest priority)

See [Credential Management](credential-management.md) for detailed credential configuration.

---

## Related Features

- **[Backend System](backend-system.md)** - Backend architecture and selection
- **[Synchronization](synchronization.md)** - Sync system and conflict resolution
- **[Credential Management](credential-management.md)** - Secure credential storage
- **[Views & Customization](views-customization.md)** - Custom view configuration
- **[CLI Interface](cli-interface.md)** - Command-line options and flags
- **[List Management](list-management.md)** - List caching behavior

---

**Last Updated**: January 2026
**Configuration Version**: 1.0
**Total Features Documented**: 16
