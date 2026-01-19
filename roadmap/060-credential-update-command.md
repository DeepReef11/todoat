# [060] Credential Update Command

## Summary
Add ability to update existing credentials stored in the system keyring without deleting and re-adding, supporting password rotation and credential verification workflows.

## Documentation Reference
- Primary: `dev-doc/CREDENTIAL_MANAGEMENT.md`
- Related: `dev-doc/CLI_INTERFACE.md`

## Dependencies
- Requires: [017] Credential Management (keyring support must exist)

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestCredentialUpdate` - `todoat credentials update nextcloud user --prompt` updates existing credential
- [ ] `TestCredentialUpdateNonExistent` - Update on non-existent credential shows appropriate error
- [ ] `TestCredentialUpdateVerify` - Updated credential can be retrieved and verified
- [ ] `TestCredentialUpdateNoChange` - Update with same password succeeds (idempotent)

### Functional Requirements
- [ ] `todoat credentials update <backend> <username> --prompt` command
- [ ] Prompts for new password with hidden input
- [ ] Verifies credential exists before update
- [ ] Updates value in system keyring
- [ ] Shows success message with masked info
- [ ] Optional `--verify` flag to test updated credential against backend

### Command Interface
```bash
# Update credential with password prompt
todoat credentials update nextcloud myuser --prompt

# Update credential and verify connection
todoat credentials update nextcloud myuser --prompt --verify

# Show help
todoat credentials update --help
```

### Output Requirements
- [ ] Success: `✓ Credential updated for nextcloud/myuser`
- [ ] With verify: `✓ Credential updated and verified for nextcloud/myuser`
- [ ] Not found: `✗ No credential found for nextcloud/myuser. Use 'credentials set' to create.`
- [ ] Verification failure: `✗ Credential updated but verification failed: <error>`

## Implementation Notes

### Keyring Update
```go
// Keyring API typically supports Set() which overwrites
// No separate Update method needed - Set() handles both create and update
err := keyring.Set(serviceName, username, newPassword)
```

### Verification Flow (optional --verify)
1. Update credential in keyring
2. Create backend connection with new credential
3. Attempt simple operation (list calendars/projects)
4. Report success or rollback with warning

### Service Naming
Same as existing: `todoat-<backend>` (e.g., `todoat-nextcloud`)

## Out of Scope
- Automatic credential rotation on schedule
- Credential expiry detection
- Bulk credential updates
- Credential backup/export
