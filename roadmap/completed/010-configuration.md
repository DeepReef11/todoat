# 010: Configuration System

## Summary
Implement YAML-based configuration system following XDG Base Directory Specification for storing user preferences and backend settings.

## Documentation Reference
- Primary: `docs/explanation/configuration.md`
- Sections: YAML Configuration, XDG Compliance, Auto-Initialization

## Dependencies
- Requires: 001-project-setup.md (internal/config directory)
- Requires: 003-sqlite-backend.md (backend to configure)
- Blocked by: none

## Complexity
**M (Medium)** - File system operations, YAML parsing, XDG path resolution, embedded defaults

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestConfigAutoCreate` - First run creates config file at XDG path with defaults
- [ ] `TestConfigCustomPath` - `todoat --config /path/to/config.yaml` uses specified config
- [ ] `TestConfigDatabasePath` - Database path from config is used for SQLite backend
- [ ] `TestConfigNoPromptDefault` - `no_prompt: true` in config enables no-prompt mode globally
- [ ] `TestConfigFlagOverride` - CLI flags override config values (e.g., `-y` overrides `no_prompt: false`)
- [ ] `TestConfigInvalid` - Invalid YAML returns clear error message
- [ ] `TestConfigMissingBackend` - Missing backend config returns helpful error

### Unit Tests (if needed)
- [ ] XDG path expansion works on Linux/macOS/Windows
- [ ] Environment variable expansion in paths (`$HOME`, `~`)
- [ ] Config validation catches invalid values

### Manual Verification
- [ ] Config file created at `~/.config/todoat/config.yaml` on first run
- [ ] Database created at `~/.local/share/todoat/tasks.db` by default
- [ ] Editing config file changes application behavior on next run

## Implementation Notes

### XDG Paths
```
Config: $XDG_CONFIG_HOME/todoat/config.yaml (default: ~/.config/todoat/config.yaml)
Data:   $XDG_DATA_HOME/todoat/ (default: ~/.local/share/todoat/)
Cache:  $XDG_CACHE_HOME/todoat/ (default: ~/.cache/todoat/)
```

### Default Config Structure
```yaml
# todoat configuration
backends:
  sqlite:
    enabled: true
    path: "~/.local/share/todoat/tasks.db"

default_backend: sqlite
no_prompt: false
output_format: text  # text or json
```

### Required Changes
1. Create `internal/config/config.go` with Config struct
2. Implement XDG path resolution (use `os.UserConfigDir()` as base)
3. Embed default config as Go embed
4. Auto-create config on first run if missing
5. Parse config at startup before command execution
6. Wire config values to relevant components

### Config Loading Order
1. Embedded defaults
2. Config file (if exists)
3. Environment variables (TODOAT_*)
4. CLI flags (highest priority)

### Path Expansion
- `~` → user home directory
- `$HOME` → user home directory
- `$XDG_*` → XDG environment variables

## Out of Scope
- Config editing via CLI (`todoat config set key value`) - separate item
- Multiple backend configurations - separate item (sync system)
- View configuration - separate item
- Config file encryption - not planned
