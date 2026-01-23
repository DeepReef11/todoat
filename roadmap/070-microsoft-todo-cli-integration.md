# [070] Microsoft To Do CLI Integration

## Summary
Wire the existing Microsoft To Do backend implementation to the CLI, enabling users to access Microsoft To Do via `--backend=mstodo` flag.

## Documentation Reference
- Primary: `docs/explanation/backends.md`
- Section: Microsoft To Do - noted as "backend code exists but not wired to CLI"

## Dependencies
- Requires: [028] Microsoft To Do Backend (backend implementation)

## Complexity
M

## Acceptance Criteria

### Tests Required
- [ ] `TestMSTodoCLIListCommand` - `todoat -b mstodo list` shows Microsoft To Do lists
- [ ] `TestMSTodoCLIGetTasks` - `todoat -b mstodo "My Tasks"` retrieves tasks
- [ ] `TestMSTodoCLIAddTask` - `todoat -b mstodo "My Tasks" add "Task"` creates task
- [ ] `TestMSTodoCLIUpdateTask` - `todoat -b mstodo "My Tasks" update "Task" -s D` updates task
- [ ] `TestMSTodoCLIDeleteTask` - `todoat -b mstodo "My Tasks" delete "Task"` removes task

### Functional Requirements
- [ ] Backend type "mstodo" is recognized in config and CLI
- [ ] `--backend=mstodo` flag works with all task commands
- [ ] Backend appears in `todoat backends` list when configured
- [ ] OAuth2 authentication flow initiated on first use
- [ ] Credentials stored via credential management system

## Implementation Notes
- Register Microsoft To Do backend type in registry
- Add "mstodo" to backend type validation
- Ensure backend factory correctly instantiates MSTodoBackend
- Update documentation to remove "not yet available" disclaimer

## Out of Scope
- Backend implementation changes (already done in 028)
- New Graph API features
- Multiple Microsoft accounts
