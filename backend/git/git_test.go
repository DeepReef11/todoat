package git_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"todoat/backend"
	"todoat/backend/git"
)

// =============================================================================
// Test Helpers
// =============================================================================

// testRepo creates a temporary git repository with an optional TODO.md file
func testRepo(t *testing.T, withMarker bool) (repoPath string, cleanup func()) {
	t.Helper()

	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git for commits
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	if withMarker {
		// Create TODO.md with marker
		todoPath := filepath.Join(tmpDir, "TODO.md")
		content := `<!-- todoat:enabled -->
# Project Tasks

## Work

- [ ] Sample task

## Personal

`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}
	}

	return tmpDir, func() {
		// Cleanup handled by t.TempDir()
	}
}

// =============================================================================
// Detection Tests (TestGitBackendDetection)
// =============================================================================

func TestGitBackendDetection(t *testing.T) {
	t.Run("detects git repo with marked TODO.md", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{
			WorkDir: repoPath,
		}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		// Check detection
		canDetect, err := be.CanDetect()
		if err != nil {
			t.Fatalf("CanDetect error: %v", err)
		}
		if !canDetect {
			t.Error("expected CanDetect to return true for repo with marked TODO.md")
		}

		// Check detection info
		info := be.DetectionInfo()
		if !strings.Contains(info, "Git repository") {
			t.Errorf("expected detection info to mention Git repository, got: %s", info)
		}
		if !strings.Contains(info, "TODO.md") {
			t.Errorf("expected detection info to mention TODO.md, got: %s", info)
		}
	})

	t.Run("does not detect repo without marker", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		// Create TODO.md without marker
		todoPath := filepath.Join(repoPath, "TODO.md")
		content := "# Tasks\n- [ ] Task\n"
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{
			WorkDir: repoPath,
		}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		canDetect, err := be.CanDetect()
		if err != nil {
			t.Fatalf("CanDetect error: %v", err)
		}
		if canDetect {
			t.Error("expected CanDetect to return false for repo without marker")
		}
	})

	t.Run("does not detect non-git directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create TODO.md with marker but no git repo
		todoPath := filepath.Join(tmpDir, "TODO.md")
		content := "<!-- todoat:enabled -->\n# Tasks\n"
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{
			WorkDir: tmpDir,
		}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		canDetect, err := be.CanDetect()
		if err != nil {
			t.Fatalf("CanDetect error: %v", err)
		}
		if canDetect {
			t.Error("expected CanDetect to return false for non-git directory")
		}
	})
}

// =============================================================================
// Marker Required Tests (TestGitBackendMarkerRequired)
// =============================================================================

func TestGitBackendMarkerRequired(t *testing.T) {
	t.Run("marker required for detection", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		// Create TODO.md without the marker
		todoPath := filepath.Join(repoPath, "TODO.md")
		content := "# Tasks\n\n## Work\n\n- [ ] Task 1\n"
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{
			WorkDir: repoPath,
		}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		canDetect, _ := be.CanDetect()
		if canDetect {
			t.Error("expected CanDetect to return false without marker")
		}
	})

	t.Run("marker at start of file", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := "<!-- todoat:enabled -->\n# Tasks\n\n## Work\n\n- [ ] Task 1\n"
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{
			WorkDir: repoPath,
		}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		canDetect, _ := be.CanDetect()
		if !canDetect {
			t.Error("expected CanDetect to return true with marker at start")
		}
	})

	t.Run("marker anywhere in file", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := "# Project Title\n\nSome description.\n\n<!-- todoat:enabled -->\n\n## Tasks\n\n- [ ] Task 1\n"
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{
			WorkDir: repoPath,
		}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		canDetect, _ := be.CanDetect()
		if !canDetect {
			t.Error("expected CanDetect to return true with marker anywhere in file")
		}
	})
}

// =============================================================================
// Fallback Files Tests (TestGitBackendFallbackFiles)
// =============================================================================

