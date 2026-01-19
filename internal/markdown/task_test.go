package markdown

import (
	"strings"
	"testing"
	"time"

	"todoat/backend"
)

func TestParseStatusChar(t *testing.T) {
	tests := []struct {
		name     string
		char     string
		expected backend.TaskStatus
	}{
		{"empty checkbox", " ", backend.StatusNeedsAction},
		{"completed x", "x", backend.StatusCompleted},
		{"completed X", "X", backend.StatusCompleted},
		{"in progress", "~", backend.StatusInProgress},
		{"cancelled", "-", backend.StatusCancelled},
		{"unknown", "?", backend.StatusNeedsAction},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseStatusChar(tt.char)
			if got != tt.expected {
				t.Errorf("ParseStatusChar(%q) = %q, want %q", tt.char, got, tt.expected)
			}
		})
	}
}

func TestFormatStatusChar(t *testing.T) {
	tests := []struct {
		name     string
		status   backend.TaskStatus
		expected string
	}{
		{"needs action", backend.StatusNeedsAction, " "},
		{"completed", backend.StatusCompleted, "x"},
		{"in progress", backend.StatusInProgress, "~"},
		{"cancelled", backend.StatusCancelled, "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatStatusChar(tt.status)
			if got != tt.expected {
				t.Errorf("FormatStatusChar(%q) = %q, want %q", tt.status, got, tt.expected)
			}
		})
	}
}

func TestParseTaskText(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		wantSummary    string
		wantPriority   int
		wantDueDate    string
		wantCategories string
	}{
		{
			name:        "plain text",
			text:        "Simple task",
			wantSummary: "Simple task",
		},
		{
			name:         "with priority",
			text:         "Task with priority !3",
			wantSummary:  "Task with priority",
			wantPriority: 3,
		},
		{
			name:        "with due date",
			text:        "Task with date @2024-06-15",
			wantSummary: "Task with date",
			wantDueDate: "2024-06-15",
		},
		{
			name:           "with single tag",
			text:           "Task with tag #work",
			wantSummary:    "Task with tag",
			wantCategories: "work",
		},
		{
			name:           "with multiple tags",
			text:           "Task with tags #work #urgent",
			wantSummary:    "Task with tags",
			wantCategories: "work,urgent",
		},
		{
			name:           "full metadata",
			text:           "Complete task !1 @2024-12-25 #important #project",
			wantSummary:    "Complete task",
			wantPriority:   1,
			wantDueDate:    "2024-12-25",
			wantCategories: "important,project",
		},
		{
			name:           "metadata in middle",
			text:           "Task !2 with stuff #tag in middle",
			wantSummary:    "Task  with stuff  in middle", // Note: extra spaces from regex replacement
			wantPriority:   2,
			wantCategories: "tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, priority, dueDate, categories := ParseTaskText(tt.text)

			if summary != tt.wantSummary {
				t.Errorf("summary = %q, want %q", summary, tt.wantSummary)
			}
			if priority != tt.wantPriority {
				t.Errorf("priority = %d, want %d", priority, tt.wantPriority)
			}
			if tt.wantDueDate != "" {
				if dueDate == nil {
					t.Errorf("dueDate = nil, want %q", tt.wantDueDate)
				} else if dueDate.Format("2006-01-02") != tt.wantDueDate {
					t.Errorf("dueDate = %q, want %q", dueDate.Format("2006-01-02"), tt.wantDueDate)
				}
			} else if dueDate != nil {
				t.Errorf("dueDate = %v, want nil", dueDate)
			}
			if categories != tt.wantCategories {
				t.Errorf("categories = %q, want %q", categories, tt.wantCategories)
			}
		})
	}
}

