# [084] Fix: config get/set missing sync.daemon.stuck_timeout and sync.daemon.task_timeout

## Summary

`sync.daemon.stuck_timeout` and `sync.daemon.task_timeout` are documented in `docs/reference/configuration.md` as configurable daemon options, and they exist in the `DaemonConfig` struct (`internal/config/config.go:88-89`), but they are **not registered** in the `config set` / `config get` command registry in `cmd/todoat/cmd/todoat.go`. Running the documented commands fails with "unknown config key".

## Documentation Reference

- Primary: `docs/reference/configuration.md`
- Section: "Daemon Configuration" and "Common Configuration Options" table
- Also referenced: `docs/how-to/sync.md` (Daemon Options table)

## Gap Type

wrong-syntax (documented config commands fail)

## Documented Command/Syntax

```bash
# From docs/reference/configuration.md
todoat config set sync.daemon.stuck_timeout 10
todoat config set sync.daemon.task_timeout "5m"
todoat config get sync.daemon.stuck_timeout
todoat config get sync.daemon.task_timeout
```

## Actual Result When Running Documented Command

```bash
$ todoat config set sync.daemon.stuck_timeout 10
{"error":"unknown config key: sync.daemon.stuck_timeout","code":1,"result":"ERROR"}

$ todoat config set sync.daemon.task_timeout "5m"
{"error":"unknown config key: sync.daemon.task_timeout","code":1,"result":"ERROR"}

$ todoat config get sync.daemon.stuck_timeout
{"error":"unknown config key: sync.daemon.stuck_timeout","code":1,"result":"ERROR"}

$ todoat config get sync.daemon.task_timeout
{"error":"unknown config key: sync.daemon.task_timeout","code":1,"result":"ERROR"}
```

## Working Alternative (if any)

Users can manually edit `~/.config/todoat/config.yaml` to add these values:

```yaml
sync:
  daemon:
    stuck_timeout: 10
    task_timeout: "5m"
```

The YAML fields are properly defined in the config struct and consumed by the daemon code.

The `stuck_timeout` can also be set via the CLI flag `todoat sync daemon start --stuck-timeout 10` at daemon start time, but this doesn't persist to config.

## Recommended Fix

FIX CODE - Add `sync.daemon.stuck_timeout` (int, minutes, non-negative) and `sync.daemon.task_timeout` (string, duration) to the `setConfigValue()` and `getConfigValue()` functions in `cmd/todoat/cmd/todoat.go`, following the pattern of existing daemon config keys like `sync.daemon.heartbeat_interval` and `sync.daemon.idle_timeout`.

## Dependencies

- Requires: none (existing config struct already has the fields)

## Complexity

S

## Acceptance Criteria

### Tests Required

- [ ] Test `todoat config set sync.daemon.stuck_timeout 10` succeeds
- [ ] Test `todoat config set sync.daemon.task_timeout "5m"` succeeds
- [ ] Test `todoat config get sync.daemon.stuck_timeout` returns correct value
- [ ] Test `todoat config get sync.daemon.task_timeout` returns correct value
- [ ] Test validation rejects invalid values (negative stuck_timeout, unparseable duration for task_timeout)

### Functional Requirements

- [ ] `config set sync.daemon.stuck_timeout <int>` sets the value and saves to config file
- [ ] `config set sync.daemon.task_timeout <duration>` sets the value and saves to config file
- [ ] `config get sync.daemon.stuck_timeout` displays the current value
- [ ] `config get sync.daemon.task_timeout` displays the current value
- [ ] `config get sync.daemon` includes both keys in its output (already shows them if set in YAML)

## Implementation Notes

The `DaemonConfig` struct already has the fields:
```go
StuckTimeout      int    `yaml:"stuck_timeout"`      // line 88
TaskTimeout       string `yaml:"task_timeout"`       // line 89
```

The `setConfigValue()` function in `cmd/todoat/cmd/todoat.go` has a switch on daemon sub-keys. Add two cases:
- `"stuck_timeout"`: validate as non-negative integer, set `cfg.Sync.Daemon.StuckTimeout`
- `"task_timeout"`: validate as parseable Go duration string, set `cfg.Sync.Daemon.TaskTimeout`

Similarly update `getConfigValue()` to handle these keys.

Follow the existing patterns for `heartbeat_interval` (int) and `background_pull_cooldown` (duration string).
