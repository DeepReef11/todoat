// Package reminder provides task reminder notification services.
package reminder

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"todoat/backend"
	"todoat/internal/notification"

	_ "modernc.org/sqlite"
)

// Config holds the reminder configuration
type Config struct {
	Enabled         bool     `yaml:"enabled" json:"enabled"`
	Intervals       []string `yaml:"intervals" json:"intervals"`
	OSNotification  bool     `yaml:"os_notification" json:"os_notification"`
	LogNotification bool     `yaml:"log_notification" json:"log_notification"`
}

// Service manages task reminders
type Service struct {
	config   *Config
	db       *sql.DB
	notifier notification.NotificationManager
}

// ReminderRecord represents a reminder stored in the database
type ReminderRecord struct {
	TaskID      string
	Interval    string
	DismissedAt *time.Time
	DisabledAt  *time.Time
}

// UpcomingReminder represents a task with reminder information
type UpcomingReminder struct {
	Task        *backend.Task
	Interval    string
	TriggerTime time.Time
}

// NewService creates a new reminder service
func NewService(cfg *Config, dbPath string) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create reminders table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS reminders (
			task_id TEXT NOT NULL,
			interval TEXT NOT NULL,
			dismissed_at DATETIME,
			disabled_at DATETIME,
			PRIMARY KEY (task_id, interval)
		)
	`)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create reminders table: %w", err)
	}

	// Create task_reminder_settings table for per-task disable
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS task_reminder_settings (
			task_id TEXT PRIMARY KEY,
			disabled BOOLEAN DEFAULT FALSE,
			disabled_at DATETIME
		)
	`)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create task_reminder_settings table: %w", err)
	}

	return &Service{
		config: cfg,
		db:     db,
	}, nil
}

// Close releases resources
func (s *Service) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// SetNotifier sets the notification manager for sending reminders
func (s *Service) SetNotifier(notifier notification.NotificationManager) {
	s.notifier = notifier
}

// CheckReminders checks tasks for due reminders and sends notifications
func (s *Service) CheckReminders(tasks []*backend.Task) ([]*backend.Task, error) {
	if !s.config.Enabled {
		return nil, nil
	}

	var triggered []*backend.Task
	seen := make(map[string]bool)
	now := time.Now()

	for _, task := range tasks {
		if task.DueDate == nil || task.Status == backend.StatusCompleted {
			continue
		}

		// Deduplicate by task ID
		if seen[task.ID] {
			continue
		}

		// Check if task-level reminders are disabled
		disabled, err := s.isTaskDisabled(task.ID)
		if err != nil {
			return nil, err
		}
		if disabled {
			continue
		}

		// Check each interval
		for _, intervalStr := range s.config.Intervals {
			duration, isAtDue, err := ParseInterval(intervalStr)
			if err != nil {
				continue
			}

			// Check if reminder should trigger
			shouldTrigger := false
			if isAtDue {
				// "at due time" - trigger if due date is today
				// Compare year, month, day in local time (not UTC truncation)
				dueY, dueM, dueD := task.DueDate.Date()
				nowY, nowM, nowD := now.Date()
				shouldTrigger = dueY == nowY && dueM == nowM && dueD == nowD
			} else {
				// Duration-based - trigger if within window
				timeUntilDue := task.DueDate.Sub(now)
				shouldTrigger = timeUntilDue >= 0 && timeUntilDue <= duration
			}

			if !shouldTrigger {
				continue
			}

			// Check if already dismissed for this interval
			dismissed, err := s.isDismissed(task.ID, intervalStr)
			if err != nil {
				return nil, err
			}
			if dismissed {
				continue
			}

			// Trigger reminder
			triggered = append(triggered, task)
			seen[task.ID] = true

			// Send notification
			if s.notifier != nil {
				notif := notification.Notification{
					Type:      notification.NotifyReminder,
					Title:     "Task Reminder",
					Message:   fmt.Sprintf("%s - Due: %s", task.Summary, task.DueDate.Format("2006-01-02")),
					Timestamp: now,
					Metadata: map[string]string{
						"task_id":  task.ID,
						"interval": intervalStr,
					},
				}
				_ = s.notifier.Send(notif)
			}

			// Mark as dismissed for this interval (to avoid re-triggering)
			_ = s.DismissReminder(task.ID, intervalStr)

			// Only trigger once per task
			break
		}
	}

	return triggered, nil
}

// DismissReminder marks a reminder as dismissed for a specific interval
func (s *Service) DismissReminder(taskID string, interval string) error {
	now := time.Now()
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO reminders (task_id, interval, dismissed_at)
		VALUES (?, ?, ?)
	`, taskID, interval, now)
	return err
}

// DisableReminder permanently disables reminders for a task
func (s *Service) DisableReminder(taskID string) error {
	now := time.Now()
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO task_reminder_settings (task_id, disabled, disabled_at)
		VALUES (?, TRUE, ?)
	`, taskID, now)
	return err
}

