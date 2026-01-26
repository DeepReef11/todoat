# [004] Time Tracking Integration

## Summary
Add basic time tracking capabilities to record time spent on tasks, with optional integration with external time tracking services.

## Source
Code analysis: No time tracking exists. Users who track time (for billing, productivity analysis, or personal metrics) must use separate tools and manually correlate with tasks.

## Motivation
Many professionals need to track time spent on tasks for billing clients, understanding productivity patterns, or meeting organizational requirements. Basic time tracking within todoat would keep this data close to the task context.

## Current Behavior
No time tracking support. Users must:
1. Start task in todoat
2. Start timer in separate app (Toggl, Clockify, etc.)
3. Stop timer manually
4. Manually associate time entries with tasks

## Proposed Behavior
```bash
# Start timer on a task
todoat Work start "Review PR #123"
# Output: Started tracking time on "Review PR #123" at 14:30

# Stop timer
todoat Work stop
# Output: Stopped tracking. Logged 45m on "Review PR #123"

# View time entries
todoat Work time "Review PR"
# Output:
# Task: Review PR #123
# Total time: 2h 15m
# Sessions:
#   2026-01-25 14:30-15:15 (45m)
#   2026-01-26 09:00-10:30 (1h 30m)

# Quick log without real-time tracking
todoat Work log "Review PR" --duration 30m

# Time report
todoat time report --since "last monday"
# Output: time by list, task, tag
```

## Estimated Value
medium - High value for freelancers/consultants, useful for productivity-minded users

## Estimated Effort
L - Requires new database tables, timer state management, reporting features, optional external integrations

## Open Questions
- Store time data in main SQLite or separate database?
- Support Pomodoro mode (25min work, 5min break)?
- Export formats (CSV, Toggl format)?
- Integration with Toggl/Clockify APIs (push time entries)?
- Handle timer running when computer sleeps/restarts?

## Related
- SQLite backend: backend/sqlite/sqlite.go
- Analytics system: internal/analytics/

## Status
unreviewed
