# [003] Migrate target-info shows different status names than main app

## Type
doc-mismatch

## Category
other

## Severity
low

## Steps to Reproduce
```bash
# View target info for a list
./todoat migrate --target-info sqlite --list TestManual001

# Output shows:
# List: TestManual001 (26 tasks)
#   - Buy groceries (COMPLETED)
#   - Urgent report (NEEDS-ACTION)
#   ...
```

## Expected Behavior
Status values should match the rest of the application's terminology:
- `TODO` (not `NEEDS-ACTION`)
- `DONE` (not `COMPLETED`)

## Actual Behavior
The migrate target-info command shows:
- `NEEDS-ACTION` instead of `TODO`
- `COMPLETED` instead of `DONE`

This is inconsistent with:
1. The main task display which uses `[TODO]`, `[DONE]`, etc.
2. The documentation which refers to statuses as TODO, IN-PROGRESS, DONE, CANCELLED

## Error Output
N/A

## Environment
- OS: Linux
- Runtime version: Go dev build

## Documentation Reference
- File: `docs/task-management.md`
- Section: Task Status Values
- Documented: `TODO`, `IN-PROGRESS`, `DONE`, `CANCELLED`

## Related Files
- migrate command implementation

## Recommended Fix
FIX CODE - Map internal iCalendar status names (NEEDS-ACTION, COMPLETED) to user-facing names (TODO, DONE) in migrate target-info output for consistency

## Resolution

**Fixed in**: this session
**Fix description**: Exported the existing `statusToString` function in `internal/views/renderer.go` as `StatusToString` and used it in `doMigrateTargetInfo` to convert internal iCalendar status names to user-facing names in both text and JSON output.
**Test added**: `TestIssue003TargetInfoStatusDisplayText` and `TestIssue003TargetInfoStatusDisplayJSON` in `internal/migrate/migrate_test.go`

### Verification Log
```bash
$ go test -v -run "TestIssue003" ./internal/migrate/...
=== RUN   TestIssue003TargetInfoStatusDisplayText
--- PASS: TestIssue003TargetInfoStatusDisplayText (0.04s)
=== RUN   TestIssue003TargetInfoStatusDisplayJSON
--- PASS: TestIssue003TargetInfoStatusDisplayJSON (0.04s)
PASS
ok  	todoat/internal/migrate	0.085s
```
**Matches expected behavior**: YES

Output now shows `TODO` and `DONE` instead of `NEEDS-ACTION` and `COMPLETED`.
