# [081] Fix: --config flag documentation not fully removed

## Summary
The `--config` flag is documented in `docs/explanation/configuration.md` as an existing feature, but this flag was never implemented. A previous fix (roadmap/completed/008-dev-doc-missing-flags.md) removed it from cli-interface.md but configuration.md still references it extensively.

## Documentation Reference
- Primary: `docs/explanation/configuration.md`
- Section: "Config Precedence" (line 55), "Custom Config Path" (lines 619-711)

## Gap Type
wrong-syntax

## Documented Command/Syntax
```bash
# From docs/explanation/configuration.md:
todoat --config /path/to/config.yaml MyList
todoat --config /path/to/config-dir/ MyList
todoat --config . MyList
todoat --config ../other-project/config.yaml MyList
```

## Actual Result When Running Documented Command
```bash
$ todoat --config /path/to/config.yaml list
Error: unknown flag: --config
```

## Working Alternative (if any)
No alternative exists - must set XDG_CONFIG_HOME or use default location.

## Recommended Fix
FIX DOCS - Remove the `--config` flag references from configuration.md to match the fix already applied to cli-interface.md:
1. Remove "Custom Path" from the Config Precedence table (line 55)
2. Remove or mark as "Planned" the entire "Custom Config Path" section (lines 619-711)
3. Update features-overview.md if it references this feature

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] No tests required - documentation fix only

### Functional Requirements
- [ ] Remove or clearly mark as unimplemented all `--config` flag references in configuration.md
- [ ] Update Config Precedence table to remove "Custom Path" entry
- [ ] Remove "Custom Config Path" section or mark as "Planned Feature"
- [ ] Check features-overview.md for any references to custom config path

## Implementation Notes
This is a continuation of the fix from roadmap item 008. The fix was partially applied (cli-interface.md was cleaned) but configuration.md was missed.

References in configuration.md that need updating:
- Line 55: Config Precedence table mentions `--config` flag
- Lines 619-711: "Custom Config Path" section documents the non-existent feature
- Line 1758: Brief mention of "Custom Path" in developer notes
