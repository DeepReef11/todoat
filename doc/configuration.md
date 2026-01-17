# Configuration

todoat is designed to work with minimal configuration. This page documents the current configuration options and defaults.

## Data Storage

### Default Location

todoat stores its data in your home directory:

```
~/.todoat/
└── todoat.db    # SQLite database
```

The directory and database file are created automatically on first use.

## Command-Line Options

todoat is primarily configured through command-line flags rather than configuration files.

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

# Get JSON output for parsing
todoat --json MyList
```

## Environment

todoat uses your system's home directory to locate its data directory. The home directory is determined using the standard Go `os.UserHomeDir()` function.

## Future Configuration

Future versions may support:
- Configuration files (e.g., `~/.todoat/config.yaml`)
- Custom database locations
- Default list settings
- Backend selection

---
*Last updated: 2026-01-17*
