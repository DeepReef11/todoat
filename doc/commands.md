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

## List Management

View and manage task lists using the `list` subcommand:

```bash
# View all lists with task counts
todoat list

# Create a new list
todoat list create "MyList"

# Get information about a list
todoat list info "MyList"

# Delete a list (moves to trash)
todoat list delete "MyList"

# View deleted lists in trash
todoat list trash

# Restore a list from trash
todoat list trash restore "MyList"

# Permanently delete a list from trash
todoat list trash purge "MyList"
```

### List Subcommands

| Command | Description |
|---------|-------------|
| `list` | View all lists with task counts |
| `list create [name]` | Create a new list |
| `list info [name]` | Show list details (name, ID, task count) |
| `list delete [name]` | Soft-delete a list (moves to trash) |
| `list trash` | View lists in trash |
| `list trash restore [name]` | Restore a list from trash |
| `list trash purge [name]` | Permanently delete a list from trash |

### Output Example

```
Available lists (2):

NAME                 TASKS
Work                 5
Personal             3
```

### Trash Example

```bash
$ todoat list delete "OldProject"
Deleted list: OldProject

$ todoat list trash
Deleted lists (1):

NAME                 DELETED
OldProject           2026-01-18 14:30

$ todoat list trash restore "OldProject"
Restored list: OldProject
```

## Viewing Tasks

List all tasks in a list:

```bash
# Using full action name
todoat MyList get

# Using alias
todoat MyList g

# Default action (get is implied when only list name provided)
todoat MyList

# Filter by status
todoat MyList get -s TODO
todoat MyList get -s IN-PROGRESS
todoat MyList get -s DONE

# Filter by priority
todoat MyList get -p 1           # Show only priority 1 tasks
todoat MyList get -p 1,2,3       # Show priorities 1, 2, or 3
todoat MyList get -p high        # Show high priority (1-4)
todoat MyList get -p medium      # Show medium priority (5)
todoat MyList get -p low         # Show low priority (6-9)

# Filter by tag
todoat MyList get --tag work              # Show tasks with "work" tag
todoat MyList get --tag work,urgent       # Show tasks with "work" OR "urgent" tag
todoat MyList get --tag work --tag urgent # Same as above (multiple flags)

# JSON output
todoat MyList get --json
```

### Get Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--status` | `-s` | Filter tasks by status (TODO, IN-PROGRESS, DONE, CANCELLED) |
| `--priority` | `-p` | Filter by priority: single value (1), comma-separated (1,2,3), or alias (high, medium, low) |
| `--tag` | | Filter by tag (can be specified multiple times or comma-separated; OR logic) |

### Priority Aliases

| Alias | Priority Values |
|-------|-----------------|
| `high` | 1, 2, 3, 4 |
| `medium` | 5 |
| `low` | 6, 7, 8, 9 |

**Output format:**
```
Tasks in 'MyList':
  [TODO] Buy groceries
  [IN-PROGRESS] Write report [P1] {work}
  [DONE] Call dentist {personal,health}
```

Tags are displayed in curly braces `{tag1,tag2}` when present.

