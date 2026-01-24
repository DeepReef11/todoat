# [006] Operations Not Auto-Synced After Execution

## Type
code-bug

## Severity
medium

## Description

When using a remote backend with `-b myremotebackend`, operations like `add`, `update`, `delete` complete but don't automatically sync with the remote. Users must manually run `todoat sync` after each operation to push changes to the remote backend.

This is inconvenient because:
1. Users explicitly specify a remote backend expecting operations to reach the remote
2. Without auto-sync, users must remember to run `todoat sync` after every operation
3. The daemon (which should handle this) doesn't actually call sync (see issue #005)

## Steps to Reproduce

```bash
# 1. Configure remote backend
cat > ~/.config/todoat/config.yaml << 'EOF'
backends:
  nextcloud-test:
    type: nextcloud
    url: nextcloud://admin:password@localhost:8080
    enabled: true

sync:
  enabled: true
  offline_mode: online  # Even with online mode!
EOF

# 2. Add a task using remote backend
todoat -b nextcloud-test Work add "Test task"
# Output: Created task: Test task

# 3. Check if task is on remote - IT'S NOT!
# (task was queued but not synced)

# 4. Check sync queue
todoat sync queue
# Shows: 1 pending create operation

# 5. Must manually sync
todoat sync
# Now task appears on remote
```

## Expected Behavior

When using `-b remotebackend` (or when `default_backend` is a remote backend), operations should either:

**Option A: Immediate sync after operation**
```
todoat -b nextcloud Work add "Task"
→ Creates task locally
→ Automatically syncs to remote
→ Reports success only after remote confirms
```

**Option B: Background sync triggered**
```
todoat -b nextcloud Work add "Task"
→ Creates task locally
→ Triggers async sync (non-blocking)
→ Reports success immediately
→ Sync happens in background
```

## Actual Behavior

```
todoat -b nextcloud Work add "Task"
→ Creates task in SQLite cache (or remote if online mode)
→ Queues operation in sync_queue
→ Reports success
→ Operation sits in queue until manual `todoat sync`
```

## Root Cause

The `syncAwareBackend` wrapper only queues operations - it never triggers sync:

```go
// CreateTask creates a task and queues a sync operation
func (b *syncAwareBackend) CreateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
    created, err := b.TaskManager.CreateTask(ctx, listID, task)
    if err != nil {
        return nil, err
    }

    // Queue create operation - but never syncs!
    _ = b.syncMgr.QueueOperationByStringID(created.ID, created.Summary, listID, "create")

    return created, nil
}
```

The wrapper should either:
1. Call `doSync()` after queueing, or
2. Signal the daemon to sync immediately, or
3. Have a config option for "sync after operation"

## Suggested Fix

### Option 1: Add `auto_sync_after_operation` config

```yaml
sync:
  enabled: true
  auto_sync_after_operation: true  # Sync immediately after each operation
```

Then in `syncAwareBackend`:

```go
func (b *syncAwareBackend) CreateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
    created, err := b.TaskManager.CreateTask(ctx, listID, task)
    if err != nil {
        return nil, err
    }

    _ = b.syncMgr.QueueOperationByStringID(created.ID, created.Summary, listID, "create")

    // Auto-sync if configured
    if b.autoSyncEnabled {
        _ = b.syncMgr.SyncNow()  // New method to trigger immediate sync
    }

    return created, nil
}
```

### Option 2: Signal running daemon

If daemon is running, signal it to sync immediately instead of waiting for interval:

```go
func (b *syncAwareBackend) CreateTask(...) (*backend.Task, error) {
    // ... create and queue ...

    // Signal daemon to sync now (if running)
    b.syncMgr.NotifyDaemon()

    return created, nil
}
```

### Option 3: Sync on Close

Sync when the backend connection closes (end of CLI command):

```go
func (b *syncAwareBackend) Close() error {
    // Sync pending operations before closing
    if b.pendingOps > 0 {
        _ = b.syncMgr.SyncNow()
    }
    _ = b.syncMgr.Close()
    return b.TaskManager.Close()
}
```

## Workaround

Until fixed, users can:
1. Run `todoat sync` after operations
2. Chain commands: `todoat -b nextcloud Work add "Task" && todoat sync`
3. Start the daemon (once issue #005 is fixed): `todoat sync daemon start`

## Related Issues

- [005-daemon-does-not-call-sync.md] - Daemon should handle periodic sync but doesn't
- [003-sync-command-does-not-actually-sync.md] - Manual sync now works (fixed)
- [001-sync-architecture-not-implemented.md] - Overall sync architecture issues

## Test Criteria

```go
// TestAutoSyncAfterOperation verifies operations trigger sync
func TestAutoSyncAfterOperation(t *testing.T) {
    cli := testutil.NewCLITestWithSync(t)
    mockRemote := testutil.NewMockRemoteBackend(t)

    cli.WriteConfig(`
sync:
  enabled: true
  auto_sync_after_operation: true
backends:
  nextcloud-test:
    type: nextcloud
    enabled: true
default_backend: nextcloud-test
`)

    // Add task
    cli.MustExecute("-y", "-b", "nextcloud-test", "Work", "add", "Test task")

    // Queue should be empty (already synced)
    queueOut := cli.MustExecute("-y", "sync", "queue")
    testutil.AssertContains(t, queueOut, "Pending Operations: 0")

    // Remote should have the task
    if mockRemote.CreateTaskCallCount() != 1 {
        t.Errorf("expected auto-sync to push task to remote")
    }
}
```

## Resolution

**Fixed in**: this session
**Fix description**: Implemented Option 1 from the suggested fix - added `auto_sync_after_operation` config option.

**Changes made**:
1. Added `AutoSyncAfterOperation` field to `SyncConfig` in `internal/config/config.go`
2. Added `IsAutoSyncAfterOperationEnabled()` helper method to Config
3. Added `cfg` field to `syncAwareBackend` struct to access config
4. Added `triggerAutoSync()` method to `syncAwareBackend` that:
   - Loads config to check if auto_sync_after_operation is enabled
   - Skips sync if in explicit offline mode
   - Calls `doSync()` to process pending queue
5. Modified `CreateTask`, `UpdateTask`, and `DeleteTask` methods to call `triggerAutoSync()` after queueing operations
6. Updated all creation points of `syncAwareBackend` to include the `cfg` field

**Test added**: `TestAutoSyncAfterOperation` and related tests in `backend/sync/sync_test.go`

### Verification Log
```bash
$ go test -v -run TestAutoSync ./backend/sync/
=== RUN   TestAutoSyncAfterOperation
--- PASS: TestAutoSyncAfterOperation (0.06s)
=== RUN   TestAutoSyncDisabledQueuesOnly
--- PASS: TestAutoSyncDisabledQueuesOnly (0.04s)
=== RUN   TestAutoSyncAfterUpdateOperation
--- PASS: TestAutoSyncAfterUpdateOperation (0.07s)
=== RUN   TestAutoSyncAfterDeleteOperation
--- PASS: TestAutoSyncAfterDeleteOperation (0.07s)
PASS
ok  	todoat/backend/sync	0.247s
```
**Matches expected behavior**: YES

**Usage**:
```yaml
sync:
  enabled: true
  auto_sync_after_operation: true  # Sync immediately after each operation
```

When enabled, operations (add/update/delete) will automatically trigger sync to push changes to the remote backend, eliminating the need to run `todoat sync` manually.
