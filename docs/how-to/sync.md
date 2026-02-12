# Synchronization

todoat supports offline-first task management with synchronization to remote backends. This guide covers configuring and using sync.

## Overview

When sync is enabled:
- All operations use a local cache for instant response
- Changes are queued for synchronization
- Run `todoat sync` to push local changes and pull remote updates
- Conflicts are resolved automatically based on your chosen strategy

## Enabling Sync

Add to your `config.yaml`:

```yaml
sync:
  enabled: true
  local_backend: sqlite
  conflict_resolution: server_wins
  offline_mode: auto
```

## Basic Sync Commands

### Full Sync

```bash
todoat sync
```

Pulls changes from remote backends and pushes local changes.

### Sync Status

```bash
todoat sync status
```

Shows:
- Last sync time
- Pending operations count
- Current sync state per backend

For JSON output:

```bash
todoat --json sync status
```

```json
{
  "offline_mode": "auto",
  "backends": [
    {
      "name": "nextcloud",
      "last_sync": "2026-01-30 10:15:00",
      "pending_operations": 2,
      "status": "Configured"
    }
  ],
  "result": "INFO_ONLY"
}
```

### View Sync Queue

```bash
todoat sync queue
```

Shows operations waiting to be synced.

For JSON output:

```bash
todoat --json sync queue
```

### Clear Sync Queue

```bash
todoat sync queue clear
```

Removes all pending sync operations from the queue. Use this when you want to discard unsynced local changes.

### View Conflicts

```bash
todoat sync conflicts
```

Lists unresolved sync conflicts requiring manual attention.

### Resolve a Conflict

```bash
todoat sync conflicts resolve [task-uid] --strategy local_wins
todoat sync conflicts resolve [task-uid] --strategy server_wins
# Also available: --strategy merge, --strategy keep_both
```

Manually resolve a specific conflict using the specified strategy.

## Background Sync Daemon

The sync daemon runs as a separate background process that periodically synchronizes tasks with remote backends. It uses a forked process architecture with Unix socket IPC, so the daemon continues running independently of the CLI.

### How the Daemon Works

When you start the daemon:
1. A new process is forked and detached from the terminal
2. The daemon creates a PID file and Unix domain socket for communication
3. It periodically syncs all configured remote backends at the configured interval
4. Normal CLI operations (add, update, delete) send an IPC notification to the daemon, triggering an immediate sync
5. If idle for the configured timeout period, the daemon exits automatically

The daemon syncs all configured remote backends with per-backend failure isolation using a circuit breaker pattern. If one backend fails repeatedly, it is temporarily skipped while other backends continue syncing normally.

### Start Daemon

```bash
todoat sync daemon start
```

Starts background sync with the default interval (300 seconds / 5 minutes).

To use a custom sync interval:

```bash
# Sync every 60 seconds
todoat sync daemon start --interval 60
```

### Check Daemon Status

```bash
todoat sync daemon status
```

For JSON output:

```bash
todoat --json sync daemon status
```

Shows daemon status including PID, sync interval, sync count, last sync time, and heartbeat health:

```
Sync daemon is running
  PID: 12345
  Interval: 60 seconds
  Sync count: 5
  Last sync: 2026-01-30T10:15:00Z
  Heartbeat: healthy
```

If heartbeat detection is enabled (via `sync.daemon.heartbeat_interval`), the status shows whether the daemon is responsive. A stale heartbeat indicates the daemon may be hung:

```
Sync daemon is running
  PID: 12345
  Interval: 60 seconds
  Sync count: 5
  Last sync: 2026-01-30T10:15:00Z
  Heartbeat: UNHEALTHY - heartbeat is stale - daemon may be hung
```

The interval shown is the actual running interval, which may differ from the config default if `--interval` was specified at start time.

### Stop Daemon

```bash
todoat sync daemon stop
```

Sends a stop request to the daemon via IPC. The daemon finishes any in-progress sync and exits gracefully.

### Force Kill Daemon

