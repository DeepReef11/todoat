# [002] Documentation shows emoji for recurring tasks but app uses [R]

## Type
doc-mismatch

## Category
other

## Severity
low

## Steps to Reproduce
```bash
# Create a recurring task
./todoat TestList add "Daily standup" --recur daily

# View the task
./todoat TestList | grep "Daily standup"
# Output: [TODO] Daily standup [R]
```

## Expected Behavior
Based on docs/task-management.md line 218, the output should show:
```
🔄 TODO   Daily standup                    Jan 20
```

## Actual Behavior
The app shows:
```
[TODO] Daily standup [R]
```

The recurring indicator is `[R]` instead of the `🔄` emoji shown in documentation.

## Error Output
N/A

## Environment
- OS: Linux
- Runtime version: Go dev build

## Documentation Reference
- File: `docs/task-management.md`
- Section: Recurring Tasks
- Documented indicator: `🔄 TODO`
- Actual indicator: `[TODO] ... [R]`

## Related Files
- `docs/task-management.md`

## Recommended Fix
FIX DOCS - Update documentation to show `[R]` indicator instead of emoji, or FIX CODE if emoji was intended
