package admin

import (
	"errors"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestChangeListBuildsDisplayRowsSortingPaginationAndFilters(t *testing.T) {
	admin := ModelAdmin{
		ListDisplay:       []string{"title", "published", "summary"},
		ListEditable:      []string{"published"},
		ListPerPage:       1,
		ListMaxShowAll:    5,
		EmptyValueDisplay: "(none)",
		DateHierarchy:     "created_at",
		ComputedColumns: map[string]ComputedColumn{
			"summary": func(row map[string]any) any { return row["title"].(string) + "!" },
		},
	}
	rows := []map[string]any{
		{"id": 1, "title": "Beta", "published": false, "created_at": time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"id": 2, "title": "Alpha", "published": true, "created_at": time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"id": 3, "title": "", "published": true, "created_at": time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)},
	}
	changeList, err := BuildChangeList(admin, rows, url.Values{"o": {"title"}, "p": {"2"}, "status": {"draft"}, "_popup": {"1"}})
	if err != nil {
		t.Fatalf("BuildChangeList() error = %v", err)
	}

	if !reflect.DeepEqual(changeList.Columns, []ChangeListColumn{{Name: "title"}, {Name: "published", Editable: true}, {Name: "summary", Computed: true}}) {
		t.Fatalf("columns = %#v", changeList.Columns)
	}
	if changeList.Total != 3 || changeList.Page != 2 || changeList.PerPage != 1 || len(changeList.Rows) != 1 {
		t.Fatalf("pagination = %#v", changeList)
	}
	if got := changeList.Rows[0].Values["title"]; got != "Beta" {
		t.Fatalf("page row title = %#v", got)
	}
	if got := changeList.Rows[0].Values["published"]; got != BooleanIcon(false) {
		t.Fatalf("boolean display = %#v", got)
	}
	if got := changeList.Rows[0].Values["summary"]; got != "Beta!" {
		t.Fatalf("computed display = %#v", got)
	}
	if !changeList.BulkSelection || !changeList.Popup || changeList.PreservedFilters != "status=draft" {
		t.Fatalf("flags/filters = %#v", changeList)
	}
	if !reflect.DeepEqual(changeList.DateHierarchy, []DateBucket{{Year: 2025, Count: 1}, {Year: 2026, Count: 2}}) {
		t.Fatalf("date hierarchy = %#v", changeList.DateHierarchy)
	}
}

func TestChangeListShowAllAndInvalidQuery(t *testing.T) {
	admin := ModelAdmin{ListDisplay: []string{"title"}, ListMaxShowAll: 5}
	rows := []map[string]any{{"title": "A"}, {"title": "B"}}
	changeList, err := BuildChangeList(admin, rows, url.Values{"all": {"1"}})
	if err != nil {
		t.Fatalf("BuildChangeList(show all) error = %v", err)
	}
	if !changeList.ShowAll || len(changeList.Rows) != 2 {
		t.Fatalf("show all change list = %#v", changeList)
	}

	if _, err := BuildChangeList(admin, rows, url.Values{"o": {"missing"}}); !errors.Is(err, ErrInvalidChangeListQuery) {
		t.Fatalf("invalid ordering error = %v, want ErrInvalidChangeListQuery", err)
	}
	if _, err := BuildChangeList(admin, rows, url.Values{"p": {"bad"}}); !errors.Is(err, ErrInvalidChangeListQuery) {
		t.Fatalf("invalid page error = %v, want ErrInvalidChangeListQuery", err)
	}
}
