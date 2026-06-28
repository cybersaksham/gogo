package templates

import (
	"testing"
	"time"
)

func TestDjangoCompatTags(t *testing.T) {
	now := time.Date(2026, 6, 28, 10, 30, 0, 0, time.UTC)
	engine := NewEngine(
		WithTemplates(map[string]string{
			"base": `<main>{{block "content" .}}base{{end}}</main>`,
			"page": `{{define "content"}}{{csrf_token .csrf}}|{{querystring "/search?q=go" .params}}|{{url "post" 7}}|{{firstof "" .fallback}}|{{cycle 3 "odd" "even"}}|{{widthratio 2 4 100}}|{{spaceless "<p> <b>x</b> </p>"}}|{{templatetag "openblock"}}|{{lorem 3}}|{{now "2006-01-02"}}{{end}}{{template "base" .}}`,
		}),
		WithDjangoCompat(HelperConfig{
			Now: now,
			URLResolver: func(name string, args ...any) (string, error) {
				return "/posts/7/", nil
			},
		}),
	)

	rendered, err := engine.Render("page", Context{
		"csrf":     "csrf-123",
		"fallback": "fallback",
		"params":   map[string]any{"page": 2, "q": "go api"},
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	want := `<main><input type="hidden" name="csrfmiddlewaretoken" value="csrf-123">|/search?page=2&amp;q=go&#43;api|/posts/7/|fallback|even|50|<p><b>x</b></p>|{%|lorem ipsum dolor|2026-06-28</main>`
	if rendered != want {
		t.Fatalf("Render() = %q, want %q", rendered, want)
	}
}

func TestDjangoCompatEveryFilter(t *testing.T) {
	now := time.Date(2026, 6, 28, 10, 30, 0, 0, time.UTC)
	engine := NewEngine(WithDjangoCompat(HelperConfig{Now: now}))
	context := Context{
		"when":      now,
		"past":      now.Add(-1 * time.Hour),
		"future":    now.Add(2 * time.Hour),
		"items":     []string{"go", "api"},
		"ints":      []int{1, 2},
		"dict":      map[string]any{"b": 2, "a": 1},
		"json":      map[string]any{"ok": true},
		"unsafeSeq": []string{"<b>x</b>"},
	}
	tests := map[string]struct {
		source string
		want   string
	}{
		"add":                {`{{add 2 3}}`, `5`},
		"addslashes":         {`{{addslashes "a'b"}}`, `a\'b`},
		"capfirst":           {`{{capfirst "go"}}`, `Go`},
		"center":             {`{{center "go" 6}}`, `  go  `},
		"cut":                {`{{cut "gogo" "go"}}`, ``},
		"date":               {`{{date .when "2006"}}`, `2026`},
		"default":            {`{{default "" "fallback"}}`, `fallback`},
		"default_if_none":    {`{{default_if_none .none "fallback"}}`, `fallback`},
		"dictsort":           {`{{dictsort .dict}}`, `a=1,b=2`},
		"divisibleby":        {`{{divisibleby 6 3}}`, `true`},
		"escape":             {`{{escape "<x>"}}`, `&lt;x&gt;`},
		"escapejs":           {"{{escapejs \"a\\n<\"}}", `a\n\u003C`},
		"filesizeformat":     {`{{filesizeformat 1536}}`, `1.5 KB`},
		"first":              {`{{first .items}}`, `go`},
		"floatformat":        {`{{floatformat 3.14159 2}}`, `3.14`},
		"force_escape":       {`{{force_escape "<x>"}}`, `&lt;x&gt;`},
		"get_digit":          {`{{get_digit 12345 2}}`, `4`},
		"join":               {`{{join .items ","}}`, `go,api`},
		"json_script":        {`{{json_script .json "payload"}}`, `<script id="payload" type="application/json">{"ok":true}</script>`},
		"last":               {`{{last .items}}`, `api`},
		"length":             {`{{length .items}}`, `2`},
		"length_is":          {`{{length_is .items 2}}`, `true`},
		"linebreaks":         {"{{linebreaks \"a\\nb\"}}", `<p>a<br>b</p>`},
		"linebreaksbr":       {"{{linebreaksbr \"a\\nb\"}}", `a<br>b`},
		"linenumbers":        {"{{linenumbers \"a\\nb\"}}", "1. a\n2. b"},
		"ljust":              {`{{ljust "go" 4}}`, `go  `},
		"lower":              {`{{lower "GO"}}`, `go`},
		"make_list":          {`{{join (make_list "go") ","}}`, `g,o`},
		"phone2numeric":      {`{{phone2numeric "1-800-FLOWERS"}}`, `1-800-3569377`},
		"pluralize":          {`{{pluralize 2 "item" "items"}}`, `items`},
		"pprint":             {`{{pprint .ints}}`, `[]int{1, 2}`},
		"random":             {`{{random .items}}`, `go`},
		"rjust":              {`{{rjust "go" 4}}`, `  go`},
		"safe":               {`{{safe "<b>x</b>"}}`, `<b>x</b>`},
		"safeseq":            {`{{first (safeseq .unsafeSeq)}}`, `<b>x</b>`},
		"slice":              {`{{slice "abcdef" "1:4"}}`, `bcd`},
		"slugify":            {`{{slugify "Go API!"}}`, `go-api`},
		"stringformat":       {`{{stringformat 7 "03d"}}`, `007`},
		"striptags":          {`{{striptags "<p>x</p>"}}`, `x`},
		"time":               {`{{time .when "15:04"}}`, `10:30`},
		"timesince":          {`{{timesince .past}}`, `1 hour`},
		"timeuntil":          {`{{timeuntil .future}}`, `2 hours`},
		"title":              {`{{title "go api"}}`, `Go Api`},
		"truncatechars":      {`{{truncatechars "abcdef" 5}}`, `ab...`},
		"truncatechars_html": {`{{truncatechars_html "<p>abcdef</p>" 5}}`, `ab...`},
		"truncatewords":      {`{{truncatewords "one two three" 2}}`, `one two ...`},
		"truncatewords_html": {`{{truncatewords_html "<p>one two three</p>" 2}}`, `one two ...`},
		"unordered_list":     {`{{unordered_list .items}}`, `<ul><li>go</li><li>api</li></ul>`},
		"upper":              {`{{upper "go"}}`, `GO`},
		"urlencode":          {`{{urlencode "a b"}}`, `a+b`},
		"urlize":             {`{{urlize "go https://example.com"}}`, `go <a href="https://example.com">https://example.com</a>`},
		"urlizetrunc":        {`{{urlizetrunc "https://example.com/long" 10}}`, `<a href="https://example.com/long">https:/...</a>`},
		"wordcount":          {`{{wordcount "one two"}}`, `2`},
		"wordwrap":           {`{{wordwrap "one two three" 7}}`, "one two\nthree"},
		"yesno":              {`{{yesno true "yes,no,maybe"}}`, `yes`},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			rendered, err := engine.RenderString(test.source, context)
			if err != nil {
				t.Fatalf("RenderString() error = %v", err)
			}
			if rendered != test.want {
				t.Fatalf("RenderString() = %q, want %q", rendered, test.want)
			}
		})
	}
}
