# [078] Fix: config set - analytics.enabled Not Supported

## Summary
The `analytics.enabled` and `analytics.retention_days` config keys are documented but not implemented in the `config set` command. Running `todoat config set analytics.enabled true` returns "unknown config key: analytics.enabled".

## Documentation Reference
- Primary: `docs/reference/configuration.md`
- Section: Analytics Configuration

## Gap Type
wrong-behavior

## Documented Command/Syntax
```bash
todoat config set analytics.enabled true
todoat config set analytics.retention_days 365
```

## Actual Result When Running Documented Command
```bash
$ todoat config set analytics.enabled true
Error: unknown config key: analytics.enabled
```

## Working Alternative (if any)
```bash
todoat config edit  # Manually edit YAML file
```

## Recommended Fix
FIX CODE - Add analytics.enabled and analytics.retention_days to the setConfigValue function in cmd/todoat/cmd/todoat.go

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] Test that `config set analytics.enabled true` sets the value correctly
- [ ] Test that `config set analytics.retention_days 365` sets the value correctly
- [ ] Test validation rejects invalid values (non-boolean for enabled, negative for retention_days)

### Functional Requirements
- [ ] `todoat config set analytics.enabled true` works as documented
- [ ] `todoat config set analytics.retention_days N` works as documented
- [ ] Values are persisted to config file

## Implementation Notes
The `setConfigValue` function at line 9295 in `cmd/todoat/cmd/todoat.go` needs a new case for "analytics" similar to the existing cases for "sync" and "trash":

```go
case "analytics":
    if len(parts) < 2 {
        return fmt.Errorf("invalid key: %s (use analytics.<setting>)", key)
    }
    switch parts[1] {
    case "enabled":
        boolVal, err := parseBool(value)
        if err != nil {
            return fmt.Errorf("invalid value for analytics.enabled: %s (valid: true, false)", value)
        }
        c.Analytics.Enabled = boolVal
        return nil
    case "retention_days":
        days, err := strconv.Atoi(value)
        if err != nil || days < 0 {
            return fmt.Errorf("invalid value for analytics.retention_days: %s (must be non-negative integer)", value)
        }
        c.Analytics.RetentionDays = days
        return nil
    }
```
