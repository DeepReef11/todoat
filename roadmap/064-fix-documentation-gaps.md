# [064] Fix documentation to match implemented CLI commands

## Summary
Documentation has minor mismatches with actual CLI implementation.

## Documentation Reference
- Primary: `docs/list-management.md`

## Gap Type
wrong-syntax (documentation mismatch)

## Gaps to Fix

### 1. `--force` flag doesn't exist
**Documented (incorrect):**
```bash
todoat list delete "List Name" --force
```

**Actual command:**
```bash
todoat list delete "List Name"  # with -y / --no-prompt global flag
todoat -y list delete "List Name"
```

**Files:** `docs/list-management.md` line 88

### 2. Missing documentation for existing commands
The following commands exist but are not documented in `docs/list-management.md`:
- `list export` - Export lists to json/csv/ical/sqlite
- `list import` - Import lists from files
- `list vacuum` - Compact SQLite database
- `list stats` - Show database statistics

## Previously Fixed (CLOSED)
These issues have been fixed in the documentation:
- ~~`list rename` vs `list update --name`~~ ✓ Fixed
- ~~`list show` vs `list info`~~ ✓ Fixed
- ~~`--filter "tags:X"` vs `--tag X`~~ ✓ Fixed

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] Run documented examples from fixed docs and verify they work

### Functional Requirements
- [ ] Remove `--force` example, replace with `-y` flag usage
- [ ] Add documentation for `list export`, `list import`, `list vacuum`, `list stats`

## Recommended Fix
FIX DOCS - The CLI works correctly; just update documentation to match.
