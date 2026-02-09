# [025] Task Move Between Lists

## Summary
Add a dedicated command to move tasks from one list to another, preserving all metadata, subtask relationships, and sync state.

## Source
Code analysis: Existing CLI update flags (`--parent`, `--no-parent`, tags, etc.) allow modifying task properties but there's no mechanism to change a task's list assignment. Users must currently delete from source list and re-add to target list, losing metadata like creation date, completion history, and recurrence state.

## Motivation
Moving tasks between lists is a common workflow need:
- Refactoring projects (splitting/merging lists)
- Changing task scope (personal to work, inbox to project)
- Reorganizing after GTD-style reviews
- Correcting task placement mistakes

Currently this requires delete + re-add, which:
- Loses task UID (breaks sync references)
- Loses creation timestamp
- Loses completion history
- Requires manual re-entry of all properties
- Breaks subtask hierarchies

## Current Behavior
No move capability exists. The workaround is:
```bash
# Current workaround (loses metadata)
todoat OldList delete "Task"
todoat NewList add "Task" -p 2 --due-date 2026-01-31 --tag work,urgent ...
```

## Proposed Behavior
```bash
# Move task to different list
todoat OldList move "Task" --to NewList

# Move with subtasks
todoat OldList move "Project/*" --to Archive

# Move using UID (for scripting)
todoat OldList move --uid <uuid> --to NewList

# Bulk move with filter
todoat Work move --status DONE --to Archive
```

The move operation would:
1. Preserve task UID, creation date, and all metadata
2. Update list assignment in backend
3. Handle subtasks (move tree together or optionally flatten)
4. Update sync metadata for remote backends
5. Support bulk moves with filters

## Estimated Value
high - Common workflow need, prevents data loss, enables proper reorganization

## Estimated Effort
M - Requires adding move logic to TaskManager interface, implementing per-backend, handling subtask trees and sync state

## Related
- [006-inbox-workflow.md](006-inbox-workflow.md) - Would use move for inbox processing
- [007-bulk-operations-cli.md](007-bulk-operations-cli.md) - Bulk move is a related pattern
- [008-task-archival.md](008-task-archival.md) - Archival could use move to archive list

## Status
unreviewed
