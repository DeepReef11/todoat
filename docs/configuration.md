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
todoat config set backends.sqlite.db_path "~/my-tasks/tasks.db"
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
| `no_prompt` | bool | Non-interactive mode |
| `json_output` | bool | JSON output by default |
| `sync.enabled` | bool | Enable synchronization |
| `sync.offline_mode` | string | `auto`, `always`, `never` |
| `sync.conflict_resolution` | string | `server_wins`, `local_wins`, `merge`, `keep_both` |

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
todoat config set json_output true
```

### Check Current Settings

```bash
# Check sync status
todoat config get sync

# Check all backends
todoat config get backends
```

## See Also

- [Getting Started](getting-started.md) - Initial setup
- [Backends](backends.md) - Configure backends
- [Synchronization](sync.md) - Sync settings