```bash
todoat sync daemon kill
```

Force kills the daemon process. Use this for emergency termination if the daemon is hung and won't respond to the normal stop command. This sends SIGTERM, waits briefly, then sends SIGKILL if needed, and cleans up the PID file and socket.

### Error Recovery

The daemon automatically handles transient errors with exponential backoff:

- After a failed sync, the daemon waits before retrying (2s, 4s, 8s, 16s, up to 60s)
- After 5 consecutive sync failures, the daemon shuts down automatically
- A successful sync resets the error counter

This prevents the daemon from spinning in a tight loop when the remote backend is unavailable (e.g., network outage, server maintenance).

### Per-Backend Circuit Breaker

When using multiple backends, each backend has its own circuit breaker. After 3 consecutive sync failures for a backend, the circuit "opens" and that backend is temporarily skipped for 30 seconds. After the cooldown, a single probe sync is attempted:

- **Success**: The circuit closes and normal syncing resumes
- **Failure**: The circuit opens again for another cooldown period

This means a flaky Nextcloud server won't prevent your Todoist tasks from syncing. You can see the circuit state per backend in the daemon status JSON output:

```bash
todoat --json sync daemon status
```

The `circuit_state` field shows `closed` (normal), `open` (skipping), or `half-open` (probing after cooldown).

To restart the daemon after it shuts down due to errors:

```bash
todoat sync daemon start
```

### Stuck Task Recovery

If the daemon crashes or is killed while processing a sync task, that task can become "stuck" in the processing state. The daemon automatically detects and recovers stuck tasks:

- Tasks in `processing` state for longer than the stuck timeout are considered stuck
- The daemon validates whether the original worker is still alive before recovery
- Stuck tasks are reset to `pending` and will be retried

Configure the stuck timeout (default: 10 minutes):

```bash
# Via CLI flag
todoat sync daemon start --stuck-timeout 15

# Via config set
todoat config set sync.daemon.stuck_timeout 15
```

Or in `config.yaml`:

```yaml
sync:
  daemon:
    stuck_timeout: 15  # minutes
```

### Per-Task Timeout

Each individual sync operation has a configurable timeout to prevent hung operations from blocking the sync queue indefinitely. If a task exceeds the timeout, it's cancelled and retried later.

Configure the per-task timeout (default: 5 minutes):

```bash
todoat config set sync.daemon.task_timeout "5m"
```

Or in `config.yaml`:

```yaml
sync:
  daemon:
    task_timeout: "5m"  # 5 minutes (default)
```

The `task_timeout` value uses Go duration format (e.g., `"30s"`, `"2m"`, `"10m"`).

### Daemon Configuration

Configure daemon behavior in `config.yaml`:

```yaml
sync:
  daemon:
    enabled: false        # Enable daemon process
    interval: 300         # Sync interval in seconds (default: 5 minutes)
    idle_timeout: 300     # Seconds before idle daemon exits (default: 5 minutes)
    heartbeat_interval: 5 # Heartbeat recording interval in seconds (default: 5)
    stuck_timeout: 10     # Minutes before a task is considered stuck (default: 10)
    task_timeout: "5m"    # Per-task timeout for sync operations (default: 5m)
```

The `--interval` flag on `sync daemon start` overrides the `interval` config value for that session.

The `heartbeat_interval` enables hung daemon detection. When set to a positive value, the daemon writes a timestamp to a heartbeat file at the specified interval. The `status` command checks this heartbeat and reports if the daemon appears hung (heartbeat older than 2x the interval).

The daemon stores its state files at:
- **PID file**: `$XDG_RUNTIME_DIR/todoat/daemon.pid` (or `/tmp/todoat-daemon-<UID>.pid`)
- **Socket**: `$XDG_RUNTIME_DIR/todoat/daemon.sock` (or `/tmp/todoat-daemon-<UID>.sock`)

When `$XDG_RUNTIME_DIR` is not set, the fallback paths include the user's numeric UID to prevent conflicts between users on shared systems.

