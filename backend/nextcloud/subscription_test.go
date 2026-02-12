package nextcloud

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"todoat/backend"
)

// =============================================================================
// CalDAV Subscription Tests (Issue #94)
// =============================================================================

// mockSubscriptionServer extends the CalDAV mock to handle MKCALENDAR requests for subscriptions
type mockSubscriptionServer struct {
	server        *httptest.Server
	calendars     map[string]*mockCalendar
	subscriptions map[string]string // calendarName -> source URL
	username      string
	password      string
}

func newMockSubscriptionServer(username, password string) *mockSubscriptionServer {
	m := &mockSubscriptionServer{
		calendars:     make(map[string]*mockCalendar),
		subscriptions: make(map[string]string),
		username:      username,
		password:      password,
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.handler))
	return m
}

func (m *mockSubscriptionServer) Close() {
	m.server.Close()
}

func (m *mockSubscriptionServer) URL() string {
	return m.server.URL
}

func (m *mockSubscriptionServer) AddCalendar(name string) {
	m.calendars[name] = &mockCalendar{
		name:           name,
		tasks:          make(map[string]string),
		ctag:           "ctag-1",
		supportedComps: []string{"VTODO", "VEVENT"},
	}
}

func (m *mockSubscriptionServer) handler(w http.ResponseWriter, r *http.Request) {
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
		m.handlePropfind(w, path)
	case "MKCALENDAR":
		m.handleMkcalendar(w, r, path)
	case "DELETE":
		m.handleDelete(w, path)
	case "REPORT":
		m.handleReport(w, path)
	case "PUT":
		// Reject writes to subscribed calendars
		for name := range m.subscriptions {
			if strings.Contains(path, "/"+name+"/") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}
		// Normal PUT for non-subscribed calendars
		for name, cal := range m.calendars {
			if strings.Contains(path, "/"+name+"/") {
				parts := strings.Split(path, "/")
				if len(parts) > 0 {
					icsFile := parts[len(parts)-1]
					uid := strings.TrimSuffix(icsFile, ".ics")
					buf := make([]byte, 4096)
					n, _ := r.Body.Read(buf)
					cal.tasks[uid] = string(buf[:n])
					w.WriteHeader(http.StatusCreated)
					return
				}
			}
		}
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (m *mockSubscriptionServer) handlePropfind(w http.ResponseWriter, path string) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	userCalendarPath := fmt.Sprintf("/remote.php/dav/calendars/%s/", m.username)
	if path == userCalendarPath || strings.HasSuffix(path, fmt.Sprintf("/calendars/%s/", m.username)) || strings.HasSuffix(path, fmt.Sprintf("/calendars/%s", m.username)) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<d:multistatus xmlns:d="DAV:" xmlns:cs="http://calendarserver.org/ns/" xmlns:cal="urn:ietf:params:xml:ns:caldav" xmlns:oc="http://owncloud.org/ns">`
		for name, cal := range m.calendars {
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

	w.WriteHeader(http.StatusNotFound)
}

func (m *mockSubscriptionServer) handleMkcalendar(w http.ResponseWriter, r *http.Request, path string) {
	// Extract calendar name from path: /remote.php/dav/calendars/user/calname/
	prefix := fmt.Sprintf("/remote.php/dav/calendars/%s/", m.username)
	if !strings.HasPrefix(path, prefix) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	calName := strings.TrimPrefix(path, prefix)
	calName = strings.TrimSuffix(calName, "/")

	if calName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if calendar already exists
	if _, exists := m.calendars[calName]; exists {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// Parse body to extract source URL
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	bodyStr := string(body)

	// Extract source URL from the MKCALENDAR body
	sourceURL := ""
	if strings.Contains(bodyStr, "cs:source") {
		// Extract URL from <d:href>URL</d:href> inside <cs:source>
		idx := strings.Index(bodyStr, "cs:source")
		if idx >= 0 {
			rest := bodyStr[idx:]
			hrefStart := strings.Index(rest, "<d:href>")
			if hrefStart >= 0 {
				rest = rest[hrefStart+8:]
				hrefEnd := strings.Index(rest, "</d:href>")
				if hrefEnd >= 0 {
					sourceURL = rest[:hrefEnd]
				}
			}
		}
	}

	// Create the calendar
	m.calendars[calName] = &mockCalendar{
		name:           calName,
		tasks:          make(map[string]string),
		ctag:           "ctag-1",
		supportedComps: []string{"VTODO"},
	}

	if sourceURL != "" {
		m.subscriptions[calName] = sourceURL
	}

	w.WriteHeader(http.StatusCreated)
}

func (m *mockSubscriptionServer) handleDelete(w http.ResponseWriter, path string) {
	prefix := fmt.Sprintf("/remote.php/dav/calendars/%s/", m.username)
	if !strings.HasPrefix(path, prefix) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	calName := strings.TrimPrefix(path, prefix)
	calName = strings.TrimSuffix(calName, "/")

	// Check if this is a calendar-level DELETE (unsubscribe)
	if _, exists := m.calendars[calName]; exists {
		delete(m.calendars, calName)
		delete(m.subscriptions, calName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (m *mockSubscriptionServer) handleReport(w http.ResponseWriter, path string) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

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

// TestListSubscribeNextcloud - todoat list subscribe "https://example.com/calendar/ical" adds subscription
func TestListSubscribeNextcloud(t *testing.T) {
	server := newMockSubscriptionServer("testuser", "testpass")
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

	// Subscribe to an external calendar feed
	list, err := be.SubscribeList(ctx, "https://example.com/calendar/ical")
	if err != nil {
		t.Fatalf("SubscribeList failed: %v", err)
	}

	if list == nil {
		t.Fatal("Expected non-nil list from subscription")
	}

	if list.Name == "" {
		t.Error("Expected subscription list to have a name")
	}

	// Verify subscription was recorded on server
	if len(server.subscriptions) != 1 {
		t.Errorf("Expected 1 subscription on server, got %d", len(server.subscriptions))
	}

	// Verify the source URL was set
	for _, srcURL := range server.subscriptions {
		if srcURL != "https://example.com/calendar/ical" {
			t.Errorf("Expected source URL 'https://example.com/calendar/ical', got %q", srcURL)
		}
	}
}

// TestListSubscribeList - Subscribed lists appear in todoat list
func TestListSubscribeList(t *testing.T) {
	server := newMockSubscriptionServer("testuser", "testpass")
	defer server.Close()

	// Add a normal calendar
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

	// Subscribe to external calendar
	_, err = be.SubscribeList(ctx, "https://example.com/calendar/ical")
	if err != nil {
		t.Fatalf("SubscribeList failed: %v", err)
	}

	// List all calendars - subscribed should appear
	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists failed: %v", err)
	}

	// Should have the normal calendar + the subscription
	if len(lists) < 2 {
		t.Errorf("Expected at least 2 lists (normal + subscription), got %d", len(lists))
		for _, l := range lists {
			t.Logf("  - %s (ID: %s)", l.Name, l.ID)
		}
	}
}

// TestListSubscribeReadOnly - Cannot modify tasks in subscribed lists
func TestListSubscribeReadOnly(t *testing.T) {
	server := newMockSubscriptionServer("testuser", "testpass")
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

	// Subscribe to external calendar
	list, err := be.SubscribeList(ctx, "https://example.com/calendar/ical")
	if err != nil {
		t.Fatalf("SubscribeList failed: %v", err)
	}

	// Try to create a task in the subscribed list - should fail
	task := &backend.Task{
		Summary: "Should Not Work",
		Status:  backend.StatusNeedsAction,
	}

	_, err = be.CreateTask(ctx, list.ID, task)
	if err == nil {
		t.Error("Expected error when creating task in subscribed (read-only) list, but got nil")
	}
}

// TestListUnsubscribe - Can remove a subscription
func TestListUnsubscribe(t *testing.T) {
	server := newMockSubscriptionServer("testuser", "testpass")
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

	// Subscribe first
	list, err := be.SubscribeList(ctx, "https://example.com/calendar/ical")
	if err != nil {
		t.Fatalf("SubscribeList failed: %v", err)
	}

	// Verify the subscribed calendar exists
	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists failed: %v", err)
	}
	initialCount := len(lists)

	// Unsubscribe
	err = be.UnsubscribeList(ctx, list.ID)
	if err != nil {
		t.Fatalf("UnsubscribeList failed: %v", err)
	}

	// Verify the calendar was removed
	lists, err = be.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists after unsubscribe failed: %v", err)
	}

	if len(lists) != initialCount-1 {
		t.Errorf("Expected %d lists after unsubscribe, got %d", initialCount-1, len(lists))
	}

	// Verify it's gone from the server
	if len(server.subscriptions) != 0 {
		t.Errorf("Expected 0 subscriptions on server after unsubscribe, got %d", len(server.subscriptions))
	}
}

// TestSubscribeInvalidURL - Validates URL format before attempting subscription
func TestSubscribeInvalidURL(t *testing.T) {
	server := newMockSubscriptionServer("testuser", "testpass")
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

	// Test with empty URL
	_, err = be.SubscribeList(ctx, "")
	if err == nil {
		t.Error("SubscribeList should fail with empty URL")
	}

	// Test with invalid URL (no scheme)
	_, err = be.SubscribeList(ctx, "not-a-url")
	if err == nil {
		t.Error("SubscribeList should fail with invalid URL")
	}
}

// TestSubscribeURLNameDerivation - Subscription creates a sensibly named calendar
func TestSubscribeURLNameDerivation(t *testing.T) {
	server := newMockSubscriptionServer("testuser", "testpass")
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

	// Subscribe with a URL
	list, err := be.SubscribeList(ctx, "https://example.com/my-calendar/feed.ics")
	if err != nil {
		t.Fatalf("SubscribeList failed: %v", err)
	}

	// The calendar name/ID should be derived from the URL
	if list.ID == "" {
		t.Error("Expected non-empty calendar ID")
	}

	if list.ID != "feed" {
		t.Errorf("Expected calendar ID 'feed', got %q", list.ID)
	}
}

// TestUnsubscribeNonExistent - Unsubscribe from non-existent calendar returns error
func TestUnsubscribeNonExistent(t *testing.T) {
	server := newMockSubscriptionServer("testuser", "testpass")
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

	err = be.UnsubscribeList(ctx, "nonexistent-calendar")
	if err == nil {
		t.Error("UnsubscribeList should fail for non-existent calendar")
	}
}

// TestSubscribeListImplementsInterface - Compile-time check
func TestSubscribeListImplementsInterface(t *testing.T) {
	var _ backend.ListSubscriber = (*Backend)(nil)
}

// TestSubscribeDuplicateURL - Subscribing with same URL derived name that already exists
func TestSubscribeDuplicateURL(t *testing.T) {
	server := newMockSubscriptionServer("testuser", "testpass")
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

	// Subscribe once
	_, err = be.SubscribeList(ctx, "https://example.com/calendar/feed.ics")
	if err != nil {
		t.Fatalf("First SubscribeList failed: %v", err)
	}

	// Subscribe again with same URL - should fail (conflict)
	_, err = be.SubscribeList(ctx, "https://example.com/calendar/feed.ics")
	if err == nil {
		t.Error("SubscribeList with duplicate URL should fail")
	}
}

// TestDeriveCalendarName - Unit test for deriving calendar name from URL
func TestDeriveCalendarName(t *testing.T) {
	tests := []struct {
		inputURL string
		expected string
	}{
		{"https://example.com/my-calendar.ics", "my-calendar"},
		{"https://example.com/path/feed.ics", "feed"},
		{"https://example.com/path/calendar", "calendar"},
		{"https://example.com/", "subscription"},
		{"https://example.com", "subscription"},
	}

	for _, tt := range tests {
		t.Run(tt.inputURL, func(t *testing.T) {
			u, err := url.Parse(tt.inputURL)
			if err != nil {
				t.Fatalf("Failed to parse URL: %v", err)
			}
			name := deriveCalendarName(u)
			if name != tt.expected {
				t.Errorf("deriveCalendarName(%q) = %q, want %q", tt.inputURL, name, tt.expected)
			}
		})
	}
}
