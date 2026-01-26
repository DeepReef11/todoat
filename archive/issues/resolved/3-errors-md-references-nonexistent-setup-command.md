# [003] Example Mismatch: errors.md references nonexistent setup command

## Type
doc-mismatch

## Category
user-journey

## Severity
high

## Location
- File: `docs/reference/errors.md`
- Line: 149, 157, 168
- Context: Error Reference documentation

## Documented Command
```bash
todoat setup <backend>
todoat setup todoist
```

## Actual Result
```bash
$ ./todoat setup --help
todoat is a command-line task manager supporting multiple backends.

Usage:
  todoat [list] [action] [task] [flags]
  todoat [command]

Available Commands:
  completion   Generate the autocompletion script for the specified shell
  config       View and manage configuration
  credentials  Manage backend credentials
  ...
```

The `setup` command does not exist. The CLI only shows `credentials` for managing credentials.

## Working Alternative (if known)
```bash
todoat credentials set <backend> <username> --prompt
todoat credentials set todoist token --prompt
todoat credentials set nextcloud myuser --prompt
```

## Recommended Fix
FIX EXAMPLE - Update docs/reference/errors.md to replace all `todoat setup <backend>` references with `todoat credentials set <backend> <username> --prompt`

Specific changes needed:
1. Line 149: Change `Run 'todoat setup <backend>' to configure credentials` to `Run 'todoat credentials set <backend> <username> --prompt' to configure credentials`
2. Line 157: Change `Run 'todoat setup todoist'` to `Run 'todoat credentials set todoist token --prompt'`
3. Line 168: Change `Re-run setup: 'todoat setup <backend>'` to `Re-run setup: 'todoat credentials set <backend> <username> --prompt'`

## Impact
Users following this error suggestion will see: `Error: unknown action: setup` or the help message, making it unclear how to actually configure credentials.

## Resolution

**Fixed in**: this session
**Fix description**: Updated docs/reference/errors.md to replace all `todoat setup <backend>` references with `todoat credentials set <backend> <username> --prompt`

### Verification Log
```bash
$ grep -n "todoat credentials set" docs/reference/errors.md
149:**Suggestion**: Run `todoat credentials set <backend> <username> --prompt` to configure credentials.
156:Suggestion: Run 'todoat credentials set todoist token --prompt' to configure credentials
168:1. Re-run setup: `todoat credentials set <backend> <username> --prompt`

$ ./todoat credentials set --help
Store credentials securely in the system keyring (macOS Keychain, Windows Credential Manager, or Linux Secret Service).

Usage:
  todoat credentials set [backend] [username] [flags]

Flags:
  -h, --help     help for set
      --prompt   Prompt for password input (required for security)
```
**Matches expected behavior**: YES
