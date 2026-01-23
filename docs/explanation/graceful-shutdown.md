# Graceful Shutdown

## Overview

The graceful shutdown system ensures that todoat exits cleanly when interrupted, preventing data loss and ensuring operations complete safely. When you press Ctrl+C or send a termination signal, todoat coordinates an orderly shutdown of all running components.

**Key Capabilities:**
- Clean handling of SIGINT (Ctrl+C) and SIGTERM signals
- Safe completion of in-progress sync and write operations
- Ordered cleanup of resources (LIFO - Last In, First Out)
- Timeout protection to prevent indefinite hangs
- Prevention of new operations during shutdown

---

## Table of Contents

- [Signal Handling](#signal-handling)
- [What Happens During Shutdown](#what-happens-during-shutdown)
- [Partial Operations](#partial-operations)
- [Timeout Behavior](#timeout-behavior)
- [Exit Codes](#exit-codes)
- [Technical Details](#technical-details)

---

## Signal Handling

### Supported Signals

**SIGINT (Ctrl+C):**
- Triggered when you press Ctrl+C in the terminal
- Initiates graceful shutdown sequence
- All registered cleanup functions are called

**SIGTERM:**
- Standard Unix termination signal
- Behaves identically to SIGINT
- Used by process managers, containers, and system shutdown

Both signals trigger the same graceful shutdown process - there's no "force quit" vs "soft quit" distinction.

### What Happens When You Interrupt

When you press Ctrl+C during an operation:

1. **Shutdown flag is set** - No new operations can start
2. **Context is cancelled** - Running operations receive a stop signal
3. **Cleanup functions run** - In reverse order of registration (LIFO)
4. **Application exits** - With appropriate exit code

---

## What Happens During Shutdown

### Sync Operations

If you press Ctrl+C during a sync operation (`todoat sync`):

**Short operations (< timeout):**
- The current sync operation completes normally
- Changes are not left in an inconsistent state
- Database transactions commit or roll back atomically

**Long operations (> timeout):**
- Operations receive a cancellation signal
- Uncommitted changes are rolled back
- Local cache remains consistent (may be behind remote)
- You can run `todoat sync` again later to complete

### Database Writes

If you interrupt during a database write (add, update, complete, delete):

- **Transaction safety**: All database operations use transactions
- **Atomic commits**: Either the full change commits or nothing does
- **No partial writes**: You won't have half-written tasks

### Cleanup Order (LIFO)

Cleanup functions run in Last-In-First-Out order:

```
Registration order:     Cleanup order:
1. Database connection  3. Database connection (last)
2. Sync manager         2. Sync manager
3. TUI components       1. TUI components (first)
```

This ensures dependent resources are cleaned up in the correct order - for example, the TUI closes before the database connection, preventing errors.

---

## Partial Operations

### Sync Queue Safety

If shutdown occurs during sync queue processing:

- **Completed operations**: Removed from queue
- **In-progress operation**: Rolled back, stays in queue
- **Pending operations**: Remain in queue for next sync

On next `todoat sync`, processing resumes from where it stopped.

### Batch Operations

For operations that process multiple items:

- Items processed before interrupt: Committed
- Current item during interrupt: Rolled back
- Remaining items: Not processed

---

## Timeout Behavior

### Default Behavior

The shutdown manager uses a context-based timeout:

- Cleanup functions receive a context that can be cancelled
- If cleanup takes too long, the timeout fires
- Timed-out cleanups are abandoned to prevent hanging

### Long-Running Cleanup

If a cleanup function exceeds the timeout:

1. The cleanup function's context is cancelled
2. The function should check `ctx.Done()` and return
3. Remaining cleanups continue (not blocked by one slow cleanup)
4. Application exits after all cleanups finish or time out

### Recommended Practice

Long-running operations should check for cancellation:

```go
func myCleanup(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            // Timeout reached, exit gracefully
            return ctx.Err()
        default:
            // Do cleanup work
            if done {
                return nil
            }
        }
    }
}
```

---

## Exit Codes

### Clean Shutdown (Exit Code 0)

Returned when:
- All cleanup functions completed successfully
- No errors during shutdown
- Application terminated gracefully

### Timeout or Error (Non-zero Exit Code)

Returned when:
- Cleanup timed out waiting for operations
- An error occurred during cleanup
- Forced termination was required

### Checking Exit Status

```bash
todoat sync
echo $?  # Shows exit code (0 = success)
```

---

## Technical Details

### Shutdown Manager

The shutdown system uses a centralized manager (`internal/shutdown/shutdown.go`):

```go
type Manager struct {
    cleanups   []cleanupEntry  // Registered cleanup functions
    shutdown   bool            // Shutdown initiated flag
    ctx        context.Context // Context cancelled on shutdown
    cancel     context.CancelFunc
    once       sync.Once       // Ensures single shutdown
}
```

### Concurrent Safety

The shutdown manager is safe for concurrent use:

- **Multiple signals**: Only the first shutdown call takes effect
- **Concurrent cleanup**: Cleanup functions run sequentially (not parallel)
- **Thread-safe checks**: `IsShutdown()` can be called from any goroutine

### Preventing New Operations

After shutdown is initiated:

```go
// Operations should check before starting
if mgr.IsShutdown() {
    return ErrShuttingDown
}

// Or use the context
select {
case <-mgr.Context().Done():
    return ctx.Err()
default:
    // Safe to proceed
}
```

---

## Related Documentation

- [Synchronization System](synchronization.md) - Sync operations that use graceful shutdown
- [Task Management](task-management.md) - Database operations protected by shutdown
- [Configuration](configuration.md) - Application settings

---

## Source Code

- `internal/shutdown/shutdown.go` - Shutdown manager implementation
- `internal/shutdown/shutdown_test.go` - Comprehensive tests covering:
  - SIGINT/SIGTERM handling
  - Shutdown during sync
  - Shutdown during writes
  - Timeout behavior
  - Cleanup ordering
  - Concurrent safety
  - Exit codes
