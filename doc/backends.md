# Backends

todoat supports multiple storage backends for task management.

## SQLite Backend (Default)

The SQLite backend stores tasks locally in a database file. This is the default backend.

### Location

By default, the database is stored following XDG conventions:

```
~/.local/share/todoat/tasks.db
```

Or if `XDG_DATA_HOME` is set:

```
$XDG_DATA_HOME/todoat/tasks.db
```

Note: The legacy location `~/.todoat/todoat.db` is also supported for backwards compatibility.

The directory and database file are created automatically on first use.

### Database Schema

The SQLite backend uses two tables:

**task_lists** - Stores task list metadata:
- `id` - Unique identifier (UUID)
- `name` - List name
- `color` - List color (optional)
- `modified` - Last modification timestamp
- `deleted_at` - Deletion timestamp for soft-deleted lists (optional)

**tasks** - Stores tasks:
- `id` - Unique identifier (UUID)
- `list_id` - Reference to parent list
- `summary` - Task title/summary
- `description` - Task description (optional)
- `status` - Task status (NEEDS-ACTION, COMPLETED, IN-PROGRESS, CANCELLED)
- `priority` - Priority level (0-9)
- `due_date` - Due date (optional)
- `start_date` - Start date (optional)
- `created` - Creation timestamp
- `modified` - Last modification timestamp
- `completed` - Completion timestamp (optional)
- `categories` - Comma-separated tags (optional)
- `parent_id` - Parent task ID for subtasks (optional)
- `deleted_at` - Deletion timestamp for soft-deleted items (optional)

### Automatic List Creation

When you reference a list that doesn't exist, it's created automatically:

```bash
# Creates "MyNewList" if it doesn't exist
todoat MyNewList add "First task"
```

### Data Persistence

All changes are immediately persisted to the database. No manual save is required.

### Soft Delete and Trash

Lists are soft-deleted when you use `todoat list delete`. Soft-deleted lists:
- Are moved to "trash" (marked with a `deleted_at` timestamp)
- No longer appear in the regular list view
- Can be restored with `todoat list trash restore`
- Can be permanently deleted with `todoat list trash purge`

```bash
# Delete a list (moves to trash)
todoat list delete "OldProject"

# View trash
todoat list trash

# Restore a list
todoat list trash restore "OldProject"

# Permanently delete
todoat list trash purge "OldProject"
```

### Backup

To backup your tasks, copy the database file:

```bash
cp ~/.local/share/todoat/tasks.db ~/backup/todoat-backup.db
```

### Reset

To start fresh, delete the database:

```bash
rm ~/.local/share/todoat/tasks.db
```

A new database will be created on the next use.

## Nextcloud CalDAV Backend

The Nextcloud backend syncs tasks with a Nextcloud server using the CalDAV protocol. Tasks are stored as VTODO items in your Nextcloud calendars.

### Configuration

Configure the Nextcloud backend using environment variables:

| Variable | Description |
|----------|-------------|
| `TODOAT_NEXTCLOUD_HOST` | Nextcloud server hostname (e.g., `cloud.example.com`) |
| `TODOAT_NEXTCLOUD_USERNAME` | Your Nextcloud username |
| `TODOAT_NEXTCLOUD_PASSWORD` | Your Nextcloud password or app password |

### Example Setup

**Option 1: Using the secure credential manager (recommended)**

```bash
# Store credentials securely in system keyring
todoat credentials set nextcloud youruser --prompt

# Verify credentials are stored
todoat credentials get nextcloud youruser
```

**Option 2: Using environment variables**

```bash
# Set environment variables
export TODOAT_NEXTCLOUD_HOST="cloud.example.com"
export TODOAT_NEXTCLOUD_USERNAME="youruser"
export TODOAT_NEXTCLOUD_PASSWORD="your-app-password"
```

For security, consider using an app password instead of your main Nextcloud password. You can generate an app password in Nextcloud under Settings > Security > Devices & sessions.

See [Credential Management](./commands.md#credential-management) for more details on secure credential storage.

### Features

The Nextcloud backend supports:
- Listing calendars as task lists
- Creating, updating, and deleting tasks (VTODOs)
- Task properties: summary, description, status, priority, due date, start date, categories
- Task completion tracking

### Limitations

Some features work differently with CalDAV compared to the SQLite backend:

| Feature | SQLite | Nextcloud |
|---------|--------|-----------|
| Create lists | ✓ | Not supported |
| Delete lists | ✓ (soft delete) | Not supported |
| Trash/restore lists | ✓ | Not supported |
| Subtasks | ✓ | Not supported |

### Security Options

| Option | Description |
|--------|-------------|
| HTTPS | Used by default for secure connections |
| Allow HTTP | Can be enabled for testing (not recommended) |
| Skip TLS verification | Can be enabled for self-signed certificates |

## Future Backends

Additional backends planned for future releases:
- Todoist API
- Git/Markdown files

---
*Last updated: 2026-01-18*
