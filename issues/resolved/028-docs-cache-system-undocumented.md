# [028] Docs: Cache system internals not documented

## Type
documentation

## Severity
low

## Test Location
- File: internal/cache/cache_test.go
- Functions:
  - TestListCacheCreation
  - TestListCacheHit
  - TestListCacheTTL
  - TestListCacheInvalidation
  - TestListCacheInvalidationOnDelete
  - TestListCacheMissingDirCreation
  - TestListCacheCorruptHandling
  - TestListCachePerformance
  - TestListCacheContainsRequiredFields
  - TestCacheBackendIsolation
  - TestCacheBackendMismatchInvalidates
  - TestCacheInvalidationOnTaskAdd
  - TestCacheInvalidationOnTaskDelete

## Feature Description
Extensive cache testing reveals:
- List caching with TTL
- Cache invalidation on task/list changes
- Per-backend cache isolation
- Corrupt cache recovery
- Performance optimizations

Users might benefit from understanding cache behavior for troubleshooting.

## Expected Documentation
- Location: docs/explanation/ (new file)
- Suggested file: docs/explanation/caching.md

Should cover:
- [ ] What data is cached and where
- [ ] Cache TTL and invalidation triggers
- [ ] Per-backend isolation
- [ ] How to clear cache if needed
- [ ] Troubleshooting stale data issues

## Alternative
If caching is transparent:
- [ ] Add note in sync.md about cache location
- [ ] Document manual cache clearing for troubleshooting

## Resolution

**Fixed in**: this session
**Fix description**: Created comprehensive caching documentation at `docs/explanation/caching.md` covering all requested topics: what data is cached, cache location (XDG paths), TTL behavior, invalidation triggers, per-backend isolation, how to clear cache, and troubleshooting stale data issues.

### Verification Log
```bash
$ ls -la docs/explanation/caching.md
-rw-rw-r-- 1 ubuntu ubuntu 4570 Jan 23 00:32 docs/explanation/caching.md

$ head -20 docs/explanation/caching.md
# Caching System

## Overview

todoat uses a local cache to improve performance when listing task lists...

$ grep -E "What Is Cached|Cache Location|Time-To-Live|Cache Invalidation|Per-Backend Isolation|Clearing the Cache|Troubleshooting" docs/explanation/caching.md
## What Is Cached
## Cache Location
### Time-To-Live (TTL)
### Cache Invalidation
### Per-Backend Isolation
### Clearing the Cache
## Troubleshooting
```
**Documentation covers all checklist items**: YES
