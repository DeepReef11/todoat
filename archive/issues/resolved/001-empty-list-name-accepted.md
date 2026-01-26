# [001] Empty list name is accepted when creating tasks

## Category
error-handling

## Severity
medium

## Steps to Reproduce
```bash
./bin/todoat "" add "Test"
```

## Expected Behavior
The CLI should return an error message indicating that a list name is required.

## Actual Behavior
The task is created successfully in a list with an empty name.

```
Created task: Test (ID: 1c712d30-9616-4020-b8eb-f934d12be75f)
Exit code: 0
```

The empty-named list then appears in list output:
```
Available lists (3):

NAME                 TASKS
MyList               4
NonExistentList12345 0
                     1
```

## Error Output
No error produced - command succeeds when it should fail.

## Environment
- OS: Linux
- Go version: go1.25.5 linux/amd64
- Config exists: yes
- DB exists: yes

## Possible Cause
The CLI does not validate that the list name argument is non-empty before creating tasks or lists.

## Related Files
- cmd/todoat/cmd/todoat.go (task creation logic)
- backend/sqlite/sqlite.go (backend list creation)

## Resolution

**Fixed in**: this session
**Fix description**: Added validation in cmd/todoat/cmd/todoat.go to check that list name is not empty or whitespace-only before processing commands. Returns error "list name cannot be empty" if validation fails.
**Test added**: `TestEmptyListNameRejectedCLI` and `TestWhitespaceOnlyListNameRejectedCLI` in `cmd/todoat/cmd/todoat_test.go`
