# [059] Recurring Tasks Support

## Summary
Implement recurring/repeating task functionality allowing tasks to automatically regenerate on a schedule (daily, weekly, monthly, custom intervals) when completed.

## Documentation Reference
- Primary: `docs/explanation/task-management.md`
- Related: `docs/explanation/synchronization.md`, `docs/explanation/backend-system.md`

## Dependencies
- Requires: [011] Task Dates
- Requires: [058] Time of Day Support (optional but recommended)

## Complexity
L

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestAddRecurringDaily` - `todoat -y MyList add "Standup" --recur daily` creates recurring task
- [ ] `TestAddRecurringWeekly` - `todoat -y MyList add "Review" --recur weekly` creates weekly task
- [ ] `TestAddRecurringMonthly` - `todoat -y MyList add "Report" --recur monthly` creates monthly task
- [ ] `TestAddRecurringCustom` - `todoat -y MyList add "Check" --recur "every 3 days"` works
- [ ] `TestCompleteRecurringTask` - Completing recurring task creates new instance with updated due date
- [ ] `TestRecurringFromDueDate` - New instance due date based on original due date, not completion date
- [ ] `TestRecurringFromCompletion` - `--recur-from-completion` bases next date on when completed
- [ ] `TestRecurringInJSON` - JSON output includes `recurrence` field with rule
- [ ] `TestRemoveRecurrence` - `todoat -y MyList update "Task" --recur none` removes recurrence
- [ ] `TestRecurringTaskDisplay` - Recurring tasks show indicator (🔄 or [R]) in list

### Functional Requirements
- [ ] Recurrence rule stored as RRULE-compatible string (RFC 5545)
- [ ] Supported frequencies: daily, weekly, monthly, yearly
- [ ] Custom intervals: `every N days/weeks/months`
- [ ] Weekday support: `every monday`, `weekdays`, `weekends`
- [ ] Day-of-month support: `monthly on 15th`, `monthly on last`
- [ ] Completing recurring task:
  - Mark current instance as DONE
  - Create new task with next due date
  - Copy all metadata (priority, tags, description)
  - Preserve parent relationship if subtask
- [ ] Two recurrence modes:
  - From due date (default): Next = due_date + interval
  - From completion: Next = completed_date + interval

### Backend Support
- [ ] SQLite: New `recurrence` column storing RRULE string
- [ ] Nextcloud/CalDAV: Use RRULE property in VTODO (native support)
- [ ] Todoist: Use recurring due string (native support)
- [ ] Git: Store as `@recur: daily` in markdown metadata

### Data Model
```go
type Task struct {
    // ... existing fields
    Recurrence     string    // RRULE string: "FREQ=WEEKLY;INTERVAL=1"
    RecurFromDue   bool      // true = from due date, false = from completion
}
```

### Output Requirements
- [ ] Recurring indicator in task list (configurable via views)
- [ ] JSON includes: `recurrence`, `recur_from_due` fields
- [ ] Next occurrence info available: `--show-next-occurrence` flag

## Implementation Notes

### RRULE Examples
- Daily: `FREQ=DAILY;INTERVAL=1`
- Weekly: `FREQ=WEEKLY;INTERVAL=1`
- Every 2 weeks: `FREQ=WEEKLY;INTERVAL=2`
- Monthly on 15th: `FREQ=MONTHLY;BYMONTHDAY=15`
- Weekdays: `FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR`

### Completion Flow
1. User completes recurring task
2. System marks current task DONE with completed timestamp
3. System calculates next due date from RRULE
4. System creates new task instance with:
   - New UID
   - Status: TODO
   - Due date: calculated next date
   - Copied: summary, description, priority, tags, parent_uid
   - Cleared: completed timestamp
5. Return both tasks in response (completed + new)

### CalDAV Sync Considerations
- Nextcloud handles RRULE natively
- Sync must not duplicate recurring tasks
- Use RECURRENCE-ID for exception handling (future)

## Out of Scope
- Exception dates (skip specific occurrences)
- Count-based recurrence (`repeat 5 times`)
- End date for recurrence series
- Bulk completion of recurring instances
- Recurring task history view
