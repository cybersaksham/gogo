package admin

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/cybersaksham/gogo/auth"
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

func TestAdminIndexRouteRendersRegisteredModels(t *testing.T) {
	site := DefaultSite()
	if err := site.ModelRegistry.RegisterMetadata(models.Metadata{AppLabel: "blog", ModelName: "Post", TableName: "blog_post"}, ModelAdmin{}); err != nil {
		t.Fatalf("RegisterMetadata() error = %v", err)
	}
	router, err := site.URLs()
	if err != nil {
		t.Fatalf("URLs() error = %v", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, staffAdminRequest(http.MethodGet, "/admin/"))
	if response.Code != http.StatusOK {
		t.Fatalf("index status = %d body=%s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("content type = %q", got)
	}
	body := response.Body.String()
	for _, want := range []string{"Gogo administration", "Site administration", "blog", "Post", "/admin/blog/post/"} {
		if !strings.Contains(body, want) {
			t.Fatalf("index body missing %q:\n%s", want, body)
		}
	}

	appResponse := httptest.NewRecorder()
	router.ServeHTTP(appResponse, staffAdminRequest(http.MethodGet, "/admin/blog/"))
	if appResponse.Code != http.StatusOK || !strings.Contains(appResponse.Body.String(), "Post") {
		t.Fatalf("app list = (%d, %q)", appResponse.Code, appResponse.Body.String())
	}
}

func TestAdminRoutesRequireActiveStaffUser(t *testing.T) {
	site := DefaultSite()
	router, err := site.URLs()
	if err != nil {
		t.Fatalf("URLs() error = %v", err)
	}

	unauthenticated := httptest.NewRecorder()
	router.ServeHTTP(unauthenticated, httptest.NewRequest(http.MethodGet, "/admin/", nil))
	if unauthenticated.Code != http.StatusFound || unauthenticated.Header().Get("Location") != "/admin/login/?next=%2Fadmin%2F" {
		t.Fatalf("unauthenticated admin response = %d location %q", unauthenticated.Code, unauthenticated.Header().Get("Location"))
	}

	login := httptest.NewRecorder()
	router.ServeHTTP(login, httptest.NewRequest(http.MethodGet, "/admin/login/", nil))
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d", login.Code)
	}

	plainUser := auth.User{AbstractUser: auth.AbstractUser{
		AbstractBaseUser: auth.AbstractBaseUser{ID: 2, IsActive: true, Authenticated: true},
		Username:         "plain",
	}}
	forbiddenRequest := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	forbiddenRequest = forbiddenRequest.WithContext(auth.ContextWithUser(forbiddenRequest.Context(), plainUser))
	forbidden := httptest.NewRecorder()
	router.ServeHTTP(forbidden, forbiddenRequest)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("non-staff admin response = %d body=%s", forbidden.Code, forbidden.Body.String())
	}
}

func staffAdminRequest(method, path string) *http.Request {
	staff := auth.User{AbstractUser: auth.AbstractUser{
		AbstractBaseUser: auth.AbstractBaseUser{ID: 1, IsActive: true, Authenticated: true},
		Username:         "staff",
		IsStaff:          true,
	}}
	request := httptest.NewRequest(method, path, nil)
	return request.WithContext(auth.ContextWithUser(request.Context(), staff))
}

func routeNames(routes []gogohttp.Route) []string {
	names := make([]string, len(routes))
	for i, route := range routes {
		names[i] = route.Name
	}
	return names
}
