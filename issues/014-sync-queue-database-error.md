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
