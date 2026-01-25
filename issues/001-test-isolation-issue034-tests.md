# [001] Test Isolation Failure in Issue034 Tests

## Type
code-bug

## Category
other

## Severity
low

## Steps to Reproduce
```bash
# Run the tests (fresh system or after clean)
go test ./cmd/todoat/cmd -run TestIssue034 -v
```

## Expected Behavior
Tests should pass with isolated database state per test run.

## Actual Behavior
Tests fail with error: "Error: list 'StatsTestAutoDetect' already exists" and "Error: list 'VacuumTestAutoDetect' already exists"

## Error Output
```
=== RUN   TestIssue034StatsWithAutoDetect
    todoat_test.go:1826: failed to create list: Error: list 'StatsTestAutoDetect' already exists
--- FAIL: TestIssue034StatsWithAutoDetect (0.03s)
=== RUN   TestIssue034VacuumWithAutoDetect
    todoat_test.go:1881: failed to create list: Error: list 'VacuumTestAutoDetect' already exists
--- FAIL: TestIssue034VacuumWithAutoDetect (0.03s)
FAIL
FAIL	todoat/cmd/todoat/cmd	0.057s
FAIL
```

## Environment
- OS: Linux
- Runtime version: Go 1.25.5

## Possible Cause
The tests use `t.TempDir()` for `dbPath` and `configPath` which should provide isolation. However, when `auto_detect_backend: true` is set in the config, the backend resolution logic may be using a different database path than the one specified in `cfg.DBPath`, potentially falling back to a shared location or reusing state from previous test runs.

The test setup passes:
```go
cfg := &Config{
    DBPath:     dbPath,     // Temp dir path
    ConfigPath: configPath, // Temp dir path
}
```

But the config YAML only sets:
```yaml
default_backend: sqlite
auto_detect_backend: true
```

Without explicitly setting `backends.sqlite.path` in the test config, the auto_detect path resolution may not respect the test's `DBPath` override.

## Related Files
- `cmd/todoat/cmd/todoat_test.go:1803-1853` (TestIssue034StatsWithAutoDetect)
- `cmd/todoat/cmd/todoat_test.go:1858-1909` (TestIssue034VacuumWithAutoDetect)
- `cmd/todoat/cmd/todoat.go` (backend resolution logic)

## Recommended Fix
FIX CODE - Either:
1. Update the test config YAML to explicitly set `backends.sqlite.path` to the temp dbPath
2. Fix the backend resolution logic to always respect `cfg.DBPath` when set, even with `auto_detect_backend: true`
3. Add test cleanup to delete the test lists if they already exist

Example fix for option 1:
```go
configYAML := fmt.Sprintf(`default_backend: sqlite
auto_detect_backend: true
backends:
  sqlite:
    enabled: true
    path: "%s"
`, dbPath)
```
