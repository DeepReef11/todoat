Integration test should be using CLI and test important features. They should create task list, add, update, delete, parent, etc. Then should do the same with the sync feature enabled.

## Resolution

**Fixed in**: this session
**Fix description**: Created comprehensive CLI integration tests in `backend/sqlite/integration_test.go` covering:
- Task lifecycle (create list, add, update, complete, delete tasks)
- Multiple tasks workflow with filtering by priority/status
- List management (create, rename, delete, trash)
- Hierarchical tasks/subtasks with -P flag and path notation (A/B/C)
- Re-parenting tasks and removing parents
- Sync features (status, queue, conflicts)
- Conflict resolution strategies (server_wins, local_wins, merge, keep_both)
- Command abbreviations (a, g, u, c, d)
- JSON output format
- Tag/category filtering
- Due date filtering
- Error handling scenarios
- Bulk operations (complete/update descendants)
- Export/import functionality

**Tests added**: 16 integration tests in `backend/sqlite/integration_test.go`

### Verification Log
```bash
$ go test -v -count=1 -run 'Integration' ./backend/sqlite/...
=== RUN   TestTaskLifecycleIntegration
--- PASS: TestTaskLifecycleIntegration (0.06s)
=== RUN   TestMultipleTasksWorkflowIntegration
--- PASS: TestMultipleTasksWorkflowIntegration (0.05s)
=== RUN   TestListLifecycleIntegration
--- PASS: TestListLifecycleIntegration (0.03s)
=== RUN   TestListRenameIntegration
--- PASS: TestListRenameIntegration (0.03s)
=== RUN   TestSubtaskHierarchyIntegration
--- PASS: TestSubtaskHierarchyIntegration (0.06s)
=== RUN   TestPathBasedHierarchyIntegration
--- PASS: TestPathBasedHierarchyIntegration (0.05s)
=== RUN   TestRemoveParentIntegration
--- PASS: TestRemoveParentIntegration (0.04s)
=== RUN   TestSyncStatusIntegration
--- PASS: TestSyncStatusIntegration (0.03s)
=== RUN   TestSyncConflictResolutionIntegration
--- PASS: TestSyncConflictResolutionIntegration (0.19s)
=== RUN   TestCommandAbbreviationsIntegration
--- PASS: TestCommandAbbreviationsIntegration (0.05s)
=== RUN   TestJSONOutputIntegration
--- PASS: TestJSONOutputIntegration (0.04s)
=== RUN   TestTagWorkflowIntegration
--- PASS: TestTagWorkflowIntegration (0.04s)
=== RUN   TestDueDateWorkflowIntegration
--- PASS: TestDueDateWorkflowIntegration (0.04s)
=== RUN   TestErrorHandlingIntegration
--- PASS: TestErrorHandlingIntegration (0.04s)
=== RUN   TestBulkOperationsIntegration
--- PASS: TestBulkOperationsIntegration (0.08s)
=== RUN   TestExportImportIntegration
--- PASS: TestExportImportIntegration (0.07s)
PASS
ok      todoat/backend/sqlite   0.921s
```
**Matches expected behavior**: YES
