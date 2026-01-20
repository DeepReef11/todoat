# List Management

Lists organize your tasks into separate containers. This guide covers creating, viewing, and managing task lists.

## Viewing Lists

### List All Task Lists

```bash
todoat list
```

Output shows:
- List name
- Number of tasks
- Color (if set)

### Interactive Selection

When you run `todoat` without specifying a list, an interactive menu appears:

```
Select a task list:
1. Work Tasks (12 tasks)
2. Personal (5 tasks)
3. Shopping (0 tasks)

Enter number:
```

## Creating Lists

### Basic Creation

```bash
todoat list create "My New List"
```

After creation, use `list update` to add description and color:

```bash
todoat list update "My New List" --description "Task list description" --color "#FF5733"
```

## Updating Lists

### Rename a List

```bash
todoat list update "Old Name" --name "New Name"
```

### Update Color

```bash
todoat list update "Work Tasks" --color "#0066CC"
```

### Update Description

```bash
todoat list update "Work Tasks" --description "Updated description text"
```

## Deleting Lists

### Soft Delete (Move to Trash)

```bash
todoat list delete "List Name"
```

You'll be prompted to confirm. Lists are moved to trash, not permanently deleted.

### Force Delete (Skip Confirmation)

```bash
todoat list delete "List Name" --force

# Or use no-prompt mode
todoat -y list delete "List Name"
```

## Trash Management

### View Trashed Lists

```bash
todoat list trash
```

Shows:
- List name
- When it was deleted
- Number of tasks (preserved)

### Restore a List

```bash
todoat list trash restore "List Name"
```

Restores the list and all its tasks.

### Permanently Delete

```bash
todoat list trash purge "List Name"
```

This is irreversible.

## List Information

### View List Details

```bash
todoat list info "Work Tasks"
```

Shows:
- Name
- Description
- Color
- Task count
- Created date

## List Properties

| Property | Description | Example |
|----------|-------------|---------|
| Name | Display name (required) | "Work Tasks" |
| Description | Purpose or notes | "All work-related tasks" |
| Color | Hex color code | "#FF5733" |
| ID | Unique identifier (auto-generated) | "abc-123-def" |

## Scripting

### Non-Interactive Mode

```bash
# List available lists (returns INFO_ONLY)
todoat -y list

# Delete without confirmation
todoat -y list delete "Temp List"
```

### JSON Output

```bash
# List all lists as JSON
todoat --json list
```

```json
{
  "lists": [
    {"id": "abc-123", "name": "Work Tasks", "task_count": 12, "color": "#0066cc"},
    {"id": "def-456", "name": "Personal", "task_count": 5, "color": "#ff5733"}
  ],
  "result": "INFO_ONLY"
}
```

## Examples

### Organize by Area

```bash
# Create lists for different areas
todoat list create "Work"
todoat list create "Personal"
todoat list create "Shopping"
todoat list create "Home"

# Add colors to lists
todoat list update "Work" --color "#0066CC"
todoat list update "Personal" --color "#00CC66"
todoat list update "Shopping" --color "#FF9900"
todoat list update "Home" --color "#FF5733"
```

### Project-Based Organization

```bash
# Create project-specific lists with descriptions
todoat list create "Project Alpha"
todoat list update "Project Alpha" --description "Q1 2026 launch project"

todoat list create "Project Beta"
todoat list update "Project Beta" --description "Mobile app development"

todoat list create "Maintenance"
todoat list update "Maintenance" --description "Bug fixes and updates"
```

### Archive Old Projects

```bash
# Move completed project to trash
todoat list delete "Project Alpha"

# Later, if you need it back
todoat list trash restore "Project Alpha"
```

## Notes

- List names must be unique within a backend
- Deleting a list moves it to trash (recoverable for 30 days by default)
- Colors are supported by most backends (Nextcloud, Todoist)
- List properties sync when synchronization is enabled

## See Also

- [Task Management](task-management.md) - Working with tasks in lists
- [Backends](backends.md) - Backend-specific list features
- [Synchronization](sync.md) - Syncing lists across devices
