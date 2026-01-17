# Command Reference

Complete reference for all todoat commands and flags.

## Basic Syntax

```
todoat [list-name] [action] [task-summary] [flags]
```

- **list-name**: Name of the task list (created automatically if it doesn't exist)
- **action**: Operation to perform (defaults to `get` if omitted)
- **task-summary**: Text to identify or describe the task

## Actions

| Action | Alias | Description |
|--------|-------|-------------|
| `get` | `g` | List tasks in a list (default action) |
| `add` | `a` | Add a new task |
| `update` | `u` | Update an existing task |
| `complete` | `c` | Mark a task as completed |
| `delete` | `d` | Delete a task |

## Viewing Tasks

List all tasks in a list:

```bash
# Using full action name
todoat MyList get

# Using alias
todoat MyList g

# Default action (get is implied when only list name provided)
todoat MyList
```

**Output format:**
```
Tasks in 'MyList':
  [ ] Buy groceries
  [>] Write report [P1]
  [x] Call dentist
```

Status icons:
- `[ ]` - TODO (needs action)
- `[>]` - IN-PROGRESS
- `[x]` - COMPLETED
- `[-]` - CANCELLED

## Adding Tasks

Add a new task to a list:

```bash
# Basic task
todoat MyList add "Buy groceries"

# Using alias
todoat MyList a "Buy groceries"

# With priority (0-9, where 1 is highest)
todoat MyList add "Urgent task" -p 1
```

### Add Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--priority` | `-p` | Task priority (0-9, 0=undefined, 1=highest) |

## Updating Tasks

Update an existing task:

```bash
# Update task status
todoat MyList update "task name" -s DONE

# Update task priority
todoat MyList update "task name" -p 2

# Rename a task
todoat MyList update "old name" --summary "new name"

# Using alias
todoat MyList u "task name" -s IN-PROGRESS
```

### Update Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--priority` | `-p` | New priority (0-9) |
| `--status` | `-s` | New status (TODO, IN-PROGRESS, DONE, CANCELLED) |
| `--summary` | | New task summary/name |

### Status Values

| Status | Aliases |
|--------|---------|
| `TODO` | Default for new tasks |
| `IN-PROGRESS` | `INPROGRESS`, `PROGRESS` |
| `DONE` | `COMPLETED` |
| `CANCELLED` | `CANCELED` |

## Completing Tasks

Mark a task as completed:

```bash
# Complete by exact name
todoat MyList complete "Buy groceries"

# Complete by partial match
todoat MyList complete "groceries"

# Using alias
todoat MyList c "groceries"
```

## Deleting Tasks

Remove a task from a list:

```bash
# Delete by exact name
todoat MyList delete "Buy groceries"

# Delete by partial match
todoat MyList delete "groceries"

# Using alias
todoat MyList d "groceries"
```

## Task Matching

When specifying a task for `update`, `complete`, or `delete`, todoat uses intelligent matching:

1. **Exact match**: First tries case-insensitive exact match
2. **Partial match**: If no exact match, searches for tasks containing the search term

```bash
# These all match "Buy groceries for dinner"
todoat MyList complete "Buy groceries for dinner"  # Exact match
todoat MyList complete "groceries"                  # Partial match
todoat MyList complete "dinner"                     # Partial match
```

### Multiple Matches

If multiple tasks match your search term:
- In interactive mode: Error prompting you to be more specific
- In no-prompt mode (`-y`): Lists all matching tasks and exits

## Global Flags

These flags work with any command:

| Flag | Short | Description |
|------|-------|-------------|
| `--help` | `-h` | Show help for the command |
| `--version` | | Show version information |
| `--no-prompt` | `-y` | Disable interactive prompts (for scripting) |
| `--verbose` | `-V` | Enable debug/verbose output |
| `--json` | | Output results in JSON format |

### No-Prompt Mode

Use `-y` or `--no-prompt` for scripting:

```bash
# Delete without confirmation
todoat -y MyList delete "task"

# Script-friendly operation
todoat --no-prompt MyList add "Automated task"
```

### JSON Output

Use `--json` for machine-readable output:

```bash
todoat --json MyList
```

## Examples

```bash
# Show help
todoat --help

# Show version
todoat --version

# List tasks in "Work" list
todoat Work

# Add high-priority task
todoat Work add "Finish report" -p 1

# Update task to in-progress
todoat Work update "report" -s IN-PROGRESS

# Complete the task
todoat Work complete "report"

# Delete a task (no confirmation in script mode)
todoat -y Work delete "old task"
```

---
*Last updated: 2026-01-17*
