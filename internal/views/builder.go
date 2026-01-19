package views

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// BuilderPanel indicates which panel has focus in the view builder
type BuilderPanel int

const (
	PanelFields BuilderPanel = iota
	PanelFilters
	PanelSort
)

// Builder is the TUI model for interactive view creation
type Builder struct {
	viewName  string
	viewsDir  string
	cancelled bool
	saved     bool

	// Current panel focus
	panel BuilderPanel

	// Fields panel state
	fieldCursor    int
	selectedFields map[string]bool
	fieldConfigs   map[string]FieldConfig

	// Field configuration dialog state
	inFieldConfig    bool
	configFieldName  string
	configCursor     int
	configTextInput  textinput.Model
	configInputField int // 0=width, 1=format

	// Filters panel state
	filters      []Filter
	filterCursor int
	addingFilter bool
	filterField  int // index in AvailableFields
	filterOp     int // index in filterOperators
	filterValue  textinput.Model
	filterStep   int // 0=field, 1=operator, 2=value

	// Sort panel state
	sortRules     []SortRule
	sortCursor    int
	addingSort    bool
	sortField     int // index in AvailableFields
	sortDirection int // 0=asc, 1=desc
	sortStep      int // 0=field, 1=direction

	// Error message
	errorMsg string

	// UI dimensions
	width  int
	height int

	// Styles
	titleStyle       lipgloss.Style
	panelStyle       lipgloss.Style
	activePanelStyle lipgloss.Style
	selectedStyle    lipgloss.Style
	helpStyle        lipgloss.Style
	errorStyle       lipgloss.Style
	checkboxStyle    lipgloss.Style
}

// FieldConfig holds configuration for a field in the view
type FieldConfig struct {
	Width  int
	Align  string
	Format string
}

// Filter operators
var filterOperators = []string{
	"eq", "ne", "lt", "lte", "gt", "gte", "contains", "in", "not_in", "regex",
}

// Alignment options
var alignments = []string{"left", "center", "right"}

// NewBuilder creates a new Builder model for creating views
func NewBuilder(name, viewsDir string) *Builder {
	ti := textinput.New()
	ti.Placeholder = "value"
	ti.CharLimit = 256

	filterVal := textinput.New()
	filterVal.Placeholder = "filter value..."
	filterVal.CharLimit = 256

	configInput := textinput.New()
	configInput.Placeholder = ""
	configInput.CharLimit = 10

	return &Builder{
		viewName:        name,
		viewsDir:        viewsDir,
		selectedFields:  make(map[string]bool),
		fieldConfigs:    make(map[string]FieldConfig),
		filterValue:     filterVal,
		configTextInput: configInput,
		width:           80,
		height:          24,
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")),
		panelStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		activePanelStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("212")).
			Padding(0, 1),
		selectedStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")),
		helpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")),
		checkboxStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")),
	}
}

// Init initializes the builder
func (b *Builder) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (b *Builder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width
		b.height = msg.Height
		return b, nil

	case tea.KeyMsg:
		// Clear error on any key press
		b.errorMsg = ""

		// Handle field configuration dialog
		if b.inFieldConfig {
			return b.handleFieldConfigInput(msg)
		}

		// Handle filter adding mode
		if b.addingFilter {
			return b.handleFilterInput(msg)
		}

		// Handle sort adding mode
		if b.addingSort {
			return b.handleSortInput(msg)
		}

		// Global shortcuts
		switch msg.Type {
		case tea.KeyCtrlC:
			b.cancelled = true
			return b, tea.Quit

		case tea.KeyEsc:
			b.cancelled = true
			return b, tea.Quit

		case tea.KeyCtrlS:
			return b.save()

		case tea.KeyTab:
			b.panel = (b.panel + 1) % 3
			return b, nil

		case tea.KeyShiftTab:
			if b.panel == 0 {
				b.panel = 2
			} else {
				b.panel--
			}
			return b, nil
		}

		// Panel-specific handling
		switch b.panel {
		case PanelFields:
			return b.handleFieldsPanel(msg)
		case PanelFilters:
			return b.handleFiltersPanel(msg)
		case PanelSort:
			return b.handleSortPanel(msg)
		}
	}

	return b, nil
}

