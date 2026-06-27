package dialects

import "testing"

func TestIntReturnsPointer(t *testing.T) {
	value := Int(42)
	if value == nil || *value != 42 {
		t.Fatalf("Int(42) = %#v", value)
	}
}

func TestQuoteIdentEscapesEmbeddedQuotes(t *testing.T) {
	if got := QuoteIdent(`bad"name`); got != `"bad""name"` {
		t.Fatalf("QuoteIdent() = %q", got)
	}
}

func TestLimitOffsetOmitsUnsetParts(t *testing.T) {
	if got := RenderLimitOffset(LimitOffset{Limit: Int(5)}); got != "LIMIT 5" {
		t.Fatalf("limit only = %q", got)
	}
	if got := RenderLimitOffset(LimitOffset{Offset: Int(7)}); got != "OFFSET 7" {
		t.Fatalf("offset only = %q", got)
	}
	if got := RenderLimitOffset(LimitOffset{}); got != "" {
		t.Fatalf("empty = %q", got)
	}
}
