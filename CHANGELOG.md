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

## v0.4.0 - 2026-06-30

Feature release for Django-style admin model pages, generated-project first-run
readiness, and downstream smoke reliability.

### Release Metadata

- Previous release: `v0.3.0`.
- New release: `v0.4.0`.
- Module path: `github.com/cybersaksham/gogo`.
- CLI install path: `github.com/cybersaksham/gogo/cmd/gogo`.

### Added

- Added Django-style admin changelist, add form, change form, delete
  confirmation, object history, autocomplete, and JavaScript catalog route
  rendering for registered model admins.
- Added Django-style admin login and password-change HTML forms with stable
  IDs, body classes, breadcrumbs, object tools, changelist controls, submit
  rows, and form field rendering.
- Added embedded admin CSS and JavaScript serving under each admin site's
  `/admin/static/` path so generated projects load admin styling without
  manual static-file setup.
- Added generated-project smoke coverage for first-run module readiness and
  the downstream admin page workflow.
- Added a client project smoke report documenting automatic scaffolded files,
  manual local-only setup, verified commands, HTTP checks, and remaining
  expected env behavior.

### Changed

- Changed generated project module hydration to be best-effort so local builds,
  unpublished test versions, or temporarily unavailable module proxies do not
  make `startproject` fail.
- Changed generated project `go.mod` output for released CLIs to include the
  framework dependency graph needed by first-run project-local commands.
- Changed admin page templates and bundled styles to follow the familiar Django
  admin structure, selectors, and visual layout more closely.

### Fixed

- Fixed generated projects requiring manual module tidying before
  `go run manage.go startapp` could run in common release installs.
- Fixed registered admin model routes returning placeholder text such as
  `admin:notes_item_changelist` instead of useful HTML pages.
- Fixed admin add forms showing invalid History/Delete links and double-slash
  delete URLs before an object exists.
- Fixed admin form widgets rendering empty values as `&lt;nil&gt;`.
- Fixed generated-project admin pages linking to `/static/admin.css` and
  `/static/admin.js`, which could be intercepted by the project's static mount
  and return `404`.

### Breaking

- None.

### Migration Notes

- Existing generated projects should update their `go.mod` requirement to
  `github.com/cybersaksham/gogo v0.4.0` and run `go mod tidy`.
- Existing generated projects that already mount admin routes will get the new
  admin pages and `/admin/static/` assets after updating the framework version.
- Required environment variables remain unchanged. Projects still need a valid
  `GOGO_SECRET_KEY` and `DATABASE_URL` for checks and runtime commands.

### Verification

- Passed `make ci` before tagging.
- Passed `go test -tags=integration ./...` before tagging.
- Passed release dry run for `v0.4.0` before tagging.

### Artifacts

- The GitHub release workflow publishes CLI binaries for Linux, macOS, and
  Windows on `amd64` and `arm64`.
- The GitHub release workflow publishes `checksums.txt` with SHA256 checksums
  for release artifacts.

## v0.3.0 - 2026-06-29

Feature release for generated-project admin authentication, real migration
application, generated module version pinning, and client smoke reliability.

### Release Metadata

- Previous release: `v0.2.0`.
- New release: `v0.3.0`.
- Module path: `github.com/cybersaksham/gogo`.
- CLI install path: `github.com/cybersaksham/gogo/cmd/gogo`.

### Added

- Added `auth.FileUserStore`, a public file-backed built-in user store used by
  generated admin projects and available for simple deployments.
- Added `admin.SessionPermissionPolicy` so admin routes can authorize requests
  from auth context or admin session cookies.
- Added generated admin auth wiring: new projects use `.gogo/auth_users.json`
  for `createsuperuser` users and `.gogo/sessions` for admin sessions.
- Added release-aware `go.mod` generation so projects created by a released
  `gogo` CLI pin `github.com/cybersaksham/gogo` to the current module version.
- Added a client smoke report covering generated-project command behavior and
  the fixes required after the `v0.2.0` smoke run.

### Changed

- Changed admin URL generation to route login, logout, and password-change views
  through configured handlers instead of placeholders.
- Changed protected admin routes to enforce the site permission policy, redirect
  anonymous users to `/admin/login/`, and reject authenticated non-staff users.
- Changed `runserver` to accept Django-style positional addresses such as
  `go run manage.go runserver :8111`, while rejecting extra positional
  arguments.
- Changed generated client agent rules and public docs to describe generated
  admin auth storage, session storage, and framework version pinning.

### Fixed

- Fixed `migrate` so it opens the configured database, applies generated app
  migrations, executes generated SQL, and records rows in `gogo_migrations`.
- Fixed `showmigrations` so it marks applied migrations with `[X]` using the
  migration recorder.
- Fixed `migrate --plan` so it lists pending migrations and reports when there
  is no migration work to apply.
- Fixed `makemigrations --check --dry-run` so existing initial migrations are
  not proposed again, and missing migrations return a failing check status.
- Fixed generated and squashed migration files so multiple migrations in one Go
  package use unique variable names and compile together.
- Fixed generated-project integration tests to run migration commands from the
  generated project root, matching real client usage.

### Breaking

- Admin routes produced by `admin.Site.URLs()` now enforce the configured
  permission policy by default. Tests, probes, or generated projects that
  expected anonymous `/admin/` access must sign in with a staff user or attach a
  staff user in request context.

### Migration Notes

- Existing generated projects should run `go run manage.go createsuperuser` to
  create a staff admin user, then sign in at `/admin/login/`.
