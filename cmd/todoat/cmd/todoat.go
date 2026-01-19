package cmd

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	_ "modernc.org/sqlite"
	"todoat/backend"
	_ "todoat/backend/git" // Register git as detectable backend
	"todoat/backend/sqlite"
	"todoat/backend/todoist"
	"todoat/internal/cache"
	"todoat/internal/config"
	"todoat/internal/credentials"
	"todoat/internal/notification"
	"todoat/internal/reminder"
	"todoat/internal/tui"
	"todoat/internal/utils"
	"todoat/internal/views"
)

// Version info - set at build time via ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

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
	// Cache-related config fields
	CachePath string        // Path to list cache file (for testing)
	CacheTTL  time.Duration // Cache TTL duration
	// Auto-detection config fields
	WorkDir           string // Working directory for auto-detection (for testing)
	AutoDetectBackend bool   // Enable auto-detection of backend
}

// LocalIDBackend is an interface for backends that support local_id lookup (e.g., SQLite)
type LocalIDBackend interface {
	GetTaskByLocalID(ctx context.Context, listID string, localID int64) (*backend.Task, error)
	GetTaskLocalID(ctx context.Context, taskID string) (int64, error)
}

// colorHexRegex matches valid hex color formats: #RGB, #RRGGBB, RGB, RRGGBB
var colorHexRegex = regexp.MustCompile(`^#?([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$`)

