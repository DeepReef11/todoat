# [031] Test: `todoat reminder dismiss` command has no CLI test

## Type
test-coverage

## Severity
low

## Documentation Location
- File: docs/reference/cli.md
- Section: ## reminder > ### Subcommands
- Line: `dismiss <task> | Dismiss current reminder`

Also documented in docs/how-to/reminders.md:
```bash
todoat reminder dismiss "Task name"
```

## Feature Description
The `reminder dismiss` command dismisses the current reminder for a task. The user will be reminded again at the next interval. There is unit test coverage for the reminder service (`TestReminderDismiss`), but no CLI integration test.

## Current Test Coverage
In `internal/reminder/reminder_test.go`:
- `TestReminderDismiss` tests the service layer
- No CLI test verifies the actual command execution

## Expected Test
- Location: internal/reminder/reminder_test.go
- Name: TestReminderDismissCLI
- Should verify:
  - [ ] Command returns success for valid task
  - [ ] Dismissed reminder doesn't reappear immediately
  - [ ] Error handling for non-existent task
  - [ ] JSON output format works

## Example Implementation
```go
func TestReminderDismissCLI(t *testing.T) {
    cli := setupTestCLI(t)
    defer cli.Cleanup()

    // Create a task with due date
    cli.MustExecute("-y", "list", "create", "TestList")
    cli.MustExecute("-y", "TestList", "add", "Meeting prep", "--due-date", "+1h")

    // Dismiss reminder for the task
    stdout := cli.MustExecute("-y", "reminder", "dismiss", "Meeting prep")
    if strings.Contains(stdout, "error") {
        t.Errorf("unexpected error in output")
    }
}
```


---
## Resolution (2026-01-22)

This issue was incorrectly created. The tests `TestReminderDisable`, `TestReminderDismiss`, and `TestReminderList` in `internal/reminder/reminder_test.go` ARE CLI integration tests - they use `cli.MustExecute()` to invoke the actual CLI commands. The issue text incorrectly stated these were only 'service layer' tests.

Status: **Invalid - Tests Already Exist**

