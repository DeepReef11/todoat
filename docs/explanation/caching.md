# Caching System

## Overview

todoat uses a local cache to improve performance when listing task lists. The cache stores list metadata (names, IDs, task counts) and avoids repeated network requests or database queries for frequently accessed information.

## What Is Cached

The cache stores **list metadata only**, not individual tasks:

| Field | Description |
|-------|-------------|
| `id` | Unique identifier for the list |
| `name` | Display name |
| `description` | Optional description |
| `color` | Hex color code (if set) |
| `task_count` | Number of tasks in the list |
| `modified` | Last modification timestamp |

**Note**: Task data is not cached. Each task operation fetches current data from the backend.

## Cache Location

The cache file follows the XDG Base Directory Specification:

```
$XDG_CACHE_HOME/todoat/lists.json
```

Default location (if `XDG_CACHE_HOME` is not set):

```
~/.cache/todoat/lists.json
```

The cache file is JSON format with permissions `0644`.

## Cache Behavior

### Time-To-Live (TTL)

The cache has a 5-minute TTL by default. After 5 minutes, the next list operation refreshes the cache from the backend.

| Scenario | Behavior |
|----------|----------|
| Cache fresh (< 5 min old) | Returns cached data immediately |
| Cache stale (> 5 min old) | Fetches fresh data, updates cache |
| Cache missing | Creates cache directory and file |
| Cache corrupt | Deletes corrupt file, regenerates |

### Cache Invalidation

The cache is automatically invalidated (forcing a refresh) when:

1. **List operations**
   - Creating a list: `todoat list create "New List"`
   - Deleting a list: `todoat list delete "Old List"`

2. **Task operations**
   - Adding a task: `todoat "MyList" add "New task"`
   - Deleting a task: `todoat "MyList" delete "Old task"`

3. **Sync operations**
   - Running explicit sync: `todoat sync`

### Per-Backend Isolation

The cache validates the backend name before use. If you switch backends (e.g., from SQLite to Nextcloud), the old cache is automatically invalidated and refreshed with data from the new backend. This prevents stale data from one backend appearing when using another.

## Performance Impact

| Operation | Without Cache | With Cache |
|-----------|---------------|------------|
| List display | 100-300ms | 5-10ms |
| Interactive selection | 100-300ms | 5-10ms |

The improvement is most noticeable with remote backends (Nextcloud, Google Tasks) where network latency is eliminated for cached operations.

## Troubleshooting

### Viewing Cache Contents

To inspect the cache:

```bash
cat ~/.cache/todoat/lists.json | jq .
```

Example output:

```json
{
  "created_at": "2024-01-15T10:30:00Z",
  "backend": "sqlite",
  "lists": [
    {
      "id": "abc123",
      "name": "Work",
      "task_count": 5,
      "modified": "2024-01-15T10:25:00Z"
    }
  ]
}
```

### Clearing the Cache

If you suspect stale data, manually delete the cache file:

```bash
rm ~/.cache/todoat/lists.json
```

Or if using a custom XDG path:

```bash
rm "$XDG_CACHE_HOME/todoat/lists.json"
```

The cache will be regenerated on the next list operation.

### Common Issues

**Stale task counts**: If list task counts appear wrong after adding/deleting tasks, the cache may not have been invalidated. Run any list operation to refresh:

```bash
todoat list
```

**Wrong lists appearing**: If lists from a previous backend configuration appear, the cache backend validation should handle this automatically. If not, clear the cache manually.

**Permission errors**: Ensure write permissions to the cache directory:

```bash
mkdir -p ~/.cache/todoat
chmod 755 ~/.cache/todoat
```

## Technical Details

### Cache File Structure

```go
type ListCache struct {
    CreatedAt time.Time    `json:"created_at"`
    Backend   string       `json:"backend"`
    Lists     []CachedList `json:"lists"`
}

type CachedList struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description,omitempty"`
    Color       string    `json:"color,omitempty"`
    TaskCount   int       `json:"task_count"`
    Modified    time.Time `json:"modified"`
}
```

### Related Configuration

The cache TTL can be configured in `config.yaml`:

```yaml
cache_ttl: 5m  # Default: 5 minutes
```

## Related Documentation

- [List Management](list-management.md#list-caching) - List caching feature details
- [Configuration](configuration.md#xdg-base-directory-support) - XDG directory configuration
- [Synchronization](synchronization.md) - Sync cache for remote backends
