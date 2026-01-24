package sqlite

import (
	"context"
	"testing"
	"time"

	"todoat/backend"
)

// helper to create a list and fail if nil
func mustCreateList(t *testing.T, b *Backend, ctx context.Context, name string) *backend.List {
	t.Helper()
	list, err := b.CreateList(ctx, name)
	if err != nil {
		t.Fatalf("CreateList error: %v", err)
	}
	if list == nil {
		t.Fatal("CreateList returned nil list")
	}
	return list
}

// helper to create a task and fail if nil
func mustCreateTask(t *testing.T, b *Backend, ctx context.Context, listID string, task *backend.Task) *backend.Task {
	t.Helper()
	created, err := b.CreateTask(ctx, listID, task)
	if err != nil {
		t.Fatalf("CreateTask error: %v", err)
	}
	if created == nil {
		t.Fatal("CreateTask returned nil task")
	}
	return created
}

// mustNewBackend creates an in-memory backend and registers cleanup
func mustNewBackend(t *testing.T) (*Backend, context.Context) {
	t.Helper()
	b, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) error: %v", err)
	}
	t.Cleanup(func() { _ = b.Close() })
	return b, context.Background()
}

// TestNewBackend verifies that New creates a backend with the given path.
func TestNewBackend(t *testing.T) {
	b, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) error: %v", err)
	}
	defer func() { _ = b.Close() }()

	if b == nil {
		t.Fatal("New(:memory:) returned nil backend")
	}
}

// TestBackendImplementsInterface verifies the Backend type implements TaskManager.
func TestBackendImplementsInterface(t *testing.T) {
	var _ backend.TaskManager = (*Backend)(nil)
}

// TestCreateAndGetList tests creating and retrieving a task list.
func TestCreateAndGetList(t *testing.T) {
	b, ctx := mustNewBackend(t)

	// Create a list
	list := mustCreateList(t, b, ctx, "Work Tasks")
	if list.Name != "Work Tasks" {
		t.Errorf("list.Name = %q, want %q", list.Name, "Work Tasks")
	}
	if list.ID == "" {
		t.Error("list.ID is empty")
	}

	// Get the list by ID
	retrieved, err := b.GetList(ctx, list.ID)
	if err != nil {
		t.Fatalf("GetList error: %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetList returned nil")
	}
	if retrieved.Name != "Work Tasks" {
		t.Errorf("retrieved.Name = %q, want %q", retrieved.Name, "Work Tasks")
	}
}

// TestGetLists tests retrieving all lists.
func TestGetLists(t *testing.T) {
	b, ctx := mustNewBackend(t)

	// Initially empty
	lists, err := b.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists error: %v", err)
	}
	if len(lists) != 0 {
		t.Errorf("initial GetLists returned %d lists, want 0", len(lists))
	}

	// Create some lists
	mustCreateList(t, b, ctx, "List One")
	mustCreateList(t, b, ctx, "List Two")

	// Now should have 2 lists
	lists, err = b.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists error: %v", err)
	}
	if len(lists) != 2 {
		t.Errorf("GetLists returned %d lists, want 2", len(lists))
	}
}

// TestDeleteList tests deleting a list.
func TestDeleteList(t *testing.T) {
	b, ctx := mustNewBackend(t)

	// Create and then delete a list
	list := mustCreateList(t, b, ctx, "Temporary")

	err := b.DeleteList(ctx, list.ID)
	if err != nil {
		t.Fatalf("DeleteList error: %v", err)
	}

	// Should no longer exist
	retrieved, err := b.GetList(ctx, list.ID)
	if err == nil && retrieved != nil {
		t.Error("GetList should return nil or error for deleted list")
	}
}

