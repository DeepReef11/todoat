// Package todoist provides a backend implementation for the Todoist REST API v2.
package todoist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"todoat/backend"
)

const (
	// DefaultBaseURL is the Todoist REST API v2 base URL
	DefaultBaseURL = "https://api.todoist.com"
)

// Config holds Todoist connection settings
type Config struct {
	APIToken        string
	BaseURL         string // Override for testing
	UseKeyring      bool
	Username        string // Account identifier for keyring lookup
	MaxRetries      int
	RetryDelay      time.Duration
	EnableRateLimit bool
}

// ConfigFromEnv creates a Config from environment variables
func ConfigFromEnv() Config {
	return Config{
		APIToken: os.Getenv("TODOAT_TODOIST_TOKEN"),
	}
}

// Backend implements backend.TaskManager using Todoist REST API v2
type Backend struct {
	config  Config
	client  *http.Client
	baseURL string
}

// New creates a new Todoist backend
func New(cfg Config) (*Backend, error) {
	if cfg.APIToken == "" && !cfg.UseKeyring {
		return nil, fmt.Errorf("todoist API token is required")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	return &Backend{
		config:  cfg,
		client:  createHTTPClient(),
		baseURL: baseURL,
	}, nil
}

// createHTTPClient creates an HTTP client with proper configuration
func createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

// Close closes the backend
func (b *Backend) Close() error {
	if b.client == nil {
		return nil
	}
	if transport, ok := b.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// doRequest performs an authenticated Todoist API request with rate limiting support
func (b *Backend) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := b.baseURL + path

	maxRetries := b.config.MaxRetries
	if maxRetries == 0 {
		maxRetries = 1
	}
	retryDelay := b.config.RetryDelay
	if retryDelay == 0 {
		retryDelay = time.Second
	}

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		var bodyReader io.Reader
		if body != nil {
			jsonBody, marshalErr := json.Marshal(body)
			if marshalErr != nil {
				return nil, marshalErr
			}
			bodyReader = bytes.NewReader(jsonBody)
		}

		req, reqErr := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if reqErr != nil {
			return nil, reqErr
		}

		req.Header.Set("Authorization", "Bearer "+b.config.APIToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err = b.client.Do(req)
		if err != nil {
			return nil, err
		}

		// Handle rate limiting
		if resp.StatusCode == http.StatusTooManyRequests && b.config.EnableRateLimit {
			_ = resp.Body.Close()
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
		}

		break
	}

	return resp, err
}

// =============================================================================
// List (Project) Operations
// =============================================================================

// GetLists returns all Todoist projects
func (b *Backend) GetLists(ctx context.Context) ([]backend.List, error) {
	resp, err := b.doRequest(ctx, http.MethodGet, "/rest/v2/projects", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed: invalid API token")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get projects: status %d", resp.StatusCode)
	}

	var projects []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}

	lists := make([]backend.List, len(projects))
	for i, p := range projects {
		lists[i] = backend.List{
			ID:       p.ID,
			Name:     p.Name,
			Color:    p.Color,
			Modified: time.Now(), // Todoist doesn't provide this
		}
	}

	return lists, nil
}

// GetList returns a specific project by ID
func (b *Backend) GetList(ctx context.Context, listID string) (*backend.List, error) {
	resp, err := b.doRequest(ctx, http.MethodGet, "/rest/v2/projects/"+listID, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get project: status %d", resp.StatusCode)
	}

	var project struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, err
	}

	return &backend.List{
		ID:       project.ID,
		Name:     project.Name,
		Color:    project.Color,
		Modified: time.Now(),
	}, nil
}

// GetListByName returns a specific project by name (case-insensitive)
func (b *Backend) GetListByName(ctx context.Context, name string) (*backend.List, error) {
	lists, err := b.GetLists(ctx)
	if err != nil {
		return nil, err
	}
	return backend.FindListByName(lists, name), nil
}

// CreateList creates a new Todoist project
func (b *Backend) CreateList(ctx context.Context, name string) (*backend.List, error) {
	body := map[string]string{"name": name}

	resp, err := b.doRequest(ctx, http.MethodPost, "/rest/v2/projects", body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create project: status %d", resp.StatusCode)
	}

	var project struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, err
	}

	return &backend.List{
		ID:       project.ID,
		Name:     project.Name,
		Color:    project.Color,
		Modified: time.Now(),
	}, nil
}

