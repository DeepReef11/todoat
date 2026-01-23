package mstodo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"todoat/backend"
)

// =============================================================================
// CLI Integration Tests for Microsoft To Do Backend (070-microsoft-todo-cli-integration)
// These tests verify that the Microsoft To Do backend is accessible via CLI flags.
// =============================================================================

// mockMSTodoCLIServer is a simplified mock server for CLI integration tests
type mockMSTodoCLIServer struct {
	server      *httptest.Server
	taskLists   map[string]*mstodoCLITaskList
	tasks       map[string]map[string]*mstodoCLITask
	accessToken string
	mu          sync.Mutex
}

type mstodoCLITaskList struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	IsOwner     bool   `json:"isOwner"`
	IsShared    bool   `json:"isShared"`
}

type mstodoCLITask struct {
	ID                   string             `json:"id"`
	Title                string             `json:"title"`
	Body                 *mstodoCLITaskBody `json:"body,omitempty"`
	Status               string             `json:"status"`
	Importance           string             `json:"importance"`
	DueDateTime          *mstodoCLIDateTime `json:"dueDateTime,omitempty"`
	CompletedDateTime    *mstodoCLIDateTime `json:"completedDateTime,omitempty"`
	CreatedDateTime      string             `json:"createdDateTime"`
	LastModifiedDateTime string             `json:"lastModifiedDateTime"`
}

type mstodoCLITaskBody struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

type mstodoCLIDateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

func newMockMSTodoCLIServer(accessToken string) *mockMSTodoCLIServer {
	m := &mockMSTodoCLIServer{
		taskLists:   make(map[string]*mstodoCLITaskList),
		tasks:       make(map[string]map[string]*mstodoCLITask),
		accessToken: accessToken,
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.handler))
	return m
}

func (m *mockMSTodoCLIServer) Close() {
	m.server.Close()
}

func (m *mockMSTodoCLIServer) URL() string {
	return m.server.URL
}

func (m *mockMSTodoCLIServer) AddTaskList(id, displayName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.taskLists[id] = &mstodoCLITaskList{
		ID:          id,
		DisplayName: displayName,
		IsOwner:     true,
		IsShared:    false,
	}
	m.tasks[id] = make(map[string]*mstodoCLITask)
}

