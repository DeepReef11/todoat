package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SetupViewsFolder creates views directory with default view files.
// Returns true if folder was created, false if it already existed.
func SetupViewsFolder(viewsDir string) (bool, error) {
	if _, err := os.Stat(viewsDir); err == nil {
		return false, nil // Already exists
	}

	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		return false, err
	}

	// Write default.yaml
	defaultYAML := `name: default
description: Standard task display for everyday use (excludes completed tasks)
fields:
  - name: status
    width: 12
  - name: summary
    width: 40
  - name: priority
    width: 10
  - name: due_date
    width: 12
  - name: tags
    width: 20
  - name: recurrence
    width: 5
filters:
  - field: status
    operator: ne
    value: DONE
`
	if err := os.WriteFile(filepath.Join(viewsDir, "default.yaml"), []byte(defaultYAML), 0644); err != nil {
		return false, err
	}

	// Write all.yaml
	allYAML := `name: all
description: Comprehensive display showing all task metadata
fields:
  - name: status
  - name: summary
  - name: description
  - name: priority
  - name: due_date
  - name: start_date
  - name: created
  - name: modified
  - name: completed
  - name: tags
  - name: uid
  - name: parent
`
	if err := os.WriteFile(filepath.Join(viewsDir, "all.yaml"), []byte(allYAML), 0644); err != nil {
		return false, err
	}

	return true, nil
}

// Loader handles loading views from disk and built-in sources
type Loader struct {
	viewsDir string
}

// NewLoader creates a new view loader
func NewLoader(viewsDir string) *Loader {
	return &Loader{viewsDir: viewsDir}
}

// ValidateViewName checks if a view name is safe to use in file paths.
// It rejects names containing path traversal sequences or invalid characters.
// This function is exported so it can be called from view creation code.
func ValidateViewName(name string) error {
	// Reject empty names
	if name == "" {
		return fmt.Errorf("view name cannot be empty")
	}

	// Reject names containing path separators
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("invalid view name '%s': contains path separator", name)
	}

	// Reject names containing ".." (path traversal)
	if strings.Contains(name, "..") {
		return fmt.Errorf("invalid view name '%s': contains path traversal sequence", name)
	}

	// Reject names starting with "." (hidden files)
	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("invalid view name '%s': cannot start with '.'", name)
	}

	return nil
}

// LoadView loads a view by name
// First checks disk for user overrides, then falls back to built-in views
func (l *Loader) LoadView(name string) (*View, error) {
	normalizedName := strings.ToLower(name)
	if normalizedName == "" {
		normalizedName = "default"
	}

	// Validate view name to prevent path traversal (skip for built-in names)
	if normalizedName != "default" && normalizedName != "all" {
		if err := ValidateViewName(name); err != nil {
			return nil, err
		}
	}

	// Check disk first for user override (even for built-in names like "default" and "all")
	if l.viewsDir != "" {
		viewPath := filepath.Join(l.viewsDir, normalizedName+".yaml")
		if _, err := os.Stat(viewPath); err == nil {
			// User file exists - load it (overrides built-in if applicable)
			return l.loadFromDisk(normalizedName, viewPath)
		}
	}

	// Fall back to built-in views
	switch normalizedName {
	case "default":
		return DefaultView(), nil
	case "all":
		return AllView(), nil
	}

	// Custom view not found on disk
	return nil, fmt.Errorf("view '%s' not found", name)
}

// loadFromDisk loads a view from a YAML file with path validation
func (l *Loader) loadFromDisk(name, viewPath string) (*View, error) {
	// Double-check that the resolved path is within the views directory
	absViewsDir, err := filepath.Abs(l.viewsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve views directory: %w", err)
	}
	absViewPath, err := filepath.Abs(viewPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve view path: %w", err)
	}

	// Ensure the view path is within the views directory
	if !strings.HasPrefix(absViewPath, absViewsDir+string(filepath.Separator)) {
		return nil, fmt.Errorf("invalid view name '%s': path traversal detected", name)
	}

	data, err := os.ReadFile(viewPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("view '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to read view '%s': %w", name, err)
	}

	var view View
	if err := yaml.Unmarshal(data, &view); err != nil {
		return nil, fmt.Errorf("failed to parse view '%s': %w", name, err)
	}

	// Validate the view
	if err := l.validateView(&view); err != nil {
		return nil, fmt.Errorf("invalid view '%s': %w", name, err)
	}

	return &view, nil
}

