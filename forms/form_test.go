package forms

import (
	"errors"
	"reflect"
	"testing"
)

func TestFormCoreBindingValidationRenderingAndGroups(t *testing.T) {
	form := NewForm(FormOptions{
		Fields: map[string]*Field{
			"title": CharField(FieldOptions{Label: "Title"}),
			"age":   IntegerField(FieldOptions{Label: "Age"}),
			"token": CharField(FieldOptions{Widget: HiddenInput()}),
		},
		FieldOrder: []string{"title", "age", "token"},
		Data: map[string]any{
			"user-title": "Go",
			"user-age":   "42",
			"user-token": "abc",
		},
		Initial: map[string]any{
			"title": "Old",
			"age":   "42",
			"token": "abc",
		},
		Prefix: "user",
		Groups: []FieldGroup{{Name: "Main", Fields: []string{"title", "age"}}},
		Clean: func(form *Form) error {
			form.CleanedData["slug"] = "go"
			return nil
		},
	})

	if !form.IsBound() {
		t.Fatal("form should be bound when data is provided")
	}
	if got := form.FieldNames(); !reflect.DeepEqual(got, []string{"title", "age", "token"}) {
		t.Fatalf("FieldNames() = %#v", got)
	}
	if !form.IsValid() {
		t.Fatalf("IsValid() = false; errors = %#v non-field = %#v", form.Errors(), form.NonFieldErrors())
	}
	if got := form.CleanedData["age"]; got != int64(42) {
		t.Fatalf("cleaned age = %#v, want int64(42)", got)
	}
	if got := form.CleanedData["slug"]; got != "go" {
		t.Fatalf("clean hook cleaned slug = %#v, want go", got)
	}
	if got := form.ChangedData(); !reflect.DeepEqual(got, []string{"title"}) {
		t.Fatalf("ChangedData() = %#v, want title only", got)
	}

	title := form.BoundField("title")
	if title.HTMLName() != "user-title" {
		t.Fatalf("HTMLName() = %q", title.HTMLName())
	}
	if got := title.Value(); got != "Go" {
		t.Fatalf("Value() = %#v", got)
	}
	if got := title.Render(); got != `<input name="user-title" type="text" value="Go">` {
		t.Fatalf("Render() = %q", got)
	}
	if got := title.LabelTag(); got != `<label for="user-title">Title</label>` {
		t.Fatalf("LabelTag() = %q", got)
	}

	if got := len(form.VisibleFields()); got != 2 {
		t.Fatalf("VisibleFields() length = %d, want 2", got)
	}
	if got := len(form.HiddenFields()); got != 1 {
		t.Fatalf("HiddenFields() length = %d, want 1", got)
	}
	groups := form.FieldGroups()
	if len(groups) != 1 || groups[0].Name != "Main" || len(groups[0].Fields) != 2 {
		t.Fatalf("FieldGroups() = %#v", groups)
	}
	if got := form.Media(); !reflect.DeepEqual(got, Media{}) {
		t.Fatalf("Media() = %#v, want empty media", got)
	}
	if got := form.Render(); got != `<div class="form-field"><label for="user-title">Title</label><input name="user-title" type="text" value="Go"></div><div class="form-field"><label for="user-age">Age</label><input name="user-age" type="number" value="42"></div><input name="user-token" type="hidden" value="abc">` {
		t.Fatalf("Render() = %q", got)
	}
}

func TestFormCoreInvalidDataAndEscapedErrorRendering(t *testing.T) {
	form := NewForm(FormOptions{
		Fields: map[string]*Field{
			"email": EmailField(FieldOptions{Required: true, Label: "Email"}),
			"age":   IntegerField(FieldOptions{}),
		},
		FieldOrder: []string{"email", "age"},
		Data: map[string]any{
			"email": "",
			"age":   "not-a-number",
		},
		Clean: func(*Form) error {
			return errors.New("cross field <bad>")
		},
	})

	if form.IsValid() {
		t.Fatal("IsValid() = true, want invalid form")
	}
	if got := form.FieldErrors("email").Messages(); !reflect.DeepEqual(got, []string{"This field is required."}) {
		t.Fatalf("email errors = %#v", got)
	}
	if got := form.FieldErrors("age").Messages(); !reflect.DeepEqual(got, []string{"enter a valid integer"}) {
		t.Fatalf("age errors = %#v", got)
	}
	if got := form.NonFieldErrors().Messages(); !reflect.DeepEqual(got, []string{"cross field <bad>"}) {
		t.Fatalf("non-field errors = %#v", got)
	}
	if got := form.FieldErrors("email").HTML(); got != `<ul class="errorlist"><li>This field is required.</li></ul>` {
		t.Fatalf("field error HTML = %q", got)
	}
	if got := form.NonFieldErrors().HTML(); got != `<ul class="errorlist nonfield"><li>cross field &lt;bad&gt;</li></ul>` {
		t.Fatalf("non-field HTML = %q", got)
	}
	if got := form.RenderErrors(); got != `<ul class="errorlist"><li>Email: This field is required.</li><li>age: enter a valid integer</li><li>cross field &lt;bad&gt;</li></ul>` {
		t.Fatalf("RenderErrors() = %q", got)
	}
}

func TestUnboundFormUsesInitialDataAndIsNotValid(t *testing.T) {
	form := NewForm(FormOptions{
		Fields: map[string]*Field{
			"title": CharField(FieldOptions{Initial: "fallback"}),
			"count": IntegerField(FieldOptions{
				Initial: int64(5),
				Widget:  NumberInput(),
			}),
		},
		FieldOrder: []string{"title", "count"},
		Initial:    map[string]any{"title": "Initial title"},
	})

	if form.IsBound() {
		t.Fatal("form should be unbound without data")
	}
	if form.IsValid() {
		t.Fatal("unbound form should not be valid")
	}
	if got := form.BoundField("title").Value(); got != "Initial title" {
		t.Fatalf("initial title = %#v", got)
	}
	if got := form.BoundField("count").Value(); got != int64(5) {
		t.Fatalf("field initial count = %#v", got)
	}
	if got := form.ChangedData(); len(got) != 0 {
		t.Fatalf("ChangedData() = %#v, want empty for unbound form", got)
	}
}