func (b *Builder) handleFieldsPanel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if b.fieldCursor > 0 {
			b.fieldCursor--
		}
	case tea.KeyDown:
		if b.fieldCursor < len(AvailableFields)-1 {
			b.fieldCursor++
		}
	case tea.KeySpace:
		fieldName := AvailableFields[b.fieldCursor]
		b.selectedFields[fieldName] = !b.selectedFields[fieldName]
	case tea.KeyEnter:
		// Open field configuration for selected field
		fieldName := AvailableFields[b.fieldCursor]
		if b.selectedFields[fieldName] {
			b.inFieldConfig = true
			b.configFieldName = fieldName
			b.configCursor = 0
			b.configInputField = -1
		}
	}
	return b, nil
}

func (b *Builder) handleFieldConfigInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If in text input mode
	if b.configInputField >= 0 {
		switch msg.Type {
		case tea.KeyEnter:
			// Save the value
			cfg := b.fieldConfigs[b.configFieldName]
			val := b.configTextInput.Value()
			switch b.configInputField {
			case 0: // Width
				var width int
				if val != "" {
					for _, c := range val {
						if c >= '0' && c <= '9' {
							width = width*10 + int(c-'0')
						}
					}
				}
				cfg.Width = width
			case 1: // Format
				cfg.Format = val
			}
			b.fieldConfigs[b.configFieldName] = cfg
			b.configInputField = -1
			b.configTextInput.Blur()
			return b, nil

		case tea.KeyEsc:
			b.configInputField = -1
			b.configTextInput.Blur()
			return b, nil
		}

		var cmd tea.Cmd
		b.configTextInput, cmd = b.configTextInput.Update(msg)
		return b, cmd
	}

	// Normal config navigation
	switch msg.Type {
	case tea.KeyEsc:
		b.inFieldConfig = false
		return b, nil

	case tea.KeyUp:
		if b.configCursor > 0 {
			b.configCursor--
		}
	case tea.KeyDown:
		if b.configCursor < 2 { // Width, Align, Format
			b.configCursor++
		}
	case tea.KeyEnter, tea.KeySpace:
		cfg := b.fieldConfigs[b.configFieldName]
		switch b.configCursor {
		case 0: // Width - open text input
			b.configInputField = 0
			b.configTextInput.SetValue("")
			if cfg.Width > 0 {
				b.configTextInput.SetValue(string(rune('0' + cfg.Width%10)))
			}
			b.configTextInput.Placeholder = "width (e.g., 20)"
			b.configTextInput.Focus()
			return b, textinput.Blink

		case 1: // Align - cycle through options
			idx := 0
			for i, a := range alignments {
				if a == cfg.Align {
					idx = i
					break
				}
			}
			idx = (idx + 1) % len(alignments)
			cfg.Align = alignments[idx]
			b.fieldConfigs[b.configFieldName] = cfg

		case 2: // Format - open text input
			b.configInputField = 1
			b.configTextInput.SetValue(cfg.Format)
			b.configTextInput.Placeholder = "format (e.g., 2006-01-02)"
			b.configTextInput.Focus()
			return b, textinput.Blink
		}
	}
	return b, nil
}

func (b *Builder) handleFiltersPanel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if b.filterCursor > 0 {
			b.filterCursor--
		}
	case tea.KeyDown:
		if b.filterCursor < len(b.filters)-1 {
			b.filterCursor++
		}
	case tea.KeyEnter:
		// Add new filter
		b.addingFilter = true
		b.filterStep = 0
		b.filterField = 0
		b.filterOp = 0
		b.filterValue.SetValue("")
	case tea.KeyBackspace, tea.KeyDelete:
		// Remove selected filter
		if len(b.filters) > 0 && b.filterCursor < len(b.filters) {
			b.filters = append(b.filters[:b.filterCursor], b.filters[b.filterCursor+1:]...)
			if b.filterCursor >= len(b.filters) && b.filterCursor > 0 {
				b.filterCursor--
			}
		}
	}
	return b, nil
}