// ListViews returns all available views (built-in and custom)
func (l *Loader) ListViews() ([]ViewInfo, error) {
	var views []ViewInfo

	// Track which built-in views have user overrides
	defaultOverridden := false
	allOverridden := false

	// Check for user overrides of built-in views
	if l.viewsDir != "" {
		if _, err := os.Stat(filepath.Join(l.viewsDir, "default.yaml")); err == nil {
			defaultOverridden = true
		}
		if _, err := os.Stat(filepath.Join(l.viewsDir, "all.yaml")); err == nil {
			allOverridden = true
		}
	}

	// Add default view entry
	if defaultOverridden {
		// User has overridden default - load their version for description
		view, err := l.LoadView("default")
		desc := "Standard task display for everyday use"
		if err == nil {
			desc = view.Description
		}
		views = append(views, ViewInfo{
			Name:        "default",
			Description: desc,
			BuiltIn:     false,
			Overrides:   true, // User file overrides built-in
		})
	} else {
		views = append(views, ViewInfo{
			Name:        "default",
			Description: "Standard task display for everyday use",
			BuiltIn:     true,
			Overrides:   false,
		})
	}

	// Add all view entry
	if allOverridden {
		// User has overridden all - load their version for description
		view, err := l.LoadView("all")
		desc := "Comprehensive display showing all task metadata"
		if err == nil {
			desc = view.Description
		}
		views = append(views, ViewInfo{
			Name:        "all",
			Description: desc,
			BuiltIn:     false,
			Overrides:   true, // User file overrides built-in
		})
	} else {
		views = append(views, ViewInfo{
			Name:        "all",
			Description: "Comprehensive display showing all task metadata",
			BuiltIn:     true,
			Overrides:   false,
		})
	}

	// Add custom views from disk (excluding built-in names already handled)
	if l.viewsDir != "" {
		entries, err := os.ReadDir(l.viewsDir)
		if err != nil {
			if os.IsNotExist(err) {
				return views, nil
			}
			return nil, fmt.Errorf("failed to read views directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
				continue
			}

			name := strings.TrimSuffix(entry.Name(), ".yaml")
			// Skip if this name matches a built-in view (already handled above)
			if name == "default" || name == "all" {
				continue
			}

			// Try to load and extract description
			view, err := l.LoadView(name)
			desc := ""
			if err == nil {
				desc = view.Description
			}

			views = append(views, ViewInfo{
				Name:        name,
				Description: desc,
				BuiltIn:     false,
				Overrides:   false, // Custom view, doesn't override a built-in
			})
		}
	}

	return views, nil
}

// ViewInfo contains metadata about a view
type ViewInfo struct {
	Name        string
	Description string
	BuiltIn     bool
	Overrides   bool // True if user file overrides a built-in view
}

// ViewExists checks if a view exists (either built-in or custom)
func (l *Loader) ViewExists(name string) bool {
	// Check built-in views first
	switch strings.ToLower(name) {
	case "default", "", "all":
		return true
	}

	// Validate view name to prevent path traversal
	if err := ValidateViewName(name); err != nil {
		return false
	}

	// Check if custom view file exists
	if l.viewsDir == "" {
		return false
	}

	viewPath := filepath.Join(l.viewsDir, name+".yaml")

	// Double-check that the resolved path is within the views directory
	absViewsDir, err := filepath.Abs(l.viewsDir)
	if err != nil {
		return false
	}
	absViewPath, err := filepath.Abs(viewPath)
	if err != nil {
		return false
	}

	// Ensure the view path is within the views directory
	if !strings.HasPrefix(absViewPath, absViewsDir+string(filepath.Separator)) {
		return false
	}

	_, err = os.Stat(viewPath)
	return err == nil
}

// validateView checks that a view configuration is valid
func (l *Loader) validateView(v *View) error {
	if len(v.Fields) == 0 {
		return fmt.Errorf("view must have at least one field")
	}

	validFields := make(map[string]bool)
	for _, f := range AvailableFields {
		validFields[f] = true
	}

	for _, f := range v.Fields {
		if !validFields[f.Name] {
			return fmt.Errorf("unknown field: %s", f.Name)
		}
	}

	// Validate filters
	for _, filter := range v.Filters {
		if !validFields[filter.Field] {
			return fmt.Errorf("unknown filter field: %s", filter.Field)
		}
		if !isValidOperator(filter.Operator) {
			return fmt.Errorf("invalid operator: %s", filter.Operator)
		}
	}

	// Validate sort rules
	for _, sort := range v.Sort {
		if !validFields[sort.Field] {
			return fmt.Errorf("unknown sort field: %s", sort.Field)
		}
		dir := strings.ToLower(sort.Direction)
		if dir != "asc" && dir != "desc" {
			return fmt.Errorf("invalid sort direction: %s (must be 'asc' or 'desc')", sort.Direction)
		}
	}

	return nil
}

// isValidOperator checks if an operator is valid
func isValidOperator(op string) bool {
	validOps := map[string]bool{
		"eq":       true,
		"ne":       true,
		"lt":       true,
		"lte":      true,
		"gt":       true,
		"gte":      true,
		"contains": true,
		"in":       true,
		"not_in":   true,
		"regex":    true,
	}
	return validOps[strings.ToLower(op)]
}
