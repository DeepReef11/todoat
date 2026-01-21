# [035] Interactive View Builder

## Summary
Implement TUI-based interactive view builder that allows users to create custom views without manually editing YAML files, with real-time preview and validation.

## Documentation Reference
- Primary: `docs/explanation/views-customization.md` (Interactive View Builder section)
- Related: `docs/explanation/cli-interface.md`

## Dependencies
- Requires: [015] Views Customization (view system must exist)
- Requires: [029] TUI Interface (for TUI components/library)

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestViewCreateCommand` - `todoat view create myview` launches interactive builder
- [ ] `TestViewBuilderSavesYAML` - Completing builder creates valid YAML in `~/.config/todoat/views/`
- [ ] `TestViewBuilderCancel` - Pressing Escape/Ctrl+C exits without saving

### Functional Requirements
- [ ] `todoat view create <name>` command launches TUI builder
- [ ] Field selection panel with checkboxes for all available fields:
  - status, summary, description, priority, due_date, start_date
  - created, modified, completed, tags, uid, parent
- [ ] Field configuration per selected field:
  - Width (integer)
  - Alignment (left/center/right)
  - Format (for dates)
  - Plugin (optional formatter script path)
- [ ] Filter builder panel:
  - Add filter rules with field, operator, value
  - Supported operators: eq, ne, lt, lte, gt, gte, contains, in, not_in, regex
  - Date special values: today, tomorrow, yesterday, +Nd, -Nd, +Nw, +Nm
- [ ] Sort rule builder:
  - Add sort rules with field and direction (asc/desc)
  - Priority ordering for multi-level sort
- [ ] Real-time validation with error highlighting
- [ ] Preview panel showing sample task output (optional)
- [ ] Keyboard navigation:
  - Arrow keys: navigate
  - Tab/Shift+Tab: next/previous panel
  - Space: toggle checkbox
  - Enter: confirm/open
  - Esc: cancel/close
  - Ctrl+S: quick save
  - Ctrl+C: cancel and exit

### Output Requirements
- [ ] Creates valid YAML file at `~/.config/todoat/views/<name>.yaml`
- [ ] At least one field must be selected to save
- [ ] View immediately available via `todoat MyList -v <name>`

## Implementation Notes
- Uses same TUI library as [029] TUI Interface (likely bubbletea or tview)
- Validation happens on every change, errors shown at bottom
- Invalid views cannot be saved until errors resolved
- Sample data for preview can be hardcoded test tasks

## Out of Scope
- Editing existing views via TUI (use text editor for now)
- Importing views from other users via TUI
