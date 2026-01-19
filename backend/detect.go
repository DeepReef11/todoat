package backend

import (
	"sort"
	"sync"
)

// DetectableBackend is an interface for backends that support auto-detection
// based on the current directory context.
type DetectableBackend interface {
	TaskManager

	// CanDetect checks if this backend can be used in the current environment.
	// It should be fast (<100ms) and non-destructive (read-only operations).
	// Returns true if the backend is usable, false otherwise.
	CanDetect() (bool, error)

	// DetectionInfo returns human-readable information about the detected environment.
	// This is shown to users when displaying detection results.
	DetectionInfo() string
}

// DetectableConstructor creates a DetectableBackend for a given working directory.
type DetectableConstructor func(workDir string) (DetectableBackend, error)

// DetectionResult holds the result of backend detection for a single backend.
type DetectionResult struct {
	Name      string // Backend name (e.g., "git", "sqlite")
	Available bool   // Whether the backend can be used
	Info      string // Human-readable detection info
	Priority  int    // Lower number = higher priority (0 = highest)
	Backend   DetectableBackend
}

// detectableRegistration holds a constructor with its priority
type detectableRegistration struct {
	constructor DetectableConstructor
	priority    int
}

// Global registry for detectable backends
var (
	detectableMu            sync.RWMutex
	detectableConstructors  = make(map[string]DetectableConstructor)
	detectableRegistrations = make(map[string]detectableRegistration)
)

// RegisterDetectable registers a detectable backend constructor.
// Backends should call this in their init() function.
func RegisterDetectable(name string, constructor DetectableConstructor) {
	RegisterDetectableWithPriority(name, constructor, 100) // Default priority
}

// RegisterDetectableWithPriority registers a detectable backend constructor with a priority.
// Lower priority numbers are preferred (git=10, sqlite=100).
func RegisterDetectableWithPriority(name string, constructor DetectableConstructor, priority int) {
	detectableMu.Lock()
	defer detectableMu.Unlock()
	detectableConstructors[name] = constructor
	detectableRegistrations[name] = detectableRegistration{
		constructor: constructor,
		priority:    priority,
	}
}

// GetDetectableConstructors returns all registered detectable backend constructors.
func GetDetectableConstructors() map[string]DetectableConstructor {
	detectableMu.RLock()
	defer detectableMu.RUnlock()

	// Return a copy to prevent modification
	result := make(map[string]DetectableConstructor, len(detectableConstructors))
	for k, v := range detectableConstructors {
		result[k] = v
	}
	return result
}

// getDetectableRegistrations returns all registered backends with their priorities.
func getDetectableRegistrations() map[string]detectableRegistration {
	detectableMu.RLock()
	defer detectableMu.RUnlock()

	result := make(map[string]detectableRegistration, len(detectableRegistrations))
	for k, v := range detectableRegistrations {
		result[k] = v
	}
	return result
}

// ClearDetectableConstructors removes all registered constructors.
// This is primarily used for testing.
func ClearDetectableConstructors() {
	detectableMu.Lock()
	defer detectableMu.Unlock()
	detectableConstructors = make(map[string]DetectableConstructor)
	detectableRegistrations = make(map[string]detectableRegistration)
}

// DetectBackends runs detection for all registered detectable backends.
// Returns a list of detection results ordered by priority (lower number = higher priority).
func DetectBackends(workDir string) ([]DetectionResult, error) {
	registrations := getDetectableRegistrations()

	var results []DetectionResult

	for name, reg := range registrations {
		be, err := reg.constructor(workDir)
		if err != nil {
			// Construction error - backend not available
			results = append(results, DetectionResult{
				Name:      name,
				Available: false,
				Info:      "failed to initialize: " + err.Error(),
				Priority:  reg.priority,
			})
			continue
		}

		canDetect, err := be.CanDetect()
		if err != nil {
			_ = be.Close()
			results = append(results, DetectionResult{
				Name:      name,
				Available: false,
				Info:      "detection error: " + err.Error(),
				Priority:  reg.priority,
			})
			continue
		}

		if canDetect {
			results = append(results, DetectionResult{
				Name:      name,
				Available: true,
				Info:      be.DetectionInfo(),
				Priority:  reg.priority,
				Backend:   be,
			})
		} else {
			_ = be.Close()
			results = append(results, DetectionResult{
				Name:      name,
				Available: false,
				Info:      "not available in current context",
				Priority:  reg.priority,
			})
		}
	}

	// Sort by priority (lower number = higher priority)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Priority < results[j].Priority
	})

	return results, nil
}

// SelectDetectedBackend runs detection and returns the first available backend.
// Returns nil if no backend is available.
func SelectDetectedBackend(workDir string) (DetectableBackend, string, error) {
	results, err := DetectBackends(workDir)
	if err != nil {
		return nil, "", err
	}

	// Close backends we don't use
	var selected DetectableBackend
	var selectedName string

	for _, r := range results {
		if r.Available && r.Backend != nil {
			if selected == nil {
				selected = r.Backend
				selectedName = r.Name
			} else {
				// Close this one, we already have a selection
				_ = r.Backend.Close()
			}
		}
	}

	return selected, selectedName, nil
}
