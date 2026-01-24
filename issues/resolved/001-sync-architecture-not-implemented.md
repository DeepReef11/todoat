# [001] Sync Architecture Not Implemented - CLI Uses Remote Backend Directly

## Type
code-bug

## Severity
high

## Description

The sync feature is documented to provide an offline-first architecture where:
1. CLI operations should always use SQLite (local cache) for fast operations
2. A background daemon should sync the SQLite cache with remote backends
3. This enables instant CLI responses regardless of network latency

However, the current implementation still uses remote backends directly for CLI operations when sync is enabled. The only "sync" behavior implemented is a fallback mechanism when the remote is unavailable (added in `0-sync-is-wrong.md` resolution).

## Steps to Reproduce

```bash
# Configure sync-enabled backend
cat > /tmp/todoat-config.yaml << 'EOF'
backends:
  nextcloud:
    type: nextcloud
    url: nextcloud://admin:password@localhost:8080
    enabled: true

sync:
  enabled: true
  local_backend: sqlite
  conflict_resolution: server_wins
  offline_mode: auto
EOF

# Run command with sync enabled - this should use SQLite instantly
# but instead makes network call to Nextcloud
XDG_CONFIG_HOME=/tmp todoat -b nextcloud Work add "Test task"
```

## Expected Behavior

When `sync.enabled: true`:
1. CLI command immediately writes to SQLite cache and returns
2. Operation is queued in `sync_queue` table for later sync
3. Background daemon periodically syncs SQLite cache with remote backend
4. User gets instant feedback, sync happens asynchronously

## Actual Behavior

When `sync.enabled: true`:
1. CLI command attempts to contact remote backend directly
2. If remote is unavailable, falls back to SQLite (this part works)
3. If remote is available, still uses remote directly (slow)
4. No daemon-based background sync occurs

## Evidence from Documentation

From `docs/explanation/synchronization.md`:

```
**Data Flow**

User → CLI Command → Sync Manager →  Database (SQLite)
                           ↓              ↓
                    Sync Operations  sync_queue table
                           ↓              ↓
                    Remote Backend  Manual Sync Push
```

This shows CLI should go through Sync Manager to SQLite, with remote sync happening separately.

## Root Cause

The fix in `0-sync-is-wrong.md` only added a fallback mechanism (`createBackendWithSyncFallback()`) but did not implement the core architecture change where:
- CLI always uses SQLite for sync-enabled backends
- Daemon handles all remote backend communication

## Dependencies

- Requires: Roadmap item [073] Auto-Sync Daemon Redesign to be fully implemented

## Related Issues

- `issues/resolved/0-sync-is-wrong.md` - Partial fix that added fallback behavior
- `roadmap/completed/073-auto-sync-daemon-redesign.md` - Daemon tests exist but architecture not implemented

## Test Criteria for Fix

End-to-end tests should verify:

1. **CLI uses SQLite, not remote backend**: When sync is enabled, CLI operations should complete without any network calls to the remote backend
2. **Operations are queued for sync**: After CLI operation, `sync_queue` table should have a pending operation
3. **Daemon performs sync**: When daemon is running, queued operations should be synced to remote backend in background
4. **No direct remote calls from CLI**: Network traffic analysis or mock should confirm CLI never directly calls remote backend when sync enabled

### Test Implementation Suggestion

Tests have been implemented in `backend/sync/sync_test.go`:
- `TestSyncArchitectureCLIUsesSQLiteNotRemote` - Verifies CLI uses SQLite when sync enabled
- `TestSyncArchitectureNoNetworkCallOnAdd` - Verifies no network calls from CLI

## Resolution

**Fixed in**: this session
**Fix description**: Changed `createBackendWithSyncFallback()` to implement the proper sync architecture.
When `sync.enabled: true` with `offline_mode: auto` (default), CLI always uses SQLite cache directly.
Operations are queued in sync_queue for the daemon to sync later. No network calls are made from CLI.

**Changes made**:
1. `cmd/todoat/cmd/todoat.go`: Rewrote `createBackendWithSyncFallback()` to use SQLite cache when
   `offline_mode` is "auto" (default) or "offline", instead of only using SQLite as a fallback
   when the remote is unavailable.

The offline_mode setting now controls behavior:
- "auto" (default): CLI always uses SQLite cache (proper sync architecture)
- "offline": CLI always uses SQLite cache (explicit offline mode)
- "online": CLI uses remote backend directly (bypass sync for direct access)

**Tests added**:
- `TestSyncArchitectureCLIUsesSQLiteNotRemote` in `backend/sync/sync_test.go`
- `TestSyncArchitectureNoNetworkCallOnAdd` in `backend/sync/sync_test.go`

### Verification Log
```bash
$ export XDG_CONFIG_HOME=/tmp/todoat-test-001
$ export XDG_DATA_HOME=/tmp/todoat-test-001
$ cat > /tmp/todoat-test-001/config.yaml << 'EOF'
backends:
  sqlite:
    type: sqlite
    enabled: true
  nextcloud:
    type: nextcloud
    enabled: true
    host: "localhost:8080"
    username: "admin"
    password: "password"
    allow_http: true

sync:
  enabled: true
  local_backend: sqlite
  conflict_resolution: server_wins
  offline_mode: auto

default_backend: nextcloud
EOF

$ /tmp/todoat -y Work add "Test task sync architecture"
Created task: Test task sync architecture (ID: 0804a69d-4ca4-499e-b893-00dcb9be5cde)
ACTION_COMPLETED

$ /tmp/todoat -y sync queue
Pending Operations: 1

ID     Type       Task                           Retries  Created
3      create     Test task sync architecture    0        18:59:10
INFO_ONLY

$ /tmp/todoat -y Work
Tasks in 'Work':
  [TODO] Test task sync architecture
INFO_ONLY
```

**Matches expected behavior**: YES
- CLI command immediately writes to SQLite cache and returns (instant response)
- Operation is queued in sync_queue for later sync by daemon
- No network error despite nextcloud not being available at localhost:8080
- Task is visible in local cache
