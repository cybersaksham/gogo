package auth

// AdminFieldset describes one admin form section.
type AdminFieldset struct {
	Name   string
	Fields []string
}

// AdminRegistration is a framework-neutral admin adapter for later admin package integration.
type AdminRegistration struct {
	Model          string
	ListDisplay    []string
	ListFilter     []string
	SearchFields   []string
	Fieldsets      []AdminFieldset
	ReadOnlyFields []string
	Actions        []string
	Ordering       []string
}

// AdminRegistrations returns built-in auth and session admin registrations.
func AdminRegistrations() []AdminRegistration {
	return []AdminRegistration{
		{
			Model:        "auth.User",
			ListDisplay:  []string{"username", "email", "first_name", "last_name", "is_staff", "is_active"},
			ListFilter:   []string{"is_staff", "is_superuser", "is_active", "groups"},
			SearchFields: []string{"username", "first_name", "last_name", "email"},
			Fieldsets: []AdminFieldset{
				{Fields: []string{"username", "password"}},
				{Name: "Personal info", Fields: []string{"first_name", "last_name", "email"}},
				{Name: "Permissions", Fields: []string{"is_active", "is_staff", "is_superuser", "groups", "user_permissions"}},
				{Name: "Important dates", Fields: []string{"last_login", "date_joined"}},
			},
			Actions:  []string{"activate_users", "deactivate_users"},
			Ordering: []string{"username"},
		},
		{
			Model:          "auth.Group",
			ListDisplay:    []string{"name"},
			SearchFields:   []string{"name"},
			Fieldsets:      []AdminFieldset{{Name: "Group", Fields: []string{"name", "permissions"}}},
			ReadOnlyFields: nil,
			Actions:        nil,
			Ordering:       []string{"name"},
		},
		{
			Model:          "auth.Permission",
			ListDisplay:    []string{"name", "codename", "content_type"},
			ListFilter:     []string{"content_type"},
			SearchFields:   []string{"name", "codename", "content_type__app_label", "content_type__model"},
			Fieldsets:      []AdminFieldset{{Name: "Permission", Fields: []string{"name", "content_type", "codename"}}},
			ReadOnlyFields: []string{"content_type", "codename"},
			Ordering:       []string{"content_type", "codename"},
		},
		{
			Model:          "auth.ContentType",
			ListDisplay:    []string{"app_label", "model"},
			ListFilter:     []string{"app_label"},
			SearchFields:   []string{"app_label", "model"},
			Fieldsets:      []AdminFieldset{{Name: "Content type", Fields: []string{"app_label", "model"}}},
			ReadOnlyFields: []string{"app_label", "model"},
			Ordering:       []string{"app_label", "model"},
		},
		{
			Model:          "sessions.Session",
			ListDisplay:    []string{"session_key", "expire_date"},
			ListFilter:     []string{"expire_date"},
			SearchFields:   []string{"session_key"},
			Fieldsets:      []AdminFieldset{{Name: "Session", Fields: []string{"session_key", "session_data", "expire_date"}}},
			ReadOnlyFields: []string{"session_key"},
			Actions:        []string{"delete_expired_sessions"},
			Ordering:       []string{"expire_date"},
		},
	}
}
