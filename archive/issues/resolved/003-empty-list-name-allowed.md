# [003] Empty list name is allowed during creation

## Type
code-bug

## Category
feature

## Severity
low

## Steps to Reproduce
```bash
todoat list create ""
todoat list show
```

## Expected Behavior
The command should fail with a validation error indicating that list name cannot be empty.

## Actual Behavior
A list with an empty name is created successfully.

```
Created list:
```

The empty list appears in `list show` output as a blank row:
```
NAME                 TASKS
...
                     0
```

## Error Output
No error is shown - the operation succeeds silently.

## Environment
- OS: Linux
- Runtime version: Go 1.25.5

## Possible Cause
Missing input validation for list name in the `list create` command handler.

## Related Files
- `cmd/todoat/cmd/todoat.go` - CLI command handlers

## Suggested Fix
Add validation to ensure list name is not empty (and potentially validate against other invalid characters or patterns) before creating the list.

## Resolution

**Fixed in**: this session
**Fix description**: Added validation at the beginning of `doListCreate()` in `cmd/todoat/cmd/todoat.go` to trim whitespace and reject empty list names with a clear error message.
**Test added**: `TestListCreateEmptyNameSQLiteCLI` in `backend/sqlite/cli_test.go`

### Verification Log
```bash
$ todoat list create ""
Error: list name cannot be empty
Exit code: 1
```
**Matches expected behavior**: YES
