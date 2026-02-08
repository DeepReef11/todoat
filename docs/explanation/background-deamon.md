# Daemon Architecture: Timeout-Based Single Instance

## Overview

todoat uses a background daemon to process sync operations with remote backends (Nextcloud, Todoist, etc.) asynchronously. This daemon is used to prevent todoat from hanging the terminal, allowing the CLI to return as fast as possible while sync operations happen in the background. The daemon uses a timeout mechanism to stay alive during active periods and gracefully shutdown when idle.

**When the daemon is used:**
- Only when sync is enabled (`sync.enabled: true`)
- Only when using remote backends (Nextcloud, Todoist, Google Tasks, etc.)
- Not needed for local-only backends (SQLite, file)

**When the daemon is NOT needed:**
- Local SQLite backend without sync
- File backend without remote sync
- Any configuration where `sync.enabled: false`

## Core Design

### Single Daemon Instance
- Only one daemon runs per user session
- Enforced via pidfile/lockfile mechanism
- Prevents resource waste and coordination complexity
- Current pidfile location: `$XDG_RUNTIME_DIR/todoat/daemon.pid` or `/tmp/todoat-daemon.pid`

### Timeout-Based Lifecycle
- Daemon starts with 5-second idle timer
- Each new task resets the timer to 5 seconds
- When timer expires with no new tasks, daemon exits
- If CLI relaunches during active daemon, it communicates with existing instance

### Communication Method
- CLI communicates directly with daemon via IPC (Unix domain socket)
- If IPC connection fails AND pidfile check confirms no daemon, CLI launches new daemon
- Direct communication eliminates queue polling race conditions

## Current todoat Implementation Status

> **Decision [FEAT-011]**: This document requires a rewrite to match the current implementation. The status table and several sections below describe a pre-implementation state that no longer exists. The actual code in `internal/daemon/daemon.go` has a fully implemented forked process daemon with Unix domain socket IPC, daemon-driven sync loop, multi-backend support, and a client library. See `docs/decisions/question-log.md` for the full decision record.

> **Important**: Several sections below are **outdated** and describe a pre-implementation state. The actual implementation includes:
> - **Forked process** via `Fork()` using `exec.Command` with `Setsid: true`
> - **Unix domain socket IPC** with JSON message protocol (notify, status, stop)
> - **Daemon-driven sync loop** with `time.NewTicker`
> - **Multi-backend support** with per-backend intervals and failure isolation
> - **Client library** (`daemon.Client`) with `Notify()`, `Status()`, `Stop()` methods
> - **`todoat sync daemon start`** command exists and works

### What Exists Today

> **OUTDATED** — This table describes a pre-implementation state. See note above for current status.

| Component | Target Design | Current Implementation |
|-----------|---------------|------------------------|
| Daemon process | Separate forked process | ~~In-process goroutine only~~ **Implemented: forked process** |
| IPC/Socket | Unix domain socket | ~~None - single process~~ **Implemented: Unix domain socket** |
| Sync mechanism | Daemon-driven | ~~CLI-driven background goroutines~~ **Implemented: daemon-driven sync loop** |
| Multi-backend | Daemon iterates all backends | ~~Single backend sync only~~ **Implemented: per-backend intervals** |
| State management | Daemon process state | ~~WaitGroup + channels in-process~~ **Implemented: daemon process state** |

### Current Background Sync Patterns

todoat uses the following mechanisms:

1. **Auto-sync after operations** (`triggerAutoSync`): Background goroutine triggered after create/update/delete
2. **Background pull sync** (`triggerBackgroundPullSync`): Pulls from remote before read operations with cooldown
3. **Daemon sync loop** (`daemonSyncLoop`): Daemon-driven periodic sync via `todoat sync daemon start`

All use `backgroundSyncWG` WaitGroup to ensure completion before CLI exits:

```go
var backgroundSyncWG sync.WaitGroup

// In Execute()
defer backgroundSyncWG.Wait()
```

## Conflicts with Existing Implementation

> **OUTDATED** — Process isolation has been implemented. The daemon now runs as a separate forked process.

### ~~No Process Isolation~~ (Resolved)

