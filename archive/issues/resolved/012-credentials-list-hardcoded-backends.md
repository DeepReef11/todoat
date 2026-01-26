# [012] Review: credentials list uses hardcoded backend list

## Type
code-bug

## Severity
high

## Source
Code review - 2026-01-20_19-24-47

## Steps to Reproduce
1. Configure a custom backend in config.yaml (e.g., a second nextcloud instance named "work-nextcloud")
2. Run `todoat credentials list`
3. Observe that only "nextcloud" and "todoist" are shown, not the custom backend

## Expected Behavior
The `credentials list` command should read configured backends from the actual configuration file and show credential status for all configured backends.

## Actual Behavior
The command returns a hardcoded list of two backends:
```go
// cmd/todoat/cmd/todoat.go:4966-4971
// TODO: Get backend configs from actual configuration
// For now, return a placeholder list
backends := []credentials.BackendConfig{
    {Name: "nextcloud", Username: ""},
    {Name: "todoist", Username: ""},
}
```

This hardcoded list doesn't reflect the user's actual configuration.

## Files Affected
- `cmd/todoat/cmd/todoat.go:4966-4971`

## Recommended Fix
1. Load the configuration using `config.LoadWithRaw()` to access custom backend configurations
2. Iterate through `backends` in the raw config to build the actual list of configured backends
3. Include both standard and custom backend names in the list

## Resolution

**Fixed in**: this session
**Fix description**: Modified `newCredentialsListCmd` in `cmd/todoat/cmd/todoat.go` to load the actual configuration using `config.LoadWithRaw()` and build the backends list from the `backends` map in the raw config instead of using a hardcoded list.
**Test added**: TestIssue012CredentialsListReadsConfiguredBackends in cmd/todoat/cmd/todoat_test.go

### Verification Log
```bash
$ TMPDIR=$(mktemp -d) && mkdir -p "$TMPDIR/todoat" && cat > "$TMPDIR/todoat/config.yaml" << 'EOF'
backends:
  sqlite:
    enabled: true
  nextcloud:
    enabled: true
  work-nextcloud:
    enabled: true
    type: nextcloud
default_backend: sqlite
EOF
$ XDG_CONFIG_HOME="$TMPDIR" todoat credentials list
Backend Credentials:

BACKEND              USERNAME             STATUS          SOURCE
nextcloud                                 Not configured  -
sqlite                                    Not configured  -
work-nextcloud                            Not configured  -
```
**Matches expected behavior**: YES - custom backend 'work-nextcloud' now appears in the list
