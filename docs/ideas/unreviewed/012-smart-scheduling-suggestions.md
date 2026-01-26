# [012] Smart Scheduling Suggestions

## Summary
Provide intelligent suggestions for when to schedule tasks based on workload, patterns, and due dates.

## Source
Analytics capability analysis: The app collects usage analytics (commands, timing) but doesn't use this data to provide scheduling assistance. Users must manually decide when to work on tasks.

## Motivation
Users often have tasks without due dates or with distant deadlines. Deciding when to work on them requires mental effort. Smart suggestions based on:
- Current workload per day
- Historical completion patterns
- Task priority and dependencies
- Buffer time before deadlines

...could help users plan more effectively without manual time management.

## Current Behavior
```bash
# Add task without scheduling
todoat Work add "Research competitors"

# Manual due date assignment
todoat Work update "Research" --due-date 2026-02-01

# No assistance in choosing optimal dates
```

## Proposed Behavior
```bash
# Get scheduling suggestion for a task
todoat Work suggest "Research competitors"
# Output:
# Suggested schedule for "Research competitors":
#   - Start: Thursday, Jan 30 (light workload: 2 tasks)
#   - Due: Friday, Feb 7 (allows 3 business days)
#   - Reason: Your Thursdays average 3 completions, Feb 3-5 is busy
#
# Apply suggestion? [Y/n/edit]

# Suggest with constraints
todoat Work suggest "Research competitors" --before "Feb 15" --effort large
# Factors in effort estimate for scheduling

# Bulk suggestions for unscheduled tasks
todoat Work suggest --all --unscheduled
# Shows suggestions for all tasks without due dates

# Auto-apply suggestions (config-driven)
# config.yaml:
#   scheduling:
#     auto_suggest: true
#     suggest_on_add: true
#     workload_balance: true
#     max_tasks_per_day: 5

# View workload distribution
todoat schedule workload
# Output:
# Week of Jan 27:
#   Mon: ████████ 8 tasks (heavy)
#   Tue: ███████ 7 tasks
#   Wed: ████ 4 tasks (light)
#   Thu: ██ 2 tasks (very light)
#   Fri: ██████ 6 tasks
```

## Estimated Value
low - Nice-to-have, may be over-engineering for CLI tool, complex ML/heuristics needed

## Estimated Effort
L - Requires workload analysis, pattern learning from analytics, suggestion algorithm

## Open Questions
- How sophisticated should suggestions be (simple heuristics vs ML)?
- Privacy: comfortable using completion history for suggestions?
- What makes a "good" suggestion (minimize overload vs meet deadlines vs balance)?
- Should this integrate with calendar availability (ICS feed input)?
- Worth the complexity for a CLI tool?

## Related
- Analytics system: already collects timing data
- Idea #004 (Time Tracking) - effort estimates feed into scheduling
- Idea #003 (Calendar Integration) - calendar awareness

## Status
unreviewed
