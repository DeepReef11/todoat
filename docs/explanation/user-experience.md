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

### 2026-02-08 Interactive Prompt Documentation Update

**Decision**: Minimal update to `docs/explanation/interactive-ux.md` — fix "empty stub" references and add config option mention.

**Context**: The interactive prompt feature was implemented in commit `b6a6151` (2026-01-31). The code now includes a full fuzzy-find task selection prompt (`internal/cli/prompt/prompt.go`), the `ui.interactive_prompt_for_all_tasks` config option, and context-aware filtering by action type.

**Alternatives Considered**:
- Full rewrite of docs/explanation/interactive-ux.md: Would document the complete fuzzy-find prompt, config option, and context-aware filtering in detail. More comprehensive but higher effort for an already-functional feature.

**Consequences**:
- The explanation doc accurately reflects the implemented TaskSelector behavior
- "Empty stub" references removed
- Config option `ui.interactive_prompt_for_all_tasks` is mentioned
- User-facing documentation in `docs/how-to/task-management.md` describes interactive selection

**Related**: [UX-009] - See `docs/decisions/question-log.md` for full discussion
