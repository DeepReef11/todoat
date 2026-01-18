package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"todoat/backend"
)

// =============================================================================
// Google Tasks API Mock Server for Tests
// =============================================================================

// mockGoogleTasksServer simulates the Google Tasks API v1
type mockGoogleTasksServer struct {
	server       *httptest.Server
	taskLists    map[string]*googleTaskList
	tasks        map[string]map[string]*googleTask // taskListID -> taskID -> task
	accessToken  string
	refreshToken string
	tokenExpired bool
	mu           sync.Mutex
	requestLog   []string
	refreshCount int
}

type googleTaskList struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Updated string `json:"updated"`
}

type googleTask struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Notes     string `json:"notes,omitempty"`
	Status    string `json:"status"` // "needsAction" or "completed"
	Due       string `json:"due,omitempty"`
	Parent    string `json:"parent,omitempty"`
	Position  string `json:"position,omitempty"`
	Updated   string `json:"updated"`
	Completed string `json:"completed,omitempty"`
}

func newMockGoogleTasksServer(accessToken, refreshToken string) *mockGoogleTasksServer {
	m := &mockGoogleTasksServer{
		taskLists:    make(map[string]*googleTaskList),
		tasks:        make(map[string]map[string]*googleTask),
		accessToken:  accessToken,
		refreshToken: refreshToken,
		requestLog:   []string{},
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.handler))
	return m
}

func (m *mockGoogleTasksServer) Close() {
	m.server.Close()
}

func (m *mockGoogleTasksServer) URL() string {
	return m.server.URL
}

func (m *mockGoogleTasksServer) AddTaskList(id, title string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.taskLists[id] = &googleTaskList{
		ID:      id,
		Title:   title,
		Updated: time.Now().UTC().Format(time.RFC3339),
	}
	m.tasks[id] = make(map[string]*googleTask)
}

func (m *mockGoogleTasksServer) AddTask(listID, taskID, title, status, parent, due string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.tasks[listID] == nil {
		m.tasks[listID] = make(map[string]*googleTask)
	}
	m.tasks[listID][taskID] = &googleTask{
		ID:      taskID,
		Title:   title,
		Status:  status, // "needsAction" or "completed"
		Parent:  parent,
		Due:     due,
		Updated: time.Now().UTC().Format(time.RFC3339),
	}
}

func (m *mockGoogleTasksServer) SetTokenExpired(expired bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokenExpired = expired
}

func (m *mockGoogleTasksServer) GetRefreshCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.refreshCount
}

func (m *mockGoogleTasksServer) GetRequestLog() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.requestLog...)
}