// TestCreateAndGetTask tests creating and retrieving a task.
func TestCreateAndGetTask(t *testing.T) {
	b, ctx := mustNewBackend(t)

	// First create a list
	list := mustCreateList(t, b, ctx, "My Tasks")

	// Create a task
	task := &backend.Task{
		Summary:  "Buy groceries",
		Status:   backend.StatusNeedsAction,
		Priority: 5,
	}

	created := mustCreateTask(t, b, ctx, list.ID, task)
	if created.ID == "" {
		t.Error("created.ID is empty (should be auto-generated)")
	}
	if created.Summary != "Buy groceries" {
		t.Errorf("created.Summary = %q, want %q", created.Summary, "Buy groceries")
	}
	if created.Status != backend.StatusNeedsAction {
		t.Errorf("created.Status = %v, want %v", created.Status, backend.StatusNeedsAction)
	}
	if created.Priority != 5 {
		t.Errorf("created.Priority = %d, want %d", created.Priority, 5)
	}
	if created.Created.IsZero() {
		t.Error("created.Created is zero (should be auto-set)")
	}
	if created.Modified.IsZero() {
		t.Error("created.Modified is zero (should be auto-set)")
	}

	// Retrieve by ID
	retrieved, err := b.GetTask(ctx, list.ID, created.ID)
	if err != nil {
		t.Fatalf("GetTask error: %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetTask returned nil")
	}
	if retrieved.Summary != "Buy groceries" {
		t.Errorf("retrieved.Summary = %q, want %q", retrieved.Summary, "Buy groceries")
	}
}

