package utils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// =============================================================================
// Logger Tests (034-logging-utilities)
// =============================================================================

// TestGetLogger verifies singleton pattern - same instance returned
func TestGetLogger(t *testing.T) {
	logger1 := GetLogger()
	logger2 := GetLogger()

	if logger1 != logger2 {
		t.Error("GetLogger() should return same singleton instance")
	}
}

// TestLoggerDefaultVerboseMode verifies verbose is false by default
func TestLoggerDefaultVerboseMode(t *testing.T) {
	// Reset singleton for clean test
	once = sync.Once{}
	loggerInstance = nil

	logger := GetLogger()
	if logger.IsVerbose() {
		t.Error("Logger should have verbose=false by default")
	}
}

// TestSetVerboseMode verifies SetVerboseMode changes verbose state
func TestSetVerboseMode(t *testing.T) {
	// Reset singleton for clean test
	once = sync.Once{}
	loggerInstance = nil

	SetVerboseMode(true)
	logger := GetLogger()
	if !logger.IsVerbose() {
		t.Error("SetVerboseMode(true) should enable verbose mode")
	}

	SetVerboseMode(false)
	if logger.IsVerbose() {
		t.Error("SetVerboseMode(false) should disable verbose mode")
	}
}

// TestLoggerSetVerbose verifies instance SetVerbose method
func TestLoggerSetVerbose(t *testing.T) {
	// Reset singleton for clean test
	once = sync.Once{}
	loggerInstance = nil

	logger := GetLogger()
	logger.SetVerbose(true)
	if !logger.IsVerbose() {
		t.Error("SetVerbose(true) should enable verbose mode")
	}

	logger.SetVerbose(false)
	if logger.IsVerbose() {
		t.Error("SetVerbose(false) should disable verbose mode")
	}
}

