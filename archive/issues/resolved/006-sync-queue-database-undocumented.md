# [006] Sync Queue Database Location Undocumented

## Type
documentation

## Severity
high

## Source
User report

## Description

The sync queue is stored in a **separate database file** (`~/.todoat/todoat.db`) from the main tasks database (`~/.local/share/todoat/tasks.db`). This is completely undocumented and causes confusion.

## Steps to Reproduce

1. Enable sync in config
2. Add a task (gets queued for sync)
3. Delete `~/.local/share/todoat/tasks.db`
4. Run `todoat sync queue` - queue still shows pending operations!

## Expected Behavior

Either:
1. Sync queue should be in the same database as tasks, OR
2. The separate database location should be documented

## Actual Behavior

- Tasks database: `~/.local/share/todoat/tasks.db`
- Sync queue database: `~/.todoat/todoat.db`
- Users don't know about the second database
- Deleting tasks.db doesn't clear sync state
- Can cause orphaned sync operations for deleted tasks

## Code Location

`cmd/todoat/cmd/todoat.go` - `getSyncManager()` function (line ~6493-6500):

```go
func getSyncManager(cfg *Config) *SyncManager {
    dbPath := cfg.DBPath
    if dbPath == "" {
        home, _ := os.UserHomeDir()
        dbPath = filepath.Join(home, ".todoat", "todoat.db")  // Different location!
    }
    return NewSyncManager(dbPath)
}
```

## Suggested Fix

Option A: Move sync queue to main tasks database
- Store sync_queue table in ~/.local/share/todoat/tasks.db
- Keep all data together

Option B: Document the separate database
- Add to docs/reference/configuration.md
- Explain what each database contains
- Add `sync queue clear` to troubleshooting docs

## Impact

- Users can't properly reset sync state
- Confusion when troubleshooting sync issues
- Orphaned operations when tasks database is reset

## Resolution

**Fixed in**: this session
**Fix description**: Documented the separate database locations (Option B from suggested fix)

### Changes Made

1. **docs/explanation/synchronization.md**: Added new "Database Locations" section with:
   - Table of all database files and their locations
   - Detailed explanation of sync queue database (`~/.todoat/todoat.db`)
   - Backend cache database locations
   - Data management guidance for resetting sync state
   - Backup considerations
   - Troubleshooting orphaned operations

2. **docs/reference/configuration.md**: Added "Data Locations" section with:
   - Table of all data locations
   - Link to synchronization docs for details

### Verification Log
```bash
# Documentation files updated successfully:
$ grep -l "Database Locations" docs/explanation/synchronization.md docs/reference/configuration.md
docs/explanation/synchronization.md
docs/reference/configuration.md

# New section documents the sync queue location:
$ grep "~/.todoat/todoat.db" docs/explanation/synchronization.md docs/reference/configuration.md
docs/explanation/synchronization.md:| **Sync Queue** | `~/.todoat/todoat.db` | Pending sync operations, retry state |
docs/reference/configuration.md:| Sync Queue | `~/.todoat/todoat.db` | Pending sync operations |
```
**Matches expected behavior**: YES - The separate database location is now documented
