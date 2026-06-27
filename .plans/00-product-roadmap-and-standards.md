# Gogo Product Roadmap And Standards Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Define the end-to-end implementation sequence, engineering standards, and parity checklist for a Go framework that provides Django-style apps, models, migrations, ORM, admin, auth, APIs, and Celery-style queues.

**Architecture:** Gogo is built as a layered Go module with stable public packages and private internals. Client projects depend only on public packages while CLI commands, schema diffing, code generation, and runtime reflection live under `internal/`.

**Tech Stack:** Go, standard `net/http`, `database/sql`, PostgreSQL first, SQLite for tests and local quickstart, Redis and RabbitMQ for queue brokers, Go templates for server-rendered admin, structured logging through `log/slog`.

---

## Product Sequence

Implement phases in this exact order:

1. `01-foundation-cli-config.md`
2. `02-app-project-lifecycle.md`
3. `03-http-routing-middleware-views.md`
4. `04-models-fields-validation.md`
5. `05-orm-query-engine.md`
6. `06-migrations-schema-management.md`
7. `07-auth-permissions-sessions.md`
8. `08-admin-panel.md`
9. `09-api-serializers-openapi.md`
10. `10-forms-templates-static-files.md`
11. `11-queue-workers-beat-canvas.md`
12. `12-cross-cutting-services.md`
13. `13-django-contrib-compatibility.md`
14. `14-testing-docs-examples.md`
15. `15-client-project-template-and-generators.md`
16. `16-production-hardening-release.md`

Each phase must leave the repository buildable and testable.

## Global Package Boundaries

- Public framework packages:
  - `app`
  - `conf`
  - `http`
  - `models`
  - `orm`
  - `migrations`
  - `admin`
  - `auth`
  - `api`
  - `forms`
  - `templates`
  - `static`
  - `queue`
  - `cache`
  - `email`
  - `files`
  - `sessions`
  - `security`
  - `signals`
  - `testing`
- Private implementation packages:
  - `internal/cli`
  - `internal/codegen`
  - `internal/inspect`
  - `internal/schema`
  - `internal/version`
  - `internal/testutil`
- Client projects must never import from `internal/`.
- Public packages must expose stable interfaces and keep runtime-specific details private.
- All public APIs that can block must accept `context.Context`.
- All public APIs that can fail must return typed errors or wrapped errors with stable predicates.

## Core Parity Checklist

- Django-style project and app structure.
- App registry and installed apps.
- Settings loader with environment validation.
- Management CLI with project, app, server, migration, admin, queue, test, and data commands.
- HTTP router with namespacing, reversing, middleware, views, and normalized errors.
- Model system with all Django field families.
- Model metadata, indexes, constraints, validation, choices, defaults, and hooks.
- ORM with lazy querysets, filters, excludes, ordering, slicing, joins, prefetching, expressions, aggregations, transactions, raw SQL, and multi-database support.
- Migration autodetection, operations, graph, executor, recorder, rollback, fake apply, squashing, raw SQL, and data migrations.
- Built-in auth with inheritable users, groups, permissions, sessions, password tooling, and admin integration.
- Admin panel with model registration, list/detail/add/change/delete/history screens, filters, search, actions, inlines, custom widgets, object permissions, templates, and static assets.
- API layer with serializers, validation, routers, permissions, pagination, filtering, throttling, uploads, and OpenAPI.
- Forms, templates, static files, media files, and storage backends.
- Queue system with task discovery, workers, brokers, result backends, retries, timeouts, routing, rate limits, ETA/countdown, periodic beat, chains, groups, chords, maps, chunks, callbacks, errbacks, and monitoring.
- Cache, email, files, sessions, security, signals, health checks, metrics, tracing, fixtures, and seed data.
- Django contrib-style packages for sites, redirects, flatpages, sitemaps, feeds, humanize filters, Postgres helpers, and GIS operations.
- Test toolkit, docs, examples, release automation, and production hardening.

## Engineering Rules

- [ ] Keep every phase on the current Git branch.
- [ ] Commit after each completed task with an imperative message such as `Add CLI Root Command`.
- [ ] Keep `.gitignore` grouped with comments for each category.
- [ ] Keep `.env.example` synced with any required environment variable.
- [ ] Required environment variables must fail fast during boot when missing.
- [ ] Do not commit secrets, machine-local paths, personal tokens, generated coverage dumps, local databases, or uploaded media.
- [ ] Keep generated files deterministic.
- [ ] Keep public errors documented and tested.
- [ ] Prefer standard library implementations unless a dependency materially reduces risk.
- [ ] Add dependencies through `go get`, then explain why the dependency exists in the relevant plan or docs.
- [ ] Require tests for every public package.
- [ ] Require integration tests for database, migrations, admin, auth, and queue behavior.
- [ ] Keep code examples compile-checked.

## Repository-Level Files

- Create: `go.mod`
- Create: `go.sum`
- Create: `Makefile`
- Create: `README.md`
- Create: `LICENSE`
- Create: `AGENTS.md`
- Create: `.gitignore`
- Create: `.env.example`
- Create: `cmd/gogo/main.go`
- Create: `gogo.go`
- Create: `docs/architecture/overview.md`
- Create: `docs/reference/package-map.md`
- Create: `docs/tutorials/quickstart.md`

## Cross-Phase Test Commands

- [ ] Run unit tests: `go test ./...`
- [ ] Run race tests for concurrency packages: `go test -race ./queue/... ./orm/... ./http/...`
- [ ] Run PostgreSQL integration tests: `GOGO_TEST_POSTGRES_DSN=$DSN go test -tags=integration ./...`
- [ ] Run Redis integration tests: `GOGO_TEST_REDIS_ADDR=$ADDR go test -tags=integration ./queue/... ./cache/...`
- [ ] Run admin browser checks after admin exists: `go test -tags=browser ./admin/...`
- [ ] Run example project tests after generators exist: `go test ./examples/...`

## Acceptance Checklist

- [ ] A developer can read `.plans` in order and implement the framework without hidden requirements.
- [ ] Every major Django and Celery parity area has a dedicated phase.
- [ ] Each phase names packages, files, commands, tests, and acceptance criteria.
- [ ] The planned client project structure matches the public framework package boundaries.
- [ ] Security-sensitive features have explicit tests and failure behavior.
