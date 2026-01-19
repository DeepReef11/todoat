// Package cache provides list metadata caching functionality.
package cache

import (
	"time"
)

// CachedList represents a cached task list with metadata.
type CachedList struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color,omitempty"`
	TaskCount   int       `json:"task_count"`
	Modified    time.Time `json:"modified"`
}

// ListCache represents the cached list metadata structure.
type ListCache struct {
	CreatedAt time.Time    `json:"created_at"`
	Backend   string       `json:"backend"`
	Lists     []CachedList `json:"lists"`
}
