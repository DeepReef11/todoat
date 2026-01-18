//go:build integration

// Package nextcloud provides integration tests for the Nextcloud CalDAV backend.
// These tests require a real Nextcloud instance and are run with `go test -tags=integration`.
package nextcloud

import (
	"context"
	"os"
	"testing"
	"time"

	"todoat/backend"
)

// getIntegrationConfig returns a Config for integration testing.
// It reads credentials from environment variables and skips the test if not configured.
func getIntegrationConfig(t *testing.T) Config {
	t.Helper()

	host := os.Getenv("TODOAT_NEXTCLOUD_HOST")
	username := os.Getenv("TODOAT_NEXTCLOUD_USERNAME")
	password := os.Getenv("TODOAT_NEXTCLOUD_PASSWORD")

	if host == "" || username == "" || password == "" {
		t.Skip("Skipping integration test: TODOAT_NEXTCLOUD_HOST, TODOAT_NEXTCLOUD_USERNAME, and TODOAT_NEXTCLOUD_PASSWORD must be set")
	}

	return Config{
		Host:               host,
		Username:           username,
		Password:           password,
		AllowHTTP:          true, // Docker test instance uses HTTP
		InsecureSkipVerify: true, // Allow self-signed certs
	}
}

// TestIntegrationNextcloudConnection connects to a real Nextcloud instance and lists calendars.
func TestIntegrationNextcloudConnection(t *testing.T) {
	cfg := getIntegrationConfig(t)

	be, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create Nextcloud backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("Failed to list calendars: %v", err)
	}

	t.Logf("Found %d calendars:", len(lists))
	for _, list := range lists {
		t.Logf("  - %s (ID: %s)", list.Name, list.ID)
	}

	// Nextcloud creates a default "Personal" calendar, so we should have at least one
	if len(lists) == 0 {
		t.Log("No calendars found (this may be expected for a fresh Nextcloud instance)")
	}
}

// TestIntegrationNextcloudCRUD creates, reads, updates, and deletes a task on real Nextcloud.
func TestIntegrationNextcloudCRUD(t *testing.T) {
	cfg := getIntegrationConfig(t)

	be, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create Nextcloud backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get available calendars
	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("Failed to list calendars: %v", err)
	}

	if len(lists) == 0 {
		t.Skip("No calendars available for CRUD test")
	}

	// Use the first calendar for testing
	calendarID := lists[0].ID
	t.Logf("Using calendar: %s (ID: %s)", lists[0].Name, calendarID)

	// CREATE a task
	testTask := &backend.Task{
		Summary:     "Integration Test Task",
		Description: "This task was created by an integration test",
		Status:      backend.StatusNeedsAction,
		Priority:    5,
	}

	created, err := be.CreateTask(ctx, calendarID, testTask)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	t.Logf("Created task with ID: %s", created.ID)

	if created.Summary != testTask.Summary {
		t.Errorf("Expected summary %q, got %q", testTask.Summary, created.Summary)
	}

	// READ the task back
	tasks, err := be.GetTasks(ctx, calendarID)
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}

	var foundTask *backend.Task
	for _, task := range tasks {
		if task.ID == created.ID {
			foundTask = &task
			break
		}
	}

	if foundTask == nil {
		t.Fatalf("Created task not found in task list")
	}

	t.Logf("Read back task: %s", foundTask.Summary)

	// UPDATE the task
	foundTask.Summary = "Updated Integration Test Task"
	foundTask.Status = backend.StatusCompleted

	updated, err := be.UpdateTask(ctx, calendarID, foundTask)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	if updated.Summary != "Updated Integration Test Task" {
		t.Errorf("Expected updated summary, got %q", updated.Summary)
	}

	if updated.Status != backend.StatusCompleted {
		t.Errorf("Expected COMPLETED status, got %s", updated.Status)
	}

	t.Logf("Updated task successfully")

	// DELETE the task
	err = be.DeleteTask(ctx, calendarID, created.ID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	t.Logf("Deleted task successfully")

	// Verify deletion
	tasks, err = be.GetTasks(ctx, calendarID)
	if err != nil {
		t.Fatalf("Failed to get tasks after deletion: %v", err)
	}

	for _, task := range tasks {
		if task.ID == created.ID {
			t.Errorf("Task should have been deleted but still exists")
		}
	}

	t.Log("CRUD test completed successfully")
}
