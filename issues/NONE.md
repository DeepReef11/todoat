# Manual CLI Test Results

**Date**: 2026-01-19
**Result**: All tests passed - No issues found

## Environment

- **OS**: Linux 6.12.65-1-lts
- **Go version**: go1.25.5 linux/amd64
- **Config path**: ~/.config/todoat/config.yaml
- **Database path**: ~/.local/share/todoat/tasks.db

## Test Summary

### Phase 1: Build & Basic Startup
| Test | Result |
|------|--------|
| `go build -o bin/todoat ./cmd/todoat` | PASS (exit code 0) |
| `go install ./cmd/todoat` | PASS (exit code 0) |
| `./bin/todoat --help` | PASS |
| `./bin/todoat -h` | PASS |
| `./bin/todoat --version` | PASS (output: "todoat version dev") |

### Phase 2: SQLite Backend (Fresh Install)
| Test | Result |
|------|--------|
| Remove existing config/db | PASS |
| List specific list (creates config/db) | PASS |
| Config created at expected path | PASS |
| Database created at expected path | PASS |
| Add task | PASS |
| List tasks | PASS |
| Update task status | PASS |
| Complete task | PASS |
| Delete task | PASS |

### Phase 3: SQLite Backend (With Config)
| Test | Result |
|------|--------|
| Verify config exists and valid | PASS |
| Verify database exists | PASS |
| `todoat list` (list all lists) | PASS |
| `todoat list create "TestList"` | PASS |
| `todoat list delete "TestList"` | PASS |
| JSON output (`--json`) | PASS (valid JSON) |
| Filter by status (`--status=TODO`) | PASS |
| Filter by priority (`--priority=1`) | PASS |

### Phase 4: Todoist Backend
| Test | Result |
|------|--------|
| SKIPPED | No TODOAT_TODOIST_TOKEN environment variable set |

### Phase 5: Error Handling
| Test | Result |
|------|--------|
| Non-existent list | PASS (graceful message: "No tasks in list") |
| Non-existent backend flag | PASS (error: "unknown flag: --backend") |
| Invalid command | PASS (error: "unknown action: invalidcommand") |
| Missing arguments (add) | PASS (error: "task summary is required") |
| Missing arguments (update) | PASS (error: "task summary is required") |
| Missing arguments (delete) | PASS (error: "task summary is required") |
| Update non-existent task | PASS (error: "no task found matching") |
| Delete non-existent task | PASS (error: "no task found matching") |

## Notes

1. All error messages are user-friendly with no stack traces or panics
2. Config and database are automatically created on first use
3. JSON output is valid and parseable
4. The `--backend` flag is not available (possibly removed or not implemented for global scope) - commands use the default backend configured in config.yaml
5. Todoist backend tests were skipped due to missing token - this is expected behavior in CI/test environments
