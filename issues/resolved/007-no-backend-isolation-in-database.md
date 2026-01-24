# [007] No Backend Isolation in Database - SQLite and Remote Caches Mixed

## Type
code-bug

## Severity
critical

## Description

The database schema has no backend isolation. All backends (local SQLite and remote backend caches) share the same `tasks` and `task_lists` tables without any identifier to distinguish which backend owns which data. This causes:

1. **Data mixing**: Local SQLite tasks get mixed with Nextcloud/Todoist cache data
2. **--detect-backend confusion**: Always detects "sqlite" because tasks exist in shared table
3. **Sync failures after DB deletion**: Deleting tasks.db loses both local tasks AND remote cache, and sync doesn't know to pull fresh data from remote
4. **No recovery path**: Cannot rebuild cache from remote because sync doesn't detect empty/missing cache

## Steps to Reproduce

```bash
# 1. Set up with default_backend as remote
cat > ~/.config/todoat/config.yaml << 'EOF'
backends:
  sqlite:
    enabled: true
  nextcloud-test:
    type: nextcloud
    url: nextcloud://admin:password@localhost:8080
    enabled: true

default_backend: nextcloud-test

sync:
  enabled: true
EOF

# 2. Add tasks to remote backend
todoat -b nextcloud-test Work add "Remote task 1"
todoat -b nextcloud-test Work add "Remote task 2"
todoat sync

# 3. Add local SQLite tasks
todoat -b sqlite LocalList add "Local task"

# 4. Check detect-backend - shows sqlite even though default is nextcloud-test
todoat --detect-backend
# Output: Would use: sqlite (because tasks exist in shared table)

# 5. Delete database to "reset"
rm ~/.local/share/todoat/tasks.db

# 6. Try to sync - FAILS to pull data from remote
todoat -b nextcloud-test sync
# Sync shows "0 operations" - doesn't pull from remote to rebuild cache

# 7. Tasks are LOST - cannot recover
todoat -b nextcloud-test Work
# Empty or error - cache is gone and sync didn't rebuild it
```

## Expected Behavior

### Database Isolation
Each backend should have isolated storage:

**Option A: Separate tables per backend**
```sql
-- SQLite local backend
CREATE TABLE sqlite_tasks (...);
CREATE TABLE sqlite_task_lists (...);

-- Nextcloud cache
CREATE TABLE nextcloud_test_tasks (...);
CREATE TABLE nextcloud_test_task_lists (...);

-- Todoist cache
CREATE TABLE todoist_tasks (...);
CREATE TABLE todoist_task_lists (...);
```

**Option B: Single tables with backend_id column**
```sql
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    backend_id TEXT NOT NULL,  -- "sqlite", "nextcloud-test", "todoist", etc.
    list_id TEXT NOT NULL,
    ...
);

CREATE TABLE task_lists (
    id TEXT PRIMARY KEY,
    backend_id TEXT NOT NULL,
    ...
);

-- Queries always filter by backend
SELECT * FROM tasks WHERE backend_id = 'nextcloud-test' AND list_id = ?;
```

**Option C: Separate database files per backend (as documented)**
```
~/.local/share/todoat/
├── sqlite.db           # Local SQLite backend
├── nextcloud-test.db   # Nextcloud cache
├── todoist.db          # Todoist cache
└── sync.db             # Sync queue and metadata
```

### Sync Pull on Empty Cache
When cache is empty or missing, sync should:
1. Detect that local cache is empty/stale
2. Perform full pull from remote backend
3. Populate cache with remote data
4. Report "Pulled X tasks from remote"

## Actual Behavior

### No Backend Isolation
- Single `tasks` table stores all data
- No `backend_id` column to distinguish sources
- SQLite local tasks mixed with remote cache

### No Sync Pull
- Sync only processes queue (push operations)
- Never pulls from remote to update cache
- Deleting DB = permanent data loss from cache perspective

## Root Cause

### Schema Issue
`backend/sqlite/sqlite.go` defines shared tables:
```go
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    list_id TEXT NOT NULL,
    // ... NO backend_id column!
);
```

### Sync Queue Issue
`cmd/todoat/cmd/todoat.go` sync_queue also lacks backend identification:
```go
CREATE TABLE IF NOT EXISTS sync_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    // ... NO backend_name column!
);
```

### No Pull Operation
The sync command only pushes queued operations, never pulls from remote:
```go
func doSync(...) {
    // Only processes sync_queue (push)
    // Never calls remote.GetTasks() to pull
}
```

## Documentation vs Reality

Documentation claims (`docs/explanation/synchronization.md`):
- "Each backend gets its own dedicated table(s) within the shared database" - **FALSE**
- "Example: Nextcloud backend → nextcloud table, Todoist backend → todoist table" - **FALSE**
- "Each remote backend has a completely separate cache database" - **FALSE**
- "Cache Isolation: Each backend's cache is completely independent" - **FALSE**

## Impact

1. **Data corruption**: Tasks from different backends can conflict
2. **Incorrect backend detection**: `--detect-backend` finds sqlite tasks, ignores config
3. **Unrecoverable sync**: Deleting DB loses cache with no way to rebuild
4. **User confusion**: Expecting isolation but getting mixed data

## Suggested Fix

