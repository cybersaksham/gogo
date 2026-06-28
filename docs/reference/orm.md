# ORM Reference

The ORM turns immutable query state into dialect-specific SQL. It is split into connections, dialects, expressions, lookups, query state, managers, transactions, raw SQL, joins, prefetches, functions, aggregates, and compiler output.

## Public Types

| Area | Types |
| --- | --- |
| Connections | `orm.DatabaseConfig`, `orm.Database`, `orm.Connections`, `orm.Logger`, `orm.HealthCheck` |
| Compiler | `orm.Compiler`, `orm.CompiledSQL` |
| Query state | `orm.Query`, `orm.Predicate`, `orm.Join`, `orm.ExpressionRef`, `orm.RelatedState`, `orm.LockState`, `orm.SetOperation`, `orm.WindowState` |
| QuerySet | `orm.QuerySet`, `orm.QuerySetState`, `orm.RawQuery`, `orm.GetOrCreatePlan`, `orm.UpdateOrCreatePlan` |
| Managers | `orm.Manager`, `orm.ManagerSet`, `orm.TypedManager` |
| Transactions | `orm.TransactionManager`, `orm.Transaction`, `orm.TxOption` |
| Expressions | `orm.Expression`, `orm.ExpressionContext`, `orm.FieldExpression`, `orm.ValueExpression`, `orm.FilterExpression`, `orm.QExpression`, `orm.FunctionExpression`, `orm.CastExpression`, `orm.CaseExpression`, `orm.WhenExpression`, `orm.RawExpression`, `orm.SubqueryExpression`, `orm.OuterRefExpression`, `orm.ExistsExpression`, `orm.BinaryExpression`, `orm.WindowExpression`, `orm.Frame` |
| Lookups | `orm.Lookup`, `orm.LookupContext`, `orm.LookupRegistry`, `orm.SQLFragment`, `orm.JSONPathValue` |
| Aggregates | `orm.AggregateExpression`, `orm.AggregateResult` |
| Joins and prefetch | `orm.RelationMeta`, `orm.JoinPlanner`, `orm.JoinClause`, `orm.PrefetchSpec`, `orm.PrefetchPlan` |
| Raw SQL | `orm.RawExecutor`, `orm.RowScanner`, `orm.RawExecResult` |
| Dialects | `orm/dialects.Dialect`, `LockOptions`, `LimitOffset`, `SchemaIntrospection`, plus `orm/dialects/postgres.Dialect` and `orm/dialects/sqlite.Dialect` |

## Database Connections

Use `orm.OpenDatabase` with a `DatabaseConfig`, then store handles in `orm.NewConnections`.

Supported dialect packages:

- `orm/dialects/postgres`
- `orm/dialects/sqlite`

## QuerySet Operations

QuerySets are immutable. Each method returns a new QuerySet:

`All`, `None`, `Filter`, `Exclude`, `OrderBy`, `Reverse`, `Distinct`, `Union`, `Intersection`, `Difference`, `Values`, `ValuesList`, `Dates`, `DateTimes`, `Only`, `Defer`, `SelectRelated`, `PrefetchRelated`, `Annotate`, `Alias`, `First`, `Last`, `Latest`, `Earliest`, `Get`, `InBulk`, `Create`, `GetOrCreate`, `UpdateOrCreate`, `BulkCreate`, `BulkUpdate`, `Update`, `Delete`, `Count`, `Exists`, `Contains`, `Using`, `SelectForUpdate`, `ComplexFilter`, `Iterator`, `Explain`, and `Raw`.

## Lookups

Built-in lookup constants include:

`exact`, `iexact`, `contains`, `icontains`, `in`, `gt`, `gte`, `lt`, `lte`, `startswith`, `istartswith`, `endswith`, `iendswith`, `range`, `date`, `year`, `iso_year`, `month`, `day`, `week`, `week_day`, `iso_week_day`, `quarter`, `time`, `hour`, `minute`, `second`, `isnull`, `regex`, `iregex`, and JSON path lookups.

Custom lookups register through `LookupRegistry.Register`.

## Expressions And Functions

Core expression helpers include `F`, `Value`, `Filter`, `Q`, `Func`, `Cast`, `Case`, `When`, `RawSQL`, `UnsafeRawSQL`, `Subquery`, `OuterRef`, `Exists`, and binary expressions.

Database functions include `Coalesce`, `Greatest`, `Least`, `NullIf`, `Now`, `Lower`, `Upper`, `Length`, `Substr`, `Replace`, `Concat`, `MD5`, `SHA1`, `SHA224`, `SHA256`, `SHA384`, `SHA512`, `Round`, `Ceil`, `Floor`, `Abs`, `Mod`, `Power`, `Random`, `Extract`, `Trunc`, JSON object/array helpers, and collation helpers.

Window functions include `RowNumber`, `Rank`, `DenseRank`, `PercentRank`, `CumeDist`, `NTile`, `Lag`, `Lead`, `FirstValue`, `LastValue`, and `NthValue`.

## Aggregates

Aggregates support aliases, filters, distinct, defaults, ordering where supported, and window compatibility.

Built-ins include `Count`, `Sum`, `Avg`, `Min`, `Max`, `StdDev`, `Variance`, and aggregate result typed accessors.

## Transactions

Use `orm.NewTransactionManager(database).Atomic(ctx, fn)` for root and nested transactions. Nested transactions use dialect savepoints. Options include `orm.WithIsolation` and `orm.ReadOnly`.

## Raw SQL

Use `orm.ParameterizedRaw` for parameterized raw SQL. Use `orm.UnsafeRawQuery` only when the query is intentionally unparameterized. `orm.ValidateRawQuery` rejects unsafe accidental raw SQL.

## Error Types

`ErrInvalidDatabaseConfig`, `ErrDatabaseExists`, `ErrDatabaseNotFound`, `ErrUnsupportedDialectFeature`, `ErrTransactionClosed`, `ErrInvalidQuery`, `ErrInvalidLookup`, `ErrUnsupportedFunction`, `ErrInvalidExpression`, `ErrUnsafeRawSQL`, `IntegrityError`, and `QueryError`.

## Example

```go
compiler := orm.NewCompiler(postgres.New())
query := orm.NewQuery(models.Metadata{AppLabel: "blog", ModelName: "Post", TableName: "blog_post"}).
	Select("id", "title")
compiled, err := compiler.CompileSelect(query)
```
