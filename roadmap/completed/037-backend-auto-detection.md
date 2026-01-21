# [037] Backend Auto-Detection

## Summary
Implement the backend auto-detection system that automatically selects the appropriate backend based on the current directory context and configuration, with a `--detect-backend` flag to show detection results.

## Documentation Reference
- Primary: `docs/explanation/backend-system.md` (Backend Selection, Auto-Detection Interface sections)
- Related: `docs/explanation/configuration.md` (auto_detect_backend setting)

## Dependencies
- Requires: [002] Core CLI (for --detect-backend flag)
- Requires: [003] SQLite Backend
- Requires: [023] Git/Markdown Backend (for git auto-detection)

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestDetectBackendFlag` - `todoat --detect-backend` shows auto-detected backend(s)
- [ ] `TestDetectBackendGit` - In git repo with TODO.md marker, auto-detects git backend
- [ ] `TestDetectBackendPriority` - Multiple detectable backends shows priority-ordered list
- [ ] `TestAutoDetectEnabled` - With `auto_detect_backend: true`, uses detected backend automatically

### Functional Requirements
- [ ] `--detect-backend` global flag shows detection results and exits
- [ ] DetectableBackend interface:
  - `CanDetect() (bool, error)` - Returns true if backend is usable in current context
  - `DetectionInfo() string` - Human-readable detection info
- [ ] Backend registry methods:
  - `RegisterDetectable(name, constructor)` - Register detectable backend
  - `GetDetectableConstructors()` - Get all registered detectable backends
- [ ] Auto-detection algorithm:
  1. If `--backend` flag provided, use it (skip detection)
  2. If `sync.enabled: true`, use SQLite cache
  3. If `auto_detect_backend: true`, run detection:
     - Iterate through registered detectable backends
     - Call `CanDetect()` on each
     - Use first backend returning (true, nil)
  4. Fall back to `default_backend` config
  5. Fall back to first enabled backend in `backend_priority`
  6. Fall back to first enabled backend
- [ ] Git backend detection:
  - Walk directory tree from CWD to find `.git`
  - Search for markdown file with `<!-- todoat:enabled -->` marker
  - File search order: configured → fallback → defaults (TODO.md, todo.md, .todoat.md)
- [ ] Detection constraints:
  - Fast execution (<100ms)
  - Non-destructive (read-only operations)

### Output Requirements
- [ ] `--detect-backend` output format:
  ```
  Auto-detected backends:
    1. git: /path/to/repo/TODO.md
    2. sqlite: ~/.local/share/todoat/tasks.db (always available)

  Would use: git
  ```
- [ ] Show why detection failed if no backends detected

## Implementation Notes
- Each backend self-registers in `init()` function
- Detection results can be cached for session duration
- SQLite is always "detectable" as fallback (no external dependencies)

## Out of Scope
- Automatic backend configuration creation
- Network-based detection (Nextcloud, Todoist require explicit config)
