# [055] Comprehensive Sample Configuration File

## Summary
Create a fully-documented sample configuration file (config.sample.yaml) with all backend types, options, and inline comments that gets embedded in the binary and copied to users on first run.

## Documentation Reference
- Primary: `docs/explanation/todo.md`
- Related: `docs/explanation/configuration.md`
- Related: `docs/explanation/backend-system.md`

## Dependencies
- Requires: [010] Configuration System

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestSampleConfigEmbedded` - config.sample.yaml is embedded in binary via go:embed
- [ ] `TestSampleConfigCopyOnFirstRun` - First run copies sample to ~/.config/todoat/config.yaml
- [ ] `TestSampleConfigAllBackends` - Sample includes examples for all backend types (sqlite, nextcloud, todoist, git, file)
- [ ] `TestSampleConfigComments` - Sample contains inline YAML comments explaining each option

### Functional Requirements
- [ ] config.sample.yaml exists in project root or internal/config/
- [ ] Sample includes complete Nextcloud configuration with TLS options
- [ ] Sample includes complete Todoist configuration with token placeholder
- [ ] Sample includes SQLite configuration with path options
- [ ] Sample includes Git backend configuration with auto-commit options
- [ ] Sample includes sync configuration with all strategies documented
- [ ] Sample includes view defaults configuration
- [ ] Sample includes notification configuration
- [ ] All options have brief inline comments explaining their purpose
- [ ] Credentials show keyring-based and environment variable patterns

## Implementation Notes

### Sample Config Structure
```yaml
# todoat Configuration File
# This file is automatically created on first run.
# Edit values as needed for your setup.

# Backend configurations - enable and configure backends you want to use
backends:
  # SQLite backend - local database storage (recommended default)
  sqlite:
    type: sqlite
    enabled: true
    # db_path: "~/.local/share/todoat/tasks.db"  # Optional: custom path

  # Nextcloud backend - sync with Nextcloud Tasks via CalDAV
  # nextcloud:
  #   type: nextcloud
  #   enabled: false
  #   host: "nextcloud.example.com"
  #   username: "your-username"
  #   # Password stored in system keyring or TODOAT_NEXTCLOUD_PASSWORD env var
  #   # TLS options (uncomment if using self-signed certificates):
  #   # insecure_skip_verify: true
  #   # suppress_ssl_warning: true

  # Todoist backend - sync with Todoist
  # todoist:
  #   type: todoist
  #   enabled: false
  #   username: "token"  # Fixed value for Todoist
  #   # API token stored in keyring as TODOAT_TODOIST_TOKEN or env var

  # Git backend - store tasks in markdown files in a git repository
  # git:
  #   type: git
  #   enabled: false
  #   file: "TODO.md"
  #   fallback_files: ["todo.md", ".todoat.md"]
  #   auto_detect: true
  #   auto_commit: false

# Default backend when multiple are enabled
default_backend: sqlite

# Auto-detect backend based on current directory (e.g., git repo)
auto_detect_backend: false

# Backend priority for selection when auto-detect enabled
# backend_priority:
#   - git
#   - sqlite

# Synchronization settings
sync:
  enabled: false
  # local_backend: sqlite  # Cache backend for remote syncing
  # conflict_resolution: server_wins  # server_wins | local_wins | merge | keep_both
  # offline_mode: auto  # auto | online | offline

# User interface settings
no_prompt: false  # Set to true for scripting (no interactive prompts)
output_format: text  # text | json

# Default view for task display (omit for built-in "default" view)
# default_view: "my-custom-view"

# Notification settings
# notification:
#   enabled: true
#   os_notification:
#     enabled: true
#     on_sync_error: true
#     on_conflict: true
#   log_notification:
#     enabled: true
#     # path: "~/.local/share/todoat/notifications.log"
```

### Files to Create/Modify
1. Create `internal/config/config.sample.yaml` with comprehensive examples
2. Modify `internal/config/config.go` to embed sample via `//go:embed config.sample.yaml`
3. Update first-run logic to copy sample (not just use defaults)

## Out of Scope
- Config validation error suggestions (covered elsewhere)
- Config migration between versions (separate item if needed)
- Interactive config wizard
