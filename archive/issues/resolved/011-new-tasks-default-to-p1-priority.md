# New Tasks Default to P1 Priority Instead of Normal

## Summary
When creating a new task without specifying a priority, it defaults to P1 (highest priority) instead of a normal/default priority level.

## Steps to Reproduce

1. Create a task without specifying priority:
   ```bash
   todoat -b todoist inbox a aa
   Created task: aa (ID: 9937152068)
   ```
2. List tasks:
   ```bash
   todoat -b todoist inbox
   ```

## Expected Behavior
New tasks should have a normal/default priority (P4 or P7) unless explicitly specified.

## Actual Behavior
```bash
❯❯ todoat -b todoist inbox
Tasks in 'Inbox':
  [TODO] a4 [P1]
  └─ [TODO] s1 [P1]
  [TODO] stat1 [P7]
  [TODO] stat1 [P7]
  [TODO] a4 [P7]
  [TODO] stat2 [P7]
  [TODO] s1 [P7]
  └─ [TODO] s2 [P1]
  [TODO] aa [P1]    <-- Newly created task has P1
```

The newly created task `aa` shows `[P1]` even though no priority was specified.

## Impact
- Users may unintentionally create high-priority tasks
- Requires manual correction of priority after task creation
- Inconsistent with typical task management behavior where unspecified = normal priority

## Questions
1. Is this Todoist-specific or affects all backends?
2. What is the intended default priority?

## Resolution

**Fixed in**: this session
**Fix description**: The `internalToTodoistPriority` function in `backend/todoist/todoist.go` was treating priority 0 (unset) the same as priority 1-2 (highest), mapping it to Todoist priority 4 (urgent). Fixed by adding explicit handling for priority 0 to map to Todoist priority 1 (no priority/lowest).

**Root cause**: The switch statement checked `internal <= 2` which matched priority 0 (unset), returning Todoist's highest priority (4). When read back, Todoist priority 4 maps to internal priority 1 (P1).

**Test added**: `TestTodoistAddTaskWithoutPriority` and updated `TestTodoistPriorityMapping` (priority 0 case) in `backend/todoist/todoist_test.go`

### Verification Log
```bash
$ go test -v ./backend/todoist/... -run "TestTodoistPriorityMapping|TestTodoistAddTaskWithoutPriority"
=== RUN   TestTodoistAddTaskWithoutPriority
--- PASS: TestTodoistAddTaskWithoutPriority (0.00s)
=== RUN   TestTodoistPriorityMapping
=== RUN   TestTodoistPriorityMapping/0
--- PASS: TestTodoistPriorityMapping/0 (0.00s)
[...all subtests pass...]
--- PASS: TestTodoistPriorityMapping (0.00s)
PASS
ok  	todoat/backend/todoist	0.005s
```
**Matches expected behavior**: YES - Tasks created without specifying priority now show P7 (low/default) instead of P1 (highest).
