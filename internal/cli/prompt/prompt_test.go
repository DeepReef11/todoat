package prompt

import (
	"strings"
	"testing"
	"time"

	"todoat/backend"
)

// =============================================================================
// Test Helpers
// =============================================================================

func dueDate(y int, m time.Month, d int) *time.Time {
	t := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	return &t
}

// =============================================================================
// TestFuzzyFindTaskSelection - Fuzzy-find select filters tasks by typed input
// =============================================================================

func TestFuzzyFindTaskSelection(t *testing.T) {
	tasks := []backend.Task{
		{ID: "uid-1", Summary: "Buy groceries", Status: backend.StatusNeedsAction, Priority: 5, DueDate: dueDate(2026, 2, 15)},
		{ID: "uid-2", Summary: "Fix bug in parser", Status: backend.StatusInProgress, Priority: 8},
		{ID: "uid-3", Summary: "Write documentation", Status: backend.StatusNeedsAction},
		{ID: "uid-4", Summary: "Buy milk", Status: backend.StatusNeedsAction, Priority: 3},
		{ID: "uid-5", Summary: "Deploy to production", Status: backend.StatusCompleted},
	}

	t.Run("filters by typed input", func(t *testing.T) {
		// Simulate typing "buy" then selecting first match
		input := "buy\n1\n"
		reader := strings.NewReader(input)

		selector := &TaskSelector{
			Tasks:  tasks,
			Prompt: "Select task:",
			Reader: reader,
		}

		selected, err := selector.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if selected == nil {
			t.Fatal("expected a task, got nil")
		}
		// Should match one of the "Buy" tasks
		if !strings.Contains(strings.ToLower(selected.Summary), "buy") {
			t.Errorf("expected a 'Buy' task, got %q", selected.Summary)
		}
	})

	t.Run("case insensitive filtering", func(t *testing.T) {
		input := "BUY\n1\n"
		reader := strings.NewReader(input)

		selector := &TaskSelector{
			Tasks:  tasks,
			Prompt: "Select task:",
			Reader: reader,
		}

		selected, err := selector.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if selected == nil {
			t.Fatal("expected a task, got nil")
		}
		if !strings.Contains(strings.ToLower(selected.Summary), "buy") {
			t.Errorf("expected a 'Buy' task, got %q", selected.Summary)
		}
	})

	t.Run("empty filter shows all tasks", func(t *testing.T) {
		// Empty filter, select third item
		input := "\n3\n"
		reader := strings.NewReader(input)

		selector := &TaskSelector{
			Tasks:  tasks,
			Prompt: "Select task:",
			Reader: reader,
		}

		selected, err := selector.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if selected == nil {
			t.Fatal("expected a task, got nil")
		}
	})

	t.Run("no match returns error", func(t *testing.T) {
		input := "zzzznonexistent\n"
		reader := strings.NewReader(input)

		selector := &TaskSelector{
			Tasks:  tasks,
			Prompt: "Select task:",
			Reader: reader,
		}

		selected, err := selector.Run()
		if err == nil {
			t.Fatal("expected error for no matches")
		}
		if selected != nil {
			t.Error("expected nil task for no matches")
		}
	})

	t.Run("displays rich metadata", func(t *testing.T) {
		// Select first item without filter
		input := "\n1\n"
		reader := strings.NewReader(input)
		var output strings.Builder

		selector := &TaskSelector{
			Tasks:  tasks,
			Prompt: "Select task:",
			Reader: reader,
			Writer: &output,
		}

		_, err := selector.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		out := output.String()
		// Should display priority
		if !strings.Contains(out, "P5") {
			t.Error("output should contain priority P5")
		}
		// Should display due date
		if !strings.Contains(out, "2026-02-15") {
			t.Error("output should contain due date 2026-02-15")
		}
		// Should display status
		if !strings.Contains(out, "NEEDS-ACTION") || !strings.Contains(out, "IN-PROGRESS") {
			t.Error("output should contain task statuses")
		}
	})

	t.Run("cancel selection returns error", func(t *testing.T) {
		input := "\n0\n"
		reader := strings.NewReader(input)

		selector := &TaskSelector{
			Tasks:  tasks,
			Prompt: "Select task:",
			Reader: reader,
		}

		_, err := selector.Run()
		if err == nil {
			t.Fatal("expected cancellation error")
		}
		if err != ErrSelectionCancelled {
			t.Errorf("expected ErrSelectionCancelled, got %v", err)
		}
	})
}

