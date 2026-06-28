# Contrib Apps

Gogo contrib packages mirror Django's optional batteries. They are installed through `InstalledApps`, use normal middleware ordering, and are checked by `gogo check` through the system checks framework.

## InstalledApps

Install only the contrib apps a project uses. Site-aware apps must list `gogo.contrib.sites` first.

```go
InstalledApps: []string{
	"gogo.contrib.sites",
	"gogo.contrib.redirects",
	"gogo.contrib.flatpages",
	"gogo.contrib.sitemaps",
	"gogo.contrib.syndication",
	"gogo.contrib.humanize",
	"gogo.contrib.admindocs",
	"gogo.contrib.postgres",
	"gogo.contrib.gis",
}
```

`gogo.contrib.sites` provides the `Site` model, `SITE_ID` based lookup, request host lookup, current-site middleware, admin metadata, migration metadata, and duplicate-domain checks.

`gogo.contrib.redirects` provides the `Redirect` model, 404 redirect middleware, permanent and temporary redirects, gone responses for empty targets, unsafe target blocking, admin metadata, and migration metadata. Keep `AllowUnsafeTargets` disabled unless redirect rows are trusted and reviewed.

`gogo.contrib.flatpages` provides the `FlatPage` model, site-filtered page resolution, default and custom templates, registration-required enforcement, admin metadata, and migration metadata.

`gogo.contrib.sitemaps` provides sitemap interfaces, static and model-backed sitemap helpers, sitemap index and section rendering, pagination, alternates, priorities, change frequency, last modified data, and XML escaping.

`gogo.contrib.syndication` provides RSS 2.0 and Atom 1.0 feed rendering with feed metadata, item metadata, object-specific feeds, enclosures, and XML escaping.

`gogo.contrib.humanize` provides template filters: `apnumber`, `intcomma`, `intword`, `naturalday`, `naturaltime`, and `ordinal`.

`gogo.contrib.admindocs` provides staff-only admin documentation for registered models, admin classes, routes, template tags, template filters, settings, and management commands.

`gogo.contrib.postgres` provides PostgreSQL full-text search helpers, trigram helpers, PostgreSQL indexes, validators, and aggregates.

`gogo.contrib.gis` provides PostGIS SQL functions and lookups, GEOS-style geometry wrappers, GDAL-style metadata hooks, LayerMapping-style import plans, inspect suggestions, measurements, and geo sitemap helpers.

## Middleware Order

Use a deterministic order. Security, host validation, session, and auth middleware should wrap requests before contrib middleware that depends on request state.

```go
Middleware: []string{
	"gogo.http.RequestIDMiddleware",
	"gogo.http.PanicRecoveryMiddleware",
	"gogo.http.HostValidationMiddleware",
	"gogo.auth.SessionMiddleware",
	"gogo.auth.AuthenticationMiddleware",
	"gogo.contrib.sites.Middleware",
	"gogo.messages.Middleware",
	"gogo.contrib.flatpages.Middleware",
	"gogo.contrib.redirects.Middleware",
}
```

`gogo.contrib.sites.Middleware` must run before site-aware flatpage or redirect resolution so request handlers can use the current site. `gogo.messages.Middleware` should run after session/auth storage is available and before views/templates read messages. `gogo.contrib.redirects.Middleware` should run late because it inspects 404 responses and must not redirect requests that existing routes handled.

## PostgreSQL Extension Requirements

`gogo.contrib.postgres` requires the PostgreSQL dialect. Some helpers require database extensions:

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS btree_gin;
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE EXTENSION IF NOT EXISTS hstore;
CREATE EXTENSION IF NOT EXISTS unaccent;
```

Full-text search helpers use built-in PostgreSQL search functions. Trigram similarity and distance helpers require `pg_trgm`. GIN and GiST indexes work best with the matching operator class and extension for the indexed data type.

## GIS Database Requirements

`gogo.contrib.gis` requires PostgreSQL with PostGIS enabled:

```sql
CREATE EXTENSION IF NOT EXISTS postgis;
```

Spatial model fields and GIS contrib SQL helpers are PostGIS-only. Non-PostgreSQL dialects must fail checks before migrations or queries use GIS features. Native GEOS/GDAL libraries are optional for deployments that need external spatial file parsing or coordinate transforms; the framework exposes metadata and transform hooks so projects can wire those libraries explicitly.

## System Checks

Call `checks.RegisterContribChecks` from project setup or feed `checks.ContribChecks` into a custom check command. `gogo check` should receive the installed app list, middleware list, site setting, database dialect, database extensions, and redirect safety settings.

Checks cover:

- Missing dependencies such as redirects or flatpages without sites.
- Invalid `SITE_ID` for site-aware contrib apps.
- Incorrect order for `gogo.contrib.sites.Middleware`, `gogo.contrib.flatpages.Middleware`, and `gogo.contrib.redirects.Middleware`.
- `AllowUnsafeTargets` being enabled for redirects.
- Unsupported dialect usage for `gogo.contrib.postgres` and `gogo.contrib.gis`.
- Missing PostGIS and trigram extension signals.
