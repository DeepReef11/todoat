# [003] No --backend flag exists despite documentation suggestions

## Category
other

## Severity
low

## Steps to Reproduce
```bash
./bin/todoat --backend=todoist MyList
./bin/todoat --backend=sqlite MyList
```

## Expected Behavior
Based on documentation and multi-backend support, a `--backend` flag might be expected to allow selecting which backend to use for a specific command.

## Actual Behavior
```
Error: unknown flag: --backend
Exit code: 1
```

## Error Output
```
Error: unknown flag: --backend
```

## Environment
- OS: Linux
- Go version: go1.25.5 linux/amd64
- Config exists: yes
- DB exists: yes

## Possible Cause
This is a design choice rather than a bug. The application uses the `default_backend` config setting to select the backend, not a CLI flag. However:

1. The config.go only supports "sqlite" as a valid backend value (line 133: `validBackends := map[string]bool{"sqlite": true}`)
2. The getBackend() function in todoat.go always returns SQLite backend
3. Todoist/Nextcloud backends are only available for migration operations, not as primary backends

This means there's currently no way to use Todoist as the primary backend from the CLI, only for migration target.

## Recommendation
Either:
1. Add documentation clarifying that Todoist is only supported as a migration target
2. Or implement a `--backend` flag to allow per-command backend selection

## Related Files
- cmd/todoat/cmd/todoat.go:669 (getBackend function - only returns SQLite)
- internal/config/config.go:133 (Validate function - only allows "sqlite")
- doc/backends.md (documents Todoist but doesn't clarify its limited scope)

## Resolution

**Fixed in**: this session
**Fix description**: Added "Backend Usage Overview" section to doc/backends.md that clearly documents:
1. SQLite is the only supported primary backend for day-to-day operations
2. Other backends (Todoist, Nextcloud, Google Tasks, MS To-Do, File, Git) are available as migration targets only
3. There is no `--backend` flag - backend selection is via config for primary operations or `--from`/`--to` flags for migrations
**Test added**: N/A (documentation change only)

## Regression Resolved

**Date**: 2026-01-20
**Previous regression**: Documentation fix was reported as missing
**Current status**: Documentation fix is now present and verified

### Verification Log
```bash
$ ./bin/todoat --backend=sqlite MyList
Error: unknown flag: --backend

$ grep -- '--backend' doc/backends.md
doc/backends.md:36:> **Note**: There is no `--backend` flag to select a backend per-command. Backend selection is determined by configuration for primary operations, or by the `--from` and `--to` flags for migration operations.
```

**Verification result**: The documentation now correctly states that there is no `--backend` flag. The "Backend Usage Overview" section at lines 5-36 of doc/backends.md clearly documents:
1. SQLite is the only supported primary backend
2. Other backends (Todoist, Nextcloud, Google Tasks, MS To-Do, File, Git) are migration targets only
3. There is no `--backend` flag - use config for primary operations or `--from`/`--to` for migrations

**Matches expected behavior**: YES
