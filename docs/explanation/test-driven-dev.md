# todoat Development Guide - TDD Approach

This document explains how todoat should be developed using Test-Driven Development. It covers the TDD workflow, not the application features themselves.

## Core Principle: CLI-First Testing

**All features should be tested through the CLI first.** The CLI is the primary interface users interact with, so tests should verify behavior from that perspective.

- **CLI tests are the primary test suite** - They validate the full user experience
- **Unit tests supplement CLI tests** - Only for complex internal logic that can't be easily tested via CLI
- **Backend tests** - Used for integration testing with real services

If a feature can be tested through the CLI, it should be. Only fall back to unit tests when:
- Testing internal algorithms or data structures
- Testing edge cases that are difficult to trigger via CLI
- Testing low-level backend protocol handling

## Generic Project Structure
```
todoat/
├── cmd/
│   └── todoat/
│       ├── main.go           # Thin entry point
│       ├── todoat.go           # Cobra setup with injectable IO
├── backend/
│   ├── interface.go          # Backend interface definition
│   ├── sync/
│   │   ├── manager.go        # Sync logic
│   ├── sqlite/
│   │   ├── sqlite.go         # SQLite implementation
│   ├── nextcloud/
│   │   ├── nextcloud.go      # Nextcloud CalDAV implementation
│   ├── todoist/
│   │   ├── todoist.go        # Todoist API implementation
│   └── file/
│       ├── file.go           # Markdown file backend
├── internal/
│   ├── app/              # Core application logic
│   ├── config/           # Configuration handling
│   ├── cli/              # CLI display and prompts
│   │   └── prompt/       # Prompt manager (no-prompt mode)
│   ├── operations/       # Task operations (add, delete, update, etc.)
│   └── views/            # Custom view rendering
```

## Development Flow

1. **Write CLI test first** - Define expected behavior through CLI commands
2. **Run test** - It fails (no code yet)
3. **Implement minimal code** - Make test pass
4. **Refactor** - Clean up if needed
5. **Run full test suite** - Ensure no regressions
6. **Open PR** - When tests are green


## Integration Test Guidelines

### File Structure
```go
//go:build integration

package mypackage

func TestNextcloudConnection(t *testing.T) {}
func TestTodoistSync(t *testing.T) {}
func TestGoogleSync(t *testing.T) {}
```

## Running Tests
```bash
# All integration tests
go test -tags=integration

# Specific service
go test -tags=integration -run Nextcloud
go test -tags=integration -run Todoist
go test -tags=integration -run Google
```

## Naming Convention

- Use single `integration` build tag
- Name tests with service prefix: `Test{Service}{Feature}`
- Filter by service using `-run` flag

## Running Tests

```bash
# Run all tests (CLI tests are primary)
go test ./...

# Run CLI tests only (primary test suite)
go test ./cmd/todoat -v

# Run specific backend tests
go test ./backend/sqlite -v
go test ./backend/nextcloud -v
go test ./backend/todoist -v

# Run with Docker test server (for Nextcloud integration tests)
make docker-up
go test ./backend/nextcloud -v -tags=integration
```

## Code Template

This is only an example:

### cmd/todoat/main.go
```go
package main

import (
    "os"
    "todoat/cmd/todoat"
)

func main() {
    os.Exit(cmd.Execute(os.Args[1:], os.Stdout, os.Stderr))
}
```

### cmd/todoat/todoat.go

This is only an example:
```go
package cmd

import (
    "io"
    "github.com/spf13/cobra"
)

func Execute(args []string, stdout, stderr io.Writer, cfg ...*Config) int {
    var rootCmd *cobra.Command
    
    rootCmd = NewTodoAt(stdout, stderr,*cfg[0])
    
    rootCmd.SetArgs(args)
    rootCmd.SetOut(stdout)
    rootCmd.SetErr(stderr)
    
    if err := rootCmd.Execute(); err != nil {
        return 1
    }
    return 0
}


func NewTodoAt(stdout, stderr io.Writer,cfg Config) *cobra.Command {
    cmd := &cobra.Command{
        Use: "todoat",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Call business logic here
            if(cfg != nil) {
                //TODO use provided cfg
            }
            else {
                //TODO default user config selection/generation
            }
            // TODO
            return nil
        },
    }

    // Add flags: --no-prompt, --json, --backend, etc.
    cmd.PersistentFlags().BoolP("no-prompt", "y", false, "Disable interactive prompts")
    cmd.PersistentFlags().Bool("json", false, "Output in JSON format")

    return cmd
}

func NewTodoAtWithConfig(cfg Config, stdout, stderr io.Writer) *cobra.Command {
    cmd := &cobra.Command{
        Use: "myapp",
        RunE: func(cmd *cobra.Command, args []string) error {
            return run(cfg, stdout)
        },
    }
    
    return cmd
}

```

