# [062] Implement view create command with interactive builder

## Summary
The `view create` command is documented with an interactive terminal interface for creating custom views, but only `view list` is implemented. Users must manually create YAML files.

## Documentation Reference
- Primary: `docs/views.md`
- Section: Creating a Custom View - Method 1: Interactive Builder

## Gap Type
missing

## Current Behavior
```bash
$ todoat view create myview
Error: unknown command "create" for "todoat view"
```

Only `todoat view list` is available.

## Expected Behavior (from docs)
```bash
todoat view create myview
```

Opens a terminal interface where users can:
- Select fields to display
- Configure field widths and formats
- Add filters
- Set sort rules

## Dependencies
- Requires: none

## Complexity
L

## Acceptance Criteria

### Tests Required
- [ ] Test view create command generates valid YAML file
- [ ] Test created view can be used with `-v` flag
- [ ] Test overwriting existing view (with confirmation)

### Functional Requirements
- [ ] `view create <name>` starts interactive builder
- [ ] User can select which fields to include
- [ ] User can configure field widths
- [ ] User can add filter conditions
- [ ] User can set sort order
- [ ] Resulting YAML is saved to views directory
- [ ] Non-interactive mode (`-y`) writes default view config

## Implementation Notes
The `internal/views/builder.go` file already exists with builder-related code. The TUI framework (charmbracelet/bubbletea) is already in use for the main TUI. Consider using a simpler approach like checkbox selection with charm libraries.

Alternative: Start with a simpler non-interactive version that creates a basic view YAML from command-line flags:
```bash
todoat view create myview --fields "status,summary,due_date" --sort "priority:asc"
```
