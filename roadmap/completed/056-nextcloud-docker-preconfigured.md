# [056] Pre-configured Nextcloud Docker Container

## Summary
Enhance the Nextcloud Docker test environment to be fully pre-configured and ready to accept tasks, with the Tasks app installed and a test calendar created automatically.

## Documentation Reference
- Primary: `docs/explanation/todo.md`
- Related: `docs/explanation/integration-testing.md`

## Dependencies
- Requires: [030] Integration Test Infrastructure

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestDockerNextcloudReady` - `docker-compose up` results in ready-to-use Nextcloud (no setup wizard)
- [ ] `TestDockerTasksAppInstalled` - Tasks app is pre-installed and enabled
- [ ] `TestDockerTestCalendarExists` - Test calendar "TestCalendar" exists for admin user
- [ ] `TestDockerCalDAVEndpoint` - CalDAV endpoint responds at /remote.php/dav/calendars/admin/

### Functional Requirements
- [ ] Container starts without prompting for admin user setup
- [ ] Tasks app is automatically installed via occ command
- [ ] Test calendar created via occ dav:create-calendar
- [ ] Health check passes only when full setup is complete
- [ ] Environment variables work as documented (.env file)
- [ ] No manual intervention required after docker-compose up

## Implementation Notes

### Docker Compose Enhancement
```yaml
services:
  nextcloud:
    image: nextcloud:latest
    container_name: todoat-nextcloud-test
    ports:
      - "8080:80"
    environment:
      - NEXTCLOUD_ADMIN_USER=${NEXTCLOUD_ADMIN_USER:-admin}
      - NEXTCLOUD_ADMIN_PASSWORD=${NEXTCLOUD_ADMIN_PASSWORD:-adminpass}
      - NEXTCLOUD_TRUSTED_DOMAINS=localhost
    volumes:
      - ./scripts/nextcloud-init.sh:/docker-entrypoint-hooks.d/post-installation/init.sh
    healthcheck:
      test: ["CMD-SHELL", "curl -sf http://localhost/remote.php/dav/calendars/admin/testcalendar/ || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 30
      start_period: 60s
    depends_on:
      db:
        condition: service_healthy

  db:
    image: mariadb:10.6
    container_name: todoat-db-test
    environment:
      - MYSQL_ROOT_PASSWORD=rootpass
      - MYSQL_DATABASE=nextcloud
      - MYSQL_USER=nextcloud
      - MYSQL_PASSWORD=nextcloud
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 5s
      timeout: 3s
      retries: 10
```

### Initialization Script (scripts/nextcloud-init.sh)
```bash
#!/bin/bash
# Post-installation hook for Nextcloud Docker

# Wait for Nextcloud to be fully initialized
sleep 5

# Install Tasks app
php /var/www/html/occ app:install tasks || true

# Enable Tasks app (in case already installed but disabled)
php /var/www/html/occ app:enable tasks || true

# Create test calendar for integration tests
php /var/www/html/occ dav:create-calendar admin testcalendar "Test Calendar" || true

echo "Nextcloud initialization complete - Tasks app installed, test calendar created"
```

### Files to Create/Modify
1. Create `scripts/nextcloud-init.sh` with post-installation commands
2. Modify `docker-compose.yml` to mount init script and update health check
3. Update `.env.example` with documented environment variables
4. Update `docs/explanation/integration-testing.md` to reflect pre-configured setup

### Health Check Strategy
- Initial health check waits for CalDAV endpoint with test calendar
- Extended start_period (60s) allows for Tasks app installation
- Retries increased to handle slow container startup

## Out of Scope
- Production Nextcloud deployment
- Nextcloud version pinning (uses :latest)
- Multiple test users (single admin user sufficient)
- Test data seeding (tasks created by tests)
