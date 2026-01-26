# [011] Backend Data Isolation Broken - SQLite Cache Mixes Data Between Backends

## Type
code-bug

## Severity
critical

## Source
User report

## Description

The SQLite cache is not properly isolating data between different backends. When switching between backends (e.g., Todoist → Nextcloud), data from one backend appears when querying another backend.

This causes:
1. Wrong tasks displayed for a backend
2. Sync failures trying to push tasks to wrong remote
3. Complete data corruption between backends

## Steps to Reproduce

```bash
# Use Todoist backend
todoat -b todoist Inbox add "Task 1"
todoat -b todoist list
# Shows: Inbox, @work, Test, TestExamples (Todoist projects)

# Switch to Nextcloud backend
todoat -b nextcloud-test list
# WRONG: Shows Todoist lists instead of Nextcloud calendars!
# Shows: Inbox, @work, Test, TestExamples

# Try to sync
todoat -b nextcloud-test sync
# Skipping task '16': list 'Test' doesn't exist on remote and cannot be created
# Pull error: failed to get lists from remote: PROPFIND failed with status 401
```

## Expected Behavior

- Each backend should have isolated data in the SQLite cache
- `-b nextcloud-test list` should show Nextcloud calendars only
- `-b todoist list` should show Todoist projects only
- No cross-contamination between backends

## Actual Behavior

- SQLite cache appears to be shared across all backends
- Todoist data shows up when querying Nextcloud backend
- Sync tries to push Todoist tasks to Nextcloud (fails)
- Complete backend confusion

## Root Cause Found

**Location**: `cmd/todoat/cmd/todoat.go:2542`

The `createSyncFallbackBackend()` function always uses `sqlite.New()` which defaults to backend_id="sqlite":

```go
func createSyncFallbackBackend(cfg *Config, dbPath string) (backend.TaskManager, error) {
    be, err := sqlite.New(dbPath)  // BUG: Always uses "sqlite" as backend_id!
    // ...
}
```

This function is called for ALL backends when sync is enabled with `offline_mode: auto` (the default). So `-b todoist`, `-b nextcloud-test`, etc. all share the same "sqlite" backend_id, mixing their data.

**The SQLite backend DOES have proper isolation** - `backend/sqlite/sqlite.go` has:
- `backend_id` column in tables
- All queries filter by `backend_id`
- `NewWithBackendID(path, backendID)` function that correctly sets isolation

**The bug is that the wrong constructor is called.**

### Fix Required

Change `cmd/todoat/cmd/todoat.go:2542` from:
```go
be, err := sqlite.New(dbPath)
```

To:
```go
be, err := sqlite.NewWithBackendID(dbPath, backendName)
```

And update the function signature to accept `backendName`:
```go
func createSyncFallbackBackend(cfg *Config, dbPath string, backendName string) (backend.TaskManager, error) {
```

## Impact

- **Data integrity compromised** - Tasks appear under wrong backend
- **Sync broken** - Pushes to wrong remote backend
- **User confusion** - Can't trust which backend they're viewing
- **Potential data loss** - Operations on wrong backend could delete/modify wrong tasks

## Expected Fix

1. Update `createSyncFallbackBackend()` to accept and use backend name
2. Update all call sites to pass the backend name
3. Users with corrupted data will need to clear cache: `rm ~/.local/share/todoat/tasks.db`

### Call Sites to Update

Search for `createSyncFallbackBackend` calls:
- Line 2500: `return createSyncFallbackBackend(cfg, dbPath)` - needs backendName
- Any other call sites

### Additional Fixes Needed

Also check `doSync()` at line 5952:
```go
localBE, err := sqlite.New(dbPath)  // Should use backend-specific ID
```

## Workaround

None known - the data is already corrupted in the cache.

Users may need to:
1. Delete the SQLite cache: `rm ~/.local/share/todoat/tasks.db`
2. Re-sync each backend separately

## Related Files

- `backend/sqlite/sqlite.go` - SQLite backend implementation
- `backend/sqlite/migrations/` - Database schema
- `cmd/todoat/cmd/todoat.go` - Backend selection logic
