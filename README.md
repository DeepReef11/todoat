# todoat

Manage your tasks seamlessly from the comfort of your terminal

[![License: BSD-2](https://img.shields.io/badge/License-BSD--2--Clause-darkred)](https://opensource.org/license/bsd-2-clause)

## Quick Start

### Install

```bash
go install todoat/cmd/todoat@latest
```

### First Run

```bash
# Creates default config at ~/.config/todoat/config.yaml
todoat

# Add your first task
todoat Work add "Review PR #123"

# View tasks
todoat Work
```

## Configuration

Config location: `~/.config/todoat/config.yaml`

### SQLite (Local Only)

```yaml
backends:
  local:
    type: sqlite
    enabled: true

default_backend: local
```

### Nextcloud (CalDAV)

```yaml
backends:
  nextcloud:
    type: nextcloud
    enabled: true
    host: "nextcloud.example.com"
    username: "myuser"
    # Password: use keyring (recommended)
    # todoat credentials set nextcloud myuser --prompt

default_backend: nextcloud
```

### SQLite + Nextcloud (Offline Sync)

```yaml
backends:
  nextcloud:
    type: nextcloud
    enabled: true
    host: "nextcloud.example.com"
    username: "myuser"

sync:
  enabled: true
  local_backend: sqlite
  conflict_resolution: remote

default_backend: nextcloud
```

## Examples

### Basic Task Management

```bash
# Add task
todoat Work add "Review PR #123"

# Add task with priority and due date
todoat Work add "Ship feature" -p 1 --due-date 2026-01-20

# Complete task
todoat Work complete "Review PR"

# Update task status
todoat Work update "Ship feature" -s IN-PROGRESS

# Delete task
todoat Work delete "Old task"
```

### Subtasks

```bash
# Add subtask under parent
todoat Work add "Write tests" -P "Ship feature"

# Create hierarchy with path notation
todoat Work add "Project/Phase 1/Task A"
```

### Filtering

```bash
# Show only TODO tasks
todoat Work -s TODO

# Show TODO and IN-PROGRESS
todoat Work -s TODO,IN-PROGRESS

# Filter by priority (1 = highest)
todoat Work -p 1
```

### Sync Operations

```bash
# Manual sync with remote
todoat sync

# Check sync status
todoat sync status

# View pending operations
todoat sync queue
```

### Scripting (No-Prompt Mode)

```bash
# Non-interactive mode for scripts
todoat -y Work complete "task"

# JSON output for parsing
todoat -y --json Work

# Select task by UID
todoat Work update --uid "550e8400-e29b-41d4-a716-446655440000" -s DONE
```

### List Management

```bash
# Create new list
todoat list create "Projects"

# View all lists
todoat list

# Delete list (moves to trash)
todoat list delete "Old List"

# Restore from trash
todoat list trash restore "Old List"
```

## Documentation

- [CLI Reference](./docs/reference/cli.md) - All commands and flags
- [Configuration Guide](./docs/reference/configuration.md) - Full config options
- [Sync Guide](./docs/how-to/sync.md) - Offline sync setup
- [Backend System](./docs/explanation/backend-system.md) - Backend configuration
- [Feature Overview](./docs/explanation/features-overview.md) - All features

## Status Values

| Status | Abbreviation | Meaning |
|--------|--------------|---------|
| TODO | T | Not started |
| IN-PROGRESS | I | In progress |
| DONE | D | Completed |
| CANCELLED | C | Abandoned |
