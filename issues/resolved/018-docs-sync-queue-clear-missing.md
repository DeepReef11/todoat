# [018] Docs: 'sync queue clear' subcommand is tested but not documented

## Type
documentation

## Severity
low

## Test Location
- File: cmd/todoat/cmd/sync_test.go (or similar)
- Function: TestSyncQueueClearCLI

## Feature Description
The `todoat sync queue clear` command is tested but not documented in the CLI reference. The test verifies that this command removes all pending sync operations from the queue.

## Current Documentation
The CLI reference (docs/reference/cli.md) documents:
- `sync status` - Show sync status
- `sync queue` - View pending sync operations
- `sync conflicts` - View and manage sync conflicts
- `sync daemon` - Manage the sync daemon

But does not document:
- `sync queue clear` - Clear pending sync operations

## Expected Documentation
- Location: docs/reference/cli.md
- Section: Under `sync` command, add `sync queue` subsection with subcommands

Should add:
```markdown
### sync queue

View and manage the sync queue.

```bash
todoat sync queue [command]
```

| Subcommand | Description |
|------------|-------------|
| (default) | View pending operations |
| `clear` | Clear all pending operations |
```

Also update docs/how-to/sync.md with an example:
```bash
# Clear all pending sync operations
todoat sync queue clear
```

## Resolution

**Fixed in**: this session
**Fix description**: Added documentation for `sync queue clear` subcommand

### Changes Made
1. **docs/reference/cli.md**:
   - Updated `sync queue` description from "View pending sync operations" to "View and manage pending sync operations"
   - Added `### sync queue` subsection with subcommand table showing `(default)` and `clear` options
   - Added `todoat sync queue clear` to the sync examples section

2. **docs/how-to/sync.md**:
   - Added `### Clear Sync Queue` section with example command and description

### Verification Log
```bash
$ grep -A5 "### sync queue" docs/reference/cli.md
### sync queue

View and manage the sync queue.

```bash
todoat sync queue [command]
```

$ grep -A5 "### Clear Sync Queue" docs/how-to/sync.md
### Clear Sync Queue

```bash
todoat sync queue clear
```

Removes all pending sync operations from the queue. Use this when you want to discard unsynced local changes.
```
**Matches expected behavior**: YES
