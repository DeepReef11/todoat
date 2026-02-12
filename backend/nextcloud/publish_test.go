package nextcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// =============================================================================
// Nextcloud OCS Share API Tests for Public Links (Issue #95)
// =============================================================================

// mockPublishServer simulates the Nextcloud OCS Share API for public link operations
type mockPublishServer struct {
	server    *httptest.Server
	calendars map[string]*mockCalendar
	shares    map[int]*ocsShare // shareID -> share
	nextID    int
	username  string
	password  string
}

// ocsShare represents a public share in the mock server
type ocsShare struct {
	ID        int
	Path      string
	ShareType int
	Token     string
	URL       string
}

func newMockPublishServer(username, password string) *mockPublishServer {
	m := &mockPublishServer{
		calendars: make(map[string]*mockCalendar),
		shares:    make(map[int]*ocsShare),
		nextID:    1,
		username:  username,
		password:  password,
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.handler))
	return m
}

func (m *mockPublishServer) Close() {
	m.server.Close()
}

func (m *mockPublishServer) URL() string {
	return m.server.URL
}

func (m *mockPublishServer) AddCalendar(name string) {
	m.calendars[name] = &mockCalendar{
		name:           name,
		tasks:          make(map[string]string),
		ctag:           "ctag-1",
		supportedComps: []string{"VTODO", "VEVENT"},
	}
}

