# insecure_skip_verify Not Working for HTTP Servers

## Summary
The `insecure_skip_verify: true` config option doesn't help when connecting to an HTTP (non-HTTPS) server. The app attempts HTTPS even when the server is HTTP-only.

## Steps to Reproduce

1. Configure nextcloud backend with local HTTP server (see `issues/examples/user-config-multi-backend.yaml`):
   ```yaml
   nextcloud:
     type: nextcloud
     enabled: true
     host: "localhost:8080"
     insecure_skip_verify: true
     suppress_ssl_warning: true
   ```
2. Run `todoat -b nextcloud Tasks`

## Expected Behavior
With `insecure_skip_verify: true`, the app should connect to the server, or there should be a way to specify HTTP protocol.

## Actual Behavior
```bash
❯❯ todoat -b nextcloud Tasks
Error: Propfind "https://localhost:8080/remote.php/dav/calendars/admin/": http: server gave HTTP response to HTTPS client
```

The app:
- Always uses HTTPS regardless of config
- `insecure_skip_verify` only affects certificate validation, not protocol selection
- No way to specify HTTP for development/local servers

## Suggested Fix
Either:
1. Add a `protocol` or `use_https` config option
2. Allow `host` to include protocol: `host: "http://localhost:8080"`
3. Auto-detect protocol based on port or response

## Related
- Issue #010: Host URL gets https:// prepended incorrectly

## Resolution

**Fixed in**: this session
**Fix description**: The `allow_http` config option already existed in the nextcloud backend Config struct and was properly handled to use HTTP protocol. However, the `buildNextcloudConfigWithKeyring` function in `cmd/todoat/cmd/todoat.go` was not reading `allow_http` from the YAML config. Added reading of `allow_http` from the backend config map, similar to how `insecure_skip_verify` is read.
**Test added**: `TestIssue009AllowHTTPNotReadFromConfig` in `backend/nextcloud/keyring_cli_test.go`

### Verification Log
```bash
$ # Config with allow_http: true, host: localhost:8080
$ todoat list
Error: Propfind "http://localhost:8080/remote.php/dav/calendars/admin/": dial tcp [::1]:8080: connect: connection refused
```
**Matches expected behavior**: YES

The error now shows `http://localhost:8080` (HTTP protocol) instead of `https://localhost:8080` (HTTPS protocol). The "connection refused" error is expected since no server is running, confirming that the protocol selection is now correct.

### Usage
To enable HTTP protocol for Nextcloud backend, add `allow_http: true` to your config:
```yaml
nextcloud:
  type: nextcloud
  enabled: true
  host: "localhost:8080"
  username: "admin"
  allow_http: true              # Required for HTTP (non-HTTPS) servers
  insecure_skip_verify: true    # Skip TLS certificate verification (for self-signed certs)
```
