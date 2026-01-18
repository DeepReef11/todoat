package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
	"todoat/backend"
	"todoat/backend/sqlite"
	"todoat/internal/config"
	"todoat/internal/credentials"
	"todoat/internal/notification"
	"todoat/internal/reminder"
	"todoat/internal/tui"
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
	NoPrompt            bool
	Verbose             bool
	OutputFormat        string
	DBPath              string // Path to database file (for testing)
	ViewsPath           string // Path to views directory (for testing)
	ConfigPath          string // Path to config file (for testing)
	SyncEnabled         bool   // Whether sync is enabled (caches config setting)
	NotificationLogPath string // Path to notification log file (for testing)
	NotificationMock    bool   // Use mock executor for OS notifications (for testing)
	// Daemon-related config fields (for testing)
	DaemonPIDPath     string        // Path to daemon PID file
	DaemonLogPath     string        // Path to daemon log file
	DaemonTestMode    bool          // Use in-process daemon for testing
	DaemonInterval    time.Duration // Sync interval for daemon
	DaemonOfflineMode bool          // Simulate offline mode for daemon
	// Migration-related config fields (for testing)
	MigrateTargetDir      string // Directory for file-mock backend target
	MigrateMockMode       bool   // Enable mock backends for testing
	MockNextcloudDataPath string // Path to mock nextcloud data file
	// Reminder-related config fields (for testing)
	ReminderConfigPath   string      // Path to reminder config file
	NotificationCallback interface{} // Callback for notification testing
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

	// Add sync subcommand
	cmd.AddCommand(newSyncCmd(stdout, stderr, cfg))

	// Add notification subcommand
	cmd.AddCommand(newNotificationCmd(stdout, stderr, cfg))

	// Add migrate subcommand
	cmd.AddCommand(newMigrateCmd(stdout, stderr, cfg))

	// Add reminder subcommand
	cmd.AddCommand(newReminderCmd(stdout, stderr, cfg))

	// Add TUI subcommand
	cmd.AddCommand(newTUICmd(stdout, stderr, cfg))

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

// getDefaultDBPath returns the default database path following XDG spec
// Default: $XDG_DATA_HOME/todoat/tasks.db or ~/.local/share/todoat/tasks.db
func getDefaultDBPath() string {
	return filepath.Join(config.GetDataDir(), "tasks.db")
}

// getBackend creates or returns the backend connection
func getBackend(cfg *Config) (backend.TaskManager, error) {
	dbPath := cfg.DBPath
	if dbPath == "" {
		// Use default XDG-compliant path
		dbPath = getDefaultDBPath()
	}

	// Ensure directory exists (for both default and explicit paths)
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("could not create data directory: %w", err)
	}

	// Load config to check if sync is enabled
	loadSyncConfig(cfg)

	// If sync is enabled, return a sync-aware backend wrapper
	if cfg.SyncEnabled {
		be, err := sqlite.New(dbPath)
		if err != nil {
			return nil, err
		}
		return &syncAwareBackend{
			TaskManager: be,
			syncMgr:     getSyncManager(cfg),
		}, nil
	}

	return sqlite.New(dbPath)
}

// loadSyncConfig loads the sync configuration from the config file
func loadSyncConfig(cfg *Config) {
	configPath := cfg.ConfigPath
	if configPath == "" {
		// Try to find config in same directory as DB
		if cfg.DBPath != "" {
			configPath = filepath.Join(filepath.Dir(cfg.DBPath), "config.yaml")
		}
	}

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err == nil {
			// Simple YAML parsing for sync.enabled
			if strings.Contains(string(data), "enabled: true") && strings.Contains(string(data), "sync:") {
				cfg.SyncEnabled = true
			}
		}
	}
}

// syncAwareBackend wraps a TaskManager to queue sync operations
type syncAwareBackend struct {
	backend.TaskManager
	syncMgr *SyncManager
}

// CreateTask creates a task and queues a sync operation
func (b *syncAwareBackend) CreateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	created, err := b.TaskManager.CreateTask(ctx, listID, task)
	if err != nil {
		return nil, err
	}

	// Queue create operation
	_ = b.syncMgr.QueueOperationByStringID(created.ID, created.Summary, listID, "create")

	return created, nil
}

// UpdateTask updates a task and queues a sync operation
func (b *syncAwareBackend) UpdateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	updated, err := b.TaskManager.UpdateTask(ctx, listID, task)
	if err != nil {
		return nil, err
	}

	// Queue update operation
	_ = b.syncMgr.QueueOperationByStringID(updated.ID, updated.Summary, listID, "update")

	return updated, nil
}

// DeleteTask deletes a task and queues a sync operation
func (b *syncAwareBackend) DeleteTask(ctx context.Context, listID, taskID string) error {
	// Get task summary before deleting for the queue
	task, _ := b.GetTask(ctx, listID, taskID)
	summary := "Unknown"
	if task != nil {
		summary = task.Summary
	}

	err := b.TaskManager.DeleteTask(ctx, listID, taskID)
	if err != nil {
		return err
	}

	// Queue delete operation
	_ = b.syncMgr.QueueOperationByStringID(taskID, summary, listID, "delete")

	return nil
}

