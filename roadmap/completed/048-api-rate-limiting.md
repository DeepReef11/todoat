# [048] API Rate Limiting

## Summary
Implement rate limit handling for REST API backends (Todoist, Google Tasks, Microsoft To Do) with automatic retry using exponential backoff and jitter to prevent thundering herd effects.

## Documentation Reference
- Primary: `docs/explanation/synchronization.md`
- Section: Sync Queue System - Retry Logic (lines 632-647, 732-737)
- Related: `docs/explanation/backend-system.md` - backend error handling

## Dependencies
- Requires: [021] Todoist Backend
- Requires: [027] Google Tasks Backend
- Requires: [028] Microsoft To Do Backend
- Requires: [018] Synchronization Core (retry infrastructure)

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestRateLimitRetry` - 429 response triggers automatic retry after backoff period
- [ ] `TestRateLimitExponentialBackoff` - Consecutive 429s increase delay (1s, 2s, 4s, 8s, 16s)
- [ ] `TestRateLimitMaxRetries` - After 5 retries, operation fails with clear error message
- [ ] `TestRateLimitJitter` - Backoff includes random jitter (±20%) to prevent synchronized retries
- [ ] `TestRateLimitHeaderRespect` - `Retry-After` header value is used when provided by API
- [ ] `TestRateLimitQueueing` - Rate-limited operations added to sync queue for later retry

### Functional Requirements
- [ ] Parse HTTP 429 responses and extract retry information
- [ ] Implement exponential backoff: base_delay * 2^attempt (capped at 32s)
- [ ] Add random jitter to prevent thundering herd on shared backends
- [ ] Respect `Retry-After` header when present (seconds or HTTP-date)
- [ ] Log rate limit events at WARN level with backend name and retry time
- [ ] Track rate limit statistics per backend for `sync status` display

## Implementation Notes

### Backoff Algorithm
```go
func calculateBackoff(attempt int, retryAfter *time.Duration) time.Duration {
    if retryAfter != nil {
        return *retryAfter
    }

    baseDelay := 1 * time.Second
    maxDelay := 32 * time.Second

    delay := baseDelay * time.Duration(1<<attempt)
    if delay > maxDelay {
        delay = maxDelay
    }

    // Add ±20% jitter
    jitter := time.Duration(rand.Float64()*0.4-0.2) * delay
    return delay + jitter
}
```

### Rate Limit Detection
- HTTP 429 Too Many Requests
- Todoist: `X-RateLimit-Remaining` header
- Google: `Retry-After` header
- Microsoft: `Retry-After` header

### Error Types
```go
type RateLimitError struct {
    Backend     string
    RetryAfter  time.Duration
    Attempt     int
    MaxAttempts int
}
```

## Out of Scope
- Proactive rate limiting (tracking request counts)
- Rate limit quotas display
- Per-endpoint rate limiting
