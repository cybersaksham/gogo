# Public API And Compatibility Rules

Use this rule when changing exported types, functions, generated project behavior, migrations, auth hashes, sessions, queues, or settings.

## Public API

- Treat top-level packages as client-facing API.
- Avoid breaking exported names, struct fields, method signatures, or serialized formats without an explicit compatibility plan.
- Do not expose internal packages to generated projects.

## Compatibility Tests

- Update `internal/compatibility` fixtures when changing supported serialized formats.
- Preserve compatibility for:
  - migration manifests and dependencies
  - auth password hashes
  - signed sessions/cookies
  - queue envelopes
  - generated project contracts
  - settings env names

## Versioning

- Release tags use `vMAJOR.MINOR.PATCH`.
- Changelog sections must exist for release tags.
- Document deprecations before removals.

