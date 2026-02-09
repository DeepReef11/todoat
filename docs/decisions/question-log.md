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
| FEAT-011 | 2026-01-26 | Should child tasks of DONE parents be auto-hidden in default view? | NOT IMPLEMENTED - Feature was documented but never implemented. Decision reversed: children are filtered individually based on their own status |
| COMPAT-012 | 2026-01-26 | Should documentation be updated to reflect that built-in views CAN be overridden? | Update documentation - views folder copied on first launch |
| UX-013 | 2026-01-26 | Should views folder creation prompt user when -y flag is NOT provided? | Silent fallback - use built-in views without prompt; use -y flag to create views folder |
| ARCH-001 | 2026-01-31 | Should the daemon Unix socket have restricted file permissions? | Restrict to owner only (0600) |
| ARCH-002 | 2026-01-31 | Should conflict resolution propagate to remote on next sync? | Queue a remote update automatically after resolution |
| FEAT-003 | 2026-01-31 | Should background logging be a runtime config option? | Make it a config option (`logging.background_enabled`) |
| FEAT-004 | 2026-01-31 | What should happen when a notification backend tool is missing? | Warn once on first failure + Auto-detect available tools at startup |
| UX-006 | 2026-01-31 | Should auto-sync wait for completion or return immediately? | Fire-and-forget with background indicator |
| ARCH-007 | 2026-01-31 | Should the daemon HeartbeatInterval field be removed or implemented? | Implement heartbeat mechanism |
| UX-008 | 2026-01-31 | How should reminder dismissal interact with multiple intervals? | Per-interval dismissal - each interval tracked independently |
| UX-010 | 2026-01-31 | Should cache TTL be user-configurable via config.yaml? | Add `cache_ttl` config option to Config struct |
| UX-012 | 2026-01-31 | Should notification configuration be user-configurable via config.yaml? | Add `notification` config to Config struct |
| FEAT-011 | 2026-01-31 | Is background-deamon.md critically outdated and needs rewrite? | Rewrite to match current forked process + IPC implementation |
| FEAT-005 | 2026-02-06 | Should cache TTL be user-configurable? | Add `cache_ttl` config option - Full user control |
| UX-009 | 2026-02-08 | docs/explanation/interactive-ux.md needs rewrite to match implemented interactive prompt | Minimal update - fix empty stub references and add config option mention |
| ARCH-025 | 2026-02-08 | Update docs/explanation/background-deamon.md to reflect implemented heartbeat mechanism | Update explanation doc - rewrite Hung Daemon Detection section |
| ARCH-020 | 2026-02-08 | Should config validation accept all 7 supported backends as default_backend? | Expand validation to all 7 backends |
| FEAT-022 | 2026-02-08 | Should reminder.enabled default to true in DefaultConfig()? | Add Reminder: ReminderConfig{Enabled: true} to DefaultConfig() |
| FEAT-027 | 2026-02-08 | Update docs/explanation/background-deamon.md: Stuck Task Detection is now implemented | Update explanation doc |
| FEAT-028 | 2026-02-08 | Update docs/explanation/background-deamon.md: Per-Task Timeout is now implemented | Update explanation doc |

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

**Status**: NOT IMPLEMENTED - The config option was documented but never implemented in code. Decision reversed in issue #26: children are now filtered individually based on their own status, not their parent's status. Documentation updated to reflect actual behavior.

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

---

### [ARCH-001] Should the daemon Unix socket have restricted file permissions?

**Asked**: 2026-01-29
**Answered**: 2026-01-31
**Documented in**: `docs/explanation/architecture.md`

**Context**: The daemon creates a Unix domain socket for IPC (`internal/daemon/daemon.go`). Currently the socket inherits the process umask, meaning other users on the system may be able to connect to the daemon and issue commands (status queries, sync triggers). This was introduced in commit `a0df401` (background sync daemon).

**Options**:
- [x] Restrict to owner only (0600) - Prevents other users from interacting with daemon

**Impact**: Security posture for multi-user systems. Single-user desktops are unaffected.

**Status**: answered

---

### [ARCH-002] Should conflict resolution propagate to remote on next sync?

**Asked**: 2026-01-29
**Answered**: 2026-01-31
**Documented in**: `docs/explanation/synchronization.md`

**Context**: Sync conflict resolution is explicitly a local-only operation (`backend/sqlite/cli_test.go`). When a conflict is resolved locally, the remote backend is not updated. This means the next sync cycle may re-detect the same conflict if the remote still holds the conflicting version.

