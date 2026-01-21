# [047] Offline Mode Configuration

## Summary
Implement explicit offline mode configuration allowing users to control application behavior when remote backends are unavailable, with options for automatic detection, forced online, or forced offline operation.

## Documentation Reference
- Primary: `docs/explanation/configuration.md`
- Section: Offline Mode Configuration (lines 1242-1386)
- Related: `docs/explanation/synchronization.md` - Offline Mode section

## Dependencies
- Requires: [018] Synchronization Core (sync queue exists)
- Requires: [010] Configuration System (config loading)

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestOfflineModeAuto` - `sync.offline_mode: auto` detects network state and queues operations when offline
- [ ] `TestOfflineModeOnline` - `sync.offline_mode: online` fails immediately if backend unreachable
- [ ] `TestOfflineModeOffline` - `sync.offline_mode: offline` always queues operations, never contacts remote
- [ ] `TestOfflineModeAutoOnline` - In auto mode, operations succeed when backend is reachable
- [ ] `TestOfflineModeAutoOffline` - In auto mode, operations queue when backend times out
- [ ] `TestOfflineQueuedOpsSync` - Queued operations sync when connectivity restored and `todoat sync` run

### Functional Requirements
- [ ] Three modes: `auto` (default), `online`, `offline`
- [ ] `auto` mode performs connectivity check with configurable timeout
- [ ] `online` mode requires backend connectivity, fails with clear error if unavailable
- [ ] `offline` mode never attempts remote operations, always uses local cache
- [ ] Status indicator shows current mode: `todoat sync status` displays offline/online state

## Implementation Notes

### Config Structure
```yaml
sync:
  enabled: true
  offline_mode: auto  # auto | online | offline
  connectivity_timeout: 5s  # for auto mode
```

### Connectivity Check
- For `auto` mode, perform lightweight HEAD/OPTIONS request before sync
- Cache connectivity state for short duration (30s) to avoid repeated checks
- Configurable timeout (default 5s) prevents long hangs

### Mode Behaviors
| Mode | Backend Available | Backend Unavailable |
|------|------------------|---------------------|
| auto | Direct operation | Queue + local cache |
| online | Direct operation | Error + suggestion |
| offline | Queue always | Queue always |

## Out of Scope
- Automatic mode switching notifications
- Per-backend offline settings
- Network change event listeners
