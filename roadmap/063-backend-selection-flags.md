# [063] Add --backend and --list-backends flags

## Summary
The documentation describes `--backend` flag for selecting a specific backend and `--list-backends` to show configured backends, but neither flag exists. Currently only `--detect-backend` is available.

## Documentation Reference
- Primary: `docs/backends.md`
- Section: Backend Selection, Listing Backends

## Gap Type
missing

## Current Behavior
```bash
$ todoat --backend sqlite MyList
Error: unknown flag: --backend

$ todoat --list-backends
Error: unknown flag: --list-backends
```

## Expected Behavior (from docs)
```bash
# Use specific backend for one command
todoat --backend sqlite MyList

# Show all configured backends
todoat --list-backends
```

## Dependencies
- Requires: none

## Complexity
M

## Acceptance Criteria

### Tests Required
- [ ] Test `--backend` flag selects specified backend
- [ ] Test `--backend` with invalid backend name returns error
- [ ] Test `--list-backends` shows all configured backends
- [ ] Test `--list-backends` shows enabled/disabled status

### Functional Requirements
- [ ] `--backend <name>` overrides default backend for the command
- [ ] `--list-backends` displays table of configured backends with:
  - Backend name
  - Type (sqlite, nextcloud, todoist, git)
  - Enabled status
  - Default indicator
- [ ] Invalid backend name gives helpful error message

## Implementation Notes
The backend resolution logic already exists in `getBackend()`. The `--backend` flag would need to be a persistent flag checked early in the execution flow, similar to how `--detect-backend` works.

For `--list-backends`, the config file already contains all backend configurations. This could be implemented similarly to `config get backends` but with a more user-friendly table format.