- Existing generated projects can adopt the new admin login/session behavior by
  updating project `admin.go` to use `auth.NewFileUserStore`,
  `sessions.NewFileStore`, `admin.LoginView`, `admin.LogoutView`,
  `admin.PasswordChangeView`, and `admin.SessionPermissionPolicy`.
- Migration commands now require a valid `DATABASE_URL` when applying or reading
  recorded migration state. `migrate --plan` still renders a pending plan when
  database state is unavailable.
- Existing generated projects can add a `require github.com/cybersaksham/gogo
  v0.3.0` line to `go.mod`, then run `go mod tidy`.

### Verification

- Passed `make ci` before tagging.
- Passed `go test -tags=integration ./...` before tagging.
- Passed release dry run for `v0.3.0` before tagging.

### Artifacts

- The GitHub release workflow publishes CLI binaries for Linux, macOS, and
  Windows on `amd64` and `arm64`.
- The GitHub release workflow publishes `checksums.txt` with SHA256 checksums
  for release artifacts.

## v0.2.0 - 2026-06-29

Feature release for generated-project runtime completeness, public API mounting,
admin usability, and management command output clarity.

### Release Metadata

- Previous release: `v0.1.4`.
- New release: `v0.2.0`.
- Module path: `github.com/cybersaksham/gogo`.
- CLI install path: `github.com/cybersaksham/gogo/cmd/gogo`.

### Added

- Added public API-to-HTTP bridge helpers (`api.Router.MountHTTP` and
  `api.Response.HTTP`) so generated app API routes can be served through the
  project HTTP router.
- Added generated-project API wiring so app `RegisterAPI` functions are
  automatically mounted under `/api/`.
- Added development static and media serving through `runserver` when static or
  media roots are configured.
- Added a rendered admin index that lists registered admin models instead of
  returning a plain placeholder route name.
- Added concrete `squashmigrations` output with generated replacement migration
  files and `Replaces` metadata.

### Changed

- Changed generated `startapp` wiring dedupe so one app import can install
  multiple markers for routes, APIs, admin, queue tasks, and app config safely.
- Changed `optimizemigration` to report a clear no-op when no safe rewrite is
  available.
- Changed `queues`, `dumpdata`, and `loaddata` to print explicit empty or
  success summaries.
- Updated generated agent rules and public/code docs for API mounting,
  development static/media serving, admin index behavior, and migration command
  behavior.

### Fixed

- Fixed generated API endpoints returning `404` despite generated `api.go`
  files and API registry metadata.
- Fixed development `runserver` not serving collected static app assets or
  configured media files.
- Fixed `squashmigrations` and `optimizemigration` appearing successful while
  not creating/changing artifacts or explaining no-op behavior.
- Fixed ambiguous empty output from queue and fixture inspection commands.
- Fixed admin root returning only `admin:index` text instead of a useful model
  index.

### Breaking

- None.

### Migration Notes

- Existing generated projects can continue using `v0.1.x` project wiring. To
  adopt automatic API mounting, regenerate or update project `urls.go` with
  `NewAPIRouter`, `RegisterAPIRoutes`, and `api.Router.MountHTTP`, then rerun
  `startapp` or wire existing app `RegisterAPI` functions manually.
- `squashmigrations` now creates a new squashed migration file instead of only
  printing a success message.

### Verification

- Passed `make ci` before tagging.
- Passed `go test -tags=integration ./...` before tagging.
- Passed release dry run for `v0.2.0` before tagging.

### Artifacts

- The GitHub release workflow publishes CLI binaries for Linux, macOS, and
  Windows on `amd64` and `arm64`.
- The GitHub release workflow publishes `checksums.txt` with SHA256 checksums
  for release artifacts.

## v0.1.4 - 2026-06-29

Patch release for generated-project runtime parity and management UX improvements.

### Release Metadata

- Previous release: `v0.1.3`.
- New release: `v0.1.4`.
- Module path: `github.com/cybersaksham/gogo`.
- CLI install path: `github.com/cybersaksham/gogo/cmd/gogo`.

### Added

- Added a public management runner and project execution path (`go run manage.go`)
  so generated projects can execute all local workflow commands against project
  settings, app configs, router, and queue app without importing
  `github.com/cybersaksham/gogo/internal`.
- Added automatic generated-project app wiring during `startapp`, including
  generated app registration in project settings, routing/admin/task markers, and
  app config wiring in generated `app.go`.
- Added automatic sync of generated app env labels into both `.env.example` and
  generated `.env`.

### Changed

- Changed migration discovery in management commands to include generated apps and
  app directories under `apps/*` by default, with generated-app-specific SQL and
  migration listing behavior.
- Updated generated project CLI guidance and templates to consistently use
  project-local management command execution.
- Refreshed public package inventory and generated project agent rule coverage for
  CLI command entry points, models, migrations, HTTP/admin/API/auth/queues,
  templates, and settings/security workflows.

### Fixed

- Fixed app registry readiness deadlocks and blocked registrations while app
  startup is preparing.
- Fixed `collectstatic` defaults to use configured static root and project/app
  static folders automatically.
- Fixed command runtime discovery so queue inspection/dispatch tests run against
  the generated project queue graph.
- Fixed persistent CLI auth user storage so generated projects retain superusers
  and password changes across command invocations.

### Verification

- Passed `go run ./internal/release/cmd/dryrun --tag v0.1.4 --commit "$(git rev-parse HEAD)" --build-date "$(date -u +'%Y-%m-%dT%H:%M:%SZ')" --changelog CHANGELOG.md --notes-out release-notes.md`.

### Artifacts

- The GitHub release workflow publishes CLI binaries for Linux, macOS, and Windows
  on `amd64` and `arm64` with `checksums.txt`.

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
