# Host URL Gets https:// Prepended Incorrectly

## Summary
When a user specifies `http://` in the host config, the app still prepends `https://`, resulting in malformed URLs like `https://http//localhost:8080`.

## Steps to Reproduce

1. Configure nextcloud-test backend with explicit http:// (see `issues/examples/user-config-multi-backend.yaml`):
   ```yaml
   nextcloud-test:
     type: nextcloud
     enabled: true
     host: "http://localhost:8080"
   ```
2. Run `todoat -b nextcloud-test Tasks`

## Expected Behavior
The app should use the URL as specified: `http://localhost:8080/remote.php/dav/...`

## Actual Behavior
```bash
❯❯ todoat -b nextcloud-test Tasks
Error: Propfind "https://http//localhost:8080/remote.php/dav/calendars/admin/": dial tcp: lookup http: no such host
```

The app:
- Prepends `https://` to the host value unconditionally
- Results in invalid URL: `https://http//localhost:8080`
- DNS lookup fails on "http" as hostname

## Suggested Fix
1. Check if host already has a protocol prefix before prepending
2. Strip existing protocol and use the specified one, or respect it
3. Parse URL properly to handle various input formats

## Related
- Issue #009: insecure_skip_verify not working for HTTP servers

## Resolution

**Fixed in**: this session
**Fix description**: Modified `New()` in `backend/nextcloud/nextcloud.go` to parse the host value and extract scheme if present (http:// or https://), preventing double-protocol URLs.
**Test added**: `TestIssue010HostURLWithProtocolPrefix` in `backend/nextcloud/nextcloud_test.go`

### Verification Log
```bash
$ go test -v -run TestIssue010 ./backend/nextcloud/
=== RUN   TestIssue010HostURLWithProtocolPrefix
=== RUN   TestIssue010HostURLWithProtocolPrefix/host_with_http_prefix
=== RUN   TestIssue010HostURLWithProtocolPrefix/host_with_https_prefix
=== RUN   TestIssue010HostURLWithProtocolPrefix/host_without_protocol
--- PASS: TestIssue010HostURLWithProtocolPrefix (0.00s)
    --- PASS: TestIssue010HostURLWithProtocolPrefix/host_with_http_prefix (0.00s)
    --- PASS: TestIssue010HostURLWithProtocolPrefix/host_with_https_prefix (0.00s)
    --- PASS: TestIssue010HostURLWithProtocolPrefix/host_without_protocol (0.00s)
PASS
ok  	todoat/backend/nextcloud	0.007s

$ # Verified URL generation:
$ # Input: host="http://localhost:8080" -> http://localhost:8080/remote.php/dav/calendars/admin/
$ # (Previously would have been: https://http://localhost:8080/...)
```
**Matches expected behavior**: YES
