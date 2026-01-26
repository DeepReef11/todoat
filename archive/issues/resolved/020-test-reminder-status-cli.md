# [020] Test: `todoat reminder status` command has no dedicated test

## Type
test-coverage

## Severity
low

## Documentation Location
- File: docs/reference/cli.md
- Section: ## reminder > ### Subcommands
- Line: `status | Show reminder configuration status`

Also documented in docs/how-to/reminders.md:
```bash
todoat reminder status
```

## Feature Description
The `reminder status` command shows the current reminder configuration and status. While this command is used within `TestReminderConfig` (as a verification step), there is no dedicated test verifying:
- The command outputs expected fields
- JSON output format works
- Error handling when config is missing

## Current Test Coverage
In `internal/reminder/reminder_test.go`:
- `TestReminderConfig` uses `todoat reminder status` but only as a smoke test
- No validation of status output format

## Expected Test
- Location: internal/reminder/reminder_test.go
- Name: TestReminderStatusCLI
- Should verify:
  - [ ] Command returns success
  - [ ] Output includes "enabled" status
  - [ ] Output includes configured intervals
  - [ ] JSON output format works (`--json` flag)
  - [ ] Output when reminders are disabled

## Example Implementation
```go
func TestReminderStatusCLI(t *testing.T) {
    cli := setupTestCLI(t)
    defer cli.Cleanup()

    // Test with default config
    stdout := cli.MustExecute("-y", "reminder", "status")
    if !strings.Contains(stdout, "enabled") {
        t.Errorf("expected 'enabled' in status output")
    }

    // Test JSON output
    jsonOut := cli.MustExecute("--json", "reminder", "status")
    if !strings.HasPrefix(strings.TrimSpace(jsonOut), "{") {
        t.Errorf("expected JSON object in output")
    }
}
```

## Resolution

**Fixed in**: this session
**Fix description**: Added dedicated `TestReminderStatusCLI` test function with subtests for enabled status, disabled status, and notification configuration verification.
**Test added**: `TestReminderStatusCLI` in `internal/reminder/reminder_test.go`

**Note**: The `--json` flag is not currently implemented for the `reminder status` command. The test covers all existing functionality: command success, enabled/disabled status, intervals display, and notification settings. JSON output support could be added as a separate enhancement if needed.

### Verification Log
```bash
$ go test ./internal/reminder/... -run TestReminderStatusCLI -v
=== RUN   TestReminderStatusCLI
=== RUN   TestReminderStatusCLI/shows_enabled_status_with_intervals
=== RUN   TestReminderStatusCLI/shows_disabled_status
=== RUN   TestReminderStatusCLI/shows_notification_configuration
--- PASS: TestReminderStatusCLI (0.00s)
    --- PASS: TestReminderStatusCLI/shows_enabled_status_with_intervals (0.00s)
    --- PASS: TestReminderStatusCLI/shows_disabled_status (0.00s)
    --- PASS: TestReminderStatusCLI/shows_notification_configuration (0.00s)
PASS
ok  	todoat/internal/reminder	0.005s
```
**Matches expected behavior**: YES
