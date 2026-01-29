# internal/utils Package Documentation

The `internal/utils` package provides common utility functions used throughout the gosynctasks application, including logging, output formatting, error handling, input validation, and user interaction.

## Overview

| File | Purpose |
|------|---------|
| `logger.go` | Leveled logging with verbose mode and background logging |
| `output.go` | JSON/YAML marshaling and output |
| `errors.go` | Error types with user-friendly suggestions |
| `validation.go` | Input validation (priority, dates) |
| `inputs.go` | User input handling and interactive prompts |

---

## logger.go - Logging System

### Logger (Main Application Logger)

A singleton logger with verbose mode support for controlling debug output.

**Type:**
```go
type Logger struct {
    verbose bool
    mu      sync.RWMutex
}
```

**Usage:**
```go
// Get the global logger instance
logger := utils.GetLogger()

// Enable verbose mode
logger.SetVerbose(true)

// Log at different levels
logger.Debug("Debug message: %s", value)  // Only shown when verbose=true
logger.Info("Info message: %s", value)
logger.Warn("Warning: %s", value)
logger.Error("Error: %s", value)

// Convenience functions (use global logger)
utils.Debugf("Debug: %v", data)
utils.Infof("Info: %v", data)
utils.Warnf("Warning: %v", data)
utils.Errorf("Error: %v", data)

// Set verbose mode globally (also configures log flags)
utils.SetVerboseMode(true)
```

**Log Levels:**
| Level | Method | Prefix | Shown When |
|-------|--------|--------|------------|
| Debug | `Debug()` | `[DEBUG]` | verbose=true only |
| Info | `Info()` | `[INFO]` | Always |
| Warn | `Warn()` | `[WARN]` | Always |
| Error | `Error()` | `[ERROR]` | Always |

**Timestamp in Verbose Mode:**

When verbose mode is enabled, all log lines are prefixed with the current local time in `HH:MM:SS` format. This helps correlate events and measure timing between operations (e.g., identifying sync latency).

```
14:48:39 [DEBUG] Verbose mode enabled
14:48:39 [DEBUG] Sync enabled with default_backend: nextcloud-test
14:48:39 [DEBUG] Background pull sync triggered
14:48:41 [DEBUG] Using custom backend 'nextcloud-test' of type 'nextcloud'
14:48:41 [DEBUG] Background auto-sync triggered
```

Non-verbose output (Info, Warn, Error) is not affected by this timestamp prefix — only `[DEBUG]` lines include the time.

**Operation Logging:**
```go
// Log operation start/end with automatic error handling
err := utils.LogOperation("fetching tasks", func() error {
    return backend.GetTasks()
})

// With formatted message
err := utils.LogOperationf("syncing list %s", func() error {
    return syncList(listName)
}, listName)
```

### BackgroundLogger (Background Process Logger)

Specialized logger for background sync processes that writes to a PID-specific temp file.

**Type:**
```go
type BackgroundLogger struct {
    logger   *log.Logger
    logFile  *os.File
    enabled  bool
    filePath string
}
```

**Configuration:**
```go
// Toggle background logging (compile-time constant)
const ENABLE_BACKGROUND_LOGGING = true
```

**Log File Location:** `/tmp/gosynctasks-_internal_background_sync-{PID}.log`

**Usage:**
```go
// Create background logger
bl, err := utils.NewBackgroundLogger()
if err != nil {
    // Logging disabled, but bl is still usable (writes to io.Discard)
}
defer bl.Close()

// Log messages
bl.Printf("Sync started at %v", time.Now())
bl.Print("Processing...")
bl.Println("Done")

// Check status
if bl.IsEnabled() {
    fmt.Printf("Logs at: %s\n", bl.GetLogPath())
}
```

**Methods:**
| Method | Description |
|--------|-------------|
| `NewBackgroundLogger()` | Create new logger instance |
| `Printf(format, args...)` | Log formatted message |
| `Print(args...)` | Log message |
| `Println(args...)` | Log message with newline |
| `Close()` | Close log file |
| `GetLogPath()` | Return log file path |
| `IsEnabled()` | Check if logging is enabled |

---

## output.go - Output Formatting

Functions for marshaling and outputting data in JSON/YAML formats.

**Functions:**
```go
// Output to stdout
utils.OutputJSON(data)  // Prints indented JSON
utils.OutputYAML(data)  // Prints YAML

// Get bytes (for further processing)
jsonBytes, err := utils.MarshalJSON(data)
yamlBytes, err := utils.MarshalYAML(data)
```

