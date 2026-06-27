package auth

import (
	"context"
	"net/http"
	"strconv"

	"github.com/cybersaksham/gogo/sessions"
)

type userContextKey struct{}

// UserIDLoader loads users by primary key for authentication middleware.
type UserIDLoader interface {
	FindByID(ctx context.Context, id int64) (User, bool, error)
}

// ContextWithUser attaches a user to a context.
func ContextWithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

// UserFromContext returns the user attached by authentication middleware.
func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userContextKey{}).(User)
	return user, ok
}

// AnonymousUser returns the built-in anonymous principal.
func AnonymousUser() User {
	return User{AbstractUser: AbstractUser{AbstractBaseUser: AbstractBaseUser{Anonymous: true, Authenticated: false}}}
}

// AuthenticationMiddleware attaches the authenticated or anonymous user.
func AuthenticationMiddleware(loader UserIDLoader) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := AnonymousUser()
			if loader != nil {
				if session, ok := sessions.SessionFromContext(r.Context()); ok {
					user = userFromSession(r.Context(), loader, session)
				}
			}
			next.ServeHTTP(w, r.WithContext(ContextWithUser(r.Context(), user)))
		})
	}
}

func userFromSession(ctx context.Context, loader UserIDLoader, session *sessions.Session) User {
	value, ok := session.Get("user_id")
	if !ok || value == "" {
		return AnonymousUser()
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return AnonymousUser()
	}
	user, ok, err := loader.FindByID(ctx, id)
	if err != nil || !ok || !user.IsActive {
		return AnonymousUser()
	}
	user.Authenticated = true
	user.Anonymous = false
	return user
}