func (m *mockPublishServer) handler(w http.ResponseWriter, r *http.Request) {
	// Check auth
	if m.username != "" {
		user, pass, ok := r.BasicAuth()
		if !ok || user != m.username || pass != m.password {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	path := r.URL.Path

	switch {
	case r.Method == "PROPFIND":
		m.handlePropfind(w, path)
	case r.Method == "POST" && strings.Contains(path, "/ocs/v2.php/apps/files_sharing/api/v1/shares"):
		m.handleCreateShare(w, r)
	case r.Method == "DELETE" && strings.Contains(path, "/ocs/v2.php/apps/files_sharing/api/v1/shares/"):
		m.handleDeleteShare(w, r, path)
	case r.Method == "GET" && strings.Contains(path, "/ocs/v2.php/apps/files_sharing/api/v1/shares"):
		m.handleGetShares(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (m *mockPublishServer) handlePropfind(w http.ResponseWriter, path string) {
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

func (m *mockPublishServer) handleCreateShare(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse form data
	params := parseFormBody(string(body))
	path := params["path"]
	shareTypeStr := params["shareType"]

	if path == "" || shareTypeStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]interface{}{
			"ocs": map[string]interface{}{
				"meta": map[string]interface{}{
					"status":     "failure",
					"statuscode": 400,
					"message":    "missing required parameters",
				},
			},
		}
		jsonBytes, _ := json.Marshal(resp)
		_, _ = w.Write(jsonBytes)
		return
	}

	shareType, _ := strconv.Atoi(shareTypeStr)

	// Only accept public link shares (type 3)
	if shareType != 3 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the calendar path exists
	calName := strings.TrimPrefix(path, "/remote.php/dav/calendars/"+m.username+"/")
	calName = strings.TrimSuffix(calName, "/")

	if _, ok := m.calendars[calName]; !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		resp := map[string]interface{}{
			"ocs": map[string]interface{}{
				"meta": map[string]interface{}{
					"status":     "failure",
					"statuscode": 404,
					"message":    "path not found",
				},
			},
		}
		jsonBytes, _ := json.Marshal(resp)
		_, _ = w.Write(jsonBytes)
		return
	}

	// Check if already published
	for _, share := range m.shares {
		if share.Path == path && share.ShareType == 3 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			resp := map[string]interface{}{
				"ocs": map[string]interface{}{
					"meta": map[string]interface{}{
						"status":     "failure",
						"statuscode": 403,
						"message":    "path is already shared via public link",
					},
				},
			}
			jsonBytes, _ := json.Marshal(resp)
			_, _ = w.Write(jsonBytes)
			return
		}
	}

	// Create the share
	shareID := m.nextID
	m.nextID++
	token := fmt.Sprintf("abc%dxyz", shareID)
	shareURL := m.server.URL + "/s/" + token

	m.shares[shareID] = &ocsShare{
		ID:        shareID,
		Path:      path,
		ShareType: 3,
		Token:     token,
		URL:       shareURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := map[string]interface{}{
		"ocs": map[string]interface{}{
			"meta": map[string]interface{}{
				"status":     "ok",
				"statuscode": 200,
				"message":    "OK",
			},
			"data": map[string]interface{}{
				"id":         shareID,
				"share_type": 3,
				"path":       path,
				"token":      token,
				"url":        shareURL,
			},
		},
	}
	jsonBytes, _ := json.Marshal(resp)
	_, _ = w.Write(jsonBytes)
}

func (m *mockPublishServer) handleDeleteShare(w http.ResponseWriter, r *http.Request, path string) {
	// Extract share ID from path: /ocs/v2.php/apps/files_sharing/api/v1/shares/{id}
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	idStr := parts[len(parts)-1]
	shareID, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, ok := m.shares[shareID]; !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		resp := map[string]interface{}{
			"ocs": map[string]interface{}{
				"meta": map[string]interface{}{
					"status":     "failure",
					"statuscode": 404,
					"message":    "share not found",
				},
			},
		}
		jsonBytes, _ := json.Marshal(resp)
		_, _ = w.Write(jsonBytes)
		return
	}

	delete(m.shares, shareID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := map[string]interface{}{
		"ocs": map[string]interface{}{
			"meta": map[string]interface{}{
				"status":     "ok",
				"statuscode": 200,
				"message":    "OK",
			},
			"data": []interface{}{},
		},
	}
	jsonBytes, _ := json.Marshal(resp)
	_, _ = w.Write(jsonBytes)
}

func (m *mockPublishServer) handleGetShares(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	var matchingShares []map[string]interface{}
	for _, share := range m.shares {
		if share.Path == path && share.ShareType == 3 {
			matchingShares = append(matchingShares, map[string]interface{}{
				"id":         share.ID,
				"share_type": 3,
				"path":       share.Path,
				"token":      share.Token,
				"url":        share.URL,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := map[string]interface{}{
		"ocs": map[string]interface{}{
			"meta": map[string]interface{}{
				"status":     "ok",
				"statuscode": 200,
				"message":    "OK",
			},
			"data": matchingShares,
		},
	}
	jsonBytes, _ := json.Marshal(resp)
	_, _ = w.Write(jsonBytes)
}

// parseFormBody parses URL-encoded form body into key-value map
func parseFormBody(body string) map[string]string {
	result := make(map[string]string)
	for _, pair := range strings.Split(body, "&") {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = strings.ReplaceAll(kv[1], "%2F", "/")
		}
	}
	return result
}

// TestListPublishNextcloud - todoat list publish "List Name" generates public URL
func TestListPublishNextcloud(t *testing.T) {
	server := newMockPublishServer("testuser", "testpass")
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

	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists failed: %v", err)
	}

	var calendarID string
	for _, l := range lists {
		if l.Name == "MyTasks" {
			calendarID = l.ID
			break
		}
	}
	if calendarID == "" {
		t.Fatal("Expected to find calendar 'MyTasks'")
	}

	// Publish the list
	publicURL, err := be.PublishList(ctx, calendarID)
	if err != nil {
		t.Fatalf("PublishList failed: %v", err)
	}

	// Verify the URL contains the expected pattern
	if publicURL == "" {
		t.Fatal("Expected a non-empty public URL")
	}
	if !strings.Contains(publicURL, "/s/") {
		t.Errorf("Expected public URL to contain '/s/', got: %s", publicURL)
	}

	// Verify the share was recorded on the server
	if len(server.shares) != 1 {
		t.Fatalf("Expected 1 share, got %d", len(server.shares))
	}
	for _, share := range server.shares {
		if share.ShareType != 3 {
			t.Errorf("Expected share type 3 (public link), got %d", share.ShareType)
		}
	}
}

// TestListPublishOutput - Output includes the public share URL
func TestListPublishOutput(t *testing.T) {
	server := newMockPublishServer("testuser", "testpass")
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

	publicURL, err := be.PublishList(ctx, calendarID)
	if err != nil {
		t.Fatalf("PublishList failed: %v", err)
	}

	// Verify the URL is well-formed with server base and token
	if !strings.HasPrefix(publicURL, server.URL()) {
		t.Errorf("Expected public URL to start with server URL %q, got: %s", server.URL(), publicURL)
	}
	if !strings.Contains(publicURL, "/s/abc") {
		t.Errorf("Expected public URL to contain a token, got: %s", publicURL)
	}
}

// TestListUnpublish - todoat list unpublish "List Name" removes public link
func TestListUnpublish(t *testing.T) {
	server := newMockPublishServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("PublishedCal")

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
		if l.Name == "PublishedCal" {
			calendarID = l.ID
			break
		}
	}

	// First publish
	_, err = be.PublishList(ctx, calendarID)
	if err != nil {
		t.Fatalf("PublishList failed: %v", err)
	}

	// Verify share exists
	if len(server.shares) != 1 {
		t.Fatalf("Expected 1 share before unpublish, got %d", len(server.shares))
	}

	// Now unpublish
	err = be.UnpublishList(ctx, calendarID)
	if err != nil {
		t.Fatalf("UnpublishList failed: %v", err)
	}

	// Verify share was removed
	if len(server.shares) != 0 {
		t.Errorf("Expected 0 shares after unpublish, got %d", len(server.shares))
	}
}

// TestListPublishJSON - JSON output includes share URL
func TestListPublishJSON(t *testing.T) {
	server := newMockPublishServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("JSONCal")

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
		if l.Name == "JSONCal" {
			calendarID = l.ID
			calendarName = l.Name
			break
		}
	}

	publicURL, err := be.PublishList(ctx, calendarID)
	if err != nil {
		t.Fatalf("PublishList failed: %v", err)
	}

	// Build a JSON-friendly result (matches what CLI would output)
	result := map[string]interface{}{
		"result": "ACTION_COMPLETED",
		"action": "published",
		"list":   calendarName,
		"url":    publicURL,
	}
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal publish result to JSON: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal publish result JSON: %v", err)
	}

	if parsed["result"] != "ACTION_COMPLETED" {
		t.Errorf("Expected result 'ACTION_COMPLETED', got %v", parsed["result"])
	}
	if parsed["action"] != "published" {
		t.Errorf("Expected action 'published', got %v", parsed["action"])
	}
	if parsed["list"] != "JSONCal" {
		t.Errorf("Expected list 'JSONCal', got %v", parsed["list"])
	}
	if parsed["url"] == nil || parsed["url"] == "" {
		t.Error("Expected url to be present in JSON output")
	}
	if urlStr, ok := parsed["url"].(string); ok {
		if !strings.Contains(urlStr, "/s/") {
			t.Errorf("Expected URL to contain '/s/', got: %s", urlStr)
		}
	}
}

// TestPublishListNonExistentCalendar verifies error when publishing a non-existent calendar
func TestPublishListNonExistentCalendar(t *testing.T) {
	server := newMockPublishServer("testuser", "testpass")
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

	_, err = be.PublishList(ctx, "nonexistent")
	if err == nil {
		t.Error("PublishList should fail for non-existent calendar")
	}
}

// TestUnpublishListNotPublished verifies error when unpublishing a list that is not published
func TestUnpublishListNotPublished(t *testing.T) {
	server := newMockPublishServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("UnpubCal")

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
		if l.Name == "UnpubCal" {
			calendarID = l.ID
			break
		}
	}

	err = be.UnpublishList(ctx, calendarID)
	if err == nil {
		t.Error("UnpublishList should fail when list is not published")
	}
	if err != nil && !strings.Contains(err.Error(), "not published") {
		t.Errorf("Expected error about list not being published, got: %v", err)
	}
}

// TestPublishListAlreadyPublished verifies error when publishing a list that's already published
func TestPublishListAlreadyPublished(t *testing.T) {
	server := newMockPublishServer("testuser", "testpass")
	defer server.Close()

	server.AddCalendar("AlreadyPub")

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
		if l.Name == "AlreadyPub" {
			calendarID = l.ID
			break
		}
	}

	// First publish should succeed
	_, err = be.PublishList(ctx, calendarID)
	if err != nil {
		t.Fatalf("First PublishList failed: %v", err)
	}

	// Second publish should fail
	_, err = be.PublishList(ctx, calendarID)
	if err == nil {
		t.Error("Second PublishList should fail when already published")
	}
	if err != nil && !strings.Contains(err.Error(), "already published") {
		t.Errorf("Expected error about list being already published, got: %v", err)
	}
}