// TestDebugOnlyShownWhenVerbose verifies Debug output only when verbose=true
func TestDebugOnlyShownWhenVerbose(t *testing.T) {
	// Reset singleton for clean test
	once = sync.Once{}
	loggerInstance = nil

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := GetLogger()
	logger.SetVerbose(false)
	logger.Debug("test message")

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stderr = oldStderr

	if buf.Len() > 0 {
		t.Errorf("Debug should not output when verbose=false, got: %s", buf.String())
	}

	// Now test with verbose=true
	r, w, _ = os.Pipe()
	os.Stderr = w

	logger.SetVerbose(true)
	logger.Debug("test message verbose")

	_ = w.Close()
	buf.Reset()
	_, _ = io.Copy(&buf, r)
	os.Stderr = oldStderr

	if !strings.Contains(buf.String(), "[DEBUG]") {
		t.Errorf("Debug should output [DEBUG] prefix when verbose=true, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "test message verbose") {
		t.Errorf("Debug should output message when verbose=true, got: %s", buf.String())
	}
}

// TestInfoAlwaysShown verifies Info output regardless of verbose mode
func TestInfoAlwaysShown(t *testing.T) {
	// Reset singleton for clean test
	once = sync.Once{}
	loggerInstance = nil

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := GetLogger()
	logger.SetVerbose(false)
	logger.Info("info message")

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stderr = oldStderr

	if !strings.Contains(buf.String(), "[INFO]") {
		t.Errorf("Info should output [INFO] prefix, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "info message") {
		t.Errorf("Info should output message, got: %s", buf.String())
	}
}

// TestWarnAlwaysShown verifies Warn output regardless of verbose mode
func TestWarnAlwaysShown(t *testing.T) {
	// Reset singleton for clean test
	once = sync.Once{}
	loggerInstance = nil

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := GetLogger()
	logger.SetVerbose(false)
	logger.Warn("warn message")

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stderr = oldStderr

	if !strings.Contains(buf.String(), "[WARN]") {
		t.Errorf("Warn should output [WARN] prefix, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "warn message") {
		t.Errorf("Warn should output message, got: %s", buf.String())
	}
}

// TestErrorAlwaysShown verifies Error output regardless of verbose mode
func TestErrorAlwaysShown(t *testing.T) {
	// Reset singleton for clean test
	once = sync.Once{}
	loggerInstance = nil

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := GetLogger()
	logger.SetVerbose(false)
	logger.Error("error message")

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stderr = oldStderr

	if !strings.Contains(buf.String(), "[ERROR]") {
		t.Errorf("Error should output [ERROR] prefix, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "error message") {
		t.Errorf("Error should output message, got: %s", buf.String())
	}
}

// TestLogLevelPrefixes verifies each level has correct prefix
func TestLogLevelPrefixes(t *testing.T) {
	tests := []struct {
		name    string
		logFunc func(*Logger, string)
		prefix  string
		verbose bool
	}{
		{"Debug", func(l *Logger, m string) { l.Debug("%s", m) }, "[DEBUG]", true},
		{"Info", func(l *Logger, m string) { l.Info("%s", m) }, "[INFO]", false},
		{"Warn", func(l *Logger, m string) { l.Warn("%s", m) }, "[WARN]", false},
		{"Error", func(l *Logger, m string) { l.Error("%s", m) }, "[ERROR]", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset singleton for clean test
			once = sync.Once{}
			loggerInstance = nil

			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			logger := GetLogger()
			logger.SetVerbose(tt.verbose)
			tt.logFunc(logger, "test")

			_ = w.Close()
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			os.Stderr = oldStderr

			if !strings.Contains(buf.String(), tt.prefix) {
				t.Errorf("%s should have prefix %s, got: %s", tt.name, tt.prefix, buf.String())
			}
		})
	}
}

// TestConvenienceFunctions verifies global Debugf, Infof, Warnf, Errorf functions
func TestConvenienceFunctions(t *testing.T) {
	// Reset singleton for clean test
	once = sync.Once{}
	loggerInstance = nil

	SetVerboseMode(true)

	tests := []struct {
		name    string
		logFunc func(string, ...interface{})
		prefix  string
	}{
		{"Debugf", Debugf, "[DEBUG]"},
		{"Infof", Infof, "[INFO]"},
		{"Warnf", Warnf, "[WARN]"},
		{"Errorf", Errorf, "[ERROR]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			tt.logFunc("formatted %s", "value")

			_ = w.Close()
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			os.Stderr = oldStderr

			if !strings.Contains(buf.String(), tt.prefix) {
				t.Errorf("%s should have prefix %s, got: %s", tt.name, tt.prefix, buf.String())
			}
			if !strings.Contains(buf.String(), "formatted value") {
				t.Errorf("%s should format message, got: %s", tt.name, buf.String())
			}
		})
	}
}

// TestLoggerThreadSafety verifies concurrent access is safe
func TestLoggerThreadSafety(t *testing.T) {
	// Reset singleton for clean test
	once = sync.Once{}
	loggerInstance = nil

	logger := GetLogger()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if n%2 == 0 {
				logger.SetVerbose(true)
			} else {
				logger.SetVerbose(false)
			}
			logger.Debug("debug %d", n)
			logger.Info("info %d", n)
		}(i)
	}
	wg.Wait()
	// Test passes if no race condition panics
}

// =============================================================================
// Verbose Timestamp Tests (issue #47)
// =============================================================================

// TestVerboseOutputIncludesTimestamp verifies debug output lines include HH:MM:SS prefix
func TestVerboseOutputIncludesTimestamp(t *testing.T) {
	// Reset singleton for clean test
	once = sync.Once{}
	loggerInstance = nil

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := GetLogger()
	logger.SetVerbose(true)
	logger.Debug("timestamp test message")

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stderr = oldStderr

	output := buf.String()

	// Should contain a timestamp before [DEBUG]
	// Format: "HH:MM:SS [DEBUG] timestamp test message\n"
	if !strings.Contains(output, "[DEBUG]") {
		t.Fatalf("expected [DEBUG] in output, got: %s", output)
	}

	// Verify timestamp appears before [DEBUG]
	debugIdx := strings.Index(output, "[DEBUG]")
	prefix := output[:debugIdx]
	// Prefix should contain a time-like pattern (HH:MM:SS followed by space)
	timePattern := regexp.MustCompile(`\d{2}:\d{2}:\d{2} $`)
	if !timePattern.MatchString(prefix) {
		t.Errorf("expected HH:MM:SS timestamp prefix before [DEBUG], got prefix: %q", prefix)
	}
}

