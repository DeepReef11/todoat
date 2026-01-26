# [010] Test Quality: TUI Tests - Race Condition

## Type
test-quality

## Severity
high

## Test Location
- File: internal/tui/tui_test.go
- Functions: TestTUICompleteTask (and likely others using mockBackend)
- Lines: 69-88

## Problem
The mockBackend struct in TUI tests has no synchronization, causing race conditions when tests run concurrently. The race detector found data races in:
- `mockBackend.UpdateTask()` at line 72
- `mockBackend.DeleteTask()` at line 80-84
- `mockBackend.CreateTask()` at line 65

Multiple goroutines can read/write to `m.tasks` map simultaneously without proper locking.

## Current Code
```go
func (m *mockBackend) UpdateTask(_ context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	for i, t := range m.tasks[listID] {
		if t.ID == task.ID {
			m.tasks[listID][i] = *task  // Race condition here
			return task, nil
		}
	}
	return nil, nil
}
```

## Suggested Fix
Add a mutex to the mockBackend struct and use it in all methods that access the tasks map:

```go
type mockBackend struct {
	mu    sync.Mutex
	lists []backend.List
	tasks map[string][]backend.Task
}

func (m *mockBackend) UpdateTask(_ context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// ... existing code
}
```

## Impact
- Tests fail intermittently with `-race` flag
- Race conditions can mask real bugs
- CI may fail sporadically on different machines

## Resolution

**Fixed in**: this session
**Fix description**: Added sync.Mutex to mockBackend struct and protected all methods (GetLists, GetTasks, GetTask, CreateTask, UpdateTask, DeleteTask) with mutex lock/unlock. GetTasks now returns a copy of the slice to prevent races when the caller iterates over returned tasks.

### Verification Log
```bash
$ go test -race -v -count=3 ./internal/tui/...
=== RUN   TestTUILaunch
--- PASS: TestTUILaunch (0.11s)
=== RUN   TestTUIListNavigation
--- PASS: TestTUIListNavigation (0.16s)
=== RUN   TestTUITaskNavigation
--- PASS: TestTUITaskNavigation (0.21s)
=== RUN   TestTUIAddTask
--- PASS: TestTUIAddTask (0.29s)
=== RUN   TestTUIEditTask
--- PASS: TestTUIEditTask (0.26s)
=== RUN   TestTUICompleteTask
--- PASS: TestTUICompleteTask (0.26s)
=== RUN   TestTUIDeleteTask
--- PASS: TestTUIDeleteTask (0.31s)
=== RUN   TestTUITreeView
--- PASS: TestTUITreeView (0.10s)
=== RUN   TestTUIFilterTasks
--- PASS: TestTUIFilterTasks (0.27s)
=== RUN   TestTUIKeyBindings
--- PASS: TestTUIKeyBindings (0.21s)
=== RUN   TestTUIQuit
--- PASS: TestTUIQuit (0.11s)
[...repeated 3x, all passed...]
PASS
ok  	todoat/internal/tui	7.832s
```
**Matches expected behavior**: YES - No race conditions detected across 3 runs
