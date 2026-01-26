# Unknown 'credential 0' List Appearing in All Backends

## Summary
A list named "credential" with 0 tasks appears across multiple backends (nextcloud, todoist), which seems unexpected and unclear to users.

## Steps to Reproduce

1. Run `todoat -b nextcloud list` or `todoat -b todoist list`

## Actual Behavior
```bash
❯❯ todoat -b nextcloud list
Available lists (1):

NAME                 TASKS
credential           0

❯❯ todoat -b todoist list
Available lists (1):

NAME                 TASKS
credential           0
```

## Questions
1. What is the "credential" list?
2. Why does it appear in multiple backends?
3. Is this a system/internal list that shouldn't be shown to users?
4. Is this related to keyring/credential storage?

## Impact
- Confuses users about what lists exist
- Unclear if this is expected behavior or a bug
- May indicate a misconfiguration or data leak between backends

## Resolution

**Fixed in**: this session
**Fix description**: The root cause was a cache isolation bug where the list cache was shared across all backends.

### Root Cause Analysis

1. **Cache not backend-aware**: The `tryReadListCache()` function stored a `Backend` field but never validated it when reading
2. **Backend name not identified**: `getBackendName()` returned "unknown" for all non-sqlite backends (todoist, nextcloud, etc.)
3. **Result**: When user ran `todoat credential` (typo) with SQLite, created a "credential" list that got cached. Later, running `todoat -b nextcloud list` or `todoat -b todoist list` would read the same cache and show the stale "credential" list.

### Changes Made

1. **cmd/todoat/cmd/todoat.go:tryReadListCache()**: Added backend name validation - cache is now only valid if `cacheData.Backend == currentBackend`
2. **cmd/todoat/cmd/todoat.go:getBackendName()**: Extended to properly identify all backend types:
   - sqlite → "sqlite"
   - todoist → "todoist"
   - nextcloud → "nextcloud"
   - syncAwareBackend → "sync-{underlying}"
   - unknown → "unknown-{type}"
3. **cmd/todoat/cmd/todoat.go:doListView()**: Passes current backend name to cache reader

**Tests added**: `TestCacheBackendIsolation` and `TestCacheBackendMismatchInvalidates` in internal/cache/cache_test.go

### Verification Log
```bash
$ go test -v -run "TestCacheBackendIsolation|TestCacheBackendMismatchInvalidates" ./internal/cache/...
=== RUN   TestCacheBackendIsolation
--- PASS: TestCacheBackendIsolation (0.02s)
=== RUN   TestCacheBackendMismatchInvalidates
--- PASS: TestCacheBackendMismatchInvalidates (0.03s)
PASS
ok  	todoat/internal/cache	0.063s
```
**Matches expected behavior**: YES (cache is now isolated per backend)
