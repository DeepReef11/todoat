# [069] Google Tasks CLI Integration

## Summary
Wire the existing Google Tasks backend implementation to the CLI, enabling users to access Google Tasks via `--backend=google` flag.

## Documentation Reference
- Primary: `docs/explanation/backends.md`
- Section: Google Tasks - noted as "backend code exists but not wired to CLI"

## Dependencies
- Requires: [027] Google Tasks Backend (backend implementation)

## Complexity
M

## Acceptance Criteria

### Tests Required
- [ ] `TestGoogleTasksCLIListCommand` - `todoat -b google list` shows Google task lists
- [ ] `TestGoogleTasksCLIGetTasks` - `todoat -b google "My Tasks"` retrieves tasks
- [ ] `TestGoogleTasksCLIAddTask` - `todoat -b google "My Tasks" add "Task"` creates task
- [ ] `TestGoogleTasksCLIUpdateTask` - `todoat -b google "My Tasks" update "Task" -s D` updates task
- [ ] `TestGoogleTasksCLIDeleteTask` - `todoat -b google "My Tasks" delete "Task"` removes task

### Functional Requirements
- [ ] Backend type "google" is recognized in config and CLI
- [ ] `--backend=google` flag works with all task commands
- [ ] Backend appears in `todoat backends` list when configured
- [ ] OAuth2 authentication flow initiated on first use
- [ ] Credentials stored via credential management system

## Implementation Notes
- Register Google Tasks backend type in registry
- Add "google" to backend type validation
- Ensure backend factory correctly instantiates GoogleTasksBackend
- Update documentation to remove "not yet available" disclaimer

## Out of Scope
- Backend implementation changes (already done in 027)
- New API features
- Multiple Google accounts
