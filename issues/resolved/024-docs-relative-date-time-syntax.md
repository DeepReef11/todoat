# [024] Docs: Relative date with time syntax not fully documented

## Type
documentation

## Severity
low

## Test Location
- File: backend/sqlite/cli_test.go
- Functions:
  - TestRelativeDatePlusDaysWithTimeSQLiteCLI
  - TestRelativeDateWithTimeSQLiteCLI

## Documentation Gap
The docs show basic relative date syntax but don't fully document time combinations:

Current in docs/how-to/task-management.md:
```bash
todoat MyList add "Morning standup" --due-date "tomorrow 09:00"
todoat MyList add "Friday meeting" --due-date "+2d 14:00"
```

But tests suggest more patterns are supported that aren't documented.

## Expected Documentation Update
- Location: docs/how-to/task-management.md
- Section: ### Relative Dates

Should add table or examples covering:
- [x] Full syntax: `+Nd HH:MM` format
- [x] Whether timezone is supported with relative dates
- [x] Any limitations (e.g., can you do `tomorrow 14:30+05:00`?)
- [x] Examples with various time formats

## Resolution

**Fixed in**: this session
**Fix description**: Added new "Relative Dates with Time" section to docs/how-to/task-management.md

### Documentation Added
- New dedicated section "### Relative Dates with Time" covering:
  - Full syntax: `<relative-date> HH:MM` or `<relative-date> HH:MM:SS`
  - Time component format table (hours 0-23, minutes/seconds 00-59)
  - Clear note that timezone offsets are NOT supported with relative dates
  - Multiple examples: tomorrow with time, +Nd with time, today with time, with seconds

### Verification Log
```bash
$ go test -v ./backend/sqlite/... -run "TestRelativeDatePlusDaysWithTimeSQLiteCLI|TestRelativeDateWithTimeSQLiteCLI"
=== RUN   TestRelativeDateWithTimeSQLiteCLI
--- PASS: TestRelativeDateWithTimeSQLiteCLI (0.03s)
=== RUN   TestRelativeDatePlusDaysWithTimeSQLiteCLI
--- PASS: TestRelativeDatePlusDaysWithTimeSQLiteCLI (0.02s)
PASS
```
**Matches expected behavior**: YES
