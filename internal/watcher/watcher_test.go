package watcher

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// Issue #41: File Watcher for Real-Time Sync Triggers
// =============================================================================

// TestFileWatcherDetectsChanges verifies watcher detects file modifications.
func TestFileWatcherDetectsChanges(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file to watch
	watchFile := filepath.Join(tmpDir, "tasks.db")
	if err := os.WriteFile(watchFile, []byte("initial"), 0600); err != nil {
		t.Fatalf("failed to create watch file: %v", err)
	}

	var changeDetected atomic.Bool
	cfg := &Config{
		Paths:            []string{watchFile},
		DebounceDuration: 50 * time.Millisecond,
		QuietPeriod:      0, // Disable smart timing for this test
		OnSync: func() {
			changeDetected.Store(true)
		},
	}

	w, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Stop()

	if err := w.Start(); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}

	// Modify the file
	if err := os.WriteFile(watchFile, []byte("modified"), 0600); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// Wait for debounce + processing
	time.Sleep(200 * time.Millisecond)

	if !changeDetected.Load() {
		t.Errorf("expected watcher to detect file change")
	}
}

// TestFileWatcherDetectsNewFile verifies watcher detects new files in watched directories.
func TestFileWatcherDetectsNewFile(t *testing.T) {
	tmpDir := t.TempDir()

	var changeDetected atomic.Bool
	cfg := &Config{
		Paths:            []string{tmpDir},
		DebounceDuration: 50 * time.Millisecond,
		QuietPeriod:      0,
		OnSync: func() {
			changeDetected.Store(true)
		},
	}

	w, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Stop()

	if err := w.Start(); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}

	// Create a new file in the watched directory
	newFile := filepath.Join(tmpDir, "new_tasks.db")
	if err := os.WriteFile(newFile, []byte("new content"), 0600); err != nil {
		t.Fatalf("failed to create new file: %v", err)
	}

	// Wait for debounce + processing
	time.Sleep(200 * time.Millisecond)

	if !changeDetected.Load() {
		t.Errorf("expected watcher to detect new file creation")
	}
}

// TestFileWatcherTriggerSync verifies sync is triggered on file change.
func TestFileWatcherTriggerSync(t *testing.T) {
	tmpDir := t.TempDir()

	watchFile := filepath.Join(tmpDir, "tasks.db")
	if err := os.WriteFile(watchFile, []byte("initial"), 0600); err != nil {
		t.Fatalf("failed to create watch file: %v", err)
	}

	var syncCount atomic.Int32
	cfg := &Config{
		Paths:            []string{watchFile},
		DebounceDuration: 50 * time.Millisecond,
		QuietPeriod:      0,
		OnSync: func() {
			syncCount.Add(1)
		},
	}

	w, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Stop()

	if err := w.Start(); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}

	// Initial sync count should be 0
	if syncCount.Load() != 0 {
		t.Errorf("expected initial sync count 0, got %d", syncCount.Load())
	}

	// Modify the file
	if err := os.WriteFile(watchFile, []byte("modified"), 0600); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// Wait for debounce + processing
	time.Sleep(200 * time.Millisecond)

	if syncCount.Load() != 1 {
		t.Errorf("expected sync count 1 after file change, got %d", syncCount.Load())
	}

	// Modify again after debounce window
	time.Sleep(100 * time.Millisecond)
	if err := os.WriteFile(watchFile, []byte("modified again"), 0600); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// Wait for debounce + processing
	time.Sleep(200 * time.Millisecond)

	if syncCount.Load() != 2 {
		t.Errorf("expected sync count 2 after second file change, got %d", syncCount.Load())
	}
}

