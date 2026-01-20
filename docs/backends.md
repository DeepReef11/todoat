# Backend Configuration

todoat supports multiple storage backends. This guide covers configuring each backend type.

## Available Backends

| Backend | Type | Description |
|---------|------|-------------|
| Nextcloud | `nextcloud` | CalDAV-based cloud storage |
| Todoist | `todoist` | Todoist cloud service |
| SQLite | `sqlite` | Local database storage |
| Git | `git` | Markdown files in Git repositories |

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

## SQLite (Local)

### Configuration

```yaml
backends:
  sqlite:
    type: sqlite
    enabled: true
    db_path: ""  # Empty = default location
```

Default database location: `~/.local/share/todoat/tasks.db`

### Custom Location

```yaml
backends:
  sqlite:
    type: sqlite
    enabled: true
    db_path: "~/my-tasks/tasks.db"
```

### Path Expansion

Paths support:
- `~` - Home directory
- `$HOME` - Environment variable

## Git (Markdown)

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

## Backend Selection

### Selection Priority

1. Sync local backend (when sync enabled)
2. Auto-detected backend (when enabled)
3. `default_backend` from config
4. First backend in `backend_priority`
5. First enabled backend

### Default Backend

```yaml
default_backend: nextcloud
```

### Backend Priority

```yaml
backend_priority:
  - git        # Try Git first
  - nextcloud  # Then Nextcloud
  - sqlite     # Finally SQLite
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

backend_priority:
  - work
  - home
  - local
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

backend_priority:
  - git
  - nextcloud
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

## See Also

- [Getting Started](getting-started.md) - Initial setup
- [Synchronization](sync.md) - Syncing between backends
- [Configuration Reference](../dev-doc/CONFIGURATION.md) - Full configuration details
