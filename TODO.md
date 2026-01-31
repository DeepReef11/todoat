# Project TODO

## Documentation Tasks

_No documentation tasks pending._

---

## Questions for Team

### [FEAT-003] Should background logging be a runtime config option instead of compile-time constant?

**Context**: `internal/utils/logger.go` defines `const ENABLE_BACKGROUND_LOGGING = true` as a compile-time constant. This means background logging (writing to `/tmp/todoat-*-{PID}.log`) is always on and cannot be toggled without recompiling. Users have no control over whether these log files are created.

**Options**:
- [ ] Make it a config option (`logging.background_enabled: true/false`) - User control, matches other config patterns
- [ ] Keep as compile-time constant but default to false - Only developers enable it for debugging
- [ ] Remove entirely and rely on verbose mode (`-v`) - Simplify, one logging mechanism

**Impact**: Affects disk usage in `/tmp`, user privacy (logs may contain task content), debugging workflow.

**Asked**: 2026-01-29
**Status**: unanswered  <!-- User changes to "answered" or removes "un" when done -->

### [FEAT-004] What should happen when a notification backend tool is missing?

**Context**: OS notifications use platform-specific tools: `notify-send` (Linux), `osascript` (macOS), `powershell` (Windows). The Linux fallback is `wall` which broadcasts to all terminal sessions. The behavior when the primary tool is missing is not explicitly handled—it may silently fail or produce a confusing error.

**Options**:
- [ ] Silent skip with debug log - Don't bother user, log for debugging
- [ ] Warn once on first failure - Alert user their notification tool is missing, then suppress
- [ ] Fail loudly - Return error so caller can decide
- [ ] Auto-detect available tools at startup - Validate config and warn proactively

**Impact**: User experience for notifications and reminders. Affects all platforms.

**Asked**: 2026-01-29
**Status**: unanswered  <!-- User changes to "answered" or removes "un" when done -->

### [FEAT-005] Should cache TTL be user-configurable?

**Context**: List cache TTL is hardcoded to 5 minutes (`internal/testutil/cli.go`, cache implementation). This means list metadata can be stale for up to 5 minutes even if the remote has changes. Users with fast-changing backends or shared task lists may want shorter TTL; users on slow connections may want longer.

**Options**:
- [ ] Add `cache.ttl` config option - Full user control
- [ ] Keep hardcoded but reduce default - 1 minute balances freshness and network usage
- [ ] Keep current 5-minute default - Acceptable for most use cases, not worth the config surface

**Impact**: Data freshness vs network usage trade-off. Affects sync-enabled users.

**Asked**: 2026-01-29
**Status**: unanswered  <!-- User changes to "answered" or removes "un" when done -->

### [UX-006] Should auto-sync wait for completion or return immediately?

**Context**: When `auto_sync_after_operation: true`, CLI operations (create, update, delete) trigger a sync. It's unclear whether the CLI waits for sync to complete before returning to the user, or fires-and-forgets. Waiting gives users confidence their change is synced; returning immediately feels faster but leaves sync status ambiguous.

**Options**:
- [ ] Wait for sync completion - User sees success/failure, slower UX
- [ ] Fire-and-forget with background indicator - Return immediately, show sync status on next command
- [ ] Configurable (`sync.wait_for_completion: true/false`) - Let users choose their preference

**Impact**: CLI responsiveness vs sync reliability. Core user experience for synced workflows.

**Asked**: 2026-01-29
**Status**: unanswered  <!-- User changes to "answered" or removes "un" when done -->

### [ARCH-007] Should the daemon HeartbeatInterval field be removed or implemented?

**Context**: `DaemonConfig` in `internal/config/config.go` includes a `HeartbeatInterval` field, and tests reference heartbeat behavior (`daemon_test.go:842`), but no heartbeat recording or checking code exists. This creates a documentation-code gap where the config field is parseable but non-functional.

**Options**:
- [ ] Implement heartbeat mechanism - Enables hung daemon detection, matches documented architecture
- [ ] Remove the field - Clean up dead code, reduce config surface area
- [ ] Keep field but mark as reserved/future - Document that it's not yet functional

**Impact**: Config clarity, daemon reliability monitoring. Dead config fields may confuse users.

**Asked**: 2026-01-29
**Status**: unanswered  <!-- User changes to "answered" or removes "un" when done -->

### [UX-008] How should reminder dismissal interact with multiple intervals?

**Context**: Reminders support multiple intervals (e.g., `[1d, 1h, "at due time"]`). When a user dismisses a reminder, the docs say "you'll be reminded again at the next interval." It's unclear whether dismissal is per-interval (dismissing the 1d reminder still allows the 1h reminder to fire) or global (dismissing suppresses all intervals until next due cycle).

