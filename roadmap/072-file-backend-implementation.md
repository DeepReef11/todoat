# [072] File Backend Full Implementation

## Summary
Replace the placeholder File backend implementation with a fully functional file-based storage backend, enabling lightweight task storage without Git dependency.

## Documentation Reference
- Primary: `docs/explanation/backend-system.md`
- Section: File Backend - noted as "placeholder only, not production-ready"

## Dependencies
- Requires: [002] Core CLI
- Requires: [004] Task Commands

## Complexity
M

## Acceptance Criteria

### Tests Required
- [ ] `TestFileBackendAddTask` - `todoat -b file "Work" add "Task"` creates task in configured file
- [ ] `TestFileBackendGetTasks` - `todoat -b file "Work"` lists tasks from file
- [ ] `TestFileBackendUpdateTask` - `todoat -b file "Work" update "Task" -s D` updates task
- [ ] `TestFileBackendDeleteTask` - `todoat -b file "Work" delete "Task"` removes task
- [ ] `TestFileBackendListManagement` - Sections in file treated as task lists
- [ ] `TestFileBackendMetadata` - Tasks store priority, dates, status, tags

### Functional Requirements
- [ ] All TaskManager interface methods implemented (not stubs)
- [ ] File created automatically if doesn't exist
- [ ] Sections (## headings) represent task lists
- [ ] Tasks stored as list items with metadata
- [ ] Hierarchical tasks via indentation
- [ ] File path configurable via config

## Implementation Notes
- Replace stub methods in `backend/file/backend.go` with real implementations
- Use similar parsing approach to Git backend markdown parser
- Support one task per line with metadata encoding
- Atomic file writes to prevent corruption
- No Git dependency (unlike Git backend)

## Out of Scope
- Multiple file support
- Binary formats
- Cloud sync (use sync manager with remote backends)
