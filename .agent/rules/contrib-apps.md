# Contrib App Rules

Use this rule for `contrib/*`.

## Common Rules

- Contrib apps must integrate through normal app registry APIs.
- Model-backed contrib apps need migrations and admin registration.
- Contrib settings and middleware ordering must be documented.
- System checks should detect missing dependencies, unsafe settings, or unsupported dialect usage.

## Package Expectations

- `contrib/sites`: current-site resolution and middleware.
- `contrib/redirects`: safe redirect handling after 404 resolution.
- `contrib/flatpages`: site-aware static page rendering and registration requirements.
- `contrib/sitemaps`: sitemap indexes, sections, pagination, XML escaping.
- `contrib/syndication`: RSS and Atom output.
- `contrib/humanize`: template filters.
- `contrib/admindocs`: staff-only documentation pages.
- `contrib/postgres`: search, indexes, validators, aggregates, extension checks.
- `contrib/gis`: geometry metadata, spatial lookups/functions, measurement helpers.

## Verification

```bash
go test ./contrib/... ./checks
make docs-verify
```

