# [001] Todoist Backend Not Configurable via CLI/Config

## Category
todoist

## Severity
medium

## Steps to Reproduce
```bash
# Try to set Todoist as default backend
./bin/todoat -y config set default_backend todoist

# Try to store credentials
source .env && ./bin/todoat -y credentials set todoist "" --prompt <<< "$TODOAT_TODOIST_TOKEN"

# Check auto-detection with token set
source .env && ./bin/todoat -y --detect-backend
```

## Expected Behavior
- Todoist should be available as a configurable backend option
- Users should be able to set Todoist as default_backend
- With TODOAT_TODOIST_TOKEN env var set, Todoist should be auto-detected

## Actual Behavior
1. `config set default_backend todoist` returns: `Error: invalid value for default_backend: todoist (valid: sqlite)`
2. `credentials set` returns: `Error: failed to store credentials: system keyring not available in this build`
3. `--detect-backend` only shows sqlite, not Todoist, even with token in environment

## Error Output
```
$ ./bin/todoat -y config set default_backend todoist
Error: invalid value for default_backend: todoist (valid: sqlite)

$ source .env && ./bin/todoat -y credentials set todoist "" --prompt <<< "$TODOAT_TODOIST_TOKEN"
Enter password for todoist (user: ): Error: failed to store credentials: system keyring not available in this build

$ source .env && ./bin/todoat -y --detect-backend
Auto-detected backends:
  1. sqlite: /home/ubuntu/.local/share/todoat/tasks.db (always available)
     git: (not available) not available in current context

Would use: sqlite
```

## Environment
- OS: Linux
- Go version: go1.25.5 linux/amd64
- Config exists: yes
- DB exists: yes

## Possible Cause
The Todoist backend code exists in `backend/todoist/` but:
1. The config system (`internal/config/config.go`) only validates "sqlite" as a valid backend (line 142)
2. The `BackendsConfig` struct only has `SQLite` field, no `Todoist` field
3. System keyring integration requires specific build tags or system libraries not present

## Related Files
- internal/config/config.go:142 - Only allows "sqlite" as valid backend
- internal/config/config.go:41-43 - BackendsConfig only has SQLite
- backend/todoist/todoist.go - Todoist backend implementation exists

## Notes
The Todoist backend code exists and has tests, so it may be intended for future integration. The config system needs to be extended to support Todoist as a configurable backend option.

## Resolution

**Fixed in**: this session
**Fix description**: Extended the config system to support Todoist as a configurable backend:
1. Added `TodoistConfig` struct to `BackendsConfig` in `internal/config/config.go`
2. Added "todoist" to valid backends in the config validation and CLI setter functions
3. Added `backends.todoist.enabled` configuration key support
4. Implemented `DetectableBackend` interface on Todoist backend (`CanDetect`, `DetectionInfo`)
5. Registered Todoist as a detectable backend with priority 50 (between git=10 and sqlite=100)
6. Fixed nil pointer dereference in `Close()` method when backend has no HTTP client

**Test added**: `TestTodoistConfigSetDefaultBackendCLI`, `TestTodoistConfigSetBackendEnabledCLI`, `TestTodoistConfigGetAllShowsTodoistCLI`, `TestTodoistConfigValidationCLI`, `TestTodoistDetectBackendWithTokenCLI` in `backend/todoist/config_cli_test.go`

**Remaining issue**: The keyring credentials storage (`credentials set todoist`) is a separate issue related to system keyring availability, not addressed by this fix.
