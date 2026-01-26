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

## Resolution

**Fixed in**: this session
**Fix description**: Documentation updated to clearly indicate that migration to real backends (nextcloud, todoist, file) is not yet implemented. The feature is scaffolded but currently only works with mock backends for testing.
**Test added**: N/A (documentation-only fix)

**Files updated**:
1. `docs/explanation/backends.md` - Added note about unimplemented status, status table showing all combinations as "Not implemented"
2. `docs/reference/cli.md` - Added note and "Current Limitations" section explaining the error users will see
3. `docs/feature-demo.sh` - Updated to show migration is not yet implemented

### Verification Log
```bash
$ grep -A 5 "Migration to real backends" docs/explanation/backends.md
> **Note**: Migration to real backends (nextcloud, todoist, file) is not yet implemented.
> Currently, migration only works for testing purposes using mock backends (`nextcloud-mock`, `todoist-mock`, `file-mock`).
> For now, to move tasks between backends, export from the source and import to the target manually.

The `migrate` command is designed to move tasks from one backend to another while preserving metadata.

$ grep -A 3 "Migration to real backends" docs/reference/cli.md
> **Note**: Migration to real backends (nextcloud, todoist, file) is not yet implemented.
> The command is scaffolded for future use. See [Backends - Migrating Between Backends](../explanation/backends.md#migrating-between-backends) for current status.

$ go test ./...
ok  	todoat/backend	(cached)
ok  	todoat/backend/file	(cached)
...
ok  	todoat/internal/views	(cached)
```

**Matches expected behavior**: YES (documentation now accurately reflects implementation status)
