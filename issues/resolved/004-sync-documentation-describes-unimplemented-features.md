# [004] Sync Documentation Describes Unimplemented Features

## Type
documentation

## Severity
high

## Description

The documentation in `docs/explanation/synchronization.md` describes a comprehensive sync architecture that does not exist in the codebase. This is misleading because:

1. Users read the docs and expect sync to work as described
2. AI assistants (like Ralph) use the docs as source of truth and think sync is implemented
3. The docs prevented identification of the real issues because they describe working features

## Documentation Claims vs Reality

### 1. SyncManager Component

**Documentation claims** (`docs/explanation/synchronization.md` lines 97-105):
```
**SyncManager Component** (`backend/syncManager.go`):
- **Primary Role**: Manages bidirectional synchronization between local cache and remote backends
- **Pull Operations**: Fetches changes from remote backend and updates local cache
- **Push Operations**: Processes sync queue and applies queued operations to remote backend
- **Conflict Resolution**: Detects and resolves conflicts based on configured strategy
- **Retry Logic**: Implements exponential backoff for failed sync operations (max 5 retries)
```

**Reality**:
- `backend/syncManager.go` does not exist
- `SyncManager` exists in `cmd/todoat/cmd/todoat.go` but only manages the queue
- No `Pull()` method exists
- No `Push()` method exists
- No conflict resolution during sync
- No retry logic

### 2. Data Flow Diagram

**Documentation claims**:
```
User → CLI Command → Sync Manager →  Database (SQLite)
                           ↓              ↓
                    Sync Operations  sync_queue table
                           ↓              ↓
                    Remote Backend  Manual Sync Push
```

**Reality**:
- CLI either uses remote directly (with -b flag) or SQLite directly (when sync enabled)
- Sync operations are never executed
- Queue is cleared without pushing to remote

### 3. SQLiteBackend Sync Methods

**Documentation claims** (lines 107-113):
```
**SQLiteBackend Sync Methods** (used by SyncManager):
- `MarkLocallyModified()` - Flag task as changed locally
- `MarkLocallyDeleted()` - Flag task for deletion on remote
- `ClearSyncFlags()` - Reset sync state after successful push
- `UpdateSyncMetadata()` - Store etags and sync timestamps
- `GetPendingSyncOperations()` - Retrieve queued operations
- `AddToSyncQueue()` - Queue operation for later synchronization
```

**Reality**: These methods may exist but are never called by any sync logic because the sync logic doesn't exist.

### 4. Pull/Push Algorithms

**Documentation provides detailed pseudocode** (lines 284-316):
```go
// Simplified pull logic
for each list in remote:
    remoteTasks = fetchRemoteTasks(list)
    for each remoteTask in remoteTasks:
        // ... conflict detection, merge, etc.

// Simplified push logic
queue = getQueuedOperations()
sortedQueue = sortHierarchically(queue)
for each operation in sortedQueue:
    switch operation.Type:
        case "create": remote.AddTask(...)
        case "update": remote.UpdateTask(...)
        case "delete": remote.DeleteTask(...)
```

**Reality**: No such code exists. The sync command just clears the queue.

### 5. Retry Logic

**Documentation claims** (lines 318-322):
```
**Retry Logic:**
- Max retries: 5 attempts
- Backoff: Exponential with jitter
- Delays: 1s, 2s, 4s, 8s, 16s
```

**Reality**: No retry logic exists. Operations aren't even attempted once.

## Why This Matters

1. **Users trust the docs**: They configure sync expecting it to work
2. **Data loss**: Users think tasks are synced but they're silently discarded
3. **AI confusion**: Ralph used these docs as source of truth and couldn't identify the real bugs
4. **Debugging difficulty**: When sync "works" but tasks don't appear on remote, users assume network issues rather than missing implementation

## Root Cause

The documentation appears to be a design document that was written before or during implementation, describing intended behavior. The implementation was never completed, but the documentation was never updated to reflect reality.

Specifically, the "Manual Sync Workflow" section (lines 758-934) acknowledges auto-sync is "temporarily disabled" but doesn't mention that manual sync is also non-functional.

## Recommended Fix

### Option 1: Mark as Design Doc (Quick Fix)
Add a prominent warning at the top of the document:

```markdown
> ⚠️ **Implementation Status**: This document describes the PLANNED sync architecture.
> Currently, only the sync queue is implemented. The actual sync operations (push/pull)
> are not yet functional. See roadmap item [073] for implementation status.
```

### Option 2: Rewrite to Match Reality
Update the documentation to accurately describe what currently works:
- Queue operations are tracked in sync_queue table
- `todoat sync` connects to remote but does not actually sync
- Manual `-b backend` flag bypasses cache and uses remote directly

### Option 3: Implement the Features
Make the code match the documentation. This is the ideal solution but requires significant work.

## Related Issues

- [001-sync-architecture-not-implemented.md]
- [002-default-backend-ignored-when-sync-enabled.md]
- [003-sync-command-does-not-actually-sync.md]

## Files to Update

- `docs/explanation/synchronization.md` - Main sync documentation
- `docs/explanation/features-overview.md` - May reference sync features
- `docs/how-to/sync.md` - Usage guide (if exists)

## Resolution

**Fixed in**: this session
**Fix description**: Added prominent implementation status warning at top of synchronization.md, and corrected the Additional Resources section to reference actual existing files.

### Changes Made

1. Added warning block at the top of `docs/explanation/synchronization.md` explaining:
   - What features ARE implemented (queue tracking, fallback behavior, manual sync command)
   - What features are NOT implemented (Pull/Push methods, actual sync, conflict resolution, retry logic)
   - References to related issues [001], [002], [003]

2. Updated Additional Resources section to:
   - Reference actual existing code paths (`backend/sqlite/`, `backend/sync/`, `cmd/todoat/cmd/todoat.go`)
   - Added note that `backend/syncManager.go` doesn't exist
   - Updated test file references to actual existing files
   - Removed reference to non-existent `../how-to/sync.md`

### Verification

Documentation now accurately reflects implementation status at the top of the file:

```markdown
> **⚠️ Implementation Status**: This document describes the **planned** sync architecture. Currently, only the following features are implemented:
> - Sync queue tracking (`sync_queue` table stores pending operations)
> - Fallback behavior (CLI falls back to SQLite when remote is unavailable)
> - Manual sync command (`todoat sync`) - connects to remote but **does not execute actual push/pull operations**
>
> The following documented features are **NOT YET IMPLEMENTED**:
> - `SyncManager.Pull()` and `SyncManager.Push()` methods
> - Actual bidirectional sync (operations are queued but not executed)
> - Conflict resolution during sync
> - Retry logic with exponential backoff
> - Hierarchical sync ordering
```

**Matches expected behavior**: YES - Users will now see upfront that the documentation describes planned, not implemented, features.
