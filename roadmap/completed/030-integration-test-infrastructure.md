# [030] Integration Test Infrastructure

## Summary

Add Docker Compose configuration and Makefile targets to enable integration testing against real Nextcloud instances and prepare infrastructure for other backend integration tests.

## Documentation Reference
- Primary: `dev-doc/INTEGRATION_TESTING.md`
- Secondary: `dev-doc/TEST_DRIVEN_DEV.md`

## Dependencies
- Requires: none (infrastructure only)

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestDockerComposeExists` - docker-compose.yml file exists in project root
- [ ] `TestMakefileDockerUp` - `make docker-up` target exists and is documented
- [ ] `TestMakefileDockerDown` - `make docker-down` target exists
- [ ] `TestMakefileTestIntegration` - `make test-integration` target exists
- [ ] `TestMakefileTestNextcloud` - `make test-nextcloud` target exists
- [ ] `TestMakefileTestTodoist` - `make test-todoist` target exists

### Integration Tests Required
- [ ] `TestIntegrationNextcloudConnection` - Connect to real Nextcloud (Docker), list calendars
- [ ] `TestIntegrationNextcloudCRUD` - Create, read, update, delete task on real Nextcloud
- [ ] `TestIntegrationTodoistConnection` - Connect to real Todoist API, list projects (skipped if no token)

## Implementation Notes

### Files to Create/Modify

1. **docker-compose.yml** (new)
   - Nextcloud service with health check
   - Optional MariaDB for persistence
   - Volume mounts for data
   - Environment variables for admin user

2. **Makefile** (modify)
   - Add `docker-up` target
   - Add `docker-down` target
   - Add `docker-wait` target (waits for health)
   - Add `test-integration` target
   - Add `test-nextcloud` target
   - Add `test-todoist` target

3. **backend/nextcloud/integration_test.go** (new)
   - Build tag: `//go:build integration`
   - Tests against real Nextcloud
   - Reads credentials from env vars
   - Skips if env vars not set

4. **backend/todoist/integration_test.go** (new)
   - Build tag: `//go:build integration`
   - Tests against real Todoist API
   - Reads token from env var
   - Skips if token not set

### Docker Compose Configuration

```yaml
services:
  nextcloud:
    image: nextcloud:latest
    container_name: todoat-nextcloud-test
    ports:
      - "8080:80"
    environment:
      - NEXTCLOUD_ADMIN_USER=admin
      - NEXTCLOUD_ADMIN_PASSWORD=adminpass
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost/status.php"]
      interval: 10s
      timeout: 5s
      retries: 5
```

### Makefile Targets

```makefile
docker-up:
	docker-compose up -d
	@$(MAKE) docker-wait

docker-wait:
	@until docker inspect --format='{{.State.Health.Status}}' todoat-nextcloud-test | grep -q healthy; do sleep 5; done

docker-down:
	docker-compose down -v

test-integration: docker-up
	TODOAT_NEXTCLOUD_HOST=localhost:8080 \
	TODOAT_NEXTCLOUD_USERNAME=admin \
	TODOAT_NEXTCLOUD_PASSWORD=adminpass \
	go test -tags=integration -v ./...

test-nextcloud: docker-up
	TODOAT_NEXTCLOUD_HOST=localhost:8080 \
	TODOAT_NEXTCLOUD_USERNAME=admin \
	TODOAT_NEXTCLOUD_PASSWORD=adminpass \
	go test -tags=integration -v -run Integration ./backend/nextcloud

test-todoist:
	go test -tags=integration -v -run Integration ./backend/todoist
```

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `TODOAT_NEXTCLOUD_HOST` | Nextcloud host:port | `localhost:8080` |
| `TODOAT_NEXTCLOUD_USERNAME` | Nextcloud user | `admin` |
| `TODOAT_NEXTCLOUD_PASSWORD` | Nextcloud password | `adminpass` |
| `TODOAT_TODOIST_TOKEN` | Todoist API token | (none) |

## Out of Scope
- GitHub Actions CI/CD configuration (future item)
- Google Tasks integration tests (requires OAuth setup)
- Microsoft To Do integration tests (requires OAuth setup)
- Performance benchmarking infrastructure
- Test data fixtures/seeding
