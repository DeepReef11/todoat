# Sample Config Contains Non-Existent Backend Commands

## Type
doc-mismatch

## Category
user-journey

## Severity
high

## Location
- File: `internal/config/config.sample.yaml`
- Lines: 111-112
- Context: sample config

## Documented Command
```bash
# From config.sample.yaml lines 111-112:
todoat backend nextcloud set-password
todoat backend todoist set-token
```

## Actual Result
```bash
$ todoat backend nextcloud set-password
Error: unknown action: nextcloud

$ todoat --help | grep backend
# No "backend" command exists - only "--backend" flag for selecting backends
```

## Working Alternative
```bash
# The actual commands are under "credentials" subcommand:
todoat credentials set nextcloud <username> --prompt
todoat credentials set todoist token --prompt
```

## Recommended Fix
FIX EXAMPLE - Update sample config lines 111-112 to use correct commands:

Change from:
```yaml
# To store credentials in keyring (more secure):
#   todoat backend nextcloud set-password
#   todoat backend todoist set-token
```

To:
```yaml
# To store credentials in keyring (more secure):
#   todoat credentials set nextcloud <username> --prompt
#   todoat credentials set todoist token --prompt
```

## Impact
Users following the sample config instructions will see "Error: unknown action: nextcloud" and won't know how to properly store credentials. The correct `credentials set` command is documented in docs/backends.md but the sample config points to non-existent commands.

## Related
- The credentials commands are correctly documented in `docs/backends.md` lines 31-34
- Issue #004: Keyring not available in standard build (separate issue about build flags)

## Resolution

**Fixed in**: this session
**Fix description**: Updated internal/config/config.sample.yaml lines 111-112 to use the correct `credentials set` commands

### Verification Log
```bash
$ cat internal/config/config.sample.yaml | sed -n '110,112p'
# To store credentials in keyring (more secure):
#   todoat credentials set nextcloud <username> --prompt
#   todoat credentials set todoist token --prompt
```
**Matches expected behavior**: YES
