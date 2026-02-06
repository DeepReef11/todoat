# [024] Personal Productivity Statistics

## Summary
Add a `todoat stats` command that shows personal productivity metrics: tasks completed over time, completion rates, common tags, busiest days, and streak tracking.

## Source
Code analysis: The app has an analytics system (`internal/analytics/`) that tracks command usage for internal purposes, but provides no user-facing productivity statistics. The existing `todoat list stats` only shows per-list counts. Users who want to understand their productivity patterns have no built-in way to do so.

## Motivation
Understanding personal productivity patterns helps users:
- See progress over time (motivational)
- Identify which days/times they're most productive
- Recognize which lists or tags consume most effort
- Track completion streaks for habit building

This turns todoat from a passive task tracker into an active productivity tool.

## Current Behavior
```bash
todoat list stats Work
# Output: Shows task counts for one list only
# Total tasks: 45
# Done: 30
# In Progress: 10
# Todo: 5

# No historical trends, no cross-list analysis, no streaks
```

## Proposed Behavior
```bash
# Overall productivity stats
todoat stats
# Output:
# Productivity Summary (Last 30 days)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Tasks completed: 47
# Completion rate: 78%
# Current streak:  5 days
# Longest streak:  12 days
#
# By List:
#   Work     │████████████░░░░│ 28 completed
#   Personal │████░░░░░░░░░░░░│ 12 completed
#   Errands  │██░░░░░░░░░░░░░░│  7 completed
#
# Top Tags: #meeting (15), #urgent (12), #review (8)
# Busiest Day: Tuesday (avg 3.2 tasks)

# Stats for specific period
todoat stats --period week
todoat stats --period year

# Completion trend
todoat stats --trend
# Shows week-over-week completion numbers

# JSON output
todoat stats --json
```

## Estimated Value
medium - Provides motivational feedback and productivity insights; differentiates from basic task lists

## Estimated Effort
M - Requires querying historical task data (completed timestamps), aggregation logic, trend calculations

## Related
- Analytics system: `internal/analytics/` (existing internal analytics)
- SQLite backend: `backend/sqlite/` (has completed timestamps)
- Existing list stats: `cmd/todoat/cmd/todoat.go` (basic per-list counts)

## Status
unreviewed
