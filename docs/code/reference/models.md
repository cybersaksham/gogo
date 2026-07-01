# Models Reference

Model metadata is the source of truth for fields, validation, migrations, content types, admin, permissions, and ORM query compilation.

## Public Packages

| Package | Purpose |
| --- | --- |
| `models` | Model metadata, registry, lifecycle state, field metadata, indexes, constraints, inheritance metadata, permissions, and validation hooks. |
| `models/fields` | Django-style model fields and field-level conversion/validation. |
| `models/constraints` | Index and constraint metadata helpers. |
| `models/hooks` | Model lifecycle hooks. |
| `models/validation` | Structured model validation errors. |

## Core Types

| Type | Purpose |
| --- | --- |
| `models.Model` | Interface exposing `ModelMeta() models.Metadata`. |
| `models.BaseModel` | Embeddable ID/timestamp/state base. |
| `models.Metadata` | App label, model name, table name, options, fields, indexes, constraints, permissions, managers, inheritance, and migration flags. |
| `models.FieldMeta` | Field-level metadata used by models, admin, serializers, migrations, and ORM. |
| `models.Registry` | Model registry for metadata, content types, and migration metadata. |
| `models.CompositePrimaryKey` | Composite primary key metadata. |
| `models.Permission` | Custom model permission metadata. |

## Field Types

`models/fields` exposes all primary Django-style field categories:

| Category | Constructors |
| --- | --- |
| Integer | `NewAutoField`, `NewBigAutoField`, `NewSmallAutoField`, `NewIntegerField`, `NewBigIntegerField`, `NewSmallIntegerField`, `NewPositiveIntegerField`, `NewPositiveBigIntegerField`, `NewPositiveSmallIntegerField` |
| Decimal and float | `NewDecimalField`, `NewFloatField` |
| Text and boolean | `NewBooleanField`, `NewCharField`, `NewTextField`, `NewEmailField`, `NewURLField`, `NewSlugField`, `NewUUIDField` |
| Temporal | `NewDateField`, `NewDateTimeField`, `NewTimeField`, `NewDurationField` |
| Binary and JSON | `NewBinaryField`, `NewJSONField` |
| Files | `NewFileField`, `NewImageField`, `NewFilePathField` |
| Network | `NewGenericIPAddressField` |
| Generated | `NewGeneratedField` |
| Relations | `NewForeignKey`, `NewOneToOneField`, `NewManyToManyField` |
| PostgreSQL | `NewArrayField`, `NewHStoreField`, `NewIntegerRangeField`, `NewBigIntegerRangeField`, `NewDecimalRangeField`, `NewDateRangeField`, `NewDateTimeRangeField` |
| GIS | `NewGeometryField`, `NewPointField`, `NewLineStringField`, `NewPolygonField`, `NewMultiPointField`, `NewMultiLineStringField`, `NewMultiPolygonField`, `NewGeometryCollectionField`, `NewRasterField` |

## Field Options

`fields.Options` covers names, columns, verbose names, help text, primary key, unique, null, blank, default, choices, validators, db index, editable, serialization, db comments, and form/admin metadata.

`fields.Choices` and `fields.NewChoices` define fixed value sets.

## Relations

`fields.RelationConfig` stores target model, through model, related name, related query name, `OnDelete`, reverse relation metadata, and self-reference behavior.

Supported delete behaviors include cascade, protect, restrict, set null, set default, set value, do nothing, and no constraint where implemented by the relation metadata.

## Indexes And Constraints

`models.Index`, `models.IndexField`, `models.Constraint`, and `models.Permission` appear on `models.Metadata`.

`models/constraints` adds validated metadata for:

- Index fields, ordering, opclasses, conditions, include columns, tablespaces, and expressions.
- Unique, check, exclusion, deferrable, null distinct, and covering constraint metadata.

## Validation

`models.ValidateMetadata` rejects invalid migration-facing metadata, including
duplicate field names or columns, declared fields without a primary key,
duplicate generated index names, duplicate generated constraint names, and
duplicate custom permission codenames. `models.Registry.ValidateRelations`
validates relation targets after all model metadata is registered.

Field validation returns `fields.ErrValidation` or `fields.ErrInvalidField`.

Model validation uses:

- Field validators through field options.
- `models/validation.Error` for one field/object error.
- `models/validation.Errors` for grouped errors.
- Model-level uniqueness hooks where provided by forms or model stores.

## Error Types

| Error | Package |
| --- | --- |
| `fields.ErrValidation` | `models/fields` |
| `fields.ErrInvalidField` | `models/fields` |
| `constraints.ErrInvalidIndex` | `models/constraints` |
| `constraints.ErrInvalidConstraint` | `models/constraints` |

## Example

```go
meta := models.Metadata{
	AppLabel:  "blog",
	ModelName: "Post",
	TableName: "blog_post",
	Fields: []models.FieldMeta{
		{Name: "id", Column: "id", PrimaryKey: true},
		{Name: "title", Column: "title"},
	},
}
label := meta.Label()
_ = label
```
