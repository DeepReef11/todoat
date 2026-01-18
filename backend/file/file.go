// Package file implements a TaskManager backend that stores tasks in plain text files.
package file

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"todoat/backend"
	"todoat/internal/markdown"
)

// Config holds file backend configuration
type Config struct {
	FilePath string // Path to task file
}

// Backend implements backend.TaskManager for file-based storage
type Backend struct {
	config     Config
	filePath   string // Resolved absolute path
	lists      []backend.List
	tasks      map[string][]backend.Task // listID -> tasks
	tasksByID  map[string]*backend.Task  // taskID -> task
	fileLoaded bool
}

// New creates a new file backend
func New(cfg Config) (*Backend, error) {
	filePath := cfg.FilePath
	if filePath == "" {
		filePath = "tasks.txt"
	}

	// Resolve relative paths
	if !filepath.IsAbs(filePath) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		filePath = filepath.Join(wd, filePath)
	}

	return &Backend{
		config:    cfg,
		filePath:  filePath,
		tasks:     make(map[string][]backend.Task),
		tasksByID: make(map[string]*backend.Task),
	}, nil
}

// Close closes the backend
func (b *Backend) Close() error {
	return nil
}

// =============================================================================
// List Operations
// =============================================================================

// GetLists returns all task lists (sections in the file)
func (b *Backend) GetLists(ctx context.Context) ([]backend.List, error) {
	if err := b.ensureLoaded(); err != nil {
		return nil, err
	}
	return b.lists, nil
}

// GetList returns a specific list by ID
func (b *Backend) GetList(ctx context.Context, listID string) (*backend.List, error) {
	if err := b.ensureLoaded(); err != nil {
		return nil, err
	}

	for i := range b.lists {
		if b.lists[i].ID == listID {
			return &b.lists[i], nil
		}
	}
	return nil, nil
}

// GetListByName returns a specific list by name (case-insensitive)
func (b *Backend) GetListByName(ctx context.Context, name string) (*backend.List, error) {
	if err := b.ensureLoaded(); err != nil {
		return nil, err
	}

	for i := range b.lists {
		if strings.EqualFold(b.lists[i].Name, name) {
			return &b.lists[i], nil
		}
	}
	return nil, nil
}

// CreateList creates a new list (adds a new section to the file)
func (b *Backend) CreateList(ctx context.Context, name string) (*backend.List, error) {
	if err := b.ensureLoaded(); err != nil {
		return nil, err
	}

	// Check if list already exists
	for i := range b.lists {
		if strings.EqualFold(b.lists[i].Name, name) {
			return &b.lists[i], nil
		}
	}

	// Create new list
	list := backend.List{
		ID:       generateID(),
		Name:     name,
		Modified: time.Now(),
	}
	b.lists = append(b.lists, list)
	b.tasks[list.ID] = []backend.Task{}

	// Save changes
	if err := b.saveFile(); err != nil {
		return nil, err
	}

	return &list, nil
}

// DeleteList removes a list and all its tasks
func (b *Backend) DeleteList(ctx context.Context, listID string) error {
	if err := b.ensureLoaded(); err != nil {
		return err
	}

	// Find and remove the list
	newLists := make([]backend.List, 0, len(b.lists))
	found := false
	for _, l := range b.lists {
		if l.ID == listID {
			found = true
			// Remove tasks for this list
			for _, task := range b.tasks[listID] {
				delete(b.tasksByID, task.ID)
			}
			delete(b.tasks, listID)
		} else {
			newLists = append(newLists, l)
		}
	}

	if !found {
		return fmt.Errorf("list not found: %s", listID)
	}

	b.lists = newLists

	// Save changes
	return b.saveFile()
}

// GetDeletedLists returns deleted lists (not supported in file backend)
func (b *Backend) GetDeletedLists(ctx context.Context) ([]backend.List, error) {
	return []backend.List{}, nil
}

// GetDeletedListByName returns a deleted list by name (not supported)
func (b *Backend) GetDeletedListByName(ctx context.Context, name string) (*backend.List, error) {
	return nil, nil
}