**Options**:
- [ ] Push resolution to remote on next sync - Prevents conflict reappearance, but adds complexity
- [ ] Keep local-only resolution - Simpler, but users may see resolved conflicts reappear
- [x] Queue a remote update automatically after resolution - Middle ground, uses existing sync queue

**Impact**: Core sync behavior. Affects user trust in conflict resolution workflow.

**Status**: answered

---

### [FEAT-003] Should background logging be a runtime config option instead of compile-time constant?

**Asked**: 2026-01-29
**Answered**: 2026-01-31
**Documented in**: `docs/explanation/logging.md`

**Context**: `internal/utils/logger.go` defines `const ENABLE_BACKGROUND_LOGGING = true` as a compile-time constant. This means background logging (writing to `/tmp/todoat-*-{PID}.log`) is always on and cannot be toggled without recompiling. Users have no control over whether these log files are created.

**Options**:
- [x] Make it a config option (`logging.background_enabled: true/false`) - User control, matches other config patterns
- [ ] Keep as compile-time constant but default to false - Only developers enable it for debugging
- [ ] Remove entirely and rely on verbose mode (`--verbose`) - Simplify, one logging mechanism

**Impact**: Affects disk usage in `/tmp`, user privacy (logs may contain task content), debugging workflow.

**Status**: answered

---

### [FEAT-004] What should happen when a notification backend tool is missing?

**Asked**: 2026-01-29
**Answered**: 2026-01-31
**Documented in**: `docs/explanation/notification-manager.md`

**Context**: OS notifications use platform-specific tools: `notify-send` (Linux), `osascript` (macOS), `powershell` (Windows). The Linux fallback is `wall` which broadcasts to all terminal sessions. The behavior when the primary tool is missing is not explicitly handled—it may silently fail or produce a confusing error.

**Options**:
- [ ] Silent skip with debug log - Don't bother user, log for debugging
- [x] Warn once on first failure - Alert user their notification tool is missing, then suppress
- [x] Auto-detect available tools at startup - Validate config and warn proactively
- [ ] Fail loudly - Return error so caller can decide

**Impact**: User experience for notifications and reminders. Affects all platforms.

**Status**: answered

---

### [UX-006] Should auto-sync wait for completion or return immediately?

**Asked**: 2026-01-29
**Answered**: 2026-01-31
**Documented in**: `docs/explanation/synchronization.md`

**Context**: When `auto_sync_after_operation: true`, CLI operations (create, update, delete) trigger a sync. It's unclear whether the CLI waits for sync to complete before returning to the user, or fires-and-forgets. Waiting gives users confidence their change is synced; returning immediately feels faster but leaves sync status ambiguous.

**Options**:
- [ ] Wait for sync completion - User sees success/failure, slower UX
- [x] Fire-and-forget with background indicator - Return immediately, show sync status on next command or use os notification
- [ ] Configurable (`sync.wait_for_completion: true/false`) - Let users choose their preference

**Impact**: CLI responsiveness vs sync reliability. Core user experience for synced workflows.

**Status**: answered

---

### [ARCH-007] Should the daemon HeartbeatInterval field be removed or implemented?

**Asked**: 2026-01-29
**Answered**: 2026-01-31
**Documented in**: `docs/explanation/architecture.md`

**Context**: `DaemonConfig` in `internal/config/config.go` includes a `HeartbeatInterval` field, and tests reference heartbeat behavior (`daemon_test.go:842`), but no heartbeat recording or checking code exists. This creates a documentation-code gap where the config field is parseable but non-functional.

**Options**:
- [x] Implement heartbeat mechanism - Enables hung daemon detection, matches documented architecture
- [ ] Remove the field - Clean up dead code, reduce config surface area
- [ ] Keep field but mark as reserved/future - Document that it's not yet functional

**Impact**: Config clarity, daemon reliability monitoring. Dead config fields may confuse users.

**Status**: answered

---

### [UX-008] How should reminder dismissal interact with multiple intervals?

**Asked**: 2026-01-29
**Answered**: 2026-01-31
**Documented in**: `docs/explanation/user-experience.md`

**Context**: Reminders support multiple intervals (e.g., `[1d, 1h, "at due time"]`). When a user dismisses a reminder, the docs say "you'll be reminded again at the next interval." It's unclear whether dismissal is per-interval (dismissing the 1d reminder still allows the 1h reminder to fire) or global (dismissing suppresses all intervals until next due cycle).

