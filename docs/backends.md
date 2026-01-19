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

Use specific backend:

```bash
todoat --backend personal-nc MyList
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

1. `--backend` flag (highest priority)
2. Sync local backend (when sync enabled)
3. Auto-detected backend (when enabled)
4. `default_backend` from config
5. First backend in `backend_priority`
6. First enabled backend

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
todoat --list-backends

# Check auto-detection
todoat --detect-backend
```

## Using Multiple Backends

### Switching Backends

```bash
# Use specific backend for one command
todoat --backend sqlite MyList

# Use default backend
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
