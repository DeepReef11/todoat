# 005: Status System

Implement the task status handling including status transitions, display formatting, and filtering by status.

## Dependencies

- `003-sqlite-backend.md` - Task struct with status field
- `004-task-commands.md` - Commands that modify status

## Acceptance Criteria

- [ ] **Status values** defined:
  - `TODO` - task not started (default for new tasks)
  - `DONE` - task completed

- [ ] **Status display** formatting:
  - Tasks show status indicator in list view
  - Format: `[TODO]` or `[DONE]`
  - Consider colored output (optional for MVP)

- [ ] **Status flag** on commands:
  - `-s, --status` flag accepts: `TODO`, `DONE` (or abbreviations `T`, `D`)
  - `todoat <list> -s TODO` filters to show only TODO tasks
  - `todoat <list> update "Task" -s DONE` sets status
  - Status values are case-insensitive

- [ ] **Status transitions**:
  - Any status can transition to any other status
  - `complete` command is shorthand for `-s DONE`
  - When transitioning to DONE, set `completed` timestamp
  - When transitioning from DONE to TODO, clear `completed` timestamp

- [ ] **Status abbreviations** work:
  - `T` = `TODO`
  - `D` = `DONE`

- [ ] **Filter by status**:
  - `todoat <list> -s TODO` shows only TODO tasks
  - `todoat <list> -s DONE` shows only completed tasks
  - Default (no filter): show all tasks

- [ ] Unit tests for status transitions and filtering

## Complexity

**Estimate:** S (Small)

## Implementation Notes

- Reference: `docs/explanation/README.md` for status mappings
- Reference: `docs/explanation/task-management.md#task-status-system`
- For MVP, only implement TODO and DONE (skip PROCESSING, CANCELLED)
- Status values stored in database as strings
- Consider a StatusType enum/const for type safety
- Status translation function: `ParseStatus(string) Status`

### Status Constants

```go
const (
    StatusTodo = "TODO"
    StatusDone = "DONE"
)

var statusAbbreviations = map[string]string{
    "T": StatusTodo,
    "D": StatusDone,
}
```

### Display Format

```
Tasks in "Work" (3 tasks):

[TODO] Review PR #123
[TODO] Write documentation
[DONE] Deploy staging
```
