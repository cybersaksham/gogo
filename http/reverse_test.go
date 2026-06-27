package http

import (
	"context"
	"errors"
	"testing"
)

func TestRouterReversesRouteByName(t *testing.T) {
	router := NewRouter()
	if err := router.Handle("article-detail", "/articles/<int:id>/", func(context.Context, *Request) Response {
		return NoContent()
	}, "GET"); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	path, err := router.Reverse("article-detail", map[string]any{"id": 42})
	if err != nil {
		t.Fatalf("Reverse() error = %v", err)
	}
	if path != "/articles/42/" {
		t.Fatalf("Reverse() = %q, want /articles/42/", path)
	}
}

func TestRouterReversesNamespacedIncludedRoute(t *testing.T) {
	subrouter := NewRouter()
	if err := subrouter.Handle("posts", "/posts/", func(context.Context, *Request) Response {
		return NoContent()
	}, "GET"); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	router := NewRouter()
	if err := router.Include("/api", "api", subrouter); err != nil {
		t.Fatalf("Include() error = %v", err)
	}

	path, err := router.Reverse("api:posts", nil)
	if err != nil {
		t.Fatalf("Reverse() error = %v", err)
	}
	if path != "/api/posts/" {
		t.Fatalf("Reverse() = %q, want /api/posts/", path)
	}
}

func TestRouterReverseRejectsMissingArgument(t *testing.T) {
	router := routerWithArticleRoute(t)

	_, err := router.Reverse("article-detail", nil)
	if !errors.Is(err, ErrReverse) {
		t.Fatalf("Reverse() error = %v, want ErrReverse", err)
	}
}

func TestRouterReverseRejectsExtraArgument(t *testing.T) {
	router := routerWithArticleRoute(t)

	_, err := router.Reverse("article-detail", map[string]any{"id": 42, "extra": "no"})
	if !errors.Is(err, ErrReverse) {
		t.Fatalf("Reverse() error = %v, want ErrReverse", err)
	}
}

func TestRouterReverseRejectsInvalidConverterValue(t *testing.T) {
	router := routerWithArticleRoute(t)

	_, err := router.Reverse("article-detail", map[string]any{"id": "not-int"})
	if !errors.Is(err, ErrReverse) {
		t.Fatalf("Reverse() error = %v, want ErrReverse", err)
	}
}

func routerWithArticleRoute(t *testing.T) *Router {
	t.Helper()

	router := NewRouter()
	if err := router.Handle("article-detail", "/articles/<int:id>/", func(context.Context, *Request) Response {
		return NoContent()
	}, "GET"); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	return router
}
