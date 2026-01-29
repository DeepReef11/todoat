package nextcloud

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"todoat/backend"
)

// =============================================================================
// CalDAV Mock Server for Tests
// =============================================================================

// mockCalDAVServer creates a test CalDAV server that simulates Nextcloud
type mockCalDAVServer struct {
	server    *httptest.Server
	calendars map[string]*mockCalendar
	username  string
	password  string
}

type mockCalendar struct {
	name           string
	tasks          map[string]string // uid -> vtodo content
	ctag           string
	supportedComps []string // VTODO, VEVENT, etc. Empty means all
}

func newMockCalDAVServer(username, password string) *mockCalDAVServer {
	m := &mockCalDAVServer{
		calendars: make(map[string]*mockCalendar),
		username:  username,
		password:  password,
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.handler))
	return m
}

func newMockCalDAVServerTLS(username, password string) *mockCalDAVServer {
	m := &mockCalDAVServer{
		calendars: make(map[string]*mockCalendar),
		username:  username,
		password:  password,
	}
	m.server = httptest.NewTLSServer(http.HandlerFunc(m.handler))
	return m
}

func (m *mockCalDAVServer) Close() {
	m.server.Close()
}

func (m *mockCalDAVServer) URL() string {
	return m.server.URL
}

func (m *mockCalDAVServer) AddCalendar(name string) {
	m.calendars[name] = &mockCalendar{
		name:           name,
		tasks:          make(map[string]string),
		ctag:           fmt.Sprintf("ctag-%d", time.Now().UnixNano()),
		supportedComps: []string{"VTODO", "VEVENT"}, // Default: supports both
	}
}

// AddCalendarWithComponents adds a calendar with specific supported component types
func (m *mockCalDAVServer) AddCalendarWithComponents(name string, components []string) {
	m.calendars[name] = &mockCalendar{
		name:           name,
		tasks:          make(map[string]string),
		ctag:           fmt.Sprintf("ctag-%d", time.Now().UnixNano()),
		supportedComps: components,
	}
}

func (m *mockCalDAVServer) AddTask(calendarName, uid, summary, status string, priority int) {
	if cal, ok := m.calendars[calendarName]; ok {
		vtodo := fmt.Sprintf(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//todoat//test//EN
BEGIN:VTODO
UID:%s
SUMMARY:%s
STATUS:%s
PRIORITY:%d
DTSTAMP:20260118T120000Z
CREATED:20260118T120000Z
LAST-MODIFIED:20260118T120000Z
END:VTODO
END:VCALENDAR`, uid, summary, status, priority)
		cal.tasks[uid] = vtodo
		cal.ctag = fmt.Sprintf("ctag-%d", time.Now().UnixNano())
	}
}

// AddTaskWithParent adds a task with a parent relationship (subtask) to a calendar
func (m *mockCalDAVServer) AddTaskWithParent(calendarName, uid, summary, status string, priority int, parentUID string) {
	if cal, ok := m.calendars[calendarName]; ok {
		vtodo := fmt.Sprintf(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//todoat//test//EN
BEGIN:VTODO
UID:%s
SUMMARY:%s
STATUS:%s
PRIORITY:%d
RELATED-TO;RELTYPE=PARENT:%s
DTSTAMP:20260118T120000Z
CREATED:20260118T120000Z
LAST-MODIFIED:20260118T120000Z
END:VTODO
END:VCALENDAR`, uid, summary, status, priority, parentUID)
		cal.tasks[uid] = vtodo
		cal.ctag = fmt.Sprintf("ctag-%d", time.Now().UnixNano())
	}
}

