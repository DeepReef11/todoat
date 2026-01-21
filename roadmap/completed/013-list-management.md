# 013: List Management

## Summary
Implement list management commands for creating, viewing, deleting, and managing task lists (calendars) with trash/restore functionality.

## Documentation Reference
- Primary: `docs/explanation/list-management.md`
- Sections: Create List, View All Lists, Trash and Restore Lists, List Properties

## Dependencies
- Requires: 003-sqlite-backend.md (task_lists table exists)
- Requires: 004-task-commands.md (list display established)
- Blocked by: none

## Complexity
**M (Medium)** - Multiple subcommands (create, delete, trash, restore), soft-delete logic, property management

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestListCreate` - `todoat -y list create "New List"` creates list, returns ACTION_COMPLETED
- [ ] `TestListCreateDuplicate` - `todoat -y list create "Existing"` returns ERROR for duplicate name
- [ ] `TestListView` - `todoat -y list` displays all lists with names and task counts
- [ ] `TestListViewJSON` - `todoat -y --json list` returns JSON array of lists
- [ ] `TestListDelete` - `todoat -y list delete "List Name"` moves list to trash, returns ACTION_COMPLETED
- [ ] `TestListDeleteNotFound` - `todoat -y list delete "NonExistent"` returns ERROR
- [ ] `TestListTrash` - `todoat -y list trash` displays deleted lists with timestamps
- [ ] `TestListTrashEmpty` - `todoat -y list trash` with no deleted lists returns INFO_ONLY
- [ ] `TestListRestore` - `todoat -y list trash restore "Deleted List"` restores list, returns ACTION_COMPLETED
- [ ] `TestListRestoreNotInTrash` - `todoat -y list trash restore "Active"` returns ERROR
- [ ] `TestListPurge` - `todoat -y list trash purge "Deleted List"` permanently deletes, returns ACTION_COMPLETED
- [ ] `TestListInfo` - `todoat -y list info "List Name"` shows list details (ID, color, description, task count)

### Unit Tests (if needed)
- [ ] Soft-delete sets deleted_at timestamp correctly
- [ ] Restore clears deleted_at timestamp
- [ ] Purge removes list and all associated tasks
- [ ] List name uniqueness validated (case-insensitive)

### Manual Verification
- [ ] `todoat list create` prompts for name if not provided
- [ ] `todoat list delete` prompts for confirmation without `-y`
- [ ] Deleted lists not shown in `todoat list` output
- [ ] Restored lists appear in `todoat list` output

## Implementation Notes

### CLI Commands Structure
```bash
todoat list                      # View all lists
todoat list create "Name"        # Create new list
todoat list delete "Name"        # Soft-delete to trash
todoat list info "Name"          # Show list details
todoat list trash                # View deleted lists
todoat list trash restore "Name" # Restore from trash
todoat list trash purge "Name"   # Permanently delete
```

### Database Schema (already exists)
```sql
-- task_lists table already has:
-- id, name, description, color, ctag, deleted_at, created_at
```

### Required Changes
1. Add `list` subcommand with nested commands (create, delete, info, trash)
2. Add `trash` subcommand with nested commands (restore, purge)
3. Implement soft-delete by setting `deleted_at` timestamp
4. Filter out deleted lists from normal `GetTaskLists()` calls
5. Add `GetDeletedLists()` method for trash view
6. Implement `RestoreList()` and `PurgeList()` methods
7. Add list info display with task counts

### List Operations Flow
```go
// Create
func CreateList(name string) (TaskList, error)

// Soft-delete
func DeleteList(id string) error  // Sets deleted_at

// Trash operations
func GetDeletedLists() ([]TaskList, error)
func RestoreList(id string) error   // Clears deleted_at
func PurgeList(id string) error     // Permanent DELETE with CASCADE
```

### JSON Output Format
```json
{
  "lists": [
    {"id": "abc-123", "name": "Work Tasks", "task_count": 12, "color": "#0066cc"},
    {"id": "def-456", "name": "Personal", "task_count": 5, "color": "#ff5733"}
  ],
  "result": "ACTION_COMPLETED"
}
```

## Out of Scope
- List color editing (separate item)
- List description editing (separate item)
- List sharing (Nextcloud-specific, advanced feature)
- Auto-purge scheduler (future enhancement)
- Interactive list selection when not specified (already exists)
