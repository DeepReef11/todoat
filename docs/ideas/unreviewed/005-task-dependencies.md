# [005] Task Dependencies

## Summary
Add support for explicit task dependencies ("blocked by" relationships), allowing users to define which tasks must be completed before others can start.

## Source
Code analysis: The app has subtask hierarchy (parent/child) but no dependency relationships. Subtasks imply containment, not sequencing. Project work often has tasks that depend on other tasks without being children.

## Motivation
Complex projects have tasks that can't start until prerequisites are complete. Dependencies help users understand what's truly actionable vs. blocked, enable dependency-aware views, and could inform priority ordering.

## Current Behavior
```bash
# Subtasks provide hierarchy, not dependencies
todoat Work add "Phase 2" -P "Phase 1"
# This makes "Phase 2" a child of "Phase 1"
# But doesn't express "Phase 2 blocked by Phase 1"
```

## Proposed Behavior
```bash
# Add dependency when creating task
todoat Work add "Deploy to prod" --depends-on "Run integration tests"

# Add dependency to existing task
todoat Work update "Deploy to prod" --add-dependency "Update documentation"

# Remove dependency
todoat Work update "Deploy to prod" --remove-dependency "Update docs"

# View dependencies
todoat Work deps "Deploy to prod"
# Output:
# Deploy to prod
#   Blocked by:
#     [ ] Run integration tests
#     [ ] Update documentation

# Show only actionable tasks (nothing blocking them)
todoat Work --filter-actionable
# or
todoat view create actionable --filter-no-dependencies

# Completing a blocking task updates dependent task status
todoat Work complete "Run integration tests"
# Output: Completed "Run integration tests"
# "Deploy to prod" is now unblocked (1 dependency remaining)
```

## Estimated Value
medium - Essential for project management workflows, enables dependency-aware filtering

## Estimated Effort
M - Database schema changes, new CLI flags, dependency resolution logic, view integration

## Open Questions
- How to handle circular dependencies?
- Should blocking tasks affect priority calculation?
- Visual representation in list view (show blocked indicator)?
- Sync dependencies with backends that support them (some CalDAV servers)?
- VTODO RELATED-TO property compliance?

## Related
- Subtask system: docs/explanation/subtasks-hierarchy.md
- Backend interface: backend/interface.go

## Status
unreviewed
