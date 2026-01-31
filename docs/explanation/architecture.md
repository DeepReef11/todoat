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
