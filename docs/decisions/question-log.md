# Question Log

Archive of team decisions made through the Ralph question process.
For current design decisions, see `docs/explanation/`.

## Index

| ID | Date | Question | Decision |
|----|------|----------|----------|
| FEAT-001 | 2026-01-25 | Should auto_sync_after_operation default to true when sync is enabled? | Yes - default to true when sync.enabled is true |
| ARCH-002 | 2026-01-26 | How should conflict resolution strategies be fully implemented? | Full implementation - implement all four strategies with proper task data manipulation |
| COMPAT-005 | 2026-01-26 | Should sync.auto_sync_after_operation documentation be updated to reflect new default? | Yes - update docs to reflect default is now true when sync enabled |
| ARCH-006 | 2026-01-26 | Should the background pull sync cooldown (30s) be configurable? | Yes - add sync.background_pull_cooldown config option |
| UX-004 | 2026-01-26 | What should the default behavior be for bulk destructive operations? | Require confirmation - always prompt on bulk delete/update affecting >1 task |
| ARCH-007 | 2026-01-26 | Is the merge conflict strategy field prioritization correct? | Field-level timestamps - track modification time per field and use most recent |
| FEAT-008 | 2026-01-26 | Should analytics be enabled by default for new installations? | Enable by default with clear notice |
| FEAT-011 | 2026-01-26 | Should child tasks of DONE parents be auto-hidden in default view? | Configurable - add views.cascade_parent_status_filter option (default true) |
| COMPAT-012 | 2026-01-26 | Should documentation be updated to reflect that built-in views CAN be overridden? | Update documentation - views folder copied on first launch |
| UX-013 | 2026-01-26 | Should views folder creation prompt user when -y flag is NOT provided? | Silent fallback - use built-in views without prompt; use -y flag to create views folder |

---

## Archived Questions

---

### [FEAT-001] Should auto_sync_after_operation default to true when sync is enabled?

**Asked**: 2026-01-25
**Answered**: 2026-01-25
**Documented in**: `docs/explanation/features-overview.md`

**Context**: Issue #3 reports that with `sync.enabled: true`, task operations don't automatically sync. Investigation shows this is because `auto_sync_after_operation` defaults to `false` even when sync is enabled. User must explicitly set `auto_sync_after_operation: true` to get automatic syncing after operations.

**Options**:
- [ ] Option A - Keep current behavior: `auto_sync_after_operation` defaults to `false`. Users who want auto-sync must explicitly enable it. This is more conservative but requires users to discover and set this option.
- [x] Option B - Change default: When `sync.enabled: true`, default `auto_sync_after_operation` to `true`. This matches user expectations that "sync enabled" means "sync happens automatically".
- [ ] Option C - Make auto_sync_after_operation required: When sync is enabled, require the user to explicitly choose (no silent default). This prevents confusion but adds configuration burden.

**Impact**: Affects user experience for all sync-enabled configurations. Option B is a behavior change that existing users might notice.

**Status**: answered

---

### [ARCH-002] How should conflict resolution strategies be fully implemented?

**Asked**: 2026-01-26
**Answered**: 2026-01-26
**Documented in**: `docs/explanation/synchronization.md`

**Context**: The sync documentation describes four conflict resolution strategies (`server_wins`, `local_wins`, `merge`, `keep_both`), and the config validation accepts all four values. However, the current `ResolveConflict()` implementation in `cmd/todoat/cmd/todoat.go:7007-7027` simply marks conflicts as "resolved" in the database without actually applying the strategy logic to modify task data. The `merge` and `keep_both` strategies in particular require complex field-level merging or task duplication that isn't currently implemented.

**Options**:
- [x] Option A - Full implementation: Implement all four strategies with proper task data manipulation (merge fields, duplicate tasks for keep_both, etc.). This matches documentation but requires significant development effort.
- [ ] Option B - Simplify to two strategies: Only support `server_wins` and `local_wins` since these are simpler (just pick one version). Remove `merge` and `keep_both` from config validation and documentation.
- [ ] Option C - Staged approach: Keep validation accepting all four but implement `server_wins`/`local_wins` first. Mark `merge`/`keep_both` as "experimental" in docs and implement later.

**Impact**: Affects sync behavior, documentation accuracy, and user expectations for conflict handling. Users relying on merge/keep_both may be surprised by current behavior.

**Status**: answered

---

### [COMPAT-005] Should sync.auto_sync_after_operation documentation be updated to reflect new default?

**Asked**: 2026-01-26
**Answered**: 2026-01-26
**Documented in**: `docs/explanation/synchronization.md`

**Context**: The recent fix (commit c0aee63) changed `auto_sync_after_operation` to default to `true` when `sync.enabled: true`. However, `docs/how-to/sync.md` still documents it as `"true  # or false (default)"`. The code and documentation are now inconsistent.

**Options**:
- [x] Option A - Update docs: Change documentation to reflect that the default is now `true` when sync is enabled. Clear and accurate.

**Impact**: Documentation accuracy and user understanding of sync behavior.

**Status**: answered

---

### [ARCH-006] Should the background pull sync cooldown (30s) be configurable?

**Asked**: 2026-01-26
**Answered**: 2026-01-26
**Documented in**: `docs/explanation/synchronization.md`

**Context**: Commit 02e2b94 implemented background pull sync on read operations. When `auto_sync_after_operation` is enabled, read operations (GetLists, GetTasks) trigger a background pull-only sync to fetch fresh data from remote backends. There's a hardcoded 30-second cooldown (`backgroundSyncCooldown = 30 * time.Second` in `cmd/todoat/cmd/todoat.go:2912`) to prevent excessive syncing. This behavior is not documented and not configurable.

