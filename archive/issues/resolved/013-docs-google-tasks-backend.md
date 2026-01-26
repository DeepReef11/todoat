# [013] Docs: Google Tasks backend has no documentation

## Type
documentation

## Severity
medium

## Test Location
- File: backend/google/google_test.go
- Tests: 30+ test functions covering full CRUD operations

## Feature Description
The Google Tasks backend is fully implemented and tested but not documented in docs/explanation/backends.md or referenced in the CLI documentation. Tests verify:
- OAuth2 authentication flow
- Token refresh
- Task list management (CRUD)
- Task management (CRUD)
- Subtask hierarchy
- Due dates
- Status mapping

## Expected Documentation
- Location: docs/explanation/backends.md
- Section: Google Tasks

Should cover:
- [x] Configuration example in config.yaml
- [x] OAuth2 setup instructions
- [x] How to authenticate and store tokens
- [x] Usage examples
- [x] Any limitations compared to other backends

## Resolution

**Fixed in**: this session
**Fix description**: Added comprehensive Google Tasks backend documentation to docs/explanation/backends.md

### Verification Log
```bash
$ grep -n "Google Tasks" docs/explanation/backends.md
11:| Google Tasks | `google` | Google Tasks cloud service |
112:## Google Tasks
125:Google Tasks requires OAuth2 authentication. You'll need to create credentials in the Google Cloud Console:
129:3. Enable the **Google Tasks API** in APIs & Services > Library
158:# Add a task to Google Tasks
178:- **No trash/restore**: Google Tasks permanently deletes tasks and lists (no trash recovery)
179:- **Status mapping**: Google Tasks only supports "needsAction" and "completed" statuses. IN-PROGRESS and CANCELLED tasks are mapped to "completed"
180:- **No priorities**: Google Tasks does not support task priorities
181:- **No tags/categories**: Google Tasks does not support labels or categories
```
**Matches expected behavior**: YES

Documentation now includes:
- Configuration example showing type: google backend
- OAuth2 setup instructions with Google Cloud Console steps
- Environment variable authentication (access/refresh tokens, client ID/secret)
- Usage examples for list, add, and view commands
- Comprehensive limitations table covering all unsupported features
