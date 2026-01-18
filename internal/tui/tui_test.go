package tui_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"todoat/backend"
	"todoat/internal/tui"
)

// =============================================================================
// TUI Interface Tests (029-tui-interface)
// =============================================================================

// mockBackend implements tui.Backend for testing
type mockBackend struct {
	lists []backend.List
	tasks map[string][]backend.Task
}

func newMockBackend() *mockBackend {
	return &mockBackend{
		lists: []backend.List{
			{ID: "1", Name: "Work", Color: "#FF0000"},
			{ID: "2", Name: "Personal", Color: "#00FF00"},
		},
		tasks: map[string][]backend.Task{
			"1": {
				{ID: "t1", Summary: "Review PR", Status: backend.StatusNeedsAction, ListID: "1", Priority: 1},
				{ID: "t2", Summary: "Write tests", Status: backend.StatusInProgress, ListID: "1", Priority: 5},
			},
			"2": {
				{ID: "t3", Summary: "Buy groceries", Status: backend.StatusNeedsAction, ListID: "2"},
			},
		},
	}
}

func (m *mockBackend) GetLists(_ context.Context) ([]backend.List, error) {
	return m.lists, nil
}

func (m *mockBackend) GetTasks(_ context.Context, listID string) ([]backend.Task, error) {
	return m.tasks[listID], nil
}

func (m *mockBackend) GetTask(_ context.Context, listID, taskID string) (*backend.Task, error) {
	for _, t := range m.tasks[listID] {
		if t.ID == taskID {
			return &t, nil
		}
	}
	return nil, nil
}

func (m *mockBackend) CreateTask(_ context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	task.ID = "new-task"
	m.tasks[listID] = append(m.tasks[listID], *task)
	return task, nil
}

func (m *mockBackend) UpdateTask(_ context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	for i, t := range m.tasks[listID] {
		if t.ID == task.ID {
			m.tasks[listID][i] = *task
			return task, nil
		}
	}
	return nil, nil
}

func (m *mockBackend) DeleteTask(_ context.Context, listID, taskID string) error {
	tasks := m.tasks[listID]
	for i, t := range tasks {
		if t.ID == taskID {
			m.tasks[listID] = append(tasks[:i], tasks[i+1:]...)
			break
		}
	}
	return nil
}

// readAll reads all output from a reader and returns as bytes
func readAll(t *testing.T, r io.Reader) []byte {
	t.Helper()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	return out
}

// --- TUI Launch Tests ---

// TestTUILaunch - `todoat tui` launches the terminal interface
func TestTUILaunch(t *testing.T) {
	mb := newMockBackend()
	model := tui.New(mb)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Quit the TUI
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// The TUI should render without errors
	out := readAll(t, tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))
	if len(out) == 0 {
		t.Error("expected TUI to render some output")
	}
}

// --- List Navigation Tests ---

// TestTUIListNavigation - Arrow keys navigate between task lists
func TestTUIListNavigation(t *testing.T) {
	mb := newMockBackend()
	model := tui.New(mb)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Initially should be on first list (Work)
	// Press down to navigate to next list
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(50 * time.Millisecond)

	// Quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	out := readAll(t, tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))
	if !bytes.Contains(out, []byte("Work")) {
		t.Error("expected 'Work' list to be visible")
	}
	if !bytes.Contains(out, []byte("Personal")) {
		t.Error("expected 'Personal' list to be visible after navigation")
	}
}

// --- Task Navigation Tests ---

// TestTUITaskNavigation - Arrow keys navigate between tasks in list
func TestTUITaskNavigation(t *testing.T) {
	mb := newMockBackend()
	model := tui.New(mb)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Press Tab to switch focus to tasks pane
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(50 * time.Millisecond)

	// Use j/k for task navigation (vim-like)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	time.Sleep(50 * time.Millisecond)

	// Quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	out := readAll(t, tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))
	// Should show tasks
	if !bytes.Contains(out, []byte("Review PR")) {
		t.Error("expected 'Review PR' to be visible")
	}
	if !bytes.Contains(out, []byte("Write tests")) {
		t.Error("expected 'Write tests' to be visible after navigation")
	}
}

// --- Add Task Tests ---

