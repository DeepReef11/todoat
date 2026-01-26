# [003] Calendar Integration Output (ICS Feed)

## Summary
Provide an HTTP server or file watcher that serves tasks as an iCalendar feed, allowing calendar apps to subscribe and display tasks alongside events.

## Source
Feature gap analysis: The app already supports iCal export (`list export --format ical`), but this is one-time export. Users want live integration with calendar apps like Google Calendar, Apple Calendar, or Outlook.

## Motivation
Many users manage their time using calendar apps. Having tasks appear in the same calendar view as meetings helps with time blocking and realistic planning. A subscribable ICS feed would make todoat tasks visible in any calendar application without manual export/import cycles.

## Current Behavior
```bash
# One-time export to file
todoat list export Work --format ical -o work-tasks.ics
# Must manually import into calendar, repeat when tasks change
```

## Proposed Behavior
```bash
# Start local server serving ICS feed
todoat serve --port 8080
# Calendar apps subscribe to: http://localhost:8080/lists/Work.ics

# Or generate static ICS with file watcher for auto-update
todoat watch --output ~/calendars/
# Creates/updates: ~/calendars/Work.ics, ~/calendars/Personal.ics
# Calendar apps point to local file

# Filter what appears in feed
todoat serve --filter-status "TODO,IN-PROGRESS" --filter-due-before "+7d"
```

## Estimated Value
medium - Bridges gap between task management and calendar/time management workflows

## Estimated Effort
M - HTTP server setup, iCal generation already exists, need endpoint routing and optional filtering

## Open Questions
- HTTP server vs file watcher approach (or both)?
- Authentication for remote access?
- How to handle task updates (polling interval, webhooks)?
- Should completed tasks appear (with COMPLETED status) or be hidden?

## Related
- Existing iCal export: cmd/todoat/cmd/todoat.go `exportICalendar()`
- CalDAV support already exists for Nextcloud backend

## Status
unreviewed
