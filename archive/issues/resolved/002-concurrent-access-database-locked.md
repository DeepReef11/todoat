# [002] Concurrent database access causes "database is locked" errors

## Type
code-bug

## Category
error-handling

## Severity
medium

## Steps to Reproduce
```bash
cd /tmp/todoat-test
for i in {1..5}; do
  todoat "@work" add "Concurrent task $i" &
done
wait
```

## Expected Behavior
All concurrent operations should either:
1. Succeed with proper locking/retry mechanism
2. Fail gracefully with a clear error message suggesting to retry

## Actual Behavior
Most concurrent operations fail with "database is locked" errors:

```
Created task: Concurrent task 3 (ID: e2fca0e3-3745-4b83-9c3c-ac0060a6f07a)
Error: database is locked (5) (SQLITE_BUSY)
Error: database is locked (5) (SQLITE_BUSY)
Error: database is locked (5) (SQLITE_BUSY)
Error: failed to create schema_version table: database is locked (5) (SQLITE_BUSY)
```

Only 1 out of 5 concurrent operations succeeds.

## Error Output
```
Error: database is locked (5) (SQLITE_BUSY)
Error: failed to create schema_version table: database is locked (5) (SQLITE_BUSY)
```

## Environment
- OS: Linux
- Runtime version: Go 1.25.5
- SQLite backend

## Possible Cause
The SQLite database connection is not configured with appropriate timeout/retry settings for concurrent access. SQLite's default behavior is to return SQLITE_BUSY immediately when the database is locked by another process.

## Suggested Fix
Consider one or more of the following:
1. Configure SQLite with `_busy_timeout` pragma to wait before returning SQLITE_BUSY
2. Implement retry logic with exponential backoff for database operations
3. Add WAL (Write-Ahead Logging) mode which allows concurrent reads and one write
4. Document the limitation that the CLI should not be run concurrently

## Related Files
- `backend/sqlite/sqlite.go` - SQLite backend implementation

## Resolution

**Fixed in**: this session
**Fix description**: Added SQLite pragmas for concurrent access in `initSchema()` in `backend/sqlite/sqlite.go`:
1. `PRAGMA busy_timeout = 5000` - wait up to 5 seconds when database is locked before returning SQLITE_BUSY
2. `PRAGMA journal_mode = WAL` - enable Write-Ahead Logging mode for better concurrent access
**Test added**: `TestConcurrentAddCommandsSQLiteCLI` in `backend/sqlite/cli_test.go`

### Verification Log
```bash
$ cd /tmp/todoat-test
$ for i in {1..5}; do todoat "@work" add "Concurrent task $i" & done; wait
Created task: Concurrent task 1 (ID: b6cb13fe-8049-48c5-b69c-dc8e46f48a11)
Created task: Concurrent task 2 (ID: d1d7181a-b6b5-4155-b065-1a9766fb99e1)
Created task: Concurrent task 4 (ID: 790f7262-218c-4e97-ac17-71cc63bd000d)
Created task: Concurrent task 3 (ID: bab04adf-e302-4fbe-96ed-0a3c126285d7)
Created task: Concurrent task 5 (ID: e0fd4376-5ee4-4dab-80b3-bd607eead973)
```
**All 5 concurrent operations succeeded. No "database is locked" errors.**
**Matches expected behavior**: YES
