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
| `CSRFCookieName` | `GOGO_CSRF_COOKIE_NAME` | no | `gogo_csrftoken` | CSRF cookie name. |
| `BrokerURL` | `GOGO_BROKER_URL` | queue workers | empty | Queue broker URL. |
| `ResultBackend` | `GOGO_RESULT_BACKEND` | task results | empty | Queue result backend URL. |
| `CacheURL` | `GOGO_CACHE_URL` | cache | empty | Cache backend URL. |
| `EmailURL` | `GOGO_EMAIL_URL` | email | empty | Email backend URL. |

## CLI Commands

Root CLI commands are registered through `internal/cli`. Public users call the `gogo` binary.

| Command | Status | Purpose |
| --- | --- | --- |
| `gogo help` | available | List commands. |
| `gogo version` | available | Print version. |
| `gogo check` | available | Load settings and run system checks. |
| `gogo runserver` | available | Build middleware and run the HTTP server skeleton. |
| `gogo startproject` | available | Generate project scaffold. |
| `gogo startapp` | available | Generate app scaffold. |
| `gogo makemigrations` | available | Write migration files. |
| `gogo migrate` | planned through migrations runtime | Apply migrations. |
| `gogo showmigrations` | planned through migrations runtime | Show migration status. |
| `gogo sqlmigrate` | planned through migrations runtime | Render migration SQL. |
| `gogo squashmigrations` | planned through migrations runtime | Squash migration history. |
| `gogo createsuperuser` | available auth command shell | Create an admin user. |
| `gogo changepassword` | available auth command shell | Change a user password. |
| `gogo collectstatic` | available static command shell | Collect static files. |
| `gogo shell` | available | Start app shell context. |
| `gogo dbshell` | planned | Open database shell. |
| `gogo test` | planned | Run project tests. |
| `gogo worker` | available queue command shell | Run queue workers. |
| `gogo beat` | available queue command shell | Run beat scheduler. |
| `gogo inspect` | available queue command shell | Inspect workers. |
| `gogo queues` | available queue command shell | Inspect queues. |
| `gogo dumpdata` | available | Dump fixtures. |
| `gogo loaddata` | available | Load fixtures. |

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
```
