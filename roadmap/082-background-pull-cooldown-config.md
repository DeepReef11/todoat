# [082] Background Pull Sync Cooldown Configuration

## Summary
Add configurable `sync.background_pull_cooldown` option to control the cooldown period between background pull syncs on read operations.

## Documentation Reference
- Primary: `docs/explanation/synchronization.md` (Team Decisions section - [2026-01-26])
- Secondary: `docs/decisions/question-log.md` (ARCH-006)
- Section: Background pull sync behavior with `auto_sync_after_operation`

## Dependencies
- Requires: none (sync system already implemented)

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] `TestBackgroundPullCooldownConfig` - Config value is parsed and applied correctly
- [ ] `TestBackgroundPullCooldownDefault` - Default value is 30 seconds when not specified
- [ ] `TestBackgroundPullCooldownValidation` - Invalid values (negative, too small) are rejected
- [ ] `TestBackgroundPullCooldownBehavior` - Cooldown actually prevents rapid syncing

### Functional Requirements
- [ ] Add `sync.background_pull_cooldown` config option accepting duration (e.g., "30s", "1m", "5s")
- [ ] Default to 30 seconds when not specified (current hardcoded behavior)
- [ ] Validate that cooldown is at least 5 seconds (prevent excessive syncing)
- [ ] Replace hardcoded `backgroundSyncCooldown = 30 * time.Second` in `cmd/todoat/cmd/todoat.go:2912` with config value
- [ ] Document the option in sample config file

## Implementation Notes
The background pull sync feature (commit 02e2b94) triggers a pull-only sync on read operations when `auto_sync_after_operation` is enabled. Currently uses hardcoded 30-second cooldown. This feature makes it configurable for:
- Power users with fast connections who want fresher data (lower values)
- Users on metered connections who want to reduce network usage (higher values)

### Files to Modify
1. `internal/config/config.go` - Add `BackgroundPullCooldown` field to SyncConfig
2. `cmd/todoat/cmd/todoat.go` - Use config value instead of hardcoded constant
3. `config/sample-config.yaml` - Add documented option

### Configuration Format
```yaml
sync:
  enabled: true
  background_pull_cooldown: 30s  # Duration string (default: 30s, minimum: 5s)
```

## Out of Scope
- Per-backend cooldown configuration
- Different cooldowns for different operation types
- Disabling background pull sync entirely (use `auto_sync_after_operation: false` instead)
