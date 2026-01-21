# 011: Task Dates

## Summary
Implement due date and start date support for tasks, including CLI flags for setting dates and display formatting.

## Documentation Reference
- Primary: `docs/explanation/task-management.md`
- Sections: Task Dates, Add Tasks, Update Tasks

## Dependencies
- Requires: 003-sqlite-backend.md (Task struct, database schema)
- Requires: 004-task-commands.md (add/update commands)
- Blocked by: none

## Complexity
**M (Medium)** - Database schema changes, date parsing, display formatting, multiple command updates

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestAddTaskWithDueDate` - `todoat -y MyList add "Task" --due-date 2026-01-31` sets due date
- [ ] `TestAddTaskWithStartDate` - `todoat -y MyList add "Task" --start-date 2026-01-15` sets start date
- [ ] `TestAddTaskWithBothDates` - `todoat -y MyList add "Task" --due-date 2026-01-31 --start-date 2026-01-15` sets both
- [ ] `TestUpdateTaskDueDate` - `todoat -y MyList update "Task" --due-date 2026-02-15` updates due date
- [ ] `TestClearTaskDueDate` - `todoat -y MyList update "Task" --due-date ""` clears due date
- [ ] `TestInvalidDateFormat` - `todoat -y MyList add "Task" --due-date "invalid"` returns ERROR
- [ ] `TestDateFormatValidation` - `todoat -y MyList add "Task" --due-date "01-31-2026"` returns ERROR (wrong format)
- [ ] `TestTaskDatesInJSON` - `todoat -y --json MyList` includes due_date and start_date fields
- [ ] `TestCompletedTimestamp` - `todoat -y MyList complete "Task"` sets completed timestamp automatically

### Unit Tests (if needed)
- [ ] Date parsing handles YYYY-MM-DD format correctly
- [ ] Date storage uses RFC3339 format in database
- [ ] Empty string clears date field (sets to NULL)

### Manual Verification
- [ ] Dates display in task list when present
- [ ] Completed timestamp auto-set when marking task DONE
- [ ] Completed timestamp cleared when changing status from DONE to TODO

## Implementation Notes

### CLI Flags
```bash
todoat MyList add "Task" --due-date 2026-01-31
todoat MyList add "Task" --start-date 2026-01-15
todoat MyList update "Task" --due-date 2026-02-15
todoat MyList update "Task" --due-date ""  # Clear date
```

### Date Format
- User input: `YYYY-MM-DD` (e.g., `2026-01-31`)
- Storage: RFC3339 (e.g., `2026-01-31T00:00:00Z`)
- Display: `YYYY-MM-DD` or localized format

### Database Schema Update
```sql
-- Add columns if not exist
ALTER TABLE tasks ADD COLUMN due_date TEXT;
ALTER TABLE tasks ADD COLUMN start_date TEXT;
-- completed column should already exist from 003-sqlite-backend
```

### Task Struct Update
```go
type Task struct {
    // ... existing fields
    DueDate   time.Time  // User-settable
    StartDate time.Time  // User-settable
    Completed time.Time  // Auto-set when status → DONE
}
```

### Required Changes
1. Add `--due-date` and `--start-date` flags to add command
2. Add `--due-date` and `--start-date` flags to update command
3. Update Task struct with date fields
4. Update SQLite schema and CRUD operations
5. Implement date parsing and validation
6. Update display formatting to show dates
7. Auto-set Completed timestamp in complete command

### Date Validation
```go
func parseDate(input string) (time.Time, error) {
    if input == "" {
        return time.Time{}, nil  // Clear date
    }
    return time.Parse("2006-01-02", input)
}
```

## Out of Scope
- Date-based filtering (`--due-before`, `--due-after`) - separate item
- Relative date input ("tomorrow", "+1week") - separate item
- Time of day support (just dates for now) - separate item
- Recurring tasks - separate item
- Overdue highlighting - separate item (views/customization)
