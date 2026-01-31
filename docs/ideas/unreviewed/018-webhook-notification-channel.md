# [018] Webhook/HTTP Notification Channel

## Summary
Add a webhook/HTTP notification backend that sends task reminders and events to external services via HTTP POST requests, enabling integration with Slack, Discord, email relays, or custom endpoints.

## Source
Architecture gap analysis: The notification system (`internal/notification/`) currently only supports OS-level desktop notifications (notify-send on Linux, with macOS/Windows planned in issue #50). There is no way to receive task notifications on a phone, remote machine, or team channel without being at the desktop.

## Motivation
Desktop notifications are only useful when the user is at their computer. Many task management workflows require notifications that reach users regardless of their current device - for example, a Slack reminder about an overdue task, a Discord message for team-shared deadlines, or a simple HTTP callback for home automation. The notification manager already supports multiple backends (OS + log), making webhook a natural addition.

## Current Behavior
```yaml
# config.yaml - notification options
reminder:
  os_notification: true    # Desktop popup (Linux only currently)
  log_notification: true   # Write to log file
  # No remote notification option
```

Reminders only fire as local desktop popups or log entries. Users away from their machine miss all notifications.

## Proposed Behavior
```yaml
# config.yaml - with webhook notification
reminder:
  os_notification: true
  log_notification: true
  webhook_notification: true
  webhook_url: "https://hooks.slack.com/services/T.../B.../xxx"
  # or: "https://ntfy.sh/my-todoat-reminders"
  # or: "http://localhost:9090/webhook"
  webhook_format: "slack"  # slack | discord | ntfy | generic
```

```bash
# Configure via CLI
todoat config set reminder.webhook_notification true
todoat config set reminder.webhook_url "https://ntfy.sh/my-tasks"
todoat config set reminder.webhook_format "ntfy"

# Test the webhook
todoat notification test --channel webhook
```

The webhook sends a JSON POST with task details (summary, due date, list, priority) formatted for the target service.

## Estimated Value
medium - Extends notifications beyond desktop, enabling mobile alerts and team integrations without building a full mobile app. Services like ntfy.sh make this useful with minimal setup.

## Estimated Effort
S - The notification manager (`internal/notification/manager.go`) already abstracts notification delivery. Adding a webhook backend requires an HTTP POST client with configurable URL/format, plus 2-3 format templates (generic JSON, Slack block, ntfy). No architectural changes needed.

## Related
- Notification system: `internal/notification/manager.go`, `notification.go`
- OS notification: `internal/notification/os_linux.go`
- Log notification: `internal/notification/log.go`
- GitHub issue #50: Cross-platform OS notifications (complementary, not overlapping)
- Reminder system: `internal/reminder/reminder.go`

## Status
unreviewed
