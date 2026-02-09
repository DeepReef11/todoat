# [027] Backend Capability Discovery

## Summary
Add a command to inspect and display what operations each backend supports, enabling users to understand backend limitations before encountering errors.

## Source
Code analysis: Each backend implements a different subset of TaskManager interface capabilities. Nextcloud cannot create lists (CalDAV limitation), some backends don't support custom fields, priorities may map differently. These limitations are only discovered through runtime errors (e.g., `backend.ErrListCreationNotSupported`). No way to proactively check.

## Motivation
Users configuring multiple backends need to understand what each can do:
- Which backends support list creation?
- Can this backend handle recurring tasks?
- What priority values are valid?
- Does it support task hierarchy (subtasks)?
- What happens to custom fields when syncing between backends?

Currently this requires trial-and-error or reading source code. A capability discovery command would make this explicit and accessible.

## Current Behavior
```bash
# User tries to create a list on Nextcloud
todoat list create "New List" --backend nextcloud
# Error: list creation not supported for nextcloud backend

# No way to know beforehand what operations are supported
# Must try operations and see what fails
```

## Proposed Behavior
```bash
# Show capabilities for a specific backend
todoat backend capabilities nextcloud
# Output:
# Backend: nextcloud (Nextcloud CalDAV)
#
# Tasks:
#   Create:     yes
#   Read:       yes
#   Update:     yes
#   Delete:     yes
#   Subtasks:   yes
#   Recurring:  yes (RRULE format)
#
# Lists:
#   Create:     no (CalDAV limitation - create calendars in Nextcloud UI)
#   Read:       yes
#   Update:     limited (name only)
#   Delete:     no
#
# Fields:
#   Priority:   1-9 (CalDAV standard)
#   Due date:   yes
#   Start date: yes
#   Tags:       yes (CATEGORIES)
#   Custom:     no
#
# Sync:
#   Bidirectional: yes
#   ETag support:  yes
#   Offline queue: yes

# Compare capabilities across backends
todoat backend capabilities --compare
# Output:
#                  nextcloud  todoist  sqlite  mstodo
# List create         -         +        +       +
# Subtasks            +         +        +       +
# Priority range    1-9       1-4      1-9     low/med/high
# Recurring           +         +        +       +
# Custom fields       -         -        +       -

# Show capabilities for configured backends
todoat backend capabilities --all

# JSON output for scripting
todoat backend capabilities nextcloud --json
# {"backend": "nextcloud", "capabilities": {"tasks": {"create": true, ...}}}
```

## Estimated Value
medium - Reduces trial-and-error, helps users choose backends and understand sync implications

## Estimated Effort
S - Capability data already implicit in code (error returns, interface implementations); needs extraction into capability registry and display

## Related
- Idea #023 (Config Validation) - validates config correctness, this validates capability understanding
- Backend interface: `backend/backend.go` TaskManager interface
- Error constants: `backend.ErrListCreationNotSupported` pattern
- Backend implementations: `backend/*/` each has different capabilities

## Status
unreviewed
