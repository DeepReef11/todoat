// Package tui provides a terminal user interface for task management.
package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"todoat/backend"
)

// Backend interface for task operations (subset of backend.TaskManager)
type Backend interface {
	GetLists(ctx context.Context) ([]backend.List, error)
	GetTasks(ctx context.Context, listID string) ([]backend.Task, error)
	GetTask(ctx context.Context, listID, taskID string) (*backend.Task, error)
	CreateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error)
	UpdateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error)
	DeleteTask(ctx context.Context, listID, taskID string) error
}

// Focus indicates which pane has focus
type Focus int

const (
	FocusLists Focus = iota
	FocusTasks
)

// Mode indicates the current input mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeAdd
	ModeEdit
	ModeFilter
	ModeHelp
	ModeConfirmDelete
)

// Model represents the TUI state
type Model struct {
	backend Backend
	ctx     context.Context

	// Data
	lists       []backend.List
	tasks       []backend.Task
	filteredIdx []int // indices into tasks slice for filtered view

	// Selection
	listCursor int
	taskCursor int
	focus      Focus

	// Mode and input
	mode      Mode
	textInput textinput.Model
	filter    string

	// UI dimensions
	width  int
	height int

	// Styles
	listPaneStyle  lipgloss.Style
	taskPaneStyle  lipgloss.Style
	selectedStyle  lipgloss.Style
	completedStyle lipgloss.Style
	subtaskStyle   lipgloss.Style
	helpStyle      lipgloss.Style
	dialogStyle    lipgloss.Style
	statusBarStyle lipgloss.Style
}

// Message types
type listsLoadedMsg struct {
	lists []backend.List
}

type tasksLoadedMsg struct {
	tasks []backend.Task
}

type taskCreatedMsg struct {
	task *backend.Task
}

type taskUpdatedMsg struct {
	task *backend.Task
}

type taskDeletedMsg struct {
	taskID string
}

type errMsg struct {
	err error
}

// New creates a new TUI model
func New(b Backend) *Model {
	ti := textinput.New()
	ti.Placeholder = "Enter text..."
	ti.CharLimit = 256

	return &Model{
		backend:   b,
		ctx:       context.Background(),
		textInput: ti,
		focus:     FocusLists,
		mode:      ModeNormal,
		listPaneStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		taskPaneStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		selectedStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")),
		completedStyle: lipgloss.NewStyle().
			Strikethrough(true).
			Foreground(lipgloss.Color("240")),
		subtaskStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		helpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		dialogStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2),
		statusBarStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),
	}
}

// Init initializes the TUI
func (m *Model) Init() tea.Cmd {
	return m.loadLists()
}

func (m *Model) loadLists() tea.Cmd {
	return func() tea.Msg {
		lists, err := m.backend.GetLists(m.ctx)
		if err != nil {
			return errMsg{err}
		}
		return listsLoadedMsg{lists}
	}
}

func (m *Model) loadTasks() tea.Cmd {
	if len(m.lists) == 0 || m.listCursor >= len(m.lists) {
		return nil
	}
	listID := m.lists[m.listCursor].ID
	return func() tea.Msg {
		tasks, err := m.backend.GetTasks(m.ctx, listID)
		if err != nil {
			return errMsg{err}
		}
		return tasksLoadedMsg{tasks}
	}
}

func (m *Model) createTask(summary string) tea.Cmd {
	if len(m.lists) == 0 || m.listCursor >= len(m.lists) {
		return nil
	}
	listID := m.lists[m.listCursor].ID
	return func() tea.Msg {
		task := &backend.Task{
			Summary: summary,
			Status:  backend.StatusNeedsAction,
			ListID:  listID,
		}
		created, err := m.backend.CreateTask(m.ctx, listID, task)
		if err != nil {
			return errMsg{err}
		}
		return taskCreatedMsg{created}
	}
}

