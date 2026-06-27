# Models Fields And Validation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Django-style model declarations, complete field coverage, metadata, relationships, validation, constraints, indexes, hooks, and model inheritance patterns for Go.

**Architecture:** Public package `models` owns model contracts and field metadata. Field definitions are declarative Go values that generate schema state, validation behavior, admin metadata, serializer metadata, and ORM query metadata.

**Tech Stack:** Go interfaces, generics, reflection where unavoidable, `database/sql/driver`, validators, deterministic metadata registries.

---

## Files

- Create: `models/model.go`
- Create: `models/instance.go`
- Create: `models/meta.go`
- Create: `models/registry.go`
- Create: `models/errors.go`
- Create: `models/validation/validator.go`
- Create: `models/validation/errors.go`
- Create: `models/constraints/constraint.go`
- Create: `models/constraints/index.go`
- Create: `models/hooks/hooks.go`
- Create: `models/fields/field.go`
- Create: `models/fields/options.go`
- Create: `models/fields/choices.go`
- Create: `models/fields/scalar.go`
- Create: `models/fields/numeric.go`
- Create: `models/fields/temporal.go`
- Create: `models/fields/text.go`
- Create: `models/fields/binary.go`
- Create: `models/fields/json.go`
- Create: `models/fields/file.go`
- Create: `models/fields/network.go`
- Create: `models/fields/generated.go`
- Create: `models/fields/relation.go`
- Create: `models/fields/postgres.go`
- Create: `models/fields/gis.go`
- Create: `models/model_test.go`
- Create: `models/fields/fields_test.go`
- Create: `models/validation/validator_test.go`
- Create: `models/constraints/constraint_test.go`

## Task 1: Define Model Contract

- [ ] Create `models/model.go`.
- [ ] Define `Model` interface with stable metadata access.
- [ ] Define `BaseModel` with optional fields:
  - `ID`
  - `CreatedAt`
  - `UpdatedAt`
- [ ] Define `CompositePrimaryKey` metadata for models with multiple primary key columns.
- [ ] Define model state values:
  - New
  - Loaded
  - Dirty
  - Deleted
- [ ] Add support for table name, app label, verbose name, verbose plural, and default manager name.
- [ ] Add tests for metadata resolution, zero-value model state, explicit table names, and composite primary key metadata.
- [ ] Run `go test ./models`.
- [ ] Commit with message `Add Model Contract`.

## Task 2: Add Model Instance API

- [ ] Create `models/instance.go`.
- [ ] Implement model instance operations:
  - `Save`
  - `Delete`
  - `FullClean`
  - `CleanFields`
  - `Clean`
  - `ValidateUnique`
  - `ValidateConstraints`
  - `RefreshFromDB`
  - `FromDB`
  - `GetAbsoluteURL`
  - `GetFieldDisplay`
  - `SerializableValue`
  - `NaturalKey`
- [ ] Support save options for force insert, force update, update fields, using database, and raw save.
- [ ] Support delete options for using database and keeping parent rows where inheritance requires it.
- [ ] Add tests for every instance operation, save option, delete option, field display, natural key, and refresh behavior.
- [ ] Run `go test ./models ./orm`.
- [ ] Commit with message `Add Model Instance API`.

## Task 3: Add Model Metadata

- [ ] Create `models/meta.go`.
- [ ] Support Django-style model options:
  - `DBTable`
  - `DBTableComment`
  - `AppLabel`
  - `VerboseName`
  - `VerboseNamePlural`
  - `Ordering`
  - `OrderWithRespectTo`
  - `GetLatestBy`
  - `DefaultRelatedName`
  - `DefaultManagerName`
  - `BaseManagerName`
  - `Abstract`
  - `Proxy`
  - `Managed`
  - `RequiredDBVendor`
  - `RequiredDBFeatures`
  - `Indexes`
  - `Constraints`
  - `Permissions`
  - `DefaultPermissions`
  - `SelectOnSave`
- [ ] Validate contradictory options such as unmanaged models with generated migrations.
- [ ] Add tests for every metadata option.
- [ ] Run `go test ./models`.
- [ ] Commit with message `Add Model Metadata Options`.

## Task 4: Add Model Registry

- [ ] Create `models/registry.go`.
- [ ] Register models by app label and model name.
- [ ] Reject duplicate model names within an app.
- [ ] Support lookup by `app_label.ModelName`.
- [ ] Expose copied metadata for app registry, migrations, ORM, admin, serializers, and content types.
- [ ] Add tests for registration, duplicate rejection, lookup, ordering, and immutability.
- [ ] Run `go test ./models`.
- [ ] Commit with message `Add Model Registry`.

## Task 5: Define Base Field Contract

