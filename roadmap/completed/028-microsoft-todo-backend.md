# [028] Microsoft To Do Backend

## Summary
Implement Microsoft To Do integration via Microsoft Graph API, supporting full CRUD operations on tasks and task lists with OAuth2 authentication.

## Documentation Reference
- Primary: `docs/explanation/features-overview.md` (Planned features - Microsoft To Do backend)
- Related: `docs/explanation/backend-system.md`, `docs/explanation/credential-management.md`

## Dependencies
- Requires: [017] Credential Management
- Requires: [003] SQLite Backend (for TaskManager interface)

## Complexity
L

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestMSTodoListTaskLists` - `todoat --backend=mstodo list` shows Microsoft To Do lists
- [ ] `TestMSTodoGetTasks` - `todoat --backend=mstodo MyList` retrieves tasks
- [ ] `TestMSTodoAddTask` - `todoat --backend=mstodo MyList add "Task"` creates task via Graph API
- [ ] `TestMSTodoUpdateTask` - `todoat --backend=mstodo MyList update "Task" -s D` updates task
- [ ] `TestMSTodoDeleteTask` - `todoat --backend=mstodo MyList delete "Task"` removes task
- [ ] `TestMSTodoOAuth2Flow` - OAuth2 authentication with Microsoft identity platform
- [ ] `TestMSTodoTokenRefresh` - Automatic token refresh when expired
- [ ] `TestMSTodoSubtasks` - Checklist items map to subtasks
- [ ] `TestMSTodoImportance` - Priority maps to importance (low/normal/high)
- [ ] `TestMSTodoStatusMapping` - Status maps: completed/notStarted/inProgress

## Implementation Notes
- Create `backend/mstodo/` package
- Use Microsoft Graph API: `https://graph.microsoft.com/v1.0/me/todo/`
- OAuth2 with Microsoft identity platform
- Store tokens in system keyring
- MS To Do has: notStarted, inProgress, completed status
- Importance levels: low, normal, high (map from 1-9 priority)
- Checklist items can represent subtasks

## Out of Scope
- Outlook Tasks (legacy API)
- Shared task lists
- Microsoft Teams integration
- Multiple Microsoft accounts
