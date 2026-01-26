# [001] config set does not recognize analytics.enabled key

## Type
code-bug

## Category
feature

## Severity
medium

## Steps to Reproduce
```bash
todoat config set analytics.enabled true
```

## Expected Behavior
The command should set analytics.enabled to true, as documented in docs/reference/configuration.md

## Actual Behavior
Returns error:
```
Error: unknown config key: analytics.enabled
```

## Error Output
```
Error: unknown config key: analytics.enabled
```

## Environment
- OS: Linux
- Runtime version: Go (from source)

## Possible Cause
The `config set` command handler does not have `analytics.enabled` registered as a settable key, even though the config struct includes it and `config get` shows it.

## Documentation Reference (if doc-mismatch)
- File: `docs/reference/configuration.md`
- Section: Common Configuration Options table and Analytics Configuration section
- Documented command: `todoat config set analytics.enabled true` (implied)

## Related Files
- internal/config/config.go
- internal/cli/config.go

## Recommended Fix
FIX CODE - Add analytics.enabled to the list of settable config keys in the config set command handler

## Resolution

**Fixed in**: 6f0bc5d (fix: support analytics config keys and update docs)
**Fix description**: Added analytics.enabled and analytics.retention_days to the list of settable config keys in the config set command handler
**Test added**: TestConfigSetAnalyticsEnabledCLI, TestConfigSetAnalyticsRetentionDaysCLI, TestConfigSetAnalyticsValidationCLI in internal/config/cli_test.go

### Verification Log
```bash
$ todoat config set analytics.enabled true
Set analytics.enabled = true

$ go test ./internal/config/... -run TestConfigSetAnalytics -v
=== RUN   TestConfigSetAnalyticsEnabledCLI
--- PASS: TestConfigSetAnalyticsEnabledCLI (0.00s)
=== RUN   TestConfigSetAnalyticsRetentionDaysCLI
--- PASS: TestConfigSetAnalyticsRetentionDaysCLI (0.00s)
=== RUN   TestConfigSetAnalyticsValidationCLI
--- PASS: TestConfigSetAnalyticsValidationCLI (0.00s)
PASS
ok  	todoat/internal/config	0.008s
```
**Matches expected behavior**: YES
