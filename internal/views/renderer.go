package views

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"todoat/backend"
)

// Renderer handles rendering tasks using a view configuration
type Renderer struct {
	view   *View
	writer io.Writer
}

// NewRenderer creates a new view renderer
func NewRenderer(view *View, writer io.Writer) *Renderer {
	return &Renderer{view: view, writer: writer}
}

// Render renders tasks according to the view configuration
func (r *Renderer) Render(tasks []backend.Task) {
	if len(tasks) == 0 {
		return
	}

	// Apply filters
	filteredTasks := FilterTasks(tasks, r.view.Filters)

	// Apply sorting
	sortedTasks := SortTasks(filteredTasks, r.view.Sort)

	// Render with hierarchy
	r.renderWithHierarchy(sortedTasks)
}

// renderWithHierarchy renders tasks preserving parent-child relationships
func (r *Renderer) renderWithHierarchy(tasks []backend.Task) {
	// Build task map
	taskMap := make(map[string]*backend.Task)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}

	// Build tree nodes
	nodeMap := make(map[string]*taskNode)
	var rootNodes []*taskNode

	// First pass: create nodes
	for i := range tasks {
		nodeMap[tasks[i].ID] = &taskNode{task: tasks[i]}
	}

	// Second pass: build relationships
	for i := range tasks {
		node := nodeMap[tasks[i].ID]
		if tasks[i].ParentID == "" {
			rootNodes = append(rootNodes, node)
		} else if parentNode, ok := nodeMap[tasks[i].ParentID]; ok {
			parentNode.children = append(parentNode.children, node)
		} else {
			// Orphan - show at root
			rootNodes = append(rootNodes, node)
		}
	}

	// Render tree
	for i, node := range rootNodes {
		isLast := i == len(rootNodes)-1
		r.renderNode(node, "", isLast)
	}
}

type taskNode struct {
	task     backend.Task
	children []*taskNode
}

// renderNode renders a task node with tree visualization
func (r *Renderer) renderNode(node *taskNode, prefix string, isLast bool) {
	// Build the display line
	var parts []string
	for _, field := range r.view.Fields {
		val := r.formatField(&node.task, field)
		parts = append(parts, val)
	}

	// Tree character
	var treeChar string
	if prefix == "" {
		treeChar = "  "
	} else if isLast {
		treeChar = "└─ "
	} else {
		treeChar = "├─ "
	}

	line := strings.Join(parts, " ")
	_, _ = fmt.Fprintf(r.writer, "%s%s%s\n", prefix, treeChar, line)

	// Children prefix
	var childPrefix string
	if prefix == "" {
		childPrefix = "  "
	} else if isLast {
		childPrefix = prefix + "   "
	} else {
		childPrefix = prefix + "│  "
	}

	for i, child := range node.children {
		isChildLast := i == len(node.children)-1
		r.renderNode(child, childPrefix, isChildLast)
	}
}

// formatField formats a task field according to field configuration
func (r *Renderer) formatField(t *backend.Task, field Field) string {
	var value string

	// Try plugin first if configured
	if field.Plugin != nil {
		if pluginValue, ok := runPlugin(t, field.Plugin); ok {
			value = pluginValue
		}
	}

	// Fall back to standard formatting if no plugin or plugin failed
	if value == "" {
		switch field.Name {
		case "status":
			value = formatStatus(t.Status)
		case "summary":
			value = t.Summary
		case "description":
			value = t.Description
		case "priority":
			if t.Priority > 0 {
				value = fmt.Sprintf("[P%d]", t.Priority)
			}
		case "due_date":
			value = formatDate(t.DueDate, field.Format)
		case "start_date":
			value = formatDate(t.StartDate, field.Format)
		case "created":
			value = formatDateTime(t.Created, field.Format)
		case "modified":
			value = formatDateTime(t.Modified, field.Format)
		case "completed":
			value = formatDate(t.Completed, field.Format)
		case "tags":
			if t.Categories != "" {
				value = fmt.Sprintf("{%s}", t.Categories)
			}
		case "uid":
			value = t.ID
		case "parent":
			value = t.ParentID
		}
	}

	// Apply width if specified
	if field.Width > 0 {
		if len(value) > field.Width && field.Truncate {
			value = value[:field.Width-3] + "..."
		}
		// Pad based on alignment
		switch field.Align {
		case "right":
			value = fmt.Sprintf("%*s", field.Width, value)
		case "center":
			pad := field.Width - len(value)
			leftPad := pad / 2
			rightPad := pad - leftPad
			value = strings.Repeat(" ", leftPad) + value + strings.Repeat(" ", rightPad)
		default: // left
			value = fmt.Sprintf("%-*s", field.Width, value)
		}
	}

	return value
}

