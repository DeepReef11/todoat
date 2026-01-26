# [001] Start date after due date is not validated

## Type
code-bug

## Category
feature

## Severity
medium

## Steps to Reproduce
```bash
todoat "@work" add "Test" --start-date 2026-02-01 --due-date 2026-01-01
```

## Expected Behavior
The command should fail with an error message indicating that start date cannot be after due date.

## Actual Behavior
The task is created successfully with an illogical date range where start date (2026-02-01) is after due date (2026-01-01).

```json
{
  "summary": "Test",
  "due_date": "2026-01-01",
  "start_date": "2026-02-01"
}
```

## Error Output
```
Created task: Test (ID: a6eb7f2b-b696-410f-8331-5b07a98d1de7)
```
No error is shown - the invalid task is created.

## Environment
- OS: Linux
- Runtime version: Go 1.25.5

## Possible Cause
The `ValidateDateRange` function exists in `internal/utils/validation.go` (lines 186-198) and correctly validates that start date is not after due date. However, this function is never called from the CLI command handlers in `cmd/todoat/cmd/todoat.go`.

The function is defined but not integrated into the add/update task flows.

## Related Files
- `internal/utils/validation.go` - Contains `ValidateDateRange` function
- `cmd/todoat/cmd/todoat.go` - CLI handlers that should call `ValidateDateRange`

## Suggested Fix
In the `doAdd` and `doUpdate` functions (and related handlers), add a call to `utils.ValidateDateRange(startDate, dueDate)` before creating/updating the task, and return an appropriate error if validation fails.

## Resolution

**Fixed in**: this session
**Fix description**: Added calls to `utils.ValidateDateRange()` in `doAdd`, `doAddHierarchy`, `doUpdate`, and `doUpdateWithTask` functions in `cmd/todoat/cmd/todoat.go` to validate that start date is not after due date before creating or updating tasks.
**Test added**: `TestAddCommandStartDateAfterDueDateSQLiteCLI` in `backend/sqlite/cli_test.go`

### Verification Log
```bash
$ todoat "@work" add "Test" --start-date 2026-02-01 --due-date 2026-01-01
Error: start date cannot be after due date
Exit code: 1
```
**Matches expected behavior**: YES
