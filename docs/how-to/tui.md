# Terminal User Interface (TUI)

todoat includes an interactive terminal interface for managing tasks with keyboard navigation.

## Launching the TUI

```bash
todoat tui
```

The TUI displays a two-pane interface:
- **Left pane**: Task lists
- **Right pane**: Tasks in the selected list

## Navigation

### Switching Focus

| Key | Action |
|-----|--------|
| `Tab` | Switch focus between lists and tasks panes |

### Moving Within a Pane

| Key | Action |
|-----|--------|
| `j` or `↓` | Move down |
| `k` or `↑` | Move up |

When in the lists pane, selecting a different list automatically loads its tasks.

## Task Operations

### Add a Task

| Key | Action |
|-----|--------|
| `a` | Enter add mode |

1. Press `a` to start adding a task
2. Type the task name
3. Press `Enter` to create the task
4. Press `Esc` to cancel

### Edit a Task

| Key | Action |
|-----|--------|
| `e` | Enter edit mode for selected task |

1. Select a task in the tasks pane
2. Press `e` to edit
3. Modify the task name
4. Press `Enter` to save
5. Press `Esc` to cancel

### Complete/Uncomplete a Task

| Key | Action |
|-----|--------|
| `c` | Toggle task completion status |

Press `c` on a selected task to mark it as done. Press again to mark it as incomplete.

### Delete a Task

| Key | Action |
|-----|--------|
| `d` | Delete selected task (with confirmation) |

1. Select a task
2. Press `d`
3. Confirm deletion when prompted (press `y` to confirm, `n` or `Esc` to cancel)

## Filtering Tasks

| Key | Action |
|-----|--------|
| `/` | Enter filter mode |

1. Press `/` to start filtering
2. Type search text
3. Press `Enter` to apply filter
4. Tasks matching the filter are shown
5. Press `Esc` to clear filter and exit filter mode

## Help and Exit

| Key | Action |
|-----|--------|
| `?` | Show help |
| `q` | Quit TUI |
| `Ctrl+C` | Quit TUI |

## Keyboard Reference

| Key | Mode | Action |
|-----|------|--------|
| `Tab` | Normal | Switch panes |
| `j` / `↓` | Normal | Move down |
| `k` / `↑` | Normal | Move up |
| `a` | Normal | Add task |
| `e` | Normal | Edit task |
| `c` | Normal | Complete/uncomplete task |
| `d` | Normal | Delete task |
| `/` | Normal | Filter tasks |
| `?` | Normal | Show help |
| `q` | Normal | Quit |
| `Enter` | Input | Confirm input |
| `Esc` | Input | Cancel input |

## Visual Indicators

- **Selected item**: Highlighted with bold text
- **Completed tasks**: Shown with strikethrough
- **Subtasks**: Indented under parent tasks
- **Status bar**: Shows current mode and active filter

## Backend Selection

The TUI uses your default backend. To use a specific backend:

```bash
todoat -b nextcloud tui
todoat -b sqlite tui
```

## Examples

### Daily Task Review

```bash
# Launch TUI
todoat tui

# Navigate to Work list (j/k to move, Tab to switch panes)
# Mark tasks complete (c)
# Add new tasks (a)
# Quit when done (q)
```

### Quick Filter

```bash
todoat tui

# Press / and type "meeting" to show only meeting-related tasks
# Press Esc to clear filter
```

## See Also

- [Task Management](task-management.md) - CLI task operations
- [Views](views.md) - Customizing task display
