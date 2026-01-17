package sqlite

import (
	"context"
	"todoat/backend"
)

// Backend implements backend.TaskManager using SQLite
type Backend struct {
	path string
}

// New creates a new SQLite backend
func New(path string) (*Backend, error) {
	return &Backend{path: path}, nil
}

// GetLists returns all task lists
func (b *Backend) GetLists(ctx context.Context) ([]backend.List, error) {
	// TODO: implement
	return nil, nil
}

// GetList returns a specific list by ID
func (b *Backend) GetList(ctx context.Context, listID string) (*backend.List, error) {
	// TODO: implement
	return nil, nil
}

// CreateList creates a new task list
func (b *Backend) CreateList(ctx context.Context, name string) (*backend.List, error) {
	// TODO: implement
	return nil, nil
}

// DeleteList removes a task list
func (b *Backend) DeleteList(ctx context.Context, listID string) error {
	// TODO: implement
	return nil
}

// GetTasks returns all tasks in a list
func (b *Backend) GetTasks(ctx context.Context, listID string) ([]backend.Task, error) {
	// TODO: implement
	return nil, nil
}

// GetTask returns a specific task
func (b *Backend) GetTask(ctx context.Context, listID, taskID string) (*backend.Task, error) {
	// TODO: implement
	return nil, nil
}

// CreateTask adds a new task to a list
func (b *Backend) CreateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	// TODO: implement
	return nil, nil
}

// UpdateTask modifies an existing task
func (b *Backend) UpdateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	// TODO: implement
	return nil, nil
}

// DeleteTask removes a task
func (b *Backend) DeleteTask(ctx context.Context, listID, taskID string) error {
	// TODO: implement
	return nil
}

// Close closes the database connection
func (b *Backend) Close() error {
	// TODO: implement
	return nil
}

// Verify interface compliance at compile time
var _ backend.TaskManager = (*Backend)(nil)
