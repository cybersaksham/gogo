package admin

import (
	"strings"
	"testing"
)

func TestAdminWidgetsRenderEscapedAttributesAndValues(t *testing.T) {
	text := TextInput(WidgetConfig{Name: "title", Value: `<script>`, Attrs: map[string]string{"class": `field"bad`}})
	if !strings.Contains(text, `name="title"`) || !strings.Contains(text, `value="&lt;script&gt;"`) || !strings.Contains(text, `class="field&#34;bad"`) {
		t.Fatalf("text input = %s", text)
	}

	textarea := Textarea(WidgetConfig{Name: "body", Value: `<b>hello</b>`})
	if !strings.Contains(textarea, `&lt;b&gt;hello&lt;/b&gt;`) {
		t.Fatalf("textarea = %s", textarea)
	}

	if got := NumberInput(WidgetConfig{Name: "count", Value: 10}); !strings.Contains(got, `type="number"`) || !strings.Contains(got, `value="10"`) {
		t.Fatalf("number input = %s", got)
	}
	if got := Checkbox(WidgetConfig{Name: "published", Value: true}); !strings.Contains(got, `type="checkbox"`) || !strings.Contains(got, `checked`) {
		t.Fatalf("checkbox = %s", got)
	}
}

func TestAdminChoiceDateFileAndRelationWidgets(t *testing.T) {
	choices := []WidgetChoice{{Value: "draft", Label: "Draft"}, {Value: "published", Label: "Published"}}
	selectHTML := Select(WidgetConfig{Name: "status", Value: "published", Choices: choices})
	if !strings.Contains(selectHTML, `<option value="published" selected>Published</option>`) {
		t.Fatalf("select = %s", selectHTML)
	}

	multiple := SelectMultiple(WidgetConfig{Name: "tags", Value: []string{"go", "admin"}, Choices: []WidgetChoice{{Value: "go", Label: "Go"}, {Value: "admin", Label: "Admin"}, {Value: "api", Label: "API"}}})
	if !strings.Contains(multiple, `multiple`) || strings.Count(multiple, `selected`) != 2 {
		t.Fatalf("select multiple = %s", multiple)
	}

	for _, rendered := range []string{
		DateInput(WidgetConfig{Name: "date", Value: "2026-06-28"}),
		TimeInput(WidgetConfig{Name: "time", Value: "12:30"}),
		DateTimeInput(WidgetConfig{Name: "dt", Value: "2026-06-28T12:30"}),
		FileInput(WidgetConfig{Name: "avatar"}),
	} {
		if !strings.Contains(rendered, `<input`) {
			t.Fatalf("widget did not render input: %s", rendered)
		}
	}

	clearable := ClearableFileInput(WidgetConfig{Name: "avatar", Value: "current.png"})
	if !strings.Contains(clearable, `avatar-clear`) || !strings.Contains(clearable, `current.png`) {
		t.Fatalf("clearable file = %s", clearable)
	}

	rawID := RawIDRelationWidget(WidgetConfig{Name: "author", Value: 7, RelationURL: "/admin/auth/user/"})
	if !strings.Contains(rawID, `data-lookup-url="/admin/auth/user/"`) {
		t.Fatalf("raw id = %s", rawID)
	}

	autocomplete := AutocompleteWidget(WidgetConfig{Name: "category", RelationURL: "/admin/blog/category/autocomplete/"})
	if !strings.Contains(autocomplete, `data-autocomplete-url="/admin/blog/category/autocomplete/"`) {
		t.Fatalf("autocomplete = %s", autocomplete)
	}

	filtered := FilteredSelectMultiple(WidgetConfig{Name: "groups", Value: []string{"staff"}, Choices: []WidgetChoice{{Value: "staff", Label: "Staff"}}})
	if !strings.Contains(filtered, `class="filtered-select-multiple"`) || !strings.Contains(filtered, `selected`) {
		t.Fatalf("filtered = %s", filtered)
	}

	readonly := ReadonlyDisplay(WidgetConfig{Name: "created_at", Value: `<now>`})
	if readonly != `<span class="readonly" data-field="created_at">&lt;now&gt;</span>` {
		t.Fatalf("readonly = %s", readonly)
	}
}
