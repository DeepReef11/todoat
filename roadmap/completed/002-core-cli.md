# 002: Core CLI with Cobra

Implement the Cobra CLI skeleton with root command, help system, and the `-y` (no-prompt) flag for scripting support.

## Dependencies

- `001-project-setup.md` - Project structure must exist

## Acceptance Criteria

- [ ] Root command `todoat` implemented with Cobra
- [ ] Injectable IO: Execute function accepts stdout/stderr writers
  ```go
  func Execute(args []string, stdout, stderr io.Writer) int
  ```
- [ ] Help system works: `todoat --help` displays usage information
- [ ] Version flag: `todoat --version` displays version string
- [ ] Global flags implemented:
  - `-y, --no-prompt` - disable interactive prompts
  - `-V, --verbose` - enable verbose/debug output
- [ ] Exit codes:
  - `0` for success
  - `1` for errors
- [ ] Command structure supports subcommands (for later: add, get, update, etc.)
- [ ] Basic argument validation (max 3 positional args as per CLI_INTERFACE.md)
- [ ] Tests pass using injectable IO pattern from TEST_DRIVEN_DEV.md:
  ```go
  var stdout, stderr bytes.Buffer
  exitCode := Execute([]string{"--help"}, &stdout, &stderr)
  // Assert exitCode == 0
  // Assert stdout contains help text
  ```

## Complexity

**Estimate:** M (Medium)

## Implementation Notes

- Reference: `docs/explanation/cli-interface.md` for command structure
- Reference: `docs/explanation/test-driven-dev.md` for testing approach
- Use `github.com/spf13/cobra` for CLI framework
- Keep the root command simple - action handling comes in later roadmap items
- The `-y` flag state should be accessible throughout the application
- Consider a simple config struct to hold global flags
- Write CLI tests first (TDD approach)