func TestGitBackendFallbackFiles(t *testing.T) {
	t.Run("finds configured file first", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		// Create custom file with marker
		customPath := filepath.Join(repoPath, "TASKS.md")
		content := "<!-- todoat:enabled -->\n# Tasks\n\n## Work\n\n- [ ] Custom task\n"
		if err := os.WriteFile(customPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TASKS.md: %v", err)
		}

		cfg := git.Config{
			WorkDir: repoPath,
			File:    "TASKS.md",
		}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		canDetect, _ := be.CanDetect()
		if !canDetect {
			t.Error("expected to detect configured file")
		}
	})

	t.Run("falls back to TODO.md", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true) // Creates TODO.md with marker
		defer cleanup()

		cfg := git.Config{
			WorkDir: repoPath,
			// No File configured, should fall back to TODO.md
		}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		canDetect, _ := be.CanDetect()
		if !canDetect {
			t.Error("expected to fall back to TODO.md")
		}
	})

	t.Run("falls back to todo.md", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		// Create lowercase todo.md with marker
		todoPath := filepath.Join(repoPath, "todo.md")
		content := "<!-- todoat:enabled -->\n# Tasks\n\n## Work\n"
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create todo.md: %v", err)
		}

		cfg := git.Config{
			WorkDir: repoPath,
		}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		canDetect, _ := be.CanDetect()
		if !canDetect {
			t.Error("expected to fall back to todo.md")
		}
	})

	t.Run("falls back to .todoat.md", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		// Create .todoat.md with marker
		todoPath := filepath.Join(repoPath, ".todoat.md")
		content := "<!-- todoat:enabled -->\n# Tasks\n\n## Work\n"
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create .todoat.md: %v", err)
		}

		cfg := git.Config{
			WorkDir: repoPath,
		}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		canDetect, _ := be.CanDetect()
		if !canDetect {
			t.Error("expected to fall back to .todoat.md")
		}
	})

	t.Run("uses fallback_files config", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		// Create custom file with marker
		customPath := filepath.Join(repoPath, "MY_TASKS.md")
		content := "<!-- todoat:enabled -->\n# Tasks\n\n## Work\n"
		if err := os.WriteFile(customPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create MY_TASKS.md: %v", err)
		}

		cfg := git.Config{
			WorkDir:       repoPath,
			FallbackFiles: []string{"MY_TASKS.md", "OTHER.md"},
		}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		canDetect, _ := be.CanDetect()
		if !canDetect {
			t.Error("expected to find file from fallback_files")
		}
	})
}

// =============================================================================
// List Management Tests (TestGitBackendListManagement)
// =============================================================================

func TestGitBackendListManagement(t *testing.T) {
	t.Run("sections parsed as lists", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := `<!-- todoat:enabled -->
# Project Tasks

## Work

- [ ] Work task 1
- [ ] Work task 2

## Personal

- [ ] Personal task

## Shopping

- [ ] Buy milk
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
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
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
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

		// Verify the list exists
		lists, err := be.GetLists(ctx)
		if err != nil {
			t.Fatalf("GetLists error: %v", err)
		}

		found := false
		for _, l := range lists {
			if l.Name == "NewSection" {
				found = true
				break
			}
		}
		if !found {
			t.Error("created list not found in GetLists")
		}
	})

	t.Run("delete list removes section", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := `<!-- todoat:enabled -->
## Work

- [ ] Task 1

## ToDelete

- [ ] Task to delete
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
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

		// Verify it's gone
		lists, err := be.GetLists(ctx)
		if err != nil {
			t.Fatalf("GetLists error: %v", err)
		}

		for _, l := range lists {
			if l.Name == "ToDelete" {
				t.Error("deleted list still exists")
			}
		}
	})
}

// =============================================================================
// Add Task Tests (TestGitBackendAddTask)
// =============================================================================

func TestGitBackendAddTask(t *testing.T) {
	t.Run("add task to existing list", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true) // Has Work and Personal lists
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()

		// Get Work list
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
		if created.ID == "" {
			t.Error("expected task to have ID")
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

	t.Run("add task with priority", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")

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
	})

	t.Run("add task with due date", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")

		dueDate := time.Now().Add(24 * time.Hour).Truncate(time.Second)
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
	})

	t.Run("add task with categories", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")

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
	})
}

// =============================================================================
// Get Tasks Tests (TestGitBackendGetTasks)
// =============================================================================