// =============================================================================
// TestContextAwareFiltering - Disambiguation filters by action status groups
// =============================================================================

func TestContextAwareFiltering(t *testing.T) {
	tasks := []backend.Task{
		{ID: "uid-1", Summary: "Active task 1", Status: backend.StatusNeedsAction},
		{ID: "uid-2", Summary: "Active task 2", Status: backend.StatusInProgress},
		{ID: "uid-3", Summary: "Done task", Status: backend.StatusCompleted},
		{ID: "uid-4", Summary: "Cancelled task", Status: backend.StatusCancelled},
		{ID: "uid-5", Summary: "Active task 3", Status: backend.StatusNeedsAction},
	}

	t.Run("complete action shows only actionable tasks", func(t *testing.T) {
		filtered := FilterTasksByAction(tasks, "complete", false)
		for _, task := range filtered {
			if task.Status == backend.StatusCompleted || task.Status == backend.StatusCancelled {
				t.Errorf("complete action should not show %s tasks, found %q", task.Status, task.Summary)
			}
		}
		if len(filtered) != 3 {
			t.Errorf("expected 3 actionable tasks, got %d", len(filtered))
		}
	})

	t.Run("update action shows only actionable tasks", func(t *testing.T) {
		filtered := FilterTasksByAction(tasks, "update", false)
		for _, task := range filtered {
			if task.Status == backend.StatusCompleted || task.Status == backend.StatusCancelled {
				t.Errorf("update action should not show terminal tasks, found %q (%s)", task.Summary, task.Status)
			}
		}
		if len(filtered) != 3 {
			t.Errorf("expected 3 actionable tasks, got %d", len(filtered))
		}
	})

	t.Run("delete action shows only actionable tasks", func(t *testing.T) {
		filtered := FilterTasksByAction(tasks, "delete", false)
		for _, task := range filtered {
			if task.Status == backend.StatusCompleted || task.Status == backend.StatusCancelled {
				t.Errorf("delete action should not show terminal tasks, found %q (%s)", task.Summary, task.Status)
			}
		}
		if len(filtered) != 3 {
			t.Errorf("expected 3 actionable tasks, got %d", len(filtered))
		}
	})

	t.Run("get action shows all tasks", func(t *testing.T) {
		filtered := FilterTasksByAction(tasks, "get", false)
		if len(filtered) != 5 {
			t.Errorf("get action should show all tasks, got %d", len(filtered))
		}
	})
}

// =============================================================================
// TestAllFlagShowsTerminalTasks - --all flag includes COMPLETED/CANCELLED tasks
// =============================================================================

func TestAllFlagShowsTerminalTasks(t *testing.T) {
	tasks := []backend.Task{
		{ID: "uid-1", Summary: "Active task", Status: backend.StatusNeedsAction},
		{ID: "uid-2", Summary: "In progress task", Status: backend.StatusInProgress},
		{ID: "uid-3", Summary: "Completed task", Status: backend.StatusCompleted},
		{ID: "uid-4", Summary: "Cancelled task", Status: backend.StatusCancelled},
	}

	t.Run("all flag overrides filtering for complete action", func(t *testing.T) {
		filtered := FilterTasksByAction(tasks, "complete", true)
		if len(filtered) != 4 {
			t.Errorf("--all flag should show all 4 tasks, got %d", len(filtered))
		}
	})

	t.Run("all flag overrides filtering for update action", func(t *testing.T) {
		filtered := FilterTasksByAction(tasks, "update", true)
		if len(filtered) != 4 {
			t.Errorf("--all flag should show all 4 tasks, got %d", len(filtered))
		}
	})

	t.Run("all flag overrides filtering for delete action", func(t *testing.T) {
		filtered := FilterTasksByAction(tasks, "delete", true)
		if len(filtered) != 4 {
			t.Errorf("--all flag should show all 4 tasks, got %d", len(filtered))
		}
	})

	t.Run("without all flag excludes terminal tasks", func(t *testing.T) {
		filtered := FilterTasksByAction(tasks, "complete", false)
		if len(filtered) != 2 {
			t.Errorf("without --all should show 2 actionable tasks, got %d", len(filtered))
		}
		for _, task := range filtered {
			if task.Status == backend.StatusCompleted || task.Status == backend.StatusCancelled {
				t.Errorf("should not include terminal task %q", task.Summary)
			}
		}
	})
}

