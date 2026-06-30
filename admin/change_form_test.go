package admin

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cybersaksham/gogo/auth"
)

func TestChangeFormBuildsAddAndEditMetadata(t *testing.T) {
	admin := ModelAdmin{
		Fields:             []string{"title", "slug", "author", "category", "status", "tags", "created_at"},
		Fieldsets:          []Fieldset{{Name: "Main", Fields: []string{"title", "slug"}}, {Name: "Relations", Fields: []string{"author", "category", "tags"}}},
		ReadonlyFields:     []string{"created_at"},
		PrepopulatedFields: map[string][]string{"slug": {"title"}},
		RawIDFields:        []string{"author"},
		AutocompleteFields: []string{"category"},
		RadioFields:        map[string]string{"status": "horizontal"},
		FilterHorizontal:   []string{"tags"},
		FilterVertical:     []string{"groups"},
		SaveAs:             true,
		SaveOnTop:          true,
		Hooks: ModelAdminHooks{
			HasAddPermission:    func(*http.Request, auth.User) bool { return true },
			HasChangePermission: func(*http.Request, auth.User) bool { return true },
			HasDeletePermission: func(*http.Request, auth.User) bool { return true },
		},
	}
	user := auth.User{AbstractUser: auth.AbstractUser{AbstractBaseUser: auth.AbstractBaseUser{ID: 1, IsActive: true, Authenticated: true}}}
	request := httptest.NewRequest("GET", "/admin/blog/post/add/?_popup=1", nil)

	form, err := BuildChangeForm(admin, ChangeFormInput{Mode: ChangeFormAdd, User: user, Request: request})
	if err != nil {
		t.Fatalf("BuildChangeForm(add) error = %v", err)
	}
	if form.Mode != ChangeFormAdd || !form.Popup || !form.SaveOnTop || form.CanDelete {
		t.Fatalf("form flags = %#v", form)
	}
	if !reflect.DeepEqual(form.Fieldsets, admin.Fieldsets) {
		t.Fatalf("fieldsets = %#v", form.Fieldsets)
	}
	if form.Fields["created_at"].Widget != WidgetReadonly || !form.Fields["created_at"].Readonly {
		t.Fatalf("created_at field = %#v", form.Fields["created_at"])
	}
	if form.Fields["author"].Widget != WidgetRawID || form.Fields["category"].Widget != WidgetAutocomplete || form.Fields["status"].Widget != WidgetRadio {
		t.Fatalf("relation/widgets = %#v", form.Fields)
	}
	if form.Fields["tags"].Widget != WidgetFilteredSelectMultiple || !reflect.DeepEqual(form.PrepopulatedFields["slug"], []string{"title"}) {
		t.Fatalf("many-to-many/prepopulated = %#v / %#v", form.Fields["tags"], form.PrepopulatedFields)
	}
	if !reflect.DeepEqual(form.SaveButtons, []SaveButton{SaveButtonSave, SaveButtonSaveAndContinue, SaveButtonSaveAndAddAnother, SaveButtonSaveAsNew}) {
		t.Fatalf("save buttons = %#v", form.SaveButtons)
	}

	edit, err := BuildChangeForm(admin, ChangeFormInput{Mode: ChangeFormEdit, ObjectID: "42", User: user, Request: httptest.NewRequest("GET", "/admin/blog/post/42/change/", nil)})
	if err != nil {
		t.Fatalf("BuildChangeForm(edit) error = %v", err)
	}
	if edit.ObjectID != "42" || edit.DeleteURL == "" || edit.JSI18NURL == "" {
		t.Fatalf("edit metadata = %#v", edit)
	}
}

func TestChangeFormEnforcesPermissionsAndSaveIntents(t *testing.T) {
	admin := ModelAdmin{Hooks: ModelAdminHooks{HasAddPermission: func(*http.Request, auth.User) bool { return false }}}
	user := auth.User{AbstractUser: auth.AbstractUser{AbstractBaseUser: auth.AbstractBaseUser{ID: 1, IsActive: true, Authenticated: true}}}
	_, err := BuildChangeForm(admin, ChangeFormInput{Mode: ChangeFormAdd, User: user, Request: httptest.NewRequest("GET", "/admin/add/", nil)})
	if !errors.Is(err, ErrAdminPermissionDenied) {
		t.Fatalf("BuildChangeForm(permission denied) error = %v, want ErrAdminPermissionDenied", err)
	}

	tests := []struct {
		form url.Values
		want SaveIntent
	}{
		{url.Values{"_continue": {"1"}}, SaveIntentContinue},
		{url.Values{"_addanother": {"1"}}, SaveIntentAddAnother},
		{url.Values{"_saveasnew": {"1"}}, SaveIntentSaveAsNew},
		{url.Values{}, SaveIntentSave},
	}
	for _, tc := range tests {
		if got := ResolveSaveIntent(tc.form); got != tc.want {
			t.Fatalf("ResolveSaveIntent(%v) = %s, want %s", tc.form, got, tc.want)
		}
	}
}

