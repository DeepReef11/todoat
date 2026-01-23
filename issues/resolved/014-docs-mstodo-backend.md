# [014] Docs: Microsoft To Do backend has no documentation

## Type
documentation

## Severity
medium

## Test Location
- File: backend/mstodo/mstodo_test.go
- Tests: 30+ test functions covering full CRUD operations

## Feature Description
The Microsoft To Do (MS Todo) backend is fully implemented and tested but not documented in docs/explanation/backends.md or referenced in the CLI documentation. Tests verify:
- OAuth2 authentication flow
- Token refresh
- Task list management
- Task CRUD operations
- Subtasks
- Importance (priority) mapping
- Status mapping
- Due dates

## Expected Documentation
- Location: docs/explanation/backends.md
- Section: Microsoft To Do

Should cover:
- [x] Configuration example in config.yaml
- [x] Microsoft Graph API setup instructions
- [x] OAuth2 authentication flow
- [x] How to store tokens
- [x] Usage examples
- [x] Importance/priority mapping explanation

## Resolution

**Fixed in**: this session
**Fix description**: Added comprehensive Microsoft To Do backend documentation to docs/explanation/backends.md

### Verification Log
```bash
$ grep -n "Microsoft To Do" docs/explanation/backends.md
12:| Microsoft To Do | `mstodo` | Microsoft To Do cloud service |
186:## Microsoft To Do
199:Microsoft To Do requires OAuth2 authentication via Microsoft Graph API. You'll need to create an app registration in the Azure portal:
241:# Add a task to Microsoft To Do
265:Microsoft To Do uses "importance" with three levels (low, normal, high). todoat maps these to numeric priorities:
284:- **No trash/restore**: Microsoft To Do permanently deletes tasks and lists (no trash recovery)
285:- **No tags/categories**: Microsoft To Do does not support labels or categories in the API
```

Documentation includes:
- Configuration section with YAML example
- OAuth2 setup instructions with Azure portal steps
- Environment variable setup for tokens
- Usage examples for common operations
- Supported features table
- Priority/Importance mapping table
- Status mapping table
- Limitations section

**Matches expected behavior**: YES
