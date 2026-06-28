package api

import (
	"encoding/base64"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

const (
	defaultPageQueryParam     = "page"
	defaultPageSizeQueryParam = "page_size"
	defaultLimitQueryParam    = "limit"
	defaultOffsetQueryParam   = "offset"
	defaultCursorQueryParam   = "cursor"
	defaultPaginationSize     = 20
)

// PaginatedResult is the normalized pagination response body.
type PaginatedResult struct {
	Count    int    `json:"count"`
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
	Results  []any  `json:"results"`
}

// PageNumberPagination paginates with page and page_size query parameters.
type PageNumberPagination struct {
	PageSize           int
	MaxPageSize        int
	PageQueryParam     string
	PageSizeQueryParam string
}

// Paginate returns one page of items.
func (p PageNumberPagination) Paginate(request *Request, items []any) (PaginatedResult, error) {
	pageParam := stringDefault(p.PageQueryParam, defaultPageQueryParam)
	pageSizeParam := stringDefault(p.PageSizeQueryParam, defaultPageSizeQueryParam)
	pageSize, err := requestedPositiveInt(request, pageSizeParam, intDefault(p.PageSize, defaultPaginationSize))
	if err != nil {
		return PaginatedResult{}, err
	}
	pageSize = clampMax(pageSize, p.MaxPageSize)
	page, err := requestedPositiveInt(request, pageParam, 1)
	if err != nil {
		return PaginatedResult{}, err
	}

	count := len(items)
	totalPages := int(math.Ceil(float64(count) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		return PaginatedResult{}, fmt.Errorf("%w: page out of range", ErrPagination)
	}
	start := (page - 1) * pageSize
	end := minInt(start+pageSize, count)
	result := PaginatedResult{Count: count, Results: copyItems(items[start:end])}
	if page < totalPages {
		result.Next = queryURL(request, map[string]string{pageParam: strconv.Itoa(page + 1), pageSizeParam: strconv.Itoa(pageSize)})
	}
	if page > 1 {
		result.Previous = queryURL(request, map[string]string{pageParam: strconv.Itoa(page - 1), pageSizeParam: strconv.Itoa(pageSize)})
	}
	return result, nil
}

// LimitOffsetPagination paginates with limit and offset query parameters.
type LimitOffsetPagination struct {
	DefaultLimit int
	MaxLimit     int
	LimitParam   string
	OffsetParam  string
}

// Paginate returns one limit/offset page of items.
func (p LimitOffsetPagination) Paginate(request *Request, items []any) (PaginatedResult, error) {
	limitParam := stringDefault(p.LimitParam, defaultLimitQueryParam)
	offsetParam := stringDefault(p.OffsetParam, defaultOffsetQueryParam)
	limit, err := requestedPositiveInt(request, limitParam, intDefault(p.DefaultLimit, defaultPaginationSize))
	if err != nil {
		return PaginatedResult{}, err
	}
	limit = clampMax(limit, p.MaxLimit)
	offset, err := requestedNonNegativeInt(request, offsetParam, 0)
	if err != nil {
		return PaginatedResult{}, err
	}

	count := len(items)
	start := minInt(offset, count)
	end := minInt(start+limit, count)
	result := PaginatedResult{Count: count, Results: copyItems(items[start:end])}
	if end < count {
		result.Next = queryURL(request, map[string]string{limitParam: strconv.Itoa(limit), offsetParam: strconv.Itoa(end)})
	}
	if offset > 0 {
		previousOffset := offset - limit
		if previousOffset < 0 {
			previousOffset = 0
		}
		result.Previous = queryURL(request, map[string]string{limitParam: strconv.Itoa(limit), offsetParam: strconv.Itoa(previousOffset)})
	}
	return result, nil
}

// CursorPagination paginates ordered items after an encoded cursor value.
type CursorPagination struct {
	PageSize           int
	MaxPageSize        int
	Ordering           string
	CursorQueryParam   string
	PageSizeQueryParam string
}

// Paginate returns one cursor page of items.
func (p CursorPagination) Paginate(request *Request, items []any) (PaginatedResult, error) {
	cursorParam := stringDefault(p.CursorQueryParam, defaultCursorQueryParam)
	pageSizeParam := stringDefault(p.PageSizeQueryParam, defaultPageSizeQueryParam)
	pageSize, err := requestedPositiveInt(request, pageSizeParam, intDefault(p.PageSize, defaultPaginationSize))
	if err != nil {
		return PaginatedResult{}, err
	}
	pageSize = clampMax(pageSize, p.MaxPageSize)
	ordering := stringDefault(p.Ordering, "id")
	descending := strings.HasPrefix(ordering, "-")
	ordering = strings.TrimPrefix(ordering, "-")

	ordered := copyItems(items)
	sort.SliceStable(ordered, func(i, j int) bool {
		cmp := compareCursorValues(orderingValue(ordered[i], ordering), orderingValue(ordered[j], ordering))
		if descending {
			return cmp > 0
		}
		return cmp < 0
	})

	cursor := request.QueryParam(cursorParam)
	cursorValue := ""
	if cursor != "" {
		decoded, err := decodeCursor(cursor)
		if err != nil {
			return PaginatedResult{}, err
		}
		cursorValue = decoded
	}
	start := cursorStart(ordered, ordering, cursorValue, descending)
	end := minInt(start+pageSize, len(ordered))
	result := PaginatedResult{Count: len(items), Results: copyItems(ordered[start:end])}
	if end < len(ordered) && len(result.Results) > 0 {
		nextValue := fmt.Sprint(orderingValue(result.Results[len(result.Results)-1], ordering))
		result.Next = queryURL(request, map[string]string{cursorParam: encodeCursor(nextValue), pageSizeParam: strconv.Itoa(pageSize)})
	}
	if start > 0 {
		previousStart := start - pageSize
		if previousStart < 0 {
			previousStart = 0
		}
		previousValue := fmt.Sprint(orderingValue(ordered[previousStart], ordering))
		result.Previous = queryURL(request, map[string]string{cursorParam: encodeCursor(previousValue), pageSizeParam: strconv.Itoa(pageSize)})
	}
	return result, nil
}

func requestedPositiveInt(request *Request, param string, fallback int) (int, error) {
	value, err := requestedNonNegativeInt(request, param, fallback)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, fmt.Errorf("%w: %s must be positive", ErrPagination, param)
	}
	return value, nil
}

func requestedNonNegativeInt(request *Request, param string, fallback int) (int, error) {
	raw := request.QueryParam(param)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%w: invalid %s", ErrPagination, param)
	}
	return value, nil
}

