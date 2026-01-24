# [005] Sync Daemon Does Not Actually Call Sync

## Type
code-bug

## Severity
high

## Description

The sync daemon (`todoat sync daemon start`) runs in the background and logs "Sync completed" at each interval, but it never actually calls the sync function. The daemon is just a timer that logs messages without performing any synchronization.

## Steps to Reproduce

```bash
# 1. Configure sync with remote backend
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
  daemon:
    interval: 10
EOF

# 2. Add a task (gets queued)
todoat Work add "Test task"

# 3. Check queue
todoat sync queue
# Shows: 1 pending operation

# 4. Start daemon
todoat sync daemon start --interval 10

# 5. Wait for daemon to "sync"
sleep 15

# 6. Check daemon status - shows sync completed
todoat sync daemon status
# Shows: Sync count: 1

# 7. Check queue - STILL HAS PENDING OPERATION!
todoat sync queue
# Shows: 1 pending operation (should be 0 if daemon actually synced)

# 8. Check daemon log
cat ~/.local/share/todoat/daemon.log
# Shows: "[timestamp] Sync completed (count: 1)" but nothing was synced
```

## Expected Behavior

When daemon interval triggers:
1. Daemon calls `doSync()` to perform actual synchronization
2. Pending operations are pushed to remote backend
3. Remote changes are pulled to local cache
4. Queue is cleared after successful sync

## Actual Behavior

When daemon interval triggers:
1. Daemon increments sync counter
2. Daemon logs "Sync completed"
3. **No actual sync occurs**
4. Queue remains unchanged

## Root Cause

In `cmd/todoat/cmd/todoat.go` lines 7032-7076, the `daemonSyncLoop` function never calls `doSync()`:

```go
func daemonSyncLoop(daemon *daemonState, logPath string) {
    ticker := time.NewTicker(daemon.interval)
    // ...
    for {
        select {
        case <-ticker.C:
            daemon.syncCount++
            daemon.lastSync = time.Now()

            // Perform sync
            var syncErr error
            if daemon.offlineMode {
                logEntry := fmt.Sprintf("[%s] Sync attempt %d (offline mode)\n", ...)
                _ = appendToLogFile(logPath, logEntry)
            } else {
                // Normal sync - JUST LOGS! Never calls doSync()!
                logEntry := fmt.Sprintf("[%s] Sync completed (count: %d)\n", ...)
                _ = appendToLogFile(logPath, logEntry)
            }
            // ...
        }
    }
}
```

The `syncErr` variable is declared but never assigned from an actual sync operation.

## Config Reference

The daemon can be configured via:

**Config file:**
```yaml
sync:
  enabled: true
  daemon:
    interval: 60  # seconds
```

**CLI flag:**
```bash
todoat sync daemon start --interval 30
```

## Suggested Fix

The `daemonSyncLoop` function needs to actually call the sync logic:

```go
case <-ticker.C:
    daemon.syncCount++
    daemon.lastSync = time.Now()

    // Actually perform sync
    var syncErr error
    if daemon.offlineMode {
        logEntry := fmt.Sprintf("[%s] Sync attempt %d (offline mode)\n",
            time.Now().Format(time.RFC3339), daemon.syncCount)
        _ = appendToLogFile(logPath, logEntry)
    } else {
        // Create a config for doSync
        syncCfg := &Config{
            ConfigPath: daemon.configPath,
            DBPath:     daemon.dbPath,
            // ... other necessary fields
        }

        // Actually call sync!
        syncErr = doSync(syncCfg, io.Discard, io.Discard)

        if syncErr != nil {
            logEntry := fmt.Sprintf("[%s] Sync error (count: %d): %v\n",
                time.Now().Format(time.RFC3339), daemon.syncCount, syncErr)
            _ = appendToLogFile(logPath, logEntry)
        } else {
            logEntry := fmt.Sprintf("[%s] Sync completed (count: %d)\n",
                time.Now().Format(time.RFC3339), daemon.syncCount)
            _ = appendToLogFile(logPath, logEntry)
        }
    }
```

Note: The `daemonState` struct may need additional fields to hold config paths needed by `doSync()`.

## Test Criteria

```go
// TestDaemonActuallySyncs verifies daemon calls sync at each interval
func TestDaemonActuallySyncs(t *testing.T) {
    cli := testutil.NewCLITestWithDaemon(t)

    // Configure with short interval
    cli.WriteConfig(`
sync:
  enabled: true
  daemon:
    interval: 1
backends:
  nextcloud-test:
    type: nextcloud
    enabled: true
default_backend: nextcloud-test
`)

    // Add task to queue
    cli.MustExecute("-y", "Work", "add", "Test task")

    // Verify task is queued
    queueOut := cli.MustExecute("-y", "sync", "queue")
    testutil.AssertContains(t, queueOut, "Pending Operations: 1")

    // Start daemon
    cli.MustExecute("-y", "sync", "daemon", "start", "--interval", "1")
    defer cli.MustExecute("-y", "sync", "daemon", "stop")

    // Wait for daemon sync cycle
    time.Sleep(2 * time.Second)

    // CRITICAL: Queue should be empty after daemon sync
    queueOut = cli.MustExecute("-y", "sync", "queue")
    testutil.AssertContains(t, queueOut, "Pending Operations: 0")
}
```

## Related Issues

- [001-sync-architecture-not-implemented.md]
- [003-sync-command-does-not-actually-sync.md] - Manual sync had similar issue (now fixed per user)
- [004-sync-documentation-describes-unimplemented-features.md]

## Resolution

**Fixed in**: this session
**Fix description**: Modified `daemonSyncLoop` in `cmd/todoat/cmd/todoat.go` to call `doSync(daemon.cfg, io.Discard, io.Discard)` when performing sync, instead of just logging "Sync completed".
**Test added**: `TestDaemonActuallySyncs` in `backend/sync/daemon_test.go`

### Verification Log
```bash
$ go test -v -run TestDaemonActuallySyncs ./backend/sync/
=== RUN   TestDaemonActuallySyncs
--- PASS: TestDaemonActuallySyncs (0.24s)
PASS
ok      todoat/backend/sync     0.243s
```
**Matches expected behavior**: YES
