# [002] Backend Flag Help Text Shows Only sqlite and todoist

## Type
doc-mismatch

## Category
user-journey

## Severity
low

## Steps to Reproduce
```bash
todoat --help | grep backend
```

## Expected Behavior
The help text should list all supported backends:
```
-b, --backend string   Backend to use (sqlite, todoist, nextcloud, google, git, file)
```

## Actual Behavior
The help text only shows two backends:
```
-b, --backend string   Backend to use (sqlite, todoist)
```

## Error Output
```
$ todoat --help | grep backend
  -b, --backend string          Backend to use (sqlite, todoist)
      --detect-backend          Show auto-detected backends and exit
```

## Environment
- OS: Linux
- Runtime version: Go (dev build)

## Possible Cause
The help text is hardcoded and not updated when new backends were added. The error message when using an invalid backend correctly lists all supported backends:
```
Error: unknown backend: mstodo (supported: sqlite, todoist, nextcloud, google, git, file)
```

This inconsistency suggests the help text string needs to be updated to match the actual backend list.

## Documentation Reference (if doc-mismatch)
- File: `docs/reference/cli.md`
- Section: Global Flags
- Documented: All backends listed correctly in docs
- CLI help: Only shows sqlite, todoist

## Related Files
- Likely in cmd/root.go or similar where the --backend flag is defined

## Recommended Fix
FIX CODE - Update the help text string for the --backend flag to include all supported backends (sqlite, todoist, nextcloud, google, git, file).

## Resolution

**Fixed in**: this session
**Fix description**: Updated the help text string for the --backend flag in cmd/todoat/cmd/todoat.go line 264 to include all supported backends (sqlite, todoist, nextcloud, google, mstodo, git, file).

### Verification Log
```bash
$ ./todoat --help | grep backend
todoat is a command-line task manager supporting multiple backends.
  credentials  Manage backend credentials
  migrate      Migrate tasks between backends
  sync         Synchronize with remote backends
  -b, --backend string          Backend to use (sqlite, todoist, nextcloud, google, mstodo, git, file)
      --detect-backend          Show auto-detected backends and exit
```
**Matches expected behavior**: YES (all backends now shown in help text)
