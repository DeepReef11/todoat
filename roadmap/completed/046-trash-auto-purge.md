# [046] Trash Auto-Purge

## Summary
Implement automatic purging of lists from trash after a configurable retention period (default: 30 days), reducing manual cleanup and preventing indefinite data accumulation.

## Documentation Reference
- Primary: `docs/explanation/list-management.md`
- Section: Trash and Restore Lists - Edge Cases (#484)

## Dependencies
- Requires: [013] List Management (trash/restore functionality exists)

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestTrashAutoPurgeDefault` - Lists deleted >30 days ago are automatically purged on next `todoat list trash` command
- [ ] `TestTrashAutoPurgeConfigurable` - `trash.retention_days: 7` in config purges lists older than 7 days
- [ ] `TestTrashAutoPurgeDisabled` - `trash.retention_days: 0` disables auto-purge
- [ ] `TestTrashAutoPurgePreservesRecent` - Lists deleted <30 days ago are NOT purged

### Functional Requirements
- [ ] Auto-purge runs lazily when trash is accessed (list trash, restore, etc.)
- [ ] Purged lists and tasks are permanently deleted (no recovery)
- [ ] Log message indicates items purged during auto-cleanup
- [ ] Default retention is 30 days if not configured

## Implementation Notes
- Add `trash.retention_days` config option (default: 30, 0 = disabled)
- Purge logic: `DELETE FROM task_lists WHERE deleted_at < NOW() - retention_days`
- Cascade delete removes associated tasks automatically (FK constraint)
- Consider adding `--dry-run` flag to preview what would be purged

## Out of Scope
- Background scheduler for auto-purge (runs on-demand only)
- Per-list retention policies
- Undo purge functionality
