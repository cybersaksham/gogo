package http

import (
	"context"
	"net"
	nethttp "net/http"
)

// Request wraps a standard HTTP request with framework metadata.
type Request struct {
	raw        *nethttp.Request
	pathParams map[string]string
	user       any
	session    any
}

// NewRequest wraps a standard HTTP request.
func NewRequest(raw *nethttp.Request) *Request {
	return &Request{
		raw:        raw,
		pathParams: make(map[string]string),
	}
}

// Raw returns the wrapped standard library request.
func (r *Request) Raw() *nethttp.Request {
	return r.raw
}

// WithPathParam returns the request after adding a path parameter.
func (r *Request) WithPathParam(name, value string) *Request {
	r.pathParams[name] = value
	return r
}

// PathParam returns one path parameter.
func (r *Request) PathParam(name string) string {
	return r.pathParams[name]
}

// QueryParam returns the first query parameter value.
func (r *Request) QueryParam(name string) string {
	return r.raw.URL.Query().Get(name)
}

// Method returns the HTTP method.
func (r *Request) Method() string {
	return r.raw.Method
}

// Host returns the request host.
func (r *Request) Host() string {
	return r.raw.Host
}

// Scheme returns the URL scheme.
func (r *Request) Scheme() string {
	if r.raw.URL.Scheme != "" {
		return r.raw.URL.Scheme
	}
	if r.raw.TLS != nil {
		return "https"
	}
	return "http"
}

// RemoteIP returns the remote IP without port.
func (r *Request) RemoteIP() string {
	host, _, err := net.SplitHostPort(r.raw.RemoteAddr)
	if err == nil {
		return host
	}
	return r.raw.RemoteAddr
}

// Context returns the request context.
func (r *Request) Context() context.Context {
	return r.raw.Context()
}

// WithUser attaches a user placeholder for the auth phase.
func (r *Request) WithUser(user any) *Request {
	r.user = user
	return r
}

// User returns the attached user placeholder.
func (r *Request) User() any {
	return r.user
}

// WithSession attaches a session placeholder for the sessions phase.
func (r *Request) WithSession(session any) *Request {
	r.session = session
	return r
}

// Session returns the attached session placeholder.
func (r *Request) Session() any {
	return r.session
}
