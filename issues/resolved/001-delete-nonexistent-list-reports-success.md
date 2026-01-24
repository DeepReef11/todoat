# [001] Delete Nonexistent List Reports Success

## Type
code-bug

## Category
error-handling

## Severity
medium

## Steps to Reproduce
```bash
todoat list delete "NonExistentList" -y
```

## Expected Behavior
The command should fail with an error indicating the list does not exist.

## Actual Behavior
The command reports success: "Deleted list: NonExistentList" and returns ACTION_COMPLETED.

```
$ todoat list delete "NonExistentList" -y
Deleted list: NonExistentList
ACTION_COMPLETED
```

The list never existed in the first place, so the success message is misleading.

## Error Output
```
$ todoat -y list delete "NonExistentList"
Deleted list: NonExistentList
ACTION_COMPLETED
```

## Environment
- OS: Linux
- Runtime version: Go 1.21+

## Possible Cause
The delete operation likely doesn't verify that the list exists before returning success. It may be checking the database delete result and finding 0 rows affected but still returning success.

## Related Files
- cmd/todoat/list_commands.go (likely location)
- internal/backend/sqlite/list.go (likely location)

## Recommended Fix
FIX CODE - Add validation to check if the list exists before deletion. Return an error if the list is not found. This is consistent with the behavior of `todoat list info "NonExistentList"` which correctly returns an error.

## Resolution

**Fixed in**: Already implemented (duplicate of resolved issue 001)
**Fix description**: The doListDelete function checks if the list exists by calling GetListByName before attempting deletion. If the list is not found, it returns "Error: list '%s' not found" with exit code 1.
**Test added**: `TestListDeleteNotFoundSQLiteCLI` in backend/sqlite/cli_test.go

### Verification Log
```bash
$ export XDG_DATA_HOME=/tmp/todoat_test_issue001/data
$ export XDG_CONFIG_HOME=/tmp/todoat_test_issue001/config
$ rm -rf /tmp/todoat_test_issue001
$ ./todoat list delete "NonExistentList" -y
Error: list 'NonExistentList' not found
ERROR
Exit code: 1
```
**Matches expected behavior**: YES - The command fails with an error indicating the list does not exist.
