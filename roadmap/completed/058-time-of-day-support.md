# [058] Time of Day Support for Task Dates

## Summary
Extend task date handling to support time of day components (hours, minutes), enabling more precise due dates and start dates with full datetime support.

## Documentation Reference
- Primary: `dev-doc/TASK_MANAGEMENT.md` (Task Dates section)
- Related: `dev-doc/CLI_INTERFACE.md`

## Dependencies
- Requires: [011] Task Dates (basic date support must exist)
- Requires: [044] Relative Date Input (for time extensions)

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestAddTaskWithTime` - `todoat -y MyList add "Meeting" --due-date "2026-01-20T14:30"` sets datetime
- [ ] `TestAddTaskWithTimeZone` - `todoat -y MyList add "Call" --due-date "2026-01-20T14:30-05:00"` handles timezone
- [ ] `TestDateOnlyStillWorks` - `todoat -y MyList add "Task" --due-date "2026-01-20"` works (date only, midnight assumed)
- [ ] `TestTimeDisplayInList` - Tasks with time show time component in output
- [ ] `TestTimeInJSON` - `todoat -y --json MyList` includes full ISO8601 datetime
- [ ] `TestRelativeDateWithTime` - `todoat -y MyList add "Task" --due-date "tomorrow 14:00"` works
- [ ] `TestTimeUpdate` - `todoat -y MyList update "Task" --due-date "2026-01-20T15:00"` updates time

### Functional Requirements
- [ ] Due date accepts ISO8601 datetime format: `YYYY-MM-DDTHH:MM` or `YYYY-MM-DDTHH:MM:SS`
- [ ] Start date accepts same datetime format
- [ ] Timezone support: `YYYY-MM-DDTHH:MM±HH:MM` or `Z` for UTC
- [ ] Date-only input defaults to midnight local time (00:00)
- [ ] Time display format in list: `Jan 20 14:30` or configurable
- [ ] Internal storage in RFC3339 format (existing)
- [ ] Relative time input: `tomorrow 14:00`, `+2d 09:00`, `monday 10:30`

### Backend Compatibility
- [ ] SQLite: Stores RFC3339 datetime strings (no change needed)
- [ ] Nextcloud/CalDAV: Uses DATE-TIME property with TZID
- [ ] Todoist: Uses due.datetime API field
- [ ] Git: Stores in ISO8601 format in markdown metadata

### Output Requirements
- [ ] Time component only shown if not midnight (00:00)
- [ ] Consistent formatting across backends
- [ ] JSON output always includes full datetime if time set

## Implementation Notes

### Date Parsing Priority
1. Try full ISO8601 with time: `2026-01-20T14:30:00`
2. Try date with time: `2026-01-20 14:30`
3. Try relative with time: `tomorrow 14:30`
4. Fall back to date-only: `2026-01-20`

### Display Format
- With time: `Jan 20 14:30`
- Date only: `Jan 20`
- Custom format via view configuration

### CalDAV DTSTART/DUE Format
```
DTSTART;TZID=America/New_York:20260120T143000
DUE;TZID=America/New_York:20260120T170000
```

## Out of Scope
- All-day vs. timed event distinction
- Timezone conversion commands
- Time-based filtering (before/after specific times)
- Recurring times
