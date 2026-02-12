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

### With Description and Color

Set description and color during creation:

```bash
todoat list create "Personal Goals" \
  --description "Goals for 2026" \
  --color "#00CC66"
```

You can also update these properties later:

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

### Delete a List

```bash
todoat list delete "List Name"
```

You'll be prompted to confirm. Behavior depends on the backend:
- **SQLite**: Soft-deletes the list (moves to trash, recoverable)
- **Nextcloud**: Returns an error (CalDAV deletion not supported to prevent data loss)
- **Todoist, Google Tasks, Microsoft To Do, Git, File**: Permanently deletes the list

### Force Delete (Skip Confirmation)

```bash
todoat -y list delete "List Name"
```

Use the global `-y` (or `--no-prompt`) flag to skip the confirmation prompt.

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

# List info as JSON
todoat --json list info "Work Tasks"

# View trash as JSON
todoat --json list trash
```

**List all lists:**
```json
{
  "lists": [
    {"id": "abc-123", "name": "Work Tasks", "tasks": 12, "color": "#0066cc"},
    {"id": "def-456", "name": "Personal", "tasks": 5, "color": "#ff5733"}
  ],
  "result": "INFO_ONLY"
}
```

**List info:**
```json
{
  "name": "Work Tasks",
  "id": "abc-123",
  "color": "#0066cc",
  "description": "All work-related tasks",
  "tasks": 12,
  "result": "INFO_ONLY"
}
```

**Trash:**
```json
{
  "lists": [
    {"name": "Old Project", "id": "xyz-789", "deleted_at": "2026-01-15T10:30:00Z"}
  ],
  "purged_count": 0,
  "result": "INFO_ONLY"
}
```

## Examples

### Organize by Area

```bash
# Create lists with colors for different areas
todoat list create "Work" --color "#0066CC"
todoat list create "Personal" --color "#00CC66"
todoat list create "Shopping" --color "#FF9900"
todoat list create "Home" --color "#FF5733"
```

### Project-Based Organization

```bash
# Create project-specific lists with descriptions
todoat list create "Project Alpha" --description "Q1 2026 launch project"
todoat list create "Project Beta" --description "Mobile app development"
todoat list create "Maintenance" --description "Bug fixes and updates"
```

### Archive Old Projects

```bash
# Move completed project to trash
todoat list delete "Project Alpha"

# Later, if you need it back
todoat list trash restore "Project Alpha"
```

## Export and Import

### Export a List

Export tasks to various formats for backup or sharing:

```bash
# Export to JSON (default)
todoat list export "Work Tasks"

# Export to specific format
todoat list export "Work Tasks" --format json
todoat list export "Work Tasks" --format csv
todoat list export "Work Tasks" --format ical
todoat list export "Work Tasks" --format sqlite

# Specify output file
todoat list export "Work Tasks" --output ~/backup/work.json
```

| Format | Extension | Description |
|--------|-----------|-------------|
| `json` | .json | JSON format (default) |
| `csv` | .csv | Comma-separated values |
| `ical` | .ics | iCalendar format |
| `sqlite` | .db | SQLite database |

JSON exports include list metadata (name) alongside tasks. This format supports both the current structure and older array-only exports:

```json
{
  "list_name": "Work Tasks",
  "tasks": [ ... ]
}
```

### Import a List

Import tasks from a file:

```bash
# Auto-detect format from file extension
todoat list import ~/backup/work.json

# Specify format explicitly
todoat list import ~/backup/tasks.txt --format csv
```

Imported tasks receive new unique IDs, so you can safely reimport a previously exported list (e.g., after deleting and recreating it) without ID conflicts.

## Database Maintenance

### View Statistics

```bash
todoat list stats
```

Shows database statistics including:
- Total tasks and lists
- Tasks by status
- Storage usage

### Compact Database

```bash
todoat list vacuum
```

Reclaims unused space in the SQLite database. Run this periodically if you frequently delete tasks.

## Notes

- List names must be unique within a backend
- Deleting a list on SQLite moves it to trash (recoverable for 30 days by default); on other backends, deletion may be permanent
- Colors are supported by most backends (Nextcloud, Todoist)
- List properties sync when synchronization is enabled

## See Also

- [Task Management](task-management.md) - Working with tasks in lists
- [Backends](../explanation/backends.md) - Backend-specific list features
- [Synchronization](sync.md) - Syncing lists across devices
