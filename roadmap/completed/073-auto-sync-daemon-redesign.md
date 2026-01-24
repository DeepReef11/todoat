# [073] Auto-Sync Daemon Redesign

## Summary
Redesign the auto-sync daemon to support the new multi-remote backend architecture, enabling automatic background synchronization with multiple backends simultaneously.

## Documentation Reference
- Primary: `docs/explanation/synchronization.md` (Auto-Sync Daemon section, Manual Sync Workflow - Future Auto-Sync Plans)
- Secondary: `docs/explanation/features-overview.md` (marked as "🚧 Being Redesigned")

## Dependencies
- Requires: [024] Auto-Sync Daemon (original implementation, now disabled)
- Requires: [018] Synchronization Core System

## Complexity
L

## Acceptance Criteria

### Tests Required
- [ ] `TestDaemonMultiBackend` - Daemon syncs with multiple backends in sequence
- [ ] `TestDaemonPerBackendInterval` - Each backend can have custom sync interval
- [ ] `TestDaemonCacheIsolation` - Backend caches remain isolated during concurrent sync
- [ ] `TestDaemonSmartTiming` - Avoid sync during active editing (debounce)
- [ ] `TestDaemonFileWatcher` - Optional file watcher triggers sync on local changes

### Functional Requirements
- [ ] Daemon iterates through all sync-enabled backends
- [ ] Per-backend sync state tracking (last sync, pending ops, status)
- [ ] Smart sync timing to avoid interrupting active CLI operations
- [ ] Status reporting shows all backends: `todoat sync daemon status`
- [ ] Graceful handling of backend-specific failures (one backend failure doesn't block others)

## Implementation Notes
- Original daemon (024) was designed for single-backend; needs refactoring for multi-backend
- Consider file watcher for real-time sync triggers on local cache changes
- Smart sync timing: debounce rapid changes, detect active CLI operations
- Each backend maintains independent sync state and interval
- Status output should show per-backend sync health
- Re-enable daemon functionality after architecture changes complete

## Out of Scope
- Per-backend daemon processes (single daemon handles all backends)
- Real-time webhooks from remote backends
- System service installation (systemd, launchd)
