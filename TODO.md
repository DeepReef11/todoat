# Project TODO

## Documentation Tasks

- [ ] Update `docs/explanation/task-management.md:203` - Todoist API reference says "REST API v2" but code was migrated to "API v1" in commit `91b911c`. The line "For Todoist: Sends POST request to REST API v2" should reference "API v1".

---

## Questions for Team

### [FEAT-013] What is the design intent for recurring tasks?

**Context**: Recurring tasks feature exists in code (`--recur` and `--recur-from-completion` CLI flags, `Recurrence` field in `backend/interface.go:32`, RRULE string format) and is documented in user-facing docs (`docs/how-to/task-management.md`, `docs/reference/cli.md`), but is not documented in `docs/explanation/`. The `docs/explanation/task-management.md` covers CRUD operations, status, priority, dates, and search but does not cover recurrence. Need to understand the design rationale (RRULE format choice, completion-based vs due-date-based recurrence, backend compatibility) before the explanation doc is complete.

**Options**:
- [ ] Add recurrence section to docs/explanation/task-management.md - Feature is part of task management, not a standalone subsystem
- [ ] Create separate docs/explanation/recurring-tasks.md - Feature is complex enough to warrant its own explanation doc
- [ ] Already covered sufficiently - The user-facing docs are adequate and no explanation doc is needed

**Impact**: Determines whether recurring tasks design rationale is documented in explanation docs, enabling complete user-facing documentation coverage.

**Asked**: 2026-01-29
**Status**: unanswered

### [FEAT-014] What is the design intent for the Terminal User Interface (TUI)?

**Context**: The TUI feature exists in code (`internal/tui/tui.go`, `todoat tui` command) and has a user-facing how-to guide (`docs/how-to/tui.md`), but is not documented in `docs/explanation/`. The TUI provides a two-pane interactive interface using Bubble Tea with keyboard-driven task management (add, edit, complete, delete, filter). Need to understand the design rationale (why Bubble Tea, interaction model choices, feature scope vs CLI) before the explanation doc is complete.

**Options**:
- [ ] Create docs/explanation/terminal-user-interface.md - TUI is a distinct interface mode warranting its own explanation doc
- [ ] Add TUI section to docs/explanation/cli-interface.md - TUI is part of the CLI interface, not a separate subsystem
- [ ] Already covered sufficiently - The how-to guide is adequate and no explanation doc is needed

**Impact**: Determines whether TUI design rationale is documented in explanation docs.

**Asked**: 2026-01-30
**Status**: unanswered

### [FEAT-015] What is the design intent for the migration system?

**Context**: The migration feature exists in code (`todoat migrate` command with `--from`, `--to`, `--list`, `--dry-run`, `--target-info` flags) and has a user-facing how-to guide (`docs/how-to/migration.md`), but has no dedicated explanation doc in `docs/explanation/`. Migration is partially mentioned in `docs/explanation/backends.md` but the design rationale for cross-backend task migration (metadata preservation, UID handling, batch processing, status mapping) is not documented.

**Options**:
- [ ] Create docs/explanation/migration-system.md - Migration is a distinct subsystem warranting its own explanation doc
- [ ] Add migration section to docs/explanation/backends.md - Migration is part of the backend system, not a standalone feature
- [ ] Already covered sufficiently - The how-to guide and backends.md coverage are adequate

**Impact**: Determines whether migration design rationale is documented in explanation docs.

**Asked**: 2026-01-30
**Status**: unanswered

### [FEAT-016] What is the design intent for the reminder system?

**Context**: The reminder feature exists in code (`internal/reminder/`, `todoat reminder` command with subcommands `check`, `list`, `disable`, `dismiss`, `status`) and has a user-facing how-to guide (`docs/how-to/reminders.md`) plus configuration reference in `docs/reference/configuration.md`. However, there is no dedicated explanation doc in `docs/explanation/`. The notification-manager.md mentions reminders only briefly. Need to understand the design rationale (interval-based reminders, dismissal semantics, integration with notification system, cron vs daemon-driven checks) before the explanation doc is complete.

**Options**:
- [ ] Create docs/explanation/reminder-system.md - Reminders are a distinct subsystem warranting their own explanation doc
- [ ] Add reminder section to docs/explanation/notification-manager.md - Reminders are part of the notification system
- [ ] Already covered sufficiently - The how-to guide and configuration reference are adequate