func (m *Model) updateTask(task *backend.Task) tea.Cmd {
	if len(m.lists) == 0 || m.listCursor >= len(m.lists) {
		return nil
	}
	listID := m.lists[m.listCursor].ID
	return func() tea.Msg {
		updated, err := m.backend.UpdateTask(m.ctx, listID, task)
		if err != nil {
			return errMsg{err}
		}
		return taskUpdatedMsg{updated}
	}
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case listsLoadedMsg:
		m.lists = msg.lists
		if len(m.lists) > 0 {
			return m, m.loadTasks()
		}
		return m, nil

	case tasksLoadedMsg:
		m.tasks = msg.tasks
		m.applyFilter()
		return m, nil

	case taskCreatedMsg:
		m.tasks = append(m.tasks, *msg.task)
		m.applyFilter()
		m.taskCursor = len(m.filteredIdx) - 1
		return m, nil

	case taskUpdatedMsg:
		for i, t := range m.tasks {
			if t.ID == msg.task.ID {
				m.tasks[i] = *msg.task
				break
			}
		}
		m.applyFilter()
		return m, nil

	case taskDeletedMsg:
		for i, t := range m.tasks {
			if t.ID == msg.taskID {
				m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
				break
			}
		}
		m.applyFilter()
		if m.taskCursor >= len(m.filteredIdx) && m.taskCursor > 0 {
			m.taskCursor--
		}
		return m, nil

	case errMsg:
		// For now just ignore errors
		return m, nil

	case tea.KeyMsg:
		// Handle mode-specific input
		switch m.mode {
		case ModeAdd:
			return m.handleAddMode(msg)
		case ModeEdit:
			return m.handleEditMode(msg)
		case ModeFilter:
			return m.handleFilterMode(msg)
		case ModeHelp:
			return m.handleHelpMode(msg)
		case ModeConfirmDelete:
			return m.handleConfirmDeleteMode(msg)
		}

		// Normal mode key handling
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			if m.focus == FocusLists {
				m.focus = FocusTasks
			} else {
				m.focus = FocusLists
			}
			return m, nil

		case "up", "k":
			if m.focus == FocusLists {
				if m.listCursor > 0 {
					m.listCursor--
					return m, m.loadTasks()
				}
			} else {
				if m.taskCursor > 0 {
					m.taskCursor--
				}
			}
			return m, nil

		case "down", "j":
			if m.focus == FocusLists {
				if m.listCursor < len(m.lists)-1 {
					m.listCursor++
					return m, m.loadTasks()
				}
			} else {
				if m.taskCursor < len(m.filteredIdx)-1 {
					m.taskCursor++
				}
			}
			return m, nil

		case "a":
			m.mode = ModeAdd
			m.textInput.Reset()
			m.textInput.Placeholder = "New task name..."
			m.textInput.Focus()
			return m, textinput.Blink

		case "e":
			if len(m.filteredIdx) > 0 && m.taskCursor < len(m.filteredIdx) {
				taskIdx := m.filteredIdx[m.taskCursor]
				m.mode = ModeEdit
				m.textInput.Reset()
				m.textInput.SetValue(m.tasks[taskIdx].Summary)
				m.textInput.Focus()
				return m, textinput.Blink
			}
			return m, nil

		case "c":
			if len(m.filteredIdx) > 0 && m.taskCursor < len(m.filteredIdx) {
				taskIdx := m.filteredIdx[m.taskCursor]
				task := m.tasks[taskIdx]
				if task.Status == backend.StatusCompleted {
					task.Status = backend.StatusNeedsAction
				} else {
					task.Status = backend.StatusCompleted
				}
				return m, m.updateTask(&task)
			}
			return m, nil

		case "d":
			if len(m.filteredIdx) > 0 && m.taskCursor < len(m.filteredIdx) {
				m.mode = ModeConfirmDelete
				return m, nil
			}
			return m, nil

		case "/":
			m.mode = ModeFilter
			m.textInput.Reset()
			m.textInput.Placeholder = "Search..."
			m.textInput.Focus()
			return m, textinput.Blink

		case "?":
			m.mode = ModeHelp
			return m, nil
		}
	}

	// Update text input for modes that use it
	if m.mode == ModeAdd || m.mode == ModeEdit || m.mode == ModeFilter {
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleAddMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.Type {
	case tea.KeyEnter:
		value := m.textInput.Value()
		if value != "" {
			m.mode = ModeNormal
			return m, m.createTask(value)
		}
		m.mode = ModeNormal
		return m, nil

	case tea.KeyEsc:
		m.mode = ModeNormal
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *Model) handleEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.Type {
	case tea.KeyEnter:
		value := m.textInput.Value()
		if value != "" && len(m.filteredIdx) > 0 && m.taskCursor < len(m.filteredIdx) {
			taskIdx := m.filteredIdx[m.taskCursor]
			task := m.tasks[taskIdx]
			task.Summary = value
			m.mode = ModeNormal
			return m, m.updateTask(&task)
		}
		m.mode = ModeNormal
		return m, nil

	case tea.KeyEsc:
		m.mode = ModeNormal
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *Model) handleFilterMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.Type {
	case tea.KeyEnter:
		m.filter = m.textInput.Value()
		m.applyFilter()
		m.mode = ModeNormal
		return m, nil

	case tea.KeyEsc:
		m.filter = ""
		m.applyFilter()
		m.mode = ModeNormal
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *Model) handleHelpMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyEnter:
		m.mode = ModeNormal
		return m, nil
	}

	if msg.String() == "q" {
		m.mode = ModeNormal
		return m, nil
	}

	return m, nil
}