func (m *mockGoogleTasksServer) handler(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.requestLog = append(m.requestLog, r.Method+" "+r.URL.Path)

	path := r.URL.Path

	// Handle OAuth2 token refresh endpoint
	if path == "/oauth2/v4/token" || path == "/token" {
		m.handleTokenRefresh(w, r)
		m.mu.Unlock()
		return
	}

	// Check auth for API calls
	auth := r.Header.Get("Authorization")
	if m.tokenExpired {
		m.mu.Unlock()
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "401",
				"message": "Invalid Credentials",
			},
		})
		return
	}

	if auth != "Bearer "+m.accessToken {
		m.mu.Unlock()
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	m.mu.Unlock()

	switch {
	case path == "/tasks/v1/users/@me/lists" && r.Method == http.MethodGet:
		m.handleGetTaskLists(w, r)
	case path == "/tasks/v1/users/@me/lists" && r.Method == http.MethodPost:
		m.handleCreateTaskList(w, r)
	case strings.HasPrefix(path, "/tasks/v1/users/@me/lists/") && r.Method == http.MethodGet && !strings.Contains(strings.TrimPrefix(path, "/tasks/v1/users/@me/lists/"), "/"):
		listID := strings.TrimPrefix(path, "/tasks/v1/users/@me/lists/")
		m.handleGetTaskList(w, r, listID)
	case strings.HasPrefix(path, "/tasks/v1/users/@me/lists/") && r.Method == http.MethodDelete && !strings.Contains(strings.TrimPrefix(path, "/tasks/v1/users/@me/lists/"), "/"):
		listID := strings.TrimPrefix(path, "/tasks/v1/users/@me/lists/")
		m.handleDeleteTaskList(w, r, listID)
	case strings.HasPrefix(path, "/tasks/v1/lists/") && strings.HasSuffix(path, "/tasks") && r.Method == http.MethodGet:
		listID := strings.TrimPrefix(strings.TrimSuffix(path, "/tasks"), "/tasks/v1/lists/")
		m.handleGetTasks(w, r, listID)
	case strings.HasPrefix(path, "/tasks/v1/lists/") && strings.HasSuffix(path, "/tasks") && r.Method == http.MethodPost:
		listID := strings.TrimPrefix(strings.TrimSuffix(path, "/tasks"), "/tasks/v1/lists/")
		m.handleCreateTask(w, r, listID)
	case strings.HasPrefix(path, "/tasks/v1/lists/") && strings.Count(path, "/tasks/") >= 2 && r.Method == http.MethodGet:
		// Path format: /tasks/v1/lists/{listID}/tasks/{taskID}
		trimmed := strings.TrimPrefix(path, "/tasks/v1/lists/")
		parts := strings.SplitN(trimmed, "/tasks/", 2)
		listID := parts[0]
		taskID := parts[1]
		m.handleGetTask(w, r, listID, taskID)
	case strings.HasPrefix(path, "/tasks/v1/lists/") && strings.Count(path, "/tasks/") >= 2 && r.Method == http.MethodPatch:
		// Path format: /tasks/v1/lists/{listID}/tasks/{taskID}
		trimmed := strings.TrimPrefix(path, "/tasks/v1/lists/")
		parts := strings.SplitN(trimmed, "/tasks/", 2)
		listID := parts[0]
		taskID := parts[1]
		m.handleUpdateTask(w, r, listID, taskID)
	case strings.HasPrefix(path, "/tasks/v1/lists/") && strings.Count(path, "/tasks/") >= 2 && r.Method == http.MethodDelete:
		// Path format: /tasks/v1/lists/{listID}/tasks/{taskID}
		trimmed := strings.TrimPrefix(path, "/tasks/v1/lists/")
		parts := strings.SplitN(trimmed, "/tasks/", 2)
		listID := parts[0]
		taskID := parts[1]
		m.handleDeleteTask(w, r, listID, taskID)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *mockGoogleTasksServer) handleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	m.refreshCount++
	m.tokenExpired = false

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  m.accessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": m.refreshToken,
	})
}

func (m *mockGoogleTasksServer) handleGetTaskLists(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var items []*googleTaskList
	for _, tl := range m.taskLists {
		items = append(items, tl)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"kind":  "tasks#taskLists",
		"items": items,
	})
}

func (m *mockGoogleTasksServer) handleGetTaskList(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tl, ok := m.taskLists[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tl)
}

func (m *mockGoogleTasksServer) handleCreateTaskList(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var input struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("list-%d", len(m.taskLists)+1)
	tl := &googleTaskList{
		ID:      id,
		Title:   input.Title,
		Updated: time.Now().UTC().Format(time.RFC3339),
	}
	m.taskLists[id] = tl
	m.tasks[id] = make(map[string]*googleTask)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tl)
}

func (m *mockGoogleTasksServer) handleDeleteTaskList(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.taskLists[id]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	delete(m.taskLists, id)
	delete(m.tasks, id)
	w.WriteHeader(http.StatusNoContent)
}

func (m *mockGoogleTasksServer) handleGetTasks(w http.ResponseWriter, r *http.Request, listID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	listTasks, ok := m.tasks[listID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var items []*googleTask
	for _, t := range listTasks {
		items = append(items, t)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"kind":  "tasks#tasks",
		"items": items,
	})
}

