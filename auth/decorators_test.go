package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginRequiredRedirectsHTMLAndRejectsAPI(t *testing.T) {
	protected := LoginRequired(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	html := httptest.NewRecorder()
	protected.ServeHTTP(html, httptest.NewRequest("GET", "/private?x=1", nil))
	if html.Code != http.StatusFound || html.Header().Get("Location") != "/login/?next=%2Fprivate%3Fx%3D1" {
		t.Fatalf("html response = %d location %q", html.Code, html.Header().Get("Location"))
	}

	apiRequest := httptest.NewRequest("GET", "/api/private", nil)
	apiRequest.Header.Set("Accept", "application/json")
	api := httptest.NewRecorder()
	protected.ServeHTTP(api, apiRequest)
	if api.Code != http.StatusUnauthorized {
		t.Fatalf("api status = %d, want 401", api.Code)
	}

	user := User{AbstractUser: AbstractUser{AbstractBaseUser: AbstractBaseUser{ID: 1, IsActive: true, Authenticated: true}}}
	okRequest := httptest.NewRequest("GET", "/private", nil)
	okRequest = okRequest.WithContext(ContextWithUser(okRequest.Context(), user))
	ok := httptest.NewRecorder()
	protected.ServeHTTP(ok, okRequest)
	if ok.Code != http.StatusNoContent {
		t.Fatalf("authenticated status = %d", ok.Code)
	}
}

func TestPermissionAndRoleHelpers(t *testing.T) {
	user := User{AbstractUser: AbstractUser{
		AbstractBaseUser: AbstractBaseUser{
			ID:              2,
			IsActive:        true,
			Authenticated:   true,
			UserPermissions: []Permission{{Codename: "view_post", ContentType: ContentType{AppLabel: "blog", Model: "post"}}},
		},
		IsStaff: true,
	}}
	superuser := user
	superuser.IsSuperuser = true

	cases := []struct {
		name    string
		handler http.Handler
		user    User
		want    int
		api     bool
	}{
		{name: "permission ok", handler: PermissionRequired("blog.view_post")(okHandler()), user: user, want: http.StatusNoContent},
		{name: "permission denied", handler: PermissionRequired("blog.change_post")(okHandler()), user: user, want: http.StatusForbidden, api: true},
		{name: "predicate ok", handler: UserPassesTest(func(u User) bool { return u.ID == 2 })(okHandler()), user: user, want: http.StatusNoContent},
		{name: "staff ok", handler: StaffRequired(okHandler()), user: user, want: http.StatusNoContent},
		{name: "superuser denied", handler: SuperuserRequired(okHandler()), user: user, want: http.StatusForbidden, api: true},
		{name: "superuser ok", handler: SuperuserRequired(okHandler()), user: superuser, want: http.StatusNoContent, api: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/api/protected", nil)
			if tc.api {
				request.Header.Set("Accept", "application/json")
			}
			request = request.WithContext(ContextWithUser(request.Context(), tc.user))
			recorder := httptest.NewRecorder()
			tc.handler.ServeHTTP(recorder, request)
			if recorder.Code != tc.want {
				t.Fatalf("status = %d, want %d", recorder.Code, tc.want)
			}
		})
	}
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}
