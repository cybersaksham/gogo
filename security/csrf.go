package security

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

var (
	ErrCSRFMissingToken = errors.New("csrf token missing")
	ErrCSRFBadToken     = errors.New("csrf token invalid")
	ErrCSRFBadOrigin    = errors.New("csrf origin invalid")
	ErrCSRFBadReferer   = errors.New("csrf referer invalid")
)

type CSRFOptions struct {
	CookieName     string
	HeaderName     string
	FormFieldName  string
	TrustedOrigins []string
	SecureCookie   bool
	SameSite       http.SameSite
	Path           string
}

type CSRFProtection struct {
	options        CSRFOptions
	trustedOrigins map[string]struct{}
}

func NewCSRFProtection(options CSRFOptions) *CSRFProtection {
	if options.CookieName == "" {
		options.CookieName = "csrftoken"
	}
	if options.HeaderName == "" {
		options.HeaderName = "X-CSRFToken"
	}
	if options.FormFieldName == "" {
		options.FormFieldName = "csrfmiddlewaretoken"
	}
	if options.Path == "" {
		options.Path = "/"
	}
	if options.SameSite == 0 {
		options.SameSite = http.SameSiteLaxMode
	}
	trusted := make(map[string]struct{}, len(options.TrustedOrigins))
	for _, origin := range options.TrustedOrigins {
		trusted[strings.ToLower(strings.TrimRight(origin, "/"))] = struct{}{}
	}
	return &CSRFProtection{options: options, trustedOrigins: trusted}
}

func (p *CSRFProtection) CookieName() string { return p.options.CookieName }
func (p *CSRFProtection) HeaderName() string { return p.options.HeaderName }

func (p *CSRFProtection) NewToken() string {
	body := make([]byte, 32)
	if _, err := rand.Read(body); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(body)
}

func (p *CSRFProtection) Mask(token string) string {
	mask := make([]byte, len(token))
	if _, err := rand.Read(mask); err != nil {
		panic(err)
	}
	tokenBytes := []byte(token)
	masked := make([]byte, len(tokenBytes))
	for i := range tokenBytes {
		masked[i] = tokenBytes[i] ^ mask[i]
	}
	return base64.RawURLEncoding.EncodeToString(append(mask, masked...))
}

func (p *CSRFProtection) Unmask(masked string) (string, error) {
	body, err := base64.RawURLEncoding.DecodeString(masked)
	if err != nil || len(body)%2 != 0 {
		return "", ErrCSRFBadToken
	}
	half := len(body) / 2
	mask := body[:half]
	value := body[half:]
	token := make([]byte, len(value))
	for i := range value {
		token[i] = value[i] ^ mask[i]
	}
	return string(token), nil
}

func (p *CSRFProtection) Check(r *http.Request) error {
	if safeMethod(r.Method) {
		return nil
	}
	cookie, err := r.Cookie(p.options.CookieName)
	if err != nil || cookie.Value == "" {
		return ErrCSRFMissingToken
	}
	submitted, err := p.extractToken(r)
	if err != nil {
		return err
	}
	unmasked, err := p.Unmask(submitted)
	if err != nil {
		unmasked = submitted
	}
	if !hmac.Equal([]byte(cookie.Value), []byte(unmasked)) {
		return ErrCSRFBadToken
	}
	if r.URL.Scheme == "https" || r.TLS != nil {
		if err := p.checkSecureOrigin(r); err != nil {
			return err
		}
	}
	return nil
}

func (p *CSRFProtection) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := p.cookieValue(r)
		if token == "" {
			token = p.NewToken()
			http.SetCookie(w, &http.Cookie{Name: p.options.CookieName, Value: token, Path: p.options.Path, Secure: p.options.SecureCookie, HttpOnly: true, SameSite: p.options.SameSite})
		}
		if err := p.Check(r); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (p *CSRFProtection) extractToken(r *http.Request) (string, error) {
	if token := r.Header.Get(p.options.HeaderName); token != "" {
		return token, nil
	}
	if err := r.ParseForm(); err == nil {
		if token := r.Form.Get(p.options.FormFieldName); token != "" {
			return token, nil
		}
	}
	return "", ErrCSRFMissingToken
}

func (p *CSRFProtection) checkSecureOrigin(r *http.Request) error {
	if origin := r.Header.Get("Origin"); origin != "" {
		if p.originAllowed(origin, r.Host) {
			return nil
		}
		return ErrCSRFBadOrigin
	}
	referer := r.Header.Get("Referer")
	if referer == "" {
		return ErrCSRFBadReferer
	}
	if p.originAllowed(referer, r.Host) {
		return nil
	}
	return ErrCSRFBadReferer
}

func (p *CSRFProtection) originAllowed(raw string, host string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	origin := strings.ToLower(parsed.Scheme + "://" + parsed.Host)
	if _, ok := p.trustedOrigins[origin]; ok {
		return true
	}
	return strings.EqualFold(parsed.Host, host)
}

func (p *CSRFProtection) cookieValue(r *http.Request) string {
	cookie, err := r.Cookie(p.options.CookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func safeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}
