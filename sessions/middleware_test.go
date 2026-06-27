package sessions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSessionMiddlewareAttachesSessionAndSetsSecureCookie(t *testing.T) {
	store := NewDatabaseStore("secret")
	options := CookieOptions{
		Name:     "sid",
		Path:     "/",
		Domain:   "example.com",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   3600,
	}
	handler := SessionMiddleware(store, options)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := SessionFromContext(r.Context())
		if !ok {
			t.Fatalf("session missing from context")
		}
		session.Set("flash", "saved")
		w.WriteHeader(http.StatusNoContent)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/", nil))

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %#v", cookies)
	}
	cookie := cookies[0]
	if cookie.Name != "sid" || cookie.Value == "" || cookie.Path != "/" || cookie.Domain != "example.com" {
		t.Fatalf("cookie identity = %#v", cookie)
	}
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode || cookie.MaxAge != 3600 {
		t.Fatalf("cookie security attributes = %#v", cookie)
	}
}

func TestSessionMiddlewareLoadsExistingAndExpiresMissingSession(t *testing.T) {
	store := NewDatabaseStore("secret")
	session := NewSession(time.Hour)
	session.Set("state", "loaded")
	if err := store.Save(context.Background(), session); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	handler := SessionMiddleware(store, CookieOptions{Name: "sid"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loaded, ok := SessionFromContext(r.Context())
		if !ok || loaded.GetString("state") != "loaded" {
			t.Fatalf("loaded session = %#v, %v", loaded, ok)
		}
	}))
	request := httptest.NewRequest("GET", "/", nil)
	request.AddCookie(&http.Cookie{Name: "sid", Value: session.Key})
	handler.ServeHTTP(httptest.NewRecorder(), request)

	expired := NewSession(time.Hour)
	expired.Set("state", "expired")
	expired.ExpireDate = time.Now().Add(-time.Second)
	if err := store.Save(context.Background(), expired); err != nil {
		t.Fatalf("Save(expired) error = %v", err)
	}
	expireHandler := SessionMiddleware(store, CookieOptions{Name: "sid"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loaded, ok := SessionFromContext(r.Context())
		if !ok || loaded.GetString("state") != "" {
			t.Fatalf("expired session should create empty session = %#v, %v", loaded, ok)
		}
	}))
	expiredRequest := httptest.NewRequest("GET", "/", nil)
	expiredRequest.AddCookie(&http.Cookie{Name: "sid", Value: expired.Key})
	recorder := httptest.NewRecorder()
	expireHandler.ServeHTTP(recorder, expiredRequest)
	if cookie := recorder.Result().Cookies()[0]; cookie.MaxAge != -1 {
		t.Fatalf("expired session cookie = %#v, want deletion cookie", cookie)
	}
}
