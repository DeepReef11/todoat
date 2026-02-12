// Package git implements a TaskManager backend that stores tasks in markdown files
// within Git repositories.
package git

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"todoat/backend"
	"todoat/internal/markdown"
)

const (
	// Marker required in markdown file for todoat to recognize it
	Marker = "<!-- todoat:enabled -->"

	// Default files to search for in order of priority
	defaultFile1 = "TODO.md"
	defaultFile2 = "todo.md"
	defaultFile3 = ".todoat.md"
)

// Config holds Git backend configuration
type Config struct {
	WorkDir       string   // Working directory (defaults to current directory)
	File          string   // Specific markdown file to use
	FallbackFiles []string // Files to try if configured file not found
	AutoCommit    bool     // Auto-commit changes
}

// Backend implements backend.TaskManager for Git/markdown storage
type Backend struct {
	config     Config
	repoPath   string // Root of git repository
	filePath   string // Path to the markdown file
	lists      []backend.List
	tasks      map[string][]backend.Task // listID -> tasks
	tasksByID  map[string]*backend.Task  // taskID -> task
	fileLoaded bool
}

// New creates a new Git backend
func New(cfg Config) (*Backend, error) {
	workDir := cfg.WorkDir
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		cfg.WorkDir = workDir
	}

	return &Backend{
		config:    cfg,
		tasks:     make(map[string][]backend.Task),
		tasksByID: make(map[string]*backend.Task),
	}, nil
}

// CanDetect checks if this backend can be used in the current environment
func (b *Backend) CanDetect() (bool, error) {
	// Find git repository
	repoPath, err := b.findGitRepo()
	if err != nil || repoPath == "" {
		return false, nil
	}
	b.repoPath = repoPath

	// Find marked TODO file
	filePath, err := b.findTodoFile()
	if err != nil || filePath == "" {
		return false, nil
	}
	b.filePath = filePath

	return true, nil
}

// DetectionInfo returns human-readable info about the detected environment
func (b *Backend) DetectionInfo() string {
	repoName := filepath.Base(b.repoPath)
	fileName := filepath.Base(b.filePath)
	return fmt.Sprintf("Git repository at %s with task file %s", repoName, fileName)
}

// findGitRepo walks up the directory tree to find a .git directory
func (b *Backend) findGitRepo() (string, error) {
	workDir := b.config.WorkDir
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	dir := workDir
	for {
		gitPath := filepath.Join(dir, ".git")
		info, err := os.Stat(gitPath)
		if err == nil && (info.IsDir() || info.Mode().IsRegular()) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding .git
			return "", nil
		}
		dir = parent
	}
}

// findTodoFile searches for a markdown file with the todoat marker
func (b *Backend) findTodoFile() (string, error) {
	if b.repoPath == "" {
		repoPath, err := b.findGitRepo()
		if err != nil || repoPath == "" {
			return "", nil
		}
		b.repoPath = repoPath
	}

	// Build list of files to check
	var filesToCheck []string

	// 1. Configured file
	if b.config.File != "" {
		filesToCheck = append(filesToCheck, b.config.File)
	}

	// 2. Fallback files from config
	filesToCheck = append(filesToCheck, b.config.FallbackFiles...)

	// 3. Default files
	filesToCheck = append(filesToCheck, defaultFile1, defaultFile2, defaultFile3)

	// Check each file for marker
	for _, fileName := range filesToCheck {
		filePath := filepath.Join(b.repoPath, fileName)
		if hasMarker(filePath) {
			return filePath, nil
		}
	}

	return "", nil
}

// hasMarker checks if a file contains the todoat marker
func hasMarker(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), Marker)
}

// Close closes the backend
func (b *Backend) Close() error {
	return nil
}

// =============================================================================
// List Operations
// =============================================================================

// GetLists returns all task lists (sections in the markdown file)
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

	for _, l := range b.lists {
		if l.ID == listID {
			return &l, nil
		}
	}
	return nil, nil
}

// GetListByName returns a specific list by name (case-insensitive)
func (b *Backend) GetListByName(ctx context.Context, name string) (*backend.List, error) {
	if err := b.ensureLoaded(); err != nil {
		return nil, err
	}

	for _, l := range b.lists {
		if strings.EqualFold(l.Name, name) {
			return &l, nil
		}
	}
	return nil, nil
}

