package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestPageNumberPaginationBoundariesInvalidPagesAndMaxPageSize(t *testing.T) {
	request := NewRequest(httptest.NewRequest(http.MethodGet, "/posts/?page=2&page_size=50", nil))
	paginator := PageNumberPagination{PageSize: 2, MaxPageSize: 3}

	result, err := paginator.Paginate(request, paginationItems("a", "b", "c", "d", "e"))
	if err != nil {
		t.Fatalf("Paginate() error = %v", err)
	}
	if result.Count != 5 || result.Previous == "" {
		t.Fatalf("pagination links/count = %#v", result)
	}
	if !reflect.DeepEqual(result.Results, paginationItems("d", "e")) {
		t.Fatalf("results = %#v", result.Results)
	}
	if result.Next != "" {
		t.Fatalf("next = %q, want empty on final page", result.Next)
	}

	badPage := NewRequest(httptest.NewRequest(http.MethodGet, "/posts/?page=99", nil))
	_, err = paginator.Paginate(badPage, paginationItems("a"))
	if !errors.Is(err, ErrPagination) {
		t.Fatalf("bad page error = %v, want ErrPagination", err)
	}
}

func TestLimitOffsetPaginationClampsLimitAndBuildsLinks(t *testing.T) {
	request := NewRequest(httptest.NewRequest(http.MethodGet, "/posts/?limit=10&offset=1", nil))
	paginator := LimitOffsetPagination{DefaultLimit: 2, MaxLimit: 2}

	result, err := paginator.Paginate(request, paginationItems("a", "b", "c", "d"))
	if err != nil {
		t.Fatalf("Paginate() error = %v", err)
	}
	if result.Count != 4 || !reflect.DeepEqual(result.Results, paginationItems("b", "c")) {
		t.Fatalf("result = %#v", result)
	}
	if result.Next != "/posts/?limit=2&offset=3" || result.Previous != "/posts/?limit=2&offset=0" {
		t.Fatalf("links = next:%q previous:%q", result.Next, result.Previous)
	}

	invalid := NewRequest(httptest.NewRequest(http.MethodGet, "/posts/?offset=-1", nil))
	_, err = paginator.Paginate(invalid, paginationItems("a"))
	if !errors.Is(err, ErrPagination) {
		t.Fatalf("invalid offset error = %v, want ErrPagination", err)
	}
}

func TestCursorPaginationOrdersResultsAndUsesCursorLinks(t *testing.T) {
	items := []any{
		map[string]any{"id": int64(3), "title": "c"},
		map[string]any{"id": int64(1), "title": "a"},
		map[string]any{"id": int64(2), "title": "b"},
	}
	paginator := CursorPagination{PageSize: 2, Ordering: "id"}
	firstRequest := NewRequest(httptest.NewRequest(http.MethodGet, "/posts/", nil))

	first, err := paginator.Paginate(firstRequest, items)
	if err != nil {
		t.Fatalf("first Paginate() error = %v", err)
	}
	if first.Results[0].(map[string]any)["id"] != int64(1) || first.Results[1].(map[string]any)["id"] != int64(2) {
		t.Fatalf("first results = %#v", first.Results)
	}
	if first.Next == "" || first.Previous != "" {
		t.Fatalf("first links = %#v", first)
	}

	secondRequest := NewRequest(httptest.NewRequest(http.MethodGet, first.Next, nil))
	second, err := paginator.Paginate(secondRequest, items)
	if err != nil {
		t.Fatalf("second Paginate() error = %v", err)
	}
	if len(second.Results) != 1 || second.Results[0].(map[string]any)["id"] != int64(3) || second.Next != "" || second.Previous == "" {
		t.Fatalf("second result = %#v", second)
	}
}

func paginationItems(values ...string) []any {
	items := make([]any, len(values))
	for index, value := range values {
		items[index] = value
	}
	return items
}
