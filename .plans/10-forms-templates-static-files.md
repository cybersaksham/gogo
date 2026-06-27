# Forms Templates Static And Files Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Django-style forms, model forms, widgets, formsets, template rendering, context processors, static files, media files, file storage, and collectstatic.

**Architecture:** Public packages `forms`, `templates`, `static`, and `files` are reusable by admin, APIs, and client applications. Templates and static assets use project override directories before embedded framework assets.

**Tech Stack:** Go templates, form validation, multipart files, embedded assets, filesystem and object-storage abstractions.

---

## Files

- Create: `forms/form.go`
- Create: `forms/fields.go`
- Create: `forms/widgets.go`
- Create: `forms/model_form.go`
- Create: `forms/formset.go`
- Create: `forms/validation.go`
- Create: `forms/errors.go`
- Create: `templates/engine.go`
- Create: `templates/loader.go`
- Create: `templates/context.go`
- Create: `templates/tags.go`
- Create: `templates/filters.go`
- Create: `templates/django_compat.go`
- Create: `static/files.go`
- Create: `static/finders.go`
- Create: `static/collect.go`
- Create: `files/storage.go`
- Create: `files/local.go`
- Create: `files/upload.go`
- Create: `files/upload_handlers.go`
- Modify: `internal/cli/data.go`

## Task 1: Add Form Field System

- [ ] Create `forms/fields.go`.
- [ ] Implement fields:
  - Boolean
  - Char
  - Choice
  - Typed choice
  - Multiple choice
  - Date
  - DateTime
  - Time
  - Duration
  - Decimal
  - Email
  - File
  - Image
  - Float
  - Integer
  - Generic IP address
  - JSON
  - Combo
  - Multi-value
  - Split date time
  - Model choice
  - Model multiple choice
  - Multiple file
  - Regex
  - Slug
  - URL
  - UUID
- [ ] Support required, label, initial, help text, validators, disabled, localize, error messages, and widget override.
- [ ] Add tests for clean values, validation errors, disabled values, and empty values.
- [ ] Run `go test ./forms`.
- [ ] Commit with message `Add Form Fields`.

## Task 2: Add Widgets

- [ ] Create `forms/widgets.go`.
- [ ] Implement widgets:
  - Text input
  - Number input
  - Email input
  - URL input
  - Password input
  - Hidden input
  - Multiple hidden input
  - Textarea
  - Checkbox input
  - Select
  - Select multiple
  - Radio select
  - Checkbox select multiple
  - Date input
  - DateTime input
  - Time input
  - File input
  - Clearable file input
  - Split date time
- [ ] Escape labels, values, and attributes.
- [ ] Support custom attributes and CSS classes.
- [ ] Add snapshot tests for rendered widgets.
- [ ] Run `go test ./forms`.
- [ ] Commit with message `Add Form Widgets`.

## Task 3: Add Form Core

- [ ] Create `forms/form.go`.
- [ ] Create `forms/validation.go`.
- [ ] Create `forms/errors.go`.
- [ ] Implement binding, initial data, changed data, field order, non-field errors, field errors, cleaned data, and valid state.
- [ ] Implement bound fields, hidden fields, visible fields, field groups, form prefixes, form rendering API, and form media assets.
- [ ] Support form-level clean hook.
- [ ] Add tests for binding, invalid data, cleaned data, changed data, non-field errors, and error rendering.
- [ ] Run `go test ./forms`.
- [ ] Commit with message `Add Form Core`.

## Task 4: Add Model Forms

- [ ] Create `forms/model_form.go`.
- [ ] Generate form fields from model metadata.
- [ ] Support include fields, exclude fields, labels, help texts, widgets, field classes, localized fields, and readonly behavior for admin.
- [ ] Implement save with commit true or false.
- [ ] Validate uniqueness through ORM hooks.
- [ ] Add tests for generated fields, save create, save update, uniqueness validation, and excluded fields.
- [ ] Run `go test ./forms ./models`.
- [ ] Commit with message `Add Model Forms`.

## Task 5: Add Formsets

- [ ] Create `forms/formset.go`.
- [ ] Support management form fields:
  - Total forms
  - Initial forms
  - Min forms
  - Max forms
- [ ] Support extra forms, can delete, can order, min validation, max validation, and inline formsets for related models.
- [ ] Add tests for management validation, add, edit, delete, order, min, max, and inline save.
- [ ] Run `go test ./forms`.
- [ ] Commit with message `Add Formsets`.

## Task 6: Add Template Engine

- [ ] Create `templates/engine.go`.
- [ ] Create `templates/context.go`.
- [ ] Wrap Go templates with framework conventions.
- [ ] Support template inheritance through named blocks where possible, partial includes, safe strings, escaping, and context processors.
- [ ] Provide context processors:
  - Request
  - User
  - Messages
  - CSRF token
  - Static URL
  - Media URL
- [ ] Add tests for rendering, escaping, context processors, missing templates, and safe strings.
- [ ] Run `go test ./templates`.
- [ ] Commit with message `Add Template Engine`.

## Task 7: Add Template Loaders Tags And Filters

