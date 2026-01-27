# [015] Task Trash and Undo Delete

## Summary
Add a soft-delete mechanism for tasks similar to the existing list trash feature, allowing users to recover accidentally deleted tasks.

## Source
Documentation gap: `docs/explanation/task-management.md` explicitly states "deletion is permanent and cannot be undone (unlike list deletion, which has trash/restore functionality)". The asymmetry between list and task deletion behavior creates a data loss risk.

## Motivation
Users can accidentally delete tasks, especially when:
1. Using bulk delete with filters that match more than expected
2. Mistyping a task name in the delete command
3. Intending to complete but pressing delete instead
4. Script errors in automated workflows

Lists already have trash/restore functionality (`todoat list trash`, `todoat list restore`), but tasks lack this safety net. This inconsistency is unintuitive and creates unnecessary data loss risk.

## Current Behavior
```bash
# Task deletion is permanent
todoat Work delete "Important task"
# Cannot be recovered - task and all history are lost

# Compare to lists - which have trash:
todoat list delete "MyList"
todoat list trash           # See deleted lists
todoat list restore "MyList"  # Recover deleted list
```

## Proposed Behavior
```bash
# Soft-delete a task (moves to trash)
todoat Work delete "Accidental task"
# Output: Task moved to trash. Use 'todoat trash' to view or restore.

# View trashed tasks
todoat trash
# or
todoat Work --trash
# Output:
# Trash (2 tasks):
#   [7d ago] Important task (from: Work)
#   [2d ago] Another task (from: Personal)

# Restore a task from trash
todoat trash restore "Important task" --to Work

# Permanently delete (purge)
todoat trash purge "Old task"

# Empty all trash older than N days
todoat trash empty --older-than 30d

# Force permanent deletion (bypass trash)
todoat Work delete "Task" --permanent
# or
todoat Work delete "Task" -f

# Auto-purge configuration
# config.yaml:
#   trash:
#     auto_purge_after_days: 30
```

## Estimated Value
high - Prevents data loss, consistent with list behavior, standard feature in most task apps

## Estimated Effort
S - Mirrors existing list trash implementation, adds `deleted_at` column to tasks table, minimal new commands

## Related
- List trash system: `todoat list trash`, `todoat list restore`
- Idea #008 (Task Archival) - different purpose: archiving completed tasks for history, vs trash for recovering deleted tasks
- Bulk operations: docs/explanation/task-management.md#bulk-operations

## Status
unreviewed
