package reminder_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"todoat/backend"
	"todoat/internal/notification"
	"todoat/internal/reminder"
	"todoat/internal/testutil"
)

// =============================================================================
// CLI Tests (032-task-reminders)
// =============================================================================

// TestReminderCheckCLI tests the 'todoat reminder check' command
func TestReminderCheckCLI(t *testing.T) {
	t.Run("returns success when no reminders due", func(t *testing.T) {
		cli := testutil.NewCLITestWithReminder(t)

		cli.SetReminderConfig(&reminder.Config{
			Enabled: true,
			Intervals: []string{
				"1 day",
			},
			OSNotification:  false,
			LogNotification: true,
		})

		// Create a task with due date far in the future (outside 1-day window)
		dueDate := time.Now().AddDate(0, 0, 30).Format("2006-01-02")
		cli.MustExecute("-y", "Work", "add", "Future task", "--due-date", dueDate)

		// Check reminders - should show "No reminders triggered"
		stdout := cli.MustExecute("-y", "reminder", "check")

		testutil.AssertContains(t, stdout, "No reminders triggered")
		testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
	})

	t.Run("outputs list of due reminders when present", func(t *testing.T) {
		cli := testutil.NewCLITestWithReminder(t)

		cli.SetReminderConfig(&reminder.Config{
			Enabled: true,
			Intervals: []string{
				"1 day",
			},
			OSNotification:  false,
			LogNotification: true,
		})

		// Create a task due tomorrow (within 1-day window)
		dueDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
		cli.MustExecute("-y", "Work", "add", "Soon task", "--due-date", dueDate)

		// Check reminders - should show triggered reminders
		stdout := cli.MustExecute("-y", "reminder", "check")

		testutil.AssertContains(t, stdout, "Triggered")
		testutil.AssertContains(t, stdout, "Soon task")
		testutil.AssertContains(t, stdout, "reminder(s)")
		testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
	})

	t.Run("handles multiple tasks with different due dates", func(t *testing.T) {
		cli := testutil.NewCLITestWithReminder(t)

		cli.SetReminderConfig(&reminder.Config{
			Enabled: true,
			Intervals: []string{
				"7 days",
			},
			OSNotification:  false,
			LogNotification: true,
		})

		// Create tasks with various due dates
		dueSoon := time.Now().AddDate(0, 0, 2).Format("2006-01-02")    // Within 7-day window
		dueLater := time.Now().AddDate(0, 0, 5).Format("2006-01-02")   // Within 7-day window
		dueFuture := time.Now().AddDate(0, 0, 30).Format("2006-01-02") // Outside window

		cli.MustExecute("-y", "Work", "add", "Task Soon", "--due-date", dueSoon)
		cli.MustExecute("-y", "Work", "add", "Task Later", "--due-date", dueLater)
		cli.MustExecute("-y", "Work", "add", "Task Future", "--due-date", dueFuture)

		// Check reminders
		stdout := cli.MustExecute("-y", "reminder", "check")

		// Tasks within window should appear
		testutil.AssertContains(t, stdout, "Task Soon")
		testutil.AssertContains(t, stdout, "Task Later")
		// Task outside window should not appear
		testutil.AssertNotContains(t, stdout, "Task Future")
		testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
	})

	t.Run("works when reminders are disabled", func(t *testing.T) {
		cli := testutil.NewCLITestWithReminder(t)

		// Note: disabled reminders config
		cli.SetReminderConfig(&reminder.Config{
			Enabled: false,
			Intervals: []string{
				"1 day",
			},
			OSNotification:  false,
			LogNotification: false,
		})

		// Create a task due tomorrow
		dueDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
		cli.MustExecute("-y", "Work", "add", "Due task", "--due-date", dueDate)

		// Check reminders - should not trigger when disabled
		stdout := cli.MustExecute("-y", "reminder", "check")

		testutil.AssertContains(t, stdout, "No reminders triggered")
		testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)
	})
}

