# [030] Test: `todoat reminder disable` command has no CLI test

## Type
test-coverage

## Severity
low

## Documentation Location
- File: docs/reference/cli.md
- Section: ## reminder > ### Subcommands
- Line: `disable <task> | Disable reminders for a task`

Also documented in docs/how-to/reminders.md:
```bash
todoat reminder disable "Task name"
```

## Feature Description
The `reminder disable` command stops reminders for a specific task. There is unit test coverage for the reminder service (`TestReminderDisable`), but no CLI integration test.

## Current Test Coverage
In `internal/reminder/reminder_test.go`:
- `TestReminderDisable` tests the service layer
- No CLI test verifies the actual command execution

## Expected Test
- Location: internal/reminder/reminder_test.go
- Name: TestReminderDisableCLI
- Should verify:
  - [ ] Command returns success for valid task
  - [ ] Task no longer shows in reminder list
  - [ ] Error handling for non-existent task
  - [ ] JSON output format works

## Example Implementation
```go
func TestReminderDisableCLI(t *testing.T) {
    cli := setupTestCLI(t)
    defer cli.Cleanup()

    // Create a task with due date
    cli.MustExecute("-y", "list", "create", "TestList")
    cli.MustExecute("-y", "TestList", "add", "Test task", "--due-date", "+1h")

    // Disable reminder for the task
    stdout := cli.MustExecute("-y", "reminder", "disable", "Test task")
    if strings.Contains(stdout, "error") {
        t.Errorf("unexpected error in output")
    }
}
```


---
## Resolution (2026-01-22)

This issue was incorrectly created. The tests `TestReminderDisable`, `TestReminderDismiss`, and `TestReminderList` in `internal/reminder/reminder_test.go` ARE CLI integration tests - they use `cli.MustExecute()` to invoke the actual CLI commands. The issue text incorrectly stated these were only 'service layer' tests.

Status: **Invalid - Tests Already Exist**

