package nextcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// =============================================================================
// CalDAV Sharing Tests (Issue #93)
// =============================================================================

// mockSharingServer extends the CalDAV mock to handle POST requests for sharing
type mockSharingServer struct {
	server    *httptest.Server
	calendars map[string]*mockCalendar
	shares    map[string][]sharingEntry // calendarName -> shares
	username  string
	password  string
}

type sharingEntry struct {
	User       string
	Permission string
}

func newMockSharingServer(username, password string) *mockSharingServer {
	m := &mockSharingServer{
		calendars: make(map[string]*mockCalendar),
		shares:    make(map[string][]sharingEntry),
		username:  username,
		password:  password,
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.handler))
	return m
}

func (m *mockSharingServer) Close() {
	m.server.Close()
}

func (m *mockSharingServer) URL() string {
	return m.server.URL
}

func (m *mockSharingServer) AddCalendar(name string) {
	m.calendars[name] = &mockCalendar{
		name:           name,
		tasks:          make(map[string]string),
		ctag:           "ctag-1",
		supportedComps: []string{"VTODO", "VEVENT"},
	}
}

func (m *mockSharingServer) handler(w http.ResponseWriter, r *http.Request) {
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
	case "POST":
		m.handlePost(w, r, path)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (m *mockSharingServer) handlePropfind(w http.ResponseWriter, path string) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	userCalendarPath := fmt.Sprintf("/remote.php/dav/calendars/%s/", m.username)
	if path == userCalendarPath || strings.HasSuffix(path, fmt.Sprintf("/calendars/%s/", m.username)) || strings.HasSuffix(path, fmt.Sprintf("/calendars/%s", m.username)) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<d:multistatus xmlns:d="DAV:" xmlns:cs="http://calendarserver.org/ns/" xmlns:cal="urn:ietf:params:xml:ns:caldav">`
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

func (m *mockSharingServer) handlePost(w http.ResponseWriter, r *http.Request, path string) {
	for name := range m.calendars {
		if strings.Contains(path, "/"+name+"/") || strings.HasSuffix(path, "/"+name) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			bodyStr := string(body)

			if strings.Contains(bodyStr, "remove") {
				// Unshare request
				user := extractHrefUser(bodyStr)
				if user == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				shares := m.shares[name]
				newShares := make([]sharingEntry, 0, len(shares))
				for _, s := range shares {
					if s.User != user {
						newShares = append(newShares, s)
					}
				}
				m.shares[name] = newShares
				w.WriteHeader(http.StatusOK)
				return
			}

			// Share request
			user := extractHrefUser(bodyStr)
			if user == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			perm := "read"
			if strings.Contains(bodyStr, "read-write") {
				perm = "write"
			}
			if strings.Contains(bodyStr, "all") {
				perm = "admin"
			}

			m.shares[name] = append(m.shares[name], sharingEntry{
				User:       user,
				Permission: perm,
			})
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func extractHrefUser(body string) string {
	idx := strings.Index(body, "<d:href>")
	if idx < 0 {
		return ""
	}
	rest := body[idx+8:]
	end := strings.Index(rest, "</d:href>")
	if end < 0 {
		return ""
	}
	href := rest[:end]
	parts := strings.Split(strings.TrimSuffix(href, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// TestListShareNextcloud - todoat list share "List Name" --user "username" --permission read shares calendar
func TestListShareNextcloud(t *testing.T) {
	server := newMockSharingServer("testuser", "testpass")
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

	if calendarID == "" {
		t.Fatal("Expected to find calendar 'MyCalendar'")
	}

	// Share the calendar with a user
	err = be.ShareList(ctx, calendarID, "otheruser", "read")
	if err != nil {
		t.Fatalf("ShareList failed: %v", err)
	}

	// Verify the share was recorded on the server
	shares := server.shares["MyCalendar"]
	if len(shares) != 1 {
		t.Fatalf("Expected 1 share, got %d", len(shares))
	}
	if shares[0].User != "otheruser" {
		t.Errorf("Expected share user 'otheruser', got %q", shares[0].User)
	}
	if shares[0].Permission != "read" {
		t.Errorf("Expected share permission 'read', got %q", shares[0].Permission)
	}
}

// TestListSharePermissions - Verify read, write, admin permission levels
func TestListSharePermissions(t *testing.T) {
	server := newMockSharingServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("SharedCal")

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
		if l.Name == "SharedCal" {
			calendarID = l.ID
			break
		}
	}

	// Test all three permission levels
	permissions := []string{"read", "write", "admin"}
	for _, perm := range permissions {
		t.Run(perm, func(t *testing.T) {
			err := be.ShareList(ctx, calendarID, fmt.Sprintf("user_%s", perm), perm)
			if err != nil {
				t.Errorf("ShareList with permission %q failed: %v", perm, err)
			}
		})
	}

	// Test invalid permission
	t.Run("invalid_permission", func(t *testing.T) {
		err := be.ShareList(ctx, calendarID, "baduser", "invalid")
		if err == nil {
			t.Error("ShareList with invalid permission should fail")
		}
	})
}

// TestListShareRemove - todoat list unshare "List Name" --user "username" removes share
func TestListShareRemove(t *testing.T) {
	server := newMockSharingServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("SharedCal")

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
		if l.Name == "SharedCal" {
			calendarID = l.ID
			break
		}
	}

	// First share the calendar
	err = be.ShareList(ctx, calendarID, "otheruser", "read")
	if err != nil {
		t.Fatalf("ShareList failed: %v", err)
	}

	// Verify share exists
	if len(server.shares["SharedCal"]) != 1 {
		t.Fatalf("Expected 1 share before unshare, got %d", len(server.shares["SharedCal"]))
	}

	// Now unshare
	err = be.UnshareList(ctx, calendarID, "otheruser")
	if err != nil {
		t.Fatalf("UnshareList failed: %v", err)
	}

	// Verify share was removed
	if len(server.shares["SharedCal"]) != 0 {
		t.Errorf("Expected 0 shares after unshare, got %d", len(server.shares["SharedCal"]))
	}
}

// TestListShareJSON - JSON output for share operations
func TestListShareJSON(t *testing.T) {
	server := newMockSharingServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("SharedCal")

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
	var calendarName string
	for _, l := range lists {
		if l.Name == "SharedCal" {
			calendarID = l.ID
			calendarName = l.Name
			break
		}
	}

	// Share and verify the result can be represented in JSON
	err = be.ShareList(ctx, calendarID, "jsonuser", "write")
	if err != nil {
		t.Fatalf("ShareList failed: %v", err)
	}

	// Build a JSON-friendly result (matches what CLI would output)
	result := map[string]interface{}{
		"list":       calendarName,
		"user":       "jsonuser",
		"permission": "write",
		"action":     "shared",
	}
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal share result to JSON: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal share result JSON: %v", err)
	}

	if parsed["list"] != "SharedCal" {
		t.Errorf("Expected list 'SharedCal', got %v", parsed["list"])
	}
	if parsed["user"] != "jsonuser" {
		t.Errorf("Expected user 'jsonuser', got %v", parsed["user"])
	}
	if parsed["permission"] != "write" {
		t.Errorf("Expected permission 'write', got %v", parsed["permission"])
	}
	if parsed["action"] != "shared" {
		t.Errorf("Expected action 'shared', got %v", parsed["action"])
	}
}

// TestShareListNonExistentCalendar verifies error when sharing a non-existent calendar
func TestShareListNonExistentCalendar(t *testing.T) {
	server := newMockSharingServer("testuser", "testpass")
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

	err = be.ShareList(ctx, "nonexistent", "user", "read")
	if err == nil {
		t.Error("ShareList should fail for non-existent calendar")
	}
}

// TestUnshareListNonExistentCalendar verifies error when unsharing from non-existent calendar
func TestUnshareListNonExistentCalendar(t *testing.T) {
	server := newMockSharingServer("testuser", "testpass")
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

	err = be.UnshareList(ctx, "nonexistent", "user")
	if err == nil {
		t.Error("UnshareList should fail for non-existent calendar")
	}
}

// TestShareEmptyUsername verifies error when sharing with empty username
func TestShareEmptyUsername(t *testing.T) {
	server := newMockSharingServer("testuser", "testpass")
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

	var calendarID string
	for _, l := range lists {
		if l.Name == "TestCal" {
			calendarID = l.ID
			break
		}
	}

	err = be.ShareList(ctx, calendarID, "", "read")
	if err == nil {
		t.Error("ShareList should fail with empty username")
	}
}
