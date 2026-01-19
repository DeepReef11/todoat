# Configuration

todoat supports configuration through both a YAML configuration file and command-line flags.

## Configuration File

### Location

todoat follows the XDG Base Directory specification. The configuration file is located at:

```
$XDG_CONFIG_HOME/todoat/config.yaml
```

If `XDG_CONFIG_HOME` is not set, it defaults to:

```
~/.config/todoat/config.yaml
```

The configuration file is created automatically with defaults on first use if it doesn't exist.

### Configuration Options

```yaml
# todoat configuration

# Backend configuration
backends:
  sqlite:
    enabled: true
    path: ~/.local/share/todoat/tasks.db

# Default backend to use
default_backend: sqlite

# Default view for task display (optional)
# Can be a built-in view ("default", "all") or a custom view name
default_view: default

# Disable interactive prompts by default
no_prompt: false

# Default output format (text or json)
output_format: text

# Enable automatic backend detection based on directory context
auto_detect_backend: false

# Trash configuration
trash:
  retention_days: 30  # Days before auto-purge (0 = disabled)

# Synchronization configuration
sync:
  enabled: false
  local_backend: sqlite
  conflict_resolution: local
  offline_mode: auto  # auto | online | offline
  connectivity_timeout: 5s  # timeout for connectivity checks

# Notification configuration
notifications:
  enabled: true
  os:
    enabled: true
    on_sync_complete: true
    on_sync_error: true
    on_conflict: true
  log:
    enabled: true
    path: ~/.local/share/todoat/notifications.log
    max_size_mb: 10
    retention_days: 30
  reminder:
    enabled: true
    intervals:
      - "1 day"
      - "1 hour"
      - "at due time"
    os_notification: true
    log_notification: true
```

### Configuration Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `backends.sqlite.enabled` | boolean | `true` | Enable SQLite backend |
| `backends.sqlite.path` | string | `~/.local/share/todoat/tasks.db` | Path to SQLite database |
| `default_backend` | string | `sqlite` | Which backend to use |
| `default_view` | string | `default` | Default view for task display (built-in or custom view name) |
| `no_prompt` | boolean | `false` | Disable interactive prompts |
| `output_format` | string | `text` | Default output format (`text` or `json`) |
| `auto_detect_backend` | boolean | `false` | Enable automatic backend detection (see below) |
| `trash.retention_days` | integer | `30` | Days before auto-purging deleted lists (0 = disabled) |
| `sync.enabled` | boolean | `false` | Enable synchronization with remote backends |
| `sync.local_backend` | string | `sqlite` | Local backend to use for caching |
| `sync.conflict_resolution` | string | `local` | Conflict resolution strategy (`local` or `remote`) |
| `sync.offline_mode` | string | `auto` | Offline behavior: `auto`, `online`, or `offline` |
| `sync.connectivity_timeout` | string | `5s` | Timeout for connectivity checks in auto mode |
| `notifications.enabled` | boolean | `true` | Enable notification system |
| `notifications.os.enabled` | boolean | `true` | Enable OS desktop notifications |
| `notifications.os.on_sync_complete` | boolean | `true` | Notify when sync completes |
| `notifications.os.on_sync_error` | boolean | `true` | Notify when sync fails |
| `notifications.os.on_conflict` | boolean | `true` | Notify when conflicts are detected |
| `notifications.log.enabled` | boolean | `true` | Enable logging notifications to file |
| `notifications.log.path` | string | `~/.local/share/todoat/notifications.log` | Path to notification log file |
| `notifications.log.max_size_mb` | integer | `10` | Maximum log file size in MB |
| `notifications.log.retention_days` | integer | `30` | Days to retain log entries |
| `notifications.reminder.enabled` | boolean | `true` | Enable task due date reminders |
| `notifications.reminder.intervals` | list | `["1 day", "1 hour", "at due time"]` | When to send reminders before due date |
| `notifications.reminder.os_notification` | boolean | `true` | Send reminder via OS notification |
| `notifications.reminder.log_notification` | boolean | `true` | Log reminders to notification log |

### Path Expansion

Paths in the configuration file support:
- `~` expansion to home directory
- Environment variable expansion (e.g., `$HOME`)

## Data Storage

### Default Location

todoat stores its data following XDG conventions:

```
$XDG_DATA_HOME/todoat/
└── tasks.db    # SQLite database
```

If `XDG_DATA_HOME` is not set, it defaults to:

```
~/.local/share/todoat/tasks.db
```

Note: The legacy location `~/.todoat/todoat.db` is also supported.

The directory and database file are created automatically on first use.

## Command-Line Options

Command-line flags override configuration file settings.

### Global Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--no-prompt` | `-y` | Disable interactive prompts | `false` |
| `--verbose` | `-V` | Enable verbose/debug output | `false` |
| `--json` | | Output in JSON format | `false` |

### Example Usage

```bash
# Run in non-interactive mode (for scripts)
todoat -y MyList add "Task"

# Get verbose output
todoat -V MyList

# Get JSON output
todoat MyList get --json
```

## Environment Variables

todoat respects these environment variables:

### XDG Directories

| Variable | Description |
|----------|-------------|
| `XDG_CONFIG_HOME` | Configuration directory (default: `~/.config`) |
| `XDG_DATA_HOME` | Data directory (default: `~/.local/share`) |
| `XDG_CACHE_HOME` | Cache directory (default: `~/.cache`) |

### Backend Credentials

Backend authentication can be configured via environment variables. See [Backends](./backends.md) for detailed setup instructions.

**Nextcloud:**

