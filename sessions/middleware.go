package sessions

import (
	"bytes"
	"context"
	"net/http"
	"time"
)

type sessionContextKey struct{}

// CookieOptions controls the session cookie written by middleware.
type CookieOptions struct {
	Name     string
	Path     string
	Domain   string
	HttpOnly bool
	Secure   bool
	SameSite http.SameSite
	MaxAge   int
}

// ContextWithSession attaches a session to a context.
func ContextWithSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, session)
}

// SessionFromContext returns the request session from a context.
func SessionFromContext(ctx context.Context) (*Session, bool) {
	session, ok := ctx.Value(sessionContextKey{}).(*Session)
	return session, ok
}

// SessionMiddleware loads, attaches, saves, and rotates session cookies.
func SessionMiddleware(store Store, options CookieOptions) func(http.Handler) http.Handler {
	options = normalizeCookieOptions(options)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, deleteCookie := loadRequestSession(r, store, options)
			recorder := newDeferredResponseWriter(w)
			next.ServeHTTP(recorder, r.WithContext(ContextWithSession(r.Context(), session)))
			if deleteCookie {
				expireSessionCookie(w, options)
			}
			if store == nil || !session.Modified {
				recorder.Flush()
				return
			}
			if options.MaxAge > 0 && session.ExpireDate.IsZero() {
				session.ExpireDate = time.Now().Add(time.Duration(options.MaxAge) * time.Second)
			}
			if err := store.Save(r.Context(), session); err != nil {
				recorder.Reset(http.StatusInternalServerError, []byte("failed to save session\n"))
				recorder.Header().Set("Content-Type", "text/plain; charset=utf-8")
				recorder.Flush()
				return
			}
			setSessionCookie(w, options, session)
			recorder.Flush()
		})
	}
}

func loadRequestSession(r *http.Request, store Store, options CookieOptions) (*Session, bool) {
	maxAge := time.Duration(options.MaxAge) * time.Second
	session := NewSession(maxAge)
	if store == nil {
		return session, false
	}
	cookie, err := r.Cookie(options.Name)
	if err != nil || cookie.Value == "" {
		return session, false
	}
	loaded, ok, err := store.Load(r.Context(), cookie.Value)
	if err != nil || !ok {
		return session, true
	}
	return loaded, false
}

func setSessionCookie(w http.ResponseWriter, options CookieOptions, session *Session) {
	cookie := baseCookie(options)
	cookie.Value = session.Key
	cookie.MaxAge = options.MaxAge
	if !session.ExpireDate.IsZero() {
		cookie.Expires = session.ExpireDate
	}
	http.SetCookie(w, cookie)
}

func expireSessionCookie(w http.ResponseWriter, options CookieOptions) {
	cookie := baseCookie(options)
	cookie.Value = ""
	cookie.MaxAge = -1
	cookie.Expires = time.Unix(1, 0).UTC()
	http.SetCookie(w, cookie)
}

func baseCookie(options CookieOptions) *http.Cookie {
	return &http.Cookie{
		Name:     options.Name,
		Path:     options.Path,
		Domain:   options.Domain,
		HttpOnly: options.HttpOnly,
		Secure:   options.Secure,
		SameSite: options.SameSite,
	}
}

func normalizeCookieOptions(options CookieOptions) CookieOptions {
	if options.Name == "" {
		options.Name = "gogo_sessionid"
	}
	if options.Path == "" {
		options.Path = "/"
	}
	if !options.HttpOnly {
		options.HttpOnly = true
	}
	if options.SameSite == 0 {
		options.SameSite = http.SameSiteLaxMode
	}
	return options
}

type deferredResponseWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func newDeferredResponseWriter(w http.ResponseWriter) *deferredResponseWriter {
	return &deferredResponseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (w *deferredResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *deferredResponseWriter) Write(body []byte) (int, error) {
	return w.body.Write(body)
}

func (w *deferredResponseWriter) Reset(status int, body []byte) {
	w.status = status
	w.body.Reset()
	_, _ = w.body.Write(body)
}

func (w *deferredResponseWriter) Flush() {
	w.ResponseWriter.WriteHeader(w.status)
	if w.body.Len() > 0 {
		_, _ = w.ResponseWriter.Write(w.body.Bytes())
	}
}
