# Terminal User Interface (TUI)

todoat includes an interactive terminal user interface for managing tasks with keyboard navigation.

## Launching the TUI

```bash
todoat tui
```

This opens a full-screen interface with two panes: lists on the left and tasks on the right.

## Interface Layout

```
┌─ Lists ─────────┬─ Tasks ────────────────────────┐
│ ▸ Work          │   [TODO] Complete report [P1]  │
│   Personal      │   [DONE] Send email            │
│   Shopping      │   [IN-PROGRESS] Review PR      │
│                 │                                │
└─────────────────┴────────────────────────────────┘
                  [ Status bar ]
```

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `Tab` | Switch focus between lists and tasks panes |
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `q` / `Ctrl+C` | Quit the TUI |

### Task Operations

| Key | Action |
|-----|--------|
| `a` | Add a new task to the current list |
| `e` | Edit the selected task's summary |
| `c` | Toggle task completion (mark done/undone) |
| `d` | Delete the selected task (with confirmation) |

### Filtering

| Key | Action |
|-----|--------|
| `/` | Enter filter mode to search tasks |
| `Esc` | Clear filter and return to normal mode |

### Help

| Key | Action |
|-----|--------|
| `?` | Show help dialog |
| `Esc` / `Enter` | Close help dialog |

## Modes

### Normal Mode

The default mode for navigating and viewing tasks. Use navigation keys to move around and action keys to modify tasks.

### Add Mode

Activated with `a`. Type the task name and press:
- `Enter` to create the task
- `Esc` to cancel

### Edit Mode

Activated with `e`. Modify the task summary and press:
- `Enter` to save changes
- `Esc` to cancel

### Filter Mode

Activated with `/`. Type a search term and press:
- `Enter` to apply the filter
- `Esc` to clear the filter and exit

### Confirm Delete Mode

Shown when pressing `d` on a task. Press:
- `y` to confirm deletion
- `n` or `Esc` to cancel

## Features

- **Two-pane layout**: Lists on the left, tasks on the right
- **Real-time updates**: Changes are immediately saved to the backend
- **Keyboard-driven**: Full functionality without mouse
- **Vim-style navigation**: Use `j`/`k` for up/down movement
- **Task filtering**: Quickly find tasks by typing a search term
- **Subtask display**: Subtasks are shown with proper indentation
- **Status indicators**: Visual distinction for completed and in-progress tasks

## Backend Support

The TUI works with all configured backends (SQLite, Nextcloud, Todoist, etc.). It uses the currently active backend as configured in your config file.

## Tips

- Use `Tab` to quickly switch between selecting lists and managing tasks
- Filter with `/` to narrow down long task lists
- Toggle completion with `c` for quick task management
- The TUI remembers your position when switching between lists

---
*Last updated: 2026-01-19*
