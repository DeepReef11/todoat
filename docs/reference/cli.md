# CLI Reference

Complete reference for all todoat commands, flags, and options.

## Synopsis

```bash
todoat [list] [action] [task] [flags]
todoat [command]
```

## Global Flags

These flags are available for all commands:

| Flag | Description |
|------|-------------|
| `-b, --backend <name>` | Backend to use (sqlite, todoist, nextcloud, google, mstodo, git, file) |
| `--detect-backend` | Show auto-detected backends and exit |
| `--json` | Output in JSON format |
| `-y, --no-prompt` | Disable interactive prompts |
| `-V, --verbose` | Enable verbose/debug output |
| `-h, --help` | Help for the command |

## Task Commands

### Task Actions

| Action | Abbreviation | Description |
|--------|--------------|-------------|
| `get` | `g` | Show tasks (default) |
| `add` | `a` | Add a new task |
| `update` | `u` | Update task properties |
| `complete` | `c` | Mark task as done |
| `delete` | `d` | Delete a task |

### Task Flags

#### For add/update operations:

| Flag | Type | Description |
|------|------|-------------|
| `-d, --description <text>` | string | Task description/notes (use "" to clear) |
| `--due-date <date>` | string | Due date (see [Date Syntax](#date-syntax) below, use "" to clear) |
| `--start-date <date>` | string | Start date (see [Date Syntax](#date-syntax) below, use "" to clear) |
| `-p, --priority <n>` | string | Priority (0-9, 1=highest) |
| `-s, --status <status>` | string | Status (TODO, IN-PROGRESS, DONE, CANCELLED) |
| `--tag <tag>` | strings | Tag/category (can specify multiple times) |
| `--tags <tags>` | strings | Alias for --tag |
| `--add-tag <tag>` | strings | Add tag to existing tags (for update) |
| `--remove-tag <tag>` | strings | Remove tag from existing tags (for update) |
| `-P, --parent <summary>` | string | Parent task summary (for subtasks) |
| `--no-parent` | bool | Remove parent relationship (make root-level) |
| `--summary <text>` | string | New task summary (for rename) |
| `-l, --literal` | bool | Treat "/" literally (don't create hierarchy) |
| `--recur <rule>` | string | Recurrence (daily, weekly, monthly, yearly, or "every N days/weeks/months") |
| `--recur-from-completion` | bool | Base recurrence on completion date |

#### For get/filter operations:

| Flag | Type | Description |
|------|------|-------------|
| `-s, --status <status>` | string | Filter by status (comma-separated) |
| `-p, --priority <filter>` | string | Filter by priority (see below) |
| `--tag <tag>` | strings | Filter by tag (can specify multiple) |
| `-v, --view <name>` | string | View to use (default, all, or custom) |
| `--due-after <date>` | string | Filter tasks due on or after date (see [Date Syntax](#date-syntax)) |
| `--due-before <date>` | string | Filter tasks due before date (see [Date Syntax](#date-syntax)) |
| `--created-after <date>` | string | Filter tasks created on or after date (see [Date Syntax](#date-syntax)) |
| `--created-before <date>` | string | Filter tasks created before date (see [Date Syntax](#date-syntax)) |

#### Pagination:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit <n>` | int | | Maximum number of tasks to show |
| `--offset <n>` | int | 0 | Number of tasks to skip |
| `--page <n>` | int | | Page number (1-indexed, alternative to offset) |
| `--page-size <n>` | int | 50 | Number of tasks per page |

**Pagination examples:**
```bash
# Show first 20 tasks
todoat MyList --limit 20

# Show second page (tasks 51-100 with default page size)
todoat MyList --page 2

# Custom page size with page number
todoat MyList --page 2 --page-size 25

# Manual offset and limit
todoat MyList --offset 50 --limit 25
```

##### Priority Filter Syntax

The priority filter supports multiple formats:

| Format | Description | Example |
|--------|-------------|---------|
| Single value | Filter by specific priority | `-p 1` |
| Range | Filter by priority range | `-p 1-3` |
| Comma-separated | Filter by multiple priorities | `-p 1,2,3` |
| Named levels | Filter by priority category | `-p high` |
| Undefined | Filter tasks without priority | `-p 0` |

**Named priority levels:**

| Name | Priority Range | Description |
|------|----------------|-------------|
| `high` | 1-4 | High priority tasks |
| `medium` | 5 | Medium priority tasks |
| `low` | 6-9 | Low priority tasks |

Priority filters can be combined with other filters like status:

```bash
# Show high priority TODO tasks
todoat MyList -s TODO -p high

# Show medium priority in-progress tasks
todoat MyList -s IN-PROGRESS -p 5

# Show tasks with priority 1, 2, or 3 using range syntax
todoat MyList -p 1-3
```

##### Date Syntax

Date flags accept multiple formats:

| Format | Example | Description |
|--------|---------|-------------|
| ISO date | `2026-01-23` | Absolute date |
| ISO datetime | `2026-01-23T14:30` | Absolute date and time |
| Relative keyword | `today`, `tomorrow` | Human-friendly relative dates |
| Relative offset | `+1d`, `+2w`, `+1m` | Days, weeks, or months from today |

**Relative date keywords:**

| Keyword | Meaning |
|---------|---------|
| `today` | Current date |
| `tomorrow` | Next day |
| `yesterday` | Previous day |

**Relative date offsets:**

| Suffix | Meaning | Example |
|--------|---------|---------|
| `d` | Days | `+3d` (3 days from today) |
| `w` | Weeks | `+2w` (2 weeks from today) |
| `m` | Months | `+1m` (1 month from today) |

**Relative dates with time:**

Combine a relative date with a time by separating with a space:

```bash
# Tomorrow at 9am
todoat MyList add "Meeting" --due-date "tomorrow 09:00"

# One week from now at 2:30pm
todoat MyList add "Review" --due-date "+7d 14:30"
```

#### Direct task selection:

| Flag | Type | Description |
|------|------|-------------|
| `--uid <uid>` | string | Select task by backend UID (bypasses summary search) |
| `--local-id <id>` | int | Select task by local SQLite ID (requires sync enabled) |

### Examples

```bash
# Show tasks from a list
todoat MyList

# Add a task with priority and due date
todoat MyList add "Finish report" -p 1 --due-date 2026-02-01

# Add a subtask
todoat MyList add "Review section 1" -P "Finish report"

# Update task status
todoat MyList update "report" -s IN-PROGRESS

# Add tags to existing task
todoat MyList update "report" --add-tag work --add-tag urgent

# Complete a task
todoat MyList complete "report"

# Filter tasks by status and priority
todoat MyList -s TODO,IN-PROGRESS -p high

# Filter by due date
todoat MyList --due-before 2026-01-31

# Use a custom view
todoat MyList -v urgent

# Pagination - show first 20 tasks
todoat MyList --limit 20

# Pagination - show page 2
todoat MyList --page 2
```

## list

Manage task lists.

### Synopsis

```bash
todoat list [flags]
todoat list [command]
```

### Subcommands

| Command | Description |
|---------|-------------|
| `list` | View all lists (default) |
| `create` | Create a new list |
| `delete` | Delete a list (move to trash) |
| `update` | Update list properties |
| `info` | Show list details |
| `export` | Export a list to a file |
| `import` | Import a list from a file |
| `trash` | View and manage deleted lists |
| `stats` | Show database statistics |
| `vacuum` | Compact the database |

### list create

Create a new task list.

```bash
todoat list create [name] [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--description` | string | Description for the list |
| `--color` | string | Hex color (e.g., #FF5733, ABC) |

### list update

Update a list's name, color, or description.

```bash
todoat list update [name] [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--name` | string | New name for the list |
| `--description` | string | Description for the list |
| `--color` | string | Hex color (e.g., #FF5733, ABC) |

### list export

Export a list to a file.

```bash
todoat list export [name] [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | `json` | Export format: sqlite, json, csv, ical |
| `--output` | string | `./<list-name>.<ext>` | Output file path |

### list import

Import a list from a file.

```bash
todoat list import [file] [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--format` | string | Import format (auto-detect from extension if not specified) |

### list trash

View and manage deleted lists.

```bash
todoat list trash [command]
```

| Subcommand | Description |
|------------|-------------|
| `restore` | Restore a list from trash |
| `purge` | Permanently delete a list from trash |

### Examples

```bash
# List all lists
todoat list

# Create a new list
todoat list create "Personal"

# Create with description and color
todoat list create "Work" --description "Work tasks" --color "#0066cc"

# Update list name and color
todoat list update "Work" --name "Work Tasks" --color "#00cc66"

# Delete a list (moves to trash)
todoat list delete "Old List"

# View trash
todoat list trash

# Restore from trash
todoat list trash restore "Old List"

# Export list to JSON
todoat list export MyList --output tasks.json

# Export to iCal format
todoat list export MyList --format ical --output tasks.ics

# Import from file
todoat list import tasks.json
```

## analytics

View usage analytics and statistics.

### Synopsis

```bash
todoat analytics [command]
```

### Subcommands

| Command | Description |
|---------|-------------|
| `stats` | Show command usage statistics |
| `backends` | Show backend performance metrics |
| `errors` | Show most common errors |

### analytics stats

Display summary of command usage including counts and success rates.

```bash
todoat analytics stats [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--since` | string | Filter events from the past duration (e.g., 7d, 30d, 1y) |

### analytics backends

Display performance metrics for each backend.

```bash
todoat analytics backends [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--since` | string | Filter events from the past duration (e.g., 7d, 30d, 1y) |

### analytics errors

Display the most common errors grouped by command and error type.

```bash
todoat analytics errors [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--since` | string | | Filter events from the past duration (e.g., 7d, 30d, 1y) |
| `--limit` | int | 10 | Maximum number of errors to show |

### Examples

```bash
# Show command usage statistics
todoat analytics stats

# Show stats from past week
todoat analytics stats --since 7d

# Show backend performance
todoat analytics backends

# Show backend performance from past month
todoat analytics backends --since 30d

# Show top 10 errors (default)
todoat analytics errors

# Show top 20 errors from past year
todoat analytics errors --since 1y --limit 20

# Output in JSON format
todoat analytics stats --json
todoat analytics backends --json
todoat analytics errors --json
```

## config

View and manage configuration.

### Synopsis

```bash
todoat config [command]
```

### Subcommands

| Command | Description |
|---------|-------------|
| `get [key]` | Display configuration value(s) |
| `set <key> <value>` | Update configuration value |
| `edit` | Open config file in editor |
| `path` | Show config file location |
| `reset` | Reset to default configuration |

### Examples

```bash
# Show all configuration
todoat config get

# Get specific value
todoat config get default_backend

# Set default backend
todoat config set default_backend sqlite

# Open in editor
todoat config edit

# Show config file path
todoat config path
```

## sync

Synchronize with remote backends.

### Synopsis

```bash
todoat sync [flags]
todoat sync [command]
```

### Subcommands

| Command | Description |
|---------|-------------|
| `sync` | Synchronize now (default) |
| `status` | Show sync status |
| `queue` | View and manage pending sync operations |
| `conflicts` | View and manage sync conflicts |
| `daemon` | Manage the sync daemon |

### sync queue

View and manage the sync queue.

```bash
todoat sync queue [command]
```

| Subcommand | Description |
|------------|-------------|
| (default) | View pending operations |
| `clear` | Clear all pending operations |

### sync conflicts

View and manage sync conflicts.

```bash
todoat sync conflicts [command]
```

| Subcommand | Description |
|------------|-------------|
| `resolve` | Resolve a specific conflict |

#### sync conflicts resolve

```bash
todoat sync conflicts resolve [task-uid] [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--strategy` | string | `server_wins` | Resolution strategy: server_wins, local_wins, merge, keep_both |

### sync daemon

Manage the background sync daemon.

```bash
todoat sync daemon [command]
```

| Subcommand | Description |
|------------|-------------|
| `start` | Start the sync daemon |
| `status` | Show daemon status |
| `stop` | Stop the sync daemon |

#### sync daemon start

```bash
todoat sync daemon start [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--interval` | int | 300 | Sync interval in seconds |

### Examples

```bash
# Sync now
todoat sync

# Show sync status
todoat sync status

# View pending operations
todoat sync queue

# Clear all pending operations
todoat sync queue clear

# View conflicts
todoat sync conflicts

# Resolve a conflict with local changes
todoat sync conflicts resolve abc123 --strategy local_wins

# Start background sync
todoat sync daemon start

# Check daemon status
todoat sync daemon status

# Stop background sync
todoat sync daemon stop
```

## view

Manage custom views.

### Synopsis

```bash
todoat view [command]
```

### Subcommands

| Command | Description |
|---------|-------------|
| `list` | List available views |
| `create` | Create a new view |

### view create

Create a new view interactively or from flags.

```bash
todoat view create <name> [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--fields` | string | Comma-separated list of fields (e.g., "status,summary,priority") |
| `--sort` | string | Sort rule in format "field:direction" (e.g., "priority:asc") |
| `--filter-status` | string | Filter by status (comma-separated, e.g., "TODO,IN-PROGRESS") |
| `--filter-priority` | string | Filter by priority (e.g., "high", "1-3", "low") |

Without `-y` flag, opens an interactive builder. With `-y`, uses provided flags or defaults.

### Examples

```bash
# List views
todoat view list

# Create a view interactively
todoat view create urgent

# Create a view with fields and sort (non-interactive)
todoat view create urgent -y --fields "status,summary,priority" --sort "priority:asc"

# Create a view with status filter
todoat view create active -y --filter-status "TODO,IN-PROGRESS"

# Create a view with priority filter
todoat view create high-priority -y --filter-priority "high"

# Create a view with combined filters
todoat view create urgent-tasks -y --filter-status "TODO" --filter-priority "1-3" --sort "priority:asc"
```

## credentials

Manage backend credentials.

### Synopsis

```bash
todoat credentials [command]
```

### Subcommands

| Command | Description |
|---------|-------------|
| `list` | List all backends with credential status |
| `get <backend> <username>` | Retrieve credentials and show source |
| `set <backend> <username>` | Store credentials in system keyring |
| `update <backend> <username>` | Update existing credentials |
| `delete <backend> <username>` | Remove credentials from keyring |

### credentials set

Store credentials securely in the system keyring.

```bash
todoat credentials set [backend] [username] [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--prompt` | bool | Prompt for password input (required for security) |

### credentials update

Update the password for existing credentials in the system keyring.

```bash
todoat credentials update [backend] [username] [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--prompt` | bool | Prompt for password input (required for security) |
| `--verify` | bool | Verify the updated credential against the backend |

### Examples

```bash
# List credential status
todoat credentials list

# Set credentials (prompts for password)
todoat credentials set nextcloud myuser --prompt

# Set Todoist API token
todoat credentials set todoist token --prompt

# Update existing credentials
todoat credentials update nextcloud myuser --prompt

# Update and verify against backend
todoat credentials update nextcloud myuser --prompt --verify

# Delete credentials
todoat credentials delete nextcloud myuser
```

## migrate

Migrate tasks between backends.

### Synopsis

```bash
todoat migrate [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--from <backend>` | Source backend (sqlite, nextcloud, todoist, file) |
| `--to <backend>` | Target backend (sqlite, nextcloud, todoist, file) |
| `--list <name>` | Migrate only specified list |
| `--dry-run` | Show what would be migrated without making changes |
| `--target-info <backend>` | Show tasks in target backend |

### Supported Migrations

| From/To | sqlite | nextcloud | todoist | file |
|---------|--------|-----------|---------|------|
| sqlite | N/A | ✓ | ✓ | ✓ |
| nextcloud | ✓ | N/A | ✓ | ✓ |
| todoist | ✓ | ✓ | N/A | ✓ |
| file | ✓ | ✓ | ✓ | N/A |

### What Gets Migrated

- Task summary and description
- Priority and status
- Due dates and start dates
- Tags/categories
- Parent-child relationships (task hierarchy)
- Recurrence rules

### Migration Notes

- UIDs are preserved where possible
- Status values are mapped between backends (e.g., IN-PROGRESS may become different values for backends that don't support it)
- Large lists are migrated in batches with progress indicators
- Use `--dry-run` first to verify the migration plan

### Examples

```bash
# Migrate from SQLite to Nextcloud
todoat migrate --from sqlite --to nextcloud

# Migrate specific list
todoat migrate --from sqlite --to nextcloud --list "Work"

# Preview migration (dry run)
todoat migrate --from sqlite --to nextcloud --dry-run

# Check target backend contents
todoat migrate --target-info nextcloud
```

## reminder

Manage task reminders.

### Synopsis

```bash
todoat reminder [command]
```

### Subcommands

| Command | Description |
|---------|-------------|
| `list` | List upcoming reminders |
| `check` | Check for due reminders |
| `dismiss <task>` | Dismiss current reminder |
| `disable <task>` | Disable reminders for a task |
| `status` | Show reminder configuration status |

### Examples

```bash
# List upcoming reminders
todoat reminder list

# Check for due reminders
todoat reminder check

# Dismiss a reminder
todoat reminder dismiss "Meeting prep"
```

## notification

Manage the notification system.

### Synopsis

```bash
todoat notification [command]
```

### Subcommands

| Command | Description |
|---------|-------------|
| `test` | Send a test notification |
| `log` | View notification log |

### notification log

View notification history.

```bash
todoat notification log [command]
```

| Subcommand | Description |
|------------|-------------|
| `clear` | Clear notification log |

### Examples

```bash
# Test notifications
todoat notification test

# View notification log
todoat notification log

# Clear notification log
todoat notification log clear
```

## tags

List all tags in use.

### Synopsis

```bash
todoat tags [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `-l, --list <name>` | Filter tags to a specific list |

### Examples

```bash
# List all tags
todoat tags

# List tags in specific list
todoat tags -l MyList
```

## tui

Launch the terminal user interface.

### Synopsis

```bash
todoat tui [flags]
```

### Description

Launches an interactive terminal interface for managing tasks with keyboard navigation.

## completion

Generate shell completion scripts.

### Synopsis

```bash
todoat completion [shell]
```

### Subcommands

| Command | Description |
|---------|-------------|
| `bash` | Generate bash completion script |
| `zsh` | Generate zsh completion script |
| `fish` | Generate fish completion script |
| `powershell` | Generate PowerShell completion script |

### Examples

```bash
# Zsh (add to .zshrc)
source <(todoat completion zsh)

# Bash (add to .bashrc)
source <(todoat completion bash)

# Fish
todoat completion fish | source

# PowerShell
todoat completion powershell | Out-String | Invoke-Expression
```

## version

Display version information.

### Synopsis

```bash
todoat version [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `-v, --verbose` | Show extended build information |

## Status Values

| Status | Abbreviation | Description |
|--------|--------------|-------------|
| `TODO` | `T` | Not started |
| `IN-PROGRESS` | `I` | Work in progress |
| `DONE` | `D` | Completed |
| `CANCELLED` | `C` | Abandoned |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error |

## See Also

- [Getting Started](../tutorials/getting-started.md) - Installation and first steps
- [Task Management](../how-to/task-management.md) - Task operations guide
- [Configuration](configuration.md) - Config file reference
