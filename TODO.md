# Project TODO

## Documentation Tasks

_No documentation tasks pending._

---

## Questions for Team

### [FEAT-005] Should cache TTL be user-configurable?

**Context**: List cache TTL is hardcoded to 5 minutes (`internal/testutil/cli.go`, cache implementation). This means list metadata can be stale for up to 5 minutes even if the remote has changes. Users with fast-changing backends or shared task lists may want shorter TTL; users on slow connections may want longer.

**Options**:
- [ ] Add `cache.ttl` config option - Full user control
- [ ] Keep hardcoded but reduce default - 1 minute balances freshness and network usage
- [ ] Keep current 5-minute default - Acceptable for most use cases, not worth the config surface

**Impact**: Data freshness vs network usage trade-off. Affects sync-enabled users.

**Asked**: 2026-01-29
**Status**: unanswered  <!-- User changes to "answered" or removes "un" when done -->

### [UX-009] docs/explanation/interactive-ux.md needs rewrite to match implemented interactive prompt

**Context**: The interactive prompt feature was implemented in commit `b6a6151` (2026-01-31). The code now includes:
- `internal/cli/prompt/prompt.go` - Full fuzzy-find task selection (318 lines)
- `internal/cli/prompt/prompt_test.go` - Comprehensive tests (661 lines)
- `ui.interactive_prompt_for_all_tasks` config option in `internal/config/config.go`
- Context-aware filtering by action type and interactive add mode with field validation

However, `docs/explanation/interactive-ux.md` still describes the prompt as "An empty stub exists at `internal/cli/prompt/prompt.go` intended for future prompt enhancements" (line 19) and lists it as "Empty, reserved for future use" (line 75). The explanation doc needs to be rewritten to document the actual implementation, including fuzzy-find behavior, context-aware filtering, auto-selection for single matches, and the `ui.interactive_prompt_for_all_tasks` config option.

**Blocks**: User-facing documentation for `ui.interactive_prompt_for_all_tasks` in `docs/reference/configuration.md` and any interactive prompt how-to guides.

**Options**:
- [ ] Rewrite docs/explanation/interactive-ux.md - Update to document actual fuzzy-find prompt, config option, and context-aware filtering
- [x] Minimal update - Just fix the "empty stub" references and add config option mention

**Impact**: Blocks user-facing documentation for the interactive prompt config option and fuzzy-find behavior.

**Asked**: 2026-01-29 (updated 2026-01-31)
**Status**: unanswered  <!-- User changes to "answered" or removes "un" when done -->

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

### [ARCH-020] Should config validation accept all 7 supported backends as default_backend?

**Context**: `Config.Validate()` in `internal/config/config.go:197` hardcodes `validBackends` to only `sqlite`, `todoist`, and `nextcloud`. However, the codebase implements 7 backends: sqlite, todoist, nextcloud, google, mstodo, file, and git. These additional backends are loaded dynamically via `GetBackendConfig()` using the raw config map, bypassing the typed `BackendsConfig` struct (which also only has 3 fields). Users setting `default_backend: google` will get a validation error even though the backend works.

**Options**:
- [ ] Expand validation to all 7 backends - Add google, mstodo, file, git to `validBackends` map and `BackendsConfig` struct
- [ ] Keep validation strict to typed backends only - The 4 additional backends are "dynamic" and don't need config struct fields; remove them from `default_backend` validation
- [ ] Remove `default_backend` validation for dynamic backends - Only validate the 3 typed backends; skip validation for unknown names (let runtime discovery fail instead)

**Impact**: Affects users of Google Tasks, MS Todo, File, and Git backends who want to set them as default. Config validation error vs runtime error trade-off.

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

### [FEAT-022] Should reminder.enabled default to true in DefaultConfig()?

**Context**: Decision FEAT-008 states analytics should be "enabled by default" and `DefaultConfig()` in `internal/config/config.go:120` sets `Analytics.Enabled: true`. However, reminders have no explicit default in `DefaultConfig()`, so `Reminder.Enabled` defaults to `false` (Go zero value for bool). The sample config (`config.sample.yaml:127-134`) has the entire reminder section commented out. This means new users get analytics enabled but reminders disabled out of the box, requiring explicit config to use reminders.

**Options**:
- [ ] Add `Reminder: ReminderConfig{Enabled: true}` to DefaultConfig() - Reminders work out of the box for new users
- [ ] Keep current behavior (disabled by default) - Reminders require explicit opt-in, avoids unexpected notifications for users who haven't configured intervals
- [ ] Enable reminders only when intervals are configured - Auto-enable if user sets reminder intervals, skip if no intervals defined

**Impact**: New user onboarding experience. Users who add `--due-date` to tasks won't get reminders unless they also enable them in config. The "enable only when intervals configured" option provides a middle ground.

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
