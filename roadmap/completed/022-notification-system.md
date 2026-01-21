# [022] Notification System

## Summary
Implement the notification manager for alerting users about background sync events through OS-native desktop notifications and a persistent log file.

## Documentation Reference
- Primary: `docs/explanation/notification-manager.md`

## Dependencies
- Requires: [018] Synchronization Core System

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestNotificationTest` - `todoat notification test` sends test notification and exits 0
- [ ] `TestNotificationLog` - `todoat notification log` displays notification history
- [ ] `TestNotificationLogClear` - `todoat notification log clear` clears the log file
- [ ] `TestOSNotificationLinux` - OS notification sent via notify-send on Linux (mock)
- [ ] `TestOSNotificationDarwin` - OS notification sent via osascript on macOS (mock)
- [ ] `TestLogNotification` - Notifications written to log file with correct format
- [ ] `TestNotificationConfig` - Configuration enables/disables notification channels

## Implementation Notes
- Create `internal/notification/` package with NotificationManager interface
- Implement OS-specific backends: `os_linux.go`, `os_darwin.go`, `os_windows.go`
- Implement log file backend in `log.go`
- Integration with sync manager for automatic notifications
- Support notification types: sync_complete, sync_error, conflict, reminder
- Log format: `2026-01-16T10:30:00Z [TYPE] Message`
- Log rotation when file exceeds configured size

## Out of Scope
- Task reminders (future feature per docs)
- Push notifications
- Email notifications
- Mobile notifications
