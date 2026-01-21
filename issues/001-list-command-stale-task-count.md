# [001] List command shows stale task count

## Type
code-bug

## Category
feature

## Severity
medium

## Steps to Reproduce
```bash
# Check current task count in Work list
./todoat list | grep "^Work"
# Output: Work                 15

# Add a new task to Work list
./todoat Work add "Test task"
# Created task: Test task (ID: ...)

# Check task count again
./todoat list | grep "^Work"
# Output: Work                 15  (still shows 15, should show 16)

# Verify actual count using other commands
./todoat list stats | grep "Work"
# Output: Work                 16  (correct)

./todoat --json Work | jq '.count'
# Output: 16  (correct)
```

## Expected Behavior
After adding a task to a list, the `list` command should show the updated task count (16).

## Actual Behavior
The `list` command shows the old task count (15) while other commands like `list stats` and JSON output show the correct count (16).

## Error Output
N/A - no error, just incorrect count displayed

## Environment
- OS: Linux
- Runtime version: Go dev build

## Possible Cause
The `list` command may be caching task counts or reading from a stale cache that isn't updated after task creation.

## Related Files
- Likely in the list management or CLI display code

## Recommended Fix
FIX CODE - Ensure task count is refreshed or calculated dynamically when displaying list overview
