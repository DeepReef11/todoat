# [012] Example Mismatch: dev-doc references non-existent commands

## Type
doc-mismatch

## Category
user-journey

## Severity
medium

## Location
- File: `dev-doc/README.md`
  - Lines: 47, 49, 52
  - Context: Quick Feature Reference table
- File: `dev-doc/CONFIGURATION.md`
  - Lines: 262, 934
  - Context: Backend configuration examples

## Documented Commands (line 47)
```
| **Lists** | Create, rename, delete, restore lists | `list create`, `list rename`, `list delete`, `list trash` |
```

## Actual Result
```bash
$ todoat list rename --help
# Shows help for `todoat list` - no rename subcommand exists
# Available subcommands: create, delete, export, import, info, stats, trash, update, vacuum
```

## Working Alternative
Use `list update --name` to rename lists:
```bash
todoat list update "Old Name" --name "New Name"
```

---

## Documented Commands (line 49)
```
| **Backends** | Multi-backend support with auto-detection | `--backend`, `--list-backends`, `--detect-backend` |
```

## Actual Result
```bash
$ todoat --list-backends
Error: unknown flag: --list-backends
```

## Working Alternative
Only `--detect-backend` exists:
```bash
todoat --detect-backend
```

---

## Documented Commands (line 52)
```
| **Views** | Customizable display formats | `view list`, `view create`, `view show`, `-v <view-name>` |
```

## Actual Result
```bash
$ todoat view show --help
# Shows help for `todoat view` - no show subcommand exists
# Available subcommands: create, list
```

## Working Alternative
Use `view list` to see available views:
```bash
todoat view list
```

---

---

## dev-doc/CONFIGURATION.md References (lines 262, 934)

Lines 262 and 934 reference `--list-backends` which doesn't exist:
```
Line 262: - Each backend appears in `--list-backends` output
Line 934: todoat --list-backends
```

## Working Alternative
Use `--detect-backend` to see available backends:
```bash
todoat --detect-backend
```

---

## Recommended Fix
FIX DOCS - Update the internal documentation to match actual CLI:

**dev-doc/README.md:**
- Line 47: Change `list rename` to `list update --name`
- Line 49: Remove `--list-backends` (non-existent flag)
- Line 52: Remove `view show` (non-existent command)

**dev-doc/CONFIGURATION.md:**
- Line 262: Change `--list-backends` to `--detect-backend` or `config get backends`
- Line 934: Change `todoat --list-backends` to `todoat --detect-backend` or `todoat config get backends`

## Impact
Developers reading the internal documentation will try commands that don't exist, causing confusion and wasted time troubleshooting.

## Resolution

**Fixed in**: this session
**Fix description**: Updated documentation to match actual CLI commands

### Changes Made
- `dev-doc/README.md` line 47: Changed `list rename` to `list update`
- `dev-doc/README.md` line 49: Removed `--list-backends` (non-existent flag)
- `dev-doc/README.md` line 52: Removed `view show` (non-existent command)
- `dev-doc/CONFIGURATION.md` line 262: Changed `--list-backends` to `--detect-backend`
- `dev-doc/CONFIGURATION.md` line 934: Changed `todoat --list-backends` to `todoat --detect-backend`

### Verification Log
```bash
$ ./todoat list --help
Available Commands:
  create      Create a new list
  delete      Delete a list (move to trash)
  export      Export a list to a file
  import      Import a list from a file
  info        Show list details
  stats       Show database statistics
  trash       View and manage deleted lists
  update      Update a list's properties
  vacuum      Compact the database
# Confirms: list update exists, list rename does not

$ ./todoat --help | grep -E "backend|detect"
  -b, --backend string          Backend to use (sqlite, todoist)
      --detect-backend          Show auto-detected backends and exit
# Confirms: --detect-backend exists, --list-backends does not

$ ./todoat view --help
Available Commands:
  create      Create a new view
  list        List available views
# Confirms: view list and view create exist, view show does not
```
**Matches expected behavior**: YES
