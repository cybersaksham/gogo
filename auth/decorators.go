package auth

import (
	"net/http"
	"net/url"
	"strings"
)

const defaultLoginURL = "/login/"

// LoginRequired allows only authenticated active users.
func LoginRequired(next http.Handler) http.Handler {
	return UserPassesTest(func(user User) bool {
		return isLoggedIn(user)
	})(next)
}

// PermissionRequired allows users with one permission.
func PermissionRequired(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := requestUser(r)
			if !isLoggedIn(user) {
				challengeLogin(w, r)
				return
			}
			if !HasPerm(user, permission) {
				denyPermission(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// UserPassesTest allows users matching a predicate.
func UserPassesTest(test func(User) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := requestUser(r)
			if !isLoggedIn(user) {
				challengeLogin(w, r)
				return
			}
			if test == nil || !test(user) {
				denyPermission(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// StaffRequired allows active staff users.
func StaffRequired(next http.Handler) http.Handler {
	return UserPassesTest(func(user User) bool {
		return user.IsStaff
	})(next)
}

// SuperuserRequired allows active superusers.
func SuperuserRequired(next http.Handler) http.Handler {
	return UserPassesTest(func(user User) bool {
		return user.IsSuperuser
	})(next)
}

func requestUser(r *http.Request) User {
	user, ok := UserFromContext(r.Context())
	if !ok {
		return AnonymousUser()
	}
	return user
}

func isLoggedIn(user User) bool {
	return user.IsActive && user.IsAuthenticated() && !user.IsAnonymous()
}

func challengeLogin(w http.ResponseWriter, r *http.Request) {
	if wantsAPIResponse(r) {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	target := defaultLoginURL + "?next=" + url.QueryEscape(r.URL.RequestURI())
	http.Redirect(w, r, target, http.StatusFound)
}

func denyPermission(w http.ResponseWriter, r *http.Request) {
	if wantsAPIResponse(r) {
		http.Error(w, "permission denied", http.StatusForbidden)
		return
	}
	http.Error(w, "Forbidden", http.StatusForbidden)
}

func wantsAPIResponse(r *http.Request) bool {
	accept := strings.ToLower(r.Header.Get("Accept"))
	return strings.Contains(accept, "application/json") || strings.HasPrefix(r.URL.Path, "/api/")
}