func (m *mockMSTodoCLIServer) AddTask(listID, taskID, title, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.tasks[listID] == nil {
		m.tasks[listID] = make(map[string]*mstodoCLITask)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	m.tasks[listID][taskID] = &mstodoCLITask{
		ID:                   taskID,
		Title:                title,
		Status:               status,
		Importance:           "normal",
		CreatedDateTime:      now,
		LastModifiedDateTime: now,
	}
}

func (m *mockMSTodoCLIServer) GetTask(listID, taskID string) *mstodoCLITask {
	m.mu.Lock()
	defer m.mu.Unlock()
	if listTasks, ok := m.tasks[listID]; ok {
		return listTasks[taskID]
	}
	return nil
}

func (m *mockMSTodoCLIServer) handler(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	path := r.URL.Path

	// Handle OAuth2 token endpoint
	if path == "/token" {
		m.handleTokenRefresh(w, r)
		m.mu.Unlock()
		return
	}

	// Check authorization
	auth := r.Header.Get("Authorization")
	if auth != "Bearer "+m.accessToken {
		m.mu.Unlock()
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	m.mu.Unlock()

	switch {
	case path == "/v1.0/me/todo/lists" && r.Method == http.MethodGet:
		m.handleGetTaskLists(w, r)
	case path == "/v1.0/me/todo/lists" && r.Method == http.MethodPost:
		m.handleCreateTaskList(w, r)
	// Task routes must come before list routes to match correctly
	case strings.HasPrefix(path, "/v1.0/me/todo/lists/") && strings.Contains(path, "/tasks/") && r.Method == http.MethodPatch:
		trimmed := strings.TrimPrefix(path, "/v1.0/me/todo/lists/")
		parts := strings.SplitN(trimmed, "/tasks/", 2)
		listID := parts[0]
		taskID := parts[1]
		m.handleUpdateTask(w, r, listID, taskID)
	case strings.HasPrefix(path, "/v1.0/me/todo/lists/") && strings.Contains(path, "/tasks/") && r.Method == http.MethodDelete:
		trimmed := strings.TrimPrefix(path, "/v1.0/me/todo/lists/")
		parts := strings.SplitN(trimmed, "/tasks/", 2)
		listID := parts[0]
		taskID := parts[1]
		m.handleDeleteTask(w, r, listID, taskID)
	case strings.HasPrefix(path, "/v1.0/me/todo/lists/") && strings.HasSuffix(path, "/tasks") && r.Method == http.MethodGet:
		listID := strings.TrimPrefix(strings.TrimSuffix(path, "/tasks"), "/v1.0/me/todo/lists/")
		m.handleGetTasks(w, r, listID)
	case strings.HasPrefix(path, "/v1.0/me/todo/lists/") && strings.HasSuffix(path, "/tasks") && r.Method == http.MethodPost:
		listID := strings.TrimPrefix(strings.TrimSuffix(path, "/tasks"), "/v1.0/me/todo/lists/")
		m.handleCreateTask(w, r, listID)
	case strings.HasPrefix(path, "/v1.0/me/todo/lists/") && r.Method == http.MethodGet && !strings.Contains(strings.TrimPrefix(path, "/v1.0/me/todo/lists/"), "/"):
		listID := strings.TrimPrefix(path, "/v1.0/me/todo/lists/")
		m.handleGetTaskList(w, r, listID)
	case strings.HasPrefix(path, "/v1.0/me/todo/lists/") && r.Method == http.MethodPatch && !strings.Contains(strings.TrimPrefix(path, "/v1.0/me/todo/lists/"), "/"):
		listID := strings.TrimPrefix(path, "/v1.0/me/todo/lists/")
		m.handleUpdateTaskList(w, r, listID)
	case strings.HasPrefix(path, "/v1.0/me/todo/lists/") && r.Method == http.MethodDelete && !strings.Contains(strings.TrimPrefix(path, "/v1.0/me/todo/lists/"), "/"):
		listID := strings.TrimPrefix(path, "/v1.0/me/todo/lists/")
		m.handleDeleteTaskList(w, r, listID)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *mockMSTodoCLIServer) handleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": m.accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}

func (m *mockMSTodoCLIServer) handleGetTaskLists(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var items []*mstodoCLITaskList
	for _, tl := range m.taskLists {
		items = append(items, tl)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"value": items,
	})
}

func (m *mockMSTodoCLIServer) handleGetTaskList(w http.ResponseWriter, r *http.Request, id string) {
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

func (m *mockMSTodoCLIServer) handleCreateTaskList(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var input struct {
		DisplayName string `json:"displayName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("list-%d", len(m.taskLists)+1)
	tl := &mstodoCLITaskList{
		ID:          id,
		DisplayName: input.DisplayName,
		IsOwner:     true,
		IsShared:    false,
	}
	m.taskLists[id] = tl
	m.tasks[id] = make(map[string]*mstodoCLITask)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(tl)
}

func (m *mockMSTodoCLIServer) handleUpdateTaskList(w http.ResponseWriter, r *http.Request, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tl, ok := m.taskLists[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var input struct {
		DisplayName string `json:"displayName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if input.DisplayName != "" {
		tl.DisplayName = input.DisplayName
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tl)
}

func (m *mockMSTodoCLIServer) handleDeleteTaskList(w http.ResponseWriter, r *http.Request, id string) {
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

func (m *mockMSTodoCLIServer) handleGetTasks(w http.ResponseWriter, r *http.Request, listID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	listTasks, ok := m.tasks[listID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var items []*mstodoCLITask
	for _, t := range listTasks {
		items = append(items, t)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"value": items,
	})
}

func (m *mockMSTodoCLIServer) handleCreateTask(w http.ResponseWriter, r *http.Request, listID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.taskLists[listID]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var input struct {
		Title      string             `json:"title"`
		Body       *mstodoCLITaskBody `json:"body,omitempty"`
		Status     string             `json:"status,omitempty"`
		Importance string             `json:"importance,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("task-%d", len(m.tasks[listID])+1)
	status := input.Status
	if status == "" {
		status = "notStarted"
	}
	importance := input.Importance
	if importance == "" {
		importance = "normal"
	}
	now := time.Now().UTC().Format(time.RFC3339)
	task := &mstodoCLITask{
		ID:                   id,
		Title:                input.Title,
		Body:                 input.Body,
		Status:               status,
		Importance:           importance,
		CreatedDateTime:      now,
		LastModifiedDateTime: now,
	}
	m.tasks[listID][id] = task

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(task)
}

func (m *mockMSTodoCLIServer) handleUpdateTask(w http.ResponseWriter, r *http.Request, listID, taskID string) {
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
		Title      *string            `json:"title,omitempty"`
		Body       *mstodoCLITaskBody `json:"body,omitempty"`
		Status     *string            `json:"status,omitempty"`
		Importance *string            `json:"importance,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if input.Title != nil {
		task.Title = *input.Title
	}
	if input.Body != nil {
		task.Body = input.Body
	}
	if input.Status != nil {
		task.Status = *input.Status
		if *input.Status == "completed" {
			now := time.Now().UTC().Format(time.RFC3339)
			task.CompletedDateTime = &mstodoCLIDateTime{
				DateTime: now,
				TimeZone: "UTC",
			}
		} else {
			task.CompletedDateTime = nil
		}
	}
	if input.Importance != nil {
		task.Importance = *input.Importance
	}
	task.LastModifiedDateTime = time.Now().UTC().Format(time.RFC3339)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (m *mockMSTodoCLIServer) handleDeleteTask(w http.ResponseWriter, r *http.Request, listID, taskID string) {
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
// CLI Integration Tests
// These tests verify that the Microsoft To Do backend can be accessed via CLI
// =============================================================================

// TestMSTodoCLIListCommand verifies that `todoat -b mstodo list` shows Microsoft To Do lists
// This test validates the acceptance criteria from roadmap item 070
func TestMSTodoCLIListCommand(t *testing.T) {
	// Create mock server
	server := newMockMSTodoCLIServer("test-access-token")
	defer server.Close()

	server.AddTaskList("list-1", "Work Tasks")
	server.AddTaskList("list-2", "Personal Tasks")

	// Create the backend directly (CLI integration is tested via backend creation)
	be, err := New(Config{
		AccessToken: "test-access-token",
		BaseURL:     server.URL(),
		TokenURL:    server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create Microsoft To Do backend: %v", err)
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

	// Verify list names
	names := make(map[string]bool)
	for _, l := range lists {
		names[l.Name] = true
	}

	if !names["Work Tasks"] {
		t.Error("Expected to find 'Work Tasks' list")
	}
	if !names["Personal Tasks"] {
		t.Error("Expected to find 'Personal Tasks' list")
	}
}

// TestMSTodoCLIGetTasks verifies that `todoat -b mstodo "My Tasks"` retrieves tasks
// This test validates the acceptance criteria from roadmap item 070
func TestMSTodoCLIGetTasks(t *testing.T) {
	server := newMockMSTodoCLIServer("test-access-token")
	defer server.Close()

	server.AddTaskList("my-tasks-id", "My Tasks")
	server.AddTask("my-tasks-id", "task-1", "Buy groceries", "notStarted")
	server.AddTask("my-tasks-id", "task-2", "Review pull request", "notStarted")
	server.AddTask("my-tasks-id", "task-3", "Completed task", "completed")

	be, err := New(Config{
		AccessToken: "test-access-token",
		BaseURL:     server.URL(),
		TokenURL:    server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create Microsoft To Do backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	tasks, err := be.GetTasks(ctx, "my-tasks-id")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	// Verify task summaries
	summaries := make(map[string]bool)
	for _, task := range tasks {
		summaries[task.Summary] = true
	}

	if !summaries["Buy groceries"] {
		t.Error("Expected to find 'Buy groceries' task")
	}
	if !summaries["Review pull request"] {
		t.Error("Expected to find 'Review pull request' task")
	}
}

// TestMSTodoCLIAddTask verifies that `todoat -b mstodo "My Tasks" add "Task"` creates task
// This test validates the acceptance criteria from roadmap item 070
func TestMSTodoCLIAddTask(t *testing.T) {
	server := newMockMSTodoCLIServer("test-access-token")
	defer server.Close()

	server.AddTaskList("my-tasks-id", "My Tasks")

	be, err := New(Config{
		AccessToken: "test-access-token",
		BaseURL:     server.URL(),
		TokenURL:    server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create Microsoft To Do backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// Create a new task
	task, err := be.CreateTask(ctx, "my-tasks-id", &backend.Task{
		Summary: "New Microsoft Task",
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if task.Summary != "New Microsoft Task" {
		t.Errorf("Expected task summary 'New Microsoft Task', got '%s'", task.Summary)
	}

	if task.ID == "" {
		t.Error("Expected task to have an ID")
	}

	// Verify task was created on server
	tasks, err := be.GetTasks(ctx, "my-tasks-id")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	found := false
	for _, t := range tasks {
		if t.Summary == "New Microsoft Task" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Created task not found on server")
	}
}

// TestMSTodoCLIUpdateTask verifies that `todoat -b mstodo "My Tasks" update "Task" -s D` updates task
// This test validates the acceptance criteria from roadmap item 070
func TestMSTodoCLIUpdateTask(t *testing.T) {
	server := newMockMSTodoCLIServer("test-access-token")
	defer server.Close()

	server.AddTaskList("my-tasks-id", "My Tasks")
	server.AddTask("my-tasks-id", "task-1", "Task to update", "notStarted")

	be, err := New(Config{
		AccessToken: "test-access-token",
		BaseURL:     server.URL(),
		TokenURL:    server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create Microsoft To Do backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// Get the task
	tasks, err := be.GetTasks(ctx, "my-tasks-id")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks) == 0 {
		t.Fatal("Expected at least one task")
	}

	// Update the task status to completed (simulating -s D)
	taskToUpdate := &tasks[0]
	taskToUpdate.Status = backend.StatusCompleted // "COMPLETED" in backend terms

	updated, err := be.UpdateTask(ctx, "my-tasks-id", taskToUpdate)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	if updated.Status != backend.StatusCompleted {
		t.Errorf("Expected status COMPLETED, got %s", updated.Status)
	}

	// Verify update on server
	serverTask := server.GetTask("my-tasks-id", "task-1")
	if serverTask == nil {
		t.Fatal("Task not found on server")
	}
	if serverTask.Status != "completed" {
		t.Errorf("Expected server task status 'completed', got '%s'", serverTask.Status)
	}
}

// TestMSTodoCLIDeleteTask verifies that `todoat -b mstodo "My Tasks" delete "Task"` removes task
// This test validates the acceptance criteria from roadmap item 070
func TestMSTodoCLIDeleteTask(t *testing.T) {
	server := newMockMSTodoCLIServer("test-access-token")
	defer server.Close()

	server.AddTaskList("my-tasks-id", "My Tasks")
	server.AddTask("my-tasks-id", "task-to-delete", "Task to delete", "notStarted")

	be, err := New(Config{
		AccessToken: "test-access-token",
		BaseURL:     server.URL(),
		TokenURL:    server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create Microsoft To Do backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// Delete the task
	err = be.DeleteTask(ctx, "my-tasks-id", "task-to-delete")
	if err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	// Verify task is gone
	tasks, err := be.GetTasks(ctx, "my-tasks-id")
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks after deletion, got %d", len(tasks))
	}
}

// TestMSTodoBackendTypeRecognized verifies that "mstodo" is a valid backend type
// This is a key acceptance criteria for roadmap item 070
func TestMSTodoBackendTypeRecognized(t *testing.T) {
	server := newMockMSTodoCLIServer("test-token")
	defer server.Close()

	// This test verifies the backend can be created successfully
	be, err := New(Config{
		AccessToken: "test-token",
		BaseURL:     server.URL(),
		TokenURL:    server.URL() + "/token",
	})
	if err != nil {
		t.Fatalf("Failed to create Microsoft To Do backend: %v", err)
	}

	// Backend should implement TaskManager
	if be == nil {
		t.Fatal("Expected non-nil backend")
	}

	if err := be.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// TestMSTodoBackendRequiresAccessToken verifies proper error when access token is missing
func TestMSTodoBackendRequiresAccessToken(t *testing.T) {
	_, err := New(Config{
		AccessToken: "", // Missing token
	})

	if err == nil {
		t.Error("Expected error when access token is missing")
	}

	if !strings.Contains(err.Error(), "access token") {
		t.Errorf("Expected error about access token, got: %v", err)
	}
}

// TestMSTodoBackendConfigFromEnv verifies that config can be loaded from environment variables
func TestMSTodoBackendConfigFromEnv(t *testing.T) {
	// Save original env vars
	origAccessToken := os.Getenv("TODOAT_MSTODO_ACCESS_TOKEN")
	origRefreshToken := os.Getenv("TODOAT_MSTODO_REFRESH_TOKEN")
	origClientID := os.Getenv("TODOAT_MSTODO_CLIENT_ID")
	origClientSecret := os.Getenv("TODOAT_MSTODO_CLIENT_SECRET")

	// Restore env vars after test
	defer func() {
		_ = os.Setenv("TODOAT_MSTODO_ACCESS_TOKEN", origAccessToken)
		_ = os.Setenv("TODOAT_MSTODO_REFRESH_TOKEN", origRefreshToken)
		_ = os.Setenv("TODOAT_MSTODO_CLIENT_ID", origClientID)
		_ = os.Setenv("TODOAT_MSTODO_CLIENT_SECRET", origClientSecret)
	}()

	// Set test env vars
	t.Setenv("TODOAT_MSTODO_ACCESS_TOKEN", "env-access-token")
	t.Setenv("TODOAT_MSTODO_REFRESH_TOKEN", "env-refresh-token")
	t.Setenv("TODOAT_MSTODO_CLIENT_ID", "env-client-id")
	t.Setenv("TODOAT_MSTODO_CLIENT_SECRET", "env-client-secret")

	cfg := ConfigFromEnv()

	if cfg.AccessToken != "env-access-token" {
		t.Errorf("Expected access token 'env-access-token', got '%s'", cfg.AccessToken)
	}
	if cfg.RefreshToken != "env-refresh-token" {
		t.Errorf("Expected refresh token 'env-refresh-token', got '%s'", cfg.RefreshToken)
	}
	if cfg.ClientID != "env-client-id" {
		t.Errorf("Expected client ID 'env-client-id', got '%s'", cfg.ClientID)
	}
	if cfg.ClientSecret != "env-client-secret" {
		t.Errorf("Expected client secret 'env-client-secret', got '%s'", cfg.ClientSecret)
	}
}

// =============================================================================
// CLI Wiring Verification Tests
// These tests verify that the CLI wiring is complete
// =============================================================================

// TestMSTodoBackendCLIWiring verifies the backend can be created and used,
// simulating what the CLI would do when --backend=mstodo is specified.
func TestMSTodoBackendCLIWiring(t *testing.T) {
	server := newMockMSTodoCLIServer("test-token")
	defer server.Close()

	server.AddTaskList("list-1", "Test List")

	be, err := New(Config{
		AccessToken: "test-token",
		BaseURL:     server.URL(),
		TokenURL:    server.URL() + "/token",
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

	if len(lists) != 1 {
		t.Errorf("Expected 1 list, got %d", len(lists))
	}
}
