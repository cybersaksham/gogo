# CLI And Generator Rules

Use this rule for `cmd/gogo`, `internal/cli`, and generated project/app templates.

## CLI

- Commands must parse flags deterministically and return typed errors where possible.
- Commands must not leak secrets in output or dry runs.
- Root `gogo help` should reflect the actual command surface.
- CLI behavior changes need tests in `internal/cli` or `cmd/gogo`, depending on
  whether the change is command behavior or binary entrypoint behavior.
- When the installed `gogo` binary runs inside a generated project, project-aware
  commands must delegate to `go run manage.go <command>` so settings, routes,
  admin, app configs, model metadata, fixtures, and queue tasks are loaded.

## Generators

- `startproject` templates live under `internal/cli/templates/project`.
- `startapp` templates live under `internal/cli/templates/app`.
- Generated projects must compile downstream with only public framework imports.
- Update generated-project compile, functional, or end-to-end tests for template changes.
- Keep generated `.env.example` and `.gitignore` grouped by use case.

## Manage Entrypoint

- Generated `manage.go` is part of the client contract. Keep public runner
  delegation, templates, docs, and generated-project tests synchronized.
