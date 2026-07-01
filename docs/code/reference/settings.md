# Settings Reference

Settings live in `conf.Settings` and are loaded from defaults, `.env`, and process environment through `conf.LoadFromEnv` and `conf.SettingsFromMap`.

## Public Types

| Type | Purpose |
| --- | --- |
| `conf.Settings` | Runtime configuration used by CLI commands, HTTP server, app registry, middleware, storage, queue, templates, and checks. |
| `conf.ErrInvalidSettings` | Error returned when required settings are missing or invalid. |

## Settings Fields And Environment Keys

| Field | Environment key | Required | Default | Notes |
| --- | --- | --- | --- | --- |
| `Env` | `GOGO_ENV` | yes | `development` | Allowed: `development`, `test`, `production`. |
| `SecretKey` | `GOGO_SECRET_KEY` | yes | empty | Required for signing, security, sessions, CSRF, and tokens. |
| `Debug` | `GOGO_DEBUG` | no | true in development | Must be false in production. |
| `AllowedHosts` | `GOGO_ALLOWED_HOSTS` | production | empty | Required in production. |
| `HTTPAddr` | `GOGO_HTTP_ADDR` | yes | `:8000` | Must parse as host:port. |
| `DatabaseURL` | `DATABASE_URL` | yes | empty | Primary database DSN. |
| `InstalledApps` | `GOGO_INSTALLED_APPS` | no | empty | Comma-separated app list. |
| `Middleware` | `GOGO_MIDDLEWARE` | no | empty | Comma-separated middleware list. |
| `RootURLConf` | `GOGO_ROOT_URLCONF` | project | empty | Root router identifier. |
| `StaticURL` | `GOGO_STATIC_URL` | no | `/static/` | Public static prefix. |
| `StaticRoot` | `GOGO_STATIC_ROOT` | deploy | empty | Collectstatic destination. |
| `MediaURL` | `GOGO_MEDIA_URL` | no | `/media/` | Public media prefix. |
| `MediaRoot` | `GOGO_MEDIA_ROOT` | upload | empty | Upload storage root. |
| `TemplateDirs` | `GOGO_TEMPLATE_DIRS` | no | empty | Comma-separated project template directories. |
| `DefaultAutoField` | `GOGO_DEFAULT_AUTO_FIELD` | no | `BigAutoField` | Default primary key type. |
| `TimeZone` | `GOGO_TIME_ZONE` | no | `UTC` | Application time zone. |
| `LanguageCode` | `GOGO_LANGUAGE_CODE` | no | `en-us` | Default language code. |
| `SessionCookieName` | `GOGO_SESSION_COOKIE_NAME` | no | `gogo_sessionid` | Session cookie name. |
| `SessionCookieSecure` | `GOGO_SESSION_COOKIE_SECURE` | deploy | false | Must be true in production deploy checks. |
| `CSRFCookieName` | `GOGO_CSRF_COOKIE_NAME` | no | `gogo_csrftoken` | CSRF cookie name. |
| `CSRFCookieSecure` | `GOGO_CSRF_COOKIE_SECURE` | deploy | false | Must be true in production deploy checks. |
| `HTTPSEnabled` | `GOGO_HTTPS_ENABLED` | deploy | false | Confirms TLS, redirects, and secure proxy handling are enabled. |
| `CSRFTrustedOrigins` | `GOGO_CSRF_TRUSTED_ORIGINS` | cross-origin forms | empty | Comma-separated HTTPS origins. |
| `AdminPath` | `GOGO_ADMIN_PATH` | no | `/admin` | Admin URL path reviewed by deploy checks. |
| `AdminPathReviewed` | `GOGO_ADMIN_PATH_REVIEWED` | deploy | false | Must be true after admin exposure is reviewed. |
| `MigrationsApplied` | `GOGO_DEPLOY_MIGRATIONS_APPLIED` | deploy | false | Release marker set after migrations are applied and verified. |
| `StaticFilesCollected` | `GOGO_DEPLOY_STATIC_COLLECTED` | deploy | false | Release marker set after static files are collected. |
| `PasswordResetEnabled` | `GOGO_PASSWORD_RESET_ENABLED` | auth email | false | Requires `GOGO_EMAIL_URL` in deploy checks. |
| `BrokerURL` | `GOGO_BROKER_URL` | queue workers | empty | Queue broker URL. |
| `ResultBackend` | `GOGO_RESULT_BACKEND` | task results | empty | Queue result backend URL. |
| `ScheduleStore` | `GOGO_SCHEDULE_STORE` | beat schedules | empty | Queue beat schedule store URL. |
| `CacheURL` | `GOGO_CACHE_URL` | cache | empty | Cache backend URL. |
| `EmailURL` | `GOGO_EMAIL_URL` | email | empty | Email backend URL. |