func TestGitBackendGetTasks(t *testing.T) {
	t.Run("get tasks from list", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := `<!-- todoat:enabled -->
## Work

- [ ] Task 1
- [ ] Task 2
- [x] Completed task
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")

		tasks, err := be.GetTasks(ctx, list.ID)
		if err != nil {
			t.Fatalf("GetTasks error: %v", err)
		}

		if len(tasks) != 3 {
			t.Errorf("expected 3 tasks, got %d", len(tasks))
		}
	})

	t.Run("task status parsed correctly", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := `<!-- todoat:enabled -->
## Work

- [ ] Todo task
- [x] Done task
- [~] In progress task
- [-] Cancelled task
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
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
			t.Errorf("expected TODO status for 'Todo task', got %s", statusMap["Todo task"])
		}
		if statusMap["Done task"] != backend.StatusCompleted {
			t.Errorf("expected DONE status for 'Done task', got %s", statusMap["Done task"])
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
// Update Task Tests (TestGitBackendUpdateTask)
// =============================================================================

func TestGitBackendUpdateTask(t *testing.T) {
	t.Run("update task summary", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := `<!-- todoat:enabled -->
## Work

- [ ] Original task
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
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

		// Verify the change persisted
		tasks, _ = be.GetTasks(ctx, list.ID)
		if tasks[0].Summary != "Updated task" {
			t.Error("task summary not persisted")
		}
	})

	t.Run("update task status to done", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := `<!-- todoat:enabled -->
## Work

- [ ] Task to complete
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
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
		tasks, _ = be.GetTasks(ctx, list.ID)
		if tasks[0].Status != backend.StatusCompleted {
			t.Errorf("expected status COMPLETED, got %s", tasks[0].Status)
		}
	})
}

// =============================================================================
// Delete Task Tests (TestGitBackendDeleteTask)
// =============================================================================

func TestGitBackendDeleteTask(t *testing.T) {
	t.Run("delete task removes from file", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := `<!-- todoat:enabled -->
## Work

- [ ] Keep this task
- [ ] Delete this task
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
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

		// Verify deletion
		tasks, _ = be.GetTasks(ctx, list.ID)
		for _, task := range tasks {
			if task.Summary == "Delete this task" {
				t.Error("deleted task still exists")
			}
		}

		// Verify other task still exists
		found := false
		for _, task := range tasks {
			if task.Summary == "Keep this task" {
				found = true
				break
			}
		}
		if !found {
			t.Error("other task was incorrectly deleted")
		}
	})
}

// =============================================================================
// Hierarchy Tests (TestGitBackendHierarchy)
// =============================================================================

func TestGitBackendHierarchy(t *testing.T) {
	t.Run("indented tasks parsed as subtasks", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := `<!-- todoat:enabled -->
## Work

- [ ] Parent task
  - [ ] Subtask 1
  - [ ] Subtask 2
    - [ ] Sub-subtask
- [ ] Another parent
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
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
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := `<!-- todoat:enabled -->
## Work

- [ ] Parent task
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
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

		// Verify hierarchy in file
		tasks, _ = be.GetTasks(ctx, list.ID)
		var foundSubtask *backend.Task
		for i := range tasks {
			if tasks[i].Summary == "New subtask" {
				foundSubtask = &tasks[i]
				break
			}
		}
		if foundSubtask == nil || foundSubtask.ParentID != parentTask.ID {
			t.Error("subtask parent relationship not persisted")
		}
	})
}

// =============================================================================
// Auto-Commit Tests (TestGitBackendAutoCommit)
// =============================================================================

func TestGitBackendAutoCommit(t *testing.T) {
	t.Run("auto-commit disabled by default", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		// Make initial commit
		cmd := exec.Command("git", "add", ".")
		cmd.Dir = repoPath
		_ = cmd.Run()
		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = repoPath
		_ = cmd.Run()

		cfg := git.Config{
			WorkDir:    repoPath,
			AutoCommit: false,
		}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")

		// Add a task
		task := &backend.Task{
			Summary: "Test task",
			Status:  backend.StatusNeedsAction,
		}
		_, err = be.CreateTask(ctx, list.ID, task)
		if err != nil {
			t.Fatalf("CreateTask error: %v", err)
		}

		// Check git status - should have uncommitted changes
		cmd = exec.Command("git", "status", "--porcelain")
		cmd.Dir = repoPath
		output, _ := cmd.Output()
		if len(output) == 0 {
			t.Error("expected uncommitted changes with auto_commit disabled")
		}
	})

	t.Run("auto-commit creates commit on task add", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		// Make initial commit
		cmd := exec.Command("git", "add", ".")
		cmd.Dir = repoPath
		_ = cmd.Run()
		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = repoPath
		_ = cmd.Run()

		cfg := git.Config{
			WorkDir:    repoPath,
			AutoCommit: true,
		}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")

		// Add a task
		task := &backend.Task{
			Summary: "Auto-commit test task",
			Status:  backend.StatusNeedsAction,
		}
		_, err = be.CreateTask(ctx, list.ID, task)
		if err != nil {
			t.Fatalf("CreateTask error: %v", err)
		}

		// Check git status - should be clean
		cmd = exec.Command("git", "status", "--porcelain")
		cmd.Dir = repoPath
		output, _ := cmd.Output()
		if len(strings.TrimSpace(string(output))) != 0 {
			t.Errorf("expected clean working directory with auto_commit enabled, got: %s", output)
		}

		// Check commit message
		cmd = exec.Command("git", "log", "-1", "--pretty=%s")
		cmd.Dir = repoPath
		output, _ = cmd.Output()
		commitMsg := strings.TrimSpace(string(output))
		if !strings.Contains(commitMsg, "todoat:") || !strings.Contains(strings.ToLower(commitMsg), "add") {
			t.Errorf("expected commit message to mention todoat and add, got: %s", commitMsg)
		}
	})

	t.Run("auto-commit creates commit on task update", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := `<!-- todoat:enabled -->
## Work

- [ ] Task to update
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		// Make initial commit
		cmd := exec.Command("git", "add", ".")
		cmd.Dir = repoPath
		_ = cmd.Run()
		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = repoPath
		_ = cmd.Run()

		cfg := git.Config{
			WorkDir:    repoPath,
			AutoCommit: true,
		}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")
		tasks, _ := be.GetTasks(ctx, list.ID)

		task := tasks[0]
		task.Status = backend.StatusCompleted

		_, err = be.UpdateTask(ctx, list.ID, &task)
		if err != nil {
			t.Fatalf("UpdateTask error: %v", err)
		}

		// Check git status - should be clean
		cmd = exec.Command("git", "status", "--porcelain")
		cmd.Dir = repoPath
		output, _ := cmd.Output()
		if len(strings.TrimSpace(string(output))) != 0 {
			t.Errorf("expected clean working directory, got: %s", output)
		}
	})

	t.Run("auto-commit creates commit on task delete", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := `<!-- todoat:enabled -->
## Work

- [ ] Task to delete
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		// Make initial commit
		cmd := exec.Command("git", "add", ".")
		cmd.Dir = repoPath
		_ = cmd.Run()
		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = repoPath
		_ = cmd.Run()

		cfg := git.Config{
			WorkDir:    repoPath,
			AutoCommit: true,
		}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")
		tasks, _ := be.GetTasks(ctx, list.ID)
		task := tasks[0]

		err = be.DeleteTask(ctx, list.ID, task.ID)
		if err != nil {
			t.Fatalf("DeleteTask error: %v", err)
		}

		// Check git status - should be clean
		cmd = exec.Command("git", "status", "--porcelain")
		cmd.Dir = repoPath
		output, _ := cmd.Output()
		if len(strings.TrimSpace(string(output))) != 0 {
			t.Errorf("expected clean working directory, got: %s", output)
		}
	})
}

// =============================================================================
// Interface Compliance
// =============================================================================

func TestGitBackendInterfaceCompliance(t *testing.T) {
	// Verify Backend implements TaskManager interface at compile time
	var _ backend.TaskManager = (*git.Backend)(nil)
}

// =============================================================================
// GetList Tests (TestGitBackendGetList)
// =============================================================================

func TestGitBackendGetList(t *testing.T) {
	t.Run("get list by ID", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := `<!-- todoat:enabled -->
## Work

- [ ] Task 1

## Personal

- [ ] Task 2
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()

		// First get all lists to find an ID
		lists, err := be.GetLists(ctx)
		if err != nil {
			t.Fatalf("GetLists error: %v", err)
		}
		if len(lists) == 0 {
			t.Fatal("expected at least one list")
		}

		// Now get that specific list by ID
		list, err := be.GetList(ctx, lists[0].ID)
		if err != nil {
			t.Fatalf("GetList error: %v", err)
		}
		if list == nil {
			t.Fatal("expected to find list by ID")
		}
		if list.ID != lists[0].ID {
			t.Errorf("expected list ID %s, got %s", lists[0].ID, list.ID)
		}
	})

	t.Run("get list by non-existent ID returns nil", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, err := be.GetList(ctx, "non-existent-id")
		if err != nil {
			t.Fatalf("GetList error: %v", err)
		}
		if list != nil {
			t.Error("expected nil for non-existent list ID")
		}
	})
}

