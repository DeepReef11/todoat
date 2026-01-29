# Project TODO

## Documentation Tasks

_No documentation tasks pending._

---

## Questions for Team

### [ARCH-001] Should the daemon Unix socket have restricted file permissions?

**Context**: The daemon creates a Unix domain socket for IPC (`internal/daemon/daemon.go`). Currently the socket inherits the process umask, meaning other users on the system may be able to connect to the daemon and issue commands (status queries, sync triggers). This was introduced in commit `a0df401` (background sync daemon).

**Options**:
- [x] Restrict to owner only (0600) - Prevents other users from interacting with daemon

**Impact**: Security posture for multi-user systems. Single-user desktops are unaffected.

**Asked**: 2026-01-29
**Status**: answered  <!-- User changes to "answered" or removes "un" when done -->

### [ARCH-002] Should conflict resolution propagate to remote on next sync?

**Context**: Sync conflict resolution is explicitly a local-only operation (`backend/sqlite/cli_test.go`). When a conflict is resolved locally, the remote backend is not updated. This means the next sync cycle may re-detect the same conflict if the remote still holds the conflicting version.

**Options**:
- [ ] Push resolution to remote on next sync - Prevents conflict reappearance, but adds complexity
- [ ] Keep local-only resolution - Simpler, but users may see resolved conflicts reappear
- [x] Queue a remote update automatically after resolution - Middle ground, uses existing sync queue

**Impact**: Core sync behavior. Affects user trust in conflict resolution workflow.

**Asked**: 2026-01-29
**Status**: answered  <!-- User changes to "answered" or removes "un" when done -->

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

### [UX-009] How does the user interact with interactive prompt features beyond -y/--no-prompt?

**Context**: Feature is documented in docs/explanation/interactive-ux.md but several described capabilities don't exist in code. The explanation doc describes `ui.interactive_prompt_for_all_tasks` config option, `--all` flag for controlling prompt task filtering, and `promptui` library integration. However, the actual Config struct has no `interactive_prompt_for_all_tasks` field, the `--all` flag is not registered in the CLI, and `promptui` is not imported anywhere in the Go code. The actual interactive mode uses `bufio.Scanner` for user input.

**Current documentation says**: "Use `--all` flag or set `ui.interactive_prompt_for_all_tasks: true` in config to include COMPLETED/CANCELLED tasks in all interactive prompts."

**Missing details**:
- [x] CLI flags/commands (`--all` for prompts not implemented)
- [x] Config file options (`ui.interactive_prompt_for_all_tasks` not in Config struct)
- [ ] UI elements (basic prompts work via bufio.Scanner)

**Options**:
- [ ] Implement described features - Add `ui.interactive_prompt_for_all_tasks` and `--all` for prompts
- [ ] Update explanation doc to match code - Remove references to unimplemented features
- [ ] Not user-facing - Current interactive behavior is sufficient as-is

**Impact**: Blocks user-facing documentation for interactive mode features. Current basic interactive mode is already documented in task-management how-to.

**Asked**: 2026-01-29
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

