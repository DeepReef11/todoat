# [008] Review: Backend Isolation Not Used in Application Code

## Type
code-improvement

## Severity
low

## Source
Code review - 2026-01-24 17:47:03

## Description

The Issue #007 fix added `NewWithBackendID()` to support backend isolation via a `backend_id` column, but the application code continues to use `sqlite.New()` which defaults to backend_id "sqlite". This means the isolation capability exists but isn't utilized in practice.

## Steps to Reproduce

1. Configure multiple SQLite-type backends with custom names sharing the same database file
2. Create tasks in each backend
3. Observe that all tasks share the same backend_id "sqlite" rather than using distinct IDs

## Expected Behavior

When creating a SQLite backend with a custom name (e.g., "sqlite-work"), it should use `sqlite.NewWithBackendID(dbPath, "sqlite-work")` to ensure proper data isolation.

## Actual Behavior

All SQLite backends use `sqlite.New(dbPath)` which defaults to backend_id "sqlite", meaning data would not be isolated if multiple SQLite backends shared the same database file.

## Files Affected
- `cmd/todoat/cmd/todoat.go:2250` - getBackend()
- `cmd/todoat/cmd/todoat.go:2347` - fallback to sqlite
- `cmd/todoat/cmd/todoat.go:2538` - createCustomBackend() for sqlite type
- `cmd/todoat/cmd/todoat.go:5720` - doSync() local backend

## Impact

**Low** in current architecture because:
1. Remote backends (Nextcloud, Todoist, Google) don't use SQLite
2. Test configurations use separate database files
3. The sync cache typically uses separate database paths

## Suggested Fix

In `createCustomBackend()`, when creating a SQLite-type backend:

```go
case "sqlite":
    if customPath, ok := backendCfg["path"].(string); ok && customPath != "" {
        dbPath = config.ExpandPath(customPath)
    }
    // Use backend name for isolation when multiple SQLite backends share same db
    return sqlite.NewWithBackendID(dbPath, name)
```

## Notes

This issue is tracked for completeness. The current behavior is safe because different backends typically use different database files. The fix would provide defense-in-depth for edge cases.

## Resolution

**Fixed in**: this session
**Fix description**: Modified `createCustomBackend()` to use `sqlite.NewWithBackendID(dbPath, name)` instead of `sqlite.New(dbPath)` for custom SQLite-type backends. The other locations (getBackend default, fallback, doSync) correctly use the default "sqlite" backend ID for their purposes.
**Test added**: `TestIssue008_CustomSQLiteBackendUsesBackendID` in `cmd/todoat/cmd/todoat_test.go`

### Verification Log
```bash
$ go test -v -run TestIssue008_CustomSQLiteBackendUsesBackendID ./cmd/todoat/cmd/
=== RUN   TestIssue008_CustomSQLiteBackendUsesBackendID
--- PASS: TestIssue008_CustomSQLiteBackendUsesBackendID (0.04s)
PASS
ok      todoat/cmd/todoat/cmd   0.043s
```
**Matches expected behavior**: YES
