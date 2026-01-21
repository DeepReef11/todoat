# [007] Fix: dev-doc CLI_INTERFACE.md - Status Abbreviation Inconsistency

## Summary
The internal docs/explanation/cli-interface.md incorrectly documents `PROCESSING/P` as a valid status abbreviation, but the actual implementation uses `IN-PROGRESS/I`.

## Documentation Reference
- Primary: `docs/explanation/cli-interface.md`
- Section: "Action Flags" (line ~206) and throughout the document

## Gap Type
wrong-syntax

## Documented Command/Syntax
```bash
# From docs/explanation/cli-interface.md:
todoat MyList -s T,P  # Using abbreviations (T=TODO, P=PROCESSING)
todoat MyList -s PROCESSING
```

## Actual Result When Running Documented Command
```bash
$ todoat TestList -s PROCESSING
Error: invalid status "PROCESSING": valid values are TODO, IN-PROGRESS, DONE, CANCELLED

$ todoat TestList -s P
Error: invalid status "P": valid values are TODO, IN-PROGRESS, DONE, CANCELLED
```

## Working Alternative (if any)
```bash
todoat MyList -s IN-PROGRESS
todoat MyList -s I
```

## Recommended Fix
FIX DOCS - The documentation should be updated to reflect the actual implementation which uses `IN-PROGRESS/I` (not `PROCESSING/P`). The actual status name `IN-PROGRESS` is more standard and matches what users expect.

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] No tests required - documentation fix only

### Functional Requirements
- [ ] Update all occurrences of `PROCESSING/P` to `IN-PROGRESS/I` in docs/explanation/cli-interface.md

## Implementation Notes
Locations in docs/explanation/cli-interface.md that need updating:
- Line ~206: Status abbreviation table
- Line ~279: Example commands with `-s T,P`
- Line ~628: Argument validation example
- Any other occurrences of "PROCESSING"

Note: The user-facing docs (docs/task-management.md) already correctly document `IN-PROGRESS/I`.
