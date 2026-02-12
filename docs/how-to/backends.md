# How to Set Up Backends

This guide covers configuring each supported backend for todoat.

## Quick Reference

| Backend | Type | Best For |
|---------|------|----------|
| SQLite | `sqlite` | Local-only use, offline-first setups |
| Nextcloud | `nextcloud` | Self-hosted cloud sync via CalDAV |
| Todoist | `todoist` | Todoist cloud service |
| Google Tasks | `google` | Google ecosystem integration |
| Microsoft To Do | `mstodo` | Microsoft ecosystem integration |
| Git | `git` | Version-controlled tasks in repositories |
| File | `file` | Lightweight plain-text storage |

## SQLite (Default)

SQLite is the default backend and requires no setup. It stores tasks in a local database.

```yaml
backends:
  sqlite:
    type: sqlite
    enabled: true
    # path: "~/.local/share/todoat/tasks.db"  # Optional: custom path
```

To use a custom database location:

```bash
todoat config set backends.sqlite.path "~/my-tasks/tasks.db"
```

## Nextcloud (CalDAV)

### Basic Setup

1. Add to your config:

   ```yaml
   backends:
     nextcloud:
       type: nextcloud
       enabled: true
       host: "nextcloud.example.com"
       username: "myuser"
   ```

2. Store your password:

   ```bash
   todoat credentials set nextcloud myuser --prompt
   ```

3. Verify it works:

   ```bash
   todoat -b nextcloud list
   ```

### Self-Signed Certificates

For development servers or self-signed certs:

```yaml
backends:
  nextcloud:
    type: nextcloud
    enabled: true
    host: "nextcloud.local"
    username: "admin"
    insecure_skip_verify: true    # Accept self-signed certs
    allow_http: true              # Allow HTTP connections
```

### Sharing Lists

