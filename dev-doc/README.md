# gosynctasks Complete Feature Inventory

**Version:** Based on current codebase (January 2026)

This documentation provides a comprehensive inventory of all features available in gosynctasks, a fast, flexible, multi-backend task synchronization tool written in Go.

## What is gosynctasks?

gosynctasks is a command-line task management tool that allows you to:

- Manage tasks seamlessly from your terminal
- Work with multiple storage backends (Nextcloud CalDAV, Todoist, Git/Markdown, SQLite)
- Sync tasks across devices with offline support
- Organize tasks hierarchically with subtasks
- Customize task views with flexible formatters
- Store credentials securely using OS keyrings

## Documentation Structure

This feature inventory is organized into logical categories. Each document provides detailed explanations of how features work, including user journeys, prerequisites, and technical details.

### Core Features

1. **[Task Management](./TASK_MANAGEMENT.md)** - Creating, reading, updating, and deleting tasks
   - Add tasks with metadata (priority, dates, descriptions)
   - Update task properties
   - Complete and delete tasks
   - Filter tasks by status, priority, tags, and dates
   - Interactive task selection for partial matches

2. **[List Management](./LIST_MANAGEMENT.md)** - Organizing tasks into lists
   - Create and configure task lists
   - Rename and delete lists
   - Trash management (soft delete, restore, permanent deletion)
   - List information and metadata
   - Color coding and descriptions

3. **[Subtasks & Hierarchy](./SUBTASKS_HIERARCHY.md)** - Nested task organization
   - Create subtasks under parent tasks
   - Path-based hierarchy creation (e.g., "parent/child/grandchild")
   - Tree visualization with box-drawing characters
   - Parent task references and navigation

### Backend & Sync

4. **[Backend System](./BACKEND_SYSTEM.md)** - Multi-backend architecture
   - Nextcloud CalDAV backend
   - Todoist REST API backend
   - Git/Markdown file backend
   - SQLite local database backend
   - Backend selection priority and auto-detection
   - Pluggable TaskManager interface

5. **[Synchronization](./SYNCHRONIZATION.md)** - Offline sync and conflict resolution
   - Bidirectional sync between local and remote backends
   - Automatic caching for remote backends
   - Offline mode with operation queuing
   - Conflict resolution strategies (server_wins, local_wins, merge, keep_both)
   - Background auto-sync daemon
   - Sync queue management and retry logic

### Customization & Security

6. **[Credential Management](./CREDENTIAL_MANAGEMENT.md)** - Secure credential storage
   - System keyring integration (macOS Keychain, Windows Credential Manager, Linux Secret Service)
   - Environment variable support for CI/CD
   - Legacy config URL support
   - Credential priority resolution
   - Per-backend credential management

7. **[Views & Customization](./VIEWS_CUSTOMIZATION.md)** - Custom display formats
   - Built-in views (default, all)
   - Custom field selection and ordering
   - Plugin-based formatters (external scripts)
   - Filtering and sorting
   - Interactive view builder TUI
   - Date, priority, and status formatters

### Interface & Configuration

8. **[CLI Interface](./CLI_INTERFACE.md)** - Command-line interface features
   - Cobra framework with subcommands
   - Shell completion (Bash, Zsh, Fish, PowerShell)
   - Interactive list selection
   - Intelligent task matching (exact, partial, multiple)
   - Terminal width detection for dynamic formatting
   - Verbose/debug logging
   - Action abbreviations (a=add, u=update, c=complete, d=delete, g=get)

9. **[Configuration System](./CONFIGURATION.md)** - Application configuration
   - XDG Base Directory compliance
   - YAML configuration format
   - Multi-backend configuration
   - Sync configuration (global and per-backend)
   - Path expansion (~, $HOME, environment variables)
   - Config file auto-initialization from embedded sample
   - Custom config path support

## Quick Feature Reference

| Feature Category | Key Capabilities | Primary Commands |
|-----------------|------------------|------------------|
| **Task Operations** | Add, update, delete, complete tasks | `add`, `update`, `complete`, `delete`, `get` |
| **Task Filtering** | Filter by status, priority, dates, tags | `-s`, `-p`, `--due-date`, `--start-date` |
| **Lists** | Create, rename, delete, restore lists | `list create`, `list rename`, `list delete`, `list trash` |
| **Subtasks** | Hierarchical task organization | `-P "Parent"`, path notation `parent/child` |
| **Backends** | Multi-backend support with auto-detection | `--backend`, `--list-backends`, `--detect-backend` |
| **Sync** | Offline sync with conflict resolution | `sync`, `sync status`, `sync queue` |
| **Credentials** | Secure keyring storage | `credentials set`, `credentials get`, `credentials delete` |
| **Views** | Customizable display formats | `view list`, `view create`, `view show`, `-v <view-name>` |
| **Shell Integration** | Tab completion for all shells | `completion bash/zsh/fish/powershell` |

