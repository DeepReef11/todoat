# Local Analytics System

## Overview

Local SQLite-based analytics to track command usage, success rates, and backend performance. Analytics data is stored separately from user data to maintain clear separation of concerns.

**Key Characteristics**:
- **Privacy-First**: All data stored locally, never transmitted
- **Opt-In**: Disabled by default, user must explicitly enable
- **Lightweight**: Minimal overhead on command execution
- **Useful**: Provides insights into usage patterns and backend reliability

---

## Database Location

```
~/.config/todoat/analytics.db
```

Follows XDG Base Directory Specification alongside the main configuration.

---

## Database Schema

### Table Structure

```sql
CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL,
    command TEXT NOT NULL,
    subcommand TEXT,
    backend TEXT,
    success INTEGER NOT NULL,
    duration_ms INTEGER,
    error_type TEXT,
    flags TEXT,
    created_at INTEGER DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_timestamp ON events(timestamp);
CREATE INDEX IF NOT EXISTS idx_command ON events(command);
CREATE INDEX IF NOT EXISTS idx_backend ON events(backend);
CREATE INDEX IF NOT EXISTS idx_success ON events(success);
CREATE INDEX IF NOT EXISTS idx_created_at ON events(created_at);
```

### Field Definitions

| Field | Type | Description |
|-------|------|-------------|
| `id` | INTEGER | Auto-incrementing primary key |
| `timestamp` | INTEGER | Unix timestamp when command was executed |
| `command` | TEXT | Main command (e.g., "add", "list", "complete", "delete", "sync") |
| `subcommand` | TEXT | Optional subcommand if applicable |
| `backend` | TEXT | Backend identifier (e.g., "sqlite", "todoist", "nextcloud") |
| `success` | INTEGER | Boolean (0 = failure, 1 = success) |
| `duration_ms` | INTEGER | Command execution time in milliseconds |
| `error_type` | TEXT | Category of error if failed (e.g., "network_timeout", "auth_failed") |
| `flags` | TEXT | JSON string of flags/options used (e.g., `["--priority", "--tag"]`) |
| `created_at` | INTEGER | Automatic timestamp for record creation |

---

## Architecture: Analytics vs Logging vs Notifications

Analytics is a **separate system** from logging and notifications. While they may seem related, each serves a distinct purpose:

| System | Purpose | Trigger | Storage | Retention |
|--------|---------|---------|---------|-----------|
| **Analytics** | Track command metrics for performance analysis | Once per command (wrapper) | SQLite database | Long-term (months/year) |
| **Logging** | Debug info and operational messages | Throughout code at various points | Rotating text files | Short-term (days/weeks) |
| **Notifications** | Alert user to background events | Specific events (sync error, conflict) | OS notifications + log file | Immediate |

### Why Keep Them Separate

1. **Different data shapes**
   - Logger: free-form text messages with levels (debug, info, warn, error)
   - Analytics: structured events with specific fields (command, duration, success, backend)

2. **Different storage needs**
   - Logger: rotating text files, possibly stdout
   - Analytics: SQLite database for queryable metrics

3. **Different retention policies**
   - Logs: short-term debugging (days/weeks)
   - Analytics: long-term trends (months/year)

4. **Different integration points**
   - Logger: called throughout code at various points
   - Analytics: called once per command execution (wrapper pattern)

### Integration Point

Analytics should be integrated at the **command execution level**, not inside the logger:

```go
// cmd/todoat/cmd/root.go - wrap command execution
func executeCommand(cmd *cobra.Command, args []string) error {
    return analytics.TrackCommand(cmd.Name(), ..., func() error {
        return cmd.RunE(cmd, args)
    })
}
```

### Optional: Analytics Can Also Log

The analytics tracker can emit debug logs for visibility:

```go
func (t *Tracker) TrackCommand(cmd, subcmd, backend string, flags []string, fn func() error) error {
    // ... track analytics ...

    // Also log if verbose mode is enabled
    utils.Debugf("Command %s completed in %dms (success=%v)", cmd, duration, success)

    return err
}
```

This keeps concerns separated while allowing cross-system visibility when needed.

---

## Implementation

### Package Structure

```
internal/analytics/
├── analytics.go      # Core analytics types and functions
├── db.go             # Database initialization and queries
├── tracker.go        # Command tracking middleware
└── analytics_test.go # Tests
```

### Core Types

```go
// Event represents a single analytics event
type Event struct {
    ID         int64
    Timestamp  int64
    Command    string
    Subcommand string
    Backend    string
    Success    bool
    DurationMs int64
    ErrorType  string
    Flags      []string
}

// Tracker handles analytics event recording
type Tracker struct {
    db      *sql.DB
    enabled bool
    mu      sync.Mutex
}
```

### Database Initialization

```go
func NewTracker(configDir string) (*Tracker, error) {
    dbPath := filepath.Join(configDir, "analytics.db")

    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    // Create tables and indexes
    if err := initSchema(db); err != nil {
        return nil, err
    }

    return &Tracker{db: db, enabled: true}, nil
}
```

### Event Tracking