func (m *Model) handleConfirmDeleteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if len(m.filteredIdx) > 0 && m.taskCursor < len(m.filteredIdx) {
			taskIdx := m.filteredIdx[m.taskCursor]
			taskID := m.tasks[taskIdx].ID
			listID := ""
			if len(m.lists) > 0 && m.listCursor < len(m.lists) {
				listID = m.lists[m.listCursor].ID
			}

			// Perform delete synchronously for immediate feedback
			if listID != "" {
				_ = m.backend.DeleteTask(m.ctx, listID, taskID)
			}

			// Remove from local state
			for i, t := range m.tasks {
				if t.ID == taskID {
					m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
					break
				}
			}
			m.applyFilter()
			if m.taskCursor >= len(m.filteredIdx) && m.taskCursor > 0 {
				m.taskCursor--
			}
			m.mode = ModeNormal
			return m, nil
		}
		m.mode = ModeNormal
		return m, nil

	case "n", "N", "esc":
		m.mode = ModeNormal
		return m, nil
	}

	if msg.Type == tea.KeyEsc {
		m.mode = ModeNormal
		return m, nil
	}

	return m, nil
}

func (m *Model) applyFilter() {
	m.filteredIdx = nil
	for i, task := range m.tasks {
		if m.filter == "" || strings.Contains(strings.ToLower(task.Summary), strings.ToLower(m.filter)) {
			m.filteredIdx = append(m.filteredIdx, i)
		}
	}
	if m.taskCursor >= len(m.filteredIdx) {
		m.taskCursor = 0
	}
}

// View renders the TUI
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		m.width = 80
		m.height = 24
	}

	var b strings.Builder

	// Calculate pane widths
	listWidth := m.width / 4
	taskWidth := m.width - listWidth - 4

	// Render list pane
	listContent := m.renderListPane(listWidth - 4)
	listPane := m.listPaneStyle.Width(listWidth).Height(m.height - 4).Render(listContent)

	// Render task pane
	taskContent := m.renderTaskPane(taskWidth - 4)
	taskPane := m.taskPaneStyle.Width(taskWidth).Height(m.height - 4).Render(taskContent)

	// Join panes horizontally
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, listPane, taskPane)

	// Status bar
	statusBar := m.renderStatusBar()

	b.WriteString(mainView)
	b.WriteString("\n")
	b.WriteString(statusBar)

	// Overlay dialogs
	switch m.mode {
	case ModeAdd:
		return m.renderAddDialog()
	case ModeEdit:
		return m.renderEditDialog()
	case ModeFilter:
		return m.renderFilterDialog()
	case ModeHelp:
		return m.renderHelpDialog()
	case ModeConfirmDelete:
		return m.renderConfirmDeleteDialog()
	}

	return b.String()
}

func (m *Model) renderListPane(width int) string {
	var b strings.Builder
	b.WriteString("Lists\n")
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	for i, list := range m.lists {
		cursor := " "
		if i == m.listCursor && m.focus == FocusLists {
			cursor = ">"
		}
		name := list.Name
		if i == m.listCursor && m.focus == FocusLists {
			name = m.selectedStyle.Render(name)
		}
		b.WriteString(cursor + " " + name + "\n")
	}

	return b.String()
}

func (m *Model) renderTaskPane(width int) string {
	var b strings.Builder
	b.WriteString("Tasks\n")
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	if len(m.filteredIdx) == 0 {
		b.WriteString("No tasks\n")
		return b.String()
	}

	// Build parent-child relationships for tree view
	taskByID := make(map[string]int)
	for i, task := range m.tasks {
		taskByID[task.ID] = i
	}

	// Track which tasks have been rendered (for tree view)
	rendered := make(map[string]bool)

	for fi, taskIdx := range m.filteredIdx {
		task := m.tasks[taskIdx]

		// Skip if already rendered as part of a parent
		if rendered[task.ID] {
			continue
		}

		// Render task with proper indentation
		m.renderTask(&b, task, fi, 0, taskByID, rendered)
	}

	return b.String()
}

