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
| `list update [name] --name [new-name]` | Rename a list |
| `list info [name]` | Show list details (name, ID, task count) |
| `list delete [name]` | Soft-delete a list (moves to trash) |
| `list trash` | View lists in trash |
| `list trash restore [name]` | Restore a list from trash |
| `list trash purge [name]` | Permanently delete a list from trash |
| `list export [name]` | Export a list to a file |
| `list import [file]` | Import a list from a file |
| `list stats [name]` | Show database statistics |
| `list vacuum` | Compact the database (SQLite only) |

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

### Renaming Lists

Rename a list using the `list update` command:

```bash
# Rename a list
todoat list update "OldName" --name "NewName"
```

#### Rename Flags

| Flag | Description |
|------|-------------|
| `--name` | New name for the list (required) |

#### Rename Example

```bash
$ todoat list update "Work" --name "Work Projects"
Renamed list: Work → Work Projects
```

### Export and Import

Export and import lists to/from files in various formats.

```bash
# Export a list to JSON (default format)
todoat list export "Work"

# Export to a specific format
todoat list export "Work" --format json
todoat list export "Work" --format csv
todoat list export "Work" --format ical
todoat list export "Work" --format sqlite

# Export to a specific file
todoat list export "Work" --output ~/backup/work-tasks.json

# Import a list from a file (format auto-detected from extension)
todoat list import ./work-tasks.json
todoat list import ./tasks.csv
todoat list import ./calendar.ics

# Import with explicit format
todoat list import ./data.txt --format csv
```

#### Export Flags

| Flag | Description |
|------|-------------|
| `--format` | Export format: sqlite, json, csv, ical (default: json) |
| `--output` | Output file path (default: ./<list-name>.<ext>) |

#### Import Flags

| Flag | Description |
|------|-------------|
| `--format` | Import format (auto-detected from extension if not specified) |

#### Supported Formats

| Format | Extension | Description |
|--------|-----------|-------------|
| `json` | .json | JSON array of tasks |
| `csv` | .csv | Comma-separated values |
| `ical` | .ics, .ical | iCalendar format (VTODO) |
| `sqlite` | .db, .sqlite, .sqlite3 | SQLite database |

### Database Maintenance

Commands for managing the SQLite database.

```bash
# View database statistics
todoat list stats

# View statistics for a specific list
todoat list stats "Work"

# Compact the database (reclaim space from deleted data)
todoat list vacuum
```

#### Stats Output

```bash
$ todoat list stats
Database Statistics
==================
Total tasks: 42

Tasks per list:
  Work                 15
  Personal             27

Tasks by status:
  TODO                 25
  IN-PROGRESS          8
  DONE                 9

Database size: 128 KB
```

#### Vacuum Output

```bash
$ todoat list vacuum
Vacuum completed
Size before: 256 KB
Size after:  128 KB
Reclaimed:   128 KB
```

**Note:** The stats and vacuum commands are only supported for the SQLite backend.

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

# Filter by due date
todoat MyList get --due-before 2026-02-01      # Tasks due before Feb 1
todoat MyList get --due-after 2026-01-15       # Tasks due on or after Jan 15
todoat MyList get --due-after 2026-01-15 --due-before 2026-02-01  # Tasks in date range

# Filter by creation date
todoat MyList get --created-after 2026-01-01   # Tasks created on or after Jan 1
todoat MyList get --created-before 2026-01-15  # Tasks created before Jan 15

# Combine filters
todoat MyList get -s TODO --due-before 2026-02-01  # TODO tasks due before Feb 1
todoat MyList get -p high --due-after 2026-01-15   # High priority tasks due after Jan 15