- [ ] Create `models/fields/field.go`.
- [ ] Define common field options:
  - `Name`
  - `Column`
  - `PrimaryKey`
  - `Unique`
  - `DBIndex`
  - `DBDefault`
  - `DBCollation`
  - `Null`
  - `Blank`
  - `Default`
  - `Choices`
  - `Editable`
  - `HelpText`
  - `VerboseName`
  - `ErrorMessages`
  - `Validators`
  - `UniqueForDate`
  - `UniqueForMonth`
  - `UniqueForYear`
  - `Serialize`
  - `DBComment`
  - `DBTablespace`
- [ ] Define field methods:
  - `Kind()`
  - `ColumnType(dialect string)`
  - `Validate(value any) error`
  - `ToDB(value any) (any, error)`
  - `FromDB(value any) (any, error)`
  - `Clone() Field`
- [ ] Add tests for option defaults, validation, cloning, and column names.
- [ ] Run `go test ./models/fields`.
- [ ] Commit with message `Add Base Field Contract`.

## Task 6: Implement Field Choices And Enumerations

- [ ] Create `models/fields/choices.go`.
- [ ] Support simple choices, grouped choices, integer choices, text choices, blank labels, display labels, and custom enumeration types.
- [ ] Generate display labels for `GetFieldDisplay`.
- [ ] Validate duplicate choice values and invalid default values.
- [ ] Add tests for simple choices, grouped choices, integer choices, text choices, blank labels, display labels, and invalid choices.
- [ ] Run `go test ./models/fields ./models`.
- [ ] Commit with message `Add Field Choices`.

## Task 7: Implement Auto And Numeric Fields

- [ ] Create `models/fields/numeric.go`.
- [ ] Implement:
  - `AutoField`
  - `BigAutoField`
  - `SmallAutoField`
  - `IntegerField`
  - `BigIntegerField`
  - `SmallIntegerField`
  - `PositiveIntegerField`
  - `PositiveBigIntegerField`
  - `PositiveSmallIntegerField`
  - `DecimalField`
  - `FloatField`
- [ ] Validate decimal max digits and decimal places.
- [ ] Validate positive field ranges.
- [ ] Add dialect column types for PostgreSQL and SQLite.
- [ ] Add tests for boundaries, conversion, validation, and column types.
- [ ] Run `go test ./models/fields`.
- [ ] Commit with message `Add Numeric Model Fields`.

## Task 8: Implement Text And Boolean Fields

- [ ] Create `models/fields/text.go`.
- [ ] Implement:
  - `BooleanField`
  - `CharField`
  - `TextField`
  - `EmailField`
  - `URLField`
  - `SlugField`
  - `UUIDField`
- [ ] Enforce max length for `CharField`.
- [ ] Validate email, URL, slug, and UUID formats.
- [ ] Support choices and empty values.
- [ ] Add tests for validation, max length, choices, and DB conversion.
- [ ] Run `go test ./models/fields`.
- [ ] Commit with message `Add Text And Boolean Model Fields`.

## Task 9: Implement Temporal Fields

- [ ] Create `models/fields/temporal.go`.
- [ ] Implement:
  - `DateField`
  - `DateTimeField`
  - `TimeField`
  - `DurationField`
- [ ] Support `AutoNow` and `AutoNowAdd`.
- [ ] Normalize timezone behavior using framework settings.
- [ ] Validate date-only and time-only values.
- [ ] Add tests for conversion, timezone normalization, auto values, and zero-value handling.
- [ ] Run `go test ./models/fields`.
- [ ] Commit with message `Add Temporal Model Fields`.

## Task 10: Implement Binary JSON File And Network Fields

- [ ] Create `models/fields/binary.go`.
- [ ] Create `models/fields/json.go`.
- [ ] Create `models/fields/file.go`.
- [ ] Create `models/fields/network.go`.
- [ ] Implement:
  - `BinaryField`
  - `JSONField`
  - `GeneratedField`
  - `FileField`
  - `ImageField`
  - `FilePathField`
  - `GenericIPAddressField`
- [ ] Validate JSON marshalability and database scan behavior.
- [ ] Validate file upload path generation and path traversal protection.
- [ ] Validate image metadata after file service exists; before that, expose a metadata hook with a tested fake inspector.
- [ ] Validate IPv4, IPv6, and unpacked IPv4-mapped IPv6 options.
- [ ] Add tests for every field.
- [ ] Run `go test ./models/fields`.
- [ ] Commit with message `Add Binary JSON File And Network Fields`.

## Task 11: Implement Relationship Fields

- [ ] Create `models/fields/relation.go`.
- [ ] Implement:
  - `ForeignKey`
  - `OneToOneField`
  - `ManyToManyField`
  - Self-referential relationships
  - Lazy references by `app_label.ModelName`
  - Through models
  - Reverse relations
  - Related names
  - Related query names
- [ ] Implement delete behaviors:
  - Cascade
  - Protect
  - Restrict
  - Set null
  - Set default
  - Set value
  - Do nothing
