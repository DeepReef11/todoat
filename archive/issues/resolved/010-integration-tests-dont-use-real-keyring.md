# [010] Integration Tests Don't Verify Real Keyring Credential Flow

## Type
test-coverage

## Severity
high

## Source
Investigation of why tests pass but real usage fails (Nextcloud 401 after credential set)

## Description

Tests for credential retrieval from keyring use mocks or environment variables as proxies. No test actually:
1. Stores credentials in the real system keyring
2. Retrieves those credentials for a Nextcloud backend
3. Makes an authenticated request using the retrieved credentials

This gap allowed issue #008 (Nextcloud 401 after credential set) to go undetected.

## Current Test Coverage

The existing tests verify:
- `TestIssue007NextcloudWithConfigAndKeyring`: Uses env var as "proxy" for keyring, tests config file is read
- `TestNextcloudCredentialsFromKeyring`: Only checks `Config.UseKeyring` field, no actual keyring call
- `TestCredentialsSetKeyring`: Tests keyring storage but not retrieval + usage

**What's NOT tested:**
- End-to-end: `credentials set → backend creation → authenticated request`
- Username mismatch between credential storage and config
- Backend name mismatch (e.g., "nextcloud-test" stored credentials retrieved for backend)

## Missing Test Scenario

```go
// This flow is not tested:
func TestNextcloudCredentialE2E(t *testing.T) {
    // 1. Set credentials via CLI (or directly to keyring)
    cli.Execute("credentials", "set", "nextcloud-test", "admin", "--password", "secret")

    // 2. Create config with username matching keyring
    cli.SetConfig(`
backends:
  nextcloud-test:
    type: nextcloud
    host: http://test-server
    username: admin  # MUST match the keyring account
`)

    // 3. Verify backend can authenticate
    stdout, stderr, code := cli.Execute("-b", "nextcloud-test", "list")

    // Should either succeed or fail with connection error (NOT 401)
}
```

## Root Cause of #008

When user runs:
```bash
todoat credentials set nextcloud-test admin --prompt
```

Credentials are stored as:
- Service: `todoat-nextcloud-test`
- Account: `admin`
- Password: [entered password]

When creating the backend:
1. Code looks up `backends.nextcloud-test.username` from config
2. If `username` not in config, `cfg.Username` is empty
3. Keyring lookup uses `cfg.Username` as the account name
4. With empty username, keyring lookup fails or finds wrong entry
5. No password → 401 Unauthorized

## Required Tests

1. **E2E Test with Real Keyring**:
   - On CI, use file-based keyring or test keyring
   - Store credentials, verify retrieval works

2. **Config + Keyring Interaction Test**:
   - Verify username from config is used for keyring lookup
   - Test case: username in config, password in keyring → success
   - Test case: no username in config, credentials in keyring → clear error

3. **Backend Name Consistency Test**:
   - `credentials set nextcloud-test` should work with `-b nextcloud-test`
   - Verify backend name normalization is consistent

## Impact

Without these tests, credential-related bugs can slip through CI even though all tests pass.

## Related Files

- `internal/credentials/credentials.go` - keyring operations
- `cmd/todoat/cmd/todoat.go` - `buildNextcloudConfigWithKeyring()`
- `backend/nextcloud/keyring_cli_test.go` - existing (insufficient) tests
