# Production Security Operations

This guide describes the baseline security controls required before running a
Gogo project in production. The framework provides settings validation,
security middleware, CSRF protection, signed sessions, auth password hashing,
and upload validators. Deployment owners must still wire those controls into
the generated project and the edge stack.

## Production Settings

Set production values explicitly. Do not rely on development defaults.

| Setting | Required production value |
| --- | --- |
| `GOGO_ENV` | `production` |
| `GOGO_DEBUG` | `false` |
| `GOGO_SECRET_KEY` | Strong random secret from a secret manager |
| `GOGO_ALLOWED_HOSTS` | Exact public hostnames, never `*` |
| `DATABASE_URL` | Production database DSN from a secret manager |
| `GOGO_STATIC_ROOT` | Collected static file directory or mounted volume |
| `GOGO_MEDIA_ROOT` | Writable media directory outside the code tree |
| `GOGO_SESSION_COOKIE_NAME` | Environment-specific cookie name |
| `GOGO_CSRF_COOKIE_NAME` | Environment-specific cookie name |
| `GOGO_BROKER_URL` | Required when workers or beat are deployed |
| `GOGO_RESULT_BACKEND` | Required when task results are persisted |
| `GOGO_EMAIL_URL` | Required when password reset or outbound mail is enabled |

`conf.Settings.Validate` currently rejects missing `GOGO_SECRET_KEY`, missing
`DATABASE_URL`, invalid `GOGO_ENV`, invalid `GOGO_HTTP_ADDR`, and missing
`GOGO_ALLOWED_HOSTS` in production. Treat any validation failure as a failed
deployment.

## HTTPS And Proxy Security

Terminate TLS at a trusted load balancer or reverse proxy and forward traffic
to the app over a private network. Redirect HTTP to HTTPS before the request
reaches views. Configure HSTS only after the HTTPS deployment is stable.

Use `security.SecurityMiddleware` with:

- `SSLRedirect` enabled when the app is responsible for HTTPS redirects.
- `SecureProxyHeaderName` and `SecureProxyHeaderValue` set only for a trusted
  proxy header that outside clients cannot forge.
- `HSTSSeconds`, `HSTSIncludeSubdomains`, and `HSTSPreload` enabled only for
  domains that are fully HTTPS-ready.
- `ContentTypeNoSniff`, `ReferrerPolicy`, `CrossOriginOpenerPolicy`, and
  `FrameOptions` set to the strictest values the project can support.
- `AllowedHosts` copied from `GOGO_ALLOWED_HOSTS`.

If the edge layer performs redirects and headers, keep equivalent rules there
and still run host validation inside the app.

## Secure Cookies

Session cookies are configured through `sessions.CookieOptions`. Production
session cookies must use:

- `Secure: true`
- `HttpOnly: true`
- `SameSite: http.SameSiteLaxMode` or stricter unless a cross-site flow needs
  a narrower exception.
- A path scoped to `/` unless the project deliberately isolates an app area.
- A unique cookie name per environment to avoid local, staging, and production
  collisions.

CSRF cookies are configured through `security.CSRFOptions`. Production CSRF
cookies must use `SecureCookie: true`, `HttpOnly: true` from the framework
middleware, and `SameSite` compatible with the project's form flow.

Enable `DiagnoseSecureCookies` on `security.SecurityMiddleware` during
pre-production checks to detect response cookies missing the `Secure` flag.

## CSRF Trusted Origins

Unsafe methods must be protected with `security.NewCSRFProtection` or the
project's equivalent CSRF middleware. Configure `TrustedOrigins` with exact
scheme and host pairs, for example `https://admin.example.com`. Do not use
wildcards.

A trusted origin is only for known cross-origin POST flows controlled by the
same organization. It is not a CORS allowlist and must not include arbitrary
customer domains.

## Admin Path And Admin Access

The admin panel must be treated as a privileged application surface.

- Review the admin route before deployment and avoid exposing unused admin
  URLs.
- Require staff status and explicit model permissions for every admin view.
- Protect admin login with rate limiting at the edge.
- Use strong passwords and the built-in password validators.
- Place admin behind SSO, VPN, IP allowlists, or another access boundary when
  the deployment environment supports it.
- Keep admin audit logs for login, logout, object creation, object changes,
  object deletion, permission changes, and failed permission checks.

Never expose debug routes, profiling handlers, generated fixtures, or internal
task inspection endpoints through the public admin host.

## Password Hashing

Use the built-in auth helpers for all framework users:

- `auth.MakePassword` for new passwords.
- `auth.CheckPassword` for verification.
- `auth.ValidatePassword` before password creation, password change, and
  password reset confirmation.
- `auth.MustUpdatePasswordHash` to identify hashes that need rehashing after a
  parameter upgrade.

Do not store raw passwords, reversible encrypted passwords, or password hints.
Treat password reset tokens as secrets and invalidate them after password
changes.

## Sessions

Sessions depend on `GOGO_SECRET_KEY` and the selected session store. Use a
strong secret, rotate it through a planned invalidation window, and assume that
rotating it invalidates signed cookies, signed server-side session keys, and
password reset tokens.

Set reasonable expiry times for the application. Rotate session keys after
login and privilege changes. Flush sessions after logout, password change, user
deactivation, staff removal, or permission downgrade.

Use a server-side session store for sensitive or large session payloads. Signed
cookie sessions must only contain non-sensitive values that can be visible to
the browser.

## CORS Policy

Gogo does not require CORS for same-origin browser apps. If a project exposes
cross-origin APIs, implement a narrow CORS middleware or configure the edge
proxy with:

- Exact allowed origins.
- Exact allowed methods.
- Exact allowed headers.
- Credentials disabled unless the browser flow requires cookies.
- No `*` origin when credentials are enabled.
- Short preflight cache durations while a policy is changing.

CORS must not be used as an authentication or CSRF control.

## Rate Limiting

Apply rate limits at the edge or with project middleware for:

- Login.
- Password reset request and confirmation.
- Admin login.
- API token creation.
- Queue task enqueue endpoints.
- Expensive export, search, and reporting endpoints.

Use both IP and account identifiers where possible. Return generic errors for
authentication throttling so attackers cannot distinguish missing users from
invalid passwords.

## Uploaded Files

Uploaded files must be validated before use. Use the `files` upload handlers
and validators to enforce:

- Maximum request size.
- Maximum per-file size.
- Allowed content types.
- Allowed extensions.
- Image dimension limits when accepting images.

Store uploaded files outside the executable and source tree. Generate storage
names instead of trusting client filenames. Do not serve sensitive uploads
through public static handlers. Do not execute uploaded files, import them as
templates, or allow user-controlled paths to select files on disk.

For high-risk uploads, add malware scanning, quarantine, asynchronous review,
and private signed download URLs.

## Operational Checklist

Before production traffic is allowed:

- Run `make ci`.
- Run documentation verification with `make docs-verify`.
- Run vulnerability scanning in CI.
- Run `gogo check` with production environment variables.
- Confirm debug mode is disabled in the running process.
- Confirm host validation rejects unexpected hosts.
- Confirm HTTP redirects to HTTPS.
- Confirm session and CSRF cookies include `Secure`, `HttpOnly`, and expected
  `SameSite` attributes.
- Confirm admin access is limited to staff users with explicit permissions.
- Confirm backup, restore, and rollback procedures have been tested.
