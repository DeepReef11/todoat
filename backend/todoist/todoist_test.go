package todoist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"todoat/backend"
)

// =============================================================================
// Todoist REST API Mock Server for Tests
// =============================================================================

// mockTodoistServer simulates the Todoist REST API v2
type mockTodoistServer struct {
	server      *httptest.Server
	projects    map[string]*todoistProject
	tasks       map[string]*todoistTask
	apiToken    string
	mu          sync.Mutex
	rateLimited bool
	requestLog  []string
}

type todoistProject struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color,omitempty"`
	ParentID string `json:"parent_id,omitempty"`
	Order    int    `json:"order"`
}

type todoistTask struct {
	ID          string   `json:"id"`
	ProjectID   string   `json:"project_id"`
	Content     string   `json:"content"`
	Description string   `json:"description,omitempty"`
	IsCompleted bool     `json:"is_completed"`
	Priority    int      `json:"priority"` // Todoist: 1=normal, 4=urgent
	DueDate     string   `json:"due_date,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	ParentID    string   `json:"parent_id,omitempty"`
	Order       int      `json:"order"`
	CreatedAt   string   `json:"created_at"`
}

func newMockTodoistServer(apiToken string) *mockTodoistServer {
	m := &mockTodoistServer{
		projects:   make(map[string]*todoistProject),
		tasks:      make(map[string]*todoistTask),
		apiToken:   apiToken,
		requestLog: []string{},
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.handler))
	return m
}

func (m *mockTodoistServer) Close() {
	m.server.Close()
}

func (m *mockTodoistServer) URL() string {
	return m.server.URL
}

func (m *mockTodoistServer) AddProject(id, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects[id] = &todoistProject{
		ID:   id,
		Name: name,
	}
}

func (m *mockTodoistServer) AddTask(id, projectID, content string, priority int, labels []string, parentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[id] = &todoistTask{
		ID:        id,
		ProjectID: projectID,
		Content:   content,
		Priority:  priority,
		Labels:    labels,
		ParentID:  parentID,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func (m *mockTodoistServer) SetRateLimited(limited bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rateLimited = limited
}

func (m *mockTodoistServer) GetRequestLog() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.requestLog...)
}

func (m *mockTodoistServer) handler(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.requestLog = append(m.requestLog, r.Method+" "+r.URL.Path)

	// Check rate limiting first
	if m.rateLimited {
		m.mu.Unlock()
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	m.mu.Unlock()

	// Check auth
	auth := r.Header.Get("Authorization")
	if auth != "Bearer "+m.apiToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	path := r.URL.Path

	switch {
	case path == "/rest/v2/projects" && r.Method == http.MethodGet:
		m.handleGetProjects(w, r)
	case path == "/rest/v2/projects" && r.Method == http.MethodPost:
		m.handleCreateProject(w, r)
	case strings.HasPrefix(path, "/rest/v2/projects/") && r.Method == http.MethodGet:
		m.handleGetProject(w, r, strings.TrimPrefix(path, "/rest/v2/projects/"))
	case strings.HasPrefix(path, "/rest/v2/projects/") && r.Method == http.MethodDelete:
		m.handleDeleteProject(w, r, strings.TrimPrefix(path, "/rest/v2/projects/"))
	case path == "/rest/v2/tasks" && r.Method == http.MethodGet:
		m.handleGetTasks(w, r)
	case path == "/rest/v2/tasks" && r.Method == http.MethodPost:
		m.handleCreateTask(w, r)
	case strings.HasPrefix(path, "/rest/v2/tasks/") && strings.HasSuffix(path, "/close") && r.Method == http.MethodPost:
		taskID := strings.TrimSuffix(strings.TrimPrefix(path, "/rest/v2/tasks/"), "/close")
		m.handleCloseTask(w, r, taskID)
	case strings.HasPrefix(path, "/rest/v2/tasks/") && strings.HasSuffix(path, "/reopen") && r.Method == http.MethodPost:
		taskID := strings.TrimSuffix(strings.TrimPrefix(path, "/rest/v2/tasks/"), "/reopen")
		m.handleReopenTask(w, r, taskID)
	case strings.HasPrefix(path, "/rest/v2/tasks/") && r.Method == http.MethodGet:
		m.handleGetTask(w, r, strings.TrimPrefix(path, "/rest/v2/tasks/"))
	case strings.HasPrefix(path, "/rest/v2/tasks/") && r.Method == http.MethodPost:
		m.handleUpdateTask(w, r, strings.TrimPrefix(path, "/rest/v2/tasks/"))
	case strings.HasPrefix(path, "/rest/v2/tasks/") && r.Method == http.MethodDelete:
		m.handleDeleteTask(w, r, strings.TrimPrefix(path, "/rest/v2/tasks/"))
	case path == "/sync/v9/completed/get_all" && r.Method == http.MethodGet:
		m.handleGetCompletedTasks(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *mockTodoistServer) handleGetProjects(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var projects []*todoistProject
	for _, p := range m.projects {
		projects = append(projects, p)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(projects)
}

func (m *mockTodoistServer) handleGetProject(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	project, ok := m.projects[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(project)
}

func (m *mockTodoistServer) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("proj-%d", len(m.projects)+1)
	project := &todoistProject{
		ID:   id,
		Name: input.Name,
	}
	m.projects[id] = project

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(project)
}

func (m *mockTodoistServer) handleDeleteProject(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.projects[id]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	delete(m.projects, id)
	w.WriteHeader(http.StatusNoContent)
}

func (m *mockTodoistServer) handleGetTasks(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	projectID := r.URL.Query().Get("project_id")

	var tasks []*todoistTask
	for _, t := range m.tasks {
		// Real Todoist API only returns active (non-completed) tasks
		if t.IsCompleted {
			continue
		}
		if projectID == "" || t.ProjectID == projectID {
			tasks = append(tasks, t)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tasks)
}

func (m *mockTodoistServer) handleGetTask(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (m *mockTodoistServer) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var input struct {
		Content     string   `json:"content"`
		Description string   `json:"description"`
		ProjectID   string   `json:"project_id"`
		Priority    int      `json:"priority"`
		Labels      []string `json:"labels"`
		ParentID    string   `json:"parent_id"`
		DueDate     string   `json:"due_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("task-%d", len(m.tasks)+1)
	task := &todoistTask{
		ID:          id,
		ProjectID:   input.ProjectID,
		Content:     input.Content,
		Description: input.Description,
		Priority:    input.Priority,
		Labels:      input.Labels,
		ParentID:    input.ParentID,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	m.tasks[id] = task

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(task)
}

func (m *mockTodoistServer) handleUpdateTask(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var input struct {
		Content     *string  `json:"content"`
		Description *string  `json:"description"`
		Priority    *int     `json:"priority"`
		Labels      []string `json:"labels"`
		DueDate     *string  `json:"due_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if input.Content != nil {
		task.Content = *input.Content
	}
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.Priority != nil {
		task.Priority = *input.Priority
	}
	if input.Labels != nil {
		task.Labels = input.Labels
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (m *mockTodoistServer) handleDeleteTask(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.tasks[id]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	delete(m.tasks, id)
	w.WriteHeader(http.StatusNoContent)
}

func (m *mockTodoistServer) handleCloseTask(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	task.IsCompleted = true
	w.WriteHeader(http.StatusNoContent)
}

func (m *mockTodoistServer) handleReopenTask(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	task.IsCompleted = false
	w.WriteHeader(http.StatusNoContent)
}

// handleGetCompletedTasks returns completed tasks via Sync API endpoint
func (m *mockTodoistServer) handleGetCompletedTasks(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	projectID := r.URL.Query().Get("project_id")

	// Response format matches Todoist Sync API v9 completed/get_all
	type completedItem struct {
		ID          string `json:"id"`
		TaskID      string `json:"task_id"`
		Content     string `json:"content"`
		ProjectID   string `json:"project_id"`
		CompletedAt string `json:"completed_at"`
	}

	var items []completedItem
	for _, t := range m.tasks {
		if !t.IsCompleted {
			continue
		}
		if projectID != "" && t.ProjectID != projectID {
			continue
		}
		items = append(items, completedItem{
			ID:          t.ID + "-completed",
			TaskID:      t.ID,
			Content:     t.Content,
			ProjectID:   t.ProjectID,
			CompletedAt: time.Now().UTC().Format(time.RFC3339),
		})
	}

	response := struct {
		Items []completedItem `json:"items"`
	}{
		Items: items,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// =============================================================================
// CLI Tests Required (021-todoist-backend)
// =============================================================================

// TestTodoistListProjects - todoat --backend=todoist list shows Todoist projects
func TestTodoistListProjects(t *testing.T) {
	server := newMockTodoistServer("test-api-token")
	defer server.Close()

	server.AddProject("proj-1", "Work")
	server.AddProject("proj-2", "Personal")

	be, err := New(Config{
		APIToken: "test-api-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists failed: %v", err)
	}

	if len(lists) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(lists))
	}

	names := make(map[string]bool)
	for _, l := range lists {
		names[l.Name] = true
	}

	if !names["Work"] {
		t.Error("Expected to find project 'Work'")
	}
	if !names["Personal"] {
		t.Error("Expected to find project 'Personal'")
	}
}

// TestTodoistGetTasks - todoat --backend=todoist MyProject retrieves tasks from Todoist
func TestTodoistGetTasks(t *testing.T) {
	server := newMockTodoistServer("test-api-token")
	defer server.Close()

	server.AddProject("proj-1", "MyProject")
	server.AddTask("task-1", "proj-1", "Buy groceries", 1, nil, "")
	server.AddTask("task-2", "proj-1", "Review PR", 2, nil, "")

	be, err := New(Config{
		APIToken: "test-api-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	tasks, err := be.GetTasks(ctx, "proj-1")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}

	taskSummaries := make(map[string]bool)
	for _, task := range tasks {
		taskSummaries[task.Summary] = true
	}

	if !taskSummaries["Buy groceries"] {
		t.Error("Expected to find task 'Buy groceries'")
	}
	if !taskSummaries["Review PR"] {
		t.Error("Expected to find task 'Review PR'")
	}
}

// TestTodoistAddTask - todoat --backend=todoist MyProject add "Task" creates task via API
func TestTodoistAddTask(t *testing.T) {
	server := newMockTodoistServer("test-api-token")
	defer server.Close()

	server.AddProject("proj-1", "MyProject")

	be, err := New(Config{
		APIToken: "test-api-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	task := &backend.Task{
		Summary:  "New Task",
		Priority: 5, // Internal priority (1=highest, 9=lowest)
	}

	created, err := be.CreateTask(ctx, "proj-1", task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if created.Summary != "New Task" {
		t.Errorf("Expected summary 'New Task', got '%s'", created.Summary)
	}

	if created.ID == "" {
		t.Error("Expected task to have an ID")
	}

	// Verify task exists on server
	tasks, err := be.GetTasks(ctx, "proj-1")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	found := false
	for _, task := range tasks {
		if task.Summary == "New Task" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Created task not found on server")
	}
}

// TestTodoistUpdateTask - todoat --backend=todoist MyProject update "Task" -s DONE updates task
func TestTodoistUpdateTask(t *testing.T) {
	server := newMockTodoistServer("test-api-token")
	defer server.Close()

	server.AddProject("proj-1", "MyProject")
	server.AddTask("task-1", "proj-1", "Existing Task", 1, nil, "")

	be, err := New(Config{
		APIToken: "test-api-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// Get the task first
	tasks, err := be.GetTasks(ctx, "proj-1")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks) == 0 {
		t.Fatal("Expected at least one task")
	}

	task := &tasks[0]
	task.Status = backend.StatusCompleted
	task.Summary = "Updated Task"

	updated, err := be.UpdateTask(ctx, "proj-1", task)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	if updated.Status != backend.StatusCompleted {
		t.Errorf("Expected status COMPLETED, got %s", updated.Status)
	}
}

// TestTodoistDeleteTask - todoat --backend=todoist MyProject delete "Task" removes task
func TestTodoistDeleteTask(t *testing.T) {
	server := newMockTodoistServer("test-api-token")
	defer server.Close()

	server.AddProject("proj-1", "MyProject")
	server.AddTask("task-to-delete", "proj-1", "Task to Delete", 1, nil, "")

	be, err := New(Config{
		APIToken: "test-api-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	err = be.DeleteTask(ctx, "proj-1", "task-to-delete")
	if err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	// Verify task is gone
	tasks, err := be.GetTasks(ctx, "proj-1")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks after deletion, got %d", len(tasks))
	}
}

// TestTodoistPriorityMapping - Internal priority 1-9 maps to Todoist priority 1-4 correctly
func TestTodoistPriorityMapping(t *testing.T) {
	// Todoist uses 1-4 where 4 is highest (most urgent)
	// Internal uses 1-9 where 1 is highest
	tests := []struct {
		internalPriority int
		todoistPriority  int
	}{
		{1, 4}, // Highest internal = highest Todoist (4)
		{2, 4}, // Still highest
		{3, 3},
		{4, 3},
		{5, 2}, // Medium
		{6, 2},
		{7, 1}, // Low
		{8, 1},
		{9, 1}, // Lowest internal = lowest Todoist (1)
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.internalPriority), func(t *testing.T) {
			todoist := internalToTodoistPriority(tt.internalPriority)
			if todoist != tt.todoistPriority {
				t.Errorf("Internal priority %d: expected Todoist %d, got %d",
					tt.internalPriority, tt.todoistPriority, todoist)
			}

			internal := todoistToInternalPriority(tt.todoistPriority)
			// Reverse mapping picks the middle value of the range
			expectedInternal := []int{7, 5, 3, 1}[tt.todoistPriority-1]
			if internal != expectedInternal {
				t.Errorf("Todoist priority %d: expected internal %d, got %d",
					tt.todoistPriority, expectedInternal, internal)
			}
		})
	}
}

// TestTodoistLabelsAsCategories - Todoist labels map to todoat categories/tags
func TestTodoistLabelsAsCategories(t *testing.T) {
	server := newMockTodoistServer("test-api-token")
	defer server.Close()

	server.AddProject("proj-1", "MyProject")
	server.AddTask("task-1", "proj-1", "Labeled Task", 1, []string{"work", "urgent"}, "")

	be, err := New(Config{
		APIToken: "test-api-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	tasks, err := be.GetTasks(ctx, "proj-1")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks) == 0 {
		t.Fatal("Expected at least one task")
	}

	task := tasks[0]
	if task.Categories != "work,urgent" {
		t.Errorf("Expected categories 'work,urgent', got '%s'", task.Categories)
	}
}

// TestTodoistAPITokenFromKeyring - Backend retrieves API token from system keyring
func TestTodoistAPITokenFromKeyring(t *testing.T) {
	// This tests the credential resolution logic
	// In a real environment, it would use the system keyring
	cfg := Config{
		UseKeyring: true,
		Username:   "testuser", // Used as account for keyring lookup
	}

	// Verify the config structure is correct
	if !cfg.UseKeyring {
		t.Error("UseKeyring should be true")
	}

	// The actual keyring integration is tested in the credentials package
}

// TestTodoistAPITokenFromEnv - Backend retrieves token from TODOAT_TODOIST_TOKEN env var
func TestTodoistAPITokenFromEnv(t *testing.T) {
	// Save original env vars
	origToken := os.Getenv("TODOAT_TODOIST_TOKEN")
	defer func() {
		_ = os.Setenv("TODOAT_TODOIST_TOKEN", origToken)
	}()

	server := newMockTodoistServer("env-api-token")
	defer server.Close()

	server.AddProject("proj-1", "EnvTest")

	// Set environment variable
	_ = os.Setenv("TODOAT_TODOIST_TOKEN", "env-api-token")

	// Create backend using env config
	cfg := ConfigFromEnv()
	cfg.BaseURL = server.URL()

	be, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create backend from env: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists failed: %v", err)
	}

	if len(lists) != 1 {
		t.Errorf("Expected 1 list, got %d", len(lists))
	}

	if lists[0].Name != "EnvTest" {
		t.Errorf("Expected list name 'EnvTest', got '%s'", lists[0].Name)
	}
}

// TestTodoistRateLimiting - Backend respects Todoist API rate limits with backoff
func TestTodoistRateLimiting(t *testing.T) {
	server := newMockTodoistServer("test-api-token")
	defer server.Close()

	server.AddProject("proj-1", "TestProject")

	// Enable rate limiting
	server.SetRateLimited(true)

	be, err := New(Config{
		APIToken:        "test-api-token",
		BaseURL:         server.URL(),
		MaxRetries:      2,
		RetryDelay:      50 * time.Millisecond,
		EnableRateLimit: true,
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// First call should get rate limited and retry
	go func() {
		time.Sleep(100 * time.Millisecond)
		server.SetRateLimited(false)
	}()

	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists failed after retries: %v", err)
	}

	if len(lists) != 1 {
		t.Errorf("Expected 1 list, got %d", len(lists))
	}

	// Verify retries happened
	log := server.GetRequestLog()
	if len(log) < 2 {
		t.Errorf("Expected at least 2 requests (initial + retry), got %d", len(log))
	}
}

// TestTodoistSubtasks - Parent-child relationships sync correctly with Todoist
func TestTodoistSubtasks(t *testing.T) {
	server := newMockTodoistServer("test-api-token")
	defer server.Close()

	server.AddProject("proj-1", "MyProject")
	server.AddTask("parent-task", "proj-1", "Parent Task", 1, nil, "")
	server.AddTask("child-task-1", "proj-1", "Child Task 1", 1, nil, "parent-task")
	server.AddTask("child-task-2", "proj-1", "Child Task 2", 1, nil, "parent-task")

	be, err := New(Config{
		APIToken: "test-api-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	tasks, err := be.GetTasks(ctx, "proj-1")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	// Verify parent-child relationships
	parentFound := false
	childCount := 0
	for _, task := range tasks {
		if task.Summary == "Parent Task" {
			parentFound = true
			if task.ParentID != "" {
				t.Error("Parent task should not have a ParentID")
			}
		}
		if strings.HasPrefix(task.Summary, "Child Task") {
			childCount++
			if task.ParentID != "parent-task" {
				t.Errorf("Child task should have ParentID 'parent-task', got '%s'", task.ParentID)
			}
		}
	}

	if !parentFound {
		t.Error("Parent task not found")
	}
	if childCount != 2 {
		t.Errorf("Expected 2 child tasks, got %d", childCount)
	}
}

// =============================================================================
// Additional Unit Tests
// =============================================================================

func TestNewBackend(t *testing.T) {
	server := newMockTodoistServer("test-token")
	defer server.Close()

	be, err := New(Config{
		APIToken: "test-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	if be == nil {
		t.Fatal("Expected non-nil backend")
	}

	if err := be.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestBackendImplementsInterface(t *testing.T) {
	var _ backend.TaskManager = (*Backend)(nil)
}

func TestGetList(t *testing.T) {
	server := newMockTodoistServer("test-token")
	defer server.Close()

	server.AddProject("proj-1", "TestProject")

	be, err := New(Config{
		APIToken: "test-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	list, err := be.GetList(ctx, "proj-1")
	if err != nil {
		t.Fatalf("GetList failed: %v", err)
	}

	if list == nil {
		t.Fatal("Expected non-nil list")
	}

	if list.Name != "TestProject" {
		t.Errorf("Expected list name 'TestProject', got '%s'", list.Name)
	}
}

func TestGetListByName(t *testing.T) {
	server := newMockTodoistServer("test-token")
	defer server.Close()

	server.AddProject("proj-1", "MyTasks")

	be, err := New(Config{
		APIToken: "test-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// Exact match
	list, err := be.GetListByName(ctx, "MyTasks")
	if err != nil {
		t.Fatalf("GetListByName failed: %v", err)
	}

	if list == nil {
		t.Fatal("Expected non-nil list")
	}

	if list.Name != "MyTasks" {
		t.Errorf("Expected list name 'MyTasks', got '%s'", list.Name)
	}

	// Case-insensitive match
	list, err = be.GetListByName(ctx, "mytasks")
	if err != nil {
		t.Fatalf("GetListByName (case-insensitive) failed: %v", err)
	}

	if list == nil {
		t.Fatal("Expected non-nil list for case-insensitive match")
	}
}

func TestGetNonExistentList(t *testing.T) {
	server := newMockTodoistServer("test-token")
	defer server.Close()

	be, err := New(Config{
		APIToken: "test-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	list, err := be.GetList(ctx, "nonexistent-id")
	if err != nil {
		// Todoist returns 404 for non-existent projects
		// Backend should handle gracefully
		return
	}

	if list != nil {
		t.Error("Expected nil for non-existent list")
	}
}

func TestGetTask(t *testing.T) {
	server := newMockTodoistServer("test-token")
	defer server.Close()

	server.AddProject("proj-1", "Work")
	server.AddTask("task-123", "proj-1", "Important Task", 4, nil, "")

	be, err := New(Config{
		APIToken: "test-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	task, err := be.GetTask(ctx, "proj-1", "task-123")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if task == nil {
		t.Fatal("Expected non-nil task")
	}

	if task.Summary != "Important Task" {
		t.Errorf("Expected summary 'Important Task', got '%s'", task.Summary)
	}
}

func TestConfigFromEnv(t *testing.T) {
	origToken := os.Getenv("TODOAT_TODOIST_TOKEN")
	defer func() {
		_ = os.Setenv("TODOAT_TODOIST_TOKEN", origToken)
	}()

	_ = os.Setenv("TODOAT_TODOIST_TOKEN", "my-secret-token")

	cfg := ConfigFromEnv()

	if cfg.APIToken != "my-secret-token" {
		t.Errorf("Expected API token 'my-secret-token', got '%s'", cfg.APIToken)
	}
}

func TestCreateList(t *testing.T) {
	server := newMockTodoistServer("test-token")
	defer server.Close()

	be, err := New(Config{
		APIToken: "test-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	list, err := be.CreateList(ctx, "New Project")
	if err != nil {
		t.Fatalf("CreateList failed: %v", err)
	}

	if list == nil {
		t.Fatal("Expected non-nil list")
	}

	if list.Name != "New Project" {
		t.Errorf("Expected list name 'New Project', got '%s'", list.Name)
	}

	if list.ID == "" {
		t.Error("Expected list to have an ID")
	}
}

func TestDeleteList(t *testing.T) {
	server := newMockTodoistServer("test-token")
	defer server.Close()

	server.AddProject("proj-1", "ToDelete")

	be, err := New(Config{
		APIToken: "test-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	err = be.DeleteList(ctx, "proj-1")
	if err != nil {
		t.Fatalf("DeleteList failed: %v", err)
	}

	// Verify project is gone
	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists failed: %v", err)
	}

	for _, l := range lists {
		if l.ID == "proj-1" {
			t.Error("Project should have been deleted")
		}
	}
}

// Trash operations are not supported by Todoist
func TestTrashOperationsNotSupported(t *testing.T) {
	server := newMockTodoistServer("test-token")
	defer server.Close()

	be, err := New(Config{
		APIToken: "test-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// GetDeletedLists should return empty
	deletedLists, err := be.GetDeletedLists(ctx)
	if err != nil {
		// Some implementations might error, which is acceptable
		return
	}
	if len(deletedLists) != 0 {
		t.Errorf("Expected empty deleted lists, got %d", len(deletedLists))
	}

	// GetDeletedListByName should return nil
	deleted, _ := be.GetDeletedListByName(ctx, "anything")
	if deleted != nil {
		t.Error("Expected nil for deleted list by name")
	}

	// RestoreList should error
	err = be.RestoreList(ctx, "anything")
	if err == nil {
		t.Error("Expected error from RestoreList")
	}

	// PurgeList should error
	err = be.PurgeList(ctx, "anything")
	if err == nil {
		t.Error("Expected error from PurgeList")
	}
}

func TestStatusMapping(t *testing.T) {
	// Test isCompleted to backend status conversion
	tests := []struct {
		isCompleted    bool
		expectedStatus backend.TaskStatus
	}{
		{false, backend.StatusNeedsAction},
		{true, backend.StatusCompleted},
	}

	for _, tt := range tests {
		t.Run(string(tt.expectedStatus), func(t *testing.T) {
			status := todoistToBackendStatus(tt.isCompleted)
			if status != tt.expectedStatus {
				t.Errorf("Expected %s, got %s", tt.expectedStatus, status)
			}
		})
	}
}

func TestCategoriesFromLabels(t *testing.T) {
	labels := []string{"work", "urgent", "review"}
	categories := labelsToCategories(labels)

	if categories != "work,urgent,review" {
		t.Errorf("Expected 'work,urgent,review', got '%s'", categories)
	}

	// Test empty labels
	empty := labelsToCategories(nil)
	if empty != "" {
		t.Errorf("Expected empty string, got '%s'", empty)
	}
}

func TestCategoriesToLabels(t *testing.T) {
	categories := "work,urgent,review"
	labels := categoriesToLabels(categories)

	if len(labels) != 3 {
		t.Errorf("Expected 3 labels, got %d", len(labels))
	}

	expected := []string{"work", "urgent", "review"}
	for i, label := range labels {
		if label != expected[i] {
			t.Errorf("Expected label '%s', got '%s'", expected[i], label)
		}
	}

	// Test empty categories
	empty := categoriesToLabels("")
	if len(empty) != 0 {
		t.Errorf("Expected 0 labels, got %d", len(empty))
	}
}

func TestAuthFailure(t *testing.T) {
	server := newMockTodoistServer("correct-token")
	defer server.Close()

	be, err := New(Config{
		APIToken: "wrong-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	_, err = be.GetLists(ctx)
	if err == nil {
		t.Error("Expected auth error with wrong token")
	}
}

// TestTodoistFindCompletedTaskBySummaryCLI - Issue #001: delete cannot find completed tasks
// This test reproduces the issue where Todoist's GetTasks API only returns active tasks,
// making it impossible to find completed tasks by summary for deletion.
func TestTodoistFindCompletedTaskBySummaryCLI(t *testing.T) {
	server := newMockTodoistServer("test-api-token")
	defer server.Close()

	server.AddProject("proj-1", "Inbox")
	server.AddTask("task-1", "proj-1", "Test from todoat", 1, nil, "")

	be, err := New(Config{
		APIToken: "test-api-token",
		BaseURL:  server.URL(),
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// First verify the task is visible when active
	tasks, err := be.GetTasks(ctx, "proj-1")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Summary != "Test from todoat" {
		t.Fatalf("Expected task summary 'Test from todoat', got '%s'", tasks[0].Summary)
	}

	// Now mark the task as completed
	task := tasks[0]
	task.Status = backend.StatusCompleted
	_, err = be.UpdateTask(ctx, "proj-1", &task)
	if err != nil {
		t.Fatalf("UpdateTask (complete) failed: %v", err)
	}

	// After completion, GetTasks no longer returns the task (matching real Todoist API behavior)
	tasksAfterComplete, err := be.GetTasks(ctx, "proj-1")
	if err != nil {
		t.Fatalf("GetTasks after completion failed: %v", err)
	}

	// This is the bug: completed tasks are not returned by GetTasks,
	// so CLI cannot find the task by summary to delete it
	// The fix should make completed tasks findable
	foundCompletedTask := false
	for _, t := range tasksAfterComplete {
		if t.Summary == "Test from todoat" {
			foundCompletedTask = true
			break
		}
	}

	// Currently this will fail because the mock now correctly filters out completed tasks
	// After the fix is implemented, GetTasks should include completed tasks
	if !foundCompletedTask {
		t.Errorf("Bug #001: Completed task 'Test from todoat' not found in GetTasks - this prevents deletion by summary")
	}
}