// RestoreList restores a deleted list (not supported)
func (b *Backend) RestoreList(ctx context.Context, listID string) error {
	return fmt.Errorf("restore not supported in file backend")
}

// PurgeList permanently deletes a list (not supported)
func (b *Backend) PurgeList(ctx context.Context, listID string) error {
	return fmt.Errorf("purge not supported in file backend")
}

// =============================================================================
// Task Operations
// =============================================================================

// GetTasks returns all tasks in a list
func (b *Backend) GetTasks(ctx context.Context, listID string) ([]backend.Task, error) {
	if err := b.ensureLoaded(); err != nil {
		return nil, err
	}

	tasks, ok := b.tasks[listID]
	if !ok {
		return []backend.Task{}, nil
	}
	return tasks, nil
}

// GetTask returns a specific task by ID
func (b *Backend) GetTask(ctx context.Context, listID, taskID string) (*backend.Task, error) {
	if err := b.ensureLoaded(); err != nil {
		return nil, err
	}

	task, ok := b.tasksByID[taskID]
	if !ok {
		return nil, nil
	}
	return task, nil
}

// CreateTask creates a new task
func (b *Backend) CreateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	if err := b.ensureLoaded(); err != nil {
		return nil, err
	}

	// Verify list exists
	var listExists bool
	for _, l := range b.lists {
		if l.ID == listID {
			listExists = true
			break
		}
	}
	if !listExists {
		return nil, fmt.Errorf("list not found: %s", listID)
	}

	// Create the task
	newTask := backend.Task{
		ID:          generateID(),
		Summary:     task.Summary,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		DueDate:     task.DueDate,
		StartDate:   task.StartDate,
		Categories:  task.Categories,
		ParentID:    task.ParentID,
		ListID:      listID,
		Created:     time.Now(),
		Modified:    time.Now(),
	}

	if newTask.Status == "" {
		newTask.Status = backend.StatusNeedsAction
	}

	// Add to data structures
	b.tasks[listID] = append(b.tasks[listID], newTask)
	b.tasksByID[newTask.ID] = &b.tasks[listID][len(b.tasks[listID])-1]

	// Save changes
	if err := b.saveFile(); err != nil {
		return nil, err
	}

	return &newTask, nil
}

// UpdateTask updates an existing task
func (b *Backend) UpdateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	if err := b.ensureLoaded(); err != nil {
		return nil, err
	}

	// Find and update the task
	existingTask, ok := b.tasksByID[task.ID]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", task.ID)
	}

	existingTask.Summary = task.Summary
	existingTask.Description = task.Description
	existingTask.Status = task.Status
	existingTask.Priority = task.Priority
	existingTask.DueDate = task.DueDate
	existingTask.StartDate = task.StartDate
	existingTask.Completed = task.Completed
	existingTask.Categories = task.Categories
	existingTask.ParentID = task.ParentID
	existingTask.Modified = time.Now()

	// Save changes
	if err := b.saveFile(); err != nil {
		return nil, err
	}

	return existingTask, nil
}

