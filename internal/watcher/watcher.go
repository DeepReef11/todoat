// Package watcher provides file system watching for real-time sync triggers.
// It monitors task database and cache files, triggering sync operations when
// changes are detected, with debouncing and smart timing to avoid interrupting
// active editing sessions.
package watcher

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	// DefaultDebounceDuration is the default debounce window for batching rapid changes.
	DefaultDebounceDuration = 1 * time.Second

	// DefaultQuietPeriod is the default quiet period before triggering sync.
	// If file modifications occur within this period, sync is deferred
	// to avoid interrupting active editing sessions.
	DefaultQuietPeriod = 2 * time.Second
)

// Config holds file watcher configuration.
type Config struct {
	Paths            []string      // Paths to watch (files or directories)
	DebounceDuration time.Duration // Debounce window to batch rapid changes
	QuietPeriod      time.Duration // Quiet period to detect active editing (0 = disabled)
	OnSync           func()        // Callback to trigger sync
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig(onSync func()) *Config {
	return &Config{
		DebounceDuration: DefaultDebounceDuration,
		QuietPeriod:      DefaultQuietPeriod,
		OnSync:           onSync,
	}
}

// Watcher monitors file system changes and triggers sync operations.
type Watcher struct {
	cfg     *Config
	fsw     *fsnotify.Watcher
	stopCh  chan struct{}
	stopped bool
	mu      sync.Mutex
}

// New creates a new Watcher instance.
func New(cfg *Config) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	return &Watcher{
		cfg:    cfg,
		fsw:    fsw,
		stopCh: make(chan struct{}),
	}, nil
}

// Start begins watching the configured paths.
func (w *Watcher) Start() error {
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return fmt.Errorf("watcher has been stopped and cannot be restarted")
	}
	w.mu.Unlock()

	// Add paths to the fsnotify watcher
	for _, path := range w.cfg.Paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Skip non-existent paths - they may be created later
			continue
		}
		if err := w.fsw.Add(path); err != nil {
			return fmt.Errorf("failed to watch path %q: %w", path, err)
		}
	}

	// Start the event processing loop
	go w.eventLoop()

	return nil
}

// Stop stops the watcher and cleans up resources.
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.stopped {
		return
	}
	w.stopped = true
	close(w.stopCh)
	_ = w.fsw.Close()
}

// eventLoop processes fsnotify events with debouncing and smart timing.
func (w *Watcher) eventLoop() {
	var debounceTimer *time.Timer
	var quietTimer *time.Timer
	var lastEvent time.Time

	// debounceCh fires when the debounce window expires
	debounceCh := make(chan struct{}, 1)
	// quietCh fires when the quiet period expires
	quietCh := make(chan struct{}, 1)

	resetDebounce := func() {
		if debounceTimer != nil {
			debounceTimer.Stop()
		}
		debounceTimer = time.AfterFunc(w.cfg.DebounceDuration, func() {
			select {
			case debounceCh <- struct{}{}:
			default:
			}
		})
	}

	resetQuiet := func() {
		if quietTimer != nil {
			quietTimer.Stop()
		}
		if w.cfg.QuietPeriod > 0 {
			quietTimer = time.AfterFunc(w.cfg.QuietPeriod, func() {
				select {
				case quietCh <- struct{}{}:
				default:
				}
			})
		}
	}

	pendingSync := false

	for {
		select {
		case <-w.stopCh:
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			if quietTimer != nil {
				quietTimer.Stop()
			}
			return

		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			// Only react to write, create, and rename events
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}

			lastEvent = time.Now()
			_ = lastEvent // Used by quiet period logic

			if w.cfg.QuietPeriod > 0 {
				// Smart timing mode: reset the quiet timer on each event.
				// Sync fires only after the quiet period elapses without new events.
				pendingSync = true
				resetQuiet()
			} else {
				// Simple debounce mode: reset the debounce timer on each event.
				resetDebounce()
			}

		case _, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			// Log errors but continue watching

		case <-debounceCh:
			// Debounce period elapsed - trigger sync
			if w.cfg.OnSync != nil {
				w.cfg.OnSync()
			}

		case <-quietCh:
			// Quiet period elapsed - no recent edits, safe to sync
			if pendingSync && w.cfg.OnSync != nil {
				w.cfg.OnSync()
				pendingSync = false
			}
		}
	}
}