## Sync Configuration Options

### enabled

```yaml
sync:
  enabled: true  # or false
```

When `true`, remote backends are cached locally for offline access.

### Background Pull Sync on Read Operations

When `auto_sync_after_operation` is enabled, todoat automatically triggers a background pull sync whenever you perform read operations like listing tasks or lists:

```bash
# These commands trigger a background pull sync
todoat MyList              # Lists tasks
todoat list                # Lists all lists
todoat MyList -s TODO      # Filtered task listing
```

The background sync:
- Runs in the background without blocking the command
- Only pulls changes from remote (never pushes local changes)
- Has a configurable cooldown (default: 30 seconds) to prevent excessive network requests
- Ensures you see fresh data from remote backends

#### Configuring the Cooldown

The background pull cooldown can be adjusted via configuration:

```yaml
sync:
  background_pull_cooldown: "30s"  # default, minimum: 5s
```

Use shorter values (e.g., `"10s"`) for faster connections, or longer values (e.g., `"2m"`) for metered connections.

This means your task list automatically stays up-to-date as you use the CLI, without needing to manually run `todoat sync` before each read.

**Note**: Write operations (add, update, delete) trigger a full sync immediately when `auto_sync_after_operation` is enabled, not just a pull.

### auto_sync_after_operation

```yaml
sync:
  auto_sync_after_operation: true  # default when sync.enabled: true
```

When `true` (the default when sync is enabled), operations (add, update, delete) automatically trigger a sync to push changes to the remote backend immediately. This eliminates the need to run `todoat sync` manually after each operation.

When set to `false`, changes are queued locally and only pushed when you explicitly run `todoat sync` or when the sync daemon runs.

### local_backend

```yaml
sync:
  local_backend: sqlite  # Only option currently
```

The storage type for local cache.

### conflict_resolution

```yaml
sync:
  conflict_resolution: server_wins
```

How to handle conflicts when the same task is modified both locally and remotely:

| Strategy | Behavior |
|----------|----------|
| `server_wins` | Remote/server changes override local (default) |
| `local_wins` | Local changes override remote |
| `merge` | Combine changes from both versions |
| `keep_both` | Keep both versions as separate tasks |

### offline_mode

```yaml
sync:
  offline_mode: auto
```

| Mode | Behavior |
|------|----------|
| `auto` | CLI always uses SQLite cache (default) |
| `offline` | CLI always uses SQLite cache (explicit offline preference) |
| `online` | CLI uses remote backend directly (bypasses sync) |

## How Sync Works

### Offline-First Architecture

When sync is enabled with `offline_mode: auto` (the default), all CLI operations use the local SQLite cache for instant response. No network calls are made during normal CLI usage.

```bash
todoat -b nextcloud MyList add "New task"
```

The task is immediately saved to the local SQLite cache:

```
Created task: New task (ID: abc123)
```

Operations are queued automatically. Run `todoat sync` when you're ready to push changes to the remote backend.

The `offline_mode` setting controls CLI behavior:
- `auto` (default): CLI always uses SQLite cache - operations are instant, sync happens separately
- `offline`: Same as auto - explicitly indicates offline-first preference
- `online`: CLI uses remote backend directly - bypasses the sync architecture entirely

### Adding a Task Offline

```bash
todoat MyList add "New task"
```

1. Task saved to local cache
2. Create operation added to sync queue
3. User sees immediate confirmation

### Syncing Later

```bash
todoat sync
```

1. Pull: Fetch remote changes, update local cache
2. Push: Send queued operations to remote
3. Resolve any conflicts

### Sync Output

```
Syncing with backend: nextcloud
Pull: 15 tasks updated, 3 new tasks, 1 deleted
Push: 5 local changes pushed
Conflicts: 2 (resolved with remote strategy)
Sync completed successfully
```

## Working Offline

### Daily Workflow

