# Backends

todoat supports multiple storage backends. Currently, SQLite is the default and only implemented backend.

## SQLite Backend

The SQLite backend stores tasks locally in a database file.

### Location

By default, the database is stored at:

```
~/.todoat/todoat.db
```

The directory and database file are created automatically on first use.

### Database Schema

The SQLite backend uses two tables:

**task_lists** - Stores task list metadata:
- `id` - Unique identifier (UUID)
- `name` - List name
- `color` - List color (optional)
- `modified` - Last modification timestamp

**tasks** - Stores tasks:
- `id` - Unique identifier (UUID)
- `list_id` - Reference to parent list
- `summary` - Task title/summary
- `description` - Task description (optional)
- `status` - Task status (NEEDS-ACTION, COMPLETED, IN-PROGRESS, CANCELLED)
- `priority` - Priority level (0-9)
- `due_date` - Due date (optional)
- `created` - Creation timestamp
- `modified` - Last modification timestamp
- `parent_id` - Parent task ID for subtasks (optional)

### Automatic List Creation

When you reference a list that doesn't exist, it's created automatically:

```bash
# Creates "MyNewList" if it doesn't exist
todoat MyNewList add "First task"
```

### Data Persistence

All changes are immediately persisted to the database. No manual save is required.

### Backup

To backup your tasks, copy the database file:

```bash
cp ~/.todoat/todoat.db ~/backup/todoat-backup.db
```

### Reset

To start fresh, delete the database:

```bash
rm ~/.todoat/todoat.db
```

A new database will be created on the next use.

## Future Backends

Additional backends are planned for future releases:
- Nextcloud CalDAV
- Todoist API
- Git/Markdown files

---
*Last updated: 2026-01-18*
