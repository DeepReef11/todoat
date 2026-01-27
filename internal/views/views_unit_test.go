package views

import (
	"bytes"
	"testing"
	"time"

	"todoat/backend"
)

// Tests for filter.go helper functions

func TestMatchesRegex(t *testing.T) {
	tests := []struct {
		name        string
		fieldValue  any
		filterValue any
		want        bool
	}{
		{"simple match", "hello world", "hello", true},
		{"no match", "hello world", "^goodbye", false},
		{"regex pattern", "task-123", "task-\\d+", true},
		{"invalid regex", "test", "[invalid", false},
		{"empty pattern", "test", "", true},
		{"case sensitive", "Hello", "hello", false},
		{"case insensitive regex", "Hello", "(?i)hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesRegex(tt.fieldValue, tt.filterValue)
			if got != tt.want {
				t.Errorf("matchesRegex(%v, %v) = %v, want %v", tt.fieldValue, tt.filterValue, got, tt.want)
			}
		})
	}
}

func TestToString(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"TaskStatus", backend.StatusCompleted, "COMPLETED"},
		{"int", 42, "42"},
		{"*time.Time", &now, now.Format(DefaultDateFormat)},
		{"time.Time", now, now.Format(DefaultDateFormat)},
		{"*time.Time nil", (*time.Time)(nil), ""},
		{"other type", 3.14, "3.14"},
		{"bool", true, "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toString(tt.input)
			if got != tt.want {
				t.Errorf("toString(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		want   int
		wantOk bool
	}{
		{"int", 42, 42, true},
		{"string int", "123", 123, true},
		{"string non-int", "abc", 0, false},
		{"float", 3.14, 0, false},
		{"nil", nil, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toInt(tt.input)
			if got != tt.want || ok != tt.wantOk {
				t.Errorf("toInt(%v) = (%v, %v), want (%v, %v)", tt.input, got, ok, tt.want, tt.wantOk)
			}
		})
	}
}

func TestToTime(t *testing.T) {
	now := time.Now()
	dateStr := "2024-06-15"
	expectedDate, _ := time.Parse(DefaultDateFormat, dateStr)

	tests := []struct {
		name    string
		input   any
		wantNil bool
	}{
		{"*time.Time", &now, false},
		{"time.Time", now, false},
		{"valid date string", dateStr, false},
		{"invalid date string", "not-a-date", true},
		{"nil", nil, true},
		{"int", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toTime(tt.input)
			if tt.wantNil && got != nil {
				t.Errorf("toTime(%v) = %v, want nil", tt.input, got)
			}
			if !tt.wantNil && got == nil {
				t.Errorf("toTime(%v) = nil, want non-nil", tt.input)
			}
			if tt.name == "valid date string" && got != nil && !got.Equal(expectedDate) {
				t.Errorf("toTime(%v) = %v, want %v", tt.input, got, expectedDate)
			}
		})
	}
}

func TestParseFilterDate(t *testing.T) {
	now := time.Now().Truncate(24 * time.Hour)

	tests := []struct {
		name     string
		input    any
		wantNil  bool
		wantDate time.Time
	}{
		{"empty string", "", true, time.Time{}},
		{"today", "today", false, now},
		{"TODAY uppercase", "TODAY", false, now},
		{"tomorrow", "tomorrow", false, now.AddDate(0, 0, 1)},
		{"yesterday", "yesterday", false, now.AddDate(0, 0, -1)},
		{"+3d", "+3d", false, now.AddDate(0, 0, 3)},
		{"-2d", "-2d", false, now.AddDate(0, 0, -2)},
		{"+1w", "+1w", false, now.AddDate(0, 0, 7)},
		{"-2w", "-2w", false, now.AddDate(0, 0, -14)},
		{"+1m", "+1m", false, now.AddDate(0, 1, 0)},
		{"-3m", "-3m", false, now.AddDate(0, -3, 0)},
		{"+2D uppercase", "+2D", false, now.AddDate(0, 0, 2)},
		{"+1W uppercase", "+1W", false, now.AddDate(0, 0, 7)},
		{"+1M uppercase", "+1M", false, now.AddDate(0, 1, 0)},
		{"absolute date", "2024-06-15", false, time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)},
		{"invalid format", "not-a-date", true, time.Time{}},
		{"invalid relative", "+abc", true, time.Time{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFilterDate(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Errorf("parseFilterDate(%v) = %v, want nil", tt.input, got)
				}
				return
			}
			if got == nil {
				t.Errorf("parseFilterDate(%v) = nil, want %v", tt.input, tt.wantDate)
				return
			}
			gotTrunc := got.Truncate(24 * time.Hour)
			wantTrunc := tt.wantDate.Truncate(24 * time.Hour)
			if !gotTrunc.Equal(wantTrunc) {
				t.Errorf("parseFilterDate(%v) = %v, want %v", tt.input, gotTrunc, wantTrunc)
			}
		})
	}
}

