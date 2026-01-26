# [002] config set does not recognize sync.auto_sync_after_operation key

## Type
code-bug

## Category
feature

## Severity
medium

## Steps to Reproduce
```bash
todoat config set sync.auto_sync_after_operation true
```

## Expected Behavior
The command should set sync.auto_sync_after_operation to true, as documented in docs/reference/configuration.md

## Actual Behavior
Returns error:
```
Error: unknown config key: sync.auto_sync_after_operation
```

## Error Output
```
Error: unknown config key: sync.auto_sync_after_operation
```

## Environment
- OS: Linux
- Runtime version: Go (from source)

## Possible Cause
The `config set` command handler does not have `sync.auto_sync_after_operation` registered as a settable key, even though the config struct includes it and `config get` shows it.

## Documentation Reference (if doc-mismatch)
- File: `docs/reference/configuration.md`
- Section: Common Configuration Options table
- Documented key: `sync.auto_sync_after_operation`

## Related Files
- internal/config/config.go
- internal/cli/config.go

## Recommended Fix
FIX CODE - Add sync.auto_sync_after_operation to the list of settable config keys in the config set command handler

## Resolution

**Fixed in**: this session
**Fix description**: Added sync.auto_sync_after_operation to both setConfigValue and getConfigValue functions in cmd/todoat/cmd/todoat.go
**Test added**: TestConfigSetSyncAutoSyncAfterOperationCLI, TestConfigSetSyncAutoSyncAfterOperationValidationCLI in internal/config/cli_test.go

### Verification Log
```bash
$ todoat config set sync.auto_sync_after_operation true
Set sync.auto_sync_after_operation = true

$ todoat config get sync.auto_sync_after_operation
true
```
**Matches expected behavior**: YES
