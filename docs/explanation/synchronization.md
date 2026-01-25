# Synchronization System

## Overview

The synchronization system provides bidirectional data synchronization between local SQLite databases (caches) and remote backends (Nextcloud, Todoist, etc.), enabling offline task management with automatic conflict resolution.

**Key Capabilities:**
- Offline-first architecture with local caching
- Bidirectional sync (pull from remote, push local changes)
- Automatic conflict detection and resolution
- Operation queueing with retry logic
- Multi-remote backend support with isolated caches

---

## Table of Contents

- [Global Sync Architecture](#global-sync-architecture)
- [Automatic Caching System](#automatic-caching-system)
- [Sync Operations](#sync-operations)
- [Conflict Resolution](#conflict-resolution)
- [Offline Mode](#offline-mode)
- [Sync Queue System](#sync-queue-system)
- [Manual Sync Workflow](#manual-sync-workflow)
- [Configuration](#configuration)
- [Database Locations](#database-locations)
- [Database Schema](#database-schema)
- [Performance Characteristics](#performance-characteristics)

---

## Global Sync Architecture

### Purpose
Enable users to work with tasks at the speed of local while maintaining consistency with remote backends through automatic caching and synchronization.

### How It Works

**1. Cache Database Creation**
- When sync.enabled = true in config, all remote backends share a single SQLite cache database
- Each backend gets its own dedicated table(s) within the shared database
- Cache database is stored at a configurable location (e.g., cache.db)
- Example: Nextcloud backend → nextcloud table, Todoist backend → todoist table

**2. Data Flow**
```
User → CLI Command → Sync Manager →  Database (SQLite)
                           ↓              ↓
                    Sync Operations  sync_queue table
                           ↓              ↓
                    Remote Backend  Manual Sync Push
                           ↓              ↓
                    Remote Server ←────────
```

**Important: Sync Manager Role**
- When sync is enabled and CLI commands are executed, the **Sync Manager** is the component that orchestrates all operations
- The Sync Manager sits between the CLI layer and the storage layer (cache database + remote backend)
- For CLI operations (add, update, complete, delete):
  1. Sync Manager receives the operation request
  2. Updates the local cache database (SQLite)
  3. Queues the operation in `sync_queue` table for later synchronization
  4. Optionally triggers background sync (when auto-sync is re-enabled)
- For sync operations (`todoat sync`):
  1. Sync Manager pulls changes from remote backend
  2. Merges with local cache, resolving conflicts
  3. Pushes queued operations from `sync_queue` to remote backend
  4. Updates sync metadata (etags, timestamps)

**3. Isolation Between Backends**
- Each remote backend has a completely separate cache database
- Tasks from different backends never mix
- Each cache maintains its own sync metadata and queue
- Each backend has its own Sync Manager instance

**4. Transparent Operation**
- When sync is enabled, CLI operations are routed through the Sync Manager
- Sync Manager updates cache databases and queues operations
- Changes are persisted locally immediately (offline-capable)
- User manually triggers sync with `todoat sync` to push/pull changes with remote backends

### User Journey

1. Enable sync in config: `sync.enabled = true`
2. Configure remote backends (e.g., Nextcloud)
3. Use CLI normally: `todoat MyList add "Task"`
4. Changes are saved to local cache and queued
5. Run `todoat sync` when online to synchronize
6. Check sync status: `todoat sync status`

### Prerequisites
- [Backend configured](backend-system.md#configuration) with valid credentials
- [Credential management](credential-management.md) setup (keyring, env vars, or config URL)
- Network connectivity for initial sync (offline mode for subsequent work)

### Technical Details

**Sync Implementation** (`cmd/todoat/cmd/todoat.go`):
- **Primary Role**: Manages synchronization between local SQLite cache and remote backends
- **CLI Operations**: When sync is enabled with `offline_mode: auto` (default), CLI operations use SQLite cache directly for instant response
- **Push Operations**: The `todoat sync` command processes the sync queue and applies queued operations to remote backend
- **Queue Management**: Operations are queued in `sync_queue` table and processed during sync
- **Hierarchical Ordering**: Ensures parents are synced before children (prevents foreign key violations)

**Sync-Aware Backend Wrapper** (`syncAwareBackend`):
- Wraps the SQLite backend to add sync queue support
- Automatically queues create/update/delete operations for later sync
- Provides `getSyncManager()` to access queue operations

**SQLiteBackend Sync Methods** (used by sync commands):
- `GetPendingSyncOperations()` - Retrieve queued operations for push
- `AddToSyncQueue()` - Queue operation for later synchronization
- `ClearQueue()` - Clear the sync queue after successful sync

### Related Features
- [Backend System](backend-system.md) - Remote backend implementations
- [Credential Management](credential-management.md) - Authentication for remotes
- [Task Management](task-management.md) - CRUD operations that trigger sync

---

## Automatic Caching System

### Purpose
Provide automatic local caching for all remote backends without manual configuration, enabling offline work and reducing network latency.

### How It Works

**1. Automatic Cache Creation**
When sync is globally enabled:
```yaml
sync:
  enabled: true
  local_backend: sqlite  # Cache implementation type
```

Every remote backend automatically gets a cache:
```yaml
backends:
  nextcloud:
    type: nextcloud
    enabled: true
    # Auto-cached at: ~/.local/share/todoat/caches/nextcloud.db

  todoist:
    type: todoist
    enabled: true
    # Auto-cached at: ~/.local/share/todoat/caches/todoist.db
```

**2. Cache Path Generation**
- Base directory: `$XDG_DATA_HOME/todoat/caches/` (default: `~/.local/share/todoat/caches/`)
- Cache filename: `[backend-name].db`
- Automatic directory creation if missing

**3. Opt-Out Mechanism**
Individual backends can disable caching:
```yaml
backends:
  nextcloud:
    type: nextcloud
    enabled: true
    sync:
      enabled: false  # Don't cache this backend
```

**4. Backend Selection with Sync**
- **Sync disabled**: CLI uses `backend_priority` or `default_backend` directly
- **Sync enabled**: CLI uses cache database for remote backends
- Explicit `--backend` flag always overrides automatic selection

### User Journey

**Initial Setup:**
1. Enable global sync in config
2. Add remote backend configuration
3. First CLI operation triggers cache creation
4. Perform initial sync: `todoat sync`

**Daily Usage:**
1. Work with tasks: `todoat MyList add "Task"`
2. Changes saved to cache instantly
3. Sync when convenient: `todoat sync`

### Outputs/Results
- Local cache database created automatically
- Fast CLI operations (no network wait)
- Persistent local storage of all task data

### Technical Details

**Cache Database Initialization:**
- Schema automatically created from `backend/schema.go`
- Includes all tables: `tasks`, `sync_metadata`, `list_sync_metadata`, `sync_queue`, `schema_version`
- Indexes created for optimal query performance

**Cache Isolation:**
- Each backend's cache is completely independent
- No shared state between different remote backends
- Prevents data corruption from mixed operations

---

## Sync Operations

### Purpose
Synchronize local changes with remote backends and fetch remote updates to local cache.

### How It Works

**Pull Operation:**
1. **Fetch Remote State**
   - Retrieve all task lists from remote backend
   - For each list, get all tasks with metadata (etags, ctags)

2. **Detect Changes**
   - Compare remote etags with local `sync_metadata.remote_etag`
   - Identify new tasks (not in local DB)
   - Identify updated tasks (etag mismatch)
   - Identify deleted tasks (in local but not remote)

3. **Apply Remote Changes**
   - Insert new tasks into local DB
   - Update changed tasks (if not locally modified)
   - Mark conflicts (if both local and remote changed)
   - Delete tasks removed from remote (if not locally modified)

4. **Update Sync Metadata**
   - Store new etags for all synced tasks
   - Update list-level ctags
   - Record sync timestamp

**Push Operation:**
1. **Process Sync Queue**
   - Read all pending operations from `sync_queue` table
   - Sort hierarchically (parents before children)
   - Group operations by type (create, update, delete)

2. **Execute on Remote**
   - Create new tasks on remote backend
   - Update modified tasks on remote
   - Delete tasks marked for deletion

3. **Handle Responses**
   - On success: Clear sync flags, update etags, remove from queue
   - On failure: Increment retry count, calculate backoff delay
   - On max retries: Log error, optionally skip

4. **Hierarchical Ordering**
   - Parent tasks pushed before child tasks
   - Prevents foreign key constraint violations
   - Maintains referential integrity

### User Journey

**Manual Sync:**
```bash
# Trigger full sync (pull then push)
todoat sync

# View what will be synced
todoat sync status

# View pending operations
todoat sync queue
```

**Sync Output:**
```
Syncing with backend: nextcloud
Pull: 15 tasks updated, 3 new tasks, 1 deleted
Push: 5 local changes pushed
Conflicts: 2 (resolved with server_wins)
Sync completed successfully
```

### Prerequisites
- Network connectivity to remote backend
- Valid credentials (see [Credential Management](credential-management.md))
- At least one backend configured with sync enabled

### Technical Details

**Pull Algorithm:**
```go
// Simplified pull logic
for each list in remote:
    remoteTasks = fetchRemoteTasks(list)
    for each remoteTask in remoteTasks:
        localTask = findLocalTask(remoteTask.UID)
        if localTask == nil:
            insertTask(remoteTask)  // New task
        else if remoteTask.ETag != localTask.RemoteETag:
            if localTask.LocallyModified:
                handleConflict(localTask, remoteTask)  // Conflict
            else:
                updateTask(remoteTask)  // Remote update
```

**Push Algorithm:**
```go
// Simplified push logic
queue = getQueuedOperations()
sortedQueue = sortHierarchically(queue)  // Parents first
for each operation in sortedQueue:
    switch operation.Type:
        case "create":
            remoteUID = remote.AddTask(operation.Task)
            updateLocalUID(operation.TaskID, remoteUID)
        case "update":
            remote.UpdateTask(operation.Task)
        case "delete":
            remote.DeleteTask(operation.TaskUID)
    clearSyncFlags(operation.TaskID)
    removeFromQueue(operation.ID)
```

**Retry Logic:**
- Max retries: 5 attempts
- Backoff: Exponential with jitter
- Delays: 1s, 2s, 4s, 8s, 16s

### Related Features
- [Task Management](task-management.md#update-task) - Operations that trigger sync
- [Conflict Resolution](#conflict-resolution) - Handling simultaneous changes
- [Backend System](backend-system.md) - Remote backend implementations

---

## Conflict Resolution

### Purpose
Automatically resolve situations where the same task has been modified both locally and remotely between syncs.

### How It Works

**1. Conflict Detection**
During pull operation:
- Remote task has different etag (indicates remote change)
- Local task has `locally_modified = true` flag (indicates local change)
- Both conditions true = conflict detected

**2. Resolution Strategies**

**Server Wins (Default):**
- Remote version completely replaces local version
- Local changes are discarded
- Simple, predictable behavior

**Local Wins:**
- Local version kept, marked for push
- Remote changes ignored
- Useful when working offline extensively

**Merge:**
- Combine non-conflicting fields
- Example: Merge categories/tags, keep latest timestamps
- Complex but preserves most data

**Keep Both:**
- Duplicate task created
- Original task updated with remote version
- New task created with local changes (appended " (local)" to summary)
- User manually resolves later

**3. Conflict Application**
```go
switch strategy:
    case "server_wins":
        updateTask(remoteTask)
        clearLocallyModified(taskID)
    case "local_wins":
        addToQueue(taskID, "update")
        keepLocalVersion()
    case "merge":
        merged = mergeFields(localTask, remoteTask)
        updateTask(merged)
    case "keep_both":
        updateTask(remoteTask)
        duplicateTask = cloneLocalTask()
        duplicateTask.Summary += " (local)"
        insertTask(duplicateTask)
```

### User Journey

**Automatic Resolution:**
1. User modifies task offline: `todoat MyList update "Task" -s DONE`
2. Same task modified remotely by another device
3. User runs sync: `todoat sync`
4. Conflict detected and auto-resolved based on strategy
5. User sees: "Conflicts: 1 (resolved with server_wins)"

**Manual Resolution (keep_both):**
1. After sync, two versions exist: "Task" and "Task (local)"
2. User reviews both versions
3. User manually merges or chooses one
4. User deletes unwanted version

### Configuration

```yaml
sync:
  enabled: true
  conflict_resolution: server_wins  # Options: server_wins, local_wins, merge, keep_both
```

### Outputs/Results
- Conflict count shown in sync output
- Resolution strategy applied automatically
- For `keep_both`: duplicate tasks created with " (local)" suffix

### Technical Details

**Conflict Detection Code:**
```go
// In SyncManager.Pull()
if remoteTask.ETag != localMeta.RemoteETag && localMeta.LocallyModified {
    // Conflict detected
    resolveConflict(localTask, remoteTask, strategy)
}
```

**Merge Strategy Implementation:**
- Timestamps: Use latest `modified` time
- Status: Use remote (safer default)
- Priority: Use local (user preference)
- Categories: Union of both sets
- Description: Use remote if changed, else keep local

### Related Features
- [Sync Operations](#sync-operations) - Context for when conflicts occur
- [Configuration](configuration.md) - Setting conflict resolution strategy
- [Task Management](task-management.md) - Understanding task modifications

---

## Offline Mode

### Purpose
Allow users to work with tasks without network connectivity, with all changes queued for later synchronization.

### How It Works

**1. Offline Mode Configuration**
```yaml
sync:
  offline_mode: auto  # Options: auto, offline, online
```

- `auto` (default): CLI always uses SQLite cache for instant operations. Sync happens only when you run `todoat sync`.
- `offline`: Same as auto - CLI always uses SQLite cache. Use this to explicitly indicate offline-first preference.
- `online`: CLI uses remote backend directly (bypasses sync architecture). Use this when you need direct remote access without local caching.

**2. Offline Operation Flow**

**User Action:**
```bash
# Add task while offline
todoat MyList add "Buy groceries"
```

**Behind the Scenes:**
1. Task inserted into local cache database
2. Operation added to `sync_queue` table with `operation_type = "create"`
3. Task marked with `locally_modified = true` in `sync_metadata`
4. User receives immediate success confirmation

**3. Queue Persistence**
- All offline changes persist in `sync_queue` table
- Queue survives application restarts
- Queue processed on next successful sync

**4. Coming Back Online**
```bash
# When network restored
todoat sync

# Output:
# Processing queued operations: 5 pending
# Push: 3 created, 2 updated, 0 deleted
# Pull: 0 new, 2 updated, 0 deleted
# Sync completed successfully
```

### User Journey

**Working Offline:**
1. User on plane/train without internet
2. Performs normal task operations:
   - `todoat Work add "Review code"`
   - `todoat Work complete "Bug fix"`
   - `todoat Personal update "Call mom" -p 1`
3. All operations complete instantly (no network wait)
4. Changes stored locally

**Synchronizing Later:**
1. User connects to internet
2. Runs `todoat sync`
3. All queued changes pushed to remote
4. Remote changes pulled to local
5. User now up-to-date across all devices

### Prerequisites
- Sync enabled in configuration
- Initial sync completed while online (to have cached data)
- Local cache database exists

### Outputs/Results
- Instant operation confirmation (no network latency)
- Queued operations visible with `todoat sync queue`
- Sync status shows pending count: `todoat sync status`

### Technical Details

**Queue Table Schema:**
```sql
CREATE TABLE sync_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER,
    task_uid TEXT,
    list_id INTEGER,
    operation_type TEXT,  -- "create", "update", "delete"
    retry_count INTEGER DEFAULT 0,
    last_attempt_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Offline Mode Logic:**
```go
// CLI backend selection based on offline_mode
func createBackendWithSyncFallback(cfg Config, backendName string) Backend {
    offlineMode := cfg.GetOfflineMode()  // defaults to "auto"

    // For "auto" and "offline": CLI always uses SQLite cache
    // Operations are queued in sync_queue for daemon to sync later
    if offlineMode == "auto" || offlineMode == "offline" {
        return createSQLiteBackend(cfg)
    }

    // For "online": CLI uses remote backend directly
    // This bypasses sync architecture for direct remote access
    return createRemoteBackend(backendName)
}
```

**Operation Queueing:**
```go
// When user adds task offline
func (b *SQLiteBackend) AddTask(task Task) error {
    tx := b.db.Begin()

    // Insert task
    tx.Create(&task)

    // Add to sync queue
    tx.Create(&SyncQueueEntry{
        TaskID: task.ID,
        TaskUID: task.UID,
        ListID: task.ListID,
        OperationType: "create",
    })

    // Mark metadata
    tx.Create(&SyncMetadata{
        TaskID: task.ID,
        LocallyModified: true,
    })

    tx.Commit()
}
```

### Related Features
- [Sync Queue System](#sync-queue-system) - Queue management details
- [Task Management](task-management.md) - Operations that work offline
- [Sync Operations](#sync-operations) - Processing the queue when online

---

## Sync Queue System

### Purpose
Manage pending operations that need to be synchronized with remote backends, with automatic retry logic and failure handling.

### How It Works

**1. Queue Entry Creation**
Every local modification creates a queue entry:

```go
type SyncQueueEntry struct {
    ID              int64
    TaskID          int64     // SQLite internal ID (local_id) - always present
    TaskUID         string    // Backend-assigned UID - empty for unsynced tasks
    ListID          int64     // Task's list
    OperationType   string    // "create", "update", "delete"
    RetryCount      int       // Number of retry attempts
    LastAttemptAt   time.Time // Last sync attempt timestamp
    CreatedAt       time.Time // When queued
}
```

**Note on Task Identification:**
- `TaskID` is the SQLite internal ID (corresponds to `local_id` in CLI output)
- `TaskUID` is the backend-assigned unique identifier (corresponds to `uid` in CLI output)
- For newly created tasks (not yet synced), `TaskUID` will be empty
- After successful sync push, `TaskUID` is populated with the remote backend's assigned UID
- CLI flag `--local-id` uses `TaskID`, while `--uid` uses `TaskUID`

**2. Operation Types**

**Create:**
- Triggered by: `todoat MyList add "Task"`
- Action: Create new task on remote
- Success: Store remote UID, clear queue entry

**Update:**
- Triggered by: `todoat MyList update "Task" -s DONE`
- Action: Update existing task on remote
- Success: Update etag, clear queue entry

**Delete:**
- Triggered by: `todoat MyList delete "Task"`
- Action: Delete task from remote
- Success: Remove from local DB, clear queue entry

**3. Processing Order**
Queue processed hierarchically during sync:
1. Parent tasks before children (respects `parent_uid` relationships)
2. Creates before updates (ensure task exists)
3. Deletes last (remove dependencies first)

**4. Retry Mechanism**

**On Failure:**
```go
entry.RetryCount++
entry.LastAttemptAt = time.Now()
backoff := calculateBackoff(entry.RetryCount)  // Exponential: 1s, 2s, 4s, 8s, 16s

if entry.RetryCount >= maxRetries {  // maxRetries = 5
    logError("Max retries exceeded", entry)
    // Optionally: move to failed_operations table
} else {
    // Keep in queue for next sync
    updateQueueEntry(entry)
}
```

**5. Queue Inspection**

View pending operations:
```bash
todoat sync queue

# Output:
# Pending Operations: 5
#
# ID  | Type   | Task Summary        | Retries | Last Attempt
# ----+--------+---------------------+---------+-------------
# 1   | create | Buy groceries       | 0       | -
# 2   | update | Review code         | 1       | 2m ago
# 3   | update | Call mom            | 0       | -
# 4   | delete | Old task            | 2       | 5m ago
# 5   | create | Team meeting notes  | 0       | -
```

### User Journey

**Normal Flow:**
1. User performs operations while offline/online
2. Operations queued automatically
3. User runs `todoat sync`
4. Queue processed, entries removed on success

**Retry Flow:**
1. User syncs but network flaky
2. Some operations fail, retry count incremented
3. User runs sync again later
4. Failed operations retried with backoff
5. Eventually succeed or reach max retries

**Failure Investigation:**
```bash
# Check queue status
todoat sync queue

# See specific failures
todoat sync status --verbose

# Manual resolution if needed
# (e.g., fix network issue, update credentials, then sync again)
```

### Prerequisites
- Sync enabled and cache database exists
- Operations performed that modify tasks

### Outputs/Results
- Queue entries created automatically
- Retry counts visible in queue listing
- Success removes entries from queue
- Max retries triggers error logging

### Technical Details

**Hierarchical Sorting Algorithm:**
```go
func sortQueueHierarchically(entries []SyncQueueEntry) []SyncQueueEntry {
    // Build parent-child map
    childMap := make(map[string][]SyncQueueEntry)
    for _, entry := range entries {
        parentUID := getParentUID(entry.TaskID)
        if parentUID != "" {
            childMap[parentUID] = append(childMap[parentUID], entry)
        }
    }

    // Process roots first, then children recursively
    var sorted []SyncQueueEntry
    for _, entry := range entries {
        if !hasParentInQueue(entry) {
            sorted = append(sorted, entry)
            sorted = append(sorted, getChildrenRecursive(entry, childMap)...)
        }
    }
    return sorted
}
```

**Backoff Calculation:**
```go
func calculateBackoff(retryCount int) time.Duration {
    base := time.Second
    backoff := base * time.Duration(math.Pow(2, float64(retryCount)))
    jitter := time.Duration(rand.Int63n(int64(backoff / 4)))
    return backoff + jitter  // Add jitter to prevent thundering herd
}
```

### Related Features
- [Offline Mode](#offline-mode) - Primary use case for queue
- [Sync Operations](#sync-operations) - Queue processing
- [Task Management](task-management.md) - Operations that create queue entries

---

### Sync Notifications

When background sync is enabled, the [Notification Manager](notification-manager.md)
can alert users to:
- Sync completion (optional)
- Sync failures and errors
- Conflict detection requiring user attention

Configure notifications via the `notification` section in config.yaml.

---

## Manual Sync Workflow

### Purpose
Provide explicit user control over synchronization timing, replacing automatic background sync for the new multi-remote architecture.

### How It Works

**Current Status:**
- **Auto-sync temporarily disabled** during architecture redesign
- All sync operations are manual via CLI commands
- Operations still queued automatically
- Queue persists between application runs

**Manual Sync Commands:**

**Full Sync:**
```bash
todoat sync
```
- Pulls changes from all enabled remote backends
- Pushes all queued local changes
- Processes all cached remotes in sequence

**Sync Status:**
```bash
todoat sync status
```
Output:
```
Sync Status:

Backend: nextcloud (cache: nextcloud.db)
  Last Sync: 5 minutes ago
  Pending Operations: 3 (2 updates, 1 create)
  Local Tasks: 47
  Remote Tasks: 45
  Status: Out of sync

Backend: todoist (cache: todoist.db)
  Last Sync: 2 hours ago
  Pending Operations: 0
  Local Tasks: 23
  Remote Tasks: 23
  Status: In sync
```

**View Queue:**
```bash
todoat sync queue
```
Shows all pending operations across all backends.

### User Journey

**Daily Workflow:**
1. **Morning**: Sync to get latest changes
   ```bash
   todoat sync
   ```

2. **Throughout Day**: Work normally
   ```bash
   todoat Work add "New task"
   todoat Personal complete "Old task"
   # Changes queued automatically
   ```

3. **Before Meetings**: Quick sync to share updates
   ```bash
   todoat sync
   ```

4. **Evening**: Final sync before closing laptop
   ```bash
   todoat sync status  # Check what will sync
   todoat sync         # Push all changes
   ```

**Troubleshooting Workflow:**
1. Check sync status for issues:
   ```bash
   todoat sync status --verbose
   ```

2. Inspect failed operations:
   ```bash
   todoat sync queue
   ```

3. Fix underlying issue (network, credentials, etc.)

4. Retry sync:
   ```bash
   todoat sync
   ```

### Prerequisites
- Sync enabled in configuration
- At least one remote backend configured
- Valid credentials for remote backends

### Outputs/Results

**Successful Sync:**
```
Syncing nextcloud...
  Pull: 5 new tasks, 3 updated, 1 deleted
  Push: 2 created, 1 updated, 0 deleted
  Conflicts: 0

Syncing todoist...
  Pull: 0 new tasks, 1 updated, 0 deleted
  Push: 1 created, 0 updated, 0 deleted
  Conflicts: 0

Sync completed successfully
```

**Sync with Errors:**
```
Syncing nextcloud...
  Error: Authentication failed
  Skipping push operations

Syncing todoist...
  Pull: 0 new tasks, 0 updated, 0 deleted
  Push: Failed (network timeout)
  Retrying in next sync

Sync completed with errors (see above)
```

### Technical Details

**Future Auto-Sync Plans:**
- Background sync daemon for automatic synchronization
- Configurable interval (currently would be `sync.sync_interval`)
- File watcher for real-time sync triggers
- Smart sync timing (avoid sync during active editing)

**Current Manual Implementation:**
```go
func syncCommand() error {
    backends := getEnabledBackends()

    for _, backend := range backends {
        if !backend.SyncEnabled {
            continue
        }

        cache := openCacheDB(backend.Name)
        syncMgr := NewSyncManager(cache, backend)

        // Pull first
        pullErr := syncMgr.Pull()
        if pullErr != nil {
            logError("Pull failed", backend.Name, pullErr)
            continue
        }

        // Then push
        pushErr := syncMgr.Push()
        if pushErr != nil {
            logError("Push failed", backend.Name, pushErr)
        }

        reportSyncResults(syncMgr.Stats)
    }
}
```

### Related Features
- [Sync Operations](#sync-operations) - Details of pull/push
- [Configuration](configuration.md) - Sync settings
- [Backend System](backend-system.md) - Remote backends that sync

---

## Configuration

### Purpose
Configure synchronization behavior, conflict resolution, offline mode, and cache settings.

### How It Works

**Global Sync Configuration:**
```yaml
sync:
  enabled: true                          # Enable/disable sync globally
  auto_sync_after_operation: false       # Sync immediately after add/update/delete
  local_backend: sqlite                  # Cache implementation (only sqlite supported)
  conflict_resolution: server_wins       # Conflict strategy
  offline_mode: auto                     # auto, offline, online
```

**Per-Backend Sync Settings:**
```yaml
backends:
  nextcloud:
    type: nextcloud
    enabled: true
    url: nextcloud://user:pass@host
    # Sync enabled by default when global sync.enabled = true

  todoist:
    type: todoist
    enabled: true
    sync:
      enabled: false  # Opt-out of caching this backend
```

**Configuration Options:**

**sync.enabled:**
- `true`: Enable automatic caching for all remote backends
- `false`: Use backends directly without caching

**sync.local_backend:**
- `sqlite`: Use SQLite for cache (only option currently)
- Future: `postgres`, `mysql`, etc.

**sync.conflict_resolution:**
- `server_wins`: Remote changes override local (default)
- `local_wins`: Local changes override remote
- `merge`: Intelligently combine changes
- `keep_both`: Create duplicate tasks for manual resolution

**sync.auto_sync_after_operation:**
- `true`: Sync immediately after each operation (add, update, delete)
- `false`: Queue operations for manual sync (default)
- When enabled, eliminates need to run `todoat sync` manually after each operation

**sync.offline_mode:**
- `auto`: Detect network automatically
- `offline`: Force offline, never attempt remote sync
- `online`: Require network, fail if unavailable

### User Journey

**Initial Setup:**
1. Edit config file:
   ```bash
   vim ~/.config/todoat/config.yaml
   ```

2. Enable sync:
   ```yaml
   sync:
     enabled: true
     conflict_resolution: server_wins
   ```

3. Save and perform initial sync:
   ```bash
   todoat sync
   ```

**Changing Conflict Resolution:**
1. Edit config to try `merge` strategy:
   ```yaml
   sync:
     conflict_resolution: merge
   ```

2. Next sync uses new strategy automatically

**Disabling Cache for Specific Backend:**
1. Add opt-out to backend config:
   ```yaml
   backends:
     nextcloud:
       sync:
         enabled: false
   ```

2. Backend used directly without cache

### Prerequisites
- Config file exists at `~/.config/todoat/config.yaml`
- Valid backend configuration (see [Backend System](backend-system.md))

### Outputs/Results
- Cache databases created/updated based on settings
- Sync behavior changes according to configuration
- `todoat sync status` reflects current config

### Technical Details

**Configuration Loading:**
```go
type SyncConfig struct {
    Enabled                 bool   `yaml:"enabled"`
    AutoSyncAfterOperation  bool   `yaml:"auto_sync_after_operation"`
    LocalBackend            string `yaml:"local_backend"`
    ConflictResolution      string `yaml:"conflict_resolution"`
    OfflineMode             string `yaml:"offline_mode"`
}

func LoadConfig() (*Config, error) {
    // Load from XDG config path
    // Validate sync settings
    // Apply defaults
}
```

**Cache Path Resolution:**
```go
func getCachePath(backendName string) string {
    xdgData := os.Getenv("XDG_DATA_HOME")
    if xdgData == "" {
        xdgData = filepath.Join(os.Getenv("HOME"), ".local/share")
    }
    return filepath.Join(xdgData, "todoat", "caches", backendName+".db")
}
```

### Related Features
- [Configuration](configuration.md) - Complete config documentation
- [Backend System](backend-system.md) - Backend-specific settings
- [Conflict Resolution](#conflict-resolution) - Strategy details

---

## Database Locations

### Purpose
Understand where todoat stores sync-related data to enable proper backup, troubleshooting, and data management.

### Database Files

todoat uses **separate database files** for different purposes:

| Database | Location | Contents |
|----------|----------|----------|
| **Sync Queue** | `~/.todoat/todoat.db` | Pending sync operations, retry state |
| **Backend Caches** | `~/.local/share/todoat/caches/[backend].db` | Cached tasks per remote backend |
| **Default SQLite Backend** | `~/.local/share/todoat/tasks.db` | Tasks for the local sqlite backend |

### Sync Queue Database (`~/.todoat/todoat.db`)

The sync queue database is **separate** from the task cache databases. It stores:

- `sync_queue` table: Pending create/update/delete operations
- `sync_conflicts` table: Unresolved sync conflicts

**Important**: This database persists across cache resets. If you delete a backend cache database, the sync queue may still contain operations referencing tasks that no longer exist locally.

### Backend Cache Databases

Each remote backend gets its own cache database at `~/.local/share/todoat/caches/`:

```
~/.local/share/todoat/caches/
├── nextcloud.db    # Nextcloud backend cache
├── todoist.db      # Todoist backend cache
└── caldav.db       # CalDAV backend cache
```

These databases contain:
- Cached tasks from the remote backend
- Sync metadata (etags, timestamps)
- List sync metadata (ctags, sync tokens)

### Implications for Data Management

**Resetting sync state completely:**
```bash
# Clear the sync queue
todoat sync queue clear

# Or manually delete sync queue database
rm ~/.todoat/todoat.db

# Delete backend caches if needed
rm ~/.local/share/todoat/caches/*.db
```

**Backup considerations:**
- Backup both `~/.todoat/` and `~/.local/share/todoat/` directories
- The sync queue database contains pending operations that would be lost

**Troubleshooting orphaned operations:**
If you see sync operations for tasks that don't exist:
1. Check `todoat sync queue` for pending operations
2. Run `todoat sync queue clear` to remove orphaned entries
3. Re-sync with `todoat sync` to restore consistent state

---

## Database Schema

### Purpose
Store tasks, sync metadata, and operation queue in structured SQLite databases for each cached backend.

### How It Works

**Schema Tables:**

**1. tasks**
```sql
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,  -- SQLite internal ID (local_id in CLI)
    uid TEXT UNIQUE,                       -- Backend-assigned UID (null until synced)
    list_id INTEGER NOT NULL,
    summary TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'TODO',  -- Internal status (TODO/DONE/IN-PROGRESS/CANCELLED)
    priority INTEGER DEFAULT 0,
    due_date TIMESTAMP,
    start_date TIMESTAMP,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed TIMESTAMP,
    parent_uid TEXT,
    categories TEXT,  -- JSON array
    FOREIGN KEY (parent_uid) REFERENCES tasks(uid)
);
```

Note: `id` is used as `local_id` in CLI output and with `--local-id` flag. `uid` is the backend-assigned identifier used with `--uid` flag. For newly created tasks (not yet synced), `uid` will be `NULL` until sync completes.

**2. sync_metadata**
```sql
CREATE TABLE sync_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER UNIQUE NOT NULL,
    remote_etag TEXT,
    local_etag TEXT,
    locally_modified BOOLEAN DEFAULT FALSE,
    locally_deleted BOOLEAN DEFAULT FALSE,
    last_synced TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);
```

**3. list_sync_metadata**
```sql
CREATE TABLE list_sync_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    list_id INTEGER UNIQUE NOT NULL,
    remote_ctag TEXT,
    sync_token TEXT,
    last_synced TIMESTAMP
);
```

**4. sync_queue**
```sql
CREATE TABLE sync_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER,             -- SQLite internal ID (local_id), always present
    task_uid TEXT,               -- Backend-assigned UID, empty for unsynced tasks
    list_id INTEGER,
    operation_type TEXT NOT NULL,  -- create, update, delete
    retry_count INTEGER DEFAULT 0,
    last_attempt_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);
```

Note: `task_id` corresponds to the CLI's `--local-id` flag, and `task_uid` corresponds to `--uid`. For create operations, `task_uid` will be empty until the sync completes and the remote backend assigns a UID.

**5. schema_version**
```sql
CREATE TABLE schema_version (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Indexes for Performance:**
```sql
CREATE INDEX idx_tasks_list_id ON tasks(list_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
CREATE INDEX idx_tasks_parent_uid ON tasks(parent_uid);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_sync_metadata_task_id ON sync_metadata(task_id);
CREATE INDEX idx_sync_metadata_modified ON sync_metadata(locally_modified);
CREATE INDEX idx_sync_queue_operation ON sync_queue(operation_type);
CREATE INDEX idx_list_sync_list_id ON list_sync_metadata(list_id);
```

### User Journey

Users don't directly interact with schema, but it supports:
1. Fast task queries (via indexes)
2. Reliable sync state tracking
3. Hierarchical task relationships (parent_uid)
4. Operation queuing and retry

### Technical Details

**Schema Migrations:**
```go
func InitDatabase(db *gorm.DB) error {
    // Check current version
    var version int
    db.Raw("SELECT MAX(version) FROM schema_version").Scan(&version)

    // Apply migrations
    for _, migration := range migrations {
        if migration.Version > version {
            db.Exec(migration.SQL)
            db.Exec("INSERT INTO schema_version (version) VALUES (?)", migration.Version)
        }
    }
}
```

**Foreign Key Enforcement:**
- `parent_uid` references `tasks(uid)` for subtask hierarchy
- Cascading deletes prevent orphaned records
- `sync_queue` entries deleted when tasks deleted

### Related Features
- [Sync Operations](#sync-operations) - Uses all tables during sync
- [Task Management](task-management.md) - CRUD on tasks table
- [Subtasks and Hierarchy](subtasks-hierarchy.md) - Uses parent_uid

---

## Performance Characteristics

### Purpose
Understand synchronization performance to set expectations and optimize usage patterns.

### How It Works

**Benchmark Results:**

**Large Dataset (1000 tasks):**
- Initial sync (full pull): ~15 seconds
- Incremental sync (50 changes): ~2 seconds
- Full push (all queued): ~10 seconds
- Total sync cycle: <30 seconds

**Small Dataset (100 tasks):**
- Initial sync: ~2 seconds
- Incremental sync: <1 second
- Full push: ~1 second

**Factors Affecting Performance:**

**Network Latency:**
- CalDAV backends: ~100-200ms per request
- Batch operations reduce round trips
- Connection pooling reuses TCP connections

**Database Operations:**
- Indexed queries: <10ms for typical operations
- Bulk inserts: ~1000 tasks/second
- Transaction batching improves throughput

**Hierarchical Sorting:**
- Overhead: ~5% for <100 tasks
- Overhead: ~15% for >1000 tasks
- Required for correct parent-child ordering

### User Journey

**Initial Sync (Heavy):**
User enables sync for first time with large Nextcloud calendar:
```bash
todoat sync
# Syncing nextcloud...
# Pull: 1000 new tasks (15s)
# Building local indexes...
# Sync completed in 18s
```

**Daily Sync (Light):**
User syncs after making few changes:
```bash
todoat sync
# Syncing nextcloud...
# Pull: 2 updated, 0 new (1s)
# Push: 3 updates (1s)
# Sync completed in 2s
```

### Technical Details

**Optimization Strategies:**

**1. Batch Operations:**
```go
// Instead of: N individual INSERTs
for _, task := range tasks {
    db.Create(&task)  // Slow
}

// Use: Single batch INSERT
db.CreateInBatches(tasks, 100)  // Fast
```

**2. Connection Pooling:**
```go
db.SetMaxOpenConns(10)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(time.Hour)
```

**3. Index Usage:**
```go
// Queries optimized with indexes
db.Where("list_id = ? AND status = ?", listID, "TODO")  // Uses idx_tasks_list_id
db.Where("due_date < ?", time.Now())  // Uses idx_tasks_due_date
```

**4. Transaction Batching:**
```go
tx := db.Begin()
for _, task := range tasks {
    tx.Create(&task)
}
tx.Commit()  // Single disk write
```

### Related Features
- [Sync Operations](#sync-operations) - Operations being benchmarked
- [Database Schema](#database-schema) - Indexes that enable performance
- [Backend System](backend-system.md) - Backend-specific performance

---

## Related Documentation

- [Backend System](backend-system.md) - Remote backend implementations
- [Task Management](task-management.md) - Operations that trigger sync
- [Credential Management](credential-management.md) - Authentication for remotes
- [Configuration](configuration.md) - Sync configuration settings
- [Subtasks and Hierarchy](subtasks-hierarchy.md) - Hierarchical sync considerations
- [Features Overview](features-overview.md) - Complete feature catalog

---

## Additional Resources

**Source Code:**
- `backend/sqlite/` - SQLite backend implementation (cache database operations)
- `backend/sync/` - Sync-related tests
- `cmd/todoat/cmd/todoat.go` - CLI implementation including sync command

> **Note**: Sync logic is implemented in `cmd/todoat/cmd/todoat.go` rather than a separate `backend/syncManager.go` file.

**Tests:**
- `backend/sync/daemon_test.go` - Daemon tests
- `backend/sync/offline_mode_test.go` - Offline mode tests
- `backend/sync/sync_test.go` - Sync tests

**Documentation:**
- [../../CLAUDE.md](../../CLAUDE.md) - Developer guidance
