# [019] Test: Documented 'yearly' recurrence pattern has no test

## Type
test-coverage

## Severity
medium

## Documentation Location
- File: docs/reference/cli.md
- Line: 56
- Section: Task Flags for add/update operations

## Feature Description
The CLI reference documents `yearly` as a valid recurrence pattern:
```
--recur <rule>  Recurrence (daily, weekly, monthly, yearly, or "every N days/weeks/months")
```

However, there are tests for:
- TestAddRecurringDailySQLiteCLI
- TestAddRecurringWeeklySQLiteCLI
- TestAddRecurringMonthlySQLiteCLI
- TestAddRecurringCustomSQLiteCLI

But no test exists for `yearly` recurrence.

## Expected Test
- Location: backend/sqlite/cli_test.go
- Name: TestAddRecurringYearlySQLiteCLI
- Should verify:
  - [ ] `todoat MyList add "Annual review" --recur yearly` creates task
  - [ ] Task displays with [R] indicator
  - [ ] Completion generates new task with next year's due date
  - [ ] JSON output shows recurrence rule

## Example Implementation
```go
func TestAddRecurringYearlySQLiteCLI(t *testing.T) {
    cli := setupTestCLI(t)
    defer cli.Cleanup()

    cli.MustExecute("-y", "list", "create", "TestList")
    stdout := cli.MustExecute("-y", "TestList", "add", "Annual review", "--recur", "yearly", "--due-date", "2026-01-01")

    // Verify task created
    if !strings.Contains(stdout, "Annual review") {
        t.Errorf("expected task summary in output")
    }

    // Verify recurrence in JSON
    json := cli.MustExecute("--json", "TestList")
    if !strings.Contains(json, "\"recurrence\"") {
        t.Errorf("expected recurrence in JSON output")
    }
}
```

## Resolution

**Fixed in**: this session
**Fix description**: Added TestAddRecurringYearlySQLiteCLI test following the existing pattern for daily/weekly/monthly recurrence tests
**Test added**: TestAddRecurringYearlySQLiteCLI in backend/sqlite/cli_test.go

### Verification Log
```bash
$ go test -v -run TestAddRecurringYearlySQLiteCLI ./backend/sqlite/
=== RUN   TestAddRecurringYearlySQLiteCLI
--- PASS: TestAddRecurringYearlySQLiteCLI (0.03s)
PASS
ok  	todoat/backend/sqlite	0.029s

$ go test -v -run 'TestAddRecurring.*SQLiteCLI' ./backend/sqlite/
=== RUN   TestAddRecurringDailySQLiteCLI
--- PASS: TestAddRecurringDailySQLiteCLI (0.03s)
=== RUN   TestAddRecurringWeeklySQLiteCLI
--- PASS: TestAddRecurringWeeklySQLiteCLI (0.03s)
=== RUN   TestAddRecurringMonthlySQLiteCLI
--- PASS: TestAddRecurringMonthlySQLiteCLI (0.02s)
=== RUN   TestAddRecurringYearlySQLiteCLI
--- PASS: TestAddRecurringYearlySQLiteCLI (0.02s)
=== RUN   TestAddRecurringCustomSQLiteCLI
--- PASS: TestAddRecurringCustomSQLiteCLI (0.03s)
PASS
ok  	todoat/backend/sqlite	0.129s
```
**Matches expected behavior**: YES