| Variable | Description |
|----------|-------------|
| `TODOAT_NEXTCLOUD_HOST` | Nextcloud server hostname (e.g., `cloud.example.com`) |
| `TODOAT_NEXTCLOUD_USERNAME` | Nextcloud username |
| `TODOAT_NEXTCLOUD_PASSWORD` | Nextcloud password or app password |

**Todoist:**

| Variable | Description |
|----------|-------------|
| `TODOAT_TODOIST_TOKEN` | Todoist API token |

**Google Tasks:**

| Variable | Description |
|----------|-------------|
| `TODOAT_GOOGLE_ACCESS_TOKEN` | Google OAuth2 access token |
| `TODOAT_GOOGLE_REFRESH_TOKEN` | Google OAuth2 refresh token (optional) |
| `TODOAT_GOOGLE_CLIENT_ID` | Google OAuth2 client ID (optional) |
| `TODOAT_GOOGLE_CLIENT_SECRET` | Google OAuth2 client secret (optional) |

**Microsoft To-Do:**

| Variable | Description |
|----------|-------------|
| `TODOAT_MSTODO_ACCESS_TOKEN` | Microsoft OAuth2 access token |
| `TODOAT_MSTODO_REFRESH_TOKEN` | Microsoft OAuth2 refresh token (optional) |
| `TODOAT_MSTODO_CLIENT_ID` | Microsoft OAuth2 client ID (optional) |
| `TODOAT_MSTODO_CLIENT_SECRET` | Microsoft OAuth2 client secret (optional) |

Note: For better security, consider using the credential manager instead of environment variables. See [Credential Management](./commands.md#credential-management).

## Trash Auto-Purge

Deleted lists are automatically purged from the trash after a configurable retention period. This prevents indefinite data accumulation.

### How It Works

- Lists deleted with `todoat list delete` are moved to trash
- When you access the trash (e.g., `todoat list trash`), lists older than the retention period are automatically purged
- Purged lists and their tasks are permanently deleted (no recovery)

### Configuration

```yaml
trash:
  retention_days: 30  # Default: purge after 30 days
```

| Value | Behavior |
|-------|----------|
| `30` (default) | Purge lists deleted more than 30 days ago |
| `7` | Purge lists deleted more than 7 days ago |
| `0` | Disable auto-purge (keep lists in trash indefinitely) |

### Example

```bash
# Delete a list (moved to trash)
todoat list delete "OldProject"

# 30+ days later, when you view trash
todoat list trash
# Output: Auto-purged 1 list(s) older than 30 days

# To disable auto-purge, set retention_days to 0 in config
```

## Offline Mode

Control how todoat behaves when remote backends are unavailable. This is useful for working without network connectivity or to prevent unintended remote operations.

### Modes

| Mode | Description |
|------|-------------|
| `auto` (default) | Automatically detect connectivity; queue operations when offline |
| `online` | Require backend connectivity; fail with error if unavailable |
| `offline` | Never contact remote backends; always use local cache and queue operations |

### Configuration

```yaml
sync:
  enabled: true
  offline_mode: auto  # auto | online | offline
  connectivity_timeout: 5s  # timeout for connectivity checks
```

### Mode Behaviors

| Mode | Backend Available | Backend Unavailable |
|------|------------------|---------------------|
| `auto` | Direct operation | Queue + local cache |
| `online` | Direct operation | Error + suggestion |
| `offline` | Queue always | Queue always |

### Connectivity Timeout

The `connectivity_timeout` setting controls how long todoat waits when checking backend connectivity in `auto` mode. Shorter timeouts provide faster fallback to offline mode but may incorrectly detect slow connections as offline.

```yaml
sync:
  connectivity_timeout: 5s   # Default: 5 seconds
  connectivity_timeout: 1s   # Faster fallback
  connectivity_timeout: 30s  # More tolerant of slow connections
```

### Checking Current Mode

View the current offline mode status:

```bash
todoat sync status
```

Output includes the configured offline mode:
```
Sync Status:

Backend: sqlite
  Last Sync: 2026-01-18 14:30:00
  Pending Operations: 0
  Offline Mode: auto
```

### Use Cases

- **`auto`**: Best for most users. Works offline when needed, syncs when possible.
- **`online`**: Use when you need guarantee that operations reach the remote backend.
- **`offline`**: Use for airplane mode, metered connections, or to batch sync operations.

## Backend Auto-Detection

When `auto_detect_backend` is enabled, todoat automatically detects which backend to use based on your current directory context. This is useful when working in project directories that have their own task files.

### How It Works

1. todoat checks the current directory for detectable backends
2. Backends are checked in priority order (Git has highest priority)
3. The first available backend is used instead of the default backend

### Supported Detectable Backends

| Backend | Detection Criteria |
|---------|-------------------|
| Git/Markdown | Current directory is in a Git repository with a TODO.md, todo.md, or .todoat.md file containing `<!-- todoat:enabled -->` |
| SQLite | Always available as fallback |

### Enabling Auto-Detection

Add to your config file:

```yaml
auto_detect_backend: true
```

### Checking Detection Results

Use the `--detect-backend` flag to see which backends are detected:

```bash
todoat --detect-backend
```

This shows all detected backends and which one would be selected.

### Example Workflow

```bash
# Enable auto-detection in config
# ~/.config/todoat/config.yaml:
# auto_detect_backend: true

# In a project with TODO.md containing <!-- todoat:enabled -->
cd ~/projects/myproject
todoat MyList add "Project task"  # Uses Git/Markdown backend

# Outside the project
cd ~
todoat MyList add "Personal task"  # Uses SQLite backend (default)
```

## Validation

The configuration file is validated on load. Invalid configurations will produce an error message. Validation checks:

- `output_format` must be `text` or `json`
- `default_backend` must be a known backend (`sqlite`)
- The default backend must be enabled

---
*Last updated: 2026-01-19*
