# todoat Feature Overview

This document provides a high-level summary of all features in todoat, organized by functional category. Each feature includes a brief description and links to detailed documentation.

## Table of Contents

- [Task Management Features](#task-management-features)
- [List Management Features](#list-management-features)
- [Subtask & Hierarchy Features](#subtask--hierarchy-features)
- [Backend System Features](#backend-system-features)
- [Synchronization Features](#synchronization-features)
- [Credential Management Features](#credential-management-features)
- [Views & Customization Features](#views--customization-features)
- [CLI Interface Features](#cli-interface-features)
- [Configuration Features](#configuration-features)
- [Analytics Features](#analytics-features)

---

## Task Management Features

**Purpose**: Core task operations for creating, modifying, and managing individual tasks.

**Status System**: todoat uses internal status values (TODO, DONE, IN-PROGRESS, CANCELLED) for all operations. These are automatically translated to/from backend-specific status formats (e.g., CalDAV uses NEEDS-ACTION, COMPLETED, IN-PROCESS).

| Feature | Description | Documentation |
|---------|-------------|---------------|
| **Add Tasks** | Create new tasks with metadata (priority, dates, descriptions, tags) | [Task Management](task-management.md#add-tasks) |
| **View Tasks** | Display tasks from a list with filtering and custom views | [Task Management](task-management.md#view-tasks) |
| **Update Tasks** | Modify task properties (summary, status, priority, dates, etc.) | [Task Management](task-management.md#update-tasks) |
| **Complete Tasks** | Mark tasks as done with automatic completion timestamp | [Task Management](task-management.md#complete-tasks) |
| **Delete Tasks** | Remove tasks from a list (immediate deletion) | [Task Management](task-management.md#delete-tasks) |
| **Filter by Status** | Show only tasks with specific statuses (TODO, DONE, IN-PROGRESS, CANCELLED) | [Task Management](task-management.md#filter-by-status) |
| **Filter by Priority** | Display tasks matching priority levels (0-9) | [Task Management](task-management.md#filter-by-priority) |
| **Filter by Dates** | Filter tasks by due date, start date, or date ranges | [Task Management](task-management.md#filter-by-dates) |
| **Filter by Tags** | Show tasks with specific categories/tags | [Task Management](task-management.md#filter-by-tags) |
| **Task Disambiguation** | Exact and partial matching with UID-based disambiguation when multiple tasks match | [Interactive UX](interactive-ux.md#task-disambiguation) |
| **Task Search** | Intelligent search with exact, partial, and multi-match support | [Task Management](task-management.md#task-search) |
| **Bulk Operations** | Operate on multiple tasks using filters | [Task Management](task-management.md#bulk-operations) |
| **Pagination** | Control output size with limit, offset, and page flags for large task lists | [Task Management](../how-to/task-management.md#pagination) |

**Key Use Cases**:
- Daily task management (adding todos, marking complete)
- Project task tracking with priorities and deadlines
- Task filtering for focused work sessions
- Status reporting and progress tracking

**Related Features**: [List Management](#list-management-features), [Subtasks](#subtask--hierarchy-features), [Views](#views--customization-features)

---

## List Management Features

**Purpose**: Organize tasks into separate lists (calendars) with distinct properties and management capabilities.

| Feature | Description | Documentation |
|---------|-------------|---------------|
| **Create Lists** | Create new task lists with names, descriptions, and colors | [List Management](list-management.md#create-lists) |
| **View Lists** | Display all available task lists with metadata | [List Management](list-management.md#view-lists) |
| **Rename Lists** | Change list display names | [List Management](list-management.md#rename-lists) |
| **Delete Lists** | Soft-delete lists (move to trash) | [List Management](list-management.md#delete-lists) |
| **Trash Management** | View deleted lists in trash | [List Management](list-management.md#trash-management) |
| **Restore Lists** | Recover lists from trash | [List Management](list-management.md#restore-lists) |
| **Purge Lists** | Permanently delete lists from trash | [List Management](list-management.md#purge-lists) |
| **List Information** | Show detailed list metadata (ID, color, description, task count) | [List Management](list-management.md#list-information) |
| **List Caching** | Local cache of lists for faster access | [List Management](list-management.md#list-caching) |
| **Interactive List Selection** | Choose list from menu when not specified | [List Management](list-management.md#interactive-list-selection) |
| **List Color Coding** | Assign hex colors to lists for visual distinction | [List Management](list-management.md#list-color-coding) |

**Key Use Cases**:
- Organizing tasks by project, context, or area of responsibility
- Separating personal and work tasks
- Archiving completed projects (move to trash)
- Recovering accidentally deleted lists

**Related Features**: [Task Management](#task-management-features), [Backend System](#backend-system-features), [Synchronization](#synchronization-features)

---

## Subtask & Hierarchy Features

**Purpose**: Create hierarchical task structures with parent-child relationships for complex projects.

| Feature | Description | Documentation |
|---------|-------------|---------------|
| **Create Subtasks** | Add child tasks under parent tasks using `-P` flag | [Subtasks & Hierarchy](subtasks-hierarchy.md#1-parent-child-relationships) |
| **Path-Based Creation** | Auto-create hierarchy using path notation (`parent/child/grandchild`) | [Subtasks & Hierarchy](subtasks-hierarchy.md#2-path-based-task-creation) |
| **Tree Visualization** | Display hierarchical structure with box-drawing characters (├─, └─, │) | [Subtasks & Hierarchy](subtasks-hierarchy.md#3-hierarchical-display-and-navigation) |
| **Multi-Level Hierarchy** | Support unlimited nesting depth | [Subtasks & Hierarchy](subtasks-hierarchy.md#1-parent-child-relationships) |
| **Parent Path Resolution** | Reference parents by path or name | [Subtasks & Hierarchy](subtasks-hierarchy.md#2-path-based-task-creation) |
| **Hierarchical Filtering** | Show subtasks when parent matches filter | [Subtasks & Hierarchy](subtasks-hierarchy.md#integration-with-other-features) |
| **Orphan Detection** | Identify subtasks with missing parents | [Subtasks & Hierarchy](subtasks-hierarchy.md#4-subtask-operations-and-management) |
| **Indented Display** | Visual indentation to show task depth | [Subtasks & Hierarchy](subtasks-hierarchy.md#3-hierarchical-display-and-navigation) |

**Key Use Cases**:
- Breaking down complex projects into manageable subtasks
- Creating work breakdown structures (WBS)
- Organizing multi-step processes
- Tracking dependencies and task relationships

**Related Features**: [Task Management](#task-management-features), [Synchronization](#synchronization-features)

---

## Backend System Features

**Purpose**: Connect to multiple task storage backends with a unified interface.

| Feature | Description | Documentation |
|---------|-------------|---------------|
| **Nextcloud CalDAV Backend** | Sync with Nextcloud Tasks using CalDAV protocol | [Backend System](backend-system.md#1-nextcloud-backend-remote-caldav) |
| **Todoist Backend** | Integrate with Todoist using REST API | [Backend System](backend-system.md#available-backends) |
| **Google Tasks Backend** | Integrate with Google Tasks using REST API | [Backend System](backend-system.md#available-backends) |
| **Microsoft To Do Backend** | Integrate with Microsoft To Do using Graph API | [Backend System](backend-system.md#available-backends) |
| **SQLite Backend** | Local database storage with full CRUD operations | [Backend System](backend-system.md#2-sqlite-backend-local-database) |
| **Git/Markdown Backend** | Store tasks as markdown files in Git repositories | [Backend System](backend-system.md#3-git-backend-markdown-in-repositories) |
| **File Backend** | Plain text file storage | [Backend System](backend-system.md#4-file-backend-plain-text-storage) |
| **Backend Auto-Detection** | Automatically detect and configure backends | [Backend System](backend-system.md#auto-detection-interface) |
| **Backend Selection Priority** | Configurable priority order for backend selection | [Backend System](backend-system.md#selection-priority) |
| **Pluggable Architecture** | TaskManager interface for adding new backends | [Backend System](backend-system.md#1-taskmanager-interface) |
| **List Backends Command** | Display all configured backends and their status | [Backend System](backend-system.md#backend-display-information) |
| **Backend-Specific Options** | Per-backend configuration and behavior | [Backend System](backend-system.md#backend-configuration-formats) |
| **Multi-Backend Support** | Use multiple backends simultaneously | [Backend System](backend-system.md#2-backend-registry) |

**Key Use Cases**:
- Syncing tasks with Nextcloud tasks
- Integrating with Todoist projects
- Syncing with Google Tasks
- Syncing with Microsoft To Do
- Offline-first task management with SQLite
- Version-controlled task files with Git
- Switching between backends based on context

**Related Features**: [Synchronization](#synchronization-features), [Credential Management](#credential-management-features), [Configuration](#configuration-features)

---

## Synchronization Features

**Purpose**: Enable offline task management with bidirectional sync and conflict resolution.

**Architecture**: When sync is enabled, the Sync Manager coordinates all operations between CLI commands, local cache (SQLite), and remote backends. All CRUD operations are routed through the Sync Manager to ensure consistency.

| Feature | Description | Documentation |
|---------|-------------|---------------|
| **Bidirectional Sync** | Sync local and remote changes in both directions | [Synchronization](synchronization.md#bidirectional-sync) |
| **Automatic Caching** | Each remote backend gets its own SQLite cache database | [Synchronization](synchronization.md#automatic-caching) |
| **Offline Mode** | Queue operations when remote backend unavailable | [Synchronization](synchronization.md#offline-mode) |
| **Manual Sync** | Trigger sync with `todoat sync` command | [Synchronization](synchronization.md#manual-sync) |
| **Auto-Sync Daemon** | Background process for automatic syncing | [Synchronization](synchronization.md#auto-sync-daemon) |
| **Conflict Resolution** | Handle conflicts with configurable strategies | [Synchronization](synchronization.md#conflict-resolution) |
| **Server Wins Strategy** | Remote changes override local changes (default) | [Synchronization](synchronization.md#server-wins-strategy) |
| **Local Wins Strategy** | Local changes override remote changes | [Synchronization](synchronization.md#local-wins-strategy) |
| **Merge Strategy** | Combine local and remote changes intelligently | [Synchronization](synchronization.md#merge-strategy) |
| **Keep Both Strategy** | Create duplicate tasks to preserve both versions | [Synchronization](synchronization.md#keep-both-strategy) |
| **Sync Status** | View sync state and pending operations | [Synchronization](synchronization.md#sync-status) |
| **Sync Queue** | Persistent queue of pending operations | [Synchronization](synchronization.md#sync-queue) |
| **Retry Logic** | Exponential backoff for failed sync operations | [Synchronization](synchronization.md#retry-logic) |
| **ETag Support** | Use ETags for efficient change detection | [Synchronization](synchronization.md#etag-support) |
| **CTag Support** | Collection tags for list-level change detection | [Synchronization](synchronization.md#ctag-support) |
| **Hierarchical Sync** | Sync parent tasks before children (FK preservation) | [Synchronization](synchronization.md#hierarchical-sync) |
| **Sync Metadata Tracking** | Track sync state per task and list | [Synchronization](synchronization.md#sync-metadata-tracking) |
| **Per-Backend Sync Config** | Enable/disable sync for individual backends | [Synchronization](synchronization.md#per-backend-sync-config) |

**Key Use Cases**:
- Working offline on flights or without internet
- Syncing tasks across multiple devices
- Preventing data loss from concurrent edits
- Managing conflicts from simultaneous changes
- Batch syncing after offline work sessions

**Related Features**: [Backend System](#backend-system-features), [Task Management](#task-management-features), [Configuration](#configuration-features)

---

## Background Operations

### Notification System

- **Desktop Notifications**: OS-native notifications for sync events
- **Log Notifications**: Persistent log file for background operations
- **Configurable Events**: Choose which events trigger notifications

See [Notification Manager](notification-manager.md) for configuration details.

---

## Credential Management Features

**Purpose**: Securely store and manage backend authentication credentials.

| Feature | Description | Documentation |
|---------|-------------|---------------|
| **Keyring Storage** | Store passwords in OS-native keyring (most secure) | [Credential Management](credential-management.md#keyring-storage) |
| **macOS Keychain Integration** | Use macOS Keychain for credential storage | [Credential Management](credential-management.md#macos-keychain-integration) |
| **Windows Credential Manager** | Use Windows Credential Manager for storage | [Credential Management](credential-management.md#windows-credential-manager) |
| **Linux Secret Service** | Use freedesktop.org Secret Service API (GNOME Keyring, KWallet) | [Credential Management](credential-management.md#linux-secret-service) |
| **Environment Variables** | Load credentials from environment variables (good for CI/CD) | [Credential Management](credential-management.md#environment-variables) |
| **Config URL Support** | Legacy support for credentials in config URLs | [Credential Management](credential-management.md#config-url-support) |
| **Credential Priority** | Resolve credentials from keyring → env vars → config URL | [Credential Management](credential-management.md#credential-priority) |
| **Set Credentials** | Store credentials securely with `credentials set` command | [Credential Management](credential-management.md#set-credentials) |
| **Get Credentials** | Retrieve and verify credentials with `credentials get` | [Credential Management](credential-management.md#get-credentials) |
| **Delete Credentials** | Remove credentials from keyring | [Credential Management](credential-management.md#delete-credentials) |
| **Per-Backend Credentials** | Separate credential management for each backend | [Credential Management](credential-management.md#per-backend-credentials) |
| **Password Prompt** | Interactive password input with hidden entry | [Credential Management](credential-management.md#password-prompt) |
| **Credential Verification** | Test credentials before saving | [Credential Management](credential-management.md#credential-verification) |

**Key Use Cases**:
- Storing Nextcloud CalDAV credentials securely
- Managing API tokens for Todoist
- Setting up CI/CD pipelines with environment variables
- Rotating credentials without editing config files
- Multi-account support for different backends

**Related Features**: [Backend System](#backend-system-features), [Configuration](#configuration-features)

---

## Views & Customization Features

**Purpose**: Customize how tasks are displayed with flexible views, formatters, and filters.

| Feature | Description | Documentation |
|---------|-------------|---------------|
| **Built-in Views** | Pre-configured views (`default`, `all`) | [Views & Customization](views-customization.md#built-in-views) |
| **Custom Views** | User-defined views with custom field selection | [Views & Customization](views-customization.md#custom-views) |
| **Field Selection** | Choose which task fields to display | [Views & Customization](views-customization.md#field-selection) |
| **Field Ordering** | Control the order of displayed fields | [Views & Customization](views-customization.md#field-ordering) |
| **View Filters** | Filter tasks by status, priority, tags, dates within views | [Views & Customization](views-customization.md#view-filters) |
| **View Sorting** | Sort tasks by any field (ascending/descending) | [Views & Customization](views-customization.md#view-sorting) |
| **Plugin Formatters** | External scripts for custom field formatting | [Views & Customization](views-customization.md#plugin-formatters) |
| **Date Formatters** | Format date fields with custom patterns | [Views & Customization](views-customization.md#date-formatters) |
| **Priority Formatters** | Customize priority display (numbers, emojis, colors) | [Views & Customization](views-customization.md#priority-formatters) |
| **Status Formatters** | Customize status display with emojis or text | [Views & Customization](views-customization.md#status-formatters) |
| **Interactive View Builder** | TUI for creating views without editing YAML | [Views & Customization](views-customization.md#interactive-view-builder) |
| **View Storage** | YAML-based view definitions in `~/.config/todoat/views/` | [Views & Customization](views-customization.md#view-storage) |
| **List Views** | Display all available views | [Views & Customization](views-customization.md#list-views) |
| **Show View Definition** | Display view YAML configuration | [Views & Customization](views-customization.md#show-view-definition) |
| **Plugin Script Support** | Bash, Python, Ruby, and other language support for plugins | [Views & Customization](views-customization.md#plugin-script-support) |
| **Plugin Timeout Management** | Prevent hanging formatters with timeout enforcement | [Views & Customization](views-customization.md#plugin-timeout-management) |
| **Hierarchical View Support** | Views respect task tree structures | [Views & Customization](views-customization.md#hierarchical-view-support) |

**Key Use Cases**:
- Creating minimal views for daily standup meetings
- Detailed views for project status reports
- Custom formatting for GTD methodology
- Priority-focused views for urgent work
- Script-based formatting for rich terminal output

**Related Features**: [Task Management](#task-management-features), [CLI Interface](#cli-interface-features)

---

## CLI Interface Features

**Purpose**: Provide a powerful, user-friendly command-line interface with intelligent defaults.

| Feature | Description | Documentation |
|---------|-------------|---------------|
| **Cobra Framework** | Robust command structure with subcommands | [CLI Interface](cli-interface.md#cobra-framework) |
| **Shell Completion** | Tab completion for Bash, Zsh, Fish, PowerShell with auto-install | [CLI Interface](cli-interface.md#shell-completion) |
| **Interactive List Selection** | Choose from menu when list name not specified | [CLI Interface](cli-interface.md#interactive-list-selection) |
| **Intelligent Task Matching** | Exact → partial → multiple match resolution | [CLI Interface](cli-interface.md#intelligent-task-matching) |
| **Terminal Width Detection** | Dynamic formatting based on terminal size | [CLI Interface](cli-interface.md#terminal-width-detection) |
| **Verbose Output** | Detailed logging with `-v` or `--verbose` flag | [CLI Interface](cli-interface.md#verbose-output) |
| **Debug Mode** | Extended diagnostic output for troubleshooting | [CLI Interface](cli-interface.md#debug-mode) |
| **Action Abbreviations** | Short aliases (a=add, u=update, c=complete, d=delete, g=get) | [CLI Interface](cli-interface.md#action-abbreviations) |
| **Flag Shortcuts** | Short flags (`-s` for status, `-p` for priority, etc.) | [CLI Interface](cli-interface.md#flag-shortcuts) |
| **Help System** | Contextual help with `--help` on all commands | [CLI Interface](cli-interface.md#help-system) |
| **Error Messages** | Clear, actionable error reporting | [CLI Interface](cli-interface.md#error-messages) |
| **Confirmation Prompts** | User confirmation for destructive operations | [CLI Interface](cli-interface.md#confirmation-prompts) |
| **Colored Output** | Color-coded display for better readability | [CLI Interface](cli-interface.md#colored-output) |
| **Table Display** | Formatted table output for list views | [CLI Interface](cli-interface.md#table-display) |
| **Version Command** | Display version and build information | [CLI Interface](cli-interface.md#version-command) |
| **No-Prompt Mode** | Non-interactive mode for scripting (`-y`, `--no-prompt`) | [CLI Interface](cli-interface.md#no-prompt-mode) |
| **JSON Output** | Machine-parseable JSON output (`--json`) | [CLI Interface](cli-interface.md#json-output-mode) |
| **Result Codes** | Standardized operation outcome indicators | [CLI Interface](cli-interface.md#result-codes) |
| **UID Selection** | Direct task selection by UID (`--uid`) | [CLI Interface](cli-interface.md#action-flags) |

**Key Use Cases**:
- Rapid task entry from terminal
- Scripting task operations
- Integrating with shell workflows
- Quick task lookups during development
- Automating task management
- CI/CD pipeline integration with no-prompt mode
- Machine-parseable output for automation tools

**Related Features**: [Task Management](#task-management-features), [Views & Customization](#views--customization-features), [Configuration](#configuration-features)

---

## Configuration Features

**Purpose**: Flexible, standards-compliant configuration system for customizing behavior.

| Feature | Description | Documentation |
|---------|-------------|---------------|
| **YAML Configuration** | Human-readable YAML format for all settings | [Configuration](configuration.md#yaml-configuration) |
| **XDG Compliance** | Follows XDG Base Directory Specification | [Configuration](configuration.md#xdg-compliance) |
| **Multi-Backend Config** | Configure multiple backends in single file | [Configuration](configuration.md#multi-backend-config) |
| **Sync Configuration** | Global and per-backend sync settings | [Configuration](configuration.md#sync-configuration) |
| **Path Expansion** | Expand `~`, `$HOME`, and environment variables in paths | [Configuration](configuration.md#path-expansion) |
| **Auto-Initialization** | Create config from embedded sample on first run | [Configuration](configuration.md#auto-initialization) |
| **Config Validation** | Verify configuration on load with helpful errors | [Configuration](configuration.md#config-validation) |
| **Default Backend** | Set preferred backend when multiple configured | [Configuration](configuration.md#default-backend) |
| **Backend Priority** | Ordered list for backend selection | [Configuration](configuration.md#backend-priority) |
| **Conflict Resolution Config** | Configure sync conflict strategies | [Configuration](configuration.md#conflict-resolution-config) |
| **View Defaults** | Set default view for task display | [Configuration](configuration.md#view-defaults) |
| **Cache Configuration** | Configure cache paths and behavior | [Configuration](configuration.md#cache-configuration) |
| **No-Prompt Mode Config** | Configure non-interactive mode default | [Configuration](configuration.md#no-prompt-mode-configuration) |
| **Output Format Config** | Configure default output format (text/json) | [Configuration](configuration.md#output-format-configuration) |
| **Singleton Pattern** | Ensure single config instance across application | [Configuration](configuration.md#singleton-pattern) |

**Key Use Cases**:
- Setting up multiple backends (work and personal)
- Configuring sync behavior and conflict resolution
- Customizing default views and display options
- Specifying credential sources
- Managing environment-specific configurations
- Enabling non-interactive mode for automated environments
- Setting JSON output as default for script-heavy workflows

**Related Features**: [Backend System](#backend-system-features), [Synchronization](#synchronization-features), [Credential Management](#credential-management-features)

---

## Analytics Features

**Purpose**: Track command usage, backend performance, and errors locally for insights into usage patterns.

| Feature | Description | Documentation |
|---------|-------------|---------------|
| **Command Statistics** | View usage counts and success rates by command | [Analytics](analytics.md#viewing-analytics-data) |
| **Backend Performance** | Track performance metrics for each backend | [Analytics](analytics.md#viewing-analytics-data) |
| **Error Tracking** | View most common errors grouped by command | [Analytics](analytics.md#viewing-analytics-data) |
| **Time Filtering** | Filter analytics by time range (7d, 30d, 1y) | [Analytics](analytics.md#viewing-analytics-data) |
| **JSON Output** | Machine-parseable analytics output | [Analytics](analytics.md#viewing-analytics-data) |
| **Privacy-First** | All data stored locally, never transmitted | [Analytics](analytics.md#privacy-considerations) |
| **Enabled by Default** | Analytics enabled by default with clear notice; can be disabled in config | [Analytics](analytics.md#configuration) |

**Key Use Cases**:
- Understanding command usage patterns
- Identifying slow or failing backends
- Debugging common errors
- Performance monitoring over time

**Related Features**: [Configuration](#configuration-features), [Backend System](#backend-system-features)

---

## Feature Interaction Map

This diagram shows how major features interact with each other:

```
┌─────────────────┐
│  Configuration  │──────┐
└─────────────────┘      │
                         ▼
┌─────────────────┐  ┌──────────────────┐  ┌─────────────────┐
│   Credentials   │──│  Backend System  │──│  List Mgmt      │
└─────────────────┘  └──────────────────┘  └─────────────────┘
                         │           │           │
                         ▼           ▼           ▼
                    ┌──────────────────────────────────┐
                    │      Task Management             │
                    │  (CRUD + Filtering + Search)     │
                    └──────────────────────────────────┘
                         │           │           │
                 ┌───────┘           │           └────────┐
                 ▼                   ▼                    ▼
        ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
        │  Subtasks    │    │     Sync     │    │    Views     │
        │  Hierarchy   │    │   Manager    │    │ Customization│
        └──────────────┘    └──────────────┘    └──────────────┘
                 │                   │                    │
                 └───────────┐       │       ┌────────────┘
                             ▼       ▼       ▼
                        ┌──────────────────────┐
                        │   CLI Interface      │
                        │ (Display + Commands) │
                        └──────────────────────┘
```

---

## Quick Feature Lookup

**I want to...**

- **Add tasks quickly**: Use [Add Tasks](task-management.md#add-tasks) with [Action Abbreviations](cli-interface.md#action-abbreviations)
- **Organize projects**: Use [Lists](list-management.md) and [Subtasks](subtasks-hierarchy.md)
- **Work offline**: Enable [Synchronization](synchronization.md) with [Offline Mode](synchronization.md#offline-mode)
- **Sync with Nextcloud**: Configure [Nextcloud Backend](backend-system.md#1-nextcloud-backend-remote-caldav) with [Keyring Storage](credential-management.md#keyring-storage)
- **Customize display**: Create [Custom Views](views-customization.md#custom-views) or [Plugin Formatters](views-customization.md#plugin-formatters)
- **Track deadlines**: Use [Filter by Dates](task-management.md#filter-by-dates) and [Due Date fields](task-management.md#add-tasks)
- **Prioritize work**: Use [Priority levels](task-management.md#add-tasks) and [Filter by Priority](task-management.md#filter-by-priority)
- **Use in scripts**: Leverage [CLI Interface](cli-interface.md) and [Shell Completion](cli-interface.md#shell-completion)
- **Automate task management**: Use [No-Prompt Mode](cli-interface.md#no-prompt-mode) with [JSON Output](cli-interface.md#json-output-mode) and [UID Selection](cli-interface.md#action-flags)
- **Integrate with CI/CD**: Enable [No-Prompt Mode](configuration.md#no-prompt-mode-configuration) in config for non-interactive operation
- **View usage statistics**: Use [Analytics Commands](analytics.md#viewing-analytics-data) to see command usage and backend performance

---

## Navigation

- **[← Back to Main Documentation](README.md)**
- **[View Detailed Category Documentation](README.md#feature-category-details)**
- **[View Architecture Overview](README.md#architecture-overview)**

---

**Last Updated**: January 2026
**Total Features Documented**: 136
**Documentation Version**: 1.0
