# [011] Daily Review and Planning Report

## Summary
Generate a formatted daily review report showing yesterday's completions, today's agenda, overdue items, and upcoming deadlines.

## Source
Feature gap: The analytics system tracks command usage but doesn't provide daily planning reports. Users must manually query tasks to understand their day, missing the opportunity for a focused daily review workflow.

## Motivation
Effective task management includes regular reviews. A daily report command would:
1. Show what was accomplished yesterday (motivation, accountability)
2. Highlight what's due today (focus)
3. Flag overdue items (urgency awareness)
4. Preview upcoming week (planning)

This is a common feature in task apps (Todoist karma, Things Today view) that helps users start their day effectively.

## Current Behavior
```bash
# Must run multiple queries manually
todoat Work -s DONE --completed-after yesterday
todoat Work -s TODO --due-before today
todoat Work --due-after today --due-before "+7d"

# No consolidated view, no formatting for review
```

## Proposed Behavior
```bash
# Generate daily report
todoat report daily
# Output:
#
# === Daily Review: Monday, January 27, 2026 ===
#
# YESTERDAY'S COMPLETIONS (5 tasks)
#   [Work]     Review PR #123
#   [Work]     Update documentation
#   [Personal] Call dentist
#   ...
#
# TODAY'S AGENDA (3 tasks)
#   [!] [Work] Ship feature v2.0            Due: today, P1
#   [ ] [Work] Team standup notes           Due: today, P5
#   [ ] [Personal] Pick up groceries        Due: today
#
# OVERDUE (2 tasks)
#   [Work] Quarterly review                 Due: 3 days ago, P2
#   [Personal] Renew subscription           Due: 5 days ago
#
# UPCOMING THIS WEEK (4 tasks)
#   Wed: [Work] Client presentation
#   Fri: [Personal] Doctor appointment
#   ...
#
# Stats: 23 tasks active, 5 completed yesterday, 2 overdue

# Weekly report
todoat report weekly

# Export report (for journaling, sharing)
todoat report daily --format markdown > ~/journal/2026-01-27.md

# Scheduled report (via cron/daemon)
# config.yaml:
#   report:
#     daily:
#       enabled: true
#       time: "08:00"
#       notification: true
```

## Estimated Value
medium - High value for daily review habit, differentiator for terminal task tools

## Estimated Effort
S - Primarily query composition and output formatting, no new data structures

## Open Questions
- Include all backends or configurable backend list?
- Default report scope (all lists or configurable)?
- Integration with notification system for morning report?
- Export formats (plain text, markdown, JSON)?
- Weekly/monthly variations?

## Related
- Analytics system: docs/explanation/analytics.md
- Notification manager: docs/explanation/notification-manager.md
- Reminder system: docs/how-to/reminders.md

## Status
unreviewed
