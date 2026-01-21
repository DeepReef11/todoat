# [026] Cross-Backend Migration

## Summary
Implement functionality to migrate tasks between different backends, enabling users to move tasks from one storage provider to another while preserving all metadata.

## Documentation Reference
- Primary: `docs/explanation/features-overview.md` (Planned features)
- Related: `docs/explanation/backend-system.md`

## Dependencies
- Requires: [003] SQLite Backend
- Requires: [016] Nextcloud Backend
- Requires: [021] Todoist Backend

## Complexity
L

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestMigrateSQLiteToNextcloud` - `todoat migrate --from sqlite --to nextcloud` moves tasks
- [ ] `TestMigrateNextcloudToTodoist` - `todoat migrate --from nextcloud --to todoist` moves tasks
- [ ] `TestMigrateListSelection` - `todoat migrate --from sqlite --to nextcloud --list MyList` migrates specific list
- [ ] `TestMigratePreservesMetadata` - Migrated tasks retain priority, dates, status, tags
- [ ] `TestMigratePreservesHierarchy` - Parent-child relationships preserved during migration
- [ ] `TestMigrateDryRun` - `todoat migrate --dry-run` shows what would be migrated without changes
- [ ] `TestMigrateStatusMapping` - Statuses correctly mapped between backends
- [ ] `TestMigrateConflictHandling` - Handles existing tasks with same UID gracefully
- [ ] `TestMigrateBatchSize` - Large lists migrated in batches with progress

## Implementation Notes
- Create `migrate` command in `cmd/migrate.go`
- Implement backend adapter layer for format translation
- Handle status differences between backends (e.g., PROCESSING not supported in Todoist)
- Preserve UIDs where possible, regenerate if backend requires different format
- Support incremental migration (new tasks only)
- Transaction support: rollback on partial failure

## Out of Scope
- Real-time synchronization (use sync feature)
- Bidirectional migration in single command
- Automatic conflict resolution
- Migration scheduling