**Options**:
- [x] Per-interval dismissal - Each interval tracked independently, more granular control
- [ ] Global dismissal until next cycle - Single dismiss suppresses all, simpler mental model
- [ ] Snooze-style with duration - "Remind me in 30 minutes" regardless of configured intervals

**Impact**: Reminder UX, notification frequency. Affects daily usage patterns.

**Status**: answered

---

### [UX-010] Should cache TTL be user-configurable via config.yaml?

**Asked**: 2026-01-29
**Answered**: 2026-01-31
**Documented in**: `docs/explanation/caching.md`

**Context**: Feature is documented in docs/explanation/caching.md which states "The cache TTL can be configured in `config.yaml`: `cache_ttl: 5m`". However, the actual Config struct in `internal/config/config.go` has no cache TTL field. The TTL is hardcoded to 5 minutes in the test utility (`internal/testutil/cli.go`). Cannot create user-facing documentation for a config option that doesn't exist.

**Current documentation says**: "The cache TTL can be configured in config.yaml: cache_ttl: 5m"

**Missing details**:
- [x] Config file options (`cache_ttl` not in Config struct)
- [ ] CLI flags/commands (no cache management commands)

**Options**:
- [x] Add `cache_ttl` config option - Implement the config field described in explanation doc
- [ ] Update explanation doc - Remove the configurable TTL claim, document it as hardcoded 5 minutes
- [ ] Not user-facing - Cache behavior is internal and doesn't need user documentation

**Impact**: Blocks documentation of cache configuration. Users cannot currently adjust cache behavior.

**Status**: answered

---

### [UX-012] Should notification configuration be user-configurable via config.yaml?

**Asked**: 2026-01-29
**Answered**: 2026-01-31
**Documented in**: `docs/explanation/notification-manager.md`

**Context**: The explanation doc `docs/explanation/notification-manager.md` describes a `notification:` YAML config block with options like `os_notification.enabled`, `os_notification.on_sync_error`, `log_notification.path`, `log_notification.max_size_mb`, etc. However, the main `Config` struct in `internal/config/config.go` has no `Notification` field. The notification system's config is hardcoded in `cmd/todoat/cmd/todoat.go` (lines 7556-7570) with all channels always enabled. Users cannot currently configure notification behavior through config.yaml — only reminder delivery channels are configurable via `reminder.os_notification` and `reminder.log_notification`.

**Current documentation says**: "Configure desktop and log notifications" with a `notification:` YAML block.

**Missing details**:
- [x] Config file options (`notification:` block not in Config struct)
- [ ] CLI flags/commands (notification commands work correctly)

**Options**:
- [x] Add `notification` config to Config struct - Implement the config described in explanation doc
- [ ] Update explanation doc - Remove the `notification:` config block, document that notification channels are always enabled and controlled only through reminder config
- [ ] Keep as internal - Notification config is internal, reminder config controls user-facing notification preferences

**Impact**: Blocks accurate documentation of notification configuration. Users may try to add `notification:` to config.yaml based on explanation doc.

**Status**: answered

---

### [FEAT-011] docs/explanation/background-deamon.md is critically outdated and needs rewrite

**Asked**: 2026-01-29
**Answered**: 2026-01-31
**Documented in**: `docs/explanation/background-deamon.md`

**Context**: The "Current todoat Implementation Status" table (lines 42-49) and subsequent sections in `docs/explanation/background-deamon.md` describe the daemon as having:
- "In-process goroutine only" (Daemon process)
- "None - single process" (IPC/Socket)
- "CLI-driven background goroutines" (Sync mechanism)
- "Single backend sync only" (Multi-backend)

However, the actual code in `internal/daemon/daemon.go` has a fully implemented:
- **Forked process** via `Fork()` using `exec.Command` with `Setsid: true`
- **Unix domain socket IPC** with JSON message protocol (notify, status, stop)
- **Daemon-driven sync loop** with `time.NewTicker`
- **Multi-backend support** with per-backend intervals and failure isolation
- **Client library** (`daemon.Client`) with `Notify()`, `Status()`, `Stop()` methods

Additionally, line 373 states "There is no `todoat daemon start` command" but `todoat sync daemon start` exists and works.

**Options**:
- [x] Rewrite the explanation doc to match current implementation - Remove outdated status table, update all code examples and descriptions to reflect real forked process + IPC architecture
- [ ] Keep as historical context with clear "OUTDATED" markers - Preserve the design evolution but clearly mark which sections are superseded

**Impact**: The outdated explanation doc blocks accurate user-facing documentation. The how-to/sync.md has been updated with correct daemon behavior, but the explanation doc still contradicts the actual implementation.