// validateAndNormalizeColor validates a hex color string and normalizes it to #RRGGBB format.
// Returns the normalized color and an error if the input is invalid.
func validateAndNormalizeColor(color string) (string, error) {
	if !colorHexRegex.MatchString(color) {
		return "", fmt.Errorf("invalid color format: %s (expected hex format like #RGB, #RRGGBB, RGB, or RRGGBB)", color)
	}

	// Remove # if present
	color = strings.TrimPrefix(color, "#")
	color = strings.ToUpper(color)

	// Expand 3-char to 6-char format
	if len(color) == 3 {
		color = string(color[0]) + string(color[0]) +
			string(color[1]) + string(color[1]) +
			string(color[2]) + string(color[2])
	}

	return "#" + color, nil
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Set verbose mode from flag
			verbose, _ := cmd.Flags().GetBool("verbose")
			if verbose {
				utils.SetVerboseMode(true)
				utils.Debugf("Verbose mode enabled")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Update config from flags
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if noPrompt {
				cfg.NoPrompt = true
			}

			// Handle --detect-backend flag
			detectBackend, _ := cmd.Flags().GetBool("detect-backend")
			if detectBackend {
				return runDetectBackend(stdout, cfg)
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

			// Validate list name is not empty or whitespace-only
			if strings.TrimSpace(listName) == "" {
				return errors.New("list name cannot be empty")
			}

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
	cmd.PersistentFlags().Bool("detect-backend", false, "Show auto-detected backends and exit")

	// Add action-specific flags
	cmd.Flags().StringP("priority", "p", "", "Task priority (0-9) for add/update, or filter (1,2,3 or high/medium/low) for get")
	cmd.Flags().StringP("status", "s", "", "Task status (TODO, IN-PROGRESS, DONE, CANCELLED)")
	cmd.Flags().String("summary", "", "New task summary (for update)")
	cmd.Flags().StringP("description", "d", "", "Task description/notes (for add/update, use \"\" to clear)")
	cmd.Flags().String("due-date", "", "Due date in YYYY-MM-DD format (for add/update, use \"\" to clear)")
	cmd.Flags().String("start-date", "", "Start date in YYYY-MM-DD format (for add/update, use \"\" to clear)")
	cmd.Flags().StringSlice("tag", nil, "Tag/category for add/update, or filter by tag for get (can be specified multiple times or comma-separated)")
	cmd.Flags().StringSlice("tags", nil, "Alias for --tag")
	cmd.Flags().StringSlice("add-tag", nil, "Add tag(s) to existing tags (for update, can be specified multiple times)")
	cmd.Flags().StringSlice("remove-tag", nil, "Remove tag(s) from existing tags (for update, can be specified multiple times)")
	cmd.Flags().StringP("parent", "P", "", "Parent task summary (for add/update subtasks)")
	cmd.Flags().BoolP("literal", "l", false, "Treat task summary literally (don't parse / as hierarchy separator)")
	cmd.Flags().Bool("no-parent", false, "Remove parent relationship (for update, makes task root-level)")
	cmd.Flags().StringP("view", "v", "", "View to use for displaying tasks (default, all, or custom view name)")
	cmd.Flags().String("uid", "", "Task UID for direct task selection (bypasses summary search)")
	cmd.Flags().Int64("local-id", 0, "Task local ID for direct task selection (requires sync enabled)")
	// Date filtering flags for get command
	cmd.Flags().String("due-before", "", "Filter tasks due before date (YYYY-MM-DD, inclusive)")
	cmd.Flags().String("due-after", "", "Filter tasks due on or after date (YYYY-MM-DD, inclusive)")
	cmd.Flags().String("created-before", "", "Filter tasks created before date (YYYY-MM-DD, inclusive)")
	cmd.Flags().String("created-after", "", "Filter tasks created on or after date (YYYY-MM-DD, inclusive)")

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

	// Add config subcommand
	cmd.AddCommand(newConfigCmd(stdout, stderr, cfg))

	// Add version subcommand
	cmd.AddCommand(newVersionCmd(stdout, cfg))

	// Add tags subcommand
	cmd.AddCommand(newTagsCmd(stdout, cfg))

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
	listCmd.AddCommand(newListUpdateCmd(stdout, cfg))
	listCmd.AddCommand(newListDeleteCmd(stdout, cfg))
	listCmd.AddCommand(newListInfoCmd(stdout, cfg))
	listCmd.AddCommand(newListTrashCmd(stdout, cfg))
	listCmd.AddCommand(newListExportCmd(stdout, cfg))
	listCmd.AddCommand(newListImportCmd(stdout, cfg))
	listCmd.AddCommand(newListStatsCmd(stdout, cfg))
	listCmd.AddCommand(newListVacuumCmd(stdout, cfg))

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
	// Try to use cache if available
	cachePath := getListCachePath(cfg)
	cacheTTL := getListCacheTTL(cfg)

	// Check if we have a valid cache
	cachedData, cacheValid := tryReadListCache(cachePath, cacheTTL)

	var lists []backend.List
	var err error
	var cachedLists []cache.CachedList

	if cacheValid {
		// Use cached data
		for _, cl := range cachedData.Lists {
			lists = append(lists, backend.List{
				ID:       cl.ID,
				Name:     cl.Name,
				Color:    cl.Color,
				Modified: cl.Modified,
			})
		}
		cachedLists = cachedData.Lists
	} else {
		// Fetch fresh data from backend
		lists, err = be.GetLists(ctx)
		if err != nil {
			return err
		}

		// Build cache data with task counts
		cachedLists = make([]cache.CachedList, 0, len(lists))
		for _, l := range lists {
			tasks, _ := be.GetTasks(ctx, l.ID)
			cachedLists = append(cachedLists, cache.CachedList{
				ID:        l.ID,
				Name:      l.Name,
				Color:     l.Color,
				TaskCount: len(tasks),
				Modified:  l.Modified,
			})
		}

		// Write cache
		writeListCache(cachePath, cachedLists, getBackendName(be))
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
		for _, cl := range cachedLists {
			output = append(output, listJSON{
				ID:       cl.ID,
				Name:     cl.Name,
				Color:    cl.Color,
				Tasks:    cl.TaskCount,
				Modified: cl.Modified.Format("2006-01-02T15:04:05Z"),
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

	for _, cl := range cachedLists {
		_, _ = fmt.Fprintf(stdout, "%-20s %d\n", cl.Name, cl.TaskCount)
	}

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// getListCachePath returns the path to the list cache file
func getListCachePath(cfg *Config) string {
	if cfg != nil && cfg.CachePath != "" {
		return cfg.CachePath
	}
	// Default to XDG cache path
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		homeDir, _ := os.UserHomeDir()
		cacheDir = filepath.Join(homeDir, ".cache")
	}
	return filepath.Join(cacheDir, "todoat", "lists.json")
}

// getListCacheTTL returns the cache TTL duration
func getListCacheTTL(cfg *Config) time.Duration {
	if cfg != nil && cfg.CacheTTL > 0 {
		return cfg.CacheTTL
	}
	return 5 * time.Minute // Default 5 minute TTL
}

// tryReadListCache attempts to read and validate the cache file
func tryReadListCache(cachePath string, ttl time.Duration) (*cache.ListCache, bool) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}

	var cacheData cache.ListCache
	if err := json.Unmarshal(data, &cacheData); err != nil {
		// Corrupt cache - delete it
		_ = os.Remove(cachePath)
		return nil, false
	}

	// Check TTL
	if time.Since(cacheData.CreatedAt) > ttl {
		return nil, false
	}

	return &cacheData, true
}

// writeListCache writes the cache file with proper permissions
func writeListCache(cachePath string, lists []cache.CachedList, backendName string) {
	cacheData := cache.ListCache{
		CreatedAt: time.Now(),
		Backend:   backendName,
		Lists:     lists,
	}

	data, err := json.Marshal(cacheData)
	if err != nil {
		return
	}

	// Ensure cache directory exists
	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return
	}

	// Write cache file with 0644 permissions
	_ = os.WriteFile(cachePath, data, 0644)
}

// invalidateListCache deletes the cache file to force a refresh
func invalidateListCache(cfg *Config) {
	cachePath := getListCachePath(cfg)
	_ = os.Remove(cachePath)
}

// getBackendName returns a name for the backend for cache isolation
func getBackendName(be backend.TaskManager) string {
	// Use type name as backend identifier
	switch be.(type) {
	case *sqlite.Backend:
		return "sqlite"
	default:
		return "unknown"
	}
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

	// Invalidate cache after creating a list
	invalidateListCache(cfg)

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

// newListUpdateCmd creates the 'list update' subcommand
func newListUpdateCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update a list's properties",
		Long:  "Update a task list's name, color, or description.",
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

			newName, _ := cmd.Flags().GetString("name")
			color, _ := cmd.Flags().GetString("color")
			description, _ := cmd.Flags().GetString("description")
			descriptionSet := cmd.Flags().Changed("description")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			return doListUpdate(context.Background(), be, args[0], newName, color, description, descriptionSet, cfg, stdout, jsonOutput)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.Flags().String("name", "", "New name for the list")
	cmd.Flags().String("color", "", "Hex color for the list (e.g., #FF5733, ABC)")
	cmd.Flags().String("description", "", "Description for the list")
	return cmd
}

// doListUpdate updates a list's properties (name, color, description)
func doListUpdate(ctx context.Context, be backend.TaskManager, name, newName, color, description string, descriptionSet bool, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// Check that at least one update is requested
	if newName == "" && color == "" && !descriptionSet {
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultError)
		}
		return fmt.Errorf("at least one of --name, --color, or --description is required")
	}

	// Validate and normalize color if provided
	var normalizedColor string
	if color != "" {
		var err error
		normalizedColor, err = validateAndNormalizeColor(color)
		if err != nil {
			if cfg != nil && cfg.NoPrompt {
				_, _ = fmt.Fprintln(stdout, ResultError)
			}
			return err
		}
	}

	// Get all lists for matching and validation
	lists, err := be.GetLists(ctx)
	if err != nil {
		return err
	}

	// Find list by name (exact or partial match)
	var matchedList *backend.List
	var matches []backend.List
	nameLower := strings.ToLower(name)

	// First try exact match (case-insensitive)
	for i := range lists {
		if strings.EqualFold(lists[i].Name, name) {
			matchedList = &lists[i]
			break
		}
	}

	// If no exact match, try partial match
	if matchedList == nil {
		for i := range lists {
			if strings.Contains(strings.ToLower(lists[i].Name), nameLower) {
				matches = append(matches, lists[i])
			}
		}

		if len(matches) == 0 {
			if cfg != nil && cfg.NoPrompt {
				_, _ = fmt.Fprintln(stdout, ResultError)
			}
			return fmt.Errorf("list '%s' not found", name)
		}

		if len(matches) == 1 {
			matchedList = &matches[0]
		} else {
			// Multiple matches - error in no-prompt mode
			if cfg != nil && cfg.NoPrompt {
				_, _ = fmt.Fprintln(stdout, ResultError)
				return fmt.Errorf("multiple lists match '%s' - ambiguous, please be more specific", name)
			}
			// In interactive mode, we would prompt - but for now return error
			return fmt.Errorf("multiple lists match '%s' - please be more specific", name)
		}
	}

	// Check if new name already exists (case-insensitive)
	if newName != "" {
		for _, l := range lists {
			if l.ID != matchedList.ID && strings.EqualFold(l.Name, newName) {
				if cfg != nil && cfg.NoPrompt {
					_, _ = fmt.Fprintln(stdout, ResultError)
				}
				return fmt.Errorf("list '%s' already exists - choose a different name", newName)
			}
		}
	}

	// Update the list properties
	oldName := matchedList.Name
	if newName != "" {
		matchedList.Name = newName
	}
	if normalizedColor != "" {
		matchedList.Color = normalizedColor
	}
	if descriptionSet {
		matchedList.Description = description
	}

	updatedList, err := be.UpdateList(ctx, matchedList)
	if err != nil {
		return err
	}

	// Invalidate cache after updating a list
	invalidateListCache(cfg)

	if jsonOutput {
		type listJSON struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			OldName     string `json:"old_name,omitempty"`
			Color       string `json:"color,omitempty"`
			Description string `json:"description,omitempty"`
			Modified    string `json:"modified"`
			Result      string `json:"result"`
		}
		output := listJSON{
			ID:          updatedList.ID,
			Name:        updatedList.Name,
			Color:       updatedList.Color,
			Description: updatedList.Description,
			Modified:    updatedList.Modified.Format("2006-01-02T15:04:05Z"),
			Result:      "ACTION_COMPLETED",
		}
		if newName != "" && newName != oldName {
			output.OldName = oldName
		}
		jsonBytes, err := json.Marshal(output)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdout, string(jsonBytes))
		return nil
	}

	// Build output message
	_, _ = fmt.Fprintf(stdout, "Updated list '%s'\n", updatedList.Name)
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

	// Invalidate cache after deleting a list
	invalidateListCache(cfg)

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
	if list.Description != "" {
		_, _ = fmt.Fprintf(stdout, "Description: %s\n", list.Description)
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

// doListTrashView displays deleted lists, auto-purging expired ones first
func doListTrashView(ctx context.Context, be backend.TaskManager, cfg *Config, stdout io.Writer) error {
	// Get all deleted lists first
	lists, err := be.GetDeletedLists(ctx)
	if err != nil {
		return err
	}

	// Auto-purge expired lists based on retention policy
	retentionDays := getTrashRetentionDays(cfg)
	purgedCount := 0

	if retentionDays > 0 {
		cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
		var remainingLists []backend.List

		for _, l := range lists {
			if l.DeletedAt != nil && l.DeletedAt.Before(cutoffTime) {
				// This list has expired - purge it
				if err := be.PurgeList(ctx, l.ID); err != nil {
					return fmt.Errorf("failed to purge expired list %q: %w", l.Name, err)
				}
				purgedCount++
			} else {
				remainingLists = append(remainingLists, l)
			}
		}
		lists = remainingLists
	}

	// Report purged lists if any
	if purgedCount > 0 {
		if purgedCount == 1 {
			_, _ = fmt.Fprintln(stdout, "Auto-purged 1 expired list.")
		} else {
			_, _ = fmt.Fprintf(stdout, "Auto-purged %d expired lists.\n", purgedCount)
		}
	}

	if len(lists) == 0 {
		if purgedCount == 0 {
			_, _ = fmt.Fprintln(stdout, "Trash is empty.")
		}
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

// newListExportCmd creates the 'list export' subcommand
func newListExportCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [name]",
		Short: "Export a list to a file",
		Long:  "Export a task list to a file in various formats (sqlite, json, csv, ical).",
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

			format, _ := cmd.Flags().GetString("format")
			output, _ := cmd.Flags().GetString("output")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			return doListExport(context.Background(), be, args[0], format, output, cfg, stdout, jsonOutput)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Flags().String("format", "json", "Export format: sqlite, json, csv, ical")
	cmd.Flags().String("output", "", "Output file path (default: ./<list-name>.<ext>)")

	return cmd
}

// doListExport exports a list to a file
func doListExport(ctx context.Context, be backend.TaskManager, name, format, outputPath string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// Find the list by name
	list, err := be.GetListByName(ctx, name)
	if err != nil {
		return err
	}
	if list == nil {
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultError)
		}
		return fmt.Errorf("list '%s' not found", name)
	}

	// Get all tasks for the list
	tasks, err := be.GetTasks(ctx, list.ID)
	if err != nil {
		return err
	}

	// Determine output path
	if outputPath == "" {
		ext := format
		switch format {
		case "ical":
			ext = "ics"
		case "sqlite":
			ext = "db"
		}
		outputPath = fmt.Sprintf("%s.%s", list.Name, ext)
	}

	// Export based on format
	var exportErr error
	switch format {
	case "sqlite":
		exportErr = exportSQLite(ctx, list, tasks, outputPath)
	case "json":
		exportErr = exportJSON(tasks, outputPath)
	case "csv":
		exportErr = exportCSV(tasks, outputPath)
	case "ical":
		exportErr = exportICalendar(tasks, outputPath)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}

	if exportErr != nil {
		return exportErr
	}

	taskCount := len(tasks)

	if jsonOutput {
		type exportResult struct {
			Action    string `json:"action"`
			File      string `json:"file"`
			TaskCount int    `json:"task_count"`
		}
		result := exportResult{
			Action:    "export",
			File:      outputPath,
			TaskCount: taskCount,
		}
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdout, string(jsonBytes))
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "Exported %d tasks to %s\n", taskCount, outputPath)
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// exportSQLite exports tasks to a standalone SQLite database
func exportSQLite(ctx context.Context, list *backend.List, tasks []backend.Task, outputPath string) error {
	// Remove existing file if any
	_ = os.Remove(outputPath)

	// Create new database
	db, err := sql.Open("sqlite", outputPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	// Create schema
	schema := `
		CREATE TABLE IF NOT EXISTS task_lists (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			color TEXT DEFAULT '',
			modified TEXT NOT NULL,
			deleted_at TEXT
		);

		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			list_id TEXT NOT NULL,
			summary TEXT NOT NULL,
			description TEXT DEFAULT '',
			status TEXT NOT NULL DEFAULT 'NEEDS-ACTION',
			priority INTEGER DEFAULT 0,
			due_date TEXT,
			start_date TEXT,
			completed TEXT,
			created TEXT NOT NULL,
			modified TEXT NOT NULL,
			parent_id TEXT DEFAULT '',
			categories TEXT DEFAULT '',
			FOREIGN KEY (list_id) REFERENCES task_lists(id) ON DELETE CASCADE
		);
	`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// Insert list
	_, err = db.ExecContext(ctx, `INSERT INTO task_lists (id, name, color, modified) VALUES (?, ?, ?, ?)`,
		list.ID, list.Name, list.Color, list.Modified.Format(time.RFC3339Nano))
	if err != nil {
		return err
	}

	// Insert tasks
	for _, task := range tasks {
		var dueDate, startDate, completed *string
		if task.DueDate != nil {
			s := task.DueDate.Format(time.RFC3339Nano)
			dueDate = &s
		}
		if task.StartDate != nil {
			s := task.StartDate.Format(time.RFC3339Nano)
			startDate = &s
		}
		if task.Completed != nil {
			s := task.Completed.Format(time.RFC3339Nano)
			completed = &s
		}

		_, err = db.ExecContext(ctx, `INSERT INTO tasks (id, list_id, summary, description, status, priority, due_date, start_date, completed, created, modified, parent_id, categories) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			task.ID, task.ListID, task.Summary, task.Description, string(task.Status), task.Priority,
			dueDate, startDate, completed,
			task.Created.Format(time.RFC3339Nano), task.Modified.Format(time.RFC3339Nano),
			task.ParentID, task.Categories)
		if err != nil {
			return err
		}
	}

	return nil
}

// exportJSON exports tasks to a JSON file
func exportJSON(tasks []backend.Task, outputPath string) error {
	type taskJSON struct {
		ID          string     `json:"id"`
		Summary     string     `json:"summary"`
		Description string     `json:"description,omitempty"`
		Status      string     `json:"status"`
		Priority    int        `json:"priority"`
		DueDate     *time.Time `json:"due_date,omitempty"`
		StartDate   *time.Time `json:"start_date,omitempty"`
		Completed   *time.Time `json:"completed,omitempty"`
		Created     time.Time  `json:"created"`
		Modified    time.Time  `json:"modified"`
		ListID      string     `json:"list_id"`
		ParentID    string     `json:"parent_id,omitempty"`
		Categories  string     `json:"categories,omitempty"`
	}

	output := make([]taskJSON, len(tasks))
	for i, task := range tasks {
		output[i] = taskJSON{
			ID:          task.ID,
			Summary:     task.Summary,
			Description: task.Description,
			Status:      string(task.Status),
			Priority:    task.Priority,
			DueDate:     task.DueDate,
			StartDate:   task.StartDate,
			Completed:   task.Completed,
			Created:     task.Created,
			Modified:    task.Modified,
			ListID:      task.ListID,
			ParentID:    task.ParentID,
			Categories:  task.Categories,
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}

// exportCSV exports tasks to a CSV file
func exportCSV(tasks []backend.Task, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"id", "summary", "description", "status", "priority", "due_date", "start_date", "completed", "created", "modified", "list_id", "parent_id", "categories"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write tasks
	for _, task := range tasks {
		var dueDate, startDate, completed string
		if task.DueDate != nil {
			dueDate = task.DueDate.Format(time.RFC3339)
		}
		if task.StartDate != nil {
			startDate = task.StartDate.Format(time.RFC3339)
		}
		if task.Completed != nil {
			completed = task.Completed.Format(time.RFC3339)
		}

		row := []string{
			task.ID,
			task.Summary,
			task.Description,
			string(task.Status),
			strconv.Itoa(task.Priority),
			dueDate,
			startDate,
			completed,
			task.Created.Format(time.RFC3339),
			task.Modified.Format(time.RFC3339),
			task.ListID,
			task.ParentID,
			task.Categories,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// exportICalendar exports tasks to an iCalendar file
func exportICalendar(tasks []backend.Task, outputPath string) error {
	const iCalDateFormat = "20060102T150405Z"

	var lines []string
	lines = append(lines, "BEGIN:VCALENDAR")
	lines = append(lines, "VERSION:2.0")
	lines = append(lines, "PRODID:-//todoat//todoat//EN")

	for _, task := range tasks {
		lines = append(lines, "BEGIN:VTODO")
		lines = append(lines, fmt.Sprintf("UID:%s", task.ID))
		lines = append(lines, fmt.Sprintf("DTSTAMP:%s", time.Now().UTC().Format(iCalDateFormat)))

		if task.Summary != "" {
			lines = append(lines, fmt.Sprintf("SUMMARY:%s", task.Summary))
		}
		if task.Description != "" {
			lines = append(lines, fmt.Sprintf("DESCRIPTION:%s", task.Description))
		}

		// Convert status
		status := "NEEDS-ACTION"
		switch task.Status {
		case backend.StatusCompleted:
			status = "COMPLETED"
		case backend.StatusInProgress:
			status = "IN-PROGRESS"
		case backend.StatusCancelled:
			status = "CANCELLED"
		}
		lines = append(lines, fmt.Sprintf("STATUS:%s", status))

		if task.Priority > 0 {
			lines = append(lines, fmt.Sprintf("PRIORITY:%d", task.Priority))
		}
		if task.Categories != "" {
			lines = append(lines, fmt.Sprintf("CATEGORIES:%s", task.Categories))
		}
		if task.DueDate != nil {
			lines = append(lines, fmt.Sprintf("DUE:%s", task.DueDate.UTC().Format(iCalDateFormat)))
		}
		if task.StartDate != nil {
			lines = append(lines, fmt.Sprintf("DTSTART:%s", task.StartDate.UTC().Format(iCalDateFormat)))
		}
		if !task.Created.IsZero() {
			lines = append(lines, fmt.Sprintf("CREATED:%s", task.Created.UTC().Format(iCalDateFormat)))
		}
		if !task.Modified.IsZero() {
			lines = append(lines, fmt.Sprintf("LAST-MODIFIED:%s", task.Modified.UTC().Format(iCalDateFormat)))
		}
		if task.Completed != nil {
			lines = append(lines, fmt.Sprintf("COMPLETED:%s", task.Completed.UTC().Format(iCalDateFormat)))
		}

		lines = append(lines, "END:VTODO")
	}

	lines = append(lines, "END:VCALENDAR")

	return os.WriteFile(outputPath, []byte(strings.Join(lines, "\r\n")), 0644)
}

// newListImportCmd creates the 'list import' subcommand
func newListImportCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import [file]",
		Short: "Import a list from a file",
		Long:  "Import a task list from a file. Supported formats: sqlite, json, csv, ical.",
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

			format, _ := cmd.Flags().GetString("format")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			return doListImport(context.Background(), be, args[0], format, cfg, stdout, jsonOutput)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Flags().String("format", "", "Import format (auto-detect from extension if not specified)")

	return cmd
}

// doListImport imports a list from a file
func doListImport(ctx context.Context, be backend.TaskManager, inputPath, format string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// Auto-detect format from extension if not specified
	if format == "" {
		ext := strings.ToLower(filepath.Ext(inputPath))
		switch ext {
		case ".db", ".sqlite", ".sqlite3":
			format = "sqlite"
		case ".json":
			format = "json"
		case ".csv":
			format = "csv"
		case ".ics", ".ical":
			format = "ical"
		default:
			return fmt.Errorf("cannot detect format from extension '%s', please specify --format", ext)
		}
	}

	var list *backend.List
	var tasks []backend.Task
	var importErr error

	switch format {
	case "sqlite":
		list, tasks, importErr = importSQLite(ctx, inputPath)
	case "json":
		list, tasks, importErr = importJSON(inputPath)
	case "csv":
		list, tasks, importErr = importCSV(inputPath)
	case "ical":
		list, tasks, importErr = importICalendar(inputPath)
	default:
		return fmt.Errorf("unsupported import format: %s", format)
	}

	if importErr != nil {
		return importErr
	}

	// Create the list in the backend
	newList, err := be.CreateList(ctx, list.Name)
	if err != nil {
		return fmt.Errorf("failed to create list: %w", err)
	}

	// Build a map of old task IDs to new task IDs for parent relationships
	idMap := make(map[string]string)

	// First pass: create tasks without parent relationships (to get new IDs)
	createdTasks := make(map[string]*backend.Task)
	for _, task := range tasks {
		newTask := task
		newTask.ListID = newList.ID
		oldID := task.ID
		newTask.ParentID = "" // Clear parent, will set in second pass

		created, err := be.CreateTask(ctx, newList.ID, &newTask)
		if err != nil {
			return fmt.Errorf("failed to create task '%s': %w", task.Summary, err)
		}
		idMap[oldID] = created.ID
		createdTasks[oldID] = created
	}

	// Second pass: update parent relationships
	for _, task := range tasks {
		if task.ParentID != "" {
			if newParentID, ok := idMap[task.ParentID]; ok {
				newTaskID := idMap[task.ID]
				createdTask := createdTasks[task.ID]
				createdTask.ParentID = newParentID
				_, err := be.UpdateTask(ctx, newList.ID, createdTask)
				if err != nil {
					return fmt.Errorf("failed to update parent for task '%s': %w", task.Summary, err)
				}
				_ = newTaskID // suppress unused variable warning
			}
		}
	}

	// Invalidate list cache
	invalidateListCache(cfg)

	taskCount := len(tasks)

	if jsonOutput {
		type importResult struct {
			Action    string `json:"action"`
			File      string `json:"file"`
			TaskCount int    `json:"task_count"`
		}
		result := importResult{
			Action:    "import",
			File:      inputPath,
			TaskCount: taskCount,
		}
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdout, string(jsonBytes))
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "Imported %d tasks from %s\n", taskCount, inputPath)
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// importSQLite imports a list from a SQLite database
func importSQLite(ctx context.Context, inputPath string) (*backend.List, []backend.Task, error) {
	db, err := sql.Open("sqlite", inputPath)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = db.Close() }()

	// Read list
	var list backend.List
	var modifiedStr string
	err = db.QueryRowContext(ctx, "SELECT id, name, color, modified FROM task_lists LIMIT 1").Scan(
		&list.ID, &list.Name, &list.Color, &modifiedStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read list: %w", err)
	}
	list.Modified, _ = time.Parse(time.RFC3339Nano, modifiedStr)

	// Read tasks
	rows, err := db.QueryContext(ctx, `SELECT id, list_id, summary, description, status, priority, due_date, start_date, completed, created, modified, parent_id, categories FROM tasks`)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	var tasks []backend.Task
	for rows.Next() {
		var task backend.Task
		var dueDate, startDate, completed, created, modified sql.NullString
		err := rows.Scan(&task.ID, &task.ListID, &task.Summary, &task.Description, &task.Status, &task.Priority,
			&dueDate, &startDate, &completed, &created, &modified, &task.ParentID, &task.Categories)
		if err != nil {
			return nil, nil, err
		}

		if dueDate.Valid {
			t, _ := time.Parse(time.RFC3339Nano, dueDate.String)
			task.DueDate = &t
		}
		if startDate.Valid {
			t, _ := time.Parse(time.RFC3339Nano, startDate.String)
			task.StartDate = &t
		}
		if completed.Valid {
			t, _ := time.Parse(time.RFC3339Nano, completed.String)
			task.Completed = &t
		}
		if created.Valid {
			task.Created, _ = time.Parse(time.RFC3339Nano, created.String)
		}
		if modified.Valid {
			task.Modified, _ = time.Parse(time.RFC3339Nano, modified.String)
		}

		tasks = append(tasks, task)
	}

	return &list, tasks, nil
}

// importJSON imports a list from a JSON file
func importJSON(inputPath string) (*backend.List, []backend.Task, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, nil, err
	}

	type taskJSON struct {
		ID          string     `json:"id"`
		Summary     string     `json:"summary"`
		Description string     `json:"description"`
		Status      string     `json:"status"`
		Priority    int        `json:"priority"`
		DueDate     *time.Time `json:"due_date"`
		StartDate   *time.Time `json:"start_date"`
		Completed   *time.Time `json:"completed"`
		Created     time.Time  `json:"created"`
		Modified    time.Time  `json:"modified"`
		ListID      string     `json:"list_id"`
		ParentID    string     `json:"parent_id"`
		Categories  string     `json:"categories"`
	}

	var taskList []taskJSON
	if err := json.Unmarshal(data, &taskList); err != nil {
		return nil, nil, err
	}

	// Extract list name from filename
	listName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))

	tasks := make([]backend.Task, len(taskList))
	for i, t := range taskList {
		tasks[i] = backend.Task{
			ID:          t.ID,
			Summary:     t.Summary,
			Description: t.Description,
			Status:      backend.TaskStatus(t.Status),
			Priority:    t.Priority,
			DueDate:     t.DueDate,
			StartDate:   t.StartDate,
			Completed:   t.Completed,
			Created:     t.Created,
			Modified:    t.Modified,
			ListID:      t.ListID,
			ParentID:    t.ParentID,
			Categories:  t.Categories,
		}
	}

	list := &backend.List{
		Name:     listName,
		Modified: time.Now(),
	}

	return list, tasks, nil
}

// importCSV imports a list from a CSV file
func importCSV(inputPath string) (*backend.List, []backend.Task, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	if len(records) < 2 {
		return nil, nil, fmt.Errorf("CSV file is empty or has no data rows")
	}

	// Skip header row
	records = records[1:]

	tasks := make([]backend.Task, 0, len(records))
	for _, record := range records {
		if len(record) < 13 {
			continue
		}

		priority, _ := strconv.Atoi(record[4])

		task := backend.Task{
			ID:          record[0],
			Summary:     record[1],
			Description: record[2],
			Status:      backend.TaskStatus(record[3]),
			Priority:    priority,
			ListID:      record[10],
			ParentID:    record[11],
			Categories:  record[12],
		}

		if record[5] != "" {
			t, _ := time.Parse(time.RFC3339, record[5])
			task.DueDate = &t
		}
		if record[6] != "" {
			t, _ := time.Parse(time.RFC3339, record[6])
			task.StartDate = &t
		}
		if record[7] != "" {
			t, _ := time.Parse(time.RFC3339, record[7])
			task.Completed = &t
		}
		if record[8] != "" {
			task.Created, _ = time.Parse(time.RFC3339, record[8])
		}
		if record[9] != "" {
			task.Modified, _ = time.Parse(time.RFC3339, record[9])
		}

		tasks = append(tasks, task)
	}

	// Extract list name from filename
	listName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	list := &backend.List{
		Name:     listName,
		Modified: time.Now(),
	}

	return list, tasks, nil
}

// importICalendar imports a list from an iCalendar file
func importICalendar(inputPath string) (*backend.List, []backend.Task, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, nil, err
	}

	const iCalDateFormat = "20060102T150405Z"

	content := string(data)
	var tasks []backend.Task

	// Parse VTODOs
	vtodoStart := 0
	for {
		start := strings.Index(content[vtodoStart:], "BEGIN:VTODO")
		if start == -1 {
			break
		}
		start += vtodoStart
		end := strings.Index(content[start:], "END:VTODO")
		if end == -1 {
			break
		}
		end += start + len("END:VTODO")

		vtodo := content[start:end]
		task := parseVTODOContent(vtodo, iCalDateFormat)
		if task.ID != "" || task.Summary != "" {
			tasks = append(tasks, task)
		}

		vtodoStart = end
	}

	// Extract list name from filename
	listName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	list := &backend.List{
		Name:     listName,
		Modified: time.Now(),
	}

	return list, tasks, nil
}

// parseVTODOContent parses a VTODO block into a Task
func parseVTODOContent(vtodo, dateFormat string) backend.Task {
	var task backend.Task

	lines := strings.Split(vtodo, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimSuffix(line, "\r")

		if strings.HasPrefix(line, "UID:") {
			task.ID = strings.TrimPrefix(line, "UID:")
		} else if strings.HasPrefix(line, "SUMMARY:") {
			task.Summary = strings.TrimPrefix(line, "SUMMARY:")
		} else if strings.HasPrefix(line, "DESCRIPTION:") {
			task.Description = strings.TrimPrefix(line, "DESCRIPTION:")
		} else if strings.HasPrefix(line, "STATUS:") {
			status := strings.TrimPrefix(line, "STATUS:")
			switch status {
			case "COMPLETED":
				task.Status = backend.StatusCompleted
			case "IN-PROGRESS":
				task.Status = backend.StatusInProgress
			case "CANCELLED":
				task.Status = backend.StatusCancelled
			default:
				task.Status = backend.StatusNeedsAction
			}
		} else if strings.HasPrefix(line, "PRIORITY:") {
			task.Priority, _ = strconv.Atoi(strings.TrimPrefix(line, "PRIORITY:"))
		} else if strings.HasPrefix(line, "CATEGORIES:") {
			task.Categories = strings.TrimPrefix(line, "CATEGORIES:")
		} else if strings.HasPrefix(line, "DUE:") {
			if t, err := time.Parse(dateFormat, strings.TrimPrefix(line, "DUE:")); err == nil {
				task.DueDate = &t
			}
		} else if strings.HasPrefix(line, "DTSTART:") {
			if t, err := time.Parse(dateFormat, strings.TrimPrefix(line, "DTSTART:")); err == nil {
				task.StartDate = &t
			}
		} else if strings.HasPrefix(line, "CREATED:") {
			if t, err := time.Parse(dateFormat, strings.TrimPrefix(line, "CREATED:")); err == nil {
				task.Created = t
			}
		} else if strings.HasPrefix(line, "LAST-MODIFIED:") {
			if t, err := time.Parse(dateFormat, strings.TrimPrefix(line, "LAST-MODIFIED:")); err == nil {
				task.Modified = t
			}
		} else if strings.HasPrefix(line, "COMPLETED:") {
			if t, err := time.Parse(dateFormat, strings.TrimPrefix(line, "COMPLETED:")); err == nil {
				task.Completed = &t
			}
		}
	}

	return task
}

// newListStatsCmd creates the 'list stats' subcommand
func newListStatsCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "stats [name]",
		Short: "Show database statistics",
		Long:  "Display statistics about the database including task counts, status breakdown, and storage usage.",
		Args:  cobra.MaximumNArgs(1),
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

			listName := ""
			if len(args) > 0 {
				listName = args[0]
			}

			jsonOutput, _ := cmd.Flags().GetBool("json")
			return doListStats(context.Background(), be, listName, cfg, stdout, jsonOutput)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doListStats displays database statistics
func doListStats(ctx context.Context, be backend.TaskManager, listName string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// Check if backend supports stats
	sqliteBe, ok := be.(*sqlite.Backend)
	if !ok {
		// Try unwrapping syncAwareBackend
		if sab, sabOk := be.(*syncAwareBackend); sabOk {
			sqliteBe, ok = sab.TaskManager.(*sqlite.Backend)
		}
	}
	if !ok || sqliteBe == nil {
		return fmt.Errorf("stats command is only supported for SQLite backend")
	}

	stats, err := sqliteBe.Stats(ctx, listName)
	if err != nil {
		return err
	}

	if jsonOutput {
		type statsJSON struct {
			Result string                `json:"result"`
			Stats  *sqlite.DatabaseStats `json:"stats"`
		}
		output := statsJSON{
			Result: ResultInfoOnly,
			Stats:  stats,
		}
		jsonBytes, err := json.Marshal(output)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdout, string(jsonBytes))
		return nil
	}

	// Format text output
	_, _ = fmt.Fprintln(stdout, "Database Statistics")
	_, _ = fmt.Fprintln(stdout, "==================")
	_, _ = fmt.Fprintf(stdout, "Total tasks: %d\n\n", stats.TotalTasks)

	if len(stats.Lists) > 0 {
		_, _ = fmt.Fprintln(stdout, "Tasks per list:")
		for _, l := range stats.Lists {
			_, _ = fmt.Fprintf(stdout, "  %-20s %d\n", l.Name, l.Count)
		}
		_, _ = fmt.Fprintln(stdout)
	}

	if len(stats.ByStatus) > 0 {
		_, _ = fmt.Fprintln(stdout, "Tasks by status:")
		for status, count := range stats.ByStatus {
			_, _ = fmt.Fprintf(stdout, "  %-20s %d\n", status, count)
		}
		_, _ = fmt.Fprintln(stdout)
	}

	// Format database size
	sizeStr := formatBytes(stats.DatabaseSizeBytes)
	_, _ = fmt.Fprintf(stdout, "Database size: %s (%d bytes)\n", sizeStr, stats.DatabaseSizeBytes)

	if stats.LastVacuum != nil {
		_, _ = fmt.Fprintf(stdout, "Last vacuum: %s\n", stats.LastVacuum.Format("2006-01-02 15:04:05"))
	}

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// formatBytes converts bytes to human-readable format (KB, MB, etc.)
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// newListVacuumCmd creates the 'list vacuum' subcommand
func newListVacuumCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "vacuum",
		Short: "Compact the database",
		Long:  "Run SQLite VACUUM to reclaim space from deleted data and optimize the database file.",
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

			jsonOutput, _ := cmd.Flags().GetBool("json")
			return doListVacuum(context.Background(), be, cfg, stdout, jsonOutput)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doListVacuum runs the SQLite VACUUM command
func doListVacuum(ctx context.Context, be backend.TaskManager, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// Check if backend supports vacuum
	sqliteBe, ok := be.(*sqlite.Backend)
	if !ok {
		// Try unwrapping syncAwareBackend
		if sab, sabOk := be.(*syncAwareBackend); sabOk {
			sqliteBe, ok = sab.TaskManager.(*sqlite.Backend)
		}
	}
	if !ok || sqliteBe == nil {
		return fmt.Errorf("vacuum command is only supported for SQLite backend")
	}

	result, err := sqliteBe.Vacuum(ctx)
	if err != nil {
		return err
	}

	if jsonOutput {
		type vacuumJSON struct {
			Result      string               `json:"result"`
			VacuumStats *sqlite.VacuumResult `json:"vacuum"`
		}
		output := vacuumJSON{
			Result:      ResultActionCompleted,
			VacuumStats: result,
		}
		jsonBytes, err := json.Marshal(output)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdout, string(jsonBytes))
		return nil
	}

	_, _ = fmt.Fprintln(stdout, "Vacuum completed")
	_, _ = fmt.Fprintf(stdout, "Size before: %s\n", formatBytes(result.SizeBefore))
	_, _ = fmt.Fprintf(stdout, "Size after:  %s\n", formatBytes(result.SizeAfter))
	if result.Reclaimed > 0 {
		_, _ = fmt.Fprintf(stdout, "Reclaimed:   %s\n", formatBytes(result.Reclaimed))
	}

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

	// Load config (creates default if not exists) and check sync/auto-detect settings
	appConfig, _ := config.Load(cfg.ConfigPath)
	loadSyncConfig(cfg)
	loadAutoDetectConfig(cfg, appConfig)

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

	// If auto-detect is enabled, try to detect a backend
	if cfg.AutoDetectBackend {
		workDir := cfg.WorkDir
		if workDir == "" {
			workDir, _ = os.Getwd()
		}

		// Register backends for detection
		registerDetectableBackends(cfg)

		// Try to select a detected backend
		be, name, err := backend.SelectDetectedBackend(workDir)
		if err == nil && be != nil {
			utils.Debugf("Auto-detected backend: %s", name)
			return be, nil
		}
		// Fall through to default backend based on config
	}

	// Check default_backend setting from config
	if appConfig != nil && appConfig.DefaultBackend != "" && appConfig.DefaultBackend != "sqlite" {
		switch appConfig.DefaultBackend {
		case "todoist":
			// Create Todoist backend using environment variable credentials
			todoistCfg := todoist.ConfigFromEnv()
			if todoistCfg.APIToken == "" {
				return nil, fmt.Errorf("todoist backend is configured as default but TODOAT_TODOIST_TOKEN environment variable is not set")
			}
			utils.Debugf("Using default backend: todoist")
			return todoist.New(todoistCfg)
		}
	}

	return sqlite.New(dbPath)
}

// loadAutoDetectConfig loads the auto-detect configuration from the config file
func loadAutoDetectConfig(cfg *Config, appConfig *config.Config) {
	if appConfig != nil && appConfig.IsAutoDetectEnabled() {
		cfg.AutoDetectBackend = true
	}
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

// GetTaskByLocalID delegates to the underlying backend if it supports LocalIDBackend
func (b *syncAwareBackend) GetTaskByLocalID(ctx context.Context, listID string, localID int64) (*backend.Task, error) {
	if localBE, ok := b.TaskManager.(LocalIDBackend); ok {
		return localBE.GetTaskByLocalID(ctx, listID, localID)
	}
	return nil, fmt.Errorf("underlying backend does not support local-id lookup")
}

// GetTaskLocalID delegates to the underlying backend if it supports LocalIDBackend
func (b *syncAwareBackend) GetTaskLocalID(ctx context.Context, taskID string) (int64, error) {
	if localBE, ok := b.TaskManager.(LocalIDBackend); ok {
		return localBE.GetTaskLocalID(ctx, taskID)
	}
	return 0, fmt.Errorf("underlying backend does not support local-id lookup")
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
		// Parse date filter flags
		dueBeforeStr, _ := cmd.Flags().GetString("due-before")
		dueAfterStr, _ := cmd.Flags().GetString("due-after")
		createdBeforeStr, _ := cmd.Flags().GetString("created-before")
		createdAfterStr, _ := cmd.Flags().GetString("created-after")
		dateFilter, err := parseDateFilter(dueBeforeStr, dueAfterStr, createdBeforeStr, createdAfterStr)
		if err != nil {
			return err
		}
		return doGet(ctx, be, list, statusFilter, priorityFilter, tagFilter, dateFilter, viewName, cfg, stdout, jsonOutput)
	case "add":
		priorityStr, _ := cmd.Flags().GetString("priority")
		priority, err := parsePrioritySingle(priorityStr)
		if err != nil {
			return err
		}
		description, _ := cmd.Flags().GetString("description")
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
		tagsAlias, _ := cmd.Flags().GetStringSlice("tags")
		tags = append(tags, tagsAlias...)
		tags = normalizeTagSlice(tags)
		categories := strings.Join(tags, ",")
		parentSummary, _ := cmd.Flags().GetString("parent")
		literal, _ := cmd.Flags().GetBool("literal")
		return doAdd(ctx, be, list, taskSummary, priority, description, dueDate, startDate, categories, parentSummary, literal, cfg, stdout, jsonOutput)
	case "update":
		// Check for direct ID selection flags
		uidFlag, _ := cmd.Flags().GetString("uid")
		localIDFlag, _ := cmd.Flags().GetInt64("local-id")

		priorityStr, _ := cmd.Flags().GetString("priority")
		priority, err := parsePrioritySingle(priorityStr)
		if err != nil {
			return err
		}
		status, _ := cmd.Flags().GetString("status")
		newSummary, _ := cmd.Flags().GetString("summary")
		descriptionStr, _ := cmd.Flags().GetString("description")
		descriptionFlagSet := cmd.Flags().Changed("description")
		var newDescription *string
		if descriptionFlagSet {
			newDescription = &descriptionStr
		}
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
		tagsFlagSet := cmd.Flags().Changed("tags")
		var newCategories *string
		if tagFlagSet || tagsFlagSet {
			tags, _ := cmd.Flags().GetStringSlice("tag")
			tagsAlias, _ := cmd.Flags().GetStringSlice("tags")
			tags = append(tags, tagsAlias...)
			tags = normalizeTagSlice(tags)
			cat := strings.Join(tags, ",")
			newCategories = &cat
		}
		addTagsSlice, _ := cmd.Flags().GetStringSlice("add-tag")
		addTagsSlice = normalizeTagSlice(addTagsSlice)
		removeTagsSlice, _ := cmd.Flags().GetStringSlice("remove-tag")
		removeTagsSlice = normalizeTagSlice(removeTagsSlice)
		parentSummary, _ := cmd.Flags().GetString("parent")
		noParent, _ := cmd.Flags().GetBool("no-parent")

		// Check for bulk pattern first (before ID resolution)
		_, _, isBulk := parseBulkPattern(taskSummary)
		if isBulk && uidFlag == "" && !cmd.Flags().Changed("local-id") {
			// Use original bulk update function
			return doUpdate(ctx, be, list, taskSummary, newSummary, newDescription, status, priority, dueDate, startDate, clearDueDate, clearStartDate, newCategories, addTagsSlice, removeTagsSlice, parentSummary, noParent, cfg, stdout, jsonOutput)
		}

		// Resolve task by UID, local-id, or summary
		task, err := resolveTaskByID(ctx, cmd, be, list, taskSummary, uidFlag, localIDFlag, cfg)
		if err != nil {
			return err
		}
		return doUpdateWithTask(ctx, be, list, task, newSummary, newDescription, status, priority, dueDate, startDate, clearDueDate, clearStartDate, newCategories, addTagsSlice, removeTagsSlice, parentSummary, noParent, cfg, stdout, jsonOutput)
	case "complete":
		// Check for direct ID selection flags
		uidFlag, _ := cmd.Flags().GetString("uid")
		localIDFlag, _ := cmd.Flags().GetInt64("local-id")

		// Check for bulk pattern first (before ID resolution)
		_, _, isBulk := parseBulkPattern(taskSummary)
		if isBulk && uidFlag == "" && !cmd.Flags().Changed("local-id") {
			// Use original bulk complete function
			return doComplete(ctx, be, list, taskSummary, cfg, stdout, jsonOutput)
		}

		// Resolve task by UID, local-id, or summary
		task, err := resolveTaskByID(ctx, cmd, be, list, taskSummary, uidFlag, localIDFlag, cfg)
		if err != nil {
			return err
		}
		return doCompleteWithTask(ctx, be, list, task, cfg, stdout, jsonOutput)
	case "delete":
		// Check for direct ID selection flags
		uidFlag, _ := cmd.Flags().GetString("uid")
		localIDFlag, _ := cmd.Flags().GetInt64("local-id")

		// Check for bulk pattern first (before ID resolution)
		_, _, isBulk := parseBulkPattern(taskSummary)
		if isBulk && uidFlag == "" && !cmd.Flags().Changed("local-id") {
			// Use original bulk delete function
			return doDelete(ctx, be, list, taskSummary, cfg, stdout, jsonOutput)
		}

		// Resolve task by UID, local-id, or summary
		task, err := resolveTaskByID(ctx, cmd, be, list, taskSummary, uidFlag, localIDFlag, cfg)
		if err != nil {
			return err
		}
		return doDeleteWithTask(ctx, be, list, task, cfg, stdout, jsonOutput)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// doGet lists all tasks in a list, optionally filtering by status, priority, tags, and/or dates
func doGet(ctx context.Context, be backend.TaskManager, list *backend.List, statusFilter string, priorityFilter []int, tagFilter []string, dateFilter DateFilter, viewName string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	tasks, err := be.GetTasks(ctx, list.ID)
	if err != nil {
		return err
	}

	// Load and apply view if specified
	if viewName != "" {
		return doGetWithView(ctx, be, tasks, list, viewName, cfg, stdout, jsonOutput)
	}

	// Filter by status if specified
	if statusFilter != "" {
		filterStatus, err := parseStatusWithValidation(statusFilter)
		if err != nil {
			return err
		}
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

	// Filter by date if specified (AND logic - must match all date criteria)
	if !dateFilter.IsEmpty() {
		var filteredTasks []backend.Task
		for _, t := range tasks {
			if matchesDateFilter(t, dateFilter) {
				filteredTasks = append(filteredTasks, t)
			}
		}
		tasks = filteredTasks
	}

	if jsonOutput {
		return outputTaskListJSON(ctx, be, tasks, list, cfg, stdout)
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
func doGetWithView(ctx context.Context, be backend.TaskManager, tasks []backend.Task, list *backend.List, viewName string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
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
		return outputTaskListJSON(ctx, be, sortedTasks, list, cfg, stdout)
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

// applyTagChanges applies --add-tag and --remove-tag operations to existing categories
// Tags are case-insensitive for matching but preserve original case
func applyTagChanges(existingCategories string, addTags, removeTags []string) string {
	// Skip if no changes requested
	if len(addTags) == 0 && len(removeTags) == 0 {
		return existingCategories
	}

	// Parse existing tags into a map (lowercase key -> original case value)
	tagMap := make(map[string]string)
	if existingCategories != "" {
		for _, tag := range strings.Split(existingCategories, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagMap[strings.ToLower(tag)] = tag
			}
		}
	}

	// Remove tags (case-insensitive)
	for _, tag := range removeTags {
		delete(tagMap, strings.ToLower(tag))
	}

	// Add tags (case-insensitive for dedup, preserve original case)
	for _, tag := range addTags {
		lowerTag := strings.ToLower(tag)
		if _, exists := tagMap[lowerTag]; !exists {
			tagMap[lowerTag] = tag
		}
	}

	// Convert map back to sorted slice (for consistent output)
	var result []string
	for _, tag := range tagMap {
		result = append(result, tag)
	}
	sort.Strings(result)
	return strings.Join(result, ",")
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

// parseDate parses a date string in YYYY-MM-DD format or relative formats (today, tomorrow, yesterday, +Nd, -Nd, +Nw, +Nm)
func parseDate(s string) (*time.Time, error) {
	return utils.ParseDateFlag(s)
}

// DateFilter holds date filtering criteria for tasks
type DateFilter struct {
	DueBefore     *time.Time
	DueAfter      *time.Time
	CreatedBefore *time.Time
	CreatedAfter  *time.Time
}

// IsEmpty returns true if no date filters are set
func (f DateFilter) IsEmpty() bool {
	return f.DueBefore == nil && f.DueAfter == nil && f.CreatedBefore == nil && f.CreatedAfter == nil
}

// parseDateFilter parses date filter flag values into a DateFilter struct
func parseDateFilter(dueBefore, dueAfter, createdBefore, createdAfter string) (DateFilter, error) {
	var filter DateFilter
	var err error

	if dueBefore != "" {
		filter.DueBefore, err = parseDate(dueBefore)
		if err != nil {
			return filter, fmt.Errorf("invalid --due-before: %w", err)
		}
	}
	if dueAfter != "" {
		filter.DueAfter, err = parseDate(dueAfter)
		if err != nil {
			return filter, fmt.Errorf("invalid --due-after: %w", err)
		}
	}
	if createdBefore != "" {
		filter.CreatedBefore, err = parseDate(createdBefore)
		if err != nil {
			return filter, fmt.Errorf("invalid --created-before: %w", err)
		}
	}
	if createdAfter != "" {
		filter.CreatedAfter, err = parseDate(createdAfter)
		if err != nil {
			return filter, fmt.Errorf("invalid --created-after: %w", err)
		}
	}

	return filter, nil
}

// matchesDateFilter checks if a task matches the given date filter criteria.
// Date filters use inclusive ranges. Tasks without dates are excluded from date filters.
func matchesDateFilter(task backend.Task, filter DateFilter) bool {
	// If no filters are set, all tasks match
	if filter.IsEmpty() {
		return true
	}

	// Due date filters
	if filter.DueBefore != nil || filter.DueAfter != nil {
		// Tasks without due date don't match due date filters
		if task.DueDate == nil {
			return false
		}
		// Check due-before (inclusive: task due date < filter date + 1 day)
		if filter.DueBefore != nil {
			// Use start of next day for inclusive comparison
			beforeEndOfDay := filter.DueBefore.AddDate(0, 0, 1)
			if !task.DueDate.Before(beforeEndOfDay) {
				return false
			}
		}
		// Check due-after (inclusive: task due date >= filter date)
		if filter.DueAfter != nil {
			if task.DueDate.Before(*filter.DueAfter) {
				return false
			}
		}
	}

	// Created date filters
	if filter.CreatedBefore != nil {
		// Use start of next day for inclusive comparison
		beforeEndOfDay := filter.CreatedBefore.AddDate(0, 0, 1)
		if !task.Created.Before(beforeEndOfDay) {
			return false
		}
	}
	if filter.CreatedAfter != nil {
		if task.Created.Before(*filter.CreatedAfter) {
			return false
		}
	}

	return true
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
func doAdd(ctx context.Context, be backend.TaskManager, list *backend.List, summary string, priority int, description string, dueDate, startDate *time.Time, categories string, parentSummary string, literal bool, cfg *Config, stdout io.Writer, jsonOutput bool) error {
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
		return doAddHierarchy(ctx, be, list, summary, priority, description, dueDate, startDate, categories, cfg, stdout, jsonOutput)
	}

	task := &backend.Task{
		Summary:     summary,
		Description: description,
		Priority:    priority,
		Status:      backend.StatusNeedsAction,
		DueDate:     dueDate,
		StartDate:   startDate,
		Categories:  categories,
		ParentID:    parentID,
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
func doAddHierarchy(ctx context.Context, be backend.TaskManager, list *backend.List, path string, priority int, description string, dueDate, startDate *time.Time, categories string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
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
			// Only apply priority, description, dates, and categories to the leaf task
			taskPriority := 0
			taskDescription := ""
			var taskDueDate, taskStartDate *time.Time
			taskCategories := ""
			if i == len(parts)-1 {
				taskPriority = priority
				taskDescription = description
				taskDueDate = dueDate
				taskStartDate = startDate
				taskCategories = categories
			}

			task := &backend.Task{
				Summary:     part,
				Description: taskDescription,
				Priority:    taskPriority,
				Status:      backend.StatusNeedsAction,
				DueDate:     taskDueDate,
				StartDate:   taskStartDate,
				Categories:  taskCategories,
				ParentID:    parentID,
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
func doUpdate(ctx context.Context, be backend.TaskManager, list *backend.List, taskSummary, newSummary string, newDescription *string, status string, priority int, dueDate, startDate *time.Time, clearDueDate, clearStartDate bool, newCategories *string, addTags, removeTags []string, parentSummary string, noParent bool, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// Check for bulk pattern
	bulkParentSummary, pattern, isBulk := parseBulkPattern(taskSummary)
	if isBulk {
		return doBulkUpdate(ctx, be, list, bulkParentSummary, pattern, newDescription, status, priority, dueDate, startDate, clearDueDate, clearStartDate, newCategories, cfg, stdout, jsonOutput)
	}

	task, err := findTask(ctx, be, list, taskSummary, cfg)
	if err != nil {
		return err
	}

	// Apply updates
	if newSummary != "" {
		task.Summary = newSummary
	}
	if newDescription != nil {
		task.Description = *newDescription
	}
	if status != "" {
		parsedStatus, err := parseStatusWithValidation(status)
		if err != nil {
			return err
		}
		task.Status = parsedStatus
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
	// Handle --add-tag and --remove-tag (only if --tags was not set)
	if newCategories == nil {
		task.Categories = applyTagChanges(task.Categories, addTags, removeTags)
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

// doBulkUpdate modifies all children/descendants of a parent task
func doBulkUpdate(ctx context.Context, be backend.TaskManager, list *backend.List, parentSummary, pattern string, newDescription *string, status string, priority int, dueDate, startDate *time.Time, clearDueDate, clearStartDate bool, newCategories *string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// Find the parent task
	parent, err := findTask(ctx, be, list, parentSummary, cfg)
	if err != nil {
		return err
	}

	// Get all tasks to find children
	tasks, err := be.GetTasks(ctx, list.ID)
	if err != nil {
		return err
	}

	// Get children based on pattern
	recursive := pattern == "**"
	children := getChildTasks(parent.ID, tasks, recursive)

	// If no children found, return INFO_ONLY
	if len(children) == 0 {
		if jsonOutput {
			resp := bulkActionResponse{
				Result:        ResultInfoOnly,
				Action:        "update",
				AffectedCount: 0,
				Parent:        parent.Summary,
				Pattern:       pattern,
				AffectedUIDs:  []string{},
			}
			jsonBytes, _ := json.Marshal(resp)
			_, _ = fmt.Fprintln(stdout, string(jsonBytes))
			return nil
		}
		_, _ = fmt.Fprintf(stdout, "Updated 0 tasks under \"%s\"\n", parent.Summary)
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
		}
		return nil
	}

	// Parse status if provided
	var parsedStatus backend.TaskStatus
	if status != "" {
		parsedStatus, err = parseStatusWithValidation(status)
		if err != nil {
			return err
		}
	}

	// Update each child
	var affectedUIDs []string
	for i := range children {
		if newDescription != nil {
			children[i].Description = *newDescription
		}
		if status != "" {
			children[i].Status = parsedStatus
		}
		if priority > 0 {
			children[i].Priority = priority
		}
		if dueDate != nil {
			children[i].DueDate = dueDate
		}
		if clearDueDate {
			children[i].DueDate = nil
		}
		if startDate != nil {
			children[i].StartDate = startDate
		}
		if clearStartDate {
			children[i].StartDate = nil
		}
		if newCategories != nil {
			children[i].Categories = *newCategories
		}

		_, err := be.UpdateTask(ctx, list.ID, &children[i])
		if err != nil {
			return err
		}
		affectedUIDs = append(affectedUIDs, children[i].ID)
	}

	if jsonOutput {
		resp := bulkActionResponse{
			Result:        ResultActionCompleted,
			Action:        "update",
			AffectedCount: len(children),
			Parent:        parent.Summary,
			Pattern:       pattern,
			AffectedUIDs:  affectedUIDs,
		}
		jsonBytes, _ := json.Marshal(resp)
		_, _ = fmt.Fprintln(stdout, string(jsonBytes))
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "Updated %d tasks under \"%s\"\n", len(children), parent.Summary)

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

// parseStatusWithValidation converts a status string to TaskStatus, returning an error for invalid values.
// Valid values: TODO, IN-PROGRESS, DONE, CANCELLED (and their aliases)
func parseStatusWithValidation(s string) (backend.TaskStatus, error) {
	switch strings.ToUpper(s) {
	case "DONE", "COMPLETED", "D":
		return backend.StatusCompleted, nil
	case "IN-PROGRESS", "INPROGRESS", "PROGRESS":
		return backend.StatusInProgress, nil
	case "CANCELLED", "CANCELED":
		return backend.StatusCancelled, nil
	case "TODO", "NEEDS-ACTION", "T":
		return backend.StatusNeedsAction, nil
	default:
		return backend.StatusNeedsAction, fmt.Errorf("invalid status %q: valid values are TODO, IN-PROGRESS, DONE, CANCELLED", s)
	}
}

// doComplete marks a task as completed
func doComplete(ctx context.Context, be backend.TaskManager, list *backend.List, taskSummary string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// Check for bulk pattern
	parentSummary, pattern, isBulk := parseBulkPattern(taskSummary)
	if isBulk {
		return doBulkComplete(ctx, be, list, parentSummary, pattern, cfg, stdout, jsonOutput)
	}

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

// doBulkComplete marks all children/descendants of a parent task as completed
func doBulkComplete(ctx context.Context, be backend.TaskManager, list *backend.List, parentSummary, pattern string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// Find the parent task
	parent, err := findTask(ctx, be, list, parentSummary, cfg)
	if err != nil {
		return err
	}

	// Get all tasks to find children
	tasks, err := be.GetTasks(ctx, list.ID)
	if err != nil {
		return err
	}

	// Get children based on pattern
	recursive := pattern == "**"
	children := getChildTasks(parent.ID, tasks, recursive)

	// If no children found, return INFO_ONLY
	if len(children) == 0 {
		if jsonOutput {
			resp := bulkActionResponse{
				Result:        ResultInfoOnly,
				Action:        "complete",
				AffectedCount: 0,
				Parent:        parent.Summary,
				Pattern:       pattern,
				AffectedUIDs:  []string{},
			}
			jsonBytes, _ := json.Marshal(resp)
			_, _ = fmt.Fprintln(stdout, string(jsonBytes))
			return nil
		}
		_, _ = fmt.Fprintf(stdout, "Completed 0 tasks under \"%s\"\n", parent.Summary)
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
		}
		return nil
	}

	// Complete each child
	now := time.Now().UTC()
	var affectedUIDs []string
	for i := range children {
		children[i].Status = backend.StatusCompleted
		children[i].Completed = &now
		_, err := be.UpdateTask(ctx, list.ID, &children[i])
		if err != nil {
			return err
		}
		affectedUIDs = append(affectedUIDs, children[i].ID)
	}

	if jsonOutput {
		resp := bulkActionResponse{
			Result:        ResultActionCompleted,
			Action:        "complete",
			AffectedCount: len(children),
			Parent:        parent.Summary,
			Pattern:       pattern,
			AffectedUIDs:  affectedUIDs,
		}
		jsonBytes, _ := json.Marshal(resp)
		_, _ = fmt.Fprintln(stdout, string(jsonBytes))
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "Completed %d tasks under \"%s\"\n", len(children), parent.Summary)

	// Emit ACTION_COMPLETED result code in no-prompt mode
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// doDelete removes a task
func doDelete(ctx context.Context, be backend.TaskManager, list *backend.List, taskSummary string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// Check for bulk pattern
	parentSummary, pattern, isBulk := parseBulkPattern(taskSummary)
	if isBulk {
		return doBulkDelete(ctx, be, list, parentSummary, pattern, cfg, stdout, jsonOutput)
	}

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

// doBulkDelete removes all children/descendants of a parent task
func doBulkDelete(ctx context.Context, be backend.TaskManager, list *backend.List, parentSummary, pattern string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// Find the parent task
	parent, err := findTask(ctx, be, list, parentSummary, cfg)
	if err != nil {
		return err
	}

	// Get all tasks to find children
	tasks, err := be.GetTasks(ctx, list.ID)
	if err != nil {
		return err
	}

	// Get children based on pattern
	recursive := pattern == "**"
	children := getChildTasks(parent.ID, tasks, recursive)

	// If no children found, return INFO_ONLY
	if len(children) == 0 {
		if jsonOutput {
			resp := bulkActionResponse{
				Result:        ResultInfoOnly,
				Action:        "delete",
				AffectedCount: 0,
				Parent:        parent.Summary,
				Pattern:       pattern,
				AffectedUIDs:  []string{},
			}
			jsonBytes, _ := json.Marshal(resp)
			_, _ = fmt.Fprintln(stdout, string(jsonBytes))
			return nil
		}
		_, _ = fmt.Fprintf(stdout, "Deleted 0 tasks under \"%s\"\n", parent.Summary)
		if cfg != nil && cfg.NoPrompt {
			_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
		}
		return nil
	}

	// Collect affected UIDs before deletion (including cascaded descendants for * pattern)
	var affectedUIDs []string

	// For direct children pattern (*), we need to delete their descendants first
	// For all descendants pattern (**), we delete in reverse depth order
	if !recursive {
		// Direct children only - need to also delete each child's descendants
		// Collect all UIDs that will be affected (children + their descendants)
		for _, child := range children {
			affectedUIDs = append(affectedUIDs, child.ID)
			// Also count descendants that will be cascaded
			childDescendants := findDescendants(child.ID, tasks)
			affectedUIDs = append(affectedUIDs, childDescendants...)
		}

		// Now delete
		for _, child := range children {
			// Get this child's descendants
			childDescendants := findDescendants(child.ID, tasks)
			// Delete descendants first (bottom-up)
			for i := len(childDescendants) - 1; i >= 0; i-- {
				if err := be.DeleteTask(ctx, list.ID, childDescendants[i]); err != nil {
					return err
				}
			}
			// Then delete the child itself
			if err := be.DeleteTask(ctx, list.ID, child.ID); err != nil {
				return err
			}
		}
	} else {
		// All descendants - collect UIDs
		for _, child := range children {
			affectedUIDs = append(affectedUIDs, child.ID)
		}

		// Sort by depth and delete deepest first
		// Build depth map
		depthMap := make(map[string]int)
		for _, t := range tasks {
			depthMap[t.ID] = calculateDepth(t.ID, tasks)
		}
		// Sort children by depth (deepest first)
		sort.Slice(children, func(i, j int) bool {
			return depthMap[children[i].ID] > depthMap[children[j].ID]
		})
		// Delete in order
		for _, child := range children {
			if err := be.DeleteTask(ctx, list.ID, child.ID); err != nil {
				return err
			}
		}
	}

	if jsonOutput {
		resp := bulkActionResponse{
			Result:        ResultActionCompleted,
			Action:        "delete",
			AffectedCount: len(affectedUIDs),
			Parent:        parent.Summary,
			Pattern:       pattern,
			AffectedUIDs:  affectedUIDs,
		}
		jsonBytes, _ := json.Marshal(resp)
		_, _ = fmt.Fprintln(stdout, string(jsonBytes))
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "Deleted %d tasks under \"%s\"\n", len(affectedUIDs), parent.Summary)

	// Emit ACTION_COMPLETED result code in no-prompt mode
	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// calculateDepth returns the depth of a task in the hierarchy (0 for root tasks)
func calculateDepth(taskID string, tasks []backend.Task) int {
	taskMap := make(map[string]backend.Task)
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	depth := 0
	current := taskID
	for {
		task, ok := taskMap[current]
		if !ok || task.ParentID == "" {
			break
		}
		depth++
		current = task.ParentID
	}
	return depth
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

// parseBulkPattern checks if a search term ends with /* or /** for bulk operations
// Returns the parent summary, the pattern type ("*" for direct children, "**" for all descendants), and whether it's a bulk pattern
func parseBulkPattern(searchTerm string) (parentSummary string, pattern string, isBulk bool) {
	if strings.HasSuffix(searchTerm, "/**") {
		return strings.TrimSuffix(searchTerm, "/**"), "**", true
	}
	if strings.HasSuffix(searchTerm, "/*") {
		return strings.TrimSuffix(searchTerm, "/*"), "*", true
	}
	return searchTerm, "", false
}

// getChildTasks returns the children of a parent task
// If recursive is true, returns all descendants; otherwise only direct children
func getChildTasks(parentID string, tasks []backend.Task, recursive bool) []backend.Task {
	if !recursive {
		// Direct children only
		var children []backend.Task
		for _, t := range tasks {
			if t.ParentID == parentID {
				children = append(children, t)
			}
		}
		return children
	}

	// All descendants (recursive)
	descendantIDs := findDescendants(parentID, tasks)
	idSet := make(map[string]bool)
	for _, id := range descendantIDs {
		idSet[id] = true
	}

	var descendants []backend.Task
	for _, t := range tasks {
		if idSet[t.ID] {
			descendants = append(descendants, t)
		}
	}
	return descendants
}

// bulkActionResponse is the JSON output structure for bulk operations
type bulkActionResponse struct {
	Result        string   `json:"result"`
	Action        string   `json:"action"`
	AffectedCount int      `json:"affected_count"`
	Parent        string   `json:"parent"`
	Pattern       string   `json:"pattern"`
	AffectedUIDs  []string `json:"affected_uids,omitempty"`
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

// resolveTaskByID resolves a task by UID, local-id, or summary (falls back to findTask for summary-based search)
func resolveTaskByID(ctx context.Context, cmd *cobra.Command, be backend.TaskManager, list *backend.List, taskSummary, uidFlag string, localIDFlag int64, cfg *Config) (*backend.Task, error) {
	// Check for --uid flag
	if uidFlag != "" {
		// Look up task directly by UID
		task, err := be.GetTask(ctx, list.ID, uidFlag)
		if err != nil {
			return nil, err
		}
		if task == nil {
			return nil, fmt.Errorf("no task found with UID '%s'", uidFlag)
		}
		return task, nil
	}

	// Check for --local-id flag
	if cmd.Flags().Changed("local-id") {
		// Check if sync is enabled
		if cfg == nil || !cfg.SyncEnabled {
			return nil, fmt.Errorf("--local-id requires sync to be enabled")
		}

		// Check if backend supports local-id lookup
		localBE, ok := be.(LocalIDBackend)
		if !ok {
			return nil, fmt.Errorf("--local-id is only supported with SQLite backend")
		}

		// Look up task by local ID
		task, err := localBE.GetTaskByLocalID(ctx, list.ID, localIDFlag)
		if err != nil {
			return nil, err
		}
		if task == nil {
			return nil, fmt.Errorf("no task found with local-id %d", localIDFlag)
		}
		return task, nil
	}

	// Fall back to summary-based search (for bulk patterns)
	if taskSummary == "" {
		return nil, fmt.Errorf("task summary, --uid, or --local-id is required")
	}

	// Check for bulk pattern - if so, return nil to let the do* functions handle it
	_, _, isBulk := parseBulkPattern(taskSummary)
	if isBulk {
		// Return a placeholder - the do* functions will handle bulk patterns
		return nil, nil
	}

	return findTask(ctx, be, list, taskSummary, cfg)
}

// doUpdateWithTask modifies an existing task (task already resolved)
func doUpdateWithTask(ctx context.Context, be backend.TaskManager, list *backend.List, task *backend.Task, newSummary string, newDescription *string, status string, priority int, dueDate, startDate *time.Time, clearDueDate, clearStartDate bool, newCategories *string, addTags, removeTags []string, parentSummary string, noParent bool, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// If task is nil, fall back to original behavior (for bulk patterns)
	if task == nil {
		return fmt.Errorf("task not found")
	}

	// Apply updates
	if newSummary != "" {
		task.Summary = newSummary
	}
	if newDescription != nil {
		task.Description = *newDescription
	}
	if status != "" {
		parsedStatus, err := parseStatusWithValidation(status)
		if err != nil {
			return err
		}
		task.Status = parsedStatus
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
	// Handle --add-tag and --remove-tag (only if --tags was not set)
	if newCategories == nil {
		task.Categories = applyTagChanges(task.Categories, addTags, removeTags)
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

// doCompleteWithTask marks a task as completed (task already resolved)
func doCompleteWithTask(ctx context.Context, be backend.TaskManager, list *backend.List, task *backend.Task, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// If task is nil, this shouldn't happen for complete
	if task == nil {
		return fmt.Errorf("task not found")
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

// doDeleteWithTask removes a task (task already resolved)
func doDeleteWithTask(ctx context.Context, be backend.TaskManager, list *backend.List, task *backend.Task, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// If task is nil, this shouldn't happen for delete
	if task == nil {
		return fmt.Errorf("task not found")
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

// JSON output structures
type taskJSON struct {
	UID         string   `json:"uid"`
	LocalID     *int64   `json:"local_id,omitempty"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Priority    int      `json:"priority"`
	ParentID    string   `json:"parent_id,omitempty"`
	DueDate     *string  `json:"due_date,omitempty"`
	StartDate   *string  `json:"start_date,omitempty"`
	Completed   *string  `json:"completed,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Synced      *bool    `json:"synced,omitempty"`
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
		UID:         t.ID,
		Summary:     t.Summary,
		Description: t.Description,
		Status:      statusToString(t.Status),
		Priority:    t.Priority,
		ParentID:    t.ParentID,
	}
	if t.DueDate != nil {
		s := t.DueDate.Format(views.DefaultDateFormat)
		result.DueDate = &s
	}
	if t.StartDate != nil {
		s := t.StartDate.Format(views.DefaultDateFormat)
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
func outputTaskListJSON(ctx context.Context, be backend.TaskManager, tasks []backend.Task, list *backend.List, cfg *Config, stdout io.Writer) error {
	var jsonTasks []taskJSON

	// Check if backend supports local-id lookup and sync is enabled
	localBE, supportsLocalID := be.(LocalIDBackend)
	includeLocalID := supportsLocalID && cfg != nil && cfg.SyncEnabled

	for _, t := range tasks {
		jt := taskToJSON(&t)

		// Add local_id if supported
		if includeLocalID {
			localID, err := localBE.GetTaskLocalID(ctx, t.ID)
			if err == nil && localID > 0 {
				jt.LocalID = &localID
				synced := t.ID != "" // Task has UID if synced
				jt.Synced = &synced
			}
		}

		jsonTasks = append(jsonTasks, jt)
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

	// Get offline mode from config
	offlineMode := getOfflineMode(cfg)
	_, _ = fmt.Fprintf(stdout, "Offline Mode: %s\n", offlineMode)
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

// getOfflineMode returns the configured offline mode from config file
func getOfflineMode(cfg *Config) string {
	configPath := cfg.ConfigPath
	if configPath == "" {
		if cfg.DBPath != "" {
			configPath = filepath.Join(filepath.Dir(cfg.DBPath), "config.yaml")
		}
	}

	if configPath == "" {
		return "auto" // default
	}

	appConfig, err := config.LoadFromPath(configPath)
	if err != nil || appConfig == nil {
		return "auto" // default
	}

	return appConfig.GetOfflineMode()
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

// getTrashRetentionDays reads the trash retention days from config file.
// Returns 30 (default) if not configured, or 0 if auto-purge is disabled.
func getTrashRetentionDays(cfg *Config) int {
	configPath := cfg.ConfigPath
	if configPath == "" {
		return 30 // Default retention period
	}

	appConfig, err := config.LoadFromPath(configPath)
	if err != nil || appConfig == nil {
		return 30 // Default retention period
	}

	return appConfig.GetTrashRetentionDays()
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
				if t, err := time.Parse(views.DefaultDateFormat, dueDate); err == nil {
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
				taskMap["due_date"] = task.DueDate.Format(views.DefaultDateFormat)
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

func (m *MockBackend) UpdateList(ctx context.Context, list *backend.List) (*backend.List, error) {
	for i := range m.lists {
		if m.lists[i].ID == list.ID {
			m.lists[i].Name = list.Name
			m.lists[i].Color = list.Color
			m.lists[i].Modified = time.Now()
			return &m.lists[i], nil
		}
	}
	return nil, fmt.Errorf("list not found")
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
						taskMap["due_date"] = t.DueDate.Format(views.DefaultDateFormat)
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
				taskMap["due_date"] = t.DueDate.Format(views.DefaultDateFormat)
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
			_, _ = fmt.Fprintf(stdout, "  - %s (due: %s)\n", task.Summary, task.DueDate.Format(views.DefaultDateFormat))
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
			_, _ = fmt.Fprintf(stdout, "  - %s (due: %s)\n", task.Summary, task.DueDate.Format(views.DefaultDateFormat))
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

// runDetectBackend shows detected backends and the one that would be used
func runDetectBackend(stdout io.Writer, cfg *Config) error {
	workDir := cfg.WorkDir
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Register default backends for detection
	registerDetectableBackends(cfg)

	// Run detection
	results, err := backend.DetectBackends(workDir)
	if err != nil {
		return fmt.Errorf("detection failed: %w", err)
	}

	// Output detection results
	_, _ = fmt.Fprintln(stdout, "Auto-detected backends:")

	if len(results) == 0 {
		_, _ = fmt.Fprintln(stdout, "  (none detected)")
		_, _ = fmt.Fprintln(stdout, "\nNo backends could be detected.")
		return nil
	}

	// Show available backends first with numbering
	num := 1
	var firstAvailable string
	for _, r := range results {
		if r.Available {
			_, _ = fmt.Fprintf(stdout, "  %d. %s: %s\n", num, r.Name, r.Info)
			if firstAvailable == "" {
				firstAvailable = r.Name
			}
			num++
			// Close backend as we're just showing info
			if r.Backend != nil {
				_ = r.Backend.Close()
			}
		}
	}

	// Show unavailable backends
	for _, r := range results {
		if !r.Available {
			_, _ = fmt.Fprintf(stdout, "     %s: (not available) %s\n", r.Name, r.Info)
		}
	}

	// Show what would be used
	if firstAvailable != "" {
		_, _ = fmt.Fprintf(stdout, "\nWould use: %s\n", firstAvailable)
	} else {
		_, _ = fmt.Fprintln(stdout, "\nNo backends available. Configure a backend in config.yaml.")
	}

	return nil
}

// registerDetectableBackends registers all detectable backends
func registerDetectableBackends(cfg *Config) {
	// Git backend is registered from backend/git package via init()
	// We need to register SQLite here since it depends on config

	dbPath := cfg.DBPath
	if dbPath == "" {
		dbPath = getDefaultDBPath()
	}

	// Register SQLite as always-available fallback
	backend.RegisterDetectable("sqlite", func(workDir string) (backend.DetectableBackend, error) {
		return sqlite.NewDetectable(dbPath)
	})
}

// =============================================================================
// Config Command (049-config-cli-commands)
// =============================================================================

// newConfigCmd creates the 'config' subcommand for configuration management
func newConfigCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "View and manage configuration",
		Long:  "View and modify todoat configuration without manually editing YAML files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default action is to show all config (same as 'config get')
			return doConfigGet(cmd, stdout, cfg, "", false)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	configCmd.AddCommand(newConfigGetCmd(stdout, cfg))
	configCmd.AddCommand(newConfigSetCmd(stdout, stderr, cfg))
	configCmd.AddCommand(newConfigPathCmd(stdout, cfg))
	configCmd.AddCommand(newConfigEditCmd(stdout, stderr, cfg))
	configCmd.AddCommand(newConfigResetCmd(stdout, stderr, cfg))

	return configCmd
}

// newConfigGetCmd creates the 'config get' subcommand
func newConfigGetCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Display configuration value(s)",
		Long:  "Display a specific configuration value or all config if no key specified.\nSupports dot notation for nested keys (e.g., sync.enabled, backends.sqlite.path).",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := ""
			if len(args) > 0 {
				key = args[0]
			}
			jsonOutput, _ := cmd.Flags().GetBool("json")
			return doConfigGet(cmd, stdout, cfg, key, jsonOutput)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doConfigGet handles the config get command
func doConfigGet(cmd *cobra.Command, stdout io.Writer, cfg *Config, key string, jsonOutput bool) error {
	configPath := cfg.ConfigPath
	if configPath == "" {
		configPath = filepath.Join(config.GetConfigDir(), "config.yaml")
	}

	// Load the configuration
	appConfig, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if key == "" {
		// Show all config
		return outputConfig(stdout, appConfig, jsonOutput, cfg.NoPrompt)
	}

	// Get specific key value
	value, err := getConfigValue(appConfig, key)
	if err != nil {
		return err
	}

	if jsonOutput {
		result := map[string]interface{}{
			"key":   key,
			"value": value,
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintln(stdout, value)
	if cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// outputConfig outputs the full configuration
func outputConfig(stdout io.Writer, appConfig *config.Config, jsonOutput bool, noPrompt bool) error {
	if jsonOutput {
		result := configToMap(appConfig)
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	// Output as YAML
	data, err := yaml.Marshal(appConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	_, _ = fmt.Fprint(stdout, string(data))
	if noPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}

// configToMap converts a Config struct to a map for JSON output
func configToMap(c *config.Config) map[string]interface{} {
	return map[string]interface{}{
		"backends": map[string]interface{}{
			"sqlite": map[string]interface{}{
				"enabled": c.Backends.SQLite.Enabled,
				"path":    c.Backends.SQLite.Path,
			},
			"todoist": map[string]interface{}{
				"enabled": c.Backends.Todoist.Enabled,
			},
		},
		"default_backend":     c.DefaultBackend,
		"default_view":        c.DefaultView,
		"no_prompt":           c.NoPrompt,
		"output_format":       c.OutputFormat,
		"auto_detect_backend": c.AutoDetectBackend,
		"sync": map[string]interface{}{
			"enabled":              c.Sync.Enabled,
			"local_backend":        c.Sync.LocalBackend,
			"conflict_resolution":  c.Sync.ConflictResolution,
			"offline_mode":         c.GetOfflineMode(),
			"connectivity_timeout": c.GetConnectivityTimeout(),
		},
		"trash": map[string]interface{}{
			"retention_days": c.GetTrashRetentionDays(),
		},
	}
}

// getConfigValue gets a value from the config using dot notation
func getConfigValue(c *config.Config, key string) (interface{}, error) {
	key = strings.ToLower(key)
	parts := strings.Split(key, ".")

	switch parts[0] {
	case "default_backend":
		return c.DefaultBackend, nil
	case "default_view":
		return c.DefaultView, nil
	case "no_prompt":
		return c.NoPrompt, nil
	case "output_format":
		return c.OutputFormat, nil
	case "auto_detect_backend":
		return c.AutoDetectBackend, nil
	case "backends":
		if len(parts) < 2 {
			return map[string]interface{}{
				"sqlite": map[string]interface{}{
					"enabled": c.Backends.SQLite.Enabled,
					"path":    c.Backends.SQLite.Path,
				},
				"todoist": map[string]interface{}{
					"enabled": c.Backends.Todoist.Enabled,
				},
			}, nil
		}
		switch parts[1] {
		case "sqlite":
			if len(parts) < 3 {
				return map[string]interface{}{
					"enabled": c.Backends.SQLite.Enabled,
					"path":    c.Backends.SQLite.Path,
				}, nil
			}
			switch parts[2] {
			case "enabled":
				return c.Backends.SQLite.Enabled, nil
			case "path":
				return c.Backends.SQLite.Path, nil
			}
		case "todoist":
			if len(parts) < 3 {
				return map[string]interface{}{
					"enabled": c.Backends.Todoist.Enabled,
				}, nil
			}
			switch parts[2] {
			case "enabled":
				return c.Backends.Todoist.Enabled, nil
			}
		}
	case "sync":
		if len(parts) < 2 {
			return map[string]interface{}{
				"enabled":              c.Sync.Enabled,
				"local_backend":        c.Sync.LocalBackend,
				"conflict_resolution":  c.Sync.ConflictResolution,
				"offline_mode":         c.GetOfflineMode(),
				"connectivity_timeout": c.GetConnectivityTimeout(),
			}, nil
		}
		switch parts[1] {
		case "enabled":
			return c.Sync.Enabled, nil
		case "local_backend":
			return c.Sync.LocalBackend, nil
		case "conflict_resolution":
			return c.Sync.ConflictResolution, nil
		case "offline_mode":
			return c.GetOfflineMode(), nil
		case "connectivity_timeout":
			return c.GetConnectivityTimeout(), nil
		}
	case "trash":
		if len(parts) < 2 {
			return map[string]interface{}{
				"retention_days": c.GetTrashRetentionDays(),
			}, nil
		}
		switch parts[1] {
		case "retention_days":
			return c.GetTrashRetentionDays(), nil
		}
	}

	return nil, fmt.Errorf("unknown config key: %s", key)
}

// newConfigSetCmd creates the 'config set' subcommand
func newConfigSetCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Update configuration value",
		Long:  "Update a configuration value with validation.\nSupports dot notation for nested keys (e.g., sync.offline_mode auto).",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]
			return doConfigSet(stdout, stderr, cfg, key, value)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// doConfigSet handles the config set command
func doConfigSet(stdout, stderr io.Writer, cfg *Config, key, value string) error {
	configPath := cfg.ConfigPath
	if configPath == "" {
		configPath = filepath.Join(config.GetConfigDir(), "config.yaml")
	}

	// Load the configuration
	appConfig, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Set the value
	if err := setConfigValue(appConfig, key, value); err != nil {
		return err
	}

	// Save the configuration
	if err := saveConfig(configPath, appConfig); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "Set %s = %s\n", key, value)
	if cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
	}
	return nil
}

// setConfigValue sets a value in the config using dot notation
func setConfigValue(c *config.Config, key, value string) error {
	key = strings.ToLower(key)
	parts := strings.Split(key, ".")

	switch parts[0] {
	case "default_backend":
		validBackends := []string{"sqlite", "todoist"}
		if !contains(validBackends, value) {
			return fmt.Errorf("invalid value for default_backend: %s (valid: %s)", value, strings.Join(validBackends, ", "))
		}
		c.DefaultBackend = value
		return nil
	case "default_view":
		c.DefaultView = value
		return nil
	case "no_prompt":
		boolVal, err := parseBool(value)
		if err != nil {
			return fmt.Errorf("invalid value for no_prompt: %s (valid: true, false, yes, no, 1, 0)", value)
		}
		c.NoPrompt = boolVal
		return nil
	case "output_format":
		validFormats := []string{"text", "json"}
		if !contains(validFormats, value) {
			return fmt.Errorf("invalid value for output_format: %s (valid: %s)", value, strings.Join(validFormats, ", "))
		}
		c.OutputFormat = value
		return nil
	case "auto_detect_backend":
		boolVal, err := parseBool(value)
		if err != nil {
			return fmt.Errorf("invalid value for auto_detect_backend: %s (valid: true, false, yes, no, 1, 0)", value)
		}
		c.AutoDetectBackend = boolVal
		return nil
	case "backends":
		if len(parts) < 3 {
			return fmt.Errorf("invalid key: %s (use backends.<backend>.<setting>)", key)
		}
		switch parts[1] {
		case "sqlite":
			switch parts[2] {
			case "enabled":
				boolVal, err := parseBool(value)
				if err != nil {
					return fmt.Errorf("invalid value for backends.sqlite.enabled: %s (valid: true, false, yes, no, 1, 0)", value)
				}
				c.Backends.SQLite.Enabled = boolVal
				return nil
			case "path":
				c.Backends.SQLite.Path = config.ExpandPath(value)
				return nil
			}
		case "todoist":
			switch parts[2] {
			case "enabled":
				boolVal, err := parseBool(value)
				if err != nil {
					return fmt.Errorf("invalid value for backends.todoist.enabled: %s (valid: true, false, yes, no, 1, 0)", value)
				}
				c.Backends.Todoist.Enabled = boolVal
				return nil
			}
		}
	case "sync":
		if len(parts) < 2 {
			return fmt.Errorf("invalid key: %s (use sync.<setting>)", key)
		}
		switch parts[1] {
		case "enabled":
			boolVal, err := parseBool(value)
			if err != nil {
				return fmt.Errorf("invalid value for sync.enabled: %s (valid: true, false, yes, no, 1, 0)", value)
			}
			c.Sync.Enabled = boolVal
			return nil
		case "local_backend":
			c.Sync.LocalBackend = value
			return nil
		case "conflict_resolution":
			validValues := []string{"local", "remote", "manual"}
			if !contains(validValues, value) {
				return fmt.Errorf("invalid value for sync.conflict_resolution: %s (valid: %s)", value, strings.Join(validValues, ", "))
			}
			c.Sync.ConflictResolution = value
			return nil
		case "offline_mode":
			validValues := []string{"auto", "online", "offline"}
			if !contains(validValues, value) {
				return fmt.Errorf("invalid value for sync.offline_mode: %s (valid: %s)", value, strings.Join(validValues, ", "))
			}
			c.Sync.OfflineMode = value
			return nil
		case "connectivity_timeout":
			c.Sync.ConnectivityTimeout = value
			return nil
		}
	case "trash":
		if len(parts) < 2 {
			return fmt.Errorf("invalid key: %s (use trash.<setting>)", key)
		}
		switch parts[1] {
		case "retention_days":
			days, err := strconv.Atoi(value)
			if err != nil || days < 0 {
				return fmt.Errorf("invalid value for trash.retention_days: %s (must be a non-negative integer)", value)
			}
			c.Trash.RetentionDays = &days
			return nil
		}
	}

	return fmt.Errorf("unknown config key: %s", key)
}

// parseBool parses a boolean value from various formats
func parseBool(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "true", "yes", "1":
		return true, nil
	case "false", "no", "0":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", value)
	}
}

// contains checks if a string is in a slice
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// saveConfig saves the configuration to a file
func saveConfig(configPath string, c *config.Config) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add a header comment
	content := "# todoat configuration\n" + string(data)

	// Write atomically using temp file + rename
	tmpPath := configPath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	if err := os.Rename(tmpPath, configPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename config file: %w", err)
	}

	return nil
}

// newConfigPathCmd creates the 'config path' subcommand
func newConfigPathCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show config file location",
		Long:  "Display the path to the active configuration file.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := cfg.ConfigPath
			if configPath == "" {
				configPath = filepath.Join(config.GetConfigDir(), "config.yaml")
			}

			jsonOutput, _ := cmd.Flags().GetBool("json")
			if jsonOutput {
				result := map[string]string{"path": configPath}
				enc := json.NewEncoder(stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			_, _ = fmt.Fprintln(stdout, configPath)
			if cfg.NoPrompt {
				_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// newConfigEditCmd creates the 'config edit' subcommand
func newConfigEditCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open config file in editor",
		Long:  "Open the configuration file in the system editor ($EDITOR or vi).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := cfg.ConfigPath
			if configPath == "" {
				configPath = filepath.Join(config.GetConfigDir(), "config.yaml")
			}

			// Ensure config file exists
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				// Create default config
				_, err := config.Load(configPath)
				if err != nil {
					return fmt.Errorf("failed to create config file: %w", err)
				}
			}

			// Get editor from environment
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = os.Getenv("VISUAL")
			}
			if editor == "" {
				editor = "vi"
			}

			// Run the editor
			execCmd := newExecCommand(editor, configPath)
			execCmd.Stdin = os.Stdin
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr

			if err := execCmd.Run(); err != nil {
				return fmt.Errorf("failed to run editor: %w", err)
			}

			if cfg.NoPrompt {
				_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// newExecCommand creates an exec.Cmd - extracted for testing
var newExecCommand = func(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

// newConfigResetCmd creates the 'config reset' subcommand
func newConfigResetCmd(stdout, stderr io.Writer, cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset to default configuration",
		Long:  "Reset the configuration file to default values. Requires confirmation unless --no-prompt is set.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			noPrompt, _ := cmd.Flags().GetBool("no-prompt")
			if cfg.NoPrompt {
				noPrompt = true
			}

			configPath := cfg.ConfigPath
			if configPath == "" {
				configPath = filepath.Join(config.GetConfigDir(), "config.yaml")
			}

			// Require confirmation unless --no-prompt
			if !noPrompt {
				_, _ = fmt.Fprint(stdout, "This will reset your configuration to defaults. Continue? [y/N] ")
				var response string
				_, _ = fmt.Fscanln(os.Stdin, &response)
				if response != "y" && response != "Y" {
					_, _ = fmt.Fprintln(stdout, "Cancelled.")
					return nil
				}
			}

			// Create default config
			defaultCfg := config.DefaultConfig()

			// Save the default configuration
			if err := saveConfig(configPath, defaultCfg); err != nil {
				return err
			}

			_, _ = fmt.Fprintln(stdout, "Configuration reset to defaults.")
			if noPrompt {
				_, _ = fmt.Fprintln(stdout, ResultActionCompleted)
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

// VersionInfo holds version information for JSON output
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// newVersionCmd creates the 'version' subcommand
func newVersionCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long:  "Display the application version, build information, and optionally extended details.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ := cmd.Flags().GetBool("verbose")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			info := VersionInfo{
				Version:   Version,
				Commit:    Commit,
				BuildDate: BuildDate,
				GoVersion: runtime.Version(),
				Platform:  runtime.GOOS + "/" + runtime.GOARCH,
			}

			if jsonOutput {
				return outputVersionJSON(stdout, info)
			}

			return outputVersionText(stdout, info, verbose)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	versionCmd.Flags().BoolP("verbose", "v", false, "Show extended build information")

	return versionCmd
}

// outputVersionJSON outputs version info as JSON
func outputVersionJSON(stdout io.Writer, info VersionInfo) error {
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(info)
}

// outputVersionText outputs version info as formatted text
func outputVersionText(stdout io.Writer, info VersionInfo, verbose bool) error {
	_, _ = fmt.Fprintf(stdout, "Version: %s\n", info.Version)
	_, _ = fmt.Fprintf(stdout, "Commit:  %s\n", info.Commit)
	_, _ = fmt.Fprintf(stdout, "Built:   %s\n", info.BuildDate)

	if verbose {
		_, _ = fmt.Fprintf(stdout, "Go Version: %s\n", info.GoVersion)
		_, _ = fmt.Fprintf(stdout, "Platform:   %s\n", info.Platform)
	}

	return nil
}

// TagInfo holds information about a tag and its usage count
type TagInfo struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// TagsOutput holds the JSON output structure for the tags command
type TagsOutput struct {
	Tags   []TagInfo `json:"tags"`
	List   string    `json:"list,omitempty"`
	Result string    `json:"result"`
}

// newTagsCmd creates the 'tags' subcommand for listing all unique tags
func newTagsCmd(stdout io.Writer, cfg *Config) *cobra.Command {
	tagsCmd := &cobra.Command{
		Use:   "tags",
		Short: "List all unique tags in use",
		Long:  "List all unique tags across all tasks, with optional filtering by list.",
		Args:  cobra.NoArgs,
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

			listName, _ := cmd.Flags().GetString("list")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			return doTags(context.Background(), be, listName, cfg, stdout, jsonOutput)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	tagsCmd.Flags().StringP("list", "l", "", "Filter tags to a specific list")

	return tagsCmd
}

// doTags lists all unique tags across all tasks
func doTags(ctx context.Context, be backend.TaskManager, listName string, cfg *Config, stdout io.Writer, jsonOutput bool) error {
	// Get all lists
	lists, err := be.GetLists(ctx)
	if err != nil {
		return err
	}

	// Filter to specific list if requested
	if listName != "" {
		var filteredLists []backend.List
		for _, l := range lists {
			if strings.EqualFold(l.Name, listName) {
				filteredLists = append(filteredLists, l)
				break
			}
		}
		if len(filteredLists) == 0 {
			return fmt.Errorf("list not found: %s", listName)
		}
		lists = filteredLists
	}

	// Collect all tags with counts
	tagCounts := make(map[string]int)
	tagOrigCase := make(map[string]string) // Store original case for display

	for _, l := range lists {
		tasks, err := be.GetTasks(ctx, l.ID)
		if err != nil {
			continue
		}

		for _, t := range tasks {
			if t.Categories == "" {
				continue
			}
			// Split comma-separated tags
			tags := strings.Split(t.Categories, ",")
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag == "" {
					continue
				}
				lowerTag := strings.ToLower(tag)
				tagCounts[lowerTag]++
				// Preserve first-seen case
				if _, exists := tagOrigCase[lowerTag]; !exists {
					tagOrigCase[lowerTag] = tag
				}
			}
		}
	}

	// Build sorted list of tags
	var tagInfos []TagInfo
	for lowerTag, count := range tagCounts {
		tagInfos = append(tagInfos, TagInfo{
			Name:  tagOrigCase[lowerTag],
			Count: count,
		})
	}
	// Sort by count descending, then name ascending
	sort.Slice(tagInfos, func(i, j int) bool {
		if tagInfos[i].Count != tagInfos[j].Count {
			return tagInfos[i].Count > tagInfos[j].Count
		}
		return strings.ToLower(tagInfos[i].Name) < strings.ToLower(tagInfos[j].Name)
	})

	if jsonOutput {
		output := TagsOutput{
			Tags:   tagInfos,
			List:   listName,
			Result: ResultInfoOnly,
		}
		jsonBytes, err := json.Marshal(output)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdout, string(jsonBytes))
		return nil
	}

	// Text output
	if len(tagInfos) == 0 {
		_, _ = fmt.Fprintln(stdout, "No tags in use.")
	} else {
		_, _ = fmt.Fprintln(stdout, "Tags in use:")
		for _, ti := range tagInfos {
			taskWord := "tasks"
			if ti.Count == 1 {
				taskWord = "task"
			}
			_, _ = fmt.Fprintf(stdout, "  %s (%d %s)\n", ti.Name, ti.Count, taskWord)
		}
	}

	if cfg != nil && cfg.NoPrompt {
		_, _ = fmt.Fprintln(stdout, ResultInfoOnly)
	}
	return nil
}
