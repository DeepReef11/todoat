# [004] Example Mismatch: docs/README.md feature-demo sections list incomplete

## Type
doc-mismatch

## Category
user-journey

## Severity
low

## Location
- File: `docs/README.md`
- Line: 35
- Context: README

## Documented Command
```bash
# Documented available sections:
Available sections: `version`, `config`, `lists`, `tasks`, `subtasks`, `dates`, `recurring`, `tags`, `priority`, `views`, `filters`, `json`, `sync`, `reminders`, `credentials`, `migration`, `export`, `scripting`, `tui`, `cleanup`
```

## Actual Result
```bash
$ ./docs/feature-demo.sh invalid 2>&1 | head -4
Unknown section: invalid
Available sections: version, help, config, lists, tasks, subtasks, dates, recurring,
                    tags, priority, views, filters, json, sync, reminders,
                    notifications, credentials, migration, export, scripting, tui, cleanup, all
```

## Working Alternative (if known)
```bash
# The actual available sections include 3 more than documented:
# - help
# - notifications
# - all
```

## Recommended Fix
FIX EXAMPLE - Update docs/README.md line 35 to include the missing sections: `help`, `notifications`, and `all`

## Impact
Users reading the docs/README.md will not know about the `help`, `notifications`, and `all` sections that are available in the feature-demo.sh script.

## Resolution

**Fixed in**: this session
**Fix description**: Updated docs/README.md line 35 to include the missing sections: `help`, `notifications`, and `all`

### Verification Log
```bash
$ ./docs/feature-demo.sh invalid 2>&1 | head -4
Unknown section: invalid
Available sections: version, help, config, lists, tasks, subtasks, dates, recurring,
                    tags, priority, views, filters, json, sync, reminders,
                    notifications, credentials, migration, export, scripting, tui, cleanup, all

$ grep "Available sections:" docs/README.md
Available sections: `version`, `help`, `config`, `lists`, `tasks`, `subtasks`, `dates`, `recurring`, `tags`, `priority`, `views`, `filters`, `json`, `sync`, `reminders`, `notifications`, `credentials`, `migration`, `export`, `scripting`, `tui`, `cleanup`, `all`
```
**Matches expected behavior**: YES
