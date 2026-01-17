# 001: Project Setup

Initialize the Go module, establish the directory structure per TEST_DRIVEN_DEV.md, and create a basic Makefile for development workflows.

## Dependencies

None - this is the first roadmap item.

## Acceptance Criteria

- [ ] Go module initialized with `go mod init` (module name: `todoat` or appropriate path)
- [ ] Directory structure created per TEST_DRIVEN_DEV.md:
  ```
  todoat/
  ├── cmd/
  │   └── todoat/
  │       ├── main.go           # Thin entry point
  │       └── todoat.go         # Cobra setup with injectable IO
  ├── backend/
  │   ├── interface.go          # Backend interface definition
  │   └── sqlite/
  │       └── sqlite.go         # SQLite implementation (stub)
  ├── internal/
  │   ├── app/                  # Core application logic
  │   ├── config/               # Configuration handling
  │   └── cli/                  # CLI display and prompts
  │       └── prompt/           # Prompt manager (no-prompt mode)
  └── Makefile
  ```
- [ ] `main.go` is a thin entry point that calls `cmd.Execute()`
- [ ] Basic Makefile with targets:
  - `build` - compile the binary
  - `test` - run all tests
  - `clean` - remove build artifacts
  - `run` - build and run with arguments
- [ ] `go build ./...` succeeds without errors
- [ ] `go test ./...` runs (even if no tests exist yet)

## Complexity

**Estimate:** S (Small)

## Implementation Notes

- Follow the project structure from `dev-doc/TEST_DRIVEN_DEV.md`
- Keep `main.go` minimal - just parse args and call Execute
- Use standard Go project conventions
- The Makefile should be simple and shell-portable
- Create placeholder files with minimal content to establish structure
- Backend interface should define the core TaskManager methods needed for MVP