// =============================================================================
// TestInteractiveAddMode - Sequential field prompts with validation
// =============================================================================

func TestInteractiveAddMode(t *testing.T) {
	t.Run("prompts for summary when empty", func(t *testing.T) {
		input := "My new task\n\n\n\n\n\n\n\n"
		reader := strings.NewReader(input)
		var output strings.Builder

		adder := &InteractiveAdder{
			Reader: reader,
			Writer: &output,
		}

		fields, err := adder.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fields.Summary != "My new task" {
			t.Errorf("expected summary 'My new task', got %q", fields.Summary)
		}
	})

	t.Run("prompts for all fields sequentially", func(t *testing.T) {
		// summary, description, priority, due date, start date, tags, parent, recurrence
		input := "Task title\nA description\n5\n2026-03-01\n2026-02-01\nwork,urgent\nParent task\nweekly\n"
		reader := strings.NewReader(input)
		var output strings.Builder

		adder := &InteractiveAdder{
			Reader: reader,
			Writer: &output,
		}

		fields, err := adder.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fields.Summary != "Task title" {
			t.Errorf("expected summary 'Task title', got %q", fields.Summary)
		}
		if fields.Description != "A description" {
			t.Errorf("expected description 'A description', got %q", fields.Description)
		}
		if fields.Priority != 5 {
			t.Errorf("expected priority 5, got %d", fields.Priority)
		}
		if fields.DueDate != "2026-03-01" {
			t.Errorf("expected due date '2026-03-01', got %q", fields.DueDate)
		}
		if fields.StartDate != "2026-02-01" {
			t.Errorf("expected start date '2026-02-01', got %q", fields.StartDate)
		}
		if fields.Tags != "work,urgent" {
			t.Errorf("expected tags 'work,urgent', got %q", fields.Tags)
		}
		if fields.Parent != "Parent task" {
			t.Errorf("expected parent 'Parent task', got %q", fields.Parent)
		}
		if fields.Recurrence != "weekly" {
			t.Errorf("expected recurrence 'weekly', got %q", fields.Recurrence)
		}
	})

	t.Run("validates priority range with re-prompt", func(t *testing.T) {
		// Invalid priority (15), then valid (7)
		input := "Task\n\n15\n7\n\n\n\n\n"
		reader := strings.NewReader(input)
		var output strings.Builder

		adder := &InteractiveAdder{
			Reader: reader,
			Writer: &output,
		}

		fields, err := adder.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fields.Priority != 7 {
			t.Errorf("expected priority 7 after re-prompt, got %d", fields.Priority)
		}
		// Check that error message was shown
		if !strings.Contains(output.String(), "0-9") || !strings.Contains(output.String(), "nvalid") {
			t.Error("should show validation error for invalid priority")
		}
	})

	t.Run("validates date format with re-prompt", func(t *testing.T) {
		// Invalid date, then valid
		input := "Task\n\n\nnot-a-date\n2026-03-01\n\n\n\n\n"
		reader := strings.NewReader(input)
		var output strings.Builder

		adder := &InteractiveAdder{
			Reader: reader,
			Writer: &output,
		}

		fields, err := adder.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fields.DueDate != "2026-03-01" {
			t.Errorf("expected due date '2026-03-01' after re-prompt, got %q", fields.DueDate)
		}
	})

	t.Run("accepts relative dates", func(t *testing.T) {
		input := "Task\n\n\ntomorrow\ntoday\n\n\n\n"
		reader := strings.NewReader(input)

		adder := &InteractiveAdder{
			Reader: reader,
		}

		fields, err := adder.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fields.DueDate != "tomorrow" {
			t.Errorf("expected due date 'tomorrow', got %q", fields.DueDate)
		}
		if fields.StartDate != "today" {
			t.Errorf("expected start date 'today', got %q", fields.StartDate)
		}
	})

	t.Run("accepts relative date offsets", func(t *testing.T) {
		input := "Task\n\n\n+7d\n+2w\n\n\n\n"
		reader := strings.NewReader(input)

		adder := &InteractiveAdder{
			Reader: reader,
		}

		fields, err := adder.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fields.DueDate != "+7d" {
			t.Errorf("expected due date '+7d', got %q", fields.DueDate)
		}
		if fields.StartDate != "+2w" {
			t.Errorf("expected start date '+2w', got %q", fields.StartDate)
		}
	})

	t.Run("requires non-empty summary", func(t *testing.T) {
		// Two empty summaries, then valid
		input := "\n\nMy task\n\n\n\n\n\n\n\n"
		reader := strings.NewReader(input)
		var output strings.Builder

		adder := &InteractiveAdder{
			Reader: reader,
			Writer: &output,
		}

		fields, err := adder.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fields.Summary != "My task" {
			t.Errorf("expected summary 'My task', got %q", fields.Summary)
		}
		// Should have re-prompted
		outStr := output.String()
		if strings.Count(outStr, "Summary") < 2 {
			t.Error("should re-prompt for empty summary")
		}
	})

	t.Run("optional fields can be skipped with empty input", func(t *testing.T) {
		// Only summary, rest empty
		input := "Just a task\n\n\n\n\n\n\n\n"
		reader := strings.NewReader(input)

		adder := &InteractiveAdder{
			Reader: reader,
		}

		fields, err := adder.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fields.Summary != "Just a task" {
			t.Errorf("expected summary 'Just a task', got %q", fields.Summary)
		}
		if fields.Description != "" {
			t.Errorf("expected empty description, got %q", fields.Description)
		}
		if fields.Priority != 0 {
			t.Errorf("expected priority 0, got %d", fields.Priority)
		}
		if fields.DueDate != "" {
			t.Errorf("expected empty due date, got %q", fields.DueDate)
		}
	})

	t.Run("validates negative priority", func(t *testing.T) {
		input := "Task\n\n-1\n3\n\n\n\n\n"
		reader := strings.NewReader(input)
		var output strings.Builder

		adder := &InteractiveAdder{
			Reader: reader,
			Writer: &output,
		}

		fields, err := adder.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fields.Priority != 3 {
			t.Errorf("expected priority 3 after re-prompt, got %d", fields.Priority)
		}
	})
}

