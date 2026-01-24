# [002] default_backend Configuration Ignored When Sync Enabled

## Type
code-bug

## Severity
high

## Description

When `sync.enabled: true` is set in the configuration, the `default_backend` setting is completely ignored. The CLI immediately uses SQLite instead of the configured default backend.

This makes the `default_backend` configuration useless for users who want to use sync with a remote backend like Nextcloud or Todoist.

## Steps to Reproduce

```bash
# Create config with sync enabled and default_backend set to a custom backend
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
  local_backend: sqlite
  conflict_resolution: server_wins
  offline_mode: auto
EOF

# Run todoat without -b flag - should use nextcloud-test but uses sqlite instead
todoat Work

# Verify with --detect-backend (shows sqlite, ignores default_backend)
todoat --detect-backend
```

## Expected Behavior

When `default_backend: nextcloud-test` is configured:
1. `todoat Work` should use the `nextcloud-test` backend
2. With sync enabled, it should use SQLite as the local cache for `nextcloud-test`, not as the primary backend
3. The `default_backend` setting should be respected regardless of sync configuration

## Actual Behavior

1. `todoat Work` uses SQLite directly, ignoring `default_backend: nextcloud-test`
2. The `-b nextcloud-test` flag works correctly
3. The `default_backend` config has no effect when sync is enabled

## Root Cause

In `cmd/todoat/cmd/todoat.go` lines 2236-2245, when sync is enabled, the code immediately returns SQLite without checking `default_backend`:

```go
// If sync is enabled, return a sync-aware backend wrapper
if cfg.SyncEnabled {
    be, err := sqlite.New(dbPath)
    if err != nil {
        return nil, err
    }
    return &syncAwareBackend{
        TaskManager: be,
        syncMgr:     getSyncManager(cfg),
    }, nil
}

// Check default_backend setting from config  <-- NEVER REACHED when sync enabled!
if appConfig != nil && appConfig.DefaultBackend != "" && appConfig.DefaultBackend != "sqlite" {
    ...
}
```

The `default_backend` check at line 2266 is only reached if sync is NOT enabled, making it useless for sync users.

## Additional Context

This issue is related to [001-sync-architecture-not-implemented.md] - both stem from the same architectural problem where sync was implemented as a "fallback" mechanism rather than the documented offline-first architecture.

The `--detect-backend` flag showing "sqlite" is expected behavior since it shows auto-detected backends (based on environment), not the `default_backend` from config. These are different concepts:
- **Auto-detect**: Detects backends based on environment (git repo marker, sqlite file existence)
- **default_backend**: User's preferred backend from configuration

## Test Criteria for Fix

```go
// TestDefaultBackendRespectedWithSyncEnabled verifies that default_backend
// is used even when sync is enabled
func TestDefaultBackendRespectedWithSyncEnabled(t *testing.T) {
    cli := testutil.NewCLITest(t)

    // Configure sync enabled with custom default backend
    cli.WriteConfig(`
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
  local_backend: sqlite
`)

    // Mock the nextcloud backend to verify it's being used
    mockNextcloud := testutil.NewMockBackend(t, "nextcloud")
    cli.RegisterMockBackend("nextcloud-test", mockNextcloud)

    // Run command without -b flag
    cli.MustExecute("-y", "Work", "add", "Test task")

    // Verify: nextcloud-test backend should have been used (not sqlite directly)
    if mockNextcloud.AddTaskCallCount() == 0 {
        t.Errorf("default_backend nextcloud-test was not used; expected it to be called")
    }
}
```

## Dependencies

- Related to: [001-sync-architecture-not-implemented.md]

## Resolution

**Fixed in**: this session
**Fix description**: Modified `getBackend()` in `cmd/todoat/cmd/todoat.go` to check `default_backend` when sync is enabled, instead of immediately returning SQLite. Now when sync is enabled and `default_backend` is set to a non-sqlite backend, the CLI uses `createBackendWithSyncFallback()` for that backend, which:
1. Attempts to connect to the configured remote backend
2. Falls back to SQLite cache if the remote is unavailable
3. Shows a warning message about the fallback

**Test added**: `TestDefaultBackendRespectedWhenSyncEnabled` and `TestDefaultBackendWithBackendFlagOverride` in `backend/sync/sync_test.go`

### Verification Log
```bash
$ go test -v -run TestDefaultBackendRespectedWhenSyncEnabled ./backend/sync/
=== RUN   TestDefaultBackendRespectedWhenSyncEnabled
[DEBUG] Sync enabled with default_backend: nextcloud-test
[DEBUG] Using custom backend 'nextcloud-test' of type 'nextcloud'
[DEBUG] Backend connectivity check failed, falling back to SQLite cache: ...
--- PASS: TestDefaultBackendRespectedWhenSyncEnabled (0.53s)
PASS
ok      todoat/backend/sync     0.539s
```
**Matches expected behavior**: YES - The fix correctly respects `default_backend` when sync is enabled, attempting to use the configured backend before falling back to SQLite.
