# [005] Sync Pull Not Implemented

## Type
code-bug

## Severity
critical

## Source
User report + code review

## Description

The `todoat sync` command only **pushes** local changes to remote backends. It never **pulls** changes from remote backends, despite documentation claiming bidirectional sync.

## Steps to Reproduce

1. Configure sync with Nextcloud backend
2. Add tasks directly in Nextcloud (via web UI or another client)
3. Run `todoat sync`
4. Run `todoat list` - remote tasks are NOT shown

## Expected Behavior

- `todoat sync` should fetch tasks from remote backend
- New remote tasks should appear in local list
- Updated remote tasks should be merged with local
- Deleted remote tasks should be removed locally
- Output should show "Pull: X tasks updated, Y new tasks, Z deleted"

## Actual Behavior

- `todoat sync` only processes the local queue (push operations)
- Remote changes are never fetched
- Tasks created on remote never appear locally
- The sync is unidirectional (push-only), not bidirectional

## Code Location

`cmd/todoat/cmd/todoat.go` - `doSync()` function (lines ~5898-6020)

The function only:
1. Gets pending operations from local queue
2. Pushes them to remote
3. Clears the queue

It never:
1. Calls `remoteBE.GetLists()` or `remoteBE.GetTasks()`
2. Compares remote state with local
3. Imports new/changed remote tasks

## Documentation Claims (False)

From `docs/how-to/sync.md`:
- "Pulls changes from remote backends and pushes local changes"
- "Pull: 15 tasks updated, 3 new tasks, 1 deleted"

From `docs/explanation/synchronization.md`:
- "Bidirectional sync (pull from remote, push local changes)"
- "Sync Manager pulls changes from remote backend"

## Impact

- Users expect sync to work bidirectionally
- Changes made on remote (web, mobile, other clients) are never synced to local
- This defeats the purpose of sync for multi-device usage

## Resolution

**Fixed in**: this session
**Fix description**: Added `syncPullFromRemote()` function to `doSync()` that pulls tasks from remote backend to local. The function:
1. Gets all lists and tasks from remote backend
2. Compares with local lists and tasks
3. Creates new local tasks for remote-only tasks
4. Updates local tasks that are older than remote versions
5. Deletes local tasks that no longer exist on remote
**Test added**: Updated `TestSyncPullCLI` in `backend/sync/sync_test.go` to verify pull functionality

### Verification Log
```bash
$ go test -v -run TestSyncPullCLI ./backend/sync/
=== RUN   TestSyncPullCLI
--- PASS: TestSyncPullCLI (0.08s)
PASS
ok      todoat/backend/sync     0.083s
```
**Matches expected behavior**: YES

The sync output now shows:
- `Push: X operations processed` - for push operations
- `Pull: X new, Y updated, Z deleted` - for pull operations from remote