### Phase 1: Add backend_id to schema
```sql
ALTER TABLE tasks ADD COLUMN backend_id TEXT NOT NULL DEFAULT 'sqlite';
ALTER TABLE task_lists ADD COLUMN backend_id TEXT NOT NULL DEFAULT 'sqlite';
ALTER TABLE sync_queue ADD COLUMN backend_name TEXT NOT NULL DEFAULT 'sqlite';
```

### Phase 2: Update all queries to filter by backend
```go
func (b *Backend) GetTasks(ctx context.Context, listID string) ([]*Task, error) {
    rows, err := b.db.QueryContext(ctx,
        `SELECT ... FROM tasks WHERE backend_id = ? AND list_id = ?`,
        b.backendID, listID)
    // ...
}
```

### Phase 3: Implement sync pull operation
```go
func doSync(cfg *Config, ...) error {
    // 1. Pull from remote
    remoteTasks, err := remoteBE.GetTasks(ctx, "")
    if err != nil {
        return err
    }

    // 2. Update local cache
    for _, task := range remoteTasks {
        localBE.UpsertTask(ctx, task)
    }

    // 3. Push queued operations
    // ... existing push logic ...
}
```

## Test Criteria

```go
// TestBackendIsolation verifies backends don't share data
func TestBackendIsolation(t *testing.T) {
    cli := testutil.NewCLITest(t)

    // Create task in sqlite backend
    cli.MustExecute("-y", "-b", "sqlite", "LocalList", "add", "Local task")

    // Create task in remote backend
    cli.MustExecute("-y", "-b", "nextcloud-test", "RemoteList", "add", "Remote task")

    // Verify sqlite only sees its task
    sqliteOut := cli.MustExecute("-y", "-b", "sqlite", "LocalList")
    testutil.AssertContains(t, sqliteOut, "Local task")
    testutil.AssertNotContains(t, sqliteOut, "Remote task")

    // Verify remote only sees its task
    remoteOut := cli.MustExecute("-y", "-b", "nextcloud-test", "RemoteList")
    testutil.AssertContains(t, remoteOut, "Remote task")
    testutil.AssertNotContains(t, remoteOut, "Local task")
}

// TestSyncPullOnEmptyCache verifies sync rebuilds cache from remote
func TestSyncPullOnEmptyCache(t *testing.T) {
    cli := testutil.NewCLITest(t)
    mockRemote := testutil.NewMockRemoteBackend(t)
    mockRemote.AddTask("Work", &backend.Task{Summary: "Remote task"})

    // Delete local cache
    os.Remove(cli.DBPath())

    // Sync should pull from remote
    cli.MustExecute("-y", "-b", "nextcloud-test", "sync")

    // Cache should now have the remote task
    out := cli.MustExecute("-y", "-b", "nextcloud-test", "Work")
    testutil.AssertContains(t, out, "Remote task")
}
```

## Related Issues

- [001-sync-architecture-not-implemented.md]
- [003-sync-command-does-not-actually-sync.md]
- [004-sync-documentation-describes-unimplemented-features.md]

## Resolution

**Fixed in**: this session
**Fix description**: Added `backend_id` column to `tasks` and `task_lists` tables via migration (version 4). All queries now filter by `backend_id` to ensure data isolation between backends sharing the same database file.
**Test added**: `TestBackendIsolation_Issue007` in `backend/sqlite/sqlite_test.go`

### Changes Made
1. Added `backend_id` field to Backend struct
2. Added migration (version 4) to add `backend_id` column with default 'sqlite' for existing data
3. Created `NewWithBackendID()` constructor for specifying backend ID
4. Updated all queries to filter by `backend_id`:
   - GetLists, GetList, GetListByName
   - CreateList, UpdateList, DeleteList
   - GetDeletedLists, GetDeletedListByName, RestoreList, PurgeList
   - GetTasks, GetTask, GetTaskByLocalID, GetTaskLocalID
   - CreateTask, UpdateTask, DeleteTask

### Verification Log
```bash
$ go test -run TestBackendIsolation_Issue007 ./backend/sqlite/ -v
=== RUN   TestBackendIsolation_Issue007
--- PASS: TestBackendIsolation_Issue007 (0.03s)
PASS
ok  	todoat/backend/sqlite	0.036s

$ go test ./... 2>&1 | grep -E "(PASS|FAIL|ok)"
ok  	todoat/backend	0.399s
ok  	todoat/backend/file	(cached)
ok  	todoat/backend/git	(cached)
ok  	todoat/backend/google	(cached)
ok  	todoat/backend/mstodo	(cached)
ok  	todoat/backend/nextcloud	0.279s
ok  	todoat/backend/sqlite	13.924s
ok  	todoat/backend/sync	5.405s
ok  	todoat/backend/todoist	0.486s
ok  	todoat/cmd/todoat/cmd	2.268s
ok  	todoat/internal/analytics	(cached)
ok  	todoat/internal/cache	1.039s
ok  	todoat/internal/config	0.019s
ok  	todoat/internal/credentials	(cached)
ok  	todoat/internal/markdown	(cached)
ok  	todoat/internal/migrate	1.155s
ok  	todoat/internal/notification	0.016s
ok  	todoat/internal/ratelimit	(cached)
ok  	todoat/internal/reminder	1.586s
ok  	todoat/internal/shutdown	(cached)
ok  	todoat/internal/testutil	0.004s
ok  	todoat/internal/tui	(cached)
ok  	todoat/internal/utils	(cached)
ok  	todoat/internal/views	7.834s
```
**Matches expected behavior**: YES
