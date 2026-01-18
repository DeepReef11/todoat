package nextcloud

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"todoat/backend"
)

// Config holds Nextcloud connection settings
type Config struct {
	Host               string
	Username           string
	Password           string
	UseKeyring         bool
	AllowHTTP          bool
	InsecureSkipVerify bool
}

// ConfigFromEnv creates a Config from environment variables
func ConfigFromEnv() Config {
	return Config{
		Host:     os.Getenv("TODOAT_NEXTCLOUD_HOST"),
		Username: os.Getenv("TODOAT_NEXTCLOUD_USERNAME"),
		Password: os.Getenv("TODOAT_NEXTCLOUD_PASSWORD"),
	}
}

// Backend implements backend.TaskManager using Nextcloud CalDAV
type Backend struct {
	config   Config
	client   *http.Client
	baseURL  string
	username string
}

// New creates a new Nextcloud CalDAV backend
func New(cfg Config) (*Backend, error) {
	// Validate config
	if cfg.Host == "" {
		return nil, fmt.Errorf("nextcloud host is required")
	}
	if cfg.Username == "" {
		return nil, fmt.Errorf("nextcloud username is required")
	}
	if cfg.Password == "" && !cfg.UseKeyring {
		return nil, fmt.Errorf("nextcloud password is required")
	}

	// Determine protocol
	// If AllowHTTP is false, we use HTTPS (the default secure option)
	// If AllowHTTP is true, we use HTTP
	// The difference is important because HTTP sends credentials in plaintext
	scheme := "https"
	if cfg.AllowHTTP {
		scheme = "http"
	}

	baseURL := fmt.Sprintf("%s://%s/remote.php/dav/calendars/%s/", scheme, cfg.Host, cfg.Username)

	return &Backend{
		config:   cfg,
		client:   createHTTPClient(cfg),
		baseURL:  baseURL,
		username: cfg.Username,
	}, nil
}

// createHTTPClient creates an HTTP client with proper connection pooling
func createHTTPClient(cfg Config) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 2,
		IdleConnTimeout:     30 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

