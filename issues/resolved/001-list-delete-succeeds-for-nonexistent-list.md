# [001] List Delete Succeeds for Non-Existent List

## Type
code-bug

## Category
error-handling

## Severity
low

## Steps to Reproduce
```bash
# Verify the list does not exist
./todoat list

# Attempt to delete a non-existent list
./todoat list delete "NonExistentList" -y
```

## Expected Behavior
The command should return an error indicating the list does not exist, such as:
```
Error: list 'NonExistentList' not found
```

## Actual Behavior
The command returns success:
```
Deleted list: NonExistentList
ACTION_COMPLETED
```

The command completes successfully and reports that it deleted the list, even though the list never existed. The "deleted" list is even added to the trash (visible via `todoat list trash`).

## Error Output
```
Deleted list: NonExistentList
ACTION_COMPLETED
```

## Environment
- OS: Linux (Docker container)
- Runtime version: Go 1.21+

## Possible Cause
The delete operation likely uses a soft-delete pattern (setting `deleted_at` timestamp) that doesn't first check if the list exists. The SQLite backend may be executing an UPDATE or INSERT that succeeds even when the list doesn't exist in the active lists.

## Related Files
- `backend/sqliteBackend.go` - List deletion logic
- `cmd/list.go` - List delete command handler

## Recommended Fix
FIX CODE - Add existence check before deletion:
1. Query for the list by name before attempting deletion
2. Return an error if the list is not found
3. Only proceed with soft-delete if list exists

## Resolution

**Fixed in**: Already implemented in doListDelete function (cmd/todoat/cmd/todoat.go:884-912)
**Fix description**: The doListDelete function already checks if the list exists by calling GetListByName before attempting deletion. If the list is not found, it returns an error message "Error: list '%s' not found" with exit code 1 and ERROR result code.
**Test added**: Enhanced `TestListDeleteNotFoundSQLiteCLI` in backend/sqlite/cli_test.go to be a comprehensive regression test for this issue

### Verification Log
```bash
$ export XDG_DATA_HOME=/tmp/todoat_test_001/data
$ export XDG_CONFIG_HOME=/tmp/todoat_test_001/config
$ ./todoat list
No lists found. Create one with: todoat list create "MyList"

$ ./todoat list delete "NonExistentList" -y
Error: list 'NonExistentList' not found
ERROR
Exit code: 1

$ ./todoat list trash
Trash is empty.
```
**Matches expected behavior**: YES - The command correctly returns an error indicating the list does not exist, instead of falsely reporting "Deleted list: NonExistentList" with ACTION_COMPLETED.
