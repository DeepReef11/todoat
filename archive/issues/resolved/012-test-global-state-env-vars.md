# [012] Test Quality: Environment Variable State Mutation

## Type
test-quality

## Severity
low

## Test Locations
- cmd/todoat/cmd/todoat_test.go (14 instances)
- backend/nextcloud/nextcloud_test.go (12 instances)
- backend/todoist/todoist_test.go (4 instances)

## Problem
Tests modify global environment variables using `os.Setenv()`. While most tests properly restore the original values using `defer`, this pattern has risks:

1. Tests cannot run in parallel (`t.Parallel()`)
2. If a test panics before defer runs, env is left in modified state
3. Shared state between tests creates potential for ordering dependencies

## Current Code Pattern
```go
oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
defer func() {
    _ = os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
}()
if err := os.Setenv("XDG_CONFIG_HOME", configDir); err != nil {
    t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
}
```

## Assessment
The current implementation is **acceptable** because:
- All instances use proper defer cleanup
- Tests save original values before modification
- Tests don't use `t.Parallel()`

However, there are improvements to consider:
1. Use `t.Setenv()` (Go 1.17+) which automatically cleans up:
```go
t.Setenv("XDG_CONFIG_HOME", configDir)  // Automatically restored after test
```

2. Consider using environment variable injection via interfaces instead of os.Setenv

## Suggested Fix
Replace manual save/restore pattern with `t.Setenv()`:
```go
// Instead of:
oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
defer func() {
    _ = os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
}()
if err := os.Setenv("XDG_CONFIG_HOME", configDir); err != nil {
    t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
}

// Use:
t.Setenv("XDG_CONFIG_HOME", configDir)
```

## Impact
- Current impact is low due to proper cleanup
- Prevents future parallelization of these tests
- Minor code smell that could be cleaned up

## Resolution

**Fixed in**: this session
**Fix description**: Replaced all manual save/defer/restore patterns with `t.Setenv()` calls across 3 test files. Removed unused `os` imports.

### Files Changed
- `cmd/todoat/cmd/todoat_test.go` - 4 test functions updated
- `backend/nextcloud/nextcloud_test.go` - 2 test functions updated, removed unused `os` import
- `backend/todoist/todoist_test.go` - 2 test functions updated, removed unused `os` import

### Verification Log
```bash
$ go test ./cmd/todoat/cmd/... -run 'TestAppStartsWithoutExistingConfigSQLiteCLI|TestDBCreatedAtCorrectPathSQLiteCLI|TestConfigCreatedAtCorrectPathSQLiteCLI|TestConfigCreatedOnCLIExecutionSQLiteCLI' -v
=== RUN   TestAppStartsWithoutExistingConfigSQLiteCLI
--- PASS: TestAppStartsWithoutExistingConfigSQLiteCLI (0.00s)
=== RUN   TestDBCreatedAtCorrectPathSQLiteCLI
--- PASS: TestDBCreatedAtCorrectPathSQLiteCLI (0.02s)
=== RUN   TestConfigCreatedAtCorrectPathSQLiteCLI
--- PASS: TestConfigCreatedAtCorrectPathSQLiteCLI (0.00s)
=== RUN   TestConfigCreatedOnCLIExecutionSQLiteCLI
--- PASS: TestConfigCreatedOnCLIExecutionSQLiteCLI (0.02s)
PASS
ok  	todoat/cmd/todoat/cmd	0.051s

$ go test ./backend/nextcloud/... -run 'TestNextcloudCredentialsFromEnv|TestConfigFromEnv' -v
=== RUN   TestNextcloudCredentialsFromEnv
--- PASS: TestNextcloudCredentialsFromEnv (0.00s)
=== RUN   TestConfigFromEnv
--- PASS: TestConfigFromEnv (0.00s)
PASS
ok  	todoat/backend/nextcloud	0.004s

$ go test ./backend/todoist/... -run 'TestTodoistAPITokenFromEnv|TestConfigFromEnv' -v
=== RUN   TestTodoistAPITokenFromEnv
--- PASS: TestTodoistAPITokenFromEnv (0.00s)
=== RUN   TestConfigFromEnv
--- PASS: TestConfigFromEnv (0.00s)
PASS
ok  	todoat/backend/todoist	0.004s

$ go test ./...
ok  	todoat/backend/nextcloud	0.179s
ok  	todoat/backend/todoist	0.389s
ok  	todoat/cmd/todoat/cmd	0.996s
# all packages pass
```
**Matches expected behavior**: YES
