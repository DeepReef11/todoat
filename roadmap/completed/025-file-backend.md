# [025] File Backend

## Summary
Implement the file backend to store tasks in plain text files, providing a lightweight file-based storage option without Git dependency.

## Documentation Reference
- Primary: `dev-doc/BACKEND_SYSTEM.md` (File Backend section)

## Dependencies
- Requires: [002] Core CLI
- Requires: [004] Task Commands

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestFileBackendAddTask` - `todoat --backend=file tasks.txt add "New task"` creates task in file
- [ ] `TestFileBackendGetTasks` - `todoat --backend=file tasks.txt` lists tasks from file
- [ ] `TestFileBackendUpdateTask` - `todoat --backend=file tasks.txt update "task" -s D` updates task
- [ ] `TestFileBackendDeleteTask` - `todoat --backend=file tasks.txt delete "task"` removes task
- [ ] `TestFileBackendListManagement` - Sections in file treated as task lists
- [ ] `TestFileBackendCreateFile` - Creates task file if not exists with proper header
- [ ] `TestFileBackendMetadataSupport` - Tasks store priority, dates, status, tags
- [ ] `TestFileBackendHierarchy` - Indented tasks parsed as subtasks
- [ ] `TestFileBackendConfigPath` - Configurable file path via config

## Implementation Notes
- Replace placeholder implementation in `backend/file/backend.go`
- Implement `TaskManager` interface fully
- Support plain text format (simpler than markdown)
- Format: One task per line with metadata encoding
- Sections via headers for list separation
- Support hierarchical tasks via indentation
- No Git dependency (unlike Git backend)

## Out of Scope
- Auto-commit (use Git backend for that)
- Multiple file support
- Binary file formats
- Cloud sync (use other backends)