// =============================================================================
// TestNoPromptBypass - All interactions bypass when no-prompt is set
// =============================================================================

func TestNoPromptBypass(t *testing.T) {
	tasks := []backend.Task{
		{ID: "uid-1", Summary: "Task A", Status: backend.StatusNeedsAction},
		{ID: "uid-2", Summary: "Task B", Status: backend.StatusNeedsAction},
	}

	t.Run("task selector returns error in no-prompt mode", func(t *testing.T) {
		selector := &TaskSelector{
			Tasks:    tasks,
			Prompt:   "Select task:",
			NoPrompt: true,
		}

		_, err := selector.Run()
		if err == nil {
			t.Fatal("expected error in no-prompt mode")
		}
		if err != ErrNoPromptMode {
			t.Errorf("expected ErrNoPromptMode, got %v", err)
		}
	})

	t.Run("interactive adder returns error in no-prompt mode", func(t *testing.T) {
		adder := &InteractiveAdder{
			NoPrompt: true,
		}

		_, err := adder.Run()
		if err == nil {
			t.Fatal("expected error in no-prompt mode")
		}
		if err != ErrNoPromptMode {
			t.Errorf("expected ErrNoPromptMode, got %v", err)
		}
	})

	t.Run("filter still works in no-prompt mode", func(t *testing.T) {
		// Filtering is not interactive, should work regardless
		filtered := FilterTasksByAction(tasks, "complete", false)
		if len(filtered) != 2 {
			t.Errorf("filtering should work in no-prompt mode, got %d tasks", len(filtered))
		}
	})
}

