# [043] Date-Based Task Filtering

## Summary
Implement date filtering flags for the get command to filter tasks by due date and creation date ranges.

## Documentation Reference
- Primary: `dev-doc/CLI_INTERFACE.md`
- Section: Filter by Dates
- Related: `dev-doc/BACKEND_SYSTEM.md` (TaskFilter struct with DueAfter/DueBefore)

## Dependencies
- Requires: [011] Task Dates

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestFilterDueBefore` - `todoat -y MyList --due-before 2026-02-01` shows only tasks due before date
- [ ] `TestFilterDueAfter` - `todoat -y MyList --due-after 2026-01-15` shows only tasks due after date
- [ ] `TestFilterDueRange` - `todoat -y MyList --due-after 2026-01-15 --due-before 2026-02-01` shows tasks in range
- [ ] `TestFilterCreatedAfter` - `todoat -y MyList --created-after 2026-01-01` shows tasks created after date
- [ ] `TestFilterCreatedBefore` - `todoat -y MyList --created-before 2026-01-15` shows tasks created before date
- [ ] `TestFilterNoDueDate` - Tasks without due dates excluded from due date filters
- [ ] `TestFilterCombined` - `todoat -y MyList -s TODO --due-before 2026-02-01` combines status and date filters

### Functional Requirements
- Date filters use inclusive range (includes boundary dates)
- Tasks without dates not matched by date filters
- Multiple filters combine with AND logic
- Date format: YYYY-MM-DD

## Implementation Notes
- Add flags to get command: `--due-before`, `--due-after`, `--created-before`, `--created-after`
- Use TaskFilter struct from backend interface
- Parse dates using same logic as 011-task-dates
- Apply filters before displaying results

## Out of Scope
- Relative date input ("today", "+7d") - separate item [044]
- Modified date filtering
- Time-of-day filtering
