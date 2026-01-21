# README Planner

This document outlines how to build the main README.md for todoat. The README should be concise, focused on getting users started quickly, and showcase key features through practical examples.

---

## README Structure

### 1. Header Section

```markdown
# todoat

Manage your tasks seamlessly from the comfort of your terminal

[![License: BSD-2](https://img.shields.io/badge/License-BSD--2--Clause-darkred)](https://opensource.org/license/bsd-2-clause)
```

**Guidelines:**
- One-line description (what it is)
- Badges: license, build status (optional)
- No lengthy feature lists in header

---

### 2. Quick Start Section

**Purpose:** Get users from zero to working in under 2 minutes.

```markdown
## Quick Start

### Install

```bash
go install github.com/user/todoat/cmd/todoat@latest
```

### First Run

```bash
# Creates default config at ~/.config/todoat/config.yaml
todoat

# Add your first task
todoat MyList add "Hello world"

# View tasks
todoat MyList
```
```

**Guidelines:**
- Single install command
- 3-4 commands maximum to show basic usage
- No explanation of flags yet

---

### 3. Configuration Section

**Purpose:** Show minimal config for each backend type.

```markdown
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
  auto_sync: true
  local_backend: sqlite
  conflict_resolution: server_wins

default_backend: nextcloud
```

```

**Guidelines:**
- Show 3 configs: [local-only, md file], [remote-only, todoist], [sync mode, multi-backend, auto_detect_backend with git, nextcloud, sqlite]
- Minimal YAML (only required fields)
- Brief comment and commands for setting keyring credentials
- Link to full config docs for advanced options. Put more advanced doc in ./doc/

---

### 4. Examples Section

**Purpose:** Showcase most important features through real commands.

```markdown
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

# Select task by UID (after ambiguous match)
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
```

**Guidelines:**
- Group by feature category
- 2-5 commands per category
- Show most common use cases
- Include comments explaining what each does
- Prioritize features in this order:
  1. Basic CRUD (add, complete, update, delete)
  2. Subtasks
  3. Filtering
  4. Sync
  5. Scripting/automation
  6. List management

---

### 5. Documentation Links Section

```markdown
## Documentation

- [CLI Reference](./reshape/doc/CLI_INTERFACE.md) - All commands and flags
- [Configuration Guide](./reshape/doc/CONFIGURATION.md) - Full config options
- [Sync Guide](./SYNC_GUIDE.md) - Offline sync setup
- [Backend System](./reshape/doc/BACKEND_SYSTEM.md) - Backend configuration

## Status Values

| Status | Abbreviation | Meaning |
|--------|--------------|---------|
| TODO | T | Not started |
| IN-PROGRESS | I | In progress |
| DONE | D | Completed |
| CANCELLED | C | Abandoned |
```

**Guidelines:**
- Link to detailed docs, don't duplicate content
- Include quick reference table for status (frequently needed)

---

**Guidelines:**
- Keep minimal
- No contributing section in README (link to CONTRIBUTING.md if exists)

---

## What NOT to Include in README

1. **Full feature lists** - Link to FEATURES_OVERVIEW.md instead
2. **All flags and options** - Link to CLI_INTERFACE.md
3. **Architecture details** - Link to detailed docs
4. **Credential management details** - Brief mention, link to docs
5. **View system** - Advanced feature, link to docs
6. **Backend-specific details** - Link to BACKEND_SYSTEM.md
7. **Troubleshooting** - Separate doc or wiki

---

## Length Guidelines

| Section | Target Lines |
|---------|--------------|
| Header | 5-10 |
| Quick Start | 15-20 |
| Configuration | 40-50 |
| Examples | 60-80 |
| Documentation Links | 15-20 |
| **Total** | **~150-180 lines** |

---

## Tone Guidelines

- **Concise**: No fluff, every line serves a purpose
- **Practical**: Show, don't tell - examples over explanations
- **Scannable**: Users should find what they need in seconds
- **Assume competence**: Don't over-explain basic concepts

---

## Checklist Before Publishing

- [ ] Can a new user go from install to first task in < 2 minutes?
- [ ] Are all 3 config scenarios covered (local, remote, sync)?
- [ ] Do examples cover the 80% use case?
- [ ] Are links to detailed docs working?
- [ ] Is total length under 200 lines?
- [ ] No duplicate information from other docs?