func (b *Builder) handleFilterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch b.filterStep {
	case 0: // Select field
		switch msg.Type {
		case tea.KeyEsc:
			b.addingFilter = false
			return b, nil
		case tea.KeyUp:
			if b.filterField > 0 {
				b.filterField--
			}
		case tea.KeyDown:
			if b.filterField < len(AvailableFields)-1 {
				b.filterField++
			}
		case tea.KeyEnter:
			b.filterStep = 1
		}
	case 1: // Select operator
		switch msg.Type {
		case tea.KeyEsc:
			b.filterStep = 0
			return b, nil
		case tea.KeyUp:
			if b.filterOp > 0 {
				b.filterOp--
			}
		case tea.KeyDown:
			if b.filterOp < len(filterOperators)-1 {
				b.filterOp++
			}
		case tea.KeyEnter:
			b.filterStep = 2
			b.filterValue.Focus()
			return b, textinput.Blink
		}
	case 2: // Enter value
		switch msg.Type {
		case tea.KeyEsc:
			b.filterStep = 1
			b.filterValue.Blur()
			return b, nil
		case tea.KeyEnter:
			// Add the filter
			filter := Filter{
				Field:    AvailableFields[b.filterField],
				Operator: filterOperators[b.filterOp],
				Value:    b.filterValue.Value(),
			}
			b.filters = append(b.filters, filter)
			b.addingFilter = false
			b.filterValue.Blur()
			b.filterValue.SetValue("")
			return b, nil
		default:
			var cmd tea.Cmd
			b.filterValue, cmd = b.filterValue.Update(msg)
			return b, cmd
		}
	}
	return b, nil
}

func (b *Builder) handleSortPanel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if b.sortCursor > 0 {
			b.sortCursor--
		}
	case tea.KeyDown:
		if b.sortCursor < len(b.sortRules)-1 {
			b.sortCursor++
		}
	case tea.KeyEnter:
		// Add new sort rule
		b.addingSort = true
		b.sortStep = 0
		b.sortField = 0
		b.sortDirection = 0
	case tea.KeyBackspace, tea.KeyDelete:
		// Remove selected sort rule
		if len(b.sortRules) > 0 && b.sortCursor < len(b.sortRules) {
			b.sortRules = append(b.sortRules[:b.sortCursor], b.sortRules[b.sortCursor+1:]...)
			if b.sortCursor >= len(b.sortRules) && b.sortCursor > 0 {
				b.sortCursor--
			}
		}
	}
	return b, nil
}

func (b *Builder) handleSortInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch b.sortStep {
	case 0: // Select field
		switch msg.Type {
		case tea.KeyEsc:
			b.addingSort = false
			return b, nil
		case tea.KeyUp:
			if b.sortField > 0 {
				b.sortField--
			}
		case tea.KeyDown:
			if b.sortField < len(AvailableFields)-1 {
				b.sortField++
			}
		case tea.KeyEnter:
			b.sortStep = 1
		}
	case 1: // Select direction
		switch msg.Type {
		case tea.KeyEsc:
			b.sortStep = 0
			return b, nil
		case tea.KeyUp, tea.KeyDown:
			b.sortDirection = 1 - b.sortDirection
		case tea.KeyEnter:
			// Add the sort rule
			dir := "asc"
			if b.sortDirection == 1 {
				dir = "desc"
			}
			rule := SortRule{
				Field:     AvailableFields[b.sortField],
				Direction: dir,
			}
			b.sortRules = append(b.sortRules, rule)
			b.addingSort = false
			return b, nil
		}
	}
	return b, nil
}

