package http

import (
	"context"
	"errors"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouterMatchesRouteAndPathParams(t *testing.T) {
	router := NewRouter()
	err := router.Handle("article-detail", "/articles/<int:id>/", func(_ context.Context, request *Request) Response {
		return Text(200, request.PathParam("id"))
	}, "GET")
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest("GET", "/articles/42/", nil))

	if recorder.Code != 200 {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if recorder.Body.String() != "42" {
		t.Fatalf("body = %q, want 42", recorder.Body.String())
	}
}

func TestRouterIncludesNamespacedSubrouter(t *testing.T) {
	subrouter := NewRouter()
	if err := subrouter.Handle("posts", "/posts/", func(context.Context, *Request) Response {
		return Text(200, "posts")
	}, "GET"); err != nil {
		t.Fatalf("subrouter Handle() error = %v", err)
	}

	router := NewRouter()
	if err := router.Include("/api", "api", subrouter); err != nil {
		t.Fatalf("Include() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest("GET", "/api/posts/", nil))

	if recorder.Code != 200 || recorder.Body.String() != "posts" {
		t.Fatalf("response = (%d, %q), want (200, posts)", recorder.Code, recorder.Body.String())
	}

	routes := router.Routes()
	if len(routes) != 1 || routes[0].Name != "api:posts" {
		t.Fatalf("Routes() = %#v, want namespaced route", routes)
	}
}

func TestRouterRejectsDuplicateRouteNames(t *testing.T) {
	router := NewRouter()
	if err := router.Handle("index", "/", func(context.Context, *Request) Response {
		return Text(200, "one")
	}, "GET"); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	err := router.Handle("index", "/other/", func(context.Context, *Request) Response {
		return Text(200, "two")
	}, "GET")
	if !errors.Is(err, ErrRouteConflict) {
		t.Fatalf("Handle() error = %v, want ErrRouteConflict", err)
	}
}

func TestRouterUsesCustomNotFoundHandler(t *testing.T) {
	router := NewRouter()
	router.SetNotFound(func(context.Context, *Request) Response {
		return Text(404, "missing")
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest("GET", "/missing/", nil))

	if recorder.Code != 404 || recorder.Body.String() != "missing" {
		t.Fatalf("response = (%d, %q), want (404, missing)", recorder.Code, recorder.Body.String())
	}
}

func TestRouterUsesCustomMethodNotAllowedHandler(t *testing.T) {
	router := NewRouter()
	if err := router.Handle("index", "/", func(context.Context, *Request) Response {
		return Text(200, "ok")
	}, "GET"); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	router.SetMethodNotAllowed(func(context.Context, *Request) Response {
		return Text(nethttp.StatusMethodNotAllowed, "no method")
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest("POST", "/", nil))

	if recorder.Code != 405 || recorder.Body.String() != "no method" {
		t.Fatalf("response = (%d, %q), want (405, no method)", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Allow"); got != "GET" {
		t.Fatalf("Allow = %q, want GET", got)
	}
}

func TestRouterUsesCustomInternalServerErrorHandler(t *testing.T) {
	router := NewRouter()
	if err := router.Handle("panic", "/panic/", func(context.Context, *Request) Response {
		panic("boom")
	}, "GET"); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	router.SetInternalServerError(func(context.Context, *Request) Response {
		return Text(500, "custom 500")
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest("GET", "/panic/", nil))

	if recorder.Code != 500 || !strings.Contains(recorder.Body.String(), "custom 500") {
		t.Fatalf("response = (%d, %q), want custom 500", recorder.Code, recorder.Body.String())
	}
}
