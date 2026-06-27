# ORM Query Engine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Django-style ORM for Go that translates model metadata and queryset operations into safe SQL with lazy evaluation, expressions, joins, transactions, and multi-database support.

**Architecture:** Public package `orm` provides database connections, managers, querysets, expressions, transactions, and raw SQL escape hatches. Dialects own SQL rendering differences while query state remains dialect-neutral until compilation.

**Tech Stack:** `database/sql`, context-aware queries, SQL dialect interfaces, PostgreSQL first, SQLite for tests, generics for typed managers where useful.

---

## Files

- Create: `orm/db.go`
- Create: `orm/connection.go`
- Create: `orm/router.go`
- Create: `orm/manager.go`
- Create: `orm/queryset.go`
- Create: `orm/query.go`
- Create: `orm/compiler.go`
- Create: `orm/lookups.go`
- Create: `orm/expressions.go`
- Create: `orm/aggregates.go`
- Create: `orm/functions.go`
- Create: `orm/window.go`
- Create: `orm/joins.go`
- Create: `orm/prefetch.go`
- Create: `orm/transactions.go`
- Create: `orm/raw.go`
- Create: `orm/errors.go`
- Create: `orm/dialects/dialect.go`
- Create: `orm/dialects/postgres/dialect.go`
- Create: `orm/dialects/sqlite/dialect.go`
- Create: `orm/tests/models_test.go`
- Create: `orm/queryset_test.go`
- Create: `orm/compiler_test.go`
- Create: `orm/transactions_test.go`

## Task 1: Define Database Connections

- [ ] Create `orm/db.go`.
- [ ] Define `Database` with connection name, driver, DSN, SQL DB handle, dialect, logger, and health check.
- [ ] Support default database and named databases.
- [ ] Implement open, close, ping, and stats.
- [ ] Add tests with SQLite in-memory database.
- [ ] Run `go test ./orm`.
- [ ] Commit with message `Add ORM Database Connections`.

## Task 2: Add Database Router

- [ ] Create `orm/router.go`.
- [ ] Support routing hooks:
  - `DBForRead(model)`
  - `DBForWrite(model)`
  - `AllowRelation(modelA, modelB)`
  - `AllowMigrate(db, appLabel, modelName)`
- [ ] Implement default router.
- [ ] Add tests for default routing, custom routing, migration allow/deny, and relation allow/deny.
- [ ] Run `go test ./orm`.
- [ ] Commit with message `Add ORM Database Router`.

## Task 3: Add Dialect Interface

- [ ] Create `orm/dialects/dialect.go`.
- [ ] Define dialect methods for:
  - Placeholder format
  - Identifier quoting
  - Column types
  - Returning support
  - Upsert support
  - JSON operations
  - Date extraction
  - Lock clauses
  - Limit and offset
  - Savepoints
  - Schema introspection hooks
- [ ] Implement PostgreSQL dialect.
- [ ] Implement SQLite dialect for tests and quickstart.
- [ ] Add tests for placeholders, quoting, returning, limit/offset, and lock clauses.
- [ ] Run `go test ./orm/dialects/...`.
- [ ] Commit with message `Add SQL Dialects`.

## Task 4: Add Query State

- [ ] Create `orm/query.go`.
- [ ] Represent query state immutably:
  - Model metadata
  - Selected columns
  - Filters
  - Excludes
  - Joins
  - Ordering
  - Grouping
  - Having
  - Limit
  - Offset
  - Distinct
  - Annotations
  - Related loading
  - Locking
  - Set operations
  - Window expressions
- [ ] Ensure queryset methods clone query state instead of mutating the original.
- [ ] Add tests for clone behavior.
- [ ] Run `go test ./orm`.
- [ ] Commit with message `Add ORM Query State`.

## Task 5: Add Lookups

