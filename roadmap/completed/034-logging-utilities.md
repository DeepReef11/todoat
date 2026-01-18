# [034] Logging and Utility Infrastructure

## Summary
Implement the internal/utils package with logging utilities, error constructors with user-friendly suggestions, input validation, and user prompt functions as described in LOGGING.md.

## Documentation Reference
- Primary: `dev-doc/LOGGING.md`
- Related: `dev-doc/CLI_INTERFACE.md` (verbose mode), `dev-doc/CONFIGURATION.md`

## Dependencies
- Requires: [002] Core CLI (for verbose flag integration)
- Requires: [010] Configuration (for log_file configuration)

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestVerboseModeEnabled` - `todoat -V MyList` outputs debug messages to stderr
- [ ] `TestVerboseModeDisabled` - `todoat MyList` does not output debug messages
- [ ] `TestBackgroundLoggerCreation` - Background logger creates PID-specific log file

### Functional Requirements
- [ ] Logger singleton with `GetLogger()` function
- [ ] Verbose mode toggle via `SetVerboseMode(bool)`
- [ ] Log levels: Debug, Info, Warn, Error with appropriate prefixes
- [ ] Debug level only visible when verbose=true
- [ ] BackgroundLogger writes to `/tmp/todoat-{PID}.log`
- [ ] BackgroundLogger gracefully degrades to io.Discard if file creation fails
- [ ] Pre-built error constructors:
  - `ErrTaskNotFound(searchTerm)` with search suggestion
  - `ErrListNotFound(listName)` with creation suggestion
  - `ErrNoListsAvailable()` with list create suggestion
  - `ErrSyncNotEnabled()` with config suggestion
  - `ErrBackendNotConfigured(name)` with config help
  - `ErrBackendOffline(name, reason)` with smart suggestions based on error type
  - `ErrInvalidPriority(priority)` with valid range
  - `ErrInvalidDate(dateStr)` with format hint
  - `ErrInvalidStatus(status, valid)` with valid options
  - `ErrCredentialsNotFound(backend, user)` with setup command
  - `ErrAuthenticationFailed(backend)` with verification suggestion
- [ ] Input validation: `ValidatePriority(int)`, `ParseDateFlag(string)`, `ValidateDateRange(start, due)`
- [ ] User prompts: `PromptYesNo()`, `PromptSelection()`, `PromptConfirmation()`

## Implementation Notes
- Logger uses sync.RWMutex for thread safety
- Error constructors return `ErrorWithSuggestion` type with `Error()` and `Suggestion()` methods
- Smart suggestion logic for backend offline errors based on DNS, connection refused, or timeout
- All prompts read from stdin with trimming and validation loops

## Out of Scope
- Log rotation (future enhancement)
- Remote logging
- Structured JSON logging
