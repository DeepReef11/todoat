# Analytics

todoat tracks local command usage statistics to help you understand your usage patterns and backend performance. This guide covers viewing and managing analytics data.

## Viewing Command Statistics

### Usage Summary

```bash
todoat analytics stats
```

Shows command usage counts and success rates.

### Filter by Time Period

```bash
# Stats from the past week
todoat analytics stats --since 7d

# Stats from the past month
todoat analytics stats --since 30d

# Stats from the past year
todoat analytics stats --since 1y
```

### JSON Output

```bash
todoat --json analytics stats
```

## Viewing Backend Performance

Compare performance across your configured backends:

```bash
todoat analytics backends
```

Shows usage count, average duration, and success rate per backend.

### Filter by Time Period

```bash
# Backend performance from the past month
todoat analytics backends --since 30d
```

### JSON Output

```bash
todoat --json analytics backends
```

## Viewing Errors

### Most Common Errors

```bash
todoat analytics errors
```

Shows the top 10 errors by default, grouped by command and error type.

### Control Output

```bash
# Show top 20 errors
todoat analytics errors --limit 20

# Errors from the past week
todoat analytics errors --since 7d

# Combine filters
todoat analytics errors --since 30d --limit 5
```

### JSON Output

```bash
todoat --json analytics errors
```

## Configuration

### Enable/Disable Analytics

Analytics is enabled by default. To disable:

```bash
todoat config set analytics.enabled false
```

Or edit `~/.config/todoat/config.yaml`:

```yaml
analytics:
  enabled: false
```

### Environment Variable Override

```bash
# Disable analytics regardless of config
export TODOAT_ANALYTICS_ENABLED=false
```

### Data Retention

Set how long analytics data is kept:

```bash
todoat config set analytics.retention_days 365
```

Set to `0` for unlimited retention.

## Data Location

Analytics data is stored at:

```
~/.config/todoat/analytics.db
```

### Delete All Analytics Data

```bash
rm ~/.config/todoat/analytics.db
```

The database is recreated automatically on next use.

### Direct Database Access

For advanced queries, access the SQLite database directly:

```bash
sqlite3 ~/.config/todoat/analytics.db

# Or run a query directly
sqlite3 ~/.config/todoat/analytics.db "SELECT command, COUNT(*) FROM events GROUP BY command;"
```

## Privacy

Analytics data is:
- Stored locally only, never transmitted
- Limited to command names, backend types, success/failure, and duration
- Does not include task content, descriptions, usernames, or credentials

## Examples

### Weekly Review

```bash
# What commands did I use most this week?
todoat analytics stats --since 7d

# How reliable were my backends?
todoat analytics backends --since 7d

# Any recurring errors?
todoat analytics errors --since 7d
```

### Troubleshoot a Slow Backend

```bash
# Compare backend performance
todoat analytics backends --since 30d

# Check for backend-specific errors
todoat analytics errors --since 30d
```

### Export for Analysis

```bash
# Export stats as JSON for external tools
todoat --json analytics stats > stats.json
todoat --json analytics backends > backends.json
```

## See Also

- [Configuration](../reference/configuration.md) - Analytics configuration options
- [CLI Reference](../reference/cli.md#analytics) - Complete command reference