// TestReminderStatusCLI tests the 'todoat reminder status' command
func TestReminderStatusCLI(t *testing.T) {
	t.Run("shows enabled status with intervals", func(t *testing.T) {
		cli := testutil.NewCLITestWithReminder(t)

		cli.SetReminderConfig(&reminder.Config{
			Enabled: true,
			Intervals: []string{
				"7 days",
				"1 day",
				"at due time",
			},
			OSNotification:  true,
			LogNotification: true,
		})

		stdout := cli.MustExecute("-y", "reminder", "status")

		// Verify command succeeds with info result code
		testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)

		// Verify status shows enabled
		testutil.AssertContains(t, stdout, "enabled")

		// Verify all configured intervals are shown
		testutil.AssertContains(t, stdout, "7 days")
		testutil.AssertContains(t, stdout, "1 day")
		testutil.AssertContains(t, stdout, "at due time")

		// Verify notification settings are shown
		testutil.AssertContains(t, stdout, "OS Notification")
		testutil.AssertContains(t, stdout, "Log Notification")
	})

	t.Run("shows disabled status", func(t *testing.T) {
		cli := testutil.NewCLITestWithReminder(t)

		cli.SetReminderConfig(&reminder.Config{
			Enabled: false,
			Intervals: []string{
				"1 day",
			},
			OSNotification:  false,
			LogNotification: false,
		})

		stdout := cli.MustExecute("-y", "reminder", "status")

		// Verify command succeeds
		testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)

		// Verify status shows disabled
		testutil.AssertContains(t, stdout, "disabled")
	})

	t.Run("shows notification configuration", func(t *testing.T) {
		cli := testutil.NewCLITestWithReminder(t)

		cli.SetReminderConfig(&reminder.Config{
			Enabled: true,
			Intervals: []string{
				"1 hour",
			},
			OSNotification:  true,
			LogNotification: false,
		})

		stdout := cli.MustExecute("-y", "reminder", "status")

		// Verify OS notification setting
		testutil.AssertContains(t, stdout, "OS Notification: true")
		// Verify Log notification setting
		testutil.AssertContains(t, stdout, "Log Notification: false")
	})
}

