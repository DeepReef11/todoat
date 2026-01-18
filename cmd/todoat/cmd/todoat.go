package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"todoat/backend"
	"todoat/backend/sqlite"
	"todoat/internal/credentials"
	"todoat/internal/views"
)

// Version is set at build time
var Version = "dev"

// Result codes for CLI output (used in no-prompt mode)
const (
	ResultActionCompleted = "ACTION_COMPLETED"
	ResultInfoOnly        = "INFO_ONLY"
	ResultError           = "ERROR"
)

// Config holds application configuration
type Config struct {
	NoPrompt     bool
	Verbose      bool
	OutputFormat string
	DBPath       string // Path to database file (for testing)
	ViewsPath    string // Path to views directory (for testing)
}

// Execute runs the CLI with the given arguments and IO writers
func Execute(args []string, stdout, stderr io.Writer, cfg *Config) int {
	rootCmd := NewTodoAt(stdout, stderr, cfg)

	rootCmd.SetArgs(args)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	if err := rootCmd.Execute(); err != nil {
		// Check if --json flag was passed to output error as JSON
		jsonOutput := containsJSONFlag(args)
		if jsonOutput {
			outputErrorJSON(err, stdout)
		} else {
			_, _ = fmt.Fprintln(stderr, "Error:", err)
			// Emit ERROR result code in no-prompt mode
			if cfg != nil && cfg.NoPrompt {
				_, _ = fmt.Fprintln(stdout, ResultError)
			}
		}
		return 1
	}
	return 0
}

// containsJSONFlag checks if args contain --json flag
func containsJSONFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--json" {
			return true
		}
	}
	return false
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

			// Check for JSON output mode
			jsonOutput, _ := cmd.Flags().GetBool("json")

			// Execute the action
			return executeAction(ctx, cmd, be, list, action, taskSummary, cfg, stdout, jsonOutput)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add global flags
	cmd.PersistentFlags().BoolP("no-prompt", "y", false, "Disable interactive prompts")
	cmd.PersistentFlags().BoolP("verbose", "V", false, "Enable verbose/debug output")
	cmd.PersistentFlags().Bool("json", false, "Output in JSON format")

	// Add action-specific flags
	cmd.Flags().StringP("priority", "p", "", "Task priority (0-9) for add/update, or filter (1,2,3 or high/medium/low) for get")
	cmd.Flags().StringP("status", "s", "", "Task status (TODO, IN-PROGRESS, DONE, CANCELLED)")
	cmd.Flags().String("summary", "", "New task summary (for update)")
	cmd.Flags().String("due-date", "", "Due date in YYYY-MM-DD format (for add/update, use \"\" to clear)")
	cmd.Flags().String("start-date", "", "Start date in YYYY-MM-DD format (for add/update, use \"\" to clear)")
	cmd.Flags().StringSlice("tag", nil, "Tag/category for add/update, or filter by tag for get (can be specified multiple times or comma-separated)")
	cmd.Flags().StringP("parent", "P", "", "Parent task summary (for add/update subtasks)")
	cmd.Flags().BoolP("literal", "l", false, "Treat task summary literally (don't parse / as hierarchy separator)")
	cmd.Flags().Bool("no-parent", false, "Remove parent relationship (for update, makes task root-level)")
	cmd.Flags().StringP("view", "v", "", "View to use for displaying tasks (default, all, or custom view name)")

	// Add list subcommand
	cmd.AddCommand(newListCmd(stdout, cfg))

	// Add view subcommand
	cmd.AddCommand(newViewCmd(stdout, cfg))

	// Add credentials subcommand
	cmd.AddCommand(newCredentialsCmd(stdout, stderr, cfg))

	return cmd
}

// newListCmd creates the 'list' subcommand for list management
func newListCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Manage task lists",
		Long:  "View all lists or manage lists with subcommands.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Update config from flags
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			be, err := getBackend(cfg)
			if err != nil {
				return err
			}
			defer func() { _ = be.Close() }()

			jsonOutput, _ := cmd.Flags().GetBool("json")
			return doListView(context.Background(), be, cfg, stdout, jsonOutput)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add subcommands
	listCmd.AddCommand(newListCreateCmd(stdout, cfg))
	listCmd.AddCommand(newListDeleteCmd(stdout, cfg))
	listCmd.AddCommand(newListInfoCmd(stdout, cfg))
	listCmd.AddCommand(newListTrashCmd(stdout, cfg))

	return listCmd
}

// newListCreateCmd creates the 'list create' subcommand
func newListCreateCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new list",
		Long:  "Create a new task list with the given name.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Update config from flags
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			be, err := getBackend(cfg)
			if err != nil {
				return err
			}
			defer func() { _ = be.Close() }()

			jsonOutput, _ := cmd.Flags().GetBool("json")
			return doListCreate(context.Background(), be, args[0], cfg, stdout, jsonOutput)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doListView displays all task lists with their task counts