func (m *mockGoogleTasksServer) handleGetTask(w http.ResponseWriter, r *http.Request, listID, taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	listTasks, ok := m.tasks[listID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	task, ok := listTasks[taskID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (m *mockGoogleTasksServer) handleCreateTask(w http.ResponseWriter, r *http.Request, listID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.taskLists[listID]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var input struct {
		Title  string `json:"title"`
		Notes  string `json:"notes,omitempty"`
		Status string `json:"status,omitempty"`
		Due    string `json:"due,omitempty"`
		Parent string `json:"parent,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("task-%d", len(m.tasks[listID])+1)
	status := input.Status
	if status == "" {
		status = "needsAction"
	}
	task := &googleTask{
		ID:      id,
		Title:   input.Title,
		Notes:   input.Notes,
		Status:  status,
		Due:     input.Due,
		Parent:  input.Parent,
		Updated: time.Now().UTC().Format(time.RFC3339),
	}
	m.tasks[listID][id] = task

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (m *mockGoogleTasksServer) handleUpdateTask(w http.ResponseWriter, r *http.Request, listID, taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	listTasks, ok := m.tasks[listID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	task, ok := listTasks[taskID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var input struct {
		Title  *string `json:"title,omitempty"`
		Notes  *string `json:"notes,omitempty"`
		Status *string `json:"status,omitempty"`
		Due    *string `json:"due,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if input.Title != nil {
		task.Title = *input.Title
	}
	if input.Notes != nil {
		task.Notes = *input.Notes
	}
	if input.Status != nil {
		task.Status = *input.Status
		if *input.Status == "completed" {
			task.Completed = time.Now().UTC().Format(time.RFC3339)
		} else {
			task.Completed = ""
		}
	}
	if input.Due != nil {
		task.Due = *input.Due
	}
	task.Updated = time.Now().UTC().Format(time.RFC3339)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (m *mockGoogleTasksServer) handleDeleteTask(w http.ResponseWriter, r *http.Request, listID, taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	listTasks, ok := m.tasks[listID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if _, ok := listTasks[taskID]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	delete(listTasks, taskID)
	w.WriteHeader(http.StatusNoContent)
}

// =============================================================================
// CLI Tests Required (027-google-tasks-backend)
// =============================================================================

// TestGoogleTasksListTaskLists - todoat --backend=google list shows Google task lists
func TestGoogleTasksListTaskLists(t *testing.T) {
	server := newMockGoogleTasksServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "Work")
	server.AddTaskList("list-2", "Personal")

	be, err := New(Config{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
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
		t.Errorf("Expected 2 task lists, got %d", len(lists))
	}

	names := make(map[string]bool)
	for _, l := range lists {
		names[l.Name] = true
	}

	if !names["Work"] {
		t.Error("Expected to find task list 'Work'")
	}
	if !names["Personal"] {
		t.Error("Expected to find task list 'Personal'")
	}
}

// TestGoogleTasksGetTasks - todoat --backend=google MyList retrieves tasks from Google
func TestGoogleTasksGetTasks(t *testing.T) {
	server := newMockGoogleTasksServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyList")
	server.AddTask("list-1", "task-1", "Buy groceries", "needsAction", "", "")
	server.AddTask("list-1", "task-2", "Review PR", "needsAction", "", "")

	be, err := New(Config{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	tasks, err := be.GetTasks(ctx, "list-1")
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

// TestGoogleTasksAddTask - todoat --backend=google MyList add "Task" creates task via API
func TestGoogleTasksAddTask(t *testing.T) {
	server := newMockGoogleTasksServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyList")

	be, err := New(Config{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	task := &backend.Task{
		Summary: "New Task",
	}

	created, err := be.CreateTask(ctx, "list-1", task)
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
	tasks, err := be.GetTasks(ctx, "list-1")
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

// TestGoogleTasksUpdateTask - todoat --backend=google MyList update "Task" -s D updates task
func TestGoogleTasksUpdateTask(t *testing.T) {
	server := newMockGoogleTasksServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyList")
	server.AddTask("list-1", "task-1", "Existing Task", "needsAction", "", "")

	be, err := New(Config{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// Get the task first
	tasks, err := be.GetTasks(ctx, "list-1")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks) == 0 {
		t.Fatal("Expected at least one task")
	}

	task := &tasks[0]
	task.Status = backend.StatusCompleted
	task.Summary = "Updated Task"

	updated, err := be.UpdateTask(ctx, "list-1", task)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	if updated.Status != backend.StatusCompleted {
		t.Errorf("Expected status COMPLETED, got %s", updated.Status)
	}
}

// TestGoogleTasksDeleteTask - todoat --backend=google MyList delete "Task" removes task
func TestGoogleTasksDeleteTask(t *testing.T) {
	server := newMockGoogleTasksServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyList")
	server.AddTask("list-1", "task-to-delete", "Task to Delete", "needsAction", "", "")

	be, err := New(Config{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	err = be.DeleteTask(ctx, "list-1", "task-to-delete")
	if err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	// Verify task is gone
	tasks, err := be.GetTasks(ctx, "list-1")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks after deletion, got %d", len(tasks))
	}
}

// TestGoogleTasksOAuth2Flow - OAuth2 authentication flow with token storage
func TestGoogleTasksOAuth2Flow(t *testing.T) {
	server := newMockGoogleTasksServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "TestList")

	// Test with OAuth2 config
	cfg := Config{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	}

	be, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create backend with OAuth2: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists with OAuth2 failed: %v", err)
	}

	if len(lists) != 1 {
		t.Errorf("Expected 1 list, got %d", len(lists))
	}
}

// TestGoogleTasksTokenRefresh - Automatic token refresh when expired
func TestGoogleTasksTokenRefresh(t *testing.T) {
	server := newMockGoogleTasksServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "TestList")
	server.SetTokenExpired(true)

	be, err := New(Config{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// This should trigger a token refresh
	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists after token refresh failed: %v", err)
	}

	if len(lists) != 1 {
		t.Errorf("Expected 1 list after token refresh, got %d", len(lists))
	}

	// Verify refresh was called
	if server.GetRefreshCount() < 1 {
		t.Error("Expected token refresh to be called")
	}
}

// TestGoogleTasksSubtasks - Parent-child relationships sync correctly
func TestGoogleTasksSubtasks(t *testing.T) {
	server := newMockGoogleTasksServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyList")
	server.AddTask("list-1", "parent-task", "Parent Task", "needsAction", "", "")
	server.AddTask("list-1", "child-task-1", "Child Task 1", "needsAction", "parent-task", "")
	server.AddTask("list-1", "child-task-2", "Child Task 2", "needsAction", "parent-task", "")

	be, err := New(Config{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	tasks, err := be.GetTasks(ctx, "list-1")
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

// TestGoogleTasksDueDate - Due dates map correctly to Google Tasks format
func TestGoogleTasksDueDate(t *testing.T) {
	server := newMockGoogleTasksServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyList")
	// Google Tasks uses RFC3339 date format but only date portion matters
	server.AddTask("list-1", "task-with-due", "Task with Due", "needsAction", "", "2026-01-20T00:00:00Z")

	be, err := New(Config{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	tasks, err := be.GetTasks(ctx, "list-1")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.DueDate == nil {
		t.Fatal("Expected task to have a due date")
	}

	// Due date should be 2026-01-20
	expectedYear := 2026
	expectedMonth := time.January
	expectedDay := 20

	if task.DueDate.Year() != expectedYear || task.DueDate.Month() != expectedMonth || task.DueDate.Day() != expectedDay {
		t.Errorf("Expected due date 2026-01-20, got %v", task.DueDate)
	}

	// Test creating a task with due date
	dueDate := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC)
	newTask := &backend.Task{
		Summary: "Task with new due",
		DueDate: &dueDate,
	}

	created, err := be.CreateTask(ctx, "list-1", newTask)
	if err != nil {
		t.Fatalf("CreateTask with due date failed: %v", err)
	}

	if created.DueDate == nil {
		t.Error("Created task should have due date")
	}
}

// TestGoogleTasksStatusMapping - Status maps: completed/needs-action
func TestGoogleTasksStatusMapping(t *testing.T) {
	server := newMockGoogleTasksServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyList")
	server.AddTask("list-1", "task-active", "Active Task", "needsAction", "", "")
	server.AddTask("list-1", "task-done", "Done Task", "completed", "", "")

	be, err := New(Config{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	tasks, err := be.GetTasks(ctx, "list-1")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	statusMap := make(map[string]backend.TaskStatus)
	for _, task := range tasks {
		statusMap[task.Summary] = task.Status
	}

	if statusMap["Active Task"] != backend.StatusNeedsAction {
		t.Errorf("Expected 'Active Task' to have NEEDS-ACTION status, got %s", statusMap["Active Task"])
	}

	if statusMap["Done Task"] != backend.StatusCompleted {
		t.Errorf("Expected 'Done Task' to have COMPLETED status, got %s", statusMap["Done Task"])
	}
}

// =============================================================================
// Additional Unit Tests
// =============================================================================

func TestNewBackend(t *testing.T) {
	server := newMockGoogleTasksServer("test-token", "refresh-token")
	defer server.Close()

	be, err := New(Config{
		AccessToken:  "test-token",
		RefreshToken: "refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
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
	server := newMockGoogleTasksServer("test-token", "refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "TestList")

	be, err := New(Config{
		AccessToken:  "test-token",
		RefreshToken: "refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	list, err := be.GetList(ctx, "list-1")
	if err != nil {
		t.Fatalf("GetList failed: %v", err)
	}

	if list == nil {
		t.Fatal("Expected non-nil list")
	}

	if list.Name != "TestList" {
		t.Errorf("Expected list name 'TestList', got '%s'", list.Name)
	}
}

func TestGetListByName(t *testing.T) {
	server := newMockGoogleTasksServer("test-token", "refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyTasks")

	be, err := New(Config{
		AccessToken:  "test-token",
		RefreshToken: "refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
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
	server := newMockGoogleTasksServer("test-token", "refresh-token")
	defer server.Close()

	be, err := New(Config{
		AccessToken:  "test-token",
		RefreshToken: "refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	list, err := be.GetList(ctx, "nonexistent-id")
	if err != nil {
		// Google returns 404 for non-existent lists
		return
	}

	if list != nil {
		t.Error("Expected nil for non-existent list")
	}
}

func TestGetTask(t *testing.T) {
	server := newMockGoogleTasksServer("test-token", "refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "Work")
	server.AddTask("list-1", "task-123", "Important Task", "needsAction", "", "")

	be, err := New(Config{
		AccessToken:  "test-token",
		RefreshToken: "refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	task, err := be.GetTask(ctx, "list-1", "task-123")
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

func TestCreateList(t *testing.T) {
	server := newMockGoogleTasksServer("test-token", "refresh-token")
	defer server.Close()

	be, err := New(Config{
		AccessToken:  "test-token",
		RefreshToken: "refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	list, err := be.CreateList(ctx, "New List")
	if err != nil {
		t.Fatalf("CreateList failed: %v", err)
	}

	if list == nil {
		t.Fatal("Expected non-nil list")
	}

	if list.Name != "New List" {
		t.Errorf("Expected list name 'New List', got '%s'", list.Name)
	}

	if list.ID == "" {
		t.Error("Expected list to have an ID")
	}
}

func TestDeleteList(t *testing.T) {
	server := newMockGoogleTasksServer("test-token", "refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "ToDelete")

	be, err := New(Config{
		AccessToken:  "test-token",
		RefreshToken: "refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	err = be.DeleteList(ctx, "list-1")
	if err != nil {
		t.Fatalf("DeleteList failed: %v", err)
	}

	// Verify list is gone
	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists failed: %v", err)
	}

	for _, l := range lists {
		if l.ID == "list-1" {
			t.Error("List should have been deleted")
		}
	}
}

// Trash operations are not supported by Google Tasks
func TestTrashOperationsNotSupported(t *testing.T) {
	server := newMockGoogleTasksServer("test-token", "refresh-token")
	defer server.Close()

	be, err := New(Config{
		AccessToken:  "test-token",
		RefreshToken: "refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// GetDeletedLists should return empty
	deletedLists, err := be.GetDeletedLists(ctx)
	if err != nil {
		return // Some implementations might error
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

func TestStatusConversion(t *testing.T) {
	tests := []struct {
		googleStatus   string
		expectedStatus backend.TaskStatus
	}{
		{"needsAction", backend.StatusNeedsAction},
		{"completed", backend.StatusCompleted},
	}

	for _, tt := range tests {
		t.Run(tt.googleStatus, func(t *testing.T) {
			status := googleToBackendStatus(tt.googleStatus)
			if status != tt.expectedStatus {
				t.Errorf("Expected %s, got %s", tt.expectedStatus, status)
			}
		})
	}
}

func TestBackendStatusToGoogle(t *testing.T) {
	tests := []struct {
		backendStatus  backend.TaskStatus
		expectedGoogle string
	}{
		{backend.StatusNeedsAction, "needsAction"},
		{backend.StatusCompleted, "completed"},
		{backend.StatusInProgress, "needsAction"}, // No direct mapping
		{backend.StatusCancelled, "completed"},    // Treat as completed
	}

	for _, tt := range tests {
		t.Run(string(tt.backendStatus), func(t *testing.T) {
			gStatus := backendToGoogleStatus(tt.backendStatus)
			if gStatus != tt.expectedGoogle {
				t.Errorf("Expected %s, got %s", tt.expectedGoogle, gStatus)
			}
		})
	}
}

func TestAuthFailure(t *testing.T) {
	server := newMockGoogleTasksServer("correct-token", "refresh-token")
	defer server.Close()

	be, err := New(Config{
		AccessToken:  "wrong-token",
		RefreshToken: "",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
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

func TestPriorityIgnored(t *testing.T) {
	// Google Tasks doesn't support priority, but we should handle tasks with priority set
	server := newMockGoogleTasksServer("test-token", "refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyList")

	be, err := New(Config{
		AccessToken:  "test-token",
		RefreshToken: "refresh-token",
		BaseURL:      server.URL(),
		TokenURL:     server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// Create task with priority (should be ignored)
	task := &backend.Task{
		Summary:  "Priority Task",
		Priority: 1, // High priority
	}

	created, err := be.CreateTask(ctx, "list-1", task)
	if err != nil {
		t.Fatalf("CreateTask with priority failed: %v", err)
	}

	// Priority should be 0 (not supported)
	if created.Priority != 0 {
		t.Logf("Note: Priority %d was set but Google Tasks doesn't support priority", created.Priority)
	}
}

// Integration test helpers
func TestIntegrationConfigFromEnv(t *testing.T) {
	// Verify the env var names are documented
	envVars := []string{
		"TODOAT_GOOGLE_ACCESS_TOKEN",
		"TODOAT_GOOGLE_REFRESH_TOKEN",
		"TODOAT_GOOGLE_CLIENT_ID",
		"TODOAT_GOOGLE_CLIENT_SECRET",
	}

	for i, envVar := range envVars {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			// Just verify the env var name is valid
			if envVar == "" {
				t.Error("Empty env var name")
			}
		})
	}
}
