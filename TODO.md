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

### [ARCH-008] Where should task archives be stored?

**Context**: The unreviewed idea `docs/ideas/unreviewed/008-task-archival.md` proposes archiving completed tasks to preserve history while keeping active lists clean. The storage location has implications for performance, sync behavior, and backup strategy.

**Options**:
- [ ] Option A - Same SQLite database: Archive table in the main tasks.db. Simplest implementation, shared backups, but increases database size over time.
- [ ] Option B - Separate archive database: Dedicated archive.db file. Keeps main database lean, easier to exclude from sync, but adds complexity for queries that span active and archived tasks.
- [ ] Option C - Per-list archive tables: Archive table per list in main database. Allows list-specific retention policies, but more complex schema.

**Impact**: Affects database performance, sync architecture, and how archives interact with existing backup/restore workflows.

**Asked**: 2026-01-26
**Status**: unanswered

### [UX-009] How should multi-backend aggregate views display task source?

**Context**: The unreviewed idea `docs/ideas/unreviewed/009-multi-backend-views.md` proposes views that show tasks from multiple backends together. Users need to know which backend a task comes from, but the display method affects readability and usability.

**Options**:
- [ ] Option A - Prefix notation: `[sqlite] Buy groceries`. Clear but adds visual noise to every line.
- [ ] Option B - Dedicated column: Add "Backend" column to table output. Clean but takes horizontal space.
- [ ] Option C - Color coding: Each backend gets a distinct color. Visual but requires color support and may be inaccessible.
- [ ] Option D - On-demand only: Show backend via `--show-backend` flag, hidden by default. Cleanest output but less discoverable.

**Impact**: Affects terminal output formatting and accessibility. Important for users who rely on aggregate views daily.

**Asked**: 2026-01-26
**Status**: unanswered

### [ARCH-010] How should task dependencies handle circular references?

**Context**: The unreviewed idea `docs/ideas/unreviewed/005-task-dependencies.md` proposes task dependencies ("blocked by" relationships). Circular dependencies (A blocks B, B blocks C, C blocks A) could occur through user error or sync from misconfigured backends.

**Options**:
- [ ] Option A - Prevent at creation: Validate dependency graph when adding dependencies, reject if cycle would be created. Safest but requires graph traversal on every add.
- [ ] Option B - Detect and warn: Allow circular dependencies but mark affected tasks with a warning. Show "circular dependency detected" in task view.
- [ ] Option C - Break on query: When calculating "actionable" tasks, detect and break cycles arbitrarily. Document that cycles may have unpredictable behavior.

**Impact**: Affects dependency validation performance and user experience when dependencies are misconfigured.

**Asked**: 2026-01-26
**Status**: unanswered
