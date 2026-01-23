# [032] Test: `todoat reminder list` command has no CLI test

## Type
test-coverage

## Severity
low

## Documentation Location
- File: docs/reference/cli.md
- Section: ## reminder > ### Subcommands
- Line: `list | List upcoming reminders`

Also documented in docs/how-to/reminders.md:
```bash
todoat reminder list
```

## Feature Description
The `reminder list` command shows tasks with upcoming due dates within the configured reminder intervals. There is unit test coverage for the reminder service (`TestReminderList`), but no CLI integration test.

## Current Test Coverage
In `internal/reminder/reminder_test.go`:
- `TestReminderList` tests the service layer
- No CLI test verifies the actual command execution

## Expected Test
- Location: internal/reminder/reminder_test.go
- Name: TestReminderListCLI
- Should verify:
  - [ ] Command returns success
  - [ ] Output shows upcoming reminders when tasks exist
  - [ ] Empty output message when no reminders due
  - [ ] JSON output format works (`--json` flag)
  - [ ] Respects configured reminder intervals

## Example Implementation
```go
func TestReminderListCLI(t *testing.T) {
    cli := setupTestCLI(t)
    defer cli.Cleanup()

    // Create a task with due date within reminder window
    cli.MustExecute("-y", "list", "create", "TestList")
    cli.MustExecute("-y", "TestList", "add", "Soon task", "--due-date", "+1h")

    // List reminders
    stdout := cli.MustExecute("-y", "reminder", "list")
    // Should include the task if within reminder interval

    // Test JSON output
    jsonOut := cli.MustExecute("--json", "reminder", "list")
    if !strings.HasPrefix(strings.TrimSpace(jsonOut), "{") && !strings.HasPrefix(strings.TrimSpace(jsonOut), "[") {
        t.Errorf("expected JSON in output")
    }
}
```


---
## Resolution (2026-01-22)

This issue was incorrectly created. The tests `TestReminderDisable`, `TestReminderDismiss`, and `TestReminderList` in `internal/reminder/reminder_test.go` ARE CLI integration tests - they use `cli.MustExecute()` to invoke the actual CLI commands. The issue text incorrectly stated these were only 'service layer' tests.

Status: **Invalid - Tests Already Exist**

