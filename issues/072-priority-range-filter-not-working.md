# [072] Priority Range Filter Not Working

## Type
doc-mismatch

## Category
feature

## Severity
low

## Steps to Reproduce
```bash
todoat Personal -p 1-3
```

## Expected Behavior
Should filter tasks with priority 1, 2, or 3 (as documented in cli.md line 80).

## Actual Behavior
Returns error:
```
Error: invalid priority value: 1-3
```

## Error Output
```
Error: invalid priority value: 1-3
```

## Environment
- OS: Linux
- Runtime version: Go 1.21+

## Possible Cause
The priority filter parser doesn't handle the range syntax (N-M) for the `-p` flag, although the view creation command `--filter-priority` does support it.

## Documentation Reference
- File: `docs/reference/cli.md`
- Section: Priority Filter Syntax (lines 77-82)
- Documented command: `-p 1-3` (Filter by priority range)

## Related Files
- `internal/cli/get.go` (likely)
- `internal/cli/flags.go` (likely)

## Workaround
Use comma-separated values instead:
```bash
todoat Personal -p 1,2,3
```

Or use named priority levels:
```bash
todoat Personal -p high
```

## Recommended Fix
FIX CODE - Implement priority range parsing for the `-p` flag to match the documented behavior and be consistent with `--filter-priority` in view creation.

## Dependencies
none
