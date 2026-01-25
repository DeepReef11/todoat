# [080] Fix: sync.conflict_resolution Documented Values Don't Match Implementation

## Summary
The documentation describes `sync.conflict_resolution` accepting values `server_wins`, `local_wins`, `merge`, `keep_both` but the implementation only accepts `local`, `remote`, `manual`.

## Documentation Reference
- Primary: `docs/reference/configuration.md`
- Secondary: `docs/how-to/sync.md`
- Section: Sync Configuration Options, Conflict Resolution

## Gap Type
wrong-syntax

## Documented Command/Syntax
```bash
todoat config set sync.conflict_resolution server_wins
# Also documented: local_wins, merge, keep_both
```

## Actual Result When Running Documented Command
```bash
$ todoat config set sync.conflict_resolution server_wins
Error: invalid value for sync.conflict_resolution: server_wins (valid: local, remote, manual)
```

## Working Alternative (if any)
```bash
todoat config set sync.conflict_resolution remote  # Instead of server_wins
todoat config set sync.conflict_resolution local   # Instead of local_wins
```

## Recommended Fix
FIX DOCS AND CODE - The documentation describes a more user-friendly naming scheme (`server_wins`, `local_wins`, `merge`, `keep_both`) which should be implemented. The current implementation values (`local`, `remote`, `manual`) are less descriptive.

Option 1 (Preferred): Update code to accept documented values
Option 2: Update docs to match code values (less user-friendly)

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] Test that documented values work (if fixing code)
- [ ] Test that error message matches documentation

### Functional Requirements
- [ ] `todoat config set sync.conflict_resolution server_wins` works (or docs updated)
- [ ] `todoat config set sync.conflict_resolution local_wins` works (or docs updated)
- [ ] `todoat config set sync.conflict_resolution merge` works (or docs updated)
- [ ] `todoat config set sync.conflict_resolution keep_both` works (or docs updated)
- [ ] Help text and error messages match valid values

## Implementation Notes
The documented values are more descriptive and user-friendly:
- `server_wins` (clearer than `remote`)
- `local_wins` (clearer than `local`)
- `merge` (same in both)
- `keep_both` (clearer than `manual`)

In `cmd/todoat/cmd/todoat.go` at line 9376-9379:

```go
// Current:
validValues := []string{"local", "remote", "manual"}

// Should be (to match docs):
validValues := []string{"server_wins", "local_wins", "merge", "keep_both"}
```

Also need to update any other places in the codebase that use these values.
