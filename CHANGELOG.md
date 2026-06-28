# Changelog

All notable changes to Gogo are documented in this file.

The format follows Keep a Changelog conventions and the project uses semantic
versioning after the first stable release.

## Unreleased

### Added

None.

### Changed

None.

### Fixed

None.

## v0.1.3 - 2026-06-29

Patch release for release readiness and agent guidance completeness.

### Release Metadata

- Previous release: `v0.1.2`.
- New release: `v0.1.3`.
- Module path: `github.com/cybersaksham/gogo`.
- CLI install path: `github.com/cybersaksham/gogo/cmd/gogo`.

### Added

- Added public documentation pages for the complete public package map and
  contrib/cross-cutting framework features.
- Added a generated client-project agent rule containing the public package
  index, contrib package index, CLI command surface, and feature discovery
  workflow.

### Changed

- Expanded generated client-project Gogo agent rules with more complete
  feature coverage for models, fields, migrations, HTTP, admin, APIs, auth,
  forms, templates, static files, uploads, queues, settings, testing,
  deployment, contrib apps, and cross-cutting framework services.
- Updated public and README installation instructions to pin `v0.1.3`.

### Fixed

- Fixed CI and release workflow setup so public documentation dependencies are
  installed before docs checks run on clean GitHub runners.
- Fixed manual release reruns so CLI binary metadata and GitHub release target
  metadata use the checked-out tag commit, and existing release assets can be
  replaced without moving the tag.

## v0.1.2 - 2026-06-28

Patch release for client-project AI-agent guidance and public documentation
publishing.

### Release Metadata

- Previous release: `v0.1.1`.
- New release: `v0.1.2`.
- Module path: `github.com/cybersaksham/gogo`.
- CLI install path: `github.com/cybersaksham/gogo/cmd/gogo`.

### Added

- Added generated `.agent/rules/gogo.md` and `.agent/rules/gogo/*` files to
  new projects created with `gogo startproject` so downstream AI agents can
  understand Gogo project structure, models, migrations, HTTP, admin, API,
  auth, forms, static files, queues, settings, security, testing, and
  deployment workflows.
- Added an Astro Starlight public documentation site under `docs/public`
  covering installation, project creation, feature areas, project layout,
  settings, models, ORM, migrations, HTTP, admin, API, auth, forms, templates,
  static files, queues, CLI usage, testing, and deployment.
- Added a GitHub Pages workflow that audits, checks, builds, uploads, and
  deploys the public documentation site from `docs/public/dist`.

### Changed

- Moved existing maintainer-oriented markdown documentation under `docs/code`
  while keeping user-facing public documentation under `docs/public`.
- Raised the required Go toolchain to `1.26.4` for the root module, generated
  projects, generated Docker builds, CI, and release workflows to avoid known
  reachable vulnerabilities in older Go `1.26.x` standard library releases.
- Updated docs verification to include MDX files, Starlight route links, public
  docs npm audit, Starlight type checks, and static docs builds.
- Updated docs Makefile targets for public docs install, audit, check, and
  build workflows.
- Configured the public docs site for the GitHub Pages project path
  `https://cybersaksham.github.io/gogo/`.

### Verification

- Passed `make ci` before tagging.
- Passed `go test -tags=integration ./...` before tagging.
- Passed `govulncheck ./...` with Go `1.26.4` before tagging.
- Passed release dry run for `v0.1.2` before tagging.

### Artifacts

- The GitHub release workflow publishes CLI binaries for Linux, macOS, and
  Windows on `amd64` and `arm64`.
- The GitHub release workflow publishes `checksums.txt` with SHA256 checksums
  for release artifacts.

## v0.1.1 - 2026-06-28

Patch release for CLI install/version behavior and repository onboarding
documentation.

### Release Metadata

- Previous release: `v0.1.0`.
- New release: `v0.1.1`.
- Module path: `github.com/cybersaksham/gogo`.
- CLI install path: `github.com/cybersaksham/gogo/cmd/gogo`.

### Added

- Added the MIT `LICENSE` file.
- Added `CONTRIBUTING.md` with setup, verification, security, compatibility,
  pull request, and release contribution guidance.
- Added an agent rule that keeps `README.md` limited to download, setup,
  contribution, security, and license information while product documentation
  stays under `docs/`.

### Changed

- Simplified `README.md` to only cover download, setup, contribution, security,
  and license information.
- Updated download instructions to pin `v0.1.1` and document the direct Git
  fallback for new tags that are not yet available from the public Go checksum
  database.
- Updated CLI settings reference documentation to include `gogo --help` and
  `gogo --version`.

### Fixed

- Added `gogo --version`, `gogo -version`, `gogo --help`, and `gogo -h` root
  aliases.
- Fixed source installs from `go install github.com/cybersaksham/gogo/cmd/gogo@vX.Y.Z`
  so `gogo version` can report the Go module version when release linker flags
  are not present.
- Moved environment-variable documentation coverage checks from `README.md` to
  `docs/reference/settings.md` so the README remains onboarding-only.

### Verification

- Passed `make ci` before tagging.
- Passed `go test -tags=integration ./...` before tagging.
- Passed release dry run for `v0.1.1` before tagging.

### Artifacts

- The GitHub release workflow publishes CLI binaries for Linux, macOS, and
  Windows on `amd64` and `arm64`.
- The GitHub release workflow publishes `checksums.txt` with SHA256 checksums
  for release artifacts.

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
