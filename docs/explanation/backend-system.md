# Backend System

## Overview

The Backend System in todoat provides a pluggable architecture that allows the application to store and retrieve tasks from multiple different storage providers. This design enables users to work with tasks in Nextcloud (CalDAV), local SQLite databases, Git repositories with markdown files, or plain files - all through a unified interface.

**Related Features:**
- [Synchronization](synchronization.md) - Syncing between local and remote backends
- [Configuration](configuration.md) - Backend configuration and setup
- [Credential Management](credential-management.md) - Secure credential storage for remote backends

---

## Core Concepts

### 1. TaskManager Interface

**Purpose:** Define a standard contract that all backends must implement, ensuring consistent behavior regardless of storage provider.

**How It Works:**
- All backends implement the `TaskManager` interface defined in `backend/taskManager.go:155`
- The interface specifies 20+ methods for task operations (CRUD), list management, status handling, and display formatting
- Key method categories:
  - **Task Operations:** `GetTasks()`, `FindTasksBySummary()`, `AddTask()`, `UpdateTask()`, `DeleteTask()`
  - **List Operations:** `GetTaskLists()`, `CreateTaskList()`, `DeleteTaskList()`, `RenameTaskList()`
  - **Trash Management:** `GetDeletedTaskLists()`, `RestoreTaskList()`, `PermanentlyDeleteTaskList()`
  - **Status Translation:** `ParseStatusFlag()`, `StatusToDisplayName()`
  - **Display Helpers:** `SortTasks()`, `GetPriorityColor()`, `GetBackendDisplayName()`

**Technical Details:**
- Interface is thread-safe when accessed from multiple goroutines
- Each backend provides its own interpretation while maintaining API compatibility
- Status values are backend-specific but translated through standard methods

---

### 2. Backend Registry

**Purpose:** Enable dynamic backend registration and discovery using a factory pattern, supporting extensibility without modifying core code.

**How It Works:**
- Global registry (`backend/registry.go:22`) maintains three types of constructors:
  1. **Scheme Constructors:** Map URL schemes (e.g., `nextcloud://`) to backend implementations
  2. **Type Constructors:** Map backend types (e.g., `"nextcloud"`) to implementations
  3. **Detectable Constructors:** Backends that support auto-detection (e.g., Git)
- Each backend registers itself in its `init()` function during package initialization
- Registration examples:
  - Nextcloud: `backend.RegisterScheme("nextcloud", NewNextcloudBackend)` and `backend.RegisterType("nextcloud", ...)`
  - Git: `backend.RegisterDetectable("git", newGitBackendWrapper)`
  - SQLite: `backend.RegisterType("sqlite", newSQLiteBackendWrapper)`

**Technical Details:**
- Registry uses `sync.RWMutex` for thread-safe concurrent access
- Detectable backends also register as regular types
- `GetSchemeConstructor()` and `GetTypeConstructor()` retrieve registered constructors

---

## Available Backends

### 1. Nextcloud Backend (Remote CalDAV)

**Purpose:** Integrate with Nextcloud's task management via standard CalDAV protocol, enabling cloud storage and multi-device access.

**How It Works:**

1. **Connection Establishment:**
   - User provides Nextcloud URL and credentials (via [Credential Management](credential-management.md))
   - Backend constructs CalDAV endpoint: `https://host/remote.php/dav/calendars/username/`
   - HTTP client configured with TLS settings, connection pooling (10 max idle, 2 per host), 30s timeout

2. **Authentication:**
   - Supports three credential sources (priority order):
     1. System keyring (most secure)
     2. Environment variables (`TODOAT_NEXTCLOUD_USERNAME`, `TODOAT_NEXTCLOUD_PASSWORD`)
     3. URL-embedded credentials (legacy, deprecated)
   - Basic auth credentials cached after first resolution

3. **Data Format:**
   - Tasks stored as iCalendar VTODO components
   - **Status Translation (Internal ↔ CalDAV Backend):**
     - Internal `TODO` ↔ CalDAV `NEEDS-ACTION`
     - Internal `DONE` ↔ CalDAV `COMPLETED`
     - Internal `IN-PROGRESS` ↔ CalDAV `IN-PROCESS`
     - Internal `CANCELLED` ↔ CalDAV `CANCELLED`
     - Translation occurs at storage/retrieval boundaries; all application logic uses internal status
   - Priority: 0-9 (0=undefined, 1=highest, 9=lowest)
   - Full VTODO properties: UID, Summary, Description, Status, Priority, Dates, Categories, Parent UID

