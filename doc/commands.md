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

## Shell Completion

Generate shell completion scripts for faster command entry.

```bash
# Generate completion for your shell
todoat completion bash > ~/.bash_completion.d/todoat
todoat completion zsh > ~/.zfunc/_todoat
todoat completion fish > ~/.config/fish/completions/todoat.fish
todoat completion powershell > todoat.ps1
```

### Completion Subcommands

| Command | Description |
|---------|-------------|
| `completion bash` | Generate bash completion script |
| `completion zsh` | Generate zsh completion script |
| `completion fish` | Generate fish completion script |
| `completion powershell` | Generate PowerShell completion script |

### Setup Examples

**Bash:**
```bash
# Add to ~/.bashrc
source <(todoat completion bash)

# Or save to a file
todoat completion bash > /etc/bash_completion.d/todoat
```

**Zsh:**
```bash
# Add to ~/.zshrc (before compinit)
source <(todoat completion zsh)

# Or save to fpath
todoat completion zsh > "${fpath[1]}/_todoat"
```

**Fish:**
```bash
todoat completion fish | source

# Or save permanently
todoat completion fish > ~/.config/fish/completions/todoat.fish
```

**PowerShell:**
```powershell
todoat completion powershell | Out-String | Invoke-Expression

# Or add to profile
todoat completion powershell >> $PROFILE
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
