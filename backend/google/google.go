// Package google provides a backend implementation for the Google Tasks API v1.
package google

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
	// DefaultBaseURL is the Google Tasks API v1 base URL
	DefaultBaseURL = "https://www.googleapis.com"
	// DefaultTokenURL is the Google OAuth2 token endpoint
	DefaultTokenURL = "https://oauth2.googleapis.com/token"
)

// Config holds Google Tasks connection settings
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
		AccessToken:  os.Getenv("TODOAT_GOOGLE_ACCESS_TOKEN"),
		RefreshToken: os.Getenv("TODOAT_GOOGLE_REFRESH_TOKEN"),
		ClientID:     os.Getenv("TODOAT_GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("TODOAT_GOOGLE_CLIENT_SECRET"),
	}
}

// Backend implements backend.TaskManager using Google Tasks API v1
type Backend struct {
	config       Config
	client       *http.Client
	baseURL      string
	tokenURL     string
	accessToken  string
	refreshToken string
}

// New creates a new Google Tasks backend
func New(cfg Config) (*Backend, error) {
	if cfg.AccessToken == "" {
		return nil, fmt.Errorf("google access token is required")
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

// doRequest performs an authenticated Google Tasks API request
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
// List (TaskList) Operations
// =============================================================================

// GetLists returns all Google task lists
func (b *Backend) GetLists(ctx context.Context) ([]backend.List, error) {
	resp, err := b.doRequest(ctx, http.MethodGet, "/tasks/v1/users/@me/lists", nil)
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
		Items []struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			Updated string `json:"updated"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	lists := make([]backend.List, len(result.Items))
	for i, item := range result.Items {
		modified, _ := time.Parse(time.RFC3339, item.Updated)
		lists[i] = backend.List{
			ID:       item.ID,
			Name:     item.Title,
			Modified: modified,
		}
	}

	return lists, nil
}

// GetList returns a specific task list by ID
func (b *Backend) GetList(ctx context.Context, listID string) (*backend.List, error) {
	resp, err := b.doRequest(ctx, http.MethodGet, "/tasks/v1/users/@me/lists/"+listID, nil)
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

	var item struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Updated string `json:"updated"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}

	modified, _ := time.Parse(time.RFC3339, item.Updated)
	return &backend.List{
		ID:       item.ID,
		Name:     item.Title,
		Modified: modified,
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

// CreateList creates a new Google task list
func (b *Backend) CreateList(ctx context.Context, name string) (*backend.List, error) {
	body := map[string]string{"title": name}

	resp, err := b.doRequest(ctx, http.MethodPost, "/tasks/v1/users/@me/lists", body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create task list: status %d", resp.StatusCode)
	}

	var item struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Updated string `json:"updated"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}

	modified, _ := time.Parse(time.RFC3339, item.Updated)
	return &backend.List{
		ID:       item.ID,
		Name:     item.Title,
		Modified: modified,
	}, nil
}

// DeleteList deletes a Google task list (permanent deletion)
func (b *Backend) DeleteList(ctx context.Context, listID string) error {
	resp, err := b.doRequest(ctx, http.MethodDelete, "/tasks/v1/users/@me/lists/"+listID, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete task list: status %d", resp.StatusCode)
	}

	return nil
}

// GetDeletedLists returns deleted lists (not supported by Google Tasks)
func (b *Backend) GetDeletedLists(ctx context.Context) ([]backend.List, error) {
	// Google Tasks doesn't have a trash concept for task lists
	return []backend.List{}, nil
}

// GetDeletedListByName returns a deleted list by name (not supported)
func (b *Backend) GetDeletedListByName(ctx context.Context, name string) (*backend.List, error) {
	return nil, nil
}

// RestoreList restores a deleted list (not supported)
func (b *Backend) RestoreList(ctx context.Context, listID string) error {
	return fmt.Errorf("restoring task lists is not supported by Google Tasks")
}

// PurgeList permanently deletes a list (not supported - already permanent)
func (b *Backend) PurgeList(ctx context.Context, listID string) error {
	return fmt.Errorf("purging task lists is not supported by Google Tasks (deletion is already permanent)")
}

// =============================================================================
// Task Operations
// =============================================================================

// GetTasks returns all tasks in a task list
func (b *Backend) GetTasks(ctx context.Context, listID string) ([]backend.Task, error) {
	resp, err := b.doRequest(ctx, http.MethodGet, "/tasks/v1/lists/"+listID+"/tasks", nil)
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
		Items []struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			Notes     string `json:"notes"`
			Status    string `json:"status"`
			Due       string `json:"due"`
			Parent    string `json:"parent"`
			Updated   string `json:"updated"`
			Completed string `json:"completed"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	tasks := make([]backend.Task, len(result.Items))
	for i, item := range result.Items {
		modified, _ := time.Parse(time.RFC3339, item.Updated)

		tasks[i] = backend.Task{
			ID:          item.ID,
			Summary:     item.Title,
			Description: item.Notes,
			Status:      googleToBackendStatus(item.Status),
			ListID:      listID,
			ParentID:    item.Parent,
			Modified:    modified,
		}

		if item.Due != "" {
			dueDate, err := time.Parse(time.RFC3339, item.Due)
			if err == nil {
				tasks[i].DueDate = &dueDate
			}
		}

		if item.Completed != "" {
			completed, err := time.Parse(time.RFC3339, item.Completed)
			if err == nil {
				tasks[i].Completed = &completed
			}
		}
	}

	return tasks, nil
}

// GetTask returns a specific task by ID
func (b *Backend) GetTask(ctx context.Context, listID, taskID string) (*backend.Task, error) {
	resp, err := b.doRequest(ctx, http.MethodGet, "/tasks/v1/lists/"+listID+"/tasks/"+taskID, nil)
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

	var item struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Notes     string `json:"notes"`
		Status    string `json:"status"`
		Due       string `json:"due"`
		Parent    string `json:"parent"`
		Updated   string `json:"updated"`
		Completed string `json:"completed"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}

	modified, _ := time.Parse(time.RFC3339, item.Updated)

	task := &backend.Task{
		ID:          item.ID,
		Summary:     item.Title,
		Description: item.Notes,
		Status:      googleToBackendStatus(item.Status),
		ListID:      listID,
		ParentID:    item.Parent,
		Modified:    modified,
	}

	if item.Due != "" {
		dueDate, err := time.Parse(time.RFC3339, item.Due)
		if err == nil {
			task.DueDate = &dueDate
		}
	}

	if item.Completed != "" {
		completed, err := time.Parse(time.RFC3339, item.Completed)
		if err == nil {
			task.Completed = &completed
		}
	}

	return task, nil
}

// CreateTask creates a new task in a task list
func (b *Backend) CreateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	body := map[string]interface{}{
		"title":  task.Summary,
		"status": backendToGoogleStatus(task.Status),
	}

	if task.Description != "" {
		body["notes"] = task.Description
	}

	if task.ParentID != "" {
		body["parent"] = task.ParentID
	}

	if task.DueDate != nil {
		// Google Tasks uses RFC3339 format for due dates
		body["due"] = task.DueDate.Format(time.RFC3339)
	}

	resp, err := b.doRequest(ctx, http.MethodPost, "/tasks/v1/lists/"+listID+"/tasks", body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create task: status %d", resp.StatusCode)
	}

	var item struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Notes     string `json:"notes"`
		Status    string `json:"status"`
		Due       string `json:"due"`
		Parent    string `json:"parent"`
		Updated   string `json:"updated"`
		Completed string `json:"completed"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}

	modified, _ := time.Parse(time.RFC3339, item.Updated)

	created := &backend.Task{
		ID:          item.ID,
		Summary:     item.Title,
		Description: item.Notes,
		Status:      googleToBackendStatus(item.Status),
		ListID:      listID,
		ParentID:    item.Parent,
		Modified:    modified,
		Created:     modified,
	}

	if item.Due != "" {
		dueDate, err := time.Parse(time.RFC3339, item.Due)
		if err == nil {
			created.DueDate = &dueDate
		}
	}

	return created, nil
}

// UpdateTask updates an existing task
func (b *Backend) UpdateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	body := map[string]interface{}{
		"title":  task.Summary,
		"status": backendToGoogleStatus(task.Status),
	}

	if task.Description != "" {
		body["notes"] = task.Description
	}

	if task.DueDate != nil {
		body["due"] = task.DueDate.Format(time.RFC3339)
	}

	resp, err := b.doRequest(ctx, http.MethodPatch, "/tasks/v1/lists/"+listID+"/tasks/"+task.ID, body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update task: status %d", resp.StatusCode)
	}

	var item struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Notes     string `json:"notes"`
		Status    string `json:"status"`
		Due       string `json:"due"`
		Parent    string `json:"parent"`
		Updated   string `json:"updated"`
		Completed string `json:"completed"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}

	modified, _ := time.Parse(time.RFC3339, item.Updated)

	updated := &backend.Task{
		ID:          item.ID,
		Summary:     item.Title,
		Description: item.Notes,
		Status:      googleToBackendStatus(item.Status),
		ListID:      listID,
		ParentID:    item.Parent,
		Modified:    modified,
	}

	if item.Due != "" {
		dueDate, err := time.Parse(time.RFC3339, item.Due)
		if err == nil {
			updated.DueDate = &dueDate
		}
	}

	if item.Completed != "" {
		completed, err := time.Parse(time.RFC3339, item.Completed)
		if err == nil {
			updated.Completed = &completed
		}
	}

	return updated, nil
}

// DeleteTask removes a task
func (b *Backend) DeleteTask(ctx context.Context, listID, taskID string) error {
	resp, err := b.doRequest(ctx, http.MethodDelete, "/tasks/v1/lists/"+listID+"/tasks/"+taskID, nil)
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
// Status Conversion Functions
// =============================================================================

// googleToBackendStatus converts Google Tasks status to backend status
func googleToBackendStatus(status string) backend.TaskStatus {
	switch status {
	case "completed":
		return backend.StatusCompleted
	default:
		return backend.StatusNeedsAction
	}
}

// backendToGoogleStatus converts backend status to Google Tasks status
func backendToGoogleStatus(status backend.TaskStatus) string {
	switch status {
	case backend.StatusCompleted, backend.StatusCancelled:
		return "completed"
	default:
		return "needsAction"
	}
}

// Verify interface compliance at compile time
var _ backend.TaskManager = (*Backend)(nil)
