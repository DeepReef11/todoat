# [001] Fix: errors.md References Nonexistent 'setup' Command - Wrong-Syntax

## Summary
The error reference documentation mentions a `todoat setup <backend>` command that does not exist. The correct command for configuring credentials is `todoat credentials set <backend> <username> --prompt`.

## Documentation Reference
- Primary: `docs/reference/errors.md`
- Section: Authentication Errors (lines 145-171)

## Gap Type
wrong-syntax

## Documented Command/Syntax
```bash
todoat setup todoist
todoat setup <backend>
```

## Actual Result When Running Documented Command
```bash
$ todoat setup todoist
todoat is a command-line task manager supporting multiple backends.
# Shows general help - no 'setup' command exists
```

## Working Alternative
```bash
todoat credentials set todoist token --prompt
todoat credentials set nextcloud myuser --prompt
```

## Recommended Fix
FIX DOCS - The `setup` command never existed. The documentation should reference the correct `credentials set` command.

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] No tests needed - documentation-only fix

### Functional Requirements
- [ ] Update errors.md lines 149, 156, 168 to use `todoat credentials set` instead of `todoat setup`
- [ ] Ensure error suggestions match the actual CLI behavior
