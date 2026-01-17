# List Management

This document details all features related to managing task lists in todoat. Lists are containers that organize tasks, similar to calendars in CalDAV or projects in task management systems.

## Table of Contents

- [Overview](#overview)
- [Create List](#create-list)
- [View All Lists](#view-all-lists)
- [Interactive List Selection](#interactive-list-selection)
- [List Caching](#list-caching)
- [Trash and Restore Lists](#trash-and-restore-lists)
- [List Properties](#list-properties)
- [Backend-Specific List Features](#backend-specific-list-features)

---

## Overview

Lists in todoat serve as organizational containers for tasks. Each backend (Nextcloud CalDAV, SQLite) manages lists differently:
- **Nextcloud**: Lists are CalDAV calendars containing VTODO items
- **SQLite**: Lists are database tables with foreign key relationships to tasks

All list operations respect the current [backend selection](./BACKEND_SYSTEM.md#backend-selection-logic) and work seamlessly across different storage systems.

---

## Create List

### Feature Name
**List Creation**

### Purpose
Allows users to create new task lists to organize tasks by project, category, or any custom grouping scheme.

### How It Works

**User Actions:**
1. Execute command: `todoat list create "List Name"`
2. System validates the name (non-empty, unique within backend)
3. Backend creates the list structure
4. Confirmation message displays with list ID

**System Processes:**
1. Command parser extracts list name from arguments
2. Backend validation checks for duplicates
3. For Nextcloud:
   - Creates new CalDAV calendar via MKCALENDAR request
   - Sets displayname, color, and supported-component properties
4. For SQLite:
   - Inserts new record into `task_lists` table
   - Generates unique ID
   - Initializes sync metadata if [sync enabled](./SYNCHRONIZATION.md#configuration)
5. Updates local [list cache](#list-caching)

**Data Flow:**
```
CLI Input → Cobra Command → Backend.CreateList()
          ↓
Nextcloud: HTTP MKCALENDAR → Update cache
SQLite: INSERT INTO task_lists → Update sync_metadata
```

**Edge Cases:**
- **Duplicate name**: Error returned, suggests using unique name
- **Permission denied** (Nextcloud): CalDAV 403 error displayed
- **Offline mode**: SQLite creates locally, queued for sync when online

### User Journey
```
User wants to organize work tasks separately from personal tasks
↓
$ todoat list create "Work Tasks"
↓
System: "Created list: Work Tasks (ID: abc-123)"
↓
User can now add tasks: todoat "Work Tasks" add "Meeting"
```

### Prerequisites
- Active [backend connection](./BACKEND_SYSTEM.md#backend-selection-logic)
- For Nextcloud: Valid [credentials](./CREDENTIAL_MANAGEMENT.md)
- For SQLite with sync: Initialized [sync system](./SYNCHRONIZATION.md#architecture)

### Outputs/Results
- New list created in backend storage
- List appears in `todoat list` output
- List available for [task operations](./TASK_MANAGEMENT.md)
- If sync enabled: Create operation queued in `sync_queue` table

### Technical Details
**Nextcloud Implementation** (`backend/nextcloudBackend.go`):
```go
func (nc *NextcloudBackend) CreateList(name string) (*TaskList, error)
```
- Uses MKCALENDAR WebDAV method
- Sets properties: displayname, calendar-color, supported-calendar-component-set
- Generates URL-safe list ID from name

**SQLite Implementation** (`backend/sqliteBackend.go`):
```go
func (s *SQLiteBackend) CreateList(name string) (*TaskList, error)
```
- Transaction-based creation with rollback on error
- Creates corresponding `list_sync_metadata` record
- separate synced backends
- use id for internal operation, let remote backend generate uid

### Related Features
- [View All Lists](#view-all-lists) - See all available lists
- [Task Creation](./TASK_MANAGEMENT.md#add-task) - Add tasks to the new list
- [Synchronization](./SYNCHRONIZATION.md#push-operations) - Sync new lists to remote

---

## View All Lists

### Feature Name
**List Directory/Enumeration**

### Purpose
Displays all available task lists across enabled backends, allowing users to see what organizational containers exist and their basic properties.

### How It Works

**User Actions:**
1. Execute: `todoat list` (no arguments)
2. View formatted table of all lists

**System Processes:**
1. Queries all enabled backends via `GetTaskLists()`
2. For Nextcloud:
   - PROPFIND request to calendar-home-set
   - Parses XML response for calendar properties
3. For SQLite:
   - SELECT query on `task_lists` table
   - Joins with task counts
4. Aggregates results from multiple backends
5. Formats output table with columns:
   - List Name
   - Backend (nextcloud/sqlite)
   - Task Count
   - Color (if supported)
   - Last Modified (if available)

**Data Flow:**
```
User Command → GetTaskLists() per backend → Aggregate results
            ↓
Format as table → Display to stdout
```

**Integration Points:**
- Uses [backend system](./BACKEND_SYSTEM.md#backend-system) to query all enabled backends
- Respects [backend priority](./BACKEND_SYSTEM.md#backend-selection-logic) for ordering

**Edge Cases:**
- **No lists found**: Displays helpful message suggesting `list create`
- **Backend offline**: Shows error for that backend, continues with others
- **Empty lists**: Shows count as "0 tasks"

### User Journey
```
User wants to see all available task lists
↓
$ todoat list
↓
Output:
┌──────────────┬──────────┬────────┬───────┐
│ Name         │ Backend  │ Tasks  │ Color │
├──────────────┼──────────┼────────┼───────┤
│ Work Tasks   │ sqlite   │ 12     │ #blue │
│ Personal     │ nextcloud│ 5      │ #red  │
│ Shopping     │ sqlite   │ 0      │ #green│
└──────────────┴──────────┴────────┴───────┘
↓
User selects a list to work with
```

### Prerequisites
- At least one [backend enabled](./CONFIGURATION.md#backend-configuration)
- For remote backends: Network connectivity and [valid credentials](./CREDENTIAL_MANAGEMENT.md)

### Outputs/Results
- Formatted table of lists with metadata
- Exit code 0 if successful, non-zero on error
- Lists cached locally for [interactive selection](#interactive-list-selection)

### Technical Details
**Backend Interface** (`backend/taskManager.go`):
```go
type TaskManager interface {
    GetTaskLists() ([]TaskList, error)
}
```

**TaskList Structure**:
```go
type TaskList struct {
    ID          string    // Unique identifier
    Name        string    // Display name
    Description string    // Optional description
    CTags       string    // CalDAV sync token
    Color       string    // Hex color code
}
```

**Caching**: Results cached in `$XDG_CACHE_HOME/todoat/lists.json` (5-minute TTL)

### Related Features
- [Interactive List Selection](#interactive-list-selection) - Pick from this list
- [Create List](#create-list) - Add new lists
- [List Caching](#list-caching) - Performance optimization

---

## Interactive List Selection

### Feature Name
**Dynamic List Picker**

### Purpose
When a command is executed without specifying a list name, provides an interactive menu to select from available lists, improving usability and reducing memorization burden.

### How It Works

**User Actions:**
1. Execute command without list name: `todoat`
2. View numbered list of available lists
3. Enter number corresponding to desired list
4. System loads tasks from selected list

**System Processes:**
1. Detects missing list argument in command
2. Checks [list cache](#list-caching) for recent list data
3. If cache stale or missing:
   - Calls `GetTaskLists()` on all enabled backends
   - Updates cache with results
4. **Normal Mode**: Displays numbered menu with dynamic terminal width formatting
5. **No-Prompt Mode** (`-y`): Displays list table and returns `INFO_ONLY` (no selection prompt)
6. Reads user input from stdin (skipped in no-prompt mode)
7. Validates selection (must be valid number in range)
8. Executes original command with selected list

**Terminal Width Detection:**
- Uses `golang.org/x/term` for cross-platform terminal size detection
- Default: 80 characters if detection fails
- Constraints: 40-100 character range
- Adjusts border drawing to terminal width

**Data Flow:**
```
Command without list → Check cache → No-Prompt Mode?
                                     ↓ Yes
                              Display list table → INFO_ONLY
                                     ↓ No
                              Display menu → User selects → Execute command
```

**Edge Cases:**
- **Single list available**: Displays the list (no auto-select in no-prompt mode)
- **No lists available**: Error message, suggests `list create`
- **Invalid input**: Re-prompts user, allows retry
- **User cancels** (Ctrl+C): Graceful exit with cleanup
- **No-prompt mode**: Always displays available lists with `INFO_ONLY` result

### User Journey

**Normal Mode:**
```
User wants to add a task but forgets list name
↓
$ todoat add "Buy milk"
↓
System detects missing list, shows menu:
Select a task list:
1. Work Tasks (12 tasks)
2. Personal (5 tasks)
3. Shopping (0 tasks)

Enter number:
↓
User types: 2
↓
Task added to "Personal" list
```

**No-Prompt Mode:**
```
Script needs to check available lists
↓
$ todoat -y
↓
Available lists:
ID:abc-123	NAME:Work Tasks	TASKS:12
ID:def-456	NAME:Personal	TASKS:5
ID:ghi-789	NAME:Shopping	TASKS:0
INFO_ONLY
↓
Script parses output and specifies list explicitly
↓
$ todoat -y "Work Tasks" add "Buy milk"
ACTION_COMPLETED
```

**No-Prompt Mode with JSON:**
```bash
$ todoat -y --json
```
```json
{
  "lists": [
    {"id": "abc-123", "name": "Work Tasks", "task_count": 12, "color": "#0066cc"},
    {"id": "def-456", "name": "Personal", "task_count": 5, "color": "#ff5733"},
    {"id": "ghi-789", "name": "Shopping", "task_count": 0, "color": "#00cc66"}
  ],
  "result": "INFO_ONLY"
}
```

### Prerequisites
- At least one list exists in any enabled backend
- **Normal Mode**: Terminal supports stdin reading (not in pipe/redirect mode)
- **No-Prompt Mode**: No stdin requirement

### Outputs/Results
- **Normal Mode**: Interactive menu displayed to stdout, selected list used for command
- **No-Prompt Mode**: List table with IDs and names, `INFO_ONLY` result code
- Cache updated if refresh occurred


**Terminal Width Calculation**:
```go
width, _, err := term.GetSize(int(os.Stdout.Fd()))
if err != nil {
    width = 80 // Default fallback
}
width = clamp(width, 40, 100)
```

### Related Features
- [List Caching](#list-caching) - Powers fast menu display
- [View All Lists](#view-all-lists) - Data source for menu
- [CLI Interface](./CLI_INTERFACE.md) - Argument parsing
- [CLI Interface - No-Prompt Mode](./CLI_INTERFACE.md#no-prompt-mode) - Non-interactive operation
- [CLI Interface - JSON Output](./CLI_INTERFACE.md#json-output-mode) - Machine-parseable output

---

## List Caching

### Feature Name
**List Metadata Cache**

### Purpose
Improves performance by caching task list metadata locally, reducing network requests and database queries for frequently accessed list information.

### How It Works

**User Actions:**
- Transparent to user - automatic background operation
- Any command that needs list information triggers cache check

**System Processes:**
1. On first list query:
   - Fetches lists from all enabled backends
   - Records timestamp in cache metadata
2. On subsequent queries:
   - Reads configured cache db (sqlite)
   - Checks age (TTL = 5 minutes by default)
   - If fresh: Returns cached data
   - If stale: Refreshes from backends
3. Cache invalidation on:
   - List create/delete operations
   - Explicit `sync` command
   - Manual cache clear (delete file)

**Edge Cases:**
- **Corrupt db**: Deleted and regenerated
- **Missing db**: Created with proper permissions

### User Journey
When sync and autosync is enabled:

```
First command of the day:
$ todoat "Work Tasks"
→ Cache miss, sync with backend (300ms)

Second command 1 minute later:
$ todoat
→ Cache hit, instant list display (5ms)

After 6 minutes:
$ todoat
→ Cache stale, auto-refreshes (300ms)
```

### Prerequisites
- Write permissions to XDG cache directory
- At least one backend enabled

### Outputs/Results
- Faster list operations (5-10x speedup for remote backends)
- Reduced network traffic for Nextcloud backend
- Improved responsiveness of [interactive selection](#interactive-list-selection)

### Technical Details
**Sync Cache TTL**: 4 hours (configurable via `config.yaml` )
Force sync before doing operation when sync did not happen since configured time (4 hours default). If autosync disabled, tell user to sync (no prompt).


### Related Features
- [Interactive List Selection](#interactive-list-selection) - Primary consumer
- [View All Lists](#view-all-lists) - Cache data source
- [Configuration](./CONFIGURATION.md#xdg-base-directory-support) - Cache location

---

## Trash and Restore Lists

### Feature Name
**List Soft Delete and Recovery**

### Purpose
Provides safety net for accidentally deleted lists by moving them to a trash/archive state instead of permanent deletion, allowing recovery of lists and their tasks.

### How It Works

**Deleting a List:**

**User Actions:**
1. Execute: `todoat list delete "List Name"`
2. **Normal Mode**: System prompts for confirmation (unless `--force` flag used)
3. **No-Prompt Mode** (`-y`): Deletes immediately without confirmation (acts as force mode)
4. Confirm deletion (if prompted)
5. List moved to trash state

**System Processes:**
1. Finds list by name across backends
2. **Check No-Prompt Mode**: If `-y` flag or `no_prompt: true` in config, skip confirmation
3. For Nextcloud:
   - Marks calendar as hidden (not supported by all servers)
   - Or moves to special "trash" calendar collection
4. For SQLite:
   - Sets `deleted_at` timestamp in `task_lists` table
   - Tasks remain in database but hidden from normal queries
5. List removed from active list cache
6. Confirmation message displays (or `ACTION_COMPLETED` in no-prompt mode)

**Viewing Trashed Lists:**

**User Actions:**
1. Execute: `todoat list trash`
2. View all deleted lists with deletion timestamps

**Restoring a List:**

**User Actions:**
1. Execute: `todoat list trash restore "List Name"`
2. List and all tasks restored to active state

**System Processes:**
1. Finds list in trash by name
2. For Nextcloud:
   - Unhides calendar
   - Or moves back to active calendar collection
3. For SQLite:
   - Clears `deleted_at` timestamp
   - All tasks become visible again
4. Updates list cache
5. If sync enabled: Restore operation queued

**Data Flow:**
```
Delete: List → Set deleted_at → Hide from queries
Restore: Find in trash → Clear deleted_at → Show in queries
```

**Edge Cases:**
- **List not found**: Error with suggestion to check `list trash`
- **Name collision on restore**: Appends "(Restored)" to name
- **Permanent delete**: `--permanent` flag bypasses trash (requires explicit confirmation even in no-prompt mode)
- **Trash cleanup**: Auto-purge after 30 days (configurable)
- **No-prompt mode**: Skips confirmation, outputs `ACTION_COMPLETED` on success

### User Journey

**Normal Mode:**
```
User accidentally deletes important list
↓
$ todoat list delete "Work Tasks"
Confirm deletion of "Work Tasks"? [y/N]: y
↓
List moved to trash
↓
User realizes mistake
↓
$ todoat list trash
Deleted Lists:
- Work Tasks (deleted 2 minutes ago)
↓
$ todoat list trash restore "Work Tasks"
↓
List and all 12 tasks restored successfully
```

**No-Prompt Mode (Scripting):**
```bash
# Delete without confirmation
$ todoat -y list delete "Temp List"
List 'Temp List' moved to trash
ACTION_COMPLETED

# With JSON output
$ todoat -y --json list delete "Another List"
```
```json
{
  "action": "delete",
  "list": {"id": "abc-123", "name": "Another List"},
  "result": "ACTION_COMPLETED"
}
```

### Prerequisites
- List must exist in active or trash state
- For restore: No active list with same name (or auto-rename)
- **No-Prompt Mode**: Be cautious - deletes without confirmation

### Outputs/Results
- **Delete**: List hidden from normal operations, tasks preserved
- **Delete (No-Prompt)**: `ACTION_COMPLETED` result code
- **View trash**: Table of deleted lists with timestamps
- **Restore**: List and tasks fully recovered
- If sync enabled: Operations propagated to remote backend


**Nextcloud Implementation**:
- Uses WebDAV MOVE operation to special collection
- Or sets custom property: `<deleted>true</deleted>`

**Auto-Purge**: Background job removes lists after TTL

### Related Features
- [Task Deletion](./TASK_MANAGEMENT.md#delete-task) - Similar soft-delete pattern
- [Synchronization](./SYNCHRONIZATION.md#delete-operations) - Propagates trash state
- [Configuration](./CONFIGURATION.md#trash-settings) - TTL and auto-purge settings
- [CLI Interface - No-Prompt Mode](./CLI_INTERFACE.md#no-prompt-mode) - Skips confirmation prompts

---

## List Properties

### Feature Name
**List Metadata and Attributes**

### Purpose
Stores and manages additional information about lists beyond just names, including colors, descriptions, sync tokens, and backend-specific properties.

### How It Works

**Supported Properties:**

1. **Name** (required)
   - Display name shown in UI
   - Used for list identification
   - Must be unique within backend

2. **Description** (optional)
   - Long-form text describing list purpose
   - Displayed in detailed views
   - Set via: `todoat list update "Name" --description "Text"`

3. **Color** (optional)
   - Hex color code (e.g., `#FF5733`)
   - Used for visual differentiation in UI
   - Set via: `todoat list update "Name" --color "#0066cc"`
   - For nextcloud, synced with CalDAV calendar-color property

4. **CTags** (automatic)
   - CalDAV sync token for change detection
   - Managed by backend, read-only to users
   - Used in [synchronization](./SYNCHRONIZATION.md#sync-tokens)

5. **ID** (automatic)
   - Unique identifier within backend
   - UUID format for SQLite
   - URL path for Nextcloud
   - Immutable after creation

**User Actions:**
```bash
# View properties
todoat list show "Work Tasks"

# Update color
todoat list update "Work Tasks" --color "#FF5733"

# Update description
todoat list update "Work Tasks" --description "All work-related tasks"
```

**System Processes:**
1. Parse property updates from CLI flags
2. Validate property values:
   - Color: Must be valid hex code
   - Description: Max length (backend-dependent)
3. Update backend storage
4. For Nextcloud: PROPPATCH request with XML properties
5. For SQLite: UPDATE query on task_lists table
6. Invalidate list cache
7. If sync enabled: Queue update operation

**Data Structure** (`backend/taskManager.go`):
```go
type TaskList struct {
    ID          string
    Name        string
    Description string
    CTags       string  // Sync token
    Color       string  // Hex color
}
```

### User Journey
```
User creates a list and wants to color-code it
↓
$ todoat list create "Urgent"
↓
$ todoat list update "Urgent" --color "#FF0000"
↓
List now displays with red color indicator
↓
User adds description for context
↓
$ todoat list update "Urgent" --description "High-priority tasks requiring immediate attention"
```

### Prerequisites
- List must exist
- Valid property values according to backend constraints

### Outputs/Results
- Updated list properties stored in backend
- Changes reflected in [list views](#view-all-lists)
- If sync enabled: Properties synced to remote

### Technical Details
**Nextcloud Property Mapping**:
- Name → `calendar:displayname`
- Color → `calendar:calendar-color`
- Description → `calendar:calendar-description`
- CTags → `calendar:ctag`

**SQLite Schema**:
```sql
CREATE TABLE task_lists (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    color TEXT,
    ctag TEXT,
    deleted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Related Features
- [View All Lists](#view-all-lists) - Displays properties
- [Synchronization](./SYNCHRONIZATION.md#metadata-sync) - Syncs properties
- [Configuration](./CONFIGURATION.md#list-defaults) - Default colors

---

## Backend-Specific List Features

### Feature Name
**Backend-Dependent List Capabilities**

### Purpose
Different backends provide unique capabilities for list management. This section documents features available only in specific backends.

These are advanced features that should be planned at very end. Also, it should be planned better since there should be more uniformity for the different commands between backends.

### Nextcloud-Specific Features (Advanced, plan later)

**1. Calendar Sharing**
- **Purpose**: Share task lists with other Nextcloud users
- **Command**: `todoat list share "List Name" --user "username" --permission read`
- **Permissions**: read, write, admin
- **How It Works**:
  - Uses CalDAV sharing protocol (RFC 5397)
  - Sends SHARE-RESOURCE request to server
  - Other user sees list in their calendar collection
  - Changes sync bidirectionally

**2. Calendar Subscriptions**
- **Purpose**: Subscribe to read-only task lists via URL
- **Command**: `todoat list subscribe "https://example.com/calendar/ical"`
- **How It Works**:
  - Adds calendar to Nextcloud via MKCALENDAR
  - Sets source URL property
  - Periodic refresh from source

**3. Public Links**
- **Purpose**: Generate public URLs for list viewing
- **Command**: `todoat list publish "List Name"`
- **How It Works**:
  - Creates public share token via OCS Share API
  - Returns URL: `https://nextcloud.example.com/s/abc123`
  - Accessible without authentication

### SQLite-Specific Features (Advanced, plan later)

**1. Database Statistics**
- **Purpose**: View internal database metrics
- **Command**: `todoat list stats "List Name"`
- **Output**:
  - Total tasks
  - Tasks by status
  - Disk space used
  - Index efficiency

**2. Vacuum/Optimize**
- **Purpose**: Reclaim space from deleted lists/tasks
- **Command**: `todoat list vacuum`
- **How It Works**:
  - Runs SQLite VACUUM command
  - Rebuilds database file
  - Reclaims deleted space

**3. Export/Import**
- **Purpose**: Backup list as standalone file
- **Command**: `todoat list export "List Name" --format sqlite`
- **Format Support**: sqlite, json, csv
- **How It Works**:
  - Creates standalone database/file with list data
  - Includes all tasks and metadata
  - Can be imported to another instance

### Comparison Matrix

| Feature | Nextcloud | SQLite |
|---------|-----------|---------|
| Sharing | ✅ Native | ❌ Not supported |
| Public Links | ✅ Via OCS API | ❌ Not supported |
| Offline Access | ⚠️ With sync | ✅ Native |
| Export | ⚠️ Via iCalendar | ✅ Multiple formats |
| Subscriptions | ✅ CalDAV | ❌ Not supported |
| Statistics | ⚠️ Limited | ✅ Detailed |

### User Journey Example (Nextcloud Sharing)
```
User wants to collaborate on project tasks
↓
$ todoat list share "Project Alpha" --user "colleague@example.com" --permission write
↓
System: Shared "Project Alpha" with colleague@example.com
↓
Colleague sees list in their todoat instance
↓
Both users can add/edit tasks
↓
Changes sync in real-time via CalDAV
```

### Prerequisites
- **Sharing**: Nextcloud server with sharing enabled
- **Export**: Write permissions to output directory
- **Vacuum**: No active connections to database

### Technical Details
**Nextcloud Sharing** (`backend/nextcloudBackend.go`):
```go
func (nc *NextcloudBackend) ShareList(listID, user string, perm Permission) error {
    // Uses OCS Share API v2
    // POST /ocs/v2.php/apps/dav/api/v1/shares
}
```

**SQLite Export** (`backend/sqliteBackend.go`):
```go
func (s *SQLiteBackend) ExportList(listID, format string) ([]byte, error) {
    // Supports: sqlite, json, csv, icalendar
}
```

### Related Features
- [Backend System](./BACKEND_SYSTEM.md) - Backend architecture
- [Synchronization](./SYNCHRONIZATION.md) - Cross-backend sync
- [Task Management](./TASK_MANAGEMENT.md) - Operating on shared lists

---

## Summary

List management in todoat provides:
- ✅ Cross-backend list creation and organization
- ✅ Interactive list selection for improved UX
- ✅ Performance-optimized caching
- ✅ Safety features (trash/restore)
- ✅ Rich metadata (colors, descriptions)
- ✅ Backend-specific advanced features

All list operations integrate seamlessly with the [backend system](./BACKEND_SYSTEM.md) and [synchronization](./SYNCHRONIZATION.md) to provide a unified experience across SQLite and Nextcloud backends.
