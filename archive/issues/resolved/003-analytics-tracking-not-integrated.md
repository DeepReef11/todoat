# [003] Analytics Tracking Not Integrated Into Commands

## Type
code-bug

## Category
feature

## Severity
medium

## Steps to Reproduce
```bash
# 1. Enable analytics in config
cat >> ~/.config/todoat/config.yaml << 'EOF'
analytics:
  enabled: true
  retention_days: 365
EOF

# 2. Run several commands
./todoat list
./todoat TestManual add "Test task"
./todoat TestManual complete "Test task"

# 3. Check for analytics database
ls ~/.config/todoat/analytics.db

# 4. Try to view analytics
./todoat analytics stats
```

## Expected Behavior
According to `docs/explanation/analytics.md`:
- Analytics database should be created at `~/.config/todoat/analytics.db`
- Commands should be tracked automatically when analytics is enabled
- `todoat analytics stats` should show command usage statistics

## Actual Behavior
- Analytics database is never created
- `todoat analytics stats` returns error: "analytics database not found at /home/ubuntu/.config/todoat/analytics.db (analytics may not be enabled)"
- No tracking occurs despite running multiple commands with analytics enabled in config

## Error Output
```
Error: analytics database not found at /home/ubuntu/.config/todoat/analytics.db (analytics may not be enabled)
```

## Environment
- OS: Linux
- Runtime version: Go (development build)

## Possible Cause
The analytics tracking code exists in `internal/analytics/tracker.go` with a `TrackCommand()` function, but this function is never called from any command execution code. The tracking middleware is implemented but not integrated.

Searching for `TrackCommand` usage:
- Only found in `internal/analytics/tracker.go` (definition)
- Only found in `internal/analytics/analytics_test.go` (tests)
- NOT found in any `cmd/` files

The `NewTracker()` function is also never called from the main application code.

## Documentation Reference (if doc-mismatch)
- File: `docs/explanation/analytics.md`
- Section: "Configuration" and "Viewing Analytics Data"
- Documented behavior: "Analytics should be integrated at the **command execution level**" and "Enable analytics in config" should make it work

## Related Files
- `internal/analytics/tracker.go` - Tracker implementation (unused)
- `internal/analytics/db.go` - Database setup (unused)
- `cmd/todoat/cmd/todoat.go` - Analytics CLI commands (work, but no data)

## Recommended Fix
FIX CODE - The analytics tracking middleware needs to be integrated into the command execution flow. The documentation describes the correct architecture, but the implementation is incomplete.

## Resolution

**Fixed in**: this session
**Fix description**: Integrated analytics tracking into the Execute function in cmd/todoat/cmd/todoat.go. The tracker is now initialized when analytics is enabled in config, and wraps command execution to track all commands.
**Test added**: TestAnalyticsTrackingIntegration and TestAnalyticsTrackingDisabled in cmd/todoat/cmd/todoat_test.go

### Changes made:
1. Added `Analytics` config struct to internal/config/config.go with `enabled` and `retention_days` fields
2. Added `IsAnalyticsEnabled()` and `GetAnalyticsRetentionDays()` methods to config
3. Modified tracker.go to use sync.WaitGroup for proper cleanup of async writes
4. Added analytics import and integration to cmd/todoat/cmd/todoat.go
5. Created `initAnalyticsTracker()` function to initialize tracker based on config
6. Wrapped command execution with `tracker.TrackCommand()` when analytics is enabled

### Verification Log
```bash
$ # Enable analytics in config
$ cat config.yaml
default_backend: sqlite
analytics:
  enabled: true
  retention_days: 365

$ # Run several commands
$ ./todoat list
No lists found. Create one with: todoat list create "MyList"

$ ./todoat TestManual add "Test task"
Created task: Test task (ID: 9238925b-b524-45b6-9df1-f941675fd25a)

$ ./todoat TestManual complete "Test task"
Completed task: Test task

$ # Check for analytics database
$ ls ~/.config/todoat/analytics.db
/tmp/todoat_test_config_1085030/todoat/analytics.db

$ # View analytics
$ ./todoat analytics stats
Command Usage Statistics
========================
Command            Total    Success Success Rate
-------            -----    ------- ------------
TestManual             2          2       100.0%
list                   1          1       100.0%
```
**Matches expected behavior**: YES
