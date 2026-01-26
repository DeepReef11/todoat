# Docker Integration Test Setup is Incomplete

## Issue

The current `docker-compose.yml` in the project root is incomplete for proper Nextcloud integration testing. It lacks:

- Database backend (PostgreSQL)
- Redis for caching
- Initialization script to install Tasks and Calendar apps
- Helper scripts for starting/stopping the test server
- Health check verification scripts

## Solution

The `issues/docker-integration/` folder contains a working example of a complete Nextcloud Docker setup for integration tests, copied from the gosynctasks project.

### What's included in `docker-integration/`

- `docker-compose.yml` - Full stack with PostgreSQL, Redis, and Nextcloud
- `scripts/init-nextcloud.sh` - Automatically installs Tasks and Calendar apps
- `scripts/start-test-server.sh` - Starts the server and waits for readiness
- `scripts/stop-test-server.sh` - Stops the server with optional data cleanup
- `scripts/wait-for-nextcloud.sh` - CI-friendly script to wait for Nextcloud readiness

### Usage

Copy the contents of `issues/docker-integration/` to the project root (or adapt as needed):

```bash
cp -r issues/docker-integration/* .
./scripts/start-test-server.sh
```

Then access Nextcloud at http://localhost:8080 with credentials `admin` / `admin123`.

## Resolution

**Fixed in**: this session
**Fix description**: Copied complete Docker integration setup from `issues/docker-integration/` to project root. Updated `docker-compose.yml` with PostgreSQL database, Redis cache, and proper health checks. Added helper scripts for starting/stopping/waiting for the test server.
**Files added/updated**:
- `docker-compose.yml` - Full stack configuration
- `scripts/init-nextcloud.sh` - Post-installation script to install Tasks and Calendar apps
- `scripts/start-test-server.sh` - Convenient start script with health check
- `scripts/stop-test-server.sh` - Stop script with data cleanup option
- `scripts/wait-for-nextcloud.sh` - CI-friendly readiness check script