**Impact**: Determines whether reminder design rationale is documented in explanation docs.

**Asked**: 2026-01-30
**Status**: unanswered

### [FEAT-017] What is the design intent for time-of-day support in task dates?

**Context**: Time-of-day support exists in code (`--due-date "2026-01-20T14:30"`, `--due-date "tomorrow 09:00"`, timezone handling) and is documented in user-facing docs (`docs/how-to/task-management.md`, `docs/reference/cli.md`), but `docs/explanation/task-management.md` only documents "YYYY-MM-DD (date only)" as the user input format. The explanation doc needs updating to cover ISO 8601 datetime input, relative dates with time (`tomorrow 14:30`, `+7d 09:00`), timezone offset support, and how time-of-day is stored and displayed. This was implemented in roadmap item 058.

**Options**:
- [ ] Update docs/explanation/task-management.md date section - Add time-of-day input formats, timezone handling, and display behavior
- [ ] Already covered sufficiently - The user-facing docs are adequate and explanation doc update is low priority

**Impact**: Ensures explanation doc accurately reflects the implemented date/time handling. User-facing docs are already correct.

**Asked**: 2026-01-30
**Status**: unanswered

### [FEAT-018] What is the design intent for the file watcher daemon feature?

**Context**: File watcher feature exists in code (`internal/watcher/watcher.go`, config fields `sync.daemon.file_watcher`, `sync.daemon.smart_timing`, `sync.daemon.debounce_ms` in `internal/config/config.go`) but is not documented in `docs/explanation/`. The explanation doc `docs/explanation/synchronization.md` only mentions "File watcher for real-time sync triggers" as a bullet point under "Future Auto-Sync Plans". The feature adds `fsnotify`-based file watching to trigger sync when local cache files change, with smart timing to avoid syncing during active editing and configurable debounce. Need to understand the design rationale before creating user-facing documentation.

**Options**:
- [ ] Add file watcher section to docs/explanation/synchronization.md - Feature is part of sync, update the "Future Plans" section to reflect implementation
- [ ] Create separate docs/explanation/file-watcher.md - Feature is complex enough to warrant its own explanation doc
- [ ] Wait until committed - Feature code is currently uncommitted, document after merge

**Impact**: Blocks user-facing documentation for file watcher config options and daemon behavior. Currently the config fields `file_watcher`, `smart_timing`, `debounce_ms` exist but are undocumented.

**Asked**: 2026-01-30
**Status**: unanswered

### [UX-019] Should the verbose flag be unified across all subcommands?

**Context**: The CLI has three different verbose flag patterns: global `-V`/`--verbose` (persistent flag on root, enables debug output via `utils.SetVerboseMode`), `sync status --verbose` (local flag, no short form, shows sync metadata), and `version --verbose/-v` (local flag with lowercase `-v`, shows build info). Commit `8a05147` documents the `sync status` inconsistency in the CLI reference. Users expecting `-v` to work globally will get an error on most commands, and `-V` doesn't affect `sync status` verbose output.

**Options**:
- `-v` is for view. Use `--verbose` only

**Impact**: CLI consistency and discoverability. Affects `sync status`, `version`, and any future commands that want "show more detail" behavior.

**Asked**: 2026-01-31
**Status**: unanswered  <!-- User changes to "answered" or removes "un" when done -->

### [ARCH-021] Should RetentionDays use consistent types across TrashConfig and AnalyticsConfig?

**Context**: `TrashConfig.RetentionDays` uses `*int` (pointer) in `internal/config/config.go:60`, allowing nil (default 30), explicit 0 (disabled), and positive values. `AnalyticsConfig.RetentionDays` uses plain `int` in `config.go:55`, where 0 (Go zero value) triggers the default of 365, making it impossible to explicitly set retention to 0 days (immediate purge). The `GetAnalyticsRetentionDays()` method at line 350 treats `<= 0` as "use default 365".

**Options**:
- [ ] Make both `*int` (pointer) - Consistent, allows explicit 0 for both (trash: keep forever, analytics: purge immediately)
- [ ] Keep current inconsistency - Analytics should never have 0 retention (data loss), so the restriction is intentional
- [ ] Make both plain `int` with documented minimum - Use 0 as "use default" for both, add explicit `disabled` boolean if needed

**Impact**: Config consistency and user expectations. A user setting `analytics.retention_days: 0` gets 365 days silently, which may be surprising.