func (b *Builder) save() (tea.Model, tea.Cmd) {
	// Validate: at least one field must be selected
	hasField := false
	for _, selected := range b.selectedFields {
		if selected {
			hasField = true
			break
		}
	}

	if !hasField {
		b.errorMsg = "Error: At least one field must be selected"
		return b, nil
	}

	// Build the view
	view := View{
		Name:    b.viewName,
		Filters: b.filters,
		Sort:    b.sortRules,
	}

	// Add fields in order
	for _, fieldName := range AvailableFields {
		if b.selectedFields[fieldName] {
			field := Field{Name: fieldName}
			if cfg, ok := b.fieldConfigs[fieldName]; ok {
				field.Width = cfg.Width
				field.Align = cfg.Align
				field.Format = cfg.Format
			}
			view.Fields = append(view.Fields, field)
		}
	}

	// Ensure views directory exists
	if err := os.MkdirAll(b.viewsDir, 0755); err != nil {
		b.errorMsg = "Error: " + err.Error()
		return b, nil
	}

	// Write YAML file
	viewPath := filepath.Join(b.viewsDir, b.viewName+".yaml")
	data, err := yaml.Marshal(view)
	if err != nil {
		b.errorMsg = "Error: " + err.Error()
		return b, nil
	}

	if err := os.WriteFile(viewPath, data, 0644); err != nil {
		b.errorMsg = "Error: " + err.Error()
		return b, nil
	}

	b.saved = true
	return b, tea.Quit
}

// View renders the builder UI
func (b *Builder) View() string {
	if b.width == 0 {
		b.width = 80
	}
	if b.height == 0 {
		b.height = 24
	}

	var out strings.Builder

	// Title
	title := b.titleStyle.Render("View Builder: " + b.viewName)
	out.WriteString(title + "\n\n")

	// Calculate panel widths
	panelWidth := (b.width - 6) / 3
	if panelWidth < 20 {
		panelWidth = 20
	}
	panelHeight := b.height - 8

	// Render panels
	fieldsPanel := b.renderFieldsPanel(panelWidth, panelHeight)
	filtersPanel := b.renderFiltersPanel(panelWidth, panelHeight)
	sortPanel := b.renderSortPanel(panelWidth, panelHeight)

	// Join panels
	panels := lipgloss.JoinHorizontal(lipgloss.Top, fieldsPanel, filtersPanel, sortPanel)
	out.WriteString(panels)
	out.WriteString("\n")

	// Error message
	if b.errorMsg != "" {
		out.WriteString(b.errorStyle.Render(b.errorMsg) + "\n")
	}

	// Help line
	help := "Arrow keys: navigate | Tab: next panel | Space: toggle | Enter: confirm/open | Ctrl+S: save | Esc/Ctrl+C: cancel"
	out.WriteString(b.helpStyle.Render(help))

	return out.String()
}

func (b *Builder) renderFieldsPanel(width, height int) string {
	var content strings.Builder
	content.WriteString("Fields\n")
	content.WriteString(strings.Repeat("─", width-4) + "\n")

	for i, field := range AvailableFields {
		cursor := " "
		if b.panel == PanelFields && i == b.fieldCursor {
			cursor = ">"
		}

		checkbox := "[ ]"
		if b.selectedFields[field] {
			checkbox = b.checkboxStyle.Render("[x]")
		}

		name := field
		if b.panel == PanelFields && i == b.fieldCursor {
			name = b.selectedStyle.Render(field)
		}

		content.WriteString(cursor + " " + checkbox + " " + name + "\n")
	}

	// Field configuration dialog overlay
	if b.inFieldConfig {
		content.WriteString("\n" + strings.Repeat("─", width-4) + "\n")
		content.WriteString("Config: " + b.configFieldName + "\n")

		cfg := b.fieldConfigs[b.configFieldName]
		items := []string{
			"Width: " + formatInt(cfg.Width),
			"Align: " + formatAlign(cfg.Align),
			"Format: " + cfg.Format,
		}

		for i, item := range items {
			cursor := " "
			if i == b.configCursor {
				cursor = ">"
				item = b.selectedStyle.Render(item)
			}
			content.WriteString(cursor + " " + item + "\n")
		}

		if b.configInputField >= 0 {
			content.WriteString(b.configTextInput.View() + "\n")
		}

		content.WriteString(b.helpStyle.Render("Enter: edit | Esc: done") + "\n")
	}

	style := b.panelStyle
	if b.panel == PanelFields {
		style = b.activePanelStyle
	}

	return style.Width(width).Height(height).Render(content.String())
}

