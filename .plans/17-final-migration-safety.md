# Final Migration Safety Implementation Checklist

Source plan: `/Users/cybersaksham/.codex/worktrees/44a5/ai-calls/docs/gogo-final-migration-safety-plan.html`

Goal: make Gogo safe for existing-schema adoption and future project-owned migrations without losing defaults, field alterations, indexes, constraints, or PostgreSQL-specific shape.

## Phases

1. Default contract and SQL rendering
   - Add typed public database defaults for literals and trusted SQL expressions.
   - Normalize model defaults into migration state/specs.
   - Render defaults in `CreateTable`, `AddColumn`, and `AlterDefault`.

2. Full field alterations
   - Decompose `AlterField` into type, default, nullability, and collation actions.
   - Reject unsafe non-null transitions unless a default or explicit acknowledgement exists.
   - Keep forwards/backwards ordering deterministic.

3. Indexes and constraints
   - Expand field `Unique` and `DBIndex` into deterministic state objects.
   - Make `CreateModel` create model constraints and indexes after table creation.
   - Ensure index/constraint diffs produce explicit operations.

4. Fake-initial and diffschema comparator
   - Extend column schema/introspection with type, default, collation, identity, and ordinal data.
   - Share one comparator between fake-initial and diffschema.
   - Fail closed when required shape cannot be inspected.

5. PostgreSQL features
   - Add index method, opclasses, include columns, partial conditions, expressions, and concurrency metadata.
   - Add constraint foreign-key metadata.
   - Keep advanced features representable through `RunSQL` and `SeparateDatabaseAndState`.

6. Autodetector and manifests
   - Emit minimal operations for default, index, unique, and rich metadata changes.
   - Preserve old manifests and round-trip new rich specs.

7. Docs, templates, verification
   - Update public docs, operations docs, generated README/rules, and app model examples.
   - Run focused tests, full tests, race tests, docs verification, generated-project smoke, and available external gates.
