package security

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSRFSafeMethodSetsCookieAndBypassesValidation(t *testing.T) {
	protector := NewCSRFProtection(CSRFOptions{CookieName: "csrftoken", SecureCookie: true})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	protector.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d", recorder.Code)
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "csrftoken" || !cookies[0].Secure || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("cookies = %#v", cookies)
	}
}

func TestCSRFValidHeaderAndFormTokens(t *testing.T) {
	protector := NewCSRFProtection(CSRFOptions{})
	token := protector.NewToken()
	masked := protector.Mask(token)
	for _, request := range []*http.Request{
		httptest.NewRequest(http.MethodPost, "/", nil),
		httptest.NewRequest(http.MethodPost, "/", strings.NewReader("csrfmiddlewaretoken="+masked)),
	} {
		request.AddCookie(&http.Cookie{Name: protector.CookieName(), Value: token})
		request.Header.Set(protector.HeaderName(), masked)
		if request.Body != nil {
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		if err := protector.Check(request); err != nil {
			t.Fatalf("Check() error = %v", err)
		}
	}
}

func TestCSRFMissingBadOriginRefererAndTrustedOrigin(t *testing.T) {
	protector := NewCSRFProtection(CSRFOptions{TrustedOrigins: []string{"https://trusted.example.com"}})
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	if err := protector.Check(request); err == nil {
		t.Fatal("missing token should fail")
	}
	token := protector.NewToken()
	request = httptest.NewRequest(http.MethodPost, "https://app.example.com/", nil)
	request.AddCookie(&http.Cookie{Name: protector.CookieName(), Value: token})
	request.Header.Set(protector.HeaderName(), protector.Mask(token))
	request.Header.Set("Origin", "https://evil.example.com")
	if err := protector.Check(request); err == nil {
		t.Fatal("bad origin should fail")
	}
	request.Header.Set("Origin", "https://trusted.example.com")
	if err := protector.Check(request); err != nil {
		t.Fatalf("trusted origin error = %v", err)
	}
	request.Header.Del("Origin")
	request.Header.Set("Referer", "https://evil.example.com/path")
	if err := protector.Check(request); err == nil {
		t.Fatal("bad referer should fail")
	}
	request.Header.Set("Referer", "https://app.example.com/path")
	if err := protector.Check(request); err != nil {
		t.Fatalf("same-origin referer error = %v", err)
	}
}
