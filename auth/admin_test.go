package auth

import (
	"reflect"
	"testing"
)

func TestAuthAdminRegistrationsExposeDjangoStyleConfiguration(t *testing.T) {
	registrations := AdminRegistrations()
	if len(registrations) != 5 {
		t.Fatalf("registrations = %#v", registrations)
	}

	user, ok := adminByModel(registrations, "auth.User")
	if !ok {
		t.Fatalf("auth.User admin registration missing")
	}
	if !reflect.DeepEqual(user.ListDisplay, []string{"username", "email", "first_name", "last_name", "is_staff", "is_active"}) {
		t.Fatalf("user ListDisplay = %#v", user.ListDisplay)
	}
	if !reflect.DeepEqual(user.ListFilter, []string{"is_staff", "is_superuser", "is_active", "groups"}) {
		t.Fatalf("user ListFilter = %#v", user.ListFilter)
	}
	if !reflect.DeepEqual(user.SearchFields, []string{"username", "first_name", "last_name", "email"}) {
		t.Fatalf("user SearchFields = %#v", user.SearchFields)
	}
	if len(user.Fieldsets) != 4 || user.Fieldsets[0].Name != "" {
		t.Fatalf("user Fieldsets = %#v", user.Fieldsets)
	}
	if len(user.ReadOnlyFields) != 0 {
		t.Fatalf("user ReadOnlyFields = %#v", user.ReadOnlyFields)
	}
	if !reflect.DeepEqual(user.FilterHorizontal, []string{"groups", "user_permissions"}) {
		t.Fatalf("user FilterHorizontal = %#v", user.FilterHorizontal)
	}
	if !reflect.DeepEqual(user.Actions, []string{"activate_users", "deactivate_users"}) {
		t.Fatalf("user Actions = %#v", user.Actions)
	}

	group, ok := adminByModel(registrations, "auth.Group")
	if !ok {
		t.Fatalf("auth.Group admin registration missing")
	}
	if !reflect.DeepEqual(group.FilterHorizontal, []string{"permissions"}) {
		t.Fatalf("group FilterHorizontal = %#v", group.FilterHorizontal)
	}

	for _, model := range []string{"auth.Group", "auth.Permission", "auth.ContentType", "sessions.Session"} {
		if _, ok := adminByModel(registrations, model); !ok {
			t.Fatalf("%s admin registration missing", model)
		}
	}
}

func adminByModel(registrations []AdminRegistration, model string) (AdminRegistration, bool) {
	for _, registration := range registrations {
		if registration.Model == model {
			return registration, true
		}
	}
	return AdminRegistration{}, false
}