// TestReminderConfig tests that reminder settings can be configured in config.yaml
func TestReminderConfig(t *testing.T) {
	cli := testutil.NewCLITestWithReminder(t)

	// Set up a config with reminder settings
	cli.SetReminderConfig(&reminder.Config{
		Enabled: true,
		Intervals: []string{
			"1 day",
			"1 hour",
			"at due time",
		},
		OSNotification:  true,
		LogNotification: true,
	})

	// Verify config is loaded properly by listing reminder status
	stdout := cli.MustExecute("-y", "reminder", "status")

	testutil.AssertContains(t, stdout, "enabled")
	testutil.AssertContains(t, stdout, "1 day")
	testutil.AssertContains(t, stdout, "1 hour")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestReminderNotification tests that reminder notification is sent when task due date approaches
func TestReminderNotification(t *testing.T) {
	cli := testutil.NewCLITestWithReminder(t)

	// Enable reminders with 1-day interval
	cli.SetReminderConfig(&reminder.Config{
		Enabled: true,
		Intervals: []string{
			"1 day",
		},
		OSNotification:  true,
		LogNotification: true,
	})

	// Add a task due tomorrow (within 1-day reminder window)
	dueDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	cli.MustExecute("-y", "Work", "add", "Upcoming task", "--due-date", dueDate)

	// Check reminders - should trigger notification
	stdout := cli.MustExecute("-y", "reminder", "check")

	testutil.AssertContains(t, stdout, "Upcoming task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify notification was sent (check notification log)
	notifLog := cli.GetNotificationLog()
	if !strings.Contains(notifLog, "REMINDER") {
		t.Errorf("expected notification log to contain REMINDER, got:\n%s", notifLog)
	}
	if !strings.Contains(notifLog, "Upcoming task") {
		t.Errorf("expected notification log to contain task name, got:\n%s", notifLog)
	}
}

// TestReminderIntervals tests configurable intervals (1 day, 1 hour, at due time)
func TestReminderIntervals(t *testing.T) {
	tests := []struct {
		name          string
		intervals     []string
		dueInDays     int
		shouldTrigger bool
	}{
		{
			name:          "1 day interval - task due tomorrow",
			intervals:     []string{"1 day"},
			dueInDays:     1,
			shouldTrigger: true,
		},
		{
			name:          "1 day interval - task due in 3 days",
			intervals:     []string{"1 day"},
			dueInDays:     3,
			shouldTrigger: false,
		},
		{
			name:          "3 days interval - task due in 2 days",
			intervals:     []string{"3 days"},
			dueInDays:     2,
			shouldTrigger: true,
		},
		{
			name:          "at due time interval - task due today",
			intervals:     []string{"at due time"},
			dueInDays:     0,
			shouldTrigger: true,
		},
		{
			name:          "at due time interval - task due tomorrow",
			intervals:     []string{"at due time"},
			dueInDays:     1,
			shouldTrigger: false, // "at due time" only triggers on due date
		},
		{
			name:          "multiple intervals - uses any matching",
			intervals:     []string{"3 days", "1 day", "at due time"},
			dueInDays:     2,
			shouldTrigger: true, // 2 days is within 3-day window
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := testutil.NewCLITestWithReminder(t)

			cli.SetReminderConfig(&reminder.Config{
				Enabled:         true,
				Intervals:       tt.intervals,
				OSNotification:  true,
				LogNotification: true,
			})

			// Add task with calculated due date
			dueDate := time.Now().AddDate(0, 0, tt.dueInDays).Format("2006-01-02")
			cli.MustExecute("-y", "Work", "add", "Test task", "--due-date", dueDate)

			// Check reminders
			stdout := cli.MustExecute("-y", "reminder", "check")

			notifLog := cli.GetNotificationLog()
			hasReminder := strings.Contains(notifLog, "REMINDER")

			if tt.shouldTrigger && !hasReminder {
				t.Errorf("expected reminder to trigger, but it didn't. Log:\n%s", notifLog)
			}
			if !tt.shouldTrigger && hasReminder {
				t.Errorf("expected reminder NOT to trigger, but it did. Log:\n%s", notifLog)
			}
			_ = stdout // Used for exit code check
		})
	}
}

// TestReminderDisable tests that individual reminders can be disabled
func TestReminderDisable(t *testing.T) {
	cli := testutil.NewCLITestWithReminder(t)

	cli.SetReminderConfig(&reminder.Config{
		Enabled: true,
		Intervals: []string{
			"1 day",
		},
		OSNotification:  true,
		LogNotification: true,
	})

	// Add a task due tomorrow
	dueDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	cli.MustExecute("-y", "Work", "add", "Task with reminder", "--due-date", dueDate)

	// Disable reminder for this specific task
	stdout := cli.MustExecute("-y", "reminder", "disable", "Task with reminder")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Check reminders - should NOT trigger for disabled task
	cli.MustExecute("-y", "reminder", "check")

	notifLog := cli.GetNotificationLog()
	if strings.Contains(notifLog, "Task with reminder") {
		t.Errorf("expected disabled task reminder NOT to appear in log, got:\n%s", notifLog)
	}
}

// TestReminderList tests that 'todoat reminder list' shows upcoming reminders
func TestReminderList(t *testing.T) {
	cli := testutil.NewCLITestWithReminder(t)

	cli.SetReminderConfig(&reminder.Config{
		Enabled: true,
		Intervals: []string{
			"7 days",
			"1 day",
		},
		OSNotification:  true,
		LogNotification: true,
	})

	// Add multiple tasks with different due dates
	dueDate1 := time.Now().AddDate(0, 0, 1).Format("2006-01-02") // Tomorrow
	dueDate2 := time.Now().AddDate(0, 0, 3).Format("2006-01-02") // 3 days
	dueDate3 := time.Now().AddDate(0, 0, 5).Format("2006-01-02") // 5 days

	cli.MustExecute("-y", "Work", "add", "Soon task", "--due-date", dueDate1)
	cli.MustExecute("-y", "Work", "add", "Later task", "--due-date", dueDate2)
	cli.MustExecute("-y", "Work", "add", "Much later task", "--due-date", dueDate3)

	// List upcoming reminders
	stdout := cli.MustExecute("-y", "reminder", "list")

	// All tasks should appear (all within 7-day window)
	testutil.AssertContains(t, stdout, "Soon task")
	testutil.AssertContains(t, stdout, "Later task")
	testutil.AssertContains(t, stdout, "Much later task")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}

// TestReminderDismiss tests that 'todoat reminder dismiss <task>' dismisses a reminder
func TestReminderDismiss(t *testing.T) {
	cli := testutil.NewCLITestWithReminder(t)

	cli.SetReminderConfig(&reminder.Config{
		Enabled: true,
		Intervals: []string{
			"1 day",
		},
		OSNotification:  true,
		LogNotification: true,
	})

	// Add a task due tomorrow
	dueDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	cli.MustExecute("-y", "Work", "add", "Dismissable task", "--due-date", dueDate)

	// First reminder check - should trigger
	cli.MustExecute("-y", "reminder", "check")
	notifLog1 := cli.GetNotificationLog()
	if !strings.Contains(notifLog1, "Dismissable task") {
		t.Errorf("expected first reminder to trigger, got:\n%s", notifLog1)
	}

	// Clear notification log for clean test
	cli.ClearNotificationLog()

	// Dismiss the reminder
	stdout := cli.MustExecute("-y", "reminder", "dismiss", "Dismissable task")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Second reminder check - should NOT trigger (dismissed)
	cli.MustExecute("-y", "reminder", "check")
	notifLog2 := cli.GetNotificationLog()
	if strings.Contains(notifLog2, "Dismissable task") {
		t.Errorf("expected dismissed reminder NOT to trigger again, got:\n%s", notifLog2)
	}
}

// =============================================================================
// Notification Integration Tests
// =============================================================================

// TestReminderOSNotification tests that reminders are sent via OS notification
func TestReminderOSNotification(t *testing.T) {
	var sentNotifications []notification.Notification

	cli := testutil.NewCLITestWithReminder(t)
	cli.SetNotificationCallback(func(n interface{}) {
		if notif, ok := n.(notification.Notification); ok {
			sentNotifications = append(sentNotifications, notif)
		}
	})

	cli.SetReminderConfig(&reminder.Config{
		Enabled: true,
		Intervals: []string{
			"1 day",
		},
		OSNotification:  true,
		LogNotification: false, // Only OS
	})

	// Add a task due tomorrow
	dueDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	cli.MustExecute("-y", "Work", "add", "OS notify task", "--due-date", dueDate)

	// Check reminders
	cli.MustExecute("-y", "reminder", "check")

	// Verify OS notification was sent
	found := false
	for _, n := range sentNotifications {
		if n.Type == notification.NotifyReminder && strings.Contains(n.Message, "OS notify task") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected OS notification for reminder")
	}
}

// TestReminderLogNotification tests that reminders are logged to notification log
func TestReminderLogNotification(t *testing.T) {
	cli := testutil.NewCLITestWithReminder(t)

	cli.SetReminderConfig(&reminder.Config{
		Enabled: true,
		Intervals: []string{
			"1 day",
		},
		OSNotification:  false, // Only log
		LogNotification: true,
	})

	// Add a task due tomorrow
	dueDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	cli.MustExecute("-y", "Work", "add", "Log notify task", "--due-date", dueDate)

	// Check reminders
	cli.MustExecute("-y", "reminder", "check")

	// Verify log notification was written
	notifLog := cli.GetNotificationLog()
	if !strings.Contains(notifLog, "[REMINDER]") {
		t.Errorf("expected log to contain [REMINDER], got:\n%s", notifLog)
	}
	if !strings.Contains(notifLog, "Log notify task") {
		t.Errorf("expected log to contain task name, got:\n%s", notifLog)
	}
}

// TestReminderFormat tests that notification includes task summary and due date
func TestReminderFormat(t *testing.T) {
	cli := testutil.NewCLITestWithReminder(t)

	cli.SetReminderConfig(&reminder.Config{
		Enabled: true,
		Intervals: []string{
			"1 day",
		},
		OSNotification:  true,
		LogNotification: true,
	})

	// Add a task with specific due date
	dueDate := time.Now().AddDate(0, 0, 1)
	dueDateStr := dueDate.Format("2006-01-02")
	cli.MustExecute("-y", "Work", "add", "Formatted task", "--due-date", dueDateStr)

	// Check reminders
	cli.MustExecute("-y", "reminder", "check")

	// Verify notification format contains task summary and due date
	notifLog := cli.GetNotificationLog()
	if !strings.Contains(notifLog, "Formatted task") {
		t.Errorf("expected notification to contain task summary, got:\n%s", notifLog)
	}
	// Due date should be included
	if !strings.Contains(notifLog, dueDateStr) {
		t.Errorf("expected notification to contain due date %s, got:\n%s", dueDateStr, notifLog)
	}
}

// =============================================================================
// Unit Tests - Reminder Service
// =============================================================================

// TestReminderServiceCheckReminders tests the reminder service logic
func TestReminderServiceCheckReminders(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create reminder service
	service, err := reminder.NewService(&reminder.Config{
		Enabled: true,
		Intervals: []string{
			"1 day",
		},
		OSNotification:  true,
		LogNotification: true,
	}, dbPath)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	defer func() { _ = service.Close() }()

	// Create mock notifier
	var sentNotifications []notification.Notification
	mockNotifier := &mockNotificationManager{
		sendFunc: func(n notification.Notification) error {
			sentNotifications = append(sentNotifications, n)
			return nil
		},
	}
	service.SetNotifier(mockNotifier)

	// Add task due tomorrow (within 1-day window)
	dueTime := time.Now().AddDate(0, 0, 1)
	task := &backend.Task{
		ID:      "test-task-1",
		Summary: "Test reminder task",
		DueDate: &dueTime,
		Status:  backend.StatusNeedsAction,
	}

	// Check reminders
	triggered, err := service.CheckReminders([]*backend.Task{task})
	if err != nil {
		t.Fatalf("CheckReminders failed: %v", err)
	}

	if len(triggered) != 1 {
		t.Errorf("expected 1 triggered reminder, got %d", len(triggered))
	}
	if len(sentNotifications) != 1 {
		t.Errorf("expected 1 notification, got %d", len(sentNotifications))
	}
}

// TestReminderServiceDismissedReminders tests that dismissed reminders are not triggered
func TestReminderServiceDismissedReminders(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	service, err := reminder.NewService(&reminder.Config{
		Enabled: true,
		Intervals: []string{
			"1 day",
		},
		OSNotification:  true,
		LogNotification: true,
	}, dbPath)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	defer func() { _ = service.Close() }()

	var sentNotifications []notification.Notification
	mockNotifier := &mockNotificationManager{
		sendFunc: func(n notification.Notification) error {
			sentNotifications = append(sentNotifications, n)
			return nil
		},
	}
	service.SetNotifier(mockNotifier)

	dueTime := time.Now().AddDate(0, 0, 1) // Tomorrow
	task := &backend.Task{
		ID:      "test-task-2",
		Summary: "Dismissed task",
		DueDate: &dueTime,
		Status:  backend.StatusNeedsAction,
	}

	// Dismiss the reminder before checking
	err = service.DismissReminder(task.ID, "1 day")
	if err != nil {
		t.Fatalf("DismissReminder failed: %v", err)
	}

	// Check reminders - should not trigger
	triggered, err := service.CheckReminders([]*backend.Task{task})
	if err != nil {
		t.Fatalf("CheckReminders failed: %v", err)
	}

	if len(triggered) != 0 {
		t.Errorf("expected 0 triggered reminders for dismissed task, got %d", len(triggered))
	}
	if len(sentNotifications) != 0 {
		t.Errorf("expected 0 notifications for dismissed task, got %d", len(sentNotifications))
	}
}

// TestParseInterval tests parsing of interval strings
func TestParseInterval(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		isAtDue  bool
		wantErr  bool
	}{
		// Full word formats
		{"1 day", 24 * time.Hour, false, false},
		{"3 days", 3 * 24 * time.Hour, false, false},
		{"7 days", 7 * 24 * time.Hour, false, false},
		{"1 week", 7 * 24 * time.Hour, false, false},
		{"1 hour", time.Hour, false, false},
		{"2 hours", 2 * time.Hour, false, false},
		{"15 minutes", 15 * time.Minute, false, false},
		{"30 minute", 30 * time.Minute, false, false},
		// Shorthand formats (documented in docs/reference/configuration.md)
		{"1d", 24 * time.Hour, false, false},
		{"3d", 3 * 24 * time.Hour, false, false},
		{"7d", 7 * 24 * time.Hour, false, false},
		{"1w", 7 * 24 * time.Hour, false, false},
		{"1h", time.Hour, false, false},
		{"2h", 2 * time.Hour, false, false},
		{"15m", 15 * time.Minute, false, false},
		{"30m", 30 * time.Minute, false, false},
		// At due time
		{"at due time", 0, true, false},
		// Invalid formats
		{"invalid", 0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			duration, isAtDue, err := reminder.ParseInterval(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if duration != tt.expected {
				t.Errorf("expected duration %v, got %v", tt.expected, duration)
			}
			if isAtDue != tt.isAtDue {
				t.Errorf("expected isAtDue %v, got %v", tt.isAtDue, isAtDue)
			}
		})
	}
}

