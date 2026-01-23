# [033] Git and File Backends Not Accessible via CLI

## Type
doc-mismatch

## Category
feature

## Severity
medium

## Steps to Reproduce
```bash
# Try to use git backend explicitly
todoat -b git list

# Try to use file backend explicitly
todoat -b file list
```

## Expected Behavior
The git and file backends should be usable via the `--backend` / `-b` flag, as documented in `docs/explanation/backends.md`.

## Actual Behavior
```
Error: unknown backend type 'git' for custom backend 'git'
```

The CLI only recognizes `sqlite`, `todoist`, and `nextcloud` as valid backend types:
- `--backend` flag help shows: `Backend to use (sqlite, todoist)`
- Error message shows: `unknown backend: %s (supported: sqlite, todoist, nextcloud)`

## Error Output
```
Error: unknown backend type 'git' for custom backend 'git'
```

## Environment
- OS: Linux
- Runtime version: Go (built from source)

## Possible Cause
The git and file backends are implemented in `backend/git/git.go` and `backend/file/file.go` but are not wired into the `createBackendByName` and `createCustomBackend` functions in `cmd/todoat/cmd/todoat.go`. The switch statements at lines 2288-2319 and 2340-2370 only handle `sqlite`, `todoist`, and `nextcloud` cases.

## Documentation Reference (if doc-mismatch)
- File: `docs/explanation/backends.md`
- Section: Available Backends, Git (Markdown)
- Documented command: The documentation describes Git and File backends as available options, including configuration examples like:
  ```yaml
  backends:
    git:
      type: git
      enabled: true
      auto_detect: true
  ```

## Related Files
- `backend/git/git.go` - Git backend implementation (exists, implements TaskManager interface)
- `backend/file/file.go` - File backend implementation (exists, implements TaskManager interface)
- `cmd/todoat/cmd/todoat.go` - CLI command handler (missing git/file cases in switch statements)
- `docs/explanation/backends.md` - Documents git/file backends as available

## Recommended Fix
FIX CODE - Add cases for `git` and `file` backend types in `createBackendByName` and `createCustomBackend` functions to wire up the existing implementations.

## Resolution

**Fixed in**: this session
**Fix description**: Added git and file backend support to CLI by:
1. Changed import from `_ "todoat/backend/git"` to `"todoat/backend/git"` and added `"todoat/backend/file"`
2. Added `case "git":` and `case "file":` in `createBackendByName` function
3. Added `case "git":` and `case "file":` in `createCustomBackend` function with config support
4. Updated error message to include git and file in supported backends list
**Test added**: TestIssue033GitBackendAccessibleViaCLI and TestIssue033FileBackendAccessibleViaCLI in cmd/todoat/cmd/todoat_test.go

### Verification Log
```bash
$ todoat -b git list
Available lists (1):

NAME                 TASKS
Inbox                1
$ todoat -b file list
Available lists (1):

NAME                 TASKS
Inbox                1
```
**Matches expected behavior**: YES
