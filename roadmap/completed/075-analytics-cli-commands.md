# [075] Analytics CLI Commands

## Summary
Add CLI commands to view analytics data (command usage, success rates, backend performance) without requiring direct sqlite3 queries.

## Documentation Reference
- Primary: `docs/explanation/analytics.md`
- Section: "Viewing Analytics Data" - explicitly notes "There are currently no built-in CLI commands for viewing analytics data"

## Dependencies
- Requires: [067] Analytics System (completed)

## Complexity
M

## Acceptance Criteria

### Tests Required
- [ ] `TestAnalyticsStatsCommand` - `todoat analytics stats` shows usage summary
- [ ] `TestAnalyticsBackendPerformance` - `todoat analytics backends` shows backend performance metrics
- [ ] `TestAnalyticsErrorsCommand` - `todoat analytics errors` shows most common errors
- [ ] `TestAnalyticsTimeRange` - `todoat analytics stats --since 7d` filters by time range
- [ ] `TestAnalyticsJSONOutput` - `todoat analytics stats --json` outputs machine-parseable JSON

### Functional Requirements
- `todoat analytics stats` displays command usage summary (count, success rate)
- `todoat analytics backends` shows backend performance (uses, avg duration, success rate)
- `todoat analytics errors` lists most common errors with counts
- Time range filtering via `--since` flag (e.g., `7d`, `30d`, `1y`)
- JSON output format support for scripting

## Implementation Notes
- Add `analytics` command group in `cmd/todoat/cmd/todoat.go`
- Query existing analytics database at `~/.config/todoat/analytics.db`
- Use same SQL queries documented in `docs/explanation/analytics.md`
- Format output as tables for human-readable mode
- Support `--json` flag for machine-parseable output

## Out of Scope
- Real-time analytics streaming
- Analytics export to external services
- Custom query builder
- Analytics dashboard (TUI)