// =============================================================================
// Additional Edge Cases
// =============================================================================

func TestTaskSelectorSingleTask(t *testing.T) {
	tasks := []backend.Task{
		{ID: "uid-1", Summary: "Only task", Status: backend.StatusNeedsAction},
	}

	t.Run("single task auto-selects without prompt", func(t *testing.T) {
		selector := &TaskSelector{
			Tasks:  tasks,
			Prompt: "Select task:",
			// No reader needed - should auto-select
		}

		selected, err := selector.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if selected == nil {
			t.Fatal("expected task, got nil")
		}
		if selected.Summary != "Only task" {
			t.Errorf("expected 'Only task', got %q", selected.Summary)
		}
	})
}

func TestTaskSelectorEmptyList(t *testing.T) {
	t.Run("empty task list returns error", func(t *testing.T) {
		selector := &TaskSelector{
			Tasks:  []backend.Task{},
			Prompt: "Select task:",
		}

		_, err := selector.Run()
		if err == nil {
			t.Fatal("expected error for empty list")
		}
	})
}

func TestFilterTasksByActionWithTags(t *testing.T) {
	tasks := []backend.Task{
		{ID: "uid-1", Summary: "Tagged task", Status: backend.StatusNeedsAction, Categories: "work,urgent"},
		{ID: "uid-2", Summary: "Untagged task", Status: backend.StatusNeedsAction},
		{ID: "uid-3", Summary: "Completed tagged", Status: backend.StatusCompleted, Categories: "work"},
	}

	t.Run("filtering preserves tags", func(t *testing.T) {
		filtered := FilterTasksByAction(tasks, "complete", false)
		// Should have 2 actionable tasks (excluding completed)
		if len(filtered) != 2 {
			t.Errorf("expected 2 actionable tasks, got %d", len(filtered))
		}
		// First task should keep its tags
		if filtered[0].Categories != "work,urgent" {
			t.Errorf("expected categories 'work,urgent', got %q", filtered[0].Categories)
		}
	})
}

func TestFuzzyFindPartialMatch(t *testing.T) {
	tasks := []backend.Task{
		{ID: "uid-1", Summary: "Fix authentication bug", Status: backend.StatusNeedsAction},
		{ID: "uid-2", Summary: "Update authentication flow", Status: backend.StatusNeedsAction},
		{ID: "uid-3", Summary: "Deploy new version", Status: backend.StatusNeedsAction},
	}

	t.Run("partial match filters correctly", func(t *testing.T) {
		input := "auth\n1\n"
		reader := strings.NewReader(input)

		selector := &TaskSelector{
			Tasks:  tasks,
			Prompt: "Select task:",
			Reader: reader,
		}

		selected, err := selector.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if selected == nil {
			t.Fatal("expected task, got nil")
		}
		if !strings.Contains(strings.ToLower(selected.Summary), "auth") {
			t.Errorf("expected task with 'auth', got %q", selected.Summary)
		}
	})
}

func TestInteractiveAdderPriorityNonNumeric(t *testing.T) {
	t.Run("non-numeric priority re-prompts", func(t *testing.T) {
		input := "Task\n\nabc\n5\n\n\n\n\n"
		reader := strings.NewReader(input)
		var output strings.Builder

		adder := &InteractiveAdder{
			Reader: reader,
			Writer: &output,
		}

		fields, err := adder.Run()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fields.Priority != 5 {
			t.Errorf("expected priority 5, got %d", fields.Priority)
		}
	})
}
