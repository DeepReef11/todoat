# Project TODO

## Questions for Team


### [ARCH-007] Is the merge conflict strategy field prioritization correct?

**Context**: The recent conflict resolution implementation (commit fdd890e) added a "merge" strategy that combines local and remote task versions. The current implementation uses remote values for summary, description, and status, but keeps local values for priority and categories. This choice is undocumented and users may have different expectations about which fields "win" during merge.

**Options**:
- [x] Option B - Field-level timestamps: Track modification time per field (if available) and use the most recent value. More accurate but requires additional metadata tracking.

**Impact**: Affects how merge conflict resolution works. Users who use "merge" strategy may not get expected results if their mental model differs from implementation.

**Asked**: 2026-01-26
**Status**: answered

