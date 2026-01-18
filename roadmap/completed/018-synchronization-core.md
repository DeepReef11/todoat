# [018] Synchronization Core System

## Summary
Implement the core synchronization infrastructure with local SQLite caching, sync queue management, and the Sync Manager component that coordinates bidirectional sync between local cache and remote backends.

## Documentation Reference
- Primary: `dev-doc/SYNCHRONIZATION.md`
- Related: `dev-doc/BACKEND_SYSTEM.md`, `dev-doc/CONFIGURATION.md`

## Dependencies
- Requires: [016] Nextcloud Backend (remote backend to sync with)
- Requires: [017] Credential Management (authentication for sync operations)
- Requires: [003] SQLite Backend (for local cache database)

## Complexity
L

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestSyncPull` - `todoat sync` pulls changes from remote backend to local cache
- [ ] `TestSyncPush` - `todoat sync` pushes queued local changes to remote backend
- [ ] `TestSyncStatus` - `todoat sync status` shows last sync time, pending operations, and connection status
- [ ] `TestSyncQueueView` - `todoat sync queue` lists pending operations with timestamps
- [ ] `TestSyncQueueClear` - `todoat sync queue clear` removes all pending operations
- [ ] `TestSyncOfflineAdd` - Adding task while offline queues operation in sync_queue table
- [ ] `TestSyncOfflineUpdate` - Updating task while offline queues operation
- [ ] `TestSyncOfflineDelete` - Deleting task while offline queues operation
- [ ] `TestSyncCacheIsolation` - Each remote backend has separate cache tables
- [ ] `TestSyncETagSupport` - Updates use If-Match header with ETag for optimistic locking
- [ ] `TestSyncConfigEnabled` - `sync.enabled: true` in config enables sync behavior
- [ ] `TestSyncConfigDisabled` - `sync.enabled: false` bypasses sync manager

## Implementation Notes
- Sync Manager sits between CLI layer and storage layer
- Cache database stored at configurable location (default: XDG data dir)
- sync_queue table schema: id, operation, entity_type, entity_id, payload, created_at, retry_count
- ETag/CTag tracking for efficient change detection
- Hierarchical sync: sync parent tasks before children to preserve FK relationships

## Out of Scope
- Auto-sync daemon (separate roadmap item)
- Advanced conflict resolution (separate roadmap item)
- Multi-remote sync (future feature)