## CLI Commands

Root CLI commands are registered through `internal/cli`. Public users call the
installed `gogo` binary for global help, version, and project creation. Inside
a generated project, `go run manage.go <command>` is the explicit project
entrypoint, and the installed `gogo` binary delegates project-aware commands to
that entrypoint so commands load project settings, routes, admin, app configs,
model metadata, fixtures, queue tasks, project checks, custom project commands,
custom middleware registries, and server lifecycle hooks.

| Command | Status | Purpose |
| --- | --- | --- |
| `gogo help` | available | List commands. |
| `gogo --help` | available | List commands. |
| `gogo version` | available | Print version. |
| `gogo --version` | available | Print version. |
| `gogo <project-aware-command>` | available | Delegate to `go run manage.go <project-aware-command>` inside a generated project. |
| `go run manage.go check` | available | Load settings and run system checks. Use `--deploy` for production readiness checks. |
| `go run manage.go runserver` | available | Build middleware and run the HTTP server with project HTTP routes, API routes, admin, development static/media mounts, project server config, readiness hooks, and shutdown hooks. |
| `gogo startproject` | available | Generate project scaffold. |
| `go run manage.go startapp` | available | Generate app scaffold. |
| `go run manage.go makemigrations` | available | Write migration files. |
| `go run manage.go migrate` | available | Apply migrations. |
| `go run manage.go showmigrations` | available | Show migration status. |
| `go run manage.go sqlmigrate` | available | Render migration SQL. |
| `go run manage.go squashmigrations` | available | Write a squashed replacement migration with `Replaces` metadata. |
| `go run manage.go optimizemigration` | available | Optimize migration operations. |
| `go run manage.go createsuperuser` | available auth command shell | Create an admin user. |
| `go run manage.go changepassword` | available auth command shell | Change a user password. |
| `go run manage.go collectstatic` | available static command shell | Collect static files. |
| `go run manage.go shell` | available | Start app shell context. |
| `go run manage.go dbshell` | available | Open database shell. |
| `go run manage.go test` | available | Run project tests. |
| `go run manage.go worker` | available queue command shell | Run queue workers. |
| `go run manage.go beat` | available queue command shell | Run beat scheduler. |
| `go run manage.go inspect` | available queue command shell | Inspect workers. |
| `go run manage.go queues` | available queue command shell | Inspect queues. |
| `go run manage.go dumpdata` | available | Dump fixtures. |
| `go run manage.go loaddata` | available | Load fixtures. |

## Validation Rules

`Settings.Validate` fails when:

- `GOGO_ENV` is not one of `development`, `test`, or `production`.
- `GOGO_SECRET_KEY` is empty.
- `DATABASE_URL` is empty.
- `GOGO_HTTP_ADDR` is empty or invalid.
- `GOGO_ALLOWED_HOSTS` is empty in production.

## Example

```go
settings := conf.DefaultSettings()
settings.SecretKey = "dev-secret"
settings.DatabaseURL = "sqlite:///tmp/gogo.sqlite3"
err := settings.Validate()
_ = err
```
