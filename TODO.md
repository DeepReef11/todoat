# Project TODO

## Documentation Tasks

_No documentation tasks pending._

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