**Status**: answered

---

### [FEAT-005] Should cache TTL be user-configurable?

**Asked**: 2026-01-29
**Answered**: 2026-02-06
**Documented in**: `docs/explanation/caching.md`

**Context**: List cache TTL is hardcoded to 5 minutes (`internal/testutil/cli.go`, cache implementation). This means list metadata can be stale for up to 5 minutes even if the remote has changes. Users with fast-changing backends or shared task lists may want shorter TTL; users on slow connections may want longer.

**Options**:
- [x] Add `cache.ttl` config option - Full user control
- [ ] Keep hardcoded but reduce default - 1 minute balances freshness and network usage
- [ ] Keep current 5-minute default - Acceptable for most use cases, not worth the config surface

**Resolution**: Implemented as `cache_ttl` config option (e.g., `"5m"`, `"30s"`, `"10m"`). Default remains 5 minutes. Documented in `docs/reference/configuration.md`, `internal/config/config.sample.yaml`, and `docs/explanation/caching.md`. Getter methods `GetCacheTTL()` and `GetCacheTTLDuration()` added to Config struct.

**Impact**: Data freshness vs network usage trade-off. Affects sync-enabled users.

**Status**: answered

---

### [UX-009] docs/explanation/interactive-ux.md needs rewrite to match implemented interactive prompt

**Asked**: 2026-01-29 (updated 2026-01-31)
**Answered**: 2026-02-08
**Documented in**: `docs/explanation/user-experience.md`

**Context**: The interactive prompt feature was implemented in commit `b6a6151` (2026-01-31). The code now includes:
- `internal/cli/prompt/prompt.go` - Full fuzzy-find task selection (318 lines)
- `internal/cli/prompt/prompt_test.go` - Comprehensive tests (661 lines)
- `ui.interactive_prompt_for_all_tasks` config option in `internal/config/config.go`
- Context-aware filtering by action type and interactive add mode with field validation

The explanation doc has been updated with minimal changes to remove "empty stub" references and document TaskSelector behavior. User-facing docs (`docs/how-to/task-management.md`) now describe interactive selection.

**Blocks**: User-facing documentation for `ui.interactive_prompt_for_all_tasks` in `docs/reference/configuration.md` and any interactive prompt how-to guides.

**Options**:
- [ ] Rewrite docs/explanation/interactive-ux.md - Update to document actual fuzzy-find prompt, config option, and context-aware filtering
- [x] Minimal update - Just fix the "empty stub" references and add config option mention

**Impact**: Blocks user-facing documentation for the interactive prompt config option and fuzzy-find behavior.

**Resolution**: Minimal update applied. The explanation doc (`docs/explanation/interactive-ux.md`) has been updated to remove empty stub references and document TaskSelector behavior with config option mention. User-facing documentation is unblocked.

**Status**: answered

---

### [ARCH-025] Update docs/explanation/background-deamon.md to reflect implemented heartbeat mechanism

**Asked**: 2026-02-06
**Answered**: 2026-02-08
**Documented in**: `docs/explanation/architecture.md`

