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
