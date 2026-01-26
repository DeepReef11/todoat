# [001] Add command accepts invalid status values silently

## Category
error-handling

## Severity
low

## Steps to Reproduce
```bash
# Add a task with invalid status
./bin/todoat -y MyList add -s INVALID "Test task"

# Check what status the task was given
./bin/todoat -y MyList
```

## Expected Behavior
The `add` command should reject invalid status values and display an error message similar to the `update` command:
```
Error: invalid status "INVALID": valid values are TODO, IN-PROGRESS, DONE, CANCELLED
```

## Actual Behavior
The `add` command silently accepts the invalid status value and defaults to `TODO`:
```
Created task: Test task (ID: ...)
ACTION_COMPLETED
```

The task is created with status `TODO` instead of showing an error.

## Error Output
No error is shown - the command succeeds with exit code 0.

## Environment
- OS: Linux
- Go version: go1.25.5 linux/amd64
- Config exists: yes
- DB exists: yes

## Possible Cause
The `add` action in the CLI does not validate the status flag before creating the task, unlike the `update` action which properly validates it.

## Related Files
- cmd/todoat/main.go (or wherever the add action is handled)
- The update action already has proper validation that could be reused

## Resolution

**Fixed in**: this session
**Fix description**: Added status flag reading and validation in the `add` action using `parseStatusWithValidation()` (the same function used by `update`). Updated `doAdd()` and `doAddHierarchy()` functions to accept and use the validated status parameter instead of hardcoding `StatusNeedsAction`.
**Test added**: `TestAddCommandInvalidStatusSQLiteCLI` and `TestAddCommandValidStatusSQLiteCLI` in `backend/sqlite/cli_test.go`

### Verification Log
```bash
$ ./bin/todoat -y MyList add -s INVALID "Test task"
Error: invalid status "INVALID": valid values are TODO, IN-PROGRESS, DONE, CANCELLED
Exit code: 1

$ ./bin/todoat -y MyList add -s TODO "Test task"
Created task: Test task (ID: f98cf160-33c4-4976-8b61-a2bea96ce89c)
ACTION_COMPLETED
Exit code: 0

$ ./bin/todoat -y MyList add -s IN-PROGRESS "Test task in progress"
Created task: Test task in progress (ID: 0712945f-ed4b-4dd4-860b-c544b53763af)
ACTION_COMPLETED
Exit code: 0
```
**Matches expected behavior**: YES
