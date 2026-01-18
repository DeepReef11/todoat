//go:build integration

// Package todoist provides integration tests for the Todoist REST API backend.
// These tests require a valid Todoist API token and are run with `go test -tags=integration`.
package todoist

import (
	"context"
	"os"
	"testing"
	"time"
)

// getIntegrationConfig returns a Config for integration testing.
// It reads the API token from environment variables and skips the test if not configured.
func getIntegrationConfig(t *testing.T) Config {
	t.Helper()

	token := os.Getenv("TODOAT_TODOIST_TOKEN")

	if token == "" {
		t.Skip("Skipping integration test: TODOAT_TODOIST_TOKEN must be set")
	}

	return Config{
		APIToken: token,
	}
}

// TestIntegrationTodoistConnection connects to the real Todoist API and lists projects.
func TestIntegrationTodoistConnection(t *testing.T) {
	cfg := getIntegrationConfig(t)

	be, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create Todoist backend: %v", err)
	}
	defer func() { _ = be.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	lists, err := be.GetLists(ctx)
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}

	t.Logf("Found %d projects:", len(lists))
	for _, list := range lists {
		t.Logf("  - %s (ID: %s)", list.Name, list.ID)
	}

	// Every Todoist account has at least an Inbox project
	if len(lists) == 0 {
		t.Error("Expected at least one project (Inbox)")
	}

	// Verify we can find a project (likely Inbox)
	hasInbox := false
	for _, list := range lists {
		if list.Name == "Inbox" {
			hasInbox = true
			break
		}
	}

	if hasInbox {
		t.Log("Found Inbox project as expected")
	} else {
		t.Log("Inbox project not found (may have been renamed)")
	}
}
