# Notification Manager

The notification manager handles user notifications for background operations where CLI output is not visible. This is primarily used by the background sync daemon.

## Purpose

When todoat runs background sync, the user no longer sees CLI output. The notification manager provides alternative channels to inform users about:
- Sync completion (success/failure)
- Sync conflicts requiring attention
- Connection errors
- Task reminders

## Configuration

```yaml
notification:
  enabled: true

  # Desktop notifications (notify-send, osascript, etc.)
  os_notification:
    enabled: true
    on_sync_complete: false    # Notify on every successful sync
    on_sync_error: true        # Notify on sync failures
    on_conflict: true          # Notify when conflicts detected

  # Log file notifications
  log_notification:
    enabled: true
    path: ""                   # Empty = XDG default (~/.local/share/todoat/notifications.log)
    max_size_mb: 10            # Rotate when log exceeds this size
    retention_days: 30         # Delete logs older than this
```

## Notification Types

| Type | Description | Default |
|------|-------------|---------|
| `sync_complete` | Background sync finished successfully | OS: off, Log: on |
| `sync_error` | Sync failed (connection, auth, etc.) | OS: on, Log: on |
| `conflict` | Sync conflict detected | OS: on, Log: on |
| `reminder` | Task due date reminder | OS: on, Log: on |

## OS Notification Backends

The notification manager auto-detects the available notification system:

| Platform | Backend | Command |
|----------|---------|---------|
| Linux (freedesktop) | notify-send | `notify-send "todoat" "message"` |
| Linux (no DE) | wall (fallback) | `wall "todoat: message"` |
| macOS | osascript | `osascript -e 'display notification "message" with title "todoat"'` |
| Windows | PowerShell | `[System.Windows.Forms.MessageBox]::Show("message")` |

## Missing Backend Tool Behavior

**Decision [FEAT-004]**: When a notification backend tool is missing (e.g., `notify-send` not installed on Linux), todoat uses a two-pronged approach:

1. **Warn once on first failure**: Alert the user that their notification tool is missing, then suppress further warnings for the session.
2. **Auto-detect available tools at startup**: Validate configured notification tools proactively and warn if the preferred tool is unavailable.

**Context**: OS notifications use platform-specific tools: `notify-send` (Linux), `osascript` (macOS), `powershell` (Windows). The Linux fallback is `wall` which broadcasts to all terminal sessions. Previously, the behavior when the primary tool was missing was not explicitly handled — it could silently fail or produce confusing errors.

**Rationale**: Combining warn-once with startup detection provides a good balance: users are informed about missing tools without being nagged repeatedly, and the auto-detection catches configuration issues before they affect runtime behavior.

**Impact**: User experience for notifications and reminders across all platforms.

**Related**: [FEAT-004] - See `docs/decisions/question-log.md` for full discussion

---

## Log Notification Format

```
2026-01-16T10:30:00Z [SYNC_COMPLETE] Synced 5 tasks with nextcloud
2026-01-16T10:35:00Z [SYNC_ERROR] Connection failed: timeout after 30s
2026-01-16T10:40:00Z [CONFLICT] Task "Review PR" has conflicting changes
```

## Implementation

### Interface

```go
// internal/notification/notification.go

type NotificationType string

const (
    NotifySyncComplete NotificationType = "sync_complete"
    NotifySyncError    NotificationType = "sync_error"
    NotifyConflict     NotificationType = "conflict"
    NotifyReminder     NotificationType = "reminder"
)

type Notification struct {
    Type      NotificationType
    Title     string
    Message   string
    Timestamp time.Time
    Metadata  map[string]string  // Additional context (task UID, backend, etc.)
}

type NotificationManager interface {
    // Send dispatches notification to all enabled channels
    Send(n Notification) error

    // SendAsync dispatches without blocking (for background sync)
    SendAsync(n Notification)

    // Close cleans up resources (log file handles, etc.)
    Close() error
}
```

### Factory

```go
// internal/notification/manager.go

func NewNotificationManager(cfg *config.NotificationConfig) (NotificationManager, error) {
    var channels []NotificationChannel

    if cfg.OSNotification.Enabled {
        osChannel, err := newOSNotificationChannel(cfg.OSNotification)
        if err != nil {
            return nil, err
        }
        channels = append(channels, osChannel)
    }

    if cfg.LogNotification.Enabled {
        logChannel, err := newLogNotificationChannel(cfg.LogNotification)
        if err != nil {
            return nil, err
        }
        channels = append(channels, logChannel)
    }

    return &manager{channels: channels}, nil
}
```

### Usage in Sync Manager

```go
// backend/sync/manager.go

func (s *SyncManager) runBackgroundSync() {
    result, err := s.Sync()

    if err != nil {
        s.notifier.Send(notification.Notification{
            Type:    notification.NotifySyncError,
            Title:   "Sync Failed",
            Message: err.Error(),
        })
        return
    }

    if len(result.Conflicts) > 0 {
        s.notifier.Send(notification.Notification{
            Type:    notification.NotifyConflict,
            Title:   "Sync Conflicts",
            Message: fmt.Sprintf("%d conflicts need attention", len(result.Conflicts)),
        })
    }
}
```

## CLI Commands

```bash
# Test notification system
todoat notification test

# View notification log
todoat notification log

# Clear notification log
todoat notification log clear
```

## Notification Configuration Design

**Decision [UX-012]**: Add `notification` config block to the Config struct in `internal/config/config.go`, implementing the configuration described in this document.

**Context**: The notification system's config was hardcoded in `cmd/todoat/cmd/todoat.go` with all channels always enabled. Users could not configure notification behavior through `config.yaml` — only reminder delivery channels were configurable via `reminder.os_notification` and `reminder.log_notification`. However, this document described a `notification:` YAML config block with options like `os_notification.enabled`, `os_notification.on_sync_error`, `log_notification.path`, etc.

**Rationale**: Adding the notification config to the Config struct makes the configuration described in this document functional. Users can control which notification channels are active and which events trigger notifications, matching the documented interface.

**Impact**: Users can now configure notification behavior through `config.yaml` as documented above. Previously, users who added `notification:` to their config based on this document would have had it silently ignored.

**Related**: [UX-012] - See `docs/decisions/question-log.md` for full discussion

---

## Disabling Notifications

To completely disable notifications:

```yaml
notification:
  enabled: false
```

Or disable specific channels:

```yaml
notification:
  enabled: true
  os_notification:
    enabled: false
  log_notification:
    enabled: true
```

## Project Structure

```
internal/
└── notification/
    ├── notification.go       # Interface and types
    ├── manager.go            # NotificationManager implementation
    ├── os_linux.go           # Linux notify-send backend
    ├── os_darwin.go          # macOS osascript backend
    ├── os_windows.go         # Windows PowerShell backend
    └── log.go                # Log file backend
```