func TestChangeFormRelatedPopupAndJavaScriptCatalog(t *testing.T) {
	popup := RelatedPopupResponse("42", "Gogo Admin")
	if popup.Action != "change" || popup.ObjectID != "42" || popup.ObjectRepr != "Gogo Admin" {
		t.Fatalf("popup response = %#v", popup)
	}
	js := JavaScriptCatalog(map[string]string{"Save": "Save", "Delete": "Delete"})
	if js.ContentType != "application/javascript" || !strings.Contains(js.Body, `"Save":"Save"`) {
		t.Fatalf("js catalog = %#v", js)
	}
}

func TestAuthUserChangeFormUsesDjangoUserAdminWidgets(t *testing.T) {
	metadata := authMetadataByLabel()["auth.User"]
	modelAdmin := ModelAdmin{
		Model:            metadata,
		Fieldsets:        []Fieldset{{Fields: []string{"username", "password"}}, {Name: "Permissions", Fields: []string{"is_active", "is_staff", "is_superuser", "groups", "user_permissions"}}, {Name: "Important dates", Fields: []string{"last_login", "date_joined"}}},
		FilterHorizontal: []string{"groups", "user_permissions"},
		Hooks: ModelAdminHooks{
			HasChangePermission: func(*http.Request, auth.User) bool { return true },
			HasDeletePermission: func(*http.Request, auth.User) bool { return true },
		},
	}
	user := auth.User{AbstractUser: auth.AbstractUser{AbstractBaseUser: auth.AbstractBaseUser{ID: 1, IsActive: true, Authenticated: true}}}
	values := map[string]any{
		"username":         "admin",
		"password":         "pbkdf2_sha256$720000$saltvalue$hashvalue",
		"is_active":        true,
		"is_staff":         true,
		"is_superuser":     true,
		"groups":           []string{},
		"user_permissions": []string{"1"},
		"last_login":       time.Date(2026, 6, 30, 15, 43, 47, 0, time.UTC),
		"date_joined":      time.Date(2026, 6, 30, 8, 59, 42, 0, time.UTC),
	}

	form, err := BuildChangeForm(modelAdmin, ChangeFormInput{
		Mode:     ChangeFormEdit,
		ObjectID: "1",
		User:     user,
		Request:  httptest.NewRequest(http.MethodGet, "/admin/auth/user/1/change/", nil),
		Values:   values,
	})
	if err != nil {
		t.Fatalf("BuildChangeForm(auth user) error = %v", err)
	}

	checks := map[string]WidgetKind{
		"password":         WidgetPasswordHash,
		"is_active":        WidgetCheckbox,
		"is_staff":         WidgetCheckbox,
		"is_superuser":     WidgetCheckbox,
		"groups":           WidgetFilteredSelectMultiple,
		"user_permissions": WidgetFilteredSelectMultiple,
		"last_login":       WidgetDateTime,
		"date_joined":      WidgetDateTime,
	}
	for field, want := range checks {
		if got := form.Fields[field].Widget; got != want {
			t.Fatalf("%s widget = %s, want %s in %#v", field, got, want, form.Fields[field])
		}
	}

	rendered, err := RenderTemplate("change_form.html", adminPageData{
		CSRFToken:  "token",
		DeleteURL:  "/admin/auth/user/1/delete/",
		HistoryURL: "/admin/auth/user/1/history/",
		Form:       changeFormViewData(modelAdmin, form),
	}, nil)
	if err != nil {
		t.Fatalf("RenderTemplate(change_form) error = %v", err)
	}
	for _, want := range []string{
		`<label>Password:</label>`,
		`id="id_password"`,
		`<strong>algorithm</strong>: <bdi>pbkdf2_sha256</bdi>`,
		`<a class="button" href="../password/" role="button">Reset password</a>`,
		`Raw passwords are not stored`,
		`<div class="flex-container checkbox-row">`,
		`<input type="checkbox" name="is_active"`,
		`<label class="vCheckboxLabel" for="id_is_active">Active</label>`,
		`<select name="groups"`,
		`class="selectfilter filtered-select-multiple selectfilter"`,
		`data-field-name="groups"`,
		`<p class="datetime">`,
		`name="last_login_0"`,
		`name="date_joined_1"`,
		`<a role="button" href="/admin/auth/user/1/delete/" class="deletelink">Delete</a>`,
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("auth user change form missing %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, `name="password"`) {
		t.Fatalf("password hash should not render as editable input:\n%s", rendered)
	}
}
