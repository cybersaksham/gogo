# Models, ORM, And Migrations Rules

Use this rule for `models`, `orm`, `migrations`, `internal/schema`, and database-facing behavior.

## Models

- `models.Metadata` is the source of truth for model structure.
- Preserve field metadata for app label, model name, table name, fields, indexes, constraints, permissions, inheritance, and relationships.
- Validate unsafe or inconsistent metadata close to registration.

## ORM

- Query compilation must be deterministic.
- Support dialect differences through dialect interfaces, not ad hoc string branching in callers.
- Preserve placeholder ordering and argument ordering.
- Raw SQL APIs must keep caller-provided SQL explicit.

## Migrations

- Migration files must be deterministic.
- Unsafe non-null additions, destructive operations, and irreversible data operations need explicit safety behavior.
- Migration history must preserve app, name, checksum, executor version, and applied state.
- Update migration compatibility fixtures when manifest formats change.

## Verification

```bash
go test ./models/... ./orm/... ./migrations/... ./internal/schema
go test -race ./orm/...
```