// =============================================================================
// GetTask Tests (TestGitBackendGetTask)
// =============================================================================

func TestGitBackendGetTask(t *testing.T) {
	t.Run("get task by ID", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := `<!-- todoat:enabled -->
## Work

- [ ] Task 1
- [ ] Task 2
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")
		tasks, err := be.GetTasks(ctx, list.ID)
		if err != nil {
			t.Fatalf("GetTasks error: %v", err)
		}
		if len(tasks) == 0 {
			t.Fatal("expected at least one task")
		}

		// Get specific task by ID
		task, err := be.GetTask(ctx, list.ID, tasks[0].ID)
		if err != nil {
			t.Fatalf("GetTask error: %v", err)
		}
		if task == nil {
			t.Fatal("expected to find task by ID")
		}
		if task.ID != tasks[0].ID {
			t.Errorf("expected task ID %s, got %s", tasks[0].ID, task.ID)
		}
	})

	t.Run("get task by non-existent ID returns nil", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")
		task, err := be.GetTask(ctx, list.ID, "non-existent-id")
		if err != nil {
			t.Fatalf("GetTask error: %v", err)
		}
		if task != nil {
			t.Error("expected nil for non-existent task ID")
		}
	})
}

// =============================================================================
// UpdateList Tests (TestGitBackendUpdateList)
// =============================================================================

func TestGitBackendUpdateList(t *testing.T) {
	t.Run("update list name", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := `<!-- todoat:enabled -->
## Work

- [ ] Task 1
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, err := be.GetListByName(ctx, "Work")
		if err != nil {
			t.Fatalf("GetListByName error: %v", err)
		}

		// Update the list name
		list.Name = "Updated Work"
		updated, err := be.UpdateList(ctx, list)
		if err != nil {
			t.Fatalf("UpdateList error: %v", err)
		}

		if updated.Name != "Updated Work" {
			t.Errorf("expected list name 'Updated Work', got '%s'", updated.Name)
		}

		// Verify the change persisted
		lists, _ := be.GetLists(ctx)
		found := false
		for _, l := range lists {
			if l.Name == "Updated Work" {
				found = true
				break
			}
		}
		if !found {
			t.Error("updated list name not persisted")
		}
	})

	t.Run("update non-existent list returns error", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list := &backend.List{
			ID:   "non-existent-id",
			Name: "Test",
		}
		_, err = be.UpdateList(ctx, list)
		if err == nil {
			t.Error("expected error when updating non-existent list")
		}
	})
}

