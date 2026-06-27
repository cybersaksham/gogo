# App Project Lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Django-style project bootstrapping, installed apps, app configs, app registry, lifecycle hooks, discovery, and management command extension points.

**Architecture:** Public package `app` defines app contracts and registry behavior. Private package `internal/inspect` handles safe reflection and generated manifests so runtime code can discover models, admin registrations, routes, tasks, and commands without client projects importing internals.

**Tech Stack:** Go interfaces, generics where useful, `context`, `sync`, deterministic registries, `testing/fstest`.

---

## Files

- Create: `app/config.go`
- Create: `app/registry.go`
- Create: `app/lifecycle.go`
- Create: `app/discovery.go`
- Create: `app/errors.go`
- Create: `app/command.go`
- Create: `app/testing.go`
- Create: `internal/inspect/package.go`
- Create: `internal/inspect/manifest.go`
- Create: `internal/cli/startproject.go`
- Create: `internal/cli/startapp.go`
- Create: `internal/cli/shell.go`
- Create: `app/registry_test.go`
- Create: `app/lifecycle_test.go`
- Create: `internal/inspect/manifest_test.go`
- Create: `internal/cli/startproject_test.go`
- Create: `internal/cli/startapp_test.go`

## Task 1: Define App Config Contract

- [ ] Create `app/config.go`.
- [ ] Define `Config` with methods:
  - `Name() string`
  - `Label() string`
  - `Path() string`
  - `VerboseName() string`
  - `Dependencies() []string`
  - `Ready(context.Context, *Registry) error`
  - `Shutdown(context.Context) error`
- [ ] Add `BaseConfig` struct for client apps to embed.
- [ ] Add validation for import-like names, label uniqueness, non-empty path, and dependency names.
- [ ] Add tests for valid config, invalid names, invalid labels, duplicate labels, and missing path.
- [ ] Run `go test ./app`.
- [ ] Commit with message `Add App Config Contract`.

## Task 2: Implement App Registry

- [ ] Create `app/registry.go`.
- [ ] Implement deterministic registration.
- [ ] Reject duplicate app names and labels.
- [ ] Resolve dependency order with cycle detection.
- [ ] Expose:
  - `Register(Config) error`
  - `Apps() []Config`
  - `Get(nameOrLabel string) (Config, bool)`
  - `MustGet(nameOrLabel string) Config`
  - `Labels() []string`
  - `Ready(context.Context) error`
  - `Shutdown(context.Context) error`
- [ ] Add typed errors:
  - `ErrDuplicateApp`
  - `ErrInvalidApp`
  - `ErrMissingDependency`
  - `ErrDependencyCycle`
  - `ErrRegistryReady`
- [ ] Add tests for registration order, dependency order, cycles, lookups, and duplicate labels.
- [ ] Run `go test ./app`.
- [ ] Commit with message `Add App Registry`.

## Task 3: Implement Lifecycle Hooks

- [ ] Create `app/lifecycle.go`.
- [ ] Ensure `Ready` runs each app exactly once in dependency order.
- [ ] Ensure `Shutdown` runs in reverse dependency order.
- [ ] Ensure failed `Ready` prevents partial ready state from being treated as successful.
- [ ] Add test app configs that record call order.
- [ ] Add tests for successful boot, ready failure, shutdown after partial ready, and context cancellation.
- [ ] Run `go test ./app`.
- [ ] Commit with message `Add App Lifecycle Hooks`.

## Task 4: Add Discovery Manifest

- [ ] Create `internal/inspect/manifest.go`.
- [ ] Define generated manifest structures:
  - `AppManifest`
  - `ModelManifest`
  - `AdminManifest`
  - `RouteManifest`
  - `TaskManifest`
  - `CommandManifest`
  - `MigrationManifest`
- [ ] Store enough metadata to discover app-owned resources without scanning arbitrary source at runtime.
- [ ] Add JSON marshal/unmarshal tests.
- [ ] Add deterministic sort tests.
- [ ] Run `go test ./internal/inspect`.
- [ ] Commit with message `Add App Discovery Manifest`.

## Task 5: Add Public Discovery API

- [ ] Create `app/discovery.go`.
- [ ] Add resource registries to `Registry`:
  - Models
  - Admin registrations
  - Routes
  - API routes
  - Forms
  - Templates
  - Static asset roots
  - Queue tasks
  - Management commands
  - Migrations
- [ ] Expose read-only discovery methods that return copied slices.
- [ ] Add tests proving callers cannot mutate registry internals.
- [ ] Run `go test ./app`.
- [ ] Commit with message `Add App Resource Discovery`.

## Task 6: Add Management Command Extension Points

- [ ] Create `app/command.go`.
- [ ] Define `ManagementCommand` interface with:
  - `Name() string`
  - `Summary() string`
  - `Run(context.Context, []string) error`
- [ ] Allow apps to register custom commands.
- [ ] Merge app commands into CLI registry after app boot.
- [ ] Reject command names that collide with built-in commands unless the command is explicitly namespaced.
- [ ] Add tests for custom command registration, namespacing, duplicate detection, and execution.
- [ ] Run `go test ./app ./internal/cli`.
- [ ] Commit with message `Add App Management Commands`.

## Task 7: Implement Startproject Generator

- [ ] Complete `internal/cli/startproject.go`.
- [ ] Generate:
  - `go.mod`
  - `manage.go`
  - `.gitignore`
  - `.env.example`
  - `Makefile`
  - `README.md`
  - `myproject/app.go`
  - `myproject/settings/base.go`
  - `myproject/settings/local.go`
  - `myproject/settings/test.go`
  - `myproject/settings/production.go`
  - `myproject/urls.go`
  - `myproject/admin.go`
  - `myproject/middleware.go`
  - `myproject/queue.go`
  - `apps/.keep`
  - `templates/base.html`
  - `static/.keep`
  - `fixtures/.keep`
  - `tests/integration/.keep`
  - `deploy/docker/Dockerfile`
  - `deploy/docker/docker-compose.yml`
- [ ] Refuse to overwrite non-empty directories unless `--force` is provided.
- [ ] Keep generated `.env.example` grouped and synced with framework requirements.
- [ ] Add filesystem tests with `testing/fstest` and temporary directories.
- [ ] Run `go test ./internal/cli`.
- [ ] Commit with message `Add Project Generator`.

## Task 8: Implement Startapp Generator

- [ ] Complete `internal/cli/startapp.go`.
- [ ] Generate app folder:
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
  - `tests/.keep`
- [ ] Validate app name, app label, and target path.
- [ ] Refuse to overwrite existing app directories unless `--force` is provided.
- [ ] Add tests for valid generation, invalid names, existing folder protection, and forced overwrite.
- [ ] Run `go test ./internal/cli`.
- [ ] Commit with message `Add App Generator`.

## Task 9: Add Shell Command Skeleton

- [ ] Create `internal/cli/shell.go`.
- [ ] Implement `gogo shell` as an interactive Go-aware environment launcher where possible and a clear guidance command otherwise.
- [ ] Load settings and app registry before launching.
- [ ] Print registered apps, database status, and useful import paths.
- [ ] Add tests for non-interactive `--command` execution.
- [ ] Run `go test ./internal/cli`.
- [ ] Commit with message `Add Shell Command`.

## Acceptance Checklist

- [ ] Generated project follows the client project folder plan.
- [ ] Generated app follows Django-style app boundaries.
- [ ] App registry is deterministic and safe under concurrent reads after boot.
- [ ] Dependencies are ordered and cycles are rejected.
- [ ] App `Ready` and `Shutdown` hooks are tested.
- [ ] Custom management commands can be registered by apps.

