Sync is supposed to make operation faster when enabled and also be able to sync later if connection to server not working (when it had sync at least once in the past).

Sync command seem to have worked (I assume it synced in background with previous commands)
```bash
todoat -b nextcloud-test sync
Sync completed with backend 'nextcloud-test'
  Operations processed: 0

```
Docker nextcloud container was paused to see if sync works properly

```bash
todoat -b nextcloud-test
Error: Propfind "http://localhost:8080/remote.php/dav/calendars/admin/": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
```

Instead of directly using sqlite and launch background deamon that would send notification that nextcloud not available, could not sync, it seems like it completely ignored the sync feature.

End to end test should have direct tests but also sync tests using config that enable sync feature:
```yaml
sync:
  enabled: true
  local_backend: sqlite                    # Cache backend for remote syncing
  conflict_resolution: server_wins         # Options: server_wins | local_wins | merge | keep_both
  offline_mode: auto                       # Options: auto | online | offline
```

sync did not work with todoist either.

## Resolution

**Fixed in**: this session
**Fix description**: When sync is enabled and a remote backend is specified with `-b`, the app now checks connectivity with a configurable timeout. If the backend is unavailable and `offline_mode` is "auto" (default) or "offline", it falls back to SQLite cache and queues operations for later sync. If `offline_mode` is "online", it fails immediately.

**Changes made**:
1. `cmd/todoat/cmd/todoat.go`: Added `createBackendWithSyncFallback()` function that attempts to connect to remote backend and falls back to SQLite if unavailable
2. `cmd/todoat/cmd/todoat.go`: Fixed `loadSyncConfigFromAppConfig()` to properly read sync config from the parsed config (was not finding config when using XDG paths)
3. `backend/sync/sync_test.go`: Added tests `TestSyncFallbackToSQLiteWhenRemoteUnavailable` and `TestSyncFallbackListTasksWhenRemoteUnavailable`

**Test added**: `TestSyncFallbackToSQLiteWhenRemoteUnavailable` and `TestSyncFallbackListTasksWhenRemoteUnavailable` in `backend/sync/sync_test.go`

### Verification Log
```bash
$ XDG_CONFIG_HOME=./config XDG_DATA_HOME=./data todoat -y -b nextcloud-test Work add "Test task when offline"
Warning: Backend 'nextcloud-test' unavailable (Propfind "http://localhost:8080/remote.php/dav/calendars/admin/": dial tcp [::1]:8080: connect: connection refused). Using SQLite cache (operations will be queued for sync).
Created task: Test task when offline (ID: 34af8b7d-c73e-4355-b670-1d24f62c3436)
ACTION_COMPLETED
$ exit code: 0

$ todoat -y Work get
Tasks in 'Work':
  [TODO] Test task when offline
INFO_ONLY

$ todoat -y sync queue
Pending Operations: 1
ID     Type       Task                           Retries  Created
1      create     Test task when offline         0        16:22:20
INFO_ONLY
```
**Matches expected behavior**: YES