func (b *Builder) renderFiltersPanel(width, height int) string {
	var content strings.Builder
	content.WriteString("Filters\n")
	content.WriteString(strings.Repeat("─", width-4) + "\n")

	if len(b.filters) == 0 && !b.addingFilter {
		content.WriteString(b.helpStyle.Render("(no filters)") + "\n")
		content.WriteString(b.helpStyle.Render("Enter: add filter") + "\n")
	} else {
		for i, f := range b.filters {
			cursor := " "
			if b.panel == PanelFilters && i == b.filterCursor {
				cursor = ">"
			}
			line := f.Field + " " + f.Operator + " " + formatValue(f.Value)
			if b.panel == PanelFilters && i == b.filterCursor {
				line = b.selectedStyle.Render(line)
			}
			content.WriteString(cursor + " " + line + "\n")
		}
	}

	// Filter adding dialog
	if b.addingFilter {
		content.WriteString("\n" + strings.Repeat("─", width-4) + "\n")
		content.WriteString("Add Filter Rule\n")

		switch b.filterStep {
		case 0:
			content.WriteString("Select field:\n")
			for i, f := range AvailableFields {
				cursor := " "
				name := f
				if i == b.filterField {
					cursor = ">"
					name = b.selectedStyle.Render(f)
				}
				content.WriteString(cursor + " " + name + "\n")
			}
		case 1:
			content.WriteString("Field: " + AvailableFields[b.filterField] + "\n")
			content.WriteString("Select operator:\n")
			for i, op := range filterOperators {
				cursor := " "
				name := op
				if i == b.filterOp {
					cursor = ">"
					name = b.selectedStyle.Render(op)
				}
				content.WriteString(cursor + " " + name + "\n")
			}
		case 2:
			content.WriteString("Field: " + AvailableFields[b.filterField] + "\n")
			content.WriteString("Operator: " + filterOperators[b.filterOp] + "\n")
			content.WriteString("Value: " + b.filterValue.View() + "\n")
		}
	}

	style := b.panelStyle
	if b.panel == PanelFilters {
		style = b.activePanelStyle
	}

	return style.Width(width).Height(height).Render(content.String())
}

func (b *Builder) renderSortPanel(width, height int) string {
	var content strings.Builder
	content.WriteString("Sort\n")
	content.WriteString(strings.Repeat("─", width-4) + "\n")

	if len(b.sortRules) == 0 && !b.addingSort {
		content.WriteString(b.helpStyle.Render("(no sort rules)") + "\n")
		content.WriteString(b.helpStyle.Render("Enter: add sort rule") + "\n")
	} else {
		for i, s := range b.sortRules {
			cursor := " "
			if b.panel == PanelSort && i == b.sortCursor {
				cursor = ">"
			}
			line := s.Field + " " + s.Direction
			if b.panel == PanelSort && i == b.sortCursor {
				line = b.selectedStyle.Render(line)
			}
			content.WriteString(cursor + " " + line + "\n")
		}
	}

	// Sort adding dialog
	if b.addingSort {
		content.WriteString("\n" + strings.Repeat("─", width-4) + "\n")
		content.WriteString("Add Sort Rule\n")

		switch b.sortStep {
		case 0:
			content.WriteString("Select field:\n")
			for i, f := range AvailableFields {
				cursor := " "
				name := f
				if i == b.sortField {
					cursor = ">"
					name = b.selectedStyle.Render(f)
				}
				content.WriteString(cursor + " " + name + "\n")
			}
		case 1:
			content.WriteString("Field: " + AvailableFields[b.sortField] + "\n")
			content.WriteString("Direction:\n")
			dirs := []string{"asc", "desc"}
			for i, d := range dirs {
				cursor := " "
				name := d
				if i == b.sortDirection {
					cursor = ">"
					name = b.selectedStyle.Render(d)
				}
				content.WriteString(cursor + " " + name + "\n")
			}
		}
	}

	style := b.panelStyle
	if b.panel == PanelSort {
		style = b.activePanelStyle
	}

	return style.Width(width).Height(height).Render(content.String())
}

func formatInt(i int) string {
	if i == 0 {
		return "(default)"
	}
	// Simple int to string
	if i < 10 {
		return string(rune('0' + i))
	}
	result := ""
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}

func formatAlign(a string) string {
	if a == "" {
		return "(default)"
	}
	return a
}

func formatValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
