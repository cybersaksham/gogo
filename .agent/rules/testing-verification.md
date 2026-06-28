# Testing And Verification Rules

Use this rule before claiming code is complete or safe.

## Standard Commands

Run focused tests first, then broaden based on risk.

```bash
go test ./...
make docs-verify
```

For generated projects and CLI templates:

```bash
go test -tags=integration ./internal/cli
```

For public release readiness:

```bash
make ci
go test -tags=integration ./...
go test -race ./queue/... ./orm/... ./http/...
make bench
```

## External Integrations

- PostgreSQL integration uses `GOGO_TEST_POSTGRES_DSN`.
- Redis integration uses `GOGO_TEST_REDIS_ADDR`.
- RabbitMQ integration uses `GOGO_TEST_RABBITMQ_URL`.
- If these are absent, say which integration coverage was not exercised.

## Reporting

- Do not claim success without fresh command output.
- Include failed commands and exact failure reasons.
- If `govulncheck` is unavailable locally, report that the vulnerability scan was skipped.

