// Package prompt handles interactive prompts with no-prompt mode support.
// It provides fuzzy-find task selection, context-aware filtering, and
// interactive add mode with field validation.
package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"todoat/backend"
	"todoat/internal/utils"
)

// Sentinel errors for prompt operations.
var (
	ErrSelectionCancelled = errors.New("selection cancelled")
	ErrNoPromptMode       = errors.New("interactive prompts disabled (--no-prompt / -y)")
	ErrNoTasks            = errors.New("no tasks available")
	ErrNoMatches          = errors.New("no tasks match the filter")
)

// TaskSelector provides fuzzy-find task selection with rich metadata display.
type TaskSelector struct {
	Tasks    []backend.Task
	Prompt   string
	Reader   io.Reader
	Writer   io.Writer
	NoPrompt bool
}

// Run executes the task selection prompt.
// If NoPrompt is true, returns ErrNoPromptMode.
// If there is exactly one task, auto-selects it.
// Otherwise, prompts the user to filter and select a task.
func (s *TaskSelector) Run() (*backend.Task, error) {
	if s.NoPrompt {
		return nil, ErrNoPromptMode
	}

	if len(s.Tasks) == 0 {
		return nil, ErrNoTasks
	}

	// Auto-select if only one task
	if len(s.Tasks) == 1 {
		return &s.Tasks[0], nil
	}

	writer := s.Writer
	if writer == nil {
		writer = io.Discard
	}

	scanner := bufio.NewScanner(s.Reader)

	// Step 1: Prompt for filter text
	_, _ = fmt.Fprintf(writer, "%s\nFilter (or press Enter to show all): ", s.Prompt)
	if !scanner.Scan() {
		return nil, ErrSelectionCancelled
	}
	filter := strings.TrimSpace(scanner.Text())

	// Step 2: Apply filter
	var filtered []backend.Task
	if filter == "" {
		filtered = s.Tasks
	} else {
		filterLower := strings.ToLower(filter)
		for _, t := range s.Tasks {
			if strings.Contains(strings.ToLower(t.Summary), filterLower) {
				filtered = append(filtered, t)
			}
		}
	}

	if len(filtered) == 0 {
		return nil, ErrNoMatches
	}

	// Auto-select if filter narrows to one
	if len(filtered) == 1 {
		_, _ = fmt.Fprintf(writer, "Auto-selected: %s\n", filtered[0].Summary)
		return &filtered[0], nil
	}

	// Step 3: Display filtered tasks with rich metadata
	for i, t := range filtered {
		_, _ = fmt.Fprintf(writer, "  %d) %s\n", i+1, formatTaskLine(t))
	}

	// Step 4: Prompt for selection number
	_, _ = fmt.Fprintf(writer, "Select (0 to cancel): ")
	if !scanner.Scan() {
		return nil, ErrSelectionCancelled
	}

	input := strings.TrimSpace(scanner.Text())
	num, err := strconv.Atoi(input)
	if err != nil {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	if num == 0 {
		return nil, ErrSelectionCancelled
	}

	if num < 1 || num > len(filtered) {
		return nil, fmt.Errorf("selection out of range: %d", num)
	}

	return &filtered[num-1], nil
}

// formatTaskLine formats a task for display with rich metadata showing
// summary, status, priority, due date, parent, and tags.
func formatTaskLine(t backend.Task) string {
	parts := []string{t.Summary}

	var meta []string
	meta = append(meta, string(t.Status))

	if t.Priority > 0 {
		meta = append(meta, fmt.Sprintf("P%d", t.Priority))
	}

	if t.DueDate != nil {
		meta = append(meta, fmt.Sprintf("due: %s", t.DueDate.Format("2006-01-02")))
	}

	if t.ParentID != "" {
		meta = append(meta, fmt.Sprintf("parent: %s", t.ParentID))
	}

	if t.Categories != "" {
		meta = append(meta, fmt.Sprintf("tags: %s", t.Categories))
	}

	if len(meta) > 0 {
		parts = append(parts, fmt.Sprintf("[%s]", strings.Join(meta, ", ")))
	}

	return strings.Join(parts, " ")
}

// FilterTasksByAction returns tasks filtered by the action being performed.
// Actions like "complete", "update", and "delete" show only actionable tasks
// (NEEDS-ACTION and IN-PROGRESS) by default.
// The showAll parameter overrides this to show all tasks including terminal states.
func FilterTasksByAction(tasks []backend.Task, action string, showAll bool) []backend.Task {
	if showAll {
		result := make([]backend.Task, len(tasks))
		copy(result, tasks)
		return result
	}

	// Actions that should only show actionable tasks
	actionableOnly := map[string]bool{
		"complete": true,
		"update":   true,
		"delete":   true,
	}

	if !actionableOnly[action] {
		result := make([]backend.Task, len(tasks))
		copy(result, tasks)
		return result
	}

	var filtered []backend.Task
	for _, t := range tasks {
		if t.Status == backend.StatusNeedsAction || t.Status == backend.StatusInProgress {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// AddFields holds the field values collected during interactive add mode.
type AddFields struct {
	Summary     string
	Description string
	Priority    int
	DueDate     string
	StartDate   string
	Tags        string
	Parent      string
	Recurrence  string
}

// InteractiveAdder provides sequential field prompts with validation
// for adding a task when no summary is provided.
type InteractiveAdder struct {
	Reader   io.Reader
	Writer   io.Writer
	NoPrompt bool
}

// Run executes the interactive add mode, prompting for each field sequentially.
// Fields: summary (required), description, priority (0-9), due date, start date,
// tags, parent, recurrence.
func (a *InteractiveAdder) Run() (*AddFields, error) {
	if a.NoPrompt {
		return nil, ErrNoPromptMode
	}

	writer := a.Writer
	if writer == nil {
		writer = io.Discard
	}

	scanner := bufio.NewScanner(a.Reader)
	fields := &AddFields{}

	// Summary (required, non-empty)
	for {
		_, _ = fmt.Fprint(writer, "Summary (required): ")
		if !scanner.Scan() {
			return nil, errors.New("no input for summary")
		}
		fields.Summary = strings.TrimSpace(scanner.Text())
		if fields.Summary != "" {
			break
		}
		_, _ = fmt.Fprintln(writer, "Summary cannot be empty.")
	}

	// Description (optional)
	_, _ = fmt.Fprint(writer, "Description (optional): ")
	if scanner.Scan() {
		fields.Description = strings.TrimSpace(scanner.Text())
	}

	// Priority (optional, 0-9)
	for {
		_, _ = fmt.Fprint(writer, "Priority (0-9, optional): ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			break
		}
		p, err := strconv.Atoi(input)
		if err != nil {
			_, _ = fmt.Fprintln(writer, "Invalid priority: must be a number 0-9")
			continue
		}
		if err := utils.ValidatePriority(p); err != nil {
			_, _ = fmt.Fprintln(writer, "Invalid priority: must be 0-9")
			continue
		}
		fields.Priority = p
		break
	}

	// Due date (optional, validated)
	for {
		_, _ = fmt.Fprint(writer, "Due date (YYYY-MM-DD, today, tomorrow, +Nd, optional): ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			break
		}
		// Validate the date format
		_, err := utils.ParseDateFlag(input)
		if err != nil {
			_, _ = fmt.Fprintf(writer, "Invalid date: %s. Use YYYY-MM-DD, today, tomorrow, +Nd, +Nw, +Nm\n", input)
			continue
		}
		fields.DueDate = input
		break
	}

	// Start date (optional, validated)
	for {
		_, _ = fmt.Fprint(writer, "Start date (YYYY-MM-DD, today, tomorrow, +Nd, optional): ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			break
		}
		_, err := utils.ParseDateFlag(input)
		if err != nil {
			_, _ = fmt.Fprintf(writer, "Invalid date: %s. Use YYYY-MM-DD, today, tomorrow, +Nd, +Nw, +Nm\n", input)
			continue
		}
		fields.StartDate = input
		break
	}

	// Tags (optional)
	_, _ = fmt.Fprint(writer, "Tags (comma-separated, optional): ")
	if scanner.Scan() {
		fields.Tags = strings.TrimSpace(scanner.Text())
	}

	// Parent (optional)
	_, _ = fmt.Fprint(writer, "Parent task (summary, optional): ")
	if scanner.Scan() {
		fields.Parent = strings.TrimSpace(scanner.Text())
	}

	// Recurrence (optional)
	_, _ = fmt.Fprint(writer, "Recurrence (e.g., daily, weekly, monthly, optional): ")
	if scanner.Scan() {
		fields.Recurrence = strings.TrimSpace(scanner.Text())
	}

	return fields, nil
}
