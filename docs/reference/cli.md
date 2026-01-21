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
| `-b, --backend <name>` | Backend to use (sqlite, todoist, nextcloud) |
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
| `--due-date <date>` | string | Due date (YYYY-MM-DD or YYYY-MM-DDTHH:MM, use "" to clear) |
| `--start-date <date>` | string | Start date (YYYY-MM-DD, use "" to clear) |
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
| `-p, --priority <filter>` | string | Filter by priority (1,2,3 or high/medium/low) |
| `--tag <tag>` | strings | Filter by tag (can specify multiple) |
| `-v, --view <name>` | string | View to use (default, all, or custom) |
| `--due-after <date>` | string | Filter tasks due on or after date (YYYY-MM-DD) |
| `--due-before <date>` | string | Filter tasks due before date (YYYY-MM-DD) |
| `--created-after <date>` | string | Filter tasks created on or after date |
| `--created-before <date>` | string | Filter tasks created before date |

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
| `queue` | View pending sync operations |
| `conflicts` | View and manage sync conflicts |
| `daemon` | Manage the sync daemon |

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

### Examples

```bash
# Sync now
todoat sync

# Show sync status
todoat sync status

# View pending operations
todoat sync queue

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
| `get <backend>` | Retrieve credentials and show source |
| `set <backend> <username>` | Store credentials in system keyring |
| `update <backend> <username>` | Update existing credentials |
| `delete <backend>` | Remove credentials from keyring |

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
todoat credentials delete nextcloud
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

### Examples

```bash
# Migrate from SQLite to Nextcloud
todoat migrate --from sqlite --to nextcloud

# Migrate specific list
todoat migrate --from sqlite --to nextcloud --list "Work"

# Preview migration
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
