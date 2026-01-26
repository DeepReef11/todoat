# [068] SQLite custom path configuration is ignored

## Type
code-bug

## Category
feature

## Severity
medium

## Steps to Reproduce
```bash
# Create config with custom database path
cat > ~/.config/todoat/config.yaml << 'EOF'
backends:
  sqlite:
    type: sqlite
    enabled: true
    path: "/tmp/custom-tasks.db"
default_backend: sqlite
EOF

# Create a list (which should trigger database creation)
todoat list create "TestList"

# Check if database was created at custom path
ls -la /tmp/custom-tasks.db
# Result: file does not exist

# Check default location
ls -la ~/.local/share/todoat/tasks.db
# Result: file exists - database was created at default location instead
```

## Expected Behavior
When `backends.sqlite.path` is configured, the SQLite database should be created at that custom path.

## Actual Behavior
The custom path is ignored and the database is always created at the default location (`~/.local/share/todoat/tasks.db`).

The config value is correctly loaded (verified via `todoat config get backends.sqlite.path`) but not used when initializing the SQLite backend.

## Error Output
No error - the feature silently fails to work.

## Environment
- OS: Linux
- Runtime version: Go

## Possible Cause
The config path is loaded into `Config.Backends.SQLite.Path` and can be retrieved via `Config.GetDatabasePath()`, but the SQLite backend initialization code does not use this value. The backend is likely being initialized with the default path instead of the configured path.

## Related Files
- `internal/config/config.go:192-195` - `GetDatabasePath()` method exists but may not be called
- `backend/sqlite/sqlite.go:89` - `New(path string)` accepts path parameter
- Backend initialization code that creates the SQLite backend

## Recommended Fix
FIX CODE - Ensure the SQLite backend initialization passes the configured path from `Config.GetDatabasePath()` to `sqlite.New()`. If the path is empty, use the default location.

## Resolution

**Fixed in**: this session
**Fix description**: Modified `getBackend()` in `cmd/todoat/cmd/todoat.go` to check `appConfig.GetDatabasePath()` when determining the database path. The priority order is now: CLI flag > config file > default.
**Test added**: `TestIssue068SQLitePathConfigUsed` in `backend/sqlite/cli_test.go`

### Verification Log
```bash
$ go test -v ./backend/sqlite/... -run "TestIssue068"
=== RUN   TestIssue068SQLitePathConfigUsed
--- PASS: TestIssue068SQLitePathConfigUsed (0.02s)
PASS
ok  	todoat/backend/sqlite	0.026s
```
**Matches expected behavior**: YES
