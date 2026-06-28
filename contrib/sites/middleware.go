package sites

import (
	"context"
	"net/http"
)

type contextKey struct{}

func Middleware(store Store, settings Settings) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			site, _ := CurrentSite(r.Context(), r, store, settings)
			ctx := context.WithValue(r.Context(), contextKey{}, site)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func FromContext(ctx context.Context) (Site, bool) {
	site, ok := ctx.Value(contextKey{}).(Site)
	return site, ok && site.ID != 0
}
