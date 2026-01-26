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