4. **Operations:**
   - **List Retrieval:** PROPFIND request to calendar collection URL
   - **Task Retrieval:** CalDAV REPORT query with VTODO component filter
   - **Task Creation:** PUT request with iCalendar VTODO to task-specific URL
   - **Task Updates:** PUT with If-Match header using ETag for optimistic locking
   - **Task Deletion:** DELETE request to task URL

**User Journey:**
1. User configures Nextcloud backend in `config.yaml` with host and username
2. User stores credentials: `todoat credentials set nextcloud myuser --prompt`
3. User runs: `todoat --backend=nextcloud MyTasks` to access Nextcloud tasks
4. Backend establishes HTTPS connection, authenticates, and fetches task lists
5. User performs task operations; changes immediately sync to Nextcloud server

**Prerequisites:**
- Nextcloud instance with Tasks app installed
- Valid user account with calendar access
- Network connectivity to Nextcloud server
- See [Credential Management](credential-management.md) for credential setup

**Technical Details:**
- Implementation: `backend/nextcloud/backend.go`
- VTODO parser: `backend/nextcloud/vtodo_parser.go`
- Security features:
  - HTTPS enforcement (HTTP requires explicit `allow_http: true` flag)
  - Self-signed cert support via `insecure_skip_verify: true` (with warnings)
  - Warning suppression via `suppress_ssl_warning` and `suppress_http_warning`

