# 006: CLI Test Infrastructure

Establish the CLI test framework following TEST_DRIVEN_DEV.md patterns, including injectable IO, result code assertions, and in-memory backend support for testing.

## Dependencies

- `002-core-cli.md` - CLI with injectable IO
- `003-sqlite-backend.md` - Backend for test data
- `004-task-commands.md` - Commands to test
- `005-status-system.md` - Status handling to verify

## Acceptance Criteria

- [ ] **Test configuration** support:
  - Tests can inject a config with in-memory SQLite database
  - Config specifies `no_prompt: true` by default for tests
  - Tests isolated from user's real config/database

- [ ] **Injectable IO pattern** established:
  ```go
  func TestListTasks(t *testing.T) {
      var stdout, stderr bytes.Buffer
      cfg := testConfig()

      exitCode := Execute([]string{"-y", "TestList"}, &stdout, &stderr, cfg)

      if exitCode != 0 {
          t.Fatalf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr.String())
      }
  }
  ```

- [ ] **Result codes** in output:
  - `ACTION_COMPLETED` - operation performed successfully
  - `INFO_ONLY` - display-only operation (get/list)
  - `ERROR` - operation failed
  - Result code appears as last line of output in no-prompt mode

- [ ] **Test helper functions**:
  - `testConfig()` - returns config with in-memory DB
  - `setupTestData(cfg, tasks)` - populates test tasks
  - `assertContains(t, output, expected)` - output assertion helper
  - `assertExitCode(t, got, want)` - exit code assertion

- [ ] **Core test cases** implemented:
  - List empty list (returns INFO_ONLY)
  - Add task (returns ACTION_COMPLETED)
  - List tasks after adding (shows task, returns INFO_ONLY)
  - Update task (returns ACTION_COMPLETED)
  - Complete task (returns ACTION_COMPLETED, status changes to DONE)
  - Delete task (returns ACTION_COMPLETED)
  - Delete with confirmation skipped in no-prompt mode
  - Error cases return ERROR result code

- [ ] **Exit codes** verified:
  - `0` for ACTION_COMPLETED and INFO_ONLY
  - `1` (or non-zero) for ERROR

- [ ] Tests run with `go test ./cmd/todoat -v`

- [ ] All existing functionality covered by CLI tests

## Complexity

**Estimate:** M (Medium)

## Implementation Notes

- Reference: `dev-doc/TEST_DRIVEN_DEV.md` for test patterns
- Reference: `dev-doc/CLI_INTERFACE.md#result-codes` for result code specs
- CLI tests should be the primary test suite (per TEST_DRIVEN_DEV.md)
- Use in-memory SQLite (`:memory:`) for fast, isolated tests
- Tests should use `-y` flag to ensure deterministic behavior
- Consider table-driven tests for command variations

### Test File Location

```
cmd/todoat/
├── main.go
├── todoat.go
└── todoat_test.go    # CLI tests go here
```

### Example Test Structure

```go
func TestAddTask(t *testing.T) {
    var stdout, stderr bytes.Buffer
    cfg := testConfig()

    // Ensure list exists
    setupTestList(cfg, "Work")

    exitCode := Execute([]string{"-y", "Work", "add", "Test task"}, &stdout, &stderr, cfg)

    if exitCode != 0 {
        t.Fatalf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr.String())
    }

    if !strings.Contains(stdout.String(), "ACTION_COMPLETED") {
        t.Errorf("expected ACTION_COMPLETED result code")
    }
}

func TestListTasks(t *testing.T) {
    var stdout, stderr bytes.Buffer
    cfg := testConfig()

    setupTestList(cfg, "Work")
    setupTestTask(cfg, "Work", "Existing task")

    exitCode := Execute([]string{"-y", "Work"}, &stdout, &stderr, cfg)

    if exitCode != 0 {
        t.Fatalf("expected exit code 0, got %d", exitCode)
    }

    output := stdout.String()
    if !strings.Contains(output, "Existing task") {
        t.Errorf("expected task in output")
    }
    if !strings.Contains(output, "INFO_ONLY") {
        t.Errorf("expected INFO_ONLY result code")
    }
}
```

### Result Code Output Format

In no-prompt mode, the last line of stdout should be the result code:

```
Task 'Test task' added to list 'Work'
UID: 550e8400-e29b-41d4-a716-446655440000
ACTION_COMPLETED
```

Or for errors:

```
Error 1: List 'NonExistent' not found
ERROR
```