**Asked**: 2026-01-31
**Status**: unanswered  <!-- User changes to "answered" or removes "un" when done -->

### [FEAT-023] Nextcloud suppress_ssl_warning and suppress_http_warning not wired in CLI

**Context**: The sample config (`internal/config/config.sample.yaml:16-19`) and explanation docs (`docs/explanation/backends.md:62-64`, `docs/explanation/configuration.md:81-84`) document `suppress_ssl_warning` and `suppress_http_warning` config options for the Nextcloud backend. However, these fields are not in the `nextcloud.Config` struct (`backend/nextcloud/nextcloud.go:25-32`) and `buildNextcloudConfigWithKeyring` in `cmd/todoat/cmd/todoat.go:2892-2930` only reads `insecure_skip_verify` and `allow_http`. No actual SSL/HTTP warnings are emitted by the code, making these settings no-ops.

**Options**:
- [ ] Implement warning system - Add SSL/HTTP security warnings when using `insecure_skip_verify` or `allow_http`, then add suppress options
- [ ] Remove from docs - These are aspirational features that don't exist; remove from sample config and explanation docs
- [ ] Low priority - Document that these are planned features, keep in sample config but mark as "not yet implemented"

**Impact**: Documentation claims features that don't work. Users setting these options get no effect.

**Asked**: 2026-02-06
**Status**: unanswered

### [FEAT-024] Git backend fallback_files and auto_detect not wired in CLI

**Context**: The sample config (`internal/config/config.sample.yaml:34-35`) and explanation docs (`docs/explanation/backends.md:336-340`, `docs/explanation/configuration.md:292-293`) document `fallback_files` and `auto_detect` config options for the Git backend. While `FallbackFiles` exists in `git.Config` struct (`backend/git/git.go:34`) and is used internally, `createCustomBackend` in `cmd/todoat/cmd/todoat.go:2849-2861` only reads `work_dir`, `file`, and `auto_commit` - it does NOT read `fallback_files` from config YAML. Additionally, `auto_detect` is documented but doesn't exist in the `git.Config` struct at all.

**Options**:
- [ ] Wire fallback_files in CLI - Add `if fallbackFiles, ok := backendCfg["fallback_files"].([]interface{}); ok` in createCustomBackend
- [ ] Wire auto_detect option - Add AutoDetect field to git.Config and read from config; clarify difference from global `auto_detect_backend`
- [ ] Remove from docs - These are aspirational features that don't work; remove from sample config and explanation docs
- [ ] Document the gap - Note that these features are internal defaults only, not user-configurable

**Impact**: Documentation claims features that don't work from config. The git backend uses hardcoded fallback files internally but config values are ignored.

**Asked**: 2026-02-06
**Status**: unanswered

### [FEAT-026] Notification config block not wired in Config struct (UX-012 implementation pending)

**Context**: Decision UX-012 (2026-01-31) approved adding a `notification:` config block to the Config struct. The sample config (`internal/config/config.sample.yaml:122-130`) and explanation docs (`docs/explanation/notification-manager.md`, `docs/explanation/configuration.md:402-417`) describe this config block with options like `os_notification.enabled`, `os_notification.on_sync_error`, `log_notification.path`, etc.

However, the Config struct in `internal/config/config.go:35-49` still has no `Notification` field. Only reminder-specific notification options (`reminder.os_notification`, `reminder.log_notification`) are wired. Users setting the documented `notification:` block will have it silently ignored.

**Options**:
- [ ] Implement UX-012 - Add `Notification NotificationConfig` to Config struct and wire it in CLI
- [ ] Update docs to reflect current state - Document that only reminder notification options are available until full notification config is implemented
- [ ] Low priority - The reminder config covers the primary use case; full notification config is a nice-to-have

**Impact**: Documentation claims features that don't work from config. Decision UX-012 approved implementation but it hasn't been done.

**Asked**: 2026-02-06
**Status**: unanswered

### [ARCH-029] Merge conflict strategy documentation does not match implementation

**Context**: Decision ARCH-007 (2026-01-26) approved field-level timestamp-based merge resolution, and `docs/explanation/synchronization.md:424-429` documents this behavior:
- "Timestamps: Use latest `modified` time"
- "Categories: Union of both sets"
- "Description: Use remote if changed, else keep local"

