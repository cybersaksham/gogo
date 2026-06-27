# Foundation CLI And Config Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create the base Go module, command-line entrypoint, configuration loader, environment validation, logging, error conventions, and repository hygiene needed by every phase.

**Architecture:** The CLI dispatches to command packages under `internal/cli` while public configuration APIs live in `conf`. The root package `gogo` exposes bootstrapping helpers for framework users without importing command internals.

**Tech Stack:** Go, `flag` or Cobra-style command dispatcher, `log/slog`, `context`, `os`, `errors`, `testing`, `fstest`.

---

## Files

- Create: `go.mod`
- Create: `go.sum`
- Create: `gogo.go`
- Create: `cmd/gogo/main.go`
- Create: `internal/version/version.go`
- Create: `internal/cli/root.go`
- Create: `internal/cli/command.go`
- Create: `internal/cli/errors.go`
- Create: `internal/cli/check.go`
- Create: `internal/cli/runserver.go`
- Create: `internal/cli/startproject.go`
- Create: `internal/cli/startapp.go`
- Create: `internal/cli/migrations.go`
- Create: `internal/cli/auth.go`
- Create: `internal/cli/queue.go`
- Create: `internal/cli/data.go`
- Create: `internal/cli/test.go`
- Create: `conf/settings.go`
- Create: `conf/env.go`
- Create: `conf/defaults.go`
- Create: `conf/errors.go`
- Create: `conf/settings_test.go`
- Create: `internal/cli/root_test.go`
- Create: `.gitignore`
- Create: `.env.example`
- Create: `Makefile`
- Create: `README.md`

## Task 1: Initialize Module And Repository Hygiene

- [ ] Create `go.mod` with module path `github.com/cybersaksham/gogo` and the supported Go version.
- [ ] Create `.gitignore` with grouped sections:
  - Go build output: `bin/`, `dist/`, `*.test`, `*.out`
  - Coverage: `coverage/`, `coverage.out`
  - Local environment: `.env`, `.env.*.local`
  - Local databases: `*.sqlite`, `*.sqlite3`, `*.db`
  - Uploads and generated media: `media/`, `uploads/`
  - Editor files: `.idea/`, `.vscode/`, `*.swp`
- [ ] Create `.env.example` with grouped sections:
  - Framework: `GOGO_ENV=development`, `GOGO_SECRET_KEY=`
  - Database: `DATABASE_URL=`
  - Server: `GOGO_HTTP_ADDR=:8000`
  - Queue: `GOGO_BROKER_URL=`, `GOGO_RESULT_BACKEND=`
  - Security: `GOGO_ALLOWED_HOSTS=localhost,127.0.0.1`
- [ ] Create `Makefile` targets:
  - `test`: `go test ./...`
  - `race`: `go test -race ./...`
  - `lint`: `go vet ./...`
  - `build`: `go build -o bin/gogo ./cmd/gogo`
  - `check`: `go test ./... && go vet ./...`
- [ ] Run `go test ./...`.
- [ ] Commit with message `Initialize Framework Repository`.

## Task 2: Add Version Package

- [ ] Create `internal/version/version.go`.
- [ ] Define `Version`, `Commit`, and `BuildDate` variables that can be overridden by linker flags.
- [ ] Add `Info() string` returning a stable human-readable version string.
- [ ] Add tests in `internal/version/version_test.go` for default and overridden values.
- [ ] Run `go test ./internal/version`.
- [ ] Commit with message `Add Version Metadata`.

## Task 3: Add CLI Command Contract

- [ ] Create `internal/cli/command.go`.
- [ ] Define `type Command interface { Name() string; Summary() string; Run(context.Context, []string) error }`.
- [ ] Define `type Registry` with deterministic registration order and duplicate-name detection.
- [ ] Create `internal/cli/errors.go`.
- [ ] Add typed errors:
  - `ErrUnknownCommand`
  - `ErrDuplicateCommand`
  - `ErrInvalidArguments`
  - `ErrCommandFailed`
- [ ] Add tests for command lookup, duplicate registration, unknown command, and help ordering.
- [ ] Run `go test ./internal/cli`.
- [ ] Commit with message `Add CLI Command Registry`.

## Task 4: Add Root CLI

- [ ] Create `cmd/gogo/main.go`.
- [ ] Create `internal/cli/root.go`.
- [ ] Support root commands:
  - `help`
  - `version`
  - `check`
  - `runserver`
  - `startproject`
  - `startapp`
  - `makemigrations`
  - `migrate`
  - `showmigrations`
  - `sqlmigrate`
  - `squashmigrations`
  - `createsuperuser`
  - `changepassword`
  - `collectstatic`
  - `shell`
  - `dbshell`
  - `test`
  - `worker`
  - `beat`
  - `inspect`
  - `queues`
  - `dumpdata`
  - `loaddata`