// TestGetTasks tests retrieving all tasks in a list.
func TestGetTasks(t *testing.T) {
	b, ctx := mustNewBackend(t)

	list := mustCreateList(t, b, ctx, "Project")

	// Initially empty
	tasks, err := b.GetTasks(ctx, list.ID)
	if err != nil {
		t.Fatalf("GetTasks error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("initial GetTasks returned %d tasks, want 0", len(tasks))
	}

	// Create multiple tasks
	mustCreateTask(t, b, ctx, list.ID, &backend.Task{Summary: "Task 1"})
	mustCreateTask(t, b, ctx, list.ID, &backend.Task{Summary: "Task 2"})
	mustCreateTask(t, b, ctx, list.ID, &backend.Task{Summary: "Task 3"})

	tasks, err = b.GetTasks(ctx, list.ID)
	if err != nil {
		t.Fatalf("GetTasks error: %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("GetTasks returned %d tasks, want 3", len(tasks))
	}
}

// TestUpdateTask tests updating a task.
func TestUpdateTask(t *testing.T) {
	b, ctx := mustNewBackend(t)

	list := mustCreateList(t, b, ctx, "Work")

	created := mustCreateTask(t, b, ctx, list.ID, &backend.Task{
		Summary:  "Original title",
		Status:   backend.StatusNeedsAction,
		Priority: 3,
	})

	// Update the task
	created.Summary = "Updated title"
	created.Status = backend.StatusCompleted
	created.Priority = 1

	updated, err := b.UpdateTask(ctx, list.ID, created)
	if err != nil {
		t.Fatalf("UpdateTask error: %v", err)
	}
	if updated == nil {
		t.Fatal("UpdateTask returned nil")
	}
	if updated.Summary != "Updated title" {
		t.Errorf("updated.Summary = %q, want %q", updated.Summary, "Updated title")
	}
	if updated.Status != backend.StatusCompleted {
		t.Errorf("updated.Status = %v, want %v", updated.Status, backend.StatusCompleted)
	}
	if updated.Priority != 1 {
		t.Errorf("updated.Priority = %d, want %d", updated.Priority, 1)
	}

	// Verify the update persisted
	retrieved, err := b.GetTask(ctx, list.ID, created.ID)
	if err != nil {
		t.Fatalf("GetTask error: %v", err)
	}
	if retrieved.Summary != "Updated title" {
		t.Errorf("after update, retrieved.Summary = %q, want %q", retrieved.Summary, "Updated title")
	}
}

// TestDeleteTask tests deleting a task.
func TestDeleteTask(t *testing.T) {
	b, ctx := mustNewBackend(t)

	list := mustCreateList(t, b, ctx, "Cleanup")

	task := mustCreateTask(t, b, ctx, list.ID, &backend.Task{Summary: "Delete me"})

	err := b.DeleteTask(ctx, list.ID, task.ID)
	if err != nil {
		t.Fatalf("DeleteTask error: %v", err)
	}

	// Should no longer exist
	retrieved, err := b.GetTask(ctx, list.ID, task.ID)
	if err == nil && retrieved != nil {
		t.Error("GetTask should return nil or error for deleted task")
	}

	// GetTasks should not include it
	tasks, err := b.GetTasks(ctx, list.ID)
	if err != nil {
		t.Fatalf("GetTasks error: %v", err)
	}
	for _, tk := range tasks {
		if tk.ID == task.ID {
			t.Error("deleted task still appears in GetTasks")
		}
	}
}

// TestTaskTimestamps verifies Created and Modified are properly set.
func TestTaskTimestamps(t *testing.T) {
	b, ctx := mustNewBackend(t)

	list := mustCreateList(t, b, ctx, "Timestamps")

	before := time.Now().Add(-time.Second)
	task := mustCreateTask(t, b, ctx, list.ID, &backend.Task{Summary: "Timestamped task"})
	after := time.Now().Add(time.Second)

	if task.Created.Before(before) || task.Created.After(after) {
		t.Errorf("Created timestamp %v not in expected range", task.Created)
	}
	if task.Modified.Before(before) || task.Modified.After(after) {
		t.Errorf("Modified timestamp %v not in expected range", task.Modified)
	}

	// Update should change Modified but not Created
	originalCreated := task.Created
	time.Sleep(10 * time.Millisecond) // ensure time difference

	task.Summary = "Updated summary"
	updated, err := b.UpdateTask(ctx, list.ID, task)
	if err != nil {
		t.Fatalf("UpdateTask error: %v", err)
	}
	if updated == nil {
		t.Fatal("UpdateTask returned nil")
	}

	if !updated.Created.Equal(originalCreated) {
		t.Errorf("Created changed after update: was %v, now %v", originalCreated, updated.Created)
	}
	if !updated.Modified.After(originalCreated) {
		t.Error("Modified should be after original Created time after update")
	}
}

// TestClose tests closing the database connection.
func TestClose(t *testing.T) {
	b, err := New(":memory:")
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	err = b.Close()
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}

	// After close, operations should fail or the backend should handle gracefully
	ctx := context.Background()
	_, err = b.GetLists(ctx)
	// We expect an error or graceful handling after close
	// The specific behavior depends on implementation
	_ = err // Some implementations may return an error, others may not
}

// TestTasksIsolatedByList verifies tasks are isolated to their lists.
func TestTasksIsolatedByList(t *testing.T) {
	b, ctx := mustNewBackend(t)

	list1 := mustCreateList(t, b, ctx, "List 1")
	list2 := mustCreateList(t, b, ctx, "List 2")

	// Add tasks to list1
	mustCreateTask(t, b, ctx, list1.ID, &backend.Task{Summary: "List1 Task"})

	// Add tasks to list2
	mustCreateTask(t, b, ctx, list2.ID, &backend.Task{Summary: "List2 Task A"})
	mustCreateTask(t, b, ctx, list2.ID, &backend.Task{Summary: "List2 Task B"})

	// Get tasks from each list
	tasks1, err := b.GetTasks(ctx, list1.ID)
	if err != nil {
		t.Fatalf("GetTasks error: %v", err)
	}
	if len(tasks1) != 1 {
		t.Errorf("list1 has %d tasks, want 1", len(tasks1))
	}

	tasks2, err := b.GetTasks(ctx, list2.ID)
	if err != nil {
		t.Fatalf("GetTasks error: %v", err)
	}
	if len(tasks2) != 2 {
		t.Errorf("list2 has %d tasks, want 2", len(tasks2))
	}
}

// TestDeleteListCascadesToTasks verifies that purging a list removes its tasks.
func TestDeleteListCascadesToTasks(t *testing.T) {
	b, ctx := mustNewBackend(t)

	list := mustCreateList(t, b, ctx, "Doomed List")

	task := mustCreateTask(t, b, ctx, list.ID, &backend.Task{Summary: "Doomed Task"})
	taskID := task.ID

	// Soft-delete the list first
	err := b.DeleteList(ctx, list.ID)
	if err != nil {
		t.Fatalf("DeleteList error: %v", err)
	}

	// Purge the list (permanent delete)
	err = b.PurgeList(ctx, list.ID)
	if err != nil {
		t.Fatalf("PurgeList error: %v", err)
	}

	// The task should be gone too (cascade delete)
	retrieved, err := b.GetTask(ctx, list.ID, taskID)
	if err == nil && retrieved != nil {
		t.Error("task should be deleted when list is purged (cascade)")
	}
}

// TestDeleteListSoftDelete verifies that DeleteList is a soft-delete.
func TestDeleteListSoftDelete(t *testing.T) {
	b, ctx := mustNewBackend(t)

	list := mustCreateList(t, b, ctx, "ToTrash")
	_ = mustCreateTask(t, b, ctx, list.ID, &backend.Task{Summary: "Task in list"})

	// Soft-delete the list
	err := b.DeleteList(ctx, list.ID)
	if err != nil {
		t.Fatalf("DeleteList error: %v", err)
	}

	// List should not be in active lists
	lists, err := b.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists error: %v", err)
	}
	for _, l := range lists {
		if l.ID == list.ID {
			t.Error("deleted list should not appear in GetLists")
		}
	}

	// List should be in deleted lists
	deleted, err := b.GetDeletedLists(ctx)
	if err != nil {
		t.Fatalf("GetDeletedLists error: %v", err)
	}
	found := false
	for _, l := range deleted {
		if l.ID == list.ID {
			found = true
			if l.DeletedAt == nil {
				t.Error("deleted_at should be set")
			}
		}
	}
	if !found {
		t.Error("list should be in deleted lists")
	}
}

