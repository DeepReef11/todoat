# [027] Docs: Orphan task detection feature not documented

## Type
documentation

## Severity
low

## Test Location
- File: backend/sqlite/cli_test.go
- Function: TestOrphanDetectionSQLiteCLI

## Feature Description
There's a test for orphan detection, suggesting the app can detect tasks whose parent has been deleted. This is a data integrity feature that users might want to know about.

## Expected Documentation
- Location: docs/how-to/task-management.md or docs/explanation/subtasks-hierarchy.md
- Section: Orphan Tasks or Data Integrity

Should cover:
- [ ] What happens when a parent task is deleted
- [ ] Whether orphans are automatically promoted to root level
- [ ] Any warnings or notifications about orphans
- [ ] How to manually detect/fix orphaned tasks

## Alternative
If orphan detection is automatic and transparent:
- [x] Document in subtasks-hierarchy.md under "What happens when parent is deleted"

## Resolution

**Fixed in**: this session
**Fix description**: Added documentation for orphan task handling in docs/explanation/subtasks-hierarchy.md under the "Edge Cases" section. The orphan detection behavior (automatic promotion to root level) is now documented.

### Implementation Details
The code at `internal/views/renderer.go:69` shows that when building the task tree, if a task has a `ParentID` but the parent is not found in the node map, it's treated as an orphan and added to root nodes. This ensures orphaned tasks remain visible.

### Verification Log
```bash
$ grep -A 6 "Edge Cases:" docs/explanation/subtasks-hierarchy.md
**Edge Cases:**
- Parent task not found: Error message, subtask creation aborted
- Parent task in different list: Error - parent and child must be in same list (this case should not happen unless maybe when using --uid)
- Parent task deleted: Child tasks also get deleted (prompt user y/n)
- Circular references: Prevented by database constraints and validation
- Orphan tasks (parent missing externally): Automatically promoted to root level for display. If a task's parent is deleted externally (e.g., via remote backend, data import, or database corruption), the orphaned child tasks are displayed at the root level rather than being hidden or causing errors. This ensures data visibility and allows manual cleanup.
```
**Matches expected behavior**: YES
