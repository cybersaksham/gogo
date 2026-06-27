package http

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"time"
)

const (
	csrfExemptAttr    = "gogo.csrf.exempt"
	CSRFCookieName    = "gogo_csrf_token"
	CSRFHeaderName    = "X-CSRFToken"
	CSRFFormFieldName = "csrfmiddlewaretoken"
)

// Decorator wraps a view with extra HTTP behavior.
type Decorator func(View) View

// RequireHTTPMethods allows only the provided HTTP methods.
func RequireHTTPMethods(methods ...string) Decorator {
	allowed := normalizeMethods(methods)
	return func(view View) View {
		return func(ctx context.Context, request *Request) Response {
			if contains(allowed, strings.ToUpper(request.Method())) {
				return view(ctx, request)
			}
			return methodNotAllowed(allowed)
		}
	}
}

// RequireGET allows only GET requests.
func RequireGET(view View) View {
	return RequireHTTPMethods(nethttp.MethodGet)(view)
}

// RequirePOST allows only POST requests.
func RequirePOST(view View) View {
	return RequireHTTPMethods(nethttp.MethodPost)(view)
}

// RequireSafeMethods allows GET and HEAD requests.
func RequireSafeMethods(view View) View {
	return RequireHTTPMethods(nethttp.MethodGet, nethttp.MethodHead)(view)
}

// ETag computes an ETag and applies conditional request handling.
func ETag(fn func(context.Context, *Request) (string, error)) Decorator {
	return Condition(fn, nil)
}

// LastModified computes a Last-Modified value and applies conditional request handling.
func LastModified(fn func(context.Context, *Request) (time.Time, error)) Decorator {
	return Condition(nil, fn)
}

// Condition applies ETag and Last-Modified conditional request behavior.
func Condition(etagFn func(context.Context, *Request) (string, error), lastModifiedFn func(context.Context, *Request) (time.Time, error)) Decorator {
	return func(view View) View {
		return func(ctx context.Context, request *Request) Response {
			etag, lastModified, err := conditionalValues(ctx, request, etagFn, lastModifiedFn)
			if err != nil {
				return internalError()
			}

			if requestNotModified(request, etag, lastModified) {
				response := Response{status: nethttp.StatusNotModified, headers: make(nethttp.Header), noBody: true}
				setConditionalHeaders(&response, etag, lastModified)
				return response
			}

			response := view(ctx, request)
			setConditionalHeaders(&response, etag, lastModified)
			return response
		}
	}
}

// GZipPage compresses an eligible response when the client accepts gzip.
func GZipPage(view View) View {
	return func(ctx context.Context, request *Request) Response {
		response := view(ctx, request)
		if !acceptsGZip(request) || response.status == nethttp.StatusNoContent || response.status == nethttp.StatusNotModified {
			return response
		}

		status, headers, body, err := materialize(response)
		if err != nil {
			return internalError()
		}
		if len(body) == 0 {
			return response
		}

		var compressed bytes.Buffer
		writer := gzip.NewWriter(&compressed)
		if _, err := writer.Write(body); err != nil {
			_ = writer.Close()
			return internalError()
		}
		if err := writer.Close(); err != nil {
			return internalError()
		}

		next := bodyResponse(status, headers.Get("Content-Type"), compressed.Bytes())
		for key, values := range headers {
			if strings.EqualFold(key, "Content-Length") {
				continue
			}
			for _, value := range values {
				next.Header().Add(key, value)
			}
		}
		next.Header().Set("Content-Encoding", "gzip")
		appendVary(next.Header(), "Accept-Encoding")
		return next
	}
}

// VaryOnHeaders appends headers to the Vary response header.
func VaryOnHeaders(headers ...string) Decorator {
	return func(view View) View {
		return func(ctx context.Context, request *Request) Response {
			response := view(ctx, request)
			appendVary(ensureHeader(&response), headers...)
			return response
		}
	}
}

// VaryOnCookie appends Cookie to the Vary response header.
func VaryOnCookie(view View) View {
	return VaryOnHeaders("Cookie")(view)
}

// NeverCache sets conservative no-cache headers.
func NeverCache(view View) View {
	return func(ctx context.Context, request *Request) Response {
		response := view(ctx, request)
		header := ensureHeader(&response)
		header.Set("Cache-Control", "max-age=0, no-cache, no-store, must-revalidate, private")
		header.Set("Expires", time.Unix(0, 0).UTC().Format(nethttp.TimeFormat))
		return response
	}
}

// CacheControl sets the Cache-Control response header.
func CacheControl(directives ...string) Decorator {
	return func(view View) View {
		return func(ctx context.Context, request *Request) Response {
			response := view(ctx, request)
			ensureHeader(&response).Set("Cache-Control", strings.Join(directives, ", "))
			return response
		}
	}
}

// XFrameOptionsDeny sets X-Frame-Options to DENY.
func XFrameOptionsDeny(view View) View {
	return frameOptions(view, "DENY")
}

// XFrameOptionsSameOrigin sets X-Frame-Options to SAMEORIGIN.
func XFrameOptionsSameOrigin(view View) View {
	return frameOptions(view, "SAMEORIGIN")
}

// XFrameOptionsExempt removes X-Frame-Options from the response.
func XFrameOptionsExempt(view View) View {
	return func(ctx context.Context, request *Request) Response {
		response := view(ctx, request)
		ensureHeader(&response).Del("X-Frame-Options")
		return response
	}
}

// CSRFProtect enforces CSRF token validation on unsafe methods.
func CSRFProtect(view View) View {
	return RequiresCSRFToken(view)
}