// TestRestoreList verifies that RestoreList restores a soft-deleted list.
func TestRestoreList(t *testing.T) {
	b, ctx := mustNewBackend(t)

	list := mustCreateList(t, b, ctx, "ToRestore")

	// Soft-delete and restore
	err := b.DeleteList(ctx, list.ID)
	if err != nil {
		t.Fatalf("DeleteList error: %v", err)
	}

	err = b.RestoreList(ctx, list.ID)
	if err != nil {
		t.Fatalf("RestoreList error: %v", err)
	}

	// List should be back in active lists
	restored, err := b.GetList(ctx, list.ID)
	if err != nil {
		t.Fatalf("GetList error: %v", err)
	}
	if restored == nil {
		t.Error("restored list should exist")
	}

	// List should not be in deleted lists
	deleted, err := b.GetDeletedLists(ctx)
	if err != nil {
		t.Fatalf("GetDeletedLists error: %v", err)
	}
	for _, l := range deleted {
		if l.ID == list.ID {
			t.Error("restored list should not be in deleted lists")
		}
	}
}

// TestGetNonExistentList tests getting a list that doesn't exist.
func TestGetNonExistentList(t *testing.T) {
	b, ctx := mustNewBackend(t)

	list, err := b.GetList(ctx, "nonexistent-id")
	// Should return nil or an error for non-existent list
	if list != nil && err == nil {
		t.Error("GetList should return nil or error for nonexistent list")
	}
}

