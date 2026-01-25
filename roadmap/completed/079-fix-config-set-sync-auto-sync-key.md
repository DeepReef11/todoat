# [079] Fix: config set - sync.auto_sync_after_operation Not Supported

## Summary
The `sync.auto_sync_after_operation` config key is documented but not implemented in the `config set` command. Running `todoat config set sync.auto_sync_after_operation true` returns "unknown config key".

## Documentation Reference
- Primary: `docs/reference/configuration.md`
- Secondary: `docs/how-to/sync.md`
- Section: Sync Configuration Options

## Gap Type
wrong-behavior

## Documented Command/Syntax
```bash
todoat config set sync.auto_sync_after_operation true
```

## Actual Result When Running Documented Command
```bash
$ todoat config set sync.auto_sync_after_operation true
Error: unknown config key: sync.auto_sync_after_operation
```

## Working Alternative (if any)
```bash
todoat config edit  # Manually edit YAML file
```

## Recommended Fix
FIX CODE - Add sync.auto_sync_after_operation to the setConfigValue function in cmd/todoat/cmd/todoat.go

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] Test that `config set sync.auto_sync_after_operation true` sets the value correctly
- [ ] Test validation rejects non-boolean values

### Functional Requirements
- [ ] `todoat config set sync.auto_sync_after_operation true/false` works as documented
- [ ] Value is persisted to config file

## Implementation Notes
Add a new case in the `switch parts[1]` block under case "sync" at line 9364 in `cmd/todoat/cmd/todoat.go`:

```go
case "auto_sync_after_operation":
    boolVal, err := parseBool(value)
    if err != nil {
        return fmt.Errorf("invalid value for sync.auto_sync_after_operation: %s (valid: true, false)", value)
    }
    c.Sync.AutoSyncAfterOperation = boolVal
    return nil
```
