# todoat

A command-line task manager with multiple backend support.

## Quick Start

1. Install todoat (see [installation.md](./installation.md))
2. Add your first task:
   ```bash
   todoat MyList add "My first task"
   ```
3. View your tasks:
   ```bash
   todoat MyList
   ```
4. Complete a task:
   ```bash
   todoat MyList complete "My first task"
   ```

## Features

- **Task Management**: Add, view, update, complete, and delete tasks
- **Task Lists**: Organize tasks into named lists (created automatically)
- **Subtasks**: Create task hierarchies with parent-child relationships
- **Priority Support**: Set task priority (0-9, where 1 is highest) and filter by priority
- **Status Tracking**: Track task status (TODO, IN-PROGRESS, DONE, CANCELLED)
- **Tags**: Categorize tasks with tags and filter by tag
- **Due Dates**: Set start and due dates for tasks
- **Views**: Customizable task display with built-in and custom views
- **JSON Output**: Machine-readable JSON output for scripting
- **Multiple Backends**: SQLite (local), Nextcloud CalDAV, Todoist, Google Tasks, Microsoft To-Do, Git/Markdown, File
- **Backend Migration**: Migrate tasks between different backends
- **Synchronization**: Sync tasks with remote backends with conflict resolution
- **Secure Credentials**: System keyring integration for secure credential storage
- **Notifications**: Desktop and log notifications for sync events
- **Reminders**: Configurable task due date reminders

## Basic Usage

```bash
# View tasks in a list
todoat MyList

# Add a task
todoat MyList add "Buy groceries"

# Add a task with priority and due date
todoat MyList add "Urgent task" -p 1 --due-date 2026-01-31

# Add a task with tags
todoat MyList add "Code review" --tag work,urgent

# Filter by tag
todoat MyList get --tag urgent

# Use a custom view
todoat MyList get --view all

# Create subtasks using path notation
todoat MyList add "Project/Backend/API"

# Complete a task
todoat MyList complete "Buy groceries"

# Delete a task
todoat MyList delete "Buy groceries"

# View all lists
todoat list

# JSON output (for scripting)
todoat MyList get --json
```

## Documentation

- [Installation](./installation.md) - How to install todoat
- [Commands](./commands.md) - Complete command reference
- [Configuration](./configuration.md) - Configuration options
- [Views](./views.md) - Custom views for task display
- [Backends](./backends.md) - Backend setup and configuration
- [TUI](./tui.md) - Terminal user interface guide
- [Examples](./examples.md) - Usage examples and workflows

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--help` | `-h` | Show help information |
| `--version` | | Show version |
| `--no-prompt` | `-y` | Disable interactive prompts |
| `--verbose` | `-V` | Enable debug output |
| `--json` | | Output in JSON format |

## License

See the project repository for license information.

---
*Last updated: 2026-01-18*