// =============================================================================
// Unsupported Operations Tests (TestGitBackendUnsupportedOps)
// =============================================================================

func TestGitBackendUnsupportedOps(t *testing.T) {
	t.Run("GetDeletedLists returns empty slice", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		lists, err := be.GetDeletedLists(ctx)
		if err != nil {
			t.Fatalf("GetDeletedLists error: %v", err)
		}
		if len(lists) != 0 {
			t.Errorf("expected empty slice, got %d items", len(lists))
		}
	})

	t.Run("GetDeletedListByName returns nil", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, err := be.GetDeletedListByName(ctx, "test")
		if err != nil {
			t.Fatalf("GetDeletedListByName error: %v", err)
		}
		if list != nil {
			t.Error("expected nil for deleted list")
		}
	})

	t.Run("RestoreList returns error", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		err = be.RestoreList(ctx, "some-id")
		if err == nil {
			t.Error("expected error for RestoreList")
		}
		if !strings.Contains(err.Error(), "not supported") {
			t.Errorf("expected 'not supported' error, got: %v", err)
		}
	})

	t.Run("PurgeList returns error", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		err = be.PurgeList(ctx, "some-id")
		if err == nil {
			t.Error("expected error for PurgeList")
		}
		if !strings.Contains(err.Error(), "not supported") {
			t.Errorf("expected 'not supported' error, got: %v", err)
		}
	})
}

