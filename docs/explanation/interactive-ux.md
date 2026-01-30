# Interactive UX (Prompt Mode)

## Overview

When `--no-prompt` (`-y`) is **not** set, todoat operates in interactive mode. This mode provides confirmation prompts and task disambiguation when needed. Interactive mode is the default. It can be disabled globally via config (`ui.no_prompt: true`) or per-command via the `-y` flag.

## Prompt Utilities

todoat uses `bufio.Scanner`-based prompts in `internal/utils/inputs.go`. Available functions:

- **`PromptYesNo`**: Prompts the user for a yes/no response, looping until valid input.
- **`PromptSelection`**: Displays a numbered list and prompts the user to select an item by number (enter 0 to cancel).
- **`PromptConfirmation`**: Wrapper around `PromptYesNo` returning `(bool, error)`.
- **`ReadInt`**: Reads an integer from stdin.
- **`ReadString`**: Reads a trimmed string from stdin.

All prompt functions have `WithReader` variants that accept `io.Reader`/`io.Writer` for testing.

An empty stub exists at `internal/cli/prompt/prompt.go` intended for future prompt enhancements.

## Task Disambiguation

### Search Behavior

When an action targets a task by summary, todoat uses a two-phase search:

1. **UUID check**: If the search term looks like a UUID, try matching by task ID first.
2. **Exact match** (case-insensitive): If exactly one task matches, use it directly.
3. **Partial match** (case-insensitive, substring): If exactly one task contains the search term, use it.
4. **Multiple matches**: Return an error listing all matches with their UIDs, priority, due date, and description snippet so the user can re-run with `--uid`.

### Multiple Match Error Format

When multiple tasks match a query, todoat returns an error with details to help distinguish them:

```
multiple tasks match 'mytask'. Use --uid to specify:
  - mytask [P5, due: 2026-02-15] (UID: a1b2c3d4-...)
  - mytask [desc: "Fix the login bug in..."] (UID: e5f6g7h8-...)
```

Fields shown per match (when available):
- Priority (e.g., `P5`)
- Due date (e.g., `due: 2026-02-15`)
- Description snippet (first 30 characters)
- Full UID (always shown)

The user can then re-run with `--uid <full-uid>` to select the exact task.

## No-Prompt Mode (`-y` / `--no-prompt`)

When `--no-prompt` is set (via flag or config), todoat:
- Skips interactive prompts
- Emits structured result codes for machine parsing:
  - `ACTION_COMPLETED` — operation succeeded
  - `INFO_ONLY` — informational response (e.g., list display)
  - `ERROR` — operation failed

This mode is designed for scripting and CI/CD integration.

| Behavior | Interactive (default) | `--no-prompt` (`-y`) |
|----------|----------------------|----------------------|
| Multiple matches | Error with match list | Error with match list + `ACTION_INCOMPLETE` |
| Add without summary | Error: summary required | Error: summary required |
| Decorative output | Standard text | Plain text with result codes |
| Single match | Proceeds silently | Proceeds + `ACTION_COMPLETED` |

## Implementation Notes

### Current Code Locations

| Component | Location | Notes |
|-----------|----------|-------|
| Input utilities | `internal/utils/inputs.go` | `bufio.Scanner`-based prompts |
| Prompt stub | `internal/cli/prompt/prompt.go` | Empty, reserved for future use |
| Task search | `cmd/todoat/cmd/todoat.go` (`findTask`) | Exact → partial → multiple match error |
| Match formatting | `cmd/todoat/cmd/todoat.go` (`formatMultipleMatchesError`) | Shows priority, due, desc, UID |
| No-prompt check | `cfg.NoPrompt` field | Gates result code output |

### Testing

- Prompt functions have `WithReader` variants for injecting `io.Reader`/`io.Writer` in tests
- Tests default to `no_prompt: true` for deterministic behavior

## Related

- [CLI Interface](cli-interface.md) — Command structure, no-prompt mode, result codes
- [CLI Interface: No-Prompt Mode](cli-interface.md#no-prompt-mode) — Non-interactive behavior details
- [Task Management](task-management.md) — Task operations and search
