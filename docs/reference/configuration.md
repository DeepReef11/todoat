# Configuration

This guide covers managing todoat configuration using the `config` command.

## Configuration Location

todoat follows the XDG Base Directory specification:

- **Linux/macOS**: `~/.config/todoat/config.yaml`
- **Windows**: `%APPDATA%\todoat\config.yaml`

## Data Locations

todoat stores data in multiple locations:

| Data | Location | Description |
|------|----------|-------------|
| Configuration | `~/.config/todoat/config.yaml` | User settings |
| Analytics | `~/.config/todoat/analytics.db` | Local usage statistics |
| Sync Queue | `~/.todoat/todoat.db` | Pending sync operations |
| Backend Caches | `~/.local/share/todoat/caches/` | Cached remote backend data |
| Default SQLite | `~/.local/share/todoat/tasks.db` | Local sqlite backend tasks |

See [Synchronization - Database Locations](../explanation/synchronization.md#database-locations) for details on sync-related databases.

## Viewing Configuration

### Show Config Path

```bash
todoat config path
```

Output:
```
/home/user/.config/todoat/config.yaml
```

### Show All Configuration

```bash
todoat config get
```

Displays the entire configuration as YAML.

### Show Specific Value

Use dot notation for nested values:

```bash
# Get default backend
todoat config get default_backend

# Get sync enabled status
todoat config get sync.enabled

# Get backend-specific setting
todoat config get backends.nextcloud.host

# Get all sync settings (including daemon)
todoat config get sync

# Get daemon settings
todoat config get sync.daemon

# Get reminder settings
todoat config get reminder

# Get background pull cooldown
todoat config get sync.background_pull_cooldown
```

### JSON Output

```bash
todoat --json config get
todoat --json config get sync.enabled
```

## Modifying Configuration

### Set a Value

```bash
# Enable no-prompt mode
todoat config set no_prompt true

# Set default backend
todoat config set default_backend sqlite

# Set nested values with dot notation
todoat config set sync.enabled true
todoat config set sync.offline_mode auto

# Set daemon configuration
todoat config set sync.daemon.enabled true
todoat config set sync.daemon.interval 60

# Set background pull cooldown
todoat config set sync.background_pull_cooldown "1m"

# Set reminder configuration
todoat config set reminder.enabled true
todoat config set reminder.os_notification false
todoat config set reminder.intervals "1d,1h,at due time"
```

### Validation

Values are validated before saving:

```bash
$ todoat config set no_prompt invalid
Error: invalid value for no_prompt: must be true or false
```

Boolean fields accept: `true`/`false`, `yes`/`no`, `1`/`0`

### Path Values

Paths support expansion:

```bash
todoat config set backends.sqlite.path "~/my-tasks/tasks.db"
```

## Editing Configuration

### Open in Editor

```bash
todoat config edit
```

Opens the config file in your system editor (`$EDITOR` or `vi`).

### Reset to Defaults

```bash
todoat config reset
```

Restores the default configuration. Requires confirmation.

## Common Configuration Options

| Key | Type | Description |
|-----|------|-------------|
| `default_backend` | string | Default backend name |
| `auto_detect_backend` | bool | Auto-detect backend based on current directory |
| `default_view` | string | Default view for task display |
| `no_prompt` | bool | Non-interactive mode |
| `output_format` | string | Default output format (`text` or `json`) |
| `sync.enabled` | bool | Enable synchronization |
| `sync.local_backend` | string | Cache backend for remote syncing |
| `sync.offline_mode` | string | CLI backend mode: `auto`/`offline` (use SQLite cache) or `online` (direct remote) |
| `sync.conflict_resolution` | string | `server_wins`, `local_wins`, `merge`, or `keep_both` |
| `sync.connectivity_timeout` | string | Network timeout for connectivity checks (default: `5s`) |
| `sync.auto_sync_after_operation` | bool | Auto-sync after add/update/delete operations (default: `true` when sync enabled) |
| `sync.background_pull_cooldown` | string | Cooldown between background pull syncs (default: `30s`, minimum: `5s`) |
| `sync.daemon.enabled` | bool | Enable background sync daemon (default: `false`) |
| `sync.daemon.interval` | int | Daemon sync interval in seconds (default: `300`) |
| `sync.daemon.idle_timeout` | int | Seconds of idle time before daemon exits (default: `300`) |
| `trash.retention_days` | int | Days to keep deleted items (default: `30`, 0 = forever) |
| `analytics.enabled` | bool | Enable command usage tracking (default: `true`) |
| `analytics.retention_days` | int | Days to keep analytics data (0 = forever) |
| `reminder.enabled` | bool | Enable task reminder notifications (default: `false`) |
| `reminder.intervals` | list | Time before due to send reminders (default: `[]`, no intervals) |
| `reminder.os_notification` | bool | Send reminders via OS notifications (default: `false`) |
| `reminder.log_notification` | bool | Log reminders to notification log (default: `false`) |

## Backend Configuration

Each backend has its own configuration keys under `backends.<name>`.

### SQLite

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `backends.sqlite.enabled` | bool | `true` | Enable SQLite backend |
| `backends.sqlite.path` | string | `~/.local/share/todoat/tasks.db` | Database file path |

### Nextcloud

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `backends.nextcloud.enabled` | bool | `false` | Enable Nextcloud backend |
| `backends.nextcloud.host` | string | | Nextcloud server hostname |
| `backends.nextcloud.username` | string | | CalDAV username |
| `backends.nextcloud.insecure_skip_verify` | bool | `false` | Accept self-signed certificates |
| `backends.nextcloud.allow_http` | bool | `false` | Allow HTTP (non-HTTPS) connections |

### Todoist

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `backends.todoist.enabled` | bool | `false` | Enable Todoist backend |
| `backends.todoist.username` | string | `"token"` | Fixed value for Todoist auth |

### Google Tasks

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `backends.google.enabled` | bool | `false` | Enable Google Tasks backend |

Credentials are set via environment variables (`TODOAT_GOOGLE_ACCESS_TOKEN`, `TODOAT_GOOGLE_REFRESH_TOKEN`, `TODOAT_GOOGLE_CLIENT_ID`, `TODOAT_GOOGLE_CLIENT_SECRET`).

### Microsoft To Do

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `backends.mstodo.enabled` | bool | `false` | Enable Microsoft To Do backend |

Credentials are set via environment variables (`TODOAT_MSTODO_ACCESS_TOKEN`, `TODOAT_MSTODO_REFRESH_TOKEN`, `TODOAT_MSTODO_CLIENT_ID`, `TODOAT_MSTODO_CLIENT_SECRET`).

### Git

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `backends.git.enabled` | bool | `false` | Enable Git backend |
| `backends.git.file` | string | `"TODO.md"` | Primary task file |
| `backends.git.auto_commit` | bool | `false` | Auto-commit changes |

### File

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `backends.file.enabled` | bool | `false` | Enable File backend |
| `backends.file.path` | string | | Path to task file |

See [Backend Setup](../how-to/backends.md) for detailed setup instructions.

## Analytics Configuration

Configure local command usage analytics:

```yaml
analytics:
  enabled: true           # Enable analytics tracking
  retention_days: 365     # Auto-cleanup after this many days (0 = forever)
```

Analytics data is stored locally at `~/.config/todoat/analytics.db` and is never transmitted. See [Analytics](../explanation/analytics.md) for details.

### Environment Variable Override

Override the config file setting with the `TODOAT_ANALYTICS_ENABLED` environment variable:

```bash
# Disable analytics regardless of config file
export TODOAT_ANALYTICS_ENABLED=false
```

### View Analytics Data

```bash
# View command usage statistics
todoat analytics stats

# View stats from past week
todoat analytics stats --since 7d

# View backend performance
todoat analytics backends

# View most common errors
todoat analytics errors
```

## Notification Commands

todoat provides notification commands for managing the notification log and testing the notification system. Notification behavior is controlled by the reminder configuration (see [Reminder Configuration](#reminder-configuration) above).

### Test Notifications

```bash
# Send a test notification
todoat notification test
```

### View Notification Log

```bash
# View notification history
todoat notification log

# View notification log as JSON
todoat --json notification log

# Clear notification log
todoat notification log clear
```

The notification log is stored at `~/.local/share/todoat/notifications.log`.

## Reminder Configuration

Configure task due date reminders. Reminders are disabled by default and require explicit configuration:

```yaml
reminder:
  enabled: true           # Must be enabled explicitly
  intervals:
    - 1d              # 1 day before due
    - 1h              # 1 hour before due
    - at due time     # When task is due
  os_notification: true   # Enable OS desktop notifications
  log_notification: true  # Enable notification log
```

### Reminder Options

| Option | Description | Default |
|--------|-------------|---------|
| `enabled` | Enable reminder system | `false` |
| `intervals` | Time before due to send reminders | `[]` (none) |
| `os_notification` | Send via OS desktop notifications | `false` |
| `log_notification` | Log to notification log file | `false` |

### Interval Format

| Format | Meaning |
|--------|---------|
| `15m` | 15 minutes before |
| `1h` | 1 hour before |
| `1d` | 1 day before |
| `7d` or `1w` | 1 week before |
| `at due time` | When the task is due |

### View Reminder Status

```bash
# Check reminder configuration and status
todoat reminder status

# List upcoming reminders
todoat reminder list

# Check for due reminders
todoat reminder check
```

See [Reminders How-To](../how-to/reminders.md) for detailed usage.

## Daemon Configuration

Configure the background sync daemon:

```yaml
sync:
  daemon:
    enabled: false        # Enable background sync daemon
    interval: 300         # Sync interval in seconds (default: 5 minutes)
    idle_timeout: 300     # Idle timeout in seconds before daemon exits (default: 5 minutes)
```

### Daemon Options

| Option | Description | Default |
|--------|-------------|---------|
| `enabled` | Enable daemon process for background sync | `false` |
| `interval` | Sync interval in seconds | `300` (5 minutes) |
| `idle_timeout` | Seconds before idle daemon exits | `300` (5 minutes) |

When `interval` or `idle_timeout` are set to 0 or left unset, the effective default of 300 seconds is used.

### Managing the Daemon

```bash
# Start daemon
todoat sync daemon start

# Start with custom interval
todoat sync daemon start --interval 60

# Check daemon status
todoat sync daemon status

# Stop daemon
todoat sync daemon stop

# Force kill if hung
todoat sync daemon kill
```

See [Synchronization](../how-to/sync.md#background-sync-daemon) for detailed usage.

## Trash Configuration

Configure automatic cleanup of deleted items:

```yaml
trash:
  retention_days: 30    # Keep deleted items for 30 days (0 = forever)
```

## Examples

### Switch Default Backend

```bash
# Switch to Todoist
todoat config set default_backend todoist

# Switch to SQLite for offline use
todoat config set default_backend sqlite
```

### Enable Offline Sync

```bash
todoat config set sync.enabled true
todoat config set sync.offline_mode auto
todoat config set sync.local_backend sqlite
```

### Script-Friendly Setup

```bash
# Enable for CI/CD pipelines
todoat config set no_prompt true
todoat config set output_format json
```

### Check Current Settings

```bash
# Check sync status
todoat config get sync

# Check all backends
todoat config get backends
```

## See Also

- [Getting Started](../tutorials/getting-started.md) - Initial setup
- [Backends](../explanation/backends.md) - Configure backends
- [Synchronization](../how-to/sync.md) - Sync settings