func TestNormalizeStatus(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"DONE", "COMPLETED"},
		{"done", "COMPLETED"},
		{"COMPLETED", "COMPLETED"},
		{"completed", "COMPLETED"},
		{"TODO", "NEEDS-ACTION"},
		{"todo", "NEEDS-ACTION"},
		{"NEEDS-ACTION", "NEEDS-ACTION"},
		{"IN-PROGRESS", "IN-PROCESS"},
		{"in-progress", "IN-PROCESS"},
		{"INPROGRESS", "IN-PROCESS"},
		{"IN_PROGRESS", "IN-PROCESS"},
		{"IN-PROCESS", "IN-PROCESS"},
		{"CANCELLED", "CANCELLED"},
		{"CANCELED", "CANCELLED"},
		{"cancelled", "CANCELLED"},
		{"  DONE  ", "COMPLETED"},
		{"UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeStatus(tt.input)
			if got != tt.want {
				t.Errorf("normalizeStatus(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetFieldValue(t *testing.T) {
	now := time.Now()
	dueDate := now.Add(24 * time.Hour)
	startDate := now.Add(-24 * time.Hour)
	completed := now.Add(-1 * time.Hour)

	task := &backend.Task{
		ID:          "test-123",
		Summary:     "Test summary",
		Description: "Test description",
		Status:      backend.StatusInProgress,
		Priority:    2,
		DueDate:     &dueDate,
		StartDate:   &startDate,
		Created:     now,
		Modified:    now,
		Completed:   &completed,
		Categories:  "work,urgent",
		ParentID:    "parent-456",
	}

	tests := []struct {
		field string
		want  any
	}{
		{"status", "IN-PROGRESS"},
		{"summary", "Test summary"},
		{"description", "Test description"},
		{"priority", 2},
		{"due_date", &dueDate},
		{"start_date", &startDate},
		{"created", now},
		{"modified", now},
		{"completed", &completed},
		{"tags", "work,urgent"},
		{"uid", "test-123"},
		{"parent", "parent-456"},
		{"unknown", nil},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got := getFieldValue(task, tt.field)
			switch want := tt.want.(type) {
			case *time.Time:
				gotTime, ok := got.(*time.Time)
				if !ok || !gotTime.Equal(*want) {
					t.Errorf("getFieldValue(%q) = %v, want %v", tt.field, got, want)
				}
			case time.Time:
				gotTime, ok := got.(time.Time)
				if !ok || !gotTime.Equal(want) {
					t.Errorf("getFieldValue(%q) = %v, want %v", tt.field, got, want)
				}
			default:
				if got != want {
					t.Errorf("getFieldValue(%q) = %v, want %v", tt.field, got, want)
				}
			}
		})
	}
}

func TestCompareValue(t *testing.T) {
	tests := []struct {
		name        string
		fieldValue  any
		operator    string
		filterValue any
		fieldName   string
		want        bool
	}{
		// eq operator
		{"eq strings match", "hello", "eq", "hello", "summary", true},
		{"eq strings no match", "hello", "eq", "world", "summary", false},
		{"eq status normalized", "DONE", "eq", "completed", "status", true},
		// ne operator
		{"ne strings different", "hello", "ne", "world", "summary", true},
		{"ne strings same", "hello", "ne", "hello", "summary", false},
		// contains operator
		{"contains found", "hello world", "contains", "world", "summary", true},
		{"contains not found", "hello", "contains", "world", "summary", false},
		// regex operator
		{"regex match", "task-123", "regex", "task-\\d+", "summary", true},
		{"regex no match", "hello", "regex", "^world", "summary", false},
		// in operator
		{"in found", "active", "in", []string{"active", "pending"}, "status", true},
		{"in not found", "done", "in", []string{"active", "pending"}, "status", false},
		// not_in operator
		{"not_in excluded", "done", "not_in", []string{"active", "pending"}, "status", true},
		{"not_in included", "active", "not_in", []string{"active", "pending"}, "status", false},
		// lt, lte, gt, gte with numbers
		{"lt number true", 5, "lt", 10, "priority", true},
		{"lt number false", 15, "lt", 10, "priority", false},
		{"lte equal", 10, "lte", 10, "priority", true},
		{"gt number true", 15, "gt", 10, "priority", true},
		{"gte equal", 10, "gte", 10, "priority", true},
		// unknown operator
		{"unknown operator", "test", "invalid", "test", "summary", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareValue(tt.fieldValue, tt.operator, tt.filterValue, tt.fieldName)
			if got != tt.want {
				t.Errorf("compareValue(%v, %q, %v, %q) = %v, want %v",
					tt.fieldValue, tt.operator, tt.filterValue, tt.fieldName, got, tt.want)
			}
		})
	}
}

