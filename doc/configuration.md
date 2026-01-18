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

# Disable interactive prompts by default
no_prompt: false

# Default output format (text or json)
output_format: text

# Synchronization configuration
sync:
  enabled: false
  local_backend: sqlite
  conflict_resolution: local

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
```

### Configuration Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `backends.sqlite.enabled` | boolean | `true` | Enable SQLite backend |
| `backends.sqlite.path` | string | `~/.local/share/todoat/tasks.db` | Path to SQLite database |
| `default_backend` | string | `sqlite` | Which backend to use |
| `no_prompt` | boolean | `false` | Disable interactive prompts |
| `output_format` | string | `text` | Default output format (`text` or `json`) |
| `sync.enabled` | boolean | `false` | Enable synchronization with remote backends |
| `sync.local_backend` | string | `sqlite` | Local backend to use for caching |
| `sync.conflict_resolution` | string | `local` | Conflict resolution strategy (`local` or `remote`) |
| `notifications.enabled` | boolean | `true` | Enable notification system |
| `notifications.os.enabled` | boolean | `true` | Enable OS desktop notifications |
| `notifications.os.on_sync_complete` | boolean | `true` | Notify when sync completes |
| `notifications.os.on_sync_error` | boolean | `true` | Notify when sync fails |
| `notifications.os.on_conflict` | boolean | `true` | Notify when conflicts are detected |
| `notifications.log.enabled` | boolean | `true` | Enable logging notifications to file |
| `notifications.log.path` | string | `~/.local/share/todoat/notifications.log` | Path to notification log file |
| `notifications.log.max_size_mb` | integer | `10` | Maximum log file size in MB |
| `notifications.log.retention_days` | integer | `30` | Days to retain log entries |

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

Note: For better security, consider using the credential manager instead of environment variables. See [Credential Management](./commands.md#credential-management).

## Validation

The configuration file is validated on load. Invalid configurations will produce an error message. Validation checks:

- `output_format` must be `text` or `json`
- `default_backend` must be a known backend (`sqlite`)
- The default backend must be enabled

---
*Last updated: 2026-01-18*
