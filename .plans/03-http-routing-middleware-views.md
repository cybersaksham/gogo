# HTTP Routing Middleware And Views Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the web runtime: HTTP server, router, URL patterns, route reversing, request/response wrappers, middleware pipeline, handlers, redirects, errors, and request lifecycle hooks.

**Architecture:** Public package `http` wraps Go's `net/http` while preserving interoperability with standard handlers. Router and middleware are framework-owned, but applications can mount raw `net/http.Handler` values when needed.

**Tech Stack:** Go `net/http`, `context`, path matching, `httptest`, structured errors.

---

## Files

- Create: `http/server.go`
- Create: `http/router.go`
- Create: `http/pattern.go`
- Create: `http/reverse.go`
- Create: `http/request.go`
- Create: `http/response.go`
- Create: `http/view.go`
- Create: `http/generic_views.go`
- Create: `http/decorators.go`
- Create: `http/middleware.go`
- Create: `http/middleware_builtin.go`
- Create: `http/errors.go`
- Create: `http/redirect.go`
- Create: `http/static_mount.go`
- Create: `http/server_test.go`
- Create: `http/router_test.go`
- Create: `http/reverse_test.go`
- Create: `http/middleware_test.go`
- Modify: `internal/cli/runserver.go`

## Task 1: Define Request And Response Types

- [ ] Create `http/request.go`.
- [ ] Define `Request` wrapping `*net/http.Request` with helpers for:
  - Path parameters
  - Query parameters
  - Method
  - Host
  - Scheme
  - Remote IP
  - Context
  - User value attachment point for auth phase
  - Session value attachment point for sessions phase
- [ ] Create `http/response.go`.
- [ ] Define helpers:
  - `Text(status int, body string)`
  - `HTML(status int, body string)`
  - `JSON(status int, value any)`
  - `NoContent()`
  - `File(path string)`
  - `Stream(contentType string, fn func(io.Writer) error)`
- [ ] Add tests for headers, status codes, JSON encoding errors, and streaming errors.
- [ ] Run `go test ./http`.
- [ ] Commit with message `Add HTTP Request And Response Types`.

## Task 2: Define View Contract

- [ ] Create `http/view.go`.
- [ ] Define `type View func(context.Context, *Request) Response`.
- [ ] Define adapter from `net/http.Handler`.
- [ ] Define adapter to `net/http.Handler`.
- [ ] Support method-specific views with clear `405 Method Not Allowed` responses.
- [ ] Add tests for standard handler interop and method dispatch.
- [ ] Run `go test ./http`.
- [ ] Commit with message `Add HTTP View Contract`.

## Task 3: Implement URL Patterns

- [ ] Create `http/pattern.go`.
- [ ] Support pattern syntax:
  - `/articles/`
  - `/articles/<int:id>/`
  - `/users/<slug:username>/`
  - `/files/<path:key>/`
  - `/uuid/<uuid:id>/`
  - Regex-backed route patterns for advanced compatibility
  - Custom converters registered by name
- [ ] Validate duplicate parameter names.
- [ ] Percent-decode path parameters safely.
- [ ] Add tests for string, int, slug, path, UUID, custom converter, invalid converter, and duplicate names.
- [ ] Run `go test ./http`.
- [ ] Commit with message `Add URL Pattern Matching`.

## Task 4: Implement Router

- [ ] Create `http/router.go`.
- [ ] Support:
  - Named routes
  - Route namespaces
  - Included subrouters
  - Method-specific handlers
  - Trailing slash behavior configured by settings
  - Custom 404, 405, and 500 handlers
  - Route introspection for docs and admin
- [ ] Return deterministic route conflict errors.
- [ ] Add tests for matching, namespaces, includes, conflicts, not found, method not allowed, and custom handlers.
- [ ] Run `go test ./http`.
- [ ] Commit with message `Add HTTP Router`.

## Task 5: Implement Route Reversing

- [ ] Create `http/reverse.go`.
- [ ] Implement `Reverse(name string, args map[string]any) (string, error)`.
- [ ] Support namespace-qualified names such as `admin:auth_user_change`.
- [ ] Validate missing args, extra args, type conversion, and converter constraints.
- [ ] Add tests for successful reverse, missing arg, invalid arg, namespace reverse, and included router reverse.
- [ ] Run `go test ./http`.
- [ ] Commit with message `Add Route Reversing`.

## Task 6: Implement Generic Views And View Decorators

- [ ] Create `http/generic_views.go`.
- [ ] Implement Django-style generic display views:
  - Base view
  - Template view
  - Redirect view
  - Detail view
  - List view
- [ ] Implement Django-style generic editing views:
  - Form view
  - Create view
  - Update view
  - Delete view
- [ ] Implement date-based views:
  - Archive index
  - Year archive
  - Month archive
  - Week archive
  - Day archive
  - Today archive
  - Date detail
