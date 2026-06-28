package security

import (
	"net"
	"net/http"
	"strings"
)

type SecurityMiddlewareOptions struct {
	SSLRedirect             bool
	SSLHost                 string
	SecureProxyHeaderName   string
	SecureProxyHeaderValue  string
	HSTSSeconds             int
	HSTSIncludeSubdomains   bool
	HSTSPreload             bool
	ContentTypeNoSniff      bool
	ReferrerPolicy          string
	CrossOriginOpenerPolicy string
	FrameOptions            string
	AllowedHosts            []string
	DiagnoseSecureCookies   bool
}

func SecurityMiddleware(options SecurityMiddlewareOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !hostAllowed(r.Host, options.AllowedHosts) {
				http.Error(w, "bad host", http.StatusBadRequest)
				return
			}
			secure, proxyOK := requestSecure(r, options)
			if !proxyOK {
				http.Error(w, "invalid secure proxy header", http.StatusBadRequest)
				return
			}
			if options.SSLRedirect && !secure {
				host := r.Host
				if options.SSLHost != "" {
					host = options.SSLHost
				}
				target := "https://" + host + r.URL.RequestURI()
				http.Redirect(w, r, target, http.StatusMovedPermanently)
				return
			}
			ApplySecurityHeaders(w.Header(), options, secure)
			next.ServeHTTP(w, r)
			if options.DiagnoseSecureCookies {
				warnInsecureCookies(w.Header())
			}
		})
	}
}

func requestSecure(r *http.Request, options SecurityMiddlewareOptions) (bool, bool) {
	if options.SecureProxyHeaderName != "" {
		value := r.Header.Get(options.SecureProxyHeaderName)
		if value != "" {
			return value == options.SecureProxyHeaderValue, value == options.SecureProxyHeaderValue
		}
	}
	return r.TLS != nil || r.URL.Scheme == "https", true
}

func hostAllowed(host string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	host = stripPort(strings.ToLower(host))
	for _, allowedHost := range allowed {
		allowedHost = strings.ToLower(strings.TrimSpace(allowedHost))
		if allowedHost == "*" || allowedHost == host {
			return true
		}
		if strings.HasPrefix(allowedHost, "*.") && strings.HasSuffix(host, strings.TrimPrefix(allowedHost, "*")) {
			return true
		}
	}
	return false
}

func stripPort(host string) string {
	if withoutPort, _, err := net.SplitHostPort(host); err == nil {
		return withoutPort
	}
	return host
}

func warnInsecureCookies(header http.Header) {
	var names []string
	for _, cookie := range header.Values("Set-Cookie") {
		if !strings.Contains(strings.ToLower(cookie), "secure") {
			name := strings.SplitN(cookie, "=", 2)[0]
			names = append(names, name)
		}
	}
	if len(names) > 0 {
		header.Set(HeaderSecureCookieDiagnostics, "missing Secure on "+strings.Join(names, ","))
	}
}