- [ ] Validate missing target models, invalid through models, duplicate reverse names, and nullable delete behaviors.
- [ ] Add tests for each relationship type and delete behavior.
- [ ] Run `go test ./models/fields ./models`.
- [ ] Commit with message `Add Relationship Model Fields`.

## Task 12: Implement PostgreSQL-Specific Fields

- [ ] Create `models/fields/postgres.go`.
- [ ] Implement optional PostgreSQL field metadata:
  - Array field
  - HStore-like key value field
  - Integer range
  - Big integer range
  - Decimal range
  - Date range
  - DateTime range
- [ ] Gate these fields behind dialect capability checks.
- [ ] Add tests for type metadata, unsupported dialect errors, and validation.
- [ ] Run `go test ./models/fields`.
- [ ] Commit with message `Add PostgreSQL Model Fields`.

## Task 13: Implement GIS Field Metadata

- [ ] Create `models/fields/gis.go`.
- [ ] Implement metadata for:
  - Geometry
  - Point
  - LineString
  - Polygon
  - MultiPoint
  - MultiLineString
  - MultiPolygon
  - GeometryCollection
  - Raster metadata
- [ ] Store SRID, geography flag, dimensionality, and spatial index preference.
- [ ] Validate unsupported dialects and invalid SRID values.
- [ ] Add tests for metadata and validation.
- [ ] Run `go test ./models/fields`.
- [ ] Commit with message `Add GIS Model Field Metadata`.

## Task 14: Implement Validation System

- [ ] Create `models/validation/validator.go`.
- [ ] Create `models/validation/errors.go`.
- [ ] Support:
  - Required validation
  - Type validation
  - Max/min value
  - Max/min length
  - Regex
  - Email
  - URL
  - Slug
  - UUID
  - Choice validation
  - Unique validation hook
  - Unique for date validation
  - Unique for month validation
  - Unique for year validation
  - Constraint validation
  - Custom validators
  - Field-level validation
  - Model-level validation
- [ ] Return structured errors keyed by field name.
- [ ] Add tests for every validator and combined error output.
- [ ] Run `go test ./models/validation ./models/fields`.
- [ ] Commit with message `Add Model Validation`.

## Task 15: Implement Constraints And Indexes

- [ ] Create `models/constraints/constraint.go`.
- [ ] Create `models/constraints/index.go`.
- [ ] Support:
  - Unique constraint
  - Check constraint
  - Exclusion constraint metadata
  - Deferrable constraints
  - Nulls distinct behavior
  - Constraint violation error code
  - Constraint violation error message
  - Conditional constraints
  - Functional indexes
  - Covering indexes
  - Partial indexes
  - Index ordering
  - Operator classes
  - Tablespace metadata
- [ ] Add tests for metadata validation and deterministic names.
- [ ] Run `go test ./models/constraints`.
- [ ] Commit with message `Add Model Constraints And Indexes`.

## Task 16: Implement Model Hooks

- [ ] Create `models/hooks/hooks.go`.
- [ ] Support:
  - Before validate
  - After validate
  - Before save
  - After save
  - Before delete
  - After delete
  - Many-to-many changed
- [ ] Keep hooks context-aware.
- [ ] Ensure hook ordering is deterministic.
- [ ] Add tests for successful hooks, failing hooks, context cancellation, and hook order.
- [ ] Run `go test ./models/hooks ./models`.
- [ ] Commit with message `Add Model Lifecycle Hooks`.

## Task 17: Implement Inheritance And Composition Rules

- [ ] Support abstract base models through embedded structs and metadata inheritance.
- [ ] Support multi-table inheritance with generated parent link relationships, parent table joins, parent save order, and parent delete behavior.
- [ ] Support proxy models at metadata level for admin and manager behavior.
- [ ] Support inheritable auth user extension through embedding and profile-style extension metadata while preserving the framework-owned auth user table.
- [ ] Add tests for abstract fields, overridden fields, multi-table parent links, proxy metadata, parent save order, parent delete behavior, and auth user extension metadata.
- [ ] Run `go test ./models`.
- [ ] Commit with message `Add Model Inheritance Rules`.

## Acceptance Checklist

- [ ] Every Django core field family has a Gogo equivalent.
- [ ] Field options include database defaults, collation, serialization, and unique-for-date/month/year behavior.
- [ ] Model instance APIs cover save, delete, clean, uniqueness, constraints, refresh, display, URLs, and natural keys.
- [ ] Relationship fields preserve forward and reverse metadata.
- [ ] Model metadata covers Django `Meta` behavior.
- [ ] Validation returns structured field errors.
- [ ] Constraints and indexes are migration-ready.
- [ ] Hooks are context-aware and deterministic.
- [ ] Abstract, multi-table, and proxy inheritance modes are implemented and tested.
