# [023] Configuration Validation Command

## Summary
Add a `todoat config validate` command that checks the configuration file for errors, warns about deprecated options, and verifies backend connectivity without performing any operations.

## Source
Code analysis: Configuration errors are only discovered when running commands that use the affected settings. Users setting up new backends or modifying config may not know if their changes are valid until they try to use the feature. Recent commits show active config work (cache_ttl, daemon heartbeat).

## Motivation
Configuration problems are frustrating to debug:
- Typos in backend host/username go unnoticed until sync fails
- Invalid YAML syntax breaks the whole app
- Deprecated options silently ignored
- Backend credentials invalid but user doesn't know until they try to use it

A validation command would catch problems early and guide users to fix them.

## Current Behavior
```bash
# User edits config with a typo
vim ~/.config/todoat/config.yaml
# Changes: sync.enabled: ture  (typo for 'true')

# Error only discovered later when using sync
todoat sync status
# Cryptic error or unexpected behavior

# No way to verify backend connectivity before using it
```

## Proposed Behavior
```bash
# Validate configuration
todoat config validate
# Output:
# ✓ YAML syntax valid
# ✓ Required fields present
# ✓ Backend 'nextcloud' configuration complete
# ⚠ Warning: 'sync.pull_interval' is deprecated, use 'sync.daemon.interval'
# ✗ Error: sync.enabled has invalid value 'ture' (expected boolean)
#
# 1 error, 1 warning

# Check specific backend connectivity
todoat config validate --check-connectivity
# Output:
# ✓ YAML syntax valid
# ✓ Required fields present
# ✓ SQLite backend: OK
# ✓ Nextcloud backend: Connected (https://cloud.example.com)
# ✗ Todoist backend: Authentication failed (invalid API token)
#
# 1 error

# JSON output for scripting
todoat config validate --json
# {
#   "valid": false,
#   "errors": [{"path": "sync.enabled", "message": "invalid value 'ture'"}],
#   "warnings": [{"path": "sync.pull_interval", "message": "deprecated"}]
# }
```

## Estimated Value
medium - Reduces configuration debugging time; especially valuable for new users and complex multi-backend setups

## Estimated Effort
S - Most validation logic already exists in config loading; needs extraction and user-friendly reporting

## Related
- Config system: `internal/config/config.go`
- Credentials: `internal/credentials/` (for connectivity checks)
- Existing `--detect-backend` flag: `cmd/todoat/cmd/todoat.go` (similar concept)

## Status
unreviewed
