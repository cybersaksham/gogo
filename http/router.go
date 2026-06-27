package http

import (
	"context"
	"fmt"
	nethttp "net/http"
	"sort"
	"strings"
)

// Router matches incoming requests to framework views.
type Router struct {
	routes              []Route
	byName              map[string]struct{}
	notFound            View
	methodNotAllowed    View
	internalServerError View
}

// Route stores route metadata for matching, docs, admin, and reversing.
type Route struct {
	Name    string
	Pattern string
	Methods []string
	View    View

	compiled Pattern
}

// NewRouter creates an empty router.
func NewRouter() *Router {
	return &Router{
		byName: make(map[string]struct{}),
		notFound: func(context.Context, *Request) Response {
			return Text(nethttp.StatusNotFound, "Not Found")
		},
		methodNotAllowed: func(context.Context, *Request) Response {
			return Text(nethttp.StatusMethodNotAllowed, "Method Not Allowed")
		},
		internalServerError: func(context.Context, *Request) Response {
			return Text(nethttp.StatusInternalServerError, "Internal Server Error")
		},
	}
}

// Handle registers one route.
func (r *Router) Handle(name, pattern string, view View, methods ...string) error {
	if _, exists := r.byName[name]; exists {
		return fmt.Errorf("%w: route name %q", ErrRouteConflict, name)
	}

	compiled, err := CompilePattern(pattern)
	if err != nil {
		return err
	}

	normalizedMethods := normalizeMethods(methods)
	for _, existing := range r.routes {
		if existing.Pattern != pattern {
			continue
		}
		for _, method := range normalizedMethods {
			if contains(existing.Methods, method) {
				return fmt.Errorf("%w: %s %s", ErrRouteConflict, method, pattern)
			}
		}
	}

	r.byName[name] = struct{}{}
	r.routes = append(r.routes, Route{
		Name:     name,
		Pattern:  pattern,
		Methods:  normalizedMethods,
		View:     view,
		compiled: compiled,
	})
	return nil
}

// Include includes a subrouter under a path prefix and namespace.
func (r *Router) Include(prefix, namespace string, subrouter *Router) error {
	prefix = strings.TrimRight(prefix, "/")
	for _, route := range subrouter.Routes() {
		name := route.Name
		if namespace != "" {
			name = namespace + ":" + name
		}
		pattern := prefix + route.Pattern
		if err := r.Handle(name, pattern, route.View, route.Methods...); err != nil {
			return err
		}
	}
	return nil
}

// Routes returns copied route metadata.
func (r *Router) Routes() []Route {
	routes := make([]Route, len(r.routes))
	copy(routes, r.routes)
	for i := range routes {
		routes[i].Methods = append([]string(nil), r.routes[i].Methods...)
	}
	return routes
}

// SetNotFound sets the 404 handler.
func (r *Router) SetNotFound(view View) {
	r.notFound = view
}

// SetMethodNotAllowed sets the 405 handler.
func (r *Router) SetMethodNotAllowed(view View) {
	r.methodNotAllowed = view
}

// SetInternalServerError sets the 500 handler.
func (r *Router) SetInternalServerError(view View) {
	r.internalServerError = view
}

// ServeHTTP serves a standard HTTP request.
func (r *Router) ServeHTTP(w nethttp.ResponseWriter, raw *nethttp.Request) {
	request := NewRequest(raw)
	response := r.resolve(raw.Context(), request)
	if err := response.Write(w); err != nil {
		_ = r.internalServerError(raw.Context(), request).Write(w)
	}
}

func (r *Router) resolve(ctx context.Context, request *Request) (response Response) {
	defer func() {
		if recover() != nil {
			response = r.internalServerError(ctx, request)
		}
	}()

	var allowed []string
	for _, route := range r.routes {
		params, ok := route.compiled.Match(request.Raw().URL.Path)
		if !ok {
			continue
		}

		if !contains(route.Methods, request.Method()) {
			allowed = append(allowed, route.Methods...)
			continue
		}

		for name, value := range params {
			request.WithPathParam(name, value)
		}
		return route.View(ctx, request)
	}

	if len(allowed) > 0 {
		allowed = uniqueSorted(allowed)
		response := r.methodNotAllowed(ctx, request)
		response.Header().Set("Allow", strings.Join(allowed, ", "))
		return response
	}

	return r.notFound(ctx, request)
}

func normalizeMethods(methods []string) []string {
	if len(methods) == 0 {
		return []string{nethttp.MethodGet}
	}

	normalized := make([]string, 0, len(methods))
	for _, method := range methods {
		normalized = append(normalized, strings.ToUpper(method))
	}
	sort.Strings(normalized)
	return normalized
}

func contains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}
	sort.Strings(unique)
	return unique
}
