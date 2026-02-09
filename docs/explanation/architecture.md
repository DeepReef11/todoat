# Architecture Design Decisions

This document records architecture-level design decisions for the todoat project.

## Decisions

### 2026-01-31 Daemon Unix Socket Restricted to Owner-Only Permissions (0600)

**Decision**: Restrict the daemon Unix domain socket to owner-only permissions (0600), preventing other users on the system from connecting to the daemon and issuing commands.

**Context**: The daemon creates a Unix domain socket for IPC (`internal/daemon/daemon.go`). Previously, the socket inherited the process umask, meaning other users on a multi-user system could potentially connect to the daemon and issue commands such as status queries or sync triggers. This was introduced in commit `a0df401` (background sync daemon).

**Alternatives Considered**:
- Inherit process umask (no restriction): Simpler, but allows other system users to interact with the daemon on shared systems.

**Consequences**:
- Multi-user systems are protected: other users cannot trigger syncs or query daemon status
- Single-user desktops are unaffected by this change
- If shared daemon access is needed in the future, a separate mechanism (e.g., group permissions) would need to be designed

**Implementation**: `internal/daemon/daemon.go` - socket creation should set file permissions to 0600.

**Related**: [ARCH-001] - See `docs/decisions/question-log.md` for full discussion

### 2026-01-31 Implement Heartbeat Mechanism for Daemon Health Monitoring

**Decision**: Implement the `HeartbeatInterval` field in `DaemonConfig` as a functional heartbeat mechanism for hung daemon detection.

**Context**: `DaemonConfig` in `internal/config/config.go` included a `HeartbeatInterval` field, and tests referenced heartbeat behavior (`daemon_test.go:842`), but no heartbeat recording or checking code existed. This created a documentation-code gap where the config field was parseable but non-functional. Dead config fields confuse users who expect them to have an effect.

**Alternatives Considered**:
- Remove the field: Clean up dead code, reduce config surface area. However, heartbeat functionality is genuinely useful for detecting hung daemons.
- Keep field but mark as reserved/future: Document that it's not yet functional. Risks remaining dead code indefinitely.

**Consequences**:
- Enables hung daemon detection, matching the documented daemon architecture
- The `HeartbeatInterval` config field becomes functional
- Daemon health can be monitored by the CLI before attempting IPC communication
- Adds a periodic write to the heartbeat store, which has minimal overhead

**Implementation**: `internal/daemon/daemon.go` - heartbeat recording loop; `internal/config/config.go` - `HeartbeatInterval` field in `DaemonConfig`.

**Related**: [ARCH-007] - See `docs/decisions/question-log.md` for full discussion

### 2026-02-08 Background Daemon Explanation Doc Updated for Heartbeat Implementation

**Decision**: Update `docs/explanation/background-deamon.md` to describe the actual file-based heartbeat implementation, removing "NOT YET IMPLEMENTED" markers.

**Context**: The daemon heartbeat mechanism was implemented in commit `de7491d` (Issue #74). User-facing documentation was complete (`docs/reference/configuration.md`, `docs/how-to/sync.md`), but the explanation doc still contained outdated content. The update documented the heartbeat file location, configuration example, status output examples, and clarified planned vs implemented sections.

**Alternatives Considered**:
- Remove planned sections entirely: Would lose context about the feature's evolution.

**Consequences**:
- Explanation doc accurately reflects the implemented heartbeat mechanism
- "NOT YET IMPLEMENTED" banners removed from implemented features
- Configuration example shows `heartbeat_interval` option (default: 5 seconds)
- Users can understand the file-based heartbeat design

**Related**: [ARCH-025] - See `docs/decisions/question-log.md` for full discussion

### 2026-02-08 Config Validation Expanded to All 7 Supported Backends

**Decision**: Expand the `validBackends` map in `Config.Validate()` to include all 7 supported backends: sqlite, todoist, nextcloud, google, mstodo, file, and git.

**Context**: The `Config.Validate()` function in `internal/config/config.go` hardcoded `validBackends` to only `sqlite`, `todoist`, and `nextcloud`. However, the codebase implements 7 backends that are loaded dynamically via `GetBackendConfig()` using the raw config map. Users setting `default_backend: google` would get a validation error even though the backend works correctly.

**Alternatives Considered**:
- Keep current validation: Users of Google Tasks, MS Todo, File, and Git backends would need to work around validation errors.

**Consequences**:
- All 7 backends can now be set as `default_backend` without validation errors
- The `BackendsConfig` struct should be expanded to include typed fields for all backends
- Config validation is consistent with the actual backend implementations

**Implementation**: `internal/config/config.go` - expand `validBackends` map and `BackendsConfig` struct.

**Related**: [ARCH-020] - See `docs/decisions/question-log.md` for full discussion

### 2026-02-08 Reminder System Enabled by Default

**Decision**: Add `Reminder: ReminderConfig{Enabled: true}` to `DefaultConfig()`, enabling reminders out of the box for new users.

**Context**: Decision FEAT-008 established that analytics should be "enabled by default" with `Analytics.Enabled: true` in `DefaultConfig()`. However, reminders had no explicit default, so `Reminder.Enabled` defaulted to `false` (Go zero value). This meant new users got analytics enabled but reminders disabled, requiring explicit configuration to use reminders. Users adding `--due-date` to tasks wouldn't get reminders unless they also enabled them in config.

**Alternatives Considered**:
- Keep reminders disabled by default: Users must explicitly enable reminders in config, which is a barrier to adoption.
- Enable only when intervals configured: Provides a middle ground but adds complexity.

**Consequences**:
- New users get reminders enabled by default, matching the analytics default
- Users can disable reminders explicitly in config if desired
- The sample config should be updated to reflect the new default

**Implementation**: `internal/config/config.go` - add `Reminder: ReminderConfig{Enabled: true}` to `DefaultConfig()`.

**Related**: [FEAT-022] - See `docs/decisions/question-log.md` for full discussion
