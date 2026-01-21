# [029] TUI Interface

## Summary
Implement a terminal user interface (TUI) for interactive task management, providing a visual interface within the terminal for browsing, editing, and organizing tasks.

## Documentation Reference
- Primary: `docs/explanation/features-overview.md` (Planned features - TUI/GUI interface)
- Related: `docs/explanation/cli-interface.md`, `docs/explanation/views-customization.md`

## Dependencies
- Requires: [002] Core CLI
- Requires: [004] Task Commands
- Requires: [007] List Commands
- Requires: [014] Subtasks Hierarchy
- Requires: [015] Views Customization

## Complexity
L

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestTUILaunch` - `todoat tui` launches the terminal interface
- [ ] `TestTUIListNavigation` - Arrow keys navigate between task lists
- [ ] `TestTUITaskNavigation` - Arrow keys navigate between tasks in list
- [ ] `TestTUIAddTask` - Press 'a' to add new task via input dialog
- [ ] `TestTUIEditTask` - Press 'e' to edit selected task
- [ ] `TestTUICompleteTask` - Press 'c' to toggle task completion
- [ ] `TestTUIDeleteTask` - Press 'd' with confirmation to delete task
- [ ] `TestTUITreeView` - Subtasks displayed in collapsible tree structure
- [ ] `TestTUIFilterTasks` - '/' opens filter/search dialog
- [ ] `TestTUIKeyBindings` - Help panel shows all available key bindings ('?')
- [ ] `TestTUIQuit` - 'q' exits the TUI gracefully

## Implementation Notes
- Create `cmd/tui.go` command entry point
- Use `bubbletea` library for TUI framework
- Use `lipgloss` for styling
- Implement components:
  - List selector (left pane)
  - Task list (main pane)
  - Task detail (right pane or popup)
  - Status bar (bottom)
- Support vim-like keybindings (j/k for navigation)
- Integrate with existing views system for formatting
- Real-time updates when backend changes

## Out of Scope
- GUI interface (separate roadmap item)
- Mouse support (keyboard-only initially)
- Multi-window/split views
- Custom color themes (use terminal colors)