- [ ] Create `orm/lookups.go`.
- [ ] Support lookup operations:
  - Exact
  - IExact
  - Contains
  - IContains
  - In
  - Greater than
  - Greater than or equal
  - Less than
  - Less than or equal
  - Range
  - StartsWith
  - IStartsWith
  - EndsWith
  - IEndsWith
  - Date
  - Year
  - Month
  - Day
  - Week
  - WeekDay
  - Quarter
  - Time
  - Hour
  - Minute
  - Second
  - IsNull
  - Regex
  - IRegex
  - JSON path
- [ ] Add custom lookup registration.
- [ ] Add compiler tests for every lookup.
- [ ] Run `go test ./orm`.
- [ ] Commit with message `Add ORM Field Lookups`.

## Task 6: Add Expressions

- [ ] Create `orm/expressions.go`.
- [ ] Support:
  - `F` field references
  - `Value`
  - `Q` objects with AND, OR, NOT
  - `Func`
  - `Case`
  - `When`
  - `Coalesce`
  - `Cast`
  - `RawSQL`
  - `Subquery`
  - `OuterRef`
  - `Exists`
  - Arithmetic expressions
  - Comparison expressions
  - Window expressions
  - Window frames
  - Row range frames
  - Value range frames
- [ ] Add SQL injection tests proving values are parameterized.
- [ ] Add compiler tests for nested expressions.
- [ ] Run `go test ./orm`.
- [ ] Commit with message `Add ORM Expressions`.

## Task 7: Add Database Functions And Window Expressions

- [ ] Create `orm/functions.go`.
- [ ] Create `orm/window.go`.
- [ ] Implement database functions:
  - Cast
  - Coalesce
  - Collate
  - Greatest
  - Least
  - NullIf
  - Extract
  - Trunc
  - Now
  - Lower
  - Upper
  - Length
  - Substr
  - Replace
  - Concat
  - MD5
  - SHA variants where dialect supports them
  - Round
  - Ceil
  - Floor
  - Abs
  - Mod
  - Power
  - Random
  - JSONObject
  - JSONArray
- [ ] Implement window helpers:
  - Row number
  - Rank
  - Dense rank
  - Percent rank
  - Cume dist
  - NTile
  - Lag
  - Lead
  - First value
  - Last value
  - Nth value
- [ ] Add tests for SQL rendering, unsupported dialect behavior, and annotation integration.
- [ ] Run `go test ./orm`.
- [ ] Commit with message `Add ORM Functions And Windows`.

## Task 8: Add Aggregates And Annotations

- [ ] Create `orm/aggregates.go`.
- [ ] Support:
  - Count
  - Sum
  - Avg
  - Min
  - Max
  - StdDev
  - Variance
  - Window-compatible aggregate rendering
  - Filtered aggregates
  - Distinct aggregates
  - Aliases
  - Annotations
- [ ] Add tests for SQL generation and scan output.
- [ ] Run `go test ./orm`.
- [ ] Commit with message `Add ORM Aggregates And Annotations`.

## Task 9: Add Query Compiler

- [ ] Create `orm/compiler.go`.
- [ ] Compile query state into parameterized SQL.
- [ ] Support SELECT, INSERT, UPDATE, DELETE, COUNT, EXISTS, aggregate, set operation, window expression, and raw subquery compilation.
- [ ] Validate unknown fields, ambiguous joins, invalid ordering, invalid annotations, and unsupported dialect features.
- [ ] Add golden SQL tests for PostgreSQL and SQLite.
- [ ] Run `go test ./orm`.
- [ ] Commit with message `Add ORM SQL Compiler`.

## Task 10: Add QuerySet API

