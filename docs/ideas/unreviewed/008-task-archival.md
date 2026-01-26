# [008] Task Archival System

## Summary
Add ability to archive completed tasks instead of deleting them, preserving history while keeping active lists clean.

## Source
Code analysis: The app has list trash (`todoat list trash`) for deleted lists, but no equivalent for individual tasks. Users who want to keep a record of completed work must either leave DONE tasks cluttering their lists or delete them permanently losing the history.

## Motivation
Users often want to:
1. Review what they accomplished in a time period
2. Reference past tasks for information
3. Keep lists focused on active work
4. Maintain a historical record without visual clutter

Currently, completed tasks either stay visible (clutter) or get deleted (lost history). An archive provides a middle ground.

## Current Behavior
```bash
# Completed tasks stay in the list forever
todoat Work -s DONE
# Shows all completed tasks mixed with active work

# Or delete permanently
todoat Work delete "Completed task"
# Task history is lost
```

## Proposed Behavior
```bash
# Archive a completed task
todoat Work archive "Review PR #123"

# Archive all completed tasks
todoat Work archive --all --filter-status DONE

# View archived tasks
todoat Work --archived
# or
todoat archive Work

# Search archives
todoat archive search "PR"

# Restore from archive
todoat archive restore "Review PR #123" --to Work

# Auto-archive completed tasks older than N days
# In config:
# archive:
#   auto_archive_completed_after_days: 30
```

## Estimated Value
medium - Common need in task management, bridges gap between clutter and lost history

## Estimated Effort
M - Requires new archive storage (could be separate SQLite table or file), archive commands, auto-archive daemon integration

## Open Questions
- Store archives in main SQLite or separate archive database?
- Should archives sync to remote backends?
- Per-list archives or global archive?
- Retention policy for archives (auto-delete after N months)?
- Include in analytics/reporting?

## Related
- List trash system: `todoat list trash`
- Sync daemon: could trigger auto-archive

## Status
unreviewed