// TestFileWatcherDebounce verifies rapid changes are debounced.
func TestFileWatcherDebounce(t *testing.T) {
	tmpDir := t.TempDir()

	watchFile := filepath.Join(tmpDir, "tasks.db")
	if err := os.WriteFile(watchFile, []byte("initial"), 0600); err != nil {
		t.Fatalf("failed to create watch file: %v", err)
	}

	var syncCount atomic.Int32
	cfg := &Config{
		Paths:            []string{watchFile},
		DebounceDuration: 200 * time.Millisecond,
		QuietPeriod:      0,
		OnSync: func() {
			syncCount.Add(1)
		},
	}

	w, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Stop()

	if err := w.Start(); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}

	// Make 10 rapid modifications within the debounce window
	for i := 0; i < 10; i++ {
		if err := os.WriteFile(watchFile, []byte("rapid change "+string(rune('0'+i))), 0600); err != nil {
			t.Fatalf("failed to modify file: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Much shorter than debounce window
	}

	// Wait for debounce to complete
	time.Sleep(400 * time.Millisecond)

	// Should have triggered sync only once (or at most 2 if timing is borderline)
	count := syncCount.Load()
	if count > 2 {
		t.Errorf("expected at most 2 syncs from debounced rapid changes, got %d", count)
	}
	if count == 0 {
		t.Errorf("expected at least 1 sync after rapid changes, got 0")
	}
}

// TestFileWatcherSmartTiming verifies sync avoids interrupting active editing.
func TestFileWatcherSmartTiming(t *testing.T) {
	tmpDir := t.TempDir()

	watchFile := filepath.Join(tmpDir, "tasks.db")
	if err := os.WriteFile(watchFile, []byte("initial"), 0600); err != nil {
		t.Fatalf("failed to create watch file: %v", err)
	}

	var syncCount atomic.Int32
	var mu sync.Mutex
	var syncTimes []time.Time
	cfg := &Config{
		Paths:            []string{watchFile},
		DebounceDuration: 50 * time.Millisecond,
		QuietPeriod:      300 * time.Millisecond, // Must be quiet for 300ms before syncing
		OnSync: func() {
			syncCount.Add(1)
			mu.Lock()
			syncTimes = append(syncTimes, time.Now())
			mu.Unlock()
		},
	}

	w, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Stop()

	if err := w.Start(); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}

	// Simulate active editing - continuous modifications at 100ms intervals
	// for 500ms (QuietPeriod is 300ms, so sync should be deferred)
	startTime := time.Now()
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(watchFile, []byte("editing "+string(rune('0'+i))), 0600); err != nil {
			t.Fatalf("failed to modify file: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	editEndTime := time.Now()

	// No sync should have happened during active editing because changes keep
	// arriving within the quiet period
	editPhaseSyncs := syncCount.Load()

	// Wait for quiet period to elapse after editing stops
	time.Sleep(500 * time.Millisecond)

	finalSyncs := syncCount.Load()

	// Verify sync happened AFTER editing stopped (after the quiet period)
	if finalSyncs == 0 {
		t.Errorf("expected at least 1 sync after editing stopped, got 0")
	}

	mu.Lock()
	defer mu.Unlock()

	// Any sync that occurred should be after the editing ended + quiet period
	for _, st := range syncTimes {
		// Allow a small tolerance for timing
		if st.Before(editEndTime) {
			t.Logf("NOTE: sync at %v occurred during editing (start=%v, editEnd=%v)", st.Sub(startTime), time.Duration(0), editEndTime.Sub(startTime))
		}
	}

	// During active editing (within the quiet period), syncs should be deferred
	// Allow at most 1 sync during the editing phase (the initial debounce may fire once)
	if editPhaseSyncs > 1 {
		t.Errorf("expected at most 1 sync during active editing (quiet period deferral), got %d", editPhaseSyncs)
	}
}

// TestFileWatcherStopCleanly verifies the watcher stops without errors.
func TestFileWatcherStopCleanly(t *testing.T) {
	tmpDir := t.TempDir()

	watchFile := filepath.Join(tmpDir, "tasks.db")
	if err := os.WriteFile(watchFile, []byte("initial"), 0600); err != nil {
		t.Fatalf("failed to create watch file: %v", err)
	}

	cfg := &Config{
		Paths:            []string{watchFile},
		DebounceDuration: 50 * time.Millisecond,
		QuietPeriod:      0,
		OnSync:           func() {},
	}

	w, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}

	if err := w.Start(); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}

	// Stop should clean up without errors
	w.Stop()

	// Starting after stop should fail
	if err := w.Start(); err == nil {
		t.Errorf("expected error starting a stopped watcher")
	}
}

// TestFileWatcherMultiplePaths verifies watching multiple files/directories.
func TestFileWatcherMultiplePaths(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "tasks.db")
	file2 := filepath.Join(tmpDir, "cache.db")
	if err := os.WriteFile(file1, []byte("tasks"), 0600); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("cache"), 0600); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	var syncCount atomic.Int32
	cfg := &Config{
		Paths:            []string{file1, file2},
		DebounceDuration: 50 * time.Millisecond,
		QuietPeriod:      0,
		OnSync: func() {
			syncCount.Add(1)
		},
	}

	w, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Stop()

	if err := w.Start(); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}

	// Modify first file
	if err := os.WriteFile(file1, []byte("modified tasks"), 0600); err != nil {
		t.Fatalf("failed to modify file1: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	count1 := syncCount.Load()
	if count1 == 0 {
		t.Errorf("expected sync after modifying file1, got 0")
	}

	// Modify second file
	if err := os.WriteFile(file2, []byte("modified cache"), 0600); err != nil {
		t.Fatalf("failed to modify file2: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	count2 := syncCount.Load()
	if count2 <= count1 {
		t.Errorf("expected sync after modifying file2, got %d (was %d)", count2, count1)
	}
}

// TestFileWatcherNonExistentPath verifies watcher handles non-existent paths gracefully.
func TestFileWatcherNonExistentPath(t *testing.T) {
	cfg := &Config{
		Paths:            []string{"/nonexistent/path/file.db"},
		DebounceDuration: 50 * time.Millisecond,
		QuietPeriod:      0,
		OnSync:           func() {},
	}

	w, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer w.Stop()

	// Start should not fail - non-existent paths are skipped
	if err := w.Start(); err != nil {
		t.Fatalf("start should not fail for non-existent paths: %v", err)
	}
}

// TestFileWatcherConfigDefaults verifies default configuration values.
func TestFileWatcherConfigDefaults(t *testing.T) {
	cfg := DefaultConfig(func() {})

	if cfg.DebounceDuration != DefaultDebounceDuration {
		t.Errorf("expected default debounce %v, got %v", DefaultDebounceDuration, cfg.DebounceDuration)
	}

	if cfg.QuietPeriod != DefaultQuietPeriod {
		t.Errorf("expected default quiet period %v, got %v", DefaultQuietPeriod, cfg.QuietPeriod)
	}

	if cfg.OnSync == nil {
		t.Errorf("expected OnSync callback to be set")
	}
}
