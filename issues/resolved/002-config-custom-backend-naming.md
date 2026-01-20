# Config Should Support Custom Backend Naming

## Summary
Configuration should support custom names for backend instances, allowing multiple backends of the same type (e.g., `nextcloud-test`, `nextcloud-prod`).

## Steps to Reproduce

1. Add a custom-named backend to config (see `issues/examples/user-config-multi-backend.yaml`):
   ```yaml
   backends:
     nextcloud-test:
       type: nextcloud
       enabled: true
       host: "localhost:8080"
       username: "admin"
   ```

2. Run `todoat --backend nextcloud-test`

## Expected Behavior
The app should recognize `nextcloud-test` as a valid backend name and use it.

## Actual Behavior
```bash
❯❯ todoat --backend nextcloud-test
Error: unknown backend: nextcloud-test (supported: sqlite, todoist, nextcloud)
```

## Impact
Users cannot:
- Have multiple instances of the same backend type (e.g., production vs test Nextcloud)
- Use descriptive names for their backend configurations

## Use Case
A developer wants to test against a local Nextcloud instance (`nextcloud-test`) while also having their production Nextcloud configured (`nextcloud`).

## Resolution

**Fixed in**: this session
**Fix description**: Added support for custom backend names in configuration. The `createBackendByName` function now checks if a backend name exists in the config file and reads its `type` field to determine the underlying backend type. Added `LoadWithRaw`, `GetBackendConfig`, and `IsBackendConfigured` functions to the config package to support custom backend lookups.

**Test added**: `TestCustomBackendNamingCLI` in `backend/nextcloud/custom_backend_cli_test.go`

### Verification Log
```bash
$ todoat --backend nextcloud-test list
Available lists (1):

NAME                 TASKS
Work                 1
INFO_ONLY
```

With unreachable host (to show it's trying the custom backend, not failing with "unknown backend"):
```bash
$ todoat --backend nextcloud-test list
Error: Propfind "https://nonexistent-host.invalid:9999/remote.php/dav/calendars/admin/": dial tcp: lookup nonexistent-host.invalid on 127.0.0.11:53: no such host
```

**Matches expected behavior**: YES - The app now recognizes custom backend names like `nextcloud-test` and uses them. The error message when connection fails is about network/connection, not "unknown backend".
