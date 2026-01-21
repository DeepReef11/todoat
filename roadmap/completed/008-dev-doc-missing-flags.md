# [008] Fix: dev-doc CLI_INTERFACE.md - Missing Flags Documentation

## Summary
The internal docs/explanation/cli-interface.md documents two global flags (`--config` and `--list-backends`) that don't exist in the actual implementation.

## Documentation Reference
- Primary: `docs/explanation/cli-interface.md`
- Section: "Global Flags" (lines ~93-108)

## Gap Type
wrong-syntax

## Documented Command/Syntax
```bash
# From docs/explanation/cli-interface.md:
# Use custom config file
todoat --config /path/to/config.yaml MyList

# Use config in current directory
todoat --config . MyList

# List all backends
todoat --list-backends
```

## Actual Result When Running Documented Command
```bash
$ todoat --config /path/to/config.yaml list
Error: unknown flag: --config

$ todoat --list-backends
Error: unknown flag: --list-backends
```

## Working Alternative (if any)
For `--list-backends`:
```bash
todoat config get backends
```

For `--config`:
No alternative exists - must set XDG_CONFIG_HOME or use default location.

## Recommended Fix
FIX DOCS - Remove the documented flags that don't exist, OR document the planned implementation status. The `--config` flag and `--list-backends` flags appear to be planned features that were never implemented.

Options:
1. Remove documentation for these flags
2. Mark as "Planned" in documentation
3. Implement the flags (would be separate roadmap item)

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] No tests required - documentation fix only

### Functional Requirements
- [ ] Either remove or clearly mark as unimplemented the `--config` flag documentation
- [ ] Either remove or clearly mark as unimplemented the `--list-backends` flag documentation
- [ ] Update the "Global Flags" section to only include flags that exist

## Implementation Notes
The documented flags that don't exist:
1. `--config` / `-c` - Custom config path (documented lines ~93-97)
2. `--list-backends` - List all configured backends (documented lines ~103-108)

The `--detect-backend` flag DOES exist and works correctly.
