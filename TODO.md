# Project TODO

## Questions for Team

### [FEAT-003] Should the inbox/quick-capture workflow (idea #006) use a special list or virtual collection?

**Context**: The unreviewed idea `docs/ideas/unreviewed/006-inbox-workflow.md` proposes a GTD-style inbox for quick task capture without specifying a list. The key architectural question is whether inbox should be a real list (persisted, syncs to backends) or a virtual collection (filtered view of unprocessed tasks across lists).

**Options**:
- [ ] Option A - Special list: Inbox is a real list named "Inbox" that syncs to backends. Simple to implement, works with existing infrastructure, but requires backend support for special list semantics.
- [ ] Option B - Virtual collection: Inbox is a local-only collection of tasks marked "unprocessed". More flexible (tasks can come from any list) but requires new metadata field and doesn't sync the inbox concept to remotes.
- [ ] Option C - Hybrid: Inbox list by default, but allow adding tasks to inbox from other lists via a flag. Combines benefits but adds complexity.

**Impact**: Affects data model, sync behavior, and how inbox integrates with existing list-based workflows.

**Asked**: 2026-01-26
**Status**: unanswered

### [UX-004] What should the default behavior be for bulk destructive operations?

**Context**: The unreviewed idea `docs/ideas/unreviewed/007-bulk-operations-cli.md` proposes bulk operations like `todoat Work delete --all --filter-status DONE`. The safety default for destructive bulk operations needs to be decided.

**Options**:
- [x] Option A - Require confirmation: Always prompt for confirmation on bulk delete/update affecting >1 task. Use `--force` and no-prompt `-y`to skip.
- [ ] Option B - Require dry-run first: Bulk destructive operations fail unless `--force` is passed OR user ran `--dry-run` in the same session within last 5 minutes showing the affected tasks.
- [ ] Option C - Trust the user: No special confirmation needed. Users are expected to use `--dry-run` voluntarily. CLI tools should be fast and scriptable.

**Impact**: Affects user safety vs. scripting convenience trade-off. Important for users who automate task management.

**Asked**: 2026-01-26
**Status**: unanswered

### [ARCH-007] Is the merge conflict strategy field prioritization correct?

**Context**: The recent conflict resolution implementation (commit fdd890e) added a "merge" strategy that combines local and remote task versions. The current implementation uses remote values for summary, description, and status, but keeps local values for priority and categories. This choice is undocumented and users may have different expectations about which fields "win" during merge.

**Options**:
- [ ] Option A - Current behavior: Remote wins for summary/description/status, local wins for priority/categories. Rationale: summary is the "canonical" task identity (should match server), but priority/categories are personal organization preferences.
- [ ] Option B - Field-level timestamps: Track modification time per field (if available) and use the most recent value. More accurate but requires additional metadata tracking.
- [ ] Option C - User-configurable: Allow users to specify per-field merge preferences in config. Most flexible but adds complexity.
- [ ] Option D - Document current behavior: Keep current implementation but document the merge semantics clearly in sync.md.

**Impact**: Affects how merge conflict resolution works. Users who use "merge" strategy may not get expected results if their mental model differs from implementation.

**Asked**: 2026-01-26
**Status**: unanswered
