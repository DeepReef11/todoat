# [024] Auto-Sync Daemon

## Summary
Implement background auto-sync daemon that periodically synchronizes tasks with remote backends, supporting configurable intervals and notification integration.

## Documentation Reference
- Primary: `docs/explanation/synchronization.md` (Auto-Sync Daemon section)
- Secondary: `docs/explanation/notification-manager.md`

## Dependencies
- Requires: [018] Synchronization Core System
- Requires: [022] Notification System

## Complexity
L

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestSyncDaemonStart` - `todoat sync daemon start` launches background process
- [ ] `TestSyncDaemonStop` - `todoat sync daemon stop` terminates running daemon
- [ ] `TestSyncDaemonStatus` - `todoat sync daemon status` shows daemon state
- [ ] `TestSyncDaemonInterval` - Sync runs at configured interval
- [ ] `TestSyncDaemonNotification` - Notifications sent on sync events
- [ ] `TestSyncDaemonOffline` - Daemon handles offline gracefully
- [ ] `TestSyncDaemonReconnect` - Daemon reconnects when network restored
- [ ] `TestSyncDaemonPIDFile` - PID file created for process management

## Implementation Notes
- Daemon runs as detached background process
- PID file at `$XDG_RUNTIME_DIR/todoat/daemon.pid` or `/tmp/todoat-daemon.pid`
- Configurable sync interval (default: 5 minutes)
- Exponential backoff on repeated failures
- Integration with notification manager for sync events
- Graceful shutdown on SIGTERM/SIGINT
- Log file for daemon operations

## Out of Scope
- System service installation (systemd, launchd)
- Multiple daemon instances
- Per-backend daemon configuration
- Real-time sync (webhooks)
