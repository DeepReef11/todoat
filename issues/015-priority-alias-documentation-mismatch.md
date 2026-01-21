# [015] Example Mismatch: Priority Alias Ranges Documented Incorrectly

## Type
doc-mismatch

## Category
user-journey

## Severity
medium

## Location
- File: `docs/task-management.md`
- Lines: 87-89
- Context: Documentation for priority filtering with named aliases

## Documented Command
```bash
todoat MyList -p high    # Same as 1,2,3
todoat MyList -p medium  # Priority 4-6
todoat MyList -p low     # Priority 7-9
```

## Actual Result
The actual implementation behavior (verified by tests in `backend/sqlite/cli_test.go:991-1047`):

```bash
todoat MyList -p high    # priorities 1-4
todoat MyList -p medium  # priority 5 only
todoat MyList -p low     # priorities 6-9
```

## Working Alternative
The correct documentation exists in `doc/examples.md:125-127`:

```bash
todoat Work get -p high     # priorities 1-4
todoat Work get -p medium   # priority 5
todoat Work get -p low      # priorities 6-9
```

## Recommended Fix
FIX EXAMPLE - Update `docs/task-management.md` lines 87-89 from:

```bash
todoat MyList -p high    # Same as 1,2,3
todoat MyList -p medium  # Priority 4-6
todoat MyList -p low     # Priority 7-9
```

To:

```bash
todoat MyList -p high    # priorities 1-4
todoat MyList -p medium  # priority 5
todoat MyList -p low     # priorities 6-9
```

## Impact
Users following the task-management.md documentation will have incorrect expectations about what tasks are included in each priority filter level. Tasks with priority 4 would unexpectedly appear in "high" results, and tasks with priority 6 would not appear in "medium" results as documented.

## Dependencies
None
