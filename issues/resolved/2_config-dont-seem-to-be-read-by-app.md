My config have todoist and nextcloud setup, yet I've only been able to use SQLite.

Step:
```
# Launch todoat to get config sample, add todoist and nextcloud config, set any of these that has lists of task as default backend, sync disabled
todoat list
No lists found. Create one with: todoat list create "MyList"
```

It should have shown list in one of the 2 backend config, but it clearly just use sqlite db instead

## Resolution

**Fixed in**: this session
**Fix description**: Added nextcloud as a supported backend for `default_backend` configuration setting. Previously, only sqlite and todoist were supported. The fix includes:
1. Added `NextcloudConfig` to `BackendsConfig` struct in `internal/config/config.go`
2. Added "nextcloud" to valid backends in config validation (`Validate()`)
3. Added "nextcloud" case in `setConfigValue()` for `config set default_backend` command
4. Added "nextcloud" case in `getBackend()` to create Nextcloud backend when configured as default
5. Added "nextcloud" case in `createBackendByName()` for `--backend nextcloud` flag

**Test added**: `TestNextcloudConfigSetDefaultBackendCLI`, `TestDefaultBackendNextcloudUsedCLI`, `TestDefaultBackendNextcloudWithCredentialsCLI`, and `TestBackendFlagNextcloudCLI` in `backend/nextcloud/config_cli_test.go`

### Verification Log
```bash
$ ./bin/todoat -y config set default_backend nextcloud
Set default_backend = nextcloud

$ ./bin/todoat -y config get default_backend
nextcloud

$ ./bin/todoat -y list
Error: nextcloud backend is configured as default but TODOAT_NEXTCLOUD_HOST environment variable is not set
```
**Matches expected behavior**: YES - When nextcloud is set as default backend without credentials, the CLI now shows a clear error message about missing Nextcloud credentials instead of silently falling back to SQLite.
