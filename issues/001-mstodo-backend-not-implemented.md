# [001] Microsoft To Do (mstodo) Backend Not Implemented

## Type
doc-mismatch

## Category
feature

## Severity
medium

## Steps to Reproduce
```bash
todoat -b mstodo list
```

## Expected Behavior
The command should either:
1. Work with Microsoft To Do backend (if implemented), or
2. Show an error about missing credentials similar to other backends

## Actual Behavior
Returns error indicating `mstodo` is an unknown backend:
```
Error: unknown backend: mstodo (supported: sqlite, todoist, nextcloud, google, git, file)
```

## Error Output
```
Error: unknown backend: mstodo (supported: sqlite, todoist, nextcloud, google, git, file)
```

## Environment
- OS: Linux
- Runtime version: Go (dev build)

## Possible Cause
The Microsoft To Do backend is documented but not yet implemented in the codebase. The error message shows the actual supported backends: sqlite, todoist, nextcloud, google, git, file - which does not include mstodo.

## Documentation Reference (if doc-mismatch)
- File: `docs/reference/cli.md`
- Section: Global Flags
- Documented command: `-b, --backend <name>` with value `mstodo`

- File: `docs/explanation/backends.md`
- Section: Available Backends
- Documents: Microsoft To Do (`mstodo`) with full configuration examples

## Related Files
- docs/reference/cli.md (line 18)
- docs/explanation/backends.md (lines 187-293)

## Recommended Fix
FIX DOCS - Remove Microsoft To Do documentation until the backend is implemented, OR implement the backend to match the documentation.