func TestFormatTaskText(t *testing.T) {
	dueDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		task     backend.Task
		expected string
	}{
		{
			name:     "plain task",
			task:     backend.Task{Summary: "Simple task"},
			expected: "Simple task",
		},
		{
			name:     "with priority",
			task:     backend.Task{Summary: "Task", Priority: 2},
			expected: "Task !2",
		},
		{
			name:     "with due date",
			task:     backend.Task{Summary: "Task", DueDate: &dueDate},
			expected: "Task @2024-06-15",
		},
		{
			name:     "with categories",
			task:     backend.Task{Summary: "Task", Categories: "work,urgent"},
			expected: "Task #work #urgent",
		},
		{
			name:     "full metadata",
			task:     backend.Task{Summary: "Task", Priority: 1, DueDate: &dueDate, Categories: "tag"},
			expected: "Task !1 @2024-06-15 #tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTaskText(&tt.task)
			if got != tt.expected {
				t.Errorf("FormatTaskText() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestOrganizeTasksHierarchically(t *testing.T) {
	tasks := []backend.Task{
		{ID: "1", Summary: "Root 1"},
		{ID: "2", Summary: "Root 2"},
		{ID: "3", Summary: "Child of 1", ParentID: "1"},
		{ID: "4", Summary: "Child of 1", ParentID: "1"},
		{ID: "5", Summary: "Child of 2", ParentID: "2"},
		{ID: "6", Summary: "Grandchild", ParentID: "3"},
	}

	rootTasks, childrenMap := OrganizeTasksHierarchically(tasks)

	// Check root tasks
	if len(rootTasks) != 2 {
		t.Errorf("expected 2 root tasks, got %d", len(rootTasks))
	}

	// Check children of task 1
	children1, ok := childrenMap["1"]
	if !ok || len(children1) != 2 {
		t.Errorf("expected 2 children of task 1, got %d", len(children1))
	}

	// Check children of task 2
	children2, ok := childrenMap["2"]
	if !ok || len(children2) != 1 {
		t.Errorf("expected 1 child of task 2, got %d", len(children2))
	}

	// Check grandchildren
	children3, ok := childrenMap["3"]
	if !ok || len(children3) != 1 {
		t.Errorf("expected 1 grandchild of task 3, got %d", len(children3))
	}
}

func TestWriteTaskTree(t *testing.T) {
	tasks := []backend.Task{
		{ID: "1", Summary: "Root task", Status: backend.StatusNeedsAction},
		{ID: "2", Summary: "Child task", ParentID: "1", Status: backend.StatusCompleted},
	}

	_, childrenMap := OrganizeTasksHierarchically(tasks)

	var sb strings.Builder
	WriteTaskTree(&sb, &tasks[0], childrenMap, 0)

	output := sb.String()

	// Verify root task is written
	if !strings.Contains(output, "- [ ] Root task") {
		t.Errorf("expected root task with empty checkbox, got: %s", output)
	}

	// Verify child task is indented and completed
	if !strings.Contains(output, "  - [x] Child task") {
		t.Errorf("expected indented child task with x checkbox, got: %s", output)
	}
}

func TestFormatStatusCharRoundTrip(t *testing.T) {
	statuses := []backend.TaskStatus{
		backend.StatusNeedsAction,
		backend.StatusCompleted,
		backend.StatusInProgress,
		backend.StatusCancelled,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			char := FormatStatusChar(status)
			parsed := ParseStatusChar(char)
			if parsed != status {
				t.Errorf("round trip failed: %q -> %q -> %q", status, char, parsed)
			}
		})
	}
}

func TestParseFormatTaskTextRoundTrip(t *testing.T) {
	dueDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	original := backend.Task{
		Summary:    "Test task",
		Priority:   3,
		DueDate:    &dueDate,
		Categories: "tag1,tag2",
	}

	// Format to text
	text := FormatTaskText(&original)

	// Parse back
	summary, priority, parsedDue, categories := ParseTaskText(text)

	if summary != original.Summary {
		t.Errorf("summary mismatch: %q != %q", summary, original.Summary)
	}
	if priority != original.Priority {
		t.Errorf("priority mismatch: %d != %d", priority, original.Priority)
	}
	if parsedDue == nil || parsedDue.Format("2006-01-02") != original.DueDate.Format("2006-01-02") {
		t.Errorf("due date mismatch")
	}
	if categories != original.Categories {
		t.Errorf("categories mismatch: %q != %q", categories, original.Categories)
	}
}
