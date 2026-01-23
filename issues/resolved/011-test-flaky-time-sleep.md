# [011] Test Quality: Multiple Files - Excessive time.Sleep Usage

## Type
test-quality

## Severity
medium

## Test Locations

### High Sleep Duration
| File | Sleep Calls | Total Duration |
|------|-------------|----------------|
| backend/sync/daemon_test.go | 6 | ~9000ms |
| internal/views/builder_test.go | 43 | ~3000ms |
| internal/tui/tui_test.go | 29 | ~2150ms |
| backend/todoist/todoist_test.go | 1 | ~100ms |

## Problem
Tests use hardcoded `time.Sleep()` calls to wait for asynchronous operations. This pattern is flaky because:
1. Sleep durations may be too short on slow CI machines
2. Sleep durations may be too long, wasting time on fast machines
3. Tests become timing-dependent and can fail intermittently

### Examples

**backend/sync/daemon_test.go:99**
```go
// Wait for at least 2 intervals
time.Sleep(2500 * time.Millisecond)
```

**internal/views/builder_test.go:41**
```go
// Wait for initial render
time.Sleep(100 * time.Millisecond)
```

**internal/tui/tui_test.go:110**
```go
// Wait for initial render
time.Sleep(100 * time.Millisecond)
```

## Suggested Fix

### For TUI tests (teatest framework)
Use `teatest.WaitFor` or assertion helpers that poll for conditions:
```go
// Instead of:
time.Sleep(100 * time.Millisecond)
tm.Send(tea.KeyMsg{Type: tea.KeySpace})

// Use:
teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
    return bytes.Contains(out, []byte("expected"))
}, teatest.WithDuration(time.Second))
```

### For daemon tests
Use channels or condition variables to signal completion:
```go
// Instead of:
time.Sleep(2500 * time.Millisecond)

// Use:
select {
case <-syncComplete:
case <-time.After(5 * time.Second):
    t.Fatal("timeout waiting for sync")
}
```

### For cache/timing tests
Use test doubles with controllable time:
```go
clock := &fakeClock{}
cache := NewCacheWithClock(clock)
clock.Advance(10 * time.Millisecond)
```

## Impact
- Tests may fail randomly in CI environments
- Test suite takes longer than necessary
- Difficult to debug failures that only occur sometimes

## Resolution

**Fixed in**: this session
**Fix description**: Replaced flaky time.Sleep calls with condition-based polling and minimal message processing waits.

### Changes Made

1. **backend/sync/daemon_test.go** (6 sleeps replaced, ~9000ms → ~600ms):
   - Added `WaitFor`, `WaitForWithInterval`, `WaitForOutput`, and `WaitForSyncCount` helpers to `internal/testutil/cli.go`
   - Replaced `time.Sleep(2500ms)` with `cli.WaitForSyncCount(5*time.Second, 2)` polling
   - Replaced `time.Sleep(1500ms)` calls with condition-based polling
   - Changed daemon interval from 1s to 100ms for faster tests

2. **internal/views/builder_test.go** (43 sleeps reduced):
   - Added `waitForRender`, `waitForContent`, `waitForRenderAndRead` helpers using `teatest.WaitFor`
   - Added `sendKeyAndWait` helper with 20ms minimal sleep for message queue processing
   - Replaced sequential 50-100ms sleeps with helper calls
   - Kept minimal 100ms initial render waits where reading output immediately is needed

3. **internal/tui/tui_test.go** (29 sleeps reduced):
   - Added `sendKeyAndWait` and `sendRunesAndWait` helpers with 20ms minimal sleep
   - Replaced sequential 50ms sleeps between key presses with helper calls
   - Kept 100ms initial render waits (necessary for TUI initialization)

4. **backend/todoist/todoist_test.go** (1 sleep, unchanged):
   - The single 100ms sleep is intentional timing simulation for rate limiting test
   - Left unchanged as it's testing time-based retry behavior

### Verification Log
```bash
$ go test -count=1 ./backend/sync/... ./internal/views/... ./internal/tui/... ./backend/todoist/...
ok      todoat/backend/sync     2.171s
ok      todoat/internal/views   6.988s
ok      todoat/internal/tui     1.692s
ok      todoat/backend/todoist  0.276s
```
**Matches expected behavior**: YES - All tests pass with improved timing
