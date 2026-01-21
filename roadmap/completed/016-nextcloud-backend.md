# [016] Nextcloud CalDAV Backend

## Summary
Implement Nextcloud Tasks integration via CalDAV protocol, supporting full CRUD operations on VTODO components with authentication via keyring, environment variables, or config.

## Documentation Reference
- Primary: `docs/explanation/backend-system.md#nextcloud-backend`
- Related: `docs/explanation/credential-management.md`, `docs/explanation/configuration.md`

## Dependencies
- Requires: [003] SQLite Backend (for TaskManager interface definition)
- Requires: [010] Configuration (for backend config structure)

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestNextcloudListTaskLists` - `todoat --backend=nextcloud --list-backends` shows Nextcloud calendars
- [ ] `TestNextcloudGetTasks` - `todoat --backend=nextcloud MyCalendar` retrieves tasks from Nextcloud
- [ ] `TestNextcloudAddTask` - `todoat --backend=nextcloud MyCalendar add "Task"` creates VTODO on server
- [ ] `TestNextcloudUpdateTask` - `todoat --backend=nextcloud MyCalendar update "Task" -s DONE` updates task status
- [ ] `TestNextcloudDeleteTask` - `todoat --backend=nextcloud MyCalendar delete "Task"` removes task from server
- [ ] `TestNextcloudStatusTranslation` - Internal `TODO` maps to CalDAV `NEEDS-ACTION`, `DONE` to `COMPLETED`
- [ ] `TestNextcloudPriorityMapping` - Priority 1-9 stored correctly in VTODO PRIORITY field
- [ ] `TestNextcloudCredentialsFromKeyring` - Backend retrieves password from system keyring
- [ ] `TestNextcloudCredentialsFromEnv` - Backend retrieves credentials from `TODOAT_NEXTCLOUD_*` env vars
- [ ] `TestNextcloudHTTPSEnforcement` - HTTP connections rejected unless `allow_http: true` configured
- [ ] `TestNextcloudSelfSignedCert` - Self-signed certs work with `insecure_skip_verify: true`

## Implementation Notes
- CalDAV endpoint: `https://host/remote.php/dav/calendars/username/`
- HTTP client with connection pooling (10 max idle, 2 per host, 30s timeout)
- VTODO parser converts iCalendar to internal Task struct
- Status mapping: TODO↔NEEDS-ACTION, DONE↔COMPLETED, PROCESSING↔IN-PROCESS, CANCELLED↔CANCELLED
- ETag handling for optimistic locking on updates

## Out of Scope
- Synchronization/caching (separate roadmap item)
- CalDAV categories/tags syncing (basic support only)
- Recurring tasks (RRULE parsing)
