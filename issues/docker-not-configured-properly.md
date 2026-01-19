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
