# Backend Configuration

todoat supports multiple storage backends. This guide covers configuring each backend type.

## Available Backends

| Backend | Type | CLI Support | Description |
|---------|------|-------------|-------------|
| SQLite | `sqlite` | ✅ Yes | Local database storage (default) |
| Nextcloud | `nextcloud` | ✅ Yes | CalDAV-based cloud storage |
| Todoist | `todoist` | ✅ Yes | Todoist cloud service |
| Google Tasks | `google` | ✅ Yes | Google Tasks cloud service |
| Microsoft To Do | `mstodo` | ✅ Yes | Microsoft Graph API cloud service |
| Git | `git` | ✅ Yes | Markdown files in Git repositories |
| File | `file` | ✅ Yes | Plain file-based storage |

## Nextcloud (CalDAV)

### Configuration

```yaml
backends:
  nextcloud:
    type: nextcloud
    enabled: true
    host: "nextcloud.example.com"
    username: "myuser"
```

### Setting Credentials

Store your password securely in the system keyring:

```bash
todoat credentials set nextcloud myuser --prompt
# Enter password when prompted
```

### Updating Credentials

Update an existing password (e.g., after password rotation):

```bash
todoat credentials update nextcloud myuser --prompt
# Enter new password when prompted
```

The update command verifies the credential exists before prompting for the new password. Use `credentials set` if the credential doesn't exist yet.

### HTTPS Options

For self-signed certificates or development servers:

```yaml
backends:
  nextcloud-dev:
    type: nextcloud
    enabled: true
    host: "localhost:8080"
    username: "admin"
    insecure_skip_verify: true    # Accept self-signed certs
    suppress_ssl_warning: true    # Suppress security warning
    allow_http: true              # Allow HTTP (not HTTPS)
    suppress_http_warning: true   # Suppress HTTP warning
```

### Multiple Nextcloud Accounts

```yaml
backends:
  work-nc:
    type: nextcloud
    enabled: true
    host: "nextcloud.work.com"
    username: "workuser"

  personal-nc:
    type: nextcloud
    enabled: true
    host: "nextcloud.home.local"
    username: "me"

default_backend: work-nc
```

To switch between backends, update `default_backend` in your config:

```bash
todoat config set default_backend personal-nc
```

## Todoist

### Configuration

```yaml
backends:
  todoist:
    type: todoist
    enabled: true
    username: "token"
```

### Setting API Token

1. Get your API token from Todoist Settings > Integrations > Developer
2. Store it securely:

```bash
todoat credentials set todoist token --prompt
# Paste your API token when prompted
```

## Google Tasks

### Configuration

```yaml
backends:
  google:
    type: google
    enabled: true
```

### OAuth2 Setup

Google Tasks requires OAuth2 authentication. You'll need to create credentials in the Google Cloud Console:

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the **Google Tasks API** in APIs & Services > Library
4. Go to APIs & Services > Credentials
5. Click "Create Credentials" > "OAuth client ID"
6. Select "Desktop app" as the application type
7. Download the credentials JSON file

### Authentication

Store your OAuth2 tokens securely:

```bash
# Set up tokens via environment variables
export TODOAT_GOOGLE_ACCESS_TOKEN="your-access-token"
export TODOAT_GOOGLE_REFRESH_TOKEN="your-refresh-token"
export TODOAT_GOOGLE_CLIENT_ID="your-client-id"
export TODOAT_GOOGLE_CLIENT_SECRET="your-client-secret"
```

The backend automatically refreshes expired access tokens using the refresh token.

### Usage Examples

```bash
# List all task lists
todoat -b google list

# View tasks in a specific list
todoat -b google "My Tasks"

# Add a task to Google Tasks
todoat -b google "My Tasks" add "Buy groceries"

# Add a task with due date
todoat -b google "My Tasks" add "Submit report" --due-date tomorrow
```

### Supported Features

| Feature | Supported |
|---------|-----------|
| Task lists (CRUD) | Yes |
| Tasks (CRUD) | Yes |
| Subtasks | Yes |
| Due dates | Yes |
| Task completion | Yes |
| Notes/Description | Yes |

### Limitations

- **No trash/restore**: Google Tasks permanently deletes tasks and lists (no trash recovery)
- **Status mapping**: Google Tasks only supports "needsAction" and "completed" statuses. IN-PROGRESS and CANCELLED tasks are mapped to "completed"
- **No priorities**: Google Tasks does not support task priorities
- **No tags/categories**: Google Tasks does not support labels or categories
- **No start dates**: Only due dates are supported
- **No recurrence**: Recurring tasks are not supported by the API

## Microsoft To Do

### Configuration

```yaml
backends:
  mstodo:
    type: mstodo
    enabled: true
```

### OAuth2 Setup

Microsoft To Do requires OAuth2 authentication via Microsoft Graph API. You'll need to create an app registration in the Azure portal:

1. Go to [Azure Portal](https://portal.azure.com/)
2. Navigate to Azure Active Directory > App registrations
3. Click "New registration"
4. Enter a name for your application
5. Select "Accounts in any organizational directory and personal Microsoft accounts"
6. Set redirect URI to `http://localhost` (or your preferred redirect)
7. Click "Register"
8. Note the **Application (client) ID**
9. Go to "Certificates & secrets" > "New client secret"
10. Create a secret and note its **Value** (this is your client secret)
11. Go to "API permissions" > "Add a permission"
12. Select "Microsoft Graph" > "Delegated permissions"
13. Add the following permissions:
    - `Tasks.ReadWrite`
    - `User.Read`
14. Grant admin consent if required by your organization

### Authentication

Store your OAuth2 tokens securely via environment variables or the system keyring:

```bash
# Set up tokens via environment variables
export TODOAT_MSTODO_ACCESS_TOKEN="your-access-token"
export TODOAT_MSTODO_REFRESH_TOKEN="your-refresh-token"
export TODOAT_MSTODO_CLIENT_ID="your-client-id"
export TODOAT_MSTODO_CLIENT_SECRET="your-client-secret"

# Or store the access token in the system keyring
todoat credentials set mstodo token --prompt
# Paste your access token when prompted
```

The backend automatically refreshes expired access tokens using the refresh token.

### Usage Examples

```bash
# List all task lists
todoat -b mstodo list

# View tasks in a specific list
todoat -b mstodo "My Tasks"

# Add a task to Microsoft To Do
todoat -b mstodo "My Tasks" add "Buy groceries"

# Add a task with due date
todoat -b mstodo "My Tasks" add "Submit report" --due-date tomorrow

# Add a high-priority task
todoat -b mstodo "My Tasks" add "Urgent meeting prep" --priority 1
```

### Supported Features

| Feature | Supported |
|---------|-----------|
| Task lists (CRUD) | Yes |
| Tasks (CRUD) | Yes |
| Due dates | Yes |
| Task completion | Yes |
| Notes/Description | Yes |
| Priority/Importance | Yes |
| In-progress status | Yes |

### Priority/Importance Mapping

Microsoft To Do uses "importance" with three levels (low, normal, high). todoat maps these to numeric priorities:

| todoat Priority | Microsoft Importance |
|-----------------|---------------------|
| 1-3 (high) | high |
| 4-6 (medium) | normal |
| 7-9 (low) | low |

### Status Mapping

| todoat Status | Microsoft Status |
|---------------|------------------|
| NEEDS-ACTION | notStarted |
| IN-PROGRESS | inProgress |
| COMPLETED | completed |
| CANCELLED | completed |

### Limitations

- **No trash/restore**: Microsoft To Do permanently deletes tasks and lists (no trash recovery)
- **No tags/categories**: Microsoft To Do does not support labels or categories in the API
- **No start dates**: Only due dates are supported
- **No recurrence**: Recurring tasks are not supported via the API
- **No subtask hierarchy**: Checklist items exist but are not exposed as subtasks

## SQLite (Local)

### Configuration

```yaml
backends:
  sqlite:
    type: sqlite
    enabled: true
    path: ""  # Empty = default location
```

Default database location: `~/.local/share/todoat/tasks.db`

### Custom Location

```yaml
backends:
  sqlite:
    type: sqlite
    enabled: true
    path: "~/my-tasks/tasks.db"
```

### Path Expansion

Paths support:
- `~` - Home directory
- `$HOME` - Environment variable

## Git (Markdown)

The Git backend stores tasks as markdown files in Git repositories, enabling version-controlled task management.

### Configuration

```yaml
backends:
  git:
    type: git
    enabled: true
    auto_detect: true
    file: "TODO.md"
    fallback_files:
      - "todo.md"
      - ".todoat.md"
    auto_commit: false
```

### Setting Up a Git Repository

1. Create or use existing Git repository
2. Create markdown file with marker:

```bash
echo "<!-- todoat:enabled -->" > TODO.md
echo "" >> TODO.md
echo "## Tasks" >> TODO.md
```

3. Enable auto-detection in config

### Auto-Detection

When `auto_detect: true`, todoat automatically uses Git backend when:
- Current directory is in a Git repository
- A task file with `<!-- todoat:enabled -->` marker exists

### Task File Format

```markdown
<!-- todoat:enabled -->

## My Tasks

- [ ] Task one
- [x] Completed task
- [>] In progress task
```

## File (Plain Text)

The File backend stores tasks in plain text files, providing a lightweight file-based storage option without Git dependency.

### Configuration

```yaml
backends:
  file:
    type: file
    enabled: true
    path: "~/tasks.txt"
```

### Usage

```bash
# Use file backend
todoat -b file MyList add "New task"

# List tasks from file backend
todoat -b file MyList
```

### File Format

Tasks are stored in a text file with sections as lists:

```markdown
# Tasks

## Work

- [ ] Pending task
- [x] Completed task
- [>] In progress task

## Personal

- [ ] Another task
```

### Features

- Lightweight file-based storage
- Sections treated as task lists
- Indented tasks parsed as subtasks
- Supports priority, dates, status, and tags
- No Git dependency (unlike Git backend)

## Backend Selection

### Selection Priority

1. Sync local backend (when sync enabled)
2. Auto-detected backend (when enabled)
3. `default_backend` from config
4. First enabled backend

### Default Backend

```yaml
default_backend: nextcloud
```

### Auto-Detection

```yaml
auto_detect_backend: true
```

When enabled, Git backend is auto-detected when in a repository with a marked task file.

## Listing Backends

```bash
# Show all configured backends
todoat config get backends

# Check auto-detection
todoat --detect-backend
```

## Using Multiple Backends

### Per-Command Backend Selection

Use the `--backend` flag (or `-b`) to select a specific backend for any command:

```bash
# View tasks from Todoist backend
todoat -b todoist MyList

# Add a task to SQLite backend
todoat -b sqlite Work add "Local task"

# List all lists from Nextcloud
todoat -b nextcloud list
```

This overrides the default backend for that command only, without changing your configuration.

### Switching Default Backend

To permanently change your default backend:

```bash
# Change default backend
todoat config set default_backend sqlite

# View tasks using current default backend
todoat MyList
```

### Environment Variables

Set credentials via environment:

```bash
# Nextcloud
export TODOAT_NEXTCLOUD_USERNAME="myuser"
export TODOAT_NEXTCLOUD_PASSWORD="secret"

# Todoist
export TODOAT_TODOIST_TOKEN="api-token-here"
```

## Managing Credentials

### Credential Commands

| Command | Description |
|---------|-------------|
| `credentials set <backend> <user> --prompt` | Store new credential |
| `credentials update <backend> <user> --prompt` | Update existing credential |
| `credentials get <backend> <user>` | Check credential status |
| `credentials delete <backend> <user>` | Remove credential |
| `credentials list` | Show all configured credentials |

### Examples

```bash
# Store credential for Nextcloud
todoat credentials set nextcloud myuser --prompt

# Update after password change
todoat credentials update nextcloud myuser --prompt

# Check if credentials are configured
todoat credentials get nextcloud myuser

# View all credentials status
todoat credentials list

# Remove credential
todoat credentials delete nextcloud myuser
```

## Credential Storage Priority

When resolving credentials:

1. System keyring (most secure, recommended)
2. Environment variables
3. Config file (not recommended)

## Example Configurations

### Home and Work Setup

```yaml
backends:
  work:
    type: nextcloud
    enabled: true
    host: "nextcloud.work.com"
    username: "workuser"

  home:
    type: todoist
    enabled: true
    username: "token"

  local:
    type: sqlite
    enabled: true

default_backend: work
```

### Developer Workflow

```yaml
backends:
  git:
    type: git
    enabled: true
    auto_detect: true
    file: "TODO.md"
    auto_commit: true

  nextcloud:
    type: nextcloud
    enabled: true
    host: "nextcloud.example.com"
    username: "dev"

auto_detect_backend: true
```

### Offline-First Setup

```yaml
backends:
  sqlite:
    type: sqlite
    enabled: true

  nextcloud:
    type: nextcloud
    enabled: true
    host: "nextcloud.example.com"
    username: "user"

default_backend: nextcloud

sync:
  enabled: true
  local_backend: sqlite
  offline_mode: auto
```

## Migrating Between Backends

Use the `migrate` command to move tasks from one backend to another while preserving metadata.

### Basic Migration

```bash
# Migrate from SQLite to Nextcloud
todoat migrate --from sqlite --to nextcloud

# Migrate from Nextcloud to Todoist
todoat migrate --from nextcloud --to todoist
```

### Migrate Specific List

```bash
# Migrate only one list
todoat migrate --from sqlite --to nextcloud --list "Work Tasks"
```

### Dry Run

Preview what would be migrated without making changes:

```bash
todoat migrate --from sqlite --to nextcloud --dry-run
```

### View Target Backend

Check existing tasks in the target before migrating:

```bash
todoat migrate --target-info nextcloud --list Work
```

### What Gets Migrated

Migration preserves:
- Task summary and description
- Priority and status
- Due dates and start dates
- Tags/categories
- Parent-child relationships (task hierarchy)
- Recurrence rules

### Migration Notes

- UIDs are preserved where possible
- Status values are mapped between backends (e.g., IN-PROGRESS may become different values)
- Large lists are migrated in batches with progress indicators
- Use `--dry-run` first to verify the migration plan

## See Also

- [Getting Started](../tutorials/getting-started.md) - Initial setup
- [Synchronization](../how-to/sync.md) - Syncing between backends
- [Configuration](../reference/configuration.md) - Configuration reference