func queryURL(request *Request, values map[string]string) string {
	copied := *request.Raw().URL
	query := copied.Query()
	for key, value := range values {
		query.Set(key, value)
	}
	copied.RawQuery = query.Encode()
	return copied.String()
}

func cursorStart(items []any, ordering, cursorValue string, descending bool) int {
	if cursorValue == "" {
		return 0
	}
	for index, item := range items {
		cmp := compareCursorValues(orderingValue(item, ordering), cursorValue)
		if descending && cmp < 0 {
			return index
		}
		if !descending && cmp > 0 {
			return index
		}
	}
	return len(items)
}

func orderingValue(item any, field string) any {
	if values, ok := item.(map[string]any); ok {
		return values[field]
	}
	return item
}

func compareCursorValues(left, right any) int {
	leftNumber, leftOK := numericCursorValue(left)
	rightNumber, rightOK := numericCursorValue(right)
	if leftOK && rightOK {
		switch {
		case leftNumber < rightNumber:
			return -1
		case leftNumber > rightNumber:
			return 1
		default:
			return 0
		}
	}
	return strings.Compare(fmt.Sprint(left), fmt.Sprint(right))
}

func numericCursorValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case string:
		parsed, err := strconv.ParseFloat(typed, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func encodeCursor(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}

func decodeCursor(value string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return "", fmt.Errorf("%w: invalid cursor", ErrPagination)
	}
	return string(decoded), nil
}

func copyItems(items []any) []any {
	return append([]any(nil), items...)
}

func intDefault(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func stringDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func clampMax(value, maxValue int) int {
	if maxValue > 0 && value > maxValue {
		return maxValue
	}
	return value
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}