// formatStatus formats a task status for display
func formatStatus(status backend.TaskStatus) string {
	switch status {
	case backend.StatusCompleted:
		return "[DONE]"
	case backend.StatusInProgress:
		return "[IN-PROGRESS]"
	case backend.StatusCancelled:
		return "[CANCELLED]"
	default:
		return "[TODO]"
	}
}

// formatDate formats a date pointer for display
func formatDate(t *time.Time, format string) string {
	if t == nil {
		return ""
	}
	if format == "" {
		format = DefaultDateFormat
	}
	return t.Format(format)
}

// formatDateTime formats a time.Time value for display
func formatDateTime(t time.Time, format string) string {
	if t.IsZero() {
		return ""
	}
	if format == "" {
		format = DefaultDateFormat
	}
	return t.Format(format)
}

// RenderTasksWithView is a convenience function for rendering tasks with a view
func RenderTasksWithView(tasks []backend.Task, view *View, writer io.Writer) {
	renderer := NewRenderer(view, writer)
	renderer.Render(tasks)
}

// pluginTaskData represents the JSON data sent to plugin stdin
type pluginTaskData struct {
	UID         string  `json:"uid"`
	Summary     string  `json:"summary"`
	Description string  `json:"description,omitempty"`
	Status      string  `json:"status"`
	Priority    int     `json:"priority"`
	DueDate     *string `json:"due_date,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	Created     string  `json:"created"`
	Modified    string  `json:"modified"`
	Completed   *string `json:"completed,omitempty"`
	Tags        string  `json:"tags,omitempty"`
	ParentID    string  `json:"parent,omitempty"`
}

// taskToPluginData converts a backend.Task to pluginTaskData
func taskToPluginData(t *backend.Task) pluginTaskData {
	data := pluginTaskData{
		UID:         t.ID,
		Summary:     t.Summary,
		Description: t.Description,
		Status:      statusToString(t.Status),
		Priority:    t.Priority,
		Created:     t.Created.Format(time.RFC3339),
		Modified:    t.Modified.Format(time.RFC3339),
		Tags:        t.Categories,
		ParentID:    t.ParentID,
	}

	if t.DueDate != nil {
		s := t.DueDate.Format(DefaultDateFormat)
		data.DueDate = &s
	}
	if t.StartDate != nil {
		s := t.StartDate.Format(DefaultDateFormat)
		data.StartDate = &s
	}
	if t.Completed != nil {
		s := t.Completed.Format(DefaultDateFormat)
		data.Completed = &s
	}

	return data
}

// statusToString converts TaskStatus to a string for JSON
func statusToString(status backend.TaskStatus) string {
	switch status {
	case backend.StatusCompleted:
		return "DONE"
	case backend.StatusInProgress:
		return "IN-PROGRESS"
	case backend.StatusCancelled:
		return "CANCELLED"
	default:
		return "TODO"
	}
}

// runPlugin executes a plugin command with task data and returns the formatted output
// Returns the formatted value or empty string if the plugin fails
func runPlugin(t *backend.Task, plugin *PluginConfig) (string, bool) {
	if plugin == nil || plugin.Command == "" {
		return "", false
	}

	// Expand ~ in command path
	command := plugin.Command
	if strings.HasPrefix(command, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			command = home + command[1:]
		}
	}

	// Check if command exists
	if _, err := os.Stat(command); os.IsNotExist(err) {
		return "", false
	}

	// Prepare task data as JSON
	taskData := taskToPluginData(t)
	jsonData, err := json.Marshal(taskData)
	if err != nil {
		return "", false
	}

	// Set timeout (default 1000ms)
	timeout := plugin.Timeout
	if timeout <= 0 {
		timeout = 1000
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, command)
	cmd.Stdin = bytes.NewReader(jsonData)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// Set environment variables
	if len(plugin.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range plugin.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", false
	}

	// Get output and trim whitespace
	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return "", false
	}

	return output, true
}
