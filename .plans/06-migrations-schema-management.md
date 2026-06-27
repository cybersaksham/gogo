# Migrations Schema Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Django-style migration generation, migration files, dependency graph, schema operations, data migrations, SQL rendering, migration execution, rollback, fake apply, squashing, and migration commands.

**Architecture:** Public package `migrations` defines operations and executor APIs. Private package `internal/schema` compares model state with database state and renders dialect-aware schema SQL.

**Tech Stack:** Go migration structs, deterministic file writer, schema graph, `database/sql`, ORM dialects, integration tests against SQLite and PostgreSQL.

---

## Files

- Create: `migrations/migration.go`
- Create: `migrations/operation.go`
- Create: `migrations/state.go`
- Create: `migrations/graph.go`
- Create: `migrations/loader.go`
- Create: `migrations/writer.go`
- Create: `migrations/autodetector.go`
- Create: `migrations/executor.go`
- Create: `migrations/recorder.go`
- Create: `migrations/errors.go`
- Create: `migrations/operations/model.go`
- Create: `migrations/operations/field.go`
- Create: `migrations/operations/index.go`
- Create: `migrations/operations/constraint.go`
- Create: `migrations/operations/sql.go`
- Create: `migrations/operations/data.go`
- Create: `internal/schema/introspection.go`
- Create: `internal/schema/editor.go`
- Create: `internal/schema/diff.go`
- Modify: `internal/cli/migrations.go`

## Task 1: Define Migration File Contract

- [ ] Create `migrations/migration.go`.
- [ ] Define `Migration` with:
  - App label
  - Name
  - Dependencies
  - Replaces
  - Operations
  - Atomic flag
  - RunBefore dependencies
- [ ] Define migration naming convention `0001_initial`, `0002_<slug>`.
- [ ] Add validation for missing app label, invalid name, duplicate dependencies, and empty operation list.
- [ ] Add tests for validation and deterministic identity.
- [ ] Run `go test ./migrations`.
- [ ] Commit with message `Add Migration Contract`.

## Task 2: Define Project State

- [ ] Create `migrations/state.go`.
- [ ] Represent historical app/model/field/index/constraint state without depending on live Go types.
- [ ] Support state clone and mutation.
- [ ] Support rendering state from model registry.
- [ ] Add tests for clone immutability and model registry conversion.
- [ ] Run `go test ./migrations`.
- [ ] Commit with message `Add Migration Project State`.

## Task 3: Define Operation Interface

- [ ] Create `migrations/operation.go`.
- [ ] Define operation methods:
  - `Name() string`
  - `StateForwards(*ProjectState) error`
  - `DatabaseForwards(context.Context, SchemaEditor) error`
  - `DatabaseBackwards(context.Context, SchemaEditor) error`
  - `Describe() string`
  - `Reversible() bool`
  - `ReferencesModel(appLabel, modelName string) bool`
  - `ReferencesField(appLabel, modelName, fieldName string) bool`
- [ ] Add tests with fake operations.
- [ ] Run `go test ./migrations`.
- [ ] Commit with message `Add Migration Operation Interface`.

## Task 4: Implement Model Operations

- [ ] Create `migrations/operations/model.go`.
- [ ] Implement:
  - `CreateModel`
  - `DeleteModel`
  - `RenameModel`
  - `AlterModelTable`
  - `AlterModelTableComment`
  - `AlterModelOptions`
  - `AlterModelManagers`
  - `AlterOrderWithRespectTo`
  - `AlterTogether` compatibility operations for legacy unique-together and index-together metadata
- [ ] Add state tests and SQL rendering tests.
- [ ] Run `go test ./migrations/...`.
- [ ] Commit with message `Add Model Migration Operations`.

## Task 5: Implement Field Operations

- [ ] Create `migrations/operations/field.go`.
- [ ] Implement:
  - `AddField`
  - `RemoveField`
  - `AlterField`
  - `RenameField`
- [ ] Handle defaults, nullability changes, column type changes, relation changes, and index changes.
- [ ] Require explicit data handling for unsafe non-null additions.
- [ ] Add tests for every operation and unsafe migration detection.
- [ ] Run `go test ./migrations/...`.
- [ ] Commit with message `Add Field Migration Operations`.

## Task 6: Implement Index And Constraint Operations

- [ ] Create `migrations/operations/index.go`.
- [ ] Create `migrations/operations/constraint.go`.
- [ ] Implement:
  - `AddIndex`
  - `RemoveIndex`
  - `RenameIndex`
  - `AddConstraint`
  - `RemoveConstraint`
- [ ] Support unique, check, exclusion metadata, partial, functional, covering, operator class, and deferrable constraints.
- [ ] Add tests for state mutation and SQL rendering.
- [ ] Run `go test ./migrations/...`.
- [ ] Commit with message `Add Index And Constraint Migration Operations`.

## Task 7: Implement SQL And Data Operations

- [ ] Create `migrations/operations/sql.go`.
- [ ] Create `migrations/operations/data.go`.
- [ ] Implement:
  - `RunSQL`
  - `RunPythonEquivalent` named `RunGo`
  - `SeparateDatabaseAndState`
- [ ] Add operation optimizer metadata including reduces-to-SQL, elidable, category, and operation reduction hooks.
- [ ] Require reversible data migrations to define reverse functions or explicitly mark irreversible.
- [ ] Add tests for reversible SQL, irreversible SQL, state-only operations, DB-only operations, and data operation execution.
- [ ] Run `go test ./migrations/...`.
- [ ] Commit with message `Add SQL And Data Migration Operations`.

