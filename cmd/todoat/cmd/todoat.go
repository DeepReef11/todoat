package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"todoat/backend"
	"todoat/backend/sqlite"
)

// Version is set at build time
var Version = "dev"

// Config holds application configuration
type Config struct {
	NoPrompt     bool
	Verbose      bool
	OutputFormat string
	DBPath       string // Path to database file (for testing)
}

// Execute runs the CLI with the given arguments and IO writers
func Execute(args []string, stdout, stderr io.Writer, cfg *Config) int {
	rootCmd := NewTodoAt(stdout, stderr, cfg)

	rootCmd.SetArgs(args)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(stderr, "Error:", err)
		return 1
	}
	return 0
}

// NewTodoAt creates the root command with injectable IO
func NewTodoAt(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	if cfg == nil {
		cfg = &Config{}
	}

	cmd := &cobra.Command{
		Use:     "todoat [list] [action] [task]",
		Short:   "A task management CLI",
		Long:    "todoat is a command-line task manager supporting multiple backends.",
		Version: Version,
		Args:    cobra.MaximumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Update config from flags
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			// If no args, show help
			if len(args) == 0 {
				return cmd.Help()
			}

			// Get or create backend
			be, err := getBackend(cfg)
			if err != nil {
				return err
			}
			defer func() { _ = be.Close() }()

			ctx := context.Background()
			listName := args[0]

			// Determine action - default is "get" if only list name provided
			action := "get"
			var taskSummary string

			if len(args) >= 2 {
				action = resolveAction(args[1])
				if action == "" {
					// If args[1] is not a known action, treat first arg as list name
					// and this as an unknown action
					return fmt.Errorf("unknown action: %s", args[1])
				}
			}

			if len(args) >= 3 {
				taskSummary = args[2]
			}

			// Get or create list
			list, err := getOrCreateList(ctx, be, listName)
			if err != nil {
				return err
			}

			// Execute the action
			return executeAction(ctx, cmd, be, list, action, taskSummary, cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add global flags
	cmd.PersistentFlags().BoolP("no-prompt", "y", false, "Disable interactive prompts")
	cmd.PersistentFlags().BoolP("verbose", "V", false, "Enable verbose/debug output")
	cmd.PersistentFlags().Bool("json", false, "Output in JSON format")

	// Add action-specific flags
	cmd.Flags().IntP("priority", "p", 0, "Task priority (0-9)")
	cmd.Flags().StringP("status", "s", "", "Task status (TODO, IN-PROGRESS, DONE, CANCELLED)")
	cmd.Flags().String("summary", "", "New task summary (for update)")

	return cmd
}

// getBackend creates or returns the backend connection
func getBackend(cfg *Config) (backend.TaskManager, error) {
	dbPath := cfg.DBPath
	if dbPath == "" {
		// Use default path in user's home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not find home directory: %w", err)
		}
		dbPath = filepath.Join(home, ".todoat", "todoat.db")

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			return nil, fmt.Errorf("could not create data directory: %w", err)
		}
	}

	return sqlite.New(dbPath)
}

// resolveAction maps action names and abbreviations to canonical action names
func resolveAction(s string) string {
	switch strings.ToLower(s) {
	case "get", "g":
		return "get"
	case "add", "a":
		return "add"
	case "update", "u":
		return "update"
	case "complete", "c":
		return "complete"
	case "delete", "d":
		return "delete"
	default:
		return ""
	}
}

// getOrCreateList finds a list by name or creates it
func getOrCreateList(ctx context.Context, be backend.TaskManager, name string) (*backend.List, error) {
	lists, err := be.GetLists(ctx)
	if err != nil {
		return nil, err
	}

	// Find existing list by name (case-insensitive)
	for _, l := range lists {
		if strings.EqualFold(l.Name, name) {
			return &l, nil
		}
	}

	// Create new list
	return be.CreateList(ctx, name)
}

