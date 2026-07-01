# Migrations Reference

Migrations are Go values that describe schema and data changes. They update project state, execute database operations, and record applied history.

## Public Types

| Area | Types |
| --- | --- |
| Migration identity | `migrations.Dependency`, `migrations.Migration` |
| Operations | `migrations.Operation`, `migrations.SchemaEditor`, `migrations.ManifestOperation` |
| State | `migrations.ProjectState`, `ModelState`, `FieldState`, `IndexState`, `ConstraintState` |
| Graph | `migrations.Graph` |
| Autodetection | `migrations.Autodetector`, `DetectedChange`, `ChangeType`, `RenameQuestioner` |
| Loading and writing | `migrations.Loader`, `migrations.Writer` |
| Recording and execution | `migrations.Recorder`, `AppliedMigration`, `Executor`, `ExecutorOptions` |
| Safety | `migrations.SafetyOptions`, `SafetyCheck`, `SafetyCheckedOperation` |

## Migration Operations

Operations live in `migrations/operations`.

| Operation | Purpose |
| --- | --- |
| `CreateModel` | Create a model table and state. |
| `DeleteModel` | Delete a model table and state. |
| `RenameModel` | Rename a model in state and database. |
| `AlterModelTable` | Rename a model table. |
| `AlterModelTableComment` | Change table comment metadata. |
| `AlterModelOptions` | Change model options. |
| `AlterModelManagers` | Change manager metadata. |
| `AlterOrderWithRespectTo` | Change order-with-respect-to metadata. |
| `AlterTogether` | Change unique/index together metadata. |
| `AddField` | Add a field. |
| `RemoveField` | Remove a field. |
| `AlterField` | Alter field metadata or column type. |
| `RenameField` | Rename a field or column. |
| `AddIndex` | Add an index. |
| `RemoveIndex` | Remove an index. |
| `RenameIndex` | Rename an index. |
| `AddConstraint` | Add a constraint. |
| `RemoveConstraint` | Remove a constraint. |
| `RunSQL` | Run raw SQL with optional reverse SQL. |
| `RunGo` | Run Go data migration functions. |
| `SeparateDatabaseAndState` | Split database and state operations. |

## Flow

1. Build current `ProjectState` from model metadata.
2. Load historical migrations with `Loader`.
3. Compare states with `Autodetector`.
4. Write migration files with `Writer`.
5. Apply operations with `Executor`.
6. Store history in `Recorder`.
7. Validate consistency and safety checks.

## Migration History Recorder

`Recorder` stores applied migration rows in `gogo_migrations` with app, name,
applied timestamp, checksum, and executor version. Recorder SQL is rendered
through the configured database dialect: PostgreSQL uses `$1` placeholders and
SQLite uses `?` placeholders. Applied history writes use `INSERT ... ON
CONFLICT(app, name) DO UPDATE`, not SQLite-only replacement syntax, so history
recording behaves consistently across supported databases.

`Executor` takes a database-backed migration lock before mutating schema or
history. The lock is stored in `gogo_migration_lock`; concurrent `migrate`
processes fail with `ErrMigrationLocked` before operations run. Plan-only
execution does not take the lock.

## Safety Checks

Safety checks detect destructive drops, non-null additions without defaults, irreversible operations, unsafe SQL, and backend-specific hazards where operation metadata exposes the required details.

## Error Types

`ErrInvalidMigration`, `ErrUnsafeMigration`, `ErrIrreversibleOperation`, `ErrDuplicateMigration`, `ErrMissingDependency`, `ErrMigrationCycle`, `ErrInconsistentMigrationHistory`, and `ErrMigrationLocked`.

## Example

```go
migration := migrations.Migration{
	AppLabel: "blog",
	Name:     migrations.InitialMigrationName(),
	Operations: []migrations.Operation{
		operations.RunSQL{SQL: `CREATE TABLE blog_post (id integer primary key, title text)`},
	},
}
err := migration.Validate()
_ = err
```
