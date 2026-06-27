# Cross Cutting Services Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement framework-wide services that Django-style applications expect: cache, email, messages, CSRF, security middleware, signals, i18n, health checks, metrics, tracing, and system checks.

**Architecture:** Each service is isolated in a public package with middleware or app hooks where needed. Services integrate with settings and app registry without introducing dependency cycles.

**Tech Stack:** Go interfaces, HTTP middleware, signed tokens, `log/slog`, OpenTelemetry-compatible hooks, cache backends, SMTP.

---

## Files

- Create: `cache/cache.go`
- Create: `cache/local.go`
- Create: `cache/redis.go`
- Create: `cache/file.go`
- Create: `cache/database.go`
- Create: `cache/memcached.go`
- Create: `cache/dummy.go`
- Create: `email/message.go`
- Create: `email/backend.go`
- Create: `email/smtp.go`
- Create: `email/console.go`
- Create: `email/file.go`
- Create: `email/memory.go`
- Create: `email/dummy.go`
- Create: `messages/message.go`
- Create: `messages/storage.go`
- Create: `messages/middleware.go`
- Create: `security/csrf.go`
- Create: `security/middleware.go`
- Create: `security/signing.go`
- Create: `security/headers.go`
- Create: `signals/signal.go`
- Create: `signals/dispatch.go`
- Create: `i18n/language.go`
- Create: `i18n/timezone.go`
- Create: `health/health.go`
- Create: `observability/metrics.go`
- Create: `observability/tracing.go`
- Create: `observability/logging.go`
- Create: `checks/checks.go`
- Modify: `internal/cli/check.go`

## Task 1: Add Cache Framework

- [ ] Create `cache/cache.go`.
- [ ] Define cache interface with get, set, add, get-or-set, delete, clear, touch, increment, decrement, get many, set many, delete many, and close.
- [ ] Create `cache/local.go`.
- [ ] Implement local in-memory cache with TTL, max entries, and cleanup.
- [ ] Create `cache/redis.go`.
- [ ] Implement Redis cache with key prefix, versioning, TTL, atomic increment, and connection health check.
- [ ] Create file, database, memcached, and dummy cache backends.
- [ ] Add tests for every cache operation, TTL expiry, key prefix, backend-specific behavior, and Redis integration gated by `GOGO_TEST_REDIS_ADDR`.
- [ ] Run `go test ./cache`.
- [ ] Commit with message `Add Cache Framework`.

## Task 2: Add Email Framework

- [ ] Create `email/message.go`.
- [ ] Create `email/backend.go`.
- [ ] Create `email/smtp.go`.
- [ ] Create `email/console.go`.
- [ ] Support plain text, HTML alternatives, attachments, headers, cc, bcc, reply-to, connection reuse, and fail-silently flag.
- [ ] Implement SMTP backend and console backend.
- [ ] Implement file, in-memory, and dummy email backends.
- [ ] Add tests for MIME rendering, attachments, SMTP connection behavior, console output, file output, memory backend capture, and dummy backend behavior.
- [ ] Run `go test ./email`.
- [ ] Commit with message `Add Email Framework`.

## Task 3: Add Messages Framework

- [ ] Create `messages/message.go`.
- [ ] Create `messages/storage.go`.
- [ ] Create `messages/middleware.go`.
- [ ] Support levels:
  - Debug
  - Info
  - Success
  - Warning
  - Error
- [ ] Support session-backed storage and cookie-backed storage.
- [ ] Attach messages to template context.
- [ ] Add tests for add, iterate, consume, level filtering, session storage, and cookie storage.
- [ ] Run `go test ./messages`.
- [ ] Commit with message `Add Messages Framework`.

## Task 4: Add Signing Utilities

- [ ] Create `security/signing.go`.
- [ ] Implement signed values with salt, timestamp, expiry, key rotation, and constant-time comparison.
- [ ] Use settings secret key.
- [ ] Add tests for valid signature, tampering, expiry, wrong salt, and rotated keys.
- [ ] Run `go test ./security`.
- [ ] Commit with message `Add Security Signing Utilities`.

## Task 5: Add CSRF Protection

- [ ] Create `security/csrf.go`.
- [ ] Implement CSRF token generation, masking, cookie storage, form token extraction, header extraction, referer/origin checks for secure requests, trusted origins, and safe-method bypass.
- [ ] Add middleware integration.
- [ ] Add tests for safe methods, valid token, missing token, bad token, origin failure, referer failure, trusted origin, and cookie attributes.
- [ ] Run `go test ./security`.
- [ ] Commit with message `Add CSRF Protection`.

