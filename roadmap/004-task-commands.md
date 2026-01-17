# 004: Task Commands

Implement the core task operation commands: add, get (list), update, complete, and delete.

## Dependencies

- `002-core-cli.md` - CLI framework must be in place
- `003-sqlite-backend.md` - Backend must support CRUD operations

## Acceptance Criteria

- [ ] **`add` command** implemented:
  - `todoat <list> add "Task summary"` creates a new task
  - `todoat <list> a "Task summary"` works (abbreviation)
  - Supports `-p <priority>` flag (0-9)
  - New task defaults to status `TODO`
  - Returns success message with task UID

- [ ] **`get` command** implemented:
  - `todoat <list>` displays all tasks (default action)
  - `todoat <list> get` explicit form
  - `todoat <list> g` abbreviation
  - Displays task list with status indicators and summaries
  - Shows priority if non-zero

- [ ] **`update` command** implemented:
  - `todoat <list> update "Task summary" --summary "New name"` renames task
  - `todoat <list> u "Task" -p 2` updates priority
  - `todoat <list> update "Task" -s DONE` updates status
  - Task matching by summary (exact or partial match)
  - Handles case where multiple tasks match (in no-prompt mode: error or list matches)

- [ ] **`complete` command** implemented:
  - `todoat <list> complete "Task summary"` marks task as DONE
  - `todoat <list> c "Task"` abbreviation
  - Sets completion timestamp
  - Task matching by summary

- [ ] **`delete` command** implemented:
  - `todoat <list> delete "Task summary"` removes task
  - `todoat <list> d "Task"` abbreviation
  - In no-prompt mode (`-y`): deletes without confirmation
  - In prompt mode: asks for confirmation
  - Task matching by summary

- [ ] **Task matching logic**:
  - Exact match first
  - Partial/substring match if no exact match
  - Error if no matches found
  - In no-prompt mode: error if multiple matches (with list of matches)

- [ ] **Action abbreviations** work: `g`, `a`, `u`, `c`, `d`

- [ ] CLI tests for each command using injectable IO pattern

## Complexity

**Estimate:** L (Large)

## Implementation Notes

- Reference: `dev-doc/CLI_INTERFACE.md` for command syntax
- Reference: `dev-doc/TASK_MANAGEMENT.md` for operation details
- Action resolution: check if arg[1] is a known action (get, add, update, complete, delete or abbreviation)
- For MVP, assume single list or require list name - interactive list selection deferred
- Task matching should be case-insensitive for better UX
- Consider extracting task search logic into a reusable function
- The `-y` flag should skip confirmation prompts and use first/only match

### Command Flow

```
todoat <list> <action> [task-summary] [flags]

Examples:
  todoat Work add "Review PR" -p 1
  todoat Work                        # lists tasks (default action: get)
  todoat Work complete "Review PR"
  todoat -y Work delete "Old task"   # no confirmation
```
