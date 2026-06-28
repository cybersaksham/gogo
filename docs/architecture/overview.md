# Architecture Overview

Gogo is organized as a Django-style backend framework for Go. Applications are the unit of structure: each app owns models, migrations, admin registrations, API views, templates, static assets, queue tasks, checks, and optional contrib integrations.

## Package Layers

The framework is layered so client projects import public packages and avoid internal runtime packages.

```mermaid
flowchart TD
    Client["Client project apps"]
    Public["Public framework packages"]
    Internal["Internal command/schema helpers"]
    Drivers["Database, broker, cache, email drivers"]

    Client --> Public
    Public --> Internal
    Public --> Drivers
    Internal --> Drivers
```

Public packages include `app`, `conf`, `http`, `models`, `migrations`, `orm`, `auth`, `admin`, `api`, `queue`, `forms`, `templates`, `cache`, `email`, `files`, `sessions`, `security`, `checks`, `health`, `signals`, `contrib`, and `testing`.

## App Lifecycle

App startup follows the same lifecycle across project apps and contrib apps.

```mermaid
sequenceDiagram
    participant Config as Settings
    participant Registry as App Registry
    participant App as App Config
    participant Checks as System Checks
    participant Server as Runtime

    Config->>Registry: Load InstalledApps
    Registry->>App: Import app config
    App->>Registry: Register models, commands, admin, tasks
    Registry->>Checks: Register app checks
    Checks->>Server: Report blocking diagnostics
    Server->>Registry: Ready
```

Each app config should keep side effects explicit. Model metadata, migrations, admin registration, API routes, template filters, and queue tasks should be discoverable through public app hooks.

## Request Lifecycle

HTTP requests pass through middleware, route resolution, view execution, response rendering, and optional post-response behavior.

```mermaid
sequenceDiagram
    participant Client
    participant Middleware
    participant Router
    participant View
    participant Templates
    participant Response

    Client->>Middleware: Request
    Middleware->>Middleware: Security, sessions, auth, site, messages
    Middleware->>Router: Wrapped request
    Router->>View: Route match and path params
    View->>Templates: Render template or serialize API response
    Templates->>Response: Body and headers
    Response->>Middleware: Post-response hooks
    Middleware->>Client: Response
```

Middleware order is deterministic. Security and host validation should run early. Sessions and auth should run before admin, API authentication, messages, flatpages, and redirects. Redirect middleware should run late because it inspects unresolved 404 responses.

## Model To Migration Flow

Model metadata is the source of truth for migrations.

```mermaid
flowchart LR
    Models["models.Metadata"]
    State["Migration state"]
    Detector["Autodetector"]
    Writer["Migration writer"]
    Executor["Migration executor"]
    Recorder["Migration recorder"]

    Models --> State
    State --> Detector
    Detector --> Writer
    Writer --> Executor
    Executor --> Recorder
```

The migration system compares historical project state with current model metadata, writes deterministic migration files, renders schema SQL through dialect-aware schema editors, applies operations, and records history checksums.

## ORM Query Flow

ORM query objects are immutable until compilation.

```mermaid
flowchart LR
    Manager["Manager"]
    QuerySet["QuerySet"]
    Query["Query state"]
    Compiler["Dialect compiler"]
    SQL["Compiled SQL"]
    DB["database/sql"]

    Manager --> QuerySet
    QuerySet --> Query
    Query --> Compiler
    Compiler --> SQL
    SQL --> DB
```

Managers expose named query entrypoints. QuerySets accumulate filters, ordering, annotations, joins, prefetches, locks, set operations, and write operations without mutating prior QuerySets. The compiler turns query state into SQL for the selected dialect.

## Admin Flow

Admin registration is metadata driven.

```mermaid
flowchart TD
    Model["Model metadata"]
    Admin["ModelAdmin"]
    Registry["Admin registry"]
    Views["Admin views"]
    Auth["Staff permissions"]

    Model --> Registry
    Admin --> Registry
    Registry --> Views
    Auth --> Views
```

ModelAdmin stores list display, filters, search fields, actions, inlines, widgets, readonly fields, fieldsets, and permission hooks. Admin views enforce active staff access and use the same model, form, and ORM primitives available to project apps.

## API Flow

API views reuse request, auth, serializer, pagination, parser, renderer, filtering, throttling, and OpenAPI primitives.

```mermaid
sequenceDiagram
    participant Request
    participant APIView
    participant Auth
    participant Serializer
    participant Renderer

    Request->>APIView: Initialize
    APIView->>Auth: Authenticate and authorize
    APIView->>Serializer: Validate or serialize data
    Serializer->>Renderer: Response payload
    Renderer->>Request: HTTP response
```

ViewSets map actions to routes. Serializers own field conversion and validation. Renderers and parsers keep transport formats separate from business logic.

## Queue Flow

The queue layer follows Celery-style task dispatch and worker execution.

```mermaid
flowchart LR
    Signature["Task signature"]
    Envelope["Broker envelope"]
    Broker["Broker"]
    Worker["Worker"]
    Task["Task function"]
    Backend["Result backend"]
    Events["Events and inspect"]

    Signature --> Envelope
    Envelope --> Broker
    Broker --> Worker
    Worker --> Task
    Task --> Backend
    Worker --> Events
```

Queue apps register tasks by name. Signatures become envelopes with retry, routing, ETA, priority, chord, chain, and group metadata. Workers consume broker messages, enforce rate limits, timeouts, revocation, acknowledgement policy, retries, and result storage.

## Extension Rules

Extensions should use public APIs:

- Apps register through `app.Config` and registry hooks.
- Models expose `models.Metadata`.
- Database features go through dialect interfaces, ORM expressions, migrations, or contrib packages.
- Admin features go through `admin.ModelAdmin`, actions, widgets, and registry APIs.
- API features go through serializers, views, routers, parsers, renderers, auth, permissions, throttles, and OpenAPI metadata.
- Queue features go through task definitions, signatures, brokers, result backends, beat schedules, canvas primitives, events, and inspectors.
- Cross-cutting checks go through `checks.Registry`.

Avoid importing internal command or schema packages from client apps. If a missing extension point requires internal access, add a public boundary first.
