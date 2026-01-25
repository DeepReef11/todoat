# [007] Sync Fails When Local List Doesn't Exist on CalDAV Remote

## Type
code-bug

## Severity
high

## Source
User report

## Description

When syncing to Nextcloud/CalDAV, if a local list doesn't exist on the remote, sync fails with "creating calendars is not supported via CalDAV". The sync should either:
1. Map to an existing remote calendar, or
2. Skip the task with a warning, or
3. Use a default/inbox calendar

## Steps to Reproduce

1. Configure Nextcloud backend
2. Create a local list "Tasks" (doesn't exist on Nextcloud as a calendar)
3. Add a task to the list
4. Run `todoat sync`

## Expected Behavior

One of:
- Task syncs to an existing Nextcloud calendar (mapping configured or auto-detected)
- Task syncs to default/inbox calendar
- Clear error explaining user needs to create the calendar on Nextcloud first
- Skip with warning, continue syncing other tasks

## Actual Behavior

```
Sync error for task '15': failed to create list 'Tasks' on remote: creating calendars is not supported via CalDAV
Error: failed to create list 'Tasks' on remote: creating calendars is not supported via CalDAV
```

Sync completely fails, even for tasks in lists that DO exist on remote.

## Code Location

`cmd/todoat/cmd/todoat.go` - `syncCreateOperation()` function

The code tries to create a list on the remote, which CalDAV doesn't support.

## Suggested Fix

1. Before sync, fetch existing calendars from remote
2. Map local lists to remote calendars by name
3. For unmapped lists, either:
   - Use a configurable default calendar
   - Skip tasks with warning
   - Prompt user to create calendar on remote

## Configuration Suggestion

```yaml
sync:
  list_mapping:
    Tasks: "Personal"  # Map local "Tasks" to remote "Personal" calendar
  default_calendar: "Inbox"  # Fallback for unmapped lists
```

## Impact

- Sync is unusable with CalDAV backends unless list names match exactly
- No way to configure list-to-calendar mapping
- Users must manually create calendars on Nextcloud first

## Resolution

**Fixed in**: this session
**Fix description**: When a task belongs to a list that doesn't exist on the remote and the remote backend doesn't support creating lists (like CalDAV/Nextcloud), the sync now skips that task with a clear warning message instead of failing completely. This allows sync to continue for tasks in lists that DO exist on the remote.

**Changes made**:
1. Added `ErrListCreationNotSupported` sentinel error to `backend/interface.go`
2. Updated `backend/nextcloud/nextcloud.go` `CreateList()` to return the sentinel error
3. Updated `cmd/todoat/cmd/todoat.go` `syncCreateOperation()` and `syncUpdateOperation()` to check for this error and skip tasks gracefully with a warning message

**Test added**: `TestCreateListReturnsNotSupported` in `backend/nextcloud/nextcloud_test.go`

### Verification Log
```bash
$ go test -v -run TestCreateListReturnsNotSupported ./backend/nextcloud/...
=== RUN   TestCreateListReturnsNotSupported
--- PASS: TestCreateListReturnsNotSupported (0.00s)
PASS
ok  	todoat/backend/nextcloud	0.004s

$ go test -v -run TestSyncSkipsTasksWhenListCreationNotSupported ./backend/sync/...
=== RUN   TestSyncSkipsTasksWhenListCreationNotSupported
--- PASS: TestSyncSkipsTasksWhenListCreationNotSupported (0.14s)
PASS
ok  	todoat/backend/sync	0.143s

$ go test ./...
ok  	todoat/backend	0.466s
ok  	todoat/backend/file	0.024s
ok  	todoat/backend/git	0.534s
ok  	todoat/backend/google	0.016s
ok  	todoat/backend/mstodo	0.018s
ok  	todoat/backend/nextcloud	0.341s
ok  	todoat/backend/sqlite	16.080s
ok  	todoat/backend/sync	5.796s
ok  	todoat/backend/todoist	0.474s
ok  	todoat/cmd/todoat/cmd	4.179s
ok  	todoat/internal/analytics	(cached)
ok  	todoat/internal/cache	1.093s
ok  	todoat/internal/config	0.024s
ok  	todoat/internal/credentials	(cached)
ok  	todoat/internal/markdown	(cached)
ok  	todoat/internal/migrate	1.185s
ok  	todoat/internal/notification	0.019s
ok  	todoat/internal/ratelimit	(cached)
ok  	todoat/internal/reminder	1.534s
ok  	todoat/internal/shutdown	(cached)
ok  	todoat/internal/testutil	0.004s
ok  	todoat/internal/tui	(cached)
ok  	todoat/internal/utils	(cached)
ok  	todoat/internal/views	11.975s
```
**Matches expected behavior**: YES - Tasks in unmapped lists are now skipped with a warning, allowing sync to continue for other tasks
