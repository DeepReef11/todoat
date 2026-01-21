# HTTP/HTTPS Config Inconsistency Between Identical Backends

## Summary
Two backends configured identically behave differently - one works with HTTP, the other attempts HTTPS and fails.

## Steps to Reproduce

1. Configure two nextcloud backends identically (see `issues/examples/user-config-multi-backend.yaml`):
   ```yaml
   nextcloud:
     type: nextcloud
     enabled: true
     host: "localhost:8080"
     username: "admin"
     allow_http: true
     insecure_skip_verify: true

   nextcloud-test:
     type: nextcloud
     enabled: true
     host: "localhost:8080"  # or "http://localhost:8080"
     username: "admin"
     allow_http: true
     insecure_skip_verify: true
   ```
2. Run `todoat -b nextcloud Tasks`
3. Run `todoat -b nextcloud-test Tasks`

## Expected Behavior
Both backends should behave the same since they have identical configuration.

## Actual Behavior
```bash
❯❯ todoat -b nextcloud Tasks
Error: Propfind "https://localhost:8080/remote.php/dav/calendars/admin/": http: server gave HTTP response to HTTPS client

❯❯ todoat -b nextcloud-test Tasks
# Works correctly
```

## Questions
1. Why do identically configured backends behave differently?
2. Is there cached state affecting the `nextcloud` backend?
3. Is the backend name "nextcloud" treated specially?

## Impact
Unpredictable behavior makes it difficult to debug configuration issues.

## Resolution

**Fixed in**: This session
**Fix description**: Modified `createBackendByName` in `cmd/todoat/cmd/todoat.go` to check if "nextcloud" is configured in the config file before falling back to environment-only mode. When config file entry exists, it now uses `createCustomBackend` which properly reads all config file settings (host, username, password, allow_http, insecure_skip_verify).

**Root cause**: The built-in backend name "nextcloud" was hardcoded to only use `nextcloud.ConfigFromEnv()` which ignores config file entirely. Custom backend names like "nextcloud-test" went through `createCustomBackend` which properly reads config file settings via `buildNextcloudConfigWithKeyring`.

**Test added**: `TestIssue012HTTPConfigInconsistency` and `TestIssue012BuiltinNextcloudReadsConfigFile` in `backend/nextcloud/issue012_config_cli_test.go`

### Verification Log
```bash
$ HOME=/tmp/todoat-test-012 ./todoat -b nextcloud list
Error: Propfind "http://localhost:8080/remote.php/dav/calendars/admin/": dial tcp [::1]:8080: connect: connection refused
Exit code: 1

$ HOME=/tmp/todoat-test-012 ./todoat -b nextcloud-test list
Error: Propfind "http://localhost:8080/remote.php/dav/calendars/admin/": dial tcp [::1]:8080: connect: connection refused
Exit code: 1
```
**Matches expected behavior**: YES - Both backends now use HTTP (http://localhost:8080) instead of HTTPS, respecting `allow_http: true` from config file. Both produce identical errors (connection refused to non-existent server).
