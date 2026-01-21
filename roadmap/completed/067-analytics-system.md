# [067] Analytics System

## Summary
Implement local SQLite-based analytics to track command usage, success rates, and backend performance as documented in `docs/explanation/analytics.md`.

## Documentation Reference
- Primary: `docs/explanation/analytics.md`
- Section: Full document (design specification)

## Dependencies
- Requires: none

## Complexity
M

## Acceptance Criteria

### Tests Required
- [ ] `TestTracker_TrackCommand` - Verify command tracking records events correctly
- [ ] `TestTracker_Cleanup` - Verify automatic retention cleanup works
- [ ] `TestAnalytics_Disabled` - Verify analytics can be disabled via config
- [ ] `TestAnalytics_EnvironmentOverride` - Verify TODOAT_ANALYTICS_ENABLED override

### Functional Requirements
- [ ] Analytics events stored in `~/.config/todoat/analytics.db`
- [ ] Events table with: timestamp, command, subcommand, backend, success, duration_ms, error_type, flags
- [ ] TrackCommand wrapper function for command execution
- [ ] Asynchronous event logging (non-blocking)
- [ ] Configuration option: `analytics.enabled` (opt-in, disabled by default)
- [ ] Configuration option: `analytics.retention_days` for auto-cleanup
- [ ] Environment variable override: `TODOAT_ANALYTICS_ENABLED`
- [ ] Error categorization for failed commands

## Implementation Notes
- Create `internal/analytics/` package with: analytics.go, db.go, tracker.go
- Database should be at `~/.config/todoat/analytics.db` (XDG compliant)
- Integrate at command execution level in root.go using wrapper pattern
- Store flag names only, NOT flag values (privacy)
- Never log task content, credentials, or personal identifiers

## Out of Scope
- Remote analytics/telemetry (all data stays local)
- Real-time analytics dashboard
- CLI commands to query analytics (can use sqlite3 directly)
