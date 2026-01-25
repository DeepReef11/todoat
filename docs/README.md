# todoat Documentation

todoat is a command-line task manager supporting multiple backends including Nextcloud, Todoist, SQLite, and Git-based storage.

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

## Try It Out

Run the feature demo script to see all todoat features in action:

```bash
# Run all demos
./docs/feature-demo.sh

# Run a specific section
./docs/feature-demo.sh tasks      # Task management demo
./docs/feature-demo.sh recurring  # Recurring tasks demo
./docs/feature-demo.sh views      # Views and filtering demo
```

Available sections: `version`, `help`, `config`, `lists`, `tasks`, `subtasks`, `dates`, `recurring`, `tags`, `priority`, `views`, `filters`, `json`, `sync`, `reminders`, `notifications`, `credentials`, `migration`, `export`, `scripting`, `tui`, `cleanup`, `all`

## Documentation

### Tutorials

Step-by-step guides for getting started:

| Guide | Description |
|-------|-------------|
| [Getting Started](tutorials/getting-started.md) | Installation, configuration, and first steps |

### How-To Guides

Task-oriented guides for accomplishing specific goals:

| Guide | Description |
|-------|-------------|
| [Task Management](how-to/task-management.md) | Adding, updating, completing, and deleting tasks |
| [List Management](how-to/list-management.md) | Creating and managing task lists |
| [Views](how-to/views.md) | Customizing how tasks are displayed |
| [Synchronization](how-to/sync.md) | Offline mode and sync configuration |
| [Reminders](how-to/reminders.md) | Task due date reminders |
| [Shell Completion](how-to/shell-completion.md) | Tab completion for Bash, Zsh, Fish, and PowerShell |
| [Tags](how-to/tags.md) | Managing task categories and tags |
| [TUI](how-to/tui.md) | Interactive terminal user interface |

### Reference

Technical reference documentation:

| Guide | Description |
|-------|-------------|
| [CLI Reference](reference/cli.md) | Complete command and flag reference |
| [Configuration](reference/configuration.md) | Managing settings with the config command |

### Explanation

Background and conceptual information:

| Guide | Description |
|-------|-------------|
| [Backends](explanation/backends.md) | Configuring Nextcloud, Todoist, SQLite, and other backends |

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
| `-s, --status` | Filter or set task status (TODO, IN-PROGRESS, DONE, CANCELLED) |
| `-p, --priority` | Set task priority (0-9, 1=highest) |
| `-d, --description` | Set task description |
| `--due-date` | Set due date (YYYY-MM-DD or with time: YYYY-MM-DDTHH:MM) |
| `--start-date` | Set start date (YYYY-MM-DD) |
| `--recur` | Set recurrence (daily, weekly, monthly, or "every N days") |
| `--recur-from-completion` | Base recurrence on completion date |
| `--tag` | Add or filter by tag (can specify multiple times) |
| `--add-tag` | Add tag to existing tags (for update) |
| `--remove-tag` | Remove tag from existing tags (for update) |
| `-P, --parent` | Parent task for subtask creation |
| `--no-parent` | Remove parent relationship (for update) |
| `--summary` | New task summary (for update/rename) |
| `-l, --literal` | Treat "/" literally (don't create hierarchy) |
| `-v, --view` | Use a custom view |
| `-b, --backend` | Select backend for this command |
| `--json` | Output in JSON format |
| `-y, --no-prompt` | Non-interactive mode for scripting |
| `-V, --verbose` | Enable verbose/debug output |

### Date Filtering Flags (for get)

| Flag | Description |
|------|-------------|
| `--due-after` | Filter tasks due on or after date (YYYY-MM-DD) |
| `--due-before` | Filter tasks due before date (YYYY-MM-DD) |
| `--created-after` | Filter tasks created on or after date |
| `--created-before` | Filter tasks created before date |

### Pagination Flags (for get)

| Flag | Description |
|------|-------------|
| `--limit` | Maximum number of tasks to show |
| `--offset` | Number of tasks to skip |
| `--page` | Page number (1-indexed) |
| `--page-size` | Tasks per page (default: 50) |

### Direct Task Selection Flags

| Flag | Description |
|------|-------------|
| `--uid` | Select task by backend UID (for synced tasks) |
| `--local-id` | Select task by local SQLite ID (requires sync enabled) |

### Other Commands

| Command | Description |
|---------|-------------|
| `todoat list` | Manage task lists |
| `todoat sync` | Synchronize with remote backend |
| `todoat tags` | List all tags in use |
| `todoat view` | Manage custom views |
| `todoat config` | View and modify configuration |
| `todoat credentials` | Manage backend credentials |
| `todoat analytics` | View usage statistics and backend performance |
| `todoat migrate` | Migrate tasks between backends |
| `todoat notification` | Manage notification system |
| `todoat reminder` | Manage task reminders |
| `todoat tui` | Launch terminal user interface |
| `todoat completion` | Generate shell completion scripts |
| `todoat version` | Show version information |
| `todoat help` | Help about any command |

## Configuration

todoat follows XDG Base Directory specification. Configuration is stored in:

- **Linux/macOS**: `~/.config/todoat/config.yaml`
- **Windows**: `%APPDATA%\todoat\config.yaml`

See [Getting Started](tutorials/getting-started.md) for configuration details.

## Getting Help

```bash
# General help
todoat --help

# Command-specific help (for subcommands like list, config, etc.)
todoat list --help
todoat config --help
todoat credentials --help
```

## Feedback

Report issues at: https://github.com/DeepReef11/todoat/issues
