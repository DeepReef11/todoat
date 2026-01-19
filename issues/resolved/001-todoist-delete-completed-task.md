# [001] Todoist delete cannot find completed tasks by summary

## Category
todoist

## Severity
medium

## Steps to Reproduce
```bash
export TODOAT_TODOIST_TOKEN="your_token"
./bin/todoat -y config set default_backend todoist
./bin/todoat -y Inbox add "Test from todoat"
./bin/todoat -y Inbox update "Test from todoat" -s DONE
./bin/todoat -y Inbox delete "Test from todoat"
```

## Expected Behavior
The delete command should find the completed task by its summary and delete it.

## Actual Behavior
The delete command returns an error saying the task cannot be found:
```
Error: no task found matching 'Test from todoat'
```

The task exists but is completed and not visible to the search. The workaround is to use `--uid` flag:
```bash
./bin/todoat -y Inbox --uid 9932973026 delete
```

## Error Output
```
Error: no task found matching 'Test from todoat'
Exit code: 1
```

## Environment
- OS: Linux
- Go version: go1.25.5 linux/amd64
- Config exists: yes
- DB exists: yes

## Possible Cause
The Todoist backend doesn't include completed tasks in its search results, so delete cannot find them. Note that the SQLite backend does NOT have this issue - it can successfully find and delete completed tasks.

This is likely due to the Todoist API not returning completed tasks in the regular task listing, or the backend implementation filtering them out.

## Related Files
- internal/backend/todoist/todoist.go
- internal/cli/tasks.go

## Resolution

**Fixed in**: this session
**Fix description**: Modified `GetTasks` in the Todoist backend to fetch both active tasks from the REST API (`/rest/v2/tasks`) and completed tasks from the Sync API (`/sync/v9/completed/get_all`), then merge them. This allows the CLI to find completed tasks by summary for operations like delete.
**Test added**: `TestTodoistFindCompletedTaskBySummaryCLI` in `backend/todoist/todoist_test.go`