// EnableReminder re-enables reminders for a task
func (s *Service) EnableReminder(taskID string) error {
	_, err := s.db.Exec(`
		DELETE FROM task_reminder_settings WHERE task_id = ?
	`, taskID)
	return err
}

// GetUpcomingReminders returns tasks with upcoming reminders
func (s *Service) GetUpcomingReminders(tasks []*backend.Task) ([]*backend.Task, error) {
	if !s.config.Enabled || len(s.config.Intervals) == 0 {
		return nil, nil
	}

	// Find the maximum interval to use as the window
	var maxDuration time.Duration
	for _, intervalStr := range s.config.Intervals {
		duration, isAtDue, err := ParseInterval(intervalStr)
		if err != nil {
			continue
		}
		if isAtDue {
			// "at due time" effectively has 0 advance warning
			continue
		}
		if duration > maxDuration {
			maxDuration = duration
		}
	}

	var upcoming []*backend.Task
	seen := make(map[string]bool)
	now := time.Now()

	for _, task := range tasks {
		if task.DueDate == nil || task.Status == backend.StatusCompleted {
			continue
		}

		// Deduplicate by task ID
		if seen[task.ID] {
			continue
		}

		// Check if task-level reminders are disabled
		disabled, err := s.isTaskDisabled(task.ID)
		if err != nil {
			return nil, err
		}
		if disabled {
			continue
		}

		// Check if due date is within the max window
		timeUntilDue := task.DueDate.Sub(now)
		if timeUntilDue >= 0 && timeUntilDue <= maxDuration {
			upcoming = append(upcoming, task)
			seen[task.ID] = true
		}
	}

	return upcoming, nil
}

// isDismissed checks if a reminder has been dismissed for a specific interval
func (s *Service) isDismissed(taskID, interval string) (bool, error) {
	var dismissedAt sql.NullTime
	err := s.db.QueryRow(`
		SELECT dismissed_at FROM reminders
		WHERE task_id = ? AND interval = ?
	`, taskID, interval).Scan(&dismissedAt)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return dismissedAt.Valid, nil
}

// isTaskDisabled checks if reminders are disabled for a task
func (s *Service) isTaskDisabled(taskID string) (bool, error) {
	var disabled bool
	err := s.db.QueryRow(`
		SELECT disabled FROM task_reminder_settings
		WHERE task_id = ?
	`, taskID).Scan(&disabled)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return disabled, nil
}

// ParseInterval parses an interval string and returns the duration.
// Returns (duration, isAtDueTime, error).
// isAtDueTime is true for "at due time" which means trigger on the due date itself.
// Supports both shorthand formats (1d, 1h, 15m, 1w) and full word formats (1 day, 1 hour, 15 minutes, 1 week).
func ParseInterval(interval string) (time.Duration, bool, error) {
	interval = strings.TrimSpace(strings.ToLower(interval))

	if interval == "at due time" {
		return 0, true, nil
	}

	// Parse patterns like "1d", "1h", "15m", "1w" (shorthand) and "1 day", "1 hour", "15 minutes", "1 week" (full word)
	re := regexp.MustCompile(`^(\d+)\s*(d|day|days|h|hour|hours|m|min|minute|minutes|w|week|weeks)$`)
	matches := re.FindStringSubmatch(interval)
	if matches == nil {
		return 0, false, fmt.Errorf("invalid interval format: %s", interval)
	}

	num, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	var duration time.Duration
	switch unit {
	case "d", "day", "days":
		duration = time.Duration(num) * 24 * time.Hour
	case "h", "hour", "hours":
		duration = time.Duration(num) * time.Hour
	case "m", "min", "minute", "minutes":
		duration = time.Duration(num) * time.Minute
	case "w", "week", "weeks":
		duration = time.Duration(num) * 7 * 24 * time.Hour
	default:
		return 0, false, fmt.Errorf("unknown time unit: %s", unit)
	}

	return duration, false, nil
}

// GetTaskByID finds a task by ID in a list of tasks
func GetTaskByID(tasks []*backend.Task, taskID string) *backend.Task {
	for _, t := range tasks {
		if t.ID == taskID {
			return t
		}
	}
	return nil
}

// GetTaskBySummary finds a task by summary in a list of tasks
func GetTaskBySummary(tasks []*backend.Task, summary string) *backend.Task {
	summary = strings.ToLower(strings.TrimSpace(summary))
	for _, t := range tasks {
		if strings.ToLower(strings.TrimSpace(t.Summary)) == summary {
			return t
		}
	}
	return nil
}
