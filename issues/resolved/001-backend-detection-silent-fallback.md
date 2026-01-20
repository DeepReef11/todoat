# Backend Detection Should Inform User of Errors Instead of Silent Fallback

## Summary
When backend detection selects a backend, it should inform the user if there were errors with other backends instead of silently using a fallback backend.

## Steps to Reproduce

1. Configure `default_backend: nextcloud` in config (see `issues/examples/user-config-multi-backend.yaml`)
2. Have nextcloud backend misconfigured (missing credentials)
3. Run `todoat`

## Expected Behavior
User should be informed that the configured default backend (nextcloud) has errors and that a fallback backend was selected.

## Actual Behavior
```bash
❯❯ todoat
Available lists (1):

NAME                 TASKS
credential           0
```

The app silently falls back to sqlite without informing the user that:
- Their configured `default_backend: nextcloud` was not used
- Why nextcloud backend failed

## Impact
Users may not realize they're using a different backend than intended, leading to confusion about where their tasks are stored.

## Suggested Fix
When falling back from the configured default backend, display a warning message like:
```
Warning: Default backend 'nextcloud' unavailable (missing credentials). Using 'sqlite' instead.
```

## Resolution

**Fixed in**: this session
**Fix description**: Modified `getBackend()` in `cmd/todoat/cmd/todoat.go` to gracefully fall back to SQLite with a warning when the configured default backend (todoist or nextcloud) is unavailable due to missing credentials. Added `warnBackendFallback()` helper function that writes warning messages to stderr.

**Test added**: `TestBackendFallbackWarning` in `cmd/todoat/cmd/todoat_test.go`

### Verification Log
```bash
$ ./todoat config set default_backend nextcloud
Set default_backend = nextcloud

$ unset TODOAT_NEXTCLOUD_HOST TODOAT_NEXTCLOUD_USERNAME TODOAT_NEXTCLOUD_PASSWORD

$ ./todoat
Warning: Default backend 'nextcloud' unavailable (TODOAT_NEXTCLOUD_HOST, TODOAT_NEXTCLOUD_USERNAME, TODOAT_NEXTCLOUD_PASSWORD environment variable(s) not set). Using 'sqlite' instead.
Available lists (10):

NAME                 TASKS
Tasks                2
...
```
**Matches expected behavior**: YES
