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
