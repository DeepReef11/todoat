# [029] Docs: Rate limiting behavior not documented

## Type
documentation

## Severity
low

## Test Location
- File: internal/ratelimit/ratelimit_test.go
- Functions:
  - TestRateLimitError
  - TestRateLimitRetry
  - TestRateLimitMaxRetries
  - TestRateLimitExponentialBackoff
  - TestRateLimitJitter
  - TestRateLimitHeaderRespect
  - TestRateLimitHeaderRespectHTTPDate
  - TestRateLimitContextCancellation
  - TestRateLimitMaxDelayCap
  - TestRateLimitQueueing
  - TestRateLimitWithBody
  - TestRateLimitNon429Passthrough
  - TestCalculateBackoff

## Feature Description
Comprehensive rate limiting implementation:
- Automatic retry on 429 responses
- Exponential backoff with jitter
- Respect for Retry-After headers
- Max retry limits
- Request queueing

This is important for users of rate-limited backends (Todoist, etc.).

## Expected Documentation
- Location: docs/explanation/backends.md or new docs/explanation/rate-limiting.md

Should cover:
- [ ] Which backends are rate-limited
- [ ] Automatic retry behavior
- [ ] How long retries continue before failing
- [ ] Any user-facing impact (command takes longer)
- [ ] Configuration options if any

## Priority Backends to Document
- Todoist (known rate limits on API)
- Nextcloud (possible CalDAV limits)

## Resolution

**Fixed in**: this session
**Fix description**: Created comprehensive rate limiting documentation at docs/explanation/rate-limiting.md

### Verification Log
```bash
$ ls docs/explanation/rate-limiting.md
docs/explanation/rate-limiting.md
```

Documentation covers:
- [x] Which backends are rate-limited (table with all backends)
- [x] Automatic retry behavior (5 steps, exponential backoff schedule)
- [x] How long retries continue before failing (5 retries, up to 32s max delay)
- [x] Any user-facing impact (delayed commands, transparent retry, error messages)
- [x] Configuration options if any (default values documented)
- [x] Jitter explanation
- [x] Best practices and troubleshooting

**Matches expected behavior**: YES
