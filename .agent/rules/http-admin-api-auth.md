# HTTP, Admin, API, And Auth Rules

Use this rule for `http`, `admin`, `api`, `auth`, `sessions`, and request-facing behavior.

## HTTP

- Keep routing and reversing deterministic.
- Middleware must preserve request context and avoid leaking internal errors when debug is disabled.
- Security middleware must be explicit and tested.

## Admin

- Admin access must require staff permissions.
- Preserve permission checks for index, list, add, change, delete, actions, history, and autocomplete.
- Admin templates and static assets must remain embedded and overridable.
- Admin options need validation tests when combinations are constrained.

## API

- Serializer validation must return stable field errors.
- API responses must not expose panic details.
- ViewSets, filtering, pagination, throttling, permissions, uploads, versioning, and OpenAPI changes need focused tests.

## Auth

- Preserve Django-style users, groups, permissions, content types, sessions, password hashing, password reset tokens, and admin registration.
- Password and token comparisons must be safe.
- Auth model extension must not break framework-owned auth tables.

