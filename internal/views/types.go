package views

// DefaultDateFormat is the standard date format used throughout the views package
const DefaultDateFormat = "2006-01-02"

// View represents a task display configuration
type View struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description,omitempty"`
	Fields      []Field    `yaml:"fields"`
	Filters     []Filter   `yaml:"filters,omitempty"`
	Sort        []SortRule `yaml:"sort,omitempty"`
	Hierarchy   *Hierarchy `yaml:"hierarchy,omitempty"`
}

// Field represents a field configuration in a view
type Field struct {
	Name     string        `yaml:"name"`
	Width    int           `yaml:"width,omitempty"`
	Align    string        `yaml:"align,omitempty"`  // left, center, right
	Format   string        `yaml:"format,omitempty"` // format string for dates
	Truncate bool          `yaml:"truncate,omitempty"`
	Plugin   *PluginConfig `yaml:"plugin,omitempty"`
}

// PluginConfig represents configuration for an external plugin formatter
type PluginConfig struct {
	Command string            `yaml:"command"`
	Timeout int               `yaml:"timeout,omitempty"` // in milliseconds, default 1000ms
	Env     map[string]string `yaml:"env,omitempty"`
}

// Filter represents a filter condition
type Filter struct {
	Field    string `yaml:"field"`
	Operator string `yaml:"operator"` // eq, ne, lt, lte, gt, gte, contains, in, not_in, regex
	Value    any    `yaml:"value"`
}

// SortRule represents a sorting rule
type SortRule struct {
	Field     string `yaml:"field"`
	Direction string `yaml:"direction"` // asc, desc
}

// Hierarchy represents hierarchy display options
type Hierarchy struct {
	Enabled        bool `yaml:"enabled,omitempty"`
	IndentSize     int  `yaml:"indent_size,omitempty"`
	ShowConnectors bool `yaml:"show_connectors,omitempty"`
}

// AvailableFields returns the list of valid field names
var AvailableFields = []string{
	"status",
	"summary",
	"description",
	"priority",
	"due_date",
	"start_date",
	"created",
	"modified",
	"completed",
	"tags",
	"uid",
	"parent",
}

// DefaultView returns the built-in default view
func DefaultView() *View {
	return &View{
		Name:        "default",
		Description: "Standard task display for everyday use",
		Fields: []Field{
			{Name: "status", Width: 12},
			{Name: "summary", Width: 40},
			{Name: "priority", Width: 10},
		},
		Filters: []Filter{
			{Field: "status", Operator: "ne", Value: "DONE"},
		},
	}
}

// AllView returns the built-in 'all' view showing all fields
func AllView() *View {
	return &View{
		Name:        "all",
		Description: "Comprehensive display showing all task metadata",
		Fields: []Field{
			{Name: "status"},
			{Name: "summary"},
			{Name: "description"},
			{Name: "priority"},
			{Name: "due_date"},
			{Name: "start_date"},
			{Name: "created"},
			{Name: "modified"},
			{Name: "completed"},
			{Name: "tags"},
			{Name: "uid"},
			{Name: "parent"},
		},
	}
}
