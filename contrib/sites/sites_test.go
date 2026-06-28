package sites

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCurrentSiteBySettingRequestMiddlewareAdminAndChecks(t *testing.T) {
	store := NewMemoryStore([]Site{{ID: 1, Domain: "example.com", Name: "Example"}, {ID: 2, Domain: "other.com", Name: "Other"}})
	site, ok := store.ByID(1)
	if !ok || site.Name != "Example" {
		t.Fatalf("ByID() = %#v, %v", site, ok)
	}
	current, err := CurrentSite(context.Background(), httptest.NewRequest(http.MethodGet, "http://other.com/path", nil), store, Settings{SiteID: 1})
	if err != nil || current.ID != 1 {
		t.Fatalf("CurrentSite(setting) = %#v, %v", current, err)
	}
	current, err = CurrentSite(context.Background(), httptest.NewRequest(http.MethodGet, "http://other.com/path", nil), store, Settings{})
	if err != nil || current.ID != 2 {
		t.Fatalf("CurrentSite(host) = %#v, %v", current, err)
	}
	var middlewareSite Site
	Middleware(store, Settings{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		middlewareSite, _ = FromContext(r.Context())
	})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://example.com/", nil))
	if middlewareSite.ID != 1 {
		t.Fatalf("middleware site = %#v", middlewareSite)
	}
	adminModel := Admin()
	if adminModel.Model.ModelName != "Site" || len(adminModel.SearchFields) != 2 {
		t.Fatalf("admin = %#v", adminModel)
	}
	if Migration().Name != "0001_initial" {
		t.Fatalf("migration = %#v", Migration())
	}
	if results := Checks(store, Settings{SiteID: 1}); len(results) != 0 {
		t.Fatalf("checks = %#v", results)
	}
	duplicate := NewMemoryStore([]Site{{ID: 1, Domain: "example.com"}, {ID: 2, Domain: "example.com"}})
	if results := Checks(duplicate, Settings{SiteID: 99}); len(results) != 2 {
		t.Fatalf("duplicate/missing checks = %#v", results)
	}
}
