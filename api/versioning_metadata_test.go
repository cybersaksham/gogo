package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersioningStrategiesResolveAllowedAndDefaultVersions(t *testing.T) {
	config := VersioningConfig{AllowedVersions: []string{"v1", "v2"}, DefaultVersion: "v1"}
	tests := []struct {
		name     string
		strategy VersioningStrategy
		request  *Request
		want     string
	}{
		{name: "path", strategy: URLPathVersioning{Config: config}, request: NewRequest(httptest.NewRequest(http.MethodGet, "/v2/posts/", nil)), want: "v2"},
		{name: "namespace", strategy: NamespaceVersioning{Config: config, Param: "version"}, request: NewRequest(httptest.NewRequest(http.MethodGet, "/posts/", nil)).WithPathParam("version", "v2"), want: "v2"},
		{name: "host", strategy: HostNameVersioning{Config: config}, request: hostVersionRequest("v2.api.example.com"), want: "v2"},
		{name: "query", strategy: QueryParameterVersioning{Config: config}, request: NewRequest(httptest.NewRequest(http.MethodGet, "/posts/?version=v2", nil)), want: "v2"},
		{name: "accept", strategy: AcceptHeaderVersioning{Config: config}, request: acceptVersionRequest("application/json; version=v2"), want: "v2"},
		{name: "default", strategy: QueryParameterVersioning{Config: config}, request: NewRequest(httptest.NewRequest(http.MethodGet, "/posts/", nil)), want: "v1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.strategy.ResolveVersion(test.request)
			if err != nil {
				t.Fatalf("ResolveVersion() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("version = %q, want %q", got, test.want)
			}
		})
	}

	invalid := NewRequest(httptest.NewRequest(http.MethodGet, "/posts/?version=v3", nil))
	_, err := (QueryParameterVersioning{Config: config}).ResolveVersion(invalid)
	if !errors.Is(err, ErrVersion) {
		t.Fatalf("invalid version error = %v, want ErrVersion", err)
	}

	hookRequest := NewRequest(httptest.NewRequest(http.MethodGet, "/posts/?version=v2", nil))
	if err := VersionRequest(QueryParameterVersioning{Config: config})(context.Background(), hookRequest); err != nil {
		t.Fatalf("VersionRequest() error = %v", err)
	}
	if hookRequest.Version() != "v2" {
		t.Fatalf("request version = %q, want v2", hookRequest.Version())
	}
}

func TestAPIMetadataIncludesRoutesSerializersFiltersAndPolicies(t *testing.T) {
	router := NewRouter()
	viewset := &ModelViewSet{Store: newMemoryViewSetStore()}
	viewset.RegisterAction("publish", ViewSetAction{Handler: func(context.Context, *Request) Response { return NoContent() }, Detail: true})
	if err := router.Register("posts", "post", viewset); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	serializer := NewSerializer(
		StringField("id", FieldOptions{ReadOnly: true, Label: "ID"}),
		StringField("title", FieldOptions{Required: true, HelpText: "Post title"}),
	)
	filterSet := FilterSet{ExactFields: []string{"status"}, SearchFields: []string{"title"}, OrderingFields: []string{"created_at"}}
	request := NewRequest(httptest.NewRequest(http.MethodOptions, "/v1/posts/", nil)).WithVersion("v1")

	metadata := BuildMetadata(request, router, MetadataOptions{
		Serializer:     serializer,
		FilterSet:      filterSet,
		Pagination:     PageNumberPagination{PageSize: 20},
		Authentication: []string{"session", "token"},
		Permissions:    []string{"is_authenticated"},
		Throttles:      []string{"user"},
	})
	if metadata.Version != "v1" || !metadata.BrowsableAPI {
		t.Fatalf("metadata header = %#v", metadata)
	}
	if len(metadata.Routes) != 7 || metadata.Routes[0].Name != "post-list" || metadata.Routes[6].Action != "publish" {
		t.Fatalf("routes = %#v", metadata.Routes)
	}
	if len(metadata.Serializer.Fields) != 2 || metadata.Serializer.Fields[1].Name != "title" || !metadata.Serializer.Fields[1].Required {
		t.Fatalf("serializer metadata = %#v", metadata.Serializer)
	}
	if metadata.Filters.Exact[0] != "status" || metadata.Pagination != "page_number" || metadata.Forms["create"][1].Name != "title" {
		t.Fatalf("metadata detail = %#v", metadata)
	}
	if metadata.Authentication[1] != "token" || metadata.Permissions[0] != "is_authenticated" || metadata.Throttles[0] != "user" {
		t.Fatalf("policies = %#v/%#v/%#v", metadata.Authentication, metadata.Permissions, metadata.Throttles)
	}
}

func hostVersionRequest(host string) *Request {
	raw := httptest.NewRequest(http.MethodGet, "https://"+host+"/posts/", nil)
	raw.Host = host
	return NewRequest(raw)
}

func acceptVersionRequest(accept string) *Request {
	raw := httptest.NewRequest(http.MethodGet, "/posts/", nil)
	raw.Header.Set("Accept", accept)
	return NewRequest(raw)
}
