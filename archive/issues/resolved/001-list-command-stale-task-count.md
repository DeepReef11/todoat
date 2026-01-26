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

## Resolution

**Fixed in**: this session
**Fix description**: Added `invalidateListCache(cfg)` calls after task add and delete operations in cmd/todoat/cmd/todoat.go. The list cache is now properly invalidated when tasks are created or deleted, ensuring the `list` command shows accurate task counts.
**Test added**: `TestCacheInvalidationOnTaskAdd` and `TestCacheInvalidationOnTaskDelete` in internal/cache/cache_test.go

### Verification Log
```bash
$ export HOME=/tmp/todoat_test_issue001
$ /workspace/go/todoat/todoat list create Work
Created list: Work

$ /workspace/go/todoat/todoat list | grep "^Work"
Work                 0

$ /workspace/go/todoat/todoat Work add "Test task"
Created task: Test task (ID: 4d5da7ee-3b1d-4e48-ae9d-a4a654bc19e1)

$ /workspace/go/todoat/todoat list | grep "^Work"
Work                 1

$ /workspace/go/todoat/todoat Work add "Second task"
Created task: Second task (ID: 100876ff-e06d-41a2-8063-712c5c863453)

$ /workspace/go/todoat/todoat list | grep "^Work"
Work                 2

$ /workspace/go/todoat/todoat --json Work | jq '.count'
2

$ /workspace/go/todoat/todoat Work delete "Test task" -y
Deleted task: Test task

$ /workspace/go/todoat/todoat list | grep "^Work"
Work                 1
```
**Matches expected behavior**: YES
