# Getting Started with todoat

This guide walks you through installing, configuring, and using todoat for the first time.

## Installation

### From Source

```bash
git clone https://github.com/DeepReef11/todoat
cd todoat
go build -o todoat ./cmd/todoat
```

### Adding to PATH

```bash
# Add to your shell profile (.bashrc, .zshrc, etc.)
export PATH="$PATH:/path/to/todoat"
```

## First Run

When you first run todoat, if no configuration file exists, it automatically creates one at `~/.config/todoat/config.yaml` with a sample configuration including documentation and examples.

```bash
$ todoat
# Config file auto-created with SQLite backend enabled
# Shows available lists or prompts for list selection
```

The default configuration uses SQLite as the local backend, which requires no additional setup.

## Configuration

Configuration is stored at `~/.config/todoat/config.yaml`.

### Minimal SQLite Configuration

For local-only task management:

```yaml
backends:
  sqlite:
    type: sqlite
    enabled: true
    path: ""  # Uses default location

default_backend: sqlite
ui: cli
```

### Nextcloud Configuration

To sync with Nextcloud:

```yaml
backends:
  nextcloud:
    type: nextcloud
    enabled: true
    host: "nextcloud.example.com"
    username: "myuser"

default_backend: nextcloud
ui: cli
```

Store your password securely:

```bash
todoat credentials set nextcloud myuser --prompt
```

### Todoist Configuration

To use Todoist:

```yaml
backends:
  todoist:
    type: todoist
    enabled: true
    username: "token"

default_backend: todoist
ui: cli
```

Store your API token:

```bash
todoat credentials set todoist token --prompt
# Enter your Todoist API token when prompted
```

## Basic Usage

### View Your Task Lists

```bash
# Interactive list selection
todoat

# Or list all available lists
todoat list
```

### View Tasks in a List

```bash
todoat MyList
```

### Add a Task

```bash
# Simple task
todoat MyList add "Buy groceries"

# Task with priority (1=highest, 9=lowest)
todoat MyList add "Urgent report" -p 1

# Task with due date
todoat MyList add "Project deadline" --due-date 2026-01-31

# Task with description
todoat MyList add "Meeting notes" -d "Prepare agenda for Monday meeting"
```

### Update a Task

```bash
# Change status
todoat MyList update "groceries" -s DONE

# Change priority
todoat MyList update "report" -p 2

# Rename a task
todoat MyList update "old name" --summary "new name"
```

### Complete a Task

```bash
todoat MyList complete "groceries"
```

### Delete a Task

```bash
todoat MyList delete "old task"
```

## Action Abbreviations

Use shortcuts for common actions:

| Abbreviation | Full Command |
|--------------|--------------|
| `a` | `add` |
| `u` | `update` |
| `c` | `complete` |
| `d` | `delete` |
| `g` | `get` |

Example:

```bash
todoat MyList a "New task"     # add
todoat MyList c "Task name"    # complete
```

## Status Values

Tasks have four possible statuses:

| Status | Abbreviation | Meaning |
|--------|--------------|---------|
| TODO | T | Not started |
| IN-PROGRESS | I | In progress |
| DONE | D | Completed |
| CANCELLED | C | Abandoned |

Filter by status:

```bash
# Show only incomplete tasks
todoat MyList -s TODO,IN-PROGRESS

# Show only completed tasks
todoat MyList -s DONE
```

## Creating Task Lists

```bash
# Create a new list
todoat list create "Personal"

# Create with description and color
todoat list create "Work" --description "Work tasks" --color "#0066cc"
```

## Shell Completion

Enable tab completion for faster command entry.

### Quick Setup

The easiest way to set up completion:

```bash
todoat completion install
```

This automatically detects your shell from `$SHELL` and installs completion to a user-writable location.

### Manual Setup

If you prefer manual control:

```bash
# Zsh (add to .zshrc)
source <(todoat completion zsh)

# Bash (add to .bashrc)
source <(todoat completion bash)

# Fish (add to config.fish)
todoat completion fish | source
```

See [Shell Completion](../how-to/shell-completion.md) for more options and troubleshooting.

## Getting Help

```bash
# General help
todoat --help

# Command-specific help (for subcommands like list, config, etc.)
todoat list --help
todoat config --help

# Version information
todoat version
```

## Next Steps

- [Task Management](../how-to/task-management.md) - Learn about all task operations
- [List Management](../how-to/list-management.md) - Organize tasks into lists
- [Backends](../explanation/backends.md) - Configure Nextcloud, Todoist, and other backends
- [Views](../how-to/views.md) - Customize how tasks are displayed
- [Synchronization](../how-to/sync.md) - Work offline and sync across devices
