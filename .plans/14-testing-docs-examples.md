# Testing Docs And Examples Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide a complete test toolkit, documentation set, examples, and verification workflow so framework users can build, test, and learn the product reliably.

**Architecture:** Public package `testing` gives client projects helpers for HTTP, ORM, migrations, auth, admin, API, queue, fixtures, and temporary settings. Docs and examples are compile-tested and run in CI.

**Tech Stack:** Go testing, `httptest`, temporary databases, fixtures, golden tests, documentation examples.

---

## Files

- Create: `testing/client.go`
- Create: `testing/database.go`
- Create: `testing/settings.go`
- Create: `testing/fixtures.go`
- Create: `testing/assertions.go`
- Create: `testing/mail.go`
- Create: `testing/queue.go`
- Create: `testing/admin.go`
- Create: `docs/architecture/overview.md`
- Create: `docs/architecture/package-boundaries.md`
- Create: `docs/reference/settings.md`
- Create: `docs/reference/models.md`
- Create: `docs/reference/orm.md`
- Create: `docs/reference/migrations.md`
- Create: `docs/reference/auth.md`
- Create: `docs/reference/admin.md`
- Create: `docs/reference/api.md`
- Create: `docs/reference/queue.md`
- Create: `docs/tutorials/quickstart.md`
- Create: `docs/tutorials/blog.md`
- Create: `docs/tutorials/admin.md`
- Create: `docs/tutorials/tasks.md`
- Create: `examples/blog/`

## Task 1: Add Test Client

- [ ] Create `testing/client.go`.
- [ ] Implement request helpers for GET, POST, PUT, PATCH, DELETE, OPTIONS, JSON requests, form requests, multipart requests, redirects, cookies, sessions, and authenticated users.
- [ ] Add response assertions for status, header, body contains, JSON path, template used, redirect target, and form errors.
- [ ] Add tests for every helper.
- [ ] Run `go test ./testing`.
- [ ] Commit with message `Add Framework Test Client`.

## Task 2: Add Test Database Helpers

- [ ] Create `testing/database.go`.
- [ ] Support temporary SQLite database, PostgreSQL test database creation from DSN, transaction wrapping, fixture loading, migration apply, and database reset.
- [ ] Add tests for SQLite lifecycle and PostgreSQL integration gated by `GOGO_TEST_POSTGRES_DSN`.
- [ ] Run `go test ./testing`.
- [ ] Commit with message `Add Test Database Helpers`.

## Task 3: Add Test Settings Helpers

- [ ] Create `testing/settings.go`.
- [ ] Support override settings, temporary installed apps, temporary middleware, temporary databases, temporary template directories, and automatic restore.
- [ ] Add tests for isolation and restore behavior.
- [ ] Run `go test ./testing`.
- [ ] Commit with message `Add Test Settings Helpers`.

## Task 4: Add Fixture Helpers

- [ ] Create `testing/fixtures.go`.
- [ ] Support loading fixtures, dumping fixtures, factory helpers, natural keys, and transactional fixture setup.
- [ ] Add tests with sample models.
- [ ] Run `go test ./testing`.
- [ ] Commit with message `Add Test Fixture Helpers`.

## Task 5: Add Mail Queue And Admin Test Helpers

- [ ] Create `testing/mail.go`.
- [ ] Create `testing/queue.go`.
- [ ] Create `testing/admin.go`.
- [ ] Provide in-memory email outbox, eager task execution, fake broker/backend, admin login helper, and admin page assertions.
- [ ] Add tests for mail capture, eager task success/failure, and admin helper behavior.
- [ ] Run `go test ./testing`.
- [ ] Commit with message `Add Mail Queue And Admin Test Helpers`.

## Task 6: Add Assertions

- [ ] Create `testing/assertions.go`.
- [ ] Add assertions:
  - Equal JSON
  - Has field error
  - Has non-field error
  - Query count
  - Signal sent
  - Email sent
  - Task enqueued
  - Permission granted
  - Permission denied
- [ ] Add tests for each assertion.
- [ ] Run `go test ./testing`.
- [ ] Commit with message `Add Framework Test Assertions`.

## Task 7: Write Architecture Docs

- [ ] Create `docs/architecture/overview.md`.
- [ ] Create `docs/architecture/package-boundaries.md`.
- [ ] Document package layers, client import rules, app lifecycle, request lifecycle, model-to-migration flow, ORM query flow, admin flow, API flow, queue flow, and extension rules.
- [ ] Add diagrams using Mermaid for lifecycle and dependency direction.
- [ ] Run markdown link check command defined in `Makefile`.
- [ ] Commit with message `Document Framework Architecture`.

## Task 8: Write Reference Docs

- [ ] Create reference docs listed in this plan.
- [ ] Cover every public package, public type, public setting, CLI command, model field, lookup, migration operation, admin option, serializer field, queue option, and error type.
- [ ] Include compile-checked examples for core APIs.
- [ ] Run docs example tests.
- [ ] Commit with message `Add Framework Reference Documentation`.

## Task 9: Write Tutorials

- [ ] Create quickstart tutorial that builds a project, creates an app, defines models, runs migrations, creates superuser, registers admin, creates API, and runs server.
- [ ] Create blog tutorial with relationships, forms, admin filters, API pagination, and background email task.
- [ ] Create admin tutorial covering custom ModelAdmin options.
- [ ] Create tasks tutorial covering retries, beat, chains, groups, and chords.
- [ ] Run tutorial commands against generated example project.
- [ ] Commit with message `Add Framework Tutorials`.

## Task 10: Add Blog Example

- [ ] Create `examples/blog`.
- [ ] Include models for author, post, tag, comment, and audit event.
- [ ] Include migrations, admin, API, forms, templates, static files, auth usage, queue tasks, and tests.
- [ ] Add Makefile target to run example tests.
- [ ] Run `go test ./examples/blog/...`.
- [ ] Commit with message `Add Blog Example Application`.

## Task 11: Add Documentation Verification

- [ ] Add scripts or Makefile targets:
  - Check markdown links
  - Check code examples compile
  - Check generated docs are current
  - Check tutorials run
- [ ] Run full docs verification.
- [ ] Commit with message `Add Documentation Verification`.

## Acceptance Checklist

- [ ] Client projects can test HTTP, ORM, migrations, auth, admin, API, queue, email, and fixtures using public helpers.
- [ ] Reference docs cover every public API introduced by earlier phases.
- [ ] Tutorials are runnable from a fresh checkout.
- [ ] Blog example demonstrates the full framework.
- [ ] Documentation verification runs in CI.