```go
// TrackCommand wraps command execution with analytics
func (t *Tracker) TrackCommand(cmd, subcmd, backend string, flags []string, fn func() error) error {
    if !t.enabled {
        return fn()
    }

    start := time.Now()
    err := fn()
    duration := time.Since(start).Milliseconds()

    event := Event{
        Timestamp:  time.Now().Unix(),
        Command:    cmd,
        Subcommand: subcmd,
        Backend:    backend,
        Success:    err == nil,
        DurationMs: duration,
        Flags:      flags,
    }

    if err != nil {
        event.ErrorType = categorizeError(err)
    }

    // Log asynchronously to avoid slowing down commands
    go t.logEvent(event)

    return err
}
```

### Usage in Commands

```go
func (c *AddCommand) Execute(args []string) error {
    flags := collectFlags(c)

    return analytics.TrackCommand("add", "", c.backend, flags, func() error {
        return c.backend.AddTask(args[0], c.priority, c.tags)
    })
}
```

---

## Analytics Queries

### Success Rate by Command

```sql
SELECT
    command,
    COUNT(*) as total,
    SUM(success) as successful,
    ROUND(100.0 * SUM(success) / COUNT(*), 2) as success_rate
FROM events
GROUP BY command
ORDER BY total DESC;
```

### Backend Performance

```sql
SELECT
    backend,
    COUNT(*) as uses,
    ROUND(AVG(duration_ms), 2) as avg_duration_ms,
    ROUND(100.0 * SUM(success) / COUNT(*), 2) as success_rate
FROM events
WHERE backend IS NOT NULL
GROUP BY backend
ORDER BY uses DESC;
```

### Most Common Errors

```sql
SELECT
    command,
    error_type,
    COUNT(*) as occurrences
FROM events
WHERE success = 0 AND error_type IS NOT NULL
GROUP BY command, error_type
ORDER BY occurrences DESC
LIMIT 10;
```

### Usage Over Time (Last 7 Days)

```sql
SELECT
    DATE(timestamp, 'unixepoch') as date,
    COUNT(*) as command_count
FROM events
WHERE timestamp > strftime('%s', 'now', '-7 days')
GROUP BY date
ORDER BY date;
```

### Command Execution Time

```sql
SELECT
    command,
    MIN(duration_ms) as min_ms,
    ROUND(AVG(duration_ms), 2) as avg_ms,
    MAX(duration_ms) as max_ms
FROM events
WHERE duration_ms IS NOT NULL
GROUP BY command
ORDER BY avg_ms DESC;
```

---

## Configuration

### Config File Entry

```yaml
analytics:
  enabled: true           # default
  retention_days: 365      # Auto-cleanup after this many days
```

### Environment Variable Override

```bash
# Disable analytics regardless of config
export TODOAT_ANALYTICS_ENABLED=false
```

---

## Privacy Considerations

| What is Tracked | What is NOT Tracked |
|-----------------|---------------------|
| Command names (add, list, complete) | Task content or descriptions |
| Backend type used | Usernames or credentials |
| Success/failure status | Personal identifiers |
| Execution duration | Server hostnames |
| Error categories | File paths |
| Flags used (names only) | Flag values |

**Privacy Guarantees**:
- All data stored locally in `~/.config/todoat/analytics.db`
- No network transmission of analytics data
- User can delete the database file at any time: `rm ~/.config/todoat/analytics.db`
- Disabled by default - requires explicit opt-in

---

## Maintenance

### Automatic Cleanup

Cleanup runs automatically on application start if `retention_days` is configured:

```go
func (t *Tracker) Cleanup(retentionDays int) error {
    cutoff := time.Now().Unix() - int64(retentionDays*86400)

    _, err := t.db.Exec("DELETE FROM events WHERE timestamp < ?", cutoff)
    if err != nil {
        return err
    }

    _, err = t.db.Exec("VACUUM")
    return err
}
```

### Manual Database Optimization

```bash
sqlite3 ~/.config/todoat/analytics.db "VACUUM;"
```

---

## Testing

### Verify Analytics Tracking

```bash
# Enable analytics in config
# analytics:
#   enabled: true

# Perform some operations
todoat add "Test task" --priority 1
todoat list
todoat complete 1
todoat -b todoist sync

# Query the database directly
sqlite3 ~/.config/todoat/analytics.db "SELECT command, success, duration_ms FROM events ORDER BY timestamp DESC LIMIT 10;"
```

### Unit Tests

```go
func TestTracker_TrackCommand(t *testing.T) {
    tracker := setupTestTracker(t)
    defer tracker.Close()

    err := tracker.TrackCommand("add", "", "sqlite", []string{"--priority"}, func() error {
        return nil
    })

    assert.NoError(t, err)

    events, err := tracker.Query("SELECT * FROM events WHERE command = 'add'")
    assert.NoError(t, err)
    assert.Len(t, events, 1)
    assert.True(t, events[0].Success)
}
```

---

## Related Documentation

- [Configuration System](configuration.md) - Config file format and locations
- [Logging](logging.md) - Application logging system
- [Backend System](backend-system.md) - Backend types tracked in analytics