// UpdateList updates a Todoist project
func (b *Backend) UpdateList(ctx context.Context, list *backend.List) (*backend.List, error) {
	body := map[string]string{"name": list.Name}
	if list.Color != "" {
		body["color"] = list.Color
	}

	resp, err := b.doRequest(ctx, http.MethodPost, "/rest/v2/projects/"+list.ID, body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update project: status %d", resp.StatusCode)
	}

	var project struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, err
	}

	return &backend.List{
		ID:       project.ID,
		Name:     project.Name,
		Color:    project.Color,
		Modified: time.Now(),
	}, nil
}

// DeleteList deletes a Todoist project (permanent deletion)
func (b *Backend) DeleteList(ctx context.Context, listID string) error {
	resp, err := b.doRequest(ctx, http.MethodDelete, "/rest/v2/projects/"+listID, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete project: status %d", resp.StatusCode)
	}

	return nil
}

// GetDeletedLists returns deleted lists (not supported by Todoist)
func (b *Backend) GetDeletedLists(ctx context.Context) ([]backend.List, error) {
	// Todoist doesn't have a trash concept for projects
	return []backend.List{}, nil
}

// GetDeletedListByName returns a deleted list by name (not supported)
func (b *Backend) GetDeletedListByName(ctx context.Context, name string) (*backend.List, error) {
	return nil, nil
}

// RestoreList restores a deleted list (not supported)
func (b *Backend) RestoreList(ctx context.Context, listID string) error {
	return fmt.Errorf("restoring projects is not supported by Todoist")
}

// PurgeList permanently deletes a list (not supported - already permanent)
func (b *Backend) PurgeList(ctx context.Context, listID string) error {
	return fmt.Errorf("purging projects is not supported by Todoist (deletion is already permanent)")
}

// =============================================================================
// Task Operations
// =============================================================================

// GetTasks returns all tasks in a project, including completed tasks.
// Active tasks are fetched from the REST API, completed tasks from the Sync API.
func (b *Backend) GetTasks(ctx context.Context, listID string) ([]backend.Task, error) {
	// Fetch active tasks from REST API
	activeTasks, err := b.getActiveTasks(ctx, listID)
	if err != nil {
		return nil, err
	}

	// Fetch completed tasks from Sync API
	completedTasks, err := b.getCompletedTasks(ctx, listID)
	if err != nil {
		// If completed tasks fetch fails, still return active tasks
		// This is a graceful degradation - the user can always use --uid
		return activeTasks, nil
	}

	// Merge active and completed tasks
	return append(activeTasks, completedTasks...), nil
}

// getActiveTasks fetches active (non-completed) tasks from the REST API
func (b *Backend) getActiveTasks(ctx context.Context, listID string) ([]backend.Task, error) {
	path := "/rest/v2/tasks"
	if listID != "" {
		path += "?project_id=" + listID
	}

	resp, err := b.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tasks: status %d", resp.StatusCode)
	}

	var todoistTasks []struct {
		ID          string   `json:"id"`
		ProjectID   string   `json:"project_id"`
		Content     string   `json:"content"`
		Description string   `json:"description"`
		IsCompleted bool     `json:"is_completed"`
		Priority    int      `json:"priority"`
		Labels      []string `json:"labels"`
		ParentID    string   `json:"parent_id"`
		CreatedAt   string   `json:"created_at"`
		Due         *struct {
			Date string `json:"date"`
		} `json:"due"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&todoistTasks); err != nil {
		return nil, err
	}

	tasks := make([]backend.Task, len(todoistTasks))
	for i, t := range todoistTasks {
		created, _ := time.Parse(time.RFC3339, t.CreatedAt)

		tasks[i] = backend.Task{
			ID:          t.ID,
			Summary:     t.Content,
			Description: t.Description,
			Status:      todoistToBackendStatus(t.IsCompleted),
			Priority:    todoistToInternalPriority(t.Priority),
			ListID:      t.ProjectID,
			ParentID:    t.ParentID,
			Categories:  labelsToCategories(t.Labels),
			Created:     created,
			Modified:    time.Now(),
		}

		if t.Due != nil && t.Due.Date != "" {
			dueDate, err := time.Parse("2006-01-02", t.Due.Date)
			if err == nil {
				tasks[i].DueDate = &dueDate
			}
		}
	}

	return tasks, nil
}

// getCompletedTasks fetches completed tasks from the Sync API
func (b *Backend) getCompletedTasks(ctx context.Context, listID string) ([]backend.Task, error) {
	path := "/sync/v9/completed/get_all"
	if listID != "" {
		path += "?project_id=" + listID
	}

	resp, err := b.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get completed tasks: status %d", resp.StatusCode)
	}

	var response struct {
		Items []struct {
			TaskID      string `json:"task_id"`
			Content     string `json:"content"`
			ProjectID   string `json:"project_id"`
			CompletedAt string `json:"completed_at"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	tasks := make([]backend.Task, len(response.Items))
	for i, item := range response.Items {
		completedAt, _ := time.Parse(time.RFC3339, item.CompletedAt)

		tasks[i] = backend.Task{
			ID:        item.TaskID,
			Summary:   item.Content,
			Status:    backend.StatusCompleted,
			Priority:  5, // Default priority for completed tasks
			ListID:    item.ProjectID,
			Completed: &completedAt,
			Created:   completedAt, // Use completed time as created (actual not available)
			Modified:  completedAt,
		}
	}

	return tasks, nil
}