- [ ] Create `templates/loader.go`.
- [ ] Create `templates/tags.go`.
- [ ] Create `templates/filters.go`.
- [ ] Load templates from:
  - Project directories
  - App template directories
  - Embedded framework templates
- [ ] Project templates must override app templates, and app templates must override framework templates.
- [ ] Add helpers for URL reversing, static URL, media URL, date formatting, default values, length, join, pluralize, line breaks, and safe escape.
- [ ] Add tests for loader precedence and every helper.
- [ ] Run `go test ./templates`.
- [ ] Commit with message `Add Template Loaders And Helpers`.

## Task 8: Add Django-Compatible Template Tags And Filters

- [ ] Create `templates/django_compat.go`.
- [ ] Implement tag equivalents:
  - Autoescape
  - Block
  - Comment
  - CSRF token
  - Cycle
  - Debug
  - Extends
  - Filter
  - FirstOf
  - For
  - For empty
  - If
  - Ifchanged
  - Include
  - Load
  - Lorem
  - Now
  - Querystring
  - Regroup
  - Resetcycle
  - Spaceless
  - Template tag
  - URL
  - Verbatim
  - Widthratio
  - With
- [ ] Implement filter equivalents:
  - Add
  - Addslashes
  - Capfirst
  - Center
  - Cut
  - Date
  - Default
  - Default if none
  - Dictsort
  - Divisibleby
  - Escape
  - EscapeJS
  - Filesizeformat
  - First
  - Floatformat
  - Force escape
  - Get digit
  - Join
  - JSON script
  - Last
  - Length
  - Length is
  - Linebreaks
  - Linebreaks br
  - Linenumbers
  - Ljust
  - Lower
  - Make list
  - Phone2numeric
  - Pluralize
  - Pprint
  - Random
  - Rjust
  - Safe
  - Safeseq
  - Slice
  - Slugify
  - Stringformat
  - Striptags
  - Time
  - Timesince
  - Timeuntil
  - Title
  - Truncatechars
  - Truncatechars html
  - Truncatewords
  - Truncatewords html
  - Unordered list
  - Upper
  - URL encode
  - URLize
  - URLizetrunc
  - Wordcount
  - Wordwrap
  - Yesno
- [ ] Add tests for escaping behavior, inheritance behavior, URL reversing, CSRF output, querystring handling, and every filter.
- [ ] Run `go test ./templates`.
- [ ] Commit with message `Add Django Compatible Template Helpers`.

## Task 9: Add File Storage

- [ ] Create `files/storage.go`.
- [ ] Create `files/local.go`.
- [ ] Create `files/upload.go`.
- [ ] Create `files/upload_handlers.go`.
- [ ] Define storage interface with open, save, delete, exists, list, size, URL, modified time, and path where supported.
- [ ] Implement local filesystem storage.
- [ ] Implement upload handlers for in-memory uploads, temporary-file uploads, chunked uploads, upload interruption, and per-request upload limits.
- [ ] Normalize names, block path traversal, and avoid overwriting by default.
- [ ] Add upload validators for size, content type, extension, and image dimensions hook.
- [ ] Add tests for every storage method, upload handler, path traversal rejection, collision naming, and upload validation.
- [ ] Run `go test ./files`.
- [ ] Commit with message `Add File Storage`.

## Task 10: Add Static Files

- [ ] Create `static/files.go`.
- [ ] Create `static/finders.go`.
- [ ] Create `static/collect.go`.
- [ ] Find static files from project directories, app static directories, and embedded framework assets.
- [ ] Implement hashed manifest storage for production.
- [ ] Implement `collectstatic` command integration.
- [ ] Detect duplicate static paths and deterministic winner ordering.
- [ ] Add tests for finders, duplicate handling, manifest hashing, and collect output.
- [ ] Run `go test ./static`.
- [ ] Commit with message `Add Static Files Pipeline`.

## Task 11: Add Data Fixtures

- [ ] Modify `internal/cli/data.go`.
- [ ] Implement `dumpdata` and `loaddata`.
- [ ] Support JSON, JSONL, XML, and custom registered serializers for fixtures with natural keys for content types and permissions.
- [ ] Support app/model filters, indentation, database selection, and transaction wrapping.
- [ ] Add tests for dumping, loading, natural keys, duplicate handling, and invalid fixture errors.
- [ ] Run `go test ./internal/cli ./models ./orm`.
- [ ] Commit with message `Add Data Fixture Commands`.

## Acceptance Checklist

- [ ] Forms support all common field and widget types.
- [ ] Forms support bound fields, media assets, prefixes, hidden fields, visible fields, and model choice fields.
- [ ] Model forms map from model metadata and can save through ORM.
- [ ] Formsets support admin inline needs.
- [ ] Templates load with correct override precedence.
- [ ] Template tags and filters cover Django-compatible behavior where the Go renderer can support it safely.
- [ ] Static files can be collected with hashed names.
- [ ] File storage blocks unsafe paths and supports memory, temporary-file, and chunked uploads.
- [ ] Fixtures can dump and load model data through JSON, JSONL, XML, and custom serializers.
