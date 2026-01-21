# No Tasks Shown in Task List Despite Tasks Existing

## Summary
Task list shows no tasks even though tasks exist on the backend.

## Steps to Reproduce

1. Have tasks on a Nextcloud backend
2. Run `todoat -b nextcloud` or similar list command

## Expected Behavior
Tasks should be displayed.

## Actual Behavior
No tasks shown in the task list, but tasks exist on the backend.

## Workaround
Using `todoat -b nextcloud-test` (different backend name) shows the tasks correctly.

## Related
- Issue #012: HTTP/HTTPS config inconsistency between identical backends
- The empty list may be a side effect of the HTTP/HTTPS protocol mismatch

## Questions
1. Is this a separate issue or a symptom of the HTTP/HTTPS problem?
2. Does the initial connection failure cause subsequent list operations to return empty?

## Resolution

**Fixed in**: Issue #012 resolution (c602696)
**Fix description**: This issue was a symptom of issue #012 (HTTP/HTTPS config inconsistency). When using `-b nextcloud`, the built-in backend name was not reading config file settings (including `allow_http`), causing the backend to default to HTTPS even when the server only supported HTTP. This resulted in connection errors which manifested as empty task lists.

The fix in issue #012 modified `createBackendByName` to check if "nextcloud" is configured in the config file before falling back to environment-only mode. When a config file entry exists, it now uses `createCustomBackend` which properly reads all config file settings.

**Root cause**: Not a separate issue - it was a symptom of the HTTP/HTTPS protocol mismatch described in issue #012.

**Test added**: `TestIssue013NoTasksShownResolved` in `backend/nextcloud/issue013_no_tasks_test.go`

### Verification Log
```bash
$ go test ./backend/nextcloud/... -run "Issue013" -v
=== RUN   TestIssue013NoTasksShownResolved
--- PASS: TestIssue013NoTasksShownResolved (0.00s)
PASS
ok      todoat/backend/nextcloud    0.005s
```

Test verifies that:
1. `-b nextcloud` and `-b nextcloud-test` now behave identically when configured the same
2. Neither backend requires environment variables when config file has credentials
3. Neither backend gets HTTP/HTTPS protocol mismatch errors when `allow_http: true` is set

**Matches expected behavior**: YES - Both backends now work identically, reading config file settings properly.
