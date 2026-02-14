# [032] Sync Health Dashboard

## Summary
Add a `todoat sync health` command that surfaces internal daemon diagnostics — circuit breaker states, per-backend error rates, sync queue depth breakdown, and daemon uptime — giving users visibility into sync system health beyond the basic `sync status` output.

## Source
Code analysis of daemon internals (`internal/daemon/daemon.go`, `internal/daemon/circuitbreaker.go`) and the existing `sync status` command (`cmd/todoat/cmd/todoat.go:7461`). The daemon tracks rich health data internally (circuit breaker states, per-backend sync states, error counts, heartbeat timestamps) but `sync status` only exposes last sync time, pending count, and connection status. Users diagnosing sync problems have no visibility into why syncs are failing.

## Motivation
When sync stops working, users have limited diagnostic tools. The current `sync status` shows "Last Sync: Never" or a timestamp and pending count, but doesn't explain:
- Why a backend stopped syncing (circuit breaker tripped? network error? auth expired?)
- How many operations are queued per backend and per operation type
- Whether the daemon is actually running and healthy
- Historical error patterns (is it intermittent or persistent?)

This forces users to check logs manually, restart the daemon blindly, or file bug reports without diagnostic context. The internal daemon already tracks this data via `BackendState`, `CircuitBreaker`, and heartbeat mechanisms — it just needs a user-facing surface.

## Current Behavior
```
$ todoat sync status
Sync Status:

Offline Mode: auto

Backend: nextcloud
  Last Sync: 2026-02-14 10:30:00
  Pending Operations: 3
  Status: Configured
```

No visibility into circuit breaker state, error history, queue composition, or daemon process health.

## Proposed Behavior
```
$ todoat sync health
Sync Health Dashboard:

Daemon: running (PID 12345, uptime 2h 14m)
  Heartbeat: 5s ago
  Next tick: in 3s

Backend: nextcloud
  Circuit Breaker: closed (healthy)
  Last Sync: 2m ago (success)
  Error Rate: 0/50 (last hour)
  Queue: 2 creates, 1 update (3 total)

Backend: todoist
  Circuit Breaker: open (tripped 5m ago, cooldown 2m remaining)
  Last Sync: 8m ago (failed: 401 Unauthorized)
  Error Rate: 5/12 (last hour)
  Queue: 0 pending

Overall: 1/2 backends healthy, 3 operations pending
```

With `--json` flag for programmatic consumption. With `--verbose` for extended error history.

## Estimated Value
medium - Directly addresses a real diagnostic gap. Users with multi-backend setups or intermittent connectivity especially benefit. Reduces support burden by enabling self-diagnosis.

## Estimated Effort
S - Most data is already tracked internally. The work is primarily creating a new subcommand that queries the daemon via the existing Unix socket IPC (`Message{Type: "status"}` response already includes backend states) and formatting the output.

## Related
- Daemon implementation: `internal/daemon/daemon.go` (BackendState, GetAllBackendStates)
- Circuit breaker: `internal/daemon/circuitbreaker.go` (State, FailureCount)
- Existing sync status: `cmd/todoat/cmd/todoat.go:7461` (doSyncStatus)
- GitHub Issue #114: Circuit Breaker Pattern for Backend-Specific Daemon Errors
- GitHub Issue #115: Daemon Notification Integration
- [023] Configuration Validation Command (validates config; this validates runtime sync health)

## Status
unreviewed
