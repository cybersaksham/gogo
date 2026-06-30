package admin

import (
	"reflect"
	"testing"
)

func TestRegisterAuthModelsRegistersDjangoStyleAuthAdmin(t *testing.T) {
	registry := NewRegistry()
	if err := RegisterAuthModels(registry); err != nil {
		t.Fatalf("RegisterAuthModels() error = %v", err)
	}

	wantModels := []string{"auth.User", "auth.Group", "auth.Permission", "auth.ContentType"}
	if got := registry.RegisteredModels(); !reflect.DeepEqual(got, wantModels) {
		t.Fatalf("registered models = %#v, want %#v", got, wantModels)
	}

	userAdmin, ok := registry.GetAdmin("auth.User")
	if !ok {
		t.Fatalf("auth.User was not registered")
	}
	if !reflect.DeepEqual(userAdmin.ListDisplay, []string{"username", "email", "first_name", "last_name", "is_staff", "is_active"}) {
		t.Fatalf("auth.User list display = %#v", userAdmin.ListDisplay)
	}
	if !reflect.DeepEqual(userAdmin.Fieldsets, []Fieldset{
		{Fields: []string{"username", "password"}},
		{Name: "Personal info", Fields: []string{"first_name", "last_name", "email"}},
		{Name: "Permissions", Fields: []string{"is_active", "is_staff", "is_superuser", "groups", "user_permissions"}},
		{Name: "Important dates", Fields: []string{"last_login", "date_joined"}},
	}) {
		t.Fatalf("auth.User fieldsets = %#v", userAdmin.Fieldsets)
	}
	if len(userAdmin.ReadonlyFields) != 0 {
		t.Fatalf("auth.User readonly fields = %#v", userAdmin.ReadonlyFields)
	}

	permissionAdmin, ok := registry.GetAdmin("auth.Permission")
	if !ok {
		t.Fatalf("auth.Permission was not registered")
	}
	if !reflect.DeepEqual(permissionAdmin.ReadonlyFields, []string{"content_type", "codename"}) {
		t.Fatalf("auth.Permission readonly fields = %#v", permissionAdmin.ReadonlyFields)
	}
}
