package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFilterSetSupportsExactLookupsSearchAndOrdering(t *testing.T) {
	items := []map[string]any{
		{"id": int64(1), "title": "Go Intro", "status": "published", "views": int64(5)},
		{"id": int64(2), "title": "Advanced Go", "status": "published", "views": int64(15)},
		{"id": int64(3), "title": "Python", "status": "published", "views": int64(25)},
		{"id": int64(4), "title": "Draft Go", "status": "draft", "views": int64(50)},
	}
	request := NewRequest(httptest.NewRequest(http.MethodGet, "/posts/?status=published&views__gte=10&search=go&ordering=-views", nil))
	filterSet := FilterSet{
		ExactFields:    []string{"status"},
		LookupFields:   map[string][]string{"views": {"gte"}},
		SearchFields:   []string{"title"},
		OrderingFields: []string{"views"},
	}

	filtered, err := filterSet.Apply(context.Background(), request, items)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if len(filtered) != 1 || filtered[0]["id"] != int64(2) {
		t.Fatalf("filtered = %#v", filtered)
	}
}

func TestFilterSetSupportsCustomBackendsAndDistinct(t *testing.T) {
	items := []map[string]any{
		{"id": int64(1), "tenant": "acme", "tag": "go"},
		{"id": int64(1), "tenant": "acme", "tag": "api"},
		{"id": int64(2), "tenant": "other", "tag": "go"},
	}
	request := NewRequest(httptest.NewRequest(http.MethodGet, "/posts/", nil))
	filterSet := FilterSet{
		DistinctFields: []string{"id"},
		Backends: []FilterBackend{
			FilterBackendFunc(func(_ context.Context, _ *Request, rows []map[string]any) ([]map[string]any, error) {
				return filterRows(rows, func(row map[string]any) bool { return row["tenant"] == "acme" }), nil
			}),
		},
	}

	filtered, err := filterSet.Apply(context.Background(), request, items)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if len(filtered) != 1 || filtered[0]["id"] != int64(1) {
		t.Fatalf("filtered = %#v", filtered)
	}
}

func TestFilterSetRejectsInvalidFields(t *testing.T) {
	request := NewRequest(httptest.NewRequest(http.MethodGet, "/posts/?unknown=value", nil))
	filterSet := FilterSet{ExactFields: []string{"status"}}

	_, err := filterSet.Apply(context.Background(), request, nil)
	if !errors.Is(err, ErrFilter) {
		t.Fatalf("Apply() error = %v, want ErrFilter", err)
	}
}
