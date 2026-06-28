package security

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecurityMiddlewareRedirectAllowedHostsAndProxyValidation(t *testing.T) {
	middleware := SecurityMiddleware(SecurityMiddlewareOptions{SSLRedirect: true, AllowedHosts: []string{"app.example.com"}})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "http://app.example.com/path", nil)
	middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMovedPermanently || recorder.Header().Get("Location") != "https://app.example.com/path" {
		t.Fatalf("redirect status=%d location=%q", recorder.Code, recorder.Header().Get("Location"))
	}
	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "http://evil.example.com/path", nil)
	middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("bad host status=%d", recorder.Code)
	}
	proxy := SecurityMiddleware(SecurityMiddlewareOptions{SecureProxyHeaderName: "X-Forwarded-Proto", SecureProxyHeaderValue: "https"})
	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "http://app.example.com/", nil)
	request.Header.Set("X-Forwarded-Proto", "http")
	proxy(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("bad proxy status=%d", recorder.Code)
	}
}

func TestSecurityMiddlewareHeadersAndCookieDiagnostics(t *testing.T) {
	middleware := SecurityMiddleware(SecurityMiddlewareOptions{
		HSTSSeconds:             3600,
		HSTSIncludeSubdomains:   true,
		HSTSPreload:             true,
		ContentTypeNoSniff:      true,
		ReferrerPolicy:          "same-origin",
		CrossOriginOpenerPolicy: "same-origin",
		FrameOptions:            FrameDeny,
		DiagnoseSecureCookies:   true,
		SecureProxyHeaderName:   "X-Forwarded-Proto",
		SecureProxyHeaderValue:  "https",
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "http://app.example.com/", nil)
	request.Header.Set("X-Forwarded-Proto", "https")
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "sessionid", Value: "abc"})
	})).ServeHTTP(recorder, request)
	header := recorder.Header()
	if !strings.Contains(header.Get(HeaderStrictTransportSecurity), "max-age=3600") || !strings.Contains(header.Get(HeaderStrictTransportSecurity), "preload") {
		t.Fatalf("HSTS = %q", header.Get(HeaderStrictTransportSecurity))
	}
	if header.Get(HeaderContentTypeOptions) != "nosniff" || header.Get(HeaderReferrerPolicy) != "same-origin" || header.Get(HeaderCrossOriginOpenerPolicy) != "same-origin" || header.Get(HeaderXFrameOptions) != FrameDeny {
		t.Fatalf("headers = %#v", header)
	}
	if !strings.Contains(header.Get(HeaderSecureCookieDiagnostics), "sessionid") {
		t.Fatalf("cookie diagnostics = %q", header.Get(HeaderSecureCookieDiagnostics))
	}
}
