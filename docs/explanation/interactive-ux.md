# Interactive UX (Prompt Mode)

## Overview

When `--no-prompt` (`-y`) is **not** set, todoat operates in interactive mode. This mode provides richer terminal output, confirmation prompts, smart task disambiguation, and guided input for task creation. The goal is to make CLI interactions feel intuitive and prevent destructive mistakes.

Interactive mode is the default. It can be disabled globally via config (`ui.no_prompt: true`) or per-command via the `-y` flag.

## Prompt Library

todoat uses [promptui](https://github.com/manifoldco/promptui) for interactive prompts. This library provides:

- **Fuzzy-find selection**: Type to filter when selecting from lists
- **Styled prompts**: Colored labels, icons, and formatted output
- **Validation**: Inline input validation with error messages
- **Templates**: Customizable display for list items

Current implementation in `internal/utils/inputs.go` uses basic `bufio.Scanner`. This should be migrated to promptui for all interactive selection and input prompts.

## Smart Task Disambiguation

### Context-Aware Filtering

When an action targets a task by summary and multiple tasks share that summary, todoat filters candidates based on whether the action makes sense for each task's current status before deciding whether to prompt.

**Status groups:**

| Group | Statuses | Description |
|-------|----------|-------------|
| Actionable | `NEEDS-ACTION`, `IN-PROGRESS` | Tasks that can still be worked on |
| Terminal | `COMPLETED`, `CANCELLED` | Tasks that are done |

**Default filtering rules by action:**

| Action | Default Candidates | With `--all` | Rationale |
|--------|-------------------|--------------|-----------|
| `complete` (`c`) | Actionable only | All statuses | Already-done tasks are excluded by default |
| `update` (`u`) with `-s COMPLETED` | Actionable only | All statuses | Same as complete |
| `update` (`u`) with `-s CANCELLED` | Actionable only | All statuses | Same as complete |
| `update` (`u`) with other flags | Actionable only | All statuses | Updating a done task is unusual |
| `delete` (`d`) | Actionable only | All statuses | Typically deleting active tasks; use `--all` to include terminal |
| `get` (`g`) | All statuses | All statuses | Viewing applies to any task |

> **Override**: Use `--all` flag or set `ui.interactive_prompt_for_all_tasks: true` in config to include `COMPLETED`/`CANCELLED` tasks in all interactive prompts. See [Showing All Tasks](#showing-all-tasks-in-prompts).

### Disambiguation Flow

```
User runs: todoat mylist c "mytask"

1. Find all tasks with summary matching "mytask"
2. Apply context-aware filter for the action
3. If 0 matches after filter → error: no matching task
4. If 1 match after filter → proceed without prompt
5. If multiple matches after filter → prompt selection
```

**Example: Complete action with 3 "mytask" entries**

```
Tasks in list:
  - mytask (NEEDS-ACTION)
  - mytask (COMPLETED)
  - mytask (CANCELLED)

Action: complete
Filter: actionable only → 1 candidate (NEEDS-ACTION)
Result: proceeds without prompt
```

**Example: Complete action with 2 actionable "mytask" entries**

```
Tasks in list:
  - mytask (NEEDS-ACTION)
  - mytask (IN-PROGRESS)
  - mytask (CANCELLED)

Action: complete
Filter: actionable only → 2 candidates
Result: prompt selection with rich details
```

### Rich Selection Display

When disambiguation requires a prompt, display enough context to differentiate tasks. Use promptui's fuzzy-find select with custom templates.

**Display fields per candidate:**

| Field | Shown | Example |
|-------|-------|---------|
| Summary | Always | `mytask` |
| Status | Always | `TODO`, `IN-PROGRESS` |
| Priority | If set (non-zero) | `Priority: 5` |
| Due date | If set | `Due: 2026-02-15` |
| Parent | If subtask | `Parent: project-alpha` |
| Description | If set (truncated) | `Desc: Fix the login bug in...` |
| Categories/Tags | If set | `Tags: backend, urgent` |
| Created date | Always | `Created: 2026-01-20` |

**Prompt format example:**

```
Multiple tasks match "mytask". Select one:
  ❯ mytask  [TODO]  Priority:5  Due:2026-02-15  Parent:project-alpha  Tags:backend
    mytask  [IN-PROGRESS]  Created:2026-01-28  Desc:Fix the login bug in...
    mytask  [TODO]  Due:2026-03-01  Tags:frontend,urgent
  (type to filter, ↑/↓ to navigate, enter to select, q to cancel)
```

## Action-Specific Interactive Behavior

### Delete Action (`d`)

Delete always requires confirmation in interactive mode to prevent accidental data loss.

**Single match — confirm:**
```
$ todoat mylist d "mytask"
Delete task "mytask" (NEEDS-ACTION, Due: 2026-02-15)? (y/n): _
```

**Multiple matches — select then confirm (actionable only by default):**
```
$ todoat mylist d "mytask"
Multiple tasks match "mytask". Select one to delete:
  ❯ mytask  [TODO]  Due:2026-02-15  Priority:5
    mytask  [IN-PROGRESS]  Created:2026-01-28

Delete task "mytask" (TODO, Due: 2026-02-15)? (y/n): _
```

With `--all`, terminal tasks are also shown:
```
$ todoat mylist d --all "mytask"
Multiple tasks match "mytask". Select one to delete:
  ❯ mytask  [TODO]  Due:2026-02-15  Priority:5
    mytask  [IN-PROGRESS]  Created:2026-01-28
    mytask  [COMPLETED]  Completed:2026-01-25
    mytask  [CANCELLED]  Created:2026-01-10

Delete task "mytask" (TODO, Due: 2026-02-15)? (y/n): _
```

**No task specified — see [No Task Specified](#no-task-specified--prompt-from-list).**

### Add Action (`a`) — Interactive Task Creation

When no summary is provided via arguments, or when invoked with just the list name, todoat enters interactive add mode. Fields are prompted sequentially. Pressing Enter without input skips optional fields (uses default). Summary is mandatory.

**Field prompts (in order):**

| # | Field | Required | Default | Format/Notes |
|---|-------|----------|---------|--------------|
| 1 | Summary | Yes | — | Free text. Cannot be empty. Use `--literal` behavior (no `/` hierarchy parsing) |
| 2 | Description | No | empty | Free text, single line |
| 3 | Priority | No | 0 (none) | Integer 0-9 (0=none, 1=highest, 9=lowest) |
| 4 | Due date | No | none | `YYYY-MM-DD`, or relative: `today`, `tomorrow`, `+7d`, `+2w`, `+1m` |
| 5 | Start date | No | none | Same format as due date |
| 6 | Tags | No | none | Comma-separated: `backend, urgent` |
| 7 | Parent task | No | none | Fuzzy-find select from existing tasks in list, or empty for top-level |
| 8 | Recurrence | No | none | RRULE format: `FREQ=WEEKLY;INTERVAL=1` |

**Interactive add flow:**

```
$ todoat mylist a
Summary (required): Buy groceries
Description (Enter to skip): Get items for the week
Priority 0-9 (Enter to skip): 3
Due date (Enter to skip, format: YYYY-MM-DD or +Nd): tomorrow
Start date (Enter to skip):
Tags (Enter to skip, comma-separated): errands, personal
Parent task (Enter to skip):
  ❯ (none - top level)
    House chores
    Weekly planning
Recurrence (Enter to skip, RRULE format):

Created task: Buy groceries (ID: a1b2c3d4-...)
```

**With summary provided — skip summary prompt:**
```
$ todoat mylist a "Buy groceries"
Description (Enter to skip): _
...
```

**Validation during input:**
- Summary: reject empty input, re-prompt
- Priority: reject non-integer or out of range 0-9, re-prompt
- Dates: validate format, show error and re-prompt on invalid input
  ```
  Due date (Enter to skip, format: YYYY-MM-DD or +Nd): abc
  ✗ Invalid date format. Use YYYY-MM-DD, today, tomorrow, +Nd, +Nw, +Nm
  Due date (Enter to skip, format: YYYY-MM-DD or +Nd): _
  ```
- Date range: if start date is after due date, warn and re-prompt start date
- Tags: accept any comma-separated text
- Recurrence: validate RRULE format if provided

### Update Action (`u`)

When a task is identified (with disambiguation if needed), update uses flags as today. No additional interactive prompts beyond disambiguation.

```
$ todoat mylist u "mytask" -p 5 --due-date tomorrow
```

If multiple actionable matches exist, disambiguate first, then apply the update.

### Complete Action (`c`)

Same disambiguation as update. No confirmation prompt needed — completing a task is not destructive.

### Get/View Action (`g`)

No disambiguation filtering by status — all tasks are valid targets for viewing. If multiple matches, prompt selection from all matches.

### No Task Specified — Prompt from List

When an action requires a task but none is specified, todoat prompts with fuzzy-find selection from all tasks in the list, filtered to actionable tasks by default.

**Complete without target — select which task to complete:**
```
$ todoat mylist c
Select a task to complete:
  ❯ buy groceries     [TODO]        Due:2026-02-15
    write report      [IN-PROGRESS] Priority:3
  (type to filter, ↑/↓ to navigate, enter to select, q to cancel)
```

Only `NEEDS-ACTION` and `IN-PROGRESS` tasks are shown. `COMPLETED`/`CANCELLED` tasks are hidden.

**Update with status change — select which task to cancel:**
```
$ todoat mylist u -s CANCELLED
Select a task to cancel:
  ❯ buy groceries     [TODO]        Due:2026-02-15
    write report      [IN-PROGRESS] Priority:3
  (type to filter)
```

**Delete without target — select which task to delete:**
```
$ todoat mylist d
Select a task to delete:
  ❯ buy groceries     [TODO]        Due:2026-02-15
    write report      [IN-PROGRESS] Priority:3
  (type to filter)

Delete task "buy groceries" (TODO, Due: 2026-02-15)? (y/n): _
```

By default, only actionable tasks shown. Use `--all` to include all statuses:

```
$ todoat mylist d --all
Select a task to delete:
  ❯ buy groceries     [TODO]        Due:2026-02-15
    write report      [IN-PROGRESS] Priority:3
    old task          [COMPLETED]   Completed:2026-01-25
    abandoned task    [CANCELLED]   Created:2026-01-10
  (type to filter)
```

**Get/view without target — shows all tasks (no filtering):**
```
$ todoat mylist g
Select a task to view:
  ❯ buy groceries     [TODO]        Due:2026-02-15
    write report      [IN-PROGRESS] Priority:3
    old task          [COMPLETED]   Completed:2026-01-25
  (type to filter)
```

## Showing All Tasks in Prompts

By default, interactive prompts hide `COMPLETED` and `CANCELLED` tasks for most actions. This keeps prompts focused on what the user is likely working with.

To include all tasks regardless of status:

**Per-command:** Use the `--all` flag:
```bash
todoat mylist d --all          # Delete prompt shows all tasks
todoat mylist c --all          # Complete prompt shows all tasks (including already done)
todoat mylist u --all -p 5     # Update prompt shows all tasks
```

**Globally:** Set in config:
```yaml
ui:
  interactive_prompt_for_all_tasks: true   # default: false
```

When enabled, all interactive prompts show every task regardless of status. The `--all` flag overrides the config on a per-command basis.

| Setting | `--all` flag | Result |
|---------|-------------|--------|
| `false` (default) | Not set | Actionable tasks only |
| `false` (default) | Set | All tasks |
| `true` | Not set | All tasks |
| `true` | N/A | All tasks (flag is redundant) |

## Decorative Output

In interactive mode, todoat uses richer terminal formatting compared to `--no-prompt` mode.

### Output Differences

| Element | Interactive Mode | No-Prompt Mode |
|---------|-----------------|----------------|
| Task created | `Created task: Buy groceries (ID: a1b2...)` | `Created task: Buy groceries (ID: a1b2...)\nACTION_COMPLETED` |
| Task completed | `Completed: "Buy groceries"` | `ACTION_COMPLETED` |
| Task deleted | `Deleted: "Buy groceries"` | `ACTION_COMPLETED` |
| Errors | Colored error with suggestion | Plain error + `ERROR` result code |
| Lists | Bordered box with colors | Plain text table |
| Multiple matches | Fuzzy-find selector | Match table + `ACTION_INCOMPLETE` |

### Color Usage

Colors enhance readability in interactive mode. They are automatically disabled when output is piped or redirected.

| Element | Color |
|---------|-------|
| Task summary | Bold white |
| Status TODO/NEEDS-ACTION | Yellow |
| Status IN-PROGRESS | Cyan |
| Status COMPLETED | Green |
| Status CANCELLED | Gray/dim |
| Priority (high, 1-3) | Red |
| Priority (medium, 4-6) | Yellow |
| Priority (low, 7-9) | Gray |
| Due date (overdue) | Red |
| Due date (today) | Yellow |
| Due date (future) | Default |
| Prompt labels | Cyan |
| Error messages | Red |
| Success messages | Green |

## Relationship with `--no-prompt` Mode

| Behavior | Interactive (default) | `--no-prompt` (`-y`) |
|----------|----------------------|----------------------|
| Delete confirmation | Prompts y/n | Deletes immediately |
| Multiple matches | Fuzzy-find selection (actionable only) | Returns match list + `ACTION_INCOMPLETE` |
| No task specified | Fuzzy-find from list (actionable only) | Error: task required |
| Add without summary | Prompts for fields | Error: summary required |
| No list specified | List picker | Returns available lists + `INFO_ONLY` |
| Decorative output | Colors, borders, icons | Plain text, result codes |
| Single match | Proceeds silently | Proceeds + `ACTION_COMPLETED` |
| `--all` flag | Shows all statuses in prompts | Shows all statuses in output |

## Implementation Notes

### promptui Integration

```go
import "github.com/manifoldco/promptui"

// Fuzzy-find task selection
prompt := promptui.Select{
    Label: "Select a task",
    Items: candidates,
    Templates: &promptui.SelectTemplates{
        Label:    "{{ . }}",
        Active:   "❯ {{ .Summary | bold }}  [{{ .Status }}]  {{ if .DueDate }}Due:{{ .DueDate }}{{ end }}",
        Inactive: "  {{ .Summary }}  [{{ .Status }}]  {{ if .DueDate }}Due:{{ .DueDate }}{{ end }}",
        Selected: "✔ {{ .Summary }}",
    },
    Searcher: func(input string, index int) bool {
        // Fuzzy match against summary, tags, description
    },
}

// Input with validation
prompt := promptui.Prompt{
    Label:    "Due date (Enter to skip, format: YYYY-MM-DD or +Nd)",
    Validate: validateDateInput,
    Default:  "",
}
```

### Current Code Locations

| Component | Current Location | Notes |
|-----------|-----------------|-------|
| Input utilities | `internal/utils/inputs.go` | Replace with promptui |
| Prompt stub | `internal/cli/prompt/prompt.go` | Empty, intended for this |
| Task struct | `backend/interface.go:18-34` | Fields for display |
| Date parsing | `internal/utils/validation.go` | Reuse for input validation |
| No-prompt check | `cfg.NoPrompt` flag | Gate interactive behavior |

### Testing

- All promptui interactions should have a bypass path using `io.Reader`/`io.Writer` injection
- Tests default to `no_prompt: true` for deterministic behavior
- Interactive flows should be testable via the existing `WithReader` pattern

## Related

- [CLI Interface](cli-interface.md) — Command structure, no-prompt mode, result codes
- [CLI Interface: No-Prompt Mode](cli-interface.md#no-prompt-mode) — Non-interactive behavior details
- [Task Management](task-management.md) — Task operations and search
- [Views Customization](views-customization.md) — Output formatting
