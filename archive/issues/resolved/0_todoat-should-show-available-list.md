When running `todoat` without args, should show list of task lists.

Example:
```bash
$ todoat
Available lists (1):

NAME                 TASKS
l                    1

```

## Resolution

**Fixed in**: this session
**Fix description**: Modified root command RunE to call doListView when no args provided instead of showing help
**Tests added**: TestNoArgsShowsListsSQLiteCLI, TestNoArgsShowsListsEmptySQLiteCLI in backend/sqlite/cli_test.go; TestRootCommandShowsListsCoreCLI in cmd/todoat/cmd/todoat_test.go

### Verification Log
```bash
$ ./todoat
Available lists (1):

NAME                 TASKS
Work                 1
```
**Matches expected behavior**: YES