func (m *Model) renderTask(b *strings.Builder, task backend.Task, filterIdx, indent int, taskByID map[string]int, rendered map[string]bool) {
	rendered[task.ID] = true

	cursor := " "
	if filterIdx == m.taskCursor && m.focus == FocusTasks {
		cursor = ">"
	}

	// Indentation for subtasks
	indentStr := strings.Repeat("  ", indent)
	if indent > 0 {
		indentStr = strings.Repeat("  ", indent-1) + "└─"
	}

	// Status indicator
	var status string
	switch task.Status {
	case backend.StatusCompleted:
		status = "[✓]"
	case backend.StatusInProgress:
		status = "[~]"
	default:
		status = "[ ]"
	}

	// Summary
	summary := task.Summary
	if task.Status == backend.StatusCompleted {
		summary = m.completedStyle.Render(summary)
	} else if filterIdx == m.taskCursor && m.focus == FocusTasks {
		summary = m.selectedStyle.Render(summary)
	}
	if indent > 0 {
		summary = m.subtaskStyle.Render(summary)
	}

	b.WriteString(cursor + " " + indentStr + status + " " + summary + "\n")

	// Render children (subtasks)
	for i, t := range m.tasks {
		if t.ParentID == task.ID && !rendered[t.ID] {
			// Find filter index for this subtask
			childFilterIdx := -1
			for fi, idx := range m.filteredIdx {
				if idx == i {
					childFilterIdx = fi
					break
				}
			}
			if childFilterIdx >= 0 {
				m.renderTask(b, t, childFilterIdx, indent+1, taskByID, rendered)
			}
		}
	}
}

func (m *Model) renderStatusBar() string {
	left := ""
	if len(m.lists) > 0 && m.listCursor < len(m.lists) {
		left = m.lists[m.listCursor].Name
	}

	right := "q:quit  ?:help"
	if m.filter != "" {
		right = "Filter: " + m.filter + "  " + right
	}

	padding := m.width - len(left) - len(right) - 2
	if padding < 1 {
		padding = 1
	}

	return m.statusBarStyle.Width(m.width).Render(left + strings.Repeat(" ", padding) + right)
}

func (m *Model) renderAddDialog() string {
	dialog := m.dialogStyle.Render(
		"Add New Task\n\n" +
			m.textInput.View() + "\n\n" +
			m.helpStyle.Render("Enter: confirm  Esc: cancel"),
	)
	return m.centerDialog(dialog)
}

func (m *Model) renderEditDialog() string {
	title := "Edit Task"
	if len(m.filteredIdx) > 0 && m.taskCursor < len(m.filteredIdx) {
		taskIdx := m.filteredIdx[m.taskCursor]
		title = "Edit: " + m.tasks[taskIdx].Summary
	}

	dialog := m.dialogStyle.Render(
		title + "\n\n" +
			m.textInput.View() + "\n\n" +
			m.helpStyle.Render("Enter: confirm  Esc: cancel"),
	)
	return m.centerDialog(dialog)
}

func (m *Model) renderFilterDialog() string {
	dialog := m.dialogStyle.Render(
		"Search/Filter Tasks\n\n" +
			m.textInput.View() + "\n\n" +
			m.helpStyle.Render("Enter: filter  Esc: clear"),
	)
	return m.centerDialog(dialog)
}

func (m *Model) renderHelpDialog() string {
	help := `Help - Key Bindings

Navigation:
  j/↓    Move down
  k/↑    Move up
  Tab    Switch focus between lists/tasks

Actions:
  a      Add new task
  e      Edit selected task
  c      Toggle task completion
  d      Delete task (with confirm)
  /      Search/filter tasks

General:
  ?      Show this help
  q      Quit

Press any key to close`

	dialog := m.dialogStyle.Render(help)
	return m.centerDialog(dialog)
}

func (m *Model) renderConfirmDeleteDialog() string {
	dialog := m.dialogStyle.Render(
		"Delete selected task?\n\n" +
			m.helpStyle.Render("y: yes  n: no"),
	)
	return m.centerDialog(dialog)
}

func (m *Model) centerDialog(dialog string) string {
	// Get dialog dimensions
	lines := strings.Split(dialog, "\n")
	dialogHeight := len(lines)
	dialogWidth := 0
	for _, line := range lines {
		if len(line) > dialogWidth {
			dialogWidth = len(line)
		}
	}

	// Calculate position
	topPad := (m.height - dialogHeight) / 2
	leftPad := (m.width - dialogWidth) / 2

	if topPad < 0 {
		topPad = 0
	}
	if leftPad < 0 {
		leftPad = 0
	}

	// Build centered output
	var b strings.Builder
	for i := 0; i < topPad; i++ {
		b.WriteString("\n")
	}
	for _, line := range lines {
		b.WriteString(strings.Repeat(" ", leftPad))
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}
