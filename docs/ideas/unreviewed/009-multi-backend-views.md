# [009] Multi-Backend Aggregate Views

## Summary
Create views that aggregate tasks from multiple backends into a single unified view, enabling cross-backend workflows.

## Source
Code analysis: The app supports multiple backends (sqlite, todoist, nextcloud, google, mstodo, git, file) but each backend is accessed independently. Users with tasks spread across backends have no unified view.

## Motivation
Power users often have tasks in multiple systems:
- Personal tasks in local SQLite
- Work tasks in Todoist (company account)
- Shared family tasks in Nextcloud

Currently they must check each backend separately. An aggregate view would show "all high priority tasks due this week" regardless of which backend holds them.

## Current Behavior
```bash
# Must check each backend separately
todoat -b sqlite Work
todoat -b todoist Work
todoat -b nextcloud Family

# No way to see all tasks together
```

## Proposed Behavior
```bash
# Define an aggregate view in config
# views/all-urgent.yaml:
#   aggregate:
#     backends: [sqlite, todoist, nextcloud]
#   filters:
#     - field: priority
#       operator: lte
#       value: 3
#     - field: status
#       operator: eq
#       value: TODO

# Use aggregate view
todoat --view all-urgent
# Shows urgent tasks from ALL configured backends

# Quick aggregate (no config needed)
todoat --aggregate -s TODO -p high
# Shows high priority TODO tasks from all enabled backends

# Tasks show their source backend
# Output:
# [sqlite]    Buy groceries                  1 (High)   Work
# [todoist]   Review quarterly report        2 (High)   Projects
# [nextcloud] Schedule family dinner         1 (High)   Family
```

## Estimated Value
medium - High value for multi-backend users, unique differentiator for terminal task tools

## Estimated Effort
L - Requires parallel backend queries, result merging, conflict handling for same-named lists across backends

## Open Questions
- How to handle list name collisions across backends?
- Performance with many backends (parallel queries)?
- Should operations (complete, update) work on aggregate results?
- Cache aggregate results or always query fresh?
- How to display backend source (prefix, column, color)?

## Related
- Backend system: docs/explanation/backend-system.md
- Views system: docs/how-to/views.md
- Multiple backend config already supported

## Status
unreviewed
