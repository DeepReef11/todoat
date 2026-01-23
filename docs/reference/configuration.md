# Configuration

This guide covers managing todoat configuration using the `config` command.

## Configuration Location

todoat follows the XDG Base Directory specification:

- **Linux/macOS**: `~/.config/todoat/config.yaml`
- **Windows**: `%APPDATA%\todoat\config.yaml`

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
| `sync.offline_mode` | string | `auto`, `always`, `never` |
| `sync.conflict_resolution` | string | `server_wins`, `local_wins`, `merge`, `keep_both` |
| `sync.connectivity_timeout` | string | Network timeout for connectivity checks (default: `5s`) |
| `trash.retention_days` | int | Days to keep deleted items (0 = forever) |

## Notification Configuration

Configure desktop and log notifications for sync events:

```yaml
notification:
  enabled: true
  os_notification:
    enabled: true
    on_sync_error: true     # Notify on sync failures
    on_conflict: true       # Notify on sync conflicts
  log_notification:
    enabled: true
    path: "~/.local/share/todoat/notifications.log"
```

### Test Notifications

```bash
# Send a test notification
todoat notification test
```

### View Notification Log

```bash
# View notification history
todoat notification log

# Clear notification log
todoat notification log clear
```

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
