# [007] Enhanced Bulk Operations CLI

## Summary
Extend bulk operations to support piping task IDs/UIDs and applying complex operations to filtered task sets in a Unix-friendly way.

## Source
Code analysis: Current bulk operations exist but are limited. Terminal users expect Unix-style composability (piping, xargs patterns). Current implementation doesn't fully support these workflows.

## Motivation
Power users want to compose todoat with other Unix tools. Operations like "complete all tasks tagged 'done-review'" or "delete all overdue tasks" should be expressible as one-liners that can be scripted and scheduled.

## Current Behavior
```bash
# Basic filtering exists but limited bulk operations
todoat Work --filter-status TODO --filter-tag review
# Shows filtered tasks but can't easily operate on them
```

## Proposed Behavior
```bash
# Get task UIDs for piping
todoat Work list --filter-tag done --output-uids
# Output: uid1 uid2 uid3 (one per line)

# Pipe to bulk operations
todoat Work list --filter-status DONE --output-uids | xargs -I{} todoat Work delete --uid {}

# Built-in bulk complete/delete/update
todoat Work complete --all --filter-tag "done-review"
todoat Work delete --all --filter-status DONE --filter-created-before "30 days ago"

# Bulk update
todoat Work update --all --filter-priority 9 --set-priority 5

# Dry run for safety
todoat Work delete --all --filter-status DONE --dry-run
# Output: Would delete 15 tasks:
#   - Old task 1
#   - Old task 2

# Confirm prompt for destructive bulk ops
todoat Work delete --all --filter-status DONE
# Output: This will delete 15 tasks. Continue? [y/N]

# Force skip confirmation
todoat Work delete --all --filter-status DONE --force
```

## Estimated Value
medium - Essential for scripting and automation, matches Unix philosophy

## Estimated Effort
S - Builds on existing filter infrastructure, mainly adding output modes and bulk operation wrappers

## Open Questions
- Output format for UIDs (newline-separated, JSON array)?
- Safety defaults (require --force or --dry-run first)?
- Rate limiting for bulk remote operations?
- Transaction support (rollback on partial failure)?

## Related
- Existing filters: cmd/todoat/cmd/todoat.go filter flags
- Scripting docs: docs/README.md scripting section

## Status
unreviewed
