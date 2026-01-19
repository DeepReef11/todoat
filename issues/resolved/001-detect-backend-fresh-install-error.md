# [001] --detect-backend shows misleading error on fresh install

## Category
startup

## Severity
low

## Steps to Reproduce
```bash
# Remove existing config and data directories
rm -rf ~/.config/todoat ~/.local/share/todoat

# Run detect-backend on fresh install
./bin/todoat --detect-backend
```

## Expected Behavior
Either:
1. Auto-create the necessary directories and show SQLite as available, OR
2. Show a clear message that directories don't exist yet and need to be created on first use

## Actual Behavior
Shows a confusing error message:
```
Auto-detected backends:
     git: (not available) not available in current context
     sqlite: (not available) failed to initialize: unable to open database file: out of memory (14)

No backends available. Configure a backend in config.yaml.
```

The error "out of memory (14)" is misleading - the actual issue is that the parent directories don't exist.

## Error Output
```
Auto-detected backends:
     git: (not available) not available in current context
     sqlite: (not available) failed to initialize: unable to open database file: out of memory (14)

No backends available. Configure a backend in config.yaml.
```

## Environment
- OS: Linux 6.12.65-1-lts
- Go version: go1.25.5 linux/amd64
- Config exists: no
- DB exists: no

## Possible Cause
The SQLite backend detection tries to open the database file without first ensuring the parent directories exist. The SQLite error code 14 (SQLITE_CANTOPEN) is being displayed as "out of memory" which is incorrect/misleading.

## Related Files
- cmd/todoat/detect.go (likely handles --detect-backend)
- internal/backend/sqlite/ (SQLite backend initialization)

## Notes
- Normal list operations (`./bin/todoat -y MyList`) correctly auto-create the directories
- Only `--detect-backend` has this issue
- The error message is confusing for new users trying to understand available backends

## Resolution

**Fixed in**: this session
**Fix description**: Modified `sqlite.NewDetectable()` in `backend/sqlite/sqlite.go` to ensure the parent directory exists (using `os.MkdirAll`) before attempting to open the SQLite database. This prevents the confusing "out of memory (14)" error on fresh installs and makes SQLite always available as intended.
**Test added**: `TestDetectBackendFreshInstallCLI` in `backend/detect_cli_test.go`
