# [010] Location-Based Task Context

## Summary
Add optional location/context metadata to tasks, enabling filtering by where tasks can be done (home, office, errands, computer, phone).

## Source
GTD methodology analysis: "Contexts" are a core GTD concept - filtering tasks by what resources/location you need. The app has tags which could serve this purpose, but dedicated context support would provide better UX and built-in common contexts.

## Motivation
Not all tasks can be done anywhere. When at the grocery store, show "errands" tasks. When at the computer, show "computer" tasks. Context-aware filtering helps users focus on what's actually actionable in their current situation.

## Current Behavior
```bash
# Users can use tags as contexts
todoat Work add "Buy milk" --tag errands
todoat Work --tag errands

# But no first-class context support, easy to forget conventions
```

## Proposed Behavior
```bash
# Built-in contexts (configurable)
# config.yaml:
#   contexts:
#     - home
#     - office
#     - errands
#     - computer
#     - phone
#     - anywhere

# Add task with context
todoat Work add "Buy milk" --context errands
todoat Work add "Review PR" --context computer
todoat Work add "Call dentist" --context phone

# Filter by context
todoat --context errands
# Shows all tasks from all lists with @errands context

# Multiple contexts (can be done in either)
todoat Work add "Email response" --context computer,phone

# Default context per list
# config.yaml:
#   lists:
#     Groceries:
#       default_context: errands

# Context-based view
todoat view create at-computer --filter-context computer --filter-status TODO

# Quick context switch in TUI
# Press 'c' to cycle contexts or filter by context
```

## Estimated Value
low - Useful for GTD practitioners, but tags already provide similar functionality

## Estimated Effort
S - Contexts are essentially special-cased tags, minimal new infrastructure

## Open Questions
- Contexts vs. tags - worth the distinction?
- Should contexts be enforced (must be from config list) or freeform?
- GPS/location-based automatic context switching (future, complex)?
- CalDAV CATEGORIES vs custom field for context storage?

## Related
- Tags system: docs/how-to/tags.md
- GTD methodology contexts
- Idea #006 (Inbox Workflow) - contexts complement inbox processing

## Status
unreviewed
