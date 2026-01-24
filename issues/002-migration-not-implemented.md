# [002] Migration command not implemented for real backends

## Type
missing-feature

## Category
feature

## Severity
medium

## Steps to Reproduce
```bash
# Try to migrate from sqlite to any real backend
./todoat migrate --from sqlite --to file --list TestManualJourney --dry-run
# Error: real file backend not yet implemented for migration

./todoat migrate --from sqlite --to todoist --list TestManualJourney --dry-run
# Error: real todoist backend not yet implemented for migration

./todoat migrate --from sqlite --to nextcloud --list TestManualJourney --dry-run
# Error: real nextcloud backend not yet implemented for migration
```

## Expected Behavior
The `todoat migrate` command should be able to migrate tasks between different backends as documented in the `--help` output and the documentation (which mentions preserving metadata and hierarchy).

## Actual Behavior
All migrations to real backends fail with "real [backend] backend not yet implemented for migration" error.

## Error Output
```
Error: real file backend not yet implemented for migration
Error: real todoist backend not yet implemented for migration
Error: real nextcloud backend not yet implemented for migration
```

## Environment
- OS: Linux
- Runtime version: Go (dev build)

## Possible Cause
Migration functionality is only scaffolded but not yet implemented for actual backend-to-backend transfers.

## Documentation Reference (if doc-mismatch)
- File: `docs/how-to/sync.md` and migrate --help
- The migrate command is documented but the implementation is incomplete

## Related Files
- migrate command implementation
- Backend interfaces

## Recommended Fix
IMPLEMENT FEATURE or FIX DOCS - Either implement the migration functionality or clearly document in --help and docs that this feature is not yet available.
