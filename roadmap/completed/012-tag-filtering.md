# 012: Tag Filtering

## Summary
Implement tag/category support for tasks, including CLI flags for adding tags and filtering tasks by tags.

## Documentation Reference
- Primary: `dev-doc/TASK_MANAGEMENT.md`
- Sections: Filter by Tags, Add Tasks (Categories field)

## Dependencies
- Requires: 004-task-commands.md (add/update commands exist)
- Requires: 005-status-system.md (filtering pattern established)
- Blocked by: none

## Complexity
**S (Small)** - Adds tag flag to add/update commands and filtering flag to get command, similar pattern to status/priority filtering

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestAddTaskWithTag` - `todoat -y MyList add "Task" --tag work` adds task with tag
- [ ] `TestAddTaskMultipleTags` - `todoat -y MyList add "Task" --tag work --tag urgent` adds task with multiple tags
- [ ] `TestAddTaskCommaSeparatedTags` - `todoat -y MyList add "Task" --tag "work,urgent"` adds task with comma-separated tags
- [ ] `TestUpdateTaskTags` - `todoat -y MyList update "Task" --tag home` updates task tags
- [ ] `TestClearTaskTags` - `todoat -y MyList update "Task" --tag ""` clears task tags
- [ ] `TestFilterByTag` - `todoat -y MyList --tag work` shows only tasks with "work" tag
- [ ] `TestFilterByMultipleTags` - `todoat -y MyList --tag work --tag urgent` shows tasks with ANY of the tags (OR logic)
- [ ] `TestFilterTagNoMatch` - `todoat -y MyList --tag nonexistent` returns INFO_ONLY with message
- [ ] `TestFilterTagJSON` - `todoat -y --json MyList --tag work` returns filtered JSON result with tags array
- [ ] `TestFilterTagCombined` - `todoat -y MyList -s TODO --tag work` combined with status filter works

### Unit Tests (if needed)
- [ ] Tag parsing handles comma-separated values
- [ ] Tag storage uses Categories field in Task struct
- [ ] Empty string clears tags

### Manual Verification
- [ ] Tags display in task list when present
- [ ] Filter message shows in output header: "filtered by tags: work, urgent"

## Implementation Notes

### CLI Flags
```bash
# Adding tags
todoat MyList add "Task" --tag work
todoat MyList add "Task" --tag work --tag urgent
todoat MyList add "Task" --tag "work,urgent,home"

# Updating tags
todoat MyList update "Task" --tag "newTag"
todoat MyList update "Task" --tag ""  # Clear tags

# Filtering by tags
todoat MyList --tag work              # Single tag
todoat MyList --tag work --tag urgent # Multiple tags (OR)
todoat MyList -s TODO --tag work      # Combined with status
```

### Tag Storage
- Uses existing Categories field in Task struct
- Stored as comma-separated string internally
- Displayed as array in JSON output

### Database Schema
```sql
-- Categories column should already exist from 003-sqlite-backend
-- Uses TEXT field with comma-separated values
```

### Required Changes
1. Add `--tag` flag to add command (StringSliceVar for multiple values)
2. Add `--tag` flag to update command
3. Add `--tag` flag to get/list command for filtering
4. Parse comma-separated tags into slice
5. Filter tasks by tag membership (client-side)
6. Update display to show tags when present
7. Update JSON output to include tags as array

### Filter Logic
```go
func matchesTagFilter(task Task, filterTags []string) bool {
    taskTags := strings.Split(task.Categories, ",")
    for _, filterTag := range filterTags {
        for _, taskTag := range taskTags {
            if strings.TrimSpace(taskTag) == strings.TrimSpace(filterTag) {
                return true  // OR logic - match any tag
            }
        }
    }
    return false
}
```

## Out of Scope
- Tag auto-completion - separate item
- Tag management commands (list all tags, rename tags) - separate item
- Tag colors/icons - views customization
- AND logic for tag filtering - use multiple filter calls if needed
