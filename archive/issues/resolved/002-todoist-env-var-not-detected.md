# [002] TODOAT_TODOIST_TOKEN Environment Variable Not Detected by Credentials Command

## Category
todoist

## Severity
medium

## Steps to Reproduce
```bash
# Set environment variable
export TODOAT_TODOIST_TOKEN="your-api-token-here"

# Check credentials status
./bin/todoat -y credentials list

# Output shows "Not configured" even with env var set
```

## Expected Behavior
The `credentials list` command should detect and report that Todoist credentials are available via the `TODOAT_TODOIST_TOKEN` environment variable.

## Actual Behavior
The credentials list command shows "Not configured" for Todoist even when the `TODOAT_TODOIST_TOKEN` environment variable is set. The environment variable IS read by the Todoist backend code (`backend/todoist/todoist.go:37`), but the credentials list command doesn't check for it.

## Error Output
```
Backend Credentials:

BACKEND              USERNAME             STATUS          SOURCE
nextcloud                                 Not configured  -
todoist                                   Not configured  -
```

## Environment
- OS: Linux 6.12.65-1-lts
- Go version: go1.25.5 linux/amd64
- Config exists: yes
- DB exists: yes

## Possible Cause
The credentials listing functionality in the CLI doesn't check for backend-specific environment variables. It only checks the keyring storage.

While `backend/todoist/todoist.go` has `ConfigFromEnv()` at line 35-39 that reads the env var, and even has `HasCredentials()` at line 644 that checks for the env var, the credentials list command doesn't call these backend-specific methods.

## Related Files
- backend/todoist/todoist.go:35-39 (ConfigFromEnv)
- backend/todoist/todoist.go:644-652 (HasCredentials, CredentialSource)
- internal/credentials/credentials.go

## Resolution

**Fixed in**: this session
**Fix description**: Modified `getEnvPassword()` function in `internal/credentials/credentials.go` to also check for `TODOAT_[BACKEND]_TOKEN` environment variable in addition to `TODOAT_[BACKEND]_PASSWORD`. The token check is done first (higher priority) since API tokens are the preferred authentication method for services like Todoist.
**Test added**: `TestCredentialsTodoistTokenEnvVar` in `internal/credentials/credentials_test.go`