func TestEquals(t *testing.T) {
	// Use UTC dates for consistent comparison
	date1 := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2024, 6, 16, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		fieldValue  any
		filterValue any
		fieldName   string
		want        bool
	}{
		{"both nil", nil, nil, "summary", true},
		{"status match normalized", "DONE", "completed", "status", true},
		{"status mismatch", "DONE", "pending", "status", false},
		// For date fields, filterValue is typically a string (from YAML config)
		{"date match with string filter", &date1, "2024-06-15", "due_date", true},
		{"date mismatch with string filter", &date1, "2024-06-16", "due_date", false},
		{"date nil both", (*time.Time)(nil), nil, "due_date", true},
		{"date field nil vs string", (*time.Time)(nil), "2024-06-15", "due_date", false},
		{"date field vs nil filter", &date1, nil, "due_date", false},
		{"string match", "test", "test", "summary", true},
		{"string mismatch", "test", "other", "summary", false},
		// Date pointer comparison (both as *time.Time - converted via toString -> parseFilterDate)
		{"date match pointers same day", &date1, &date1, "due_date", true},
		{"date mismatch pointers different days", &date1, &date2, "due_date", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := equals(tt.fieldValue, tt.filterValue, tt.fieldName)
			if got != tt.want {
				t.Errorf("equals(%v, %v, %q) = %v, want %v",
					tt.fieldValue, tt.filterValue, tt.fieldName, got, tt.want)
			}
		})
	}
}

func TestInList(t *testing.T) {
	tests := []struct {
		name        string
		fieldValue  any
		filterValue any
		want        bool
	}{
		{"in []any found", "active", []any{"active", "pending"}, true},
		{"in []any not found", "done", []any{"active", "pending"}, false},
		{"in []string found", "active", []string{"active", "pending"}, true},
		{"in []string not found", "done", []string{"active", "pending"}, false},
		{"status normalized in []any", "DONE", []any{"completed", "cancelled"}, true},
		{"status normalized in []string", "DONE", []string{"completed", "cancelled"}, true},
		{"case insensitive in []any", "ACTIVE", []any{"active", "pending"}, true},
		{"case insensitive in []string", "ACTIVE", []string{"active", "pending"}, true},
		{"empty list", "active", []any{}, false},
		{"nil value", nil, []any{"active"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inList(tt.fieldValue, tt.filterValue)
			if got != tt.want {
				t.Errorf("inList(%v, %v) = %v, want %v", tt.fieldValue, tt.filterValue, got, tt.want)
			}
		})
	}
}

func TestCompareForSort(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)

	tests := []struct {
		name      string
		a         any
		b         any
		fieldName string
		want      int
	}{
		{"both nil", nil, nil, "summary", 0},
		{"a nil", nil, "test", "summary", 1},
		{"b nil", "test", nil, "summary", -1},
		{"dates a before b", &now, &later, "due_date", -1},
		{"dates a after b", &later, &now, "due_date", 1},
		{"dates equal", &now, &now, "due_date", 0},
		{"date a nil", (*time.Time)(nil), &now, "due_date", 1},
		{"date b nil", &now, (*time.Time)(nil), "due_date", -1},
		{"ints a < b", 1, 5, "priority", -1},
		{"ints a > b", 5, 1, "priority", 1},
		{"ints equal", 3, 3, "priority", 0},
		{"strings a < b", "apple", "banana", "summary", -1},
		{"strings a > b", "banana", "apple", "summary", 1},
		{"strings equal", "test", "test", "summary", 0},
		{"strings case insensitive", "Apple", "banana", "summary", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareForSort(tt.a, tt.b, tt.fieldName)
			if got != tt.want {
				t.Errorf("compareForSort(%v, %v, %q) = %v, want %v",
					tt.a, tt.b, tt.fieldName, got, tt.want)
			}
		})
	}
}

