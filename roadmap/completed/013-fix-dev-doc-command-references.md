# [013] Fix: dev-doc references non-existent commands

## Summary
Internal documentation files (dev-doc/README.md and dev-doc/CONFIGURATION.md) reference commands and flags that don't exist in the implementation. This causes confusion for developers.

## Documentation Reference
- Primary: `issues/012-dev-doc-readme-command-mismatches.md`
- Files affected:
  - `dev-doc/README.md` (lines 47, 49, 52)
  - `dev-doc/CONFIGURATION.md` (lines 262, 934)

## Gap Type
wrong-syntax

## Documented Command/Syntax
```bash
# dev-doc/README.md line 47
list rename

# dev-doc/README.md line 49, CONFIGURATION.md lines 262, 934
todoat --list-backends

# dev-doc/README.md line 52
todoat view show
```

## Actual Result When Running Documented Command
```bash
$ todoat list rename
# Shows help for `todoat list` - no rename subcommand exists

$ todoat --list-backends
Error: unknown flag: --list-backends

$ todoat view show
# Shows help for `todoat view` - no show subcommand exists
```

## Working Alternative (if any)
```bash
# Instead of list rename:
todoat list update "Old Name" --name "New Name"

# Instead of --list-backends:
todoat --detect-backend
# or
todoat config get backends

# Instead of view show:
todoat view list
```

## Recommended Fix
FIX DOCS - Update internal documentation to match actual CLI implementation.

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] N/A (documentation change only)

### Functional Requirements
- [ ] dev-doc/README.md line 47: Change `list rename` to `list update --name`
- [ ] dev-doc/README.md line 49: Remove `--list-backends`
- [ ] dev-doc/README.md line 52: Remove `view show`
- [ ] dev-doc/CONFIGURATION.md line 262: Replace `--list-backends` reference
- [ ] dev-doc/CONFIGURATION.md line 934: Replace `todoat --list-backends` example

## Implementation Notes
The user-facing documentation (docs/*.md) is correct and does not reference these non-existent commands. Only the internal dev-doc needs updating. This was partially addressed in commit 6f7e421 which fixed dev-doc/CLI_INTERFACE.md but missed README.md and CONFIGURATION.md.
