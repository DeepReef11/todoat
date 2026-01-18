// Package mstodo provides a backend implementation for the Microsoft Graph API To Do.
package mstodo

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
	// DefaultBaseURL is the Microsoft Graph API base URL
	DefaultBaseURL = "https://graph.microsoft.com"
	// DefaultTokenURL is the Microsoft identity platform token endpoint
	DefaultTokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
)

// Config holds Microsoft To Do connection settings
type Config struct {
	AccessToken  string
	RefreshToken string
	ClientID     string
	ClientSecret string
	BaseURL      string // Override for testing
	TokenURL     string // Override for testing
}

// ConfigFromEnv creates a Config from environment variables
func ConfigFromEnv() Config {
	return Config{
		AccessToken:  os.Getenv("TODOAT_MSTODO_ACCESS_TOKEN"),
		RefreshToken: os.Getenv("TODOAT_MSTODO_REFRESH_TOKEN"),
		ClientID:     os.Getenv("TODOAT_MSTODO_CLIENT_ID"),
		ClientSecret: os.Getenv("TODOAT_MSTODO_CLIENT_SECRET"),
	}
}

// Backend implements backend.TaskManager using Microsoft Graph API To Do
type Backend struct {
	config       Config
	client       *http.Client
	baseURL      string
	tokenURL     string
	accessToken  string
	refreshToken string
}

// New creates a new Microsoft To Do backend
func New(cfg Config) (*Backend, error) {
	if cfg.AccessToken == "" {
		return nil, fmt.Errorf("microsoft access token is required")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	tokenURL := cfg.TokenURL
	if tokenURL == "" {
		tokenURL = DefaultTokenURL
	}

	return &Backend{
		config:       cfg,
		client:       createHTTPClient(),
		baseURL:      baseURL,
		tokenURL:     tokenURL,
		accessToken:  cfg.AccessToken,
		refreshToken: cfg.RefreshToken,
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
	if transport, ok := b.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// refreshAccessToken refreshes the OAuth2 access token using the refresh token
func (b *Backend) refreshAccessToken(ctx context.Context) error {
	if b.refreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	data := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": b.refreshToken,
	}

	if b.config.ClientID != "" {
		data["client_id"] = b.config.ClientID
	}
	if b.config.ClientSecret != "" {
		data["client_secret"] = b.config.ClientSecret
	}

	jsonBody, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.tokenURL, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed: status %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	b.accessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		b.refreshToken = tokenResp.RefreshToken
	}

	return nil
}

// doRequest performs an authenticated Microsoft Graph API request
func (b *Backend) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := b.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+b.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}

	// Handle token expiration - attempt refresh and retry
	if resp.StatusCode == http.StatusUnauthorized && b.refreshToken != "" {
		_ = resp.Body.Close()

		if err := b.refreshAccessToken(ctx); err != nil {
			return nil, fmt.Errorf("token refresh failed: %w", err)
		}

		// Retry with new token
		if body != nil {
			jsonBody, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(jsonBody)
		}

		req, err = http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+b.accessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err = b.client.Do(req)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

// =============================================================================
// Microsoft Graph API Types
// =============================================================================

type msTaskList struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	IsOwner           bool   `json:"isOwner"`
	IsShared          bool   `json:"isShared"`
	WellknownListName string `json:"wellknownListName,omitempty"`
}

type msTask struct {
	ID                   string            `json:"id"`
	Title                string            `json:"title"`
	Body                 *msTaskBody       `json:"body,omitempty"`
	Status               string            `json:"status"`     // notStarted, inProgress, completed
	Importance           string            `json:"importance"` // low, normal, high
	DueDateTime          *msDateTime       `json:"dueDateTime,omitempty"`
	CompletedDateTime    *msDateTime       `json:"completedDateTime,omitempty"`
	CreatedDateTime      string            `json:"createdDateTime"`
	LastModifiedDateTime string            `json:"lastModifiedDateTime"`
	ChecklistItems       []msChecklistItem `json:"checklistItems,omitempty"`
}

type msTaskBody struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"` // text or html
}

type msDateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type msChecklistItem struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	IsChecked   bool   `json:"isChecked"`
}

// =============================================================================
// List (TaskList) Operations
// =============================================================================

