# [017] Credential Management

## Summary
Implement secure credential storage and retrieval using OS-native keyrings (macOS Keychain, Windows Credential Manager, Linux Secret Service) with fallback to environment variables and config URLs.

## Documentation Reference
- Primary: `dev-doc/CREDENTIAL_MANAGEMENT.md`
- Related: `dev-doc/BACKEND_SYSTEM.md`, `dev-doc/CONFIGURATION.md`

## Dependencies
- Requires: [016] Nextcloud Backend (needs credentials for authentication)
- Requires: [010] Configuration (for credential config structure)

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestCredentialsSetKeyring` - `todoat credentials set nextcloud myuser --prompt` stores password in system keyring
- [ ] `TestCredentialsGetKeyring` - `todoat credentials get nextcloud myuser` retrieves credentials from keyring
- [ ] `TestCredentialsGetEnvVar` - Credentials retrieved from `TODOAT_NEXTCLOUD_USERNAME` and `TODOAT_NEXTCLOUD_PASSWORD` env vars
- [ ] `TestCredentialsPriority` - Keyring takes precedence over env vars, env vars over config URL
- [ ] `TestCredentialsDelete` - `todoat credentials delete nextcloud myuser` removes credentials from keyring
- [ ] `TestCredentialsNotFound` - `todoat credentials get nonexistent user` returns clear error message
- [ ] `TestCredentialsHiddenInput` - Password input is hidden during `--prompt` mode
- [ ] `TestCredentialsJSON` - `todoat --json credentials get nextcloud myuser` returns JSON with source info
- [ ] `TestCredentialsListBackends` - `todoat credentials list` shows all backends with credential status

## Implementation Notes
- Use `zalando/go-keyring` library for cross-platform keyring access
- Keyring entry format: service=`todoat-[backend]`, account=`[username]`
- Environment variable pattern: `TODOAT_[BACKEND]_USERNAME`, `TODOAT_[BACKEND]_PASSWORD`
- Cache resolved credentials after first successful retrieval

## Out of Scope
- OAuth/token-based authentication (future feature)
- Credential rotation automation
- Multi-factor authentication
