// Package analytics provides local SQLite-based analytics for tracking command
// usage, success rates, and backend performance.
package analytics

import "os"

// Event represents a single analytics event
type Event struct {
	ID         int64
	Timestamp  int64
	Command    string
	Subcommand string
	Backend    string
	Success    bool
	DurationMs int64
	ErrorType  string
	Flags      string // JSON string of flags
}

// IsEnabledFromEnv checks the TODOAT_ANALYTICS_ENABLED environment variable
// and returns the effective enabled state. Environment variable overrides the
// config value.
func IsEnabledFromEnv(configEnabled bool) bool {
	envVal := os.Getenv("TODOAT_ANALYTICS_ENABLED")
	if envVal == "" {
		return configEnabled
	}
	return envVal == "true" || envVal == "1"
}
