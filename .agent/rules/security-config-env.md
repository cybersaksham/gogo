# Security, Config, And Environment Rules

Use this rule for settings, auth, sessions, CSRF, signing, deployment checks, or environment variables.

## Secrets

- Never commit `.env`, passwords, tokens, private keys, local databases, uploaded media, or machine-local paths.
- Keep `.gitignore` grouped and update it when adding generated or local-only artifacts.
- Do not print secrets in CLI output, dry runs, errors, logs, or tests.

## Environment Files

- Root `.env.example` documents framework env variables.
- `internal/cli/templates/project/env.example.tmpl` defines generated client project variables.
- When env variables change, update both files and `docs/reference/settings.md`.
- Blank values in `.env.example` mean required values unless a documented default exists.

## Runtime Validation

- Required settings must fail fast through `conf.Settings.Validate` or command startup.
- Production deploy checks must reject unsafe debug settings, weak secrets, wildcard hosts, insecure cookies, missing HTTPS confirmation, unreviewed admin paths, unapplied migrations, uncollected static files, unreachable services, and password reset without email.

## Auth And Session Safety

- Preserve constant-time comparisons for tokens and passwords.
- Preserve password hash compatibility fixtures.
- Keep session and CSRF cookies secure in production settings.