**Options**:
- [x] Option B - Make configurable: Add `sync.background_pull_cooldown` config option (default: 30s). Power users with fast connections might want lower values; metered connections might want higher.

**Impact**: Affects sync behavior, network usage, and data freshness for users with auto_sync enabled.

**Status**: answered

---

### [UX-004] What should the default behavior be for bulk destructive operations?

**Asked**: 2026-01-26
**Answered**: 2026-01-26
**Documented in**: `docs/explanation/cli-interface.md`

**Context**: The unreviewed idea `docs/ideas/unreviewed/007-bulk-operations-cli.md` proposes bulk operations like `todoat Work delete --all --filter-status DONE`. The safety default for destructive bulk operations needs to be decided.

**Options**:
- [x] Option A - Require confirmation: Always prompt for confirmation on bulk delete/update affecting >1 task. Use `--force` and no-prompt `-y`to skip.
- [ ] Option B - Require dry-run first: Bulk destructive operations fail unless `--force` is passed OR user ran `--dry-run` in the same session within last 5 minutes showing the affected tasks.
- [ ] Option C - Trust the user: No special confirmation needed. Users are expected to use `--dry-run` voluntarily. CLI tools should be fast and scriptable.

**Impact**: Affects user safety vs. scripting convenience trade-off. Important for users who automate task management.

**Status**: answered

---

### [ARCH-007] Is the merge conflict strategy field prioritization correct?

**Asked**: 2026-01-26
**Answered**: 2026-01-26
**Documented in**: `docs/explanation/synchronization.md`

**Context**: The recent conflict resolution implementation (commit fdd890e) added a "merge" strategy that combines local and remote task versions. The current implementation uses remote values for summary, description, and status, but keeps local values for priority and categories. This choice is undocumented and users may have different expectations about which fields "win" during merge.

**Options**:
- [x] Option B - Field-level timestamps: Track modification time per field (if available) and use the most recent value. More accurate but requires additional metadata tracking.

**Impact**: Affects how merge conflict resolution works. Users who use "merge" strategy may not get expected results if their mental model differs from implementation.

**Status**: answered

---

### [FEAT-008] Should analytics be enabled by default for new installations?

**Asked**: 2026-01-26
**Answered**: 2026-01-26
**Documented in**: `docs/explanation/analytics.md`

**Context**: The analytics system (docs/explanation/analytics.md) is documented as "opt-in" and "disabled by default". However, the documentation also states `enabled: true # default` in the config example section. The sample config at internal/config/config.sample.yaml shows analytics as commented out. This inconsistency could confuse users.

**Options**:
- [x] Enable by default with clear notice - Better insights for users, but more intrusive

**Impact**: Affects new user onboarding experience and privacy expectations. Analytics data is local-only and never transmitted.

**Status**: answered

---

### [FEAT-011] Should child tasks of DONE parents be auto-hidden in default view?

**Asked**: 2026-01-26
**Answered**: 2026-01-26
**Documented in**: `docs/explanation/subtasks-hierarchy.md`

**Context**: The docs/explanation/subtasks-hierarchy.md states: "Child with parent status DONE are considered like DONE. For instance, if DONE tasks are filtered out, childs of DONE tasks will also be filtered out." However, commit 3fc5620 recently changed the default view to filter DONE tasks. It's unclear if this cascades to children of DONE parents correctly.

**Options**:
- [x] Configurable - Add config option `views.cascade_parent_status_filter: true/false` default true (filter children of parent)

**Impact**: Affects how users see task hierarchies in the default view. May cause confusion if children of completed parents still show as TODO.

**Status**: answered

---

### [COMPAT-012] Should documentation be updated to reflect that built-in views CAN be overridden?

**Asked**: 2026-01-26
**Answered**: 2026-01-26
**Documented in**: `docs/explanation/views-customization.md`

**Context**: The documentation at `docs/explanation/views-customization.md` (line 1018) states: "Built-in views (default, all) are hard-coded in the application and cannot be overridden. Custom views must use different names." However, the staged changes in `internal/views/loader.go` implement exactly the opposite behavior - the code now checks disk first for user overrides before falling back to built-in views. This change was part of roadmap item 034 (views-folder-setup.md) which explicitly requires "User view overrides built-in" functionality.

**Options**:
- [x] Update documentation - Change the note to explain that users CAN override built-in views by editing `default.yaml` or `all.yaml` in their views folder. The views/ folder with builtin views should be copied if views/ folder doesn't exist at config path on first app launch (like config.yaml).

**Impact**: Documentation accuracy and user expectations. If docs say "cannot be overridden" but code allows it, users may not discover this useful customization option.

**Status**: answered

---

### [UX-013] Should views folder creation prompt user when -y flag is NOT provided?

**Asked**: 2026-01-26
**Answered**: 2026-01-26
**Documented in**: `docs/explanation/views-customization.md`

**Context**: Roadmap item 034 specifies: "When views/ folder doesn't exist and -y not used, should prompt: 'Views folder not found. Create with default views? [Y/n]'". However, the current implementation in `cmd/todoat/cmd/todoat.go:3387` only auto-creates the views folder when `cfg.NoPrompt` (i.e., `-y` flag) is true. Without `-y`, the folder is never created and no prompt is shown - the user silently gets built-in views.

**Options**:
- [x] Silent fallback - Use built-in views without prompt when views folder doesn't exist; users can run any command with `-y` flag to create the views folder

**Impact**: Affects first-run user experience. Silent fallback is simpler. Users who want to customize views can use the `-y` flag to initialize the views folder.

**Status**: answered
