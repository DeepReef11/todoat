# Task Management

Task Management encompasses all core CRUD (Create, Read, Update, Delete) operations for individual tasks within todoat. These operations form the foundation of daily task management workflows.

---

## Table of Contents

1. [Get Tasks (List/Read)](#get-tasks-listread)
2. [Add Tasks (Create)](#add-tasks-create)
3. [Update Tasks](#update-tasks)
4. [Complete Tasks](#complete-tasks)
5. [Delete Tasks](#delete-tasks)
6. [Task Filtering](#task-filtering)
7. [Task Search and Selection](#task-search-and-selection)
8. [Task Status System](#task-status-system)
9. [Task Priority System](#task-priority-system)
10. [Task Dates](#task-dates)
11. [Task Metadata](#task-metadata)

---

## Get Tasks (List/Read)

### Feature Name
**Display Tasks from a Task List**

### Purpose
Allows users to view all tasks within a specific task list, providing a quick overview of pending work, completed items, and task status. This is the default and most frequently used operation in todoat.

### How It Works

**User Actions:**
1. User executes one of the following commands:
   - `todoat` - Interactive list selection via terminal UI
   - `todoat MyList` - Direct list name specification
   - `todoat MyList get` - Explicit get action (optional)
   - `todoat MyList g` - Using abbreviation

**System Processes:**
1. **Backend Selection**: System identifies which backend to use based on:
   - Explicit `--backend` flag (highest priority)
   - Sync local backend if sync is enabled
   - Auto-detected backend (if `auto_detect_backend: true`)
   - Default backend from config
   - First enabled backend

2. **List Resolution**:
   - If no list name provided, display interactive list selector
   - If list name provided, search for exact match
   - If no exact match, search for partial matches
   - If multiple matches, present selection menu

3. **Task Retrieval**:
   - Backend executes `GetTasks(listID)` method
   - Tasks are retrieved from storage (CalDAV server, SQLite database, Git file, etc.)
   - Task data includes: UID, Summary, Description, Status, Priority, Dates, Parent relationships

4. **Rendering**:
   - Tasks are formatted according to the active view (see [Views & Customization](views-customization.md))
   - Default view shows: Status, Summary, Priority
   - Hierarchical tasks are displayed with tree-drawing characters (├─, └─, │)
   - Terminal width is auto-detected for dynamic formatting (40-100 chars, default 80)

**Data Flow:**
```
User Command → Backend Selection → List Resolution → GetTasks() →
View Formatter → Terminal Display
```

**Integration Points:**
- [Backend System](backend-system.md) - Determines which storage backend to query
- [Views & Customization](views-customization.md) - Controls display formatting
- [Subtasks & Hierarchy](subtasks-hierarchy.md) - Tree-based task display

**Edge Cases:**
- Empty list: Displays message "No tasks found in list 'MyList'"
- List not found: Error message with suggestion to use `todoat list` to see available lists
- No lists exist: Prompts user to create first list with `todoat list create "Name"`
- Network failure (remote backend): Falls back to cached data if sync is enabled

### User Journey
1. User opens terminal
2. Types `todoat MyList`
3. System displays formatted task list with status indicators, priorities, and hierarchical relationships
4. User reviews tasks and decides next actions

### Prerequisites
- At least one task list must exist in the backend
- Backend must be configured and accessible (see [Configuration](configuration.md))
- For remote backends: Network connectivity (or cached data available)

### Outputs/Results
**Terminal Output Example:**
```
Tasks in "Work" (5 tasks):

├─ [TODO] Project Alpha (Priority: 1)
│  ├─ [IN-PROGRESS] Design architecture
│  └─ [TODO] Write documentation
├─ [TODO] Review PR #123 (Priority: 2)
└─ [DONE] Deploy staging environment
```

### Technical Details
- **Terminal Width Detection**: Uses `golang.org/x/term` for cross-platform detection
- **Task Data Model** (from `backend/taskManager.go`):
  ```go
  type Task struct {
      UID          string
      Summary      string
      Description  string
      Status       string
      Priority     int        // 0-9 (0=undefined, 1=highest, 9=lowest)
      DueDate      time.Time
      StartDate    time.Time
      Created      time.Time
      Modified     time.Time
      Completed    time.Time
      Categories   []string
      ParentUID    string     // For subtask relationships
  }
  ```
- **Status Translation**:
  - **Internal Status**: todoat uses standardized internal statuses (TODO, DONE, IN-PROGRESS, CANCELLED) for all application logic
  - **Backend Status**: Each backend's native status format (e.g., NEEDS-ACTION, COMPLETED, IN-PROCESS for CalDAV) is automatically translated to/from internal status at storage boundaries
  - This dual-status system ensures consistent behavior across all backends while respecting each backend's protocol
- **Caching**: When sync is enabled, tasks are served from local SQLite cache for instant display

### Related Features
- [Task Filtering](#task-filtering) - Filter displayed tasks by status, priority, etc.
- [List Management](list-management.md) - Create and manage task lists
- [Subtasks & Hierarchy](subtasks-hierarchy.md) - View parent-child task relationships
- [Views & Customization](views-customization.md) - Customize task display format

---

## Add Tasks (Create)

### Feature Name
**Create New Tasks**

### Purpose
Enables users to add new tasks to a task list with optional metadata (description, priority, status, dates, parent task). Supports both simple task creation and advanced hierarchical task structures.

### How It Works

**User Actions:**
1. Execute one of the following commands:
   - `todoat MyList add "Task summary"` - Basic task creation
   - `todoat MyList a "Task summary"` - Using abbreviation
   - `todoat MyList add` - Prompts for summary interactively
   - `todoat MyList add "Task" -d "Details" -p 1` - With metadata
   - `todoat MyList add "Subtask" -P "Parent"` - Create subtask (see [Subtasks & Hierarchy](subtasks-hierarchy.md))
   - `todoat MyList add "parent/child/grandchild"` - Auto-create hierarchy
   - `todoat MyList add -l "literal/text"` - Disable path parsing with `-l` flag

**System Processes:**
1. **Input Validation**:
   - If summary is empty and not provided interactively, prompt user
   - Validate date formats (YYYY-MM-DD)
   - Validate priority range (0-9)
   - Validate status values (TODO/T, DONE/D, IN-PROGRESS/I, CANCELLED/C)

2. **UID Generation**:
   - System generates unique identifier for task
   - Format: UUID or backend-specific ID format
   - Ensures no collisions within the list

3. **Parent Task Resolution** (if `-P` flag used):
   - Search for parent task by summary or path
   - If path format ("Parent/Child"), resolve hierarchically
   - If not found, return error
   - See [Subtasks & Hierarchy](subtasks-hierarchy.md) for details

4. **Path-Based Hierarchy Creation** (if summary contains "/"):
   - Unless `-l` (literal) flag is used
   - Split summary by "/" delimiter
   - Auto-create parent tasks if they don't exist
   - Set parent-child relationships automatically
   - Example: "Project/Phase 1/Task A" creates three tasks with proper nesting

5. **Task Object Construction**:
   ```
   Task:
     UID: [generated]
     Summary: [user input]
     Description: [from -d flag or empty]
     Status: [from -S flag or default "TODO"]
     Priority: [from -p flag or 0]
     DueDate: [from --due-date or empty]
     StartDate: [from --start-date or empty]
     Created: [current timestamp]
     Modified: [current timestamp]
     ParentUID: [from parent resolution or empty]
   ```

6. **Backend Storage**:
   - Backend executes `AddTask(listID, task)` method
   - For Nextcloud: Creates iCalendar VTODO object, sends PUT request to CalDAV server
   - For SQLite: Inserts row into tasks table, marks as locally modified if sync enabled
   - For Git: Appends markdown task to TODO.md file
   - For Todoist: Sends POST request to REST API v2

7. **Sync Queue (if sync enabled)**:
   - Task is added to `sync_queue` table with operation type "create"
   - Background sync daemon picks up and syncs to remote backend
   - See [Synchronization](synchronization.md)

**Data Flow:**
```
User Input → Validation → UID Generation → Parent Resolution →
Path Hierarchy Processing → Task Construction → Backend Storage →
Sync Queue (if enabled) → Success Confirmation
```

**Integration Points:**
- [Backend System](backend-system.md) - Determines storage mechanism
- [Subtasks & Hierarchy](subtasks-hierarchy.md) - Parent-child relationships
- [Synchronization](synchronization.md) - Queues task for remote sync
- [List Management](list-management.md) - Validates target list exists

**Edge Cases:**
- Empty summary: Prompts user interactively or returns error
- Parent not found: Error message with suggestion to create parent first or use path syntax
- Invalid date format: Error message with correct format (YYYY-MM-DD)
- Invalid priority: Error message, must be 0-9
- Duplicate task: Allowed (tasks are distinguished by UID, not summary)
- Network failure (remote backend): Task queued locally, synced later (if sync enabled)
- Path syntax with `-l` flag: Treated as literal text, "/" not parsed

### User Journey
1. User identifies task to add
2. Types command: `todoat Work add "Review PR #456" -p 2 --due-date 2026-01-20`
3. System validates input and generates UID
4. Task is created in backend storage
5. Confirmation message: "Task 'Review PR #456' added to list 'Work'"
6. If sync enabled, task is queued for synchronization
7. User can immediately see task with `todoat Work`

### Prerequisites
- Target task list must exist (see [List Management](list-management.md))
- Backend must be configured and accessible
- For subtasks: Parent task must exist (or use path syntax for auto-creation)
- For network backends: Connectivity required (or sync must be enabled for offline operation)

### Outputs/Results
**Success Message:**
```
Task 'Review PR #456' added to list 'Work'
UID: 550e8400-e29b-41d4-a716-446655440000
```

**With Sync Enabled:**
```
Task 'Review PR #456' added to list 'Work'
UID: 550e8400-e29b-41d4-a716-446655440000
Queued for synchronization (will sync within 5 minutes)
```

### Technical Details
- **UID Generation**:
  - SQLite backend: Uses UUID v4
  - Nextcloud backend: Uses UUID formatted as CalDAV UID
  - Git backend: Generates hash-based UID
  - Todoist backend: Uses API-provided ID

- **Status Default**: If `-S` flag not provided, defaults to "TODO" (internal) or "NEEDS-ACTION" (CalDAV)

- **Timestamp Handling**:
  - Created and Modified timestamps set to current UTC time
  - Dates are stored in RFC3339 format internally
  - CalDAV dates converted to iCalendar format

- **Transaction Safety**:
  - SQLite: Operations wrapped in database transaction
  - Rollback on failure ensures data consistency

- **Shell Completion**: Task list names and status values auto-complete in supported shells (bash, zsh, fish, PowerShell)

### Related Features
- [Update Tasks](#update-tasks) - Modify existing tasks
- [Subtasks & Hierarchy](subtasks-hierarchy.md) - Create hierarchical task structures
- [Task Status System](#task-status-system) - Status value meanings
- [Task Priority System](#task-priority-system) - Priority value meanings
- [Synchronization](synchronization.md) - Offline task creation and sync

---

## Update Tasks

### Feature Name
**Modify Existing Task Properties**

### Purpose
Allows users to change task attributes including summary (rename), description, status, priority, due date, and start date. Supports partial matching for task identification.

### How It Works

**User Actions:**
1. Execute update command with task identifier and new values:
   - `todoat MyList update "Task name" -s DONE` - Change status
   - `todoat MyList u "partial" -p 5` - Using abbreviation, partial match
   - `todoat MyList update "old" --summary "New name"` - Rename task
   - `todoat MyList update "task" -d "New description"` - Update description
   - `todoat MyList update "task" --due-date 2026-02-15` - Change due date
   - `todoat MyList update "task" --due-date ""` - Clear due date (empty string)
   - `todoat MyList update "task" -p 0` - Clear priority (0 = undefined)

**System Processes:**
1. **Task Search** (see [Task Search and Selection](#task-search-and-selection)):
   - Search for task by summary using intelligent matching:
     - Exact match (case-insensitive) → select immediately
     - Single partial match → confirm with user
     - Multiple partial matches → present selection menu
     - No matches → error with suggestion

2. **Update Validation**:
   - Validate new status value (TODO/T, DONE/D, IN-PROGRESS/I, CANCELLED/C)
   - Validate priority range (0-9)
   - Validate date formats (YYYY-MM-DD) or empty string to clear
   - Validate at least one update flag is provided

3. **Task Modification**:
   - Retrieve current task object
   - Apply new values to specified fields only (unchanged fields preserved)
   - Update Modified timestamp to current time
   - If status changed to DONE, set Completed timestamp
   - If status changed from DONE to other, clear Completed timestamp

4. **Backend Update**:
   - Backend executes `UpdateTask(listID, task)` method
   - For Nextcloud: Modifies VTODO object, sends PUT request with If-Match etag header
   - For SQLite: Updates task row, marks as locally modified if sync enabled
   - For Git: Modifies markdown task line in TODO.md
   - For Todoist: Sends POST request to update endpoint

5. **Sync Queue (if sync enabled)**:
   - Task update added to `sync_queue` table with operation type "update"
   - Existing queue entries for same task are preserved (not deduplicated)
   - Background sync daemon processes updates in order

**Data Flow:**
```
User Command → Task Search → User Confirmation (if needed) →
Validation → Task Modification → Backend Update →
Sync Queue (if enabled) → Success Confirmation
```

**Integration Points:**
- [Backend System](backend-system.md) - Storage update mechanism
- [Task Search and Selection](#task-search-and-selection) - Intelligent task finding
- [Synchronization](synchronization.md) - Conflict detection and resolution
- [Task Status System](#task-status-system) - Status value translation

**Edge Cases:**
- Task not found: Error message with available similar tasks
- Multiple matches without user selection: User must choose from menu
- User cancels selection: Operation aborted gracefully
- No update flags provided: Error message listing available flags
- Invalid status/priority/date: Error message with valid values
- Concurrent modification (remote backend): Conflict resolution based on config (see [Synchronization](synchronization.md))
- Empty string for dates: Clears the date field
- Network failure: Update queued locally if sync enabled

### User Journey
1. User realizes task needs modification
2. Types: `todoat Work update "groceries" -s DONE --summary "Buy milk and eggs"`
3. If multiple tasks match "groceries", system presents selection menu
4. User confirms selection (or task is auto-selected if unique)
5. System updates task with new values
6. Confirmation: "Task 'Buy milk and eggs' updated successfully"
7. Modified timestamp updated
8. If sync enabled, update queued for synchronization

### Prerequisites
- Task list must exist and contain tasks
- Task to update must be identifiable by summary
- Backend must be accessible
- For remote backends: Network connectivity (or sync enabled for offline updates)

### Outputs/Results
**Success Message:**
```
Task 'Buy groceries' updated successfully
New summary: 'Buy milk and eggs'
Status: DONE
Modified: 2026-01-14 10:30:00 UTC
```

**With Multiple Matches:**
```
Multiple tasks found matching "groceries":
1. Buy groceries (TODO, Priority: 3)
2. Review grocery list (IN-PROGRESS)
Select task to update (1-2, or 'c' to cancel): 1

Task 'Buy groceries' updated successfully
Status: DONE
```

### Technical Details
- **Partial Matching Algorithm** (from `internal/operations/subtasks.go`):
  ```
  1. Search for exact match (case-insensitive)
  2. If not found, search for partial matches (case-insensitive substring)
  3. If single match, prompt for confirmation
  4. If multiple matches, present numbered selection menu
  5. Allow user to cancel at any prompt
  ```

- **Status Translation**:
  - CLI abbreviations (T/D/P/C) expanded to full status
  - Internal status (TODO/DONE) translated to CalDAV status (NEEDS-ACTION/COMPLETED)
  - Function: `StatusStringTranslateToStandardStatus()` and `StatusStringTranslateToAppStatus()`

- **Etag Handling** (Nextcloud backend):
  - Etag retrieved during task fetch
  - Included in If-Match header during update
  - Prevents overwriting concurrent modifications
  - If etag mismatch, triggers conflict resolution

- **Transaction Safety**:
  - SQLite updates wrapped in transactions
  - Rollback on validation failure

### Related Features
- [Complete Tasks](#complete-tasks) - Shortcut for marking tasks as DONE
- [Task Search and Selection](#task-search-and-selection) - How tasks are found
- [Synchronization](synchronization.md) - Conflict resolution for concurrent updates
- [Task Status System](#task-status-system) - Available status values

---

## Complete Tasks

### Feature Name
**Quick Task Completion**

### Purpose
Provides a convenient shortcut to mark tasks as DONE without requiring the full update command syntax. This is one of the most common operations in daily task management.

### How It Works

**User Actions:**
1. Execute complete command:
   - `todoat MyList complete "Task name"` - Mark as DONE
   - `todoat MyList c "partial"` - Using abbreviation
   - `todoat MyList complete "task"` - Partial match supported

**System Processes:**
1. **Task Search**: Uses same intelligent search as update (see [Task Search and Selection](#task-search-and-selection))
2. **Status Change**: Internally executes update operation with `-s DONE` flag
3. **Timestamp Update**: Sets Completed timestamp to current UTC time
4. **Backend Update**: Same mechanism as general update operation
5. **Sync Queue**: Queued as update operation if sync enabled

**Data Flow:**
```
User Command → Task Search → Status Change (→ DONE) →
Completed Timestamp → Backend Update → Sync Queue → Success
```

**Integration Points:**
- [Update Tasks](#update-tasks) - Complete is a specialized update operation
- [Task Search and Selection](#task-search-and-selection) - Task identification
- [Synchronization](synchronization.md) - Offline completion support

**Edge Cases:**
- Task already DONE: Success message, no change made
- Task not found: Same error handling as update
- Multiple matches: Same selection menu as update
- Network failure: Completion queued locally if sync enabled

### User Journey
1. User finishes a task
2. Types: `todoat Work complete "Review PR"`
3. System finds task and marks as DONE
4. Confirmation: "Task 'Review PR #456' marked as DONE"
5. Completed timestamp recorded
6. Task appears as completed in next get operation

### Prerequisites
- Same as [Update Tasks](#update-tasks)
- Task must exist and be identifiable

### Outputs/Results
**Success Message:**
```
Task 'Review PR #456' marked as DONE
Completed: 2026-01-14 10:30:00 UTC
```

### Technical Details
- **Implementation**: Wrapper around update operation with hardcoded `-s DONE` flag
- **Completed Timestamp**: Set to `time.Now().UTC()` when status changed to DONE
- **CalDAV Translation**: Status translated to "COMPLETED" for CalDAV backends
- **Idempotent**: Completing an already-done task is safe (no error)

### Related Features
- [Update Tasks](#update-tasks) - General task modification
- [Task Status System](#task-status-system) - Status meanings
- [Get Tasks](#get-tasks-listread) - View completed tasks

---

## Delete Tasks

### Feature Name
**Permanently Remove Tasks**

### Purpose
Allows users to delete tasks from a task list. Note that deletion is permanent and cannot be undone (unlike list deletion, which has trash/restore functionality).

### How It Works

**User Actions:**
1. Execute delete command:
   - `todoat MyList delete "Task name"` - Delete task
   - `todoat MyList d "partial"` - Using abbreviation

**System Processes:**
1. **Task Search**: Uses intelligent search to identify task (see [Task Search and Selection](#task-search-and-selection))
2. **Confirmation**: System may prompt for confirmation (depends on implementation)
3. **Backend Deletion**:
   - Backend executes `DeleteTask(listID, taskUID)` method
   - For Nextcloud: Sends DELETE request to CalDAV server
   - For SQLite: Deletes row from tasks table, or marks as locally deleted if sync enabled
   - For Git: Removes markdown task line from TODO.md
   - For Todoist: Sends DELETE request to REST API

4. **Sync Queue (if sync enabled)**:
   - Task deletion added to `sync_queue` table with operation type "delete"
   - Remote backend deletion processed by background sync daemon

**Data Flow:**
```
User Command → Task Search → Confirmation → Backend Deletion →
Sync Queue (if enabled) → Success Confirmation
```

**Integration Points:**
- [Backend System](backend-system.md) - Storage deletion mechanism
- [Task Search and Selection](#task-search-and-selection) - Task identification
- [Synchronization](synchronization.md) - Propagate deletion to remote
- [Subtasks & Hierarchy](subtasks-hierarchy.md) - Handle child task deletion

**Edge Cases:**
- Task not found: Error message
- Task has children (subtasks): May prompt for confirmation to delete entire subtree (depends on implementation)
- Network failure: Deletion queued locally if sync enabled
- Concurrent deletion (remote backend): No error if task already deleted

### User Journey
1. User decides task is no longer needed
2. Types: `todoat Work delete "outdated task"`
3. System finds task
4. Optional confirmation prompt
5. Task deleted from backend
6. Confirmation: "Task 'outdated task' deleted successfully"
7. Task no longer appears in task list

### Prerequisites
- Task list must exist
- Task must exist and be identifiable
- Backend must be accessible (or sync enabled for offline deletion)

### Outputs/Results
**Success Message:**
```
Task 'outdated task' deleted successfully
```

### Technical Details
- **Permanent Deletion**: Unlike list deletion, tasks cannot be restored (no trash)
- **SQLite Sync Behavior**:
  - With sync enabled: Row marked with `locally_deleted` flag, actually deleted after remote sync confirms
  - Without sync: Row immediately deleted from database
- **Child Task Handling**: If task has children, implementation may delete entire subtree or orphan children
- **Foreign Key Constraints**: SQLite schema enforces referential integrity for parent-child relationships

### Related Features
- [Subtasks & Hierarchy](subtasks-hierarchy.md) - Impact of deleting parent tasks
- [List Management](list-management.md#trash-management) - List trash/restore (not available for tasks)
- [Synchronization](synchronization.md) - Offline deletion support

---

## Task Filtering

### Feature Name
**Filter Displayed Tasks by Criteria**

### Purpose
Allows users to narrow down displayed tasks based on status, reducing visual clutter and focusing on relevant tasks (e.g., only TODO items, only completed tasks).

### How It Works

**User Actions:**
1. Execute get command with filter flag:
   - `todoat MyList -s TODO` - Show only TODO tasks
   - `todoat MyList -s TODO,IN-PROGRESS` - Multiple statuses
   - `todoat MyList -s T,D,I` - Using abbreviations
   - `todoat MyList -s D` - Show only completed tasks

**System Processes:**
1. **Flag Parsing**: System parses `-s` / `--status` flag value(s)
2. **Status Translation**: Abbreviations (T/D/I/C) expanded to full status names
3. **Backend Query**: Some backends support server-side filtering, others filter client-side
4. **Task Retrieval**: Same as normal get operation
5. **Client-Side Filtering**: Tasks filtered before rendering
6. **Display**: Only matching tasks shown in view

**Data Flow:**
```
User Command → Flag Parsing → Status Translation → Backend Query →
Client Filtering → View Rendering → Terminal Display
```

**Integration Points:**
- [Get Tasks](#get-tasks-listread) - Base task retrieval operation
- [Task Status System](#task-status-system) - Valid status values
- [Views & Customization](views-customization.md) - Display after filtering

**Edge Cases:**
- Invalid status value: Error message with valid statuses
- No tasks match filter: Message "No tasks found matching filter"
- Empty list: Same as unfiltered empty list

### User Journey
1. User has large task list, wants to focus on incomplete work
2. Types: `todoat Work -s TODO,IN-PROGRESS`
3. System retrieves tasks and filters by status
4. Only TODO and IN-PROGRESS tasks displayed
5. User reviews focused task list

### Prerequisites
- Task list must exist
- User must know valid status values (or use abbreviations)

### Outputs/Results
**Filtered Output:**
```
Tasks in "Work" (3 tasks, filtered by status: TODO, IN-PROGRESS):

├─ [TODO] Project Alpha (Priority: 1)
├─ [IN-PROGRESS] Design architecture
└─ [TODO] Write documentation
```

### Technical Details
- **Multiple Values**: `-s` flag accepts comma-separated list of statuses
- **Abbreviation Support**: T, D, P, C automatically expanded
- **Filter Logic**: Logical OR (task matches if status in filter list)
- **Performance**: Client-side filtering for most backends; Nextcloud supports server-side filtering via CalDAV REPORT queries

### Related Features
- [Task Status System](#task-status-system) - Available status values
- [Views & Customization](views-customization.md) - Advanced filtering in custom views
- [Get Tasks](#get-tasks-listread) - Base retrieval operation

---

## Task Search and Selection

### Feature Name
**Intelligent Task Identification**

### Purpose
Provides smart, user-friendly task lookup that handles exact matches, partial matches, and multiple matches with interactive selection. Reduces need for typing full task names. Also supports direct ID-based selection for unambiguous operations in scripts.

### How It Works

**User Actions:**
- User references task by summary in update/complete/delete commands
- Summary can be full name or partial match
- Alternatively, user specifies task directly by:
  - `--uid` flag: Select by backend-assigned UID (for synced tasks)
  - `--local-id` flag: Select by SQLite internal ID (only when sync enabled)

**System Processes:**

**0. Direct ID Selection (Bypass Search):**

   **UID Selection (`--uid`):**
   - If `--uid` flag provided, bypass all search logic
   - Look up task directly by backend-assigned UID
   - If found → proceed immediately
   - If not found → error with message
   - Used for synced tasks that have a backend-assigned UID

   **Local ID Selection (`--local-id`):**
   - Only available when sync is enabled (`sync.enabled: true`)
   - Look up task directly by SQLite internal ID
   - Useful for tasks that were just created locally and not yet synced
   - Example: `--local-id 42` looks up task with SQLite ID 42
   - Returns error if sync is disabled (local IDs only exist in SQLite cache)

1. **Exact Match Search** (Case-Insensitive):
   - Search all tasks in list for exact summary match
   - If found → return task immediately (no confirmation needed)

2. **Partial Match Search** (if exact match not found):
   - Search for tasks containing the search string as substring (case-insensitive)
   - Examples: "groceries" matches "Buy groceries", "grocery list", "Groceries for party"

3. **Single Match Handling**:
   - If exactly one partial match found
   - **Normal Mode**: Display task details, prompt "Is this the task you want? (y/n)"
   - **No-Prompt Mode** (`-y`): Use match automatically without confirmation
   - If user confirms → proceed with operation
   - If user declines → abort operation

4. **Multiple Match Handling**:
   - If multiple matches found
   - **Normal Mode**:
     - Display numbered list with task details (summary, status, priority)
     - Prompt: "Select task to [operation] (1-N, or 'c' to cancel):"
     - User enters number or 'c'
     - Selected task used for operation
   - **No-Prompt Mode** (`-y`):
     - Output match table with UIDs and parent context
     - Return `ACTION_INCOMPLETE` result code
     - User/script must re-run with `--uid` flag

5. **No Match Handling**:
   - Display error: "Task not found matching 'search term'"
   - Suggest: "Use 'todoat MyList' to see all tasks"
   - List similar tasks if any exist (fuzzy matching)

**Data Flow:**
```
--uid or --local-id Flag? → Yes → Look up by ID → Found? → Proceed
                                                   ↓ No → Error

No ID flags:
Search String → Exact Match → Found? → Return Task
                              ↓ No
                           Partial Match → Count Matches
                                           ↓
                              Single → Confirm (or auto in -y) → Proceed
                                           ↓
                              Multiple → Select (or ACTION_INCOMPLETE in -y) → Proceed
                                           ↓
                              None → Error + Suggestions
```

**Integration Points:**
- [Update Tasks](#update-tasks) - Used for task identification
- [Complete Tasks](#complete-tasks) - Used for task identification
- [Delete Tasks](#delete-tasks) - Used for task identification
- [CLI Interface](cli-interface.md) - Interactive prompts and menus
- [CLI Interface - No-Prompt Mode](cli-interface.md#no-prompt-mode) - Non-interactive behavior
- [CLI Interface - Result Codes](cli-interface.md#result-codes) - `ACTION_INCOMPLETE` for ambiguous matches

**Edge Cases:**
- User cancels confirmation: Operation aborted gracefully
- User cancels selection: Operation aborted gracefully
- User enters invalid selection number: Re-prompt with error
- No tasks in list: Immediate error "No tasks found"
- Search string matches all tasks: Present full list for selection
- **No-prompt mode with multiple matches**: Returns `ACTION_INCOMPLETE` with match list including `local_id` for each task
- **UID not found**: Error with message, suggests listing tasks
- **Local ID not found**: Error with message
- **--local-id used with sync disabled**: Error "Local ID selection requires sync to be enabled"

### User Journey

**ID-Based Selection (Scripting):**
1. User runs: `todoat -y Work complete "review"`
2. Multiple matches found, returns:
   ```
   Multiple tasks match "review":
   ID:42	UID:550e8400-e29b-41d4-a716-446655440000	TASK:Review PR #456	PARENT:Project Alpha
   ID:43	UID:660e8400-e29b-41d4-a716-446655440001	TASK:Code review	PARENT:
   ID:44	UID:	TASK:Review notes	PARENT:Documentation
   ACTION_INCOMPLETE
   ```
3. User/script re-runs with UID (for synced task): `todoat Work complete --uid "550e8400-e29b-41d4-a716-446655440000"`
4. Or with local ID (for any task): `todoat Work complete --local-id 44`
5. Task completed immediately (no prompts)
6. Output: `ACTION_COMPLETED`

**Unsynced Task Selection:**
1. User creates task locally while offline: `todoat -y Work add "Draft proposal"`
2. Task created with SQLite ID (no remote UID yet)
3. Response includes: `local_id: 44` (the SQLite internal ID)
4. User can immediately operate on it using local ID: `todoat Work update --local-id 44 -p 1`
5. User lists tasks: `todoat -y --json Work`
6. Output shows: `"local_id": 44, "uid": null, "synced": false`
7. After sync, `uid` becomes backend-assigned UUID: `"local_id": 44, "uid": "550e8400...", "synced": true`

**Single Partial Match (Normal Mode):**
1. User types: `todoat Work update "PR" -s DONE`
2. System finds one task: "Review PR #456"
3. Prompt: "Found: Review PR #456 [TODO, Priority: 2]. Is this correct? (y/n)"
4. User types: `y`
5. Task updated

**Single Partial Match (No-Prompt Mode):**
1. User types: `todoat -y Work update "PR" -s DONE`
2. System finds one task: "Review PR #456"
3. Auto-confirms (no prompt)
4. Task updated
5. Output: `ACTION_COMPLETED`

**Multiple Matches (Normal Mode):**
1. User types: `todoat Work complete "review"`
2. System finds three matches:
   ```
   Multiple tasks found matching "review":
   1. Review PR #456 (TODO, Priority: 2)
   2. Code review guidelines (DONE)
   3. Review meeting notes (IN-PROGRESS)
   Select task (1-3, or 'c' to cancel):
   ```
3. User types: `1`
4. Task "Review PR #456" marked as DONE

**Multiple Matches (No-Prompt Mode with JSON):**
1. User types: `todoat -y --json Work complete "review"`
2. System returns:
   ```json
   {
     "matches": [
       {"local_id": 42, "uid": "550e8400...", "summary": "Review PR #456", "status": "TODO", "parents": ["Project Alpha"], "synced": true},
       {"local_id": 43, "uid": "660e8400...", "summary": "Code review guidelines", "status": "DONE", "parents": [], "synced": true},
       {"local_id": 44, "uid": null, "summary": "Review meeting notes", "status": "IN-PROGRESS", "parents": [], "synced": false}
     ],
     "result": "ACTION_INCOMPLETE",
     "message": "Multiple tasks match 'review'. Use --uid or --local-id to specify exact task."
   }
   ```
3. Script parses JSON, selects appropriate task by `uid` (for synced) or `local_id` (for any)
4. Script re-runs: `todoat -y Work complete --uid "550e8400..."` or `todoat -y Work complete --local-id 44`

### Prerequisites
- Task list must contain tasks
- User must provide search string, `--uid` flag, or `--local-id` flag
- For `--uid`: UID must exist in the task list
- For `--local-id`: Sync must be enabled, and local ID must exist in the SQLite cache

### Outputs/Results
See User Journey examples above.

**Result Codes (No-Prompt Mode):**
| Scenario | Result Code | Message Format |
|----------|-------------|----------------|
| Task found and operation completed | `ACTION_COMPLETED` | - |
| Multiple matches, cannot proceed | `ACTION_INCOMPLETE` | "Multiple tasks match '...'. Use --uid to specify exact task." |
| Task not found | `ERROR` | "Error 1: Task '...' not found in list '...'" |
| Invalid input | `ERROR` | "Error 2: ..." |
| Backend error | `ERROR` | "Error 3: ..." |

### Technical Details
- **Function**: `findTaskBySummary()` in `internal/operations/subtasks.go`
- **UID Lookup**: Direct database/backend query by UID field
- **Local ID Lookup**: Direct SQLite query by internal ID (only when sync enabled)
- **Case Sensitivity**: All searches are case-insensitive
- **Substring Matching**: Uses `strings.Contains()` after lowercasing
- **User Input**: Uses `bufio.NewScanner()` for terminal input (skipped in no-prompt mode)
- **Cancellation**: User can type 'c' or 'cancel' at any prompt
- **No-Prompt Mode**: Controlled by `-y` flag or `no_prompt: true` in config

**ID Flags:**
```go
// CLI flag registration
cmd.Flags().String("uid", "", "Select task by backend-assigned UID (bypasses search)")
cmd.Flags().Int64("local-id", 0, "Select task by SQLite internal ID (requires sync enabled)")
```

**ID Resolution Logic:**
```go
func resolveTask(uid string, localID int64, syncEnabled bool, backend TaskManager) (*Task, error) {
    // Local ID takes precedence if provided
    if localID > 0 {
        if !syncEnabled {
            return nil, errors.New("Local ID selection requires sync to be enabled")
        }
        return backend.GetTaskByLocalID(localID)
    }
    // UID lookup for synced tasks
    if uid != "" {
        return backend.GetTaskByUID(uid)
    }
    return nil, errors.New("No task identifier provided")
}
```

**Task ID Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `local_id` | int64 | SQLite internal ID (always present when sync enabled) |
| `uid` | string | Backend-assigned UID (present after sync, null/empty before) |
| `synced` | bool | Whether task has been synced to remote backend |

### Related Features
- [Update Tasks](#update-tasks) - Primary consumer
- [Complete Tasks](#complete-tasks) - Primary consumer
- [Delete Tasks](#delete-tasks) - Primary consumer
- [Subtasks & Hierarchy](subtasks-hierarchy.md) - Parent task resolution uses same logic
- [CLI Interface - No-Prompt Mode](cli-interface.md#no-prompt-mode) - Non-interactive operation
- [CLI Interface - JSON Output](cli-interface.md#json-output-mode) - Machine-parseable output
- [CLI Interface - Result Codes](cli-interface.md#result-codes) - Operation outcome indicators
- [Synchronization](synchronization.md) - `--local-id` requires sync to be enabled

---

## Task Status System

### Feature Name
**Task Lifecycle State Management**

### Purpose
Provides standardized status values to track task progress through its lifecycle from creation to completion or cancellation. Enables filtering, reporting, and workflow management.

**Important: Internal vs Backend Status**

todoat uses a **dual status system** to maintain consistency across different backends while respecting each backend's native format:

- **Internal Status** (Application Level): The canonical status used throughout todoat for all operations, filtering, and logic (TODO, DONE, IN-PROGRESS, CANCELLED)
- **Backend-Specific Status** (Storage Level): The native status format used by each backend's protocol or API

This separation allows todoat to provide a consistent user experience while working with multiple backend systems that have different status vocabularies.

### How It Works

**Status Values and Translation:**

| Internal Status (App) | CalDAV/Nextcloud Status | CLI Abbreviation | Meaning |
|----------------------|------------------------|------------------|---------|
| TODO | NEEDS-ACTION | T | Task not started, pending work |
| DONE | COMPLETED | D | Task finished successfully |
| IN-PROGRESS | IN-PROCESS | I | Task in progress, actively being worked on |
| CANCELLED | CANCELLED | C | Task abandoned, no longer relevant |

**Note:** Different backends use different status names:
- **CalDAV/Nextcloud**: NEEDS-ACTION, COMPLETED, IN-PROCESS, CANCELLED
- **SQLite**: Stores internal status directly (TODO, DONE, IN-PROGRESS, CANCELLED)
- **Git**: Maps to markdown checkbox syntax (`[ ]`, `[x]`, `[>]`, `[-]`)
- **Todoist**: Maps to API-native status (open/closed)

**User Actions:**
- Set status when adding: `-S TODO` or `-S T`
- Update status: `update "task" -s DONE`
- Filter by status: `-s TODO,IN-PROGRESS`

**System Processes:**
1. **Status Translation Chain**:
   - **User Input** → **Internal Status** → **Backend Status**
   - CLI abbreviations (T/D/I/C) → Internal status (TODO/DONE/IN-PROGRESS/CANCELLED)
   - Internal status → Backend-specific status (e.g., NEEDS-ACTION/COMPLETED/IN-PROCESS for CalDAV)
   - Functions: `StatusStringTranslateToStandardStatus()` (CLI→Internal) and `StatusStringTranslateToAppStatus()` (Internal→Backend)
   - All application logic operates on internal status only
   - Backend translation happens only at storage/retrieval boundaries

2. **Status-Driven Behavior**:
   - When status → DONE: Set Completed timestamp
   - When status DONE → other: Clear Completed timestamp
   - Filtering uses internal status values

3. **Backend-Specific Handling**:
   - Nextcloud: Stores as CalDAV VTODO STATUS property
   - SQLite: Stores as text in status column
   - Git: Translates to markdown checkbox syntax (`[ ]`, `[x]`, `[>]`, `[-]`)
   - Todoist: Uses API-native status (open/closed)

**Data Flow:**
```
User Input (CLI abbreviation) → Internal Status Translation →
Application Logic (filtering, updates, etc.) → Backend-Specific Translation →
Storage Format (NEEDS-ACTION for CalDAV, TODO for SQLite, etc.)
```

**Integration Points:**
- [Add Tasks](#add-tasks-create) - Set initial status
- [Update Tasks](#update-tasks) - Change task status
- [Complete Tasks](#complete-tasks) - Shortcut to DONE status
- [Task Filtering](#task-filtering) - Filter by status
- [Backend System](backend-system.md) - Backend-specific status storage

**Edge Cases:**
- Invalid status value: Error with list of valid values
- Case variations: Status is case-insensitive ("done", "DONE", "Done" all work)
- Unknown CalDAV status from server: Defaults to TODO with warning

### User Journey
1. User creates task: defaults to TODO
2. User starts work: `update "task" -s IN-PROGRESS`
3. Task in progress state
4. User finishes: `complete "task"` (sets to DONE)
5. Task marked complete with timestamp
6. Or user abandons: `update "task" -s CANCELLED`

### Prerequisites
- None (status is fundamental task property)

### Outputs/Results
Status is visible in all task displays:
```
[TODO] Review PR #456
[IN-PROGRESS] Write documentation
[DONE] Deploy staging
[CANCELLED] Old feature request
```

### Technical Details
- **Default Status**: New tasks default to TODO (NEEDS-ACTION in CalDAV)
- **Git Markdown Mapping**:
  ```
  [ ] → TODO (NEEDS-ACTION)
  [x] → DONE (COMPLETED)
  [>] → IN-PROGRESS (IN-PROCESS)
  [-] → CANCELLED
  ```
- **Todoist Mapping**: Only supports open/closed, IN-PROGRESS mapped to open
- **Status Completion**: Shell completion provides status values when typing flags

### Related Features
- [Task Filtering](#task-filtering) - Filter tasks by status
- [Complete Tasks](#complete-tasks) - Shortcut to DONE
- [Update Tasks](#update-tasks) - General status changes
- [Views & Customization](views-customization.md) - Custom status display formatting

---

## Task Priority System

### Feature Name
**Task Importance Ranking**

### Purpose
Allows users to assign importance levels to tasks, enabling prioritization of work and visual highlighting of high-priority items. Follows iCalendar VTODO standard (0-9 scale).

### How It Works

**Priority Scale:**
- **0**: Undefined (no priority set)
- **1**: Highest priority (most urgent/important)
- **2-4**: High priority
- **5**: Medium priority (default if manually setting)
- **6-8**: Low priority
- **9**: Lowest priority

**User Actions:**
- Set priority when adding: `-p 1` (highest)
- Update priority: `update "task" -p 5` (medium)
- Clear priority: `update "task" -p 0` (undefined)

**System Processes:**
1. **Priority Validation**: Ensure value is 0-9 integer
2. **Storage**: Stored as integer in task object
3. **Display**:
   - Backend method `GetPriorityColor()` returns color code for terminal display
   - High priorities (1-4) typically displayed in red/orange
   - Medium (5) in yellow
   - Low (6-9) in blue/gray
   - Undefined (0) may be hidden or shown neutrally

4. **Sorting**: Tasks can be sorted by priority in custom views (see [Views & Customization](views-customization.md))

**Data Flow:**
```
User Input (-p flag) → Validation (0-9) → Task Object → Backend Storage →
Display Rendering (with color)
```

**Integration Points:**
- [Add Tasks](#add-tasks-create) - Set initial priority
- [Update Tasks](#update-tasks) - Change priority
- [Views & Customization](views-customization.md) - Priority-based sorting and coloring
- [Backend System](backend-system.md) - Backend-specific color schemes

**Edge Cases:**
- Invalid priority (negative, > 9, non-integer): Error message
- Priority 0 vs. not setting priority: Both result in undefined priority
- Backend doesn't support priority: Stored but may not be used

### User Journey
1. User creates urgent task: `todoat Work add "Fix production bug" -p 1`
2. Task displayed with high-priority indicator (red color)
3. User creates low-priority task: `add "Clean up docs" -p 8`
4. Task displayed with low-priority indicator (gray color)
5. User re-prioritizes: `update "docs" -p 3` (now high priority)

### Prerequisites
- None (priority is optional task metadata)

### Outputs/Results
**Task Display with Priority:**
```
Tasks in "Work":

├─ [TODO] Fix production bug (Priority: 1) ← Red color
├─ [TODO] Review PR (Priority: 3) ← Orange color
├─ [TODO] Update README (Priority: 5) ← Yellow color
└─ [TODO] Clean up docs (Priority: 8) ← Gray color
```

### Technical Details
- **iCalendar Standard**: Follows RFC 5545 VTODO PRIORITY property specification
- **CalDAV Storage**: Stored as PRIORITY:N in VTODO object
- **SQLite Storage**: Stored as INTEGER in priority column
- **Git Markdown**: Stored as metadata tag `@priority:N`
- **Todoist**: Mapped to Todoist's 1-4 priority scale (with conversion)

- **Color Mapping** (example from Nextcloud backend):
  ```go
  func (nb *NextcloudBackend) GetPriorityColor(priority int) string {
      switch {
      case priority >= 1 && priority <= 4:
          return "\033[91m" // Red
      case priority == 5:
          return "\033[93m" // Yellow
      case priority >= 6:
          return "\033[94m" // Blue
      default:
          return "\033[0m"  // Reset
      }
  }
  ```

### Related Features
- [Add Tasks](#add-tasks-create) - Set priority on creation
- [Update Tasks](#update-tasks) - Modify priority
- [Views & Customization](views-customization.md) - Priority-based sorting and custom colors
- [Get Tasks](#get-tasks-listread) - Display with priority indicators

---

## Task Dates

### Feature Name
**Task Temporal Metadata**

### Purpose
Enables users to set due dates, start dates, and track creation/modification/completion timestamps. Supports deadline management, scheduling, and task lifecycle tracking.

### How It Works

**Date Types:**

| Date Field | User-Settable | Auto-Set | Purpose | Format |
|------------|---------------|----------|---------|--------|
| Due Date | Yes (`--due-date`) | No | Task deadline | YYYY-MM-DD |
| Start Date | Yes (`--start-date`) | No | When to begin task | YYYY-MM-DD |
| Created | No | On creation | Task creation timestamp | RFC3339 |
| Modified | No | On update | Last modification timestamp | RFC3339 |
| Completed | No | When status → DONE | Task completion timestamp | RFC3339 |

**User Actions:**
- Set due date: `add "Task" --due-date 2026-01-31`
- Set start date: `add "Task" --start-date 2026-01-15`
- Update dates: `update "task" --due-date 2026-02-15`
- Clear dates: `update "task" --due-date ""`

**System Processes:**
1. **Date Parsing**:
   - User input format: YYYY-MM-DD
   - Parsed to time.Time object
   - Validated for format correctness
   - Empty string clears date field

2. **Automatic Timestamps**:
   - **Created**: Set to current UTC time when task added
   - **Modified**: Updated to current UTC time on every update
   - **Completed**: Set when status changes to DONE, cleared when status changes from DONE

3. **Backend Storage**:
   - **CalDAV**: Dates stored as iCalendar DATE or DATE-TIME properties
     - DUE, DTSTART (start), CREATED, LAST-MODIFIED, COMPLETED
   - **SQLite**: Stored as TEXT in RFC3339 format
   - **Git**: Stored as metadata tags `@due:YYYY-MM-DD`, `@created:YYYY-MM-DD`, etc.
   - **Todoist**: Uses API-native date format

4. **Display**:
   - Dates shown in custom views (see [Views & Customization](views-customization.md))
   - Not shown in default view (only in "all" view or custom views)
   - Time zone conversion for display (UTC stored, local displayed)

**Data Flow:**
```
User Input (--due-date) → Date Parsing → Validation → Task Object →
Backend Translation → Storage Format → Display Rendering
```

**Integration Points:**
- [Add Tasks](#add-tasks-create) - Set initial dates
- [Update Tasks](#update-tasks) - Modify dates
- [Complete Tasks](#complete-tasks) - Sets Completed timestamp
- [Views & Customization](views-customization.md) - Display date fields
- [Backend System](backend-system.md) - Backend-specific date formats

**Edge Cases:**
- Invalid date format: Error message "Date must be YYYY-MM-DD format"
- Past due date: Accepted without warning (user may intentionally set past dates)
- Future completed date: Accepted (system trusts user input)
- Empty string for dates: Clears the date field (removes due date)
- Start date after due date: Accepted without warning (validation could be added)

### User Journey
1. User creates task with deadline: `todoat Work add "Submit report" --due-date 2026-01-20`
2. Task stored with due date
3. View with dates: `todoat Work -v all` shows due date
4. User extends deadline: `update "report" --due-date 2026-01-25`
5. User completes task: `complete "report"`
6. System sets Completed timestamp automatically
7. View shows: Created, Modified, Completed, Due Date

### Prerequisites
- None (dates are optional task metadata)
- For date display: Must use view that includes date fields

### Outputs/Results
**Task with All Date Metadata (using -v all):**
```
Task: Submit report
Status: DONE
Priority: 2
Due Date: 2026-01-25
Start Date: 2026-01-15
Created: 2026-01-10T09:00:00Z
Modified: 2026-01-20T14:30:00Z
Completed: 2026-01-20T14:30:00Z
```

### Technical Details
- **Date Storage Format**: RFC3339 (ISO 8601 with time zone)
  - Example: `2026-01-14T10:30:00Z`
- **User Input Format**: YYYY-MM-DD (date only)
  - Converted to RFC3339 with time 00:00:00 in user's timezone
- **CalDAV Format Conversion**:
  - DATE property: `20260114` (yyyymmdd)
  - DATE-TIME property: `20260114T103000Z` (yyyymmddThhmmssZ)
- **Time Zone Handling**:
  - Stored in UTC
  - Displayed in user's local timezone (or UTC if local TZ unavailable)
- **Nil/Empty Dates**: Represented as zero time.Time value or NULL in database

### Related Features
- [Add Tasks](#add-tasks-create) - Set dates on creation
- [Update Tasks](#update-tasks) - Modify dates
- [Complete Tasks](#complete-tasks) - Auto-sets completion timestamp
- [Views & Customization](views-customization.md) - Display and filter by dates
- [Task Metadata](#task-metadata) - Complete metadata overview

---

## Task Metadata

### Feature Name
**Complete Task Information Schema**

### Purpose
Provides comprehensive overview of all data associated with a task, including system-managed and user-managed properties. Essential for understanding the full task data model.

### How It Works

**Task Data Model:**

```go
type Task struct {
    // Identifiers
    UID         string    // Unique identifier (UUID or backend-specific)

    // User-Managed Content
    Summary     string    // Task name/title (required)
    Description string    // Detailed notes (optional)

    // Status & Priority
    Status      string    // TODO/DONE/IN-PROGRESS/CANCELLED
    Priority    int       // 0-9 (0=undefined, 1=highest, 9=lowest)

    // Dates
    DueDate     time.Time // Task deadline (optional)
    StartDate   time.Time // When to begin (optional)
    Created     time.Time // Creation timestamp (auto-set)
    Modified    time.Time // Last modification (auto-updated)
    Completed   time.Time // Completion timestamp (auto-set when DONE)

    // Organization
    Categories  []string  // Tags/labels (optional, backend-dependent)

    // Hierarchy
    ParentUID   string    // Parent task UID for subtasks (optional)
}
```

**Viewing All Metadata:**
- Command: `todoat MyList -v all`
- Shows all fields for all tasks
- Or create custom view with desired fields (see [Views & Customization](views-customization.md))

**Data Flow:**
```
Task Creation/Update → Field Population → Backend Storage →
Retrieval → View Formatting → Display
```

**Integration Points:**
- [Add Tasks](#add-tasks-create) - Populates metadata on creation
- [Update Tasks](#update-tasks) - Modifies metadata
- [Views & Customization](views-customization.md) - Controls which metadata is displayed
- [Subtasks & Hierarchy](subtasks-hierarchy.md) - ParentUID field usage
- [Backend System](backend-system.md) - Storage of metadata

**Edge Cases:**
- Optional fields empty: Displayed as empty or hidden depending on view
- Long descriptions: May be truncated in list view, full text in detail view
- Categories not supported by backend: Silently ignored
- Invalid UID: System regenerates on next sync

### User Journey
1. User wants to see all task details
2. Types: `todoat Work -v all`
3. System displays comprehensive task information including all metadata
4. User sees creation dates, modification history, due dates, priorities, etc.

### Prerequisites
- Task must exist
- Use `-v all` flag or custom view with desired fields

### Outputs/Results
**Complete Task Metadata Display:**
```
Tasks in "Work" (all fields):

Task: Fix production bug
  UID: 550e8400-e29b-41d4-a716-446655440000
  Status: IN-PROGRESS
  Priority: 1
  Description: Critical bug affecting user login flow
  Due Date: 2026-01-15
  Start Date: 2026-01-14
  Created: 2026-01-10T09:00:00Z
  Modified: 2026-01-14T08:30:00Z
  Parent: Project Alpha
  Categories: [bug, urgent, backend]
```

### Technical Details
- **UID Generation**: UUID v4 for most backends, API-provided for Todoist
- **Zero Values**: Empty time.Time, empty strings, priority 0 represent "not set"
- **Categories**:
  - CalDAV: Stored as CATEGORIES property (comma-separated)
  - Git: Stored as inline tags `#tag`
  - SQLite: JSON array in categories column
  - Todoist: Mapped to Todoist labels
- **Field Limits**:
  - Summary: Typically 255 chars (backend-dependent)
  - Description: No practical limit (multi-KB supported)
  - UID: 36 chars (UUID) or backend-specific

### Related Features
- [Get Tasks](#get-tasks-listread) - Display metadata
- [Views & Customization](views-customization.md) - Control metadata display
- [Add Tasks](#add-tasks-create) - Set metadata on creation
- [Update Tasks](#update-tasks) - Modify metadata
- [Task Status System](#task-status-system) - Status field details
- [Task Priority System](#task-priority-system) - Priority field details
- [Task Dates](#task-dates) - Date field details
- [Subtasks & Hierarchy](subtasks-hierarchy.md) - ParentUID usage

---

## Summary

Task Management in todoat provides comprehensive CRUD operations with intelligent search, flexible filtering, and rich metadata support. Key capabilities include:

- **CRUD Operations**: Create, read, update, delete tasks with validation and error handling
- **Intelligent Search**: Partial matching with confirmation, multi-select menus
- **Status Management**: Four-state lifecycle (TODO/DONE/IN-PROGRESS/CANCELLED)
- **Priority System**: 0-9 scale with visual indicators
- **Date Support**: Due dates, start dates, automatic timestamps
- **Filtering**: By status, priority, dates (in custom views)
- **Offline Support**: Queue operations when sync enabled
- **Multi-Backend**: Consistent interface across Nextcloud, SQLite, Git, Todoist

For hierarchical task organization, see [Subtasks & Hierarchy](subtasks-hierarchy.md).
For backend-specific behavior, see [Backend System](backend-system.md).
For customizing task display, see [Views & Customization](views-customization.md).
