# [015] Feature: Analytics system lacks CLI commands for viewing data

## Type
feature-request

## Severity
low

## Documentation Location
- File: docs/explanation/analytics.md
- Section: Full analytics system design

## Feature Description
The analytics system is documented and implemented with tests (internal/analytics/analytics_test.go), but there are no CLI commands to:
1. View analytics data/statistics
2. Query command usage patterns
3. Export analytics reports

The documentation shows SQL queries users can run manually with sqlite3, but suggests no built-in CLI interface exists yet.

## Expected CLI Commands
- Location: New `todoat analytics` subcommand group

Should provide:
- [ ] `todoat analytics summary` - Show usage statistics
- [ ] `todoat analytics report` - Generate detailed report
- [ ] `todoat analytics clear` - Clear analytics data (with confirmation)

Alternatively, this could be a documentation-only fix to explicitly state that analytics data must be queried directly via sqlite3.

## Resolution

**Fixed in**: this session
**Fix description**: Added documentation section clarifying that CLI commands are not yet implemented and users should use sqlite3 directly to query analytics data.
**Test added**: N/A (documentation-only fix)

### Verification Log
```bash
$ cat docs/explanation/analytics.md | grep -A 15 "Viewing Analytics Data"
## Viewing Analytics Data

There are currently no built-in CLI commands for viewing analytics data. To query your usage statistics, use `sqlite3` directly:

```bash
# Open the analytics database
sqlite3 ~/.config/todoat/analytics.db

# Or run a query directly
sqlite3 ~/.config/todoat/analytics.db "SELECT command, COUNT(*) FROM events GROUP BY command;"
```

The sections below provide useful queries you can run to analyze your usage patterns.
```
**Matches expected behavior**: YES - Documentation now explicitly states that analytics must be queried via sqlite3.
