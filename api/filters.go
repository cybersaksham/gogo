package api

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

const (
	defaultSearchParam   = "search"
	defaultOrderingParam = "ordering"
)

// FilterBackend filters API result rows.
type FilterBackend interface {
	Filter(context.Context, *Request, []map[string]any) ([]map[string]any, error)
}

// FilterBackendFunc adapts a function into a filter backend.
type FilterBackendFunc func(context.Context, *Request, []map[string]any) ([]map[string]any, error)

// Filter runs the function backend.
func (f FilterBackendFunc) Filter(ctx context.Context, request *Request, rows []map[string]any) ([]map[string]any, error) {
	return f(ctx, request, rows)
}

// FilterSet configures safe API filtering, search, ordering, and distinct handling.
type FilterSet struct {
	ExactFields    []string
	LookupFields   map[string][]string
	SearchFields   []string
	OrderingFields []string
	Distinct       bool
	DistinctFields []string
	Backends       []FilterBackend
	SearchParam    string
	OrderingParam  string
}

// Apply filters rows using request query parameters and custom backends.
func (f FilterSet) Apply(ctx context.Context, request *Request, rows []map[string]any) ([]map[string]any, error) {
	if err := f.validateQueryFields(request); err != nil {
		return nil, err
	}
	filtered := append([]map[string]any(nil), rows...)
	var err error
	filtered, err = f.applyExactFilters(request, filtered)
	if err != nil {
		return nil, err
	}
	filtered, err = f.applyLookupFilters(request, filtered)
	if err != nil {
		return nil, err
	}
	filtered = f.applySearch(request, filtered)
	if err := f.applyOrdering(request, filtered); err != nil {
		return nil, err
	}
	for _, backend := range f.Backends {
		if backend == nil {
			continue
		}
		filtered, err = backend.Filter(ctx, request, filtered)
		if err != nil {
			return nil, err
		}
	}
	if f.Distinct || len(f.DistinctFields) > 0 {
		filtered = distinctRows(filtered, f.DistinctFields)
	}
	return filtered, nil
}

func (f FilterSet) applyExactFilters(request *Request, rows []map[string]any) ([]map[string]any, error) {
	for _, field := range f.ExactFields {
		value := request.QueryParam(field)
		if value == "" {
			continue
		}
		rows = filterRows(rows, func(row map[string]any) bool {
			return matchesLookup(row[field], "exact", value)
		})
	}
	return rows, nil
}

func (f FilterSet) applyLookupFilters(request *Request, rows []map[string]any) ([]map[string]any, error) {
	query := request.Raw().URL.Query()
	for field, lookups := range f.LookupFields {
		for _, lookup := range lookups {
			param := field + "__" + lookup
			value := query.Get(param)
			if value == "" {
				continue
			}
			lookupName := lookup
			rows = filterRows(rows, func(row map[string]any) bool {
				return matchesLookup(row[field], lookupName, value)
			})
		}
	}
	return rows, nil
}

func (f FilterSet) applySearch(request *Request, rows []map[string]any) []map[string]any {
	searchParam := stringDefault(f.SearchParam, defaultSearchParam)
	term := strings.ToLower(strings.TrimSpace(request.QueryParam(searchParam)))
	if term == "" || len(f.SearchFields) == 0 {
		return rows
	}
	return filterRows(rows, func(row map[string]any) bool {
		for _, field := range f.SearchFields {
			if strings.Contains(strings.ToLower(fmt.Sprint(row[field])), term) {
				return true
			}
		}
		return false
	})
}

func (f FilterSet) applyOrdering(request *Request, rows []map[string]any) error {
	orderingParam := stringDefault(f.OrderingParam, defaultOrderingParam)
	ordering := strings.TrimSpace(request.QueryParam(orderingParam))
	if ordering == "" {
		return nil
	}
	allowed := stringSet(f.OrderingFields)
	fields := splitCSV(ordering)
	for _, field := range fields {
		name := strings.TrimPrefix(field, "-")
		if _, ok := allowed[name]; !ok {
			return fmt.Errorf("%w: invalid ordering field %s", ErrFilter, name)
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		for _, field := range fields {
			descending := strings.HasPrefix(field, "-")
			name := strings.TrimPrefix(field, "-")
			cmp := compareFilterValues(rows[i][name], rows[j][name])
			if cmp == 0 {
				continue
			}
			if descending {
				return cmp > 0
			}
			return cmp < 0
		}
		return false
	})
	return nil
}