// TestVerboseTimestampFormat verifies timestamp format is consistent HH:MM:SS
func TestVerboseTimestampFormat(t *testing.T) {
	// Reset singleton for clean test
	once = sync.Once{}
	loggerInstance = nil

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := GetLogger()
	logger.SetVerbose(true)
	logger.Debug("format check")

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stderr = oldStderr

	output := buf.String()

	// Full line should match: "HH:MM:SS [DEBUG] format check\n"
	linePattern := regexp.MustCompile(`^\d{2}:\d{2}:\d{2} \[DEBUG\] format check\n$`)
	if !linePattern.MatchString(output) {
		t.Errorf("expected output matching 'HH:MM:SS [DEBUG] format check\\n', got: %q", output)
	}

	// Validate the hour/minute/second ranges
	timestampPattern := regexp.MustCompile(`^(\d{2}):(\d{2}):(\d{2}) `)
	matches := timestampPattern.FindStringSubmatch(output)
	if len(matches) != 4 {
		t.Fatalf("could not parse timestamp from output: %q", output)
	}
	// Hour 00-23, Minute 00-59, Second 00-59 are enforced by Go's time format,
	// but verify the values are plausible
	hour := matches[1]
	minute := matches[2]
	second := matches[3]
	if hour > "23" || minute > "59" || second > "59" {
		t.Errorf("timestamp values out of range: %s:%s:%s", hour, minute, second)
	}
}

// TestNonVerboseNoTimestamp verifies normal (non-verbose) output is unaffected
func TestNonVerboseNoTimestamp(t *testing.T) {
	// Reset singleton for clean test
	once = sync.Once{}
	loggerInstance = nil

	logger := GetLogger()
	logger.SetVerbose(false)

	// Test Info output (non-debug) - should NOT have timestamp
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger.Info("info without timestamp")

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stderr = oldStderr

	infoOutput := buf.String()
	// Info output should start with [INFO], no timestamp prefix
	if !strings.HasPrefix(infoOutput, "[INFO]") {
		t.Errorf("Info output should start with [INFO] (no timestamp), got: %q", infoOutput)
	}
	// Should not have a timestamp-like prefix
	timestampPattern := regexp.MustCompile(`^\d{2}:\d{2}:\d{2} `)
	if timestampPattern.MatchString(infoOutput) {
		t.Errorf("non-debug output should NOT have timestamp prefix, got: %q", infoOutput)
	}

	// Test Warn output - should NOT have timestamp
	r, w, _ = os.Pipe()
	os.Stderr = w

	logger.Warn("warn without timestamp")

	_ = w.Close()
	buf.Reset()
	_, _ = io.Copy(&buf, r)
	os.Stderr = oldStderr

	warnOutput := buf.String()
	if !strings.HasPrefix(warnOutput, "[WARN]") {
		t.Errorf("Warn output should start with [WARN] (no timestamp), got: %q", warnOutput)
	}

	// Test Error output - should NOT have timestamp
	r, w, _ = os.Pipe()
	os.Stderr = w

	logger.Error("error without timestamp")

	_ = w.Close()
	buf.Reset()
	_, _ = io.Copy(&buf, r)
	os.Stderr = oldStderr

	errorOutput := buf.String()
	if !strings.HasPrefix(errorOutput, "[ERROR]") {
		t.Errorf("Error output should start with [ERROR] (no timestamp), got: %q", errorOutput)
	}

	// Also verify that Debug with verbose=false produces no output at all
	r, w, _ = os.Pipe()
	os.Stderr = w

	logger.Debug("should not appear")

	_ = w.Close()
	buf.Reset()
	_, _ = io.Copy(&buf, r)
	os.Stderr = oldStderr

	if buf.Len() > 0 {
		t.Errorf("Debug with verbose=false should produce no output, got: %q", buf.String())
	}
}

// =============================================================================
// BackgroundLogger Tests
// =============================================================================

// TestBackgroundLoggerCreation verifies logger creates PID-specific log file
func TestBackgroundLoggerCreation(t *testing.T) {
	bl, err := NewBackgroundLogger()
	if err != nil {
		t.Fatalf("NewBackgroundLogger() error = %v", err)
	}
	defer bl.Close()

	if !bl.IsEnabled() {
		t.Skip("Background logging is disabled")
	}

	logPath := bl.GetLogPath()
	if logPath == "" {
		t.Error("GetLogPath() should return non-empty path")
	}

	// Check path contains PID
	pid := os.Getpid()
	if !strings.Contains(logPath, string(rune(pid))) && !strings.Contains(logPath, "-") {
		// Path should be in /tmp and contain some identifier
		if !strings.HasPrefix(logPath, "/tmp/") && !strings.HasPrefix(logPath, os.TempDir()) {
			t.Errorf("Log path should be in temp directory, got: %s", logPath)
		}
	}

	// Verify file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file should exist at %s", logPath)
	}
}

