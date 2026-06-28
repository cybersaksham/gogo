# Gogo

Gogo is a Go backend framework built around Django-style applications, models, migrations, admin, auth, APIs, forms, templates, static files, and Celery-style queues.

## Current Status

The framework implementation follows the incremental product plans in `.plans/`. The repository now contains the planned framework packages, generated client-project templates, docs, CI/release workflows, compatibility tests, benchmarks, deployment checks, and end-to-end generated-project verification.

Core implemented areas:

- App/project lifecycle with `startproject`, `startapp`, app registry, discovery, and lifecycle hooks.
- HTTP routing, middleware, views, URL reversing, static mounting, and `runserver`.
- Models, fields, metadata, validation, constraints, indexes, hooks, inheritance, and composition.
- ORM query engine, dialects, expressions, lookups, querysets, managers, transactions, and raw SQL.
- Migrations, operations, graph, writer/loader, recorder, executor, safety checks, and migration CLI commands.
- Built-in auth, users, groups, permissions, content types, sessions, password hashing, forms, middleware, admin registration, and auth CLI commands.
- Admin site, model registration, change lists, forms, inlines, actions, widgets, templates, static assets, delete, and history.
- API request/response handling, parsers, renderers, serializers, model serializers, viewsets, router, auth, permissions, pagination, filters, throttling, uploads, versioning, metadata, and OpenAPI generation.
- Forms, widgets, formsets, model forms, template engine, Django-compatible tags/filters, file storage, static collection, and fixtures.
- Queue task registry, brokers, result backends, workers, retries, timeouts, revocation, routing, beat schedules, canvas workflows, events, inspection, monitoring, message security, and queue CLI/admin integration.
- Cross-cutting cache, email, messages, signing, CSRF, security middleware, signals, i18n, health checks, observability, and system checks.
- Django contrib-style sites, redirects, flatpages, sitemaps, syndication, humanize, admin docs, PostgreSQL helpers, and GIS metadata/helpers.

## CLI

Available commands:

- `gogo help`
- `gogo version`
- `gogo check`
- `gogo runserver`
- `gogo startproject`
- `gogo startapp`
- `gogo makemigrations`
- `gogo migrate`
- `gogo showmigrations`
- `gogo sqlmigrate`
- `gogo squashmigrations`
- `gogo optimizemigration`
- `gogo createsuperuser`
- `gogo changepassword`
- `gogo collectstatic`
- `gogo shell`
- `gogo dbshell`
- `gogo test`
- `gogo worker`
- `gogo beat`
- `gogo inspect`
- `gogo queues`
- `gogo dumpdata`
- `gogo loaddata`

## Environment

Copy `.env.example` to `.env` for local development. Keep `.env` out of Git. Generated client projects must commit `.env.example` and must not commit `.env`.

Framework variables:

- `GOGO_ENV`: runtime environment. Defaults to `development`. Allowed values are `development`, `test`, and `production`.
- `GOGO_SECRET_KEY`: required secret key for signing and security-sensitive features.
- `GOGO_DEBUG`: optional debug flag. Defaults to true only in development.
- `GOGO_INSTALLED_APPS`: comma-separated installed app list.
- `GOGO_MIDDLEWARE`: comma-separated middleware list.
- `GOGO_ROOT_URLCONF`: root URL configuration identifier.
- `GOGO_DEFAULT_AUTO_FIELD`: default model auto field. Defaults to `BigAutoField`.
- `GOGO_TIME_ZONE`: application time zone. Defaults to `UTC`.
- `GOGO_LANGUAGE_CODE`: application language code. Defaults to `en-us`.

Database variables:

- `DATABASE_URL`: required database connection URL.

Server variables:

- `GOGO_HTTP_ADDR`: HTTP bind address. Defaults to `:8000`.

Static and media variables:

- `GOGO_STATIC_URL`: public static URL prefix. Defaults to `/static/`.
- `GOGO_STATIC_ROOT`: filesystem path for collected static files.
- `GOGO_MEDIA_URL`: public media URL prefix. Defaults to `/media/`.
- `GOGO_MEDIA_ROOT`: filesystem path for uploaded media.
- `GOGO_TEMPLATE_DIRS`: comma-separated template directories.

Queue variables:

- `GOGO_BROKER_URL`: queue broker URL. Required when queue workers are enabled.
- `GOGO_RESULT_BACKEND`: queue result backend URL. Required when task results are enabled.

Cache and email variables:

- `GOGO_CACHE_URL`: cache backend URL.
- `GOGO_EMAIL_URL`: email backend URL.

Session and CSRF variables:

- `GOGO_SESSION_COOKIE_NAME`: session cookie name. Defaults to `gogo_sessionid`.
- `GOGO_SESSION_COOKIE_SECURE`: set to true for production session cookies.
- `GOGO_CSRF_COOKIE_NAME`: CSRF cookie name. Defaults to `gogo_csrftoken`.
- `GOGO_CSRF_COOKIE_SECURE`: set to true for production CSRF cookies.
- `GOGO_CSRF_TRUSTED_ORIGINS`: comma-separated HTTPS origins allowed for trusted CSRF flows.

Security variables:

- `GOGO_ALLOWED_HOSTS`: comma-separated allowed hosts. Required in production.
- `GOGO_HTTPS_ENABLED`: set to true after HTTPS redirects and secure proxy handling are enabled.
- `GOGO_ADMIN_PATH`: admin URL path. Defaults to `/admin`.
- `GOGO_ADMIN_PATH_REVIEWED`: set to true after admin exposure has been reviewed.
- `GOGO_DEPLOY_MIGRATIONS_APPLIED`: set to true after production migrations are applied and verified.
- `GOGO_DEPLOY_STATIC_COLLECTED`: set to true after static files are collected for the release.
- `GOGO_PASSWORD_RESET_ENABLED`: set to true when password reset flows are enabled; requires `GOGO_EMAIL_URL` in deploy checks.

## Development

Run tests:

```bash
make test
```

Run checks:

```bash
GOGO_SECRET_KEY=dev DATABASE_URL=postgres://dev gogo check
```

Build the CLI:

```bash
make build
```

## Security

Gogo is intended for production-grade public products. Secrets, local databases, generated uploads, and machine-local files must not be committed. Required environment variables must fail fast during boot or checks when missing.