// GetTask returns a specific task by ID
func (b *Backend) GetTask(ctx context.Context, listID, taskID string) (*backend.Task, error) {
	resp, err := b.doRequest(ctx, http.MethodGet, "/rest/v2/tasks/"+taskID, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get task: status %d", resp.StatusCode)
	}

	var t struct {
		ID          string   `json:"id"`
		ProjectID   string   `json:"project_id"`
		Content     string   `json:"content"`
		Description string   `json:"description"`
		IsCompleted bool     `json:"is_completed"`
		Priority    int      `json:"priority"`
		Labels      []string `json:"labels"`
		ParentID    string   `json:"parent_id"`
		CreatedAt   string   `json:"created_at"`
		Due         *struct {
			Date string `json:"date"`
		} `json:"due"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}

	created, _ := time.Parse(time.RFC3339, t.CreatedAt)

	task := &backend.Task{
		ID:          t.ID,
		Summary:     t.Content,
		Description: t.Description,
		Status:      todoistToBackendStatus(t.IsCompleted),
		Priority:    todoistToInternalPriority(t.Priority),
		ListID:      t.ProjectID,
		ParentID:    t.ParentID,
		Categories:  labelsToCategories(t.Labels),
		Created:     created,
		Modified:    time.Now(),
	}

	if t.Due != nil && t.Due.Date != "" {
		dueDate, err := time.Parse("2006-01-02", t.Due.Date)
		if err == nil {
			task.DueDate = &dueDate
		}
	}

	return task, nil
}

// CreateTask creates a new task in a project
func (b *Backend) CreateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	body := map[string]interface{}{
		"content":    task.Summary,
		"project_id": listID,
		"priority":   internalToTodoistPriority(task.Priority),
	}

	if task.Description != "" {
		body["description"] = task.Description
	}

	if task.Categories != "" {
		body["labels"] = categoriesToLabels(task.Categories)
	}

	if task.ParentID != "" {
		body["parent_id"] = task.ParentID
	}

	if task.DueDate != nil {
		body["due_date"] = task.DueDate.Format("2006-01-02")
	}

	resp, err := b.doRequest(ctx, http.MethodPost, "/rest/v2/tasks", body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create task: status %d", resp.StatusCode)
	}

	var created struct {
		ID        string `json:"id"`
		Content   string `json:"content"`
		ProjectID string `json:"project_id"`
		Priority  int    `json:"priority"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, err
	}

	return &backend.Task{
		ID:          created.ID,
		Summary:     created.Content,
		Description: task.Description,
		Status:      backend.StatusNeedsAction,
		Priority:    todoistToInternalPriority(created.Priority),
		ListID:      created.ProjectID,
		Categories:  task.Categories,
		Created:     time.Now(),
		Modified:    time.Now(),
	}, nil
}

