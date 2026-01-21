# 007: List Management Commands

## Summary
Implement list management commands to create, view, and manage task lists. This enables users to organize tasks into separate lists.

## Documentation Reference
- Primary: `docs/explanation/list-management.md`
- Sections: Create List, View All Lists

## Dependencies
- Requires: 003-sqlite-backend.md (TaskList struct and GetTaskLists)
- Requires: 002-core-cli.md (CLI framework)
- Blocked by: none

## Complexity
**M (Medium)** - Multiple subcommands, database operations, display formatting

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestListCreate` - `todoat -y list create "MyList"` returns ACTION_COMPLETED and creates list
- [ ] `TestListCreateDuplicate` - `todoat -y list create "ExistingList"` returns ERROR for duplicate name
- [ ] `TestListView` - `todoat -y list` displays all available lists with task counts
- [ ] `TestListViewEmpty` - `todoat -y list` with no lists returns INFO_ONLY with helpful message
- [ ] `TestListViewJSON` - `todoat -y --json list` returns valid JSON with list array
- [ ] `TestListCreateJSON` - `todoat -y --json list create "Test"` returns JSON with created list info

### Manual Verification
- [ ] `todoat list` displays formatted table with list names and task counts
- [ ] Newly created lists appear immediately in list view

## Implementation Notes

### Command Structure
```
todoat list              # View all lists (default action)
todoat list create "Name"  # Create new list
```

### Required Changes
1. Add `list` subcommand to root command
2. Add `create` subcommand under `list`
3. Implement `CreateList(name string)` in SQLite backend
4. Display formatting with task counts

### Database Operations
- CreateList: INSERT INTO task_lists with generated UUID
- GetTaskLists: SELECT with COUNT of tasks per list

### Display Format
```
Available lists (3):

NAME          TASKS
Work          12
Personal      5
Shopping      0

INFO_ONLY
```

## Out of Scope
- List deletion (trash/restore) - separate roadmap item
- List properties (color, description) - separate roadmap item
- Interactive list selection - separate roadmap item
- List rename - separate roadmap item
