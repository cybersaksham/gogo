# Client Project Template And Generators Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Finalize the generated client project experience so applications using Gogo get Django-like project structure, app structure, settings, routes, admin, API, migrations, queue, tests, deploy files, and safe environment defaults.

**Architecture:** CLI generators use embedded templates owned by `internal/cli/templates`. The generated project imports only public framework packages and is tested as a real downstream module.

**Tech Stack:** Embedded templates, Go modules, Makefile, Docker Compose, framework CLI.

---

## Files

- Create: `internal/cli/templates/project/go.mod.tmpl`
- Create: `internal/cli/templates/project/manage.go.tmpl`
- Create: `internal/cli/templates/project/gitignore.tmpl`
- Create: `internal/cli/templates/project/env.example.tmpl`
- Create: `internal/cli/templates/project/Makefile.tmpl`
- Create: `internal/cli/templates/project/README.md.tmpl`
- Create: `internal/cli/templates/project/settings/base.go.tmpl`
- Create: `internal/cli/templates/project/settings/local.go.tmpl`
- Create: `internal/cli/templates/project/settings/test.go.tmpl`
- Create: `internal/cli/templates/project/settings/production.go.tmpl`
- Create: `internal/cli/templates/project/app.go.tmpl`
- Create: `internal/cli/templates/project/urls.go.tmpl`
- Create: `internal/cli/templates/project/admin.go.tmpl`
- Create: `internal/cli/templates/project/middleware.go.tmpl`
- Create: `internal/cli/templates/project/queue.go.tmpl`
- Create: `internal/cli/templates/project/deploy/docker/Dockerfile.tmpl`
- Create: `internal/cli/templates/project/deploy/docker/docker-compose.yml.tmpl`
- Create: `internal/cli/templates/app/app.go.tmpl`
- Create: `internal/cli/templates/app/models.go.tmpl`
- Create: `internal/cli/templates/app/admin.go.tmpl`
- Create: `internal/cli/templates/app/urls.go.tmpl`
- Create: `internal/cli/templates/app/api.go.tmpl`
- Create: `internal/cli/templates/app/serializers.go.tmpl`
- Create: `internal/cli/templates/app/forms.go.tmpl`
- Create: `internal/cli/templates/app/services.go.tmpl`
- Create: `internal/cli/templates/app/tasks.go.tmpl`
- Create: `internal/cli/templates/app/permissions.go.tmpl`
- Create: `internal/cli/templates/app/tests.go.tmpl`
- Create: `internal/cli/templates/templates_test.go`

## Task 1: Finalize Project Template

- [ ] Create project templates listed in this plan.
- [ ] Generated project structure must be:
  - `manage.go`
  - `myproject/app.go`
  - `myproject/settings/base.go`
  - `myproject/settings/local.go`
  - `myproject/settings/test.go`
  - `myproject/settings/production.go`
  - `myproject/urls.go`
  - `myproject/admin.go`
  - `myproject/middleware.go`
  - `myproject/queue.go`
  - `apps/`
  - `templates/base.html`
  - `static/`
  - `media/`
  - `fixtures/`
  - `tests/integration/`
  - `deploy/docker/Dockerfile`
  - `deploy/docker/docker-compose.yml`
- [ ] `.gitignore` must group environment files, Go outputs, local databases, media, coverage, and editor files.
- [ ] `.env.example` must include every required setting with defaults where safe and blanks where required.
- [ ] Add template tests that generated files match expected paths.
- [ ] Run `go test ./internal/cli`.
- [ ] Commit with message `Finalize Project Template`.

## Task 2: Finalize App Template

- [ ] Create app templates listed in this plan.
- [ ] Generated app structure must be:
  - `app.go`
  - `models.go`
  - `admin.go`
  - `urls.go`
  - `api.go`
  - `serializers.go`
  - `forms.go`
  - `services.go`
  - `tasks.go`
  - `permissions.go`
  - `migrations/.keep`
  - `templates/<app_label>/.keep`
  - `static/<app_label>/.keep`
  - `tests/<app_label>_test.go`
- [ ] App template must register AppConfig, routes, admin, API, tasks, and migrations through public framework APIs.
- [ ] Add tests for generated app compile.
- [ ] Run `go test ./internal/cli`.
- [ ] Commit with message `Finalize App Template`.

## Task 3: Add Generated Project Compile Test

- [ ] Add test that runs `gogo startproject sampleproject` in a temp directory.
- [ ] Add test that runs `gogo startapp blog` inside generated project.
- [ ] Run `go mod tidy` for generated project using local replace directive.
- [ ] Run generated project tests.
- [ ] Assert generated project imports no `gogo/internal` package.
- [ ] Run `go test ./internal/cli`.
- [ ] Commit with message `Add Generated Project Compile Test`.

## Task 4: Add Generated Project Functional Test

- [ ] In a temp generated project, create a blog app with a sample model.
- [ ] Run `gogo makemigrations`.
- [ ] Run `gogo migrate`.
- [ ] Run `gogo createsuperuser --username admin --email admin@example.com --password <test-password>`.
- [ ] Start test server through injectable server runtime.
- [ ] Verify homepage, admin login, API route, and static file path.
- [ ] Run `go test -tags=integration ./internal/cli`.
- [ ] Commit with message `Add Generated Project Functional Test`.

## Task 5: Add Deployment Template Checks

- [ ] Dockerfile must use multi-stage build.
- [ ] Runtime image must run as non-root user.
- [ ] Compose file must include app, PostgreSQL, Redis, and optional RabbitMQ.
- [ ] Compose environment must use `.env`.
- [ ] Volumes must separate database, Redis, static, and media data.
- [ ] Add tests that generated Dockerfile and Compose include required services and do not contain secrets.
- [ ] Run `go test ./internal/cli`.
- [ ] Commit with message `Add Deployment Template Checks`.

## Task 6: Add Generator Documentation

- [ ] Document `startproject` and `startapp` flags.
- [ ] Document generated file responsibilities.
- [ ] Document environment variable rules.
- [ ] Document how to add apps, models, admin entries, API routes, and tasks.
- [ ] Run documentation verification.
- [ ] Commit with message `Document Project Generators`.

## Acceptance Checklist

- [ ] Generated projects compile and run tests.
- [ ] Generated apps compile and register with app registry.
- [ ] Generated project `.env.example` is synced with settings.
- [ ] Generated `.gitignore` excludes local-only files.
- [ ] Docker templates are production-safe by default.
- [ ] Generated code imports only public framework packages.

