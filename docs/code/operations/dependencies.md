# Dependency Policy

Gogo is a framework, so dependencies must not replace the framework surfaces
that the project exists to provide. Add dependencies only when they are
production-grade, actively maintained, and scoped to infrastructure that the
framework should not implement itself.

## Audit Command

List the full dependency graph with:

```bash
go list -m all
```

This command runs in CI through the dependency audit job and the local `deps`
Makefile target. Vulnerability scanning runs separately through `govulncheck`.

## Approved Dependency Categories

Dependencies are acceptable in these categories when they are narrowly scoped
and reviewed.

| Category | Allowed purpose |
| --- | --- |
| Database drivers | Wire protocol and database connectivity, for example PostgreSQL and SQLite drivers. |
| Redis client | Cache backend, broker transport, and result backend connectivity. |
| RabbitMQ client | AMQP broker transport connectivity. |
| Password hashing packages | Well-reviewed password hashing and cryptographic primitives. |
| OpenTelemetry interfaces | Metrics, tracing, context propagation, and instrumentation adapters. |
| YAML parser | Documentation tooling, configuration import/export, or fixture tooling when JSON/TOML is not enough. |
| Standard compression/archive packages | Release artifacts, backups, and import/export helpers. |
| Testing helpers | Test-only assertions, containers, or fixtures when they do not leak into public APIs. |

Approved dependencies must:

- Use a compatible open-source license.
- Avoid global mutable state unless isolated by an adapter.
- Have a clear maintenance signal.
- Avoid network calls during package initialization.
- Avoid panics for normal input errors.
- Avoid logging secrets.
- Be wrapped behind Gogo interfaces when used by public framework features.

## Rejected Dependency Categories

Do not add dependencies that replace Gogo's core product surface.

| Category | Reason rejected |
| --- | --- |
| Web frameworks that replace Gogo HTTP | Routing, middleware, request/response handling, views, and decorators are framework-owned. |
| ORM libraries that replace Gogo ORM | Model metadata, query compilation, transactions, managers, and dialects are framework-owned. |
| Migration libraries that replace Gogo migrations | Autodetection, operations, graph loading, migration records, and writers are framework-owned. |
| Admin generators that replace Gogo admin | Admin sites, registries, views, forms, actions, and permissions are framework-owned. |
| Queue frameworks that replace Gogo queue | Task registry, broker abstraction, worker, beat, retries, chords, chains, groups, and inspection are framework-owned. |
| Authentication frameworks that replace built-in auth | Users, groups, permissions, password hashing, sessions, and auth middleware are framework-owned. |
| Unmaintained transitive-heavy utility bundles | They increase attack surface and make upgrades harder. |
| Packages requiring secrets or telemetry at import time | They are unsafe for libraries and tests. |

Adapters are acceptable when they integrate with an external service without
replacing the framework API. For example, a Redis client is acceptable behind a
Gogo broker implementation; a full external queue framework is not.

## Review Checklist

Before adding a dependency:

- Confirm the standard library cannot reasonably cover the use case.
- Confirm the dependency does not replace a framework-owned package.
- Check license compatibility.
- Check maintenance activity and release history.
- Check known vulnerabilities.
- Check transitive dependency size with `go list -m all`.
- Add a narrow adapter instead of exposing third-party types directly from
  public APIs.
- Add tests for failure behavior and context cancellation.
- Update operations docs when the dependency affects deployment.

## Upgrade Policy

Dependency upgrades should be small and reviewable.

- Security updates take priority.
- Patch updates can ship in patch releases when tests pass.
- Minor updates can ship in minor releases when public behavior is unchanged.
- Major dependency updates require compatibility review and may require a major
  Gogo release if public behavior changes.

Every dependency update must pass:

```bash
go mod tidy
go test ./...
go vet ./...
go list -m all
```

When `govulncheck` is available, run:

```bash
govulncheck ./...
```

## Removal Policy

Remove dependencies when:

- The standard library can replace them.
- The dependency becomes unmaintained.
- A vulnerability cannot be patched promptly.
- The package pulls in unnecessary transitive dependencies.
- The package encourages bypassing Gogo public APIs.

Dependency removal must preserve public API compatibility within the same major
version or provide a documented migration path.
