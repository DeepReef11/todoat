# [002] Fix: cli.md Documents Unsupported Relative Date Keywords - Wrong-Behavior

## Summary
The CLI reference documentation claims that `next week` and `next month` are valid relative date keywords, but they are not implemented and produce an error when used.

## Documentation Reference
- Primary: `docs/reference/cli.md`
- Section: Date Syntax (lines 114-130)

## Gap Type
wrong-behavior

## Documented Command/Syntax
```bash
todoat MyList add "Task" --due-date "next week"
todoat MyList add "Task" --due-date "next month"
```

Per the documentation table:
| Keyword | Meaning |
|---------|---------|
| `next week` | 7 days from today |
| `next month` | 1 month from today |

## Actual Result When Running Documented Command
```bash
$ todoat MyList add "task" --due-date "next week"
Error: invalid due-date: invalid date: next week

Suggestion: Use date format YYYY-MM-DD (e.g., 2026-01-15)

$ todoat MyList add "task" --due-date "next month"
Error: invalid due-date: invalid date: next month

Suggestion: Use date format YYYY-MM-DD (e.g., 2026-01-15)
```

## Working Alternative
```bash
# These work correctly:
todoat MyList add "Task" --due-date "+7d"    # 7 days from now
todoat MyList add "Task" --due-date "+1m"    # 1 month from now
todoat MyList add "Task" --due-date "today"
todoat MyList add "Task" --due-date "tomorrow"
todoat MyList add "Task" --due-date "yesterday"
```

## Recommended Fix
FIX DOCS - Remove `next week` and `next month` from the documentation table since they are not implemented. The `+7d` and `+1m` syntax achieves the same result and is already documented.

Alternatively: FIX CODE - Implement support for `next week` and `next month` keywords, but this adds complexity for minimal benefit since the `+7d` and `+1m` syntax already works.

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] No tests needed if fixing docs only
- [ ] If implementing keywords: add tests for `next week` and `next month` parsing

### Functional Requirements
- [ ] Remove `next week` and `next month` from cli.md date syntax table (lines 120-121)
- [ ] Or implement these keywords in the date parser
