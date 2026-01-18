# [036] List Metadata Caching

## Summary
Implement local caching for task list metadata to improve performance when listing available task lists, reducing network calls to remote backends.

## Documentation Reference
- Primary: `dev-doc/LIST_MANAGEMENT.md` (List Caching section)
- Related: `dev-doc/CONFIGURATION.md`

## Dependencies
- Requires: [007] List Commands (list operations must exist)
- Requires: [010] Configuration (for XDG cache paths)

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestListCacheCreation` - First `todoat list` creates cache file at XDG cache path
- [ ] `TestListCacheHit` - Subsequent `todoat list` within TTL uses cached data (no network call)
- [ ] `TestListCacheInvalidation` - `todoat list create "New"` invalidates cache
- [ ] `TestListCacheTTL` - Cache expires after configured TTL (default 5 minutes)

### Functional Requirements
- [ ] Cache stored at `$XDG_CACHE_HOME/todoat/lists.json` (or SQLite)
- [ ] Default cache TTL: 5 minutes (configurable via `cache_ttl` in config)
- [ ] Cache contains: list ID, name, description, color, task count, backend, last_modified
- [ ] Cache invalidation triggers:
  - List create operation
  - List delete operation
  - List update/rename operation
  - Explicit `todoat sync` command
  - Manual cache clear (delete file)
  - Stale cache (past TTL)
- [ ] Corrupt cache handling: delete and regenerate
- [ ] Missing cache: create with proper permissions (0644)
- [ ] Per-backend cache isolation (multi-backend support)

### Performance Requirements
- [ ] Cache lookup: <10ms
- [ ] Cache hit skips network call entirely for remote backends
- [ ] Expected speedup: 5-10x for remote backend list operations

## Implementation Notes
- Use JSON format for simplicity and human-readability
- Include cache creation timestamp for TTL calculation
- Thread-safe cache access with file locking or sync.RWMutex
- Cache directory created automatically if missing

## Out of Scope
- Task-level caching (handled by sync system)
- Distributed cache invalidation
- Cache compression
