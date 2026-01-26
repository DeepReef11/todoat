# [014] Focus Mode (Single Task View)

## Summary
Add a focus mode that displays only the current working task, hides distractions, and optionally integrates with system Do Not Disturb settings.

## Source
Productivity methodology analysis: Deep work requires focus on one task. The TUI and CLI show full task lists which can be distracting. A focus mode would help users concentrate on the single most important task.

## Motivation
Task lists are useful for planning but counterproductive during execution - seeing 20 other tasks creates anxiety and temptation to switch. Focus mode would:
1. Display only the current task prominently
2. Provide quick "done" and "next" actions
3. Optionally suppress notifications
4. Support time-boxing (Pomodoro-style)

## Current Behavior
```bash
# See full task list always
todoat Work

# TUI shows all tasks in the list
todoat tui
```

## Proposed Behavior
```bash
# Enter focus mode for highest priority task
todoat focus
# Output:
# ╔════════════════════════════════════════╗
# ║  FOCUS: Ship feature v2.0              ║
# ║  Priority: 1 (High) | Due: Today       ║
# ║                                        ║
# ║  Press [d] done | [s] skip | [q] quit  ║
# ╚════════════════════════════════════════╝

# Focus on specific task
todoat focus "Review PR #123"

# Focus mode in TUI
# Press 'f' to enter focus mode for selected task

# Focus with timer (Pomodoro)
todoat focus --timer 25m
# Counts down, notifies when time's up

# Focus with DND integration
todoat focus --dnd
# Enables system Do Not Disturb during focus

# Auto-select next task after completion
todoat focus --auto-next
# Moves to next highest priority task after completing current

# Exit focus mode
# Press 'q' or complete all focused tasks
```

## Estimated Value
low - Niche feature for productivity enthusiasts, might be over-engineering

## Estimated Effort
M - New TUI mode, timer implementation, OS DND integration varies by platform

## Open Questions
- Worth adding if users can just filter to one task?
- Timer/Pomodoro scope creep (see Idea #004 Time Tracking)?
- DND integration platform complexity worth it?
- Focus history for analytics?
- Integration with time tracking (auto-start timer on focus)?

## Related
- TUI: docs/how-to/tui.md
- Idea #004 (Time Tracking) - timer overlap
- Pomodoro technique

## Status
unreviewed
