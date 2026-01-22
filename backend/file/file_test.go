package file_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"todoat/backend"
	"todoat/backend/file"
)

// =============================================================================
// Test Helpers
// =============================================================================

// testFile creates a temporary file for testing
func testFile(t *testing.T, content string) (filePath string, cleanup func()) {
	t.Helper()

	tmpDir := t.TempDir()
	filePath = filepath.Join(tmpDir, "tasks.txt")

	if content != "" {
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	return filePath, func() {
		// Cleanup handled by t.TempDir()
	}
}

// =============================================================================
// TestFileBackendAddTask - todoat --backend=file tasks.txt add "New task"
// =============================================================================

func TestFileBackendAddTask(t *testing.T) {
	t.Run("creates task in new file", func(t *testing.T) {
		filePath, cleanup := testFile(t, "")
		defer cleanup()

		cfg := file.Config{
			FilePath: filePath,
		}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()

		// Create default list
		list, err := be.CreateList(ctx, "Tasks")
		if err != nil {
			t.Fatalf("CreateList error: %v", err)
		}

		// Add a task
		task := &backend.Task{
			Summary: "New task",
			Status:  backend.StatusNeedsAction,
		}
		created, err := be.CreateTask(ctx, list.ID, task)
		if err != nil {
			t.Fatalf("CreateTask error: %v", err)
		}

		if created.Summary != "New task" {
			t.Errorf("expected summary 'New task', got '%s'", created.Summary)
		}
		if created.ID == "" {
			t.Error("expected task to have ID")
		}

		// Verify task is in file
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if !strings.Contains(string(data), "New task") {
			t.Errorf("expected file to contain 'New task', got:\n%s", data)
		}
	})

	t.Run("add task to existing list", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Existing task

`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, err := be.GetListByName(ctx, "Work")
		if err != nil {
			t.Fatalf("GetListByName error: %v", err)
		}
		if list == nil {
			t.Fatal("expected to find Work list")
		}

		// Add a task
		task := &backend.Task{
			Summary: "New task from test",
			Status:  backend.StatusNeedsAction,
		}
		created, err := be.CreateTask(ctx, list.ID, task)
		if err != nil {
			t.Fatalf("CreateTask error: %v", err)
		}

		if created.Summary != "New task from test" {
			t.Errorf("expected summary 'New task from test', got '%s'", created.Summary)
		}

		// Verify task is in list
		tasks, err := be.GetTasks(ctx, list.ID)
		if err != nil {
			t.Fatalf("GetTasks error: %v", err)
		}

		found := false
		for _, task := range tasks {
			if task.Summary == "New task from test" {
				found = true
				break
			}
		}
		if !found {
			t.Error("created task not found in GetTasks")
		}
	})
}

// =============================================================================
// TestFileBackendGetTasks - todoat --backend=file tasks.txt lists tasks
// =============================================================================

func TestFileBackendGetTasks(t *testing.T) {
	t.Run("get tasks from list", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1
- [ ] Task 2
- [x] Completed task
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")
		if list == nil {
			t.Fatal("expected to find Work list")
		}

		tasks, err := be.GetTasks(ctx, list.ID)
		if err != nil {
			t.Fatalf("GetTasks error: %v", err)
		}

		if len(tasks) != 3 {
			t.Errorf("expected 3 tasks, got %d", len(tasks))
		}
	})

	t.Run("task status parsed correctly", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Todo task
- [x] Done task
- [~] In progress task
- [-] Cancelled task
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")
		tasks, _ := be.GetTasks(ctx, list.ID)

		statusMap := make(map[string]backend.TaskStatus)
		for _, task := range tasks {
			statusMap[task.Summary] = task.Status
		}

		if statusMap["Todo task"] != backend.StatusNeedsAction {
			t.Errorf("expected NEEDS-ACTION status for 'Todo task', got %s", statusMap["Todo task"])
		}
		if statusMap["Done task"] != backend.StatusCompleted {
			t.Errorf("expected COMPLETED status for 'Done task', got %s", statusMap["Done task"])
		}
		if statusMap["In progress task"] != backend.StatusInProgress {
			t.Errorf("expected IN-PROGRESS status for 'In progress task', got %s", statusMap["In progress task"])
		}
		if statusMap["Cancelled task"] != backend.StatusCancelled {
			t.Errorf("expected CANCELLED status for 'Cancelled task', got %s", statusMap["Cancelled task"])
		}
	})
}

// =============================================================================
// TestFileBackendUpdateTask - todoat --backend=file tasks.txt update "task" -s D
// =============================================================================

func TestFileBackendUpdateTask(t *testing.T) {
	t.Run("update task summary", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Original task
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")
		tasks, _ := be.GetTasks(ctx, list.ID)

		task := tasks[0]
		task.Summary = "Updated task"

		updated, err := be.UpdateTask(ctx, list.ID, &task)
		if err != nil {
			t.Fatalf("UpdateTask error: %v", err)
		}

		if updated.Summary != "Updated task" {
			t.Errorf("expected summary 'Updated task', got '%s'", updated.Summary)
		}

		// Verify the change persisted in file
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if !strings.Contains(string(data), "Updated task") {
			t.Error("task summary not persisted to file")
		}
		if strings.Contains(string(data), "Original task") {
			t.Error("old task summary still in file")
		}
	})

	t.Run("update task status to done", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task to complete
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")
		tasks, _ := be.GetTasks(ctx, list.ID)

		task := tasks[0]
		task.Status = backend.StatusCompleted
		now := time.Now()
		task.Completed = &now

		_, err = be.UpdateTask(ctx, list.ID, &task)
		if err != nil {
			t.Fatalf("UpdateTask error: %v", err)
		}

		// Verify the change persisted
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if !strings.Contains(string(data), "[x]") {
			t.Error("task status not updated to done in file")
		}
	})
}

// =============================================================================
// TestFileBackendDeleteTask - todoat --backend=file tasks.txt delete "task"
// =============================================================================

func TestFileBackendDeleteTask(t *testing.T) {
	t.Run("delete task removes from file", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Keep this task
- [ ] Delete this task
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")
		tasks, _ := be.GetTasks(ctx, list.ID)

		// Find task to delete
		var taskToDelete *backend.Task
		for _, task := range tasks {
			if task.Summary == "Delete this task" {
				taskToDelete = &task
				break
			}
		}
		if taskToDelete == nil {
			t.Fatal("task to delete not found")
		}

		// Delete the task
		if err := be.DeleteTask(ctx, list.ID, taskToDelete.ID); err != nil {
			t.Fatalf("DeleteTask error: %v", err)
		}

		// Verify deletion in file
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if strings.Contains(string(data), "Delete this task") {
			t.Error("deleted task still in file")
		}
		if !strings.Contains(string(data), "Keep this task") {
			t.Error("other task was incorrectly deleted")
		}
	})
}

// =============================================================================
// TestFileBackendListManagement - Sections in file treated as task lists
// =============================================================================

func TestFileBackendListManagement(t *testing.T) {
	t.Run("sections parsed as lists", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Work task 1
- [ ] Work task 2

## Personal

- [ ] Personal task

## Shopping

- [ ] Buy milk
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		lists, err := be.GetLists(ctx)
		if err != nil {
			t.Fatalf("GetLists error: %v", err)
		}

		if len(lists) != 3 {
			t.Errorf("expected 3 lists, got %d", len(lists))
		}

		// Check list names
		names := make(map[string]bool)
		for _, l := range lists {
			names[l.Name] = true
		}
		if !names["Work"] {
			t.Error("expected 'Work' list")
		}
		if !names["Personal"] {
			t.Error("expected 'Personal' list")
		}
		if !names["Shopping"] {
			t.Error("expected 'Shopping' list")
		}
	})

	t.Run("create list adds section", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Sample task
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		newList, err := be.CreateList(ctx, "NewSection")
		if err != nil {
			t.Fatalf("CreateList error: %v", err)
		}

		if newList.Name != "NewSection" {
			t.Errorf("expected list name 'NewSection', got '%s'", newList.Name)
		}

		// Verify the section exists in file
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if !strings.Contains(string(data), "## NewSection") {
			t.Error("new section not added to file")
		}
	})

	t.Run("delete list removes section", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1

## ToDelete

- [ ] Task to delete
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()

		// Find the list to delete
		list, err := be.GetListByName(ctx, "ToDelete")
		if err != nil {
			t.Fatalf("GetListByName error: %v", err)
		}
		if list == nil {
			t.Fatal("expected to find ToDelete list")
		}

		// Delete the list
		if err := be.DeleteList(ctx, list.ID); err != nil {
			t.Fatalf("DeleteList error: %v", err)
		}

		// Verify it's gone from file
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if strings.Contains(string(data), "ToDelete") {
			t.Error("deleted list still in file")
		}
	})
}

// =============================================================================
// TestFileBackendCreateFile - Creates task file if not exists with proper header
// =============================================================================

func TestFileBackendCreateFile(t *testing.T) {
	t.Run("creates file if not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "new_tasks.txt")

		// Verify file doesn't exist
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Fatal("expected file to not exist")
		}

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()

		// Create a list (should trigger file creation)
		_, err = be.CreateList(ctx, "Work")
		if err != nil {
			t.Fatalf("CreateList error: %v", err)
		}

		// Verify file now exists with proper header
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("expected file to exist: %v", err)
		}
		if !strings.Contains(string(data), "# Tasks") {
			t.Error("expected file to have header")
		}
	})
}

// =============================================================================
// TestFileBackendMetadataSupport - Tasks store priority, dates, status, tags
// =============================================================================

func TestFileBackendMetadataSupport(t *testing.T) {
	t.Run("task with priority", func(t *testing.T) {
		filePath, cleanup := testFile(t, "")
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.CreateList(ctx, "Work")

		task := &backend.Task{
			Summary:  "High priority task",
			Status:   backend.StatusNeedsAction,
			Priority: 1,
		}
		created, err := be.CreateTask(ctx, list.ID, task)
		if err != nil {
			t.Fatalf("CreateTask error: %v", err)
		}

		if created.Priority != 1 {
			t.Errorf("expected priority 1, got %d", created.Priority)
		}

		// Verify priority is in file
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if !strings.Contains(string(data), "!1") {
			t.Error("expected priority marker in file")
		}
	})

	t.Run("task with due date", func(t *testing.T) {
		filePath, cleanup := testFile(t, "")
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.CreateList(ctx, "Work")

		dueDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
		task := &backend.Task{
			Summary: "Task with due date",
			Status:  backend.StatusNeedsAction,
			DueDate: &dueDate,
		}
		created, err := be.CreateTask(ctx, list.ID, task)
		if err != nil {
			t.Fatalf("CreateTask error: %v", err)
		}

		if created.DueDate == nil {
			t.Error("expected due date to be set")
		}

		// Verify due date is in file
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if !strings.Contains(string(data), "@2024-06-15") {
			t.Error("expected due date marker in file")
		}
	})

	t.Run("task with categories/tags", func(t *testing.T) {
		filePath, cleanup := testFile(t, "")
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.CreateList(ctx, "Work")

		task := &backend.Task{
			Summary:    "Tagged task",
			Status:     backend.StatusNeedsAction,
			Categories: "urgent,review",
		}
		created, err := be.CreateTask(ctx, list.ID, task)
		if err != nil {
			t.Fatalf("CreateTask error: %v", err)
		}

		if created.Categories != "urgent,review" {
			t.Errorf("expected categories 'urgent,review', got '%s'", created.Categories)
		}

		// Verify tags are in file
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if !strings.Contains(string(data), "#urgent") || !strings.Contains(string(data), "#review") {
			t.Error("expected tag markers in file")
		}
	})

	t.Run("parse metadata from file", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task with metadata !1 @2024-06-15 #urgent #review
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")
		tasks, _ := be.GetTasks(ctx, list.ID)

		if len(tasks) != 1 {
			t.Fatalf("expected 1 task, got %d", len(tasks))
		}

		task := tasks[0]
		if task.Priority != 1 {
			t.Errorf("expected priority 1, got %d", task.Priority)
		}
		if task.DueDate == nil {
			t.Error("expected due date to be set")
		}
		if task.Categories != "urgent,review" {
			t.Errorf("expected categories 'urgent,review', got '%s'", task.Categories)
		}
	})
}

// =============================================================================
// TestFileBackendHierarchy - Indented tasks parsed as subtasks
// =============================================================================

func TestFileBackendHierarchy(t *testing.T) {
	t.Run("indented tasks parsed as subtasks", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Parent task
  - [ ] Subtask 1
  - [ ] Subtask 2
    - [ ] Sub-subtask
- [ ] Another parent
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")
		tasks, err := be.GetTasks(ctx, list.ID)
		if err != nil {
			t.Fatalf("GetTasks error: %v", err)
		}

		// Build task map
		taskMap := make(map[string]*backend.Task)
		for i := range tasks {
			taskMap[tasks[i].Summary] = &tasks[i]
		}

		// Verify parent-child relationships
		parentTask := taskMap["Parent task"]
		subtask1 := taskMap["Subtask 1"]
		subtask2 := taskMap["Subtask 2"]
		subSubtask := taskMap["Sub-subtask"]

		if parentTask == nil {
			t.Fatal("Parent task not found")
		}
		if subtask1 == nil || subtask1.ParentID != parentTask.ID {
			t.Error("Subtask 1 should have Parent task as parent")
		}
		if subtask2 == nil || subtask2.ParentID != parentTask.ID {
			t.Error("Subtask 2 should have Parent task as parent")
		}
		if subSubtask == nil || subSubtask.ParentID != subtask2.ID {
			t.Error("Sub-subtask should have Subtask 2 as parent")
		}

		// Verify Another parent has no parent
		anotherParent := taskMap["Another parent"]
		if anotherParent != nil && anotherParent.ParentID != "" {
			t.Error("Another parent should have no parent")
		}
	})

	t.Run("create subtask with parent", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Parent task
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")
		tasks, _ := be.GetTasks(ctx, list.ID)
		parentTask := tasks[0]

		// Create subtask
		subtask := &backend.Task{
			Summary:  "New subtask",
			Status:   backend.StatusNeedsAction,
			ParentID: parentTask.ID,
		}
		created, err := be.CreateTask(ctx, list.ID, subtask)
		if err != nil {
			t.Fatalf("CreateTask error: %v", err)
		}

		if created.ParentID != parentTask.ID {
			t.Errorf("expected parent ID '%s', got '%s'", parentTask.ID, created.ParentID)
		}

		// Verify hierarchy in file (subtask should be indented)
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		// Should have indented subtask
		if !strings.Contains(string(data), "  - [ ] New subtask") {
			t.Errorf("expected indented subtask in file, got:\n%s", data)
		}
	})
}

// =============================================================================
// TestFileBackendConfigPath - Configurable file path via config
// =============================================================================

func TestFileBackendConfigPath(t *testing.T) {
	t.Run("uses configured file path", func(t *testing.T) {
		tmpDir := t.TempDir()
		customPath := filepath.Join(tmpDir, "custom", "my_tasks.txt")

		// Create directory
		if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}

		cfg := file.Config{FilePath: customPath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()

		// Create a list
		_, err = be.CreateList(ctx, "Test")
		if err != nil {
			t.Fatalf("CreateList error: %v", err)
		}

		// Verify file was created at custom path
		if _, err := os.Stat(customPath); os.IsNotExist(err) {
			t.Error("expected file to be created at custom path")
		}
	})

	t.Run("relative path resolved", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalWd) }()
		_ = os.Chdir(tmpDir)

		cfg := file.Config{FilePath: "tasks.txt"}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()

		// Create a list
		_, err = be.CreateList(ctx, "Test")
		if err != nil {
			t.Fatalf("CreateList error: %v", err)
		}

		// Verify file was created in current directory
		expectedPath := filepath.Join(tmpDir, "tasks.txt")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("expected file at %s", expectedPath)
		}
	})
}

// =============================================================================
// TestFileBackendGetList - GetList returns specific list by ID
// =============================================================================

func TestFileBackendGetList(t *testing.T) {
	t.Run("get existing list by ID", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1

## Personal

- [ ] Task 2
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		lists, err := be.GetLists(ctx)
		if err != nil {
			t.Fatalf("GetLists error: %v", err)
		}

		// Get list by ID
		list, err := be.GetList(ctx, lists[0].ID)
		if err != nil {
			t.Fatalf("GetList error: %v", err)
		}
		if list == nil {
			t.Fatal("expected to find list")
		}
		if list.ID != lists[0].ID {
			t.Errorf("expected list ID %s, got %s", lists[0].ID, list.ID)
		}
	})

	t.Run("get non-existent list returns nil", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, err := be.GetList(ctx, "non-existent-id")
		if err != nil {
			t.Fatalf("GetList error: %v", err)
		}
		if list != nil {
			t.Error("expected nil for non-existent list")
		}
	})
}

// =============================================================================
// TestFileBackendGetTask - GetTask returns specific task by ID
// =============================================================================

func TestFileBackendGetTask(t *testing.T) {
	t.Run("get existing task by ID", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Find me task
- [ ] Another task
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")
		tasks, _ := be.GetTasks(ctx, list.ID)

		// Get task by ID
		task, err := be.GetTask(ctx, list.ID, tasks[0].ID)
		if err != nil {
			t.Fatalf("GetTask error: %v", err)
		}
		if task == nil {
			t.Fatal("expected to find task")
		}
		if task.ID != tasks[0].ID {
			t.Errorf("expected task ID %s, got %s", tasks[0].ID, task.ID)
		}
	})

	t.Run("get non-existent task returns nil", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")

		task, err := be.GetTask(ctx, list.ID, "non-existent-id")
		if err != nil {
			t.Fatalf("GetTask error: %v", err)
		}
		if task != nil {
			t.Error("expected nil for non-existent task")
		}
	})
}

// =============================================================================
// TestFileBackendUpdateList - UpdateList modifies list properties
// =============================================================================

func TestFileBackendUpdateList(t *testing.T) {
	t.Run("update list name", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## OldName

- [ ] Task 1
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "OldName")

		// Update the list name
		list.Name = "NewName"
		list.Color = "blue"
		updated, err := be.UpdateList(ctx, list)
		if err != nil {
			t.Fatalf("UpdateList error: %v", err)
		}

		if updated.Name != "NewName" {
			t.Errorf("expected name 'NewName', got '%s'", updated.Name)
		}
		if updated.Color != "blue" {
			t.Errorf("expected color 'blue', got '%s'", updated.Color)
		}

		// Verify change persisted to file
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if !strings.Contains(string(data), "## NewName") {
			t.Error("list name not updated in file")
		}
	})

	t.Run("update non-existent list returns error", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()

		// Try to update a non-existent list
		fakeList := &backend.List{
			ID:   "non-existent-id",
			Name: "Fake",
		}
		_, err = be.UpdateList(ctx, fakeList)
		if err == nil {
			t.Error("expected error when updating non-existent list")
		}
		if !strings.Contains(err.Error(), "list not found") {
			t.Errorf("expected 'list not found' error, got: %v", err)
		}
	})
}

// =============================================================================
// TestFileBackendDeletedListOperations - GetDeletedLists, RestoreList, PurgeList
// =============================================================================

func TestFileBackendDeletedListOperations(t *testing.T) {
	t.Run("GetDeletedLists returns empty list", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		deletedLists, err := be.GetDeletedLists(ctx)
		if err != nil {
			t.Fatalf("GetDeletedLists error: %v", err)
		}
		if len(deletedLists) != 0 {
			t.Errorf("expected empty list, got %d items", len(deletedLists))
		}
	})

	t.Run("GetDeletedListByName returns nil", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, err := be.GetDeletedListByName(ctx, "Anything")
		if err != nil {
			t.Fatalf("GetDeletedListByName error: %v", err)
		}
		if list != nil {
			t.Error("expected nil for deleted list by name")
		}
	})

	t.Run("RestoreList returns error", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		err = be.RestoreList(ctx, "any-id")
		if err == nil {
			t.Error("expected error from RestoreList")
		}
		if !strings.Contains(err.Error(), "not supported") {
			t.Errorf("expected 'not supported' error, got: %v", err)
		}
	})

	t.Run("PurgeList returns error", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		err = be.PurgeList(ctx, "any-id")
		if err == nil {
			t.Error("expected error from PurgeList")
		}
		if !strings.Contains(err.Error(), "not supported") {
			t.Errorf("expected 'not supported' error, got: %v", err)
		}
	})
}

// =============================================================================
// TestFileBackendErrorCases - Error handling scenarios
// =============================================================================

func TestFileBackendErrorCases(t *testing.T) {
	t.Run("create task in non-existent list returns error", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		task := &backend.Task{
			Summary: "New task",
			Status:  backend.StatusNeedsAction,
		}
		_, err = be.CreateTask(ctx, "non-existent-list-id", task)
		if err == nil {
			t.Error("expected error when creating task in non-existent list")
		}
		if !strings.Contains(err.Error(), "list not found") {
			t.Errorf("expected 'list not found' error, got: %v", err)
		}
	})

	t.Run("update non-existent task returns error", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")

		fakeTask := &backend.Task{
			ID:      "non-existent-id",
			Summary: "Fake",
		}
		_, err = be.UpdateTask(ctx, list.ID, fakeTask)
		if err == nil {
			t.Error("expected error when updating non-existent task")
		}
		if !strings.Contains(err.Error(), "task not found") {
			t.Errorf("expected 'task not found' error, got: %v", err)
		}
	})

	t.Run("delete non-existent task returns error", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")

		err = be.DeleteTask(ctx, list.ID, "non-existent-id")
		if err == nil {
			t.Error("expected error when deleting non-existent task")
		}
		if !strings.Contains(err.Error(), "task not found") {
			t.Errorf("expected 'task not found' error, got: %v", err)
		}
	})

	t.Run("delete non-existent list returns error", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		err = be.DeleteList(ctx, "non-existent-id")
		if err == nil {
			t.Error("expected error when deleting non-existent list")
		}
		if !strings.Contains(err.Error(), "list not found") {
			t.Errorf("expected 'list not found' error, got: %v", err)
		}
	})

	t.Run("GetListByName returns nil for non-existent", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, err := be.GetListByName(ctx, "NonExistent")
		if err != nil {
			t.Fatalf("GetListByName error: %v", err)
		}
		if list != nil {
			t.Error("expected nil for non-existent list by name")
		}
	})

	t.Run("GetTasks returns empty for unknown list", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		tasks, err := be.GetTasks(ctx, "unknown-list-id")
		if err != nil {
			t.Fatalf("GetTasks error: %v", err)
		}
		if len(tasks) != 0 {
			t.Errorf("expected empty task list, got %d", len(tasks))
		}
	})
}

// =============================================================================
// TestFileBackendMalformedFiles - Parsing edge cases
// =============================================================================

func TestFileBackendMalformedFiles(t *testing.T) {
	t.Run("empty file initializes empty state", func(t *testing.T) {
		filePath, cleanup := testFile(t, "")
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		lists, err := be.GetLists(ctx)
		if err != nil {
			t.Fatalf("GetLists error: %v", err)
		}
		if len(lists) != 0 {
			t.Errorf("expected 0 lists, got %d", len(lists))
		}
	})

	t.Run("file with only header", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

Just some random text here
No sections or tasks
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		lists, err := be.GetLists(ctx)
		if err != nil {
			t.Fatalf("GetLists error: %v", err)
		}
		if len(lists) != 0 {
			t.Errorf("expected 0 lists, got %d", len(lists))
		}
	})

	t.Run("tasks before any section are ignored", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

- [ ] Orphan task without section

## Work

- [ ] Task in section
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		lists, _ := be.GetLists(ctx)
		if len(lists) != 1 {
			t.Fatalf("expected 1 list, got %d", len(lists))
		}

		tasks, _ := be.GetTasks(ctx, lists[0].ID)
		if len(tasks) != 1 {
			t.Errorf("expected 1 task, got %d", len(tasks))
		}
		if tasks[0].Summary != "Task in section" {
			t.Errorf("expected 'Task in section', got '%s'", tasks[0].Summary)
		}
	})

	t.Run("empty section", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Empty Section

## Section With Tasks

- [ ] A task
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		lists, _ := be.GetLists(ctx)
		if len(lists) != 2 {
			t.Fatalf("expected 2 lists, got %d", len(lists))
		}

		emptyList, _ := be.GetListByName(ctx, "Empty Section")
		if emptyList == nil {
			t.Fatal("expected to find 'Empty Section' list")
		}

		tasks, _ := be.GetTasks(ctx, emptyList.ID)
		if len(tasks) != 0 {
			t.Errorf("expected 0 tasks in empty section, got %d", len(tasks))
		}
	})

	t.Run("CreateList returns existing if name matches", func(t *testing.T) {
		filePath, cleanup := testFile(t, `# Tasks

## Work

- [ ] Task 1
`)
		defer cleanup()

		cfg := file.Config{FilePath: filePath}
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()

		// Try to create a list with existing name (case-insensitive)
		existingList, _ := be.GetListByName(ctx, "Work")
		list, err := be.CreateList(ctx, "work") // lowercase
		if err != nil {
			t.Fatalf("CreateList error: %v", err)
		}

		if list.ID != existingList.ID {
			t.Error("expected CreateList to return existing list with same name")
		}
	})
}

// =============================================================================
// TestFileBackendDefaultPath - Default path when none configured
// =============================================================================

func TestFileBackendDefaultPath(t *testing.T) {
	t.Run("uses default tasks.txt when no path configured", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalWd) }()
		_ = os.Chdir(tmpDir)

		cfg := file.Config{} // No FilePath set
		be, err := file.New(cfg)
		if err != nil {
			t.Fatalf("failed to create file backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()

		// Create a list to trigger file creation
		_, err = be.CreateList(ctx, "Test")
		if err != nil {
			t.Fatalf("CreateList error: %v", err)
		}

		// Verify default file was created
		expectedPath := filepath.Join(tmpDir, "tasks.txt")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Error("expected tasks.txt to be created in current directory")
		}
	})
}

// =============================================================================
// Interface Compliance
// =============================================================================

func TestFileBackendInterfaceCompliance(t *testing.T) {
	// Verify Backend implements TaskManager interface at compile time
	var _ backend.TaskManager = (*file.Backend)(nil)
}
