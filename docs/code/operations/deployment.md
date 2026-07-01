# Deployment Operations

This runbook covers the production deployment shape for a Gogo project. Read it
with `security.md`, `database.md`, `queues.md`, `admin.md`, and `upgrades.md`.

## Runtime Processes

A production deployment normally contains separate processes:

| Process | Responsibility |
| --- | --- |
| Web | HTTP routing, admin, APIs, auth, sessions, static/media mounts, health endpoints. |
| Worker | Background task execution from configured queues. |
| Beat | Periodic task scheduling. |
| Database | Application data, auth data, admin logs, migrations, optional SQL result backend. |
| Broker | Redis or RabbitMQ task transport. |
| Cache/result backend | Redis, SQL, or another configured backend for cache and task results. |

Run web, worker, and beat from the same application version. Do not deploy new
workers against an old web process when task envelopes or task names changed.

## Environment Variables

Required production variables:

```bash
GOGO_ENV=production
GOGO_DEBUG=false
GOGO_SECRET_KEY=
GOGO_ALLOWED_HOSTS=example.com,www.example.com
DATABASE_URL=
GOGO_HTTP_ADDR=:8000
GOGO_SESSION_COOKIE_SECURE=true
GOGO_CSRF_COOKIE_SECURE=true
GOGO_HTTPS_ENABLED=true
GOGO_ADMIN_PATH=/admin
GOGO_ADMIN_PATH_REVIEWED=true
GOGO_DEPLOY_MIGRATIONS_APPLIED=true
GOGO_DEPLOY_STATIC_COLLECTED=true
```

Common deployment variables:

```bash
GOGO_STATIC_URL=/static/
GOGO_STATIC_ROOT=/app/staticfiles
GOGO_MEDIA_URL=/media/
GOGO_MEDIA_ROOT=/app/media
GOGO_BROKER_URL=redis://redis:6379/0
GOGO_RESULT_BACKEND=redis://redis:6379/1
GOGO_CACHE_URL=redis://redis:6379/2
GOGO_EMAIL_URL=
GOGO_SESSION_COOKIE_NAME=gogo_sessionid
GOGO_CSRF_COOKIE_NAME=gogo_csrftoken
GOGO_CSRF_TRUSTED_ORIGINS=
GOGO_PASSWORD_RESET_ENABLED=
```

Keep `.env` out of Git. Store production secrets in the platform secret manager
and inject them into the process environment.

## Build And Start

Build a release binary from a tagged version:

```bash
go build -trimpath -o bin/gogo ./cmd/gogo
```

Generated projects can either call the installed `gogo` binary directly or wire
their project-specific `manage.go` to the framework runner once the application
has custom command registration.

Minimum startup order:

1. Database is reachable.
2. Broker and result backend are reachable when queue features are enabled.
3. Migrations are applied.
4. Static files are collected.
5. Web process starts.
6. Workers start.
7. Beat starts after at least one worker pool is ready.

Run deploy checks before allowing traffic:

```bash
go run manage.go check --deploy
```

Production deploy checks fail when `GOGO_BROKER_URL` or `GOGO_RESULT_BACKEND`
is set to `memory` or `memory://`. Use a durable backend such as Redis for
workers, or leave those variables empty when the deployment does not run queue
processes.

Validate worker runtime URLs before starting workers:

```bash
go run manage.go worker --check \
  --broker-url "$GOGO_BROKER_URL" \
  --result-backend "$GOGO_RESULT_BACKEND"
```

`redis://` and `rediss://` broker and result backend URLs must construct real
Redis runtime objects and fail if Redis is unreachable. `amqp://` and
`amqps://` remain unsupported runtime URLs until a real RabbitMQ transport is
registered; they never fall back to memory.

## Migrations

Review migration plans before applying them:

```bash
go run manage.go makemigrations --check --dry-run
go run manage.go migrate --plan
go run manage.go migrate
go run manage.go showmigrations
```

Run migrations once per deployment, not from every web or worker replica. Use a
single release job, init job, or operator action. `go run manage.go migrate`
takes a database-backed migration lock and fails before operations run if
another migration process already holds it.

For risky migrations:

- Back up the database first.
- Run the migration in staging against production-like data.
- Split schema additions, backfills, and destructive changes across releases.
- Stop or drain workers if they use tables being migrated.

## Static Files

Set `GOGO_STATIC_ROOT` to a directory or volume that is writable during release
and readable by the web process or fronting static server.

Use hashed manifests when static files are served by a CDN or immutable cache.
Duplicate static paths should be reviewed because project files override app
files and app files override framework files.

Collect static files during release, not on every request. Keep collected files
versioned by release or clear the destination before writing a complete new set.

## Media Storage

Set `GOGO_MEDIA_ROOT` to a writable directory or external mount outside the
source tree. Do not store uploads inside the executable, static root, or Git
checkout.

Production media storage must provide:

- Durable storage across restarts.
- Backups.
- Size limits.
- Type and extension validation.
- Private serving for sensitive files.
- Malware scanning for high-risk uploads.

## Health Checks

Expose liveness and readiness separately.

- Liveness should confirm the process event loop is alive.
- Readiness should include required dependencies such as database, broker,
  result backend, cache, and media root.

The `health.Registry` package supports required and optional checks, timeouts,
JSON reports, and HTTP handlers. Required readiness failures should remove the
instance from load balancing.

## Logging

Use structured logs. Include request ID, deployment version, process type,
hostname, route, status code, task name, task ID, queue name, migration name,
and admin action where relevant.

Never log passwords, password reset tokens, session cookies, CSRF tokens,
authorization headers, database URLs, broker URLs, or secret key values.

Use `observability.ConfigureLogger` and framework instrumentation hooks where
the project has not already integrated a platform logger.

## Metrics And Tracing

Track at minimum:

- HTTP request count, status, and duration.
- ORM query count and duration.
- Migration duration and failures.
- Admin actions and permission failures.
- Auth login success and failure.
- Queue task state, duration, retries, failures, and revocations.
- Broker queue depth.
- Worker concurrency and active tasks.
- Beat scheduler enqueue count and lock failures.

The `observability` package exposes meter, tracer, and instrumentation hooks
that can be adapted to OpenTelemetry or the platform metrics client.

## Rollbacks

Prepare rollback instructions before deployment.

- If migrations were not applied, roll back binary and static assets.
- If additive migrations were applied, roll back only after confirming the old
  code tolerates the added schema.
- If destructive migrations or data migrations ran, restore from backup or run
  a reviewed reverse migration.
- Stop workers before rolling back queue envelope or task-signature changes.
- Keep old static assets available until all old web replicas are drained.

Use `upgrades.md` for the detailed rollback decision tree.
