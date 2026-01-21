# Reminders

todoat supports task reminders to notify you before tasks are due. This guide covers configuring and using the reminder system.

## Overview

Reminders work with tasks that have due dates. When a task is due within a configured interval, you can receive notifications to stay on track.

## Viewing Reminders

### List Upcoming Reminders

```bash
todoat reminder list
```

Shows tasks with upcoming due dates within the configured reminder intervals.

### Check for Due Reminders

```bash
todoat reminder check
```

Checks all tasks and sends reminders for those due within the configured intervals.

## Managing Reminders

### Disable Reminders for a Task

```bash
todoat reminder disable "Task name"
```

Stops reminders for a specific task.

### Dismiss a Reminder

```bash
todoat reminder dismiss "Task name"
```

Dismisses the current reminder for a task. You'll be reminded again at the next interval.

### View Reminder Status

```bash
todoat reminder status
```

Shows the current reminder configuration and status.

## Configuration

Configure reminders in `~/.config/todoat/config.yaml`:

```yaml
reminder:
  enabled: true
  intervals:
    - 1d    # 1 day before due
    - 1h    # 1 hour before due
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `enabled` | Enable reminder system | `true` |
| `intervals` | Time before due to send reminders | `["1d", "1h"]` |

### Interval Format

| Format | Meaning |
|--------|---------|
| `15m` | 15 minutes before |
| `1h` | 1 hour before |
| `1d` | 1 day before |
| `1w` | 1 week before |

## Automated Reminder Checks

### Using Cron

Run reminder checks periodically:

```bash
# Check reminders every 15 minutes
*/15 * * * * /path/to/todoat reminder check
```

### With Sync Daemon

If using the sync daemon, reminders are checked automatically during sync cycles.

## Examples

### Daily Workflow

```bash
# Morning: Check what's due today
todoat reminder list

# Dismiss reminders you've seen
todoat reminder dismiss "Meeting prep"

# Check again later
todoat reminder check
```

### Disable Reminders for Repeating Tasks

```bash
# Some recurring tasks don't need reminders
todoat reminder disable "Daily standup"
```

## See Also

- [Task Management](task-management.md) - Setting due dates on tasks
- [Synchronization](sync.md) - Background sync and notifications
