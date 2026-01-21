# 009: Priority Filtering

## Summary
Implement `-p, --priority` flag for filtering tasks by priority level when listing tasks.

## Documentation Reference
- Primary: `docs/explanation/task-management.md`
- Sections: Task Priority System, Task Filtering

## Dependencies
- Requires: 004-task-commands.md (get command exists)
- Requires: 005-status-system.md (filtering pattern established)
- Blocked by: none

## Complexity
**S (Small)** - Adds filtering flag to existing get command, similar pattern to status filtering

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestPriorityFilterSingle` - `todoat -y MyList -p 1` shows only priority 1 tasks
- [ ] `TestPriorityFilterRange` - `todoat -y MyList -p 1,2,3` shows tasks with priority 1, 2, or 3
- [ ] `TestPriorityFilterHigh` - `todoat -y MyList -p high` shows priorities 1-4
- [ ] `TestPriorityFilterMedium` - `todoat -y MyList -p medium` shows priority 5
- [ ] `TestPriorityFilterLow` - `todoat -y MyList -p low` shows priorities 6-9
- [ ] `TestPriorityFilterUndefined` - `todoat -y MyList -p 0` shows tasks with no priority set
- [ ] `TestPriorityFilterNoMatch` - `todoat -y MyList -p 1` with no matching tasks returns INFO_ONLY with message
- [ ] `TestPriorityFilterJSON` - `todoat -y --json MyList -p 1` returns filtered JSON result
- [ ] `TestPriorityFilterInvalid` - `todoat -y MyList -p 10` returns ERROR for invalid priority

### Manual Verification
- [ ] Priority filter message shows in output header: "filtered by priority: 1-4"
- [ ] Combined status and priority filters work: `todoat MyList -s TODO -p high`

## Implementation Notes

### Priority Aliases
```
high   = 1,2,3,4
medium = 5
low    = 6,7,8,9
```

### Flag Usage
```bash
todoat MyList -p 1           # Single priority
todoat MyList -p 1,2,3       # Multiple priorities
todoat MyList -p high        # Alias for 1-4
todoat MyList -p 0           # Undefined/no priority
todoat MyList -s TODO -p 1   # Combined with status filter
```

### Required Changes
1. Add `-p, --priority` flag to get/list command
2. Parse priority values (numbers 0-9 and aliases)
3. Filter tasks after retrieval (client-side filtering)
4. Update display header to show active filters

### Filter Logic
```go
func matchesPriorityFilter(task Task, priorities []int) bool {
    for _, p := range priorities {
        if task.Priority == p {
            return true
        }
    }
    return false
}
```

## Out of Scope
- Server-side priority filtering (backend-specific optimization)
- Priority ranges like "1-4" syntax (use alias "high" instead)
- Default priority filter (show all by default)
