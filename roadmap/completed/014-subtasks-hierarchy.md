# [014] Subtasks and Hierarchical Task Support

## Summary
Implement parent-child task relationships with path-based creation, tree visualization using box-drawing characters, and comprehensive subtask operations including re-parenting and cascade deletion.

## Documentation Reference
- Primary: `docs/explanation/subtasks-hierarchy.md`
- Related: `docs/explanation/task-management.md`, `docs/explanation/views-customization.md`

## Dependencies
- Requires: [004] Task Commands
- Requires: [003] SQLite Backend (for parent_uid foreign key)

## Complexity
L

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestAddSubtaskWithParentFlag` - `todoat MyList add "Child" -P "Parent"` creates subtask under existing parent
- [ ] `TestPathBasedHierarchyCreation` - `todoat MyList add "A/B/C"` creates 3-level hierarchy with auto-parent creation
- [ ] `TestTreeVisualization` - `todoat MyList` displays tasks with box-drawing characters (├─, └─, │)
- [ ] `TestUpdateParent` - `todoat MyList update "Task" -P "NewParent"` re-parents task
- [ ] `TestRemoveParent` - `todoat MyList update "Task" --no-parent` moves subtask to root level
- [ ] `TestCascadeDelete` - `todoat MyList delete "Parent"` prompts and deletes all descendants
- [ ] `TestLiteralSlashFlag` - `todoat MyList add -l "UI/UX Design"` creates single task with slash in summary
- [ ] `TestPathResolutionExisting` - Adding `A/B/C` when `A/B` exists only creates `C` under existing `B`
- [ ] `TestCircularReferenceBlocked` - Cannot set task as parent of its own ancestor
- [ ] `TestOrphanDetection` - System handles tasks whose parent was deleted externally

## Implementation Notes
- Add `parent_uid` column to tasks table with foreign key constraint and ON DELETE CASCADE
- Path parsing should split on `/` unless `--literal` flag set
- Tree building uses two-pass algorithm: first create node map, then link relationships
- Box-drawing characters: `├─` for branch, `└─` for last child, `│` for vertical continuation
- Hierarchical sorting ensures parents displayed before children regardless of sort order

## Out of Scope
- Cross-list parent-child relationships (parent and child must be in same list)
- Automatic parent status updates when all children complete
- Bulk operations on hierarchies (e.g., `complete "Parent/*"`) - separate roadmap item
