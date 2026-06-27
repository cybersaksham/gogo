package http

import (
	"errors"
	"testing"
)

func TestPatternMatchesBuiltInConverters(t *testing.T) {
	cases := []struct {
		pattern string
		path    string
		name    string
		value   string
	}{
		{pattern: "/articles/<int:id>/", path: "/articles/42/", name: "id", value: "42"},
		{pattern: "/users/<slug:username>/", path: "/users/jane-doe_1/", name: "username", value: "jane-doe_1"},
		{pattern: "/files/<path:key>/", path: "/files/a/b/c.txt/", name: "key", value: "a/b/c.txt"},
		{pattern: "/uuid/<uuid:id>/", path: "/uuid/123e4567-e89b-12d3-a456-426614174000/", name: "id", value: "123e4567-e89b-12d3-a456-426614174000"},
	}

	for _, tc := range cases {
		t.Run(tc.pattern, func(t *testing.T) {
			pattern, err := CompilePattern(tc.pattern)
			if err != nil {
				t.Fatalf("CompilePattern() error = %v", err)
			}

			params, ok := pattern.Match(tc.path)
			if !ok {
				t.Fatalf("Match(%q) = false, want true", tc.path)
			}
			if params[tc.name] != tc.value {
				t.Fatalf("param %s = %q, want %q", tc.name, params[tc.name], tc.value)
			}
		})
	}
}

func TestPatternDoesNotMatchInvalidConverterValue(t *testing.T) {
	pattern, err := CompilePattern("/articles/<int:id>/")
	if err != nil {
		t.Fatalf("CompilePattern() error = %v", err)
	}

	if _, ok := pattern.Match("/articles/not-int/"); ok {
		t.Fatalf("Match() = true, want false")
	}
}

func TestPatternPercentDecodesParameters(t *testing.T) {
	pattern, err := CompilePattern("/search/<str:query>/")
	if err != nil {
		t.Fatalf("CompilePattern() error = %v", err)
	}

	params, ok := pattern.Match("/search/hello%20world/")
	if !ok {
		t.Fatalf("Match() = false, want true")
	}
	if params["query"] != "hello world" {
		t.Fatalf("query = %q, want decoded value", params["query"])
	}
}

func TestPatternSupportsCustomConverters(t *testing.T) {
	RegisterConverter("year", `[0-9]{4}`)
	pattern, err := CompilePattern("/archive/<year:year>/")
	if err != nil {
		t.Fatalf("CompilePattern() error = %v", err)
	}

	params, ok := pattern.Match("/archive/2026/")
	if !ok {
		t.Fatalf("Match() = false, want true")
	}
	if params["year"] != "2026" {
		t.Fatalf("year = %q, want 2026", params["year"])
	}
}

func TestRegexPatternMatchesNamedGroups(t *testing.T) {
	pattern, err := CompileRegexPattern(`^/archive/(?P<year>[0-9]{4})/$`)
	if err != nil {
		t.Fatalf("CompileRegexPattern() error = %v", err)
	}

	params, ok := pattern.Match("/archive/2026/")
	if !ok {
		t.Fatalf("Match() = false, want true")
	}
	if params["year"] != "2026" {
		t.Fatalf("year = %q, want 2026", params["year"])
	}
}

func TestPatternRejectsInvalidConverter(t *testing.T) {
	_, err := CompilePattern("/x/<missing:id>/")
	if !errors.Is(err, ErrInvalidPattern) {
		t.Fatalf("CompilePattern() error = %v, want ErrInvalidPattern", err)
	}
}

func TestPatternRejectsDuplicateParameterNames(t *testing.T) {
	_, err := CompilePattern("/x/<int:id>/<slug:id>/")
	if !errors.Is(err, ErrInvalidPattern) {
		t.Fatalf("CompilePattern() error = %v, want ErrInvalidPattern", err)
	}
}