// TestGetNonExistentTask tests getting a task that doesn't exist.
func TestGetNonExistentTask(t *testing.T) {
	b, ctx := mustNewBackend(t)

	list := mustCreateList(t, b, ctx, "Test List")

	task, err := b.GetTask(ctx, list.ID, "nonexistent-task-id")
	// Should return nil or an error for non-existent task
	if task != nil && err == nil {
		t.Error("GetTask should return nil or error for nonexistent task")
	}
}

// TestTaskDescription tests that task description is properly stored.
func TestTaskDescription(t *testing.T) {
	b, ctx := mustNewBackend(t)

	list := mustCreateList(t, b, ctx, "Test")

	task := mustCreateTask(t, b, ctx, list.ID, &backend.Task{
		Summary:     "Task with description",
		Description: "This is a detailed description of the task.",
	})

	retrieved, err := b.GetTask(ctx, list.ID, task.ID)
	if err != nil {
		t.Fatalf("GetTask error: %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetTask returned nil")
	}
	if retrieved.Description != "This is a detailed description of the task." {
		t.Errorf("Description = %q, want %q", retrieved.Description, "This is a detailed description of the task.")
	}
}

// TestTaskDueDate tests that task due date is properly stored.
func TestTaskDueDate(t *testing.T) {
	b, ctx := mustNewBackend(t)

	list := mustCreateList(t, b, ctx, "Test")

	dueDate := time.Date(2025, 12, 31, 23, 59, 0, 0, time.UTC)
	task := mustCreateTask(t, b, ctx, list.ID, &backend.Task{
		Summary: "Task with due date",
		DueDate: &dueDate,
	})

	retrieved, err := b.GetTask(ctx, list.ID, task.ID)
	if err != nil {
		t.Fatalf("GetTask error: %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetTask returned nil")
	}
	if retrieved.DueDate == nil {
		t.Fatal("DueDate is nil")
	}
	if !retrieved.DueDate.Equal(dueDate) {
		t.Errorf("DueDate = %v, want %v", *retrieved.DueDate, dueDate)
	}
}

// TestAllTaskStatuses tests all task statuses are properly stored.
func TestAllTaskStatuses(t *testing.T) {
	b, ctx := mustNewBackend(t)

	list := mustCreateList(t, b, ctx, "Status Test")

	statuses := []backend.TaskStatus{
		backend.StatusNeedsAction,
		backend.StatusCompleted,
		backend.StatusInProgress,
		backend.StatusCancelled,
	}

	for _, status := range statuses {
		task := mustCreateTask(t, b, ctx, list.ID, &backend.Task{
			Summary: "Task with status " + string(status),
			Status:  status,
		})

		retrieved, err := b.GetTask(ctx, list.ID, task.ID)
		if err != nil {
			t.Fatalf("GetTask error: %v", err)
		}
		if retrieved == nil {
			t.Fatalf("GetTask returned nil for status %v", status)
		}
		if retrieved.Status != status {
			t.Errorf("status = %v, want %v", retrieved.Status, status)
		}
	}
}

