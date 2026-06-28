# Forms, Templates, Static Files, And Fixtures Rules

Use this rule for `forms`, `templates`, `files`, `static`, and fixture CLI behavior.

## Forms

- Preserve binding, cleaned data, changed data, errors, field order, prefixes, field groups, media, formsets, and model forms.
- Widgets must escape labels, values, and attributes.

## Templates

- Preserve context processors, safe strings, escaping, template helpers, and Django-compatible tags/filters.
- Project templates should override app templates, and app templates should override framework templates where loaders are involved.

## Files And Uploads

- Block path traversal.
- Avoid overwriting by default.
- Validate size, content type, extension, and image hooks where supported.
- Keep local storage behavior deterministic.

## Static And Fixtures

- Static collection must support duplicate detection, deterministic winner ordering, hashed manifests, and `collectstatic`.
- Fixtures must support JSON, JSONL, XML, custom serializers, natural keys, app/model filtering, indentation, database selection, and transaction options.

