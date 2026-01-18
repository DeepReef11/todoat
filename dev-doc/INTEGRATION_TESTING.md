# Integration Testing

## Overview

This document describes how to run integration tests against real backend services (Nextcloud, Todoist, Google Tasks, Microsoft To Do). Integration tests validate the full communication stack with actual services rather than mock servers.

**Related Documentation:**
- [Test Driven Development](./TEST_DRIVEN_DEV.md) - General TDD approach
- [Backend System](./BACKEND_SYSTEM.md) - Backend implementations
- [Synchronization](./SYNCHRONIZATION.md) - Sync testing considerations

---

## Table of Contents

- [Test Architecture](#test-architecture)
- [Running Integration Tests](#running-integration-tests)
- [Docker Test Environment](#docker-test-environment)
- [Nextcloud Integration Tests](#nextcloud-integration-tests)
- [Todoist Integration Tests](#todoist-integration-tests)
- [Google Tasks Integration Tests](#google-tasks-integration-tests)
- [Microsoft To Do Integration Tests](#microsoft-to-do-integration-tests)
- [CI/CD Integration](#cicd-integration)

---

## Test Architecture

### Test Hierarchy

```
Tests
├── Unit Tests (no external dependencies)
│   ├── Mock CalDAV server tests
│   ├── Mock REST API server tests
│   └── Pure logic tests
└── Integration Tests (real services)
    ├── Nextcloud (Docker or remote)
    ├── Todoist (API sandbox)
    ├── Google Tasks (OAuth sandbox)
    └── Microsoft To Do (OAuth sandbox)
```

### Build Tags

Integration tests use Go build tags to separate them from unit tests:

```go
//go:build integration

package mypackage

func TestRealNextcloudConnection(t *testing.T) {}
```

This allows:
- `go test ./...` - Runs only unit tests (fast, no dependencies)
- `go test -tags=integration ./...` - Runs all tests including integration
- `go test -tags=integration -run Nextcloud ./...` - Runs only Nextcloud integration tests

---

## Running Integration Tests

### Quick Reference

```bash
# Unit tests only (default)
go test ./...

# All integration tests
go test -tags=integration ./...

# Specific backend integration tests
go test -tags=integration -run Nextcloud ./backend/nextcloud
go test -tags=integration -run Todoist ./backend/todoist
go test -tags=integration -run Google ./backend/google
go test -tags=integration -run MSTodo ./backend/mstodo

# With verbose output
go test -tags=integration -v ./backend/nextcloud
```

### Makefile Targets

```bash
# Start Docker test environment (Nextcloud)
make docker-up

# Stop Docker test environment
make docker-down

# Run all integration tests (starts Docker if needed)
make test-integration

# Run specific backend integration tests
make test-nextcloud
make test-todoist
```

---

## Docker Test Environment

### Purpose

The Docker test environment provides a local Nextcloud instance for integration testing. This enables:
- Testing against a real CalDAV server
- Reproducible test environment
- CI/CD integration without external dependencies

### docker-compose.yml

Create `docker-compose.yml` in the project root:

```yaml
version: '3.8'

services:
  nextcloud:
    image: nextcloud:latest
    container_name: todoat-nextcloud-test
    ports:
      - "8080:80"
    environment:
      - NEXTCLOUD_ADMIN_USER=admin
      - NEXTCLOUD_ADMIN_PASSWORD=adminpass
      - NEXTCLOUD_TRUSTED_DOMAINS=localhost
    volumes:
      - nextcloud_data:/var/www/html
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost/status.php"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s

  # Optional: Database for better performance
  db:
    image: mariadb:10.6
    container_name: todoat-db-test
    environment:
      - MYSQL_ROOT_PASSWORD=rootpass
      - MYSQL_DATABASE=nextcloud
      - MYSQL_USER=nextcloud
      - MYSQL_PASSWORD=nextcloudpass
    volumes:
      - db_data:/var/lib/mysql

volumes:
  nextcloud_data:
  db_data:
```

### Makefile Targets

Add these targets to `Makefile`:

```makefile
.PHONY: docker-up docker-down docker-wait test-integration test-nextcloud

# Start Docker test environment
docker-up:
	docker-compose up -d
	@echo "Waiting for Nextcloud to be ready..."
	@$(MAKE) docker-wait

# Wait for Nextcloud to be healthy
docker-wait:
	@until docker inspect --format='{{.State.Health.Status}}' todoat-nextcloud-test 2>/dev/null | grep -q healthy; do \
		echo "Waiting for Nextcloud..."; \
		sleep 5; \
	done
	@echo "Nextcloud is ready!"

# Stop Docker test environment
docker-down:
	docker-compose down -v

# Run all integration tests
test-integration: docker-up
	TODOAT_NEXTCLOUD_HOST=localhost:8080 \
	TODOAT_NEXTCLOUD_USERNAME=admin \
	TODOAT_NEXTCLOUD_PASSWORD=adminpass \
	go test -tags=integration -v ./...

# Run Nextcloud integration tests only
test-nextcloud: docker-up
	TODOAT_NEXTCLOUD_HOST=localhost:8080 \
	TODOAT_NEXTCLOUD_USERNAME=admin \
	TODOAT_NEXTCLOUD_PASSWORD=adminpass \
	go test -tags=integration -v -run Integration ./backend/nextcloud

# Run Todoist integration tests (requires TODOAT_TODOIST_TOKEN env var)
test-todoist:
	go test -tags=integration -v -run Integration ./backend/todoist
```

### Starting the Environment

```bash
# Start Nextcloud container
make docker-up

# Verify it's running
curl http://localhost:8080/status.php

# Output should show Nextcloud is installed and running
```

### First-Time Setup

After starting the container for the first time:

1. **Wait for initialization** (30-60 seconds)
2. **Install Tasks app** (if not pre-installed):
   ```bash
   docker exec todoat-nextcloud-test su -s /bin/bash www-data -c \
     "php occ app:install tasks"
   ```
3. **Create test calendar** (optional):
   ```bash
   docker exec todoat-nextcloud-test su -s /bin/bash www-data -c \
     "php occ dav:create-calendar admin TestCalendar"
   ```

---

## Nextcloud Integration Tests

### Test File Structure

```go
//go:build integration

package nextcloud

import (
    "context"
    "os"
    "testing"
)

// TestIntegrationNextcloudConnection tests real CalDAV connection
func TestIntegrationNextcloudConnection(t *testing.T) {
    host := os.Getenv("TODOAT_NEXTCLOUD_HOST")
    if host == "" {
        t.Skip("TODOAT_NEXTCLOUD_HOST not set, skipping integration test")
    }

    cfg := Config{
        Host:      host,
        Username:  os.Getenv("TODOAT_NEXTCLOUD_USERNAME"),
        Password:  os.Getenv("TODOAT_NEXTCLOUD_PASSWORD"),
        AllowHTTP: true, // For local Docker testing
    }

    be, err := New(cfg)
    if err != nil {
        t.Fatalf("Failed to create backend: %v", err)
    }
    defer be.Close()

    ctx := context.Background()
    lists, err := be.GetLists(ctx)
    if err != nil {
        t.Fatalf("GetLists failed: %v", err)
    }

    t.Logf("Found %d calendars", len(lists))
}
```

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `TODOAT_NEXTCLOUD_HOST` | Nextcloud host:port | `localhost:8080` |
| `TODOAT_NEXTCLOUD_USERNAME` | Nextcloud username | `admin` |
| `TODOAT_NEXTCLOUD_PASSWORD` | Nextcloud password | `adminpass` |

### Running Nextcloud Integration Tests

```bash
# With Docker environment
make test-nextcloud

# Manual (if Docker already running)
TODOAT_NEXTCLOUD_HOST=localhost:8080 \
TODOAT_NEXTCLOUD_USERNAME=admin \
TODOAT_NEXTCLOUD_PASSWORD=adminpass \
go test -tags=integration -v ./backend/nextcloud

# Against remote Nextcloud (HTTPS)
TODOAT_NEXTCLOUD_HOST=cloud.example.com \
TODOAT_NEXTCLOUD_USERNAME=myuser \
TODOAT_NEXTCLOUD_PASSWORD=mypass \
go test -tags=integration -v ./backend/nextcloud
```

### Test Coverage

Integration tests should cover:
- Connection establishment
- List (calendar) CRUD operations
- Task CRUD operations
- Status translation (NEEDS-ACTION <-> TODO, etc.)
- Priority mapping
- Subtask hierarchy
- ETag/CTag change detection
- Error handling (network, auth, conflicts)

---

## Todoist Integration Tests

### Test File Structure

```go
//go:build integration

package todoist

import (
    "context"
    "os"
    "testing"
)

// TestIntegrationTodoistConnection tests real Todoist API connection
func TestIntegrationTodoistConnection(t *testing.T) {
    token := os.Getenv("TODOAT_TODOIST_TOKEN")
    if token == "" {
        t.Skip("TODOAT_TODOIST_TOKEN not set, skipping integration test")
    }

    cfg := Config{
        APIToken: token,
    }

    be, err := New(cfg)
    if err != nil {
        t.Fatalf("Failed to create backend: %v", err)
    }
    defer be.Close()

    ctx := context.Background()
    lists, err := be.GetLists(ctx)
    if err != nil {
        t.Fatalf("GetLists failed: %v", err)
    }

    t.Logf("Found %d projects", len(lists))
}
```

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `TODOAT_TODOIST_TOKEN` | Todoist API token | `abc123...` |

### Getting a Test API Token

1. Go to [Todoist Settings](https://todoist.com/prefs/integrations)
2. Scroll to "API token" section
3. Copy the token
4. **For CI/CD**: Use a dedicated test account

### Running Todoist Integration Tests

```bash
# Set token and run
TODOAT_TODOIST_TOKEN=your-api-token \
go test -tags=integration -v ./backend/todoist

# Or use Makefile (if token in env)
make test-todoist
```

### Test Coverage

Integration tests should cover:
- API authentication
- Project (list) CRUD operations
- Task CRUD operations
- Priority mapping (internal 1-9 <-> Todoist 1-4)
- Labels as categories
- Subtask hierarchy (parent_id)
- Rate limiting and retry logic
- Due date handling

### Test Isolation

**Important**: Integration tests create real data. Clean up after tests:

```go
func TestIntegrationTodoistCreateTask(t *testing.T) {
    // ... setup ...

    // Create test project
    list, err := be.CreateList(ctx, "todoat-test-project")
    if err != nil {
        t.Fatalf("CreateList failed: %v", err)
    }

    // ALWAYS clean up
    defer func() {
        _ = be.DeleteList(ctx, list.ID)
    }()

    // ... test logic ...
}
```

---

## Google Tasks Integration Tests

### Environment Variables

| Variable | Description |
|----------|-------------|
| `TODOAT_GOOGLE_CLIENT_ID` | OAuth client ID |
| `TODOAT_GOOGLE_CLIENT_SECRET` | OAuth client secret |
| `TODOAT_GOOGLE_REFRESH_TOKEN` | OAuth refresh token |

### Getting Test Credentials

1. Create a Google Cloud project
2. Enable Google Tasks API
3. Create OAuth credentials (Desktop app)
4. Run OAuth flow once to get refresh token
5. Store refresh token for CI/CD

### Running Google Tasks Integration Tests

```bash
TODOAT_GOOGLE_CLIENT_ID=xxx \
TODOAT_GOOGLE_CLIENT_SECRET=xxx \
TODOAT_GOOGLE_REFRESH_TOKEN=xxx \
go test -tags=integration -v ./backend/google
```

---

## Microsoft To Do Integration Tests

### Environment Variables

| Variable | Description |
|----------|-------------|
| `TODOAT_MSTODO_CLIENT_ID` | Azure AD app client ID |
| `TODOAT_MSTODO_CLIENT_SECRET` | Azure AD app client secret |
| `TODOAT_MSTODO_REFRESH_TOKEN` | OAuth refresh token |

### Getting Test Credentials

1. Register app in Azure AD
2. Add Microsoft Graph Tasks.ReadWrite permission
3. Run OAuth flow once to get refresh token
4. Store refresh token for CI/CD

### Running Microsoft To Do Integration Tests

```bash
TODOAT_MSTODO_CLIENT_ID=xxx \
TODOAT_MSTODO_CLIENT_SECRET=xxx \
TODOAT_MSTODO_REFRESH_TOKEN=xxx \
go test -tags=integration -v ./backend/mstodo
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  nextcloud-integration:
    runs-on: ubuntu-latest
    services:
      nextcloud:
        image: nextcloud:latest
        ports:
          - 8080:80
        env:
          NEXTCLOUD_ADMIN_USER: admin
          NEXTCLOUD_ADMIN_PASSWORD: adminpass
        options: >-
          --health-cmd "curl -f http://localhost/status.php || exit 1"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
          --health-start-period 30s

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Wait for Nextcloud
        run: |
          until curl -s http://localhost:8080/status.php | grep -q installed; do
            sleep 5
          done

      - name: Run Nextcloud Integration Tests
        env:
          TODOAT_NEXTCLOUD_HOST: localhost:8080
          TODOAT_NEXTCLOUD_USERNAME: admin
          TODOAT_NEXTCLOUD_PASSWORD: adminpass
        run: go test -tags=integration -v ./backend/nextcloud

  todoist-integration:
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run Todoist Integration Tests
        env:
          TODOAT_TODOIST_TOKEN: ${{ secrets.TODOAT_TODOIST_TOKEN }}
        run: go test -tags=integration -v ./backend/todoist
```

### Secrets Management

For CI/CD, store credentials as secrets:
- GitHub Actions: Repository secrets
- GitLab CI: CI/CD variables (masked)
- Other: Environment-specific secret management

**Never commit credentials to the repository.**

---

## Best Practices

### 1. Test Isolation

Each test should be independent:
- Create test-specific resources (projects, lists)
- Clean up resources in `defer` blocks
- Use unique names with timestamps or UUIDs

### 2. Skip When Not Configured

Always check for required environment variables:

```go
func TestIntegrationFeature(t *testing.T) {
    if os.Getenv("REQUIRED_VAR") == "" {
        t.Skip("REQUIRED_VAR not set, skipping integration test")
    }
    // ... test code ...
}
```

### 3. Timeouts

Set appropriate timeouts for network operations:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

### 4. Parallel Test Safety

Be careful with parallel tests against shared services:

```go
func TestIntegrationConcurrent(t *testing.T) {
    t.Parallel() // Only if test is isolated
    // ...
}
```

### 5. Rate Limit Awareness

External APIs have rate limits. Add delays if needed:

```go
func TestIntegrationBulkOperations(t *testing.T) {
    for i := 0; i < 10; i++ {
        // ... create task ...
        time.Sleep(100 * time.Millisecond) // Respect rate limits
    }
}
```

---

## Troubleshooting

### Nextcloud Docker Issues

**Problem**: Container won't start
```bash
docker logs todoat-nextcloud-test
```

**Problem**: Tasks app not available
```bash
docker exec todoat-nextcloud-test su -s /bin/bash www-data -c "php occ app:list"
docker exec todoat-nextcloud-test su -s /bin/bash www-data -c "php occ app:install tasks"
```

**Problem**: CalDAV endpoint not responding
```bash
curl -v http://localhost:8080/remote.php/dav/calendars/admin/
```

### Todoist API Issues

**Problem**: 401 Unauthorized
- Verify token is correct
- Check token hasn't expired
- Ensure no extra whitespace in token

**Problem**: 429 Rate Limited
- Add delays between requests
- Use test account with fresh rate limits

---

## Related Documentation

- [Test Driven Development](./TEST_DRIVEN_DEV.md) - TDD workflow
- [Backend System](./BACKEND_SYSTEM.md) - Backend implementations
- [Credential Management](./CREDENTIAL_MANAGEMENT.md) - Credential handling
- [Configuration](./CONFIGURATION.md) - Backend configuration

---

**Navigation:**
- [Back to Overview](./README.md)
- [Back to Features Overview](./FEATURES_OVERVIEW.md)
