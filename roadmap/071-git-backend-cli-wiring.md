# [071] Git Backend CLI Wiring

## Summary
Complete the CLI wiring for the Git/Markdown backend, enabling users to explicitly access it via `--backend=git` flag in addition to auto-detection.

## Documentation Reference
- Primary: `docs/explanation/backend-system.md`
- Section: Git Backend - noted as "implemented but not yet wired to CLI"

## Dependencies
- Requires: [023] Git/Markdown Backend (backend implementation)
- Requires: [037] Backend Auto-Detection

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] `TestGitBackendExplicitFlag` - `todoat -b git "Project Tasks"` works with explicit flag
- [ ] `TestGitBackendConfigType` - Backend type "git" recognized in config `backends:` section
- [ ] `TestGitBackendListsCommand` - `todoat -b git list` shows sections from TODO.md

### Functional Requirements
- [ ] `--backend=git` flag explicitly selects Git backend
- [ ] Backend type "git" is valid in config file
- [ ] Git backend appears in `todoat backends` list when configured
- [ ] Auto-detection continues to work alongside explicit selection

## Implementation Notes
- Verify Git backend is registered for explicit type selection (not just auto-detection)
- Add "git" to backend type validation if missing
- Ensure `--backend=git` bypasses auto-detection and uses Git directly
- Update documentation to show explicit usage alongside auto-detection

## Out of Scope
- Auto-detection changes (already done in 037)
- Markdown parser changes
- Auto-commit feature changes
