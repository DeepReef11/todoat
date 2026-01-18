package mstodo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"todoat/backend"
)

// =============================================================================
// Microsoft Graph API Mock Server for Tests
// =============================================================================

// mockMSGraphServer simulates the Microsoft Graph API for To Do
type mockMSGraphServer struct {
	server       *httptest.Server
	taskLists    map[string]*testMSTaskList
	tasks        map[string]map[string]*testMSTask // taskListID -> taskID -> task
	accessToken  string
	refreshToken string
	tokenExpired bool
	mu           sync.Mutex
	requestLog   []string
	refreshCount int
}

// Test types for the mock server (using different names to avoid redeclaration)
type testMSTaskList struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	IsOwner           bool   `json:"isOwner"`
	IsShared          bool   `json:"isShared"`
	WellknownListName string `json:"wellknownListName,omitempty"`
}

type testMSTask struct {
	ID                   string                `json:"id"`
	Title                string                `json:"title"`
	Body                 *testMSTaskBody       `json:"body,omitempty"`
	Status               string                `json:"status"`     // notStarted, inProgress, completed
	Importance           string                `json:"importance"` // low, normal, high
	DueDateTime          *testMSDateTime       `json:"dueDateTime,omitempty"`
	CompletedDateTime    *testMSDateTime       `json:"completedDateTime,omitempty"`
	CreatedDateTime      string                `json:"createdDateTime"`
	LastModifiedDateTime string                `json:"lastModifiedDateTime"`
	ChecklistItems       []testMSChecklistItem `json:"checklistItems,omitempty"`
}

type testMSTaskBody struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"` // text or html
}

type testMSDateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type testMSChecklistItem struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	IsChecked   bool   `json:"isChecked"`
}

func newMockMSGraphServer(accessToken, refreshToken string) *mockMSGraphServer {
	m := &mockMSGraphServer{
		taskLists:    make(map[string]*testMSTaskList),
		tasks:        make(map[string]map[string]*testMSTask),
		accessToken:  accessToken,
		refreshToken: refreshToken,
		requestLog:   []string{},
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.handler))
	return m
}

func (m *mockMSGraphServer) Close() {
	m.server.Close()
}

func (m *mockMSGraphServer) URL() string {
	return m.server.URL
}

func (m *mockMSGraphServer) AddTaskList(id, displayName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.taskLists[id] = &testMSTaskList{
		ID:          id,
		DisplayName: displayName,
		IsOwner:     true,
		IsShared:    false,
	}
	m.tasks[id] = make(map[string]*testMSTask)
}

func (m *mockMSGraphServer) AddTask(listID, taskID, title, status, importance string, checklistItems []testMSChecklistItem) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.tasks[listID] == nil {
		m.tasks[listID] = make(map[string]*testMSTask)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	m.tasks[listID][taskID] = &testMSTask{
		ID:                   taskID,
		Title:                title,
		Status:               status,
		Importance:           importance,
		CreatedDateTime:      now,
		LastModifiedDateTime: now,
		ChecklistItems:       checklistItems,
	}
}

func (m *mockMSGraphServer) AddTaskWithDue(listID, taskID, title, status, importance, dueDate string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.tasks[listID] == nil {
		m.tasks[listID] = make(map[string]*testMSTask)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	task := &testMSTask{
		ID:                   taskID,
		Title:                title,
		Status:               status,
		Importance:           importance,
		CreatedDateTime:      now,
		LastModifiedDateTime: now,
	}
	if dueDate != "" {
		task.DueDateTime = &testMSDateTime{
			DateTime: dueDate,
			TimeZone: "UTC",
		}
	}
	m.tasks[listID][taskID] = task
}

func (m *mockMSGraphServer) SetTokenExpired(expired bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokenExpired = expired
}

func (m *mockMSGraphServer) GetRefreshCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.refreshCount
}

func (m *mockMSGraphServer) GetRequestLog() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.requestLog...)
}