## Feature Maturity

| Status | Features |
|--------|----------|
| ✅ **Stable** | Task CRUD, List management, Nextcloud backend, Todoist backend, SQLite sync, Credential management, Custom views, Subtasks, Shell completion |
| 🚧 **In Development** | File backend implementation, Auto-sync (currently disabled, being redesigned) |
| 📋 **Planned** | Cross-backend migration, Google Tasks backend, Microsoft To Do backend, TUI/GUI interface |

## Getting Started

For installation and basic usage, see the main [README.md](../README.md) in the project root.

For detailed sync documentation, see [SYNC_GUIDE.md](../SYNC_GUIDE.md).

For development guidance, see [CLAUDE.md](../CLAUDE.md).

## Feature Category Details

Click on any category below to explore detailed documentation:

### 📝 [Task Management](./TASK_MANAGEMENT.md)
Learn how to create, update, filter, and manage individual tasks with all available metadata fields and operations.

### 📋 [List Management](./LIST_MANAGEMENT.md)
Discover how to organize tasks into lists, manage list properties, and use the trash system for safe deletion.

### 🌳 [Subtasks & Hierarchy](./SUBTASKS_HIERARCHY.md)
Understand hierarchical task organization with parent-child relationships and path-based shortcuts.

### 🔌 [Backend System](./BACKEND_SYSTEM.md)
Explore the multi-backend architecture, backend selection logic, and how each backend type works.

### 🔄 [Synchronization](./SYNCHRONIZATION.md)
Master offline synchronization, conflict resolution, and the background sync daemon system.

### 🔐 [Credential Management](./CREDENTIAL_MANAGEMENT.md)
Learn secure credential storage using OS keyrings, environment variables, and credential resolution.

### 🎨 [Views & Customization](./VIEWS_CUSTOMIZATION.md)
Customize task display with views, formatters, plugins, and filtering/sorting options.

### 💻 [CLI Interface](./CLI_INTERFACE.md)
Understand the command-line interface, shell completion, interactive features, and usage patterns.

### ⚙️ [Configuration System](./CONFIGURATION.md)
Configure gosynctasks with YAML files, manage multiple backends, and customize application behavior.


## Status Mappings

gosynctasks uses different status representations for different backends:

| Internal Status | CalDAV Status | CLI Abbreviation | Display |
|----------------|---------------|------------------|---------|
| TODO | NEEDS-ACTION | T | ⬜ TODO |
| DONE | COMPLETED | D | ✅ DONE |
| PROCESSING | IN-PROCESS | P | ▶️ PROCESSING |
| CANCELLED | CANCELLED | C | ❌ CANCELLED |

## Data Model

### Task Object
Tasks follow the iCalendar VTODO standard with these fields:
- **UID**: Unique identifier (auto-generated)
- **Summary**: Task title/name
- **Description**: Detailed task description
- **Status**: TODO, DONE, PROCESSING, CANCELLED
- **Priority**: 0-9 (0=undefined, 1=highest, 9=lowest)
- **Due Date**: When task should be completed
- **Start Date**: When task should begin
- **Created**: Task creation timestamp
- **Modified**: Last modification timestamp
- **Completed**: Completion timestamp
- **Categories**: Tags/labels (comma-separated)
- **ParentUID**: Parent task for subtasks

### TaskList Object
Task lists/calendars with these properties:
- **ID**: Unique identifier
- **Name**: List display name
- **Description**: List description
- **CTag**: Change tag for sync (CalDAV)
- **Color**: Hex color code

## XDG Base Directory Compliance

gosynctasks follows the XDG Base Directory Specification:

| Type | Location | Purpose |
|------|----------|---------|
| **Config** | `$XDG_CONFIG_HOME/gosynctasks/` | Configuration files (`config.yaml`) |
| **Data** | `$XDG_DATA_HOME/gosynctasks/` | SQLite databases, sync state |
| **Cache** | `$XDG_CACHE_HOME/gosynctasks/` | Task list cache (`lists.json`) |
| **Views** | `$XDG_CONFIG_HOME/gosynctasks/views/` | Custom view definitions |

Default locations (if XDG variables not set):
- Config: `~/.config/gosynctasks/`
- Data: `~/.local/share/gosynctasks/`
- Cache: `~/.cache/gosynctasks/`

## Support & Resources

- **Project Repository**: https://github.com/DeepReef11/gosynctasks
- **Issue Tracker**: https://github.com/DeepReef11/gosynctasks/issues
- **Discussions**: https://github.com/DeepReef11/gosynctasks/discussions
- **License**: BSD-2-Clause

## Contributing

For contribution guidelines, see [CONTRIBUTING.md](../.github/CONTRIBUTING.md) in the project root.

For testing procedures, see [TESTING.md](../TESTING.md).

---

**Note**: This documentation reflects the current state of gosynctasks. Some features may be under active development. Check the project repository for the latest updates.