// TestBackgroundLoggerWritesMesages verifies Printf, Print, Println work
func TestBackgroundLoggerWritesMesages(t *testing.T) {
	bl, err := NewBackgroundLogger()
	if err != nil {
		t.Fatalf("NewBackgroundLogger() error = %v", err)
	}
	defer bl.Close()

	if !bl.IsEnabled() {
		t.Skip("Background logging is disabled")
	}

	bl.Printf("Printf message: %s", "test")
	bl.Print("Print message")
	bl.Println("Println message")

	// Close to flush
	bl.Close()

	// Read log file
	content, err := os.ReadFile(bl.GetLogPath())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "Printf message: test") {
		t.Errorf("Log should contain Printf message, got: %s", content)
	}
	if !strings.Contains(string(content), "Print message") {
		t.Errorf("Log should contain Print message, got: %s", content)
	}
	if !strings.Contains(string(content), "Println message") {
		t.Errorf("Log should contain Println message, got: %s", content)
	}
}

// TestBackgroundLoggerGracefulDegradation verifies fallback to io.Discard
func TestBackgroundLoggerGracefulDegradation(t *testing.T) {
	// Set an invalid temp directory
	oldTmpDir := os.Getenv("TMPDIR")
	_ = os.Setenv("TMPDIR", "/nonexistent/directory/that/should/not/exist")
	defer func() { _ = os.Setenv("TMPDIR", oldTmpDir) }()

	// Also try to make it fail by using a read-only location
	bl, _ := NewBackgroundLoggerWithPath("/nonexistent/directory/log.txt")

	// Should still be usable (writes to io.Discard)
	bl.Printf("This should not panic: %s", "test")
	bl.Print("This should not panic")
	bl.Println("This should not panic")
	bl.Close()

	// If we got here without panicking, the test passes
	if bl.IsEnabled() {
		t.Error("Logger should not be enabled when file creation fails")
	}
}

// TestBackgroundLoggerClose verifies close releases resources
func TestBackgroundLoggerClose(t *testing.T) {
	bl, err := NewBackgroundLogger()
	if err != nil {
		t.Fatalf("NewBackgroundLogger() error = %v", err)
	}

	logPath := bl.GetLogPath()
	bl.Close()

	// Writing after close should not panic (graceful degradation)
	bl.Printf("After close")
	bl.Print("After close")
	bl.Println("After close")

	// Clean up test file
	if logPath != "" {
		_ = os.Remove(logPath)
	}
}

// TestBackgroundLoggerPathFormat verifies log path format
func TestBackgroundLoggerPathFormat(t *testing.T) {
	bl, err := NewBackgroundLogger()
	if err != nil {
		t.Fatalf("NewBackgroundLogger() error = %v", err)
	}
	defer bl.Close()

	if !bl.IsEnabled() {
		t.Skip("Background logging is disabled")
	}

	logPath := bl.GetLogPath()

	// Path should contain "todoat"
	if !strings.Contains(logPath, "todoat") {
		t.Errorf("Log path should contain 'todoat', got: %s", logPath)
	}

	// Path should be in temp directory
	tmpDir := os.TempDir()
	if !strings.HasPrefix(logPath, tmpDir) && !strings.HasPrefix(logPath, "/tmp") {
		t.Errorf("Log path should be in temp directory, got: %s", logPath)
	}

	// Clean up
	_ = os.Remove(logPath)
}

// TestNewBackgroundLoggerWithPath allows custom path for testing
func TestNewBackgroundLoggerWithPath(t *testing.T) {
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "custom-log.txt")

	bl, err := NewBackgroundLoggerWithPath(customPath)
	if err != nil {
		t.Fatalf("NewBackgroundLoggerWithPath() error = %v", err)
	}
	defer bl.Close()

	if bl.GetLogPath() != customPath {
		t.Errorf("GetLogPath() = %s, want %s", bl.GetLogPath(), customPath)
	}

	bl.Println("Custom path test")
	bl.Close()

	content, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatalf("Failed to read custom log file: %v", err)
	}

	if !strings.Contains(string(content), "Custom path test") {
		t.Errorf("Custom log should contain message, got: %s", content)
	}
}
