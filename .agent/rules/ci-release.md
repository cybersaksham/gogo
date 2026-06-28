# CI And Release Rules

Use this rule for `.github/workflows`, release code, changelog, dependency policy, and distribution behavior.

## CI

- CI must protect formatting, vet, unit tests, race-focused tests, external integration jobs, generated project tests, examples, docs, dependency listing, and vulnerability scanning.
- Keep workflow permissions minimal.
- Avoid duplicating long shell logic when Makefile targets already express the behavior.

## Release

- Release runs on semver tags and manual dispatch.
- Release must verify before publishing.
- Release dry-run must validate tag, commit, build date, changelog section, artifacts, and notes.
- CLI artifacts must include version, commit, and build date linker flags.
- Checksums must be generated for artifacts.

## Changelog

- Every release tag must have a matching `CHANGELOG.md` section.
- Keep unreleased notes concise and user-facing.

## Dependencies

- Dependency changes need a clear reason.
- Run vulnerability checks in CI. Locally, report when `govulncheck` is unavailable.