**Options**:
- [ ] Per-interval dismissal - Each interval tracked independently, more granular control
- [ ] Global dismissal until next cycle - Single dismiss suppresses all, simpler mental model
- [ ] Snooze-style with duration - "Remind me in 30 minutes" regardless of configured intervals

**Impact**: Reminder UX, notification frequency. Affects daily usage patterns.

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
- [ ] Minimal update - Just fix the "empty stub" references and add config option mention

**Impact**: Blocks user-facing documentation for the interactive prompt config option and fuzzy-find behavior.

**Asked**: 2026-01-29 (updated 2026-01-31)
**Status**: unanswered  <!-- User changes to "answered" or removes "un" when done -->

### [UX-010] Should cache TTL be user-configurable via config.yaml?

**Context**: Feature is documented in docs/explanation/caching.md which states "The cache TTL can be configured in `config.yaml`: `cache_ttl: 5m`". However, the actual Config struct in `internal/config/config.go` has no cache TTL field. The TTL is hardcoded to 5 minutes in the test utility (`internal/testutil/cli.go`). Cannot create user-facing documentation for a config option that doesn't exist.

**Current documentation says**: "The cache TTL can be configured in config.yaml: cache_ttl: 5m"

**Missing details**:
- [x] Config file options (`cache_ttl` not in Config struct)
- [ ] CLI flags/commands (no cache management commands)

**Options**:
- [ ] Add `cache_ttl` config option - Implement the config field described in explanation doc
- [ ] Update explanation doc - Remove the configurable TTL claim, document it as hardcoded 5 minutes
- [ ] Not user-facing - Cache behavior is internal and doesn't need user documentation

**Impact**: Blocks documentation of cache configuration. Users cannot currently adjust cache behavior.

**Asked**: 2026-01-29
**Status**: unanswered  <!-- User changes to "answered" or removes "un" when done -->

### [FEAT-011] docs/explanation/background-deamon.md is critically outdated and needs rewrite

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

The entire "Current Implementation Status", "Current Background Sync Patterns", "No Unix Socket Infrastructure", and "Conflicts with Existing Implementation" sections describe a pre-implementation state that no longer exists.

**Options**:
- [ ] Rewrite the explanation doc to match current implementation - Remove outdated status table, update all code examples and descriptions to reflect real forked process + IPC architecture
- [ ] Keep as historical context with clear "OUTDATED" markers - Preserve the design evolution but clearly mark which sections are superseded

**Impact**: The outdated explanation doc blocks accurate user-facing documentation. The how-to/sync.md has been updated with correct daemon behavior, but the explanation doc still contradicts the actual implementation.

**Asked**: 2026-01-29
**Status**: unanswered  <!-- User changes to "answered" or removes "un" when done -->

### [UX-012] Should notification configuration be user-configurable via config.yaml?

**Context**: The explanation doc `docs/explanation/notification-manager.md` describes a `notification:` YAML config block with options like `os_notification.enabled`, `os_notification.on_sync_error`, `log_notification.path`, `log_notification.max_size_mb`, etc. However, the main `Config` struct in `internal/config/config.go` has no `Notification` field. The notification system's config is hardcoded in `cmd/todoat/cmd/todoat.go` (lines 7556-7570) with all channels always enabled. Users cannot currently configure notification behavior through config.yaml — only reminder delivery channels are configurable via `reminder.os_notification` and `reminder.log_notification`.

**Current documentation says**: "Configure desktop and log notifications" with a `notification:` YAML block.

**Missing details**:
- [x] Config file options (`notification:` block not in Config struct)
- [ ] CLI flags/commands (notification commands work correctly)

**Options**:
- [ ] Add `notification` config to Config struct - Implement the config described in explanation doc
- [ ] Update explanation doc - Remove the `notification:` config block, document that notification channels are always enabled and controlled only through reminder config
- [ ] Keep as internal - Notification config is internal, reminder config controls user-facing notification preferences

**Impact**: Blocks accurate documentation of notification configuration. Users may try to add `notification:` to config.yaml based on explanation doc.

**Asked**: 2026-01-29
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
- [ ] Unify all verbose flags under global `-V` - Each subcommand checks the global verbose flag, remove local verbose flags
- [ ] Keep separate but rename local flags - Use `--detailed` or `--extended` for subcommand-specific extra output to avoid confusion with global `-V`
- [ ] Keep current behavior - Document the distinction between global debug verbosity (`-V`) and subcommand-specific detail flags (`--verbose`)

**Impact**: CLI consistency and discoverability. Affects `sync status`, `version`, and any future commands that want "show more detail" behavior.

**Asked**: 2026-01-31
**Status**: unanswered  <!-- User changes to "answered" or removes "un" when done -->
