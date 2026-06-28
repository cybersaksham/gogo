# Repository Structure Rules

Use this rule when adding, moving, or reviewing files.

## Package Boundaries

- Public framework APIs belong in top-level packages such as `app`, `http`, `models`, `orm`, `migrations`, `auth`, `admin`, `api`, `forms`, `templates`, `files`, `static`, `queue`, `cache`, `email`, `messages`, `security`, `sessions`, `signals`, `i18n`, `health`, `observability`, `contrib`, and `testing`.
- Internal-only implementation belongs under `internal/`.
- CLI implementation belongs under `internal/cli`; the executable entrypoint is `cmd/gogo/main.go`.
- Generated downstream project templates belong under `internal/cli/templates/project`.
- Generated downstream app templates belong under `internal/cli/templates/app`.
- Release-only code belongs under `internal/release`.
- Code-maintainer documentation belongs under `docs/code/architecture`, `docs/code/reference`, `docs/code/tutorials`, `docs/code/operations`, or `docs/code/generated`.
- Public static documentation belongs under `docs/public` and must not replace the maintainer docs under `docs/code`.

## Generated Project Contract

- Generated projects must import only public packages.
- Generated projects must compile as independent downstream modules.
- Keep generated code deterministic.
- When templates change, update generated-project tests.

## Tests

- Public packages need package tests unless the package only holds generated migration declarations.
- Put behavioral tests beside the package being changed.
- Put downstream generated-project checks in `internal/cli/*generated_project*_test.go`.
