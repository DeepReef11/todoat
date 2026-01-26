# [003] sync.conflict_resolution documented values don't match code

## Type
doc-mismatch

## Category
feature

## Severity
medium

## Steps to Reproduce
```bash
todoat config set sync.conflict_resolution server_wins
```

## Expected Behavior
Based on documentation, `server_wins` should be a valid value for sync.conflict_resolution

## Actual Behavior
Returns error indicating different valid values:
```
Error: invalid value for sync.conflict_resolution: server_wins (valid: local, remote, manual)
```

## Error Output
```
Error: invalid value for sync.conflict_resolution: server_wins (valid: local, remote, manual)
```

## Environment
- OS: Linux
- Runtime version: Go (from source)

## Possible Cause
Documentation was written with one set of values (`server_wins`, `local_wins`, `merge`, `keep_both`) but code implements different values (`local`, `remote`, `manual`).

## Documentation Reference (if doc-mismatch)
- File: `docs/reference/configuration.md`
- Section: Common Configuration Options table
- Documented values: `server_wins`, `local_wins`, `merge`, `keep_both`
- Actual valid values: `local`, `remote`, `manual`

## Related Files
- docs/reference/configuration.md
- docs/how-to/sync.md
- README.md (mentions `server_wins`)
- internal/cli/config.go

## Recommended Fix
FIX DOCS - Update documentation to match actual valid values (`local`, `remote`, `manual`), or FIX CODE if the documented values are the intended ones

## Resolution

**Fixed in**: this session
**Fix description**: Changed code validation to accept the documented values (`server_wins`, `local_wins`, `merge`, `keep_both`) instead of the incorrect values (`local`, `remote`, `manual`). Also updated `docs/reference/configuration.md` to list the correct values.

### Verification Log
```bash
$ todoat config set sync.conflict_resolution server_wins
Set sync.conflict_resolution = server_wins
```
**Matches expected behavior**: YES
