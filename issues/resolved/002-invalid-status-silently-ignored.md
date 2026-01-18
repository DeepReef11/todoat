# [002] Invalid status value silently ignored on add

## Category
error-handling

## Severity
low

## Steps to Reproduce
```bash
./bin/todoat MyList add "Status test" -s INVALID
```

## Expected Behavior
The CLI should return an error message indicating that "INVALID" is not a valid status value. According to the help text, valid values are: TODO, IN-PROGRESS, DONE, CANCELLED.

## Actual Behavior
The task is created successfully with status defaulting to "TODO":

```
Created task: Status test (ID: 6c394a12-5fbf-4314-b104-f6de0cd7f562)
Exit code: 0
```

When viewing the task:
```json
{"summary":"Status test","status":"TODO","priority":0}
```

## Error Output
No error produced - command succeeds with invalid input.

## Environment
- OS: Linux
- Go version: go1.25.5 linux/amd64
- Config exists: yes
- DB exists: yes

## Possible Cause
The status flag value is not validated against the list of valid status values before being used. The backend likely normalizes unknown values to the default.

## Related Files
- cmd/todoat/cmd/todoat.go (flag handling and task creation)
- backend/backend.go (TaskStatus enum definition)

## Resolution

**Fixed in**: this session
**Fix description**: Added `parseStatusWithValidation` function that returns an error for invalid status values. Applied validation in both `doUpdate` and `doGet` functions. Returns error with message listing valid status values (TODO, IN-PROGRESS, DONE, CANCELLED).
**Test added**: `TestInvalidStatusRejectedCLI` and `TestValidStatusesAcceptedCLI` in `cmd/todoat/cmd/todoat_test.go`
