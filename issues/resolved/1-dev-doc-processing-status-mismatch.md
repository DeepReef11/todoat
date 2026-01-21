# [001] Example Mismatch: dev-doc files use invalid PROCESSING status

## Type
doc-mismatch

## Category
user-journey

## Severity
medium

## Location
Multiple dev-doc files contain command examples using "PROCESSING" status which is invalid.

Files affected:
- `dev-doc/README_PLANNER.md` (lines 144, 166, 167, 246)
- `dev-doc/SUBTASKS_HIERARCHY.md` (line 582)
- `dev-doc/TASK_MANAGEMENT.md` (lines 602, 603, 632)
- `dev-doc/README.md` (lines 79, 89)
- `dev-doc/BACKEND_SYSTEM.md` (multiple references)
- `dev-doc/FEATURES_OVERVIEW.md` (lines 23, 32)
- `dev-doc/SYNCHRONIZATION.md` (line 1098)

## Documented Command
```bash
# From dev-doc/README_PLANNER.md:144
todoat Work update "Ship feature" -s PROCESSING

# From dev-doc/TASK_MANAGEMENT.md:602
todoat MyList -s TODO,PROCESSING
```

## Actual Result
```bash
$ todoat Work update "Ship feature" -s PROCESSING
Error: invalid status "PROCESSING": valid values are TODO, IN-PROGRESS, DONE, CANCELLED

$ todoat MyList -s TODO,PROCESSING
Error: invalid status "TODO,PROCESSING": valid values are TODO, IN-PROGRESS, DONE, CANCELLED
```

## Working Alternative (if known)
```bash
todoat Work update "Ship feature" -s IN-PROGRESS
todoat MyList -s TODO,IN-PROGRESS
```

## Recommended Fix
FIX EXAMPLE - Replace all occurrences of "PROCESSING" with "IN-PROGRESS" and "P" abbreviation with "I" in dev-doc files.

This was previously identified in `roadmap/completed/007-dev-doc-status-abbreviation-inconsistency.md` but only addressed `dev-doc/CLI_INTERFACE.md`. Other dev-doc files still contain the incorrect status.

## Impact
Developers following internal documentation examples will see errors. The user-facing docs (docs/) correctly use IN-PROGRESS.

## Resolution

**Fixed in**: this session
**Fix description**: Replaced all occurrences of "PROCESSING" with "IN-PROGRESS" and "P" abbreviation with "I" across all dev-doc files.

**Files modified**:
- `dev-doc/README_PLANNER.md` - Fixed status in examples and status table
- `dev-doc/SUBTASKS_HIERARCHY.md` - Fixed status in tree displays and examples
- `dev-doc/TASK_MANAGEMENT.md` - Fixed all status references, examples, and tables
- `dev-doc/README.md` - Fixed status table and data model
- `dev-doc/BACKEND_SYSTEM.md` - Fixed status mappings, examples, and code samples
- `dev-doc/FEATURES_OVERVIEW.md` - Fixed status descriptions
- `dev-doc/SYNCHRONIZATION.md` - Fixed schema comment

### Verification Log
```bash
$ grep -n "PROCESSING" dev-doc/*.md
No PROCESSING found

$ grep -n "IN-PROGRESS" dev-doc/README_PLANNER.md | head -3
144:todoat Work update "Ship feature" -s IN-PROGRESS
166:# Show TODO and IN-PROGRESS
167:todoat Work -s TODO,IN-PROGRESS
```
**Matches expected behavior**: YES
