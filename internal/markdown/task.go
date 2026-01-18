// Package markdown provides shared utilities for parsing and formatting
// markdown-based task files used by the file and git backends.
package markdown

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"todoat/backend"
)

// OrganizeTasksHierarchically separates root tasks from children.
// Returns root tasks (those without ParentID) and a map of parentID -> children.
func OrganizeTasksHierarchically(tasks []backend.Task) ([]backend.Task, map[string][]backend.Task) {
	childrenMap := make(map[string][]backend.Task)
	var rootTasks []backend.Task

	for _, task := range tasks {
		if task.ParentID == "" {
			rootTasks = append(rootTasks, task)
		} else {
			childrenMap[task.ParentID] = append(childrenMap[task.ParentID], task)
		}
	}

	return rootTasks, childrenMap
}

// WriteTaskTree writes a task and its children with proper indentation to a strings.Builder.
func WriteTaskTree(sb *strings.Builder, task *backend.Task, childrenMap map[string][]backend.Task, level int) {
	indent := strings.Repeat("  ", level)

	// Format status character
	statusChar := FormatStatusChar(task.Status)

	// Format task line
	sb.WriteString(indent)
	sb.WriteString("- [")
	sb.WriteString(statusChar)
	sb.WriteString("] ")
	sb.WriteString(FormatTaskText(task))
	sb.WriteString("\n")

	// Write children
	if children, ok := childrenMap[task.ID]; ok {
		for i := range children {
			WriteTaskTree(sb, &children[i], childrenMap, level+1)
		}
	}
}

// ParseStatusChar converts a markdown checkbox character to TaskStatus.
func ParseStatusChar(char string) backend.TaskStatus {
	switch strings.ToLower(char) {
	case "x":
		return backend.StatusCompleted
	case "~":
		return backend.StatusInProgress
	case "-":
		return backend.StatusCancelled
	default:
		return backend.StatusNeedsAction
	}
}

// FormatStatusChar converts TaskStatus to markdown checkbox character.
func FormatStatusChar(status backend.TaskStatus) string {
	switch status {
	case backend.StatusCompleted:
		return "x"
	case backend.StatusInProgress:
		return "~"
	case backend.StatusCancelled:
		return "-"
	default:
		return " "
	}
}

// ParseTaskText extracts summary, priority, due date, and categories from task text.
// Format: "Task summary !1 @2024-01-15 #tag1 #tag2"
func ParseTaskText(text string) (summary string, priority int, dueDate *time.Time, categories string) {
	summary = text

	// Extract priority: !1, !2, etc.
	priorityPattern := regexp.MustCompile(`!(\d)`)
	if matches := priorityPattern.FindStringSubmatch(text); len(matches) == 2 {
		_, _ = fmt.Sscanf(matches[1], "%d", &priority)
		summary = strings.TrimSpace(priorityPattern.ReplaceAllString(summary, ""))
	}

	// Extract due date: @2024-01-15
	dueDatePattern := regexp.MustCompile(`@(\d{4}-\d{2}-\d{2})`)
	if matches := dueDatePattern.FindStringSubmatch(text); len(matches) == 2 {
		if t, err := time.Parse("2006-01-02", matches[1]); err == nil {
			dueDate = &t
		}
		summary = strings.TrimSpace(dueDatePattern.ReplaceAllString(summary, ""))
	}

	// Extract categories/tags: #tag1 #tag2
	tagPattern := regexp.MustCompile(`#(\w+)`)
	var tags []string
	for _, match := range tagPattern.FindAllStringSubmatch(text, -1) {
		if len(match) == 2 {
			tags = append(tags, match[1])
		}
	}
	if len(tags) > 0 {
		categories = strings.Join(tags, ",")
		summary = strings.TrimSpace(tagPattern.ReplaceAllString(summary, ""))
	}

	return summary, priority, dueDate, categories
}

// FormatTaskText formats a task back to markdown text.
func FormatTaskText(task *backend.Task) string {
	parts := []string{task.Summary}

	if task.Priority > 0 {
		parts = append(parts, fmt.Sprintf("!%d", task.Priority))
	}

	if task.DueDate != nil {
		parts = append(parts, "@"+task.DueDate.Format("2006-01-02"))
	}

	if task.Categories != "" {
		for _, tag := range strings.Split(task.Categories, ",") {
			parts = append(parts, "#"+strings.TrimSpace(tag))
		}
	}

	return strings.Join(parts, " ")
}
