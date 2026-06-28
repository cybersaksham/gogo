package postgres

import (
	"errors"
	"strings"
	"testing"
)

func TestPostgresSearchTrigramIndexesValidatorsAndAggregates(t *testing.T) {
	vector := SearchVector("title", "body")
	if vector.SQL() != `to_tsvector('simple', coalesce("title", '') || ' ' || coalesce("body", ''))` {
		t.Fatalf("SearchVector SQL = %q", vector.SQL())
	}
	query := SearchQuery("gogo")
	if query.SQL() != `plainto_tsquery('simple', 'gogo')` {
		t.Fatalf("SearchQuery SQL = %q", query.SQL())
	}
	if SearchRank(vector, query).SQL() == "" || SearchHeadline("body", query).SQL() == "" {
		t.Fatal("search rank/headline SQL missing")
	}
	if Similarity("title", "gogo").SQL() != `"title" % 'gogo'` || Distance("title", "gogo").SQL() != `"title" <-> 'gogo'` || WordSimilarity("title", "gogo").SQL() == "" {
		t.Fatal("trigram SQL missing")
	}
	if sql, err := (Index{Name: "idx", Table: "blog_post", Columns: []string{"title"}, Method: GIN}).SQL("postgres"); err != nil || !strings.Contains(sql, "USING gin") {
		t.Fatalf("index SQL = %q, %v", sql, err)
	}
	if _, err := (Index{Name: "idx", Table: "blog_post", Columns: []string{"title"}, Method: GIN}).SQL("sqlite"); !errors.Is(err, ErrUnsupportedDialect) {
		t.Fatalf("unsupported index error = %v", err)
	}
	if err := ValidateArrayLength([]int{1, 2}, 1, 2); err != nil {
		t.Fatalf("ValidateArrayLength() error = %v", err)
	}
	if err := ValidateRangeBounds(1, 3, true, true); err != nil {
		t.Fatalf("ValidateRangeBounds() error = %v", err)
	}
	if err := ValidateJSONStructure(map[string]any{"name": "gogo"}, []string{"name"}); err != nil {
		t.Fatalf("ValidateJSONStructure() error = %v", err)
	}
	if ArrayAgg("id").SQL() != `array_agg("id")` || JSONObjectAgg("key", "value").SQL() == "" || StringAgg("name", ",").SQL() == "" || BoolAnd("active").SQL() == "" || BitXor("flags").SQL() == "" {
		t.Fatal("aggregate SQL missing")
	}
}