The ~~current~~ previous in-process goroutine approach meant:
- ~~Daemon cannot outlive the CLI process~~ **Resolved: daemon runs as forked process**
- ~~No true background processing after CLI exits~~ **Resolved: daemon continues after CLI exits**
- ~~Network latency blocks CLI responsiveness~~ **Resolved: fire-and-forget with background indicator**
- ~~Users experience "terminal hanging" during sync (Issue #36)~~ **Resolved**

### Database Schema Differences

Current todoat uses separate databases with different schemas:

**Task database** (`~/.local/share/todoat/tasks.db`):
```sql
-- No worker_id, claimed_at, or status for daemon task claiming
tasks (id, list_id, summary, status, ...)  -- status is task completion, not processing
```

**Sync queue database** (`~/.local/share/todoat/sync.db`):
```sql
sync_queue (id, task_id, task_uid, operation_type, retry_count, ...)
-- No daemon-specific claiming mechanism
-- retry_count exists but no exponential backoff enforcement
```

To implement the planned failure recovery features, todoat would need:
- Add `daemon_tasks` table for daemon work queue (or repurpose sync_queue)
- Add `daemon_heartbeat` table for health monitoring (see issue #74)
- Add `worker_id`, `claimed_at` columns for atomic task claiming

> **Note**: These schema additions are **not yet implemented**. They are described here as part of the planned design.

### ~~No Unix Socket Infrastructure~~ (Resolved)

> **OUTDATED** — Unix socket IPC has been implemented in `internal/daemon/daemon.go`.

The daemon now includes:
- Unix domain socket creation/cleanup
- Socket listener for IPC
- JSON message protocol for CLI-daemon communication (notify, status, stop)
- `daemon.Client` library with `Notify()`, `Status()`, `Stop()` methods

## Task Deduplication

> **NOT YET IMPLEMENTED** — This section describes a planned design. The `worker_id`, `claimed_at`, and `status` columns do not exist in the `sync_queue` table. Current deduplication relies on the single-daemon-instance guarantee via pidfile locking.

### Problem
Edge cases could allow two daemon instances briefly, risking double-execution of tasks.

### Planned Solution
Atomic task claiming in SQLite database.

```sql
-- Daemon atomically claims next pending task
BEGIN IMMEDIATE;

UPDATE sync_queue
SET status = 'processing',
    worker_id = ?,
    claimed_at = CURRENT_TIMESTAMP
WHERE id = (
    SELECT id FROM sync_queue
    WHERE status = 'pending'
    ORDER BY created_at
    LIMIT 1
);

-- Check if claim succeeded
SELECT changes();  -- Returns 1 if successful, 0 if no tasks or already claimed

COMMIT;
```

### Key Properties
- `BEGIN IMMEDIATE` acquires write lock immediately
- Only one daemon can successfully update any given task
- Failed claim (0 rows) means task already taken or no work available
- todoat already uses WAL mode: `PRAGMA journal_mode=WAL`

## Implementation Flow

### CLI Execution
1. If sync disabled or using local-only backend: perform operation directly, no daemon needed
2. If sync enabled with remote backend:
   a. Perform local operation immediately (fast return to user)
   b. Insert sync task into sync_queue with status='pending'
   c. Check if pidfile exists and process is alive
   d. Attempt IPC connection to daemon
   e. If connection succeeds: send "new task" notification, daemon resets timer
   f. If connection fails: launch new daemon, establish connection, send notification

### Daemon Lifecycle
1. Acquire exclusive lock on pidfile (exit if already locked)
2. Create Unix socket for IPC
3. Start 5-second idle timer
4. Enter main loop:
   - Listen for IPC notifications (non-blocking)
   - When notified: reset 5-second timer
   - Attempt to claim next pending task (atomic UPDATE)
   - If claim succeeds: process task, mark complete
   - If claim fails: continue waiting
   - If timer expires: cleanup and exit
5. On exit: remove pidfile and socket

### Task Processing
```sql
-- After successful processing
BEGIN IMMEDIATE;
UPDATE sync_queue
SET status = 'completed',
    completed_at = CURRENT_TIMESTAMP
WHERE id = ?;
COMMIT;
```

## Race Condition Prevention

### Scenario: CLI thinks daemon dead, but it exists
- Pidfile check confirms process alive → don't launch new daemon
- Even if pidfile stale, new daemon fails to acquire lock → exits immediately

### Scenario: Two daemons briefly exist due to bug
- Atomic task claiming ensures each task processed exactly once
- Daemon that fails to claim (0 rows updated) skips that task

### Scenario: Daemon exits between CLI check and task insert
- CLI inserts task, attempts notification, connection fails
- CLI launches new daemon, which picks up task from database

## Hung Daemon Detection and Recovery

### Problem
Daemon may hang or loop infinitely due to:
- Network timeouts to Nextcloud/Todoist
- Unhandled exceptions in task processing
- Infinite error retry loops
- Deadlocks or blocking I/O
- Resource exhaustion

### Heartbeat Mechanism (Implemented)

The daemon writes timestamps to a heartbeat file at a configurable interval. The `todoat sync daemon status` command checks this heartbeat to detect hung daemons.

**Heartbeat file location:**
- `$XDG_RUNTIME_DIR/todoat/daemon.heartbeat` (preferred)
- `/tmp/todoat-daemon-<UID>.heartbeat` (fallback)

**Configuration:**
```yaml
sync:
  daemon:
    heartbeat_interval: 5  # seconds (default: 5)
```

**Status output with heartbeat:**
```
Sync daemon is running
  PID: 12345
  Interval: 60 seconds
  Sync count: 5
  Last sync: 2026-01-30T10:15:00Z
  Heartbeat: healthy
```

A heartbeat is considered stale if older than 2x the configured interval, indicating the daemon may be hung:

```
  Heartbeat: UNHEALTHY - heartbeat is stale - daemon may be hung
```

### CLI Health Check

The `status` command reads the heartbeat file and compares the timestamp:

```go
// Check heartbeat health
timeSinceHeartbeat := time.Since(lastHeartbeat)
if timeSinceHeartbeat > 2 * heartbeatInterval {
    // Daemon is likely hung
    return "UNHEALTHY - heartbeat is stale"
}
return "healthy"
```

### Force Kill Command

When the daemon appears hung (stale heartbeat or unresponsive), use the kill command:

```bash
todoat sync daemon kill
```

This sends SIGTERM, waits briefly, then sends SIGKILL if needed, and cleans up the PID file and socket.

### Planned Enhancements

> The following features are not yet implemented but are planned for future versions.

#### Per-Task Timeout
```go
// Planned: In daemon task processing loop
const MaxTaskDuration = 5 * time.Minute

func processTaskWithTimeout(task SyncTask) error {
    ctx, cancel := context.WithTimeout(context.Background(), MaxTaskDuration)
    defer cancel()
    // ... process with timeout
}
```

#### Stuck Task Detection
```sql
-- Planned: CLI or monitoring can detect stuck tasks
SELECT id, worker_id, claimed_at
FROM sync_queue
WHERE status = 'processing'
  AND claimed_at < datetime('now', '-10 minutes');

-- Planned: Reset stuck tasks (if daemon confirmed dead/hung)
UPDATE sync_queue
SET status = 'pending',
    worker_id = NULL,
    claimed_at = NULL
WHERE id IN (stuck_task_ids);
```

### Graceful Shutdown Signal

#### Signal Handling in Daemon
```go
// Already partially implemented in todoat via stopChan/doneChan
func setupSignalHandlers(daemon *daemonState) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

    go func() {
        <-sigChan
        log.Println("Shutdown signal received, finishing current task...")
        close(daemon.stopChan)
    }()
}
```

#### CLI Force Kill Command
```bash
# User can manually kill hung daemon if needed (emergency only)
todoat daemon kill

# Implementation:
# 1. Read PID from pidfile
# 2. Send SIGTERM (graceful)
# 3. Wait 5 seconds
# 4. Send SIGKILL if still alive
# 5. Cleanup pidfile and socket
# 6. Reset any stuck tasks in sync_queue
```

> **Note**: The `todoat sync daemon start` command exists and works. The daemon can also be started automatically when the CLI detects pending background work and no running daemon.

### Planned Error Loop Prevention

#### Exponential Backoff on Errors
```go
// todoat sync_queue already has retry_count, but needs enforcement
const MaxConsecutiveErrors = 5

func (d *daemon) processLoop() {
    consecutiveErrors := 0

    for {
        select {
        case <-d.stopChan:
            return
        default:
        }

        task, err := d.claimNextTask()
        if err != nil {
            consecutiveErrors++
            log.Printf("Task claim error: %v", err)

            if consecutiveErrors >= MaxConsecutiveErrors {
                log.Println("Too many consecutive errors, shutting down daemon")
                return
            }

            // Exponential backoff: 1s, 2s, 4s, 8s, 16s (max 60s)
            backoff := time.Duration(math.Min(math.Pow(2, float64(consecutiveErrors)), 60)) * time.Second
            time.Sleep(backoff)
            continue
        }

        if task != nil {
            if err := d.processTask(task); err != nil {
                consecutiveErrors++
            } else {
                consecutiveErrors = 0 // Reset on success
            }
        }
    }
}
```

## Configuration

> **PARTIALLY IMPLEMENTED** — The daemon uses `PIDPath`, `SocketPath`, `LogPath`, `Interval`, `IdleTimeout`, `ConfigPath`, `DBPath`, and `CachePath` from its Config struct. The heartbeat, task timeout, and error threshold settings below are **not yet implemented**.

todoat-specific paths and values:

```yaml
# Daemon configuration (proposed additions to config.yaml)
daemon:
  pidfile: "$XDG_RUNTIME_DIR/todoat/daemon.pid"  # or /tmp/todoat-daemon.pid
  socket: "$XDG_RUNTIME_DIR/todoat/daemon.sock"
  logfile: "$XDG_DATA_HOME/todoat/daemon.log"
  idle_timeout: "5s"
  heartbeat_interval: "2s"
  heartbeat_timeout: "30s"
  task_timeout: "5m"
  max_consecutive_errors: 5

sync:
  # Existing config
  enabled: true
  auto_sync_after_operation: true
  background_pull_cooldown: "30s"
  # New daemon-related
  daemon_enabled: false  # Feature flag for new architecture
```

## Error Handling

### Database Locked
- Retry with exponential backoff (max 3 attempts)
- Log contention events for monitoring
- todoat already uses WAL mode which helps with this

### IPC Connection Timeout
- 500ms timeout on connection attempt
- Fallback to daemon launch if timeout or connection refused

### Stale Pidfile
- Verify process exists with `syscall.Kill(pid, 0)`
- Remove pidfile if process dead, launch new daemon

### Hung Daemon (Not Yet Implemented)
- Detect via heartbeat timeout (30 seconds)
- Send SIGTERM for graceful shutdown
- Send SIGKILL after 5 second grace period
- Reset stuck tasks back to pending status
- Launch new daemon to continue processing

### Task Processing Errors (Not Yet Implemented)
- Log error with full context
- Increment consecutive error counter
- Apply exponential backoff
- Shutdown daemon after 5 consecutive errors
- Mark failed tasks for manual review

### Backend-Specific Errors
- Nextcloud/CalDAV connection failures
- Todoist API rate limiting
- Network connectivity issues
- Use circuit breaker pattern per backend

## Integration with Existing todoat Components

### Notification Manager
The daemon should integrate with the existing `NotificationManager`:

```go
// Already exists in internal/notification/manager.go
daemon.notifyMgr.SendAsync(notification.Notification{
    Type:      notification.NotifySyncComplete,
    Title:     "todoat sync",
    Message:   fmt.Sprintf("Sync completed (count: %d)", syncCount),
    Timestamp: time.Now(),
})
```

### Sync Queue

> **NOT YET IMPLEMENTED** — The schema additions below are planned but not applied. The current `sync_queue` table does not have `status`, `worker_id`, `claimed_at`, or `completed_at` columns.

Planned: use existing `sync_queue` table with schema additions:

```sql
-- Planned: Add columns to existing sync_queue table
ALTER TABLE sync_queue ADD COLUMN status TEXT DEFAULT 'pending';
ALTER TABLE sync_queue ADD COLUMN worker_id TEXT;
ALTER TABLE sync_queue ADD COLUMN claimed_at TEXT;
ALTER TABLE sync_queue ADD COLUMN completed_at TEXT;

-- Planned: Add index for efficient task claiming
CREATE INDEX idx_sync_queue_status ON sync_queue(status);
```

### Multi-Backend Support (Future)
Per roadmap item 073, the daemon should eventually support:
- Iterating through all sync-enabled backends
- Per-backend sync state tracking
- Backend-specific failure isolation
- Per-backend configurable intervals

## Benefits of This Approach

- **No race conditions:** Atomic operations prevent double-execution
- **Efficient:** Daemon stays alive during bursts, exits when idle
- **Simple:** Single daemon model, no complex coordination
- **Reliable:** Direct communication eliminates polling delays
- **Safe:** Multiple layers of protection against edge cases
- **Recoverable:** Hung daemon detection and automatic recovery
- **Observable:** Heartbeat and logging provide visibility into daemon health
- **Resilient:** Error handling prevents infinite loops and cascading failures
- **Responsive CLI:** Users get immediate feedback, sync happens in background

## Problems and Considerations for todoat

### Migration Path
1. Current in-process goroutine approach must continue working during transition
2. Feature flag (`daemon.daemon_enabled`) to toggle between old and new architecture
3. Database migrations needed for new columns/tables

### Platform Compatibility
- Unix sockets work on Linux/macOS but not Windows
- Windows would need named pipes or TCP localhost
- Pidfile semantics differ across platforms

### Testing Complexity
- Daemon tests need to spawn real processes
- Race condition tests are timing-sensitive
- Current test infrastructure uses in-process testing

### User Experience
- Daemon starts automatically - no manual `start` command needed
- Users may need `todoat daemon kill` for recovery from hung state
- `todoat daemon status` could show daemon health (optional)

## Related Issues and Roadmap Items

- Issue #36: Sync not truly in background - needs background daemon
- Issue #37: Subtasks not properly synced to Nextcloud backend
- Issue #38: Config parsing errors should be logged, not silent
- Roadmap 024: Auto-Sync Daemon (original, partially implemented)
- Roadmap 073: Auto-Sync Daemon Redesign (multi-backend support)
