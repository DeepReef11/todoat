# [009] Auto-Sync Not Triggering After Task Operations

## Type
code-bug

## Severity
high

## Source
User report

## Description

With `sync.enabled: true` in config (which should enable auto-sync by default), operations like `todoat add` do not automatically trigger sync. User must manually run `todoat sync` to push changes.

## Steps to Reproduce

```bash
# Config has sync enabled:
# sync:
#   enabled: true
#   local_backend: sqlite
#   conflict_resolution: server_wins
#   offline_mode: auto

# Add a task
todoat -b todoist Inbox add "11"
# Output: Created task: 11 (ID: f68132bb-ce6c-4434-8195-c68149c39f44)
# No sync happens

# Check queue - operation is queued, not synced
todoat -b todoist sync queue
# Pending Operations: 1
# ID     Type       Task                           Retries  Created
# 1      create     11                             0        15:11:05

# Must manually sync
todoat -b todoist sync
# Now it syncs
```

## Expected Behavior

With `sync.enabled: true`:
- After `todoat add`, sync should automatically run
- Task should appear on remote immediately (or within short delay)
- Queue should be empty after operation

## Actual Behavior

- Task added to local database only
- Operation queued but not executed
- Manual `todoat sync` required
- `sync.enabled: true` does not enable auto-sync

## Documentation Claims

From sync documentation, `sync.enabled: true` should enable sync functionality. If auto-sync requires additional config like `auto_sync_after_operation: true`, this should be:
1. Documented clearly
2. Possibly default to true when `sync.enabled: true`

## Code Location

- `cmd/todoat/cmd/todoat.go` - task operation handlers (add, update, delete, complete)
- Config key `sync.auto_sync_after_operation` may exist but not be honored

## Root Cause Analysis

Investigation found this is NOT a bug - the code works as designed, but the UX/documentation is confusing:

1. `sync.enabled: true` enables the sync **infrastructure** (queue, sync command)
2. `auto_sync_after_operation: true` enables **automatic** sync after each operation
3. By default, `auto_sync_after_operation: false` - user must run `todoat sync` manually

The documentation at `docs/reference/configuration.md:136` clearly states:
> `sync.auto_sync_after_operation` - Auto-sync after add/update/delete operations (default: `false`)

However, the user expectation is that `sync.enabled: true` means "sync is enabled" (including auto-sync), not just "sync infrastructure is enabled".

## Expected Behavior (User Request)

**Auto-sync should be enabled by default when `sync.enabled: true`** and `auto_sync_after_operation` is not explicitly set.

Current behavior requires two settings which is unintuitive:
```yaml
sync:
  enabled: true                    # This alone should enable auto-sync
  auto_sync_after_operation: true  # Shouldn't be required
```

## Required Fix

FIX CODE - When `sync.enabled: true` and `auto_sync_after_operation` is not explicitly set to `false`, auto-sync should be enabled by default.

Logic should be:
```go
// In IsAutoSyncAfterOperationEnabled()
func (c *Config) IsAutoSyncAfterOperationEnabled() bool {
    // If sync is enabled and auto_sync_after_operation is not explicitly false,
    // default to true (auto-sync enabled)
    if c.Sync.Enabled && !c.Sync.AutoSyncAfterOperationExplicitlyFalse {
        return true
    }
    return c.Sync.AutoSyncAfterOperation
}
```

Or simpler: Change the default value of `AutoSyncAfterOperation` to `true`.

## Workaround (Until Fixed)

User can add to their config:
```yaml
sync:
  enabled: true
  auto_sync_after_operation: true  # Shouldn't be needed but is for now
```

## Related Files

- `cmd/todoat/cmd/todoat.go` - `doAdd()`, `doUpdate()`, `doComplete()`, `doDelete()`
- `internal/config/config.go` - sync config structure
- `docs/how-to/sync.md` or `docs/explanation/synchronization.md`