// TestReminderServiceNoDuplicates tests that duplicate tasks in input produce no duplicate reminders (Issue #69)
func TestReminderServiceNoDuplicates(t *testing.T) {
	t.Run("CheckReminders deduplicates by task ID", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		service, err := reminder.NewService(&reminder.Config{
			Enabled: true,
			Intervals: []string{
				"1 day",
				"at due time",
			},
			OSNotification:  true,
			LogNotification: true,
		}, dbPath)
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}
		defer func() { _ = service.Close() }()

		var sentNotifications []notification.Notification
		mockNotifier := &mockNotificationManager{
			sendFunc: func(n notification.Notification) error {
				sentNotifications = append(sentNotifications, n)
				return nil
			},
		}
		service.SetNotifier(mockNotifier)

		dueTime := time.Now().AddDate(0, 0, 1) // Tomorrow
		taskA := &backend.Task{
			ID:      "task-dup-1",
			Summary: "Do tomorrow",
			DueDate: &dueTime,
			Status:  backend.StatusNeedsAction,
		}
		taskB := &backend.Task{
			ID:      "task-dup-2",
			Summary: "Weekly review",
			DueDate: &dueTime,
			Status:  backend.StatusNeedsAction,
		}

		// Simulate getAllTasks returning duplicate entries (same task appearing twice)
		tasks := []*backend.Task{taskA, taskB, taskA, taskB}

		triggered, err := service.CheckReminders(tasks)
		if err != nil {
			t.Fatalf("CheckReminders failed: %v", err)
		}

		// Each task should appear only once despite being passed twice
		if len(triggered) != 2 {
			t.Errorf("expected 2 triggered reminders (deduplicated), got %d", len(triggered))
			for i, task := range triggered {
				t.Logf("  triggered[%d]: %s (ID: %s)", i, task.Summary, task.ID)
			}
		}

		// Notifications should also be deduplicated
		if len(sentNotifications) != 2 {
			t.Errorf("expected 2 notifications (deduplicated), got %d", len(sentNotifications))
		}
	})

	t.Run("GetUpcomingReminders deduplicates by task ID", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		service, err := reminder.NewService(&reminder.Config{
			Enabled: true,
			Intervals: []string{
				"7 days",
				"1 day",
			},
			OSNotification:  true,
			LogNotification: true,
		}, dbPath)
		if err != nil {
			t.Fatalf("failed to create service: %v", err)
		}
		defer func() { _ = service.Close() }()

		dueTime := time.Now().AddDate(0, 0, 1)
		taskA := &backend.Task{
			ID:      "task-dup-3",
			Summary: "Task A",
			DueDate: &dueTime,
			Status:  backend.StatusNeedsAction,
		}
		taskB := &backend.Task{
			ID:      "task-dup-4",
			Summary: "Task B",
			DueDate: &dueTime,
			Status:  backend.StatusNeedsAction,
		}

		// Simulate duplicate task entries
		tasks := []*backend.Task{taskA, taskB, taskA, taskB}

		upcoming, err := service.GetUpcomingReminders(tasks)
		if err != nil {
			t.Fatalf("GetUpcomingReminders failed: %v", err)
		}

		// Each task should appear only once despite being passed twice
		if len(upcoming) != 2 {
			t.Errorf("expected 2 upcoming reminders (deduplicated), got %d", len(upcoming))
			for i, task := range upcoming {
				t.Logf("  upcoming[%d]: %s (ID: %s)", i, task.Summary, task.ID)
			}
		}
	})
}

