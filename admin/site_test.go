package admin

import (
	"errors"
	"net/http"
	"testing"
)

func TestDefaultAdminSite(t *testing.T) {
	site := DefaultSite()
	if site.Name != "admin" || site.Header != "Gogo administration" || site.Title != "Gogo site admin" {
		t.Fatalf("default site branding = %#v", site)
	}
	if site.IndexTitle != "Site administration" || site.URLPrefix != "/admin" {
		t.Fatalf("default site navigation = %#v", site)
	}
	if site.LoginView == nil || site.LogoutView == nil || site.PasswordChangeView == nil {
		t.Fatalf("default auth views must be configured")
	}
	if site.PermissionPolicy == nil || site.ModelRegistry == nil {
		t.Fatalf("default policy/registry must be configured")
	}
}

func TestCustomAdminSiteAndMultipleNamedSites(t *testing.T) {
	custom, err := NewSite(SiteOptions{
		Name:               "staff",
		Header:             "Staff console",
		Title:              "Staff admin",
		IndexTitle:         "Staff tools",
		URLPrefix:          "/staff-admin/",
		LoginView:          http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		LogoutView:         http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		PasswordChangeView: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		PermissionPolicy:   StaffPermissionPolicy{},
	})
	if err != nil {
		t.Fatalf("NewSite(custom) error = %v", err)
	}
	if custom.URLPrefix != "/staff-admin" {
		t.Fatalf("URLPrefix = %q, want normalized /staff-admin", custom.URLPrefix)
	}

	sites := NewSiteCollection()
	if err := sites.Register(DefaultSite()); err != nil {
		t.Fatalf("Register(default) error = %v", err)
	}
	if err := sites.Register(custom); err != nil {
		t.Fatalf("Register(custom) error = %v", err)
	}
	if got, ok := sites.Get("staff"); !ok || got.Header != "Staff console" {
		t.Fatalf("Get(staff) = %#v, %v", got, ok)
	}
	if err := sites.Register(custom); !errors.Is(err, ErrDuplicateSite) {
		t.Fatalf("Register(duplicate) error = %v, want ErrDuplicateSite", err)
	}
}

func TestAdminSiteRejectsInvalidURLPrefixes(t *testing.T) {
	for _, prefix := range []string{"", "admin", "/bad prefix", "http://admin"} {
		if _, err := NewSite(SiteOptions{Name: "bad", URLPrefix: prefix}); !errors.Is(err, ErrInvalidURLPrefix) {
			t.Fatalf("NewSite(%q) error = %v, want ErrInvalidURLPrefix", prefix, err)
		}
	}
}