// Tests for types.go

func TestDefaultView(t *testing.T) {
	view := DefaultView()

	if view == nil {
		t.Fatal("DefaultView() returned nil")
	}

	if view.Name != "default" {
		t.Errorf("DefaultView().Name = %q, want %q", view.Name, "default")
	}

	if view.Description == "" {
		t.Error("DefaultView().Description is empty")
	}

	if len(view.Fields) == 0 {
		t.Error("DefaultView().Fields is empty")
	}

	// Check expected fields
	expectedFields := []string{"status", "summary", "priority"}
	for i, expected := range expectedFields {
		if i >= len(view.Fields) {
			t.Errorf("DefaultView missing field %q at index %d", expected, i)
			continue
		}
		if view.Fields[i].Name != expected {
			t.Errorf("DefaultView().Fields[%d].Name = %q, want %q", i, view.Fields[i].Name, expected)
		}
	}
}

func TestAllView(t *testing.T) {
	view := AllView()

	if view == nil {
		t.Fatal("AllView() returned nil")
	}

	if view.Name != "all" {
		t.Errorf("AllView().Name = %q, want %q", view.Name, "all")
	}

	// Should have all available fields
	if len(view.Fields) != len(AvailableFields) {
		t.Errorf("AllView().Fields has %d fields, want %d", len(view.Fields), len(AvailableFields))
	}
}

// Tests for renderer.go

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		status backend.TaskStatus
		want   string
	}{
		{backend.StatusCompleted, "[DONE]"},
		{backend.StatusInProgress, "[IN-PROGRESS]"},
		{backend.StatusCancelled, "[CANCELLED]"},
		{backend.StatusNeedsAction, "[TODO]"},
		{"UNKNOWN", "[TODO]"}, // default
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			got := formatStatus(tt.status)
			if got != tt.want {
				t.Errorf("formatStatus(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestStatusToString(t *testing.T) {
	tests := []struct {
		status backend.TaskStatus
		want   string
	}{
		{backend.StatusCompleted, "DONE"},
		{backend.StatusInProgress, "IN-PROGRESS"},
		{backend.StatusCancelled, "CANCELLED"},
		{backend.StatusNeedsAction, "TODO"},
		{"UNKNOWN", "TODO"}, // default
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			got := StatusToString(tt.status)
			if got != tt.want {
				t.Errorf("StatusToString(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	date := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	dateWithTime := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name   string
		t      *time.Time
		format string
		want   string
	}{
		{"nil", nil, "", ""},
		{"date only default format", &date, "", "Jun 15"},
		{"date with time default format", &dateWithTime, "", "Jun 15 14:30"},
		{"custom format", &date, "02/01/2006", "15/06/2024"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDate(tt.t, tt.format)
			if got != tt.want {
				t.Errorf("formatDate(%v, %q) = %q, want %q", tt.t, tt.format, got, tt.want)
			}
		})
	}
}

func TestFormatDateTime(t *testing.T) {
	datetime := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	zero := time.Time{}

	tests := []struct {
		name   string
		t      time.Time
		format string
		want   string
	}{
		{"zero time", zero, "", ""},
		{"default format", datetime, "", "2024-06-15"},
		{"custom format", datetime, "02/01/2006 15:04", "15/06/2024 14:30"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDateTime(tt.t, tt.format)
			if got != tt.want {
				t.Errorf("formatDateTime(%v, %q) = %q, want %q", tt.t, tt.format, got, tt.want)
			}
		})
	}
}

func TestFormatDateForJSON(t *testing.T) {
	date := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	dateWithTime := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name string
		t    *time.Time
		want string
	}{
		{"nil", nil, ""},
		{"date only", &date, "2024-06-15"},
		{"date with time", &dateWithTime, "2024-06-15T14:30:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDateForJSON(tt.t)
			if got != tt.want {
				t.Errorf("formatDateForJSON(%v) = %q, want %q", tt.t, got, tt.want)
			}
		})
	}
}

func TestHasTimeComponent(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		want bool
	}{
		{"midnight", time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC), false},
		{"with hour", time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC), true},
		{"with minute", time.Date(2024, 6, 15, 0, 30, 0, 0, time.UTC), true},
		{"with second", time.Date(2024, 6, 15, 0, 0, 45, 0, time.UTC), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasTimeComponent(tt.t)
			if got != tt.want {
				t.Errorf("hasTimeComponent(%v) = %v, want %v", tt.t, got, tt.want)
			}
		})
	}
}