// GetLists returns all Microsoft To Do task lists
func (b *Backend) GetLists(ctx context.Context) ([]backend.List, error) {
	resp, err := b.doRequest(ctx, http.MethodGet, "/v1.0/me/todo/lists", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed: invalid access token")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get task lists: status %d", resp.StatusCode)
	}

	var result struct {
		Value []msTaskList `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	lists := make([]backend.List, len(result.Value))
	for i, item := range result.Value {
		lists[i] = backend.List{
			ID:   item.ID,
			Name: item.DisplayName,
		}
	}

	return lists, nil
}

// GetList returns a specific task list by ID
func (b *Backend) GetList(ctx context.Context, listID string) (*backend.List, error) {
	resp, err := b.doRequest(ctx, http.MethodGet, "/v1.0/me/todo/lists/"+listID, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get task list: status %d", resp.StatusCode)
	}

	var item msTaskList
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}

	return &backend.List{
		ID:   item.ID,
		Name: item.DisplayName,
	}, nil
}

// GetListByName returns a specific task list by name (case-insensitive)
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

// CreateList creates a new Microsoft To Do task list
func (b *Backend) CreateList(ctx context.Context, name string) (*backend.List, error) {
	body := map[string]string{"displayName": name}

	resp, err := b.doRequest(ctx, http.MethodPost, "/v1.0/me/todo/lists", body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create task list: status %d", resp.StatusCode)
	}

	var item msTaskList
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}

	return &backend.List{
		ID:   item.ID,
		Name: item.DisplayName,
	}, nil
}

// DeleteList deletes a Microsoft To Do task list (permanent deletion)
func (b *Backend) DeleteList(ctx context.Context, listID string) error {
	resp, err := b.doRequest(ctx, http.MethodDelete, "/v1.0/me/todo/lists/"+listID, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete task list: status %d", resp.StatusCode)
	}

	return nil
}

// GetDeletedLists returns deleted lists (not supported by Microsoft To Do)
func (b *Backend) GetDeletedLists(ctx context.Context) ([]backend.List, error) {
	// Microsoft To Do doesn't have a trash concept for task lists
	return []backend.List{}, nil
}

// GetDeletedListByName returns a deleted list by name (not supported)
func (b *Backend) GetDeletedListByName(ctx context.Context, name string) (*backend.List, error) {
	return nil, nil
}

// RestoreList restores a deleted list (not supported)
func (b *Backend) RestoreList(ctx context.Context, listID string) error {
	return fmt.Errorf("restoring task lists is not supported by Microsoft To Do")
}

// PurgeList permanently deletes a list (not supported - already permanent)
func (b *Backend) PurgeList(ctx context.Context, listID string) error {
	return fmt.Errorf("purging task lists is not supported by Microsoft To Do (deletion is already permanent)")
}

// =============================================================================
// Task Operations
// =============================================================================

// GetTasks returns all tasks in a task list
func (b *Backend) GetTasks(ctx context.Context, listID string) ([]backend.Task, error) {
	resp, err := b.doRequest(ctx, http.MethodGet, "/v1.0/me/todo/lists/"+listID+"/tasks", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("task list not found: %s", listID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tasks: status %d", resp.StatusCode)
	}

	var result struct {
		Value []msTask `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	tasks := make([]backend.Task, len(result.Value))
	for i, item := range result.Value {
		modified, _ := time.Parse(time.RFC3339, item.LastModifiedDateTime)
		created, _ := time.Parse(time.RFC3339, item.CreatedDateTime)

		tasks[i] = backend.Task{
			ID:       item.ID,
			Summary:  item.Title,
			Status:   msToBackendStatus(item.Status),
			Priority: importanceToPriority(item.Importance),
			ListID:   listID,
			Modified: modified,
			Created:  created,
		}

		if item.Body != nil {
			tasks[i].Description = item.Body.Content
		}

		if item.DueDateTime != nil && item.DueDateTime.DateTime != "" {
			dueDate := parseMSDateTime(item.DueDateTime)
			if dueDate != nil {
				tasks[i].DueDate = dueDate
			}
		}

		if item.CompletedDateTime != nil && item.CompletedDateTime.DateTime != "" {
			completed := parseMSDateTime(item.CompletedDateTime)
			if completed != nil {
				tasks[i].Completed = completed
			}
		}
	}

	return tasks, nil
}

// GetTask returns a specific task by ID
func (b *Backend) GetTask(ctx context.Context, listID, taskID string) (*backend.Task, error) {
	resp, err := b.doRequest(ctx, http.MethodGet, "/v1.0/me/todo/lists/"+listID+"/tasks/"+taskID, nil)
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

	var item msTask
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}

	modified, _ := time.Parse(time.RFC3339, item.LastModifiedDateTime)
	created, _ := time.Parse(time.RFC3339, item.CreatedDateTime)

	task := &backend.Task{
		ID:       item.ID,
		Summary:  item.Title,
		Status:   msToBackendStatus(item.Status),
		Priority: importanceToPriority(item.Importance),
		ListID:   listID,
		Modified: modified,
		Created:  created,
	}

	if item.Body != nil {
		task.Description = item.Body.Content
	}

	if item.DueDateTime != nil && item.DueDateTime.DateTime != "" {
		task.DueDate = parseMSDateTime(item.DueDateTime)
	}

	if item.CompletedDateTime != nil && item.CompletedDateTime.DateTime != "" {
		task.Completed = parseMSDateTime(item.CompletedDateTime)
	}

	return task, nil
}

