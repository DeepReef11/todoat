# [001] default_backend Configuration Setting Is Ignored

## Category
startup | todoist

## Severity
high

## Steps to Reproduce
```bash
# Set todoist as default backend
./bin/todoat -y config set default_backend todoist
./bin/todoat -y config set backends.todoist.enabled true

# Verify config
./bin/todoat -y config get default_backend
# Output: todoist

# List tasks - still uses SQLite
./bin/todoat -y list
# Output: Shows SQLite lists, not Todoist projects
```

## Expected Behavior
When `default_backend` is set to `todoist`, the CLI should use the Todoist backend for all operations.

## Actual Behavior
The CLI always uses SQLite backend regardless of the `default_backend` configuration setting. The `getBackend()` function in `cmd/todoat/cmd/todoat.go:2090` always returns a SQLite backend without checking the `default_backend` config value.

## Error Output
```
No error is shown - the CLI silently uses SQLite instead of the configured backend.
```

## Environment
- OS: Linux 6.12.65-1-lts
- Go version: go1.25.5 linux/amd64
- Config exists: yes
- DB exists: yes

## Possible Cause
The `getBackend()` function at line 2090 in `cmd/todoat/cmd/todoat.go` does not read the `default_backend` setting from the configuration. It:
1. Always defaults to SQLite at line 2138
2. Only considers sync mode or auto-detect mode
3. Never checks the `DefaultBackend` field from config

The `DefaultBackend` setting is read and stored in config (lines 7257-7258, 7399) but never used to select the actual backend.

## Related Files
- cmd/todoat/cmd/todoat.go:2090-2139 (getBackend function)
- internal/config/config.go:17 (DefaultBackend field)

## Resolution

**Fixed in**: this session
**Fix description**: Modified `getBackend()` function in `cmd/todoat/cmd/todoat.go` to check `appConfig.DefaultBackend` after auto-detect and before falling back to SQLite. When `default_backend` is set to `todoist`, the function now creates a Todoist backend using `todoist.ConfigFromEnv()`. If the token is missing, a clear error message is returned instead of silently using SQLite.
**Test added**: `TestDefaultBackendTodoistUsedCLI` and `TestDefaultBackendTodoistWithTokenCLI` in `backend/todoist/config_cli_test.go`