// Tests for loader.go

func TestIsValidOperator(t *testing.T) {
	validOps := []string{"eq", "ne", "lt", "lte", "gt", "gte", "contains", "in", "not_in", "regex"}
	invalidOps := []string{"invalid", "like", "between", ""}

	for _, op := range validOps {
		t.Run("valid_"+op, func(t *testing.T) {
			if !isValidOperator(op) {
				t.Errorf("isValidOperator(%q) = false, want true", op)
			}
		})
	}

	for _, op := range invalidOps {
		t.Run("invalid_"+op, func(t *testing.T) {
			if isValidOperator(op) {
				t.Errorf("isValidOperator(%q) = true, want false", op)
			}
		})
	}
}

func TestViewExists(t *testing.T) {
	loader := NewLoader("")

	// Built-in views should exist
	builtIns := []string{"default", "", "all"}
	for _, name := range builtIns {
		t.Run("builtin_"+name, func(t *testing.T) {
			if !loader.ViewExists(name) {
				t.Errorf("ViewExists(%q) = false, want true for built-in", name)
			}
		})
	}

	// Non-existent view
	if loader.ViewExists("nonexistent-view-xyz") {
		t.Error("ViewExists(\"nonexistent-view-xyz\") = true, want false")
	}
}

func TestRendererEmptyTasks(t *testing.T) {
	var buf bytes.Buffer
	view := DefaultView()
	renderer := NewRenderer(view, &buf)

	renderer.Render([]backend.Task{})

	if buf.Len() != 0 {
		t.Errorf("Render of empty tasks produced output: %q", buf.String())
	}
}

func TestTaskToPluginData(t *testing.T) {
	now := time.Now()
	dueDate := now.Add(24 * time.Hour)
	startDate := now.Add(-24 * time.Hour)
	completed := now.Add(-1 * time.Hour)

	task := &backend.Task{
		ID:          "test-123",
		Summary:     "Test summary",
		Description: "Test description",
		Status:      backend.StatusCompleted,
		Priority:    2,
		DueDate:     &dueDate,
		StartDate:   &startDate,
		Created:     now,
		Modified:    now,
		Completed:   &completed,
		Categories:  "work,urgent",
		ParentID:    "parent-456",
	}

	data := taskToPluginData(task)

	if data.UID != "test-123" {
		t.Errorf("taskToPluginData.UID = %q, want %q", data.UID, "test-123")
	}
	if data.Summary != "Test summary" {
		t.Errorf("taskToPluginData.Summary = %q, want %q", data.Summary, "Test summary")
	}
	if data.Status != "DONE" {
		t.Errorf("taskToPluginData.Status = %q, want %q", data.Status, "DONE")
	}
	if data.Priority != 2 {
		t.Errorf("taskToPluginData.Priority = %d, want %d", data.Priority, 2)
	}
	if data.DueDate == nil {
		t.Error("taskToPluginData.DueDate is nil, want non-nil")
	}
	if data.StartDate == nil {
		t.Error("taskToPluginData.StartDate is nil, want non-nil")
	}
	if data.Completed == nil {
		t.Error("taskToPluginData.Completed is nil, want non-nil")
	}
	if data.Tags != "work,urgent" {
		t.Errorf("taskToPluginData.Tags = %q, want %q", data.Tags, "work,urgent")
	}
	if data.ParentID != "parent-456" {
		t.Errorf("taskToPluginData.ParentID = %q, want %q", data.ParentID, "parent-456")
	}
}

func TestRunPluginNilConfig(t *testing.T) {
	task := &backend.Task{ID: "test"}

	// Nil plugin
	result, ok := runPlugin(task, nil)
	if ok || result != "" {
		t.Errorf("runPlugin(task, nil) = (%q, %v), want (\"\", false)", result, ok)
	}

	// Empty command
	result, ok = runPlugin(task, &PluginConfig{Command: ""})
	if ok || result != "" {
		t.Errorf("runPlugin with empty command = (%q, %v), want (\"\", false)", result, ok)
	}

	// Non-existent command
	result, ok = runPlugin(task, &PluginConfig{Command: "/nonexistent/command/xyz123"})
	if ok || result != "" {
		t.Errorf("runPlugin with nonexistent command = (%q, %v), want (\"\", false)", result, ok)
	}
}
