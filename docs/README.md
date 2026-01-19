# todoat Documentation

todoat is a powerful command-line task manager that works with multiple backends including Nextcloud, Todoist, SQLite, and Git-based storage.

## Quick Start

```bash
# Show tasks from a list
todoat MyList

# Add a new task
todoat MyList add "Complete the report"

# Complete a task
todoat MyList complete "report"

# Interactive list selection (no list specified)
todoat
```

## Documentation

| Guide | Description |
|-------|-------------|
| [Getting Started](getting-started.md) | Installation, configuration, and first steps |
| [Task Management](task-management.md) | Adding, updating, completing, and deleting tasks |
| [List Management](list-management.md) | Creating and managing task lists |
| [Backends](backends.md) | Configuring Nextcloud, Todoist, SQLite, and other backends |
| [Configuration](configuration.md) | Managing settings with the config command |
| [Views](views.md) | Customizing how tasks are displayed |
| [Synchronization](sync.md) | Offline mode and sync configuration |
| [Shell Completion](shell-completion.md) | Tab completion for Bash, Zsh, Fish, and PowerShell |
| [Tags](tags.md) | Managing task categories and tags |

## Command Overview

### Task Operations

| Command | Description |
|---------|-------------|
| `todoat <list>` | Show tasks (default action) |
| `todoat <list> add "task"` | Add a new task |
| `todoat <list> update "task" [flags]` | Update task properties |
| `todoat <list> complete "task"` | Mark task as done |
| `todoat <list> delete "task"` | Delete a task |

### Action Abbreviations

| Abbreviation | Full Command |
|--------------|--------------|
| `g` | `get` |
| `a` | `add` |
| `u` | `update` |
| `c` | `complete` |
| `d` | `delete` |

### Common Flags

| Flag | Description |
|------|-------------|
| `-s, --status` | Filter or set task status (TODO, DONE, PROCESSING, CANCELLED) |
| `-p, --priority` | Set task priority (0-9, 1=highest) |
| `-d, --description` | Set task description |
| `--due-date` | Set due date (YYYY-MM-DD) |
| `-v, --view` | Use a custom view |
| `--json` | Output in JSON format |
| `-y, --no-prompt` | Non-interactive mode for scripting |

### Other Commands

| Command | Description |
|---------|-------------|
| `todoat list` | Manage task lists |
| `todoat sync` | Synchronize with remote backend |
| `todoat tags` | List all tags in use |
| `todoat view` | Manage custom views |
| `todoat config` | View and modify configuration |
| `todoat credentials` | Manage backend credentials |
| `todoat version` | Show version information |

## Configuration

todoat follows XDG Base Directory specification. Configuration is stored in:

- **Linux/macOS**: `~/.config/todoat/config.yaml`
- **Windows**: `%APPDATA%\todoat\config.yaml`

See [Getting Started](getting-started.md) for configuration details.

## Getting Help

```bash
# General help
todoat --help

# Command-specific help
todoat add --help
todoat list --help
```

## Feedback

Report issues at: https://github.com/anthropics/claude-code/issues