- [ ] Commands that depend on later phases must return a typed unavailable error with the target phase name.
- [ ] Add tests for `gogo help`, `gogo version`, and unavailable command messages.
- [ ] Run `go test ./cmd/gogo ./internal/cli`.
- [ ] Commit with message `Add Root CLI Entrypoint`.

## Task 5: Add Settings Model

- [ ] Create `conf/settings.go`.
- [ ] Define `Settings` with:
  - `Env string`
  - `SecretKey string`
  - `Debug bool`
  - `AllowedHosts []string`
  - `HTTPAddr string`
  - `DatabaseURL string`
  - `InstalledApps []string`
  - `Middleware []string`
  - `RootURLConf string`
  - `StaticURL string`
  - `StaticRoot string`
  - `MediaURL string`
  - `MediaRoot string`
  - `TemplateDirs []string`
  - `DefaultAutoField string`
  - `TimeZone string`
  - `LanguageCode string`
  - `SessionCookieName string`
  - `CSRFCookieName string`
  - `BrokerURL string`
  - `ResultBackend string`
  - `CacheURL string`
  - `EmailURL string`
- [ ] Add `Validate() error` that fails for missing `SecretKey`, missing `DatabaseURL`, invalid `Env`, empty `AllowedHosts` in production, and invalid HTTP address.
- [ ] Add tests for valid development, valid production, missing required values, and invalid enum values.
- [ ] Run `go test ./conf`.
- [ ] Commit with message `Add Framework Settings Model`.

## Task 6: Add Environment Loader

- [ ] Create `conf/env.go`.
- [ ] Implement `LoadEnvFile(path string) (map[string]string, error)` for simple `KEY=VALUE` files with comments and quoted values.
- [ ] Implement `LoadFromEnv() (Settings, error)` using process environment plus `.env` when present.
- [ ] Ensure `.env.example` remains synced with every environment variable read by `conf`.
- [ ] Add tests using `t.Setenv` and temporary files.
- [ ] Run `go test ./conf`.
- [ ] Commit with message `Add Environment Settings Loader`.

## Task 7: Add Defaults

- [ ] Create `conf/defaults.go`.
- [ ] Provide safe defaults:
  - `Env=development`
  - `Debug=true` only for development
  - `HTTPAddr=:8000`
  - `StaticURL=/static/`
  - `MediaURL=/media/`
  - `DefaultAutoField=BigAutoField`
  - `TimeZone=UTC`
  - `LanguageCode=en-us`
  - `SessionCookieName=gogo_sessionid`
  - `CSRFCookieName=gogo_csrftoken`
- [ ] Ensure required values remain required after defaults.
- [ ] Add tests proving `SecretKey` and `DatabaseURL` are still required.
- [ ] Run `go test ./conf`.
- [ ] Commit with message `Add Configuration Defaults`.

## Task 8: Add Check Command

- [ ] Create `internal/cli/check.go`.
- [ ] `gogo check` must load settings, validate required values, validate installed app names are importable once app registry exists, and print structured diagnostics.
- [ ] In this phase, implement config checks and return unavailable diagnostics for app checks.
- [ ] Add tests for passing and failing config checks.
- [ ] Run `go test ./internal/cli ./conf`.
- [ ] Commit with message `Add Configuration Check Command`.

## Task 9: Add Runserver Skeleton

- [ ] Create `internal/cli/runserver.go`.
- [ ] Parse `--addr`, `--settings`, and `--reload=false`.
- [ ] Load settings and pass the address to an injectable server function.
- [ ] Return a clear error when the HTTP phase has not wired the server yet.
- [ ] Add tests that the command resolves the final address from flag, env, and default precedence.
- [ ] Run `go test ./internal/cli`.
- [ ] Commit with message `Add Runserver Command Skeleton`.

## Task 10: Add Documentation

- [ ] Create `README.md` with project purpose, planned phases, current status, CLI command list, and security expectations.
- [ ] Document every environment variable from `.env.example`.
- [ ] Document that generated client projects must keep `.env` out of Git and commit `.env.example`.
- [ ] Run `go test ./...`.
- [ ] Commit with message `Document Foundation Usage`.

## Acceptance Checklist

- [ ] `go test ./...` passes.
- [ ] `go build -o bin/gogo ./cmd/gogo` succeeds.
- [ ] `bin/gogo help` lists all planned commands in stable order.
- [ ] Missing required environment variables fail with actionable errors.
- [ ] `.gitignore` and `.env.example` are grouped and synced.
- [ ] No command silently succeeds when its backing feature is not implemented.
