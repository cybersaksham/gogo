package forms

import (
	"testing"
	"time"
)

func TestWidgetsRenderEscapedHTMLSnapshots(t *testing.T) {
	tests := []struct {
		name   string
		widget Widget
		field  string
		value  any
		attrs  Attrs
		want   string
	}{
		{
			name:   "text input",
			widget: TextInput(),
			field:  "title",
			value:  `<bad>`,
			attrs:  Attrs{"class": "input", "data-note": `<x>`},
			want:   `<input class="input" data-note="&lt;x&gt;" name="title" type="text" value="&lt;bad&gt;">`,
		},
		{
			name:   "number input",
			widget: NumberInput(),
			field:  "amount",
			value:  42,
			attrs:  Attrs{"class": "number"},
			want:   `<input class="number" name="amount" type="number" value="42">`,
		},
		{
			name:   "email input",
			widget: EmailInput(),
			field:  "email",
			value:  "dev@example.com",
			attrs:  nil,
			want:   `<input name="email" type="email" value="dev@example.com">`,
		},
		{
			name:   "url input",
			widget: URLInput(),
			field:  "site",
			value:  "https://example.com?a=1&b=2",
			attrs:  nil,
			want:   `<input name="site" type="url" value="https://example.com?a=1&amp;b=2">`,
		},
		{
			name:   "password input",
			widget: PasswordInput(),
			field:  "password",
			value:  "secret",
			attrs:  nil,
			want:   `<input name="password" type="password" value="">`,
		},
		{
			name:   "hidden input",
			widget: HiddenInput(),
			field:  "token",
			value:  "abc<123>",
			attrs:  nil,
			want:   `<input name="token" type="hidden" value="abc&lt;123&gt;">`,
		},
		{
			name:   "multiple hidden input",
			widget: MultipleHiddenInput(),
			field:  "tag",
			value:  []string{"go", "api"},
			attrs:  nil,
			want:   "<input name=\"tag\" type=\"hidden\" value=\"go\">\n<input name=\"tag\" type=\"hidden\" value=\"api\">",
		},
		{
			name:   "textarea",
			widget: Textarea(),
			field:  "body",
			value:  "Go <framework>",
			attrs:  Attrs{"class": "textarea", "rows": "4"},
			want:   `<textarea class="textarea" name="body" rows="4">Go &lt;framework&gt;</textarea>`,
		},
		{
			name:   "checkbox input",
			widget: CheckboxInput(),
			field:  "active",
			value:  true,
			attrs:  Attrs{"class": "check"},
			want:   `<input checked class="check" name="active" type="checkbox" value="true">`,
		},
		{
			name:   "select",
			widget: Select([]Choice{{Value: "draft", Label: "Draft"}, {Value: "pub", Label: "Published & Live"}}),
			field:  "status",
			value:  "pub",
			attrs:  Attrs{"class": "select"},
			want:   `<select class="select" name="status"><option value="draft">Draft</option><option selected value="pub">Published &amp; Live</option></select>`,
		},
		{
			name:   "select multiple",
			widget: SelectMultiple([]Choice{{Value: "go", Label: "Go"}, {Value: "api", Label: "API"}}),
			field:  "tags",
			value:  []string{"api"},
			attrs:  Attrs{"class": "select"},
			want:   `<select class="select" multiple name="tags"><option value="go">Go</option><option selected value="api">API</option></select>`,
		},
		{
			name:   "radio select",
			widget: RadioSelect([]Choice{{Value: "admin", Label: "Admin"}, {Value: "user", Label: "User & Staff"}}),
			field:  "role",
			value:  "user",
			attrs:  nil,
			want:   "<label><input name=\"role\" type=\"radio\" value=\"admin\"> Admin</label>\n<label><input checked name=\"role\" type=\"radio\" value=\"user\"> User &amp; Staff</label>",
		},
		{
			name:   "checkbox select multiple",
			widget: CheckboxSelectMultiple([]Choice{{Value: "read", Label: "Read"}, {Value: "write", Label: "Write"}}),
			field:  "perms",
			value:  []string{"read", "write"},
			attrs:  nil,
			want:   "<label><input checked name=\"perms\" type=\"checkbox\" value=\"read\"> Read</label>\n<label><input checked name=\"perms\" type=\"checkbox\" value=\"write\"> Write</label>",
		},
		{
			name:   "date input",
			widget: DateInput(),
			field:  "starts",
			value:  time.Date(2026, 6, 28, 10, 30, 0, 0, time.UTC),
			attrs:  nil,
			want:   `<input name="starts" type="date" value="2026-06-28">`,
		},
		{
			name:   "datetime input",
			widget: DateTimeInput(),
			field:  "starts",
			value:  time.Date(2026, 6, 28, 10, 30, 0, 0, time.UTC),
			attrs:  nil,
			want:   `<input name="starts" type="datetime-local" value="2026-06-28T10:30:00">`,
		},
		{
			name:   "time input",
			widget: TimeInput(),
			field:  "starts",
			value:  time.Date(2026, 6, 28, 10, 30, 0, 0, time.UTC),
			attrs:  nil,
			want:   `<input name="starts" type="time" value="10:30:00">`,
		},
		{
			name:   "file input",
			widget: FileInput(),
			field:  "avatar",
			value:  "ignored.txt",
			attrs:  Attrs{"class": "file"},
			want:   `<input class="file" name="avatar" type="file">`,
		},
		{
			name:   "clearable file input",
			widget: ClearableFileInput(),
			field:  "avatar",
			value:  UploadedFile{Name: "avatar<1>.png"},
			attrs:  Attrs{"class": "file"},
			want:   `<span class="current-file">avatar&lt;1&gt;.png</span><label><input name="avatar-clear" type="checkbox" value="true"> Clear</label><input class="file" name="avatar" type="file">`,
		},
		{
			name:   "split date time",
			widget: SplitDateTimeWidget(),
			field:  "starts",
			value:  time.Date(2026, 6, 28, 10, 30, 0, 0, time.UTC),
			attrs:  Attrs{"class": "split"},
			want:   `<input class="split" name="starts_0" type="date" value="2026-06-28"><input class="split" name="starts_1" type="time" value="10:30:00">`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.widget.Render(test.field, test.value, test.attrs)
			if got != test.want {
				t.Fatalf("Render() = %q\nwant %q", got, test.want)
			}
		})
	}
}