### Example
```go
package cmd

import (
    "bytes"
    "testing"
)

func TestListTasks(t *testing.T) {
    var stdout, stderr bytes.Buffer
    var cfg = ... //TODO

    // Tests run in no-prompt mode by default
    exitCode := Execute([]string{"-y", "MyList"}, &stdout, &stderr, cfg)

    if exitCode != 0 {
        t.Fatalf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr.String())
    }

    // Check for INFO_ONLY result code
    if !strings.Contains(stdout.String(), "INFO_ONLY") {
        t.Errorf("expected INFO_ONLY result code")
    }
}

func TestAddTask(t *testing.T) {
    var stdout, stderr bytes.Buffer
    var cfg = ... //TODO

    exitCode := Execute([]string{"-y", "MyList", "add", "Test task"}, &stdout, &stderr, cfg)

    if exitCode != 0 {
        t.Fatalf("expected exit code 0, got %d", exitCode)
    }

    if !strings.Contains(stdout.String(), "ACTION_COMPLETED") {
        t.Errorf("expected ACTION_COMPLETED result code")
    }
}

func TestJSONOutput(t *testing.T) {
    var stdout, stderr bytes.Buffer
    var cfg = ... //TODO 

    exitCode := Execute([]string{"-y", "--json", "MyList"}, &stdout, &stderr, cfg)

    if exitCode != 0 {
        t.Fatalf("expected exit code 0, got %d", exitCode)
    }

    // Parse JSON output
    var response map[string]interface{}
    if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
        t.Fatalf("invalid JSON output: %v", err)
    }

    if response["result"] != "INFO_ONLY" {
        t.Errorf("expected result INFO_ONLY, got %v", response["result"])
    }
}
```

## Key Principles

- **Inject IO**: Pass `stdout`/`stderr` to all functions that write output
- **SetArgs**: Use `cmd.SetArgs(args)` in tests to simulate CLI flags
- **Same path**: Tests execute identical code as real CLI invocation
- **Test first**: Write the test showing desired behavior before implementation
- **Thin main**: Keep `main.go` minimal, no logic
- **No-prompt by default**: Tests use `-y` flag to avoid stdin dependencies
- **Result codes**: Assert on `ACTION_COMPLETED`, `ACTION_INCOMPLETE`, `INFO_ONLY`, `ERROR` in output


### Testing Error Handling

```go
func TestErrorNotFound(t *testing.T) {
    var stdout, stderr bytes.Buffer
    var cfg = ...

    exitCode := Execute([]string{"-y", "--json", "NonExistentList"}, &stdout, &stderr, cfg)

    if exitCode != 1 {
        t.Fatalf("expected exit code 1 for error, got %d", exitCode)
    }

    var response map[string]interface{}
    json.Unmarshal(stdout.Bytes(), &response)

    if response["result"] != "ERROR" {
        t.Errorf("expected ERROR result")
    }

    msg := response["message"].(string)
    if !strings.HasPrefix(msg, "Error 1:") {
        t.Errorf("expected Error 1 (not found), got: %s", msg)
    }
}
```


## Test Categories

| Priority | Category | Location | Purpose |
|----------|----------|----------|---------|
| 1 | **CLI tests** | `cmd/todoat/todoat_test.go` | **Primary**: Basic tests features through CLI |
| 2 | Unit tests | `*_test.go` next to source | Secondary: Complex internal logic |
| 3 | Backend integration | `backend/*/integration_test.go` | Test with real services, focus on using CLI, may also have unit tests |


## When to Write Unit Tests (Not CLI Tests)

Unit tests are appropriate when CLI tests aren't sufficient:

### 1. Backend Protocol Internals
```go
// backend/nextcloud/caldav_test.go
func TestParseVTODO(t *testing.T) {
    ical := `BEGIN:VTODO
SUMMARY:Test task
STATUS:NEEDS-ACTION
END:VTODO`

    task, err := parseVTODO(ical)
    // Test internal parsing logic
}
```

### 2. Complex Algorithms
```go
// backend/sync/conflict_test.go
func TestMergeConflictResolution(t *testing.T) {
    local := &Task{Summary: "Local version", Modified: time.Now()}
    remote := &Task{Summary: "Remote version", Modified: time.Now().Add(-1 * time.Hour)}

    result := mergeConflict(local, remote, MergeStrategyServerWins)
    // Test merge algorithm directly
}
```

### 3. Edge Cases Difficult to Trigger via CLI
```go
// backend/sqlite/query_test.go
func TestSQLInjectionPrevention(t *testing.T) {
    // Test that malicious input is properly escaped
    query := buildQuery("Robert'; DROP TABLE tasks;--")
    // Verify parameterized query
}
```

**Rule of thumb:** If you're testing user-facing behavior, use CLI tests. If you're testing internal implementation details, unit tests are acceptable.

## Test Configuration

Tests should use a test config that enables no-prompt mode:

```yaml
# test_config.yaml
backends:
  test:
    type: sqlite
    enabled: true
    path: ":memory:"

default_backend: test
no_prompt: true
output_format: text
```

Tests should feed configs