- [ ] Create `orm/queryset.go`.
- [ ] Implement lazy operations:
  - `All`
  - `None`
  - `Filter`
  - `Exclude`
  - `OrderBy`
  - `Reverse`
  - `Distinct`
  - `Union`
  - `Intersection`
  - `Difference`
  - `Values`
  - `ValuesList`
  - `Dates`
  - `DateTimes`
  - `Only`
  - `Defer`
  - `SelectRelated`
  - `PrefetchRelated`
  - `Annotate`
  - `Alias`
  - `Aggregate`
  - `First`
  - `Last`
  - `Latest`
  - `Earliest`
  - `Get`
  - `InBulk`
  - `Create`
  - `GetOrCreate`
  - `UpdateOrCreate`
  - `BulkCreate`
  - `BulkUpdate`
  - `Update`
  - `Delete`
  - `Count`
  - `Exists`
  - `Contains`
  - `Using`
  - `SelectForUpdate`
  - `ComplexFilter`
  - `Iterator`
  - `Explain`
  - `Raw`
- [ ] Add tests for lazy evaluation and every public operation.
- [ ] Run `go test ./orm`.
- [ ] Commit with message `Add QuerySet API`.

## Task 11: Add Managers

- [ ] Create `orm/manager.go`.
- [ ] Add typed model managers.
- [ ] Support default managers and base managers from model metadata.
- [ ] Allow custom managers with custom queryset methods.
- [ ] Add tests for default manager, custom manager, base manager, and manager inheritance.
- [ ] Run `go test ./orm ./models`.
- [ ] Commit with message `Add ORM Managers`.

## Task 12: Add Joins And Related Loading

- [ ] Create `orm/joins.go`.
- [ ] Create `orm/prefetch.go`.
- [ ] Implement foreign key joins, one-to-one joins, reverse joins, and many-to-many joins.
- [ ] Implement `SelectRelated` as SQL joins.
- [ ] Implement `PrefetchRelated` as separate batched queries.
- [ ] Support custom prefetch querysets and target attributes.
- [ ] Add tests for N+1 prevention, nested relations, nullable relations, and many-to-many prefetch.
- [ ] Run `go test ./orm`.
- [ ] Commit with message `Add ORM Related Loading`.

## Task 13: Add Transactions

- [ ] Create `orm/transactions.go`.
- [ ] Support:
  - `Atomic`
  - Nested savepoints
  - Commit
  - Rollback
  - Rollback on panic
  - Isolation levels
  - Read-only transactions
  - `OnCommit` callbacks
  - Row locking through `SelectForUpdate`
- [ ] Add tests for commit, rollback, nested rollback, panic rollback, on-commit callbacks, and lock SQL.
- [ ] Run `go test ./orm`.
- [ ] Commit with message `Add ORM Transactions`.

## Task 14: Add Raw SQL

- [ ] Create `orm/raw.go`.
- [ ] Support raw SELECT mapped to models.
- [ ] Support raw exec with parameters.
- [ ] Require explicit unsafe marker for unparameterized SQL fragments.
- [ ] Add tests for parameter binding, model scanning, exec rows affected, and unsafe marker rejection.
- [ ] Run `go test ./orm`.
- [ ] Commit with message `Add Raw SQL Support`.

## Task 15: Add Error Semantics

- [ ] Create `orm/errors.go`.
- [ ] Add typed errors:
  - `ErrDoesNotExist`
  - `ErrMultipleObjectsReturned`
  - `ErrIntegrity`
  - `ErrInvalidQuery`
  - `ErrUnsupportedDialectFeature`
  - `ErrTransactionClosed`
  - `ErrNoRowsAffected`
- [ ] Add `errors.Is` and `errors.As` tests.
- [ ] Run `go test ./orm`.
- [ ] Commit with message `Add ORM Error Types`.

## Acceptance Checklist

- [ ] QuerySets are lazy and immutable.
- [ ] Every lookup is parameterized.
- [ ] QuerySet set operations, date truncation helpers, latest/earliest, in-bulk, contains, using, and select-for-update are implemented.
- [ ] Database functions and window expressions are available.
- [ ] Expressions and raw SQL cannot bypass parameterization accidentally.
- [ ] Transactions support savepoints and on-commit callbacks.
- [ ] Related loading supports forward, reverse, nested, and many-to-many relations.
- [ ] Multi-database routing is available to reads, writes, and migrations.
- [ ] PostgreSQL and SQLite tests pass.