However, the actual implementation in `cmd/todoat/cmd/todoat.go:7399-7413` uses simple hardcoded field selection:
- Summary: Always remote (not timestamp-based)
- Description: Always remote (not conditional)
- Priority: Always local
- Categories: Always local (not union)
- Status: Always remote

The field-level timestamp tracking mentioned in ARCH-007 does not appear to be implemented. Users expecting intelligent merge behavior per the documentation will get the simpler hardcoded version.

**Options**:
- [ ] Implement ARCH-007 - Add field-level modification tracking and use timestamps for merge resolution
- [ ] Update documentation - Change `docs/explanation/synchronization.md:424-429` to reflect actual behavior (hardcoded field selection)
- [ ] Keep as-is - The simple implementation is intentional; ARCH-007 is a future enhancement

**Impact**: Documentation accuracy. Users who choose "merge" strategy may not get expected results if their mental model matches the documented behavior rather than the actual implementation.

**Asked**: 2026-02-08
**Status**: unanswered

### [ARCH-031] ClaimNextOperation uses BEGIN instead of BEGIN IMMEDIATE despite comment

**Context**: Commit `34edf10` added atomic task claiming for the sync queue to prevent race conditions when multiple daemon instances coexist. The code at `cmd/todoat/cmd/todoat.go:7814-7815` has a comment saying "Use BEGIN IMMEDIATE to acquire exclusive write lock immediately" but calls `sm.db.Begin()`, which starts a deferred transaction in Go's `database/sql`. A deferred transaction only acquires the write lock on the first write statement, creating a window for SQLITE_BUSY errors. To achieve `BEGIN IMMEDIATE` semantics with `modernc.org/sqlite`, you would need `sm.db.BeginTx(ctx, &sql.TxOptions{})` with a pragma or execute `BEGIN IMMEDIATE` via raw SQL.

**Options**:
- [ ] Use raw SQL `BEGIN IMMEDIATE` - Execute `sm.db.Exec("BEGIN IMMEDIATE")` and manage the transaction manually
- [ ] Accept deferred BEGIN - The current behavior is sufficient because concurrent daemons are short-lived and SQLITE_BUSY retries handle contention
- [ ] Use BeginTx with serializable isolation - Use `sm.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})` if the driver supports it

**Impact**: The atomicity guarantee from Issue #81 may not hold under concurrent daemon scenarios. The write lock acquisition window allows two daemons to both begin transactions before either writes.

**Asked**: 2026-02-12
**Status**: unanswered

### [API-032] DeleteList interface says "soft-delete" but 5 of 7 backends hard-delete

**Context**: The `TaskManager` interface at `backend/interface.go:64` defines `DeleteList` with the comment "Soft-delete (move to trash)". However, only SQLite implements soft-delete (sets `deleted_at`). Nextcloud returns an error ("would be permanent"). Todoist, Git, File, Google, and MS Todo all perform permanent hard deletes. The interface also defines `PurgeList` at line 70 as "Permanent delete", creating an unclear distinction when `DeleteList` is already permanent for most backends. A user on SQLite who deletes a list can restore it via `RestoreList`, but switching to any other backend and deleting a list loses data permanently with no warning.

**Options**:
- [ ] Update interface contract - Change comment to "Delete list (may be permanent depending on backend)" and add backend-specific documentation
- [ ] Add CLI warning - Before `DeleteList` on non-SQLite backends, warn the user that deletion is permanent and require confirmation
- [ ] Standardize to soft-delete - Implement soft-delete in all backends (may not be possible for API-based backends like Todoist/Google)
- [ ] Keep as-is - Document the difference in backend docs, accept that behavior varies

**Impact**: Silent data loss when users switch backends. A user accustomed to SQLite's trash/restore behavior will permanently lose data on other backends without warning.

**Asked**: 2026-02-12
**Status**: unanswered

### [ARCH-035] Multi-backend daemon bypasses error loop prevention (consecutiveErrors never increments)

**Context**: The daemon error loop prevention from commit `7c1d1ed` uses a `consecutiveErrors` counter at `internal/daemon/daemon.go:383` to shut down the daemon after repeated failures. However, when multi-backend sync is used (line 531-535), `performMultiBackendSync` always sets `syncFailed = false`, so `consecutiveErrors` never increments. Per-backend error tracking exists (`state.ErrorCount`) but there is no per-backend or global threshold that triggers daemon shutdown. If a backend has persistently invalid credentials or is permanently unreachable, the daemon retries forever with no backoff and no shutdown.

