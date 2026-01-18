package views

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"todoat/backend"
)

// FilterTasks applies all filters to a list of tasks
func FilterTasks(tasks []backend.Task, filters []Filter) []backend.Task {
	if len(filters) == 0 {
		return tasks
	}

	var result []backend.Task
	for _, t := range tasks {
		if matchesAllFilters(&t, filters) {
			result = append(result, t)
		}
	}
	return result
}

// matchesAllFilters checks if a task matches all filters (AND logic)
func matchesAllFilters(t *backend.Task, filters []Filter) bool {
	for _, f := range filters {
		if !matchesFilter(t, f) {
			return false
		}
	}
	return true
}

// matchesFilter checks if a task matches a single filter
func matchesFilter(t *backend.Task, f Filter) bool {
	fieldValue := getFieldValue(t, f.Field)
	return compareValue(fieldValue, f.Operator, f.Value, f.Field)
}

// getFieldValue extracts a field value from a task
func getFieldValue(t *backend.Task, field string) any {
	switch field {
	case "status":
		return string(t.Status)
	case "summary":
		return t.Summary
	case "description":
		return t.Description
	case "priority":
		return t.Priority
	case "due_date":
		return t.DueDate
	case "start_date":
		return t.StartDate
	case "created":
		return t.Created
	case "modified":
		return t.Modified
	case "completed":
		return t.Completed
	case "tags":
		return t.Categories
	case "uid":
		return t.ID
	case "parent":
		return t.ParentID
	default:
		return nil
	}
}

// compareValue compares a field value against a filter value using the specified operator
func compareValue(fieldValue any, operator string, filterValue any, fieldName string) bool {
	op := strings.ToLower(operator)

	switch op {
	case "eq":
		return equals(fieldValue, filterValue, fieldName)
	case "ne":
		return !equals(fieldValue, filterValue, fieldName)
	case "lt":
		return lessThan(fieldValue, filterValue, fieldName, false)
	case "lte":
		return lessThan(fieldValue, filterValue, fieldName, true)
	case "gt":
		return greaterThan(fieldValue, filterValue, fieldName, false)
	case "gte":
		return greaterThan(fieldValue, filterValue, fieldName, true)
	case "contains":
		return contains(fieldValue, filterValue)
	case "in":
		return inList(fieldValue, filterValue)
	case "not_in":
		return !inList(fieldValue, filterValue)
	case "regex":
		return matchesRegex(fieldValue, filterValue)
	default:
		return false
	}
}

// equals checks if two values are equal
func equals(fieldValue, filterValue any, fieldName string) bool {
	// Handle nil/empty cases
	if fieldValue == nil && filterValue == nil {
		return true
	}

	// Handle status comparison
	if fieldName == "status" {
		return normalizeStatus(toString(fieldValue)) == normalizeStatus(toString(filterValue))
	}

	// Handle date comparison
	if isDateField(fieldName) {
		fv := toTime(fieldValue)
		filterv := parseFilterDate(filterValue)
		if fv == nil || filterv == nil {
			return fv == nil && filterv == nil
		}
		// Compare dates only (not time)
		return fv.Truncate(24 * time.Hour).Equal(filterv.Truncate(24 * time.Hour))
	}

	// String comparison
	return toString(fieldValue) == toString(filterValue)
}

// lessThan checks if fieldValue < filterValue
func lessThan(fieldValue, filterValue any, fieldName string, orEqual bool) bool {
	if isDateField(fieldName) {
		fv := toTime(fieldValue)
		filterv := parseFilterDate(filterValue)
		if fv == nil || filterv == nil {
			return false
		}
		fvDay := fv.Truncate(24 * time.Hour)
		filtervDay := filterv.Truncate(24 * time.Hour)
		if orEqual {
			return fvDay.Before(filtervDay) || fvDay.Equal(filtervDay)
		}
		return fvDay.Before(filtervDay)
	}

	// Numeric comparison
	fvNum, okFv := toInt(fieldValue)
	filterNum, okFilter := toInt(filterValue)
	if okFv && okFilter {
		if orEqual {
			return fvNum <= filterNum
		}
		return fvNum < filterNum
	}

	return false
}