// TestReminderServiceGetUpcoming tests listing upcoming reminders
func TestReminderServiceGetUpcoming(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	service, err := reminder.NewService(&reminder.Config{
		Enabled: true,
		Intervals: []string{
			"7 days",
			"1 day",
		},
		OSNotification:  true,
		LogNotification: true,
	}, dbPath)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	defer func() { _ = service.Close() }()

	// Create tasks with various due dates
	dueTime1 := time.Now().AddDate(0, 0, 1)  // 1 day (within 7-day window)
	dueTime2 := time.Now().AddDate(0, 0, 5)  // 5 days (within 7-day window)
	dueTime3 := time.Now().AddDate(0, 0, 10) // 10 days (beyond 7-day window)

	tasks := []*backend.Task{
		{ID: "task-1", Summary: "Soon", DueDate: &dueTime1, Status: backend.StatusNeedsAction},
		{ID: "task-2", Summary: "Later", DueDate: &dueTime2, Status: backend.StatusNeedsAction},
		{ID: "task-3", Summary: "Much later", DueDate: &dueTime3, Status: backend.StatusNeedsAction},
	}

	upcoming, err := service.GetUpcomingReminders(tasks)
	if err != nil {
		t.Fatalf("GetUpcomingReminders failed: %v", err)
	}

	// Tasks 1 and 2 should be in upcoming (within 7-day window), task 3 should not
	if len(upcoming) != 2 {
		t.Errorf("expected 2 upcoming reminders, got %d", len(upcoming))
	}
}