**Options**:
- [ ] Add per-backend error threshold - After N consecutive errors per backend, disable that backend and log a warning; shut down if all backends disabled
- [ ] Apply global threshold to multi-backend - If all backends fail in a single sync cycle, increment `consecutiveErrors` normally
- [ ] Accept current behavior - Multi-backend is designed to be resilient; one failing backend shouldn't affect others, and the daemon should keep running

**Impact**: Resource consumption from infinite retry loops when backends are permanently unreachable. The error loop prevention feature from Issue #82 does not apply to the multi-backend path.

**Asked**: 2026-02-12
**Status**: unanswered

### [ARCH-036] Stuck task recovery resets actively-processing tasks when heartbeat is disabled

**Context**: `isWorkerDead` at `cmd/todoat/cmd/todoat.go:7998-8001` returns `true` when `heartbeatDir == ""` (heartbeat not configured). Since heartbeat settings are commented out in `config.sample.yaml`, this is the default. With default config (`stuck_timeout: 10m`, no heartbeat), if a sync operation takes longer than 10 minutes (slow network, large dataset), `GetStuckOperationsWithValidation` treats it as stuck and `RecoverStuckOperations` resets it to `pending`. The task is then re-processed, causing duplicate sync operations. The daemon heartbeat mechanism (commit `de7491d`) was designed to differentiate between actually-stuck tasks and slow-but-healthy ones, but it's disabled by default.

**Options**:
- [ ] Enable heartbeat by default - Set reasonable defaults for `heartbeat_path` and `heartbeat_interval` so stuck detection works correctly out of the box
- [ ] Increase default stuck_timeout - Use a much larger default (e.g., 30m or 1h) to reduce false positives when heartbeat is disabled
- [ ] Change isWorkerDead default - When heartbeat is not configured, assume worker is alive (return false) instead of dead; only enable stuck recovery when heartbeat is configured
- [ ] Accept current behavior - 10 minutes is generous; if sync takes longer, the duplicate is a reasonable trade-off for recovering from actual stuck tasks

**Impact**: False positive stuck detection causes duplicate sync operations, which could create duplicate tasks on remote backends or trigger unnecessary conflict resolution.

**Asked**: 2026-02-12
**Status**: unanswered

### [FEAT-037] validBackends map missing google, mstodo, git, and file backends

**Context**: The `validBackends` map in `internal/config/config.go:207` only includes `sqlite`, `todoist`, and `nextcloud`. Setting `default_backend` to `google`, `mstodo`, `git`, or `file` via `config set` will fail validation with "unknown default_backend" even though these backends are fully functional and documented. The `question-log.md` already identifies this as an action item ("Expand validation to all 7 backends") but it hasn't been implemented.

**Options**:
- [ ] Expand validBackends map - Add all 7 backends to the validation map
- [ ] Dynamic validation - Build valid backend list from registered backends instead of a hardcoded map

**Impact**: Users following documentation to configure google/mstodo/git/file as default backend will get a validation error. Workaround: edit config YAML directly instead of using `config set`.

**Asked**: 2026-02-12
**Status**: unanswered

### [ARCH-038] Circuit breaker half-open state allows unlimited probes (docs say "single probe")

**Context**: The circuit breaker documentation at `docs/how-to/sync.md:201` states "a single probe sync is attempted" after the cooldown expires. However, the implementation at `internal/daemon/circuitbreaker.go:80-82` allows ALL requests through in half-open state (`case CircuitHalfOpen: return true`). There is no counter limiting half-open to a single probe. If multiple sync triggers fire during the half-open window (e.g., ticker + IPC notify), all of them proceed, potentially overwhelming a recovering backend.

**Options**:
- [ ] Limit to single probe - Add a counter so only one request passes in half-open; subsequent requests are blocked until the probe resolves
- [ ] Keep unlimited half-open - The sync mutex (`syncMu` in daemon.go:552) already serializes concurrent syncs, so multiple probes can't actually happen simultaneously
- [ ] Update documentation - Change "single probe" to reflect actual behavior

**Impact**: Documentation accuracy and backend recovery behavior. The sync mutex may already prevent the thundering herd concern, making this a documentation-only fix.

**Asked**: 2026-02-14
**Status**: unanswered