## Task 6: Add Security Middleware

- [ ] Create `security/middleware.go`.
- [ ] Create `security/headers.go`.
- [ ] Implement:
  - HTTPS redirect
  - Secure proxy header validation
  - HSTS
  - Content type nosniff
  - Referrer policy
  - Cross-origin opener policy
  - X-Frame-Options clickjacking protection
  - Allowed hosts enforcement hook
  - Secure cookies enforcement diagnostics
- [ ] Add tests for every header and redirect behavior.
- [ ] Run `go test ./security ./http`.
- [ ] Commit with message `Add Security Middleware`.

## Task 7: Add Signals

- [ ] Create `signals/signal.go`.
- [ ] Create `signals/dispatch.go`.
- [ ] Support typed signals, weak-style disconnect semantics through handles, sync send, async-safe send through queue hook, receiver ordering, sender filtering, and panic recovery option.
- [ ] Add built-in signal names:
  - App ready
  - Request started
  - Request finished
  - Got request exception
  - Pre save
  - Post save
  - Pre delete
  - Post delete
  - Many-to-many changed
  - User logged in
  - User logged out
  - User login failed
  - Settings changed
  - Template rendered
  - Database connection created
  - Pre migrate
  - Post migrate
  - Check registered
- [ ] Add tests for connect, disconnect, sender filtering, ordering, panic handling, and built-in signal dispatch.
- [ ] Run `go test ./signals`.
- [ ] Commit with message `Add Signal Dispatcher`.

## Task 8: Add Internationalization And Time Zones

- [ ] Create `i18n/language.go`.
- [ ] Create `i18n/timezone.go`.
- [ ] Support active language per request, default language, accepted language parsing, translation catalog interface, lazy translation values, time zone activation, local date formatting, and UTC storage convention.
- [ ] Add tests for language negotiation, time zone activation, formatting, and request context storage.
- [ ] Run `go test ./i18n`.
- [ ] Commit with message `Add I18N And Time Zone Services`.

## Task 9: Add Health Checks

- [ ] Create `health/health.go`.
- [ ] Provide health check registry for database, cache, queue broker, result backend, storage, and custom checks.
- [ ] Support readiness and liveness HTTP handlers.
- [ ] Ensure readiness fails when required dependencies fail.
- [ ] Add tests for passing, failing, timeout, readiness, and liveness behavior.
- [ ] Run `go test ./health`.
- [ ] Commit with message `Add Health Checks`.

## Task 10: Add Observability Hooks

- [ ] Create `observability/metrics.go`.
- [ ] Create `observability/tracing.go`.
- [ ] Create `observability/logging.go`.
- [ ] Provide interfaces for counters, histograms, gauges, spans, and trace propagation.
- [ ] Provide logging configuration for handlers, levels, formatters, filters, structured fields, request IDs, SQL logging, task logging, security logging, and admin audit logging.
- [ ] Instrument HTTP requests, ORM queries, migrations, admin actions, auth login attempts, and queue tasks.
- [ ] Add no-op default implementation.
- [ ] Add tests proving instrumentation and logging hooks are called without requiring a vendor-specific backend.
- [ ] Run `go test ./observability ./http ./orm ./queue`.
- [ ] Commit with message `Add Observability Hooks`.

## Task 11: Add System Checks Framework

- [ ] Create `checks/checks.go`.
- [ ] Allow apps and packages to register checks with IDs, tags, severity, hint, object, and message.
- [ ] Implement severities:
  - Debug
  - Info
  - Warning
  - Error
  - Critical
- [ ] Add checks for settings, installed apps, models, migrations, auth, admin, static files, security, database, and queue.
- [ ] Modify `internal/cli/check.go` to run all registered checks.
- [ ] Add tests for check registration, filtering, severity thresholds, and CLI output.
- [ ] Run `go test ./checks ./internal/cli`.
- [ ] Commit with message `Add System Checks Framework`.

## Acceptance Checklist

- [ ] Cache supports local and Redis backends.
- [ ] Email supports SMTP, console, and test backends.
- [ ] Messages integrate with sessions and templates.
- [ ] CSRF and security middleware protect unsafe requests.
- [ ] Signals cover request, model, auth, and app lifecycle events.
- [ ] I18N and time zones are request-aware.
- [ ] Health checks expose readiness and liveness.
- [ ] Observability hooks are vendor-neutral.
- [ ] `gogo check` reports framework-wide diagnostics.
