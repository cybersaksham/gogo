# Queue Operations

Gogo includes Celery-style queue primitives for task registration, routing,
retries, time limits, workers, beat scheduling, inspection, revocation, chains,
groups, chunks, maps, and chords.

## Configuration

Set queue environment variables only when queues are enabled. `memory://` is
for local development and tests. Production deploy checks reject memory broker
and result backend or schedule-store URLs. Leave `GOGO_BROKER_URL`,
`GOGO_RESULT_BACKEND`, and `GOGO_SCHEDULE_STORE` empty when the deployment does
not run workers or beat.

```bash
GOGO_BROKER_URL=redis://redis:6379/0
GOGO_RESULT_BACKEND=redis://redis:6379/1
GOGO_SCHEDULE_STORE=redis://redis:6379/2
```

Runtime support:

| Integration | Status |
| --- | --- |
| Memory broker/backend/schedule store | Local development and tests |
| Redis broker/backend/schedule URLs | Real Redis runtime for `redis://` and `rediss://` |
| RabbitMQ broker URLs | Unsupported as runtime factories; no memory fallback |
| SQL result backend URLs | Fail fast unless a SQL backend factory is registered |

Use separate Redis databases, prefixes, or clusters for broker, result backend,
schedule store, and cache when operational isolation matters.

Redis workers use a visibility timeout for in-flight deliveries. If a worker
stops before acking a task, another worker can reclaim it after the timeout and
the attempt count increments. Delayed tasks are held until their ETA before
being consumed. Result backends store terminal state, children, group metadata,
and chord counters in Redis so independent processes can wait for task results.
Redis schedule stores persist beat entries and use Redis locks so only one beat
process enqueues a given schedule at a time.

Validate Redis reachability before deploying workers:

```bash
go run manage.go worker --check \
  --broker-url "$GOGO_BROKER_URL" \
  --result-backend "$GOGO_RESULT_BACKEND"
```

## Worker Deployment

Run workers as a separate process from the web server.

```bash
go run manage.go worker \
  --broker-url "$GOGO_BROKER_URL" \
  --result-backend "$GOGO_RESULT_BACKEND" \
  --queues default \
  --concurrency 4 \
  --prefetch-multiplier 1 \
  --hostname "worker-${HOSTNAME}" \
  --log-level info
```

Important worker options:

- `--queues` selects the queues consumed by the worker.
- `--concurrency` controls concurrent task execution.
- `--autoscale min,max` allows worker pool growth and shrink.
- `--prefetch-multiplier` limits in-flight tasks per worker slot.
- `--max-tasks-per-child` recycles workers after task count.
- `--max-memory-per-child` recycles workers after memory growth.
- `--soft-time-limit` and `--hard-time-limit` set task limits.
- `--accepted-serializers` restricts payload content types.
- `--check` validates worker configuration without processing tasks.
- `--once` processes one task for smoke tests and release checks.

Use graceful shutdown for deploys. Give workers enough termination time to
finish or requeue in-flight tasks according to the configured ack policy.

## Beat Deployment

Run beat as exactly one active scheduler per schedule store unless the store is
configured with locks that prevent duplicate enqueue.

```bash
go run manage.go beat \
  --broker-url "$GOGO_BROKER_URL" \
  --schedule-path "$GOGO_SCHEDULE_STORE" \
  --interval 1s
```

Use `--once` in CI and release checks:

```bash
go run manage.go beat --once
```

Beat needs the same task registry and routing configuration as workers.
Deploying beat without matching workers can build queue backlog.

## Queue Monitoring

Use inspection commands during operations:

```bash
go run manage.go inspect --report
go run manage.go inspect --ping
go run manage.go queues
go run manage.go queues --queue default
```

Monitor:

- Queue ready count.
- In-flight count.
- Worker active tasks.
- Worker processed, succeeded, failed, revoked, retried, acked, and nacked
  counters.
- Oldest queued task age.
- Beat enqueue count and lock errors.
- Chord unlock and callback failures.
- Result backend write and expiry failures.

Enable events when operational visibility is needed:

```bash
go run manage.go inspect --enable-events
```

Disable event collection when it creates too much broker or storage pressure.

## Retries And Timeouts

Every non-idempotent task must define clear retry behavior. Retried tasks must
be safe to run more than once or must use application-level idempotency keys.

Set time limits for tasks that call external services or process user uploads.
Use soft limits for cleanup and hard limits for safety.

Avoid unbounded retry loops. Use backoff, maximum retries, and dead-letter or
manual review queues for repeated failures.

## Queue Security

Restrict accepted serializers to known safe content types. Do not accept
arbitrary binary or executable payloads from untrusted producers.

Broker credentials are secrets. Keep them out of logs and release artifacts.
Use TLS or private networking for broker connections in production.

Task names are part of the public queue contract for deployed producers and
consumers. Rename tasks with compatibility aliases or a drained queue.

## Rollbacks

Queue rollback depends on message compatibility.

- Stop beat before rolling back scheduler or message format changes.
- Drain queues before removing task names.
- Keep old workers running until old messages are processed.
- Purge only when the queued work can be safely discarded.
- Stop workers before database rollback if tasks use the changed schema.

Inspect queue depth before and after rollback:

```bash
go run manage.go queues
go run manage.go inspect --report
```
