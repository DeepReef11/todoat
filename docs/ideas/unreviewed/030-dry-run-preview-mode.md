# [030] Dry-Run Preview Mode

## Summary
Add a global `--dry-run` flag that shows what changes a command would make without executing them, enabling safe verification before destructive or batch operations.

## Source
Code analysis: Several commands modify state (delete, update, sync push, complete) with no preview capability. Idea #007 (Bulk Operations) mentions dry-run for bulk delete specifically, but a global dry-run mode would benefit all modifying operations and provide consistency.

## Motivation
Users need confidence before:
- Deleting tasks (especially with filters)
- Bulk updates that could affect many tasks
- Sync operations that push local changes to remote
- Import operations that add many tasks
- Any command in a script before running in production

A global dry-run flag provides consistent preview behavior across all commands.

## Current Behavior
```bash
# Delete happens immediately - no preview
todoat Work delete "important task"
# Task is gone

# Sync push - no way to see what will be sent
todoat sync push
# Changes sent to remote immediately

# Bulk operations (proposed) - would need their own dry-run
todoat Work complete --all --filter-status TODO
# All tasks completed immediately
```

## Proposed Behavior
```bash
# Global flag works on any modifying command
todoat --dry-run Work delete "task name"
# Output:
# [DRY RUN] Would delete task:
#   Summary: task name
#   UID: abc123
#   Status: TODO
#   No changes made.

# Works with updates
todoat --dry-run Work update "task" --priority 1 --tag urgent
# Output:
# [DRY RUN] Would update task 'task':
#   Priority: 5 -> 1
#   Tags: [] -> [urgent]
#   No changes made.

# Works with bulk operations
todoat --dry-run Work complete --all --filter-tag review
# Output:
# [DRY RUN] Would complete 7 tasks:
#   - Code review PR #42       (TODO -> DONE)
#   - Review design doc        (TODO -> DONE)
#   - Review budget proposal   (IN-PROGRESS -> DONE)
#   ...
#   No changes made.

# Works with sync
todoat --dry-run sync push
# Output:
# [DRY RUN] Would push to 'nextcloud':
#   Create: 2 tasks
#     - New task 1
#     - New task 2
#   Update: 3 tasks
#     - Modified task A (priority changed)
#     - Modified task B (status changed)
#     - Modified task C (due date changed)
#   Delete: 1 task
#     - Deleted task X
#   No changes made.

# Works with import
todoat --dry-run Work import tasks.ics
# Output:
# [DRY RUN] Would import 15 tasks from tasks.ics:
#   - Task 1 (priority 1, due 2026-02-15)
#   - Task 2 (priority 3, no due date)
#   ...
#   No changes made.

# Short form
todoat -n Work delete "task"

# JSON output for scripting
todoat --dry-run --json Work delete "task"
# {"dry_run": true, "action": "delete", "tasks": [{"uid": "abc123", ...}]}
```

## Estimated Value
medium - Safety feature that builds trust; essential for scripting and automation

## Estimated Effort
S - Pattern can be implemented at command execution layer; modifying commands check flag and return preview instead of executing

## Related
- Idea #007 (Bulk Operations) - mentions dry-run for bulk delete; this generalizes it
- Sync system: `backend/sync/` would benefit from push preview
- Import: `cmd/todoat/cmd/todoat.go` import commands
- Delete confirmation: some commands already have `-y` flag; dry-run is complementary

## Status
unreviewed