**Example:**
```go
tasks := []Task{{Summary: "Task 1"}, {Summary: "Task 2"}}

// Output as JSON
utils.OutputJSON(tasks)
// Output:
// [
//   {
//     "summary": "Task 1"
//   },
//   {
//     "summary": "Task 2"
//   }
// ]
```

---

## errors.go - Error Handling

Error types that include user-friendly suggestions for common issues.

### ErrorWithSuggestion Type

```go
type ErrorWithSuggestion struct {
    Err        error
    Suggestion string
}
```

**Output format:**
```
<error message>

Suggestion: <helpful suggestion>
```

### Pre-built Error Constructors

| Function | Use Case |
|----------|----------|
| `ErrTaskNotFound(searchTerm)` | Task search returned no results |
| `ErrListNotFound(listName)` | List doesn't exist |
| `ErrNoListsAvailable()` | No lists configured |
| `ErrSyncNotEnabled()` | Sync operation when sync is disabled |
| `ErrBackendNotConfigured(name)` | Backend not in config |
| `ErrBackendOffline(name, reason)` | Connection failed (smart suggestions) |
| `ErrInvalidPriority(priority)` | Priority outside 0-9 range |
| `ErrInvalidDate(dateStr)` | Date parsing failed |
| `ErrInvalidStatus(status, valid)` | Unknown status value |
| `ErrCredentialsNotFound(backend, user)` | Keyring/env credentials missing |
| `ErrAuthenticationFailed(backend)` | Auth rejected by server |
| `ErrConfigFileNotFound(path)` | Config file missing |
| `ErrInvalidConfig(field, reason)` | Config validation failed |

**Usage:**
```go
// Return pre-built error
if task == nil {
    return utils.ErrTaskNotFound(searchTerm)
}

// Wrap existing error with suggestion
if err != nil {
    return utils.WrapWithSuggestion(err, "Try running 'gosynctasks sync' first")
}
```

### Smart Suggestions

`ErrBackendOffline` provides context-aware suggestions:
- DNS errors: "Check your DNS settings and internet connection"
- Connection refused: "Check if the server is running and accessible"
- Timeout: "The server may be slow or unreachable. Try again later"
- Default: "Check your internet connection and try again"

---

## validation.go - Input Validation

### Priority Validation

```go
// Returns nil if valid, ErrInvalidPriority otherwise
err := utils.ValidatePriority(5)  // Valid: 0-9

err := utils.ValidatePriority(10) // Error: invalid priority 10
                                  // Suggestion: Priority must be between 0 and 9
```

### Date Parsing and Validation

```go
// Parse date string (YYYY-MM-DD format)
date, err := utils.ParseDateFlag("2026-01-15")  // Returns *time.Time
date, err := utils.ParseDateFlag("")            // Returns nil, nil (clear date)
date, err := utils.ParseDateFlag("invalid")    // Returns error

// Validate date range
err := utils.ValidateDates(startDate, dueDate)
// Returns error if startDate > dueDate
```

---

## inputs.go - User Input

Functions for reading user input and interactive prompts.

### Basic Input

```go
// Read integer from stdin
num, err := utils.ReadInt()

// Read string from stdin (trimmed)
str, err := utils.ReadString()
```

### Interactive Prompts

**Yes/No Prompt:**
```go
// Loops until valid y/n response
if utils.PromptYesNo("Delete this task?") {
    deleteTask()
}
```

**Selection Prompt:**
```go
// Display numbered list and get selection
tasks := []Task{{Summary: "Task A"}, {Summary: "Task B"}}

idx, err := utils.PromptSelection(tasks, "Select a task", func(i int, t Task) {
    fmt.Printf("%d. %s\n", i+1, t.Summary)
})
// Returns 0-based index, or error if cancelled (user enters 0)
```

**Confirmation Prompt:**
```go
confirmed, err := utils.PromptConfirmation("Are you sure?")
if confirmed {
    proceed()
}
```

---

## Best Practices

1. **Logging**: Use `utils.Debugf()` for development/debug info; use `utils.Infof()` sparingly for user-facing messages
2. **Errors**: Always use the pre-built error constructors when applicable for consistent UX
3. **Validation**: Validate user input early using `ValidatePriority()` and `ParseDateFlag()`
4. **Background Logging**: Only enable `ENABLE_BACKGROUND_LOGGING` during development/debugging

## Testing

Each file has corresponding tests:
- `logger_test.go`
- `output_test.go`
- `errors_test.go`
- `validation_test.go`
- `inputs_test.go`

Run tests: `go test ./internal/utils/...`
