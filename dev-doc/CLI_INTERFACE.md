# CLI Interface Features

This document details the command-line interface features of gosynctasks, including command structure, flags, interactive modes, shell completion, and terminal integration.

## Table of Contents

1. [Command Structure](#command-structure)
2. [Global Flags](#global-flags)
3. [Action Flags](#action-flags)
4. [Interactive List Selection](#interactive-list-selection)
5. [Shell Completion](#shell-completion)
6. [Terminal Width Detection](#terminal-width-detection)
7. [Argument Parsing and Validation](#argument-parsing-and-validation)
8. [Verbose Mode](#verbose-mode)
9. [Backend Selection Flags](#backend-selection-flags)
10. [Graceful Shutdown](#graceful-shutdown)
11. [No-Prompt Mode](#no-prompt-mode)
12. [JSON Output Mode](#json-output-mode)
13. [Result Codes](#result-codes)

---

## Command Structure

### Feature: Flexible Command Syntax

**Purpose:** Provides an intuitive, concise command syntax that supports both explicit and implicit operations, reducing typing while maintaining clarity.

**How It Works:**

1. **Basic Syntax:** `gosynctasks [list-name] [action] [task-summary]`
2. **Default Actions:** When action is omitted, defaults to `get` (show tasks)
3. **Action Abbreviations:** Short one-letter aliases for common actions:
   - `g` → `get`
   - `a` → `add`
   - `u` → `update`
   - `c` → `complete`
   - `d` → `delete`
4. **Argument Validation:** Maximum 3 positional arguments enforced by Cobra framework
5. **Subcommands:** Special operations use subcommands (`list`, `view`, `sync`, `credentials`)

**User Journey:**

```bash
# No arguments - interactive list selection
gosynctasks

# List name only - show tasks from that list
gosynctasks MyList

# Explicit action
gosynctasks MyList get

# Action abbreviation
gosynctasks MyList a "New task"

# Full command with task summary
gosynctasks MyList add "Complete the report"
```

**Prerequisites:**
- Application initialized with valid configuration (see [Configuration](./CONFIGURATION.md))
- At least one backend configured (see [Backend System](./BACKEND_SYSTEM.md))

**Outputs/Results:**
- Command executes appropriate action (CRUD operation, list display, etc.)
- Error messages for invalid syntax or missing arguments
- Help text when `--help` flag is used

**Technical Details:**
- Built on Cobra framework (`github.com/spf13/cobra`)
- Root command defined in `cmd/gosynctasks/main.go:26-114`
- Action resolution in `internal/app/app.go`
- Argument parsing handles both full names and abbreviations

**Related Features:**
- [Interactive List Selection](#interactive-list-selection) - Used when no list name provided
- [Task Management](./TASK_MANAGEMENT.md) - All CRUD operations
- [List Management](./LIST_MANAGEMENT.md) - List-specific commands

---

## Global Flags

### Feature: Persistent Configuration Flags

**Purpose:** Allow users to override default behavior, specify custom configurations, and control application-wide settings without modifying configuration files.

**How It Works:**

1. **Config Path Flag (`--config`):**
   - Overrides default XDG config location
   - Accepts file path or directory path
   - Special value `.` uses current directory: `./gosynctasks/config.yaml`
   - Set before application initialization via `config.SetCustomConfigPath()`

2. **Backend Override (`--backend`, `-b`):**
   - Forces use of specific backend config
   - Overrides auto-detection and config defaults
   - Value passed to `app.NewApp(backendName)`
   - See [Backend Selection](./BACKEND_SYSTEM.md#backend-selection)

3. **List Backends (`--list-backends`):**
   - Displays all configured backends with status
   - Shows enabled/disabled state
   - Exits immediately after display
   - See [Backend System](./BACKEND_SYSTEM.md)

4. **Detect Backend (`--detect-backend`):**
   - Shows auto-detected backends in current directory
   - Useful for Git backend auto-detection
   - Exits immediately after display
   - See [Backend Auto-Detection](./BACKEND_SYSTEM.md#auto-detection)

5. **Verbose Mode (`--verbose` or `-V`):**
   - Enables debug logging throughout application
   - See [Verbose Mode](#verbose-mode) below

6. **No-Prompt Mode (`--no-prompt` or `-y`):**
   - Disables all interactive prompts for scripting and automation
   - Skips confirmation dialogs (delete, etc.) - acts as force mode
   - Auto-confirms single task matches
   - Outputs structured data for ambiguous cases instead of prompting
   - See [No-Prompt Mode](#no-prompt-mode) below

7. **JSON Output (`--json`):**
   - Outputs results in machine-parseable JSON format
   - Pairs well with `--no-prompt` for scripting
   - See [JSON Output Mode](#json-output-mode) below

**User Journey:**

```bash
# Use custom config file
gosynctasks --config /path/to/config.yaml MyList

# Use config in current directory
gosynctasks --config . MyList

# Force specific backend
gosynctasks --backend nextcloud-prod WorkTasks

# List all backends
gosynctasks --list-backends

# Check auto-detection
gosynctasks --detect-backend

# Enable debug logging
gosynctasks --verbose MyList add "Debug this"
gosynctasks -V MyList add "Debug this"

# No-prompt mode for scripting
gosynctasks -y MyList delete "Task"           # Delete without confirmation
gosynctasks --no-prompt MyList complete "Task" # Complete without prompts

# JSON output for automation
gosynctasks --json MyList                      # List tasks as JSON
gosynctasks -y --json MyList add "Task"        # Add task, JSON output

# Combined scripting workflow
gosynctasks -y --json MyList update "review" -s DONE
```

**Prerequisites:**
- For `--config`: Valid config file at specified path
- For `--backend`: Backend name must exist in configuration
- For `--list-backends` and `--detect-backend`: Valid configuration loaded

**Outputs/Results:**
- `--config`: Application uses specified configuration
- `--backend`: All operations use specified backend
- `--list-backends`: Table of backends with status
- `--detect-backend`: List of auto-detected backends
- `--verbose`: Debug messages printed to stderr
- `--no-prompt`: All outputs include result codes (see [Result Codes](#result-codes))
- `--json`: All outputs in JSON format with structured data

**Technical Details:**
- Flags registered in `cmd/gosynctasks/main.go:118-122`
- Processed in `PersistentPreRunE` hook before command execution
- `--config` sets global state via `config.SetCustomConfigPath()`
- `--verbose` sets global state via `utils.SetVerboseMode(true)`
- Flags available to all commands and subcommands

**Related Features:**
- [Backend System](./BACKEND_SYSTEM.md) - Backend selection and auto-detection
- [Configuration](./CONFIGURATION.md) - Config file structure and locations
- [Verbose Mode](#verbose-mode) - Debug logging details

---

## Action Flags

### Feature: Operation-Specific Flags

**Purpose:** Provide fine-grained control over task operations, enabling users to set multiple properties in a single command without interactive prompts.

**How It Works:**

**Status Flags:**
1. **Filter/Set Status (`-s, --status`):**
   - For `get`: Filters tasks by one or more statuses (array)
   - For `update`: Sets task status (single value)
   - Accepts full names or abbreviations: TODO/T, DONE/D, PROCESSING/P, CANCELLED/C
   - Example: `-s TODO,PROCESSING` or `-s T,P`

**Task Property Flags:**
2. **Description (`-d, --description`):**
   - Sets or updates task description
   - Supports multi-line text
   - Example: `-d "Detailed notes about the task"`

3. **Priority (`-p, --priority`):**
   - Sets task priority (0-9)
   - 0 = undefined, 1 = highest, 9 = lowest
   - Example: `-p 1`

5. **Due Date (`--due-date`):**
   - Sets task due date in YYYY-MM-DD format
   - Empty string clears the date: `--due-date ""`
   - Example: `--due-date 2025-01-31`

6. **Start Date (`--start-date`):**
   - Sets task start date in YYYY-MM-DD format
   - Empty string clears the date: `--start-date ""`
   - Example: `--start-date 2025-01-15`

7. **Summary (`--summary`):**
   - For `update`: Renames the task
   - Example: `--summary "New task name"`

**Hierarchy Flags:**
8. **Parent (`-P, --parent`):**
   - Specifies parent task when adding subtasks
   - Supports full summary or path references: "Parent/Child"
   - See [Subtask Features](./SUBTASKS_HIERARCHY.md)
   - Example: `-P "Feature Development"`

9. **Literal (`-l, --literal`):**
   - Disables automatic path-based hierarchy creation
   - Treats "/" in task summary as literal character
   - Example: `-l "Design UI/UX mockups"` (creates one task, not hierarchy)

**View Flags:**
10. **View Mode (`-v, --view`):**
    - Selects custom view for displaying tasks
    - Default: "default" view
    - Built-in: "default", "all"
    - Custom: User-defined view names
    - See [Views Customization](./VIEWS_CUSTOMIZATION.md)
    - Example: `-v all` or `-v myview`

**Task Selection Flags:**
11. **UID Selection (`--uid`):**
    - Selects a task by its unique identifier instead of summary search
    - Bypasses task search and matching entirely
    - Required for unambiguous task operations in scripts
    - Useful after receiving `ACTION_INCOMPLETE` with multiple matches
    - Example: `--uid "550e8400-e29b-41d4-a716-446655440000"`
    - **Unsynced tasks**: For tasks not yet synced to remote (no backend-assigned UID), use `NOT-SYNCED-<sqlite-id>` format
    - Example: `--uid "NOT-SYNCED-123"` (uses SQLite internal ID)
    - Supported by: `update`, `complete`, `delete`

**User Journey:**

```bash
# Filter tasks by status
gosynctasks MyList -s TODO,PROCESSING
gosynctasks MyList -s T,D  # Using abbreviations

# Add task with multiple properties
gosynctasks MyList add "Complete report" -d "Q4 financial report" -p 1 -s TODO --due-date 2025-01-31

# Add task with dates
gosynctasks MyList add "Project" --start-date 2025-01-15 --due-date 2025-02-28

# Update task status
gosynctasks MyList update "task name" -s DONE

# Update task priority and description
gosynctasks MyList update "task" -p 5 -d "Updated notes"

# Rename task
gosynctasks MyList update "old name" --summary "new name"

# Add subtask
gosynctasks MyList add "Subtask" -P "Parent Task"

# Add task with literal slash (disable hierarchy)
gosynctasks MyList add -l "Design UI/UX mockups"

# Use custom view
gosynctasks MyList -v all
gosynctasks MyList -v myview

# UID-based task selection (for scripting)
gosynctasks MyList update --uid "550e8400-e29b-41d4-a716-446655440000" -s DONE
gosynctasks MyList delete --uid "550e8400-e29b-41d4-a716-446655440000"
gosynctasks MyList complete --uid "550e8400-e29b-41d4-a716-446655440000"

# Unsynced task (created locally, not yet synced to remote)
gosynctasks MyList update --uid "NOT-SYNCED-123" -s DONE  # Uses SQLite internal ID

# Workflow: search → get UID → operate
gosynctasks -y --json MyList update "ambiguous" -s DONE  # Returns matches with UIDs
gosynctasks -y MyList update --uid "returned-uid-here" -s DONE  # Use specific UID
```

**Prerequisites:**
- Task list must exist for list-specific operations
- For `update`: Task must exist (uses intelligent search, or `--uid` for direct selection)
- For `-P`: Parent task must exist
- For `--uid`: Valid UID must exist in the task list
- For custom views: View must be defined in `~/.config/gosynctasks/views/`

**Outputs/Results:**
- `get` with `-s`: Filtered task list
- `add` with flags: New task created with specified properties
- `update` with flags: Task updated with new values
- Invalid flag values produce error messages
- Date validation errors for malformed dates

**Technical Details:**
- Flags registered in `cmd/gosynctasks/main.go:`
- Status flag has shell completion for valid values 
- View flag has dynamic completion from available views (lines 145-154)
- Date parsing uses Go's time.Parse with `2006-01-02` format
- Status abbreviations translated via `StatusStringTranslateToStandardStatus()`
- Flag values extracted and passed to operation handlers in `internal/app/app.go`

**Related Features:**
- [Task Management](./TASK_MANAGEMENT.md) - CRUD operations using these flags
- [Subtasks Hierarchy](./SUBTASKS_HIERARCHY.md) - Parent and literal flags
- [Views Customization](./VIEWS_CUSTOMIZATION.md) - View flag and custom views
- [Shell Completion](#shell-completion) - Tab completion for flag values

---

## Interactive List Selection

### Feature: Visual Task List Picker

**Purpose:** Enables users to browse and select task lists visually when no list name is provided, eliminating the need to remember exact list names.

**How It Works:**

1. **Trigger Condition:** Command invoked with no arguments or only flags
2. **List Retrieval:** Fetches all task lists from active backend
3. **Task Count Calculation:** Queries each list for task count
4. **Terminal Width Detection:** Adapts display to terminal size (see [Terminal Width Detection](#terminal-width-detection))
5. **Formatted Display:**
   - Cyan bordered box with header "Available Task Lists"
   - Numbered list with bold white names
   - Task counts in gray (e.g., "(5 tasks)")
   - Optional descriptions shown indented below list name
   - Dynamic width adjustment (40-100 chars)
6. **User Prompt:** Waits for numeric selection or 'q' to quit
7. **Selection Handling:** Executes default action (get) on selected list

**User Journey:**

```bash
# User invokes without list name
$ gosynctasks

# System displays formatted list:
┌─ Available Task Lists ──────────────────────────────────┐
   1. Work                        (12 tasks)
      Professional tasks and projects
   2. Personal                    (5 tasks)
   3. Shopping                    (0 tasks)
└─────────────────────────────────────────────────────────┘

Select a list (1-3, or 'q' to quit): 1

# System shows tasks from "Work" list
```

**Prerequisites:**
- At least one task list exists in the backend
- Backend connection is active
- Terminal supports ANSI color codes (gracefully degrades if not)

**Outputs/Results:**
- Visual list picker with color-coded formatting
- Selected list's tasks displayed using default view
- "No task lists found" message if backend is empty
- Error message if backend connection fails

**Technical Details:**
- Display logic in `internal/cli/display.go:23-78`
- Terminal width detection using `golang.org/x/term` package
- ANSI color codes:
  - `\033[1;36m` - Bold cyan (borders)
  - `\033[1;37m` - Bold white (list names)
  - `\033[36m` - Cyan (numbers)
  - `\033[90m` - Gray (task counts, descriptions)
  - `\033[0m` - Reset
- Border calculation: `borderWidth = min(max(termWidth - 2, 40), 100)`
- Task count queries run in sequence (potential optimization opportunity)

**Related Features:**
- [List Management](./LIST_MANAGEMENT.md) - Creating and managing lists
- [Terminal Width Detection](#terminal-width-detection) - Dynamic sizing
- [Backend System](./BACKEND_SYSTEM.md) - Source of task lists

---

## Shell Completion

### Feature: Tab Completion for Commands and Arguments

**Purpose:** Speeds up command entry and reduces errors by providing context-aware suggestions for list names, actions, flags, and flag values.

**How It Works:**

**Completion Types:**

1. **List Name Completion (Position 1):**
   - Fetches all task lists from backend
   - Filters by prefix match (case-insensitive)
   - Returns matching list names
   - Example: `gosynctasks Wo<TAB>` → `gosynctasks Work`

2. **Action Completion (Position 2):**
   - After list name is entered
   - Suggests: `get`, `add`, `update`, `complete`, `delete`
   - Full names only (not abbreviations)
   - Example: `gosynctasks Work u<TAB>` → `gosynctasks Work update`

3. **Task Summary (Position 3):**
   - No completion offered (user enters free text)
   - Returns `NoFileComp` directive to prevent file path completion

4. **Status Flag Completion (`--status`, `--add-status`):**
   - Suggests: `TODO`, `DONE`, `PROCESSING`, `CANCELLED`
   - Example: `gosynctasks MyList -s T<TAB>` → `gosynctasks MyList -s TODO`

5. **View Flag Completion (`--view`):**
   - Dynamically loads available views from `~/.config/gosynctasks/views/`
   - Includes built-in views: `default`, `all`
   - Example: `gosynctasks MyList -v my<TAB>` → `gosynctasks MyList -v myview`

**Shell Support:**
- **Zsh:** Full support with descriptions
- **Bash:** Full support
- **Fish:** Full support
- **PowerShell:** Full support

**Activation:**

```bash
# Zsh (add to .zshrc)
eval "$(gosynctasks completion zsh)"

# Bash (add to .bashrc)
eval "$(gosynctasks completion bash)"

# Fish (add to config.fish)
gosynctasks completion fish | source

# PowerShell (add to profile)
gosynctasks completion powershell | Out-String | Invoke-Expression
```

**User Journey:**

```bash
# User starts typing list name
$ gosynctasks Wo<TAB>

# Shell completes to matching list
$ gosynctasks Work

# User types action prefix
$ gosynctasks Work u<TAB>

# Shell shows options: update
$ gosynctasks Work update

# User sets status flag
$ gosynctasks Work -s T<TAB>

# Shell completes
$ gosynctasks Work -s TODO
```

**Prerequisites:**
- Completion script loaded in shell configuration
- For list/view completion: Application can initialize (valid config)
- Shell supports completion (all modern shells do)

**Outputs/Results:**
- Tab key triggers suggestions
- Multiple matches show selection menu (shell-dependent)
- No matches: no action (shell beeps or shows nothing)
- Completion failures fallback gracefully

**Related Features:**
- [Task Management](./TASK_MANAGEMENT.md) - Commands being completed
- [Views Customization](./VIEWS_CUSTOMIZATION.md) - View name completion
- [List Management](./LIST_MANAGEMENT.md) - List name completion

---

## Terminal Width Detection

### Feature: Adaptive Display Formatting

**Purpose:** Automatically adjusts output width to fit the user's terminal, ensuring readability across different screen sizes and preventing line wrapping issues.

**How It Works:**

1. **Detection Process:**
   - Queries terminal size using `term.GetSize(os.Stdout.Fd())`
   - Returns width in columns and height in rows
   - Falls back to 80 columns if detection fails (non-TTY, error)

2. **Width Constraints:**
   - Minimum width: 40 characters (prevents cramped display)
   - Maximum width: 100 characters (prevents overly wide lines)
   - Applied formula: `borderWidth = min(max(width - 2, 40), 100)`
   - Padding of 2 chars for borders

3. **Usage Locations:**
   - Interactive list selection borders
   - Task display tables
   - Help text formatting
   - Progress indicators

4. **Platform Support:**
   - Cross-platform via `golang.org/x/term` package
   - Works on Linux, macOS, Windows
   - Graceful degradation for pipes/redirects

**User Journey:**

```bash
# Wide terminal (120 chars)
$ gosynctasks
┌─ Available Task Lists ──────────────────────────────────────────────────────────────────────────┐
  # Border extends to 100 chars (max)

# Narrow terminal (60 chars)
$ gosynctasks
┌─ Available Task Lists ──────────────────────────────┐
  # Border fits 58 chars

# Very narrow terminal (30 chars)
$ gosynctasks
┌─ Available Task Lists ──────────┐
  # Border uses minimum 40 chars

# Redirected output (not a TTY)
$ gosynctasks | less
  # Uses default 80 chars
```

**Prerequisites:**
- None (always active)
- TTY environment for accurate detection
- Non-TTY uses safe default

**Outputs/Results:**
- Properly sized borders and tables
- No horizontal scrolling needed
- Consistent alignment across rows
- Readable on all terminal sizes

**Technical Details:**
- Detection function: `internal/cli/display.go:13-20`
- Package: `golang.org/x/term` from Go extended libraries
- Error handling returns 80 as fallback
- Width applied in `ShowTaskLists()` function (lines 23-78)
- Algorithm: `borderWidth = termWidth - 2; clamp(borderWidth, 40, 100)`
- Used for header/footer borders and padding calculations

**Related Features:**
- [Interactive List Selection](#interactive-list-selection) - Primary user
- Task display tables (uses same detection)
- View rendering (inherits terminal width)

---

## Argument Parsing and Validation

### Feature: Intelligent Command Line Parsing

**Purpose:** Validates user input, provides helpful error messages, and supports flexible argument formats while maintaining strict correctness.

**How It Works:**

1. **Argument Count Validation:**
   - Maximum 3 positional arguments enforced by Cobra
   - `Args: cobra.MaximumNArgs(3)` in root command
   - Extra arguments trigger error with usage help

2. **Action Resolution:**
   - Checks if argument 2 matches known action
   - Supports full names: `get`, `add`, `update`, `complete`, `delete`
   - Supports abbreviations: `g`, `a`, `u`, `c`, `d`
   - Case-insensitive matching
   - Invalid actions show error with valid action list

3. **List Name Validation:**
   - First argument checked against available task lists
   - Exact match required (case-sensitive)
   - Unknown list names trigger interactive selection or error

4. **Flag Validation:**
   - Type checking: integers for priority, strings for dates
   - Date format validation: must match YYYY-MM-DD
   - Priority range checking: 0-9
   - Status validation: must be TODO/T, DONE/D, PROCESSING/P, CANCELLED/C
   - Empty strings allowed for clearing date values

5. **Task Summary Handling:**
   - Third argument treated as free text
   - Supports spaces and special characters
   - Slash handling depends on `--literal` flag
   - Missing summary prompts user or shows error (action-dependent)

**User Journey:**

```bash
# Too many arguments
$ gosynctasks MyList add "Task" extra
Error: accepts at most 3 arg(s), received 4

# Invalid action
$ gosynctasks MyList show
Error: unknown action "show". Valid actions: get, add, update, complete, delete

# Invalid status
$ gosynctasks MyList add "Task" -s RUNNING
Error: invalid status "RUNNING". Valid: TODO/T, DONE/D, PROCESSING/P, CANCELLED/C

# Invalid date format
$ gosynctasks MyList add "Task" --due-date 01/31/2025
Error: invalid date format. Use YYYY-MM-DD

# Invalid priority
$ gosynctasks MyList add "Task" -p 10
Error: priority must be 0-9 (0=undefined, 1=highest, 9=lowest)

# Valid commands
$ gosynctasks MyList add "Task"  # OK
$ gosynctasks MyList a "Task"    # OK (abbreviation)
$ gosynctasks MyList              # OK (default to get)
```

**Prerequisites:**
- Valid configuration loaded
- Backend accessible
- For list operations: List must exist

**Outputs/Results:**
- Valid input: Command executes successfully
- Invalid input: Error message with specific problem and valid options
- Missing required input: Prompts user or shows usage help
- Help text available via `--help` flag

**Related Features:**
- [Task Management](./TASK_MANAGEMENT.md) - Actions being validated
- [Action Flags](#action-flags) - Flag validation details
- [Command Structure](#command-structure) - Argument positions

---

## Verbose Mode

### Feature: Debug Logging

# CLI Verbosity and Logging in Go

For a well-structured CLI, I'd recommend a tiered verbosity system with centralized logging:

## Verbosity Levels

**Standard approach:**
- Default (0): Only show critical info and errors
- Verbose (--verbose): Add informational messages about what's happening
- Debug (-vv or --debug): Include detailed debug information
- Trace (-vvv): Everything including function calls, state changes

**Example log output:**
```
2026/01/14 15:30:45 INFO:  Starting application
2026/01/14 15:30:45 DEBUG: Configuration loaded: verbose=2, logFile=/var/log/app.log
2026/01/14 15:30:45 TRACE: Entering processing loop
2026/01/14 15:30:45 DEBUG: Processing item: example
2026/01/14 15:30:45 INFO:  Application completed successfully
```

**Implementation pattern:**

```go
package main

import (
    "fmt"
    "io"
    "log"
    "os"
)

type LogLevel int

const (
    LevelError LogLevel = iota
    LevelWarn
    LevelInfo
    LevelDebug
    LevelTrace
)

type Logger struct {
    level      LogLevel
    errorLog   *log.Logger
    warnLog    *log.Logger
    infoLog    *log.Logger
    debugLog   *log.Logger
    traceLog   *log.Logger
    logFile    *os.File
}

func NewLogger(level LogLevel, logFilePath string, fileOnly bool) (*Logger, error) {
    flags := log.Ldate | log.Ltime | log.Lmsgprefix
    
    var logFile *os.File
    var err error
    var output io.Writer = os.Stdout
    var errOutput io.Writer = os.Stderr
    
    if logFilePath != "" {
        logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
        if err != nil {
            return nil, fmt.Errorf("failed to open log file: %w", err)
        }
        
        if fileOnly {
            output = logFile
            errOutput = logFile
        } else {
            output = io.MultiWriter(os.Stdout, logFile)
            errOutput = io.MultiWriter(os.Stderr, logFile)
        }
    }
    
    return &Logger{
        level:      level,
        errorLog:   log.New(errOutput, "ERROR: ", flags),
        warnLog:    log.New(errOutput, "WARN:  ", flags),
        infoLog:    log.New(output, "INFO:  ", flags),
        debugLog:   log.New(output, "DEBUG: ", flags),
        traceLog:   log.New(output, "TRACE: ", flags),
        logFile:    logFile,
    }, nil
}

func (l *Logger) Close() error {
    if l.logFile != nil {
        return l.logFile.Close()
    }
    return nil
}

func (l *Logger) Error(format string, v ...interface{}) {
    l.errorLog.Printf(format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
    if l.level >= LevelWarn {
        l.warnLog.Printf(format, v...)
    }
}

func (l *Logger) Info(format string, v ...interface{}) {
    if l.level >= LevelInfo {
        l.infoLog.Printf(format, v...)
    }
}

func (l *Logger) Debug(format string, v ...interface{}) {
    if l.level >= LevelDebug {
        l.debugLog.Printf(format, v...)
    }
}

func (l *Logger) Trace(format string, v ...interface{}) {
    if l.level >= LevelTrace {
        l.traceLog.Printf(format, v...)
    }
}

func (l *Logger) SetOutput(w io.Writer) {
    l.errorLog.SetOutput(w)
    l.warnLog.SetOutput(w)
    l.infoLog.SetOutput(w)
    l.debugLog.SetOutput(w)
    l.traceLog.SetOutput(w)
}
```

## Usage in CLI

```go
package main

import (
    "flag"
    "fmt"
    "os"
)

type Config struct {
    Verbose int
    LogFile string
    Logger  *Logger
}

func main() {
    config := &Config{}
    level := ...
    config.Logfile := ...
    
    logger, err := NewLogger(level, config.LogFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "ERROR: failed to initialize logger: %v\n", err)
        os.Exit(1)
    }
    defer logger.Close()
    
    config.Logger = logger
    
    if err := run(config); err != nil {
        config.Logger.Error("fatal: %v", err)
        os.Exit(1)
    }
}

func run(cfg *Config) error {
    cfg.Logger.Info("Starting application")
    cfg.Logger.Debug("Configuration loaded: verbose=%d, logFile=%s", cfg.Verbose, cfg.LogFile)
    
    cfg.Logger.Trace("Entering processing loop")
    
    if err := doWork(cfg); err != nil {
        return fmt.Errorf("work failed: %w", err)
    }
    
    cfg.Logger.Info("Application completed successfully")
    return nil
}

func doWork(cfg *Config) error {
    cfg.Logger.Debug("Processing item: %s", "example")
    cfg.Logger.Trace("Item state: %+v", map[string]string{"key": "value"})
    return nil
}
```

**Basic file logging:**
Set in config.yaml

**Log rotation considerations:**

For production CLIs that run as services or frequently, consider using a log rotation library:

```go
import "gopkg.in/natefinch/lumberjack.v2"

func NewLogger(level LogLevel, logFilePath string) (*Logger, error) {
    flags := log.Ldate | log.Ltime | log.Lmsgprefix
    
    var output io.Writer = os.Stdout
    var errOutput io.Writer = os.Stderr
    
    if logFilePath != "" {
        fileWriter := &lumberjack.Logger{
            Filename:   logFilePath,
            MaxSize:    100,
            MaxBackups: 3,
            MaxAge:     28,
            Compress:   true,
        }
        
        output = io.MultiWriter(os.Stdout, fileWriter)
        errOutput = io.MultiWriter(os.Stderr, fileWriter)
    }
    
    return &Logger{
        level:      level,
        errorLog:   log.New(errOutput, "ERROR: ", flags),
        warnLog:    log.New(errOutput, "WARN:  ", flags),
        infoLog:    log.New(output, "INFO:  ", flags),
        debugLog:   log.New(output, "DEBUG: ", flags),
        traceLog:   log.New(output, "TRACE: ", flags),
    }, nil
}
```

## Best Practices

**For flags:** Use `--verbose` for verbose, `-vv` for debug, `-vvv` for trace.

**Error handling:** Always log errors at Error level. Return them up the stack for the caller to decide what to do.

**Warnings:** Use for recoverable issues or deprecated features.

**Info:** Default operational messages users should see (started, completed, processed N items).

**Debug:** Internal state useful for troubleshooting.

**Trace:** Function entry/exit, variable dumps—very noisy.

**Silent mode:** Consider adding `--quiet` that suppresses everything except errors.

**File logging patterns:**
- Use `io.MultiWriter` to output to both console and file simultaneously
- Always close the log file with `defer logger.Close()`
- Consider log rotation for long-running processes
- Use absolute paths or paths relative to a well-known location for log files

The key is making your default output clean and useful, with verbosity flags that progressively reveal more detail when debugging.

2. **Logging Output:**
   - Does not interfere with normal command output

3. **What Gets Logged:**
   - Configuration file loading and parsing
   - Backend selection and initialization
   - Credential resolution (source: keyring, env, config)
   - Database queries and sync operations
   - Requests to CalDAV/API servers
   - Task search and matching logic
   - View rendering and filtering decisions
   - Error stack traces and context


**Prerequisites:**
- None (flag always available)
- Output visible when stderr is not redirected

**Outputs/Results:**
- Debug messages do not affect command exit codes
- Useful for diagnosing sync conflicts, credential issues, backend errors

**Related Features:**
- [Credential Management](./CREDENTIAL_MANAGEMENT.md) - Logs credential sources
- [Synchronization](./SYNCHRONIZATION.md) - Logs sync operations
- [Backend System](./BACKEND_SYSTEM.md) - Logs backend selection
- All operations (provides diagnostic visibility)

---

## Backend Selection Flags

### Feature: Backend Discovery and Override

**Purpose:** Allows users to discover configured backends, see auto-detected backends, and override default backend selection for specific commands.

**How It Works:**

**Three Backend-Related Flags:**

1. **`--backend <name>` or `-b <name>`:**
   - Forces use of specified backend for the command
   - Overrides config `default_backend` and auto-detection
   - Backend name must match config entry
   - Passed to `app.NewApp(backendName)`
   - Affects entire command execution

2. **`--list-backends`:**
   - Displays all backends defined in configuration
   - Shows status: enabled/disabled
   - Shows type: nextcloud, sqlite, git, todoist, file
   - Shows connection details (host, file, path)
   - Exits immediately after display (no other action taken)

3. **`--detect-backend`:**
   - Runs auto-detection algorithm
   - Shows which backends would be auto-selected
   - Shows detection result and priority order
   - Exits immediately after display

**Selection Priority:**
1. Explicit `--backend` flag (highest priority)
3. Auto-detected backend (if `auto_detect_backend = true`)
4. Config `default_backend`
5. First enabled backend in config

**User Journey:**

```bash
# List all configured backends
$ gosynctasks --list-backends

Available Backends:
  nextcloud-prod (enabled)  - Nextcloud CalDAV @ nextcloud.example.com
  nextcloud-test (enabled)  - Nextcloud CalDAV @ localhost:8080
  sqlite (enabled)          - SQLite @ ~/.local/share/gosynctasks/tasks.db
  git (disabled)            - Git Markdown @ TODO.md
  todoist (enabled)         - Todoist API

# Detect backends in current directory
$ gosynctasks --detect-backend

Auto-Detection Results:
  git: Detected (TODO.md with marker found)

# Use specific backend for one command
$ gosynctasks --backend nextcloud-test MyList

# Use specific backend with other flags
$ gosynctasks --backend git --verbose ProjectTasks add "Fix bug"
```

**Prerequisites:**
- `--backend`: Backend name must exist in config
- `--list-backends`: Valid configuration file loaded
- `--detect-backend`: Auto-detection enabled in config

**Outputs/Results:**
- `--backend`: Command uses specified backend
- `--list-backends`: Table of backends with status/details
- `--detect-backend`: Detection results with matched backends
- Invalid backend name: Error message with available backends

**Related Features:**
- [Backend System](./BACKEND_SYSTEM.md) - Full backend selection details
- [Backend Auto-Detection](./BACKEND_SYSTEM.md#auto-detection) - Git detection logic
- [Configuration](./CONFIGURATION.md) - Backend configuration

---

## Graceful Shutdown

### Feature: Signal Handling and Cleanup

**Purpose:** Ensures application shuts down cleanly when interrupted (Ctrl+C, SIGTERM), preventing database corruption, incomplete sync operations, and orphaned background processes.

**How It Works:**

1. **Signal Registration:**
   - Creates buffered channel for OS signals
   - Registers handlers for SIGINT (Ctrl+C) and SIGTERM
   - Channel capacity: 1 signal

2. **Background Goroutine:**
   - Spawned at application startup
   - Blocks waiting for signal on channel
   - When signal received:
     - Calls `application.Shutdown()` if app initialized
     - Exits with code 0

3. **Shutdown Sequence (in `application.Shutdown()`):**
   - Closes active backend connections
   - Flushes pending database transactions
   - Stops background sync processes (if running)
   - Cleans up temporary files
   - Closes log files

4. **Background Sync Handling:**
   - Background sync runs in detached process
   - Not affected by main process shutdown
   - Continues sync operations independently
   - See [Synchronization](./SYNCHRONIZATION.md) for details

**User Journey:**

```bash
# User starts long-running operation
$ gosynctasks sync

Syncing with remote backend...
[=============>          ] 60% complete

# User presses Ctrl+C
^C

# System performs cleanup
Shutting down gracefully...
Database connections closed
Background sync detached (continues in background)
Exit code: 0

# Application exits cleanly
$ echo $?
0
```

**Prerequisites:**
- Application initialized (signal handler always registered)
- OS supports SIGINT and SIGTERM signals

**Outputs/Results:**
- Clean shutdown with exit code 0
- Database transactions flushed (no corruption)
- Background processes handled appropriately
- Open files closed
- Network connections terminated gracefully

**Related Features:**
- [Synchronization](./SYNCHRONIZATION.md) - Background sync detachment
- [Backend System](./BACKEND_SYSTEM.md) - Connection cleanup
- Database operations - Transaction flushing

---

## No-Prompt Mode

### Feature: Non-Interactive Operation for Scripting

**Purpose:** Enables fully non-interactive operation of gosynctasks for use in scripts, automation, CI/CD pipelines, and programmatic task management. When enabled, all interactive prompts are bypassed with deterministic behavior.

**How It Works:**

1. **Activation:**
   - CLI flag: `-y` or `--no-prompt`
   - Config setting: `ui.no_prompt: true`
   - CLI flag overrides config setting
   - Application tests default to no-prompt mode

2. **Behavior Changes:**

| Scenario | Normal Mode | No-Prompt Mode |
|----------|-------------|----------------|
| Delete confirmation | Prompts "Confirm? [y/N]" | Deletes immediately (force mode) |
| Single task match | Prompts "Is this correct? (y/n)" | Uses match automatically |
| Multiple task matches | Shows selection menu | Outputs match table + `ACTION_INCOMPLETE` |
| No list specified | Interactive list picker | Outputs available lists + `INFO_ONLY` |
| Config auto-init | Prompts to create config | Creates config silently |

3. **Result Codes:**
   - All commands output a result code (see [Result Codes](#result-codes))
   - `ACTION_COMPLETED`: Operation performed successfully
   - `ACTION_INCOMPLETE`: Ambiguous input, user decision required
   - `INFO_ONLY`: Display-only operation, no changes made
   - `ERROR`: Operation failed with error message

**User Journey:**

```bash
# Script: Complete all tasks matching "review"
$ gosynctasks -y --json MyList complete "review"

# If single match: Task completed
{
  "action": "complete",
  "task": {"uid": "550e8400...", "summary": "Review PR #456", "status": "DONE"},
  "result": "ACTION_COMPLETED"
}

# If multiple matches: Returns candidates for user/script to choose
{
  "matches": [
    {"uid": "550e8400...", "summary": "Review PR #456", "parents": ["Project Alpha"]},
    {"uid": "660e8400...", "summary": "Code review", "parents": []}
  ],
  "result": "ACTION_INCOMPLETE",
  "message": "Multiple tasks match 'review'. Use --uid to specify exact task."
}

# Script then uses specific UID
$ gosynctasks -y MyList complete --uid "550e8400-e29b-41d4-a716-446655440000"
ACTION_COMPLETED
```

**Multiple Match Output (Text Mode):**

```
Multiple tasks match "review":
UID:550e8400-e29b-41d4-a716-446655440000	TASK:Review PR #456	PARENT:Project Alpha
UID:660e8400-e29b-41d4-a716-446655440001	TASK:Code review guidelines	PARENT:
UID:NOT-SYNCED-42	TASK:Review meeting notes	PARENT:Documentation/Meetings
ACTION_INCOMPLETE
```

Note: `NOT-SYNCED-<id>` indicates a task created locally but not yet synced to the remote backend. The number is the SQLite internal ID. Use this value with `--uid` to operate on unsynced tasks.

**No List Specified Output:**

```bash
$ gosynctasks -y
Available lists:
ID:abc-123	NAME:Work	TASKS:12
ID:def-456	NAME:Personal	TASKS:5
ID:ghi-789	NAME:Shopping	TASKS:0
INFO_ONLY
```

**Prerequisites:**
- None (always available)

**Outputs/Results:**
- Deterministic, parseable output for all operations
- Result code always appears at end of output
- No stdin reads (script-safe)
- Exit codes reflect result: 0 for success/info, non-zero for errors

**Technical Details:**

**Configuration:**
```yaml
ui:
  no_prompt: false  # Default: prompts enabled
```

**Flag Registration:**
```go
rootCmd.PersistentFlags().BoolP("no-prompt", "y", false, "Disable interactive prompts")
```

**Prompt Manager:**
- Centralized prompt handling in `internal/cli/prompt/`
- All prompts routed through PromptManager
- PromptManager checks no-prompt mode before prompting
- In no-prompt mode: returns default/structured output instead of prompting

**Testing:**
- Tests set `no_prompt: true` by default for deterministic behavior
- Enables automated testing without mock stdin

**Related Features:**
- [JSON Output Mode](#json-output-mode) - Machine-parseable output format
- [Result Codes](#result-codes) - Output status indicators
- [Task Search and Selection](./TASK_MANAGEMENT.md#task-search-and-selection) - Affected by no-prompt mode
- [Interactive List Selection](#interactive-list-selection) - Bypassed in no-prompt mode

---

## JSON Output Mode

### Feature: Machine-Parseable JSON Output

**Purpose:** Provides structured JSON output for all operations, enabling robust integration with scripts, automation tools, and other programs that consume gosynctasks output.

**How It Works:**

1. **Activation:**
   - CLI flag: `--json`
   - Config setting: `ui.output_format: json`
   - CLI flag overrides config setting

2. **Output Structure:**
   - All responses are valid JSON objects
   - Consistent field names across all operations
   - Includes result code in every response
   - Error messages included in JSON structure

**User Journey:**

**List Tasks (JSON):**
```bash
$ gosynctasks --json MyList
```
```json
{
  "list": {
    "id": "abc-123",
    "name": "MyList",
    "task_count": 3
  },
  "tasks": [
    {
      "uid": "550e8400-e29b-41d4-a716-446655440000",
      "summary": "Review PR #456",
      "status": "TODO",
      "priority": 2,
      "due_date": "2026-01-20",
      "parents": ["Project Alpha"]
    },
    {
      "uid": "660e8400-e29b-41d4-a716-446655440001",
      "summary": "Write documentation",
      "status": "PROCESSING",
      "priority": 5,
      "parents": []
    }
  ],
  "result": "INFO_ONLY"
}
```

**Add Task (JSON):**
```bash
$ gosynctasks --json MyList add "New task" -p 1 --due-date 2026-01-25
```
```json
{
  "action": "add",
  "task": {
    "uid": "770e8400-e29b-41d4-a716-446655440002",
    "summary": "New task",
    "status": "TODO",
    "priority": 1,
    "due_date": "2026-01-25",
    "created": "2026-01-16T10:30:00Z"
  },
  "result": "ACTION_COMPLETED"
}
```

**Update Task (JSON):**
```bash
$ gosynctasks --json MyList update "PR" -s DONE
```
```json
{
  "action": "update",
  "task": {
    "uid": "550e8400-e29b-41d4-a716-446655440000",
    "summary": "Review PR #456",
    "status": "DONE",
    "completed": "2026-01-16T10:35:00Z"
  },
  "result": "ACTION_COMPLETED"
}
```

**List Available Lists (JSON):**
```bash
$ gosynctasks --json
```
```json
{
  "lists": [
    {"id": "abc-123", "name": "Work", "task_count": 12, "color": "#0066cc"},
    {"id": "def-456", "name": "Personal", "task_count": 5, "color": "#ff5733"},
    {"id": "ghi-789", "name": "Shopping", "task_count": 0, "color": "#00cc66"}
  ],
  "result": "INFO_ONLY"
}
```

**Multiple Matches (JSON with no-prompt):**
```bash
$ gosynctasks -y --json MyList complete "review"
```
```json
{
  "matches": [
    {
      "uid": "550e8400-e29b-41d4-a716-446655440000",
      "summary": "Review PR #456",
      "status": "TODO",
      "priority": 2,
      "parents": ["Project Alpha"],
      "synced": true
    },
    {
      "uid": "660e8400-e29b-41d4-a716-446655440001",
      "summary": "Code review guidelines",
      "status": "DONE",
      "priority": 0,
      "parents": [],
      "synced": true
    },
    {
      "uid": "NOT-SYNCED-42",
      "summary": "Review meeting notes",
      "status": "PROCESSING",
      "priority": 5,
      "parents": ["Documentation", "Meetings"],
      "synced": false
    }
  ],
  "result": "ACTION_INCOMPLETE",
  "message": "Multiple tasks match 'review'. Use --uid to specify exact task."
}
```

Note: Tasks with `"synced": false` have `uid` in `NOT-SYNCED-<sqlite-id>` format. This is the SQLite internal ID and can be used with `--uid` until the task is synced to the remote backend.

**Error Response (JSON):**
```bash
$ gosynctasks --json NonExistentList
```
```json
{
  "result": "ERROR",
  "message": "Error 1: List 'NonExistentList' not found"
}
```

**Error with Context (JSON):**
```bash
$ gosynctasks --json Work update "task" --due-date "bad-date"
```
```json
{
  "result": "ERROR",
  "message": "Error 2: Invalid date format 'bad-date', expected YYYY-MM-DD"
}
```

**Prerequisites:**
- None (always available)

**Outputs/Results:**
- All output is valid JSON
- Single JSON object per command (not JSON lines)
- UTF-8 encoded
- Parseable with `jq`, Python `json` module, or any JSON parser

**Technical Details:**

**Configuration:**
```yaml
ui:
  output_format: text  # text | json
```

**Benefits over Text Output:**
| Aspect | Text Output | JSON Output |
|--------|-------------|-------------|
| Parsing | Regex/awk (fragile) | Native JSON parsing |
| Structure | Flat, tabular | Hierarchical, typed |
| Parent hierarchy | Awkward encoding | Clean arrays |
| Extensibility | Breaking changes likely | Add fields safely |
| Error handling | Parse stderr | Structured error object |

**JSON Field Reference:**

| Field | Type | Description |
|-------|------|-------------|
| `result` | string | Result code: `ACTION_COMPLETED`, `ACTION_INCOMPLETE`, `INFO_ONLY`, `ERROR` |
| `action` | string | Operation performed: `add`, `update`, `complete`, `delete` |
| `task` | object | Task object for single-task operations |
| `tasks` | array | Array of task objects for list operations |
| `matches` | array | Array of matching tasks when ambiguous |
| `list` | object | List metadata for list-specific operations |
| `lists` | array | Array of list objects |
| `message` | string | Human-readable message: info for `ACTION_INCOMPLETE`, error details for `ERROR` (format: `Error #: description`) |
| `synced` | boolean | Whether task has been synced to remote (in task objects) |

**UID Format:**
| Format | Description |
|--------|-------------|
| `<uuid>` | Backend-assigned UID (e.g., `550e8400-e29b-41d4-a716-446655440000`) |
| `NOT-SYNCED-<id>` | SQLite internal ID for unsynced tasks (e.g., `NOT-SYNCED-42`) |

**Related Features:**
- [No-Prompt Mode](#no-prompt-mode) - Pairs well for scripting
- [Result Codes](#result-codes) - Included in JSON output
- [Views & Customization](./VIEWS_CUSTOMIZATION.md) - Custom views also support JSON output

---

## Result Codes

### Feature: Standardized Operation Outcome Indicators

**Purpose:** Provides clear, machine-parseable indicators of command outcomes, enabling scripts and automation to reliably determine what happened and take appropriate action.

**How It Works:**

1. **Output Location:**
   - Text mode: Last line of output (e.g., `ACTION_COMPLETED`)
   - JSON mode: `result` field in JSON object

2. **Result Code Values:**

| Code | Meaning | When Used | Exit Code |
|------|---------|-----------|-----------|
| `ACTION_COMPLETED` | Operation performed successfully | Task added, updated, deleted, completed | 0 |
| `ACTION_INCOMPLETE` | Operation requires user decision | Multiple matches found, ambiguous input | 0 |
| `INFO_ONLY` | No action performed, display only | List tasks, show lists, view status | 0 |
| `ERROR` | Operation failed | Invalid input, backend error, not found | Non-zero |

3. **ERROR Format:**

**Text Mode:**
```
Error 1: List 'NonExistent' not found
ERROR
```

**JSON Mode:**
```json
{
  "result": "ERROR",
  "message": "Error 1: List 'NonExistent' not found"
}
```

**Common Error Codes:**
| Error # | Description |
|---------|-------------|
| 1 | Resource not found (list, task) |
| 2 | Invalid input (bad flag value, malformed date) |
| 3 | Backend connection failed |
| 4 | Permission denied |
| 5 | Sync conflict |
| 6 | Validation error |

4. **Behavior by Operation:**

| Operation | Success Result | Ambiguous Result |
|-----------|---------------|------------------|
| `add` | `ACTION_COMPLETED` | N/A (always creates new) |
| `update` | `ACTION_COMPLETED` | `ACTION_INCOMPLETE` (multiple matches) |
| `complete` | `ACTION_COMPLETED` | `ACTION_INCOMPLETE` (multiple matches) |
| `delete` | `ACTION_COMPLETED` | `ACTION_INCOMPLETE` (multiple matches) |
| `get` (list tasks) | `INFO_ONLY` | N/A |
| No list specified | `INFO_ONLY` | N/A |
| `sync` | `ACTION_COMPLETED` | N/A |
| `sync status` | `INFO_ONLY` | N/A |

**User Journey:**

**Successful Operation:**
```bash
$ gosynctasks -y MyList complete "unique task"
Task 'unique task' marked as DONE
ACTION_COMPLETED

$ echo $?
0
```

**Ambiguous Operation:**
```bash
$ gosynctasks -y MyList complete "review"
Multiple tasks match "review":
UID:550e8400...	TASK:Review PR #456	PARENT:Project Alpha
UID:660e8400...	TASK:Code review	PARENT:
ACTION_INCOMPLETE

$ echo $?
0
```

**Display-Only Operation:**
```bash
$ gosynctasks -y MyList
Tasks in "MyList" (5 tasks):
[TODO] Task 1
[DONE] Task 2
INFO_ONLY

$ echo $?
0
```

**Error (Text Mode):**
```bash
$ gosynctasks -y NonExistent
Error 1: List 'NonExistent' not found
ERROR

$ echo $?
1
```

**Error (JSON Mode):**
```bash
$ gosynctasks -y --json NonExistent
```
```json
{
  "result": "ERROR",
  "message": "Error 1: List 'NonExistent' not found"
}
```

**Multiple Errors:**
```bash
$ gosynctasks -y Work update "task" --due-date "invalid"
Error 2: Invalid date format 'invalid', expected YYYY-MM-DD
ERROR
```

**Script Integration Example:**

```bash
#!/bin/bash
# Complete a task and handle results

output=$(gosynctasks -y MyList complete "$1" 2>&1)
result=$(echo "$output" | tail -1)

case "$result" in
  "ACTION_COMPLETED")
    echo "Task completed successfully"
    ;;
  "ACTION_INCOMPLETE")
    echo "Multiple matches found. Please specify UID:"
    echo "$output" | grep "^UID:"
    ;;
  "ERROR")
    # Extract error number and message
    error_line=$(echo "$output" | grep "^Error")
    echo "Operation failed: $error_line"
    exit 1
    ;;
esac
```

**JSON Script Example:**

```bash
#!/bin/bash
# Complete a task using JSON output

response=$(gosynctasks -y --json MyList complete "$1")
result=$(echo "$response" | jq -r '.result')

case "$result" in
  "ACTION_COMPLETED")
    uid=$(echo "$response" | jq -r '.task.uid')
    echo "Completed task: $uid"
    ;;
  "ACTION_INCOMPLETE")
    echo "Multiple matches. UIDs:"
    echo "$response" | jq -r '.matches[].uid'
    ;;
  "ERROR")
    message=$(echo "$response" | jq -r '.message')
    echo "Error: $message"
    exit 1
    ;;
esac
```

**Prerequisites:**
- Result codes always present in no-prompt mode
- In normal (prompt) mode, result codes only shown with `--show-result` flag (optional)

**Outputs/Results:**
- Consistent result code format
- Exit codes align with result (0 for non-errors, non-zero for errors)
- Result codes never change mid-stream (always at end)

**Technical Details:**

**Implementation:**
```go
type ResultCode string

const (
    ResultActionCompleted  ResultCode = "ACTION_COMPLETED"
    ResultActionIncomplete ResultCode = "ACTION_INCOMPLETE"
    ResultInfoOnly         ResultCode = "INFO_ONLY"
    ResultError            ResultCode = "ERROR"
)

func (r ResultCode) ExitCode() int {
    if r == ResultError {
        return 1
    }
    return 0
}

// Error codes
const (
    ErrNotFound       = 1  // Resource not found
    ErrInvalidInput   = 2  // Invalid input
    ErrConnection     = 3  // Backend connection failed
    ErrPermission     = 4  // Permission denied
    ErrSyncConflict   = 5  // Sync conflict
    ErrValidation     = 6  // Validation error
)

type AppError struct {
    Code    int
    Message string
}

func (e AppError) Error() string {
    return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

// JSON output for errors
type ErrorResponse struct {
    Result  string `json:"result"`  // Always "ERROR"
    Message string `json:"message"` // "Error #: description"
}
```

**Result Code Selection Logic:**
```
IF error occurred:
    result = ERROR
ELSE IF operation mutated data (add/update/delete/complete):
    IF operation succeeded:
        result = ACTION_COMPLETED
    ELSE IF multiple matches (no-prompt mode):
        result = ACTION_INCOMPLETE
ELSE (read-only operation):
    result = INFO_ONLY
```

**Related Features:**
- [No-Prompt Mode](#no-prompt-mode) - Primary use case for result codes
- [JSON Output Mode](#json-output-mode) - Result codes in JSON response
- [Task Search and Selection](./TASK_MANAGEMENT.md#task-search-and-selection) - Source of `ACTION_INCOMPLETE`

---

## Summary

The CLI interface of gosynctasks provides a comprehensive command-line experience with:

- **Flexible syntax** supporting full commands and abbreviations
- **Rich flag system** for fine-grained control over operations
- **Interactive modes** for browsing lists and selecting tasks
- **Shell completion** for faster, error-free command entry
- **Adaptive display** that adjusts to terminal size
- **Robust validation** with helpful error messages
- **Verbose mode** for troubleshooting
- **Backend discovery** for multi-backend environments
- **Graceful shutdown** ensuring data integrity
- **No-prompt mode** for scripting and automation (`-y`, `--no-prompt`)
- **JSON output** for machine-parseable results (`--json`)
- **Result codes** for reliable script integration (`ACTION_COMPLETED`, `ACTION_INCOMPLETE`, `INFO_ONLY`)
- **UID-based selection** for unambiguous task operations (`--uid`)

All interface features are built on the Cobra framework, providing consistent behavior, automatic help generation, and cross-platform compatibility.

**Next Steps:**
- See [Configuration](./CONFIGURATION.md) for setup details
- See [Task Management](./TASK_MANAGEMENT.md) for operation specifics
- See [Backend System](./BACKEND_SYSTEM.md) for backend selection details
- See [Views Customization](./VIEWS_CUSTOMIZATION.md) for output formatting
