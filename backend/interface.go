package backend

import (
	"context"
	"time"
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
	ID       string
	Name     string
	Color    string
	Modified time.Time
}

// TaskManager defines the interface for task storage backends
type TaskManager interface {
	// List operations
	GetLists(ctx context.Context) ([]List, error)
	GetList(ctx context.Context, listID string) (*List, error)
	CreateList(ctx context.Context, name string) (*List, error)
	DeleteList(ctx context.Context, listID string) error

	// Task operations
	GetTasks(ctx context.Context, listID string) ([]Task, error)
	GetTask(ctx context.Context, listID, taskID string) (*Task, error)
	CreateTask(ctx context.Context, listID string, task *Task) (*Task, error)
	UpdateTask(ctx context.Context, listID string, task *Task) (*Task, error)
	DeleteTask(ctx context.Context, listID, taskID string) error

	// Connection management
	Close() error
}