// greaterThan checks if fieldValue > filterValue
func greaterThan(fieldValue, filterValue any, fieldName string, orEqual bool) bool {
	if isDateField(fieldName) {
		fv := toTime(fieldValue)
		filterv := parseFilterDate(filterValue)
		if fv == nil || filterv == nil {
			return false
		}
		fvDay := fv.Truncate(24 * time.Hour)
		filtervDay := filterv.Truncate(24 * time.Hour)
		if orEqual {
			return fvDay.After(filtervDay) || fvDay.Equal(filtervDay)
		}
		return fvDay.After(filtervDay)
	}

	// Numeric comparison
	fvNum, okFv := toInt(fieldValue)
	filterNum, okFilter := toInt(filterValue)
	if okFv && okFilter {
		if orEqual {
			return fvNum >= filterNum
		}
		return fvNum > filterNum
	}

	return false
}

// contains checks if fieldValue contains filterValue (for strings and arrays)
func contains(fieldValue, filterValue any) bool {
	fvStr := toString(fieldValue)
	filterStr := toString(filterValue)

	// For tags/categories, check if any tag matches
	if strings.Contains(fvStr, ",") {
		tags := strings.Split(fvStr, ",")
		for _, tag := range tags {
			if strings.EqualFold(strings.TrimSpace(tag), filterStr) {
				return true
			}
		}
		return false
	}

	return strings.Contains(strings.ToLower(fvStr), strings.ToLower(filterStr))
}

// inList checks if fieldValue is in the filterValue list
func inList(fieldValue, filterValue any) bool {
	fvStr := toString(fieldValue)

	// Handle list as []interface{} or []string
	switch list := filterValue.(type) {
	case []any:
		for _, v := range list {
			if strings.EqualFold(fvStr, toString(v)) {
				return true
			}
		}
	case []string:
		for _, v := range list {
			if strings.EqualFold(fvStr, v) {
				return true
			}
		}
	}

	return false
}

// matchesRegex checks if fieldValue matches the regex pattern
func matchesRegex(fieldValue, filterValue any) bool {
	pattern := toString(filterValue)
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(toString(fieldValue))
}

// SortTasks sorts tasks according to sort rules, preserving hierarchy
func SortTasks(tasks []backend.Task, rules []SortRule) []backend.Task {
	if len(rules) == 0 {
		return tasks
	}

	// Build parent-child map for hierarchy preservation
	taskMap := make(map[string]*backend.Task)
	childMap := make(map[string][]*backend.Task)
	var rootTasks []*backend.Task

	for i := range tasks {
		t := &tasks[i]
		taskMap[t.ID] = t
		if t.ParentID == "" {
			rootTasks = append(rootTasks, t)
		} else {
			childMap[t.ParentID] = append(childMap[t.ParentID], t)
		}
	}

	// Sort root tasks
	sortTaskSlice(rootTasks, rules)

	// Sort children at each level
	for _, children := range childMap {
		sortTaskSlice(children, rules)
	}

	// Build result preserving hierarchy (parents before children)
	var result []backend.Task
	var addTaskWithChildren func(t *backend.Task)
	addTaskWithChildren = func(t *backend.Task) {
		result = append(result, *t)
		children := childMap[t.ID]
		for _, child := range children {
			addTaskWithChildren(child)
		}
	}

	for _, t := range rootTasks {
		addTaskWithChildren(t)
	}

	return result
}

// sortTaskSlice sorts a slice of task pointers by sort rules
func sortTaskSlice(tasks []*backend.Task, rules []SortRule) {
	if len(tasks) <= 1 || len(rules) == 0 {
		return
	}

	// Simple bubble sort for now (can optimize later)
	for i := 0; i < len(tasks); i++ {
		for j := i + 1; j < len(tasks); j++ {
			if compareTasksForSort(tasks[i], tasks[j], rules) > 0 {
				tasks[i], tasks[j] = tasks[j], tasks[i]
			}
		}
	}
}

