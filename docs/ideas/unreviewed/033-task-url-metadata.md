# [033] Task URL/Link Metadata Field

## Summary
Add an optional URL field to tasks, allowing users to associate a web link (GitHub issue, Jira ticket, document, PR, wiki page) with a task for quick reference and context.

## Source
Analysis of the Task struct (`backend/interface.go:18-34`) — the struct has Summary, Description, Status, Priority, DueDate, Categories, Recurrence, but no URL/link field. Many task management workflows originate from or reference web resources. The CalDAV VTODO spec supports a `URL` property, and most cloud task services (Todoist, Google Tasks, Microsoft To Do) support link/URL fields natively.

## Motivation
Tasks frequently relate to external resources: a GitHub issue to fix, a document to review, a PR to merge, a Jira ticket to close. Currently, users must embed URLs in the Description field with no structured access. This means:
- No way to filter or list tasks by whether they have linked resources
- No way for TUI or views to render a clickable/openable link
- No standard field for backends that natively support URLs (CalDAV `URL` property, Todoist `url` field)
- Shell scripts/integrations can't programmatically access the task's related URL

A dedicated URL field solves all of these and aligns with what the underlying protocols already support.

## Current Behavior
Users embed URLs in the task description:
```bash
todoat Work add "Fix login bug" -d "See https://github.com/org/repo/issues/42"
```

The URL is buried in free text with no structured access.

## Proposed Behavior
```bash
# Add task with URL
todoat Work add "Fix login bug" --url "https://github.com/org/repo/issues/42"

# Update task URL
todoat Work update "Fix login bug" --url "https://github.com/org/repo/issues/42"

# View shows URL field
todoat Work get "Fix login bug"
# Summary: Fix login bug
# URL: https://github.com/org/repo/issues/42
# Status: TODO

# Open URL in browser (convenience)
todoat Work open "Fix login bug"
# Opens https://github.com/org/repo/issues/42 in default browser

# JSON output includes URL
todoat Work --json | jq '.[].url'

# Filter tasks that have URLs
todoat Work --has-url
```

Backend mapping:
- SQLite: new `url` column
- Nextcloud/CalDAV: maps to VTODO `URL` property
- Todoist: maps to task `url`/`content` link
- Google Tasks: maps to `links` field
- File/Git: rendered as `URL: <url>` in markdown

## Estimated Value
medium - Natural metadata that many task workflows need. Aligns with existing backend capabilities. Low friction to use.

## Estimated Effort
M - Requires Task struct change (new field), schema migration for SQLite, backend mapping for each of 7 backends, CLI flags, view rendering, and export/import support. Each individual change is small but touches many files.

## Related
- Task struct: `backend/interface.go:18-34`
- CalDAV VTODO `URL` property (RFC 5545)
- SQLite schema migrations: `backend/sqlite/sqlite.go` + `cmd/todoat/cmd/todoat.go:8272`
- [019] Task Activity Log (another metadata extension idea)
- [010] Location-Based Task Context (another metadata field idea)

## Status
unreviewed
