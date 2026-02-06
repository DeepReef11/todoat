package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// defaultBackgroundLoggingEnabled is the default value when no config is available.
// This is used by NewBackgroundLogger() when called without explicit enabled parameter.
// The runtime config option logging.background_enabled overrides this default.
const defaultBackgroundLoggingEnabled = true

// Logger provides leveled logging with verbose mode support.
type Logger struct {
	verbose bool
	mu      sync.RWMutex
}

var (
	loggerInstance *Logger
	once           sync.Once
)

// GetLogger returns the singleton logger instance.
func GetLogger() *Logger {
	once.Do(func() {
		loggerInstance = &Logger{
			verbose: false,
		}
	})
	return loggerInstance
}

// SetVerboseMode sets the verbose mode globally.
func SetVerboseMode(verbose bool) {
	logger := GetLogger()
	logger.SetVerbose(verbose)
}

// SetVerbose sets the verbose mode for this logger instance.
func (l *Logger) SetVerbose(verbose bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.verbose = verbose
}

// IsVerbose returns whether verbose mode is enabled.
func (l *Logger) IsVerbose() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.verbose
}

// formatMessage formats a message with optional printf-style arguments.
func formatMessage(msgOrFormat string, args ...interface{}) string {
	if len(args) > 0 {
		return fmt.Sprintf(msgOrFormat, args...)
	}
	return msgOrFormat
}

// Debug logs a debug message (only shown when verbose=true).
// Can be used with a simple message or printf-style format string with args.
func (l *Logger) Debug(msgOrFormat string, args ...interface{}) {
	if !l.IsVerbose() {
		return
	}
	fmt.Fprintf(os.Stderr, "%s [DEBUG] %s\n", time.Now().Format("15:04:05"), formatMessage(msgOrFormat, args...))
}

// Info logs an info message (always shown).
// Can be used with a simple message or printf-style format string with args.
func (l *Logger) Info(msgOrFormat string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[INFO] %s\n", formatMessage(msgOrFormat, args...))
}

// Warn logs a warning message (always shown).
// Can be used with a simple message or printf-style format string with args.
func (l *Logger) Warn(msgOrFormat string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[WARN] %s\n", formatMessage(msgOrFormat, args...))
}

// Error logs an error message (always shown).
// Can be used with a simple message or printf-style format string with args.
func (l *Logger) Error(msgOrFormat string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] %s\n", formatMessage(msgOrFormat, args...))
}

// Debugf is a convenience function that logs a debug message using the global logger.
func Debugf(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

// Infof is a convenience function that logs an info message using the global logger.
func Infof(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

// Warnf is a convenience function that logs a warning message using the global logger.
func Warnf(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

// Errorf is a convenience function that logs an error message using the global logger.
func Errorf(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

// BackgroundLogger provides logging for background processes to a PID-specific file.
type BackgroundLogger struct {
	logger   *log.Logger
	logFile  *os.File
	enabled  bool
	filePath string
}

// NewBackgroundLogger creates a new background logger with a PID-specific log file.
// Uses the default enabled value. For runtime config control, use NewBackgroundLoggerWithEnabled.
func NewBackgroundLogger() (*BackgroundLogger, error) {
	return NewBackgroundLoggerWithEnabled(defaultBackgroundLoggingEnabled)
}

// NewBackgroundLoggerWithEnabled creates a background logger with explicit enabled control.
// This is the runtime config-aware version of NewBackgroundLogger.
// Pass config.IsBackgroundLoggingEnabled() to honor the logging.background_enabled config.
func NewBackgroundLoggerWithEnabled(enabled bool) (*BackgroundLogger, error) {
	if !enabled {
		return &BackgroundLogger{
			logger:  log.New(io.Discard, "", log.LstdFlags),
			enabled: false,
		}, nil
	}

	pid := os.Getpid()
	logPath := fmt.Sprintf("%s/todoat-%d.log", os.TempDir(), pid)
	return NewBackgroundLoggerWithPath(logPath)
}

// NewBackgroundLoggerWithPath creates a background logger with a custom path.
func NewBackgroundLoggerWithPath(path string) (*BackgroundLogger, error) {
	bl := &BackgroundLogger{
		filePath: path,
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Gracefully degrade to io.Discard
		bl.logger = log.New(io.Discard, "", log.LstdFlags)
		bl.enabled = false
		return bl, err
	}

	bl.logFile = file
	bl.logger = log.New(file, "", log.LstdFlags)
	bl.enabled = true
	return bl, nil
}

// Printf logs a formatted message.
func (bl *BackgroundLogger) Printf(format string, args ...interface{}) {
	if bl.logger != nil {
		bl.logger.Printf(format, args...)
	}
}

// Print logs a message.
func (bl *BackgroundLogger) Print(args ...interface{}) {
	if bl.logger != nil {
		bl.logger.Print(args...)
	}
}

// Println logs a message with a newline.
func (bl *BackgroundLogger) Println(args ...interface{}) {
	if bl.logger != nil {
		bl.logger.Println(args...)
	}
}

// Close closes the log file.
func (bl *BackgroundLogger) Close() {
	if bl.logFile != nil {
		_ = bl.logFile.Close()
		bl.logFile = nil
	}
	// After close, switch to io.Discard for graceful degradation
	bl.logger = log.New(io.Discard, "", log.LstdFlags)
	bl.enabled = false
}

// GetLogPath returns the log file path.
func (bl *BackgroundLogger) GetLogPath() string {
	return bl.filePath
}

// IsEnabled returns whether background logging is enabled.
func (bl *BackgroundLogger) IsEnabled() bool {
	return bl.enabled
}
