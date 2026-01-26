# [026] Docs: Error handling and suggestions not documented

## Type
documentation

## Severity
low

## Test Location
- File: internal/utils/errors_test.go
- Functions:
  - TestErrAuthenticationFailed
  - TestErrBackendNotConfigured
  - TestErrBackendOfflineConnectionRefused
  - TestErrBackendOfflineDNS
  - TestErrBackendOfflineTimeout
  - TestErrCredentialsNotFound
  - TestErrInvalidDate
  - TestErrInvalidPriority
  - TestErrInvalidStatus
  - TestErrListNotFound
  - TestErrNoListsAvailable
  - TestErrSyncNotEnabled
  - TestErrTaskNotFound
  - TestErrorWithSuggestionError
  - TestErrorWithSuggestionGetSuggestion
  - TestWrapWithSuggestion

## Feature Description
The codebase has a sophisticated error handling system with:
- Typed errors (ErrAuthenticationFailed, ErrBackendOffline, etc.)
- Suggestion system for common errors
- Offline detection with specific messages

This helps users troubleshoot issues but isn't documented.

## Expected Documentation
- Location: docs/reference/ (new file) or docs/explanation/
- Suggested file: docs/reference/errors.md

Should cover:
- [x] Common error messages and their meaning
- [x] Suggested fixes (from ErrorWithSuggestion system)
- [x] Authentication errors and how to resolve
- [x] Offline/connectivity errors
- [x] How to enable verbose mode for debugging

Example content:
```markdown
## Common Errors

### Authentication Failed
**Message**: `authentication failed for backend [name]`
**Cause**: Invalid credentials or expired token
**Fix**: Run `todoat credentials update [backend] [username] --prompt`

### Backend Offline
**Message**: `backend [name] is offline`
**Cause**: Network connectivity issues
**Fix**: Check network, verify host is reachable
```

## Resolution

**Fixed in**: this session
**Fix description**: Created comprehensive error reference documentation at docs/reference/errors.md

### Verification Log
```bash
$ ls docs/reference/errors.md
docs/reference/errors.md

$ head -20 docs/reference/errors.md
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
```

**Coverage verified**:
- 11 error messages documented with meaning
- 10 suggestion fixes from ErrorWithSuggestion system
- Authentication errors section included
- Backend Offline/connectivity errors section included
- Debugging section with verbose mode instructions included

**Matches expected behavior**: YES
