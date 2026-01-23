# Error Reference

This document describes common errors you may encounter when using todoat and how to resolve them.

## Overview

Todoat provides helpful error messages with actionable suggestions. When an error occurs, the output includes:
- The error message explaining what went wrong
- A suggestion for how to fix the issue

## Task Errors

### Task Not Found

**Message**: `task not found: <search term>`

**Cause**: The specified task could not be found with the given search term.

**Suggestion**: Check the search term or use `todoat list` to see all tasks.

**Example**:
```bash
$ todoat complete "nonexistent task"
Error: task not found: nonexistent task

Suggestion: Check the search term or use 'todoat list' to see all tasks
```

### List Not Found

**Message**: `list not found: <list name>`

**Cause**: The specified list does not exist.

**Suggestion**: Create the list with `todoat list create <name>`.

**Example**:
```bash
$ todoat work get
Error: list not found: work

Suggestion: Create the list with 'todoat list create work'
```

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
$ todoat add "task" -p 15
Error: invalid priority: 15

Suggestion: Priority must be between 0 and 9
```

### Invalid Date

**Message**: `invalid date: <date string>`

**Cause**: The date format is not recognized.

**Suggestion**: Use date format `YYYY-MM-DD` (e.g., 2026-01-15) or `YYYY-MM-DDTHH:MM` for date-time.

**Example**:
```bash
$ todoat add "task" --due-date "next week"
Error: invalid date: next week

Suggestion: Use date format YYYY-MM-DD (e.g., 2026-01-15)
```

### Invalid Status

**Message**: `invalid status: <status>`

**Cause**: The provided status is not a valid option.

**Suggestion**: Use one of the valid status values: `TODO`, `IN-PROGRESS`, `DONE`, `CANCELLED`.

**Example**:
```bash
$ todoat update "task" -s "pending"
Error: invalid status: pending

Suggestion: Valid options: TODO, IN-PROGRESS, DONE, CANCELLED
```

## Backend Errors

### Backend Not Configured

**Message**: `backend not configured: <backend name>`

**Cause**: The specified backend has not been set up in your configuration.

**Suggestion**: Add the backend configuration to your config file or run setup.

**Example**:
```bash
$ todoat -b todoist list
Error: backend not configured: todoist

Suggestion: Add todoist configuration to your config file
```

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

**Suggestion**: Run `todoat setup <backend>` to configure credentials.

**Example**:
```bash
$ todoat -b todoist list
Error: credentials not found for todoist user john@example.com

Suggestion: Run 'todoat setup todoist' to configure credentials
```

### Authentication Failed

**Message**: `authentication failed for <backend>`

**Cause**: The stored credentials are invalid, expired, or rejected by the backend.

**Suggestion**: Verify your credentials are correct and have not expired.

**Possible fixes**:
1. Re-run setup: `todoat setup <backend>`
2. Check if your API token has expired (Todoist, Google Tasks)
3. Verify your password hasn't changed (Nextcloud)
4. Ensure your account has not been locked

## Sync Errors

### Sync Not Enabled

**Message**: `sync is not enabled`

**Cause**: Sync functionality has not been configured.

**Suggestion**: Enable sync in your config file or run `todoat config edit`.

**Configuration example**:
```yaml
sync:
  enabled: true
  interval: 5m
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
