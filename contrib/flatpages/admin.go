package flatpages

import "github.com/cybersaksham/gogo/admin"

func Admin() admin.ModelAdmin {
	return admin.ModelAdmin{
		Model:          Metadata(),
		AllowUnmanaged: true,
		ListDisplay:    []string{"url", "title", "registration_required"},
		ListFilter:     []string{"enable_comments", "registration_required"},
		SearchFields:   []string{"url", "title", "content"},
		Fieldsets: []admin.Fieldset{
			{Name: "Content", Fields: []string{"url", "title", "content", "template_name"}},
			{Name: "Options", Fields: []string{"enable_comments", "registration_required", "sites"}},
		},
		FilterHorizontal: []string{"sites"},
	}
}
