# [061] Add --description and --color flags to list create command

## Summary
The `list create` command is documented to support `--description` and `--color` flags, but these flags are not implemented. The `list update` command has these flags, so the implementation pattern already exists.

## Documentation Reference
- Primary: `docs/list-management.md`
- Section: Creating Lists

## Gap Type
missing

## Current Behavior
```bash
$ todoat list create "My List" --description "test" --color "#FF0000"
Error: unknown flag: --description
```

The `list create` command only accepts a positional name argument.

## Expected Behavior (from docs)
```bash
todoat list create "Personal Goals" \
  --description "Goals for 2026" \
  --color "#00CC66"
```

Should create a list with the specified description and color.

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] Test `list create` with `--description` flag
- [ ] Test `list create` with `--color` flag
- [ ] Test `list create` with both flags
- [ ] Test color validation (hex format)

### Functional Requirements
- [ ] `list create "Name" --description "text"` creates list with description
- [ ] `list create "Name" --color "#RRGGBB"` creates list with color
- [ ] Invalid color format returns appropriate error

## Implementation Notes
The `list update` command already has these flags implemented at lines 632-644 of `cmd/todoat/cmd/todoat.go`. The same pattern can be applied to `newListCreateCmd`.

The backend `CreateList` method already supports these properties through the `List` struct.
