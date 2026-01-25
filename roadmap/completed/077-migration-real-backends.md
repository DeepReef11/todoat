# [077] Migration to Real Backends

## Summary
Complete the cross-backend migration feature by implementing migration support for real backends (Nextcloud, Todoist, File) instead of only mock/SQLite backends.

## Documentation Reference
- Primary: `docs/explanation/backends.md`
- Section: Notes "Migration to real backends (nextcloud, todoist, file) is not yet implemented"

## Dependencies
- Requires: [026] Cross-Backend Migration (completed - but only for SQLite/mock)
- Requires: [016] Nextcloud Backend (completed)
- Requires: [021] Todoist Backend (completed)
- Requires: [025] File Backend (completed)

## Complexity
L

## Acceptance Criteria

### Tests Required
- [ ] `TestMigrateSQLiteToNextcloud` - `todoat migrate --from sqlite --to nextcloud` moves tasks to real Nextcloud
- [ ] `TestMigrateSQLiteToTodoist` - `todoat migrate --from sqlite --to todoist` moves tasks to real Todoist
- [ ] `TestMigrateSQLiteToFile` - `todoat migrate --from sqlite --to file` moves tasks to file backend
- [ ] `TestMigrateNextcloudToSQLite` - `todoat migrate --from nextcloud --to sqlite` imports from Nextcloud
- [ ] `TestMigrateTodoistToSQLite` - `todoat migrate --from todoist --to sqlite` imports from Todoist
- [ ] `TestMigratePreservesMetadata` - Migrated tasks retain priority, dates, status, tags
- [ ] `TestMigratePreservesHierarchy` - Parent-child relationships preserved during migration
- [ ] `TestMigrateStatusMapping` - Statuses correctly mapped between backend formats
- [ ] `TestMigrateRateLimiting` - Migration respects API rate limits

### Functional Requirements
- Complete `getMigrateBackend()` in `cmd/todoat/cmd/todoat.go` for nextcloud, todoist, file backends
- Handle authentication via existing credential management system
- Map status values between backend-specific formats
- Preserve task hierarchy (parent-child relationships)
- Support dry-run mode for all backend combinations
- Progress reporting for large migrations

## Implementation Notes
- Reuse existing backend implementations from `backend/` packages
- Use credential resolution from `internal/credentials/`
- Handle rate limiting for Todoist/Google Tasks via `internal/ratelimit/`
- Map CalDAV statuses (NEEDS-ACTION, COMPLETED) to internal statuses (TODO, DONE)
- Consider batch operations to minimize API calls

## Out of Scope
- Bidirectional sync during migration
- Incremental/delta migration
- Migration scheduling/automation
- Cross-account migration (same backend type, different accounts)
