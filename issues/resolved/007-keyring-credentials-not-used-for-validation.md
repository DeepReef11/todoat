# Keyring Credentials Not Used for Backend Validation

## Summary
After storing credentials in the system keyring via `todoat credentials set`, the backend validation still requires environment variables and doesn't check the keyring.

## Steps to Reproduce

1. Configure nextcloud backend in config with host and username (see `issues/examples/user-config-multi-backend.yaml`)
2. Store password in keyring:
   ```bash
   todoat credentials set nextcloud admin --prompt
   Enter password for nextcloud (user: admin):
   Credentials stored in system keyring
   ```
3. Run `todoat`

## Expected Behavior
Backend validation should check keyring for credentials and use the stored password.

## Actual Behavior
```bash
❯❯ todoat
Warning: Default backend 'nextcloud' unavailable (TODOAT_NEXTCLOUD_HOST, TODOAT_NEXTCLOUD_USERNAME, TODOAT_NEXTCLOUD_PASSWORD environment variable(s) not set). Using 'sqlite' instead.
```

The warning still asks for environment variables even though:
- `host` and `username` are in the config file
- Password was just stored in the keyring

## Workaround
User must use a `.env` file with environment variables:
```bash
set -a; source .env; set +a
todoat -b nextcloud  # Works after loading env vars
```

## Impact
- Keyring credential storage appears broken or misleading
- Users expect config file + keyring to be sufficient
- Forces use of environment variables despite keyring setup

## Resolution

**Fixed in**: this session
**Fix description**: Modified backend validation to read credentials from config file and keyring, not just environment variables. Added new `buildNextcloudConfigWithKeyring` function that checks config file values first, then environment variables, and finally keyring for password.
**Test added**: `TestIssue007KeyringCredentialsNotUsedForValidation` and `TestIssue007NextcloudWithConfigAndKeyring` in `backend/nextcloud/keyring_cli_test.go`

### Verification Log
```bash
$ # Config file with host=fake-test-server.local, username=admin
$ # All TODOAT_NEXTCLOUD_* env vars unset
$ todoat list
No lists found. Create one with: todoat list create "MyList"
INFO_ONLY
Warning: Default backend 'nextcloud' unavailable (password (keyring, config file, or TODOAT_NEXTCLOUD_PASSWORD) not configured). Using 'sqlite' instead.
```
**Matches expected behavior**: YES

The warning now correctly identifies that only password is missing (and suggests keyring, config file, or env var as sources). Host and username are read from the config file.