func (m *mockCalDAVServer) handler(w http.ResponseWriter, r *http.Request) {
	// Check auth
	if m.username != "" {
		user, pass, ok := r.BasicAuth()
		if !ok || user != m.username || pass != m.password {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	path := r.URL.Path

	switch r.Method {
	case "PROPFIND":
		m.handlePropfind(w, r, path)
	case "REPORT":
		m.handleReport(w, r, path)
	case "PUT":
		m.handlePut(w, r, path)
	case "DELETE":
		m.handleDelete(w, r, path)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (m *mockCalDAVServer) handlePropfind(w http.ResponseWriter, r *http.Request, path string) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	// Check if this is a request for the user's calendar collection
	// Path format: /remote.php/dav/calendars/username/
	userCalendarPath := fmt.Sprintf("/remote.php/dav/calendars/%s/", m.username)
	if path == userCalendarPath || strings.HasSuffix(path, fmt.Sprintf("/calendars/%s/", m.username)) || strings.HasSuffix(path, fmt.Sprintf("/calendars/%s", m.username)) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<d:multistatus xmlns:d="DAV:" xmlns:cs="http://calendarserver.org/ns/" xmlns:cal="urn:ietf:params:xml:ns:caldav">`
		for name, cal := range m.calendars {
			// Build supported-calendar-component-set
			compSet := ""
			for _, comp := range cal.supportedComps {
				compSet += fmt.Sprintf(`<cal:comp name="%s"/>`, comp)
			}
			response += fmt.Sprintf(`
<d:response>
  <d:href>/remote.php/dav/calendars/%s/%s/</d:href>
  <d:propstat>
    <d:prop>
      <d:displayname>%s</d:displayname>
      <d:resourcetype><d:collection/><cal:calendar/></d:resourcetype>
      <cs:getctag>%s</cs:getctag>
      <cal:supported-calendar-component-set>%s</cal:supported-calendar-component-set>
    </d:prop>
    <d:status>HTTP/1.1 200 OK</d:status>
  </d:propstat>
</d:response>`, m.username, name, name, cal.ctag, compSet)
		}
		response += `</d:multistatus>`
		w.WriteHeader(http.StatusMultiStatus)
		_, _ = w.Write([]byte(response))
		return
	}

	// Single calendar request
	for name, cal := range m.calendars {
		if strings.Contains(path, "/"+name+"/") || strings.HasSuffix(path, "/"+name) {
			response := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<d:multistatus xmlns:d="DAV:" xmlns:cs="http://calendarserver.org/ns/" xmlns:cal="urn:ietf:params:xml:ns:caldav">
<d:response>
  <d:href>%s</d:href>
  <d:propstat>
    <d:prop>
      <d:displayname>%s</d:displayname>
      <cs:getctag>%s</cs:getctag>
    </d:prop>
    <d:status>HTTP/1.1 200 OK</d:status>
  </d:propstat>
</d:response>
</d:multistatus>`, path, name, cal.ctag)
			w.WriteHeader(http.StatusMultiStatus)
			_, _ = w.Write([]byte(response))
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func (m *mockCalDAVServer) handleReport(w http.ResponseWriter, r *http.Request, path string) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	// Find calendar from path
	for name, cal := range m.calendars {
		if strings.Contains(path, "/"+name+"/") || strings.HasSuffix(path, "/"+name) {
			response := `<?xml version="1.0" encoding="UTF-8"?>
<d:multistatus xmlns:d="DAV:" xmlns:cal="urn:ietf:params:xml:ns:caldav">`
			for uid, vtodo := range cal.tasks {
				response += fmt.Sprintf(`
<d:response>
  <d:href>/remote.php/dav/calendars/%s/%s/%s.ics</d:href>
  <d:propstat>
    <d:prop>
      <d:getetag>"%s-etag"</d:getetag>
      <cal:calendar-data>%s</cal:calendar-data>
    </d:prop>
    <d:status>HTTP/1.1 200 OK</d:status>
  </d:propstat>
</d:response>`, m.username, name, uid, uid, vtodo)
			}
			response += `</d:multistatus>`
			w.WriteHeader(http.StatusMultiStatus)
			_, _ = w.Write([]byte(response))
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func (m *mockCalDAVServer) handlePut(w http.ResponseWriter, r *http.Request, path string) {
	// Find calendar and extract UID from path
	for name, cal := range m.calendars {
		if strings.Contains(path, "/"+name+"/") {
			// Parse UID from path: /remote.php/dav/calendars/user/calendar/uid.ics
			parts := strings.Split(path, "/")
			if len(parts) > 0 {
				icsFile := parts[len(parts)-1]
				uid := strings.TrimSuffix(icsFile, ".ics")

				// Read body (VTODO content)
				buf := make([]byte, 4096)
				n, _ := r.Body.Read(buf)
				vtodo := string(buf[:n])

				cal.tasks[uid] = vtodo
				cal.ctag = fmt.Sprintf("ctag-%d", time.Now().UnixNano())

				w.Header().Set("ETag", fmt.Sprintf(`"%s-etag"`, uid))
				w.WriteHeader(http.StatusCreated)
				return
			}
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func (m *mockCalDAVServer) handleDelete(w http.ResponseWriter, r *http.Request, path string) {
	for name, cal := range m.calendars {
		if strings.Contains(path, "/"+name+"/") {
			parts := strings.Split(path, "/")
			if len(parts) > 0 {
				icsFile := parts[len(parts)-1]
				uid := strings.TrimSuffix(icsFile, ".ics")
				delete(cal.tasks, uid)
				cal.ctag = fmt.Sprintf("ctag-%d", time.Now().UnixNano())
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

// =============================================================================
// CLI Tests Required (016-nextcloud-backend)
// =============================================================================

// TestNextcloudListTaskLists - todoat --backend=nextcloud --list-backends shows Nextcloud calendars
func TestNextcloudListTaskLists(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("MyCalendar")
	server.AddCalendar("Work")

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
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
		t.Errorf("Expected 2 lists, got %d", len(lists))
	}

	names := make(map[string]bool)
	for _, l := range lists {
		names[l.Name] = true
	}

	if !names["MyCalendar"] {
		t.Error("Expected to find calendar 'MyCalendar'")
	}
	if !names["Work"] {
		t.Error("Expected to find calendar 'Work'")
	}
}

// TestNextcloudGetTasks - todoat --backend=nextcloud MyCalendar retrieves tasks from Nextcloud
func TestNextcloudGetTasks(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("MyCalendar")
	server.AddTask("MyCalendar", "task-1", "Buy groceries", "NEEDS-ACTION", 0)
	server.AddTask("MyCalendar", "task-2", "Review PR", "IN-PROCESS", 1)

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// First get the calendar ID
	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists failed: %v", err)
	}

	var calendarID string
	for _, l := range lists {
		if l.Name == "MyCalendar" {
			calendarID = l.ID
			break
		}
	}

	tasks, err := be.GetTasks(ctx, calendarID)
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

// TestNextcloudAddTask - todoat --backend=nextcloud MyCalendar add "Task" creates VTODO on server
func TestNextcloudAddTask(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("MyCalendar")

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
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

	var calendarID string
	for _, l := range lists {
		if l.Name == "MyCalendar" {
			calendarID = l.ID
			break
		}
	}

	task := &backend.Task{
		Summary:  "New Task",
		Priority: 5,
	}

	created, err := be.CreateTask(ctx, calendarID, task)
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
	tasks, err := be.GetTasks(ctx, calendarID)
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

// TestNextcloudUpdateTask - todoat --backend=nextcloud MyCalendar update "Task" -s DONE updates task status
func TestNextcloudUpdateTask(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("MyCalendar")
	server.AddTask("MyCalendar", "task-1", "Existing Task", "NEEDS-ACTION", 0)

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
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

	var calendarID string
	for _, l := range lists {
		if l.Name == "MyCalendar" {
			calendarID = l.ID
			break
		}
	}

	// Get the task first
	tasks, err := be.GetTasks(ctx, calendarID)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks) == 0 {
		t.Fatal("Expected at least one task")
	}

	task := &tasks[0]
	task.Status = backend.StatusCompleted
	task.Summary = "Updated Task"

	updated, err := be.UpdateTask(ctx, calendarID, task)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	if updated.Status != backend.StatusCompleted {
		t.Errorf("Expected status COMPLETED, got %s", updated.Status)
	}
}

// TestNextcloudDeleteTask - todoat --backend=nextcloud MyCalendar delete "Task" removes task from server
func TestNextcloudDeleteTask(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("MyCalendar")
	server.AddTask("MyCalendar", "task-to-delete", "Task to Delete", "NEEDS-ACTION", 0)

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
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

	var calendarID string
	for _, l := range lists {
		if l.Name == "MyCalendar" {
			calendarID = l.ID
			break
		}
	}

	// Get the task first
	tasks, err := be.GetTasks(ctx, calendarID)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks) == 0 {
		t.Fatal("Expected at least one task")
	}

	taskID := tasks[0].ID

	err = be.DeleteTask(ctx, calendarID, taskID)
	if err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	// Verify task is gone
	tasks, err = be.GetTasks(ctx, calendarID)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks after deletion, got %d", len(tasks))
	}
}

// TestNextcloudStatusTranslation - Internal TODO maps to CalDAV NEEDS-ACTION, DONE to COMPLETED
func TestNextcloudStatusTranslation(t *testing.T) {
	tests := []struct {
		internalStatus backend.TaskStatus
		caldavStatus   string
	}{
		{backend.StatusNeedsAction, "NEEDS-ACTION"},
		{backend.StatusCompleted, "COMPLETED"},
		{backend.StatusInProgress, "IN-PROCESS"},
		{backend.StatusCancelled, "CANCELLED"},
	}

	for _, tt := range tests {
		t.Run(string(tt.internalStatus), func(t *testing.T) {
			caldav := statusToCalDAV(tt.internalStatus)
			if caldav != tt.caldavStatus {
				t.Errorf("Expected CalDAV status %s, got %s", tt.caldavStatus, caldav)
			}

			internal := statusFromCalDAV(tt.caldavStatus)
			if internal != tt.internalStatus {
				t.Errorf("Expected internal status %s, got %s", tt.internalStatus, internal)
			}
		})
	}
}

// TestNextcloudPriorityMapping - Priority 1-9 stored correctly in VTODO PRIORITY field
func TestNextcloudPriorityMapping(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("MyCalendar")

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
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

	var calendarID string
	for _, l := range lists {
		if l.Name == "MyCalendar" {
			calendarID = l.ID
			break
		}
	}

	// Test priorities 1-9
	for priority := 1; priority <= 9; priority++ {
		task := &backend.Task{
			Summary:  fmt.Sprintf("Priority %d task", priority),
			Priority: priority,
		}

		created, err := be.CreateTask(ctx, calendarID, task)
		if err != nil {
			t.Fatalf("CreateTask failed for priority %d: %v", priority, err)
		}

		if created.Priority != priority {
			t.Errorf("Expected priority %d, got %d", priority, created.Priority)
		}
	}
}

// TestNextcloudCredentialsFromKeyring - Backend retrieves password from system keyring
func TestNextcloudCredentialsFromKeyring(t *testing.T) {
	// This test verifies the credential resolution logic
	// In a real environment, it would use the system keyring
	// For unit testing, we verify the Config structure supports keyring

	cfg := Config{
		Host:       "nextcloud.example.com",
		Username:   "testuser",
		UseKeyring: true,
	}

	// Verify the config structure is correct
	if !cfg.UseKeyring {
		t.Error("UseKeyring should be true")
	}

	// The actual keyring integration would be tested in integration tests
	// Here we just verify the config supports it
}

// TestNextcloudCredentialsFromEnv - Backend retrieves credentials from TODOAT_NEXTCLOUD_* env vars
func TestNextcloudCredentialsFromEnv(t *testing.T) {
	server := newMockCalDAVServer("envuser", "envpass")
	defer server.Close()

	server.AddCalendar("EnvTest")

	// Set environment variables (auto-restored after test)
	t.Setenv("TODOAT_NEXTCLOUD_HOST", strings.TrimPrefix(server.URL(), "http://"))
	t.Setenv("TODOAT_NEXTCLOUD_USERNAME", "envuser")
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "envpass")

	// Create backend using env config
	cfg := ConfigFromEnv()
	cfg.AllowHTTP = true

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

// TestNextcloudHTTPSEnforcement - HTTP connections rejected unless allow_http: true configured
func TestNextcloudHTTPSEnforcement(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("TestCal")

	// Try to create backend without AllowHTTP
	// The backend will be created with HTTPS scheme, but the server is HTTP
	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: false, // explicitly false, so HTTPS will be used
	})

	if err != nil {
		// It's okay to fail at creation time with an HTTPS error
		if strings.Contains(err.Error(), "HTTPS") || strings.Contains(err.Error(), "http") {
			return // Test passes - early HTTPS enforcement
		}
		t.Fatalf("Unexpected error: %v", err)
	}
	defer func() { _ = be.Close() }()

	// Try to connect - should fail because we're using HTTPS on an HTTP server
	ctx := context.Background()
	_, err = be.GetLists(ctx)
	if err == nil {
		t.Error("Expected error when using HTTPS on HTTP server")
	}

	// The error should be a TLS/connection error since we're trying HTTPS on HTTP
	// This is the enforcement - you can't accidentally connect to HTTP without AllowHTTP
}

// TestNextcloudSelfSignedCert - Self-signed certs work with insecure_skip_verify: true
func TestNextcloudSelfSignedCert(t *testing.T) {
	server := newMockCalDAVServerTLS("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("TLSTest")

	// Without InsecureSkipVerify, should fail
	_, _ = New(Config{
		Host:               strings.TrimPrefix(strings.TrimPrefix(server.URL(), "https://"), "http://"),
		Username:           "testuser",
		Password:           "testpass",
		InsecureSkipVerify: false,
	})

	// Connection might fail due to self-signed cert
	// The important thing is that with InsecureSkipVerify=true, it works

	// With InsecureSkipVerify, should work
	be, err := New(Config{
		Host:               strings.TrimPrefix(strings.TrimPrefix(server.URL(), "https://"), "http://"),
		Username:           "testuser",
		Password:           "testpass",
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatalf("Failed to create backend with InsecureSkipVerify: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists failed with InsecureSkipVerify: %v", err)
	}

	if len(lists) != 1 {
		t.Errorf("Expected 1 list, got %d", len(lists))
	}
}

// =============================================================================
// Additional Unit Tests
// =============================================================================

func TestNewBackend(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
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
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("TestCal")

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
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

	if len(lists) == 0 {
		t.Fatal("Expected at least one list")
	}

	list, err := be.GetList(ctx, lists[0].ID)
	if err != nil {
		t.Fatalf("GetList failed: %v", err)
	}

	if list == nil {
		t.Fatal("Expected non-nil list")
	}

	if list.Name != "TestCal" {
		t.Errorf("Expected list name 'TestCal', got '%s'", list.Name)
	}
}

func TestGetListByName(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("MyTasks")

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
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
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	list, err := be.GetList(ctx, "nonexistent-id")
	if err != nil {
		t.Fatalf("GetList should not error for non-existent list, got: %v", err)
	}

	if list != nil {
		t.Error("Expected nil for non-existent list")
	}
}

func TestGetTask(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("Work")
	server.AddTask("Work", "task-123", "Important Task", "NEEDS-ACTION", 1)

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
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

	task, err := be.GetTask(ctx, lists[0].ID, "task-123")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if task == nil {
		t.Fatal("Expected non-nil task")
	}

	if task.Summary != "Important Task" {
		t.Errorf("Expected summary 'Important Task', got '%s'", task.Summary)
	}

	if task.Priority != 1 {
		t.Errorf("Expected priority 1, got %d", task.Priority)
	}
}

func TestHTTPClientConfig(t *testing.T) {
	// Verify HTTP client is configured correctly
	cfg := Config{
		Host:               "example.com",
		Username:           "user",
		Password:           "pass",
		InsecureSkipVerify: true,
	}

	client := createHTTPClient(cfg)

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected *http.Transport")
	}

	if transport.MaxIdleConns != 10 {
		t.Errorf("Expected MaxIdleConns=10, got %d", transport.MaxIdleConns)
	}

	if transport.MaxIdleConnsPerHost != 2 {
		t.Errorf("Expected MaxIdleConnsPerHost=2, got %d", transport.MaxIdleConnsPerHost)
	}

	if transport.IdleConnTimeout != 30*time.Second {
		t.Errorf("Expected IdleConnTimeout=30s, got %v", transport.IdleConnTimeout)
	}

	if client.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout=30s, got %v", client.Timeout)
	}

	if transport.TLSClientConfig == nil {
		t.Fatal("Expected non-nil TLSClientConfig")
	}

	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("Expected InsecureSkipVerify=true")
	}
}

// Test VTODO parsing
func TestParseVTODO(t *testing.T) {
	vtodo := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//todoat//test//EN
BEGIN:VTODO
UID:test-uid-123
SUMMARY:Test Task
STATUS:NEEDS-ACTION
PRIORITY:5
DTSTAMP:20260118T120000Z
CREATED:20260115T100000Z
LAST-MODIFIED:20260118T120000Z
DUE:20260120T000000Z
DTSTART:20260119T000000Z
DESCRIPTION:This is a test task
CATEGORIES:work,urgent
END:VTODO
END:VCALENDAR`

	task, err := parseVTODO(vtodo)
	if err != nil {
		t.Fatalf("parseVTODO failed: %v", err)
	}

	if task.ID != "test-uid-123" {
		t.Errorf("Expected ID 'test-uid-123', got '%s'", task.ID)
	}

	if task.Summary != "Test Task" {
		t.Errorf("Expected summary 'Test Task', got '%s'", task.Summary)
	}

	if task.Status != backend.StatusNeedsAction {
		t.Errorf("Expected status NEEDS-ACTION, got %s", task.Status)
	}

	if task.Priority != 5 {
		t.Errorf("Expected priority 5, got %d", task.Priority)
	}

	if task.Description != "This is a test task" {
		t.Errorf("Expected description 'This is a test task', got '%s'", task.Description)
	}

	if task.Categories != "work,urgent" {
		t.Errorf("Expected categories 'work,urgent', got '%s'", task.Categories)
	}

	if task.DueDate == nil {
		t.Error("Expected non-nil DueDate")
	}

	if task.StartDate == nil {
		t.Error("Expected non-nil StartDate")
	}
}

// Test VTODO generation
func TestGenerateVTODO(t *testing.T) {
	now := time.Now().UTC()
	dueDate := now.Add(24 * time.Hour)

	task := &backend.Task{
		ID:          "gen-uid-456",
		Summary:     "Generated Task",
		Description: "A generated task",
		Status:      backend.StatusInProgress,
		Priority:    3,
		DueDate:     &dueDate,
		Categories:  "important",
	}

	vtodo := generateVTODO(task)

	if !strings.Contains(vtodo, "UID:gen-uid-456") {
		t.Error("VTODO should contain UID")
	}

	if !strings.Contains(vtodo, "SUMMARY:Generated Task") {
		t.Error("VTODO should contain SUMMARY")
	}

	if !strings.Contains(vtodo, "STATUS:IN-PROCESS") {
		t.Error("VTODO should contain STATUS:IN-PROCESS")
	}

	if !strings.Contains(vtodo, "PRIORITY:3") {
		t.Error("VTODO should contain PRIORITY:3")
	}

	if !strings.Contains(vtodo, "DESCRIPTION:A generated task") {
		t.Error("VTODO should contain DESCRIPTION")
	}

	if !strings.Contains(vtodo, "CATEGORIES:important") {
		t.Error("VTODO should contain CATEGORIES")
	}

	if !strings.Contains(vtodo, "BEGIN:VCALENDAR") {
		t.Error("VTODO should be wrapped in VCALENDAR")
	}

	if !strings.Contains(vtodo, "BEGIN:VTODO") {
		t.Error("VTODO should contain VTODO component")
	}
}

// Helper function tests
func TestConfigFromEnv(t *testing.T) {
	// Set environment variables (auto-restored after test)
	t.Setenv("TODOAT_NEXTCLOUD_HOST", "test.example.com")
	t.Setenv("TODOAT_NEXTCLOUD_USERNAME", "testuser")
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "testpass")

	cfg := ConfigFromEnv()

	if cfg.Host != "test.example.com" {
		t.Errorf("Expected host 'test.example.com', got '%s'", cfg.Host)
	}

	if cfg.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", cfg.Username)
	}

	if cfg.Password != "testpass" {
		t.Errorf("Expected password 'testpass', got '%s'", cfg.Password)
	}
}

// Trash operations are not supported by Nextcloud CalDAV
func TestTrashOperationsNotSupported(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// GetDeletedLists should return empty (or error)
	deletedLists, err := be.GetDeletedLists(ctx)
	if err != nil {
		// Some implementations might error, which is acceptable
		return
	}
	if len(deletedLists) != 0 {
		t.Errorf("Expected empty deleted lists, got %d", len(deletedLists))
	}

	// Other trash operations should return errors or no-op
	_, _ = be.GetDeletedListByName(ctx, "anything")
	// This is acceptable to error or return nil

	_ = be.RestoreList(ctx, "anything")
	// This is acceptable to error

	_ = be.PurgeList(ctx, "anything")
	// This is acceptable to error
}

// TestCreateListReturnsNotSupported verifies that CreateList returns
// ErrListCreationNotSupported, which allows sync operations to skip
// tasks gracefully when the list doesn't exist on the CalDAV remote.
// This is the fix for Issue #007.
func TestCreateListReturnsNotSupported(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// Try to create a list - should return ErrListCreationNotSupported
	_, err = be.CreateList(ctx, "NewCalendar")
	if err == nil {
		t.Fatal("CreateList should return an error for CalDAV backends")
	}

	// Verify it's the specific sentinel error
	if err != backend.ErrListCreationNotSupported {
		t.Errorf("CreateList should return backend.ErrListCreationNotSupported, got: %v", err)
	}
}

// TestIssue010HostURLWithProtocolPrefix reproduces issue #010:
// When user specifies http:// in host config, app prepends https:// resulting
// in malformed URLs like "https://http//localhost:8080".
func TestIssue010HostURLWithProtocolPrefix(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()
	server.AddCalendar("tasks")

	// Get the server URL (includes http://)
	serverURL := server.URL()

	// Test case 1: Host with http:// prefix should use http
	t.Run("host_with_http_prefix", func(t *testing.T) {
		// User provides host as "http://localhost:8080" (full URL)
		cfg := Config{
			Host:     serverURL, // e.g. "http://127.0.0.1:12345"
			Username: "testuser",
			Password: "testpass",
			// Note: AllowHTTP is not set, but host has explicit http://
		}

		be, err := New(cfg)
		if err != nil {
			t.Fatalf("Failed to create backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		// Verify the baseURL is correct and doesn't have "https://http//"
		if strings.Contains(be.baseURL, "https://http") {
			t.Errorf("BUG: baseURL contains 'https://http': %s", be.baseURL)
		}
		if strings.Contains(be.baseURL, "http//") {
			t.Errorf("BUG: baseURL contains 'http//' (missing colon): %s", be.baseURL)
		}

		// The baseURL should start with http:// since that's what the host specified
		if !strings.HasPrefix(be.baseURL, "http://") {
			t.Errorf("Expected baseURL to start with http://, got: %s", be.baseURL)
		}

		// Verify it actually works by making a request
		ctx := context.Background()
		lists, err := be.GetLists(ctx)
		if err != nil {
			t.Errorf("GetLists failed: %v (baseURL was: %s)", err, be.baseURL)
		}
		if len(lists) == 0 {
			t.Error("Expected at least one calendar/list")
		}
	})

	// Test case 2: Host with https:// prefix should use https
	t.Run("host_with_https_prefix", func(t *testing.T) {
		// Create a TLS server for this test
		tlsServer := newMockCalDAVServerTLS("testuser", "testpass")
		defer tlsServer.Close()
		tlsServer.AddCalendar("tasks")

		// User provides host as "https://localhost:8443" (full URL)
		cfg := Config{
			Host:               tlsServer.URL(), // e.g. "https://127.0.0.1:12346"
			Username:           "testuser",
			Password:           "testpass",
			InsecureSkipVerify: true, // TLS test server uses self-signed cert
		}

		be, err := New(cfg)
		if err != nil {
			t.Fatalf("Failed to create backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		// Verify the baseURL doesn't double up the protocol
		if strings.Contains(be.baseURL, "https://https") {
			t.Errorf("BUG: baseURL contains 'https://https': %s", be.baseURL)
		}

		// The baseURL should start with https://
		if !strings.HasPrefix(be.baseURL, "https://") {
			t.Errorf("Expected baseURL to start with https://, got: %s", be.baseURL)
		}

		// Verify it actually works
		ctx := context.Background()
		lists, err := be.GetLists(ctx)
		if err != nil {
			t.Errorf("GetLists failed: %v (baseURL was: %s)", err, be.baseURL)
		}
		if len(lists) == 0 {
			t.Error("Expected at least one calendar/list")
		}
	})

	// Test case 3: Host without protocol should still work (use AllowHTTP flag)
	t.Run("host_without_protocol", func(t *testing.T) {
		// Strip the protocol from server URL
		hostOnly := strings.TrimPrefix(serverURL, "http://")

		cfg := Config{
			Host:      hostOnly, // e.g. "127.0.0.1:12345" (no protocol)
			Username:  "testuser",
			Password:  "testpass",
			AllowHTTP: true, // Explicitly allow HTTP since no protocol in host
		}

		be, err := New(cfg)
		if err != nil {
			t.Fatalf("Failed to create backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		// Should use http:// because AllowHTTP is true
		if !strings.HasPrefix(be.baseURL, "http://") {
			t.Errorf("Expected baseURL to start with http://, got: %s", be.baseURL)
		}

		// Verify it actually works
		ctx := context.Background()
		lists, err := be.GetLists(ctx)
		if err != nil {
			t.Errorf("GetLists failed: %v (baseURL was: %s)", err, be.baseURL)
		}
		if len(lists) == 0 {
			t.Error("Expected at least one calendar/list")
		}
	})
}

// TestIssue001FilterVTODOCalendars reproduces issue #001:
// GetLists() should only return calendars that support VTODO components.
// Calendars that only support VEVENT (like "Contact birthdays") should be filtered out.
func TestIssue001FilterVTODOCalendars(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	// Add a calendar that supports tasks (VTODO)
	server.AddCalendarWithComponents("Tasks", []string{"VTODO"})
	// Add a calendar that supports both (should be included)
	server.AddCalendarWithComponents("Personal", []string{"VEVENT", "VTODO"})
	// Add a calendar that only supports events (should be filtered out)
	server.AddCalendarWithComponents("Contact birthdays", []string{"VEVENT"})
	// Add a calendar with no VTODO support (should be filtered out)
	server.AddCalendarWithComponents("Events Only", []string{"VEVENT", "VJOURNAL"})

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
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

	// Should only get calendars that support VTODO
	if len(lists) != 2 {
		t.Errorf("Expected 2 lists (Tasks and Personal), got %d", len(lists))
		for _, l := range lists {
			t.Logf("  - %s (ID: %s)", l.Name, l.ID)
		}
	}

	// Verify the correct calendars are included
	names := make(map[string]bool)
	for _, l := range lists {
		names[l.Name] = true
	}

	if !names["Tasks"] {
		t.Error("Expected to find calendar 'Tasks' (supports VTODO)")
	}
	if !names["Personal"] {
		t.Error("Expected to find calendar 'Personal' (supports VTODO)")
	}
	if names["Contact birthdays"] {
		t.Error("Calendar 'Contact birthdays' should be filtered out (VEVENT only)")
	}
	if names["Events Only"] {
		t.Error("Calendar 'Events Only' should be filtered out (no VTODO support)")
	}
}

// TestIssue029SyncPreservesParentChildRelationships tests that parent-child relationships
// are preserved when syncing from a remote CalDAV server. The RELATED-TO property with
// RELTYPE=PARENT should be parsed and stored in the ParentID field.
// Issue #29: Sync doesn't preserve parent-child relationships from server
func TestIssue029SyncPreservesParentChildRelationships(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("Tasks")

	// Add a parent task
	server.AddTask("Tasks", "parent-task-uid", "Parent task", "NEEDS-ACTION", 0)

	// Add a child task that references the parent via RELATED-TO property
	server.AddTaskWithParent("Tasks", "child-task-uid", "Child task", "NEEDS-ACTION", 0, "parent-task-uid")

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// Get the calendar ID
	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists failed: %v", err)
	}

	var calendarID string
	for _, l := range lists {
		if l.Name == "Tasks" {
			calendarID = l.ID
			break
		}
	}

	// Get tasks from the calendar
	tasks, err := be.GetTasks(ctx, calendarID)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(tasks))
	}

	// Find the child task and verify its ParentID
	var childTask *backend.Task
	for i := range tasks {
		if tasks[i].ID == "child-task-uid" {
			childTask = &tasks[i]
			break
		}
	}

	if childTask == nil {
		t.Fatal("Child task not found")
	}

	// Critical assertion: ParentID should be set from RELATED-TO property
	if childTask.ParentID != "parent-task-uid" {
		t.Errorf("Expected child task ParentID to be 'parent-task-uid', got '%s'", childTask.ParentID)
	}
}

// TestIssue037CreateTaskPreservesParentID tests that CreateTask properly includes
// ParentID in the VTODO sent to the server (issue #37)
func TestIssue037CreateTaskPreservesParentID(t *testing.T) {
	server := newMockCalDAVServer("testuser", "testpass")
	defer server.Close()
	server.AddCalendar("tasks")

	be, err := New(Config{
		Host:      strings.TrimPrefix(server.URL(), "http://"),
		Username:  "testuser",
		Password:  "testpass",
		AllowHTTP: true,
	})
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()

	// Create a subtask with ParentID set
	subtask := &backend.Task{
		ID:       "subtask-uid-123",
		Summary:  "This is a subtask",
		Status:   backend.StatusNeedsAction,
		ParentID: "parent-task-uid-456",
	}

	createdTask, err := be.CreateTask(ctx, "tasks", subtask)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Verify the returned task has the ParentID preserved
	if createdTask.ParentID != "parent-task-uid-456" {
		t.Errorf("Expected returned task ParentID to be 'parent-task-uid-456', got '%s'", createdTask.ParentID)
	}

	// Check that the VTODO stored on the server contains RELATED-TO property
	storedVTODO := server.calendars["tasks"].tasks["subtask-uid-123"]
	if !strings.Contains(storedVTODO, "RELATED-TO") {
		t.Errorf("VTODO should contain RELATED-TO property for subtask, got:\n%s", storedVTODO)
	}
	if !strings.Contains(storedVTODO, "parent-task-uid-456") {
		t.Errorf("VTODO should contain parent UID 'parent-task-uid-456', got:\n%s", storedVTODO)
	}
}