// Close closes the backend
func (b *Backend) Close() error {
	// Close idle connections
	if transport, ok := b.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// doRequest performs an authenticated CalDAV request
func (b *Backend) doRequest(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(b.config.Username, b.config.Password)
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	req.Header.Set("Depth", "1")

	return b.client.Do(req)
}

// GetLists returns all calendars (task lists) from Nextcloud
func (b *Backend) GetLists(ctx context.Context) ([]backend.List, error) {
	propfindBody := `<?xml version="1.0" encoding="UTF-8"?>
<d:propfind xmlns:d="DAV:" xmlns:cs="http://calendarserver.org/ns/" xmlns:cal="urn:ietf:params:xml:ns:caldav">
  <d:prop>
    <d:displayname/>
    <d:resourcetype/>
    <cs:getctag/>
  </d:prop>
</d:propfind>`

	resp, err := b.doRequest(ctx, "PROPFIND", b.baseURL, []byte(propfindBody))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusMultiStatus {
		return nil, fmt.Errorf("PROPFIND failed with status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseCalendarList(string(bodyBytes), b.username)
}

// GetList returns a specific calendar by ID
func (b *Backend) GetList(ctx context.Context, listID string) (*backend.List, error) {
	lists, err := b.GetLists(ctx)
	if err != nil {
		return nil, err
	}

	for _, l := range lists {
		if l.ID == listID {
			return &l, nil
		}
	}

	return nil, nil
}

// GetListByName returns a specific calendar by name (case-insensitive)
func (b *Backend) GetListByName(ctx context.Context, name string) (*backend.List, error) {
	lists, err := b.GetLists(ctx)
	if err != nil {
		return nil, err
	}

	for _, l := range lists {
		if strings.EqualFold(l.Name, name) {
			return &l, nil
		}
	}

	return nil, nil
}

// CreateList creates a new calendar (not commonly supported via CalDAV)
func (b *Backend) CreateList(ctx context.Context, name string) (*backend.List, error) {
	// Creating calendars via CalDAV is typically done with MKCALENDAR
	// For now, return an error as this is complex and not all servers support it
	return nil, fmt.Errorf("creating calendars is not supported via CalDAV")
}

// DeleteList soft-deletes a calendar (moves to trash)
func (b *Backend) DeleteList(ctx context.Context, listID string) error {
	// CalDAV DELETE is a hard delete, not soft delete
	return fmt.Errorf("deleting calendars is not supported via CalDAV (would be permanent)")
}

// GetDeletedLists returns deleted calendars (not supported in CalDAV)
func (b *Backend) GetDeletedLists(ctx context.Context) ([]backend.List, error) {
	// CalDAV doesn't have a trash concept
	return []backend.List{}, nil
}

// GetDeletedListByName returns a deleted calendar by name (not supported)
func (b *Backend) GetDeletedListByName(ctx context.Context, name string) (*backend.List, error) {
	return nil, nil
}

// RestoreList restores a deleted calendar (not supported)
func (b *Backend) RestoreList(ctx context.Context, listID string) error {
	return fmt.Errorf("restoring calendars is not supported via CalDAV")
}

// PurgeList permanently deletes a calendar (not supported)
func (b *Backend) PurgeList(ctx context.Context, listID string) error {
	return fmt.Errorf("purging calendars is not supported via CalDAV")
}

// GetTasks returns all tasks (VTODOs) in a calendar
func (b *Backend) GetTasks(ctx context.Context, listID string) ([]backend.Task, error) {
	calendarURL := b.baseURL + listID + "/"

	reportBody := `<?xml version="1.0" encoding="UTF-8"?>
<cal:calendar-query xmlns:d="DAV:" xmlns:cal="urn:ietf:params:xml:ns:caldav">
  <d:prop>
    <d:getetag/>
    <cal:calendar-data/>
  </d:prop>
  <cal:filter>
    <cal:comp-filter name="VCALENDAR">
      <cal:comp-filter name="VTODO"/>
    </cal:comp-filter>
  </cal:filter>
</cal:calendar-query>`

	resp, err := b.doRequest(ctx, "REPORT", calendarURL, []byte(reportBody))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusMultiStatus {
		return nil, fmt.Errorf("REPORT failed with status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseTaskList(string(bodyBytes), listID)
}

// GetTask returns a specific task by ID
func (b *Backend) GetTask(ctx context.Context, listID, taskID string) (*backend.Task, error) {
	tasks, err := b.GetTasks(ctx, listID)
	if err != nil {
		return nil, err
	}

	for _, t := range tasks {
		if t.ID == taskID {
			return &t, nil
		}
	}

	return nil, nil
}

// CreateTask creates a new VTODO on the server
func (b *Backend) CreateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	// Generate a new UID if not provided
	uid := task.ID
	if uid == "" {
		uid = uuid.New().String()
	}

	newTask := &backend.Task{
		ID:          uid,
		Summary:     task.Summary,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		DueDate:     task.DueDate,
		StartDate:   task.StartDate,
		Categories:  task.Categories,
		ListID:      listID,
		Created:     time.Now().UTC(),
		Modified:    time.Now().UTC(),
	}

	if newTask.Status == "" {
		newTask.Status = backend.StatusNeedsAction
	}

	vtodo := generateVTODO(newTask)
	taskURL := b.baseURL + listID + "/" + uid + ".ics"

	resp, err := b.doRequest(ctx, "PUT", taskURL, []byte(vtodo))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PUT failed with status %d", resp.StatusCode)
	}

	return newTask, nil
}

// UpdateTask updates an existing VTODO on the server
func (b *Backend) UpdateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	task.Modified = time.Now().UTC()
	vtodo := generateVTODO(task)
	taskURL := b.baseURL + listID + "/" + task.ID + ".ics"

	resp, err := b.doRequest(ctx, "PUT", taskURL, []byte(vtodo))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PUT failed with status %d", resp.StatusCode)
	}

	return task, nil
}

// DeleteTask removes a VTODO from the server
func (b *Backend) DeleteTask(ctx context.Context, listID, taskID string) error {
	taskURL := b.baseURL + listID + "/" + taskID + ".ics"

	resp, err := b.doRequest(ctx, "DELETE", taskURL, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("DELETE failed with status %d", resp.StatusCode)
	}

	return nil
}

// =============================================================================
// Status Conversion Functions
// =============================================================================

// statusToCalDAV converts internal status to CalDAV status
func statusToCalDAV(status backend.TaskStatus) string {
	switch status {
	case backend.StatusNeedsAction:
		return "NEEDS-ACTION"
	case backend.StatusCompleted:
		return "COMPLETED"
	case backend.StatusInProgress:
		return "IN-PROCESS"
	case backend.StatusCancelled:
		return "CANCELLED"
	default:
		return "NEEDS-ACTION"
	}
}

// statusFromCalDAV converts CalDAV status to internal status
func statusFromCalDAV(caldavStatus string) backend.TaskStatus {
	switch caldavStatus {
	case "NEEDS-ACTION":
		return backend.StatusNeedsAction
	case "COMPLETED":
		return backend.StatusCompleted
	case "IN-PROCESS":
		return backend.StatusInProgress
	case "CANCELLED":
		return backend.StatusCancelled
	default:
		return backend.StatusNeedsAction
	}
}

// =============================================================================
// VTODO Parsing and Generation
// =============================================================================

// parseVTODO parses a VTODO iCalendar component into a Task
func parseVTODO(vtodo string) (*backend.Task, error) {
	task := &backend.Task{}

	// Extract UID
	if uid := extractProperty(vtodo, "UID"); uid != "" {
		task.ID = uid
	}

	// Extract SUMMARY
	if summary := extractProperty(vtodo, "SUMMARY"); summary != "" {
		task.Summary = summary
	}

	// Extract DESCRIPTION
	if desc := extractProperty(vtodo, "DESCRIPTION"); desc != "" {
		task.Description = desc
	}

	// Extract STATUS
	if status := extractProperty(vtodo, "STATUS"); status != "" {
		task.Status = statusFromCalDAV(status)
	} else {
		task.Status = backend.StatusNeedsAction
	}

	// Extract PRIORITY
	if priorityStr := extractProperty(vtodo, "PRIORITY"); priorityStr != "" {
		if p, err := strconv.Atoi(priorityStr); err == nil {
			task.Priority = p
		}
	}

	// Extract CATEGORIES
	if categories := extractProperty(vtodo, "CATEGORIES"); categories != "" {
		task.Categories = categories
	}

	// Extract DUE
	if due := extractProperty(vtodo, "DUE"); due != "" {
		if t, err := parseCalendarDate(due); err == nil {
			task.DueDate = &t
		}
	}

	// Extract DTSTART
	if dtstart := extractProperty(vtodo, "DTSTART"); dtstart != "" {
		if t, err := parseCalendarDate(dtstart); err == nil {
			task.StartDate = &t
		}
	}

	// Extract CREATED
	if created := extractProperty(vtodo, "CREATED"); created != "" {
		if t, err := parseCalendarDate(created); err == nil {
			task.Created = t
		}
	}

	// Extract LAST-MODIFIED
	if modified := extractProperty(vtodo, "LAST-MODIFIED"); modified != "" {
		if t, err := parseCalendarDate(modified); err == nil {
			task.Modified = t
		}
	}

	// Extract COMPLETED
	if completed := extractProperty(vtodo, "COMPLETED"); completed != "" {
		if t, err := parseCalendarDate(completed); err == nil {
			task.Completed = &t
		}
	}

	return task, nil
}

// extractProperty extracts a property value from iCalendar content
func extractProperty(content, property string) string {
	// Handle properties that might have parameters (e.g., DUE;VALUE=DATE:20260120)
	pattern := regexp.MustCompile(`(?m)^` + property + `(?:;[^:]*)?:(.*)$`)
	matches := pattern.FindStringSubmatch(content)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// parseCalendarDate parses various iCalendar date formats
func parseCalendarDate(dateStr string) (time.Time, error) {
	// Try different formats
	formats := []string{
		"20060102T150405Z",
		"20060102T150405",
		"20060102",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// generateVTODO generates a VTODO iCalendar component from a Task
func generateVTODO(task *backend.Task) string {
	now := time.Now().UTC()
	dtstamp := now.Format("20060102T150405Z")

	var lines []string
	lines = append(lines, "BEGIN:VCALENDAR")
	lines = append(lines, "VERSION:2.0")
	lines = append(lines, "PRODID:-//todoat//todoat//EN")
	lines = append(lines, "BEGIN:VTODO")
	lines = append(lines, fmt.Sprintf("UID:%s", task.ID))
	lines = append(lines, fmt.Sprintf("DTSTAMP:%s", dtstamp))

	if task.Summary != "" {
		lines = append(lines, fmt.Sprintf("SUMMARY:%s", task.Summary))
	}

	if task.Description != "" {
		lines = append(lines, fmt.Sprintf("DESCRIPTION:%s", task.Description))
	}

	lines = append(lines, fmt.Sprintf("STATUS:%s", statusToCalDAV(task.Status)))

	if task.Priority > 0 {
		lines = append(lines, fmt.Sprintf("PRIORITY:%d", task.Priority))
	}

	if task.Categories != "" {
		lines = append(lines, fmt.Sprintf("CATEGORIES:%s", task.Categories))
	}

	if task.DueDate != nil {
		lines = append(lines, fmt.Sprintf("DUE:%s", task.DueDate.UTC().Format("20060102T150405Z")))
	}

	if task.StartDate != nil {
		lines = append(lines, fmt.Sprintf("DTSTART:%s", task.StartDate.UTC().Format("20060102T150405Z")))
	}

	if !task.Created.IsZero() {
		lines = append(lines, fmt.Sprintf("CREATED:%s", task.Created.UTC().Format("20060102T150405Z")))
	} else {
		lines = append(lines, fmt.Sprintf("CREATED:%s", dtstamp))
	}

	if !task.Modified.IsZero() {
		lines = append(lines, fmt.Sprintf("LAST-MODIFIED:%s", task.Modified.UTC().Format("20060102T150405Z")))
	} else {
		lines = append(lines, fmt.Sprintf("LAST-MODIFIED:%s", dtstamp))
	}

	if task.Completed != nil {
		lines = append(lines, fmt.Sprintf("COMPLETED:%s", task.Completed.UTC().Format("20060102T150405Z")))
	}

	lines = append(lines, "END:VTODO")
	lines = append(lines, "END:VCALENDAR")

	return strings.Join(lines, "\r\n")
}

// =============================================================================
// XML Response Parsing
// =============================================================================

// Prop represents a CalDAV property
type Prop struct {
	DisplayName  string `xml:"displayname"`
	ResourceType struct {
		Collection bool `xml:"collection"`
		Calendar   bool `xml:"calendar"`
	} `xml:"resourcetype"`
	CTag string `xml:"getctag"`
}

// PropStat represents a property status
type PropStat struct {
	Prop   Prop   `xml:"prop"`
	Status string `xml:"status"`
}

// Response represents a CalDAV response
type Response struct {
	Href     string     `xml:"href"`
	PropStat []PropStat `xml:"propstat"`
}

// MultiStatus represents a CalDAV multistatus response
type MultiStatus struct {
	Responses []Response `xml:"response"`
}

// parseCalendarList parses a PROPFIND response to extract calendars
func parseCalendarList(xmlBody, username string) ([]backend.List, error) {
	var lists []backend.List

	var ms MultiStatus
	if err := xml.Unmarshal([]byte(xmlBody), &ms); err != nil {
		// Try regex fallback for simpler parsing
		return parseCalendarListRegex(xmlBody, username)
	}

	for _, resp := range ms.Responses {
		for _, ps := range resp.PropStat {
			if strings.Contains(ps.Status, "200") && ps.Prop.DisplayName != "" {
				// Extract calendar ID from href
				calID := extractCalendarID(resp.Href, username)
				if calID != "" {
					lists = append(lists, backend.List{
						ID:       calID,
						Name:     ps.Prop.DisplayName,
						Modified: time.Now(), // CalDAV doesn't provide modified time for collections
					})
				}
			}
		}
	}

	return lists, nil
}

// parseCalendarListRegex is a fallback parser using regex
func parseCalendarListRegex(xmlBody, username string) ([]backend.List, error) {
	var lists []backend.List

	// Extract displayname and href pairs
	hrefPattern := regexp.MustCompile(`<d:href>([^<]+)</d:href>`)
	displayPattern := regexp.MustCompile(`<d:displayname>([^<]+)</d:displayname>`)

	hrefs := hrefPattern.FindAllStringSubmatch(xmlBody, -1)
	displays := displayPattern.FindAllStringSubmatch(xmlBody, -1)

	// Match them up (they appear in order in responses)
	for i, href := range hrefs {
		if i < len(displays) {
			calID := extractCalendarID(href[1], username)
			if calID != "" && displays[i][1] != "" {
				lists = append(lists, backend.List{
					ID:       calID,
					Name:     displays[i][1],
					Modified: time.Now(),
				})
			}
		}
	}

	return lists, nil
}

// extractCalendarID extracts the calendar ID from a CalDAV href
func extractCalendarID(href, username string) string {
	// href format: /remote.php/dav/calendars/username/calendarid/
	prefix := fmt.Sprintf("/remote.php/dav/calendars/%s/", username)
	if strings.HasPrefix(href, prefix) {
		id := strings.TrimPrefix(href, prefix)
		id = strings.TrimSuffix(id, "/")
		if id != "" && !strings.Contains(id, "/") {
			return id
		}
	}
	return ""
}

// parseTaskList parses a REPORT response to extract tasks
func parseTaskList(xmlBody, listID string) ([]backend.Task, error) {
	var tasks []backend.Task

	// Extract calendar-data elements which contain VTODO
	dataPattern := regexp.MustCompile(`<cal:calendar-data>([^<]*(?:<!\[CDATA\[.*?\]\]>)?[^<]*)</cal:calendar-data>`)
	matches := dataPattern.FindAllStringSubmatch(xmlBody, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			vtodo := match[1]
			// Unescape XML entities
			vtodo = strings.ReplaceAll(vtodo, "&lt;", "<")
			vtodo = strings.ReplaceAll(vtodo, "&gt;", ">")
			vtodo = strings.ReplaceAll(vtodo, "&amp;", "&")
			vtodo = strings.ReplaceAll(vtodo, "&quot;", "\"")
			vtodo = strings.ReplaceAll(vtodo, "&apos;", "'")

			if strings.Contains(vtodo, "BEGIN:VTODO") {
				task, err := parseVTODO(vtodo)
				if err == nil && task.ID != "" {
					task.ListID = listID
					tasks = append(tasks, *task)
				}
			}
		}
	}

	return tasks, nil
}

// Verify interface compliance at compile time
var _ backend.TaskManager = (*Backend)(nil)
