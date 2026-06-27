package admin

import (
	"reflect"
	"testing"

	gogohttp "github.com/cybersaksham/gogo/http"
	"github.com/cybersaksham/gogo/models"
)

func TestAdminURLsGenerateNamespacedRoutesAndReverse(t *testing.T) {
	site := DefaultSite()
	if err := site.ModelRegistry.RegisterMetadata(models.Metadata{AppLabel: "blog", ModelName: "Post", TableName: "blog_post"}, ModelAdmin{
		CustomURLs: []URLPattern{{Name: "stats", Path: "stats/"}},
	}); err != nil {
		t.Fatalf("RegisterMetadata() error = %v", err)
	}

	router, err := site.URLs()
	if err != nil {
		t.Fatalf("URLs() error = %v", err)
	}
	names := routeNames(router.Routes())
	wantNames := []string{
		"admin:index",
		"admin:login",
		"admin:logout",
		"admin:password_change",
		"admin:app_list",
		"admin:blog_post_changelist",
		"admin:blog_post_add",
		"admin:blog_post_change",
		"admin:blog_post_delete",
		"admin:blog_post_history",
		"admin:blog_post_autocomplete",
		"admin:blog_post_jsi18n",
		"admin:blog_post_stats",
	}
	if !reflect.DeepEqual(names, wantNames) {
		t.Fatalf("route names = %#v, want %#v", names, wantNames)
	}

	reversed, err := router.Reverse("admin:blog_post_change", map[string]any{"object_id": 42})
	if err != nil {
		t.Fatalf("Reverse(change) error = %v", err)
	}
	if reversed != "/admin/blog/post/42/change/" {
		t.Fatalf("Reverse(change) = %q", reversed)
	}
	index, err := router.Reverse("admin:index", nil)
	if err != nil || index != "/admin/" {
		t.Fatalf("Reverse(index) = %q, %v", index, err)
	}
}

func routeNames(routes []gogohttp.Route) []string {
	names := make([]string, len(routes))
	for i, route := range routes {
		names[i] = route.Name
	}
	return names
}
