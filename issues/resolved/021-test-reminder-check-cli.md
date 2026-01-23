# [021] Test: `todoat reminder check` command has no CLI test

## Type
test-coverage

## Severity
low

## Documentation Location
- File: docs/reference/cli.md
- Section: ## reminder > ### Subcommands
- Line: `check | Check for due reminders`

Also documented in docs/how-to/reminders.md:
```bash
todoat reminder check
```

## Feature Description
The `reminder check` command checks all tasks and sends reminders for those due within the configured intervals. There is unit test coverage for the reminder service (`TestReminderServiceCheckReminders`), but no CLI integration test.

## Current Test Coverage
In `internal/reminder/reminder_test.go`:
- `TestReminderServiceCheckReminders` tests the service layer
- No CLI test verifies the actual command execution

## Expected Test
- Location: internal/reminder/reminder_test.go
- Name: TestReminderCheckCLI
- Should verify:
  - [ ] Command returns success when no reminders due
  - [ ] Command outputs list of due reminders when present
  - [ ] JSON output format works
  - [ ] Exit code is appropriate (success when no action needed)

## Example Implementation
```go
func TestReminderCheckCLI(t *testing.T) {
    cli := setupTestCLI(t)
    defer cli.Cleanup()

    // Create a task with due date
    cli.MustExecute("-y", "list", "create", "TestList")
    cli.MustExecute("-y", "TestList", "add", "Soon task", "--due-date", "+1h")

    // Check reminders
    stdout := cli.MustExecute("-y", "reminder", "check")
    // Output may vary based on whether task falls within reminder window
    if stdout == "" {
        t.Logf("No reminders triggered (expected if not in window)")
    }
}
```

## Resolution

**Fixed in**: this session
**Fix description**: Added comprehensive CLI test for `todoat reminder check` command
**Test added**: TestReminderCheckCLI in internal/reminder/reminder_test.go

### Verification Log
```bash
$ go test -v -run TestReminderCheckCLI ./internal/reminder/...
=== RUN   TestReminderCheckCLI
=== RUN   TestReminderCheckCLI/returns_success_when_no_reminders_due
=== RUN   TestReminderCheckCLI/outputs_list_of_due_reminders_when_present
=== RUN   TestReminderCheckCLI/handles_multiple_tasks_with_different_due_dates
=== RUN   TestReminderCheckCLI/works_when_reminders_are_disabled
--- PASS: TestReminderCheckCLI (0.16s)
    --- PASS: TestReminderCheckCLI/returns_success_when_no_reminders_due (0.03s)
    --- PASS: TestReminderCheckCLI/outputs_list_of_due_reminders_when_present (0.04s)
    --- PASS: TestReminderCheckCLI/handles_multiple_tasks_with_different_due_dates (0.06s)
    --- PASS: TestReminderCheckCLI/works_when_reminders_are_disabled (0.03s)
PASS
ok  	todoat/internal/reminder	0.169s
```
**Matches expected behavior**: YES

### Test Coverage
The new test verifies:
- [x] Command returns success when no reminders due
- [x] Command outputs list of due reminders when present
- [x] Exit code is appropriate (ResultActionCompleted)
- [x] Multiple tasks with different due dates handled correctly
- [x] Disabled reminders config properly prevents triggering

Note: JSON output format was not tested as the `reminder check` command uses standard text output format (consistent with other reminder subcommands). The global `--json` flag exists but is primarily used for error output formatting.
