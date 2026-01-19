# Manual CLI Testing Results

**Date**: 2026-01-19
**Go Version**: go1.25.5 linux/amd64
**todoat Version**: dev

## Summary

All manual CLI tests passed successfully. No critical issues found.

## Test Results

### Phase 1: Build & Basic Startup

| Test | Result |
|------|--------|
| `go build -o bin/todoat ./cmd/todoat` | Exit code: 0 |
| `go install ./cmd/todoat` | Exit code: 0 |
| `./bin/todoat -y --help` | Exit code: 0, shows full help |
| `./bin/todoat -y -h` | Exit code: 0, shows full help |
| `./bin/todoat -y --version` | Exit code: 0, shows "todoat version dev" |

### Phase 2: SQLite Backend (Fresh Install)

Tested with clean slate (removed ~/.config/todoat and ~/.local/share/todoat).

| Test | Result |
|------|--------|
| `./bin/todoat -y` | Exit code: 0, shows help |
| `./bin/todoat -y MyList` | Exit code: 0, creates list, shows "No tasks" |
| `./bin/todoat -y MyList add "Test task"` | Exit code: 0, task created with UUID |
| `./bin/todoat -y MyList` | Exit code: 0, shows task |
| `./bin/todoat -y MyList update "Test task" -s DONE` | Exit code: 0, task updated |
| `./bin/todoat -y MyList complete "Test task"` | Exit code: 0, task completed |
| `./bin/todoat -y MyList delete "Test task"` | Exit code: 0, task deleted |

**Config created**: ~/.config/todoat/config.yaml
**Database created**: ~/.local/share/todoat/tasks.db

### Phase 3: SQLite Backend (With Config)

| Test | Result |
|------|--------|
| `./bin/todoat -y list` | Exit code: 0, shows available lists |
| `./bin/todoat -y list create "TestList"` | Exit code: 0, list created |
| `./bin/todoat -y list delete "TestList"` | Exit code: 0, list deleted |
| `./bin/todoat -y --json MyList` | Exit code: 0, valid JSON output |
| `./bin/todoat -y --json MyList add "JSON test"` | Exit code: 0, valid JSON output |
| `./bin/todoat -y MyList --status=TODO` | Exit code: 0, filters correctly |
| `./bin/todoat -y MyList --priority=1` | Exit code: 0, filters correctly |

### Phase 4: Todoist Backend

The Todoist backend is intentionally **not implemented as a primary backend**. Per design documentation (issues/resolved/003-no-backend-flag-documented.md):

1. SQLite is the only supported primary backend for day-to-day operations
2. Todoist is available as a **migration target only**
3. Migration to Todoist returns: "real todoist backend not yet implemented for migration" - this is intentional placeholder code

This is documented behavior, not a bug.

**Tests Run:**
| Test | Result |
|------|--------|
| `./bin/todoat -y --detect-backend` | Exit code: 0, shows sqlite as available |
| `./bin/todoat -y credentials list` | Exit code: 0, shows todoist as "Not configured" |
| `./bin/todoat -y migrate --dry-run --from=sqlite --to=todoist` | Exit code: 1, expected "not yet implemented" message |

### Phase 5: Error Handling

| Test | Expected | Result |
|------|----------|--------|
| `./bin/todoat -y NonExistentList12345` | Graceful message | "No tasks in list" - Exit code: 0 |
| `./bin/todoat -y MyList invalidcommand` | Error message | "Error: unknown action" - Exit code: 1 |
| `./bin/todoat -y MyList add` | Error message | "Error: task summary is required" - Exit code: 1 |
| `./bin/todoat -y MyList update` | Error message | "Error: task summary is required" - Exit code: 1 |
| `./bin/todoat -y MyList delete` | Error message | "Error: task summary is required" - Exit code: 1 |
| `./bin/todoat -y MyList complete` | Error message | "Error: task summary is required" - Exit code: 1 |
| `./bin/todoat -y --invalid-flag` | Error message | "Error: unknown flag" - Exit code: 1 |

All error cases handled gracefully with clear messages. No panics or stack traces observed.

## Notes

1. **Backend selection**: There is no `--backend` flag. The app uses auto-detection or config file settings. This is by design.

2. **Todoist integration**: Todoist backend code exists but is only partially implemented. It cannot be used as a primary backend or for migrations yet. This is documented.

3. **All CRUD operations** work correctly with the SQLite backend.

4. **JSON output** produces valid, parseable JSON.

5. **Error handling** is robust with user-friendly messages.

6. **Config auto-creation** works correctly on fresh install.

## Conclusion

The todoat CLI is functioning correctly for all tested scenarios with the SQLite backend. The Todoist backend limitations are documented and intentional.
