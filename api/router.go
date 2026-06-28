package api

import (
	"context"
	"fmt"
	nethttp "net/http"
	"sort"
	"strings"

	frameworkhttp "github.com/cybersaksham/gogo/http"
)

// RouterOption configures an API router.
type RouterOption func(*Router)

// Router stores generated API routes and resolves requests to API views.
type Router struct {
	prefix        string
	trailingSlash bool
	routes        []Route
	byName        map[string]struct{}
}

// Route stores generated API route metadata.
type Route struct {
	Name    string
	Pattern string
	Methods []string
	Action  string
	Detail  bool
	View    View

	compiled frameworkhttp.Pattern
}

// NewRouter creates an API router.
func NewRouter(options ...RouterOption) *Router {
	router := &Router{
		trailingSlash: true,
		byName:        map[string]struct{}{},
	}
	for _, option := range options {
		option(router)
	}
	return router
}

// WithAPIPrefix configures a path prefix for routes registered on the router.
func WithAPIPrefix(prefix string) RouterOption {
	return func(router *Router) {
		router.prefix = prefix
	}
}

// WithTrailingSlash enables or disables generated trailing slashes.
func WithTrailingSlash(enabled bool) RouterOption {
	return func(router *Router) {
		router.trailingSlash = enabled
	}
}

// Register registers all standard routes and custom actions for a viewset.
func (r *Router) Register(prefix, basename string, viewset *ModelViewSet) error {
	if viewset == nil {
		return fmt.Errorf("%w: nil viewset", ErrRouteConflict)
	}
	base := joinRoutePath(r.prefix, prefix)
	lookup := viewset.LookupParam
	if lookup == "" {
		lookup = "id"
	}
	standardRoutes := []struct {
		nameAction string
		action     string
		pattern    string
		methods    []string
		detail     bool
	}{
		{nameAction: "list", action: "list", pattern: r.withSlash(base), methods: []string{nethttp.MethodGet}},
		{nameAction: "create", action: "create", pattern: r.withSlash(base), methods: []string{nethttp.MethodPost}},
		{nameAction: "detail", action: "retrieve", pattern: r.withSlash(joinRoutePath(base, "<str:"+lookup+">")), methods: []string{nethttp.MethodGet}, detail: true},
		{nameAction: "update", action: "update", pattern: r.withSlash(joinRoutePath(base, "<str:"+lookup+">")), methods: []string{nethttp.MethodPut}, detail: true},
		{nameAction: "partial_update", action: "partial_update", pattern: r.withSlash(joinRoutePath(base, "<str:"+lookup+">")), methods: []string{nethttp.MethodPatch}, detail: true},
		{nameAction: "destroy", action: "destroy", pattern: r.withSlash(joinRoutePath(base, "<str:"+lookup+">")), methods: []string{nethttp.MethodDelete}, detail: true},
	}
	for _, route := range standardRoutes {
		if err := r.addRoute(routeName(basename, route.nameAction), route.pattern, route.action, route.detail, viewset.AsView(route.action), route.methods...); err != nil {
			return err
		}
	}

	for _, actionName := range sortedActionNames(viewset.Actions) {
		action := viewset.Actions[actionName]
		actionPath := strings.ReplaceAll(actionName, "_", "-")
		pattern := r.withSlash(joinRoutePath(base, actionPath))
		if action.Detail {
			pattern = r.withSlash(joinRoutePath(base, "<str:"+lookup+">", actionPath))
		}
		methods := action.Methods
		if len(methods) == 0 {
			methods = []string{nethttp.MethodGet}
		}
		if err := r.addRoute(routeName(basename, actionName), pattern, actionName, action.Detail, viewset.AsView(actionName), methods...); err != nil {
			return err
		}
	}
	return nil
}

// Handle registers one custom API route.
func (r *Router) Handle(name, pattern string, view View, methods ...string) error {
	return r.addRoute(name, r.withSlash(joinRoutePath(r.prefix, pattern)), "", false, view, methods...)
}