// =============================================================================
// Error Handling Tests (TestGitBackendErrorHandling)
// =============================================================================

func TestGitBackendErrorHandling(t *testing.T) {
	t.Run("create task in non-existent list returns error", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		task := &backend.Task{
			Summary: "Test task",
			Status:  backend.StatusNeedsAction,
		}
		_, err = be.CreateTask(ctx, "non-existent-list-id", task)
		if err == nil {
			t.Error("expected error when creating task in non-existent list")
		}
	})

	t.Run("update non-existent task returns error", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")
		task := &backend.Task{
			ID:      "non-existent-task-id",
			Summary: "Test task",
			Status:  backend.StatusNeedsAction,
		}
		_, err = be.UpdateTask(ctx, list.ID, task)
		if err == nil {
			t.Error("expected error when updating non-existent task")
		}
	})

	t.Run("delete non-existent task returns error", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		list, _ := be.GetListByName(ctx, "Work")
		err = be.DeleteTask(ctx, list.ID, "non-existent-task-id")
		if err == nil {
			t.Error("expected error when deleting non-existent task")
		}
	})

	t.Run("delete non-existent list returns error", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		err = be.DeleteList(ctx, "non-existent-list-id")
		if err == nil {
			t.Error("expected error when deleting non-existent list")
		}
	})

	t.Run("operations fail when not a git repo", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create TODO.md with marker but no git repo
		todoPath := filepath.Join(tmpDir, "TODO.md")
		content := "<!-- todoat:enabled -->\n## Work\n- [ ] Task\n"
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{WorkDir: tmpDir}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()

		// GetLists should fail because no git repo
		_, err = be.GetLists(ctx)
		if err == nil {
			t.Error("expected error when getting lists without git repo")
		}
	})

	t.Run("operations fail when no todo file", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false) // No TODO.md
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()

		// GetLists should fail because no todo file found
		_, err = be.GetLists(ctx)
		if err == nil {
			t.Error("expected error when getting lists without todo file")
		}
	})
}

// =============================================================================
// GetTasks Edge Cases (TestGitBackendGetTasksEdgeCases)
// =============================================================================

func TestGitBackendGetTasksEdgeCases(t *testing.T) {
	t.Run("get tasks from non-existent list returns empty slice", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, true)
		defer cleanup()

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()
		tasks, err := be.GetTasks(ctx, "non-existent-list-id")
		if err != nil {
			t.Fatalf("GetTasks error: %v", err)
		}
		if len(tasks) != 0 {
			t.Errorf("expected empty slice for non-existent list, got %d tasks", len(tasks))
		}
	})
}

// =============================================================================
// Create List Edge Cases (TestGitBackendCreateListEdgeCases)
// =============================================================================

func TestGitBackendCreateListEdgeCases(t *testing.T) {
	t.Run("creating existing list returns existing list", func(t *testing.T) {
		repoPath, cleanup := testRepo(t, false)
		defer cleanup()

		todoPath := filepath.Join(repoPath, "TODO.md")
		content := `<!-- todoat:enabled -->
## Work

- [ ] Task 1
`
		if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create TODO.md: %v", err)
		}

		cfg := git.Config{WorkDir: repoPath}
		be, err := git.New(cfg)
		if err != nil {
			t.Fatalf("failed to create git backend: %v", err)
		}
		defer func() { _ = be.Close() }()

		ctx := context.Background()

		// Try to create a list with the same name (case-insensitive)
		list, err := be.CreateList(ctx, "work")
		if err != nil {
			t.Fatalf("CreateList error: %v", err)
		}

		// Should return the existing list
		if list.Name != "Work" {
			t.Errorf("expected existing list name 'Work', got '%s'", list.Name)
		}

		// Should not have duplicates
		lists, _ := be.GetLists(ctx)
		workCount := 0
		for _, l := range lists {
			if strings.EqualFold(l.Name, "work") {
				workCount++
			}
		}
		if workCount != 1 {
			t.Errorf("expected 1 Work list, got %d", workCount)
		}
	})
}