// TestBackendIsolation_Issue007 verifies that backends don't share data.
// Issue #007: When multiple backends use the same database file (sync cache),
// they must have isolated data via backend_id column.
func TestBackendIsolation_Issue007(t *testing.T) {
	// Create two backends pointing to the same database file
	// (simulating the sync cache scenario)
	tmpFile := t.TempDir() + "/shared.db"

	// Create "sqlite" backend (local)
	sqliteBE, err := NewWithBackendID(tmpFile, "sqlite")
	if err != nil {
		t.Fatalf("NewWithBackendID for sqlite: %v", err)
	}
	defer func() { _ = sqliteBE.Close() }()

	// Create "nextcloud-test" backend (remote cache)
	nextcloudBE, err := NewWithBackendID(tmpFile, "nextcloud-test")
	if err != nil {
		t.Fatalf("NewWithBackendID for nextcloud-test: %v", err)
	}
	defer func() { _ = nextcloudBE.Close() }()

	ctx := context.Background()

	// Create a list and task in sqlite backend
	sqliteList, err := sqliteBE.CreateList(ctx, "LocalList")
	if err != nil {
		t.Fatalf("CreateList for sqlite: %v", err)
	}
	sqliteTask, err := sqliteBE.CreateTask(ctx, sqliteList.ID, &backend.Task{Summary: "Local task"})
	if err != nil {
		t.Fatalf("CreateTask for sqlite: %v", err)
	}

	// Create a list and task in nextcloud backend
	nextcloudList, err := nextcloudBE.CreateList(ctx, "RemoteList")
	if err != nil {
		t.Fatalf("CreateList for nextcloud: %v", err)
	}
	nextcloudTask, err := nextcloudBE.CreateTask(ctx, nextcloudList.ID, &backend.Task{Summary: "Remote task"})
	if err != nil {
		t.Fatalf("CreateTask for nextcloud: %v", err)
	}

	// Verify sqlite backend only sees its own data
	sqliteLists, err := sqliteBE.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists for sqlite: %v", err)
	}
	if len(sqliteLists) != 1 {
		t.Errorf("sqlite backend has %d lists, want 1", len(sqliteLists))
	}
	if len(sqliteLists) > 0 && sqliteLists[0].Name != "LocalList" {
		t.Errorf("sqlite list name = %q, want %q", sqliteLists[0].Name, "LocalList")
	}

	sqliteTasks, err := sqliteBE.GetTasks(ctx, sqliteList.ID)
	if err != nil {
		t.Fatalf("GetTasks for sqlite: %v", err)
	}
	if len(sqliteTasks) != 1 {
		t.Errorf("sqlite list has %d tasks, want 1", len(sqliteTasks))
	}
	if len(sqliteTasks) > 0 && sqliteTasks[0].Summary != "Local task" {
		t.Errorf("sqlite task summary = %q, want %q", sqliteTasks[0].Summary, "Local task")
	}

	// Verify nextcloud backend only sees its own data
	nextcloudLists, err := nextcloudBE.GetLists(ctx)
	if err != nil {
		t.Fatalf("GetLists for nextcloud: %v", err)
	}
	if len(nextcloudLists) != 1 {
		t.Errorf("nextcloud backend has %d lists, want 1", len(nextcloudLists))
	}
	if len(nextcloudLists) > 0 && nextcloudLists[0].Name != "RemoteList" {
		t.Errorf("nextcloud list name = %q, want %q", nextcloudLists[0].Name, "RemoteList")
	}

	nextcloudTasks, err := nextcloudBE.GetTasks(ctx, nextcloudList.ID)
	if err != nil {
		t.Fatalf("GetTasks for nextcloud: %v", err)
	}
	if len(nextcloudTasks) != 1 {
		t.Errorf("nextcloud list has %d tasks, want 1", len(nextcloudTasks))
	}
	if len(nextcloudTasks) > 0 && nextcloudTasks[0].Summary != "Remote task" {
		t.Errorf("nextcloud task summary = %q, want %q", nextcloudTasks[0].Summary, "Remote task")
	}

	// Verify cross-backend isolation - sqlite backend shouldn't see nextcloud's task
	crossCheckTask, err := sqliteBE.GetTask(ctx, nextcloudList.ID, nextcloudTask.ID)
	if err != nil {
		t.Fatalf("GetTask cross-check: %v", err)
	}
	if crossCheckTask != nil {
		t.Errorf("sqlite backend should NOT see nextcloud's task, but found: %+v", crossCheckTask)
	}

	// Verify cross-backend isolation - nextcloud backend shouldn't see sqlite's task
	crossCheckTask2, err := nextcloudBE.GetTask(ctx, sqliteList.ID, sqliteTask.ID)
	if err != nil {
		t.Fatalf("GetTask cross-check 2: %v", err)
	}
	if crossCheckTask2 != nil {
		t.Errorf("nextcloud backend should NOT see sqlite's task, but found: %+v", crossCheckTask2)
	}
}
