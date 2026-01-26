# [022] Docs: Priority filter flag documentation incomplete in CLI reference

## Type
documentation

## Severity
low

## Test Location
- File: backend/sqlite/cli_test.go
- Functions:
  - TestPriorityFilterHighSQLiteCLI
  - TestPriorityFilterMediumSQLiteCLI
  - TestPriorityFilterLowSQLiteCLI
  - TestPriorityFilterRangeSQLiteCLI
  - TestPriorityFilterSingleSQLiteCLI
  - TestPriorityFilterUndefinedSQLiteCLI
  - TestPriorityFilterCombinedWithStatusSQLiteCLI

## Documentation Gap
The CLI reference (docs/reference/cli.md) shows priority filter as:
```
-p, --priority <filter> | string | Filter by priority (1,2,3 or high/medium/low)
```

But tests reveal additional functionality not documented:
1. Undefined priority filtering (0)
2. Range syntax (e.g., "1-3" for priorities 1 through 3)
3. Combination with status filter
4. "high", "medium", "low" named levels with specific ranges

## Current Documentation
docs/reference/cli.md line 65:
```
| `-p, --priority <filter>` | string | Filter by priority (1,2,3 or high/medium/low) |
```

docs/how-to/task-management.md shows more detail but still incomplete:
```bash
todoat MyList -p 1,2,3,4    # High priority tasks
todoat MyList -p high       # priorities 1-4
todoat MyList -p medium     # priority 5
todoat MyList -p low        # priorities 6-9
```

## Expected Documentation Update
- Location: docs/reference/cli.md
- Section: Task Flags > For get/filter operations

Should add:
- [x] Range syntax: `--priority 1-3` filters priorities 1, 2, and 3
- [x] Undefined filter: `--priority 0` shows tasks without priority
- [x] Clarify that high=1-4, medium=5, low=6-9
- [x] Document that filters can be combined: `-s TODO -p high`

## Resolution

**Fixed in**: this session
**Fix description**: Added comprehensive "Priority Filter Syntax" subsection to docs/reference/cli.md under "For get/filter operations" section.

### Changes Made
1. Updated priority filter description from "(1,2,3 or high/medium/low)" to "(see below)"
2. Added new "Priority Filter Syntax" subsection documenting:
   - Single value, comma-separated, range, named levels, and undefined (0) formats
   - Named priority levels table (high=1-4, medium=5, low=6-9)
   - Examples of combining priority with status filters

### Verification Log
```bash
$ go test ./backend/sqlite/... -run "TestPriorityFilter" -v
=== RUN   TestPriorityFilterSingleSQLiteCLI
--- PASS: TestPriorityFilterSingleSQLiteCLI (0.04s)
=== RUN   TestPriorityFilterRangeSQLiteCLI
--- PASS: TestPriorityFilterRangeSQLiteCLI (0.05s)
=== RUN   TestPriorityFilterHighSQLiteCLI
--- PASS: TestPriorityFilterHighSQLiteCLI (0.04s)
=== RUN   TestPriorityFilterMediumSQLiteCLI
--- PASS: TestPriorityFilterMediumSQLiteCLI (0.04s)
=== RUN   TestPriorityFilterLowSQLiteCLI
--- PASS: TestPriorityFilterLowSQLiteCLI (0.04s)
=== RUN   TestPriorityFilterUndefinedSQLiteCLI
--- PASS: TestPriorityFilterUndefinedSQLiteCLI (0.03s)
=== RUN   TestPriorityFilterNoMatchSQLiteCLI
--- PASS: TestPriorityFilterNoMatchSQLiteCLI (0.02s)
=== RUN   TestPriorityFilterJSONSQLiteCLI
--- PASS: TestPriorityFilterJSONSQLiteCLI (0.03s)
=== RUN   TestPriorityFilterInvalidSQLiteCLI
--- PASS: TestPriorityFilterInvalidSQLiteCLI (0.03s)
=== RUN   TestPriorityFilterCombinedWithStatusSQLiteCLI
--- PASS: TestPriorityFilterCombinedWithStatusSQLiteCLI (0.04s)
PASS
ok      todoat/backend/sqlite   0.360s
```
**Matches expected behavior**: YES
