# Production Hardening And Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prepare the framework for production-grade public use with security review, compatibility policy, CI, benchmarks, dependency policy, release automation, migration guarantees, and operational documentation.

**Architecture:** Production hardening validates the whole framework as a cohesive product, not only individual packages. CI runs unit, integration, race, security, examples, generated project, docs, and benchmark smoke checks.

**Tech Stack:** Go test, Go vet, race detector, vulnerability scanning, GitHub Actions or equivalent CI, release tags, changelog, reproducible builds.

---

## Files

- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/release.yml`
- Create: `SECURITY.md`
- Create: `CHANGELOG.md`
- Create: `docs/operations/deployment.md`
- Create: `docs/operations/security.md`
- Create: `docs/operations/database.md`
- Create: `docs/operations/queues.md`
- Create: `docs/operations/admin.md`
- Create: `docs/operations/upgrades.md`
- Create: `benchmarks/orm_bench_test.go`
- Create: `benchmarks/router_bench_test.go`
- Create: `benchmarks/queue_bench_test.go`
- Create: `internal/release/checks.go`

## Task 1: Add CI Pipeline

- [ ] Create `.github/workflows/ci.yml`.
- [ ] Include jobs:
  - Format check
  - `go vet ./...`
  - Unit tests
  - Race tests for concurrency packages
  - PostgreSQL integration tests
  - Redis integration tests
  - RabbitMQ integration tests
  - Generated project tests
  - Example tests
  - Documentation verification
  - Vulnerability scan
- [ ] Ensure secrets are referenced only through CI secret names.
- [ ] Add local `make ci` target matching CI commands.
- [ ] Commit with message `Add CI Pipeline`.

## Task 2: Add Security Policy

- [ ] Create `SECURITY.md`.
- [ ] Document supported versions, vulnerability reporting process, expected response timeline, secret handling policy, dependency update policy, and disclosure process.
- [ ] Create `docs/operations/security.md`.
- [ ] Document production settings for secret key, allowed hosts, HTTPS, secure cookies, CSRF trusted origins, admin path, password hashing, sessions, CORS policy, rate limiting, and uploaded files.
- [ ] Run documentation verification.
- [ ] Commit with message `Add Security Policy`.

## Task 3: Add Release Workflow

- [ ] Create `.github/workflows/release.yml`.
- [ ] Build CLI binaries for supported platforms.
- [ ] Generate checksums.
- [ ] Publish release notes from `CHANGELOG.md`.
- [ ] Run full CI before release job.
- [ ] Add dry-run release script under `internal/release`.
- [ ] Commit with message `Add Release Workflow`.

## Task 4: Add Compatibility Policy

- [ ] Document semantic versioning policy.
- [ ] Define public API stability rules.
- [ ] Define migration file compatibility rules.
- [ ] Define generated project compatibility rules.
- [ ] Define supported Go versions.
- [ ] Define supported databases and minimum versions.
- [ ] Define supported brokers and minimum versions.
- [ ] Commit with message `Document Compatibility Policy`.

## Task 5: Add Upgrade Documentation

- [ ] Create `docs/operations/upgrades.md`.
- [ ] Document how to upgrade framework version in client projects.
- [ ] Document migration safety checks.
- [ ] Document breaking change handling.
- [ ] Document generated template drift and how to compare new generated templates with existing projects.
- [ ] Add upgrade checklist for production apps.
- [ ] Run documentation verification.
- [ ] Commit with message `Document Upgrade Process`.

## Task 6: Add Operational Docs

- [ ] Create `docs/operations/deployment.md`.
- [ ] Create `docs/operations/database.md`.
- [ ] Create `docs/operations/queues.md`.
- [ ] Create `docs/operations/admin.md`.
- [ ] Cover:
  - Environment variables
  - Database migrations
  - Static collection
  - Media storage
  - Admin security
  - Worker deployment
  - Beat deployment
  - Queue monitoring
  - Health checks
  - Logging
  - Metrics
  - Backups
  - Rollbacks
- [ ] Run documentation verification.
- [ ] Commit with message `Add Operational Documentation`.

## Task 7: Add Benchmarks

- [ ] Create benchmark files listed in this plan.
- [ ] Benchmark:
  - Router match
  - Route reverse
  - Simple ORM select compile
  - ORM insert compile
  - ORM scan
  - Migration autodetect on small app
  - Admin changelist query planning
  - Serializer validation
  - Queue publish
  - Queue worker task execution with fake broker
- [ ] Add `make bench`.
- [ ] Store benchmark guidance in docs without committing machine-specific numbers as pass/fail thresholds.
- [ ] Commit with message `Add Framework Benchmarks`.

## Task 8: Add Dependency Policy

- [ ] Document approved dependency categories:
  - Database drivers
  - Redis client
  - RabbitMQ client
  - Password hashing packages
  - OpenTelemetry interfaces
  - YAML parser if needed for docs or config
- [ ] Document rejected dependency categories:
  - Web frameworks that replace framework HTTP
  - ORM libraries that replace Gogo ORM
  - Migration libraries that replace Gogo migrations
  - Admin generators that replace Gogo admin
- [ ] Add `go list -m all` dependency audit command to CI.
- [ ] Commit with message `Document Dependency Policy`.

## Task 9: Add Backward Compatibility Tests

- [ ] Add fixtures representing older migration files, settings files, generated projects, serialized queue messages, signed sessions, password hashes, and admin URLs.
- [ ] Ensure current framework can read or produce clear upgrade errors for those fixtures.
- [ ] Add tests for each compatibility fixture.
- [ ] Run `go test ./...`.
- [ ] Commit with message `Add Compatibility Tests`.

## Task 10: Add Production Readiness Check

- [ ] Create `internal/release/checks.go`.
- [ ] Add `gogo check --deploy`.
- [ ] Check:
  - Debug disabled
  - Secret key strong
  - Allowed hosts explicit
  - Secure cookies enabled
  - HTTPS settings enabled
  - CSRF trusted origins valid
  - Database reachable
  - Migrations applied
  - Static files collected
  - Media storage writable
  - Admin path reviewed
  - Queue broker reachable when tasks are enabled
  - Result backend reachable when result storage is enabled
  - Email backend configured when password reset is enabled
- [ ] Add tests for every deploy check.
- [ ] Run `go test ./checks ./internal/cli ./internal/release`.
- [ ] Commit with message `Add Production Readiness Checks`.

## Task 11: Add Final End-To-End Verification

- [ ] Generate a fresh project.
- [ ] Generate two apps.
- [ ] Add models with scalar fields, relationship fields, constraints, and indexes.
- [ ] Generate migrations.
- [ ] Apply migrations.
- [ ] Create superuser.
- [ ] Register admin.
- [ ] Add API viewsets.
- [ ] Add forms and templates.
- [ ] Add queue task, periodic schedule, chain, group, and chord.
- [ ] Run app server test.
- [ ] Run worker and beat test with fake or integration broker.
- [ ] Run `gogo check --deploy` against production-like settings.
- [ ] Run `go test ./...`.
- [ ] Commit with message `Verify Framework End To End`.

## Acceptance Checklist

- [ ] CI covers unit, integration, race, docs, examples, generated projects, security, and compatibility.
- [ ] Security policy and production docs are written.
- [ ] Release workflow produces reproducible artifacts and checksums.
- [ ] Compatibility and upgrade policies are explicit.
- [ ] Benchmarks exist for key framework paths.
- [ ] Deploy checks catch unsafe production settings.
- [ ] Fresh generated project can exercise the whole framework end to end.

