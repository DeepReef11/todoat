# [021] Task Effort Estimation

## Summary
Add an optional effort/size field to tasks (S/M/L/XL or time estimates) to help with workload planning and realistic scheduling.

## Source
Code analysis: The Task struct in `backend/interface.go` has priority but no effort/estimate field. Users doing capacity planning or sprint work have no way to express how big a task is. Idea 004 (Time Tracking) focuses on actual time spent; this focuses on predicted effort.

## Motivation
Not all tasks are equal in scope. A priority-1 task might take 5 minutes or 5 days. Without effort estimates, users can't realistically plan their day or week. Effort estimation enables:
- "What can I realistically finish today?"
- "Should I start this task before the meeting?"
- "Is this week overloaded?"

## Current Behavior
```bash
todoat Work add "Implement user auth" -p 1
todoat Work add "Fix typo in docs" -p 1
# Both show as high priority but one is a 5-minute fix, the other is multi-day work
# No way to distinguish scope
```

## Proposed Behavior
```bash
# Add task with effort estimate
todoat Work add "Implement user auth" --effort L
todoat Work add "Fix typo in docs" --effort S

# Update effort on existing task
todoat Work update "user auth" --effort XL

# Filter by effort
todoat Work --filter-effort S,M    # Show only small/medium tasks

# Views can include effort column
todoat view create planning --fields "summary,effort,priority,due_date"

# Task list output includes effort
$ todoat Work
Work
  [ ] Implement user auth  [XL] p1
  [ ] Fix typo in docs     [S]  p1

# Supported values: S, M, L, XL (or time estimates: 15m, 1h, 4h, 1d)
```

## Estimated Value
medium - Enables realistic workload planning; essential for professional task management

## Estimated Effort
M - Requires schema change (new field), CLI flags, filter logic, view integration

## Related
- Task model: `backend/interface.go` (Task struct needs new field)
- Idea 004: Time tracking integration (tracks actual time; effort tracks estimated time)
- Idea 012: Smart scheduling (could use effort estimates for suggestions)
- Views system: `internal/views/` (effort as a display field)

## Status
unreviewed