func (f FilterSet) validateQueryFields(request *Request) error {
	query := request.Raw().URL.Query()
	allowedExact := stringSet(f.ExactFields)
	allowedOrdering := stringDefault(f.OrderingParam, defaultOrderingParam)
	allowedSearch := stringDefault(f.SearchParam, defaultSearchParam)
	for key := range query {
		if isFilterControlParam(key) || key == allowedOrdering || key == allowedSearch {
			continue
		}
		if _, ok := allowedExact[key]; ok {
			continue
		}
		field, lookup, ok := strings.Cut(key, "__")
		if ok && lookupAllowed(f.LookupFields, field, lookup) {
			continue
		}
		return fmt.Errorf("%w: invalid field %s", ErrFilter, key)
	}
	return nil
}

func lookupAllowed(lookups map[string][]string, field, lookup string) bool {
	for _, allowed := range lookups[field] {
		if allowed == lookup {
			return true
		}
	}
	return false
}

func matchesLookup(value any, lookup, expected string) bool {
	switch lookup {
	case "exact":
		return fmt.Sprint(value) == expected
	case "iexact":
		return strings.EqualFold(fmt.Sprint(value), expected)
	case "contains":
		return strings.Contains(fmt.Sprint(value), expected)
	case "icontains":
		return strings.Contains(strings.ToLower(fmt.Sprint(value)), strings.ToLower(expected))
	case "startswith":
		return strings.HasPrefix(fmt.Sprint(value), expected)
	case "istartswith":
		return strings.HasPrefix(strings.ToLower(fmt.Sprint(value)), strings.ToLower(expected))
	case "endswith":
		return strings.HasSuffix(fmt.Sprint(value), expected)
	case "iendswith":
		return strings.HasSuffix(strings.ToLower(fmt.Sprint(value)), strings.ToLower(expected))
	case "gt", "gte", "lt", "lte":
		cmp := compareFilterValues(value, expected)
		switch lookup {
		case "gt":
			return cmp > 0
		case "gte":
			return cmp >= 0
		case "lt":
			return cmp < 0
		default:
			return cmp <= 0
		}
	case "in":
		for _, item := range splitCSV(expected) {
			if matchesLookup(value, "exact", item) {
				return true
			}
		}
		return false
	case "isnull":
		wantNull := strings.EqualFold(expected, "true") || expected == "1"
		return (value == nil) == wantNull
	default:
		return false
	}
}

func compareFilterValues(left, right any) int {
	return compareCursorValues(left, right)
}

func filterRows(rows []map[string]any, keep func(map[string]any) bool) []map[string]any {
	filtered := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if keep(row) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func distinctRows(rows []map[string]any, fields []string) []map[string]any {
	seen := map[string]struct{}{}
	distinct := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		key := distinctKey(row, fields)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		distinct = append(distinct, row)
	}
	return distinct
}

func distinctKey(row map[string]any, fields []string) string {
	if len(fields) == 0 {
		keys := make([]string, 0, len(row))
		for key := range row {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		fields = keys
	}
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		parts = append(parts, field+"="+fmt.Sprint(row[field]))
	}
	return strings.Join(parts, "\x00")
}

func splitCSV(value string) []string {
	raw := strings.Split(value, ",")
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		item = strings.TrimSpace(item)
		if item != "" {
			values = append(values, item)
		}
	}
	return values
}

func isFilterControlParam(param string) bool {
	switch param {
	case defaultPageQueryParam, defaultPageSizeQueryParam, defaultLimitQueryParam, defaultOffsetQueryParam, defaultCursorQueryParam:
		return true
	default:
		return false
	}
}