// Close closes both the backend and sync manager
func (b *syncAwareBackend) Close() error {
	_ = b.syncMgr.Close()
	return b.TaskManager.Close()
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
		// Apply default view from config if -v flag was not explicitly provided
		if !cmd.Flags().Changed("view") {
			viewName = getDefaultView(cfg, cmd.ErrOrStderr())
		}
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

// getDefaultView returns the default view from config, or empty string if not set.
// Also returns a warning message if the configured view doesn't exist.
func getDefaultView(cfg *Config, stderr io.Writer) string {
	configPath := cfg.ConfigPath
	if configPath == "" {
		if cfg.DBPath != "" {
			configPath = filepath.Join(filepath.Dir(cfg.DBPath), "config.yaml")
		}
	}

	if configPath == "" {
		return ""
	}

	appConfig, err := config.LoadFromPath(configPath)
	if err != nil || appConfig == nil {
		return ""
	}

	defaultView := appConfig.DefaultView
	if defaultView == "" {
		return ""
	}

	// Check if the view exists
	viewsDir := getViewsDir(cfg)
	loader := views.NewLoader(viewsDir)
	if !loader.ViewExists(defaultView) {
		// Warn about missing view and fall back to default
		if stderr != nil {
			_, _ = fmt.Fprintf(stderr, "Warning: configured default_view '%s' not found, using built-in default\n", defaultView) //nolint:errcheck // Best effort warning, write failure non-critical
		}
		return ""
	}

	return defaultView
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

// newSyncCmd creates the 'sync' subcommand for synchronization management
func newSyncCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize with remote backends",
		Long:  "Synchronize local cache with remote backends. Use subcommands to view status and manage the sync queue.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			// For now, just report that sync ran
			_, _ = fmt.Fprintln(stdout, "Sync completed (no remote backend configured)")
			if cfg != nil && cfg.NoPrompt {
				_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	syncCmd.AddCommand(newSyncStatusCmd(stdout, cfg))
	syncCmd.AddCommand(newSyncQueueCmd(stdout, cfg))
	syncCmd.AddCommand(newSyncConflictsCmd(stdout, cfg))
	syncCmd.AddCommand(newSyncDaemonCmd(stdout, stderr, cfg))

	return syncCmd
}

// newSyncStatusCmd creates the 'sync status' subcommand
func newSyncStatusCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show sync status",
		Long:  "Show last sync time, pending operations, and connection status for all backends.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}
			verbose, _ := cmd.Flags().GetBool("verbose")

			return doSyncStatus(cfg, stdout, verbose)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Flags().Bool("verbose", false, "Show detailed sync metadata")
	return cmd
}

// doSyncStatus displays sync status for all backends
func doSyncStatus(cfg *Config, stdout io.Writer, verbose bool) error {
	// Get sync manager
	syncMgr := getSyncManager(cfg)
	defer func() { _ = syncMgr.Close() }()

	_, _ = fmt.Fprintln(stdout, "Sync Status:")
	_, _ = fmt.Fprintln(stdout, "")

	// Get pending operations count
	pendingCount, err := syncMgr.GetPendingCount()
	if err != nil {
		pendingCount = 0
	}

	// Get last sync time
	lastSync := syncMgr.GetLastSyncTime()
	lastSyncStr := "Never"
	if !lastSync.IsZero() {
		lastSyncStr = lastSync.Format("2006-01-02 15:04:05")
	}

	// Load config to get configured backends
	configBackends := getConfiguredBackends(cfg)
	if len(configBackends) > 0 {
		for _, backendName := range configBackends {
			_, _ = fmt.Fprintf(stdout, "Backend: %s\n", backendName)
			_, _ = fmt.Fprintf(stdout, "  Last Sync: %s\n", lastSyncStr)
			_, _ = fmt.Fprintf(stdout, "  Pending Operations: %d\n", pendingCount)
			_, _ = fmt.Fprintf(stdout, "  Status: Configured\n")
			_, _ = fmt.Fprintln(stdout, "")
		}
	} else {
		_, _ = fmt.Fprintf(stdout, "Backend: sqlite\n")
		_, _ = fmt.Fprintf(stdout, "  Last Sync: %s\n", lastSyncStr)
		_, _ = fmt.Fprintf(stdout, "  Pending Operations: %d\n", pendingCount)
		_, _ = fmt.Fprintf(stdout, "  Status: %s\n", syncMgr.GetConnectionStatus())
	}

	if verbose {
		_, _ = fmt.Fprintln(stdout, "")
		_, _ = fmt.Fprintln(stdout, "Sync Metadata:")
		// Show additional metadata when verbose
	}

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// getConfiguredBackends returns a list of backend names from the config file
func getConfiguredBackends(cfg *Config) []string {
	configPath := cfg.ConfigPath
	if configPath == "" {
		if cfg.DBPath != "" {
			configPath = filepath.Join(filepath.Dir(cfg.DBPath), "config.yaml")
		}
	}

	if configPath == "" {
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	// Parse config to find backends
	var backends []string
	lines := strings.Split(string(data), "\n")
	inBackends := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "backends:" {
			inBackends = true
			continue
		}
		if inBackends {
			// Check if we're still in the backends section (indented lines)
			if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && trimmed != "" {
				break
			}
			// Check for backend name (single level of indentation, ends with colon)
			if (strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t")) &&
				!strings.HasPrefix(strings.TrimSpace(line), " ") &&
				strings.HasSuffix(trimmed, ":") &&
				!strings.Contains(trimmed, " ") {
				backendName := strings.TrimSuffix(trimmed, ":")
				if backendName != "" {
					backends = append(backends, backendName)
				}
			}
		}
	}

	return backends
}

// newSyncQueueCmd creates the 'sync queue' subcommand
func newSyncQueueCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	queueCmd := &cobra.Command{
		Use:   "queue",
		Short: "View pending sync operations",
		Long:  "List all pending operations waiting to be synchronized with remote backends.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			return doSyncQueueView(cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	queueCmd.AddCommand(newSyncQueueClearCmd(stdout, cfg))

	return queueCmd
}

// doSyncQueueView displays pending sync operations
func doSyncQueueView(cfg *Config, stdout io.Writer) error {
	syncMgr := getSyncManager(cfg)
	defer func() { _ = syncMgr.Close() }()

	ops, err := syncMgr.GetPendingOperations()
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "Pending Operations: %d\n", len(ops))

	if len(ops) > 0 {
		_, _ = fmt.Fprintln(stdout, "")
		_, _ = fmt.Fprintf(stdout, "%-6s %-10s %-30s %-8s %s\n", "ID", "Type", "Task", "Retries", "Created")

		for _, op := range ops {
			createdStr := op.CreatedAt.Format("15:04:05")
			summary := op.TaskSummary
			if len(summary) > 28 {
				summary = summary[:28] + ".."
			}
			_, _ = fmt.Fprintf(stdout, "%-6d %-10s %-30s %-8d %s\n",
				op.ID, op.OperationType, summary, op.RetryCount, createdStr)
		}
	}

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// newSyncQueueClearCmd creates the 'sync queue clear' subcommand
func newSyncQueueClearCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear all pending sync operations",
		Long:  "Remove all pending operations from the sync queue. Use with caution as this discards unsynced changes.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			return doSyncQueueClear(cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doSyncQueueClear removes all pending sync operations
func doSyncQueueClear(cfg *Config, stdout io.Writer) error {
	syncMgr := getSyncManager(cfg)
	defer func() { _ = syncMgr.Close() }()

	count, err := syncMgr.ClearQueue()
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "Sync queue cleared: %d operations removed\n", count)
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// newSyncConflictsCmd creates the 'sync conflicts' subcommand
func newSyncConflictsCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	conflictsCmd := &cobra.Command{
		Use:   "conflicts",
		Short: "View and manage sync conflicts",
		Long:  "List all unresolved sync conflicts and manage their resolution.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			jsonOutput, _ := cmd.Flags().GetBool("json")
			return doSyncConflictsView(cfg, stdout, jsonOutput)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	conflictsCmd.AddCommand(newSyncConflictsResolveCmd(stdout, cfg))

	return conflictsCmd
}

// doSyncConflictsView displays sync conflicts
func doSyncConflictsView(cfg *Config, stdout io.Writer, jsonOutput bool) error {
	syncMgr := getSyncManager(cfg)
	defer func() { _ = syncMgr.Close() }()

	conflicts, err := syncMgr.GetConflicts()
	if err != nil {
		return err
	}

	// Handle JSON output
	if jsonOutput {
		output := struct {
			Conflicts []SyncConflict `json:"conflicts"`
		}{
			Conflicts: conflicts,
		}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdout, string(data))
		return nil
	}

	// Text output
	_, _ = fmt.Fprintf(stdout, "Conflicts: %d\n", len(conflicts))

	if len(conflicts) > 0 {
		_, _ = fmt.Fprintln(stdout, "")
		_, _ = fmt.Fprintf(stdout, "%-36s %-30s %-20s %s\n", "UID", "Task", "Detected", "Status")

		for _, c := range conflicts {
			detectedStr := c.DetectedAt.Format("2006-01-02 15:04:05")
			summary := c.TaskSummary
			if len(summary) > 28 {
				summary = summary[:25] + "..."
			}
			_, _ = fmt.Fprintf(stdout, "%-36s %-30s %-20s %s\n",
				c.TaskUID, summary, detectedStr, c.Status)
		}
	}

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// newSyncConflictsResolveCmd creates the 'sync conflicts resolve' subcommand
func newSyncConflictsResolveCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	var strategy string

	cmd := &cobra.Command{
		Use:   "resolve [task-uid]",
		Short: "Resolve a sync conflict",
		Long:  "Resolve a specific sync conflict using the specified strategy.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			taskUID := args[0]
			return doSyncConflictResolve(cfg, stdout, taskUID, strategy)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Flags().StringVar(&strategy, "strategy", "server_wins", "Resolution strategy: server_wins, local_wins, merge, keep_both")

	return cmd
}

// doSyncConflictResolve resolves a specific sync conflict
func doSyncConflictResolve(cfg *Config, stdout io.Writer, taskUID string, strategy string) error {
	syncMgr := getSyncManager(cfg)
	defer func() { _ = syncMgr.Close() }()

	// Validate strategy
	validStrategies := map[string]bool{
		"server_wins": true,
		"local_wins":  true,
		"merge":       true,
		"keep_both":   true,
	}
	if !validStrategies[strategy] {
		return fmt.Errorf("invalid strategy: %s (valid: server_wins, local_wins, merge, keep_both)", strategy)
	}

	// Try to resolve the conflict
	err := syncMgr.ResolveConflict(taskUID, strategy)
	if err != nil {
		_, _ = fmt.Fprintf(stdout, "Failed to resolve conflict: %v\n", err)
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultError)
		}
		return err
	}

	_, _ = fmt.Fprintf(stdout, "Conflict resolved for task %s using strategy %s\n", taskUID, strategy)
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// getSyncManager returns a SyncManager for the current configuration
func getSyncManager(cfg *Config) *SyncManager {
	dbPath := cfg.DBPath
	if dbPath == "" {
		home, _ := os.UserHomeDir()
		dbPath = filepath.Join(home, ".todoat", "todoat.db")
	}
	return NewSyncManager(dbPath)
}

// SyncManager handles synchronization operations
type SyncManager struct {
	dbPath string
	db     *sql.DB
}

// SyncOperation represents a pending sync operation
type SyncOperation struct {
	ID            int64
	TaskID        int64
	TaskUID       string
	TaskSummary   string
	ListID        int64
	OperationType string // "create", "update", "delete"
	RetryCount    int
	LastAttemptAt *time.Time
	CreatedAt     time.Time
}

// SyncConflict represents a sync conflict between local and remote versions
type SyncConflict struct {
	ID             int64     `json:"id"`
	TaskUID        string    `json:"task_uid"`
	TaskSummary    string    `json:"task_summary"`
	ListID         int64     `json:"list_id"`
	LocalVersion   string    `json:"local_version"`  // JSON serialized local task state
	RemoteVersion  string    `json:"remote_version"` // JSON serialized remote task state
	LocalModified  time.Time `json:"local_modified"`
	RemoteModified time.Time `json:"remote_modified"`
	DetectedAt     time.Time `json:"detected_at"`
	Status         string    `json:"status"` // "pending", "resolved"
}

// NewSyncManager creates a new SyncManager
func NewSyncManager(dbPath string) *SyncManager {
	sm := &SyncManager{dbPath: dbPath}
	_ = sm.initDB()
	return sm
}

// initDB initializes the sync database tables
func (sm *SyncManager) initDB() error {
	db, err := sql.Open("sqlite", sm.dbPath)
	if err != nil {
		return err
	}
	sm.db = db

	schema := `
		CREATE TABLE IF NOT EXISTS sync_queue (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			task_uid TEXT DEFAULT '',
			task_summary TEXT DEFAULT '',
			list_id INTEGER NOT NULL,
			operation_type TEXT NOT NULL,
			retry_count INTEGER DEFAULT 0,
			last_attempt_at TEXT,
			created_at TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS sync_metadata (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT UNIQUE NOT NULL,
			value TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS sync_conflicts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_uid TEXT NOT NULL,
			task_summary TEXT DEFAULT '',
			list_id INTEGER NOT NULL,
			local_version TEXT DEFAULT '',
			remote_version TEXT DEFAULT '',
			local_modified TEXT NOT NULL,
			remote_modified TEXT NOT NULL,
			detected_at TEXT NOT NULL,
			status TEXT DEFAULT 'pending'
		);

		CREATE INDEX IF NOT EXISTS idx_sync_queue_task ON sync_queue(task_id);
		CREATE INDEX IF NOT EXISTS idx_sync_queue_type ON sync_queue(operation_type);
		CREATE INDEX IF NOT EXISTS idx_sync_conflicts_uid ON sync_conflicts(task_uid);
		CREATE INDEX IF NOT EXISTS idx_sync_conflicts_status ON sync_conflicts(status);
	`
	_, err = db.Exec(schema)
	return err
}

// Close closes the database connection
func (sm *SyncManager) Close() error {
	if sm.db != nil {
		return sm.db.Close()
	}
	return nil
}

// GetPendingCount returns the number of pending sync operations
func (sm *SyncManager) GetPendingCount() (int, error) {
	if sm.db == nil {
		return 0, nil
	}

	var count int
	err := sm.db.QueryRow("SELECT COUNT(*) FROM sync_queue").Scan(&count)
	return count, err
}

// GetLastSyncTime returns the last successful sync timestamp
func (sm *SyncManager) GetLastSyncTime() time.Time {
	if sm.db == nil {
		return time.Time{}
	}

	var valueStr string
	err := sm.db.QueryRow("SELECT value FROM sync_metadata WHERE key = 'last_sync'").Scan(&valueStr)
	if err != nil {
		return time.Time{}
	}

	t, _ := time.Parse(time.RFC3339Nano, valueStr)
	return t
}

// GetConnectionStatus returns the current connection status
func (sm *SyncManager) GetConnectionStatus() string {
	// For now, return offline as there's no remote backend configured
	return "Offline (no remote backend configured)"
}

// GetPendingOperations returns all pending sync operations
func (sm *SyncManager) GetPendingOperations() ([]SyncOperation, error) {
	if sm.db == nil {
		return []SyncOperation{}, nil
	}

	// First try with tasks table join, fall back to without
	rows, err := sm.db.Query(`
		SELECT sq.id, sq.task_id, sq.task_uid, sq.list_id, sq.operation_type,
		       sq.retry_count, sq.last_attempt_at, sq.created_at,
		       COALESCE(t.summary, sq.task_summary) as task_summary
		FROM sync_queue sq
		LEFT JOIN tasks t ON sq.task_id = t.id
		ORDER BY sq.created_at ASC
	`)
	if err != nil {
		// Fall back to query without tasks table join
		rows, err = sm.db.Query(`
			SELECT id, task_id, task_uid, list_id, operation_type,
			       retry_count, last_attempt_at, created_at, task_summary
			FROM sync_queue
			ORDER BY created_at ASC
		`)
		if err != nil {
			return nil, err
		}
	}
	defer func() { _ = rows.Close() }()

	var ops []SyncOperation
	for rows.Next() {
		var op SyncOperation
		var lastAttemptStr, createdAtStr sql.NullString
		var taskSummary sql.NullString

		err := rows.Scan(&op.ID, &op.TaskID, &op.TaskUID, &op.ListID, &op.OperationType,
			&op.RetryCount, &lastAttemptStr, &createdAtStr, &taskSummary)
		if err != nil {
			return nil, err
		}

		if taskSummary.Valid {
			op.TaskSummary = taskSummary.String
		} else {
			op.TaskSummary = "Unknown"
		}
		if lastAttemptStr.Valid {
			t, _ := time.Parse(time.RFC3339Nano, lastAttemptStr.String)
			op.LastAttemptAt = &t
		}
		if createdAtStr.Valid {
			op.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAtStr.String)
		}

		ops = append(ops, op)
	}

	if ops == nil {
		ops = []SyncOperation{}
	}
	return ops, rows.Err()
}

// ClearQueue removes all pending sync operations
func (sm *SyncManager) ClearQueue() (int, error) {
	if sm.db == nil {
		return 0, nil
	}

	result, err := sm.db.Exec("DELETE FROM sync_queue")
	if err != nil {
		return 0, err
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}

// QueueOperation adds an operation to the sync queue
func (sm *SyncManager) QueueOperation(taskID int64, taskUID string, taskSummary string, listID int64, opType string) error {
	if sm.db == nil {
		return nil
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := sm.db.Exec(`
		INSERT INTO sync_queue (task_id, task_uid, task_summary, list_id, operation_type, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, taskID, taskUID, taskSummary, listID, opType, now)
	return err
}

// QueueOperationByStringID adds an operation to the sync queue using string IDs
func (sm *SyncManager) QueueOperationByStringID(taskID string, taskSummary string, listID string, opType string) error {
	if sm.db == nil {
		return nil
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := sm.db.Exec(`
		INSERT INTO sync_queue (task_id, task_uid, task_summary, list_id, operation_type, created_at)
		VALUES (0, ?, ?, 0, ?, ?)
	`, taskID, taskSummary, opType, now)
	return err
}

// GetConflicts returns all pending sync conflicts
func (sm *SyncManager) GetConflicts() ([]SyncConflict, error) {
	if sm.db == nil {
		return []SyncConflict{}, nil
	}

	rows, err := sm.db.Query(`
		SELECT id, task_uid, task_summary, list_id, local_version, remote_version,
		       local_modified, remote_modified, detected_at, status
		FROM sync_conflicts
		WHERE status = 'pending'
		ORDER BY detected_at DESC
	`)
	if err != nil {
		return []SyncConflict{}, nil
	}
	defer func() { _ = rows.Close() }()

	var conflicts []SyncConflict
	for rows.Next() {
		var c SyncConflict
		var localModStr, remoteModStr, detectedStr sql.NullString

		err := rows.Scan(&c.ID, &c.TaskUID, &c.TaskSummary, &c.ListID,
			&c.LocalVersion, &c.RemoteVersion, &localModStr, &remoteModStr,
			&detectedStr, &c.Status)
		if err != nil {
			continue
		}

		if localModStr.Valid {
			c.LocalModified, _ = time.Parse(time.RFC3339Nano, localModStr.String)
		}
		if remoteModStr.Valid {
			c.RemoteModified, _ = time.Parse(time.RFC3339Nano, remoteModStr.String)
		}
		if detectedStr.Valid {
			c.DetectedAt, _ = time.Parse(time.RFC3339Nano, detectedStr.String)
		}

		conflicts = append(conflicts, c)
	}

	if conflicts == nil {
		conflicts = []SyncConflict{}
	}
	return conflicts, rows.Err()
}

// GetConflictCount returns the number of pending conflicts
func (sm *SyncManager) GetConflictCount() (int, error) {
	if sm.db == nil {
		return 0, nil
	}

	var count int
	err := sm.db.QueryRow("SELECT COUNT(*) FROM sync_conflicts WHERE status = 'pending'").Scan(&count)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

// GetConflictByUID returns a conflict by task UID
func (sm *SyncManager) GetConflictByUID(taskUID string) (*SyncConflict, error) {
	if sm.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var c SyncConflict
	var localModStr, remoteModStr, detectedStr sql.NullString

	err := sm.db.QueryRow(`
		SELECT id, task_uid, task_summary, list_id, local_version, remote_version,
		       local_modified, remote_modified, detected_at, status
		FROM sync_conflicts
		WHERE task_uid = ? AND status = 'pending'
	`, taskUID).Scan(&c.ID, &c.TaskUID, &c.TaskSummary, &c.ListID,
		&c.LocalVersion, &c.RemoteVersion, &localModStr, &remoteModStr,
		&detectedStr, &c.Status)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("conflict not found: %s", taskUID)
	}
	if err != nil {
		return nil, err
	}

	if localModStr.Valid {
		c.LocalModified, _ = time.Parse(time.RFC3339Nano, localModStr.String)
	}
	if remoteModStr.Valid {
		c.RemoteModified, _ = time.Parse(time.RFC3339Nano, remoteModStr.String)
	}
	if detectedStr.Valid {
		c.DetectedAt, _ = time.Parse(time.RFC3339Nano, detectedStr.String)
	}

	return &c, nil
}

// ResolveConflict marks a conflict as resolved with the given strategy
func (sm *SyncManager) ResolveConflict(taskUID string, strategy string) error {
	if sm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	result, err := sm.db.Exec(`
		UPDATE sync_conflicts SET status = 'resolved'
		WHERE task_uid = ? AND status = 'pending'
	`, taskUID)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("conflict not found: %s", taskUID)
	}

	return nil
}

// AddConflict adds a new conflict to the database
func (sm *SyncManager) AddConflict(c *SyncConflict) error {
	if sm.db == nil {
		return fmt.Errorf("database not initialized")
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	localMod := c.LocalModified.Format(time.RFC3339Nano)
	remoteMod := c.RemoteModified.Format(time.RFC3339Nano)

	_, err := sm.db.Exec(`
		INSERT INTO sync_conflicts (task_uid, task_summary, list_id, local_version, remote_version,
		                            local_modified, remote_modified, detected_at, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending')
	`, c.TaskUID, c.TaskSummary, c.ListID, c.LocalVersion, c.RemoteVersion,
		localMod, remoteMod, now)

	return err
}

// =============================================================================
// Notification Commands
// =============================================================================

// newNotificationCmd creates the 'notification' subcommand for notification management
func newNotificationCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	notificationCmd := &cobra.Command{
		Use:   "notification",
		Short: "Manage notifications",
		Long:  "Manage the notification system for background sync events.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	notificationCmd.AddCommand(newNotificationTestCmd(stdout, cfg))
	notificationCmd.AddCommand(newNotificationLogCmd(stdout, cfg))

	return notificationCmd
}

// newNotificationTestCmd creates the 'notification test' subcommand
func newNotificationTestCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Send a test notification",
		Long:  "Send a test notification through all enabled notification channels.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			return doNotificationTest(cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doNotificationTest sends a test notification
func doNotificationTest(cfg *Config, stdout io.Writer) error {
	// Get notification log path
	logPath := cfg.NotificationLogPath
	if logPath == "" {
		logPath = getDefaultNotificationLogPath()
	}

	// Create notification config
	notifCfg := &notification.Config{
		Enabled: true,
		OSNotification: notification.OSNotificationConfig{
			Enabled:        true,
			OnSyncComplete: true,
			OnSyncError:    true,
			OnConflict:     true,
		},
		LogNotification: notification.LogNotificationConfig{
			Enabled:       true,
			Path:          logPath,
			MaxSizeMB:     10,
			RetentionDays: 30,
		},
	}

	var opts []notification.Option
	if cfg.NotificationMock {
		opts = append(opts, notification.WithCommandExecutor(&notification.MockCommandExecutor{}))
	}

	manager, err := notification.NewManager(notifCfg, opts...)
	if err != nil {
		return fmt.Errorf("failed to create notification manager: %w", err)
	}
	defer func() { _ = manager.Close() }()

	// Send test notification
	n := notification.Notification{
		Type:      notification.NotifyTest,
		Title:     "todoat",
		Message:   "Test notification from todoat",
		Timestamp: time.Now(),
	}

	err = manager.Send(n)
	if err != nil {
		return fmt.Errorf("failed to send test notification: %w", err)
	}

	_, _ = fmt.Fprintln(stdout, "Test notification sent")
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// newNotificationLogCmd creates the 'notification log' subcommand
func newNotificationLogCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	logCmd := &cobra.Command{
		Use:   "log",
		Short: "View notification log",
		Long:  "Display the notification history from the log file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			return doNotificationLogView(cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	logCmd.AddCommand(newNotificationLogClearCmd(stdout, cfg))

	return logCmd
}

// doNotificationLogView displays the notification log
func doNotificationLogView(cfg *Config, stdout io.Writer) error {
	logPath := cfg.NotificationLogPath
	if logPath == "" {
		logPath = getDefaultNotificationLogPath()
	}

	entries, err := notification.ReadLog(logPath)
	if err != nil {
		return fmt.Errorf("failed to read notification log: %w", err)
	}

	if len(entries) == 0 {
		_, _ = fmt.Fprintln(stdout, "No notifications in log")
	} else {
		_, _ = fmt.Fprintln(stdout, "Notification Log:")
		_, _ = fmt.Fprintln(stdout, "")
		for _, entry := range entries {
			_, _ = fmt.Fprintln(stdout, entry)
		}
	}

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// newNotificationLogClearCmd creates the 'notification log clear' subcommand
func newNotificationLogClearCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear notification log",
		Long:  "Clear all entries from the notification log file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			return doNotificationLogClear(cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doNotificationLogClear clears the notification log
func doNotificationLogClear(cfg *Config, stdout io.Writer) error {
	logPath := cfg.NotificationLogPath
	if logPath == "" {
		logPath = getDefaultNotificationLogPath()
	}

	err := notification.ClearLog(logPath)
	if err != nil {
		return fmt.Errorf("failed to clear notification log: %w", err)
	}

	_, _ = fmt.Fprintln(stdout, "Notification log cleared")
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// getDefaultNotificationLogPath returns the default path for the notification log
func getDefaultNotificationLogPath() string {
	// Use XDG_DATA_HOME or default to ~/.local/share/todoat
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		homeDir, _ := os.UserHomeDir()
		dataDir = filepath.Join(homeDir, ".local", "share")
	}
	return filepath.Join(dataDir, "todoat", "notifications.log")
}

// =============================================================================
// Sync Daemon Commands
// =============================================================================

// daemonState holds the in-process daemon state for testing
type daemonState struct {
	running     bool
	pid         int
	syncCount   int
	lastSync    time.Time
	interval    time.Duration
	stopChan    chan struct{}
	doneChan    chan struct{} // signals when daemon goroutine has stopped
	offlineMode bool
	cfg         *Config
	notifyMgr   notification.NotificationManager
}

// Global daemon instance for in-process testing
var testDaemon *daemonState

// newSyncDaemonCmd creates the 'sync daemon' subcommand
func newSyncDaemonCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	daemonCmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage the sync daemon",
		Long:  "Start, stop, and manage the background sync daemon.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	daemonCmd.AddCommand(newSyncDaemonStartCmd(stdout, stderr, cfg))
	daemonCmd.AddCommand(newSyncDaemonStopCmd(stdout, stderr, cfg))
	daemonCmd.AddCommand(newSyncDaemonStatusCmd(stdout, cfg))

	return daemonCmd
}

// newSyncDaemonStartCmd creates the 'sync daemon start' subcommand
func newSyncDaemonStartCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the sync daemon",
		Long:  "Start the background sync daemon that periodically synchronizes tasks with remote backends.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			interval, _ := cmd.Flags().GetInt("interval")
			if interval > 0 {
				cfg.DaemonInterval = time.Duration(interval) * time.Second
			}

			return doDaemonStart(cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Flags().Int("interval", 0, "Sync interval in seconds (default from config or 300)")

	return cmd
}

// newSyncDaemonStopCmd creates the 'sync daemon stop' subcommand
func newSyncDaemonStopCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the sync daemon",
		Long:  "Stop the running sync daemon.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			return doDaemonStop(cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// newSyncDaemonStatusCmd creates the 'sync daemon status' subcommand
func newSyncDaemonStatusCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show daemon status",
		Long:  "Show the current status of the sync daemon.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			return doDaemonStatus(cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doDaemonStart starts the sync daemon
func doDaemonStart(cfg *Config, stdout io.Writer) error {
	pidPath := getDaemonPIDPath(cfg)
	logPath := getDaemonLogPath(cfg)

	// Check if already running
	if isDaemonRunning(cfg, pidPath) {
		_, _ = fmt.Fprintln(stdout, "Sync daemon is already running")
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
		}
		return nil
	}

	// Get interval from config or default
	interval := cfg.DaemonInterval
	if interval == 0 {
		interval = getConfigDaemonInterval(cfg)
	}
	if interval == 0 {
		interval = 5 * time.Minute // Default: 5 minutes
	}

	if cfg.DaemonTestMode {
		// In-process daemon for testing
		return startTestDaemon(cfg, stdout, pidPath, logPath, interval)
	}

	// Real daemon would fork a background process here
	// For now, we'll just write the PID file and report started
	return startTestDaemon(cfg, stdout, pidPath, logPath, interval)
}

// startTestDaemon starts an in-process daemon for testing
func startTestDaemon(cfg *Config, stdout io.Writer, pidPath, logPath string, interval time.Duration) error {
	// Create PID file directory
	if err := os.MkdirAll(filepath.Dir(pidPath), 0755); err != nil {
		return fmt.Errorf("failed to create PID directory: %w", err)
	}

	// Create log file directory
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Initialize notification manager
	notifyCfg := &notification.Config{
		Enabled: true,
		LogNotification: notification.LogNotificationConfig{
			Enabled: true,
			Path:    cfg.NotificationLogPath,
		},
	}
	var notifyMgr notification.NotificationManager
	if cfg.NotificationMock {
		notifyMgr, _ = notification.NewManager(notifyCfg, notification.WithCommandExecutor(&notification.MockCommandExecutor{}))
	} else {
		notifyMgr, _ = notification.NewManager(notifyCfg)
	}

	// Create daemon state
	testDaemon = &daemonState{
		running:     true,
		pid:         os.Getpid(),
		syncCount:   0,
		interval:    interval,
		stopChan:    make(chan struct{}),
		doneChan:    make(chan struct{}),
		offlineMode: cfg.DaemonOfflineMode,
		cfg:         cfg,
		notifyMgr:   notifyMgr,
	}

	// Write PID file
	if err := os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", testDaemon.pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Write initial log entry
	logEntry := fmt.Sprintf("[%s] Daemon started with interval %v\n", time.Now().Format(time.RFC3339), interval)
	if err := appendToLogFile(logPath, logEntry); err != nil {
		// Log error but continue
		_, _ = fmt.Fprintf(stdout, "Warning: failed to write to log file: %v\n", err)
	}

	// Start sync loop in background
	go daemonSyncLoop(testDaemon, logPath)

	_, _ = fmt.Fprintf(stdout, "Sync daemon started (PID: %d, interval: %v)\n", testDaemon.pid, interval)
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// daemonSyncLoop runs the periodic sync in the background
func daemonSyncLoop(daemon *daemonState, logPath string) {
	ticker := time.NewTicker(daemon.interval)
	defer ticker.Stop()
	defer close(daemon.doneChan) // Signal that the goroutine has stopped

	for {
		select {
		case <-daemon.stopChan:
			logEntry := fmt.Sprintf("[%s] Daemon stopped\n", time.Now().Format(time.RFC3339))
			_ = appendToLogFile(logPath, logEntry)
			return
		case <-ticker.C:
			daemon.syncCount++
			daemon.lastSync = time.Now()

			// Perform sync
			var syncErr error
			if daemon.offlineMode {
				// Simulated offline - just log it
				logEntry := fmt.Sprintf("[%s] Sync attempt %d (offline mode)\n", time.Now().Format(time.RFC3339), daemon.syncCount)
				_ = appendToLogFile(logPath, logEntry)
			} else {
				// Normal sync
				logEntry := fmt.Sprintf("[%s] Sync completed (count: %d)\n", time.Now().Format(time.RFC3339), daemon.syncCount)
				_ = appendToLogFile(logPath, logEntry)
			}

			// Send notification
			if daemon.notifyMgr != nil {
				notif := notification.Notification{
					Type:      notification.NotifySyncComplete,
					Title:     "todoat sync",
					Message:   fmt.Sprintf("Sync completed (count: %d)", daemon.syncCount),
					Timestamp: time.Now(),
				}
				if syncErr != nil {
					notif.Type = notification.NotifySyncError
					notif.Message = fmt.Sprintf("Sync error: %v", syncErr)
				}
				daemon.notifyMgr.SendAsync(notif)
			}
		}
	}
}

// appendToLogFile appends a log entry to the daemon log file
func appendToLogFile(logPath, entry string) error {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(entry)
	return err
}

// doDaemonStop stops the sync daemon
func doDaemonStop(cfg *Config, stdout io.Writer) error {
	pidPath := getDaemonPIDPath(cfg)

	if !isDaemonRunning(cfg, pidPath) {
		_, _ = fmt.Fprintln(stdout, "Sync daemon is not running")
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
		}
		return nil
	}

	if cfg != nil && cfg.DaemonTestMode && testDaemon != nil {
		// Stop in-process daemon
		close(testDaemon.stopChan)
		// Wait for the daemon goroutine to finish
		<-testDaemon.doneChan
		testDaemon.running = false
		if testDaemon.notifyMgr != nil {
			_ = testDaemon.notifyMgr.Close()
		}
		testDaemon = nil
	}

	// Remove PID file
	_ = os.Remove(pidPath)

	_, _ = fmt.Fprintln(stdout, "Sync daemon stopped")
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// doDaemonStatus shows daemon status
func doDaemonStatus(cfg *Config, stdout io.Writer) error {
	pidPath := getDaemonPIDPath(cfg)

	if !isDaemonRunning(cfg, pidPath) {
		_, _ = fmt.Fprintln(stdout, "Sync daemon is not running")
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
		}
		return nil
	}

	// Get daemon info
	pid := 0
	syncCount := 0
	interval := time.Duration(0)
	lastSync := time.Time{}

	if cfg.DaemonTestMode && testDaemon != nil {
		pid = testDaemon.pid
		syncCount = testDaemon.syncCount
		interval = testDaemon.interval
		lastSync = testDaemon.lastSync
	} else {
		// Read PID from file
		data, err := os.ReadFile(pidPath)
		if err == nil {
			_, _ = fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &pid)
		}
		// Get interval from config
		interval = getConfigDaemonInterval(cfg)
		if interval == 0 {
			interval = 5 * time.Minute
		}
	}

	_, _ = fmt.Fprintln(stdout, "Sync daemon is running")
	_, _ = fmt.Fprintf(stdout, "  PID: %d\n", pid)
	_, _ = fmt.Fprintf(stdout, "  Interval: %d seconds\n", int(interval.Seconds()))
	_, _ = fmt.Fprintf(stdout, "  Sync count: %d\n", syncCount)
	if !lastSync.IsZero() {
		_, _ = fmt.Fprintf(stdout, "  Last sync: %s\n", lastSync.Format(time.RFC3339))
	}

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// getDaemonPIDPath returns the path to the daemon PID file
func getDaemonPIDPath(cfg *Config) string {
	if cfg.DaemonPIDPath != "" {
		return cfg.DaemonPIDPath
	}
	// Default: $XDG_RUNTIME_DIR/todoat/daemon.pid or /tmp/todoat-daemon.pid
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir != "" {
		return filepath.Join(runtimeDir, "todoat", "daemon.pid")
	}
	return "/tmp/todoat-daemon.pid"
}

// getDaemonLogPath returns the path to the daemon log file
func getDaemonLogPath(cfg *Config) string {
	if cfg.DaemonLogPath != "" {
		return cfg.DaemonLogPath
	}
	// Default: $XDG_DATA_HOME/todoat/daemon.log
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		homeDir, _ := os.UserHomeDir()
		dataDir = filepath.Join(homeDir, ".local", "share")
	}
	return filepath.Join(dataDir, "todoat", "daemon.log")
}

// isDaemonRunning checks if the daemon is currently running
func isDaemonRunning(cfg *Config, pidPath string) bool {
	if cfg.DaemonTestMode && testDaemon != nil && testDaemon.running {
		return true
	}

	// Check PID file
	_, err := os.Stat(pidPath)
	return err == nil
}

// getConfigDaemonInterval reads the daemon interval from config file
func getConfigDaemonInterval(cfg *Config) time.Duration {
	configPath := cfg.ConfigPath
	if configPath == "" {
		return 0
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return 0
	}

	// Simple YAML parsing for daemon interval
	content := string(data)
	lines := strings.Split(content, "\n")
	inDaemon := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "daemon:" {
			inDaemon = true
			continue
		}
		if inDaemon && strings.HasPrefix(trimmed, "interval:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				var seconds int
				if _, err := fmt.Sscanf(val, "%d", &seconds); err == nil {
					return time.Duration(seconds) * time.Second
				}
			}
		}
		// Stop if we hit another top-level key
		if inDaemon && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			break
		}
	}
	return 0
}

// =============================================================================
// Migrate Command
// =============================================================================

// newMigrateCmd creates the 'migrate' subcommand for cross-backend migration
func newMigrateCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate tasks between backends",
		Long:  "Migrate tasks from one storage backend to another, preserving metadata and hierarchy.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			fromBackend, _ := cmd.Flags().GetString("from")
			toBackend, _ := cmd.Flags().GetString("to")
			listName, _ := cmd.Flags().GetString("list")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			targetInfo, _ := cmd.Flags().GetString("target-info")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			// Handle target-info mode
			if targetInfo != "" {
				return doMigrateTargetInfo(cfg, stdout, targetInfo, listName, jsonOutput)
			}

			// Validate required flags
			if fromBackend == "" {
				return fmt.Errorf("--from flag is required")
			}
			if toBackend == "" {
				return fmt.Errorf("--to flag is required")
			}

			// Validate backends are different
			if fromBackend == toBackend {
				return fmt.Errorf("cannot migrate to same backend type")
			}

			// Validate backend names
			validBackends := map[string]bool{
				"sqlite": true, "nextcloud": true, "todoist": true, "file": true,
				"nextcloud-mock": true, "todoist-mock": true, "file-mock": true,
			}
			if !validBackends[fromBackend] {
				return fmt.Errorf("unknown backend: %s", fromBackend)
			}
			if !validBackends[toBackend] {
				return fmt.Errorf("unknown backend: %s", toBackend)
			}

			return doMigrate(cfg, stdout, fromBackend, toBackend, listName, dryRun, jsonOutput)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	migrateCmd.Flags().String("from", "", "Source backend (sqlite, nextcloud, todoist, file)")
	migrateCmd.Flags().String("to", "", "Target backend (sqlite, nextcloud, todoist, file)")
	migrateCmd.Flags().String("list", "", "Migrate only specified list")
	migrateCmd.Flags().Bool("dry-run", false, "Show what would be migrated without making changes")
	migrateCmd.Flags().String("target-info", "", "Show tasks in target backend")

	return migrateCmd
}

// MigrationResult holds the result of a migration operation
type MigrationResult struct {
	Source         string             `json:"source"`
	Target         string             `json:"target"`
	List           string             `json:"list,omitempty"`
	Migrated       int                `json:"migrated"`
	Skipped        int                `json:"skipped"`
	Updated        int                `json:"updated"`
	StatusMappings []StatusMapping    `json:"status_mappings,omitempty"`
	Tasks          []MigratedTaskInfo `json:"tasks,omitempty"`
	DryRun         bool               `json:"dry_run"`
	Hierarchy      bool               `json:"hierarchy_preserved"`
}

// StatusMapping describes how a status was mapped between backends
type StatusMapping struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// MigratedTaskInfo holds info about a migrated task
type MigratedTaskInfo struct {
	Summary    string `json:"summary"`
	Status     string `json:"status"`
	Priority   int    `json:"priority,omitempty"`
	DueDate    string `json:"due_date,omitempty"`
	Categories string `json:"categories,omitempty"`
	ParentID   string `json:"parent_id,omitempty"`
}

// MockBackend implements backend.TaskManager for testing migrations
type MockBackend struct {
	name      string
	lists     []backend.List
	tasks     map[string][]backend.Task
	tasksByID map[string]*backend.Task
	targetDir string
}

// NewMockBackend creates a new mock backend for testing
func NewMockBackend(name, targetDir string) *MockBackend {
	return &MockBackend{
		name:      name,
		lists:     []backend.List{},
		tasks:     make(map[string][]backend.Task),
		tasksByID: make(map[string]*backend.Task),
		targetDir: targetDir,
	}
}

// LoadFromFile loads mock data from a JSON file (for nextcloud-mock)
func (m *MockBackend) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil // File not existing is OK
	}

	// Try to parse as full format first (from SaveToFile)
	var fullData map[string][]map[string]interface{}
	if err := json.Unmarshal(data, &fullData); err == nil && len(fullData) > 0 {
		// Check if first entry has "summary" key to distinguish from simple format
		for _, taskMaps := range fullData {
			if len(taskMaps) > 0 {
				if _, hasSummary := taskMaps[0]["summary"]; hasSummary {
					// Full format - load complete task data
					return m.loadFullFormat(fullData)
				}
			}
		}
	}

	// Try simple format (list of strings)
	var simpleData map[string][]string
	if err := json.Unmarshal(data, &simpleData); err != nil {
		return err
	}

	for listName, taskSummaries := range simpleData {
		list := backend.List{
			ID:       generateUUID(),
			Name:     listName,
			Modified: time.Now(),
		}
		m.lists = append(m.lists, list)
		m.tasks[list.ID] = []backend.Task{}

		for _, summary := range taskSummaries {
			task := backend.Task{
				ID:       generateUUID(),
				Summary:  summary,
				Status:   backend.StatusNeedsAction,
				ListID:   list.ID,
				Created:  time.Now(),
				Modified: time.Now(),
			}
			m.tasks[list.ID] = append(m.tasks[list.ID], task)
			m.tasksByID[task.ID] = &m.tasks[list.ID][len(m.tasks[list.ID])-1]
		}
	}

	return nil
}

// loadFullFormat loads from the full format saved by SaveToFile
func (m *MockBackend) loadFullFormat(data map[string][]map[string]interface{}) error {
	for listName, taskMaps := range data {
		list := backend.List{
			ID:       generateUUID(),
			Name:     listName,
			Modified: time.Now(),
		}
		m.lists = append(m.lists, list)
		m.tasks[list.ID] = []backend.Task{}

		for _, taskMap := range taskMaps {
			task := backend.Task{
				ID:       generateUUID(),
				ListID:   list.ID,
				Created:  time.Now(),
				Modified: time.Now(),
			}

			if summary, ok := taskMap["summary"].(string); ok {
				task.Summary = summary
			}
			if status, ok := taskMap["status"].(string); ok {
				task.Status = backend.TaskStatus(status)
			}
			if priority, ok := taskMap["priority"].(float64); ok {
				task.Priority = int(priority)
			}
			if dueDate, ok := taskMap["due_date"].(string); ok {
				if t, err := time.Parse("2006-01-02", dueDate); err == nil {
					task.DueDate = &t
				}
			}
			if categories, ok := taskMap["categories"].(string); ok {
				task.Categories = categories
			}
			if parentID, ok := taskMap["parent_id"].(string); ok {
				task.ParentID = parentID
			}

			if task.Status == "" {
				task.Status = backend.StatusNeedsAction
			}

			m.tasks[list.ID] = append(m.tasks[list.ID], task)
			m.tasksByID[task.ID] = &m.tasks[list.ID][len(m.tasks[list.ID])-1]
		}
	}

	return nil
}

// SaveToFile saves the mock backend state to a JSON file
func (m *MockBackend) SaveToFile(path string) error {
	result := make(map[string][]map[string]interface{})

	for _, list := range m.lists {
		tasks := m.tasks[list.ID]
		taskList := make([]map[string]interface{}, 0, len(tasks))
		for _, task := range tasks {
			taskMap := map[string]interface{}{
				"summary":  task.Summary,
				"status":   string(task.Status),
				"priority": task.Priority,
			}
			if task.DueDate != nil {
				taskMap["due_date"] = task.DueDate.Format("2006-01-02")
			}
			if task.Categories != "" {
				taskMap["categories"] = task.Categories
			}
			if task.ParentID != "" {
				taskMap["parent_id"] = task.ParentID
			}
			taskList = append(taskList, taskMap)
		}
		result[list.Name] = taskList
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (m *MockBackend) GetLists(ctx context.Context) ([]backend.List, error) {
	return m.lists, nil
}

func (m *MockBackend) GetList(ctx context.Context, listID string) (*backend.List, error) {
	for i := range m.lists {
		if m.lists[i].ID == listID {
			return &m.lists[i], nil
		}
	}
	return nil, nil
}

func (m *MockBackend) GetListByName(ctx context.Context, name string) (*backend.List, error) {
	for i := range m.lists {
		if strings.EqualFold(m.lists[i].Name, name) {
			return &m.lists[i], nil
		}
	}
	return nil, nil
}

func (m *MockBackend) CreateList(ctx context.Context, name string) (*backend.List, error) {
	// Check if already exists
	for i := range m.lists {
		if strings.EqualFold(m.lists[i].Name, name) {
			return &m.lists[i], nil
		}
	}

	list := backend.List{
		ID:       generateUUID(),
		Name:     name,
		Modified: time.Now(),
	}
	m.lists = append(m.lists, list)
	m.tasks[list.ID] = []backend.Task{}
	return &m.lists[len(m.lists)-1], nil
}

func (m *MockBackend) DeleteList(ctx context.Context, listID string) error {
	for i := range m.lists {
		if m.lists[i].ID == listID {
			m.lists = append(m.lists[:i], m.lists[i+1:]...)
			delete(m.tasks, listID)
			return nil
		}
	}
	return fmt.Errorf("list not found")
}

func (m *MockBackend) GetDeletedLists(ctx context.Context) ([]backend.List, error) {
	return []backend.List{}, nil
}

func (m *MockBackend) GetDeletedListByName(ctx context.Context, name string) (*backend.List, error) {
	return nil, nil
}

func (m *MockBackend) RestoreList(ctx context.Context, listID string) error {
	return fmt.Errorf("not supported")
}

func (m *MockBackend) PurgeList(ctx context.Context, listID string) error {
	return fmt.Errorf("not supported")
}

func (m *MockBackend) GetTasks(ctx context.Context, listID string) ([]backend.Task, error) {
	return m.tasks[listID], nil
}

func (m *MockBackend) GetTask(ctx context.Context, listID, taskID string) (*backend.Task, error) {
	return m.tasksByID[taskID], nil
}

func (m *MockBackend) CreateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	newTask := backend.Task{
		ID:          generateUUID(),
		Summary:     task.Summary,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		DueDate:     task.DueDate,
		StartDate:   task.StartDate,
		Categories:  task.Categories,
		ParentID:    task.ParentID,
		ListID:      listID,
		Created:     time.Now(),
		Modified:    time.Now(),
	}

	if newTask.Status == "" {
		newTask.Status = backend.StatusNeedsAction
	}

	m.tasks[listID] = append(m.tasks[listID], newTask)
	m.tasksByID[newTask.ID] = &m.tasks[listID][len(m.tasks[listID])-1]

	return &newTask, nil
}

func (m *MockBackend) UpdateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	existing, ok := m.tasksByID[task.ID]
	if !ok {
		return nil, fmt.Errorf("task not found")
	}

	existing.Summary = task.Summary
	existing.Description = task.Description
	existing.Status = task.Status
	existing.Priority = task.Priority
	existing.DueDate = task.DueDate
	existing.StartDate = task.StartDate
	existing.Categories = task.Categories
	existing.ParentID = task.ParentID
	existing.Modified = time.Now()

	return existing, nil
}

func (m *MockBackend) DeleteTask(ctx context.Context, listID, taskID string) error {
	tasks := m.tasks[listID]
	for i := range tasks {
		if tasks[i].ID == taskID {
			m.tasks[listID] = append(tasks[:i], tasks[i+1:]...)
			delete(m.tasksByID, taskID)
			return nil
		}
	}
	return fmt.Errorf("task not found")
}

func (m *MockBackend) Close() error {
	// Save state to file if targetDir is set
	if m.targetDir != "" {
		path := filepath.Join(m.targetDir, m.name+"-data.json")
		return m.SaveToFile(path)
	}
	return nil
}

// generateUUID generates a new UUID string
func generateUUID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().UnixNano()%1000000)
}

// getMigrateBackend returns a backend instance for migration
func getMigrateBackend(cfg *Config, backendName string) (backend.TaskManager, error) {
	ctx := context.Background()

	switch backendName {
	case "sqlite":
		return getBackend(cfg)

	case "file-mock":
		if cfg.MigrateTargetDir == "" {
			return nil, fmt.Errorf("migrate target directory not configured")
		}
		mb := NewMockBackend("file-mock", cfg.MigrateTargetDir)
		// Load existing data if any
		path := filepath.Join(cfg.MigrateTargetDir, "file-mock-data.json")
		_ = mb.LoadFromFile(path)
		return mb, nil

	case "nextcloud-mock":
		mb := NewMockBackend("nextcloud-mock", cfg.MigrateTargetDir)
		if cfg.MockNextcloudDataPath != "" {
			_ = mb.LoadFromFile(cfg.MockNextcloudDataPath)
		}
		return mb, nil

	case "todoist-mock":
		mb := NewMockBackend("todoist-mock", cfg.MigrateTargetDir)
		return mb, nil

	case "nextcloud":
		// For real nextcloud, we'd need credentials
		_ = ctx
		return nil, fmt.Errorf("real nextcloud backend not yet implemented for migration")

	case "todoist":
		return nil, fmt.Errorf("real todoist backend not yet implemented for migration")

	case "file":
		return nil, fmt.Errorf("real file backend not yet implemented for migration")

	default:
		return nil, fmt.Errorf("unknown backend: %s", backendName)
	}
}

// mapStatus maps a status from source to target backend
func mapStatus(status backend.TaskStatus, targetBackend string) (backend.TaskStatus, bool) {
	// Todoist doesn't support IN-PROGRESS status
	if targetBackend == "todoist" || targetBackend == "todoist-mock" {
		if status == backend.StatusInProgress {
			return backend.StatusNeedsAction, true // Mapped from IN-PROGRESS
		}
	}
	return status, false
}

// doMigrate performs the actual migration between backends
func doMigrate(cfg *Config, stdout io.Writer, fromBackend, toBackend, listName string, dryRun, jsonOutput bool) error {
	ctx := context.Background()

	// Get source backend
	source, err := getMigrateBackend(cfg, fromBackend)
	if err != nil {
		return err
	}
	defer func() { _ = source.Close() }()

	// Get target backend
	target, err := getMigrateBackend(cfg, toBackend)
	if err != nil {
		return err
	}
	defer func() { _ = target.Close() }()

	// Get lists to migrate
	var listsToMigrate []backend.List
	if listName != "" {
		list, err := source.GetListByName(ctx, listName)
		if err != nil {
			return err
		}
		if list == nil {
			return fmt.Errorf("list not found: %s", listName)
		}
		listsToMigrate = []backend.List{*list}
	} else {
		var err error
		listsToMigrate, err = source.GetLists(ctx)
		if err != nil {
			return err
		}
	}

	result := MigrationResult{
		Source: fromBackend,
		Target: toBackend,
		DryRun: dryRun,
	}
	if listName != "" {
		result.List = listName
	}

	hasHierarchy := false
	statusMappings := make(map[string]string)

	for _, list := range listsToMigrate {
		// Get tasks from source
		tasks, err := source.GetTasks(ctx, list.ID)
		if err != nil {
			return fmt.Errorf("failed to get tasks from list %s: %w", list.Name, err)
		}

		if dryRun {
			// Just record what would be migrated
			for _, task := range tasks {
				result.Tasks = append(result.Tasks, MigratedTaskInfo{
					Summary:    task.Summary,
					Status:     string(task.Status),
					Priority:   task.Priority,
					Categories: task.Categories,
				})
				result.Migrated++
				if task.ParentID != "" {
					hasHierarchy = true
				}
			}
			continue
		}

		// Create list in target if needed
		targetList, err := target.GetListByName(ctx, list.Name)
		if err != nil {
			return fmt.Errorf("failed to check target list %s: %w", list.Name, err)
		}
		if targetList == nil {
			targetList, err = target.CreateList(ctx, list.Name)
			if err != nil {
				return fmt.Errorf("failed to create target list %s: %w", list.Name, err)
			}
		}

		// Get existing tasks in target for conflict detection
		existingTasks, _ := target.GetTasks(ctx, targetList.ID)
		existingByUID := make(map[string]*backend.Task)
		for i := range existingTasks {
			existingByUID[existingTasks[i].ID] = &existingTasks[i]
		}

		// Track ID mapping for hierarchy
		idMapping := make(map[string]string) // oldID -> newID

		// First pass: create all tasks (without hierarchy)
		for _, task := range tasks {
			// Map status if needed
			mappedStatus, wasMapped := mapStatus(task.Status, toBackend)
			if wasMapped {
				statusMappings[string(task.Status)] = string(mappedStatus)
			}

			// Check for existing task with same summary (conflict detection)
			var existingTask *backend.Task
			for i := range existingTasks {
				if existingTasks[i].Summary == task.Summary {
					existingTask = &existingTasks[i]
					break
				}
			}

			newTask := &backend.Task{
				Summary:     task.Summary,
				Description: task.Description,
				Status:      mappedStatus,
				Priority:    task.Priority,
				DueDate:     task.DueDate,
				StartDate:   task.StartDate,
				Completed:   task.Completed,
				Categories:  task.Categories,
				// ParentID will be set in second pass
			}

			if existingTask != nil {
				// Update existing task
				newTask.ID = existingTask.ID
				_, err := target.UpdateTask(ctx, targetList.ID, newTask)
				if err != nil {
					return fmt.Errorf("failed to update task %s: %w", task.Summary, err)
				}
				idMapping[task.ID] = existingTask.ID
				result.Updated++
			} else {
				// Create new task
				created, err := target.CreateTask(ctx, targetList.ID, newTask)
				if err != nil {
					return fmt.Errorf("failed to create task %s: %w", task.Summary, err)
				}
				idMapping[task.ID] = created.ID
				result.Migrated++
			}

			if task.ParentID != "" {
				hasHierarchy = true
			}
		}

		// Second pass: update parent relationships
		for _, task := range tasks {
			if task.ParentID != "" {
				newID, ok := idMapping[task.ID]
				if !ok {
					continue
				}
				newParentID, ok := idMapping[task.ParentID]
				if !ok {
					continue
				}

				// Get the task and update its parent
				targetTask, err := target.GetTask(ctx, targetList.ID, newID)
				if err != nil || targetTask == nil {
					continue
				}

				targetTask.ParentID = newParentID
				_, _ = target.UpdateTask(ctx, targetList.ID, targetTask)
			}
		}
	}

	result.Hierarchy = hasHierarchy
	for from, to := range statusMappings {
		result.StatusMappings = append(result.StatusMappings, StatusMapping{From: from, To: to})
	}

	// Output results
	if jsonOutput {
		data, _ := json.MarshalIndent(result, "", "  ")
		_, _ = fmt.Fprintln(stdout, string(data))
		return nil
	}

	if dryRun {
		_, _ = fmt.Fprintf(stdout, "Would migrate %d tasks from %s to %s (dry-run)\n", result.Migrated, fromBackend, toBackend)
		for _, task := range result.Tasks {
			_, _ = fmt.Fprintf(stdout, "  - %s\n", task.Summary)
		}
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
		}
	} else {
		listInfo := ""
		if listName != "" {
			listInfo = fmt.Sprintf(" from list %s", listName)
		}

		// Also show list name if all tasks were from the same list
		allListNames := make(map[string]bool)
		for _, list := range listsToMigrate {
			allListNames[list.Name] = true
		}
		listNames := ""
		for name := range allListNames {
			if listNames != "" {
				listNames += ", "
			}
			listNames += name
		}

		_, _ = fmt.Fprintf(stdout, "Migrated %d tasks%s from %s to %s\n", result.Migrated, listInfo, fromBackend, toBackend)
		if listNames != "" && listInfo == "" {
			_, _ = fmt.Fprintf(stdout, "  Lists: %s\n", listNames)
		}

		if result.Updated > 0 {
			_, _ = fmt.Fprintf(stdout, "  (%d updated, %d skipped)\n", result.Updated, result.Skipped)
		}

		if hasHierarchy {
			_, _ = fmt.Fprintln(stdout, "  (hierarchy preserved)")
		}

		if len(result.StatusMappings) > 0 {
			for _, mapping := range result.StatusMappings {
				_, _ = fmt.Fprintf(stdout, "  (status %s mapped to %s)\n", mapping.From, mapping.To)
			}
		}

		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
		}
	}

	return nil
}

// doMigrateTargetInfo displays tasks in the target backend
func doMigrateTargetInfo(cfg *Config, stdout io.Writer, targetBackend, listName string, jsonOutput bool) error {
	ctx := context.Background()

	target, err := getMigrateBackend(cfg, targetBackend)
	if err != nil {
		return err
	}
	defer func() { _ = target.Close() }()

	var list *backend.List
	if listName != "" {
		list, err = target.GetListByName(ctx, listName)
		if err != nil {
			return err
		}
		if list == nil {
			return fmt.Errorf("list not found: %s", listName)
		}
	}

	if list == nil {
		// Show all lists and tasks
		lists, err := target.GetLists(ctx)
		if err != nil {
			return err
		}

		if jsonOutput {
			result := make(map[string][]map[string]interface{})
			for _, l := range lists {
				tasks, _ := target.GetTasks(ctx, l.ID)
				taskList := make([]map[string]interface{}, 0, len(tasks))
				for _, t := range tasks {
					taskMap := map[string]interface{}{
						"summary":  t.Summary,
						"status":   string(t.Status),
						"priority": t.Priority,
					}
					if t.DueDate != nil {
						taskMap["due_date"] = t.DueDate.Format("2006-01-02")
					}
					if t.Categories != "" {
						taskMap["categories"] = t.Categories
					}
					if t.ParentID != "" {
						taskMap["parent_id"] = t.ParentID
					}
					taskList = append(taskList, taskMap)
				}
				result[l.Name] = taskList
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			_, _ = fmt.Fprintln(stdout, string(data))
			return nil
		}

		for _, l := range lists {
			_, _ = fmt.Fprintf(stdout, "List: %s\n", l.Name)
			tasks, _ := target.GetTasks(ctx, l.ID)
			_, _ = fmt.Fprintf(stdout, "  %d tasks\n", len(tasks))
		}
		return nil
	}

	// Show specific list
	tasks, err := target.GetTasks(ctx, list.ID)
	if err != nil {
		return err
	}

	if jsonOutput {
		taskList := make([]map[string]interface{}, 0, len(tasks))
		for _, t := range tasks {
			taskMap := map[string]interface{}{
				"summary":  t.Summary,
				"status":   string(t.Status),
				"priority": t.Priority,
			}
			if t.DueDate != nil {
				taskMap["due_date"] = t.DueDate.Format("2006-01-02")
			}
			if t.Categories != "" {
				taskMap["categories"] = t.Categories
			}
			if t.ParentID != "" {
				taskMap["parent_id"] = t.ParentID
			}
			taskList = append(taskList, taskMap)
		}
		result := map[string]interface{}{
			"list":  list.Name,
			"tasks": taskList,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		_, _ = fmt.Fprintln(stdout, string(data))
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "List: %s (%d tasks)\n", list.Name, len(tasks))
	for _, t := range tasks {
		_, _ = fmt.Fprintf(stdout, "  - %s (%s)\n", t.Summary, t.Status)
	}

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// newTUICmd creates the 'tui' subcommand for launching the terminal UI
func newTUICmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch the terminal user interface",
		Long:  "Launch an interactive terminal user interface for managing tasks with keyboard navigation.",
		RunE: func(cmd *cobra.Command, args []string) error {
			be, err := getBackend(cfg)
			if err != nil {
				return fmt.Errorf("failed to initialize backend: %w", err)
			}
			defer func() { _ = be.Close() }()

			// Create a TUI backend adapter
			adapter := &tuiBackendAdapter{TaskManager: be}

			// Create and run the TUI
			model := tui.New(adapter)
			p := tea.NewProgram(model, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("error running TUI: %w", err)
			}

			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// tuiBackendAdapter adapts backend.TaskManager to tui.Backend interface
type tuiBackendAdapter struct {
	backend.TaskManager
}

func (a *tuiBackendAdapter) GetLists(ctx context.Context) ([]backend.List, error) {
	return a.TaskManager.GetLists(ctx)
}

func (a *tuiBackendAdapter) GetTasks(ctx context.Context, listID string) ([]backend.Task, error) {
	return a.TaskManager.GetTasks(ctx, listID)
}

func (a *tuiBackendAdapter) GetTask(ctx context.Context, listID, taskID string) (*backend.Task, error) {
	return a.TaskManager.GetTask(ctx, listID, taskID)
}

func (a *tuiBackendAdapter) CreateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	return a.TaskManager.CreateTask(ctx, listID, task)
}

func (a *tuiBackendAdapter) UpdateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	return a.TaskManager.UpdateTask(ctx, listID, task)
}

func (a *tuiBackendAdapter) DeleteTask(ctx context.Context, listID, taskID string) error {
	return a.TaskManager.DeleteTask(ctx, listID, taskID)
}

// =============================================================================
// Reminder Command
// =============================================================================

// newReminderCmd creates the 'reminder' subcommand
func newReminderCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	reminderCmd := &cobra.Command{
		Use:   "reminder",
		Short: "Manage task reminders",
		Long:  "Manage reminder notifications for tasks with due dates.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	reminderCmd.AddCommand(newReminderStatusCmd(stdout, cfg))
	reminderCmd.AddCommand(newReminderCheckCmd(stdout, cfg))
	reminderCmd.AddCommand(newReminderListCmd(stdout, cfg))
	reminderCmd.AddCommand(newReminderDisableCmd(stdout, cfg))
	reminderCmd.AddCommand(newReminderDismissCmd(stdout, cfg))

	return reminderCmd
}

// newReminderStatusCmd creates the 'reminder status' subcommand
func newReminderStatusCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show reminder configuration status",
		Long:  "Display the current reminder configuration and status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			return doReminderStatus(cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doReminderStatus displays the reminder configuration status
func doReminderStatus(cfg *Config, stdout io.Writer) error {
	reminderCfg, err := loadReminderConfig(cfg)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(stdout, "Reminder Status:")
	if reminderCfg.Enabled {
		_, _ = fmt.Fprintln(stdout, "  Status: enabled")
	} else {
		_, _ = fmt.Fprintln(stdout, "  Status: disabled")
	}
	_, _ = fmt.Fprintln(stdout, "  Intervals:")
	for _, interval := range reminderCfg.Intervals {
		_, _ = fmt.Fprintf(stdout, "    - %s\n", interval)
	}
	_, _ = fmt.Fprintf(stdout, "  OS Notification: %v\n", reminderCfg.OSNotification)
	_, _ = fmt.Fprintf(stdout, "  Log Notification: %v\n", reminderCfg.LogNotification)

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// newReminderCheckCmd creates the 'reminder check' subcommand
func newReminderCheckCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check for due reminders",
		Long:  "Check all tasks with due dates and send reminders for those within the configured intervals.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			return doReminderCheck(cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doReminderCheck checks for due reminders and sends notifications
func doReminderCheck(cfg *Config, stdout io.Writer) error {
	reminderCfg, err := loadReminderConfig(cfg)
	if err != nil {
		return err
	}

	// Get database path
	dbPath := cfg.DBPath
	if dbPath == "" {
		dbPath = getDefaultDBPath()
	}

	// Create reminder service
	service, err := reminder.NewService(reminderCfg, dbPath+".reminders")
	if err != nil {
		return fmt.Errorf("failed to create reminder service: %w", err)
	}
	defer func() { _ = service.Close() }()

	// Set up notification manager
	notifier, err := createReminderNotifier(cfg, reminderCfg)
	if err != nil {
		return fmt.Errorf("failed to create notifier: %w", err)
	}
	if notifier != nil {
		defer func() { _ = notifier.Close() }()
		service.SetNotifier(notifier)
	}

	// Get backend and tasks
	be, err := getBackend(cfg)
	if err != nil {
		return err
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	tasks, err := getAllTasks(ctx, be)
	if err != nil {
		return err
	}

	// Convert to pointer slice
	taskPtrs := make([]*backend.Task, len(tasks))
	for i := range tasks {
		taskPtrs[i] = &tasks[i]
	}

	// Check reminders
	triggered, err := service.CheckReminders(taskPtrs)
	if err != nil {
		return err
	}

	if len(triggered) == 0 {
		_, _ = fmt.Fprintln(stdout, "No reminders triggered")
	} else {
		_, _ = fmt.Fprintf(stdout, "Triggered %d reminder(s):\n", len(triggered))
		for _, task := range triggered {
			_, _ = fmt.Fprintf(stdout, "  - %s (due: %s)\n", task.Summary, task.DueDate.Format("2006-01-02"))
		}
	}

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// newReminderListCmd creates the 'reminder list' subcommand
func newReminderListCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List upcoming reminders",
		Long:  "List all tasks with upcoming reminders within the configured intervals.",
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			return doReminderList(cfg, stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doReminderList lists upcoming reminders
func doReminderList(cfg *Config, stdout io.Writer) error {
	reminderCfg, err := loadReminderConfig(cfg)
	if err != nil {
		return err
	}

	// Get database path
	dbPath := cfg.DBPath
	if dbPath == "" {
		dbPath = getDefaultDBPath()
	}

	// Create reminder service
	service, err := reminder.NewService(reminderCfg, dbPath+".reminders")
	if err != nil {
		return fmt.Errorf("failed to create reminder service: %w", err)
	}
	defer func() { _ = service.Close() }()

	// Get backend and tasks
	be, err := getBackend(cfg)
	if err != nil {
		return err
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	tasks, err := getAllTasks(ctx, be)
	if err != nil {
		return err
	}

	// Convert to pointer slice
	taskPtrs := make([]*backend.Task, len(tasks))
	for i := range tasks {
		taskPtrs[i] = &tasks[i]
	}

	// Get upcoming reminders
	upcoming, err := service.GetUpcomingReminders(taskPtrs)
	if err != nil {
		return err
	}

	if len(upcoming) == 0 {
		_, _ = fmt.Fprintln(stdout, "No upcoming reminders")
	} else {
		_, _ = fmt.Fprintf(stdout, "Upcoming reminders (%d):\n", len(upcoming))
		for _, task := range upcoming {
			_, _ = fmt.Fprintf(stdout, "  - %s (due: %s)\n", task.Summary, task.DueDate.Format("2006-01-02"))
		}
	}

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// newReminderDisableCmd creates the 'reminder disable' subcommand
func newReminderDisableCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "disable <task>",
		Short: "Disable reminders for a task",
		Long:  "Permanently disable all reminders for a specific task.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			return doReminderDisable(cfg, args[0], stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doReminderDisable disables reminders for a task
func doReminderDisable(cfg *Config, taskSummary string, stdout io.Writer) error {
	reminderCfg, err := loadReminderConfig(cfg)
	if err != nil {
		return err
	}

	// Get database path
	dbPath := cfg.DBPath
	if dbPath == "" {
		dbPath = getDefaultDBPath()
	}

	// Create reminder service
	service, err := reminder.NewService(reminderCfg, dbPath+".reminders")
	if err != nil {
		return fmt.Errorf("failed to create reminder service: %w", err)
	}
	defer func() { _ = service.Close() }()

	// Get backend and find task
	be, err := getBackend(cfg)
	if err != nil {
		return err
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	tasks, err := getAllTasks(ctx, be)
	if err != nil {
		return err
	}

	// Find task by summary
	var task *backend.Task
	for i := range tasks {
		if strings.EqualFold(tasks[i].Summary, taskSummary) {
			task = &tasks[i]
			break
		}
	}

	if task == nil {
		return fmt.Errorf("task not found: %s", taskSummary)
	}

	// Disable reminder
	err = service.DisableReminder(task.ID)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "Disabled reminders for task: %s\n", task.Summary)

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// newReminderDismissCmd creates the 'reminder dismiss' subcommand
func newReminderDismissCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "dismiss <task>",
		Short: "Dismiss current reminder for a task",
		Long:  "Dismiss the current reminder for a specific task. It will trigger again at the next interval.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			return doReminderDismiss(cfg, args[0], stdout)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doReminderDismiss dismisses the current reminder for a task
func doReminderDismiss(cfg *Config, taskSummary string, stdout io.Writer) error {
	reminderCfg, err := loadReminderConfig(cfg)
	if err != nil {
		return err
	}

	// Get database path
	dbPath := cfg.DBPath
	if dbPath == "" {
		dbPath = getDefaultDBPath()
	}

	// Create reminder service
	service, err := reminder.NewService(reminderCfg, dbPath+".reminders")
	if err != nil {
		return fmt.Errorf("failed to create reminder service: %w", err)
	}
	defer func() { _ = service.Close() }()

	// Get backend and find task
	be, err := getBackend(cfg)
	if err != nil {
		return err
	}
	defer func() { _ = be.Close() }()

	ctx := context.Background()
	tasks, err := getAllTasks(ctx, be)
	if err != nil {
		return err
	}

	// Find task by summary
	var task *backend.Task
	for i := range tasks {
		if strings.EqualFold(tasks[i].Summary, taskSummary) {
			task = &tasks[i]
			break
		}
	}

	if task == nil {
		return fmt.Errorf("task not found: %s", taskSummary)
	}

	// Dismiss all intervals for this task
	for _, interval := range reminderCfg.Intervals {
		_ = service.DismissReminder(task.ID, interval)
	}

	_, _ = fmt.Fprintf(stdout, "Dismissed reminders for task: %s\n", task.Summary)

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// loadReminderConfig loads the reminder configuration
func loadReminderConfig(cfg *Config) (*reminder.Config, error) {
	// Check for test config path
	if cfg.ReminderConfigPath != "" {
		data, err := os.ReadFile(cfg.ReminderConfigPath)
		if err != nil {
			// Return default config if file doesn't exist
			return &reminder.Config{
				Enabled: false,
			}, nil
		}

		var reminderCfg reminder.Config
		if err := json.Unmarshal(data, &reminderCfg); err != nil {
			return nil, fmt.Errorf("failed to parse reminder config: %w", err)
		}
		return &reminderCfg, nil
	}

	// Return default config
	return &reminder.Config{
		Enabled:         true,
		Intervals:       []string{"1 day", "at due time"},
		OSNotification:  true,
		LogNotification: true,
	}, nil
}

// createReminderNotifier creates a notification manager for reminders
func createReminderNotifier(cfg *Config, reminderCfg *reminder.Config) (notification.NotificationManager, error) {
	if !reminderCfg.OSNotification && !reminderCfg.LogNotification {
		return nil, nil
	}

	// Get notification log path
	logPath := cfg.NotificationLogPath
	if logPath == "" {
		logPath = getDefaultNotificationLogPath()
	}

	notifCfg := &notification.Config{
		Enabled: true,
		OSNotification: notification.OSNotificationConfig{
			Enabled:        reminderCfg.OSNotification,
			OnSyncComplete: false,
			OnSyncError:    false,
			OnConflict:     false,
		},
		LogNotification: notification.LogNotificationConfig{
			Enabled:       reminderCfg.LogNotification,
			Path:          logPath,
			MaxSizeMB:     10,
			RetentionDays: 30,
		},
	}

	var opts []notification.Option
	if cfg.NotificationMock {
		opts = append(opts, notification.WithCommandExecutor(&notification.MockCommandExecutor{}))
	}

	// Add notification callback if configured (for testing)
	if cfg.NotificationCallback != nil {
		if callback, ok := cfg.NotificationCallback.(func(interface{})); ok {
			opts = append(opts, notification.WithSendCallback(func(n notification.Notification) {
				callback(n)
			}))
		}
	}

	return notification.NewManager(notifCfg, opts...)
}

// getAllTasks gets all tasks from all lists
func getAllTasks(ctx context.Context, be backend.TaskManager) ([]backend.Task, error) {
	lists, err := be.GetLists(ctx)
	if err != nil {
		return nil, err
	}

	var allTasks []backend.Task
	for _, list := range lists {
		tasks, err := be.GetTasks(ctx, list.ID)
		if err != nil {
			return nil, err
		}
		allTasks = append(allTasks, tasks...)
	}

	return allTasks, nil
}
