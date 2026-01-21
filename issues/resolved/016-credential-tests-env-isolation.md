# [016] Review: Credential tests not isolated from environment variables

## Type
code-bug

## Severity
critical

## Source
Code review - 2026-01-21_03-25-25

## Steps to Reproduce
1. Set environment variable: `export TODOAT_TODOIST_TOKEN=sometoken`
2. Run credential tests: `go test todoat/internal/credentials -v`
3. Observe failures in:
   - `TestCredentialsListJSONCLI`
   - `TestCredentialsListBackends`
   - `TestCredentialsEnvVarPriority`

## Expected Behavior
Tests should be isolated from external environment variables. Each test should save and restore ALL relevant environment variables including `TODOAT_TODOIST_TOKEN`.

## Actual Behavior
Tests save/restore `TODOAT_TODOIST_USERNAME` and `TODOAT_TODOIST_PASSWORD` but NOT `TODOAT_TODOIST_TOKEN`. When `TODOAT_TODOIST_TOKEN` is already set in the environment:
- Tests expecting todoist to have no credentials find credentials from the token
- `TestCredentialsEnvVarPriority` finds the pre-existing token instead of the test-set password

## Files Affected
- `internal/credentials/cli_test.go`
- `internal/credentials/credentials_test.go`

## Fix Required
Update tests to save and unset `TODOAT_TODOIST_TOKEN` at the start, restoring it at the end:

```go
// Save and restore env vars
origToken := os.Getenv("TODOAT_TODOIST_TOKEN")
defer func() {
    if origToken != "" {
        _ = os.Setenv("TODOAT_TODOIST_TOKEN", origToken)
    } else {
        _ = os.Unsetenv("TODOAT_TODOIST_TOKEN")
    }
}()
_ = os.Unsetenv("TODOAT_TODOIST_TOKEN")
```

## Resolution

**Fixed in**: 712e1db
**Fix description**: Added TODOAT_TODOIST_TOKEN to the save/unset/restore pattern in cli_test.go and credentials_test.go
**Tests added**: Existing tests now properly isolated

### Verification Log
```bash
$ export TODOAT_TODOIST_TOKEN=sometoken && go test todoat/internal/credentials -v -run "TestCredentialsListJSONCLI|TestCredentialsListBackends|TestCredentialsEnvVarPriority"
=== RUN   TestCredentialsListJSONCLI
--- PASS: TestCredentialsListJSONCLI (0.00s)
=== RUN   TestCredentialsListBackends
--- PASS: TestCredentialsListBackends (0.00s)
=== RUN   TestCredentialsEnvVarPriority
--- PASS: TestCredentialsEnvVarPriority (0.00s)
PASS
ok      todoat/internal/credentials     0.003s
```
**Matches expected behavior**: YES
