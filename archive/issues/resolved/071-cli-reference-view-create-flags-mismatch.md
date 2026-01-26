# [071] Example Mismatch: view create --filter-status/--filter-priority flags don't exist

## Type
doc-mismatch

## Category
user-journey

## Severity
high

## Location
- File: `docs/reference/cli.md`
- Line: 258
- Context: CLI reference documentation

## Documented Command
```bash
todoat view create urgent --filter-status "TODO,IN-PROGRESS" --filter-priority "high"
```

## Actual Result
```bash
$ todoat view create --help
Create a new view interactively or from command-line flags.

Without -y flag, opens an interactive builder where you can:
- Select which fields to display
- Configure field widths and formats
- Add filter conditions
- Set sort rules

With -y flag (non-interactive mode), creates a default view or uses provided flags:
  todoat view create myview -y --fields "status,summary,priority" --sort "priority:asc"

Usage:
  todoat view create <name> [flags]

Flags:
      --fields string   Comma-separated list of fields (e.g., "status,summary,priority")
  -h, --help            help for create
      --sort string     Sort rule in format "field:direction" (e.g., "priority:asc")
```

The `--filter-status` and `--filter-priority` flags do not exist. Only `--fields` and `--sort` are available.

## Working Alternative (if known)
```bash
# Create a view with fields and sort only
todoat view create urgent -y --fields "status,summary,priority,due_date" --sort "priority:asc"

# Filters must be added manually by editing the YAML file at ~/.config/todoat/views/urgent.yaml
# or using the interactive view builder (without -y flag)
```

## Recommended Fix
FIX EXAMPLE - Update the CLI reference to show the actual available flags for view create command:
```bash
# Create a view with fields and sort
todoat view create urgent -y --fields "status,summary,priority" --sort "priority:asc"
```

Or alternatively, implement the `--filter-status` and `--filter-priority` flags if desired.

## Impact
Users following this example will see an error when trying to create views with filter flags, causing confusion about how to properly create custom views from the command line.

## Resolution

**Fixed in**: this session
**Fix description**: Updated CLI reference example to use actual available flags (`--fields`, `--sort`, `-y`) instead of non-existent `--filter-status` and `--filter-priority` flags.

### Verification Log
```bash
$ sed -n '257,259p' docs/reference/cli.md
# Create a view with fields and sort
todoat view create urgent -y --fields "status,summary,priority" --sort "priority:asc"
```
**Matches expected behavior**: YES
