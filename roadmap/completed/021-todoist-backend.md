# [021] Todoist Backend

## Summary
Implement Todoist integration via REST API v2, supporting full CRUD operations on tasks and projects with authentication via API token stored in keyring or environment variables.

## Documentation Reference
- Primary: `dev-doc/BACKEND_SYSTEM.md#todoist-backend`
- Related: `dev-doc/CREDENTIAL_MANAGEMENT.md`, `dev-doc/CONFIGURATION.md`

## Dependencies
- Requires: [017] Credential Management (for API token storage)
- Requires: [003] SQLite Backend (for TaskManager interface definition)

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestTodoistListProjects` - `todoat --backend=todoist list` shows Todoist projects
- [ ] `TestTodoistGetTasks` - `todoat --backend=todoist MyProject` retrieves tasks from Todoist
- [ ] `TestTodoistAddTask` - `todoat --backend=todoist MyProject add "Task"` creates task via API
- [ ] `TestTodoistUpdateTask` - `todoat --backend=todoist MyProject update "Task" -s DONE` updates task
- [ ] `TestTodoistDeleteTask` - `todoat --backend=todoist MyProject delete "Task"` removes task
- [ ] `TestTodoistPriorityMapping` - Internal priority 1-9 maps to Todoist priority 1-4 correctly
- [ ] `TestTodoistLabelsAsCategories` - Todoist labels map to todoat categories/tags
- [ ] `TestTodoistAPITokenFromKeyring` - Backend retrieves API token from system keyring
- [ ] `TestTodoistAPITokenFromEnv` - Backend retrieves token from `TODOAT_TODOIST_TOKEN` env var
- [ ] `TestTodoistRateLimiting` - Backend respects Todoist API rate limits with backoff
- [ ] `TestTodoistSubtasks` - Parent-child relationships sync correctly with Todoist

## Implementation Notes
- Use Todoist REST API v2: `https://api.todoist.com/rest/v2/`
- API token required (no OAuth flow in scope)
- Priority mapping: Todoist uses 1-4 (4=highest), todoat uses 1-9 (1=highest)
- Status mapping: Todoist tasks are either completed or not (no in-progress status)
- Projects map to todoat lists
- Sections can be used for subtask grouping (optional enhancement)

## Out of Scope
- Todoist Premium features (reminders, comments)
- Todoist Sync API (only REST API)
- OAuth authentication flow
- Todoist sections/labels management commands
