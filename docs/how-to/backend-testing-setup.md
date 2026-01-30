# Backend Testing Setup

Quick setup guide for running integration tests against each backend.

## Quick Reference

| Backend | Env Vars | Get Credentials |
|---------|----------|-----------------|
| [Todoist](#todoist) | `TODOAT_TODOIST_TOKEN` | [todoist.com/app/settings/integrations/developer](https://todoist.com/app/settings/integrations/developer) |
| [Nextcloud](#nextcloud) | `TODOAT_NEXTCLOUD_HOST`, `TODOAT_NEXTCLOUD_USERNAME`, `TODOAT_NEXTCLOUD_PASSWORD` | Docker or self-hosted |
| [Google Tasks](#google-tasks) | `TODOAT_GOOGLE_CLIENT_ID`, `TODOAT_GOOGLE_CLIENT_SECRET`, `TODOAT_GOOGLE_REFRESH_TOKEN` | [console.cloud.google.com](https://console.cloud.google.com/) |
| [Microsoft To Do](#microsoft-to-do) | `TODOAT_MSTODO_CLIENT_ID`, `TODOAT_MSTODO_CLIENT_SECRET`, `TODOAT_MSTODO_REFRESH_TOKEN` | [portal.azure.com](https://portal.azure.com/) |

## Using .env File

Create `.env` in project root (git-ignored):

```bash
# Todoist
TODOAT_TODOIST_TOKEN="your-api-token"

# Nextcloud
TODOAT_NEXTCLOUD_HOST=localhost:8080
TODOAT_NEXTCLOUD_USERNAME=admin
TODOAT_NEXTCLOUD_PASSWORD=admin123
```

Run tests with:

```bash
set -a && source .env && set +a && go test -tags=integration -v ./backend/todoist
```

---

## Todoist

### Get API Token

1. Go to [todoist.com/app/settings/integrations/developer](https://todoist.com/app/settings/integrations/developer)
2. Copy the API token

### Environment Variables

```bash
export TODOAT_TODOIST_TOKEN="your-api-token"
```

### Run Tests

```bash
go test -tags=integration -v ./backend/todoist
# or
make test-todoist
```

### CI Setup (GitHub Actions)

Add `TODOAT_TODOIST_TOKEN` as a repository secret.

---

## Nextcloud

### Option 1: Docker (Recommended for Testing)

```bash
# Start Nextcloud
make docker-up

# Or manually:
docker compose up -d
```

Default credentials: `admin` / `admin123`

### Option 2: Remote Instance

Use your own Nextcloud server.

### Environment Variables

```bash
# Docker setup
export TODOAT_NEXTCLOUD_HOST=localhost:8080
export TODOAT_NEXTCLOUD_USERNAME=admin
export TODOAT_NEXTCLOUD_PASSWORD=admin123

# Remote setup (from container, use host IP)
export TODOAT_NEXTCLOUD_HOST=172.17.0.1:8080
```

### Run Tests

```bash
go test -tags=integration -v ./backend/nextcloud
# or
make test-nextcloud
```

### Install Tasks App (if CRUD tests skip)

```bash
docker exec -u www-data todoat-nextcloud-1 php occ app:install tasks
```

---

## Google Tasks

### Get Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create project or select existing
3. Enable **Google Tasks API**
4. Create OAuth 2.0 credentials (Desktop app)
5. Run OAuth flow to get refresh token

### Environment Variables

```bash
export TODOAT_GOOGLE_CLIENT_ID="xxx.apps.googleusercontent.com"
export TODOAT_GOOGLE_CLIENT_SECRET="xxx"
export TODOAT_GOOGLE_REFRESH_TOKEN="xxx"
```

### Run Tests

```bash
go test -tags=integration -v ./backend/google
```

---

## Microsoft To Do

### Get Credentials

1. Go to [Azure Portal](https://portal.azure.com/)
2. Register new app in Azure AD
3. Add API permission: **Microsoft Graph > Tasks.ReadWrite**
4. Create client secret
5. Run OAuth flow to get refresh token

### Environment Variables

```bash
export TODOAT_MSTODO_CLIENT_ID="xxx"
export TODOAT_MSTODO_CLIENT_SECRET="xxx"
export TODOAT_MSTODO_REFRESH_TOKEN="xxx"
```

### Run Tests

```bash
go test -tags=integration -v ./backend/mstodo
```

---

## Run All Integration Tests

```bash
# With .env file
set -a && source .env && set +a && go test -tags=integration -v ./...

# Or with Makefile (starts Docker for Nextcloud)
make test-integration
```

## See Also

- [Integration Testing](../explanation/integration-testing.md) - Detailed explanation
- [Backend Configuration](../explanation/backends.md) - Backend setup for regular use