Nextcloud supports sharing task lists with other users. See [List Management - Sharing](list-management.md#sharing-lists-nextcloud) for details.

### Public Links

Generate public read-only URLs for task lists:

```bash
todoat list publish "Work Tasks"
```

See [List Management - Public Links](list-management.md#public-links-nextcloud) for details.

### Calendar Subscriptions

Subscribe to external calendar feeds as read-only task lists:

```bash
todoat list subscribe "https://example.com/calendar/ical"
```

See [List Management - Calendar Subscriptions](list-management.md#calendar-subscriptions-nextcloud) for details.

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

Switch between them:

```bash
# Temporarily use a different backend
todoat -b personal-nc list

# Change the default
todoat config set default_backend personal-nc
```

## Todoist

1. Get your API token from [Todoist Settings > Integrations > Developer](https://todoist.com/prefs/integrations)

2. Add to your config:

   ```yaml
   backends:
     todoist:
       type: todoist
       enabled: true
   ```

3. Store the API token:

   ```bash
   todoat credentials set todoist token --prompt
   # Paste your API token when prompted
   ```

## Google Tasks

### OAuth2 Setup

Google Tasks requires OAuth2 authentication:

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the **Google Tasks API** in APIs & Services > Library
4. Go to APIs & Services > Credentials
5. Click "Create Credentials" > "OAuth client ID"
6. Select "Desktop app" as the application type
7. Download the credentials JSON file

### Configuration

```yaml
backends:
  google:
    type: google
    enabled: true
```

### Authentication

Set your OAuth2 tokens via environment variables:

```bash
export TODOAT_GOOGLE_ACCESS_TOKEN="your-access-token"
export TODOAT_GOOGLE_REFRESH_TOKEN="your-refresh-token"
export TODOAT_GOOGLE_CLIENT_ID="your-client-id"
export TODOAT_GOOGLE_CLIENT_SECRET="your-client-secret"
```

The backend automatically refreshes expired access tokens using the refresh token.

### Usage

```bash
todoat -b google list
todoat -b google "My Tasks"
todoat -b google "My Tasks" add "Buy groceries"
todoat -b google "My Tasks" add "Submit report" --due-date tomorrow
```

### Limitations

- No task priorities
- No tags/categories
- No start dates (only due dates)
- No recurring tasks
- No trash/restore (permanent delete)
- Status limited to "needsAction" and "completed" (IN-PROGRESS maps to completed)

## Microsoft To Do

### OAuth2 Setup

Microsoft To Do uses the Microsoft Graph API:

1. Go to [Azure Portal](https://portal.azure.com/)
2. Navigate to Azure Active Directory > App registrations
3. Click "New registration"
4. Enter a name and select "Accounts in any organizational directory and personal Microsoft accounts"
5. Set redirect URI to `http://localhost`
6. Note the **Application (client) ID**
7. Go to "Certificates & secrets" > "New client secret" and note its value
8. Go to "API permissions" > "Add a permission"
9. Select "Microsoft Graph" > "Delegated permissions"
10. Add `Tasks.ReadWrite` and `User.Read`
11. Grant admin consent if required

### Configuration

```yaml
backends:
  mstodo:
    type: mstodo
    enabled: true
```

### Authentication

```bash
export TODOAT_MSTODO_ACCESS_TOKEN="your-access-token"
export TODOAT_MSTODO_REFRESH_TOKEN="your-refresh-token"
export TODOAT_MSTODO_CLIENT_ID="your-client-id"
export TODOAT_MSTODO_CLIENT_SECRET="your-client-secret"
```

Or store the access token in the system keyring:

```bash
todoat credentials set mstodo token --prompt
```

### Usage

```bash
todoat -b mstodo list
todoat -b mstodo "My Tasks"
todoat -b mstodo "My Tasks" add "Buy groceries"
todoat -b mstodo "My Tasks" add "Urgent meeting" --priority 1
```

### Priority Mapping

| todoat Priority | Microsoft Importance |
|-----------------|---------------------|
| 1-3 (high) | high |
| 4-6 (medium) | normal |
| 7-9 (low) | low |

### Limitations

- No tags/categories
- No start dates (only due dates)
- No recurring tasks via API
- No subtask hierarchy (checklist items are separate)
- No trash/restore (permanent delete)

## Git (Markdown)

The Git backend stores tasks as markdown files in Git repositories.

### Configuration

```yaml
backends:
  git:
    type: git
    enabled: true
    file: "TODO.md"
    auto_commit: false
```

### Setup

1. Create a markdown file with the todoat marker in your repository:

   ```bash
   echo "<!-- todoat:enabled -->" > TODO.md
   echo "" >> TODO.md
   echo "## Tasks" >> TODO.md
   ```

2. Use the git backend:

   ```bash
   todoat -b git "Tasks" add "New feature"
   ```

### Task File Format

```markdown
<!-- todoat:enabled -->

## My Tasks

- [ ] Pending task
- [x] Completed task
- [>] In progress task
```

## File (Plain Text)

The File backend stores tasks in a plain text file without Git dependency.

### Configuration

```yaml
backends:
  file:
    type: file
    enabled: true
    path: "~/tasks.md"
```

### Usage

```bash
todoat -b file "Work" add "New task"
todoat -b file "Work"
```

### File Format

Sections are treated as task lists:

```markdown
## Work

- [ ] Pending task
- [x] Completed task
- [>] In progress task

## Personal

- [ ] Another task
```

Indented tasks are parsed as subtasks.

## Selecting a Backend

### Per-Command Selection

Use `-b` or `--backend` to override the default for a single command:

```bash
todoat -b todoist MyList
todoat -b sqlite Work add "Local task"
```

### Default Backend

Set your preferred default:

```bash
todoat config set default_backend sqlite
```

### Auto-Detection

Enable automatic backend detection (uses Git backend when in a repository with a marked task file):

```bash
todoat config set auto_detect_backend true
```

### Selection Priority

When determining which backend to use:

1. Explicit `-b` flag (highest priority)
2. Sync local backend (when sync is enabled)
3. Auto-detected backend (when enabled)
4. `default_backend` from config
5. First enabled backend

### Check Detection

```bash
todoat --detect-backend
```

## See Also

- [Getting Started](../tutorials/getting-started.md) - First-time setup
- [Credentials](credentials.md) - Credential management details
- [Synchronization](sync.md) - Syncing between backends
- [Migration](migration.md) - Moving tasks between backends
- [Configuration Reference](../reference/configuration.md) - All config options
