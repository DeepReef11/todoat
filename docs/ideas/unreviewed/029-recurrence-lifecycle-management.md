# [029] Recurrence Lifecycle Management

## Summary
Add commands to manage recurring task lifecycle: skip occurrences, edit recurrence patterns, view upcoming instances, and set recurrence end conditions.

## Source
Code analysis: Recurrence exists (`Recurrence` field, `parseRecurrence()`, `calculateNextOccurrence()`) but management is limited. Users cannot skip a single occurrence, see upcoming instances, or modify the recurrence pattern without recreating the task. No end date or count limit for recurrences.

## Motivation
Recurring tasks need ongoing management beyond initial creation:
- Skip this week's occurrence (vacation, holiday)
- See next 5 instances to verify pattern is correct
- Change recurrence from weekly to bi-weekly
- End recurrence after a date or number of occurrences
- Convert completed instance back to one-off task

Current implementation creates new instances on completion but lacks lifecycle controls.

## Current Behavior
```bash
# Create recurring task
todoat Work add "Weekly standup" --recur weekly

# On completion, new instance is auto-created
todoat Work complete "Weekly standup"
# New task appears with next due date

# No way to:
# - Skip a specific occurrence
# - View upcoming occurrences
# - Modify recurrence pattern
# - Set end date/count
# - Pause/resume recurrence
```

## Proposed Behavior
```bash
# View upcoming occurrences
todoat Work recur show "Weekly standup" --next 5
# Output:
# Weekly standup (recurring weekly on Monday)
# Next 5 occurrences:
#   1. Mon 2026-02-10
#   2. Mon 2026-02-17
#   3. Mon 2026-02-24
#   4. Mon 2026-03-03
#   5. Mon 2026-03-10

# Skip specific occurrence
todoat Work recur skip "Weekly standup"
# Output: Skipped next occurrence (Mon 2026-02-10)
# Next occurrence is now Mon 2026-02-17

todoat Work recur skip "Weekly standup" --date 2026-02-24
# Output: Marked 2026-02-24 occurrence as skipped

# View skipped occurrences
todoat Work recur skipped "Weekly standup"
# Output: Skipped occurrences:
#   - Mon 2026-02-10 (skipped)
#   - Mon 2026-02-24 (skipped)

# Modify recurrence pattern
todoat Work recur edit "Weekly standup" --pattern "every 2 weeks"
todoat Work recur edit "Monthly review" --pattern "every month on the 1st"

# Set recurrence end date
todoat Work recur edit "Weekly standup" --until 2026-12-31
# Output: Recurrence will end after 2026-12-31

# Set occurrence limit
todoat Work recur edit "Weekly standup" --count 10
# Output: Recurrence will end after 10 occurrences (5 remaining)

# Pause/resume recurrence
todoat Work recur pause "Weekly standup"
# Output: Recurrence paused. No new instances will be created on completion.

todoat Work recur resume "Weekly standup"
# Output: Recurrence resumed.

# Convert recurring task to one-off
todoat Work recur stop "Weekly standup"
# Output: Converted to non-recurring task

# Show all recurring tasks
todoat Work --filter-recurring
# Or: todoat Work list --recurring-only
```

## Estimated Value
medium - Common need for anyone using recurring tasks; missing controls cause workarounds like recreating tasks

## Estimated Effort
M - Extends existing recurrence infrastructure; needs skip tracking, pattern parsing, RRULE modifications

## Related
- Recurrence implementation: `cmd/todoat/cmd/todoat.go` (parseRecurrence, calculateNextOccurrence)
- Task struct: `backend/backend.go` Recurrence field
- RRULE format: CalDAV standard for recurrence patterns
- Idea #005 (Dependencies) - recurring tasks may have dependency implications

## Status
unreviewed
