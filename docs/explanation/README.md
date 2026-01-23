# todoat Complete Feature Inventory

**Version:** Based on current codebase (January 2026)

This documentation provides a comprehensive inventory of all features available in todoat, a fast, flexible, multi-backend task synchronization tool written in Go.

## What is todoat?

todoat is a command-line task management tool that allows you to:

- Manage tasks seamlessly from your terminal
- Work with multiple storage backends (Nextcloud CalDAV, Todoist, Google Tasks, Git/Markdown, File, SQLite)
- Sync tasks across devices with offline support
- Organize tasks hierarchically with subtasks
- Customize task views with flexible formatters
- Store credentials securely using OS keyrings

## Documentation Files

### Core Concepts
- [Features Overview](features-overview.md)
- [CLI Interface](cli-interface.md)
- [Configuration](configuration.md)
- [Task Management](task-management.md)
- [List Management](list-management.md)

### Backend & Sync
- [Backend System](backend-system.md)
- [Synchronization](synchronization.md)
- [Credential Management](credential-management.md)

### Advanced Features
- [Subtasks & Hierarchy](subtasks-hierarchy.md)
- [Views Customization](views-customization.md)
- [Notification Manager](notification-manager.md)
- [Analytics](analytics.md)
- [Logging](logging.md)

### Development
- [Test-Driven Development](test-driven-dev.md)
- [README Planner](readme-planner.md)

## Quick Feature Reference

| Feature Category | Key Capabilities | Primary Commands |
|-----------------|------------------|------------------|
| **Task Operations** | Add, update, delete, complete tasks | `add`, `update`, `complete`, `delete`, `get` |
| **Task Filtering** | Filter by status, priority, dates, tags | `-s`, `-p`, `--due-date`, `--start-date` |
| **Lists** | Create, update, delete, restore lists | `list create`, `list update`, `list delete`, `list trash` |
| **Subtasks** | Hierarchical task organization | `-P "Parent"`, path notation `parent/child` |
| **Backends** | Multi-backend support with auto-detection | `--backend`, `--detect-backend` |
| **Sync** | Offline sync with conflict resolution | `sync`, `sync status`, `sync queue` |
| **Credentials** | Secure keyring storage | `credentials set`, `credentials get`, `credentials delete` |
| **Views** | Customizable display formats | `view list`, `view create`, `-v <view-name>` |
| **Shell Integration** | Tab completion for all shells | `completion bash/zsh/fish/powershell` |

## Feature Maturity

| Status | Features |
|--------|----------|
| ✅ **Stable** | Task CRUD, List management, Nextcloud backend, Todoist backend, Google Tasks backend, Git backend, File backend, SQLite sync, Credential management, Custom views, Subtasks, Shell completion, Cross-backend migration, TUI interface |
| 🚧 **In Development** | Auto-sync daemon (being redesigned) |
| 📋 **Planned** | Microsoft To Do backend (backend code exists, not yet wired to CLI) |

## Getting Started

For detailed sync documentation, see [Synchronization](synchronization.md).

## Status Mappings

todoat uses different status representations for different backends:

| Internal Status | CalDAV Status | CLI Abbreviation | Display |
|----------------|---------------|------------------|---------|
| TODO | NEEDS-ACTION | T | ⬜ TODO |
| DONE | COMPLETED | D | ✅ DONE |
| IN-PROGRESS | IN-PROCESS | I | ▶️ IN-PROGRESS |
| CANCELLED | CANCELLED | C | ❌ CANCELLED |

## Data Model

### Task Object
Tasks follow the iCalendar VTODO standard with these fields:
- **UID**: Unique identifier (auto-generated)
- **Summary**: Task title/name
- **Description**: Detailed task description
- **Status**: TODO, DONE, IN-PROGRESS, CANCELLED
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

todoat follows the XDG Base Directory Specification:

| Type | Location | Purpose |
|------|----------|---------|
| **Config** | `$XDG_CONFIG_HOME/todoat/` | Configuration files (`config.yaml`) |
| **Data** | `$XDG_DATA_HOME/todoat/` | SQLite databases, sync state |
| **Cache** | `$XDG_CACHE_HOME/todoat/` | Task list cache (`lists.json`) |
| **Views** | `$XDG_CONFIG_HOME/todoat/views/` | Custom view definitions |

Default locations (if XDG variables not set):
- Config: `~/.config/todoat/`
- Data: `~/.local/share/todoat/`
- Cache: `~/.cache/todoat/`

## Support & Resources

- **Project Repository**: https://github.com/DeepReef11/todoat
- **Issue Tracker**: https://github.com/DeepReef11/todoat/issues
- **Discussions**: https://github.com/DeepReef11/todoat/discussions
- **License**: BSD-2-Clause

## Contributing

For testing procedures, see [Test-Driven Development](test-driven-dev.md).

---

**Note**: This documentation reflects the current state of todoat. Some features may be under active development. Check the project repository for the latest updates.
