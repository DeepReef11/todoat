# [027] Google Tasks Backend

## Summary
Implement Google Tasks integration via Google Tasks API, supporting full CRUD operations on tasks and task lists with OAuth2 authentication.

## Documentation Reference
- Primary: `dev-doc/FEATURES_OVERVIEW.md` (Planned features - Google Tasks backend)
- Related: `dev-doc/BACKEND_SYSTEM.md`, `dev-doc/CREDENTIAL_MANAGEMENT.md`

## Dependencies
- Requires: [017] Credential Management
- Requires: [003] SQLite Backend (for TaskManager interface)

## Complexity
L

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestGoogleTasksListTaskLists` - `todoat --backend=google list` shows Google task lists
- [ ] `TestGoogleTasksGetTasks` - `todoat --backend=google MyList` retrieves tasks from Google
- [ ] `TestGoogleTasksAddTask` - `todoat --backend=google MyList add "Task"` creates task via API
- [ ] `TestGoogleTasksUpdateTask` - `todoat --backend=google MyList update "Task" -s D` updates task
- [ ] `TestGoogleTasksDeleteTask` - `todoat --backend=google MyList delete "Task"` removes task
- [ ] `TestGoogleTasksOAuth2Flow` - OAuth2 authentication flow with token storage
- [ ] `TestGoogleTasksTokenRefresh` - Automatic token refresh when expired
- [ ] `TestGoogleTasksSubtasks` - Parent-child relationships sync correctly
- [ ] `TestGoogleTasksDueDate` - Due dates map correctly to Google Tasks format
- [ ] `TestGoogleTasksStatusMapping` - Status maps: completed/needs-action

## Implementation Notes
- Create `backend/google/` package
- Use Google Tasks API v1: `https://www.googleapis.com/tasks/v1/`
- OAuth2 required (no API key option)
- Store tokens in system keyring
- Google Tasks has limited status: only completed or needsAction
- No priority support in Google Tasks (use notes field or ignore)
- Due date is date-only (no time) in Google Tasks

## Out of Scope
- Google Calendar integration (different API)
- Recurring tasks
- Task notes as description (API limitation)
- Multiple Google accounts
