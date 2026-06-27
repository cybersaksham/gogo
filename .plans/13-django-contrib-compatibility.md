# Django Contrib Compatibility Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Django contrib-style optional applications and helpers that sit outside the core request/model/admin/auth/queue stack but are expected in a full Django-parity framework.

**Architecture:** Contrib packages live under public `contrib/` namespaces and register as normal Gogo apps. Each contrib app ships models, migrations, admin registrations, views, templates, system checks, and tests through the same public framework APIs available to client apps.

**Tech Stack:** Gogo app registry, models, migrations, ORM, admin, templates, HTTP, cache, security middleware, PostgreSQL dialect extensions, GIS field metadata and spatial query operations.

---

## Files

- Create: `contrib/sites/app.go`
- Create: `contrib/sites/models.go`
- Create: `contrib/sites/middleware.go`
- Create: `contrib/sites/admin.go`
- Create: `contrib/sites/migrations/0001_initial.go`
- Create: `contrib/redirects/app.go`
- Create: `contrib/redirects/models.go`
- Create: `contrib/redirects/middleware.go`
- Create: `contrib/redirects/admin.go`
- Create: `contrib/redirects/migrations/0001_initial.go`
- Create: `contrib/flatpages/app.go`
- Create: `contrib/flatpages/models.go`
- Create: `contrib/flatpages/views.go`
- Create: `contrib/flatpages/admin.go`
- Create: `contrib/flatpages/templates/flatpages/default.html`
- Create: `contrib/flatpages/migrations/0001_initial.go`
- Create: `contrib/sitemaps/sitemap.go`
- Create: `contrib/sitemaps/views.go`
- Create: `contrib/syndication/feed.go`
- Create: `contrib/syndication/rss.go`
- Create: `contrib/syndication/atom.go`
- Create: `contrib/humanize/filters.go`
- Create: `contrib/admindocs/views.go`
- Create: `contrib/admindocs/templates/admindocs/index.html`
- Create: `contrib/postgres/search.go`
- Create: `contrib/postgres/indexes.go`
- Create: `contrib/postgres/validators.go`
- Create: `contrib/postgres/aggregates.go`
- Create: `contrib/gis/functions.go`
- Create: `contrib/gis/lookups.go`
- Create: `contrib/gis/measure.go`
- Create: `contrib/gis/geos.go`
- Create: `contrib/gis/gdal.go`
- Create: `contrib/gis/layermapping.go`
- Create: `contrib/gis/inspect.go`
- Create: `contrib/gis/sitemaps.go`

## Task 1: Add Sites Framework

- [ ] Create sites app files listed in this plan.
- [ ] Implement `Site` model with domain and name fields.
- [ ] Add `SITE_ID` setting support.
- [ ] Add current-site lookup by ID and request host.
- [ ] Add current-site middleware.
- [ ] Add admin registration with search by domain and name.
- [ ] Add migrations and system checks for duplicate domains and missing configured site.
- [ ] Add tests for current site by setting, current site by request, middleware attachment, admin metadata, and migration creation.
- [ ] Run `go test ./contrib/sites`.
- [ ] Commit with message `Add Sites Contrib App`.

## Task 2: Add Redirects Framework

- [ ] Create redirects app files listed in this plan.
- [ ] Implement `Redirect` model with site, old path, new path, and HTTP status behavior.
- [ ] Add redirect middleware that runs after 404 resolution.
- [ ] Support permanent redirects, temporary redirects, gone responses for empty target, and host-aware site filtering.
- [ ] Block unsafe redirect targets unless explicitly allowed by settings.
- [ ] Add admin registration with list display, filters, and search.
- [ ] Add tests for 404 redirect, no redirect on existing route, permanent redirect, temporary redirect, gone response, unsafe target rejection, and site filtering.
- [ ] Run `go test ./contrib/redirects`.
- [ ] Commit with message `Add Redirects Contrib App`.

## Task 3: Add Flatpages Framework

- [ ] Create flatpages app files listed in this plan.
- [ ] Implement `FlatPage` model with URL, title, content, enable comments flag, template name, registration required flag, and sites many-to-many.
- [ ] Add flatpage view that resolves page by path and current site.
- [ ] Render custom template when configured and default template otherwise.
- [ ] Enforce registration requirement through built-in auth.
- [ ] Add admin registration with fieldsets, sites widget, search, and filters.
- [ ] Add tests for lookup, site filtering, custom template, registration required, not found, and admin metadata.
- [ ] Run `go test ./contrib/flatpages`.
- [ ] Commit with message `Add Flatpages Contrib App`.

## Task 4: Add Sitemaps Framework

- [ ] Create sitemaps files listed in this plan.
- [ ] Define sitemap interface with items, location, last modified, change frequency, priority, alternates, protocol, and limit.
- [ ] Implement sitemap index view and sitemap section view.
- [ ] Support static views and model-backed sitemaps.
- [ ] Support pagination for large sitemaps.
- [ ] Escape XML safely.
- [ ] Add tests for index XML, section XML, pagination, last modified headers, alternate language links, and unsafe XML escaping.
- [ ] Run `go test ./contrib/sitemaps`.
- [ ] Commit with message `Add Sitemaps Contrib App`.

## Task 5: Add Syndication Feeds

