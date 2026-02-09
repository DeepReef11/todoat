# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Stuck task detection and recovery for sync queue (#83)
  - `GetStuckOperations` and `RecoverStuckOperations` methods detect tasks stuck in 'processing' state
  - `GetStuckOperationsWithValidation` validates worker daemon liveness via heartbeat files before recovery
  - `--stuck-timeout` flag for `sync daemon start` command (default: 10 minutes)
  - `stuck_timeout` config option under `sync.daemon` section
  - Prevents sync queue stalls when daemon crashes mid-task
- Daemon error loop prevention with exponential backoff (#82)
  - `MaxConsecutiveErrors` constant (5) triggers graceful shutdown after repeated sync failures
  - Exponential backoff between retries: 2^n seconds, capped at 60 seconds
  - Prevents infinite error loops that could consume system resources
- Sync queue atomic task claiming with deduplication (#81)
  - New `ClaimNextOperation` method uses BEGIN IMMEDIATE for exclusive lock
  - Prevents race conditions when multiple daemon instances briefly coexist
  - New `status`, `worker_id`, `claimed_at` columns with automatic migration
- Interactive task selection via TaskSelector when multiple tasks match a search term (#79)
  - Wired TaskSelector to CLI commands (complete, delete, update, add with parent)
  - Falls back to error with disambiguation info when `--no-prompt` mode is enabled

### Fixed
- Fixed sync queue schema initialization to create indexes after column migrations (#86)
  - Prevents `no such column: status` error on fresh databases with old schema
  - Index creation now runs after `migrateSyncQueueSchema` ensures columns exist
- `todoat sync` now uses backends configured in the `backends:` section even when `default_backend` is not set (#80)
  - Sync command iterates over all enabled remote backends in the `backends:` section
  - Per-backend failure isolation ensures one backend failure doesn't block others

### Added
- `cache_ttl` config option for user-configurable list metadata cache TTL (e.g., `"5m"`, `"30s"`, `"10m"`)
  - `GetCacheTTL()` and `GetCacheTTLDuration()` getter methods on Config struct
  - Default remains 5 minutes for backwards compatibility

### Changed
- Removed redundant `ResultInfoOnly` output from info-only CLI commands in no-prompt mode (list, get, stats, status, queue views)
- Updated `docs/explanation/background-deamon.md` to document implemented daemon features: atomic task claiming (#81), error loop prevention (#82), sync queue schema
- Updated `docs/how-to/sync.md` with error recovery section documenting exponential backoff behavior
- Updated `docs/explanation/interactive-ux.md` to document TaskSelector component and interactive task selection behavior
- OS notification channel now consolidated into single cross-platform implementation (`internal/notification/os.go`)
- Default view config path lookup now uses `config.GetConfigDir()` directly instead of deriving from DBPath

### Added
- `logging.background_enabled` config option to control background log file creation (default: true)
- `NewBackgroundLoggerWithEnabled()` function for runtime config-aware background logging

### Changed
- Background logging is now controlled via config instead of compile-time constant
- Updated logging docs to reflect config-based background logging control

### Security
- `columnExists` function now validates table names against an allowlist to prevent SQL injection in PRAGMA queries (#72)

### Fixed
- Plugin commands are now validated to be within the plugin directory, preventing arbitrary command execution via malicious view YAML files (#73)

### Added
- Daemon heartbeat mechanism for hung process detection (#74)
  - `sync.daemon.heartbeat_interval` config option (seconds, default: 5)
  - `daemon status` now displays heartbeat health when enabled
  - Automatic cleanup of heartbeat file on daemon stop
- User experience design decision document (`docs/explanation/user-experience.md`)

### Changed
- Updated explanation docs (architecture, background-daemon, caching, logging, notification-manager, synchronization) to match current implementation
- Updated `docs/how-to/reminders.md` with design decision references
- Updated `docs/reference/configuration.md` with config behavior clarifications
- Recorded multiple design decisions in question log

### Fixed
- `daemon start` now always forks a real background daemon regardless of `daemon.enabled` config setting; the feature-flag check is only used for auto-start gating, preventing stale PID files when config omits `daemon.enabled` (#59)

### Changed
- Updated `-P/--parent` flag description to clarify it accepts path-style names (e.g., `"Parent/Child"`)

### Added
- Interactive prompt package (`internal/cli/prompt`) with fuzzy-find task selection, context-aware filtering by action, and interactive add mode with field validation (#48)
- `ui.interactive_prompt_for_all_tasks` config option to include completed/cancelled tasks in interactive prompts
- `--all` / `-a` flag support for showing terminal tasks in interactive selection
- Backend setup how-to guide with configuration examples for all backends (SQLite, Nextcloud, Todoist, Google Tasks, Microsoft To Do, Git, File)
- Backend configuration reference table in configuration docs

### Changed
- `list` command JSON output now wraps results in an object with `lists` and `result` fields instead of a bare array (#67)
- Updated list-management docs JSON examples to use correct field name `tasks` instead of `task_count`
- Updated cross-reference links to point to new backend setup guide

### Fixed
- Fixed separate-process regression test for Issue #59 to use correct XDG config directory structure (`XDG_CONFIG_HOME/todoat/config.yaml`)
- Daemon feature check, config interval lookup, and daemon start now fall back to default config path when `ConfigPath` is empty instead of silently returning early
- Corrected `--json` flag position in analytics docs (global flag goes before subcommand)
- Fixed backend error message in errors.md to match actual CLI output (no quoted backend name)
- Updated completion subcommand descriptions to match Cobra-generated help text
- Added note that `--verbose` flag is not available on `version` command

### Added
- Documentation for `--json` flag on `sync queue`, `sync daemon status`, `reminder check`, `notification log`, and `view list` commands
- Regression test for Issue #60: backend error message matches documentation
- Regression test for Issue #59: in-process daemon status returns actual running interval via IPC instead of config default
- Separate-process regression test for Issue #59: verifies daemon status shows actual interval when start and status are separate CLI invocations
- Rate limit error documentation with causes, examples, and troubleshooting steps
- Tests for verbose debug timestamp output (#47)

### Changed
- Verbose debug output now includes HH:MM:SS timestamp prefix for easier log correlation (#47)
- Improved CLI reference descriptions for `config get`, `config set`, `credentials set`, `sync daemon kill`, and `--recur` flag
- Updated sync daemon status docs with example output showing PID, interval, sync count, and last sync time
- Updated task matching docs: single match completes directly without confirmation prompt; multiple matches show error with UIDs instead of interactive menu
- Added daemon configuration section to sample config with `enabled`, `interval`, and `idle_timeout` options
- Improved CLI reference: clearer flag descriptions, date filter inclusivity notes, sync subcommand clarifications
- Improved CLI reference descriptions for `list`, `analytics`, `config`, `sync`, `view`, `credentials`, `migrate`, `reminder`, `notification`, `tags`, `tui`, `completion`, and `version` commands
- Added `yearly` and `every N months` recurrence patterns to task management how-to guide

### Added
- File watcher for real-time sync triggers with debouncing and smart timing (#41)
  - `internal/watcher` package using `fsnotify` for file system monitoring
  - Config fields: `sync.daemon.file_watcher`, `sync.daemon.smart_timing`, `sync.daemon.debounce_ms`
  - Defers sync during active editing sessions (quiet period detection)
- `regex` filter operator for views (e.g., `summary regex ^Project`)
- JSON output support for `list info`, `list trash`, `sync status`, `reminder list`, `reminder status`
- Documented `list info`, `list stats`, and `list vacuum` commands in CLI reference
- How-to guide for analytics (viewing stats, backend performance, errors, configuration)
- How-to guide for credential management (keyring, environment variables, rotation)
- How-to guide for migration between backends (migrate commands, supported backends, safe migration steps)

### Fixed
- Fixed broken relative link in backend testing setup doc (pointed to `backends.md` instead of `../explanation/backends.md`)
- Fixed alignment of `idle_timeout` key in config map literals (removed extra space)

### Changed
- Documented `--target-dir` flag for `completion install` in CLI reference
- Expanded configuration reference with more `config get`/`config set` examples for sync, daemon, reminder, and background pull cooldown settings
- Clarified credential resolution order: environment variables are used as fallback when keyring has no entry
- Expanded sync daemon documentation with architecture details, IPC behavior, and state file locations
- Updated notification configuration docs to reflect actual behavior (reminder-controlled, not separate config block)
- Clarified trash retention default value (`30` days) in configuration reference
- Removed outdated `ui: cli` config entries from getting-started tutorial examples
- Removed incorrect `interval: 5m` from sync config example in error reference

### Added
- IPC notify support in daemon sync loop for immediate sync triggering via Unix socket
- Documented daemon configuration options (`sync.daemon.enabled`, `interval`, `idle_timeout`)
- Documented `sync status` command and `--version` global flag in CLI reference

### Fixed
- Daemon test cleanup now uses `defer` to ensure daemon is stopped even on test failure
- Daemon sync loop refactored to extract `daemonPerformSync` and handle IPC notify signals

### Security
- PowerShell notification escaping now covers `$` to prevent subexpression injection (#44)

### Changed
- JSON export now includes list metadata (list_name) in output structure
- JSON import supports both new format (with list_name) and legacy format (array of tasks)

### Fixed
- JSON reimport now generates new UUIDs to avoid conflicts with soft-deleted tasks (#43)
- List restore now invalidates cache, ensuring restored list appears in `list` output (#42)

### Added
- Regression test for issue #43: reimport tasks after deleting list with soft-deleted UIDs
- Background sync daemon with forked process architecture for async sync operations (#36, #39)
  - `todoat sync daemon start/stop/status/kill` commands
  - Daemon runs as separate process, CLI returns immediately after local operations
  - IPC communication via Unix domain socket
  - Configurable idle timeout and auto-shutdown
  - Process isolation with proper signal handling (SIGTERM/SIGINT)
- Shorthand reminder interval formats: `1d`, `1h`, `15m`, `1w` (in addition to full word formats)
- `internal/daemon` package for background daemon process management

### Changed
- Updated auto-sync daemon documentation status (feature now stable)
- Enhanced CLI help text with task action documentation and usage examples
- Sync operations can now be delegated to background daemon when enabled

### Fixed
- CreateTask now preserves ParentID in VTODO sent to server (#37)
- Config parsing errors now produce a warning instead of silently using defaults (#38)
- Reminder configuration now properly loads from config.yaml instead of only using defaults (#034)

### Added
- Nextcloud/CalDAV backend now parses and generates RELATED-TO property for parent-child task relationships (subtask support) (#29)
- `--parent` flag now accepts task UID directly when multiple tasks have the same name (#28)
- Background sync now completes before program exit, ensuring auto-sync operations fully sync to remote (#032)
- Custom-named backends are now recognized when defined at config top-level for backwards compatibility (#031)
- `config set analytics.enabled` and `config set analytics.retention_days` now work correctly (#078)
- Documentation command syntax examples now use correct list-based format (#001)

### Changed
- Default value for `analytics.enabled` changed from `false` to `true`
- Documentation improvements for configuration and synchronization
- Clarified that default view excludes completed tasks (use `-v all` to see them)
- Simplified feature documentation tables by removing redundant status column (all features are stable)
- Updated getting-started tutorial with `completion install` quick setup

### Added
- `completion install` command to auto-detect shell and install completion scripts (#035)
- `completion uninstall` command to remove installed completion scripts
- Tests for analytics configuration settings
- Issue tracking for config key and value validation bugs
- Integration tests for credential manager keyring flow (#010)
- Documentation for background pull sync on read operations
- Configurable `sync.background_pull_cooldown` option to control cooldown between background pull syncs (default: 30s, minimum: 5s) (#082)
- Reminder configuration documentation in reference guide (#083)
- Reminder configuration example in sample config (#083)
- Acceptance criteria tests for task reminders (#083)