### [FEAT-039] Circuit breaker threshold and cooldown are hardcoded, not configurable

**Context**: The circuit breaker at `internal/daemon/circuitbreaker.go:13,17` uses hardcoded constants: `DefaultCircuitBreakerThreshold = 3` failures and `DefaultCircuitBreakerCooldown = 30s`. These values are not exposed in the config registry (`cmd/todoat/cmd/todoat.go` config set section has no circuit breaker entries) and cannot be changed via `config set` or the YAML file. Users with flaky networks may need a higher threshold (5+ failures), and users on stable corporate networks may want tighter settings (2 failures, 15s cooldown).

**Options**:
- [ ] Add config options - Add `sync.daemon.circuit_breaker_threshold` and `sync.daemon.circuit_breaker_cooldown` to config struct, sample config, and config set registry
- [ ] Keep hardcoded - The defaults are reasonable for most users; adding config complexity is not worth it for an internal resilience mechanism
- [ ] Add config but keep current defaults - Wire the config but don't document heavily; power users can discover it

**Impact**: Affects daemon resilience tuning. Users with specific network conditions cannot adjust circuit breaker sensitivity without code changes.

**Asked**: 2026-02-14
**Status**: unanswered

### [FEAT-040] backend_priority in sample config has no Config struct field

**Context**: The sample config at `internal/config/config.sample.yaml:54` documents `backend_priority: [nextcloud, git, sqlite]` and commit `ae74316` added it. However, there is no `BackendPriority` field in the Config struct at `internal/config/config.go`. Setting this value in config YAML is silently ignored. The feature implies backends should be tried in a specific order during auto-detection or multi-backend sync, but no code reads this setting.

**Options**:
- [ ] Implement backend_priority - Add `BackendPriority []string` to Config struct and wire it into auto-detection and/or multi-backend sync order
- [ ] Remove from sample config - The setting is aspirational and misleading; remove until implemented
- [ ] Document as planned - Add a comment in sample config noting it's not yet implemented

**Impact**: Users setting `backend_priority` expect it to affect behavior. Silent no-op creates false confidence in backend ordering.

**Asked**: 2026-02-14
**Status**: unanswered

### [ARCH-041] Documentation uses ParentUID but code uses ParentID

**Context**: The Task struct at `backend/interface.go:30` defines the field as `ParentID string`. However, all explanation docs consistently use `ParentUID`: `docs/explanation/task-management.md:121,1285`, `docs/explanation/subtasks-hierarchy.md:284-416`, `docs/explanation/synchronization.md:734`, `docs/explanation/backend-system.md:477`, and `docs/explanation/README.md:88`. This naming inconsistency means developers reading docs will look for `ParentUID` in code and not find it.

**Options**:
- [ ] Rename code field to ParentUID - Aligns with docs; "UID" is more accurate for CalDAV-based backends where IDs are UUIDs
- [ ] Update docs to ParentID - Aligns with code; simpler name, consistent with other ID fields (ListID, ID)
- [ ] Keep both - Add `ParentUID` as an alias comment on the struct field; update docs to mention both names

**Impact**: Developer confusion when cross-referencing docs and code. No runtime impact since the field works regardless of name.

**Asked**: 2026-02-14
**Status**: unanswered

### [UX-042] Multi-backend partial failure sends "sync completed" notification

**Context**: When multi-backend sync runs and some backends succeed while others fail, `performSync()` at `internal/daemon/daemon.go:569-579` sets `result` to `syncSuccess` (zero value) because `allFailed` is false. The notification at line 1006 then sends "Sync completed" even though one or more backends failed. For example, if Nextcloud succeeds but Todoist fails due to expired token, the user sees "Sync completed" and has no indication of the Todoist failure. The `syncFailed` notification only fires when ALL backends fail.

**Options**:
- [ ] Add partial failure notification - Introduce `syncPartial` result type that sends "Sync completed with errors (1 of 3 backends failed)" notification
- [ ] Keep current behavior - Users can check `daemon status` for per-backend errors; notifications should only alert on total failure
- [ ] Add per-backend failure notification - Send individual "Backend X sync failed" notifications for each failing backend, in addition to the overall result

**Impact**: Users may miss persistent backend failures masked by other backends succeeding. Particularly risky when the failing backend contains the user's primary task data.

**Asked**: 2026-02-14
**Status**: unanswered