// DeleteTask deletes a task
func (b *Backend) DeleteTask(ctx context.Context, listID, taskID string) error {
	if err := b.ensureLoaded(); err != nil {
		return err
	}

	// Find the task
	_, ok := b.tasksByID[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Remove from task list
	tasks := b.tasks[listID]
	newTasks := make([]backend.Task, 0, len(tasks))
	for _, t := range tasks {
		if t.ID != taskID {
			newTasks = append(newTasks, t)
		}
	}
	b.tasks[listID] = newTasks

	// Remove from tasksByID
	delete(b.tasksByID, taskID)

	// Save changes
	return b.saveFile()
}

// =============================================================================
// File Operations
// =============================================================================

// ensureLoaded loads the file if not already loaded, or initializes empty state
func (b *Backend) ensureLoaded() error {
	if b.fileLoaded {
		return nil
	}

	// Check if file exists
	if _, err := os.Stat(b.filePath); os.IsNotExist(err) {
		// File doesn't exist - initialize with empty state
		b.lists = []backend.List{}
		b.tasks = make(map[string][]backend.Task)
		b.tasksByID = make(map[string]*backend.Task)
		b.fileLoaded = true
		return nil
	}

	return b.loadFile()
}

// loadFile parses the file into lists and tasks
func (b *Backend) loadFile() error {
	data, err := os.ReadFile(b.filePath)
	if err != nil {
		return err
	}

	b.lists = nil
	b.tasks = make(map[string][]backend.Task)
	b.tasksByID = make(map[string]*backend.Task)

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	var currentList *backend.List
	var parentStack []*taskWithIndent // Stack for tracking parent hierarchy

	// Regex patterns
	sectionPattern := regexp.MustCompile(`^##\s+(.+)$`)
	taskPattern := regexp.MustCompile(`^(\s*)-\s+\[([ xX~-])\]\s+(.*)$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for section header
		if matches := sectionPattern.FindStringSubmatch(line); len(matches) == 2 {
			listName := strings.TrimSpace(matches[1])
			list := backend.List{
				ID:       generateID(),
				Name:     listName,
				Modified: time.Now(),
			}
			b.lists = append(b.lists, list)
			b.tasks[list.ID] = []backend.Task{}
			currentList = &b.lists[len(b.lists)-1]
			parentStack = nil // Reset parent stack for new section
			continue
		}

		// Check for task item
		if matches := taskPattern.FindStringSubmatch(line); len(matches) == 4 && currentList != nil {
			indent := len(matches[1])
			statusChar := matches[2]
			taskText := strings.TrimSpace(matches[3])

			// Parse task text for metadata
			summary, priority, dueDate, categories := markdown.ParseTaskText(taskText)

			task := backend.Task{
				ID:         generateID(),
				Summary:    summary,
				Status:     markdown.ParseStatusChar(statusChar),
				Priority:   priority,
				DueDate:    dueDate,
				Categories: categories,
				ListID:     currentList.ID,
				Created:    time.Now(),
				Modified:   time.Now(),
			}

			// Handle hierarchy based on indentation
			// Pop parents that are at same or deeper indentation level
			for len(parentStack) > 0 && parentStack[len(parentStack)-1].indent >= indent {
				parentStack = parentStack[:len(parentStack)-1]
			}

			// Set parent if we have one
			if len(parentStack) > 0 {
				task.ParentID = parentStack[len(parentStack)-1].task.ID
			}

			// Add task to list
			b.tasks[currentList.ID] = append(b.tasks[currentList.ID], task)
			taskPtr := &b.tasks[currentList.ID][len(b.tasks[currentList.ID])-1]
			b.tasksByID[task.ID] = taskPtr

			// Push this task as potential parent
			parentStack = append(parentStack, &taskWithIndent{task: taskPtr, indent: indent})
		}
	}

	b.fileLoaded = true
	return nil
}

// taskWithIndent holds a task pointer and its indentation level
type taskWithIndent struct {
	task   *backend.Task
	indent int
}

// saveFile writes the lists and tasks back to the file
func (b *Backend) saveFile() error {
	// Ensure parent directory exists
	dir := filepath.Dir(b.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var sb strings.Builder

	// Write header
	sb.WriteString("# Tasks\n\n")

	// Write each list and its tasks
	for _, list := range b.lists {
		sb.WriteString("## ")
		sb.WriteString(list.Name)
		sb.WriteString("\n\n")

		// Organize tasks hierarchically
		rootTasks, childrenMap := markdown.OrganizeTasksHierarchically(b.tasks[list.ID])

		// Write tasks with proper indentation
		for i := range rootTasks {
			markdown.WriteTaskTree(&sb, &rootTasks[i], childrenMap, 0)
		}

		sb.WriteString("\n")
	}

	return os.WriteFile(b.filePath, []byte(sb.String()), 0644)
}

// =============================================================================
// Helpers
// =============================================================================

// generateID generates a unique identifier
func generateID() string {
	return uuid.New().String()
}

// Verify interface compliance at compile time
var _ backend.TaskManager = (*Backend)(nil)
