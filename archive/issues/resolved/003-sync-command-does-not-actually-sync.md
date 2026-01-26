# [003] Sync Command Does Not Actually Sync - Just Clears Queue

## Type
code-bug

## Severity
critical

## Description

The `todoat sync` command does not actually synchronize tasks with remote backends. It creates a connection to the remote backend, counts pending operations, then clears the queue without ever executing the operations. **Tasks are never actually pushed to or pulled from the remote.**

This means users who rely on the sync feature are losing their data - operations are queued and then silently discarded.

## Steps to Reproduce

```bash
# 1. Configure sync with a remote backend
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

# 2. Add a task (gets queued)
todoat -b nextcloud-test Work add "Test task for sync"

# 3. Check queue - task is pending
todoat sync queue
# Shows: 1 pending create operation

# 4. Run sync
todoat sync
# Output: Sync completed with backend 'nextcloud-test'
#         Operations processed: 1

# 5. Check queue - queue is empty
todoat sync queue
# Shows: Pending Operations: 0

# 6. Check remote backend - TASK IS NOT THERE!
todoat -b nextcloud-test Work
# Task does not exist on remote - it was never synced
```

## Expected Behavior

`todoat sync` should:
1. Pull changes from remote backend to local cache
2. Push queued operations (create/update/delete) to remote backend
3. Handle conflicts according to configured strategy
4. Only clear queue entries after successful sync

## Actual Behavior

`todoat sync` does:
1. Connects to remote backend
2. Counts pending operations
3. **Immediately clears the queue without syncing**
4. Reports "Sync completed" misleadingly

## Root Cause

In `cmd/todoat/cmd/todoat.go` lines 5683-5706, the sync implementation is a stub:

```go
// Process pending operations
// Note: This is a simplified sync that marks queued operations as synced.
// Full bidirectional sync with conflict resolution would require more complex logic.
successCount := 0
errorCount := 0

for _, op := range pendingOps {
    // For now, we simply count operations as successful since we've established
    // connectivity to the remote backend. Full sync implementation would actually
    // execute create/update/delete operations on the remote.
    switch op.OperationType {
    case "create", "update", "delete":
        successCount++  // Just counting, not actually syncing!
    default:
        errorCount++
    }
}

// Mark remoteBE as used (Go compiler requirement)
_ = remoteBE  // REMOTE BACKEND IS NEVER USED!

// Clear successfully processed operations
if successCount > 0 {
    _, _ = syncMgr.ClearQueue()  // Clears queue without syncing!
}
```

The comment explicitly states this is a "simplified sync" but the behavior is deceptive - it reports success while doing nothing.

## Documentation vs Reality

The documentation in `docs/explanation/synchronization.md` describes a complete sync architecture with:
- Pull operations fetching from remote
- Push operations executing queued operations
- Conflict resolution
- Retry logic with exponential backoff
- Hierarchical ordering for parent/child tasks

**None of this exists in the code.** The documentation describes an aspirational design, not the actual implementation.

Key documented features that don't exist:
1. `SyncManager.Pull()` - doesn't exist
2. `SyncManager.Push()` - doesn't exist
3. Conflict resolution during sync - doesn't exist
4. Retry logic - doesn't exist
5. Hierarchical sync ordering - doesn't exist

## Impact

**Data Loss**: Users think their tasks are being synced but they're not. Queued operations are silently discarded.

## Related Issues

- [001-sync-architecture-not-implemented.md] - CLI architecture issue
- [002-default-backend-ignored-when-sync-enabled.md] - Config issue

## Test Criteria for Fix

```go
// TestSyncActuallySyncsToRemote verifies sync command executes operations on remote
func TestSyncActuallySyncsToRemote(t *testing.T) {
    cli := testutil.NewCLITestWithSync(t)
    mockRemote := testutil.NewMockRemoteBackend(t)
    cli.RegisterMockBackend("nextcloud-test", mockRemote)

    cli.WriteConfig(`
sync:
  enabled: true
backends:
  nextcloud-test:
    type: nextcloud
    enabled: true
default_backend: nextcloud-test
`)

    // Add task (queued locally)
    cli.MustExecute("-y", "Work", "add", "Test task")

    // Verify task is queued
    queueOut := cli.MustExecute("-y", "sync", "queue")
    testutil.AssertContains(t, queueOut, "Pending Operations: 1")

    // Run sync
    cli.MustExecute("-y", "sync")

    // CRITICAL: Verify remote backend received the create operation
    if mockRemote.CreateTaskCallCount() != 1 {
        t.Errorf("sync did not push task to remote; expected 1 CreateTask call, got %d",
            mockRemote.CreateTaskCallCount())
    }

    // Verify task exists on remote
    tasks := mockRemote.GetTasks("Work")
    if len(tasks) != 1 || tasks[0].Summary != "Test task" {
        t.Errorf("task not found on remote after sync")
    }
}
```

## Recommendation

Either:
1. **Implement actual sync** - Make the sync command actually push/pull as documented
2. **Remove the feature** - If sync isn't going to work, remove it entirely rather than pretending it works
3. **Warn users** - At minimum, change the output from "Sync completed" to "Sync not implemented - queue cleared (operations NOT synced)"

Option 3 is the minimum viable fix to prevent data loss while the feature is properly implemented.

## Resolution

**Fixed in**: this session
**Fix description**: Implemented actual sync functionality in `doSync()`. Added three helper functions (`syncCreateOperation`, `syncUpdateOperation`, `syncDeleteOperation`) that execute pending operations on the remote backend:
- For "create" operations: Reads task from local SQLite, ensures list exists on remote, creates task on remote
- For "update" operations: Reads task from local SQLite, updates task on remote
- For "delete" operations: Deletes task from remote (or succeeds if already deleted)

**Test added**: `TestSyncActuallySyncsToRemote` in `backend/sync/sync_test.go`

### Verification Log
```bash
$ go test -v -run TestSyncActuallySyncsToRemote ./backend/sync/
=== RUN   TestSyncActuallySyncsToRemote
[DEBUG] Sync enabled with default_backend: sqlite-remote
[DEBUG] Offline mode: using SQLite cache (sync enabled, offline_mode=offline)
[DEBUG] Using custom backend 'sqlite-remote' of type 'sqlite'
--- PASS: TestSyncActuallySyncsToRemote (0.06s)
PASS
ok      todoat/backend/sync     0.069s
```

The test verifies:
1. Task is queued when added in offline mode (Pending Operations: 1)
2. Sync command runs successfully (Operations processed: 1)
3. Queue is cleared after sync (Pending Operations: 0)
4. **CRITICAL**: Task actually exists in the remote database after sync (verified by direct SQL query)

**Matches expected behavior**: YES