// Include includes another API router under a nested prefix.
func (r *Router) Include(prefix string, subrouter *Router) error {
	if subrouter == nil {
		return nil
	}
	for _, route := range subrouter.Routes() {
		pattern := joinRoutePath(r.prefix, prefix, route.Pattern)
		if err := r.addRoute(route.Name, pattern, route.Action, route.Detail, route.View, route.Methods...); err != nil {
			return err
		}
	}
	return nil
}

// Routes returns a copy of registered route metadata.
func (r *Router) Routes() []Route {
	routes := make([]Route, len(r.routes))
	copy(routes, r.routes)
	for index := range routes {
		routes[index].Methods = append([]string(nil), r.routes[index].Methods...)
	}
	return routes
}

// Resolve matches one request and runs the route view.
func (r *Router) Resolve(ctx context.Context, request *Request) Response {
	var allowed []string
	for _, route := range r.routes {
		params, ok := route.compiled.Match(request.Raw().URL.Path)
		if !ok {
			continue
		}
		if !stringIn(route.Methods, request.Method()) {
			allowed = append(allowed, route.Methods...)
			continue
		}
		for name, value := range params {
			request.WithPathParam(name, value)
		}
		return route.View(ctx, request)
	}
	if len(allowed) > 0 {
		return DefaultExceptionHandler(ctx, request, ErrMethodNotAllowed)
	}
	return DefaultExceptionHandler(ctx, request, ErrNotFound)
}

// ServeHTTP serves API routes directly as a standard HTTP handler.
func (r *Router) ServeHTTP(w nethttp.ResponseWriter, raw *nethttp.Request) {
	response := r.Resolve(raw.Context(), NewRequest(raw))
	if err := response.Write(w); err != nil {
		_ = DefaultExceptionHandler(raw.Context(), NewRequest(raw), ErrInternal).Write(w)
	}
}

// Reverse resolves a route name into a URL path.
func (r *Router) Reverse(name string, args map[string]any) (string, error) {
	proxy := frameworkhttp.NewRouter()
	for _, route := range r.routes {
		err := proxy.Handle(route.Name, route.Pattern, func(context.Context, *frameworkhttp.Request) frameworkhttp.Response {
			return frameworkhttp.Text(nethttp.StatusNoContent, "")
		}, route.Methods...)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrReverse, err)
		}
	}
	value, err := proxy.Reverse(name, args)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrReverse, err)
	}
	return value, nil
}

func (r *Router) addRoute(name, pattern, action string, detail bool, view View, methods ...string) error {
	if _, exists := r.byName[name]; exists {
		return fmt.Errorf("%w: route name %q", ErrRouteConflict, name)
	}
	normalizedMethods := normalizeAPIMethods(methods)
	for _, existing := range r.routes {
		if existing.Pattern != pattern {
			continue
		}
		for _, method := range normalizedMethods {
			if stringIn(existing.Methods, method) {
				return fmt.Errorf("%w: %s %s", ErrRouteConflict, method, pattern)
			}
		}
	}
	compiled, err := frameworkhttp.CompilePattern(pattern)
	if err != nil {
		return err
	}
	r.byName[name] = struct{}{}
	r.routes = append(r.routes, Route{
		Name:     name,
		Pattern:  pattern,
		Methods:  normalizedMethods,
		Action:   action,
		Detail:   detail,
		View:     view,
		compiled: compiled,
	})
	return nil
}

func (r *Router) withSlash(pattern string) string {
	if pattern == "" {
		pattern = "/"
	}
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}
	pattern = strings.TrimRight(pattern, "/")
	if pattern == "" {
		pattern = "/"
	}
	if r.trailingSlash && pattern != "/" {
		return pattern + "/"
	}
	return pattern
}

func joinRoutePath(parts ...string) string {
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, "/")
		if part == "" {
			continue
		}
		segments = append(segments, part)
	}
	if len(segments) == 0 {
		return "/"
	}
	return "/" + strings.Join(segments, "/")
}

func routeName(basename, action string) string {
	return strings.Trim(basename, "-") + "-" + strings.ReplaceAll(action, "_", "-")
}

func sortedActionNames(actions map[string]ViewSetAction) []string {
	names := make([]string, 0, len(actions))
	for name := range actions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func normalizeAPIMethods(methods []string) []string {
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

func stringIn(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