- [ ] Create syndication files listed in this plan.
- [ ] Implement feed base with title, link, description, author, categories, item title, item description, item link, item publication date, item updated date, item author, item categories, enclosures, and feed URL.
- [ ] Render RSS 2.0 and Atom 1.0.
- [ ] Support per-request feed data and object-specific feeds.
- [ ] Escape XML safely.
- [ ] Add tests for RSS output, Atom output, item fields, enclosures, object-specific feed, and escaping.
- [ ] Run `go test ./contrib/syndication`.
- [ ] Commit with message `Add Syndication Feeds`.

## Task 6: Add Humanize Template Filters

- [ ] Create `contrib/humanize/filters.go`.
- [ ] Implement filters:
  - `apnumber`
  - `intcomma`
  - `intword`
  - `naturalday`
  - `naturaltime`
  - `ordinal`
- [ ] Register filters with template engine through app config.
- [ ] Add tests for each filter, localization hooks, invalid input handling, and template integration.
- [ ] Run `go test ./contrib/humanize ./templates`.
- [ ] Commit with message `Add Humanize Template Filters`.

## Task 7: Add Admin Documentation

- [ ] Create admindocs files listed in this plan.
- [ ] Generate admin documentation pages for registered models, admin classes, URL routes, template tags, template filters, settings, and management commands.
- [ ] Require staff permission to access admin docs.
- [ ] Add tests for model docs, route docs, filter docs, command docs, and permission enforcement.
- [ ] Run `go test ./contrib/admindocs`.
- [ ] Commit with message `Add Admin Documentation App`.

## Task 8: Add PostgreSQL Contrib Helpers

- [ ] Create postgres files listed in this plan.
- [ ] Implement full text search helpers:
  - Search vector
  - Search query
  - Search rank
  - Search headline
- [ ] Implement trigram helpers:
  - Similarity
  - Distance
  - Word similarity
- [ ] Implement PostgreSQL indexes:
  - B-tree
  - Hash
  - GIN
  - GiST
  - SP-GiST
  - BRIN
  - Bloom where extension is available
- [ ] Implement validators for array length, range bounds, and JSON structure.
- [ ] Implement aggregates:
  - Array aggregate
  - JSON object aggregate
  - String aggregate
  - Bool and
  - Bool or
  - Bit and
  - Bit or
  - Bit xor
- [ ] Add tests for SQL rendering, unsupported dialect errors, extension checks, and ORM integration.
- [ ] Run `go test ./contrib/postgres ./orm`.
- [ ] Commit with message `Add PostgreSQL Contrib Helpers`.

## Task 9: Add GIS Operations

- [ ] Create GIS files listed in this plan.
- [ ] Implement GEOS-style geometry wrappers for points, lines, polygons, collections, WKT, WKB, HEXEWKB, GeoJSON, SRID handling, prepared geometry hooks, and topology predicates.
- [ ] Implement GDAL-style data source metadata hooks for vector layers, fields, feature counts, spatial references, and coordinate transforms where supported by deployment dependencies.
- [ ] Implement LayerMapping-style import helpers for loading spatial data into models with transform, encoding, uniqueness, transaction, and progress options.
- [ ] Implement GIS inspect command support that can generate model field suggestions from spatial data sources.
- [ ] Implement spatial lookups:
  - Equals
  - Contains
  - Covered by
  - Covers
  - Crosses
  - Disjoint
  - Intersects
  - Overlaps
  - Relate
  - Touches
  - Within
  - Distance less than
  - Distance less than or equal
  - Distance greater than
  - Distance greater than or equal
- [ ] Implement spatial functions:
  - Area
  - AsGeoJSON
  - AsKML
  - AsSVG
  - Centroid
  - Difference
  - Distance
  - Envelope
  - Intersection
  - Length
  - Perimeter
  - PointOnSurface
  - Scale
  - SnapToGrid
  - SymDifference
  - Transform
  - Translate
  - Union
- [ ] Implement measurement types for distance and area.
- [ ] Add Geo sitemap helpers.
- [ ] Add tests for SQL rendering, measurement conversion, unsupported dialect errors, and sitemap output.
- [ ] Run `go test ./contrib/gis ./models/fields ./orm`.
- [ ] Commit with message `Add GIS Contrib Operations`.

## Task 10: Add Contrib Docs And Checks

- [ ] Document how to install every contrib app in `InstalledApps`.
- [ ] Document required middleware ordering for sites, redirects, flatpages, and messages.
- [ ] Document PostgreSQL extension requirements.
- [ ] Document GIS database requirements.
- [ ] Add system checks for missing dependencies, invalid settings, unsafe redirects, and unsupported dialect usage.
- [ ] Run docs verification and `go test ./contrib/... ./checks`.
- [ ] Commit with message `Document Contrib Apps`.

## Acceptance Checklist

- [ ] Sites, redirects, flatpages, sitemaps, feeds, humanize, admindocs, PostgreSQL helpers, and GIS operations are implemented.
- [ ] Every contrib app registers through normal app registry APIs.
- [ ] Every contrib app has migrations where it owns models.
- [ ] Admin registrations exist for model-backed contrib apps.
- [ ] System checks catch missing settings and unsupported database features.
- [ ] Docs explain installation, middleware, settings, and database requirements.
