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
	if got := DateInput(WidgetConfig{Name: "date", Value: "2026-06-28"}); !strings.Contains(got, `<p class="date">`) || !strings.Contains(got, `class="vDateField"`) || !strings.Contains(got, `type="text"`) || !strings.Contains(got, `size="10"`) || !strings.Contains(got, `</p>`) {
		t.Fatalf("date input = %s", got)
	}
	if got := TimeInput(WidgetConfig{Name: "time", Value: "12:30"}); !strings.Contains(got, `<p class="time">`) || !strings.Contains(got, `class="vTimeField"`) || !strings.Contains(got, `size="8"`) || !strings.Contains(got, `</p>`) {
		t.Fatalf("time input = %s", got)
	}
	if got := DateTimeInput(WidgetConfig{Name: "dt", Value: "2026-06-28T12:30"}); !strings.Contains(got, `<p class="datetime">`) || !strings.Contains(got, `<label for="id_dt_0">Date:</label>`) || !strings.Contains(got, `name="dt_0"`) || !strings.Contains(got, `<label for="id_dt_1">Time:</label>`) || !strings.Contains(got, `name="dt_1"`) {
		t.Fatalf("datetime input = %s", got)
	}

	clearable := ClearableFileInput(WidgetConfig{Name: "avatar", Value: "current.png", InitialURL: "/media/current.png"})
	if !strings.Contains(clearable, `<p class="file-upload">`) || !strings.Contains(clearable, `Currently: <a href="/media/current.png">current.png</a>`) || !strings.Contains(clearable, `<span class="clearable-file-input">`) || !strings.Contains(clearable, `avatar-clear_id`) || !strings.Contains(clearable, `Change:`) || !strings.Contains(clearable, `</p>`) {
		t.Fatalf("clearable file = %s", clearable)
	}

	rawID := RawIDRelationWidget(WidgetConfig{Name: "author", Value: 7, RelationURL: "/admin/auth/user/"})
	if !strings.HasPrefix(rawID, `<div><input`) || !strings.Contains(rawID, `class="related-lookup"`) || !strings.Contains(rawID, `id="lookup_id_author"`) || !strings.Contains(rawID, `title="Lookup"`) || !strings.HasSuffix(rawID, `</div>`) {
		t.Fatalf("raw id = %s", rawID)
	}

	autocomplete := AutocompleteWidget(WidgetConfig{Name: "category", RelationURL: "/admin/blog/category/autocomplete/"})
	if !strings.Contains(autocomplete, `data-autocomplete-url="/admin/blog/category/autocomplete/"`) {
		t.Fatalf("autocomplete = %s", autocomplete)
	}

	filtered := FilteredSelectMultiple(WidgetConfig{Name: "groups", Value: []string{"staff"}, Choices: []WidgetChoice{{Value: "staff", Label: "Staff"}}})
	if !strings.Contains(filtered, `filtered-select-multiple`) || !strings.Contains(filtered, `selectfilter`) || !strings.Contains(filtered, `selected`) {
		t.Fatalf("filtered = %s", filtered)
	}

	readonly := ReadonlyDisplay(WidgetConfig{Name: "created_at", Value: `<now>`})
	if readonly != `<span class="readonly" data-field="created_at">&lt;now&gt;</span>` {
		t.Fatalf("readonly = %s", readonly)
	}
}

func TestAdminRelatedWidgetWrapperMatchesDjangoStructure(t *testing.T) {
	rendered := RelatedWidgetWrapper(WidgetConfig{
		Name:                     "author",
		RelatedModelName:         "user",
		RelatedModelLabel:        "user",
		AddRelatedURL:            "/admin/auth/user/add/",
		ChangeRelatedTemplateURL: "/admin/auth/user/__fk__/change/",
		DeleteRelatedTemplateURL: "/admin/auth/user/__fk__/delete/",
		ViewRelatedTemplateURL:   "/admin/auth/user/__fk__/change/",
		URLParams:                "_to_field=id&_popup=1",
		ViewURLParams:            "_to_field=id",
		CanAddRelated:            true,
		CanChangeRelated:         true,
		CanDeleteRelated:         true,
		CanViewRelated:           true,
	}, `<select name="author"></select>`)

	for _, want := range []string{
		`<div class="related-widget-wrapper" data-model-ref="user">`,
		`<select name="author"></select>`,
		`class="related-widget-wrapper-link change-related" id="change_id_author"`,
		`data-href-template="/admin/auth/user/__fk__/change/?_to_field=id&amp;_popup=1"`,
		`src="/admin/static/admin/img/icon-changelink.svg" alt="" width="24" height="24"`,
		`class="related-widget-wrapper-link add-related" id="add_id_author"`,
		`href="/admin/auth/user/add/?_to_field=id&amp;_popup=1"`,
		`src="/admin/static/admin/img/icon-addlink.svg" alt="" width="24" height="24"`,
		`class="related-widget-wrapper-link delete-related" id="delete_id_author"`,
		`src="/admin/static/admin/img/icon-deletelink.svg" alt="" width="24" height="24"`,
		`class="related-widget-wrapper-link view-related" id="view_id_author"`,
		`data-href-template="/admin/auth/user/__fk__/change/?_to_field=id"`,
		`src="/admin/static/admin/img/icon-viewlink.svg" alt="" width="24" height="24"`,
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("related widget wrapper missing %q in %s", want, rendered)
		}
	}
}
