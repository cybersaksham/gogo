# Audit And Review Rules

Use this rule for final passes, reviews, and production-readiness checks.

## Audit Checklist

- Worktree is clean or all changes are intentionally scoped.
- No `.env`, local database, key, token, coverage dump, upload, or media artifact is committed.
- No stale placeholder language remains in shipped docs or code.
- Generated projects do not import internal packages.
- Env examples are synchronized.
- Public docs match current CLI and API behavior.
- Security-sensitive changes have tests.
- Compatibility fixtures are updated for serialized format changes.
- Release and CI workflows still match Makefile targets.

## Review Stance

- Lead with concrete bugs, risks, regressions, and missing tests.
- Cite exact files and lines when reporting issues.
- Do not report style preferences as defects unless they affect maintainability, security, or usability.

## Final Verification

Use the smallest meaningful verification set for the change, then broaden for release-impacting changes.

```bash
go test ./...
go test -tags=integration ./internal/cli
make docs-verify
```

For release-impacting changes:

```bash
make ci
go test -tags=integration ./...
go test -race ./queue/... ./orm/... ./http/...
```

