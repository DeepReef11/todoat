# Error Reference

This document describes common errors you may encounter when using todoat and how to resolve them.

## Overview

Todoat provides helpful error messages with actionable suggestions. When an error occurs, the output includes:
- The error message explaining what went wrong
- A suggestion for how to fix the issue

## Task Errors

### Task Not Found

**Message**: `no task found matching '<search term>'`

**Cause**: The specified task could not be found with the given search term.

**Example**:
```bash
$ todoat MyList complete "nonexistent task"
Error: no task found matching 'nonexistent task'
```

### List Not Found

**Message**: `list not found: <list name>`

**Cause**: The specified list does not exist.

**Example**:
```bash
$ todoat work
Error: list not found: work

Suggestion: Create the list with 'todoat list create work'
```

**Note**: The `add` action auto-creates the list if it does not exist, so `todoat work add "task"` will create the list and add the task in one step.

### No Lists Available

**Message**: `no lists available`

**Cause**: No task lists have been created yet.

**Suggestion**: Create a list with `todoat list create <name>`.

## Validation Errors

### Invalid Priority

**Message**: `invalid priority: <value>`

**Cause**: The priority value is outside the valid range.

**Suggestion**: Priority must be between 0 and 9 (1 = highest, 9 = lowest, 0 = none).

**Example**:
```bash
$ todoat MyList add "task" -p 15
Error: invalid priority: 15

Suggestion: Priority must be between 0 and 9
```

### Invalid Date

**Message**: `invalid date: <date string>`

**Cause**: The date format is not recognized.

**Suggestion**: Use date format `YYYY-MM-DD` (e.g., 2026-01-15) or `YYYY-MM-DDTHH:MM` for date-time.

**Example**:
```bash
$ todoat MyList add "task" --due-date "next week"
Error: invalid date: next week

Suggestion: Use date format YYYY-MM-DD (e.g., 2026-01-15)
```

### Invalid Status

**Message**: `invalid status: <status>`

**Cause**: The provided status is not a valid option.

**Suggestion**: Use one of the valid status values: `TODO`, `IN-PROGRESS`, `DONE`, `CANCELLED`.

**Example**:
```bash
$ todoat MyList update "task" -s "pending"
Error: invalid status: pending

Suggestion: Valid options: TODO, IN-PROGRESS, DONE, CANCELLED
```

## Backend Errors

### List Deletion Not Supported (Nextcloud)

**Message**: `deleting calendars is not supported via CalDAV (would be permanent)`

**Cause**: The Nextcloud backend does not support list deletion to prevent accidental data loss.

**Example**:
```bash
$ todoat -b nextcloud list delete "Work"
Error: deleting calendars is not supported via CalDAV (would be permanent)
```

**Fix**: Delete the calendar directly in the Nextcloud web interface if you need to remove it.

### Sharing Not Supported

**Message**: `sharing is not supported by this backend (requires Nextcloud)`

**Cause**: The `list share` or `list unshare` command was used with a backend that does not support list sharing.

**Example**:
```bash
$ todoat -b sqlite list share "Work" --user alice
Error: sharing is not supported by this backend (requires Nextcloud)
```

**Fix**: Use the Nextcloud backend for list sharing:
```bash
todoat -b nextcloud list share "Work" --user alice --permission write
```

### Publishing Not Supported

**Message**: `publishing is not supported by this backend (requires Nextcloud)`

**Cause**: The `list publish` or `list unpublish` command was used with a backend that does not support public link publishing.

**Example**:
```bash
$ todoat -b sqlite list publish "Work"
Error: publishing is not supported by this backend (requires Nextcloud)
```

**Fix**: Use the Nextcloud backend for public link publishing:
```bash
todoat -b nextcloud list publish "Work"
```

### Subscriptions Not Supported

**Message**: `subscriptions are not supported by this backend (requires Nextcloud)`

**Cause**: The `list subscribe` or `list unsubscribe` command was used with a backend that does not support calendar subscriptions.

**Example**:
```bash
$ todoat -b sqlite list subscribe "https://example.com/calendar.ics"
Error: subscriptions are not supported by this backend (requires Nextcloud)
```

**Fix**: Use the Nextcloud backend for calendar subscriptions:
```bash
todoat -b nextcloud list subscribe "https://example.com/calendar.ics"
```

### Backend Not Configured

**Message**: `todoist backend requires API token (use 'credentials set todoist token' or set TODOAT_TODOIST_TOKEN)`

**Cause**: The specified backend has not been set up with the required credentials.

**Example**:
```bash
$ todoat -b todoist list
Error: todoist backend requires API token (use 'credentials set todoist token' or set TODOAT_TODOIST_TOKEN)
```

**Fix**: Set the API token using one of these methods:
1. Run `todoat credentials set todoist token` to configure interactively
2. Set the environment variable `TODOAT_TODOIST_TOKEN`

### Backend Offline

**Message**: `backend <name> is offline: <reason>`

**Cause**: The backend service cannot be reached.

**Suggestions** (context-aware based on error):
- DNS/hostname errors: "Check your DNS settings and internet connection"
- Connection refused: "Check if the server is running and accessible"
- Timeout: "The server may be slow or unreachable. Try again later"
- Other: "Check your internet connection and try again"

**Example**:
```bash
$ todoat -b nextcloud list
Error: backend nextcloud is offline: connection refused

Suggestion: Check if the server is running and accessible
```

## Authentication Errors

### Credentials Not Found

**Message**: `credentials not found for <backend> user <username>`

**Cause**: No stored credentials for the specified backend and user.

**Suggestion**: Run `todoat credentials set <backend> <username> --prompt` to configure credentials.

**Example**:
```bash
$ todoat -b todoist list
Error: credentials not found for todoist user john@example.com

Suggestion: Run 'todoat credentials set todoist token --prompt' to configure credentials
```

### Authentication Failed

**Message**: `authentication failed for <backend>`

**Cause**: The stored credentials are invalid, expired, or rejected by the backend.

**Suggestion**: Verify your credentials are correct and have not expired.

**Possible fixes**:
1. Update credentials: `todoat credentials set <backend> <username> --prompt`
2. Check if your API token has expired (Todoist, Google Tasks)
3. Verify your password hasn't changed (Nextcloud)
4. Ensure your account has not been locked

## Rate Limit Errors

### Rate Limit Exceeded

**Message**: `<backend> rate limit exceeded after <n> retries (max <max>)`

**Cause**: The backend API returned HTTP 429 (Too Many Requests) and all automatic retries were exhausted. todoat retries rate-limited requests up to 5 times with exponential backoff before failing.

**Example**:
```bash
$ todoat -b todoist list
Error: todoist rate limit exceeded after 5 retries (max 5)
```

**Fix**:
1. Wait a few minutes before trying again
2. Check if you're running multiple instances of todoat against the same backend
3. Consider spacing out bulk operations
4. Verify your API token has adequate quotas

## Sync Errors

### Sync Not Enabled

**Message**: `sync is not enabled`

**Cause**: Sync functionality has not been configured.

**Suggestion**: Enable sync in your config file or run `todoat config edit`.

**Configuration example**:
```yaml
sync:
  enabled: true
```

## Debugging

### Enable Verbose Mode

For more detailed error information, use the `-V` or `--verbose` flag:

```bash
$ todoat -V list
```

This shows additional debug output including:
- Configuration being used
- Backend connection details
- Full error stack traces

### Check Configuration

View your current configuration:

```bash
$ todoat config get
```

Edit configuration:

```bash
$ todoat config edit
```