// CreateTask creates a new task in a task list
func (b *Backend) CreateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	body := map[string]interface{}{
		"title":      task.Summary,
		"status":     backendToMSStatus(task.Status),
		"importance": priorityToImportance(task.Priority),
	}

	if task.Description != "" {
		body["body"] = map[string]string{
			"content":     task.Description,
			"contentType": "text",
		}
	}

	if task.DueDate != nil {
		body["dueDateTime"] = map[string]string{
			"dateTime": task.DueDate.Format("2006-01-02T15:04:05.0000000"),
			"timeZone": "UTC",
		}
	}

	resp, err := b.doRequest(ctx, http.MethodPost, "/v1.0/me/todo/lists/"+listID+"/tasks", body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create task: status %d", resp.StatusCode)
	}

	var item msTask
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}

	modified, _ := time.Parse(time.RFC3339, item.LastModifiedDateTime)
	created, _ := time.Parse(time.RFC3339, item.CreatedDateTime)

	createdTask := &backend.Task{
		ID:       item.ID,
		Summary:  item.Title,
		Status:   msToBackendStatus(item.Status),
		Priority: importanceToPriority(item.Importance),
		ListID:   listID,
		Modified: modified,
		Created:  created,
	}

	if item.Body != nil {
		createdTask.Description = item.Body.Content
	}

	if item.DueDateTime != nil && item.DueDateTime.DateTime != "" {
		createdTask.DueDate = parseMSDateTime(item.DueDateTime)
	}

	return createdTask, nil
}

// UpdateTask updates an existing task
func (b *Backend) UpdateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	body := map[string]interface{}{
		"title":      task.Summary,
		"status":     backendToMSStatus(task.Status),
		"importance": priorityToImportance(task.Priority),
	}

	if task.Description != "" {
		body["body"] = map[string]string{
			"content":     task.Description,
			"contentType": "text",
		}
	}

	if task.DueDate != nil {
		body["dueDateTime"] = map[string]string{
			"dateTime": task.DueDate.Format("2006-01-02T15:04:05.0000000"),
			"timeZone": "UTC",
		}
	}

	resp, err := b.doRequest(ctx, http.MethodPatch, "/v1.0/me/todo/lists/"+listID+"/tasks/"+task.ID, body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update task: status %d", resp.StatusCode)
	}

	var item msTask
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}

	modified, _ := time.Parse(time.RFC3339, item.LastModifiedDateTime)

	updated := &backend.Task{
		ID:       item.ID,
		Summary:  item.Title,
		Status:   msToBackendStatus(item.Status),
		Priority: importanceToPriority(item.Importance),
		ListID:   listID,
		Modified: modified,
	}

	if item.Body != nil {
		updated.Description = item.Body.Content
	}

	if item.DueDateTime != nil && item.DueDateTime.DateTime != "" {
		updated.DueDate = parseMSDateTime(item.DueDateTime)
	}

	if item.CompletedDateTime != nil && item.CompletedDateTime.DateTime != "" {
		updated.Completed = parseMSDateTime(item.CompletedDateTime)
	}

	return updated, nil
}

// DeleteTask removes a task
func (b *Backend) DeleteTask(ctx context.Context, listID, taskID string) error {
	resp, err := b.doRequest(ctx, http.MethodDelete, "/v1.0/me/todo/lists/"+listID+"/tasks/"+taskID, nil)
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
// Status and Priority Conversion Functions
// =============================================================================

// msToBackendStatus converts Microsoft To Do status to backend status
func msToBackendStatus(status string) backend.TaskStatus {
	switch status {
	case "completed":
		return backend.StatusCompleted
	case "inProgress":
		return backend.StatusInProgress
	default: // notStarted
		return backend.StatusNeedsAction
	}
}

// backendToMSStatus converts backend status to Microsoft To Do status
func backendToMSStatus(status backend.TaskStatus) string {
	switch status {
	case backend.StatusCompleted, backend.StatusCancelled:
		return "completed"
	case backend.StatusInProgress:
		return "inProgress"
	default:
		return "notStarted"
	}
}

// importanceToPriority converts Microsoft importance (low/normal/high) to priority (1-9)
func importanceToPriority(importance string) int {
	switch importance {
	case "high":
		return 1
	case "low":
		return 9
	default: // normal
		return 5
	}
}

// priorityToImportance converts priority (1-9) to Microsoft importance (low/normal/high)
func priorityToImportance(priority int) string {
	switch {
	case priority >= 1 && priority <= 3:
		return "high"
	case priority >= 7 && priority <= 9:
		return "low"
	default:
		return "normal"
	}
}

// parseMSDateTime parses Microsoft dateTime format to Go time.Time
func parseMSDateTime(dt *msDateTime) *time.Time {
	if dt == nil || dt.DateTime == "" {
		return nil
	}

	// Try various formats
	formats := []string{
		"2006-01-02T15:04:05.0000000",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
		"2006-01-02T15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dt.DateTime); err == nil {
			return &t
		}
	}

	return nil
}

// Verify interface compliance at compile time
var _ backend.TaskManager = (*Backend)(nil)