// compareTasksForSort compares two tasks according to sort rules
// Returns: -1 if a < b, 0 if equal, 1 if a > b
func compareTasksForSort(a, b *backend.Task, rules []SortRule) int {
	for _, rule := range rules {
		aVal := getFieldValue(a, rule.Field)
		bVal := getFieldValue(b, rule.Field)

		cmp := compareForSort(aVal, bVal, rule.Field)
		if cmp != 0 {
			if strings.ToLower(rule.Direction) == "desc" {
				return -cmp
			}
			return cmp
		}
	}
	return 0
}

// compareForSort compares two values for sorting
func compareForSort(a, b any, fieldName string) int {
	// Handle nil - nil values sort last
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return 1
	}
	if b == nil {
		return -1
	}

	// Date comparison
	if isDateField(fieldName) {
		aTime := toTime(a)
		bTime := toTime(b)
		if aTime == nil && bTime == nil {
			return 0
		}
		if aTime == nil {
			return 1
		}
		if bTime == nil {
			return -1
		}
		if aTime.Before(*bTime) {
			return -1
		}
		if aTime.After(*bTime) {
			return 1
		}
		return 0
	}

	// Integer comparison
	if aInt, okA := toInt(a); okA {
		if bInt, okB := toInt(b); okB {
			if aInt < bInt {
				return -1
			}
			if aInt > bInt {
				return 1
			}
			return 0
		}
	}

	// String comparison
	aStr := strings.ToLower(toString(a))
	bStr := strings.ToLower(toString(b))
	if aStr < bStr {
		return -1
	}
	if aStr > bStr {
		return 1
	}
	return 0
}

// Helper functions

func isDateField(field string) bool {
	return field == "due_date" || field == "start_date" || field == "created" || field == "modified" || field == "completed"
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case backend.TaskStatus:
		return string(val)
	case int:
		return strconv.Itoa(val)
	case *time.Time:
		if val == nil {
			return ""
		}
		return val.Format(DefaultDateFormat)
	case time.Time:
		return val.Format(DefaultDateFormat)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func toInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i, true
		}
	}
	return 0, false
}

func toTime(v any) *time.Time {
	switch val := v.(type) {
	case *time.Time:
		return val
	case time.Time:
		return &val
	case string:
		if t, err := time.Parse(DefaultDateFormat, val); err == nil {
			return &t
		}
	}
	return nil
}

// parseFilterDate parses a filter date value which can be a date string or relative date
func parseFilterDate(v any) *time.Time {
	str := toString(v)
	if str == "" {
		return nil
	}

	// Handle relative dates
	now := time.Now().Truncate(24 * time.Hour)

	switch strings.ToLower(str) {
	case "today":
		return &now
	case "tomorrow":
		t := now.AddDate(0, 0, 1)
		return &t
	case "yesterday":
		t := now.AddDate(0, 0, -1)
		return &t
	}

	// Handle +Nd, -Nd, +Nw, +Nm formats
	if len(str) > 1 && (str[0] == '+' || str[0] == '-') {
		sign := 1
		if str[0] == '-' {
			sign = -1
		}
		numStr := str[1 : len(str)-1]
		unit := str[len(str)-1]
		if n, err := strconv.Atoi(numStr); err == nil {
			switch unit {
			case 'd', 'D':
				t := now.AddDate(0, 0, sign*n)
				return &t
			case 'w', 'W':
				t := now.AddDate(0, 0, sign*n*7)
				return &t
			case 'm', 'M':
				t := now.AddDate(0, sign*n, 0)
				return &t
			}
		}
	}

	// Try to parse as absolute date
	if t, err := time.Parse(DefaultDateFormat, str); err == nil {
		return &t
	}

	return nil
}

// normalizeStatus normalizes status strings for comparison
func normalizeStatus(s string) string {
	upper := strings.ToUpper(strings.TrimSpace(s))
	switch upper {
	case "DONE", "COMPLETED":
		return "COMPLETED"
	case "TODO", "NEEDS-ACTION":
		return "NEEDS-ACTION"
	case "IN-PROGRESS", "INPROGRESS", "IN_PROGRESS":
		return "IN-PROCESS"
	case "IN-PROCESS":
		return "IN-PROCESS"
	case "CANCELLED", "CANCELED":
		return "CANCELLED"
	default:
		return upper
	}
}
