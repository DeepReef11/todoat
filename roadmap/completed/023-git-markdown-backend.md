# [023] Git/Markdown Backend

## Summary
Implement the Git backend to store tasks in human-readable markdown files within Git repositories, with auto-detection and optional auto-commit functionality.

## Documentation Reference
- Primary: `docs/explanation/backend-system.md` (Git Backend section)

## Dependencies
- Requires: [002] Core CLI
- Requires: [004] Task Commands

## Complexity
L

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestGitBackendDetection` - `todoat` auto-detects git repo with marked TODO.md file
- [ ] `TestGitBackendAddTask` - `todoat "Project Tasks" add "New task"` adds task to markdown file
- [ ] `TestGitBackendGetTasks` - `todoat "Project Tasks"` lists tasks from markdown file
- [ ] `TestGitBackendUpdateTask` - `todoat "Project Tasks" update "task" -s D` updates task in file
- [ ] `TestGitBackendDeleteTask` - `todoat "Project Tasks" delete "task"` removes task from file
- [ ] `TestGitBackendListManagement` - Sections in markdown treated as task lists
- [ ] `TestGitBackendHierarchy` - Indented tasks parsed as subtasks
- [ ] `TestGitBackendMarkerRequired` - `<!-- todoat:enabled -->` marker required
- [ ] `TestGitBackendAutoCommit` - Changes auto-committed when `auto_commit: true`
- [ ] `TestGitBackendFallbackFiles` - Search order: configured file, fallbacks, defaults

## Implementation Notes
- Create `backend/git/` package
- Implement `DetectableBackend` interface for auto-detection
- Walk up directory tree to find `.git` directory
- Search for markdown file with `<!-- todoat:enabled -->` marker
- Parse markdown sections as task lists (## headings)
- Parse list items as tasks with metadata support
- Support indentation for hierarchical subtasks
- Optional auto-commit with descriptive messages
- File caching to avoid unnecessary re-parsing

## Out of Scope
- Automatic push to remote (user handles git push)
- Branch management
- Merge conflict resolution
- Multiple markdown files per repo
