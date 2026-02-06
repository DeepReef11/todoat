# User Experience Design Decisions

This document records user experience design decisions for the todoat project.

## Decisions

### 2026-01-31 Per-Interval Reminder Dismissal

**Decision**: Reminder dismissal is per-interval — each interval is tracked independently, providing more granular control.

**User Story**: As a user with multiple reminder intervals (e.g., `[1d, 1h, "at due time"]`), I expect that dismissing the 1-day reminder still allows the 1-hour reminder to fire later, because each reminder interval serves a different purpose.

**Context**: Reminders support multiple intervals (e.g., `[1d, 1h, "at due time"]`). When a user dismisses a reminder, the behavior for subsequent intervals needed clarification: does dismissal suppress only the current interval or all intervals until the next due cycle?

**Alternatives Considered**:
- Global dismissal until next cycle: A single dismiss suppresses all intervals. Simpler mental model, but users lose the benefit of having multiple intervals.
- Snooze-style with duration: "Remind me in 30 minutes" regardless of configured intervals. More complex to implement and changes the interval-based paradigm.

**Consequences**:
- Each reminder interval fires independently and can be dismissed independently
- Users get more granular control over their notification experience
- Dismissal state must be tracked per-interval, not just per-task
- The mental model is slightly more complex but matches user expectations for multi-interval reminders

**Related**: [UX-008] - See `docs/decisions/question-log.md` for full discussion
