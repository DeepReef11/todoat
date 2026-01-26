# [001] List stats shows PROCESSING instead of IN-PROGRESS

## Type
code-bug

## Category
feature

## Severity
low

## Steps to Reproduce
```bash
# Create a task and set status to IN-PROGRESS
./todoat Tasks add "Test task" -y
./todoat Tasks update "Test task" -s IN-PROGRESS -y

# Run list stats
./todoat list stats
```

## Expected Behavior
Tasks by status should show "IN-PROGRESS" to match the user-facing CLI terminology:
```
Tasks by status:
  DONE                 2
  TODO                 7
  IN-PROGRESS          1
```

## Actual Behavior
Tasks by status shows "PROCESSING" instead:
```
Tasks by status:
  DONE                 2
  TODO                 7
  PROCESSING           1
```

## Error Output
```
Database Statistics
==================
Total tasks: 10

Tasks per list:
  MyList               0
  @work                0
  get                  0
  Tasks                10

Tasks by status:
  DONE                 2
  TODO                 7
  PROCESSING           1

Database size: 32.00 KB (32768 bytes)
```

## Environment
- OS: Linux 6.12.65-1-lts
- Runtime version: Go 1.25.5

## Possible Cause
The stats command is likely using the internal iCalendar status representation ("IN-PROCESS") instead of mapping it to the user-facing "IN-PROGRESS" term that's used elsewhere in the CLI.

## Related Files
- cmd/list.go (likely contains the stats command)
- internal/backend/ (status mapping functions)

## Resolution

**Fixed in**: this session
**Fix description**: Changed the status mapping in `Stats()` function from "PROCESSING" to "IN-PROGRESS" for StatusInProgress tasks
**Test added**: TestListStatsInProgressStatusSQLiteCLI in backend/sqlite/cli_test.go

### Verification Log
```bash
$ TODOAT_DB=/tmp/todoat-test-001.db ./todoat Tasks add "Test task" -y
Created task: Test task (ID: f1c27758-988e-467f-a860-41828baf4e6d)
ACTION_COMPLETED

$ TODOAT_DB=/tmp/todoat-test-001.db ./todoat Tasks update "Test task" -s IN-PROGRESS -y
Updated task: Test task
ACTION_COMPLETED

$ TODOAT_DB=/tmp/todoat-test-001.db ./todoat list stats
Database Statistics
==================
Total tasks: 7

Tasks per list:
  @work                6
  Tasks                1

Tasks by status:
  TODO                 6
  IN-PROGRESS          1

Database size: 32.00 KB (32768 bytes)
```
**Matches expected behavior**: YES
