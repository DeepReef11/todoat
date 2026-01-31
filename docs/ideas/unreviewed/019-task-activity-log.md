# [019] Task Activity Log

## Summary
Add an append-only activity log per task that records status changes, edits, and user notes with timestamps, providing a history trail for task progression.

## Source
Competitive analysis: Most task management tools (Todoist, Jira, Linear, GitHub Issues) maintain an activity history showing when tasks were modified, status changed, or commented on. Todoat currently overwrites task fields with no record of what changed or when.

## Motivation
When working on tasks over multiple days or weeks, users lose context about what happened previously. Questions like "when did I start this?", "why was this re-opened?", or "what did I try last time?" have no answer in the current system. An activity log also enables audit trails for shared backends and helps users remember context when returning to a task after a break.

## Current Behavior
```bash
todoat Work update "Fix login bug" --status IN-PROGRESS
# Previous status (TODO) is lost
# No record of when the change happened

todoat Work update "Fix login bug" --description "New approach: use OAuth"
# Previous description is overwritten
# No record of old approach
```

Tasks have only their current state. The `Modified` timestamp shows the last change but not what changed.

## Proposed Behavior
```bash
# Add a note to a task (most common use case)
todoat Work note "Fix login bug" "Tried session-based approach, didn't work with mobile clients"
todoat Work note "Fix login bug" "Switching to OAuth flow per team discussion"

# View activity log
todoat Work log "Fix login bug"
# Output:
# 2026-01-28 09:15  Created (status: TODO)
# 2026-01-28 14:30  Status changed: TODO → IN-PROGRESS
# 2026-01-28 14:30  Note: Tried session-based approach, didn't work with mobile clients
# 2026-01-29 10:00  Note: Switching to OAuth flow per team discussion
# 2026-01-29 10:05  Description updated
# 2026-01-31 16:00  Status changed: IN-PROGRESS → COMPLETED

# JSON output for scripting
todoat Work log "Fix login bug" --json
```

Activity entries are stored locally (SQLite) even for remote backends, since most remote APIs don't support arbitrary notes on tasks.

## Estimated Value
medium - Provides context continuity across work sessions and decision history. Particularly valuable for long-running tasks and tasks returned to after breaks.

## Estimated Effort
M - Requires a new SQLite table for activity entries (keyed by task UID), a `note` subcommand, a `log` subcommand, and hooks into existing update/complete/delete operations to auto-record status changes. Remote backends don't need modification since logs are stored locally.

## Related
- Task model: `backend/interface.go` (Task struct)
- SQLite backend: `backend/sqlite/` (storage layer)
- Existing task commands: `cmd/todoat/cmd/todoat.go` (task CRUD)
- Idea 008 (Task Archival): complementary - archival preserves completed tasks, activity log preserves the journey

## Status
unreviewed
