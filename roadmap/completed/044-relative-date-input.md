# [044] Relative Date Input

## Summary
Implement relative date parsing for CLI date flags, supporting human-friendly input like "today", "tomorrow", "+7d", "+1w", "+1m".

## Documentation Reference
- Primary: `docs/explanation/views-customization.md`
- Section: Date Filter Special Values
- Related: `docs/explanation/task-management.md`

## Dependencies
- Requires: [011] Task Dates
- Requires: [043] Date Filtering

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestRelativeDateToday` - `todoat -y MyList add "Task" --due-date today` sets due date to current date
- [ ] `TestRelativeDateTomorrow` - `todoat -y MyList add "Task" --due-date tomorrow` sets due date to next day
- [ ] `TestRelativeDateYesterday` - `todoat -y MyList --due-after yesterday` filters from yesterday
- [ ] `TestRelativeDateDaysAhead` - `todoat -y MyList add "Task" --due-date +7d` sets due date 7 days from now
- [ ] `TestRelativeDateDaysBack` - `todoat -y MyList --due-after -3d` filters from 3 days ago
- [ ] `TestRelativeDateWeeks` - `todoat -y MyList add "Task" --due-date +2w` sets due date 2 weeks from now
- [ ] `TestRelativeDateMonths` - `todoat -y MyList add "Task" --due-date +1m` sets due date 1 month from now
- [ ] `TestAbsoluteDateStillWorks` - `todoat -y MyList add "Task" --due-date 2026-01-31` still works

### Supported Formats
| Input | Meaning |
|-------|---------|
| `today` | Current date (00:00:00) |
| `tomorrow` | Next day |
| `yesterday` | Previous day |
| `+Nd` | N days from now |
| `-Nd` | N days ago |
| `+Nw` | N weeks from now |
| `+Nm` | N months from now |

### Functional Requirements
- Relative dates work with `--due-date`, `--start-date`, `--due-before`, `--due-after`, `--created-before`, `--created-after`
- Absolute dates (YYYY-MM-DD) continue to work
- Invalid relative format returns clear error message

## Implementation Notes
- Extend date parsing function in internal/utils or internal/cli
- Check for relative format before trying absolute parse
- Calculate relative dates from current time

## Out of Scope
- Time-of-day in relative dates
- Natural language parsing ("next Monday", "end of month")
