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
- **Priority Support**: Set task priority (0-9, where 1 is highest)
- **Status Tracking**: Track task status (TODO, IN-PROGRESS, DONE, CANCELLED)
- **SQLite Backend**: Local task storage in `~/.todoat/todoat.db`

## Basic Usage

```bash
# View tasks in a list
todoat MyList

# Add a task
todoat MyList add "Buy groceries"

# Add a task with priority
todoat MyList add "Urgent task" -p 1

# Complete a task
todoat MyList complete "Buy groceries"

# Delete a task
todoat MyList delete "Buy groceries"
```

## Documentation

- [Installation](./installation.md) - How to install todoat
- [Commands](./commands.md) - Complete command reference
- [Backends](./backends.md) - Backend setup and configuration
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
*Last updated: 2026-01-17*