**Related Features:**
- [Synchronization](synchronization.md#nextcloud-sync) - Offline caching with Nextcloud
- [Credential Management](credential-management.md#nextcloud-credentials) - Secure credential storage

---

### 2. SQLite Backend (Local Database)

**Purpose:** Provide fast local task storage with full SQL query capabilities and offline support, serving as cache for remote backends.

**How It Works:**

1. **Database Initialization:**
   - Database path determined by `db_path` config or default: `$XDG_DATA_HOME/todoat/tasks.db`
   - Schema creation with 5 tables (`backend/sqlite/schema.go:11`):
     - `tasks`: Task data (UID, Summary, Description, Status, Priority, Dates, Parent UID)
     - `sync_metadata`: Per-task sync state (ETag, sync flags, timestamps)
     - `list_sync_metadata`: Per-list sync state (CTag, sync tokens)
     - `sync_queue`: Pending operations for remote sync
     - `schema_version`: Migration tracking
   - Indexes optimize: `list_id`, `status`, `due_date`, `parent_uid`, `priority`, sync flags

2. **Multi-Backend Support:**
   - Each backend (Nextcloud, Todoist) gets isolated storage via `backend_name` field
   - Separate cache databases per remote: `~/.local/share/todoat/caches/nextcloud.db`
   - Tasks from different backends never mix

3. **CRUD Operations (via Sync Manager when sync enabled):**
   - When sync is enabled, all CRUD operations are coordinated by the Sync Manager
   - Sync Manager ensures consistency between cache and remote backends
   - **Create:** Auto-generates UID, sets timestamps, queues sync operation via Sync Manager
   - **Read:** SQL queries with filters for status, dates, hierarchy
   - **Update:** Transactional updates, Sync Manager marks as locally modified and queues sync
   - **Delete:** Sync Manager marks as locally deleted, queues deletion in sync queue

4. **Sync Support (Sync Manager Integration):**
   - **Sync Manager Role:** Orchestrates all synchronization between SQLite cache and remote backends
   - **Operation Flow:** CLI → Sync Manager → SQLite Backend → sync_queue table
   - **Metadata Tracking:** `locally_modified`, `locally_deleted`, `last_synced` flags managed by Sync Manager
   - **Remote ETags:** Stored for conflict detection during pull operations
   - **Operation Queue:** Maintains queued operations with retry counters for push operations
   - **Bidirectional Sync:** Sync Manager handles both pull (remote → cache) and push (queue → remote)
   - See [Synchronization](synchronization.md) for detailed Sync Manager workflow

**User Journey:**
1. User enables sync in config: `sync.enabled: true`
2. User configures Nextcloud as remote backend
3. System automatically creates `~/.local/share/todoat/caches/nextcloud.db`
4. User runs: `todoat sync` to pull remote tasks into local cache
5. User works offline: `todoat MyList add "Task"` - changes queued locally
6. User syncs when online: `todoat sync` - pushes queued changes to remote

**Prerequisites:**
- Write access to data directory (`$XDG_DATA_HOME` or `~/.local/share`)
- SQLite driver (included: `modernc.org/sqlite`)
- For sync features: configured remote backend

**Outputs/Results:**
- SQLite database file with task data
- Automatic schema migrations for version upgrades
- Operation queue persisted between application runs

**Technical Details:**
- Implementation: `backend/sqlite/backend.go`, `backend/sqlite/database.go`
- Schema: `backend/sqlite/schema.go`
- **Status Storage:** Stores internal application statuses (`TODO`, `DONE`, `IN-PROGRESS`, `CANCELLED`) directly
  - No translation needed for SQLite backend (unlike CalDAV which uses NEEDS-ACTION, COMPLETED, IN-PROCESS)
  - Sync Manager translates between internal status and backend-specific status when syncing with remote backends
- Hierarchical task support via `parent_uid` foreign key with `ON DELETE CASCADE`

**Related Features:**
- [Synchronization](synchronization.md) - Bidirectional sync with remote backends
- [Subtasks & Hierarchy](subtasks-hierarchy.md) - Parent-child relationships via `parent_uid`

---

### 3. Git Backend (Markdown in Repositories)

> **Note:** This backend is implemented but not yet wired to the CLI. It cannot be accessed via `--backend=git`. This is a planned feature.

**Purpose:** Store tasks in human-readable markdown files within Git repositories, enabling version control and collaboration workflows.

**How It Works:**

1. **Repository Detection:**
   - Walks up directory tree from current working directory
   - Searches for `.git` directory or file (submodule support)
   - Identifies repository root path

2. **Task File Location:**
   - Searches for markdown file with special marker: `<!-- todoat:enabled -->`
   - Search order (configurable via `file` and `fallback_files`):
     1. Configured file path
     2. Fallback files from config
     3. Defaults: `TODO.md`, `todo.md`, `.todoat.md`
   - File must exist and contain marker to be recognized

3. **Markdown Format:**
   - Task lists represented as `## Heading` sections
   - Tasks as markdown list items with metadata
   - Hierarchical structure via indentation
   - Parser: `backend/git/markdown_parser.go`
   - Writer: `backend/git/markdown_writer.go`

4. **Auto-Detection:**
   - Implements `DetectableBackend` interface
   - `CanDetect()` checks for git repo + marked TODO file
   - Returns detection info: "Git repository at {repo} with task file {file}"

5. **Auto-Commit (Optional):**
   - When `auto_commit: true` in config
   - Commits changes after task modifications
   - Commit messages auto-generated based on operation

**User Journey:**
1. User creates Git repository: `git init myproject && cd myproject`
2. User creates `TODO.md` with marker: `echo "<!-- todoat:enabled -->" > TODO.md`
3. User enables Git backend in config with `auto_detect: true`
4. User runs: `todoat` (no backend flag needed)
5. System auto-detects Git backend and shows tasks from TODO.md
6. User adds task: `todoat "Project Tasks" add "Implement feature"`
7. Task added to TODO.md; if auto-commit enabled, automatically committed

**Prerequisites:**
- Git repository (`.git` directory in current path or ancestors)
- Markdown file with `<!-- todoat:enabled -->` marker
- Read/write permissions to repository and task file

**Outputs/Results:**
- Markdown file updated with task changes
- Git commits (if auto-commit enabled)
- Human-readable task format for manual editing

**Technical Details:**
- Implementation: `backend/git/backend.go`
- Registered as detectable: `backend.RegisterDetectable("git", newGitBackendWrapper)`
- Status values: Uses canonical app statuses (`TODO`, `DONE`, `IN-PROGRESS`, `CANCELLED`)
- File caching: Tracks modification time to avoid unnecessary re-parsing

**Related Features:**
- [CLI Interface](cli-interface.md#auto-detection) - Automatic backend detection
- [Configuration](configuration.md#git-backend) - Git-specific settings

---

### 4. File Backend (Placeholder)

> **Note:** This backend is implemented as a placeholder but not yet wired to the CLI. It cannot be accessed via `--backend=file`.

**Purpose:** Reserved for future file-based storage implementations; currently non-functional.

**How It Works:**
- Implements `TaskManager` interface with stub methods
- All methods return `nil` or "not implemented" errors
- Registered for `file://` URL scheme and `"file"` type
- No actual file I/O operations

**Technical Details:**
- Implementation: `backend/file/backend.go`
- Status: Placeholder only, not production-ready
- Use Git backend instead for file-based storage

---

## Backend Selection

### Selection Priority

**Purpose:** Automatically choose the most appropriate backend when user doesn't specify one explicitly, considering sync state, auto-detection, and configuration.

**How It Works:**

The system follows this priority order to select a backend:

1. **Explicit Flag (`--backend=name`):**
   - User specifies backend on command line: `todoat --backend=nextcloud MyTasks`
   - Highest priority - always overrides other methods
   - Returns error if specified backend not found or disabled

2. **Sync Local Backend (when sync enabled):**
   - If `sync.enabled: true` in config
   - Uses SQLite cache database for the remote backend
   - Example: With Nextcloud configured, uses `caches/nextcloud.db`
   - See [Synchronization](synchronization.md#cache-databases)

3. **Auto-Detection (when enabled):**
   - If `auto_detect_backend: true` in config
   - Iterates through detectable backends (currently: Git)
   - Calls `CanDetect()` on each backend
   - Uses first backend that returns `true`
   - Shows detection info: "Using detected git backend: Git repository at myproject with task file TODO.md"

4. **Default Backend (`default_backend`):**
   - Uses backend specified in `default_backend` config field
   - Example: `default_backend: nextcloud`
   - Falls back if no explicit flag, sync, or detection

5. **Backend Priority List (`backend_priority`):**
   - Uses first enabled backend from `backend_priority` list
   - Example: `backend_priority: [nextcloud, git, sqlite]`
   - Iterates list, returns first enabled backend

6. **First Enabled Backend:**
   - Last resort: uses first enabled backend found in config
   - No guaranteed order (depends on map iteration)

**User Journey - Auto-Detection:**
1. User is in Git repository with TODO.md file
2. User enables auto-detection: `auto_detect_backend: true`
3. User runs: `todoat` (no arguments)
4. System checks for detectable backends
5. Git backend's `CanDetect()` finds repository and marked file
6. System uses Git backend automatically
7. User sees: "Using detected git backend: Git repository at myproject with task file TODO.md"

**User Journey - Explicit Selection:**
1. User has multiple backends configured (Nextcloud, Git)
2. User wants to use Nextcloud specifically
3. User runs: `todoat --backend=nextcloud MyTasks`
4. System bypasses auto-detection and priority
5. Nextcloud backend used for this command

**Prerequisites:**
- At least one enabled backend in configuration
- For auto-detection: Git repository with marked TODO file
- For sync: `sync.enabled: true` and remote backend configured

**Technical Details:**
- Selection logic in: `internal/operations/actions.go`, backend initialization code
- Error handling: Returns descriptive errors for missing/disabled backends
- Detection interface: `backend.DetectableBackend` with `CanDetect()` and `DetectionInfo()` methods

**Related Features:**
- [Configuration](configuration.md#backend-configuration) - Backend setup
- [CLI Interface](cli-interface.md#backend-flag) - `--backend` flag usage

---

### Backend Configuration Formats

**Purpose:** Support multiple configuration styles for backward compatibility and flexibility.

**How It Works:**

**Modern Multi-Backend Format (Recommended):**
```yaml
backends:
  nextcloud:
    type: nextcloud
    enabled: true
    host: "nextcloud.example.com"
    username: "myuser"
    # Password from keyring/environment

  work_nc:
    type: nextcloud
    enabled: true
    host: "work.example.com"
    username: "workuser"

  git:
    type: git
    enabled: true
    auto_detect: true
    file: "TODO.md"
    auto_commit: true

default_backend: nextcloud
auto_detect_backend: true
backend_priority: [work_nc, nextcloud, git]
```

**Legacy URL Format (Deprecated):**
```yaml
connector:
  url: "nextcloud://username:password@host"
  insecure_skip_verify: false
```

**Format Comparison:**
- **Modern:** Supports multiple backends of same type, keyring integration, better isolation
- **Legacy:** Single backend, credentials in config file (insecure), limited to one connection
- **Migration Path:** See [Configuration](configuration.md#migrating-from-legacy-config)

**Technical Details:**
- Modern: `BackendConfig` struct with type-specific fields
- Legacy: `ConnectorConfig` struct with URL parsing
- Both supported for backward compatibility
- Constructors: `BackendConfig.TaskManager()` vs `ConnectorConfig.TaskManager()`

---

## Backend Display Information

### Backend Display Names

**Purpose:** Provide contextual information about which backend and connection is currently in use, helping users track multi-backend workflows.

**How It Works:**
- Each backend implements `GetBackendDisplayName()` returning formatted string
- Displayed in task list headers via `TaskList.StringWithBackend()`
- Format: `[type:context]`

**Backend-Specific Formats:**
- **Nextcloud:** `[nextcloud:username@host:port]`
  - Example: `[nextcloud:admin@localhost:8080]`
- **SQLite (with sync):** `[sqlite → nextcloud]`
  - Shows cache syncing to remote
- **Git:** `[git:repo-name/TODO.md]`
  - Example: `[git:todoat/TODO.md]`
- **File:** `[file:/path/to/file]`

**User Journey:**
1. User runs: `todoat MyTasks`
2. Task list header displays: `┌─ MyTasks ──── [nextcloud:admin@localhost:8080] ┐`
3. User knows exactly which backend and account is active
4. Helpful when switching between multiple Nextcloud accounts or backends

**Outputs/Results:**
- Header line showing list name and backend info
- Right-aligned backend context within terminal width
- Truncation handling for narrow terminals

**Technical Details:**
- Methods: `GetBackendDisplayName()`, `GetBackendType()`, `GetBackendContext()`
- Formatting in: `backend/taskManager.go:688` (`StringWithWidthAndBackend`)
- Terminal width adaptation: 40-100 chars

---

## Data Models

### Task Structure

**Purpose:** Represent individual task items with metadata following iCalendar VTODO specification for cross-platform compatibility.

**How It Works:**
- Struct defined in `backend/taskManager.go:352`
- Fields map to VTODO properties:
  - **UID:** Unique identifier (auto-generated if not provided)
  - **Summary:** Task title/name (required)
  - **Description:** Additional details (optional)
  - **Status:** Backend-specific state (e.g., `NEEDS-ACTION`, `TODO`)
  - **Priority:** 0-9 (0=undefined, 1=highest, 9=lowest)
  - **Created/Modified:** Automatic timestamps
  - **DueDate/StartDate:** Optional deadline and start time
  - **Completed:** Timestamp when marked done
  - **Categories:** Tags/labels for organization
  - **ParentUID:** Links to parent task for hierarchy

**Technical Details:**
- JSON serializable for API/storage
- Status translation via `TaskManager.ParseStatusFlag()` and `StatusToDisplayName()`
- Formatting methods: `FormatWithView()`, `FormatWithIndentLevel()`
- Hierarchical organization: `OrganizeTasksHierarchically()` (`backend/taskManager.go:549`)

**Related Features:**
- [Task Management](task-management.md) - CRUD operations
- [Subtasks & Hierarchy](subtasks-hierarchy.md) - Parent-child relationships

---

### TaskList Structure

**Purpose:** Represent collections/categories of tasks, corresponding to calendars in CalDAV or sections in markdown.

**How It Works:**
- Struct defined in `backend/taskManager.go:609`
- Fields:
  - **ID:** Unique identifier within backend
  - **Name:** Human-readable list name
  - **Description:** Optional context
  - **Color:** Hex color code for UI (e.g., `#0082c9`)
  - **URL:** Backend-specific access URL
  - **CTags:** Synchronization token for change detection
  - **DeletedAt:** Trash timestamp (Nextcloud-specific)

**Technical Details:**
- String formatting: `String()`, `StringWithWidth()`, `StringWithBackend()`
- Border rendering with terminal width adaptation
- Bottom border: `BottomBorder()`, `BottomBorderWithWidth()`

**Related Features:**
- [List Management](list-management.md) - List operations
- [Synchronization](synchronization.md#ctag-tracking) - CTag-based sync

---

### TaskFilter Structure

**Purpose:** Specify filtering criteria for task queries using AND/OR logic combinations.

**How It Works:**
- Struct defined in `backend/taskManager.go:274`
- Optional filter fields (nil means no filtering):
  - **Statuses:** Include tasks with these statuses (OR logic)
  - **ExcludeStatuses:** Exclude tasks with these statuses
  - **DueAfter/DueBefore:** Date range filtering (inclusive)
  - **CreatedAfter/CreatedBefore:** Creation date filtering (inclusive)
- Multiple criteria combined with AND logic

**Example Usage:**
```go
filter := &backend.TaskFilter{
    Statuses: &[]string{"TODO", "IN-PROGRESS"},
    DueBefore: &tomorrow,
}
tasks, err := taskManager.GetTasks(listID, filter)
```

**Technical Details:**
- Backend-specific status values required
- Time filtering uses `*time.Time` for optional dates
- Used throughout CLI for `--status`, `--due-before`, etc. flags

**Related Features:**
- [Task Management](task-management.md#filtering) - Filtering tasks
- [Views & Customization](views-customization.md#filters) - Custom view filters

---

## Error Handling

### Backend Errors

**Purpose:** Provide detailed, structured error information for backend operations to aid debugging and user feedback.

**How It Works:**
- Custom error types per backend:
  - **SQLiteError:** Includes operation, list ID, task UID (`backend/sqlite/backend.go:24`)
  - **NextcloudError:** HTTP status codes, CalDAV errors
  - **UnsupportedSchemeError:** Invalid backend type/scheme (`backend/taskManager.go:12`)
- Error wrapping using Go's `Unwrap()` pattern
- `backend.BackendError` interface for type checking

**Error Categories:**
- **Not Found:** Task or list doesn't exist (`IsNotFound()`)
- **Authentication:** Invalid credentials, expired tokens
- **Network:** Connection failures, timeouts
- **Conflict:** Concurrent modifications, ETag mismatches
- **Validation:** Invalid input, constraint violations

**Technical Details:**
- Implementation: `backend/errors.go`
- Error checking: `errors.Is()`, `errors.As()` for error chains
- HTTP status mapping for Nextcloud: 404→NotFound, 401→Auth, 412→Conflict

---

## Extension Points

### Adding a New Backend

**Purpose:** Guide for developers to add support for new storage providers (e.g., Todoist, GitHub Issues, Trello).

**How It Works:**

**Step-by-Step Process:**

1. **Create Backend Package:**
   ```
   backend/mybackend/
   ├── backend.go         # Main implementation
   ├── backend_test.go    # Unit tests
   └── parser.go          # Format-specific parsing (if needed)
   ```

2. **Implement TaskManager Interface:**
   ```go
   type MyBackend struct {
       config backend.BackendConfig
       client *http.Client  // Or other connection
   }

   func (mb *MyBackend) GetTaskLists() ([]backend.TaskList, error) {
       // Implementation
   }
   // ... implement all 20+ TaskManager methods
   ```

3. **Register Backend:**
   ```go
   func init() {
       backend.RegisterType("mybackend", NewMyBackend)
       // Optional: for auto-detection
       backend.RegisterDetectable("mybackend", NewMyBackend)
   }
   ```

4. **Add Configuration Support:**
   - Update `BackendConfig` type validation: `oneof=nextcloud git file sqlite mybackend`
   - Add backend-specific config fields if needed
   - Document configuration in CLAUDE.md

5. **Implement Status Translation:**
   ```go
   func (mb *MyBackend) ParseStatusFlag(statusFlag string) (string, error) {
       // Map user input to backend's status format
   }

   func (mb *MyBackend) StatusToDisplayName(backendStatus string) string {
       // Map backend status to canonical names: TODO, DONE, IN-PROGRESS, CANCELLED
   }
   ```

6. **Add Tests:**
   - Unit tests for all CRUD operations
   - Integration tests if backend requires external service
   - Mock/test server for CI/CD

**Prerequisites:**
- Understanding of TaskManager interface contract
- Access to target service API documentation
- Test environment for new backend

**Technical Details:**
- Interface definition: `backend/taskManager.go:155`
- Registry: `backend/registry.go`
- Example implementations: `backend/nextcloud/`, `backend/git/`

**Related Features:**
- [Configuration](configuration.md#adding-backend-config) - Config schema updates

---

### Auto-Detection Interface

**Purpose:** Enable backends to automatically activate when their environment is detected, reducing configuration burden.

**How It Works:**

**DetectableBackend Interface:**
```go
type DetectableBackend interface {
    TaskManager

    // Check if backend can be used in current environment
    CanDetect() (bool, error)

    // Return human-readable detection info
    DetectionInfo() string
}
```

**Implementation Example (Git):**
```go
func (gb *GitBackend) CanDetect() (bool, error) {
    // Check for .git directory
    repoPath, err := gb.findGitRepo()
    if err != nil {
        return false, nil
    }

    // Check for marked TODO file
    filePath, err := gb.findTodoFile()
    if err != nil {
        return false, nil
    }

    return true, nil
}

func (gb *GitBackend) DetectionInfo() string {
    return fmt.Sprintf("Git repository at %s with task file %s",
        filepath.Base(gb.RepoPath), filepath.Base(gb.FilePath))
}
```

**Registration:**
```go
backend.RegisterDetectable("git", newGitBackendWrapper)
```

**Detection Flow:**
1. User enables: `auto_detect_backend: true`
2. CLI calls `GetDetectableConstructors()` from registry
3. For each detectable backend, instantiate and call `CanDetect()`
4. Use first backend returning `(true, nil)`
5. Show detection info to user

**Technical Details:**
- Registry method: `GetDetectableConstructors()` (`backend/registry.go:76`)
- Detection should be fast (<100ms) and non-destructive
- Return `false` for unsupported environments, not errors

---

## Status Translation

### Status Mapping

**Purpose:** Translate between user-friendly status names, app-internal statuses, and backend-specific formats for seamless cross-backend compatibility.

**How It Works:**

**Three Status Formats:**

1. **User Input (Abbreviations):**
   - `T` → TODO
   - `D` → DONE
   - `I` → IN-PROGRESS
   - `C` → CANCELLED

2. **App Canonical Names:**
   - `TODO` - Task not started
   - `DONE` - Task completed
   - `IN-PROGRESS` - Task in progress
   - `CANCELLED` - Task abandoned

3. **Backend-Specific:**
   - **CalDAV (Nextcloud):** `NEEDS-ACTION`, `COMPLETED`, `IN-PROCESS`, `CANCELLED`
   - **SQLite/File/Git:** Use canonical app names directly

**Translation Methods:**

**Input Translation (User → Backend):**
```go
backendStatus, err := taskManager.ParseStatusFlag("T")
// Returns: "NEEDS-ACTION" for Nextcloud, "TODO" for SQLite
```

**Output Translation (Backend → Display):**
```go
displayName := taskManager.StatusToDisplayName("NEEDS-ACTION")
// Returns: "TODO" for all backends
```

**Per-Backend Mapping (Nextcloud):**
```go
var statusToCalDAV = map[string]string{
    "T":            "NEEDS-ACTION",
    "TODO":         "NEEDS-ACTION",
    "D":            "COMPLETED",
    "DONE":         "COMPLETED",
    "I":            "IN-PROCESS",
    "IN-PROGRESS":  "IN-PROCESS",
    "C":            "CANCELLED",
    "CANCELLED":    "CANCELLED",
    // CalDAV names also accepted
    "NEEDS-ACTION": "NEEDS-ACTION",
    "COMPLETED":    "COMPLETED",
    "IN-PROCESS":   "IN-PROCESS",
}
```

**User Journey:**
1. User runs: `todoat MyTasks update "task" -s D`
2. CLI parses abbreviation: `D`
3. Backend translates: `D` → `COMPLETED` (Nextcloud) or `DONE` (SQLite)
4. Backend stores task with backend-specific status
5. On display, backend translates back: `COMPLETED` → `DONE`
6. User sees: `✓ DONE task`

**Technical Details:**
- Interface methods: `ParseStatusFlag()`, `StatusToDisplayName()` (`backend/taskManager.go:220-228`)
- Nextcloud mapping: `backend/nextcloud/backend.go:39`
- Legacy functions (deprecated): `StatusStringTranslateToStandardStatus()`, `StatusStringTranslateToAppStatus()`

---

## Performance Considerations

### Connection Pooling

**Purpose:** Optimize Nextcloud backend performance by reusing HTTP connections and reducing TLS handshake overhead.

**How It Works:**
- HTTP client configured with `http.Transport`
- Settings (`backend/nextcloud/backend.go:61`):
  - **MaxIdleConns:** 10 - Total idle connections across all hosts
  - **MaxIdleConnsPerHost:** 2 - Idle connections per Nextcloud server
  - **IdleConnTimeout:** 30s - Keep connections alive for reuse
  - **Timeout:** 30s - Overall request timeout

**Technical Details:**
- Lazy initialization: Client created on first use
- TLS config includes `InsecureSkipVerify` support for self-signed certs
- Connection reuse reduces latency by 50-200ms per request

---

### Caching Strategies

**Purpose:** Minimize redundant backend queries and improve responsiveness through intelligent caching.

**How It Works:**

**Git Backend File Caching:**
- Tracks markdown file modification time (`fileModTime`)
- Skips re-parsing if file unchanged since last load
- Parser caching in `loadFile()` method

**SQLite Indexes:**
- Optimized indexes for common queries (`backend/sqlite/schema.go`):
  - `idx_tasks_list_status`: Fast status filtering
  - `idx_tasks_due_date`: Quick due date queries
  - `idx_tasks_parent`: Efficient hierarchy traversal
  - `idx_sync_metadata_flags`: Sync operation queries
  - `idx_sync_queue_backend`: Queue processing per backend

**Task List Cache:**
- CLI maintains list cache in db (sqlite) 
- Speeds up list name auto-completion
- Invalidated on list modifications

**Technical Details:**
- Cache invalidation: File mtime checks, explicit invalidation on writes
- No in-memory caching of task data to avoid stale data issues
- Sync metadata caching in SQLite for offline support

**Related Features:**
- [Synchronization](synchronization.md#performance) - Sync-specific caching

---

## Security Considerations

### Credential Handling

**Purpose:** Protect user credentials through secure storage mechanisms and avoid plain-text exposure.

**How It Works:**
- Priority order: Keyring > Environment Variables > Config URL
- Never log or display passwords
- See [Credential Management](credential-management.md) for detailed security model

**Technical Details:**
- Credential resolution: `backend/nextcloud/backend.go:76` (`getUsername()`, `getPassword()`)
- No credential caching in memory beyond HTTP client lifecycle
- URL-based credentials deprecated due to security risks

---

### TLS/SSL Configuration

**Purpose:** Enforce secure HTTPS connections by default while supporting development and self-signed certificate scenarios.

**How It Works:**

**Default Behavior:**
- HTTPS enforced for all Nextcloud connections
- TLS certificate validation enabled
- Rejects invalid/expired certificates

**Development Options:**
```yaml
backends:
  nextcloud:
    insecure_skip_verify: true      # Allow self-signed certs (shows warning)
    suppress_ssl_warning: true      # Suppress warning message
    allow_http: true                # Allow HTTP connections (shows warning)
    suppress_http_warning: true     # Suppress HTTP warning
```

**Security Warnings:**
- Insecure options trigger console warnings by default
- Warnings include security implications
- Suppression flags for CI/CD environments

**Technical Details:**
- TLS config: `backend/nextcloud/backend.go:65` (`TLSClientConfig`)
- HTTP detection: Port-based (80, 8080, 8000) when `allow_http: true`
- HTTPS upgrade: Automatic unless HTTP explicitly allowed

---

## Cross-References

**Prerequisite Features:**
- [Configuration](configuration.md) - Backend configuration setup
- [Credential Management](credential-management.md) - Secure credential storage

**Related Features:**
- [Task Management](task-management.md) - Using backends for task operations
- [List Management](list-management.md) - List operations across backends
- [Synchronization](synchronization.md) - Multi-backend sync architecture
- [Subtasks & Hierarchy](subtasks-hierarchy.md) - Parent-child relationships
- [Views & Customization](views-customization.md) - Backend-aware formatting

**Integration Points:**
- [CLI Interface](cli-interface.md) - Backend selection flags and auto-detection
- [Configuration](configuration.md) - Backend-specific settings

---

**Navigation:**
- [← Back to Overview](README.md)
- [← Back to Features Overview](features-overview.md)
- [Next: Synchronization →](synchronization.md)
