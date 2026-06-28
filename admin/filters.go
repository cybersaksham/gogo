package admin

import (
	"fmt"
	"net/url"
	"sort"
	"time"
)

// FilterKind identifies a change-list filter type.
type FilterKind string

const (
	FilterBoolean FilterKind = "boolean"
	FilterChoices FilterKind = "choices"
	FilterDate    FilterKind = "date"
	FilterRelated FilterKind = "related"
	FilterEmpty   FilterKind = "empty"
	FilterSimple  FilterKind = "simple"
)

// FilterChoice describes one selectable filter option.
type FilterChoice struct {
	Value    string
	Label    string
	Selected bool
	Count    int
}

// FilterSpec describes a configured admin filter.
type FilterSpec struct {
	Field   string
	Title   string
	Kind    FilterKind
	Choices []FilterChoice
	Match   func(map[string]any, string) bool
}

// FilterState stores render-ready filter options.
type FilterState struct {
	Field   string
	Title   string
	Kind    FilterKind
	Choices []FilterChoice
}

// FilterResult stores filtered rows and render-ready filter states.
type FilterResult struct {
	Filters []FilterState
	Rows    []map[string]any
}

// BooleanFilter creates a boolean list filter.
func BooleanFilter(field, title string) FilterSpec {
	return FilterSpec{
		Field: field,
		Title: title,
		Kind:  FilterBoolean,
		Choices: []FilterChoice{
			{Value: "1", Label: "Yes"},
			{Value: "0", Label: "No"},
		},
	}
}

// ChoicesFilter creates a fixed-choice list filter.
func ChoicesFilter(field, title string, choices []FilterChoice) FilterSpec {
	return FilterSpec{Field: field, Title: title, Kind: FilterChoices, Choices: cloneFilterChoices(choices)}
}

// DateFilter creates a year-based date hierarchy filter.
func DateFilter(field, title string) FilterSpec {
	return FilterSpec{Field: field, Title: title, Kind: FilterDate}
}

// RelatedFilter creates a related-object filter.
func RelatedFilter(field, title string, choices []FilterChoice) FilterSpec {
	return FilterSpec{Field: field, Title: title, Kind: FilterRelated, Choices: cloneFilterChoices(choices)}
}

// EmptyFieldFilter creates an empty/non-empty field filter.
func EmptyFieldFilter(field, title string) FilterSpec {
	return FilterSpec{
		Field: field,
		Title: title,
		Kind:  FilterEmpty,
		Choices: []FilterChoice{
			{Value: "1", Label: "Empty"},
			{Value: "0", Label: "Not empty"},
		},
	}
}

// SimpleListFilter creates a custom list filter.
func SimpleListFilter(field, title string, choices []FilterChoice, match func(map[string]any, string) bool) FilterSpec {
	return FilterSpec{Field: field, Title: title, Kind: FilterSimple, Choices: cloneFilterChoices(choices), Match: match}
}

// BuildFilters builds filter states and applies active filters to rows.
func BuildFilters(filters []FilterSpec, rows []map[string]any, query url.Values, facets bool) FilterResult {
	result := FilterResult{Rows: cloneRows(rows)}
	for _, spec := range filters {
		state := buildFilterState(spec, rows, query, facets)
		result.Filters = append(result.Filters, state)
		if selected := selectedFilterValue(spec, query); selected != "" {
			result.Rows = applyFilter(spec, result.Rows, selected)
		}
	}
	return result
}

func buildFilterState(spec FilterSpec, rows []map[string]any, query url.Values, facets bool) FilterState {
	choices := cloneFilterChoices(spec.Choices)
	if spec.Kind == FilterDate {
		choices = dateFilterChoices(spec, rows)
	}
	selected := selectedFilterValue(spec, query)
	for i := range choices {
		choices[i].Selected = choices[i].Value == selected
		if facets {
			choices[i].Count = filterCount(spec, rows, choices[i].Value)
		}
	}
	return FilterState{Field: spec.Field, Title: spec.Title, Kind: spec.Kind, Choices: choices}
}

func selectedFilterValue(spec FilterSpec, query url.Values) string {
	key := spec.Field
	if spec.Kind == FilterDate {
		key = spec.Field + "__year"
	}
	if spec.Kind == FilterEmpty {
		key = spec.Field + "__empty"
	}
	return query.Get(key)
}

func applyFilter(spec FilterSpec, rows []map[string]any, value string) []map[string]any {
	filtered := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if matchesFilter(spec, row, value) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func filterCount(spec FilterSpec, rows []map[string]any, value string) int {
	count := 0
	for _, row := range rows {
		if matchesFilter(spec, row, value) {
			count++
		}
	}
	return count
}

func matchesFilter(spec FilterSpec, row map[string]any, value string) bool {
	switch spec.Kind {
	case FilterBoolean:
		want := value == "1"
		got, _ := row[spec.Field].(bool)
		return got == want
	case FilterDate:
		date, ok := row[spec.Field].(time.Time)
		return ok && fmt.Sprint(date.Year()) == value
	case FilterEmpty:
		empty := row[spec.Field] == nil || row[spec.Field] == ""
		return (value == "1" && empty) || (value == "0" && !empty)
	case FilterSimple:
		if spec.Match == nil {
			return false
		}
		return spec.Match(row, value)
	default:
		return fmt.Sprint(row[spec.Field]) == value
	}
}

func dateFilterChoices(spec FilterSpec, rows []map[string]any) []FilterChoice {
	counts := map[int]int{}
	for _, row := range rows {
		if date, ok := row[spec.Field].(time.Time); ok {
			counts[date.Year()]++
		}
	}
	years := make([]int, 0, len(counts))
	for year := range counts {
		years = append(years, year)
	}
	sort.Ints(years)
	choices := make([]FilterChoice, len(years))
	for i, year := range years {
		choices[i] = FilterChoice{Value: fmt.Sprint(year), Label: fmt.Sprint(year)}
	}
	return choices
}

func cloneFilterChoices(choices []FilterChoice) []FilterChoice {
	return append([]FilterChoice(nil), choices...)
}