**Context**: The daemon heartbeat mechanism was implemented in commit `de7491d` (Issue #74). User-facing documentation is now complete:
- `docs/reference/configuration.md` documents `sync.daemon.heartbeat_interval` (default: 5 seconds)
- `docs/how-to/sync.md` explains heartbeat monitoring and status output
- `internal/config/config.sample.yaml` includes the `heartbeat_interval` option

The explanation doc has been updated to describe the actual file-based heartbeat implementation:
- Heartbeat file location documented
- Configuration example added
- Status output examples added
- "NOT YET IMPLEMENTED" banner removed
- Planned vs implemented sections clarified

**Options**:
- [x] Update explanation doc - Rewrite the "Hung Daemon Detection" section to describe the actual file-based heartbeat implementation
- [ ] Remove planned sections - Delete the "NOT YET IMPLEMENTED" sections for features that are now implemented

**Impact**: Explanation doc accuracy. User-facing docs are already correct.

**Resolution**: Explanation doc updated to reflect the implemented heartbeat mechanism. The "Hung Daemon Detection" section now describes the actual file-based heartbeat implementation.

**Status**: answered

---

### [ARCH-020] Should config validation accept all 7 supported backends as default_backend?

**Asked**: 2026-01-31
**Answered**: 2026-02-08
**Documented in**: `docs/explanation/architecture.md`

**Context**: `Config.Validate()` in `internal/config/config.go:197` hardcodes `validBackends` to only `sqlite`, `todoist`, and `nextcloud`. However, the codebase implements 7 backends: sqlite, todoist, nextcloud, google, mstodo, file, and git. These additional backends are loaded dynamically via `GetBackendConfig()` using the raw config map, bypassing the typed `BackendsConfig` struct (which also only has 3 fields). Users setting `default_backend: google` will get a validation error even though the backend works.

**Options**:
- [x] Expand validation to all 7 backends - Add google, mstodo, file, git to `validBackends` map and `BackendsConfig` struct

**Impact**: Affects users of Google Tasks, MS Todo, File, and Git backends who want to set them as default. Config validation error vs runtime error trade-off.

**Status**: answered

---

### [FEAT-022] Should reminder.enabled default to true in DefaultConfig()?

**Asked**: 2026-01-31
**Answered**: 2026-02-08
**Documented in**: `docs/explanation/architecture.md`

**Context**: Decision FEAT-008 states analytics should be "enabled by default" and `DefaultConfig()` in `internal/config/config.go:120` sets `Analytics.Enabled: true`. However, reminders have no explicit default in `DefaultConfig()`, so `Reminder.Enabled` defaults to `false` (Go zero value for bool). The sample config (`config.sample.yaml:127-134`) has the entire reminder section commented out. This means new users get analytics enabled but reminders disabled out of the box, requiring explicit config to use reminders.

**Options**:
- [x] Add `Reminder: ReminderConfig{Enabled: true}` to DefaultConfig() - Reminders work out of the box for new users

**Impact**: New user onboarding experience. Users who add `--due-date` to tasks won't get reminders unless they also enable them in config. The "enable only when intervals configured" option provides a middle ground.

**Status**: answered

---

### [FEAT-027] Update docs/explanation/background-deamon.md: Stuck Task Detection is now implemented

**Asked**: 2026-02-08
**Answered**: 2026-02-08
**Documented in**: `docs/explanation/background-deamon.md`

**Context**: Stuck task detection and recovery was implemented in commit `a3659d3` (Issue #83). The code now includes:
- `--stuck-timeout` flag for `todoat sync daemon start` command (default: 10 minutes)
- `stuck_timeout` config option under `sync.daemon` section
- `GetStuckOperations` and `RecoverStuckOperations` methods in `backend/sync/sync.go`
- `GetStuckOperationsWithValidation` validates worker daemon liveness via heartbeat files before recovery

**Resolution**: Explanation doc updated (2026-02-08) - "Stuck Task Detection" section now marked as "(Implemented)" with actual implementation details.

**User-facing docs updated** (2026-02-08):
- `docs/reference/cli.md` - now documents `--stuck-timeout` flag
- `docs/reference/configuration.md` - now documents `sync.daemon.stuck_timeout` config option
- `docs/how-to/sync.md` - now documents stuck task recovery workflow

**Options**:
- [x] Update explanation doc - Move "Stuck Task Detection" from "Planned Enhancements" to an "Implemented" section, document the actual implementation

**Impact**: Explanation doc accuracy. User-facing docs are now complete.

**Status**: answered

---

### [FEAT-028] Update docs/explanation/background-deamon.md: Per-Task Timeout is now implemented

**Asked**: 2026-02-08
**Answered**: 2026-02-08
**Documented in**: `docs/explanation/background-deamon.md`

**Context**: Per-task timeout protection was implemented in commit `42f1b09` (Issue #84). The code now includes:
- `task_timeout` config option under `sync.daemon` section (default: "5m")
- `GetDaemonTaskTimeout()` method in `internal/config/config.go`
- `AddBackendSyncFuncWithContext` for context-aware sync functions
- `syncBackendWithTimeout` to wrap backend syncs with timeout

**Resolution**: Explanation doc updated (2026-02-08) - "Per-Task Timeout" section now marked as "(Implemented)" with actual implementation details. The "not yet implemented" statement at line 384 has been removed.

**User-facing docs updated** (2026-02-08):
- `docs/reference/configuration.md` - now documents `sync.daemon.task_timeout` config option
- `docs/how-to/sync.md` - now documents per-task timeout configuration
- `internal/config/config.sample.yaml` - now includes `task_timeout` example

**Options**:
- [x] Update explanation doc - Move "Per-Task Timeout" from "Planned Enhancements" to implemented, update line 364 to remove "not yet implemented"

**Impact**: Explanation doc accuracy. User-facing docs are now complete.

**Status**: answered
