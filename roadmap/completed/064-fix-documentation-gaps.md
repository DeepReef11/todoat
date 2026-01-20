# [064] Fix documentation to match implemented CLI commands

## Summary
Documentation has minor mismatches with actual CLI implementation.

## Documentation Reference
- Primary: `docs/list-management.md`, `docs/sync.md`

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

### 2. Sync conflict resolution flag mismatch
**Documented (incorrect):**
```bash
todoat sync conflicts resolve <conflict-id> --use local
todoat sync conflicts resolve <conflict-id> --use remote
```

**Actual command:**
```bash
todoat sync conflicts resolve [task-uid] --strategy local_wins
todoat sync conflicts resolve [task-uid] --strategy server_wins
# Also available: --strategy merge, --strategy keep_both
```

**Files:** `docs/sync.md` lines 65-66

## Previously Fixed (CLOSED)
These issues have been fixed in the documentation:
- ~~`list rename` vs `list update --name`~~ ✓ Fixed
- ~~`list show` vs `list info`~~ ✓ Fixed
- ~~`--filter "tags:X"` vs `--tag X`~~ ✓ Fixed
- ~~Missing export/import/vacuum/stats docs~~ ✓ Already documented in list-management.md

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] Run documented examples from fixed docs and verify they work

### Functional Requirements
- [ ] Remove `--force` example, replace with `-y` flag usage
- [ ] Fix sync conflicts resolve examples to use `--strategy` flag

## Recommended Fix
FIX DOCS - The CLI works correctly; just update documentation to match.
