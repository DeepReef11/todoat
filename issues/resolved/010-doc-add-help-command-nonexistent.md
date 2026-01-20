# [010] Documentation References Non-Existent 'todoat add --help' Command

## Type
doc-mismatch

## Category
user-journey

## Severity
low

## Steps to Reproduce
```bash
# User follows docs to get help on add command
todoat add --help
```

## Expected Behavior
According to the documentation (README.md and getting-started.md), the command `todoat add --help` should show help specific to the add action.

From docs/README.md lines 103-104:
```bash
# Command-specific help
todoat add --help
```

From docs/getting-started.md lines 225-227:
```bash
# Command-specific help
todoat add --help
todoat list --help
```

## Actual Behavior
`todoat add --help` shows the main help output (same as `todoat --help`) because `add` is not a standalone subcommand. It's a positional action used in the context `todoat <list> add <task>`.

```bash
$ todoat add --help
# Shows main todoat help, not add-specific help

$ todoat list --help
# This works because 'list' IS a subcommand
```

## Error Output
```
$ todoat add --help
todoat is a command-line task manager supporting multiple backends.

Usage:
  todoat [list] [action] [task] [flags]
  todoat [command]

Available Commands:
  completion   Generate the autocompletion script...
[... main help output ...]
```

## Environment
- OS: Linux
- Runtime version: Go 1.21+

## Possible Cause
The application uses positional arguments where `add`, `update`, `complete`, `delete` are actions within a list context, not standalone subcommands. The docs incorrectly suggest these have their own help pages.

## Documentation Reference
- File: `docs/README.md`
- Section: "Getting Help" (lines 99-106)
- Documented command: `todoat add --help`

- File: `docs/getting-started.md`
- Section: "Getting Help" (lines 222-233)
- Documented command: `todoat add --help`

## Related Files
- `docs/README.md` lines 103-104
- `docs/getting-started.md` lines 225-226

## Recommended Fix
FIX DOCS - Either:

Option A: Remove the `todoat add --help` example and only show subcommands that exist:
```bash
# Command-specific help
todoat list --help
todoat config --help
todoat credentials --help
```

Option B: Add a note explaining that actions like add/update/complete are positional arguments, and document their flags in the main help.

## Dependencies
None