- [ ] Create `http/decorators.go`.
- [ ] Implement decorators:
  - Require HTTP methods
  - Require GET
  - Require POST
  - Require safe methods
  - Condition
  - ETag
  - Last modified
  - GZip page
  - Vary on headers
  - Vary on cookie
  - Never cache
  - Cache control
  - X-Frame-Options deny
  - X-Frame-Options sameorigin
  - X-Frame-Options exempt
  - CSRF protect
  - CSRF exempt
  - Ensure CSRF cookie
  - Requires CSRF token
- [ ] Add tests for each generic view family, method decorators, cache decorators, conditional decorators, and frame options decorators.
- [ ] Run `go test ./http`.
- [ ] Commit with message `Add Generic Views And Decorators`.

## Task 7: Implement Middleware Pipeline

- [ ] Create `http/middleware.go`.
- [ ] Define `Middleware func(Handler) Handler`.
- [ ] Preserve order from settings.
- [ ] Support before/after behavior through handler wrapping.
- [ ] Add built-in request ID middleware.
- [ ] Add built-in panic recovery middleware.
- [ ] Add built-in structured access log middleware.
- [ ] Add built-in host validation middleware using `AllowedHosts`.
- [ ] Add tests for order, short-circuiting, panic recovery, access log fields, and host validation.
- [ ] Run `go test ./http`.
- [ ] Commit with message `Add HTTP Middleware Pipeline`.

## Task 8: Implement Django-Style Built-In Middleware

- [ ] Create `http/middleware_builtin.go`.
- [ ] Implement common middleware behavior:
  - URL append slash redirect when enabled
  - URL prepend www redirect when enabled
  - Broken link reporting hook
  - User agent denylist hook
- [ ] Implement conditional GET middleware with ETag and Last-Modified support.
- [ ] Implement GZip middleware with safe content-length handling.
- [ ] Implement locale middleware hook using the `i18n` package.
- [ ] Implement cache middleware hooks for fetch-from-cache and update-cache behavior.
- [ ] Implement clickjacking middleware integration with security headers.
- [ ] Add tests for append slash, prepend www, conditional GET, GZip, locale activation, cache hit, cache update, and clickjacking headers.
- [ ] Run `go test ./http ./security ./cache ./i18n`.
- [ ] Commit with message `Add Built In HTTP Middleware`.

## Task 9: Implement Redirect And Error Responses

- [ ] Create `http/redirect.go`.
- [ ] Create `http/errors.go`.
- [ ] Support:
  - Temporary redirects
  - Permanent redirects
  - Redirect to route name
  - Bad request
  - Forbidden
  - Not found
  - Method not allowed
  - Conflict
  - Internal server error
- [ ] Include safe public messages and private log details.
- [ ] Add tests for every response status.
- [ ] Run `go test ./http`.
- [ ] Commit with message `Add HTTP Redirects And Errors`.

## Task 10: Implement Server Runtime

- [ ] Create `http/server.go`.
- [ ] Build server from settings, app registry, router, and middleware.
- [ ] Support graceful shutdown on context cancellation.
- [ ] Set secure timeouts:
  - Read header timeout
  - Read timeout
  - Write timeout
  - Idle timeout
- [ ] Add health endpoint registration hook.
- [ ] Add tests with `httptest.Server` for startup, request handling, shutdown, and timeout config.
- [ ] Run `go test ./http`.
- [ ] Commit with message `Add HTTP Server Runtime`.

## Task 11: Wire Runserver Command

- [ ] Modify `internal/cli/runserver.go`.
- [ ] Replace unavailable response with actual server boot.
- [ ] Load settings, build app registry, load routes, build middleware, and start server.
- [ ] Support `--addr`, `--settings`, and `--reload=false`.
- [ ] Ensure Ctrl+C triggers graceful shutdown.
- [ ] Add command tests with injectable server.
- [ ] Run `go test ./internal/cli ./http`.
- [ ] Commit with message `Wire Runserver To HTTP Runtime`.

## Task 12: Add Static Mount Hook

- [ ] Create `http/static_mount.go`.
- [ ] Provide development-only static and media serving hooks.
- [ ] Refuse to serve media/static in production unless explicitly enabled.
- [ ] Add tests for development allowed, production denied, path traversal blocked, and missing file behavior.
- [ ] Run `go test ./http`.
- [ ] Commit with message `Add Development Static Mounts`.

## Acceptance Checklist

- [ ] Standard `net/http` handlers can be mounted.
- [ ] Framework views can be served by standard `net/http`.
- [ ] URL reversing works for named and namespaced routes.
- [ ] Middleware order is deterministic.
- [ ] Host validation prevents unexpected host headers.
- [ ] Server timeouts and graceful shutdown are tested.
- [ ] `gogo runserver` starts a real development server.
- [ ] Generic views, view decorators, and built-in middleware are covered.