// UpdateTask updates an existing task
func (b *Backend) UpdateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	// First update the task content/priority/etc
	body := map[string]interface{}{
		"content":  task.Summary,
		"priority": internalToTodoistPriority(task.Priority),
	}

	if task.Description != "" {
		body["description"] = task.Description
	}

	if task.Categories != "" {
		body["labels"] = categoriesToLabels(task.Categories)
	}

	resp, err := b.doRequest(ctx, http.MethodPost, "/rest/v2/tasks/"+task.ID, body)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update task: status %d", resp.StatusCode)
	}

	// Handle status changes separately (Todoist uses close/reopen endpoints)
	switch task.Status {
	case backend.StatusCompleted:
		resp, err = b.doRequest(ctx, http.MethodPost, "/rest/v2/tasks/"+task.ID+"/close", nil)
		if err != nil {
			return nil, err
		}
		_ = resp.Body.Close()
	case backend.StatusNeedsAction:
		resp, err = b.doRequest(ctx, http.MethodPost, "/rest/v2/tasks/"+task.ID+"/reopen", nil)
		if err != nil {
			return nil, err
		}
		_ = resp.Body.Close()
	}

	task.Modified = time.Now()
	return task, nil
}

// DeleteTask removes a task
func (b *Backend) DeleteTask(ctx context.Context, listID, taskID string) error {
	resp, err := b.doRequest(ctx, http.MethodDelete, "/rest/v2/tasks/"+taskID, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete task: status %d", resp.StatusCode)
	}

	return nil
}

// =============================================================================
// Priority Conversion Functions
// =============================================================================

// internalToTodoistPriority converts internal priority (1-9, 1=highest) to Todoist (1-4, 4=highest)
// Priority 0 (unset) maps to Todoist priority 1 (lowest/no priority) - fixes issue #011
func internalToTodoistPriority(internal int) int {
	// Internal: 1-9 where 1 is highest, 0 = unset/default
	// Todoist: 1-4 where 4 is highest, 1 = no priority
	switch {
	case internal <= 0:
		return 1 // Unset (0) = no priority (Todoist 1)
	case internal <= 2:
		return 4 // Highest
	case internal <= 4:
		return 3
	case internal <= 6:
		return 2
	default:
		return 1 // Lowest
	}
}

// todoistToInternalPriority converts Todoist priority (1-4, 4=highest) to internal (1-9, 1=highest)
func todoistToInternalPriority(todoist int) int {
	// Todoist: 1-4 where 4 is highest
	// Internal: 1-9 where 1 is highest
	switch todoist {
	case 4:
		return 1 // Highest
	case 3:
		return 3
	case 2:
		return 5
	default:
		return 7 // Lowest
	}
}

// =============================================================================
// Status Conversion Functions
// =============================================================================

// todoistToBackendStatus converts Todoist isCompleted to backend status
func todoistToBackendStatus(isCompleted bool) backend.TaskStatus {
	if isCompleted {
		return backend.StatusCompleted
	}
	return backend.StatusNeedsAction
}

// =============================================================================
// Label/Category Conversion Functions
// =============================================================================

// labelsToCategories converts Todoist labels to comma-separated categories
func labelsToCategories(labels []string) string {
	return strings.Join(labels, ",")
}

// categoriesToLabels converts comma-separated categories to Todoist labels
func categoriesToLabels(categories string) []string {
	if categories == "" {
		return nil
	}
	return strings.Split(categories, ",")
}

// =============================================================================
// DetectableBackend Interface Implementation
// =============================================================================

// CanDetect checks if the Todoist backend can be used.
// It returns true if TODOAT_TODOIST_TOKEN environment variable is set.
func (b *Backend) CanDetect() (bool, error) {
	// Todoist is available if we have an API token
	return b.config.APIToken != "", nil
}

// DetectionInfo returns human-readable information about the Todoist backend.
func (b *Backend) DetectionInfo() string {
	return "Todoist API (token from TODOAT_TODOIST_TOKEN)"
}

// Verify interface compliance at compile time
var _ backend.TaskManager = (*Backend)(nil)
var _ backend.DetectableBackend = (*Backend)(nil)

// init registers the todoist backend as detectable
func init() {
	// Todoist has priority 50 (higher than sqlite's 100, lower than git's 10)
	backend.RegisterDetectableWithPriority("todoist", func(workDir string) (backend.DetectableBackend, error) {
		cfg := ConfigFromEnv()
		if cfg.APIToken == "" {
			// Return a backend that will report not available
			return &Backend{config: cfg}, nil
		}
		return New(cfg)
	}, 50)
}
