# [003] System Keyring Not Available in Build

## Category
startup

## Severity
low

## Steps to Reproduce
```bash
# Try to store credentials via CLI
echo "your-token" | ./bin/todoat credentials set todoist apitoken --prompt

# Output: Error: failed to store credentials: system keyring not available in this build
```

## Expected Behavior
The credentials set command should either:
1. Work with the system keyring if available, OR
2. Provide an alternative storage method (e.g., encrypted file), OR
3. Show a clear message about how to use environment variables instead

## Actual Behavior
The command fails with "system keyring not available in this build" without suggesting alternatives.

## Error Output
```
Enter password for todoist (user: apitoken): Error: failed to store credentials: system keyring not available in this build
```

## Environment
- OS: Linux 6.12.65-1-lts
- Go version: go1.25.5 linux/amd64
- Config exists: yes
- DB exists: yes

## Possible Cause
The binary was built without keyring support. This may be intentional for certain deployments (e.g., Docker containers, headless servers) but the user experience could be improved by:
1. Detecting keyring availability at build time and documenting it
2. Suggesting environment variable usage as an alternative
3. Providing fallback encrypted file storage

## Related Files
- internal/credentials/credentials.go
- internal/credentials/cli.go
- internal/credentials/keyring.go

## Notes
This may be by design for certain deployment scenarios, but combined with issue #002 (env var not detected), users have no clear path to configure Todoist credentials.

## Resolution

**Fixed in**: this session
**Fix description**: Improved the error message in `credentials set` command when keyring is not available. The error now provides helpful guidance about using environment variables as an alternative, including the specific environment variable names for the backend (e.g., `TODOAT_TODOIST_TOKEN` and `TODOAT_TODOIST_PASSWORD`).
**Test added**: `TestCLICredentialsSetKeyringNotAvailable` in `internal/credentials/cli_test.go`
