# [030] Example Mismatch: credentials commands missing username argument

## Type
doc-mismatch

## Category
user-journey

## Severity
high

## Location
- File: `docs/reference/cli.md`
- Lines: 430, 433, 479
- Context: CLI reference documentation

## Documented Command
```bash
# In CLI reference (line 430)
| `get <backend>` | Retrieve credentials and show source |

# In CLI reference (line 433)
| `delete <backend>` | Remove credentials from keyring |

# In CLI reference example (line 479)
todoat credentials delete nextcloud
```

## Actual Result
```bash
$ ./todoat credentials get nextcloud
Error: accepts 2 arg(s), received 1

$ ./todoat credentials delete nextcloud
Error: accepts 2 arg(s), received 1

$ ./todoat credentials get --help
Usage:
  todoat credentials get [backend] [username] [flags]

$ ./todoat credentials delete --help
Usage:
  todoat credentials delete [backend] [username] [flags]
```

## Working Alternative (if known)
```bash
# Both commands require backend AND username
todoat credentials get nextcloud myuser
todoat credentials delete nextcloud myuser
```

## Recommended Fix
FIX DOCS - Update docs/reference/cli.md

Change line 430 from:
```markdown
| `get <backend>` | Retrieve credentials and show source |
```
To:
```markdown
| `get <backend> <username>` | Retrieve credentials and show source |
```

Change line 433 from:
```markdown
| `delete <backend>` | Remove credentials from keyring |
```
To:
```markdown
| `delete <backend> <username>` | Remove credentials from keyring |
```

Change line 479 from:
```bash
todoat credentials delete nextcloud
```
To:
```bash
todoat credentials delete nextcloud myuser
```

## Impact
Users following this documentation will see "Error: accepts 2 arg(s), received 1" when trying to use the credentials get/delete commands with only the backend argument.

## Resolution

**Fixed in**: this session
**Fix description**: Updated docs/reference/cli.md to include `<username>` argument in `get` and `delete` subcommands table, and fixed the example to include `myuser` argument.

### Verification Log
```bash
$ ./todoat credentials get --help
Retrieve credentials from the priority chain (keyring > environment > config URL) and display the source.

Usage:
  todoat credentials get [backend] [username] [flags]

$ ./todoat credentials delete --help
Remove stored credentials from the system keyring. Environment variables and config URL credentials are not affected.

Usage:
  todoat credentials delete [backend] [username] [flags]
```

**Documentation now matches CLI**: YES
- Line 476: `get <backend> <username>` matches `[backend] [username]`
- Line 479: `delete <backend> <username>` matches `[backend] [username]`
- Line 525: `todoat credentials delete nextcloud myuser` is valid syntax
