# [040] Bulk Hierarchy Operations

## Summary
Implement wildcard pattern support for operating on task hierarchies, enabling bulk completion, updates, and deletion of subtasks using `*` (direct children) and `**` (all descendants) patterns.

## Documentation Reference
- Primary: `dev-doc/SUBTASKS_HIERARCHY.md` (Bulk Operations section)
- Related: `dev-doc/TASK_MANAGEMENT.md`

## Dependencies
- Requires: [014] Subtasks Hierarchy (parent-child relationships)
- Requires: [004] Task Commands (update, complete, delete operations)

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestBulkCompleteDirectChildren` - `todoat MyList complete "Parent/*"` completes all direct children of Parent
- [ ] `TestBulkCompleteAllDescendants` - `todoat MyList complete "Parent/**"` completes all descendants recursively
- [ ] `TestBulkUpdatePriority` - `todoat MyList update "Parent/**" --priority 1` updates priority on all descendants
- [ ] `TestBulkDeleteChildren` - `todoat MyList delete "Parent/*"` deletes direct children only
- [ ] `TestBulkNoMatchError` - `todoat MyList complete "NonExistent/*"` returns ERROR with message
- [ ] `TestBulkEmptyMatch` - `todoat MyList complete "LeafTask/*"` returns INFO_ONLY (no children)
- [ ] `TestBulkCountOutput` - Bulk operation returns count of affected tasks

### Functional Requirements
- [ ] `*` pattern matches only direct children of specified parent
- [ ] `**` pattern matches all descendants (children, grandchildren, etc.)
- [ ] Pattern applied after resolving parent task by summary
- [ ] Single transaction for entire bulk operation
- [ ] Confirmation prompt for destructive operations (delete) with task count
- [ ] No-prompt mode (`-y`) skips confirmation
- [ ] JSON output includes affected task count and list of UIDs

### Output Requirements
- [ ] Text mode: `Completed 5 tasks under "Release v2.0"` with result code
- [ ] JSON mode:
  ```json
  {
    "result": "ACTION_COMPLETED",
    "action": "complete",
    "affected_count": 5,
    "parent": "Release v2.0",
    "pattern": "**"
  }
  ```

## Implementation Notes
- Extend task search logic to detect `/*` and `/**` suffix patterns
- After parent resolution, use `GetChildTasks(parentUID, recursive bool)`
- Apply operation to each matched child within single transaction
- Depth-first ordering for delete (children before grandchildren for FK safety)

## Out of Scope
- Glob patterns beyond `*` and `**`
- Cross-list bulk operations
- Complex filter expressions in wildcards
