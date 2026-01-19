package backend

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Task represents a todo item
type Task struct {
	ID          string
	Summary     string
	Description string
	Status      TaskStatus
	Priority    int
	DueDate     *time.Time
	StartDate   *time.Time
	Completed   *time.Time
	Created     time.Time
	Modified    time.Time
	ListID      string
	ParentID    string // For subtasks
	Categories  string // Comma-separated list of tags/categories
}

// TaskStatus represents the completion state of a task
type TaskStatus string

const (
	StatusNeedsAction TaskStatus = "NEEDS-ACTION"
	StatusCompleted   TaskStatus = "COMPLETED"
	StatusInProgress  TaskStatus = "IN-PROGRESS"
	StatusCancelled   TaskStatus = "CANCELLED"
)

// List represents a task list
type List struct {
	ID          string
	Name        string
	Color       string
	Description string
	Modified    time.Time
	DeletedAt   *time.Time // nil if not deleted, timestamp if in trash
}

// TaskManager defines the interface for task storage backends
type TaskManager interface {
	// List operations
	GetLists(ctx context.Context) ([]List, error)
	GetList(ctx context.Context, listID string) (*List, error)
	GetListByName(ctx context.Context, name string) (*List, error)
	CreateList(ctx context.Context, name string) (*List, error)
	UpdateList(ctx context.Context, list *List) (*List, error)
	DeleteList(ctx context.Context, listID string) error // Soft-delete (move to trash)

	// Trash operations
	GetDeletedLists(ctx context.Context) ([]List, error)
	GetDeletedListByName(ctx context.Context, name string) (*List, error)
	RestoreList(ctx context.Context, listID string) error
	PurgeList(ctx context.Context, listID string) error // Permanent delete

	// Task operations
	GetTasks(ctx context.Context, listID string) ([]Task, error)
	GetTask(ctx context.Context, listID, taskID string) (*Task, error)
	CreateTask(ctx context.Context, listID string, task *Task) (*Task, error)
	UpdateTask(ctx context.Context, listID string, task *Task) (*Task, error)
	DeleteTask(ctx context.Context, listID, taskID string) error

	// Connection management
	Close() error
}

// FindListByName searches for a list by name (case-insensitive) in a slice of lists.
// Returns nil if no match is found. This helper reduces code duplication across backends.
func FindListByName(lists []List, name string) *List {
	for _, l := range lists {
		if strings.EqualFold(l.Name, name) {
			return &l
		}
	}
	return nil
}

// GenerateID generates a unique identifier using UUID v4.
// This is used by backends that need to generate task/list IDs locally.
func GenerateID() string {
	return uuid.New().String()
}
