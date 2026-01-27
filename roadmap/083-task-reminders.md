# [083] Task Reminders

## Summary
Implement task reminder functionality that notifies users when tasks are due or approaching their due date, integrating with the existing notification manager system.

## Documentation Reference
- Primary: `docs/explanation/notification-manager.md`
- Section: "Task reminders (future feature)" listed under Purpose
- Related: `docs/explanation/task-management.md` (Task Dates section)

## Dependencies
- Requires: 022-notification-system.md (notification manager infrastructure)
- Requires: 011-task-dates.md (due date support)

## Complexity
M

## Acceptance Criteria

### Tests Required
- [ ] `TestReminderCreation` - Creating a task with due date schedules reminder
- [ ] `TestReminderNotification` - Reminder triggers OS/log notification at configured time
- [ ] `TestReminderConfig` - Reminder settings configurable (lead time, enabled/disabled)
- [ ] `TestReminderSnooze` - User can snooze/dismiss reminders
- [ ] `TestReminderPersistence` - Reminders survive application restart

### Functional Requirements
- Reminders trigger at configurable time before task due date
- Integrates with existing notification channels (OS notification, log)
- Configurable reminder lead time (e.g., 15 minutes, 1 hour, 1 day before)
- Option to enable/disable reminders per task or globally
- Reminder state persisted across application restarts

## Implementation Notes
- Extend notification manager to support `NotifyReminder` type (already defined in notification-manager.md)
- Add reminder configuration to notification config section
- Consider background scheduler for reminder timing
- Integrate with sync system to sync reminder state across devices

## Out of Scope
- Recurring reminders (handle via recurring tasks feature)
- Location-based reminders
- Integration with external calendar reminder systems