## Task 8: Implement Migration Graph

- [ ] Create `migrations/graph.go`.
- [ ] Build dependency graph across apps.
- [ ] Detect cycles, missing dependencies, duplicate nodes, and conflicting leaf migrations.
- [ ] Compute forwards plan and backwards plan.
- [ ] Support replacements for squashed migrations.
- [ ] Add tests for graph ordering, cycles, missing dependencies, conflicts, and squashed replacements.
- [ ] Run `go test ./migrations`.
- [ ] Commit with message `Add Migration Graph`.

## Task 9: Implement Loader And Writer

- [ ] Create `migrations/loader.go`.
- [ ] Create `migrations/writer.go`.
- [ ] Load migration manifests from app migration packages.
- [ ] Write deterministic Go migration files.
- [ ] Preserve imports, operation ordering, and migration names.
- [ ] Add tests using temporary app folders.
- [ ] Run `go test ./migrations`.
- [ ] Commit with message `Add Migration Loader And Writer`.

## Task 10: Implement Schema Editor

- [ ] Create `internal/schema/editor.go`.
- [ ] Render SQL for:
  - Create table
  - Drop table
  - Rename table
  - Add column
  - Drop column
  - Alter column type
  - Alter null
  - Alter default
  - Rename column
  - Add index
  - Drop index
  - Rename index
  - Add constraint
  - Drop constraint
  - Create many-to-many table
  - Drop many-to-many table
- [ ] Add PostgreSQL and SQLite golden SQL tests.
- [ ] Run `go test ./internal/schema`.
- [ ] Commit with message `Add Schema Editor`.

## Task 11: Implement Autodetector

- [ ] Create `migrations/autodetector.go`.
- [ ] Compare previous migration state with current model registry.
- [ ] Detect:
  - New model
  - Deleted model
  - Renamed model
  - Model option change
  - New field
  - Removed field
  - Renamed field
  - Altered field
  - New index
  - Removed index
  - Renamed index
  - New constraint
  - Removed constraint
  - Many-to-many through table change
- [ ] Add interactive question hooks for rename detection that CLI can answer with flags.
- [ ] Preserve manual migration operations when merging autodetected operations with existing empty migrations.
- [ ] Add tests for every detected change.
- [ ] Run `go test ./migrations`.
- [ ] Commit with message `Add Migration Autodetector`.

## Task 12: Implement Recorder

- [ ] Create `migrations/recorder.go`.
- [ ] Create migration history table `gogo_migrations`.
- [ ] Store app, name, applied timestamp, checksum, and executor version.
- [ ] Support applied lookup, record applied, record unapplied, and history consistency checks.
- [ ] Add tests against SQLite and PostgreSQL integration tag.
- [ ] Run `go test ./migrations`.
- [ ] Commit with message `Add Migration Recorder`.

## Task 13: Implement Executor

- [ ] Create `migrations/executor.go`.
- [ ] Apply forwards plan in dependency order.
- [ ] Roll back backwards plan in reverse order.
- [ ] Support atomic and non-atomic migrations.
- [ ] Support fake apply and fake unapply.
- [ ] Stop on irreversible reverse operations with a clear error.
- [ ] Add tests for apply, rollback, fake apply, non-atomic behavior, partial failure rollback, and irreversible reverse.
- [ ] Run `go test ./migrations`.
- [ ] Commit with message `Add Migration Executor`.

## Task 14: Wire Migration CLI Commands

- [ ] Modify `internal/cli/migrations.go`.
- [ ] Implement:
  - `makemigrations`
  - `migrate`
  - `showmigrations`
  - `sqlmigrate`
  - `squashmigrations`
  - `migrate --prune`
  - `optimizemigration`
- [ ] Support flags:
  - `--app`
  - `--name`
  - `--empty`
  - `--check`
  - `--dry-run`
  - `--database`
  - `--fake`
  - `--fake-initial`
  - `--plan`
  - `--verbosity`
  - `--merge`
  - `--noinput`
  - `--prune`
- [ ] Add CLI tests with fake apps and temporary migration files.
- [ ] Run `go test ./internal/cli ./migrations`.
- [ ] Commit with message `Wire Migration Commands`.

## Task 15: Add Migration Safety Checks

- [ ] Add checks for destructive operations:
  - Dropping tables
  - Dropping columns
  - Type narrowing
  - Adding non-null column without default
  - Removing unique constraints
  - Renaming fields with ambiguous data movement
- [ ] Require explicit confirmation flags in non-interactive mode.
- [ ] Add tests for each safety check.
- [ ] Run `go test ./migrations ./internal/cli`.
- [ ] Commit with message `Add Migration Safety Checks`.

## Acceptance Checklist

- [ ] `makemigrations` writes deterministic migration files.
- [ ] `migrate` applies and rolls back migrations.
- [ ] `showmigrations` reflects applied database state.
- [ ] `sqlmigrate` renders dialect-specific SQL.
- [ ] `squashmigrations` preserves replacement metadata.
- [ ] Migration optimization, pruning, and conflict merging are planned and tested.
- [ ] Unsafe schema changes require explicit confirmation.
- [ ] Multi-database migration routing is honored.
