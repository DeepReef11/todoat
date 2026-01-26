# [001] Fix: Documentation Wrong Command Syntax

## Summary
Several documentation files show incorrect command syntax that omits the required list name parameter. The documented commands `todoat add "task"` and `todoat complete "task"` are wrong - the correct syntax is `todoat <listname> add "task"` and `todoat <listname> complete "task"`.

## Documentation Reference
- Primary: `docs/reference/errors.md`
- Secondary: `docs/explanation/list-management.md`, `docs/explanation/analytics.md`
- Section: Various error examples

## Gap Type
wrong-syntax

## Documented Command/Syntax
```bash
# From docs/reference/errors.md:23
$ todoat complete "nonexistent task"

# From docs/reference/errors.md:65
$ todoat add "task" -p 15

# From docs/reference/errors.md:81
$ todoat add "task" --due-date "next week"

# From docs/reference/errors.md:97
$ todoat update "task" -s "pending"

# From docs/explanation/list-management.md:273
$ todoat add "Buy milk"

# From docs/explanation/analytics.md:439
todoat add "Test task" --priority 1
todoat complete 1
```

## Actual Result When Running Documented Command
```bash
$ todoat complete "nonexistent task"
Error: unknown action: nonexistent task

$ todoat add "task" -p 15
Error: unknown action: task
```

The CLI parses this as `todoat <list="complete"> <action="nonexistent"> <task="task">` because the syntax is `todoat [list] [action] [task]`.

## Working Alternative (if any)
```bash
# Correct syntax includes list name:
$ todoat MyList complete "nonexistent task"
Error: no task found matching 'nonexistent task'

$ todoat MyList add "task" -p 15
Error: priority must be between 0 and 9, got: 15
```

## Recommended Fix
FIX DOCS - Update examples to use correct command syntax with list name parameter.

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] No tests required (documentation fix only)

### Functional Requirements
- [ ] All command examples in docs/reference/errors.md include list name
- [ ] Example in docs/explanation/list-management.md:273 uses correct syntax
- [ ] Examples in docs/explanation/analytics.md:439-441 use correct syntax
- [ ] Error message text matches actual error messages

## Implementation Notes
The following files need updates:

1. **docs/reference/errors.md**:
   - Line 23: `todoat complete` → `todoat MyList complete`
   - Line 65: `todoat add` → `todoat MyList add`
   - Line 81: `todoat add` → `todoat MyList add`
   - Line 97: `todoat update` → `todoat MyList update`

2. **docs/explanation/list-management.md**:
   - Line 273: The example shows `todoat add "Buy milk"` which demonstrates interactive list selection, but the current CLI gives "unknown action" error instead. Either fix the docs to use correct syntax, or note that this is aspirational behavior.

3. **docs/explanation/analytics.md**:
   - Line 439: `todoat add` → `todoat MyList add`
   - Line 441: `todoat complete 1` → `todoat MyList complete "task-summary"` (numeric IDs don't work for task selection by default)