# JSON output
todoat MyList get --json
```

### Get Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--status` | `-s` | Filter tasks by status (TODO, IN-PROGRESS, DONE, CANCELLED) |
| `--priority` | `-p` | Filter by priority: single value (1), comma-separated (1,2,3), or alias (high, medium, low) |
| `--tag` | | Filter by tag (can be specified multiple times or comma-separated; OR logic) |
| `--due-before` | | Filter tasks due before date (see [Date Formats](#date-formats)) |
| `--due-after` | | Filter tasks due on or after date (see [Date Formats](#date-formats)) |
| `--created-before` | | Filter tasks created before date (see [Date Formats](#date-formats)) |
| `--created-after` | | Filter tasks created on or after date (see [Date Formats](#date-formats)) |

### Date Formats

Date flags (`--due-date`, `--start-date`, `--due-before`, `--due-after`, `--created-before`, `--created-after`) accept both absolute and relative formats:

**Absolute format:**
- `YYYY-MM-DD` (e.g., `2026-01-31`)

**Relative formats:**

| Format | Description | Example |
|--------|-------------|---------|
| `today` | Current date | `--due-date today` |
| `tomorrow` | Next day | `--due-date tomorrow` |
| `yesterday` | Previous day | `--due-after yesterday` |
| `+Nd` | N days from today | `--due-date +7d` (in 7 days) |
| `-Nd` | N days ago | `--created-after -30d` (last 30 days) |
| `+Nw` | N weeks from today | `--due-date +2w` (in 2 weeks) |
| `+Nm` | N months from today | `--due-date +1m` (in 1 month) |

**Examples:**
```bash
# Due tomorrow
todoat Work add "Follow up" --due-date tomorrow

# Due in one week
todoat Work add "Weekly review" --due-date +7d

# Tasks created in the last 7 days
todoat Work get --created-after -7d

# Tasks due within the next 2 weeks
todoat Work get --due-before +2w
```

### Date Filter Notes

- Date filters use inclusive range (boundary dates are included)
- Tasks without dates are excluded from date filters
- Multiple filters combine with AND logic

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
| `--description` | `-d` | Task description/notes (multi-line text supported) |
| `--due-date` | | Due date (see [Date Formats](#date-formats)) |
| `--start-date` | | Start date (see [Date Formats](#date-formats)) |
| `--tag` | | Tag/category (can be specified multiple times or comma-separated) |
| `--parent` | `-P` | Parent task summary (creates subtask under specified parent) |
| `--literal` | `-l` | Treat task summary literally (don't parse `/` as hierarchy separator) |

### Add Examples with Dates, Tags, and Descriptions

```bash
# Add task with due date (absolute)
todoat Work add "Submit report" --due-date 2026-01-31

# Add task with relative due date
todoat Work add "Follow up" --due-date tomorrow
todoat Work add "Weekly review" --due-date +7d
todoat Work add "Next month check-in" --due-date +1m

# Add task with start and due date
todoat Work add "Project milestone" --start-date 2026-01-20 --due-date 2026-02-15

# Add task with priority and due date
todoat Work add "Urgent deadline" -p 1 --due-date 2026-01-25

# Add task with tags
todoat Work add "Review PR" --tag code-review
todoat Work add "Urgent fix" --tag urgent,bug
todoat Work add "Feature work" --tag feature --tag frontend

# Add task with description
todoat Work add "Research options" -d "Compare pricing, features, and integration complexity"

# Add task with description and other flags
todoat Work add "Write proposal" -p 1 -d "Include budget estimates and timeline" --due-date 2026-02-01
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
| `--description` | `-d` | Task description/notes (use "" to clear) |
| `--due-date` | | Due date (see [Date Formats](#date-formats), use "" to clear) |
| `--start-date` | | Start date (see [Date Formats](#date-formats), use "" to clear) |
| `--tag` | | Set tags (replaces existing; can be multiple or comma-separated) |
| `--parent` | `-P` | Set parent task (move task under specified parent) |
| `--no-parent` | | Remove parent relationship (make task root-level) |
| `--uid` | | Select task by unique identifier (for scripting) |
| `--local-id` | | Select task by SQLite internal ID (requires sync enabled) |

### Update Date, Tag, and Description Examples

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

# Set a description
todoat Work update "task" -d "Updated notes with more details"

# Clear a description
todoat Work update "task" -d ""
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

# Complete by UID (for scripting)
todoat MyList complete --uid "550e8400-e29b-41d4-a716-446655440000"

# Complete by local ID (requires sync enabled)
todoat MyList complete --local-id 42

# Complete all children of a task (bulk operation)
todoat MyList complete "Project/*"
```

### Complete Flags

| Flag | Description |
|------|-------------|
| `--uid` | Select task by unique identifier |
| `--local-id` | Select task by SQLite internal ID (requires sync enabled) |

## Deleting Tasks

Remove a task from a list:

```bash
# Delete by exact name
todoat MyList delete "Buy groceries"

# Delete by partial match
todoat MyList delete "groceries"

# Using alias
todoat MyList d "groceries"

# Delete by UID (for scripting)
todoat MyList delete --uid "550e8400-e29b-41d4-a716-446655440000"

# Delete by local ID (requires sync enabled)
todoat MyList delete --local-id 42

# Delete all children of a task (bulk operation)
todoat MyList delete "Project/*"
```

### Delete Flags

| Flag | Description |
|------|-------------|
| `--uid` | Select task by unique identifier |
| `--local-id` | Select task by SQLite internal ID (requires sync enabled) |

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

### Direct Task Selection

For scripting and automation, you can select tasks directly by their unique identifiers instead of summary matching:

```bash
# Select by UID (backend-assigned unique identifier)
todoat MyList update --uid "550e8400-e29b-41d4-a716-446655440000" -s DONE
todoat MyList complete --uid "550e8400-e29b-41d4-a716-446655440000"
todoat MyList delete --uid "550e8400-e29b-41d4-a716-446655440000"

# Select by local ID (SQLite internal ID, requires sync enabled)
todoat MyList update --local-id 42 -s DONE
todoat MyList complete --local-id 42
todoat MyList delete --local-id 42
```

#### Task Selection Flags

| Flag | Description |
|------|-------------|
| `--uid` | Select task by its unique identifier (UID) |
| `--local-id` | Select task by its SQLite internal ID (requires sync enabled) |

**Notes:**
- `--uid` works with tasks that have been synced and have a backend-assigned UID
- `--local-id` requires sync to be enabled and only works with the SQLite backend
- These flags bypass summary-based search entirely for unambiguous selection
- Use `--json` output to retrieve task UIDs and local IDs for scripting

#### Scripting Workflow

```bash
# 1. Get tasks with their IDs in JSON format
todoat MyList get --json | jq '.tasks[] | {local_id, uid, summary}'

# 2. Use specific ID to operate on exact task
todoat MyList complete --uid "550e8400-e29b-41d4-a716-446655440000"
```

## Bulk Hierarchy Operations

Operate on multiple tasks at once using wildcard patterns. This is useful for completing, updating, or deleting entire task hierarchies.

### Pattern Syntax

| Pattern | Description |
|---------|-------------|
| `Parent/*` | Matches only direct children of Parent |
| `Parent/**` | Matches all descendants of Parent (children, grandchildren, etc.) |

### Examples

```bash
# Complete all direct children of a parent task
todoat Work complete "Project/*"

# Complete all descendants (recursive) of a parent task
todoat Work complete "Project/**"

# Update priority for all descendants
todoat Work update "Release v2.0/**" -p 1

# Delete all direct children only
todoat Work delete "Old Project/*"

# Delete all descendants recursively
todoat Work delete "Archived/**"
```

### Output

**Text mode:**
```
Completed 5 tasks under "Release v2.0"
```

**JSON mode:**
```json
{
  "result": "ACTION_COMPLETED",
  "action": "complete",
  "affected_count": 5,
  "parent": "Release v2.0",
  "pattern": "**"
}
```

### Notes

- Parent task is resolved first by summary matching
- If parent has no children, returns INFO_ONLY with zero affected tasks
- Delete operations prompt for confirmation (use `-y` to skip)
- All matched tasks are processed in a single transaction

## Global Flags

These flags work with any command:

| Flag | Short | Description |
|------|-------|-------------|
| `--help` | `-h` | Show help for the command |
| `--version` | | Show version information |
| `--no-prompt` | `-y` | Disable interactive prompts (for scripting) |
| `--verbose` | `-V` | Enable debug/verbose output |
| `--detect-backend` | | Show auto-detected backends and exit |

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

## Views

Views allow you to customize how tasks are displayed. You can use built-in views or create custom views with specific fields, filters, and sorting.

### Using Views

```bash
# Use a built-in view
todoat MyList get --view default
todoat MyList get -v all

# Use a custom view
todoat MyList get --view my-custom-view
```

### Built-in Views

| View | Description |
|------|-------------|
| `default` | Standard task display showing status, summary, and priority |
| `all` | Comprehensive display showing all task metadata |

### View Management Commands

```bash
# List all available views (built-in and custom)
todoat view list
```

### View Flag

| Flag | Short | Description |
|------|-------|-------------|
| `--view` | `-v` | View to use for displaying tasks |

### Creating Custom Views

Custom views are YAML files stored in `~/.config/todoat/views/`. See [views.md](./views.md) for details on creating custom views.

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
      "description": "Get milk, bread, and eggs",
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
    "description": "Detailed notes about this task",
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


## Configuration Management

View and modify todoat configuration from the command line without manually editing YAML files.

```bash
# Show all configuration
todoat config

# Show specific configuration value
todoat config get default_backend

# Show nested configuration value
todoat config get sync.enabled

# Set a configuration value
todoat config set no_prompt true

# Set a nested configuration value
todoat config set sync.offline_mode auto

# Show config file path
todoat config path

# Open config file in editor
todoat config edit

# Reset to default configuration
todoat config reset
```

### Config Subcommands

| Command | Description |
|---------|-------------|
| `config` | Show all configuration (alias for `config get`) |
| `config get [key]` | Display config value(s), supports dot notation for nested keys |
| `config set <key> <value>` | Update a config value with validation |
| `config path` | Show the path to the active config file |
| `config edit` | Open config file in system editor ($EDITOR or vi) |
| `config reset` | Reset configuration to defaults (requires confirmation) |

### Supported Configuration Keys

| Key | Type | Valid Values |
|-----|------|--------------|
| `default_backend` | string | `sqlite` |
| `default_view` | string | Any view name |
| `no_prompt` | boolean | `true`, `false`, `yes`, `no`, `1`, `0` |
| `output_format` | string | `text`, `json` |
| `auto_detect_backend` | boolean | `true`, `false`, `yes`, `no`, `1`, `0` |
| `backends.sqlite.enabled` | boolean | `true`, `false`, `yes`, `no`, `1`, `0` |
| `backends.sqlite.path` | string | File path (supports `~` expansion) |
| `sync.enabled` | boolean | `true`, `false`, `yes`, `no`, `1`, `0` |
| `sync.local_backend` | string | Backend name |
| `sync.conflict_resolution` | string | `local`, `remote`, `manual` |
| `sync.offline_mode` | string | `auto`, `online`, `offline` |
| `sync.connectivity_timeout` | string | Duration (e.g., `5s`, `30s`) |
| `trash.retention_days` | integer | Non-negative integer (0 disables auto-purge) |

### Config Get Examples

```bash
# Show all configuration as YAML
todoat config get

# Show all configuration as JSON
todoat config get --json

# Get a specific value
todoat config get default_backend
# Output: sqlite

# Get a nested value
todoat config get sync.offline_mode
# Output: auto

# Get value as JSON
todoat config get sync.enabled --json
# Output: {"key": "sync.enabled", "value": true}
```

### Config Set Examples

```bash
# Enable non-interactive mode
todoat config set no_prompt true

# Change output format to JSON
todoat config set output_format json

# Set sync offline mode
todoat config set sync.offline_mode offline

# Set trash retention period
todoat config set trash.retention_days 7

# Change database path
todoat config set backends.sqlite.path ~/.local/share/todoat/my-tasks.db
```

### Config Edit

Opens the configuration file in your preferred editor:

```bash
todoat config edit
```

The editor is selected in this order:
1. `$EDITOR` environment variable
2. `$VISUAL` environment variable
3. `vi` (fallback)

### Config Reset

Reset all configuration to default values:

```bash
# Interactive confirmation
todoat config reset
# Output: This will reset your configuration to defaults. Continue? [y/N]

# Skip confirmation
todoat -y config reset
# Output: Configuration reset to defaults.
```

## Credential Management

Manage credentials for backend services (Nextcloud, Todoist, etc.) securely using system keyrings.

```bash
# Store credentials in system keyring
todoat credentials set nextcloud myuser --prompt

# Retrieve credentials and show source
todoat credentials get nextcloud myuser

# Remove credentials from keyring
todoat credentials delete nextcloud myuser

# List all backends with credential status
todoat credentials list
```

### Credentials Subcommands

| Command | Description |
|---------|-------------|
| `credentials set [backend] [username] --prompt` | Store credentials in system keyring |
| `credentials get [backend] [username]` | Retrieve credentials and show source |
| `credentials delete [backend] [username]` | Remove credentials from system keyring |
| `credentials list` | List all backends with credential status |

### Credential Sources

Credentials are retrieved in priority order:

1. **System keyring** - OS-native secure storage (macOS Keychain, Windows Credential Manager, Linux Secret Service)
2. **Environment variables** - `TODOAT_[BACKEND]_PASSWORD` and `TODOAT_[BACKEND]_USERNAME`

### Set Command

```bash
# Store credentials (password prompted securely)
todoat credentials set nextcloud myuser --prompt
```

The `--prompt` flag is required for security - passwords are never passed on the command line.

### Get Command

```bash
# Check credential status
todoat credentials get nextcloud myuser

# Output example when found:
# Source: keyring
# Username: myuser
# Password: ******** (hidden)
# Backend: nextcloud
# Status: Available

# JSON output
todoat credentials get nextcloud myuser --json
```

### Delete Command

```bash
# Remove stored credentials
todoat credentials delete nextcloud myuser
```

This only removes credentials from the system keyring. Environment variable credentials are not affected.

### List Command

```bash
# View all backends and their credential status
todoat credentials list

# Output example:
# Backend Credentials:
#
# BACKEND              USERNAME             STATUS          SOURCE
# nextcloud            myuser               Available       keyring
# todoist                                   Not configured  -

# JSON output
todoat credentials list --json
```

## Synchronization

Synchronize local cache with remote backends using the `sync` subcommand.

```bash
# Run sync (synchronize with remote backends)
todoat sync

# View sync status
todoat sync status

# View sync status with verbose output
todoat sync status --verbose

# View pending sync operations
todoat sync queue

# Clear all pending sync operations
todoat sync queue clear
```

### Sync Subcommands

| Command | Description |
|---------|-------------|
| `sync` | Synchronize with remote backends |
| `sync status` | Show last sync time, pending operations, and connection status |
| `sync status --verbose` | Show detailed sync metadata |
| `sync queue` | View pending sync operations |
| `sync queue clear` | Remove all pending operations from the queue |
| `sync conflicts` | View unresolved sync conflicts |
| `sync conflicts resolve [uid]` | Resolve a specific conflict with a strategy |
| `sync daemon start` | Start the background sync daemon |
| `sync daemon stop` | Stop the running sync daemon |
| `sync daemon status` | Show daemon status |

### Sync Status Output

```bash
$ todoat sync status
Sync Status:

Backend: sqlite
  Last Sync: 2026-01-18 14:30:00
  Pending Operations: 3
  Status: Offline (no remote backend configured)
```

### Sync Queue Output

```bash
$ todoat sync queue
Pending Operations: 3

ID     TYPE       TASK                           RETRIES  CREATED
1      create     Buy groceries                  0        14:30:15
2      update     Finish report                  1        14:32:00
3      delete     Old task                       0        14:35:22
```

### Sync Queue Clear

Use with caution - this discards unsynced changes:

```bash
$ todoat sync queue clear
Sync queue cleared: 3 operations removed
```

### Sync Conflicts

View and manage synchronization conflicts that occur when local and remote changes are incompatible.

```bash
# View all conflicts
todoat sync conflicts

# View conflicts in JSON format
todoat sync conflicts --json

# Resolve a specific conflict
todoat sync conflicts resolve [task-uid] --strategy server_wins
```

### Sync Conflicts Subcommands

| Command | Description |
|---------|-------------|
| `sync conflicts` | List all unresolved sync conflicts |
| `sync conflicts resolve [task-uid]` | Resolve a specific conflict using a strategy |

### Resolution Strategies

| Strategy | Description |
|----------|-------------|
| `server_wins` | Remote/server version overwrites local (default) |
| `local_wins` | Local version overwrites remote |
| `merge` | Attempt to merge both versions |
| `keep_both` | Keep both versions as separate tasks |

### Sync Conflicts Output

```bash
$ todoat sync conflicts
Conflicts: 1

UID                                  Task                           Detected             Status
abc123-def456                        Update report                  2026-01-18 14:30:00  pending
```

### Resolve Example

```bash
$ todoat sync conflicts resolve abc123-def456 --strategy local_wins
Conflict resolved for task abc123-def456 using strategy local_wins
```

### Sync Daemon

The sync daemon runs in the background and periodically synchronizes tasks with remote backends.

```bash
# Start the sync daemon
todoat sync daemon start

# Start with custom interval (in seconds)
todoat sync daemon start --interval 60

# Check daemon status
todoat sync daemon status

# Stop the daemon
todoat sync daemon stop
```

### Sync Daemon Subcommands

| Command | Description |
|---------|-------------|
| `sync daemon start` | Start the background sync daemon |
| `sync daemon start --interval N` | Start with sync interval of N seconds |
| `sync daemon stop` | Stop the running sync daemon |
| `sync daemon status` | Show whether daemon is running and its configuration |

### Daemon Status Output

```bash
$ todoat sync daemon status
Sync Daemon Status:
  Running: Yes
  PID: 12345
  Interval: 300s
  Last Sync: 2026-01-18 14:30:00
```

### Daemon Configuration

The sync daemon reads its configuration from the config file (`~/.config/todoat/config.yaml`):

```yaml
sync:
  enabled: true
  daemon:
    interval: 300  # Sync interval in seconds (default: 300 = 5 minutes)
```

The daemon automatically:
- Synchronizes pending local changes to remote backends
- Fetches remote changes and applies them locally
- Sends notifications for sync events and conflicts

## Reminders

Manage task reminders that notify you when tasks are approaching their due dates.

```bash
# Show reminder configuration status
todoat reminder status

# Check for due reminders and send notifications
todoat reminder check

# List upcoming reminders
todoat reminder list

# Disable reminders for a specific task
todoat reminder disable "Task name"

# Dismiss current reminder for a task (will trigger at next interval)
todoat reminder dismiss "Task name"
```

### Reminder Subcommands

| Command | Description |
|---------|-------------|
| `reminder status` | Show current reminder configuration and status |
| `reminder check` | Check all tasks and send reminders for those within configured intervals |
| `reminder list` | List all tasks with upcoming reminders |
| `reminder disable <task>` | Permanently disable reminders for a specific task |
| `reminder dismiss <task>` | Dismiss current reminder (will trigger again at next interval) |

### Reminder Configuration

Reminders are configured in the config file (`~/.config/todoat/config.yaml`):

```yaml
notifications:
  reminder:
    enabled: true
    intervals:
      - "1 day"
      - "1 hour"
      - "at due time"
    os_notification: true
    log_notification: true
```

### Interval Formats

| Format | Description |
|--------|-------------|
| `N day` / `N days` | N days before due date |
| `N hour` / `N hours` | N hours before due date |
| `N week` / `N weeks` | N weeks before due date |
| `at due time` | On the due date itself |

### Examples

```bash
# Check reminder status
$ todoat reminder status
Reminder Status:
  Status: enabled
  Intervals:
    - 1 day
    - 1 hour
    - at due time
  OS Notification: true
  Log Notification: true

# Run a reminder check
$ todoat reminder check
Triggered 2 reminder(s):
  - Submit report (due: 2026-01-19)
  - Review PR (due: 2026-01-18)

# List upcoming reminders
$ todoat reminder list
Upcoming reminders (3):
  - Submit report (due: 2026-01-19)
  - Review PR (due: 2026-01-18)
  - Team meeting (due: 2026-01-20)

# Disable reminders for a task permanently
$ todoat reminder disable "Review PR"
Disabled reminders for task: Review PR

# Dismiss current reminder (snooze until next interval)
$ todoat reminder dismiss "Submit report"
Dismissed reminders for task: Submit report
```

## Notifications

Manage the notification system for background sync events.

```bash
# Send a test notification
todoat notification test

# View notification history
todoat notification log

# Clear notification log
todoat notification log clear
```

### Notification Subcommands

| Command | Description |
|---------|-------------|
| `notification test` | Send a test notification through all enabled channels |
| `notification log` | View notification history from the log file |
| `notification log clear` | Clear all entries from the notification log |

### Test Command

Send a test notification to verify your notification setup is working:

```bash
$ todoat notification test
Test notification sent
```

This sends a notification through all enabled channels (OS notifications and log file).

### Log Commands

View and manage the notification log:

```bash
# View all logged notifications
$ todoat notification log
Notification Log:

[2026-01-18 14:30:00] sync_complete: Sync completed successfully
[2026-01-18 14:35:00] conflict: Conflict detected for task "Report"

# Clear the notification log
$ todoat notification log clear
Notification log cleared
```

### Notification Configuration

Notifications are configured in the config file (`~/.config/todoat/config.yaml`):

```yaml
notifications:
  enabled: true
  os:
    enabled: true
    on_sync_complete: true
    on_sync_error: true
    on_conflict: true
  log:
    enabled: true
    path: ~/.local/share/todoat/notifications.log
    max_size_mb: 10
    retention_days: 30
```

### Notification Types

| Type | Description |
|------|-------------|
| `sync_complete` | Synchronization completed successfully |
| `sync_error` | An error occurred during synchronization |
| `conflict` | A sync conflict was detected |
| `reminder` | Task reminder notification |
| `test` | Test notification |

## Migration

Migrate tasks between different backends using the `migrate` command.

```bash
# Migrate all tasks from SQLite to file backend
todoat migrate --from sqlite --to file

# Migrate from Todoist to SQLite
todoat migrate --from todoist --to sqlite

# Migrate only a specific list
todoat migrate --from sqlite --to file --list "Work"

# Preview what would be migrated (dry run)
todoat migrate --from sqlite --to file --dry-run
```

### Migrate Subcommands

| Command | Description |
|---------|-------------|
| `migrate --from [backend] --to [backend]` | Migrate all tasks between backends |
| `migrate --list [name]` | Migrate only the specified list |
| `migrate --dry-run` | Preview migration without making changes |

### Migrate Flags

| Flag | Description |
|------|-------------|
| `--from` | Source backend (sqlite, nextcloud, todoist, file) |
| `--to` | Target backend (sqlite, nextcloud, todoist, file) |
| `--list` | Migrate only the specified list (optional) |
| `--dry-run` | Show what would be migrated without making changes |

### Supported Backends for Migration

| Backend | Description |
|---------|-------------|
| `sqlite` | Local SQLite database (default) |
| `nextcloud` | Nextcloud CalDAV server |
| `todoist` | Todoist REST API |
| `file` | Plain text markdown file |

Note: Google Tasks, Microsoft To-Do, and Git backends are available for task management but do not yet support migration.

### Migration Examples

**SQLite to File (for portability):**
```bash
todoat migrate --from sqlite --to file
```

**Todoist to SQLite (for local backup):**
```bash
todoat migrate --from todoist --to sqlite
```

**Preview migration:**
```bash
$ todoat migrate --from sqlite --to file --dry-run
Would migrate 15 tasks from sqlite to file (dry-run)
```

### Notes

- Migration copies tasks; it does not delete from the source
- Existing tasks in the target are not affected
- Some metadata may not transfer between backends with different capabilities
- Use `--dry-run` first to preview what will be migrated

## Terminal User Interface (TUI)

Launch an interactive terminal interface for managing tasks.

```bash
todoat tui
```

The TUI provides a two-pane interface with lists on the left and tasks on the right. Navigate using keyboard shortcuts.

### Key Bindings

| Key | Action |
|-----|--------|
| `Tab` | Switch between lists/tasks panes |
| `↑`/`k` | Move up |
| `↓`/`j` | Move down |
| `a` | Add new task |
| `e` | Edit task |
| `c` | Toggle completion |
| `d` | Delete task |
| `/` | Filter tasks |
| `?` | Show help |
| `q` | Quit |

For full TUI documentation, see [tui.md](./tui.md).

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

## Version

Display version and build information.

```bash
# Show version information
todoat version

# Show extended build information
todoat version -v

# Get version as JSON
todoat --json version
```

### Version Output

**Standard output:**
```
Version: 1.0.0
Commit:  abc1234
Built:   2026-01-19T12:00:00Z
```

**Verbose output (`-v`):**
```
Version: 1.0.0
Commit:  abc1234
Built:   2026-01-19T12:00:00Z
Go Version: go1.23.0
Platform:   linux/amd64
```

**JSON output:**
```json
{
  "version": "1.0.0",
  "commit": "abc1234",
  "build_date": "2026-01-19T12:00:00Z",
  "go_version": "go1.23.0",
  "platform": "linux/amd64"
}
```

### Version Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--verbose` | `-v` | Show extended build information (Go version, platform) |

**Note:** `todoat --version` is also available as a shorthand for `todoat version`.

---
*Last updated: 2026-01-19*
