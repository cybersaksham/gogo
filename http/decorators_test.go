package http

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMethodDecoratorsRestrictRequests(t *testing.T) {
	view := RequireGET(func(context.Context, *Request) Response {
		return Text(200, "get")
	})
	assertViewBody(t, view, NewRequest(httptest.NewRequest("GET", "/", nil)), "get")
	assertViewStatusAndBody(t, view, NewRequest(httptest.NewRequest("POST", "/", nil)), 405, "Method Not Allowed")

	post := RequirePOST(func(context.Context, *Request) Response {
		return Text(200, "post")
	})
	assertViewBody(t, post, NewRequest(httptest.NewRequest("POST", "/", nil)), "post")

	safe := RequireSafeMethods(func(context.Context, *Request) Response {
		return Text(200, "safe")
	})
	assertViewBody(t, safe, NewRequest(httptest.NewRequest("HEAD", "/", nil)), "safe")
}

func TestConditionalAndCacheDecoratorsSetHeaders(t *testing.T) {
	base := func(context.Context, *Request) Response {
		return Text(200, "cacheable")
	}
	lastModified := time.Date(2026, 6, 27, 10, 0, 0, 0, time.UTC)

	etagView := ETag(func(context.Context, *Request) (string, error) {
		return `"abc123"`, nil
	})(base)
	notModifiedRequest := httptest.NewRequest("GET", "/", nil)
	notModifiedRequest.Header.Set("If-None-Match", `"abc123"`)
	assertViewStatusAndBody(t, etagView, NewRequest(notModifiedRequest), 304, "")

	conditional := Condition(
		func(context.Context, *Request) (string, error) { return `"abc123"`, nil },
		func(context.Context, *Request) (time.Time, error) { return lastModified, nil },
	)(base)
	response := conditional(context.Background(), NewRequest(httptest.NewRequest("GET", "/", nil)))
	if got := response.Header().Get("ETag"); got != `"abc123"` {
		t.Fatalf("ETag = %q, want \"abc123\"", got)
	}
	if got := response.Header().Get("Last-Modified"); got == "" {
		t.Fatalf("Last-Modified header is empty")
	}

	decorated := NeverCache(CacheControl("public", "max-age=60")(VaryOnCookie(VaryOnHeaders("Accept-Language")(base))))
	cached := decorated(context.Background(), NewRequest(httptest.NewRequest("GET", "/", nil)))
	if got := cached.Header().Get("Vary"); got != "Accept-Language, Cookie" {
		t.Fatalf("Vary = %q, want Accept-Language, Cookie", got)
	}
	if got := cached.Header().Get("Cache-Control"); !strings.Contains(got, "no-store") {
		t.Fatalf("Cache-Control = %q, want no-store directive", got)
	}
}

func TestGZipPageCompressesEligibleResponse(t *testing.T) {
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("Accept-Encoding", "gzip")

	response := GZipPage(func(context.Context, *Request) Response {
		return Text(200, "compress me")
	})(context.Background(), NewRequest(request))

	if got := response.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", got)
	}
}

func TestFrameOptionsDecoratorsSetHeaders(t *testing.T) {
	base := func(context.Context, *Request) Response {
		return Text(200, "frame")
	}

	deny := XFrameOptionsDeny(base)(context.Background(), NewRequest(httptest.NewRequest("GET", "/", nil)))
	if got := deny.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("X-Frame-Options = %q, want DENY", got)
	}

	sameOrigin := XFrameOptionsSameOrigin(base)(context.Background(), NewRequest(httptest.NewRequest("GET", "/", nil)))
	if got := sameOrigin.Header().Get("X-Frame-Options"); got != "SAMEORIGIN" {
		t.Fatalf("X-Frame-Options = %q, want SAMEORIGIN", got)
	}

	exempt := XFrameOptionsExempt(XFrameOptionsDeny(base))(context.Background(), NewRequest(httptest.NewRequest("GET", "/", nil)))
	if got := exempt.Header().Get("X-Frame-Options"); got != "" {
		t.Fatalf("X-Frame-Options = %q, want empty", got)
	}
}

func TestCSRFDecoratorsProtectUnsafeRequestsAndSetCookie(t *testing.T) {
	base := func(context.Context, *Request) Response {
		return Text(200, "ok")
	}

	assertViewBody(t, CSRFProtect(base), NewRequest(httptest.NewRequest("GET", "/", nil)), "ok")
	assertViewStatusAndBody(t, CSRFProtect(base), NewRequest(httptest.NewRequest("POST", "/", nil)), 403, "CSRF verification failed")
	assertViewBody(t, CSRFExempt(CSRFProtect(base)), NewRequest(httptest.NewRequest("POST", "/", nil)), "ok")

	request := httptest.NewRequest("POST", "/", nil)
	request.AddCookie(&nethttp.Cookie{Name: CSRFCookieName, Value: "token"})
	request.Header.Set(CSRFHeaderName, "token")
	assertViewBody(t, RequiresCSRFToken(base), NewRequest(request), "ok")

	response := EnsureCSRFCookie(base)(context.Background(), NewRequest(httptest.NewRequest("GET", "/", nil)))
	if got := response.Header().Get("Set-Cookie"); !strings.Contains(got, CSRFCookieName+"=") {
		t.Fatalf("Set-Cookie = %q, want CSRF cookie", got)
	}
}
