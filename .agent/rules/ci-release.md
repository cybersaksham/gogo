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

## Manual Release Publishing

Only run this workflow when the user explicitly asks to create, publish, cut, or ship a release. Do not publish releases, create release tags, or update release changelog sections automatically during normal implementation, review, cleanup, or CI work.

When asked to publish a release:

1. Confirm the current branch is the intended release branch and do not switch branches unless the user explicitly requests it.
2. Ensure the worktree is clean, except for release files you are about to update. If unrelated changes exist, stop and ask how to proceed.
3. Refresh release state from the remote with tags and releases before deciding the version. Prefer GitHub Releases as the source of truth, then semver Git tags if releases are unavailable.
4. Identify the previous release tag. If no prior release exists, treat the new release as the initial release and choose the first appropriate `v0.x.0` tag.
5. Inspect every change since the previous release:
   - Review commits between the previous tag and `HEAD`.
   - Review changed files and user-facing behavior.
   - Review existing `CHANGELOG.md` `Unreleased` notes.
   - If GitHub is available, review merged PR titles/descriptions included in the range.
6. Decide the next version from the actual changes:
   - Use a major version bump for breaking public APIs, incompatible generated project structure, incompatible migration formats, incompatible auth/session/permission behavior, incompatible queue contracts, or removed features after `v1.0.0`.
   - Use a minor version bump for backward-compatible framework features, new packages, new CLI commands, new admin/auth/API/ORM/queue capabilities, or compatible generated project improvements.
   - Use a patch version bump for backward-compatible bug fixes, security fixes, hardening, documentation corrections, CI fixes, and small compatibility repairs.
   - Before `v1.0.0`, use `v0.MINOR.0` for breaking or feature releases and `v0.MINOR.PATCH` for fixes to the current minor line.
   - If there are no meaningful changes since the previous release, report that no release is needed unless the user explicitly wants a rebuild/republication.
7. Update `CHANGELOG.md` before tagging:
   - Keep a fresh `## Unreleased` section at the top.
   - Add `## vX.Y.Z - YYYY-MM-DD` for the new version.
   - Move relevant unreleased notes into the new section.
   - Add missing user-facing changes discovered from commits or PRs.
   - Include clear subsections when relevant: `Added`, `Changed`, `Fixed`, `Security`, `Deprecated`, `Removed`, `Breaking`, and `Migration Notes`.
8. Build the release notes from the changelog section and the inspected changes. The release description must include:
   - The previous tag and new tag.
   - A concise summary of what changed.
   - Breaking changes and migration steps, if any.
   - Security fixes, if any.
   - Verification commands that passed.
   - Artifact/checksum information when publishing binaries.
9. Run verification before creating or pushing the tag:
   - `make ci`
   - `go test -tags=integration ./...` when release risk touches integrations beyond the default CI target.
   - `go run ./internal/release/cmd/dryrun --tag vX.Y.Z --commit "$(git rev-parse HEAD)" --build-date "$(date -u +'%Y-%m-%dT%H:%M:%SZ')" --changelog CHANGELOG.md --notes-out release-notes.md`
10. Commit the changelog update using the repository commit rules.
11. Create an annotated semver tag only after verification passes.
12. Push the changelog commit and tag only when the user has asked to publish the release.
13. Let the GitHub release workflow publish the release when the tag is pushed, or manually trigger the release workflow for the chosen tag if the user requested that path.
14. After publishing, verify the GitHub release exists, release notes match the changelog, artifacts are attached, checksums are present, and the tag resolves for Go users.
15. Report the final version, release URL, verification results, and any follow-up risks.

Never rewrite, delete, or move an already-published release tag unless the user explicitly asks for that destructive operation and understands the downstream Go module/proxy impact.

## Changelog

- Every release tag must have a matching `CHANGELOG.md` section.
- Keep unreleased notes concise and user-facing.

## Dependencies

- Dependency changes need a clear reason.
- Run vulnerability checks in CI. Locally, report when `govulncheck` is unavailable.
