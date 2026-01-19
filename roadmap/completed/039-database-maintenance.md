# [039] Database Maintenance Commands

## Summary
Implement database maintenance commands for SQLite backend including statistics display and vacuum/optimization, helping users monitor storage usage and reclaim disk space.

## Documentation Reference
- Primary: `dev-doc/LIST_MANAGEMENT.md` (Database Statistics, Vacuum/Optimize sections)
- Related: `dev-doc/BACKEND_SYSTEM.md`

## Dependencies
- Requires: [003] SQLite Backend
- Requires: [007] List Commands

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestListStats` - `todoat list stats` displays database statistics
- [ ] `TestListStatsJSON` - `todoat list stats --json` returns JSON statistics
- [ ] `TestListVacuum` - `todoat list vacuum` reclaims space from deleted data
- [ ] `TestListVacuumConfirmation` - Vacuum prompts for confirmation (skipped with -y)

### Functional Requirements
- [ ] `todoat list stats` command displays:
  - Total tasks across all lists
  - Tasks per list (name and count)
  - Tasks by status (TODO, DONE, PROCESSING, CANCELLED counts)
  - Database file size (human-readable: KB, MB)
  - Index usage statistics
  - Last vacuum timestamp (if available)
- [ ] `todoat list stats "<name>"` shows stats for specific list
- [ ] `todoat list vacuum` command:
  - Runs SQLite VACUUM command
  - Rebuilds database file to reclaim space
  - Shows before/after file size comparison
  - Requires confirmation (destructive in sense of changing file)
- [ ] Vacuum can be run with `-y` to skip confirmation
- [ ] Statistics cached briefly (30s) to avoid repeated queries

### Output Requirements
- [ ] Stats output in formatted table for text mode
- [ ] JSON mode returns structured statistics object:
  ```json
  {
    "result": "INFO_ONLY",
    "stats": {
      "total_tasks": 150,
      "lists": [{"name": "Work", "count": 80}],
      "by_status": {"TODO": 100, "DONE": 40, "PROCESSING": 5, "CANCELLED": 5},
      "database_size_bytes": 1048576,
      "last_vacuum": "2026-01-15T10:30:00Z"
    }
  }
  ```
- [ ] Vacuum output shows space reclaimed

## Implementation Notes
- Use `PRAGMA page_count` and `PRAGMA page_size` for database size
- Use `PRAGMA freelist_count` to show reclaimable pages before vacuum
- Store last_vacuum timestamp in schema_version or separate metadata table
- VACUUM requires exclusive lock - warn if database in use
- Consider `VACUUM INTO` for safer operation (SQLite 3.27+)

## Out of Scope
- Automatic scheduled vacuum
- Database integrity checks (PRAGMA integrity_check)
- Index rebuilding (separate from VACUUM)
- Database backup before vacuum