// =============================================================================
// Acceptance Criteria Tests (083-task-reminders)
// =============================================================================

// TestReminderCreation tests that creating a task with due date schedules reminder
func TestReminderCreation(t *testing.T) {
	cli := testutil.NewCLITestWithReminder(t)

	cli.SetReminderConfig(&reminder.Config{
		Enabled: true,
		Intervals: []string{
			"1 day",
		},
		OSNotification:  true,
		LogNotification: true,
	})

	// Create a task with a due date (within reminder window)
	dueDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	cli.MustExecute("-y", "Work", "add", "Scheduled task", "--due-date", dueDate)

	// Check that reminder is scheduled (will trigger on check)
	stdout := cli.MustExecute("-y", "reminder", "check")

	// Verify the task's reminder was triggered
	testutil.AssertContains(t, stdout, "Scheduled task")
	testutil.AssertContains(t, stdout, "Triggered")
	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify notification log contains the reminder
	notifLog := cli.GetNotificationLog()
	if !strings.Contains(notifLog, "REMINDER") {
		t.Errorf("expected notification log to contain REMINDER, got:\n%s", notifLog)
	}
	if !strings.Contains(notifLog, "Scheduled task") {
		t.Errorf("expected notification log to contain task name, got:\n%s", notifLog)
	}
}

