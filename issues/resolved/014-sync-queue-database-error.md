# [014] sync queue command fails with database error when no remote backend configured

## Type
code-bug

## Category
feature

## Severity
low

## Steps to Reproduce
```bash
# Start with a fresh config (SQLite backend only, no remote configured)
todoat list create "Test"
todoat Test add "Sample task"
todoat sync queue
```

## Expected Behavior
Should either:
1. Display an empty queue message, or
2. Indicate that sync queue is not available without a configured remote backend

## Actual Behavior
Command fails with a database error:
```
Error: unable to open database file: out of memory (14)
```

## Error Output
```
Error: unable to open database file: out of memory (14)
```

## Environment
- OS: Linux
- Runtime version: Go 1.21+

## Possible Cause
The `sync queue` command tries to access a cache database in `~/.local/share/todoat/caches/` directory, but this directory doesn't exist when no remote backend is configured. The error message is misleading - it says "out of memory" but the actual issue is that the caches directory/database doesn't exist.

## Related Files
- sync queue command handler

## Recommended Fix
FIX CODE - Handle the case where no remote backend is configured by either:
1. Creating the caches directory if it doesn't exist
2. Displaying a helpful message like "No sync queue available - no remote backends configured"
3. Fixing the error message to accurately describe the issue

## Resolution

**Fixed in**: this session
**Fix description**: Added directory creation in `initDB()` function before opening SQLite database. The fix creates the parent directory of the database file if it doesn't exist, preventing the misleading "out of memory" error.
**Test added**: TestSyncQueueMissingDBDirectory in backend/sync/sync_test.go

### Verification Log
```bash
$ todoat list create "Test"
Created list: Test
$ todoat Test add "Sample task"
Created task: Sample task (ID: 3fad6c23-58ea-49d8-ab16-2b0c4b2a9a4f)
$ todoat sync queue
Pending Operations: 0
```
**Matches expected behavior**: YES
