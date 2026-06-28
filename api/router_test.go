package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestAPIRouterRegistersViewSetRoutesCustomActionsAndReverse(t *testing.T) {
	viewset := &ModelViewSet{Store: newMemoryViewSetStore()}
	viewset.RegisterAction("publish", ViewSetAction{
		Handler: func(context.Context, *Request) Response {
			return JSON(http.StatusOK, map[string]any{"published": true})
		},
		Detail:  true,
		Methods: []string{http.MethodPost},
	})

	router := NewRouter()
	if err := router.Register("posts", "post", viewset); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	routes := router.Routes()
	names := make([]string, 0, len(routes))
	patterns := map[string]string{}
	methods := map[string][]string{}
	for _, route := range routes {
		names = append(names, route.Name)
		patterns[route.Name] = route.Pattern
		methods[route.Name] = route.Methods
	}
	wantNames := []string{"post-list", "post-create", "post-detail", "post-update", "post-partial-update", "post-destroy", "post-publish"}
	if !reflect.DeepEqual(names, wantNames) {
		t.Fatalf("route names = %#v, want %#v", names, wantNames)
	}
	if patterns["post-detail"] != "/posts/<str:id>/" || patterns["post-publish"] != "/posts/<str:id>/publish/" {
		t.Fatalf("patterns = %#v", patterns)
	}
	if !reflect.DeepEqual(methods["post-create"], []string{http.MethodPost}) || !reflect.DeepEqual(methods["post-destroy"], []string{http.MethodDelete}) {
		t.Fatalf("methods = %#v", methods)
	}

	detailURL, err := router.Reverse("post-detail", map[string]any{"id": "42"})
	if err != nil {
		t.Fatalf("Reverse(detail) error = %v", err)
	}
	if detailURL != "/posts/42/" {
		t.Fatalf("detailURL = %q", detailURL)
	}
	customURL, err := router.Reverse("post-publish", map[string]any{"id": "42"})
	if err != nil {
		t.Fatalf("Reverse(custom) error = %v", err)
	}
	if customURL != "/posts/42/publish/" {
		t.Fatalf("customURL = %q", customURL)
	}

	request := NewRequest(httptest.NewRequest(http.MethodPost, "/posts/1/publish/", nil))
	response := router.Resolve(context.Background(), request)
	if response.status != http.StatusOK || response.body.(map[string]any)["published"] != true || request.PathParam("id") != "1" {
		t.Fatalf("custom action response=%#v id=%q", response, request.PathParam("id"))
	}
}

func TestAPIRouterSupportsNestedPrefixesTrailingSlashConfigAndInclude(t *testing.T) {
	router := NewRouter(WithAPIPrefix("api/v1"), WithTrailingSlash(false))
	if err := router.Register("orgs/<str:org_id>/posts", "org-post", &ModelViewSet{Store: newMemoryViewSetStore()}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	url, err := router.Reverse("org-post-detail", map[string]any{"org_id": "acme", "id": "9"})
	if err != nil {
		t.Fatalf("Reverse() error = %v", err)
	}
	if url != "/api/v1/orgs/acme/posts/9" {
		t.Fatalf("url = %q", url)
	}

	child := NewRouter(WithTrailingSlash(false))
	if err := child.Register("comments", "comment", &ModelViewSet{Store: newMemoryViewSetStore()}); err != nil {
		t.Fatalf("child Register() error = %v", err)
	}
	root := NewRouter(WithTrailingSlash(false))
	if err := root.Include("api/v1", child); err != nil {
		t.Fatalf("Include() error = %v", err)
	}
	includedURL, err := root.Reverse("comment-detail", map[string]any{"id": "7"})
	if err != nil {
		t.Fatalf("Reverse(included) error = %v", err)
	}
	if includedURL != "/api/v1/comments/7" {
		t.Fatalf("includedURL = %q", includedURL)
	}
}