Status indicators:
- `[TODO]` - Needs action
- `[IN-PROGRESS]` - Currently being worked on
- `[DONE]` - Completed
- `[CANCELLED]` - No longer needed

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
| `--due-date` | | Due date in YYYY-MM-DD format |
| `--start-date` | | Start date in YYYY-MM-DD format |
| `--tag` | | Tag/category (can be specified multiple times or comma-separated) |
| `--parent` | `-P` | Parent task summary (creates subtask under specified parent) |
| `--literal` | `-l` | Treat task summary literally (don't parse `/` as hierarchy separator) |

### Add Examples with Dates and Tags

```bash
# Add task with due date
todoat Work add "Submit report" --due-date 2026-01-31

# Add task with start and due date
todoat Work add "Project milestone" --start-date 2026-01-20 --due-date 2026-02-15

# Add task with priority and due date
todoat Work add "Urgent deadline" -p 1 --due-date 2026-01-25

# Add task with tags
todoat Work add "Review PR" --tag code-review
todoat Work add "Urgent fix" --tag urgent,bug
todoat Work add "Feature work" --tag feature --tag frontend
```

### Subtasks and Hierarchy

Create task hierarchies using the `/` separator or the `--parent` flag:

```bash
# Create a hierarchy using path notation (Project > Backend > API)
todoat Work add "Project/Backend/API"

# Creates:
#   Project
#     └─ Backend
#          └─ API

# Create a subtask using --parent flag
todoat Work add "Write tests" --parent "API"

# Use --literal to treat "/" literally (not as hierarchy separator)
todoat Work add "Fix bug in A/B test" --literal
```

**Path notation behavior:**
- Intermediate tasks (Project, Backend) are created if they don't exist
- If a task at any level already exists, it is reused
- Priority, dates, and tags are applied only to the leaf task (API in the example)

### Subtask Display

Subtasks are displayed in a tree structure:

```
Tasks in 'Work':
  [TODO] Project
  └─ [TODO] Backend
       └─ [IN-PROGRESS] API [P1]
       └─ [TODO] Write tests
  [DONE] Other task
```

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
| `--due-date` | | Due date (YYYY-MM-DD format, use "" to clear) |
| `--start-date` | | Start date (YYYY-MM-DD format, use "" to clear) |
| `--tag` | | Set tags (replaces existing; can be multiple or comma-separated) |
| `--parent` | `-P` | Set parent task (move task under specified parent) |
| `--no-parent` | | Remove parent relationship (make task root-level) |

### Update Date and Tag Examples

```bash
# Set a due date
todoat Work update "task" --due-date 2026-02-01

# Clear a due date
todoat Work update "task" --due-date ""

# Set both dates
todoat Work update "task" --start-date 2026-01-15 --due-date 2026-01-30

# Set tags (replaces all existing tags)
todoat Work update "task" --tag urgent
todoat Work update "task" --tag work,meeting

# Clear all tags
todoat Work update "task" --tag ""
```

### Update Parent Examples

```bash
# Move a task under a parent (make it a subtask)
todoat Work update "Write tests" --parent "API"

# Move a subtask to root level (remove parent)
todoat Work update "Write tests" --no-parent

# Move a task to a different parent
todoat Work update "Write tests" --parent "Backend"
```

**Note:** Circular references are automatically prevented. You cannot set a task's descendant as its parent.

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

### Cascade Deletion

When you delete a parent task, all its subtasks are also deleted:

```bash
# Given hierarchy: Project > Backend > API
todoat Work delete "Backend"
# Deletes both "Backend" and "API"
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

### No-Prompt Mode

Use `-y` or `--no-prompt` for scripting:

```bash
# Delete without confirmation
todoat -y MyList delete "task"

# Script-friendly operation
todoat --no-prompt MyList add "Automated task"
```

In no-prompt mode, todoat outputs result codes to help with scripting:

| Result Code | Meaning |
|-------------|---------|
| `ACTION_COMPLETED` | A modification was made (add, update, complete, delete) |
| `INFO_ONLY` | Read-only operation (get, list view) |
| `ERROR` | An error occurred |

## JSON Output

Use `--json` flag to get machine-readable JSON output:

```bash
# Get tasks as JSON
todoat MyList get --json

# List all lists as JSON
todoat list --json

# Add task and get JSON response
todoat MyList add "New task" --json
```

### JSON Response Examples

**Task list response:**
```json
{
  "tasks": [
    {
      "uid": "abc123",
      "summary": "Buy groceries",
      "status": "TODO",
      "priority": 1,
      "due_date": "2026-01-31",
      "tags": ["shopping", "errands"]
    }
  ],
  "list": "Shopping",
  "count": 1,
  "result": "INFO_ONLY"
}
```

**Action response:**
```json
{
  "action": "add",
  "task": {
    "uid": "abc123",
    "summary": "New task",
    "status": "TODO",
    "priority": 0
  },
  "result": "ACTION_COMPLETED"
}
```

**Error response:**
```json
{
  "error": "task summary is required",
  "code": 1,
  "result": "ERROR"
}
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
*Last updated: 2026-01-18*
