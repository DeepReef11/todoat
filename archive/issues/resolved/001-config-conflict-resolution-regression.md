# [001] sync.conflict_resolution documented values still not working (regression)

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
Based on documentation (docs/reference/configuration.md, docs/how-to/sync.md, README.md), `server_wins` should be a valid value for sync.conflict_resolution. The documentation lists valid values as: `server_wins`, `local_wins`, `merge`, `keep_both`.

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
- Runtime version: Go 1.25.5

## Possible Cause
This appears to be a regression. The issue was previously documented in `issues/resolved/003-config-conflict-resolution-values-mismatch.md` which claims it was fixed, but the bug still exists. Either the fix was never committed or was reverted.

## Documentation Reference (if doc-mismatch)
- File: `docs/reference/configuration.md`
- Section: Common Configuration Options table (line ~134)
- Documented values: `server_wins`, `local_wins`, `merge`, `keep_both`
- Actual valid values: `local`, `remote`, `manual`

Also documented incorrectly in:
- `docs/how-to/sync.md` (lines 21, 73-79, 149-160, etc.)
- `README.md` (shows `conflict_resolution: remote` which is the actual valid value)

## Related Files
- docs/reference/configuration.md
- docs/how-to/sync.md
- README.md
- internal/cli/config.go (validation logic)
- issues/resolved/003-config-conflict-resolution-values-mismatch.md (previous "fix")

## Recommended Fix
FIX CODE - Update the validation in internal/cli/config.go to accept the documented values (`server_wins`, `local_wins`, `merge`, `keep_both`) as intended, since the resolved issue indicates this was the intended fix direction. The fix may not have been properly applied.

## Dependencies
None