// CSRFExempt marks a request as exempt before calling the wrapped view.
func CSRFExempt(view View) View {
	return func(ctx context.Context, request *Request) Response {
		request.setAttr(csrfExemptAttr, true)
		return view(ctx, request)
	}
}

// EnsureCSRFCookie ensures a CSRF cookie exists on the response.
func EnsureCSRFCookie(view View) View {
	return func(ctx context.Context, request *Request) Response {
		response := view(ctx, request)
		if _, err := request.Raw().Cookie(CSRFCookieName); err == nil {
			return response
		}

		token, err := newCSRFToken()
		if err != nil {
			return internalError()
		}
		cookie := &nethttp.Cookie{
			Name:     CSRFCookieName,
			Value:    token,
			Path:     "/",
			SameSite: nethttp.SameSiteLaxMode,
			Secure:   request.Scheme() == "https",
		}
		ensureHeader(&response).Add("Set-Cookie", cookie.String())
		return response
	}
}

// RequiresCSRFToken validates the CSRF token on unsafe methods.
func RequiresCSRFToken(view View) View {
	return func(ctx context.Context, request *Request) Response {
		if request.boolAttr(csrfExemptAttr) || isSafeMethod(request.Method()) {
			return view(ctx, request)
		}
		if !validCSRFToken(request) {
			return Text(nethttp.StatusForbidden, "CSRF verification failed")
		}
		return view(ctx, request)
	}
}

func methodNotAllowed(methods []string) Response {
	response := Text(nethttp.StatusMethodNotAllowed, "Method Not Allowed")
	response.Header().Set("Allow", strings.Join(methods, ", "))
	return response
}

func conditionalValues(ctx context.Context, request *Request, etagFn func(context.Context, *Request) (string, error), lastModifiedFn func(context.Context, *Request) (time.Time, error)) (string, time.Time, error) {
	var etag string
	var lastModified time.Time
	var err error

	if etagFn != nil {
		etag, err = etagFn(ctx, request)
		if err != nil {
			return "", time.Time{}, err
		}
	}
	if lastModifiedFn != nil {
		lastModified, err = lastModifiedFn(ctx, request)
		if err != nil {
			return "", time.Time{}, err
		}
	}
	return etag, lastModified, nil
}

func setConditionalHeaders(response *Response, etag string, lastModified time.Time) {
	header := ensureHeader(response)
	if etag != "" {
		header.Set("ETag", etag)
	}
	if !lastModified.IsZero() {
		header.Set("Last-Modified", lastModified.UTC().Format(nethttp.TimeFormat))
	}
}

func requestNotModified(request *Request, etag string, lastModified time.Time) bool {
	if etag != "" {
		for _, candidate := range strings.Split(request.Raw().Header.Get("If-None-Match"), ",") {
			if strings.TrimSpace(candidate) == etag {
				return true
			}
		}
	}

	if !lastModified.IsZero() {
		header := request.Raw().Header.Get("If-Modified-Since")
		if header == "" {
			return false
		}
		since, err := nethttp.ParseTime(header)
		if err == nil && !lastModified.UTC().After(since) {
			return true
		}
	}
	return false
}

func acceptsGZip(request *Request) bool {
	for _, value := range strings.Split(request.Raw().Header.Get("Accept-Encoding"), ",") {
		if strings.TrimSpace(strings.ToLower(value)) == "gzip" {
			return true
		}
	}
	return false
}

func materialize(response Response) (int, nethttp.Header, []byte, error) {
	recorder := httptest.NewRecorder()
	if err := response.Write(recorder); err != nil {
		return 0, nil, nil, err
	}
	result := recorder.Result()
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return 0, nil, nil, err
	}
	return result.StatusCode, result.Header.Clone(), body, nil
}

func appendVary(header nethttp.Header, names ...string) {
	existing := map[string]string{}
	for _, name := range strings.Split(header.Get("Vary"), ",") {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		existing[strings.ToLower(trimmed)] = trimmed
	}

	ordered := make([]string, 0, len(existing)+len(names))
	for _, name := range strings.Split(header.Get("Vary"), ",") {
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			ordered = append(ordered, trimmed)
		}
	}
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := existing[key]; ok {
			continue
		}
		existing[key] = trimmed
		ordered = append(ordered, trimmed)
	}
	header.Set("Vary", strings.Join(ordered, ", "))
}

func frameOptions(view View, value string) View {
	return func(ctx context.Context, request *Request) Response {
		response := view(ctx, request)
		ensureHeader(&response).Set("X-Frame-Options", value)
		return response
	}
}

func isSafeMethod(method string) bool {
	switch strings.ToUpper(method) {
	case nethttp.MethodGet, nethttp.MethodHead, nethttp.MethodOptions, nethttp.MethodTrace:
		return true
	default:
		return false
	}
}

func validCSRFToken(request *Request) bool {
	cookie, err := request.Raw().Cookie(CSRFCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	token := request.Raw().Header.Get(CSRFHeaderName)
	if token == "" {
		if err := request.Raw().ParseForm(); err == nil {
			token = request.Raw().FormValue(CSRFFormFieldName)
		}
	}
	return token != "" && subtle.ConstantTimeCompare([]byte(token), []byte(cookie.Value)) == 1
}

func newCSRFToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate csrf token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

func ensureHeader(response *Response) nethttp.Header {
	if response.headers == nil {
		response.headers = make(nethttp.Header)
	}
	return response.headers
}
