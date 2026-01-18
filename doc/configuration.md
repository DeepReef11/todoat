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
```

### Configuration Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `backends.sqlite.enabled` | boolean | `true` | Enable SQLite backend |
| `backends.sqlite.path` | string | `~/.local/share/todoat/tasks.db` | Path to SQLite database |
| `default_backend` | string | `sqlite` | Which backend to use |
| `no_prompt` | boolean | `false` | Disable interactive prompts |
| `output_format` | string | `text` | Default output format (`text` or `json`) |

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

| Variable | Description |
|----------|-------------|
| `XDG_CONFIG_HOME` | Configuration directory (default: `~/.config`) |
| `XDG_DATA_HOME` | Data directory (default: `~/.local/share`) |
| `XDG_CACHE_HOME` | Cache directory (default: `~/.cache`) |

## Validation

The configuration file is validated on load. Invalid configurations will produce an error message. Validation checks:

- `output_format` must be `text` or `json`
- `default_backend` must be a known backend (`sqlite`)
- The default backend must be enabled

---
*Last updated: 2026-01-18*
