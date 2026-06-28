# Database Operations

Gogo supports PostgreSQL as the production reference database and SQLite for
tests, local development, generated project smoke tests, and narrow
single-process deployments.

## Configuration

Set `DATABASE_URL` in every environment. The framework treats it as required.

Examples:

```bash
DATABASE_URL=postgres://gogo:password@db:5432/gogo?sslmode=require
DATABASE_URL=sqlite://./db.sqlite3
```

Keep database credentials in a secret manager. Do not commit `.env` files,
database dumps, local SQLite databases, or connection URLs containing passwords.

## Supported Engines

| Engine | Production status | Notes |
| --- | --- | --- |
| PostgreSQL 17+ | Recommended | Primary production target and CI integration database. |
| SQLite through `modernc.org/sqlite v1.53.0` | Limited | Use for local, test, and single-process deployments only. |

PostgreSQL-specific contrib packages can require extensions such as `pg_trgm`,
`btree_gin`, `btree_gist`, `hstore`, `unaccent`, and `postgis`.

## Migrations

Use migrations as the only production schema change path.

```bash
gogo makemigrations --check --dry-run
gogo migrate --plan
gogo sqlmigrate app_label 0001
gogo migrate
gogo showmigrations
```

Operational rules:

- Run migrations once per release.
- Take a backup before production migration.
- Review generated SQL for locks, table rewrites, destructive operations, and
  long-running backfills.
- Prefer nullable columns, backfills, then constraints across separate releases.
- Avoid editing applied migrations.
- Keep raw SQL reversible when rollback is required.
- Use `--fake` or `--fake-initial` only after manual inspection confirms the
  database already matches the migration state.
- Use `--prune` only when stale migration records are understood and backed up.

## Connection Management

The app should use a bounded database pool sized for total web and worker
replicas. Keep database max connections below the database server limit after
reserving capacity for migrations, admin access, backups, and emergency shells.

Recommended checks:

- Connection acquisition timeout.
- Query timeout for request-scoped work.
- Separate long-running backfills from request handlers.
- Readiness check that opens and pings the database.
- Slow query logging in the database server.

## Backups

Backups must include application tables, migration records, auth tables, admin
logs, queue result tables when SQL results are used, and uploaded media metadata
stored in the database.

Minimum backup policy:

- Automated scheduled backups.
- Backup before every production migration.
- Restore test after backup configuration changes.
- Encrypted storage.
- Retention matching product and compliance requirements.
- Separate storage account or bucket from the primary database environment.

SQLite deployments must copy the database only when no writer is active or use
the database engine's backup API through project-specific tooling.

## Restore

A restore runbook must define:

- Target environment.
- Database owner and permissions.
- Extension setup.
- Restore command.
- Post-restore migration check.
- Application smoke tests.
- Queue and worker posture after restore.

After restore, run:

```bash
gogo showmigrations
gogo migrate --plan
```

If the plan wants to reapply migrations that should already be applied, stop and
inspect the migration recorder before allowing traffic.

## Rollbacks

Database rollback is safe only when the data shape is compatible with the old
application.

- Additive schema changes can often stay in place.
- Dropped columns, changed data types, destructive data migrations, and renamed
  tables usually require restore or a reviewed reverse migration.
- Worker processes must be stopped before rolling back schema used by tasks.
- Beat should be stopped before rolling back schedule storage changes.

Do not guess on destructive rollback. Restore from the verified backup when
compatibility is uncertain.
