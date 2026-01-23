# [2] Example Mismatch: --due flag should be --due-date in backends.md

## Type
doc-mismatch

## Category
user-journey

## Severity
medium

## Location
- File: `docs/explanation/backends.md`
- Lines: 168, 252
- Context: Google Tasks and Microsoft To Do usage examples

## Documented Command
```bash
todoat -b google "My Tasks" add "Submit report" --due tomorrow
todoat -b mstodo "My Tasks" add "Submit report" --due tomorrow
```

## Actual Result
```bash
$ todoat "MyList" add "Submit report" --due tomorrow
Error: unknown flag: --due
```

## Working Alternative
```bash
todoat "MyList" add "Submit report" --due-date tomorrow
```

## Recommended Fix
FIX EXAMPLE - Update the examples in `docs/explanation/backends.md` at lines 168 and 252 to use `--due-date` instead of `--due`:

Line 168 (Google Tasks):
```bash
todoat -b google "My Tasks" add "Submit report" --due-date tomorrow
```

Line 252 (Microsoft To Do):
```bash
todoat -b mstodo "My Tasks" add "Submit report" --due-date tomorrow
```

## Impact
Users following these examples will see "unknown flag: --due" error when trying to set due dates for tasks.

## Note
These examples are in the Google Tasks and Microsoft To Do sections, which are documented as "not yet available via CLI". However, the examples still show incorrect flag syntax that would confuse users if/when these backends become available.

## Resolution

**Fixed in**: this session
**Fix description**: Updated `--due` to `--due-date` in docs/explanation/backends.md at lines 168 and 252

### Verification Log
```bash
$ grep "due-date tomorrow" docs/explanation/backends.md
todoat -b google "My Tasks" add "Submit report" --due-date tomorrow
todoat -b mstodo "My Tasks" add "Submit report" --due-date tomorrow
```
**Matches expected behavior**: YES