// CreateList creates a new list (adds a new section to the markdown file)
func (b *Backend) CreateList(ctx context.Context, name string) (*backend.List, error) {
	if err := b.ensureLoaded(); err != nil {
		return nil, err
	}

	// Check if list already exists
	for _, l := range b.lists {
		if strings.EqualFold(l.Name, name) {
			return &l, nil
		}
	}

	// Create new list
	list := backend.List{
		ID:       backend.GenerateID(),
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

// UpdateList updates a list's properties
func (b *Backend) UpdateList(ctx context.Context, list *backend.List) (*backend.List, error) {
	if err := b.ensureLoaded(); err != nil {
		return nil, err
	}

	// Find and update the list
	for i, l := range b.lists {
		if l.ID == list.ID {
			b.lists[i].Name = list.Name
			b.lists[i].Color = list.Color
			b.lists[i].Modified = time.Now()

			// Save changes
			if err := b.saveFile(); err != nil {
				return nil, err
			}

			return &b.lists[i], nil
		}
	}

	return nil, fmt.Errorf("list not found")
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

// GetDeletedLists returns deleted lists (not supported in Git backend)
func (b *Backend) GetDeletedLists(ctx context.Context) ([]backend.List, error) {
	return []backend.List{}, nil
}

// GetDeletedListByName returns a deleted list by name (not supported)
func (b *Backend) GetDeletedListByName(ctx context.Context, name string) (*backend.List, error) {
	return nil, nil
}

// RestoreList restores a deleted list (not supported)
func (b *Backend) RestoreList(ctx context.Context, listID string) error {
	return fmt.Errorf("restore not supported in Git backend")
}

// PurgeList permanently deletes a list (not supported - deletion is already permanent)
func (b *Backend) PurgeList(ctx context.Context, listID string) error {
	return fmt.Errorf("purge not supported in Git backend")
}

// SupportsTrash returns false because Git backend deletion is permanent.
func (b *Backend) SupportsTrash() bool { return false }

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
		ID:          backend.GenerateID(),
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

	// Auto-commit if enabled
	if b.config.AutoCommit {
		_ = b.autoCommit(fmt.Sprintf("todoat: add task '%s'", newTask.Summary))
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

	// Auto-commit if enabled
	if b.config.AutoCommit {
		_ = b.autoCommit(fmt.Sprintf("todoat: update task '%s'", existingTask.Summary))
	}

	return existingTask, nil
}

// DeleteTask deletes a task
func (b *Backend) DeleteTask(ctx context.Context, listID, taskID string) error {
	if err := b.ensureLoaded(); err != nil {
		return err
	}

	// Find the task summary for commit message
	task, ok := b.tasksByID[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	summary := task.Summary

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
	if err := b.saveFile(); err != nil {
		return err
	}

	// Auto-commit if enabled
	if b.config.AutoCommit {
		_ = b.autoCommit(fmt.Sprintf("todoat: delete task '%s'", summary))
	}

	return nil
}

// =============================================================================
// File Operations
// =============================================================================

// ensureLoaded loads the markdown file if not already loaded
func (b *Backend) ensureLoaded() error {
	if b.fileLoaded {
		return nil
	}

	// Find repo and file paths if not already set
	if b.repoPath == "" {
		repoPath, err := b.findGitRepo()
		if err != nil {
			return err
		}
		if repoPath == "" {
			return fmt.Errorf("not a git repository")
		}
		b.repoPath = repoPath
	}

	if b.filePath == "" {
		filePath, err := b.findTodoFile()
		if err != nil {
			return err
		}
		if filePath == "" {
			return fmt.Errorf("no todoat-enabled markdown file found")
		}
		b.filePath = filePath
	}

	return b.loadFile()
}

// loadFile parses the markdown file into lists and tasks
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
	var parentStack []*backend.Task // Stack for tracking parent hierarchy

	// Regex patterns
	sectionPattern := regexp.MustCompile(`^##\s+(.+)$`)
	taskPattern := regexp.MustCompile(`^(\s*)-\s+\[([ xX~-])\]\s+(.*)$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for section header
		if matches := sectionPattern.FindStringSubmatch(line); len(matches) == 2 {
			listName := strings.TrimSpace(matches[1])
			list := backend.List{
				ID:       backend.GenerateID(),
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
				ID:         backend.GenerateID(),
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
			indentLevel := indent / 2 // Assuming 2-space indentation

			// Pop parents that are at same or deeper level
			for len(parentStack) > indentLevel {
				parentStack = parentStack[:len(parentStack)-1]
			}

			// Set parent if we have one
			if len(parentStack) > 0 {
				task.ParentID = parentStack[len(parentStack)-1].ID
			}

			// Add task to list
			b.tasks[currentList.ID] = append(b.tasks[currentList.ID], task)
			taskPtr := &b.tasks[currentList.ID][len(b.tasks[currentList.ID])-1]
			b.tasksByID[task.ID] = taskPtr

			// Push this task as potential parent
			parentStack = append(parentStack, taskPtr)
		}
	}

	b.fileLoaded = true
	return nil
}

// saveFile writes the lists and tasks back to the markdown file
func (b *Backend) saveFile() error {
	var sb strings.Builder

	// Write marker
	sb.WriteString(Marker)
	sb.WriteString("\n\n")

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
// Git Operations
// =============================================================================

// autoCommit commits the TODO file with the given message
func (b *Backend) autoCommit(message string) error {
	// Stage the file
	cmd := exec.Command("git", "add", b.filePath)
	cmd.Dir = b.repoPath
	if err := cmd.Run(); err != nil {
		return err
	}

	// Commit
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = b.repoPath
	return cmd.Run()
}

// Verify interface compliance at compile time
var _ backend.TaskManager = (*Backend)(nil)
var _ backend.DetectableBackend = (*Backend)(nil)

// init registers the git backend as detectable
func init() {
	// Git backend has priority 10 (higher than sqlite's 100)
	backend.RegisterDetectableWithPriority("git", func(workDir string) (backend.DetectableBackend, error) {
		cfg := Config{WorkDir: workDir}
		return New(cfg)
	}, 10)
}