func doListView(ctx context.Context, be backend.TaskManager, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	lists, err := be.GetLists(ctx)
	if err != nil {
		return err
	}

	if jsonOutput {
		// Build JSON output with task counts
		type listJSON struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Color    string `json:"color,omitempty"`
			Tasks    int    `json:"tasks"`
			Modified string `json:"modified"`
		}
		var output []listJSON
		for _, l := range lists {
			tasks, _ := be.GetTasks(ctx, l.ID)
			output = append(output, listJSON{
				ID:       l.ID,
				Name:     l.Name,
				Color:    l.Color,
				Tasks:    len(tasks),
				Modified: l.Modified.Format("2006-01-02T15:04:05Z"),
			})
		}
		if output == nil {
			output = []listJSON{}
		}
		jsonBytes, err := json.Marshal(output)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdout, string(jsonBytes))
		return nil
	}

	if len(lists) == 0 {
		_, _ = fmt.Fprintln(stdout, "No lists found. Create one with: todoat list create \"MyList\"")
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
		}
		return nil
	}

	// Display formatted list with task counts
	_, _ = fmt.Fprintf(stdout, "Available lists (%d):\n\n", len(lists))
	_, _ = fmt.Fprintf(stdout, "%-20s %s\n", "NAME", "TASKS")

	for _, l := range lists {
		tasks, _ := be.GetTasks(ctx, l.ID)
		_, _ = fmt.Fprintf(stdout, "%-20s %d\n", l.Name, len(tasks))
	}

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// doListCreate creates a new task list
func doListCreate(ctx context.Context, be backend.TaskManager, name string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// Check for duplicate list name
	lists, err := be.GetLists(ctx)
	if err != nil {
		return err
	}

	for _, l := range lists {
		if strings.EqualFold(l.Name, name) {
			return fmt.Errorf("list '%s' already exists", name)
		}
	}

	// Create the list
	list, err := be.CreateList(ctx, name)
	if err != nil {
		return err
	}

	if jsonOutput {
		type listJSON struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Color    string `json:"color,omitempty"`
			Modified string `json:"modified"`
		}
		output := listJSON{
			ID:       list.ID,
			Name:     list.Name,
			Color:    list.Color,
			Modified: list.Modified.Format("2006-01-02T15:04:05Z"),
		}
		jsonBytes, err := json.Marshal(output)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdout, string(jsonBytes))
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "Created list: %s\n", list.Name)
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// newListDeleteCmd creates the 'list delete' subcommand
func newListDeleteCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a list (move to trash)",
		Long:  "Soft-delete a task list by moving it to trash.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			be, err := getBackend(cfg)
			if err != nil {
				return err
			}
			defer func() { _ = be.Close() }()

			return doListDelete(context.Background(), be, args[0], cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doListDelete soft-deletes a list by name
func doListDelete(ctx context.Context, be backend.TaskManager, name string, cfg *Config, stdout io.Writer) error {
	// Find the list by name
	list, err := be.GetListByName(ctx, name)
	if err != nil {
		return err
	}
	if list == nil {
		_, _ = fmt.Fprintf(stdout, "Error: list '%s' not found\n", name)
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultError)
		}
		return fmt.Errorf("list '%s' not found", name)
	}

	// Delete the list
	if err := be.DeleteList(ctx, list.ID); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "Deleted list: %s\n", list.Name)
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// newListInfoCmd creates the 'list info' subcommand
func newListInfoCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "info [name]",
		Short: "Show list details",
		Long:  "Display detailed information about a task list.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			be, err := getBackend(cfg)
			if err != nil {
				return err
			}
			defer func() { _ = be.Close() }()

			return doListInfo(context.Background(), be, args[0], cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doListInfo displays details about a list
func doListInfo(ctx context.Context, be backend.TaskManager, name string, cfg *Config, stdout io.Writer) error {
	// Find the list by name
	list, err := be.GetListByName(ctx, name)
	if err != nil {
		return err
	}
	if list == nil {
		_, _ = fmt.Fprintf(stdout, "Error: list '%s' not found\n", name)
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultError)
		}
		return fmt.Errorf("list '%s' not found", name)
	}

	// Get task count
	tasks, err := be.GetTasks(ctx, list.ID)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "Name:  %s\n", list.Name)
	_, _ = fmt.Fprintf(stdout, "ID:    %s\n", list.ID)
	if list.Color != "" {
		_, _ = fmt.Fprintf(stdout, "Color: %s\n", list.Color)
	}
	_, _ = fmt.Fprintf(stdout, "Tasks: %d\n", len(tasks))

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// newListTrashCmd creates the 'list trash' subcommand
func newListTrashCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	trashCmd := &cobra.Command{
		Use:   "trash",
		Short: "View and manage deleted lists",
		Long:  "View lists in trash or use subcommands to restore/purge.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			be, err := getBackend(cfg)
			if err != nil {
				return err
			}
			defer func() { _ = be.Close() }()

			return doListTrashView(context.Background(), be, cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	trashCmd.AddCommand(newListTrashRestoreCmd(stdout, cfg))
	trashCmd.AddCommand(newListTrashPurgeCmd(stdout, cfg))

	return trashCmd
}

// doListTrashView displays deleted lists
func doListTrashView(ctx context.Context, be backend.TaskManager, cfg *Config, stdout io.Writer) error {
	lists, err := be.GetDeletedLists(ctx)
	if err != nil {
		return err
	}

	if len(lists) == 0 {
		_, _ = fmt.Fprintln(stdout, "Trash is empty.")
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
		}
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "Deleted lists (%d):\n\n", len(lists))
	_, _ = fmt.Fprintf(stdout, "%-20s %s\n", "NAME", "DELETED")

	for _, l := range lists {
		deletedStr := ""
		if l.DeletedAt != nil {
			deletedStr = l.DeletedAt.Format("2006-01-02 15:04")
		}
		_, _ = fmt.Fprintf(stdout, "%-20s %s\n", l.Name, deletedStr)
	}

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// newListTrashRestoreCmd creates the 'list trash restore' subcommand
func newListTrashRestoreCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "restore [name]",
		Short: "Restore a list from trash",
		Long:  "Restore a deleted task list from trash.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			be, err := getBackend(cfg)
			if err != nil {
				return err
			}
			defer func() { _ = be.Close() }()

			return doListRestore(context.Background(), be, args[0], cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doListRestore restores a list from trash
func doListRestore(ctx context.Context, be backend.TaskManager, name string, cfg *Config, stdout io.Writer) error {
	// Find the deleted list by name
	list, err := be.GetDeletedListByName(ctx, name)
	if err != nil {
		return err
	}
	if list == nil {
		_, _ = fmt.Fprintf(stdout, "Error: list '%s' not found in trash\n", name)
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultError)
		}
		return fmt.Errorf("list '%s' not found in trash", name)
	}

	// Restore the list
	if err := be.RestoreList(ctx, list.ID); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "Restored list: %s\n", list.Name)
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// newListTrashPurgeCmd creates the 'list trash purge' subcommand
func newListTrashPurgeCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "purge [name]",
		Short: "Permanently delete a list from trash",
		Long:  "Permanently delete a task list and all its tasks from trash.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			be, err := getBackend(cfg)
			if err != nil {
				return err
			}
			defer func() { _ = be.Close() }()

			return doListPurge(context.Background(), be, args[0], cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doListPurge permanently deletes a list from trash
func doListPurge(ctx context.Context, be backend.TaskManager, name string, cfg *Config, stdout io.Writer) error {
	// Find the deleted list by name
	list, err := be.GetDeletedListByName(ctx, name)
	if err != nil {
		return err
	}
	if list == nil {
		_, _ = fmt.Fprintf(stdout, "Error: list '%s' not found in trash\n", name)
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultError)
		}
		return fmt.Errorf("list '%s' not found in trash", name)
	}

	// Purge the list
	if err := be.PurgeList(ctx, list.ID); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "Permanently deleted list: %s\n", list.Name)
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
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
func executeAction(ctx context.Context, cmd *cobra.Command, be backend.TaskManager, list *backend.List, action, taskSummary string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	switch action {
	case "get":
		statusFilter, _ := cmd.Flags().GetString("status")
		priorityFilterStr, _ := cmd.Flags().GetString("priority")
		priorityFilter, err := parsePriorityFilter(priorityFilterStr)
		if err != nil {
			return err
		}
		tagFilter, _ := cmd.Flags().GetStringSlice("tag")
		tagFilter = normalizeTagSlice(tagFilter)
		viewName, _ := cmd.Flags().GetString("view")
		return doGet(ctx, be, list, statusFilter, priorityFilter, tagFilter, viewName, cfg, stdout, jsonOutput)
	case "add":
		priorityStr, _ := cmd.Flags().GetString("priority")
		priority, err := parsePrioritySingle(priorityStr)
		if err != nil {
			return err
		}
		dueDateStr, _ := cmd.Flags().GetString("due-date")
		startDateStr, _ := cmd.Flags().GetString("start-date")
		dueDate, err := parseDate(dueDateStr)
		if err != nil {
			return fmt.Errorf("invalid due-date: %w", err)
		}
		startDate, err := parseDate(startDateStr)
		if err != nil {
			return fmt.Errorf("invalid start-date: %w", err)
		}
		tags, _ := cmd.Flags().GetStringSlice("tag")
		tags = normalizeTagSlice(tags)
		categories := strings.Join(tags, ",")
		parentSummary, _ := cmd.Flags().GetString("parent")
		literal, _ := cmd.Flags().GetBool("literal")
		return doAdd(ctx, be, list, taskSummary, priority, dueDate, startDate, categories, parentSummary, literal, cfg, stdout, jsonOutput)
	case "update":
		priorityStr, _ := cmd.Flags().GetString("priority")
		priority, err := parsePrioritySingle(priorityStr)
		if err != nil {
			return err
		}
		status, _ := cmd.Flags().GetString("status")
		newSummary, _ := cmd.Flags().GetString("summary")
		dueDateStr, dueDateChanged := cmd.Flags().GetString("due-date")
		startDateStr, startDateChanged := cmd.Flags().GetString("start-date")
		// Check if flags were actually set (not just empty)
		dueDateFlagSet := cmd.Flags().Changed("due-date")
		startDateFlagSet := cmd.Flags().Changed("start-date")
		_ = dueDateChanged
		_ = startDateChanged
		var dueDate, startDate *time.Time
		var clearDueDate, clearStartDate bool
		if dueDateFlagSet {
			if dueDateStr == "" {
				clearDueDate = true
			} else {
				d, err := parseDate(dueDateStr)
				if err != nil {
					return fmt.Errorf("invalid due-date: %w", err)
				}
				dueDate = d
			}
		}
		if startDateFlagSet {
			if startDateStr == "" {
				clearStartDate = true
			} else {
				d, err := parseDate(startDateStr)
				if err != nil {
					return fmt.Errorf("invalid start-date: %w", err)
				}
				startDate = d
			}
		}
		tagFlagSet := cmd.Flags().Changed("tag")
		var newCategories *string
		if tagFlagSet {
			tags, _ := cmd.Flags().GetStringSlice("tag")
			tags = normalizeTagSlice(tags)
			cat := strings.Join(tags, ",")
			newCategories = &cat
		}
		parentSummary, _ := cmd.Flags().GetString("parent")
		noParent, _ := cmd.Flags().GetBool("no-parent")
		return doUpdate(ctx, be, list, taskSummary, newSummary, status, priority, dueDate, startDate, clearDueDate, clearStartDate, newCategories, parentSummary, noParent, cfg, stdout, jsonOutput)
	case "complete":
		return doComplete(ctx, be, list, taskSummary, cfg, stdout, jsonOutput)
	case "delete":
		return doDelete(ctx, be, list, taskSummary, cfg, stdout, jsonOutput)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// doGet lists all tasks in a list, optionally filtering by status, priority, and/or tags
func doGet(ctx context.Context, be backend.TaskManager, list *backend.List, statusFilter string, priorityFilter []int, tagFilter []string, viewName string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	tasks, err := be.GetTasks(ctx, list.ID)
	if err != nil {
		return err
	}

	// Load and apply view if specified
	if viewName != "" {
		return doGetWithView(ctx, tasks, list, viewName, cfg, stdout, jsonOutput)
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

	// Filter by priority if specified
	if len(priorityFilter) > 0 {
		var filteredTasks []backend.Task
		for _, t := range tasks {
			if matchesPriorityFilter(t.Priority, priorityFilter) {
				filteredTasks = append(filteredTasks, t)
			}
		}
		tasks = filteredTasks
	}

	// Filter by tags if specified (OR logic - match any tag)
	if len(tagFilter) > 0 {
		var filteredTasks []backend.Task
		for _, t := range tasks {
			if matchesTagFilter(t.Categories, tagFilter) {
				filteredTasks = append(filteredTasks, t)
			}
		}
		tasks = filteredTasks
	}

	if jsonOutput {
		return outputTaskListJSON(tasks, list, stdout)
	}

	if len(tasks) == 0 {
		_, _ = fmt.Fprintf(stdout, "No tasks in list '%s'\n", list.Name)
	} else {
		_, _ = fmt.Fprintf(stdout, "Tasks in '%s':\n", list.Name)
		printTaskTree(tasks, stdout)
	}

	// Emit INFO_ONLY result code in no-prompt mode
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// doGetWithView lists tasks using a view configuration
func doGetWithView(ctx context.Context, tasks []backend.Task, list *backend.List, viewName string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// Load view
	viewsDir := getViewsDir(cfg)
	loader := views.NewLoader(viewsDir)
	view, err := loader.LoadView(viewName)
	if err != nil {
		return err
	}

	// Apply view filters and sorting
	filteredTasks := views.FilterTasks(tasks, view.Filters)
	sortedTasks := views.SortTasks(filteredTasks, view.Sort)

	if jsonOutput {
		return outputTaskListJSON(sortedTasks, list, stdout)
	}

	if len(sortedTasks) == 0 {
		_, _ = fmt.Fprintf(stdout, "No tasks in list '%s'\n", list.Name)
	} else {
		_, _ = fmt.Fprintf(stdout, "Tasks in '%s':\n", list.Name)
		views.RenderTasksWithView(sortedTasks, view, stdout)
	}

	// Emit INFO_ONLY result code in no-prompt mode
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// getViewsDir returns the path to the views directory
func getViewsDir(cfg *Config) string {
	if cfg != nil && cfg.ViewsPath != "" {
		return cfg.ViewsPath
	}

	// Default to XDG config directory
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "todoat", "views")
}

// newViewCmd creates the 'view' subcommand for view management
func newViewCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	viewCmd := &cobra.Command{
		Use:           "view",
		Short:         "Manage views",
		Long:          "View management commands for listing and working with views.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	viewCmd.AddCommand(newViewListCmd(stdout, cfg))

	return viewCmd
}

// newViewListCmd creates the 'view list' subcommand
func newViewListCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available views",
		Long:  "List all available views including built-in and custom views.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			return doViewList(cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doViewList displays all available views
func doViewList(cfg *Config, stdout io.Writer) error {
	viewsDir := getViewsDir(cfg)
	loader := views.NewLoader(viewsDir)

	viewList, err := loader.ListViews()
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(stdout, "Available views:")
	for _, v := range viewList {
		viewType := "custom"
		if v.BuiltIn {
			viewType = "built-in"
		}
		_, _ = fmt.Fprintf(stdout, "  - %s (%s)\n", v.Name, viewType)
	}

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// taskNode represents a task with its children for tree building
type taskNode struct {
	task     backend.Task
	children []*taskNode
}

// printTaskTree prints tasks in a tree structure with box-drawing characters
func printTaskTree(tasks []backend.Task, stdout io.Writer) {
	// Build a map from task ID to task
	taskMap := make(map[string]*backend.Task)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}

	// Build tree nodes
	nodeMap := make(map[string]*taskNode)
	var rootNodes []*taskNode

	// First pass: create nodes for all tasks
	for i := range tasks {
		nodeMap[tasks[i].ID] = &taskNode{task: tasks[i]}
	}

	// Second pass: build parent-child relationships
	for i := range tasks {
		node := nodeMap[tasks[i].ID]
		if tasks[i].ParentID == "" {
			// Root-level task
			rootNodes = append(rootNodes, node)
		} else if parentNode, ok := nodeMap[tasks[i].ParentID]; ok {
			// Has valid parent
			parentNode.children = append(parentNode.children, node)
		} else {
			// Orphan task (parent not in list) - show at root level
			rootNodes = append(rootNodes, node)
		}
	}

	// Print the tree recursively
	for i, node := range rootNodes {
		isLast := i == len(rootNodes)-1
		printTaskNode(node, "", isLast, stdout)
	}
}

// printTaskNode recursively prints a task node with tree visualization
func printTaskNode(node *taskNode, prefix string, isLast bool, stdout io.Writer) {
	t := node.task

	// Build the display line
	statusIcon := getStatusIcon(t.Status)
	priorityStr := ""
	if t.Priority > 0 {
		priorityStr = fmt.Sprintf(" [P%d]", t.Priority)
	}
	tagsStr := ""
	if t.Categories != "" {
		tagsStr = fmt.Sprintf(" {%s}", t.Categories)
	}

	// Choose the appropriate tree character
	var treeChar string
	if prefix == "" {
		// Root level - no tree character
		treeChar = "  "
	} else if isLast {
		treeChar = "└─ "
	} else {
		treeChar = "├─ "
	}

	_, _ = fmt.Fprintf(stdout, "%s%s%s %s%s%s\n", prefix, treeChar, statusIcon, t.Summary, priorityStr, tagsStr)

	// Build the prefix for children
	var childPrefix string
	if prefix == "" {
		childPrefix = "  "
	} else if isLast {
		childPrefix = prefix + "   "
	} else {
		childPrefix = prefix + "│  "
	}

	// Print children
	for i, child := range node.children {
		isChildLast := i == len(node.children)-1
		printTaskNode(child, childPrefix, isChildLast, stdout)
	}
}

// matchesTagFilter checks if a task's categories match any of the filter tags (OR logic)
func matchesTagFilter(categories string, filterTags []string) bool {
	if categories == "" {
		return false
	}
	taskTags := strings.Split(categories, ",")
	for _, filterTag := range filterTags {
		filterTag = strings.TrimSpace(filterTag)
		for _, taskTag := range taskTags {
			if strings.EqualFold(strings.TrimSpace(taskTag), filterTag) {
				return true
			}
		}
	}
	return false
}

// normalizeTagSlice processes a tag slice to handle comma-separated values
func normalizeTagSlice(tags []string) []string {
	var result []string
	for _, tag := range tags {
		// Split by comma in case of comma-separated values
		parts := strings.Split(tag, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				result = append(result, part)
			}
		}
	}
	return result
}

// matchesPriorityFilter checks if a task's priority matches any of the filter priorities
func matchesPriorityFilter(taskPriority int, priorities []int) bool {
	for _, p := range priorities {
		if taskPriority == p {
			return true
		}
	}
	return false
}

// parsePriorityFilter parses a priority filter string into a slice of priority values
// Supports: single value (1), comma-separated (1,2,3), aliases (high, medium, low)
func parsePriorityFilter(s string) ([]int, error) {
	if s == "" {
		return nil, nil
	}

	s = strings.TrimSpace(strings.ToLower(s))

	// Handle aliases
	switch s {
	case "high":
		return []int{1, 2, 3, 4}, nil
	case "medium":
		return []int{5}, nil
	case "low":
		return []int{6, 7, 8, 9}, nil
	}

	// Handle comma-separated values
	parts := strings.Split(s, ",")
	var priorities []int
	for _, part := range parts {
		part = strings.TrimSpace(part)
		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid priority value: %s", part)
		}
		if val < 0 || val > 9 {
			return nil, fmt.Errorf("priority must be between 0 and 9, got: %d", val)
		}
		priorities = append(priorities, val)
	}

	return priorities, nil
}

// parsePrioritySingle parses a single priority value from string
func parsePrioritySingle(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid priority value: %s", s)
	}
	if val < 0 || val > 9 {
		return 0, fmt.Errorf("priority must be between 0 and 9, got: %d", val)
	}
	return val, nil
}

// parseDate parses a date string in YYYY-MM-DD format
func parseDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, fmt.Errorf("date must be in YYYY-MM-DD format (e.g., 2026-01-31)")
	}
	return &t, nil
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
func doAdd(ctx context.Context, be backend.TaskManager, list *backend.List, summary string, priority int, dueDate, startDate *time.Time, categories string, parentSummary string, literal bool, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	if summary == "" {
		return fmt.Errorf("task summary is required")
	}

	var parentID string

	// If -P/--parent flag is provided, find the parent task
	if parentSummary != "" {
		parent, err := findTask(ctx, be, list, parentSummary, cfg)
		if err != nil {
			return fmt.Errorf("parent task not found: %w", err)
		}
		parentID = parent.ID
	}

	// Handle path-based hierarchy creation unless --literal flag is set
	if !literal && strings.Contains(summary, "/") && parentSummary == "" {
		return doAddHierarchy(ctx, be, list, summary, priority, dueDate, startDate, categories, cfg, stdout, jsonOutput)
	}

	task := &backend.Task{
		Summary:    summary,
		Priority:   priority,
		Status:     backend.StatusNeedsAction,
		DueDate:    dueDate,
		StartDate:  startDate,
		Categories: categories,
		ParentID:   parentID,
	}

	created, err := be.CreateTask(ctx, list.ID, task)
	if err != nil {
		return err
	}

	if jsonOutput {
		return outputActionJSON("add", created, stdout)
	}

	_, _ = fmt.Fprintf(stdout, "Created task: %s (ID: %s)\n", created.Summary, created.ID)

	// Emit ACTION_COMPLETED result code in no-prompt mode
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// doAddHierarchy creates a task hierarchy from a path like "A/B/C"
func doAddHierarchy(ctx context.Context, be backend.TaskManager, list *backend.List, path string, priority int, dueDate, startDate *time.Time, categories string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return fmt.Errorf("invalid path")
	}

	tasks, err := be.GetTasks(ctx, list.ID)
	if err != nil {
		return err
	}

	var parentID string
	var lastCreated *backend.Task

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check if this task already exists at the current level
		var existingTask *backend.Task
		for _, t := range tasks {
			if strings.EqualFold(t.Summary, part) && t.ParentID == parentID {
				existingTask = &t
				break
			}
		}

		if existingTask != nil {
			// Task exists, use it as parent for next level
			parentID = existingTask.ID
			lastCreated = existingTask
		} else {
			// Create the task
			// Only apply priority, dates, and categories to the leaf task
			taskPriority := 0
			var taskDueDate, taskStartDate *time.Time
			taskCategories := ""
			if i == len(parts)-1 {
				taskPriority = priority
				taskDueDate = dueDate
				taskStartDate = startDate
				taskCategories = categories
			}

			task := &backend.Task{
				Summary:    part,
				Priority:   taskPriority,
				Status:     backend.StatusNeedsAction,
				DueDate:    taskDueDate,
				StartDate:  taskStartDate,
				Categories: taskCategories,
				ParentID:   parentID,
			}

			created, err := be.CreateTask(ctx, list.ID, task)
			if err != nil {
				return err
			}

			parentID = created.ID
			lastCreated = created

			// Add to tasks slice so subsequent iterations can find it
			tasks = append(tasks, *created)
		}
	}

	if lastCreated == nil {
		return fmt.Errorf("no task created")
	}

	if jsonOutput {
		return outputActionJSON("add", lastCreated, stdout)
	}

	_, _ = fmt.Fprintf(stdout, "Created task: %s (ID: %s)\n", lastCreated.Summary, lastCreated.ID)

	// Emit ACTION_COMPLETED result code in no-prompt mode
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// doUpdate modifies an existing task
func doUpdate(ctx context.Context, be backend.TaskManager, list *backend.List, taskSummary, newSummary, status string, priority int, dueDate, startDate *time.Time, clearDueDate, clearStartDate bool, newCategories *string, parentSummary string, noParent bool, cfg *Config, stdout io.Writer, jsonOutput bool) error {
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
	if dueDate != nil {
		task.DueDate = dueDate
	}
	if clearDueDate {
		task.DueDate = nil
	}
	if startDate != nil {
		task.StartDate = startDate
	}
	if clearStartDate {
		task.StartDate = nil
	}
	if newCategories != nil {
		task.Categories = *newCategories
	}

	// Handle parent updates
	if noParent {
		task.ParentID = ""
	} else if parentSummary != "" {
		parent, err := findTask(ctx, be, list, parentSummary, cfg)
		if err != nil {
			return fmt.Errorf("parent task not found: %w", err)
		}

		// Check for circular reference
		if err := checkCircularReference(ctx, be, list, task.ID, parent.ID); err != nil {
			return err
		}

		task.ParentID = parent.ID
	}

	updated, err := be.UpdateTask(ctx, list.ID, task)
	if err != nil {
		return err
	}

	if jsonOutput {
		return outputActionJSON("update", updated, stdout)
	}

	_, _ = fmt.Fprintf(stdout, "Updated task: %s\n", updated.Summary)

	// Emit ACTION_COMPLETED result code in no-prompt mode
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// checkCircularReference checks if setting parentID as the parent of taskID would create a circular reference
func checkCircularReference(ctx context.Context, be backend.TaskManager, list *backend.List, taskID, parentID string) error {
	if taskID == parentID {
		return fmt.Errorf("circular reference: cannot set task as its own parent")
	}

	tasks, err := be.GetTasks(ctx, list.ID)
	if err != nil {
		return err
	}

	// Build a map for quick lookup
	taskMap := make(map[string]*backend.Task)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}

	// Walk up the parent chain from parentID to check if we hit taskID
	currentID := parentID
	visited := make(map[string]bool)
	for currentID != "" {
		if visited[currentID] {
			return fmt.Errorf("circular reference detected in existing hierarchy")
		}
		visited[currentID] = true

		if currentID == taskID {
			return fmt.Errorf("circular reference: cannot set a descendant as parent")
		}

		current, ok := taskMap[currentID]
		if !ok {
			break
		}
		currentID = current.ParentID
	}

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
func doComplete(ctx context.Context, be backend.TaskManager, list *backend.List, taskSummary string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	task, err := findTask(ctx, be, list, taskSummary, cfg)
	if err != nil {
		return err
	}

	task.Status = backend.StatusCompleted
	// Auto-set completed timestamp
	now := time.Now().UTC()
	task.Completed = &now

	updated, err := be.UpdateTask(ctx, list.ID, task)
	if err != nil {
		return err
	}

	if jsonOutput {
		return outputActionJSON("complete", updated, stdout)
	}

	_, _ = fmt.Fprintf(stdout, "Completed task: %s\n", updated.Summary)

	// Emit ACTION_COMPLETED result code in no-prompt mode
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// doDelete removes a task
func doDelete(ctx context.Context, be backend.TaskManager, list *backend.List, taskSummary string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	task, err := findTask(ctx, be, list, taskSummary, cfg)
	if err != nil {
		return err
	}

	// Store task info before deletion for JSON output
	deletedTask := *task

	// Find all descendants for cascade delete
	tasks, err := be.GetTasks(ctx, list.ID)
	if err != nil {
		return err
	}

	descendantIDs := findDescendants(task.ID, tasks)

	// Delete descendants first (bottom-up to avoid FK issues), then parent
	for i := len(descendantIDs) - 1; i >= 0; i-- {
		if err := be.DeleteTask(ctx, list.ID, descendantIDs[i]); err != nil {
			return err
		}
	}

	if err := be.DeleteTask(ctx, list.ID, task.ID); err != nil {
		return err
	}

	if jsonOutput {
		return outputActionJSON("delete", &deletedTask, stdout)
	}

	_, _ = fmt.Fprintf(stdout, "Deleted task: %s\n", task.Summary)

	// Emit ACTION_COMPLETED result code in no-prompt mode
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// findDescendants returns a list of task IDs that are descendants of the given parent
func findDescendants(parentID string, tasks []backend.Task) []string {
	var result []string
	// Build a map of parent to children
	childMap := make(map[string][]string)
	for _, t := range tasks {
		if t.ParentID != "" {
			childMap[t.ParentID] = append(childMap[t.ParentID], t.ID)
		}
	}

	// BFS to find all descendants
	queue := []string{parentID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, childID := range childMap[current] {
			result = append(result, childID)
			queue = append(queue, childID)
		}
	}

	return result
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

// JSON output structures
type taskJSON struct {
	UID       string   `json:"uid"`
	Summary   string   `json:"summary"`
	Status    string   `json:"status"`
	Priority  int      `json:"priority"`
	ParentID  string   `json:"parent_id,omitempty"`
	DueDate   *string  `json:"due_date,omitempty"`
	StartDate *string  `json:"start_date,omitempty"`
	Completed *string  `json:"completed,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

type listTasksResponse struct {
	Tasks  []taskJSON `json:"tasks"`
	List   string     `json:"list"`
	Count  int        `json:"count"`
	Result string     `json:"result"`
}

type actionResponse struct {
	Action string   `json:"action"`
	Task   taskJSON `json:"task"`
	Result string   `json:"result"`
}

type errorResponse struct {
	Error  string `json:"error"`
	Code   int    `json:"code"`
	Result string `json:"result"`
}

// taskToJSON converts a backend.Task to taskJSON
func taskToJSON(t *backend.Task) taskJSON {
	result := taskJSON{
		UID:      t.ID,
		Summary:  t.Summary,
		Status:   statusToString(t.Status),
		Priority: t.Priority,
		ParentID: t.ParentID,
	}
	if t.DueDate != nil {
		s := t.DueDate.Format("2006-01-02")
		result.DueDate = &s
	}
	if t.StartDate != nil {
		s := t.StartDate.Format("2006-01-02")
		result.StartDate = &s
	}
	if t.Completed != nil {
		s := t.Completed.Format(time.RFC3339)
		result.Completed = &s
	}
	if t.Categories != "" {
		result.Tags = strings.Split(t.Categories, ",")
		// Trim whitespace from each tag
		for i, tag := range result.Tags {
			result.Tags[i] = strings.TrimSpace(tag)
		}
	}
	return result
}

// statusToString converts TaskStatus to string representation
func statusToString(s backend.TaskStatus) string {
	switch s {
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

// outputTaskListJSON outputs tasks in JSON format
func outputTaskListJSON(tasks []backend.Task, list *backend.List, stdout io.Writer) error {
	var jsonTasks []taskJSON
	for _, t := range tasks {
		jsonTasks = append(jsonTasks, taskToJSON(&t))
	}
	if jsonTasks == nil {
		jsonTasks = []taskJSON{}
	}

	response := listTasksResponse{
		Tasks:  jsonTasks,
		List:   list.Name,
		Count:  len(jsonTasks),
		Result: ResultInfoOnly,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(stdout, string(jsonBytes))
	return nil
}

// outputActionJSON outputs action result in JSON format
func outputActionJSON(action string, task *backend.Task, stdout io.Writer) error {
	response := actionResponse{
		Action: action,
		Task:   taskToJSON(task),
		Result: ResultActionCompleted,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(stdout, string(jsonBytes))
	return nil
}

// outputErrorJSON outputs error in JSON format
func outputErrorJSON(err error, stdout io.Writer) {
	response := errorResponse{
		Error:  err.Error(),
		Code:   1,
		Result: ResultError,
	}

	jsonBytes, _ := json.Marshal(response)
	_, _ = fmt.Fprintln(stdout, string(jsonBytes))
}

// newCredentialsCmd creates the 'credentials' subcommand for credential management
func newCredentialsCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	credentialsCmd := &cobra.Command{
		Use:   "credentials",
		Short: "Manage backend credentials",
		Long:  "Store, retrieve, and manage credentials for backend services securely.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	credentialsCmd.AddCommand(newCredentialsSetCmd(stdout, stderr, cfg))
	credentialsCmd.AddCommand(newCredentialsGetCmd(stdout, stderr, cfg))
	credentialsCmd.AddCommand(newCredentialsDeleteCmd(stdout, stderr, cfg))
	credentialsCmd.AddCommand(newCredentialsListCmd(stdout, stderr, cfg))

	return credentialsCmd
}

// newCredentialsSetCmd creates the 'credentials set' subcommand
func newCredentialsSetCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [backend] [username]",
		Short: "Store credentials in system keyring",
		Long:  "Store credentials securely in the system keyring (macOS Keychain, Windows Credential Manager, or Linux Secret Service).",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			backend := args[0]
			username := args[1]
			prompt, _ := cmd.Flags().GetBool("prompt")

			manager := credentials.NewManager()
			handler := credentials.NewCLIHandler(manager, os.Stdin, stdout, stderr)
			return handler.Set(backend, username, prompt)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Flags().Bool("prompt", false, "Prompt for password input (required for security)")
	return cmd
}

// newCredentialsGetCmd creates the 'credentials get' subcommand
func newCredentialsGetCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "get [backend] [username]",
		Short: "Retrieve credentials and show source",
		Long:  "Retrieve credentials from the priority chain (keyring > environment > config URL) and display the source.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			backend := args[0]
			username := args[1]
			jsonOutput, _ := cmd.Flags().GetBool("json")

			manager := credentials.NewManager()
			handler := credentials.NewCLIHandler(manager, nil, stdout, stderr)
			return handler.Get(backend, username, jsonOutput)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// newCredentialsDeleteCmd creates the 'credentials delete' subcommand
func newCredentialsDeleteCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "delete [backend] [username]",
		Short: "Remove credentials from system keyring",
		Long:  "Remove stored credentials from the system keyring. Environment variables and config URL credentials are not affected.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			backend := args[0]
			username := args[1]

			manager := credentials.NewManager()
			handler := credentials.NewCLIHandler(manager, nil, stdout, stderr)
			return handler.Delete(backend, username)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// newCredentialsListCmd creates the 'credentials list' subcommand
func newCredentialsListCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all backends with credential status",
		Long:  "Show all configured backends and whether credentials are available for each.",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOutput, _ := cmd.Flags().GetBool("json")

			// TODO: Get backend configs from actual configuration
			// For now, return a placeholder list
			backends := []credentials.BackendConfig{
				{Name: "nextcloud", Username: ""},
				{Name: "todoist", Username: ""},
			}

			manager := credentials.NewManager()
			handler := credentials.NewCLIHandler(manager, nil, stdout, stderr)
			return handler.List(backends, jsonOutput)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}
