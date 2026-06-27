# API Serializers And OpenAPI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide a built-in API toolkit with routers, request parsing, serializers, validation, permissions, authentication integration, pagination, filtering, throttling, file uploads, and OpenAPI generation.

**Architecture:** Public package `api` builds on framework HTTP, auth, models, ORM, and validation. API handlers can be function views, class-like resources, or model-backed viewsets.

**Tech Stack:** JSON, multipart parsing, model metadata, ORM querysets, auth permissions, OpenAPI 3.1 schema generation.

---

## Files

- Create: `api/router.go`
- Create: `api/view.go`
- Create: `api/viewset.go`
- Create: `api/request.go`
- Create: `api/response.go`
- Create: `api/parsers.go`
- Create: `api/renderers.go`
- Create: `api/serializer.go`
- Create: `api/fields.go`
- Create: `api/model_serializer.go`
- Create: `api/validation.go`
- Create: `api/permissions.go`
- Create: `api/authentication.go`
- Create: `api/pagination.go`
- Create: `api/filters.go`
- Create: `api/throttling.go`
- Create: `api/uploads.go`
- Create: `api/versioning.go`
- Create: `api/metadata.go`
- Create: `api/openapi.go`
- Create: `api/errors.go`
- Create: `api/tests/api_test.go`

## Task 1: Add API Request And Response Types

- [ ] Create `api/request.go`.
- [ ] Create `api/response.go`.
- [ ] Wrap framework HTTP request with parsed body, user, auth, version, accepted renderer, and query params.
- [ ] Provide response helpers for JSON, errors, created, accepted, no content, and file responses.
- [ ] Add normalized error body shape with code, message, field errors, and request ID.
- [ ] Add tests for response encoding and error shapes.
- [ ] Run `go test ./api`.
- [ ] Commit with message `Add API Request And Response Types`.

## Task 2: Add Parsers And Renderers

- [ ] Create `api/parsers.go`.
- [ ] Create `api/renderers.go`.
- [ ] Implement parsers:
  - JSON
  - Form URL encoded
  - Multipart form
  - Plain text
  - Raw bytes
- [ ] Implement renderers:
  - JSON
  - Browsable API metadata renderer if enabled
  - Plain text error renderer
- [ ] Enforce body size limits.
- [ ] Add tests for content negotiation, unsupported media type, invalid JSON, multipart files, and body limits.
- [ ] Run `go test ./api`.
- [ ] Commit with message `Add API Parsers And Renderers`.

## Task 3: Add Serializer Field System

- [ ] Create `api/serializer.go`.
- [ ] Create `api/fields.go`.
- [ ] Implement serializer fields:
  - Boolean
  - Integer
  - Float
  - Decimal
  - String
  - Email
  - URL
  - Slug
  - UUID
  - Date
  - DateTime
  - Time
  - Duration
  - Choice
  - Multiple choice
  - JSON
  - List
  - Dict
  - Nested object
  - Primary key related
  - Slug related
  - Hyperlinked related
  - File
  - Image
  - Read only
  - Write only
  - Method field
- [ ] Support required, allow null, allow blank, default, source, label, help text, validators, and error messages.
- [ ] Add tests for parsing, rendering, validation, defaults, read-only, write-only, nested errors, and source mapping.
- [ ] Run `go test ./api`.
- [ ] Commit with message `Add API Serializer Fields`.

## Task 4: Add Model Serializer

- [ ] Create `api/model_serializer.go`.
- [ ] Generate serializer fields from model metadata.
- [ ] Support include fields, exclude fields, read-only fields, extra kwargs, depth, nested serializers, and relationship fields.
- [ ] Implement create and update from validated data.
- [ ] Add tests for model field mapping, create, update, nested read, read-only fields, and invalid field config.
- [ ] Run `go test ./api ./models`.
- [ ] Commit with message `Add Model Serializer`.

## Task 5: Add API Validation

- [ ] Create `api/validation.go`.
- [ ] Support field validators, object validators, unique validators, unique together validators, and model validation reuse.
- [ ] Return deterministic field error order.
- [ ] Add tests for every validator and combined serializer errors.
- [ ] Run `go test ./api`.
- [ ] Commit with message `Add API Validation`.

## Task 6: Add API Views And ViewSets

- [ ] Create `api/view.go`.
- [ ] Create `api/viewset.go`.
- [ ] Support:
  - Function API views
  - Struct-based API views
  - Model viewsets
  - List
  - Retrieve
  - Create
  - Update
  - Partial update
  - Destroy
  - Custom actions
- [ ] Add request lifecycle hooks:
  - Initialize request
  - Check authentication
  - Check permissions
  - Check throttles
  - Parse body
  - Handle exception
  - Finalize response
