# [034] List Stats and Vacuum Commands Require Explicit Backend Flag

## Type
code-bug

## Category
feature

## Severity
low

## Steps to Reproduce
```bash
# Verify default backend is sqlite
todoat config get default_backend
# Output: sqlite

# Try to use stats command without explicit backend
todoat list stats

# Try to use vacuum command without explicit backend
todoat list vacuum
```

## Expected Behavior
The `list stats` and `list vacuum` commands should work when the default backend is already `sqlite`, without requiring the explicit `-b sqlite` flag.

## Actual Behavior
```
Error: stats command is only supported for SQLite backend
Error: vacuum command is only supported for SQLite backend
```

However, when using the explicit flag, both commands work:
```bash
todoat -b sqlite list stats   # Works
todoat -b sqlite list vacuum  # Works
```

## Error Output
```
Error: stats command is only supported for SQLite backend
```

## Environment
- OS: Linux
- Runtime version: Go (built from source)

## Possible Cause
The backend resolution logic for these commands may be checking the backend type incorrectly, or using a different code path than normal task operations that bypasses the default backend selection.

## Related Files
- `cmd/todoat/cmd/todoat.go` - CLI command handler

## Recommended Fix
FIX CODE - Ensure the stats and vacuum commands use the same backend resolution logic as other commands, so they correctly detect and use the default SQLite backend.
