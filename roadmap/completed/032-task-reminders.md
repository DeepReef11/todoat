# [032] Task Reminders

## Summary
Implement task reminder notifications that alert users when tasks are approaching or past their due dates, with configurable reminder intervals.

## Documentation Reference
- Primary: `docs/explanation/notification-manager.md` (future feature: Task reminders)
- Secondary: `docs/explanation/task-management.md` (due dates)

## Dependencies
- Requires: [022] Notification System
- Requires: [011] Task Dates

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestReminderConfig` - Reminder settings can be configured in config.yaml
- [ ] `TestReminderNotification` - Reminder notification sent when task due date approaches
- [ ] `TestReminderIntervals` - Configurable intervals (1 day, 1 hour, at due time)
- [ ] `TestReminderDisable` - Individual reminders can be disabled
- [ ] `TestReminderList` - `todoat reminder list` shows upcoming reminders
- [ ] `TestReminderDismiss` - `todoat reminder dismiss <task>` dismisses a reminder

### Notification Integration
- [ ] `TestReminderOSNotification` - Reminders sent via OS notification
- [ ] `TestReminderLogNotification` - Reminders logged to notification log
- [ ] `TestReminderFormat` - Notification includes task summary and due date

## Implementation Notes

### Configuration
```yaml
notification:
  reminder:
    enabled: true
    intervals:
      - "1 day"
      - "1 hour"
      - "at due time"
    os_notification: true
    log_notification: true
```

### CLI Commands
```bash
# List upcoming reminders
todoat reminder list

# Dismiss reminder for a task
todoat reminder dismiss "Task name"

# Snooze reminder
todoat reminder snooze "Task name" --duration "1h"
```

### Implementation Approach
- Reminder check runs during sync operations
- Compare task due dates with current time and intervals
- Track dismissed/snoozed reminders in SQLite
- Use existing notification infrastructure

### Files to Create/Modify
1. `internal/reminder/` - Reminder logic package
2. `cmd/todoat/reminder.go` - Reminder CLI commands
3. Modify sync manager to check reminders

## Out of Scope
- Recurring reminders (separate feature)
- Calendar integration for reminders
- SMS/email reminders
- Mobile push notifications
