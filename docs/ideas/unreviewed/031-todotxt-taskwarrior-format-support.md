# [031] todo.txt and Taskwarrior Format Support

## Summary
Add todo.txt and Taskwarrior data format support to the export/import system, enabling interoperability with the two most widely-used CLI task management ecosystems.

## Source
Code analysis of export/import system (`cmd/todoat/cmd/todoat.go:1516-1946`) — currently supports sqlite, json, csv, and ical formats. No support for the two dominant CLI task management formats (todo.txt, Taskwarrior JSON), creating an interoperability gap for users migrating from or working alongside these tools.

## Motivation
The CLI task management space is dominated by todo.txt (simplicity-focused, plain text) and Taskwarrior (power-user, JSON-based). Users migrating from either tool have no direct import path, and users who want to use todoat alongside these tools cannot round-trip data. Adding these formats would:
- Lower the barrier for users migrating from existing CLI task managers
- Enable todoat to participate in the broader CLI productivity ecosystem (scripts, integrations, and tools that consume/produce these formats)
- Provide familiar export formats for users who want plain-text backups

## Current Behavior
`todoat list export` supports: sqlite, json, csv, ical
`todoat list import` supports: sqlite, json, csv, ical

Users wanting to migrate from todo.txt or Taskwarrior must manually convert their data or write custom scripts.

## Proposed Behavior
`todoat list export --format todotxt` produces a compliant todo.txt file:
```
(A) 2026-01-15 Submit quarterly report +Work @office due:2026-01-31
x 2026-01-10 2026-01-05 Buy groceries +Personal @errands
```

`todoat list import --format todotxt tasks.txt` parses todo.txt format, mapping:
- `(A)-(Z)` priorities to todoat's 1-9 priority scale
- `+Project` to todoat categories/tags
- `@Context` to todoat categories/tags
- `due:YYYY-MM-DD` to due date
- `x` prefix to DONE status

`todoat list export --format taskwarrior` produces Taskwarrior-compatible JSON.
`todoat list import --format taskwarrior` parses Taskwarrior JSON export.

## Estimated Value
medium - Removes migration friction for the two largest CLI task manager user bases; enables ecosystem interop.

## Estimated Effort
M - Both formats are well-documented. todo.txt is a simple line-based format. Taskwarrior uses JSON. Main work is mapping fields between todoat's data model and each format's conventions.

## Related
- Existing export/import infrastructure at `cmd/todoat/cmd/todoat.go:1516-1946`
- [038] List Export/Import (completed roadmap — established the format-extensible export system)
- todo.txt format spec: http://todotxt.org/
- Taskwarrior export format: https://taskwarrior.org/docs/design/task.html

## Status
unreviewed
