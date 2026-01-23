# Rate Limiting

todoat handles API rate limiting automatically for cloud backends, using exponential backoff with jitter to retry failed requests transparently.

## Overview

When a backend API returns HTTP 429 (Too Many Requests), todoat automatically retries the request with increasing delays. This happens transparently without user intervention.

## Rate-Limited Backends

| Backend | Rate Limited | Notes |
|---------|-------------|-------|
| Todoist | Yes | Known API rate limits |
| Google Tasks | Yes | Google API quotas |
| Microsoft To Do | Yes | Microsoft Graph limits |
| Nextcloud | Possible | Server-dependent CalDAV limits |
| SQLite | No | Local storage |
| Git | No | Local files |

## Automatic Retry Behavior

When rate limited, todoat:

1. Detects the HTTP 429 response
2. Checks for `Retry-After` header from the server
3. Calculates backoff delay using exponential backoff
4. Waits the calculated delay
5. Retries the request

### Retry Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| Max retries | 5 | Maximum retry attempts before failing |
| Base delay | 1 second | Initial delay before first retry |
| Max delay | 32 seconds | Maximum delay between retries |
| Jitter | ±20% | Random variation to prevent thundering herd |

### Exponential Backoff Schedule

Without a server-provided `Retry-After` header:

| Attempt | Base Delay | With Jitter (range) |
|---------|-----------|---------------------|
| 1 | 1s | 0.8s - 1.2s |
| 2 | 2s | 1.6s - 2.4s |
| 3 | 4s | 3.2s - 4.8s |
| 4 | 8s | 6.4s - 9.6s |
| 5 | 16s | 12.8s - 19.2s |

Delays are capped at 32 seconds maximum.

## User-Facing Impact

### What Users Experience

- **Delayed commands**: Operations may take longer when rate limited
- **Transparent retry**: No user action required; retries happen automatically
- **Eventual success**: Most rate-limited requests succeed after backoff

### When Retries Are Exhausted

If all 5 retries fail, you'll see an error message:

```
Todoist rate limit exceeded after 5 retries (max 5)
```

In this case:
- Wait a few minutes before trying again
- Check if you're running multiple instances of todoat
- Consider spacing out bulk operations

## Retry-After Header

When the server provides a `Retry-After` header, todoat uses that value instead of calculating its own backoff. This respects:

- **Seconds format**: `Retry-After: 60` (wait 60 seconds)
- **HTTP-date format**: `Retry-After: Fri, 31 Dec 2024 23:59:59 GMT`

## Best Practices

### Avoid Rate Limits

1. **Batch operations**: Use sync instead of individual operations when possible
2. **Avoid rapid polling**: Don't refresh too frequently
3. **Single instance**: Don't run multiple todoat instances against the same backend

### If Frequently Rate Limited

1. **Check API quotas**: Verify your API token has adequate quotas
2. **Review usage patterns**: Consider if batch sync is more appropriate
3. **Space operations**: Add delays between bulk operations

## Jitter Explained

Jitter adds randomness to retry delays to prevent the "thundering herd" problem. Without jitter, many clients hitting rate limits simultaneously would all retry at exactly the same time, causing another rate limit cascade.

With ±20% jitter:
- A 10-second backoff becomes 8-12 seconds
- Different clients retry at different times
- Reduces load spikes on the server

## Troubleshooting

### Persistent Rate Limiting

If you consistently hit rate limits:

1. **Check for runaway processes**: Ensure only one todoat instance is running
2. **Review API token permissions**: Some tokens have lower quotas
3. **Contact backend provider**: Request higher quotas if needed

### Rate Limit Errors in Sync

Rate limit errors during sync are handled by the sync queue system, which will retry failed operations automatically.

## See Also

- [Backend Configuration](backends.md) - Configuring backends
- [Synchronization](synchronization.md) - Sync queue and retry behavior
