# [017] Command File Decomposition

## Summary
Split the monolithic `cmd/todoat/cmd/todoat.go` (12,500+ lines) into focused subcommand files to improve maintainability and reduce merge conflicts.

## Source
Code pattern analysis: `todoat.go` is 12,507 lines and was modified 18 times in the last 50 commits, making it the most actively changed file by a 3x margin. This concentration creates real friction for development.

## Motivation
A single 12,500-line file containing all CLI command definitions, flag parsing, and handler logic makes it difficult to work on any one command without navigating thousands of unrelated lines. With 18 modifications in 50 recent commits, this file is a hotspot for merge conflicts and cognitive overhead. Decomposing it would make each subsystem independently navigable and testable.

## Current Behavior
All CLI commands (task CRUD, list management, sync, views, credentials, config, reminders, analytics, TUI, migration, notifications, tags, completion) are defined in a single file. Developers must scroll through 12,500+ lines to find relevant sections.

## Proposed Behavior
Split into focused files by command group:
```
cmd/todoat/cmd/
  root.go          # Root command, global flags, shared helpers
  task.go          # Task CRUD (add, get, update, complete, delete)
  list.go          # List management (create, update, delete, info, trash, export/import, stats, vacuum)
  sync.go          # Sync operations (manual sync, status, queue, conflicts, daemon)
  view.go          # View management (list, create)
  config.go        # Config commands
  credentials.go   # Credential management
  reminder.go      # Reminder commands
  analytics.go     # Analytics commands
  tui.go           # TUI launch
  migrate.go       # Migration commands
  notification.go  # Notification commands
  tags.go          # Tag operations
  completion.go    # Shell completion
  helpers.go       # Shared formatting, output, validation helpers
```

Each file registers its commands with the root command using `init()` or explicit registration functions.

## Estimated Value
high - Directly reduces development friction on the most-modified file. Improves code navigation, reduces merge conflicts, and makes individual subsystems easier to test and review.

## Estimated Effort
L - Mechanical refactoring but requires careful attention to shared state (global flags, output helpers, backend initialization). No behavior changes, but extensive file movement and import adjustments. Should be done incrementally per command group.

## Related
- Main command file: `cmd/todoat/cmd/todoat.go` (12,507 lines)
- Test file: `cmd/todoat/cmd/todoat_test.go`
- Go CLI pattern: cobra library supports multi-file command registration

## Status
unreviewed
