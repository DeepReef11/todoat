# [001] Config file not created on first run

## Category
startup

## Severity
low

## Steps to Reproduce
```bash
# Remove existing config and database
rm -rf ~/.config/todoat ~/.local/share/todoat

# Run any command that creates the database
./bin/todoat MyList

# Check for config file
ls ~/.config/todoat/config.yaml
```

## Expected Behavior
The config file should be created at `~/.config/todoat/config.yaml` (or `$XDG_CONFIG_HOME/todoat/config.yaml`) with default values when the application first runs.

The `config.Load()` function in `internal/config/config.go:57-72` is designed to create a default config if none exists, but this function is not called during normal CLI execution.

## Actual Behavior
The database is created at `~/.local/share/todoat/tasks.db`, but no config file is created at `~/.config/todoat/config.yaml`.

## Error Output
```
$ ls ~/.config/todoat/config.yaml
ls: cannot access '/home/ubuntu/.config/todoat/config.yaml': No such file or directory
```

## Environment
- OS: Linux
- Go version: go1.25.5 linux/amd64
- Config exists: no
- DB exists: yes

## Possible Cause
The CLI's `Execute()` function and `NewTodoAt()` in `cmd/todoat/cmd/todoat.go` don't call `config.Load()`. Instead, `getBackend()` directly creates a SQLite backend using default paths.

The `config.Load()` function exists and is tested (`TestAppStartsWithoutExistingConfigSQLiteCLI`), but it's only called in tests, not in the actual CLI code path.

## Related Files
- cmd/todoat/cmd/todoat.go:64-99 (Execute function)
- cmd/todoat/cmd/todoat.go:100-210 (NewTodoAt function)
- cmd/todoat/cmd/todoat.go:669-686 (getBackend function)
- internal/config/config.go:57-72 (Load function)

## Impact
- Users cannot configure the application via config file
- Default view settings (`default_view`) cannot be persisted
- Sync settings cannot be configured
- The `config.yaml` is created only if users manually call sync status or other features that trigger config loading

## Resolution

**Fixed in**: this session
**Fix description**: Added a call to `config.Load(cfg.ConfigPath)` in `getBackend()` function at `cmd/todoat/cmd/todoat.go:682-683`. This ensures the config file is created with defaults when the CLI is first executed.
**Test added**: `TestConfigCreatedOnCLIExecutionSQLiteCLI` in `cmd/todoat/cmd/todoat_test.go` (already existed)