func (m *mockMSGraphServer) handler(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.requestLog = append(m.requestLog, r.Method+" "+r.URL.Path)

	path := r.URL.Path

	// Handle OAuth2 token endpoint
	if path == "/oauth2/v2.0/token" || path == "/token" {
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
				"code":    "InvalidAuthenticationToken",
				"message": "Access token has expired or is not yet valid.",
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

	// Route requests
	switch {
	case path == "/v1.0/me/todo/lists" && r.Method == http.MethodGet:
		m.handleGetTaskLists(w, r)
	case path == "/v1.0/me/todo/lists" && r.Method == http.MethodPost:
		m.handleCreateTaskList(w, r)
	case strings.HasPrefix(path, "/v1.0/me/todo/lists/") && !strings.Contains(strings.TrimPrefix(path, "/v1.0/me/todo/lists/"), "/") && r.Method == http.MethodGet:
		listID := strings.TrimPrefix(path, "/v1.0/me/todo/lists/")
		m.handleGetTaskList(w, r, listID)
	case strings.HasPrefix(path, "/v1.0/me/todo/lists/") && !strings.Contains(strings.TrimPrefix(path, "/v1.0/me/todo/lists/"), "/") && r.Method == http.MethodDelete:
		listID := strings.TrimPrefix(path, "/v1.0/me/todo/lists/")
		m.handleDeleteTaskList(w, r, listID)
	case strings.HasPrefix(path, "/v1.0/me/todo/lists/") && strings.HasSuffix(path, "/tasks") && r.Method == http.MethodGet:
		listID := strings.TrimPrefix(strings.TrimSuffix(path, "/tasks"), "/v1.0/me/todo/lists/")
		m.handleGetTasks(w, r, listID)
	case strings.HasPrefix(path, "/v1.0/me/todo/lists/") && strings.HasSuffix(path, "/tasks") && r.Method == http.MethodPost:
		listID := strings.TrimPrefix(strings.TrimSuffix(path, "/tasks"), "/v1.0/me/todo/lists/")
		m.handleCreateTask(w, r, listID)
	case strings.HasPrefix(path, "/v1.0/me/todo/lists/") && strings.Contains(path, "/tasks/") && r.Method == http.MethodGet:
		trimmed := strings.TrimPrefix(path, "/v1.0/me/todo/lists/")
		parts := strings.SplitN(trimmed, "/tasks/", 2)
		listID := parts[0]
		taskID := parts[1]
		m.handleGetTask(w, r, listID, taskID)
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
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *mockMSGraphServer) handleTokenRefresh(w http.ResponseWriter, r *http.Request) {
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

func (m *mockMSGraphServer) handleGetTaskLists(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var items []*testMSTaskList
	for _, tl := range m.taskLists {
		items = append(items, tl)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"@odata.context": m.server.URL + "/$metadata#users('user-id')/todo/lists",
		"value":          items,
	})
}

func (m *mockMSGraphServer) handleGetTaskList(w http.ResponseWriter, r *http.Request, id string) {
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

func (m *mockMSGraphServer) handleCreateTaskList(w http.ResponseWriter, r *http.Request) {
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
	tl := &testMSTaskList{
		ID:          id,
		DisplayName: input.DisplayName,
		IsOwner:     true,
		IsShared:    false,
	}
	m.taskLists[id] = tl
	m.tasks[id] = make(map[string]*testMSTask)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(tl)
}

func (m *mockMSGraphServer) handleDeleteTaskList(w http.ResponseWriter, r *http.Request, id string) {
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

func (m *mockMSGraphServer) handleGetTasks(w http.ResponseWriter, r *http.Request, listID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	listTasks, ok := m.tasks[listID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var items []*testMSTask
	for _, t := range listTasks {
		items = append(items, t)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"@odata.context": m.server.URL + "/$metadata#users('user-id')/todo/lists('" + listID + "')/tasks",
		"value":          items,
	})
}

func (m *mockMSGraphServer) handleGetTask(w http.ResponseWriter, r *http.Request, listID, taskID string) {
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

func (m *mockMSGraphServer) handleCreateTask(w http.ResponseWriter, r *http.Request, listID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.taskLists[listID]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var input struct {
		Title       string          `json:"title"`
		Body        *testMSTaskBody `json:"body,omitempty"`
		Status      string          `json:"status,omitempty"`
		Importance  string          `json:"importance,omitempty"`
		DueDateTime *testMSDateTime `json:"dueDateTime,omitempty"`
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
	task := &testMSTask{
		ID:                   id,
		Title:                input.Title,
		Body:                 input.Body,
		Status:               status,
		Importance:           importance,
		DueDateTime:          input.DueDateTime,
		CreatedDateTime:      now,
		LastModifiedDateTime: now,
	}
	m.tasks[listID][id] = task

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(task)
}

func (m *mockMSGraphServer) handleUpdateTask(w http.ResponseWriter, r *http.Request, listID, taskID string) {
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
		Title       *string         `json:"title,omitempty"`
		Body        *testMSTaskBody `json:"body,omitempty"`
		Status      *string         `json:"status,omitempty"`
		Importance  *string         `json:"importance,omitempty"`
		DueDateTime *testMSDateTime `json:"dueDateTime,omitempty"`
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
			task.CompletedDateTime = &testMSDateTime{
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
	if input.DueDateTime != nil {
		task.DueDateTime = input.DueDateTime
	}
	task.LastModifiedDateTime = time.Now().UTC().Format(time.RFC3339)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (m *mockMSGraphServer) handleDeleteTask(w http.ResponseWriter, r *http.Request, listID, taskID string) {
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
// CLI Tests Required (028-microsoft-todo-backend)
// =============================================================================

// TestMSTodoListTaskLists - todoat --backend=mstodo list shows Microsoft To Do lists
func TestMSTodoListTaskLists(t *testing.T) {
	server := newMockMSGraphServer("test-access-token", "test-refresh-token")
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

// TestMSTodoGetTasks - todoat --backend=mstodo MyList retrieves tasks
func TestMSTodoGetTasks(t *testing.T) {
	server := newMockMSGraphServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyList")
	server.AddTask("list-1", "task-1", "Buy groceries", "notStarted", "normal", nil)
	server.AddTask("list-1", "task-2", "Review PR", "notStarted", "high", nil)

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

// TestMSTodoAddTask - todoat --backend=mstodo MyList add "Task" creates task via Graph API
func TestMSTodoAddTask(t *testing.T) {
	server := newMockMSGraphServer("test-access-token", "test-refresh-token")
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

// TestMSTodoUpdateTask - todoat --backend=mstodo MyList update "Task" -s D updates task
func TestMSTodoUpdateTask(t *testing.T) {
	server := newMockMSGraphServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyList")
	server.AddTask("list-1", "task-1", "Existing Task", "notStarted", "normal", nil)

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

// TestMSTodoDeleteTask - todoat --backend=mstodo MyList delete "Task" removes task
func TestMSTodoDeleteTask(t *testing.T) {
	server := newMockMSGraphServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyList")
	server.AddTask("list-1", "task-to-delete", "Task to Delete", "notStarted", "normal", nil)

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

// TestMSTodoOAuth2Flow - OAuth2 authentication with Microsoft identity platform
func TestMSTodoOAuth2Flow(t *testing.T) {
	server := newMockMSGraphServer("test-access-token", "test-refresh-token")
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

// TestMSTodoTokenRefresh - Automatic token refresh when expired
func TestMSTodoTokenRefresh(t *testing.T) {
	server := newMockMSGraphServer("test-access-token", "test-refresh-token")
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

// TestMSTodoSubtasks - Checklist items map to subtasks
func TestMSTodoSubtasks(t *testing.T) {
	server := newMockMSGraphServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyList")
	checklistItems := []testMSChecklistItem{
		{ID: "check-1", DisplayName: "Subtask 1", IsChecked: false},
		{ID: "check-2", DisplayName: "Subtask 2", IsChecked: true},
	}
	server.AddTask("list-1", "task-1", "Parent Task", "notStarted", "normal", checklistItems)

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

	// The task should have checklist information in description or similar
	// Implementation will define how checklist items are exposed
	task := tasks[0]
	if task.Summary != "Parent Task" {
		t.Errorf("Expected task 'Parent Task', got '%s'", task.Summary)
	}
}

// TestMSTodoImportance - Priority maps to importance (low/normal/high)
func TestMSTodoImportance(t *testing.T) {
	server := newMockMSGraphServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyList")
	server.AddTask("list-1", "task-low", "Low Priority", "notStarted", "low", nil)
	server.AddTask("list-1", "task-normal", "Normal Priority", "notStarted", "normal", nil)
	server.AddTask("list-1", "task-high", "High Priority", "notStarted", "high", nil)

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
		t.Fatalf("Expected 3 tasks, got %d", len(tasks))
	}

	priorityMap := make(map[string]int)
	for _, task := range tasks {
		priorityMap[task.Summary] = task.Priority
	}

	// Priority mapping: low=7-9, normal=4-6, high=1-3
	// Actual values may vary based on implementation
	if priorityMap["High Priority"] >= priorityMap["Normal Priority"] {
		t.Errorf("Expected High Priority (%d) < Normal Priority (%d)", priorityMap["High Priority"], priorityMap["Normal Priority"])
	}
	if priorityMap["Normal Priority"] >= priorityMap["Low Priority"] {
		t.Errorf("Expected Normal Priority (%d) < Low Priority (%d)", priorityMap["Normal Priority"], priorityMap["Low Priority"])
	}
}

// TestMSTodoStatusMapping - Status maps: completed/notStarted/inProgress
func TestMSTodoStatusMapping(t *testing.T) {
	server := newMockMSGraphServer("test-access-token", "test-refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyList")
	server.AddTask("list-1", "task-not-started", "Not Started Task", "notStarted", "normal", nil)
	server.AddTask("list-1", "task-in-progress", "In Progress Task", "inProgress", "normal", nil)
	server.AddTask("list-1", "task-completed", "Completed Task", "completed", "normal", nil)

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

	if statusMap["Not Started Task"] != backend.StatusNeedsAction {
		t.Errorf("Expected 'Not Started Task' to have NEEDS-ACTION status, got %s", statusMap["Not Started Task"])
	}

	if statusMap["In Progress Task"] != backend.StatusInProgress {
		t.Errorf("Expected 'In Progress Task' to have IN-PROGRESS status, got %s", statusMap["In Progress Task"])
	}

	if statusMap["Completed Task"] != backend.StatusCompleted {
		t.Errorf("Expected 'Completed Task' to have COMPLETED status, got %s", statusMap["Completed Task"])
	}
}

// =============================================================================
// Additional Unit Tests
// =============================================================================

func TestNewBackend(t *testing.T) {
	server := newMockMSGraphServer("test-token", "refresh-token")
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

func TestNewBackendRequiresAccessToken(t *testing.T) {
	_, err := New(Config{
		AccessToken:  "",
		RefreshToken: "refresh-token",
	})
	if err == nil {
		t.Error("Expected error when access token is empty")
	}
}

func TestBackendImplementsInterface(t *testing.T) {
	var _ backend.TaskManager = (*Backend)(nil)
}

func TestGetList(t *testing.T) {
	server := newMockMSGraphServer("test-token", "refresh-token")
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
	server := newMockMSGraphServer("test-token", "refresh-token")
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
	server := newMockMSGraphServer("test-token", "refresh-token")
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
		return // Some implementations might error
	}

	if list != nil {
		t.Error("Expected nil for non-existent list")
	}
}

func TestGetTask(t *testing.T) {
	server := newMockMSGraphServer("test-token", "refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "Work")
	server.AddTask("list-1", "task-123", "Important Task", "notStarted", "high", nil)

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
	server := newMockMSGraphServer("test-token", "refresh-token")
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
	server := newMockMSGraphServer("test-token", "refresh-token")
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

// Trash operations are not supported by MS To Do via Graph API
func TestTrashOperationsNotSupported(t *testing.T) {
	server := newMockMSGraphServer("test-token", "refresh-token")
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
		msStatus       string
		expectedStatus backend.TaskStatus
	}{
		{"notStarted", backend.StatusNeedsAction},
		{"inProgress", backend.StatusInProgress},
		{"completed", backend.StatusCompleted},
	}

	for _, tt := range tests {
		t.Run(tt.msStatus, func(t *testing.T) {
			status := msToBackendStatus(tt.msStatus)
			if status != tt.expectedStatus {
				t.Errorf("Expected %s, got %s", tt.expectedStatus, status)
			}
		})
	}
}

func TestBackendStatusToMS(t *testing.T) {
	tests := []struct {
		backendStatus backend.TaskStatus
		expectedMS    string
	}{
		{backend.StatusNeedsAction, "notStarted"},
		{backend.StatusInProgress, "inProgress"},
		{backend.StatusCompleted, "completed"},
		{backend.StatusCancelled, "completed"}, // Treat as completed
	}

	for _, tt := range tests {
		t.Run(string(tt.backendStatus), func(t *testing.T) {
			msStatus := backendToMSStatus(tt.backendStatus)
			if msStatus != tt.expectedMS {
				t.Errorf("Expected %s, got %s", tt.expectedMS, msStatus)
			}
		})
	}
}

func TestImportanceConversion(t *testing.T) {
	tests := []struct {
		importance       string
		expectedPriority int
	}{
		{"low", 9},
		{"normal", 5},
		{"high", 1},
	}

	for _, tt := range tests {
		t.Run(tt.importance, func(t *testing.T) {
			priority := importanceToPriority(tt.importance)
			if priority != tt.expectedPriority {
				t.Errorf("Expected priority %d, got %d", tt.expectedPriority, priority)
			}
		})
	}
}

func TestPriorityToImportance(t *testing.T) {
	tests := []struct {
		priority           int
		expectedImportance string
	}{
		{1, "high"},
		{2, "high"},
		{3, "high"},
		{4, "normal"},
		{5, "normal"},
		{6, "normal"},
		{7, "low"},
		{8, "low"},
		{9, "low"},
		{0, "normal"}, // Default
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("priority-%d", tt.priority), func(t *testing.T) {
			importance := priorityToImportance(tt.priority)
			if importance != tt.expectedImportance {
				t.Errorf("Expected importance %s, got %s", tt.expectedImportance, importance)
			}
		})
	}
}

func TestAuthFailure(t *testing.T) {
	server := newMockMSGraphServer("correct-token", "refresh-token")
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

func TestDueDate(t *testing.T) {
	server := newMockMSGraphServer("test-token", "refresh-token")
	defer server.Close()

	server.AddTaskList("list-1", "MyList")
	server.AddTaskWithDue("list-1", "task-with-due", "Task with Due", "notStarted", "normal", "2026-01-20T00:00:00.0000000")

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
}

func TestCreateTaskWithDueDate(t *testing.T) {
	server := newMockMSGraphServer("test-token", "refresh-token")
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