- [ ] Add tests for every viewset action and lifecycle failure.
- [ ] Run `go test ./api`.
- [ ] Commit with message `Add API Views And ViewSets`.

## Task 7: Add API Router

- [ ] Create `api/router.go`.
- [ ] Register viewsets and generate routes.
- [ ] Support nested route prefixes, basename, route names, trailing slash config, and route reversing.
- [ ] Add tests for generated routes and custom actions.
- [ ] Run `go test ./api ./http`.
- [ ] Commit with message `Add API Router`.

## Task 8: Add Authentication And Permissions

- [ ] Create `api/authentication.go`.
- [ ] Create `api/permissions.go`.
- [ ] Support session authentication through built-in auth.
- [ ] Support token authentication using framework-owned token model.
- [ ] Support permission classes:
  - Allow any
  - Is authenticated
  - Is admin user
  - Is authenticated or read only
  - Model permissions
  - Object permissions hook
  - Custom permission function
- [ ] Add tests for session auth, token auth, permission denials, safe methods, and object permission checks.
- [ ] Run `go test ./api ./auth`.
- [ ] Commit with message `Add API Authentication And Permissions`.

## Task 9: Add Pagination

- [ ] Create `api/pagination.go`.
- [ ] Implement:
  - Page number pagination
  - Limit offset pagination
  - Cursor pagination
- [ ] Include count, next, previous, and results keys where applicable.
- [ ] Add tests for page boundaries, invalid pages, cursor ordering, and max page size.
- [ ] Run `go test ./api`.
- [ ] Commit with message `Add API Pagination`.

## Task 10: Add Filtering Sorting And Search

- [ ] Create `api/filters.go`.
- [ ] Support:
  - Exact field filters
  - Lookup filters
  - Search fields
  - Ordering fields
  - Custom filter backends
  - Distinct handling for relationship filters
- [ ] Add tests for every filter mode and invalid field rejection.
- [ ] Run `go test ./api ./orm`.
- [ ] Commit with message `Add API Filtering Sorting And Search`.

## Task 11: Add Throttling

- [ ] Create `api/throttling.go`.
- [ ] Support:
  - Anonymous user rate throttle
  - Authenticated user rate throttle
  - Scoped rate throttle
  - Custom throttle store
- [ ] Return retry-after headers.
- [ ] Add tests for rate windows, user keys, anonymous keys, scopes, and retry headers.
- [ ] Run `go test ./api`.
- [ ] Commit with message `Add API Throttling`.

## Task 12: Add Uploads

- [ ] Create `api/uploads.go`.
- [ ] Support multipart file upload, streamed upload, image validation hook, file size limits, extension allowlists, and storage backend integration.
- [ ] Block path traversal and executable filename surprises by normalizing stored names.
- [ ] Add tests for upload success, size rejection, extension rejection, path traversal rejection, and storage errors.
- [ ] Run `go test ./api`.
- [ ] Commit with message `Add API Upload Handling`.

## Task 13: Add API Versioning And Metadata

- [ ] Create `api/versioning.go`.
- [ ] Implement versioning strategies:
  - URL path versioning
  - Namespace versioning
  - Host name versioning
  - Query parameter versioning
  - Accept header versioning
- [ ] Enforce allowed versions and default version settings.
- [ ] Create `api/metadata.go`.
- [ ] Generate endpoint metadata for browsable API, OPTIONS responses, forms, serializers, actions, filters, pagination, authentication, permissions, and throttles.
- [ ] Add tests for every versioning strategy, invalid version handling, default version behavior, and metadata output.
- [ ] Run `go test ./api`.
- [ ] Commit with message `Add API Versioning And Metadata`.

## Task 14: Add OpenAPI Generation

- [ ] Create `api/openapi.go`.
- [ ] Generate OpenAPI 3.1 schema from routers, serializers, model metadata, auth schemes, versioning, pagination, filters, throttling, uploads, metadata, and error responses.
- [ ] Provide JSON endpoint and optional admin documentation link.
- [ ] Add tests with golden OpenAPI output for list, retrieve, create, update, custom action, auth, and error responses.
- [ ] Run `go test ./api`.
- [ ] Commit with message `Add OpenAPI Generation`.

## Acceptance Checklist

- [ ] API can expose model CRUD through viewsets.
- [ ] Serializers support primitive, relationship, nested, file, read-only, write-only, and method fields.
- [ ] Validation errors are structured and deterministic.
- [ ] Authentication and permissions integrate with built-in auth.
- [ ] Pagination, filtering, searching, ordering, throttling, and uploads are implemented.
- [ ] Versioning and endpoint metadata are implemented.
- [ ] OpenAPI output covers all registered API routes.
