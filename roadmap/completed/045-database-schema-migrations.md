# [045] Database Schema Migrations

## Summary
Implement automatic database schema migration system to handle version upgrades gracefully, tracking schema versions and applying migrations in order.

## Documentation Reference
- Primary: `docs/explanation/backend-system.md`
- Section: SQLite Backend - Database Schema
- Related: `docs/explanation/synchronization.md` (schema_version table)

## Dependencies
- Requires: [003] SQLite Backend

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestSchemaVersionTracking` - Database includes schema_version table with current version
- [ ] `TestMigrationOnUpgrade` - Opening older database triggers migration to current schema
- [ ] `TestMigrationIdempotent` - Running migrations multiple times is safe
- [ ] `TestMigrationOrder` - Migrations apply in version order

### Functional Requirements
- `schema_version` table tracks applied migrations
- Migrations run automatically on database open
- Each migration has version number and up/down functions
- Migrations are idempotent (safe to run multiple times)
- Failed migrations roll back and report error
- Backup created before destructive migrations (optional)

## Implementation Notes

### Schema Version Table
```sql
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Migration Structure
```go
type Migration struct {
    Version int
    Name    string
    Up      func(db *sql.DB) error
    Down    func(db *sql.DB) error  // Optional for rollback
}
```

### Migration Flow
1. Check current schema version
2. Find migrations with version > current
3. Apply migrations in order
4. Update schema_version after each success
5. On error: log, rollback current migration, return error

### Example Migrations
- v1: Initial schema (tasks, task_lists tables)
- v2: Add sync_metadata table
- v3: Add parent_uid column to tasks
- v4: Add categories column to tasks

## Out of Scope
- CLI command to manually run migrations
- Migration rollback command
- Data migrations (only schema changes)
