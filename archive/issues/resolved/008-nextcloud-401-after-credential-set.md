# [008] Nextcloud Returns 401 During Sync Pull

## Type
code-bug

## Severity
high

## Source
User report

## Description

When running `todoat -b nextcloud-test sync`, the pull phase fails with 401 Unauthorized even though listing works correctly.

**UPDATE**: This is likely a SYMPTOM of issue #011 (backend data isolation broken). The `-b nextcloud-test list` shows Todoist data due to the isolation bug. The sync pull tries to authenticate to the real Nextcloud but may be using wrong credentials or state.

**This should be re-evaluated after #011 is fixed.**

## Steps to Reproduce

```bash
# Set credentials in keyring
todoat credentials set nextcloud-test admin --prompt
# Enter password when prompted
# Output: "Credentials stored in system keyring"

# Try to use Nextcloud backend
todoat -b nextcloud-test list
# Error: PROPFIND failed with status 401

# Sync also fails
todoat sync
# Pull error: failed to get lists from remote: PROPFIND failed with status 401
```

## User Config (Working)

```yaml
backends:
  nextcloud-test:
    type: nextcloud
    enabled: true
    host: "localhost:8080"
    username: "admin"
    allow_http: true
    insecure_skip_verify: true
    suppress_ssl_warning: true
```

## Expected Behavior

- Credentials from keyring should be used for Nextcloud CalDAV authentication
- PROPFIND request should include Basic Auth header with stored credentials
- Backend should connect successfully

## Actual Behavior

- 401 Unauthorized returned from Nextcloud
- Credentials appear to not be retrieved from keyring correctly
- Same error even after re-setting credentials

## Environment
- Todoist backend works correctly (same keyring mechanism)
- Credentials are successfully stored (confirmed by keyring output)
- Nextcloud server is accessible (was working before)
- Config file has correct host, username, allow_http settings

## Investigation Notes

### Code Path Analysis

1. `createBackendByName()` calls `createCustomBackend()` for "nextcloud-test"
2. `createCustomBackend()` calls `buildNextcloudConfigWithKeyring(name, rawConfig)`
3. `buildNextcloudConfigWithKeyring()` should:
   - Read host, username, allow_http from config file ✓
   - Try keyring for password using `credMgr.Get(ctx, "nextcloud-test", "admin")`
   - Pass all values to `nextcloud.New(cfg)`

4. `nextcloud.New()` constructs URL and stores credentials
5. `doRequest()` calls `req.SetBasicAuth(b.config.Username, b.config.Password)`

### Possible Root Causes

1. **Keyring retrieval failing silently** - `credMgr.Get()` might be returning an error that's ignored
2. **Service name mismatch** - Stored under wrong service name
3. **Password not reaching Backend** - Password retrieved but not passed to New()
4. **Regression in URL construction** - Auth header not sent due to URL issues

### Recent Related Commits

- `17dd340` - fix: parse protocol prefix from host URL
- `9a66d35` - fix: handle unsupported list creation in CalDAV sync
- `35915d2` - fix: implement real keyring support using go-keyring

## Debug Steps

1. Enable debug logging: `TODOAT_DEBUG=1 todoat -b nextcloud-test list`
2. Verify credential retrieval: `todoat credentials get nextcloud-test admin`
3. Check if password is in config object before New() call

## Code Location

- `cmd/todoat/cmd/todoat.go:2729` - `buildNextcloudConfigWithKeyring()`
- `cmd/todoat/cmd/todoat.go:2758-2763` - Keyring lookup
- `backend/nextcloud/nextcloud.go:131` - `SetBasicAuth()` call

## Potential Fix

Add debug logging to `buildNextcloudConfigWithKeyring()` to trace:
1. Whether config values are read correctly
2. Whether keyring lookup is attempted
3. What password value (or empty) results
4. What's passed to `nextcloud.New()`

## Related Files

- `cmd/todoat/cmd/todoat.go` - `buildNextcloudConfigWithKeyring()`
- `internal/credentials/credentials.go` - keyring operations
- `backend/nextcloud/nextcloud.go` - Backend creation and auth
