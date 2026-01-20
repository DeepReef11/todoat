# [009] Sample Config Contains Outdated Todoist Credential Command

## Type
doc-mismatch

## Category
user-journey

## Severity
medium

## Steps to Reproduce
```bash
# User reads the sample config and tries to set token
cat ~/.config/todoat/config.yaml | grep -A2 "set token"
# Shows: # To set token in keyring: todoat backend todoist set-token

# User tries the command
todoat backend todoist set-token
```

## Expected Behavior
The sample config should show the correct command to set Todoist API token.

## Actual Behavior
The sample config at line 24 contains an outdated command:
```yaml
#   # To set token in keyring: todoat backend todoist set-token
```

The `todoat backend` command does not exist. The correct command is:
```bash
todoat credentials set todoist token --prompt
```

## Error Output
```
$ todoat backend todoist set-token
Error: unknown action: todoist
```

## Environment
- OS: Linux
- Runtime version: Go 1.21+

## Possible Cause
Issue #006 fixed the credentials section (lines 111-112) but missed the comment in the Todoist backend section (line 24).

## Documentation Reference
- File: `internal/config/config.sample.yaml`
- Section: Todoist backend configuration (line 24)
- Documented command: `todoat backend todoist set-token`

## Related Files
- `internal/config/config.sample.yaml` line 24

## Recommended Fix
FIX CODE - Update line 24 in `internal/config/config.sample.yaml` from:
```yaml
#   # To set token in keyring: todoat backend todoist set-token
```

To:
```yaml
#   # To set token in keyring: todoat credentials set todoist token --prompt
```

## Dependencies
None