// TestReminderSnooze tests that user can snooze/dismiss reminders
func TestReminderSnooze(t *testing.T) {
	t.Run("dismiss prevents reminder from triggering again", func(t *testing.T) {
		cli := testutil.NewCLITestWithReminder(t)

		cli.SetReminderConfig(&reminder.Config{
			Enabled: true,
			Intervals: []string{
				"1 day",
			},
			OSNotification:  true,
			LogNotification: true,
		})

		// Create a task due tomorrow
		dueDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
		cli.MustExecute("-y", "Work", "add", "Snooze test task", "--due-date", dueDate)

		// First check triggers the reminder
		cli.MustExecute("-y", "reminder", "check")
		notifLog1 := cli.GetNotificationLog()
		if !strings.Contains(notifLog1, "Snooze test task") {
			t.Errorf("expected first reminder to trigger, got:\n%s", notifLog1)
		}

		// Clear notification log
		cli.ClearNotificationLog()

		// Dismiss the reminder
		stdout := cli.MustExecute("-y", "reminder", "dismiss", "Snooze test task")
		testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

		// Second check should NOT trigger (dismissed/snoozed)
		cli.MustExecute("-y", "reminder", "check")
		notifLog2 := cli.GetNotificationLog()
		if strings.Contains(notifLog2, "Snooze test task") {
			t.Errorf("expected dismissed reminder NOT to trigger again, got:\n%s", notifLog2)
		}
	})

	t.Run("disable prevents all reminders for a task", func(t *testing.T) {
		cli := testutil.NewCLITestWithReminder(t)

		cli.SetReminderConfig(&reminder.Config{
			Enabled: true,
			Intervals: []string{
				"1 day",
			},
			OSNotification:  true,
			LogNotification: true,
		})

		// Create a task due tomorrow
		dueDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
		cli.MustExecute("-y", "Work", "add", "Disable test task", "--due-date", dueDate)

		// Disable reminders for this task before checking
		stdout := cli.MustExecute("-y", "reminder", "disable", "Disable test task")
		testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

		// Check reminders - should NOT trigger for disabled task
		cli.MustExecute("-y", "reminder", "check")
		notifLog := cli.GetNotificationLog()
		if strings.Contains(notifLog, "Disable test task") {
			t.Errorf("expected disabled task reminder NOT to appear in log, got:\n%s", notifLog)
		}
	})
}