```bash
# Morning: Sync to get latest
todoat sync

# Throughout day: Work normally
todoat Work add "Meeting notes"
todoat Work complete "Review PR"
# Changes saved locally, queued for sync

# Before leaving: Sync to share updates
todoat sync status  # Check pending changes
todoat sync         # Push all changes
```

### Checking Pending Changes

```bash
todoat sync queue
```

Output:
```
Pending Operations: 3

ID  | Type   | Task Summary        | Retries
----+--------+---------------------+--------
1   | create | Meeting notes       | 0
2   | update | Review PR           | 0
3   | delete | Old reminder        | 0
```

## Conflict Resolution

### When Conflicts Occur

A conflict happens when:
1. You modify a task locally
2. The same task is modified remotely (e.g., on another device)
3. You sync

### Resolution Strategies

**Server Wins (Recommended for teams)**

```yaml
sync:
  conflict_resolution: server_wins
```

Remote/server version replaces local. Safe when remote is the source of truth.

**Local Wins**

```yaml
sync:
  conflict_resolution: local_wins
```

Local version kept, pushed to remote. Good for single-user offline work.

**Merge**

```yaml
sync:
  conflict_resolution: merge
```

Combine changes from both versions where possible. Useful when different fields were modified.

**Keep Both**

```yaml
sync:
  conflict_resolution: keep_both
```

Keep both versions as separate tasks. You decide which to keep later.

## Per-Backend Sync Settings

Disable sync for specific backends:

```yaml
backends:
  nextcloud:
    type: nextcloud
    enabled: true
    # Uses sync (default)

  local-only:
    type: sqlite
    enabled: true
    sync:
      enabled: false  # No sync for this backend
```

## Cache Location

Cache databases are stored at:

```
~/.local/share/todoat/caches/
├── nextcloud.db
├── todoist.db
└── ...
```

Each remote backend gets a separate cache.

## Troubleshooting

### Sync Fails

1. Check network connectivity
2. Verify credentials: `todoat credentials list`
3. Check sync status: `todoat sync status --verbose`
4. If sync times out on slow networks, increase the timeout:
   ```bash
   todoat config set sync.connectivity_timeout 15s
   ```
5. Retry: `todoat sync`

### Stuck Operations

Operations that fail repeatedly stay in queue:

```bash
# View queue with retry counts
todoat sync queue

# If max retries reached, check logs
todoat sync status --verbose
```

### Reset Sync

To start fresh (loses unsynced changes):

```bash
# Remove cache database
rm ~/.local/share/todoat/caches/nextcloud.db

# Re-sync
todoat sync
```

## Example Configurations

### Single Remote, Offline-First

```yaml
backends:
  nextcloud:
    type: nextcloud
    enabled: true
    host: "nextcloud.example.com"
    username: "user"

sync:
  enabled: true
  conflict_resolution: local_wins
  offline_mode: auto
```

### Team Setup (Server Authoritative)

```yaml
backends:
  work:
    type: nextcloud
    enabled: true
    host: "nextcloud.work.com"
    username: "team-member"

sync:
  enabled: true
  conflict_resolution: server_wins
  offline_mode: auto
```

### Multiple Remotes

```yaml
backends:
  work:
    type: nextcloud
    enabled: true
    host: "work.example.com"
    username: "user"

  personal:
    type: todoist
    enabled: true

sync:
  enabled: true
  conflict_resolution: server_wins
```

Each backend gets its own isolated cache.

## Performance

### Initial Sync

First sync downloads all tasks:
- 100 tasks: ~2 seconds
- 1000 tasks: ~15 seconds

### Incremental Sync

Subsequent syncs are faster:
- 50 changes: ~2 seconds
- Only changed tasks transferred

### Tips

- Sync regularly to keep incremental syncs fast
- Use filters in views to reduce displayed data
- Large task lists work best with incremental sync

## See Also

- [Backends](../explanation/backends.md) - Configuring remote backends
- [Getting Started](../tutorials/getting-started.md) - Initial setup
- [Task Management](task-management.md) - Working with tasks