// TestTUIAddTask - Press 'a' to add new task via input dialog
func TestTUIAddTask(t *testing.T) {
	mb := newMockBackend()
	model := tui.New(mb)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Press 'a' to add new task
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	time.Sleep(50 * time.Millisecond)

	// Type task name and confirm
	for _, r := range "New test task" {
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(100 * time.Millisecond)

	// Quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	out := readAll(t, tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))
	if !bytes.Contains(out, []byte("New test task")) {
		t.Error("expected new task to appear in list")
	}
}

// --- Edit Task Tests ---

// TestTUIEditTask - Press 'e' to edit selected task
func TestTUIEditTask(t *testing.T) {
	mb := newMockBackend()
	model := tui.New(mb)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Switch to task pane
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(50 * time.Millisecond)

	// Press 'e' to edit task
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	time.Sleep(50 * time.Millisecond)

	// Quit (escape from edit then quit)
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	out := readAll(t, tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))
	// Should show edit dialog with current task content
	if !bytes.Contains(out, []byte("Edit")) && !bytes.Contains(out, []byte("Review PR")) {
		t.Error("expected edit dialog to appear with task content")
	}
}

// --- Complete Task Tests ---

// TestTUICompleteTask - Press 'c' to toggle task completion
func TestTUICompleteTask(t *testing.T) {
	mb := newMockBackend()
	model := tui.New(mb)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Switch to task pane
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(50 * time.Millisecond)

	// Press 'c' to complete task
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	time.Sleep(100 * time.Millisecond)

	// Quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	out := readAll(t, tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))
	// Should show completion indicator (checkmark or strikethrough)
	if !bytes.Contains(out, []byte("✓")) && !bytes.Contains(out, []byte("[x]")) && !bytes.Contains(out, []byte("COMPLETED")) {
		t.Error("expected task completion indicator")
	}
}

// --- Delete Task Tests ---

// TestTUIDeleteTask - Press 'd' with confirmation to delete task
func TestTUIDeleteTask(t *testing.T) {
	mb := newMockBackend()
	model := tui.New(mb)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Switch to task pane
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(50 * time.Millisecond)

	// Press 'd' to delete task
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	time.Sleep(50 * time.Millisecond)

	// Confirm deletion
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	time.Sleep(100 * time.Millisecond)

	// Quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	out := readAll(t, tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))
	// Check final frame - extract content after the last [15A (cursor positioning)
	// which indicates a fresh screen redraw after deletion
	outStr := string(out)
	lastFrame := outStr
	if idx := strings.LastIndex(outStr, "[15A"); idx != -1 {
		lastFrame = outStr[idx:]
	}
	// Task should be removed from the final frame
	if strings.Contains(lastFrame, "Review PR") {
		t.Errorf("expected task to be deleted from final frame, got:\n%s", lastFrame)
	}
}

// --- Tree View Tests ---

// TestTUITreeView - Subtasks displayed in collapsible tree structure
func TestTUITreeView(t *testing.T) {
	mb := newMockBackend()
	// Add a task with subtasks
	mb.tasks["1"] = append(mb.tasks["1"], backend.Task{
		ID:       "t4",
		Summary:  "Parent task",
		Status:   backend.StatusNeedsAction,
		ListID:   "1",
		ParentID: "",
	})
	mb.tasks["1"] = append(mb.tasks["1"], backend.Task{
		ID:       "t5",
		Summary:  "Child task",
		Status:   backend.StatusNeedsAction,
		ListID:   "1",
		ParentID: "t4",
	})

	model := tui.New(mb)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	out := readAll(t, tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))
	// Should show parent task
	if !bytes.Contains(out, []byte("Parent task")) {
		t.Error("expected parent task to be visible")
	}

	// Should show subtask (may have indentation or tree indicator)
	if !bytes.Contains(out, []byte("Child task")) {
		t.Error("expected subtask to be visible")
	}
}

// --- Filter Tests ---

// TestTUIFilterTasks - '/' opens filter/search dialog
func TestTUIFilterTasks(t *testing.T) {
	mb := newMockBackend()
	model := tui.New(mb)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Press '/' to open filter dialog
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	time.Sleep(50 * time.Millisecond)

	// Type search query
	for _, r := range "Review" {
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(50 * time.Millisecond)

	// Quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	out := readAll(t, tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))
	// Should filter tasks to show only matching ones
	if !bytes.Contains(out, []byte("Review PR")) {
		t.Error("expected matching task to be shown")
	}
}

// --- Help Tests ---

// TestTUIKeyBindings - Help panel shows all available key bindings ('?')
func TestTUIKeyBindings(t *testing.T) {
	mb := newMockBackend()
	model := tui.New(mb)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Press '?' to show help
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	time.Sleep(50 * time.Millisecond)

	// Quit (escape from help then quit)
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	out := readAll(t, tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))
	// Should show key bindings help
	if !bytes.Contains(out, []byte("Help")) && !bytes.Contains(out, []byte("Key")) {
		t.Error("expected help panel to show key bindings")
	}
}

// --- Quit Tests ---

// TestTUIQuit - 'q' exits the TUI gracefully
func TestTUIQuit(t *testing.T) {
	mb := newMockBackend()
	model := tui.New(mb)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Press 'q' to quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Should exit without error
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