// TestReminderPersistence tests that reminders survive application restart
func TestReminderPersistence(t *testing.T) {
	t.Run("dismissed reminders persist across service restart", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "reminder-persist.db")

		cfg := &reminder.Config{
			Enabled: true,
			Intervals: []string{
				"1 day",
			},
			OSNotification:  true,
			LogNotification: true,
		}

		// Create task due tomorrow
		dueTime := time.Now().AddDate(0, 0, 1)
		task := &backend.Task{
			ID:      "persist-test-1",
			Summary: "Persistent task",
			DueDate: &dueTime,
			Status:  backend.StatusNeedsAction,
		}

		// First service instance - dismiss the reminder
		{
			service1, err := reminder.NewService(cfg, dbPath)
			if err != nil {
				t.Fatalf("failed to create first service: %v", err)
			}

			// Dismiss the reminder
			err = service1.DismissReminder(task.ID, "1 day")
			if err != nil {
				t.Fatalf("DismissReminder failed: %v", err)
			}

			// Close first service (simulates app shutdown)
			_ = service1.Close()
		}

		// Second service instance - verify dismissed state persisted
		{
			service2, err := reminder.NewService(cfg, dbPath)
			if err != nil {
				t.Fatalf("failed to create second service: %v", err)
			}
			defer func() { _ = service2.Close() }()

			var sentNotifications []notification.Notification
			mockNotifier := &mockNotificationManager{
				sendFunc: func(n notification.Notification) error {
					sentNotifications = append(sentNotifications, n)
					return nil
				},
			}
			service2.SetNotifier(mockNotifier)

			// Check reminders - should NOT trigger (dismissed in previous session)
			triggered, err := service2.CheckReminders([]*backend.Task{task})
			if err != nil {
				t.Fatalf("CheckReminders failed: %v", err)
			}

			if len(triggered) != 0 {
				t.Errorf("expected 0 triggered reminders (persisted dismiss), got %d", len(triggered))
			}
			if len(sentNotifications) != 0 {
				t.Errorf("expected 0 notifications (persisted dismiss), got %d", len(sentNotifications))
			}
		}
	})

	t.Run("disabled task reminders persist across service restart", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "reminder-persist-disabled.db")

		cfg := &reminder.Config{
			Enabled: true,
			Intervals: []string{
				"1 day",
			},
			OSNotification:  true,
			LogNotification: true,
		}

		// Create task due tomorrow
		dueTime := time.Now().AddDate(0, 0, 1)
		task := &backend.Task{
			ID:      "persist-test-2",
			Summary: "Disabled persistent task",
			DueDate: &dueTime,
			Status:  backend.StatusNeedsAction,
		}

		// First service instance - disable reminders for task
		{
			service1, err := reminder.NewService(cfg, dbPath)
			if err != nil {
				t.Fatalf("failed to create first service: %v", err)
			}

			// Disable reminders for this task
			err = service1.DisableReminder(task.ID)
			if err != nil {
				t.Fatalf("DisableReminder failed: %v", err)
			}

			// Close first service (simulates app shutdown)
			_ = service1.Close()
		}

		// Second service instance - verify disabled state persisted
		{
			service2, err := reminder.NewService(cfg, dbPath)
			if err != nil {
				t.Fatalf("failed to create second service: %v", err)
			}
			defer func() { _ = service2.Close() }()

			var sentNotifications []notification.Notification
			mockNotifier := &mockNotificationManager{
				sendFunc: func(n notification.Notification) error {
					sentNotifications = append(sentNotifications, n)
					return nil
				},
			}
			service2.SetNotifier(mockNotifier)

			// Check reminders - should NOT trigger (disabled in previous session)
			triggered, err := service2.CheckReminders([]*backend.Task{task})
			if err != nil {
				t.Fatalf("CheckReminders failed: %v", err)
			}

			if len(triggered) != 0 {
				t.Errorf("expected 0 triggered reminders (persisted disable), got %d", len(triggered))
			}
			if len(sentNotifications) != 0 {
				t.Errorf("expected 0 notifications (persisted disable), got %d", len(sentNotifications))
			}
		}
	})
}

// =============================================================================
// Mock types for testing
// =============================================================================

type mockNotificationManager struct {
	sendFunc func(n notification.Notification) error
}

func (m *mockNotificationManager) Send(n notification.Notification) error {
	if m.sendFunc != nil {
		return m.sendFunc(n)
	}
	return nil
}

func (m *mockNotificationManager) SendAsync(n notification.Notification) {
	_ = m.Send(n)
}

func (m *mockNotificationManager) Close() error {
	return nil
}

func (m *mockNotificationManager) ChannelCount() int {
	return 1
}

// =============================================================================
// Issue #34: Reminder configuration not loaded from config.yaml
// =============================================================================

// TestReminderConfigFromYAML tests that reminder settings are loaded from config.yaml
// This test reproduces issue #34 where custom intervals in config.yaml were ignored.
func TestReminderConfigFromYAML(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set up a config.yaml with custom reminder settings
	// This matches the user's config format from the issue
	configYAML := `backends:
  sqlite:
    type: sqlite
    enabled: true

default_backend: sqlite

reminder:
  enabled: true
  intervals:
    - "1 day"
    - "2 hours"
    - "at due time"
  os_notification: true
  log_notification: true
`
	cli.SetFullConfig(configYAML)

	// Check reminder status - should show custom intervals from config.yaml
	stdout := cli.MustExecute("-y", "reminder", "status")

	// Verify custom interval "2 hours" is loaded (not just defaults)
	testutil.AssertContains(t, stdout, "2 hours")
	// Also verify other intervals are present
	testutil.AssertContains(t, stdout, "1 day")
	testutil.AssertContains(t, stdout, "at due time")
	testutil.AssertResultCode(t, stdout, testutil.ResultInfoOnly)
}
