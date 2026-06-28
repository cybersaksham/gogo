# Changelog

All notable changes to Gogo are documented in this file.

The format follows Keep a Changelog conventions and the project uses semantic
versioning after the first stable release.

## Unreleased

No unreleased changes.

## v0.1.0 - 2026-06-28

Initial pre-release of the app-structured Gogo framework.

### Release Metadata

- Previous release: none. This is the first public release.
- New release: `v0.1.0`.
- Module path: `github.com/cybersaksham/gogo`.
- CLI install path: `github.com/cybersaksham/gogo/cmd/gogo`.

### Added

- App-structured framework foundation with project/app lifecycle, application
  registry, settings loading, environment parsing, system checks, and CLI
  command wiring.
- Project and app generators for client projects, including generated settings,
  URL routing, app scaffolds, deployment templates, tests, `.env.example`, and
  framework-safe imports.
- HTTP package with routing, URL patterns, request/response helpers,
  middleware, redirects, generic views, decorators, and error handling.
- Model metadata and field packages covering scalar, text, numeric, temporal,
  binary, JSON, relation, file, network, PostgreSQL-specific, GIS, generated,
  choice, validation, index, and constraint behavior.
- ORM query construction and SQL compilation for filtering, lookups,
  expressions, aggregates, joins, managers, querysets, ordering, slicing,
  prefetch planning, transactions, locking, PostgreSQL dialect support, and
  SQLite dialect support.
- Migration and schema tooling with operations, plans, graph handling,
  migration records, autodetection, SQL rendering, executors, and CLI
  integration.
- Built-in auth toolkit with users, groups, permissions, content types,
  password hashing and validation, authentication, decorators, forms, tokens,
  middleware, admin metadata, and extensible model structures.
- Session, message, CSRF, security, cache, email, file storage, static file,
  template, form, signal, i18n, health, and observability packages.
- Admin package with model registration, sites, index views, change lists,
  change forms, delete views, history, permissions, filters, search,
  autocomplete, actions, inlines, widgets, URL wiring, and static assets.
- API package with serializers, model serializers, request/response helpers,
  routers, views, viewsets, authentication, permissions, pagination, parsers,
  renderers, filtering, throttling, uploads, validation, metadata, versioning,
  and OpenAPI generation.
- Queue package with task registry, brokers, result backends, workers,
  retries, timeouts, revocation, routing, beat scheduling, crontab/interval
  schedules, canvas groups/chains/chords, events, inspection, monitoring,
  message signing, CLI hooks, and admin integration points.
- Django-style contrib packages for content types, sites, redirects,
  flatpages, sitemaps, syndication feeds, humanize filters, admin docs,
  PostgreSQL helpers, and GIS helpers.
- Testing helpers for HTTP clients, settings, test databases, fixtures, mail,
  queues, admin assertions, and framework behavior.
- Documentation covering architecture, tutorials, references, operations,
  deployment, compatibility, upgrades, generators, contrib apps, and examples.
- GitHub Actions for CI and release publishing, release dry-run validation,
  benchmark coverage, dependency policy, security policy, production readiness
  checks, compatibility tests, and agent development harness rules.

### Verification

- Passed `make ci` before tagging.
- Passed `go test -tags=integration ./...` before tagging.
- Passed release dry run for `v0.1.0` before tagging.

### Artifacts

- The GitHub release workflow publishes CLI binaries for Linux, macOS, and
  Windows on `amd64` and `arm64`.
- The GitHub release workflow publishes `checksums.txt` with SHA256 checksums
  for release artifacts.
