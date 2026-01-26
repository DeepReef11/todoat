# [016] Docs: Bulk operations feature documented in overview but not in how-to guide

## Type
documentation

## Severity
medium

## Test Location
- File: backend/sqlite/cli_test.go
- Functions:
  - TestBulkCompleteDirectChildrenSQLiteCLI
  - TestBulkCompleteAllDescendantsSQLiteCLI
  - TestBulkUpdatePrioritySQLiteCLI
  - TestBulkDeleteChildrenSQLiteCLI
  - TestBulkNoMatchErrorSQLiteCLI
  - TestBulkEmptyMatchSQLiteCLI
  - TestBulkCountOutputSQLiteCLI
  - TestBulkCompleteJSONOutputSQLiteCLI
- File: backend/sqlite/integration_test.go
  - TestBulkOperationsIntegration

## Documentation Gap
The feature `docs/explanation/features-overview.md` lists Bulk Operations as a stable feature with a link to `task-management.md#bulk-operations`, but this anchor does not exist in `docs/how-to/task-management.md`.

## Feature Description (from tests)
Bulk operations allow operating on multiple tasks using glob patterns:
- `todoat MyList complete "Parent/*"` - Complete direct children only
- `todoat MyList complete "Parent/**"` - Complete all descendants
- `todoat MyList update "Parent/**" --priority 1` - Update priority on all descendants
- `todoat MyList delete "Parent/*"` - Delete direct children only

## Expected Documentation
- Location: docs/how-to/task-management.md
- Section: ## Bulk Operations (anchor: #bulk-operations)

Should cover:
- [ ] Glob pattern syntax (`*` for direct children, `**` for all descendants)
- [ ] Supported operations (complete, update, delete)
- [ ] Examples of bulk complete
- [ ] Examples of bulk update (priority, status)
- [ ] Examples of bulk delete
- [ ] JSON output format for bulk operations
- [ ] Error handling (no match, empty match)
- [x] Count output showing affected tasks

## Resolution

**Fixed in**: this session
**Fix description**: Added comprehensive Bulk Operations section to docs/how-to/task-management.md with all required documentation.

### Verification Log
```bash
$ grep -n "## Bulk Operations" docs/how-to/task-management.md
376:## Bulk Operations

$ grep "bulk-operations" docs/explanation/features-overview.md
| **Bulk Operations** | Operate on multiple tasks using filters | ✅ Stable | [Task Management](task-management.md#bulk-operations) |
```
**Matches expected behavior**: YES

The documentation now includes:
- Glob pattern syntax (`*` and `**`)
- Supported operations (complete, update, delete)
- Examples of bulk complete with project hierarchy
- Examples of bulk update (priority)
- Examples of bulk delete with cascade note
- JSON output format structure
- Error handling (no match, empty match) with result codes
- Count output showing affected tasks
