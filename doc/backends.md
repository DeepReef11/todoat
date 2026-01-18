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

## Todoist Backend

The Todoist backend syncs tasks with Todoist using the REST API v2. Tasks are stored as Todoist tasks within projects.

### Configuration

Configure the Todoist backend using environment variables or the credential manager:

| Variable | Description |
|----------|-------------|
| `TODOAT_TODOIST_TOKEN` | Todoist API token |

### Getting Your API Token

1. Log in to your Todoist account at [todoist.com](https://todoist.com)
2. Go to Settings > Integrations > Developer
3. Copy your API token

### Example Setup

**Option 1: Using the secure credential manager (recommended)**

```bash
# Store API token securely in system keyring
todoat credentials set todoist myaccount --prompt

# Verify credentials are stored
todoat credentials get todoist myaccount
```

**Option 2: Using environment variables**

```bash
# Set environment variable
export TODOAT_TODOIST_TOKEN="your-api-token"
```

For security, consider using the credential manager rather than environment variables, especially on shared systems.

### Features

The Todoist backend supports:
- Listing projects as task lists
- Creating, updating, and deleting projects
- Creating, updating, and deleting tasks
- Task properties: summary (content), description, priority, due date, labels
- Task completion tracking
- Rate limiting with automatic retry

### Todoist Mapping

| todoat Concept | Todoist Concept |
|----------------|-----------------|
| List | Project |
| Task | Task |
| Tags/Categories | Labels |
| Priority (1-9) | Priority (1-4, mapped) |
| Status | Completed flag |

### Priority Mapping

Todoist uses priority 1-4 (4 being highest), while todoat uses 1-9 (1 being highest). The priorities are mapped as follows:

| todoat Priority | Todoist Priority |
|-----------------|------------------|
| 1 | 4 (urgent) |
| 2-3 | 3 (high) |
| 4-5 | 2 (medium) |
| 6-9 | 1 (low) |

### Limitations

Some features work differently with Todoist compared to the SQLite backend:

| Feature | SQLite | Todoist |
|---------|--------|---------|
| Soft delete lists | Yes | No (permanent) |
| Trash/restore lists | Yes | No |
| Subtasks | Yes | Yes |
| Start date | Yes | No |
| Custom status values | Yes | No (completed/not completed only) |

### Rate Limiting

The Todoist API has rate limits. todoat handles this automatically by:
- Detecting 429 (Too Many Requests) responses
- Waiting and retrying up to the configured number of times
- Displaying an error if rate limit persists

## File Backend

The File backend stores tasks in plain text markdown files. This is a simple, portable storage option that works without Git or databases.

### How It Works

The File backend:
1. Stores tasks in a markdown file (default: `tasks.txt`)
2. Uses `##` headers for task lists (sections)
3. Uses markdown checkbox syntax for tasks
4. Supports subtasks via indentation

### Configuration

The file backend can be configured in your config file:

```yaml
backends:
  file:
    enabled: true
    path: ~/tasks.txt
```

Or specify a file path when using the backend.

### File Format

Tasks are stored as markdown checkbox lists within sections:

```markdown
# Tasks

## Work

- [ ] Complete report !1 @2026-01-31 #urgent
  - [ ] Gather data
  - [ ] Write summary
- [x] Send email #communication

## Personal

- [ ] Buy groceries
- [~] Read book
```

### Status Characters

| Character | Status |
|-----------|--------|
| `[ ]` | TODO (needs action) |
| `[x]` or `[X]` | Completed |
| `[~]` | In progress |
| `[-]` | Cancelled |

### Task Metadata

Tasks support inline metadata with the same syntax as the Git backend:

| Syntax | Meaning |
|--------|---------|
| `!1` - `!9` | Priority (1 is highest) |
| `@2026-01-31` | Due date (YYYY-MM-DD format) |
| `#tag` | Category/tag |

### Subtasks

Subtasks are created using 2-space indentation:

```markdown
- [ ] Main task
  - [ ] Subtask 1
    - [ ] Sub-subtask
  - [ ] Subtask 2
```

### Features

The File backend supports:
- Creating, updating, and deleting tasks
- Task properties: summary, status, priority, due date, categories
- Subtasks via indentation
- Multiple lists (as `##` sections)

### Limitations

| Feature | SQLite | File |
|---------|--------|------|
| Soft delete lists | Yes | No |
| Trash/restore lists | Yes | No |
| Start date | Yes | No |
| Task descriptions | Yes | No |

### Use Cases

The File backend is ideal for:
- Simple, human-readable task storage
- Sharing tasks via file sync (Dropbox, Syncthing, etc.)
- Editing tasks directly in a text editor
- Portable task files that work without special software

## Google Tasks Backend

The Google Tasks backend syncs tasks with Google Tasks using the Google Tasks API v1. Tasks are stored as Google Tasks within task lists.

### Configuration

Configure the Google Tasks backend using environment variables:

| Variable | Description |
|----------|-------------|
| `TODOAT_GOOGLE_ACCESS_TOKEN` | Google OAuth2 access token |
| `TODOAT_GOOGLE_REFRESH_TOKEN` | Google OAuth2 refresh token (optional, for automatic token refresh) |
| `TODOAT_GOOGLE_CLIENT_ID` | Google OAuth2 client ID (optional, required for token refresh) |
| `TODOAT_GOOGLE_CLIENT_SECRET` | Google OAuth2 client secret (optional, required for token refresh) |

### Getting API Credentials

1. Go to the [Google Cloud Console](https://console.cloud.google.com)
2. Create a new project or select an existing one
3. Enable the Google Tasks API
4. Create OAuth 2.0 credentials (Desktop application type)
5. Use the OAuth flow to obtain access and refresh tokens

### Example Setup

```bash
# Set environment variables
export TODOAT_GOOGLE_ACCESS_TOKEN="your-access-token"
export TODOAT_GOOGLE_REFRESH_TOKEN="your-refresh-token"
export TODOAT_GOOGLE_CLIENT_ID="your-client-id"
export TODOAT_GOOGLE_CLIENT_SECRET="your-client-secret"
```

### Features

The Google Tasks backend supports:
- Listing task lists
- Creating, updating, and deleting task lists
- Creating, updating, and deleting tasks
- Task properties: summary (title), notes (description), status, due date
- Subtasks via parent task ID
- Automatic OAuth2 token refresh

### Google Tasks Mapping

| todoat Concept | Google Tasks Concept |
|----------------|----------------------|
| List | Task List |
| Task | Task |
| Summary | Title |
| Description | Notes |
| Status | Status (needsAction/completed) |
| Due Date | Due (RFC3339 format) |
| Subtask | Task with parent |

### Status Mapping

Google Tasks only supports two status values:

| todoat Status | Google Tasks Status |
|---------------|---------------------|
| TODO | needsAction |
| IN-PROGRESS | needsAction |
| DONE | completed |
| CANCELLED | completed |

### Limitations

Some features work differently with Google Tasks compared to the SQLite backend:

| Feature | SQLite | Google Tasks |
|---------|--------|--------------|
| Soft delete lists | Yes | No (permanent) |
| Trash/restore lists | Yes | No |
| Priority | Yes | No |
| Start date | Yes | No |
| Tags/categories | Yes | No |
| Custom status values | Yes | No (only completed/needsAction) |

### Notes

- Google Tasks API requires OAuth2 authentication
- Access tokens expire; provide a refresh token for automatic renewal
- Rate limits apply to the Google Tasks API
- Task descriptions are stored in the "notes" field

## Git/Markdown Backend

The Git backend stores tasks in markdown files within Git repositories. This allows you to manage tasks alongside your code and use Git for version control and collaboration.

### How It Works

The Git backend:
1. Detects if you're inside a Git repository
2. Looks for a markdown file with the `<!-- todoat:enabled -->` marker
3. Reads and writes tasks as markdown checkbox lists
4. Optionally auto-commits changes

### Setup

Create a markdown file in your Git repository with the todoat marker:

```markdown
<!-- todoat:enabled -->

## My Tasks

- [ ] First task
- [x] Completed task
```

The file must contain `<!-- todoat:enabled -->` for todoat to recognize it.

### Default File Locations

todoat searches for these files in order:
1. `TODO.md`
2. `todo.md`
3. `.todoat.md`

The first file found with the `<!-- todoat:enabled -->` marker is used.

### Markdown Format

Tasks are stored as standard markdown checkbox lists:

```markdown
## Project Tasks

- [ ] Incomplete task
- [x] Completed task
- [~] In progress task
- [-] Cancelled task
  - [ ] Subtask (indented with 2 spaces)
```

### Status Characters

| Character | Status |
|-----------|--------|
| `[ ]` | TODO (needs action) |
| `[x]` | Completed |
| `[~]` | In progress |
| `[-]` | Cancelled |

### Task Metadata

Tasks can include inline metadata:

```markdown
- [ ] Task summary !1 @2026-01-31 #work #urgent
```

| Syntax | Meaning |
|--------|---------|
| `!1` - `!9` | Priority (1 is highest) |
| `@2026-01-31` | Due date (YYYY-MM-DD format) |
| `#tag` | Category/tag |

### Subtasks

Subtasks are created using indentation (2 spaces per level):

```markdown
## Work

- [ ] Main project
  - [ ] Backend tasks
    - [ ] API design
    - [ ] Database schema
  - [ ] Frontend tasks
```

### Features

The Git/Markdown backend supports:
- Creating, updating, and deleting tasks
- Task properties: summary, status, priority, due date, categories
- Subtasks via indentation
- Multiple lists (as `##` sections)
- Optional auto-commit of changes

### Limitations

| Feature | SQLite | Git/Markdown |
|---------|--------|--------------|
| Soft delete lists | Yes | No |
| Trash/restore lists | Yes | No |
| Start date | Yes | No |
| Task descriptions | Yes | No |

### Auto-Commit

When auto-commit is enabled, changes are automatically committed to Git with descriptive messages:
- `todoat: add task 'Task name'`
- `todoat: update task 'Task name'`
- `todoat: delete task 'Task name'`

---
*Last updated: 2026-01-18*
