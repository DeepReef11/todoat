# [001] Status filter with comma-separated values does not work

## Type
doc-mismatch

## Category
feature

## Severity
medium

## Steps to Reproduce
```bash
# Setup a list with tasks
todoat list create "Test"
todoat Test add "Task 1"
todoat Test add "Task 2"
todoat Test update "Task 2" -s IN-PROGRESS

# Try to filter by multiple statuses (as documented)
todoat Test -s TODO,IN-PROGRESS
```

## Expected Behavior
According to the documentation in `docs/getting-started.md` and `docs/task-management.md`, filtering by multiple statuses using comma-separated values should work:

```bash
# Show only incomplete tasks
todoat MyList -s TODO,IN-PROGRESS

# Using abbreviations (T=TODO, D=DONE, I=IN-PROGRESS, C=CANCELLED)
todoat MyList -s T,I
```

The command should return tasks matching either TODO or IN-PROGRESS status.

## Actual Behavior
The command fails with an error:

```
Error: invalid status "TODO,IN-PROGRESS": valid values are TODO, IN-PROGRESS, DONE, CANCELLED
```

## Error Output
```
Error: invalid status "TODO,IN-PROGRESS": valid values are TODO, IN-PROGRESS, DONE, CANCELLED
```

## Environment
- OS: Linux
- Runtime version: Go (built from source)

## Possible Cause
The status flag parser does not handle comma-separated values. It treats the entire string "TODO,IN-PROGRESS" as a single status value instead of splitting it.

Note: The priority filter (`-p`) correctly handles comma-separated values (e.g., `-p 1,2,3` works), suggesting this is a specific implementation gap in the status filter.

## Documentation Reference (if doc-mismatch)
- File: `docs/getting-started.md`
- Section: Status Values
- Documented command: `todoat MyList -s TODO,IN-PROGRESS`

- File: `docs/task-management.md`
- Section: Filtering by Status
- Documented commands:
  - `todoat MyList -s TODO,IN-PROGRESS`
  - `todoat MyList -s T,I`

## Related Files
- Likely in cmd/todoat parsing code for the `-s/--status` flag

## Recommended Fix
FIX CODE - The status flag handler should be updated to split comma-separated values and accept multiple statuses, similar to how the priority filter works.

## Resolution

**Fixed in**: this session
**Fix description**: Added `parseStatusFilter` and `matchesStatusFilter` functions to handle comma-separated status values. Also added single-letter abbreviations "I" for IN-PROGRESS and "C" for CANCELLED to `parseStatusWithValidation`.
**Test added**: TestStatusFilterCommaSeparatedSQLiteCLI and TestStatusFilterAbbreviationCommaSeparatedSQLiteCLI in backend/sqlite/cli_test.go

### Verification Log
```bash
$ todoat -y list create "Test" && todoat -y Test add "Task 1" && todoat -y Test add "Task 2" && todoat -y Test update "Task 2" -s IN-PROGRESS && todoat -y Test -s TODO,IN-PROGRESS
Created list: Test
ACTION_COMPLETED
Created task: Task 1 (ID: f90f76c0-66fd-4336-bad0-7593d575e0b8)
ACTION_COMPLETED
Created task: Task 2 (ID: 8edc512a-84b8-46da-909b-896b4fe217ea)
ACTION_COMPLETED
Updated task: Task 2
ACTION_COMPLETED
Tasks in 'Test':
  [TODO] Task 1
  [IN-PROGRESS] Task 2
INFO_ONLY
```
**Matches expected behavior**: YES
