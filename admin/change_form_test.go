package admin

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

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
	if form.Mode != ChangeFormAdd || !form.Popup || !form.SaveOnTop || !form.CanDelete {
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
