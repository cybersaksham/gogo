package http

import (
	"fmt"
	nethttp "net/http"
	"strings"
)

// TemporaryRedirect returns a 302 redirect response.
func TemporaryRedirect(location string) Response {
	return redirect(nethttp.StatusFound, location)
}

// PermanentRedirect returns a 301 redirect response.
func PermanentRedirect(location string) Response {
	return redirect(nethttp.StatusMovedPermanently, location)
}

// RedirectToRoute reverses a route name and returns a temporary redirect.
func RedirectToRoute(router *Router, name string, args map[string]any) (Response, error) {
	if router == nil {
		return Response{}, fmt.Errorf("%w: nil router", ErrReverse)
	}
	location, err := router.Reverse(name, args)
	if err != nil {
		return Response{}, err
	}
	return TemporaryRedirect(location), nil
}

// PermanentRedirectToRoute reverses a route name and returns a permanent redirect.
func PermanentRedirectToRoute(router *Router, name string, args map[string]any) (Response, error) {
	if router == nil {
		return Response{}, fmt.Errorf("%w: nil router", ErrReverse)
	}
	location, err := router.Reverse(name, args)
	if err != nil {
		return Response{}, err
	}
	return PermanentRedirect(location), nil
}

// BadRequest returns a safe 400 response.
func BadRequest(publicMessage string, private error) Response {
	return errorResponse(nethttp.StatusBadRequest, publicMessage, private)
}

// Forbidden returns a safe 403 response.
func Forbidden(publicMessage string, private error) Response {
	return errorResponse(nethttp.StatusForbidden, publicMessage, private)
}

// NotFound returns a safe 404 response.
func NotFound(publicMessage string, private error) Response {
	return errorResponse(nethttp.StatusNotFound, publicMessage, private)
}

// MethodNotAllowed returns a safe 405 response with the Allow header.
func MethodNotAllowed(allowed []string, private error) Response {
	response := errorResponse(nethttp.StatusMethodNotAllowed, "Method Not Allowed", private)
	response.Header().Set("Allow", strings.Join(allowed, ", "))
	return response
}

// Conflict returns a safe 409 response.
func Conflict(publicMessage string, private error) Response {
	return errorResponse(nethttp.StatusConflict, publicMessage, private)
}

// InternalServerError returns a safe 500 response.
func InternalServerError(private error) Response {
	return errorResponse(nethttp.StatusInternalServerError, "Internal Server Error", private)
}

func redirect(status int, location string) Response {
	if invalidRedirectLocation(location) {
		return BadRequest("Bad Request", fmt.Errorf("%w: %q", ErrInvalidRedirect, location))
	}
	response := Text(status, "")
	response.Header().Set("Location", location)
	return response
}

func errorResponse(status int, publicMessage string, private error) Response {
	if strings.TrimSpace(publicMessage) == "" {
		publicMessage = nethttp.StatusText(status)
	}
	response := Text(status, publicMessage)
	response.privateErr = private
	return response
}

func invalidRedirectLocation(location string) bool {
	return strings.ContainsAny(location, "\r\n")
}