// executeAction performs the requested action on the list
func executeAction(ctx context.Context, cmd *cobra.Command, be backend.TaskManager, list *backend.List, action, taskSummary string, cfg *Config, stdout io.Writer) error {
	switch action {
	case "get":
		statusFilter, _ := cmd.Flags().GetString("status")
		return doGet(ctx, be, list, statusFilter, stdout)
	case "add":
		priority, _ := cmd.Flags().GetInt("priority")
		return doAdd(ctx, be, list, taskSummary, priority, stdout)
	case "update":
		priority, _ := cmd.Flags().GetInt("priority")
		status, _ := cmd.Flags().GetString("status")
		newSummary, _ := cmd.Flags().GetString("summary")
		return doUpdate(ctx, be, list, taskSummary, newSummary, status, priority, cfg, stdout)
	case "complete":
		return doComplete(ctx, be, list, taskSummary, cfg, stdout)
	case "delete":
		return doDelete(ctx, be, list, taskSummary, cfg, stdout)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// doGet lists all tasks in a list, optionally filtering by status
func doGet(ctx context.Context, be backend.TaskManager, list *backend.List, statusFilter string, stdout io.Writer) error {
	tasks, err := be.GetTasks(ctx, list.ID)
	if err != nil {
		return err
	}

	// Filter by status if specified
	if statusFilter != "" {
		filterStatus := parseStatus(statusFilter)
		var filteredTasks []backend.Task
		for _, t := range tasks {
			if t.Status == filterStatus {
				filteredTasks = append(filteredTasks, t)
			}
		}
		tasks = filteredTasks
	}

	if len(tasks) == 0 {
		_, _ = fmt.Fprintf(stdout, "No tasks in list '%s'\n", list.Name)
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "Tasks in '%s':\n", list.Name)
	for _, t := range tasks {
		statusIcon := getStatusIcon(t.Status)
		priorityStr := ""
		if t.Priority > 0 {
			priorityStr = fmt.Sprintf(" [P%d]", t.Priority)
		}
		_, _ = fmt.Fprintf(stdout, "  %s %s%s\n", statusIcon, t.Summary, priorityStr)
	}
	return nil
}

// getStatusIcon returns a visual indicator for task status
func getStatusIcon(status backend.TaskStatus) string {
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

// doAdd creates a new task
func doAdd(ctx context.Context, be backend.TaskManager, list *backend.List, summary string, priority int, stdout io.Writer) error {
	if summary == "" {
		return fmt.Errorf("task summary is required")
	}

	task := &backend.Task{
		Summary:  summary,
		Priority: priority,
		Status:   backend.StatusNeedsAction,
	}

	created, err := be.CreateTask(ctx, list.ID, task)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "Created task: %s (ID: %s)\n", created.Summary, created.ID)
	return nil
}

// doUpdate modifies an existing task
func doUpdate(ctx context.Context, be backend.TaskManager, list *backend.List, taskSummary, newSummary, status string, priority int, cfg *Config, stdout io.Writer) error {
	task, err := findTask(ctx, be, list, taskSummary, cfg)
	if err != nil {
		return err
	}

	// Apply updates
	if newSummary != "" {
		task.Summary = newSummary
	}
	if status != "" {
		task.Status = parseStatus(status)
	}
	if priority > 0 {
		task.Priority = priority
	}

	updated, err := be.UpdateTask(ctx, list.ID, task)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "Updated task: %s\n", updated.Summary)
	return nil
}

// parseStatus converts a status string to TaskStatus
func parseStatus(s string) backend.TaskStatus {
	switch strings.ToUpper(s) {
	case "DONE", "COMPLETED", "D":
		return backend.StatusCompleted
	case "IN-PROGRESS", "INPROGRESS", "PROGRESS":
		return backend.StatusInProgress
	case "CANCELLED", "CANCELED":
		return backend.StatusCancelled
	case "TODO", "NEEDS-ACTION", "T":
		return backend.StatusNeedsAction
	default:
		return backend.StatusNeedsAction
	}
}

// doComplete marks a task as completed
func doComplete(ctx context.Context, be backend.TaskManager, list *backend.List, taskSummary string, cfg *Config, stdout io.Writer) error {
	task, err := findTask(ctx, be, list, taskSummary, cfg)
	if err != nil {
		return err
	}

	task.Status = backend.StatusCompleted

	updated, err := be.UpdateTask(ctx, list.ID, task)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "Completed task: %s\n", updated.Summary)
	return nil
}

// doDelete removes a task
func doDelete(ctx context.Context, be backend.TaskManager, list *backend.List, taskSummary string, cfg *Config, stdout io.Writer) error {
	task, err := findTask(ctx, be, list, taskSummary, cfg)
	if err != nil {
		return err
	}

	if err := be.DeleteTask(ctx, list.ID, task.ID); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "Deleted task: %s\n", task.Summary)
	return nil
}

// findTask searches for a task by summary using exact then partial matching
func findTask(ctx context.Context, be backend.TaskManager, list *backend.List, searchTerm string, cfg *Config) (*backend.Task, error) {
	if searchTerm == "" {
		return nil, fmt.Errorf("task summary is required")
	}

	tasks, err := be.GetTasks(ctx, list.ID)
	if err != nil {
		return nil, err
	}

	searchLower := strings.ToLower(searchTerm)

	// First try exact match (case-insensitive)
	for _, t := range tasks {
		if strings.EqualFold(t.Summary, searchTerm) {
			return &t, nil
		}
	}

	// Then try partial match (case-insensitive)
	var matches []backend.Task
	for _, t := range tasks {
		if strings.Contains(strings.ToLower(t.Summary), searchLower) {
			matches = append(matches, t)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no task found matching '%s'", searchTerm)
	}

	if len(matches) == 1 {
		return &matches[0], nil
	}

	// Multiple matches - error in no-prompt mode
	if cfg != nil && cfg.NoPrompt {
		var matchNames []string
		for _, m := range matches {
			matchNames = append(matchNames, fmt.Sprintf("  - %s", m.Summary))
		}
		return nil, fmt.Errorf("multiple tasks match '%s':\n%s", searchTerm, strings.Join(matchNames, "\n"))
	}

	// In interactive mode, we would prompt - but for now return error
	return nil, fmt.Errorf("multiple tasks match '%s' - please be more specific", searchTerm)
}
