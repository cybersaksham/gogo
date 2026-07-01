package http

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestRouterHandleHTTPPathValue(t *testing.T) {
	router := NewRouter()
	err := router.HandleHTTP("item-detail", "/items/<uuid:id>/", nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = w.Write([]byte(r.PathValue("id")))
	}), "GET")
	if err != nil {
		t.Fatalf("HandleHTTP() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest("GET", "/items/123e4567-e89b-12d3-a456-426614174000/", nil))

	if recorder.Code != 200 {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if got := recorder.Body.String(); got != "123e4567-e89b-12d3-a456-426614174000" {
		t.Fatalf("PathValue body = %q, want UUID", got)
	}
}

func TestRouterHandleHTTPDoesNotBuffer(t *testing.T) {
	router := NewRouter()
	release := make(chan struct{})
	err := router.HandleHTTP("stream", "/stream/", nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		flusher, ok := w.(nethttp.Flusher)
		if !ok {
			t.Errorf("ResponseWriter does not implement http.Flusher")
			return
		}
		_, _ = w.Write([]byte("first\n"))
		flusher.Flush()
		<-release
		_, _ = w.Write([]byte("second\n"))
	}), "GET")
	if err != nil {
		t.Fatalf("HandleHTTP() error = %v", err)
	}

	server := httptest.NewServer(router)
	defer server.Close()
	defer close(release)

	response, err := server.Client().Get(server.URL + "/stream/")
	if err != nil {
		t.Fatalf("GET stream error = %v", err)
	}
	defer response.Body.Close()

	lineReady := make(chan string, 1)
	errReady := make(chan error, 1)
	go func() {
		line, err := bufio.NewReader(response.Body).ReadString('\n')
		if err != nil {
			errReady <- err
			return
		}
		lineReady <- line
	}()

	select {
	case line := <-lineReady:
		if line != "first\n" {
			t.Fatalf("first streamed line = %q, want first\\n", line)
		}
	case err := <-errReady:
		t.Fatalf("read first streamed line error = %v", err)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for flushed partial response")
	}
}

func TestRouterHandleHTTPWebSocketUpgrade(t *testing.T) {
	router := NewRouter()
	err := router.HandleHTTP("upgrade", "/upgrade/", nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		hijacker, ok := w.(nethttp.Hijacker)
		if !ok {
			nethttp.Error(w, "hijacker unavailable", nethttp.StatusInternalServerError)
			return
		}
		conn, rw, err := hijacker.Hijack()
		if err != nil {
			t.Errorf("Hijack() error = %v", err)
			return
		}
		defer conn.Close()
		_, _ = rw.WriteString("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: test\r\n\r\nupgraded")
		_ = rw.Flush()
	}), "GET")
	if err != nil {
		t.Fatalf("HandleHTTP() error = %v", err)
	}

	server := httptest.NewServer(router)
	defer server.Close()

	conn, err := net.Dial("tcp", strings.TrimPrefix(server.URL, "http://"))
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer conn.Close()
	_, _ = io.WriteString(conn, "GET /upgrade/ HTTP/1.1\r\nHost: example.test\r\nConnection: Upgrade\r\nUpgrade: test\r\n\r\n")

	response, err := nethttp.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatalf("ReadResponse() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != nethttp.StatusSwitchingProtocols {
		t.Fatalf("status = %d, want 101", response.StatusCode)
	}
	if got := response.Header.Get("Upgrade"); got != "test" {
		t.Fatalf("Upgrade = %q, want test", got)
	}
}

func TestRouterHandleHTTPMiddlewareOrder(t *testing.T) {
	router := NewRouter()
	err := router.HandleHTTP("raw", "/raw/", nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = w.Write([]byte("handler"))
	}), "GET")
	if err != nil {
		t.Fatalf("HandleHTTP() error = %v", err)
	}

	var order []string
	outer := func(next Handler) Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			order = append(order, "outer-before")
			next.ServeHTTP(w, r)
			order = append(order, "outer-after")
		})
	}
	inner := func(next Handler) Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			order = append(order, "inner-before")
			next.ServeHTTP(w, r)
			order = append(order, "inner-after")
		})
	}

	recorder := httptest.NewRecorder()
	Chain(router, outer, inner).ServeHTTP(recorder, httptest.NewRequest("GET", "/raw/", nil))

	if recorder.Code != 200 || recorder.Body.String() != "handler" {
		t.Fatalf("response = (%d, %q), want raw handler", recorder.Code, recorder.Body.String())
	}
	want := []string{"outer-before", "inner-before", "inner-after", "outer-after"}
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Fatalf("middleware order = %#v, want %#v", order, want)
	}
}

func TestRouterHandleHTTPConflictDetection(t *testing.T) {
	router := NewRouter()
	if err := router.Handle("item", "/items/", func(context.Context, *Request) Response {
		return Text(200, "framework")
	}, "GET"); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if err := router.HandleHTTP("item-raw", "/items/", nethttp.HandlerFunc(func(nethttp.ResponseWriter, *nethttp.Request) {}), "GET"); !errors.Is(err, ErrRouteConflict) {
		t.Fatalf("HandleHTTP() duplicate method/pattern error = %v, want ErrRouteConflict", err)
	}
	if err := router.HandleHTTP("item", "/other/", nethttp.HandlerFunc(func(nethttp.ResponseWriter, *nethttp.Request) {}), "GET"); !errors.Is(err, ErrRouteConflict) {
		t.Fatalf("HandleHTTP() duplicate name error = %v, want ErrRouteConflict", err)
	}
}

func TestReverseRawHandlerRoute(t *testing.T) {
	router := NewRouter()
	err := router.HandleHTTP("item-detail", "/items/<int:id>/", nethttp.HandlerFunc(func(nethttp.ResponseWriter, *nethttp.Request) {}), "GET")
	if err != nil {
		t.Fatalf("HandleHTTP() error = %v", err)
	}

	path, err := router.Reverse("item-detail", map[string]any{"id": 42})
	if err != nil {
		t.Fatalf("Reverse() error = %v", err)
	}
	if path != "/items/42/" {
		t.Fatalf("Reverse() = %q, want /items/42/", path)
	}
}
