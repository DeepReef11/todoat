# [028] Multi-Profile Workspaces

## Summary
Support multiple configuration profiles (workspaces) that can be quickly switched, enabling separate contexts for work, personal, and project-specific task management.

## Source
Code analysis: Configuration is loaded from a single path (`~/.config/todoat/config.yaml`). Users with multiple task contexts (work vs personal, different projects, different teams) must either merge all backends into one config or manually swap config files.

## Motivation
Common scenarios needing separate configurations:
- Work and personal task separation (different backends, different views)
- Multiple clients/projects with their own task systems
- Testing configurations without affecting production setup
- Shared config for team vs personal customizations

Currently users resort to:
- Environment variable overrides for each session
- Shell aliases to swap config files
- Merging all contexts into one large config

## Current Behavior
```bash
# Single config file at fixed location
~/.config/todoat/config.yaml

# To use different configs, must:
# 1. Set TODOAT_CONFIG env var (if supported)
# 2. Use symlinks to swap config files
# 3. Write wrapper scripts

# No native way to manage multiple configs
```

## Proposed Behavior
```bash
# Create named profiles
todoat profile create work
todoat profile create personal
todoat profile create "client-acme"

# Each profile has its own config
# ~/.config/todoat/profiles/work/config.yaml
# ~/.config/todoat/profiles/personal/config.yaml
# ~/.config/todoat/profiles/client-acme/config.yaml

# Switch active profile
todoat profile use work
# Output: Switched to profile 'work'

# Show current profile
todoat profile current
# Output: work

# List profiles
todoat profile list
# Output:
#   personal
# * work (active)
#   client-acme

# One-off command with different profile
todoat --profile personal list
todoat -P work sync status

# Profile-specific state
# - Current list selection
# - View preferences
# - Sync daemon (per-profile)

# Clone profile
todoat profile clone work work-test

# Delete profile
todoat profile delete work-test

# Export/import profiles
todoat profile export work > work-profile.yaml
todoat profile import contractor < contractor-profile.yaml
```

## Estimated Value
medium - Significant quality-of-life for users managing multiple contexts; reduces config complexity

## Estimated Effort
M - Requires config path abstraction, profile registry, state separation; daemon may need profile awareness

## Related
- Config system: `internal/config/config.go`
- XDG paths: currently uses single XDG config path
- Daemon: `internal/daemon/` may need per-profile instances
- Credentials: `internal/credentials/` may need profile namespacing

## Status
unreviewed
