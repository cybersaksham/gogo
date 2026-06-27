# Queue Workers Beat And Canvas Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Celery-style queue system with task registration, brokers, workers, result backends, retries, timeouts, routing, scheduling, beat, workflows, monitoring, and admin visibility.

**Architecture:** Public package `queue` owns task APIs, signatures, workers, scheduling, result states, and workflow canvas. Broker and result backend packages are pluggable behind stable interfaces.

**Tech Stack:** Go concurrency, context cancellation, Redis, RabbitMQ, SQL result backend, durable task envelopes, structured events, admin integration.

---

## Files

- Create: `queue/app.go`
- Create: `queue/task.go`
- Create: `queue/signature.go`
- Create: `queue/envelope.go`
- Create: `queue/serialization.go`
- Create: `queue/result.go`
- Create: `queue/state.go`
- Create: `queue/worker.go`
- Create: `queue/pool.go`
- Create: `queue/autoscale.go`
- Create: `queue/retry.go`
- Create: `queue/routing.go`
- Create: `queue/rate_limit.go`
- Create: `queue/scheduler.go`
- Create: `queue/beat.go`
- Create: `queue/events.go`
- Create: `queue/inspect.go`
- Create: `queue/security.go`
- Create: `queue/errors.go`
- Create: `queue/canvas/chain.go`
- Create: `queue/canvas/group.go`
- Create: `queue/canvas/chord.go`
- Create: `queue/canvas/map.go`
- Create: `queue/canvas/chunks.go`
- Create: `queue/brokers/broker.go`
- Create: `queue/brokers/redis/broker.go`
- Create: `queue/brokers/rabbitmq/broker.go`
- Create: `queue/backends/backend.go`
- Create: `queue/backends/redis/backend.go`
- Create: `queue/backends/sql/backend.go`
- Create: `queue/admin.go`
- Modify: `internal/cli/queue.go`

## Task 1: Add Queue App And Task Registry

- [ ] Create `queue/app.go`.
- [ ] Create `queue/task.go`.
- [ ] Define task registration with:
  - Name
  - Function
  - Serializer
  - Queue
  - Routing key
  - Priority
  - Max retries
  - Default retry delay
  - Retry backoff
  - Retry jitter
  - Soft timeout
  - Hard timeout
  - Rate limit
  - Acknowledgement policy
  - Ignore result
  - Track started
- [ ] Discover tasks from installed apps.
- [ ] Reject duplicate task names.
- [ ] Add tests for registration, duplicate detection, defaults, and app discovery.
- [ ] Run `go test ./queue`.
- [ ] Commit with message `Add Queue Task Registry`.

## Task 2: Add Task Envelope And Serialization

- [ ] Create `queue/envelope.go`.
- [ ] Create `queue/signature.go`.
- [ ] Create `queue/serialization.go`.
- [ ] Define task envelope fields:
  - ID
  - Root ID
  - Parent ID
  - Group ID
  - Chord ID
  - Name
  - Args
  - Kwargs
  - Headers
  - Retries
  - ETA
  - Expires
  - Queue
  - Priority
  - ReplyTo
  - CorrelationID
  - CreatedAt
- [ ] Support serializers:
  - JSON
  - Gob for trusted internal deployments
  - Raw bytes
  - Custom registered serializers
- [ ] Support compression:
  - None
  - Gzip
  - Zstd where dependency is approved during implementation
- [ ] Reject untrusted serializers unless explicitly allowed by settings.
- [ ] Implement signatures with immutable options and cloning.
- [ ] Add tests for serialization, compression, trusted serializer enforcement, signatures, countdown, ETA, expires, priority, headers, and clone behavior.
- [ ] Run `go test ./queue`.
- [ ] Commit with message `Add Queue Task Envelopes`.

## Task 3: Add Broker Interface

- [ ] Create `queue/brokers/broker.go`.
- [ ] Define broker methods:
  - Publish
  - Consume
  - Ack
  - Nack
  - Requeue
  - DeclareQueue
  - PurgeQueue
  - InspectQueues
  - Close
- [ ] Support visibility timeout and acknowledgement semantics.
- [ ] Add fake broker tests for publish, consume, ack, nack, requeue, and close.
- [ ] Run `go test ./queue/brokers`.
- [ ] Commit with message `Add Queue Broker Interface`.

## Task 4: Add Redis Broker

- [ ] Create `queue/brokers/redis/broker.go`.
- [ ] Support Redis queues, delayed task sorted set, priority buckets, visibility timeout, and dead-letter list.
- [ ] Add integration tests gated by `GOGO_TEST_REDIS_ADDR`.
- [ ] Add unit tests for key naming and envelope encoding.
- [ ] Run `go test ./queue/brokers/redis`.
- [ ] Commit with message `Add Redis Queue Broker`.

## Task 5: Add RabbitMQ Broker

- [ ] Create `queue/brokers/rabbitmq/broker.go`.
- [ ] Support exchanges, routing keys, durable queues, acknowledgements, negative acknowledgements, priority queues, delayed delivery strategy, and dead-letter exchanges.
- [ ] Add integration tests gated by `GOGO_TEST_RABBITMQ_URL`.
- [ ] Add unit tests for route declaration planning.
- [ ] Run `go test ./queue/brokers/rabbitmq`.
- [ ] Commit with message `Add RabbitMQ Queue Broker`.

## Task 6: Add Result Backend Interface

- [ ] Create `queue/result.go`.
- [ ] Create `queue/state.go`.
- [ ] Create `queue/backends/backend.go`.
- [ ] Define states:
  - Pending
  - Received
  - Started
  - Retry
  - Success
  - Failure
  - Revoked
  - Ignored
- [ ] Define backend methods:
  - StoreResult
  - GetResult
  - Forget
  - Wait
  - Children
  - GroupResult
  - ChordCounter
- [ ] Add tests for state transitions and result expiry.
- [ ] Run `go test ./queue ./queue/backends`.
- [ ] Commit with message `Add Queue Result Backend Interface`.

## Task 7: Add Redis And SQL Result Backends

- [ ] Create `queue/backends/redis/backend.go`.
- [ ] Create `queue/backends/sql/backend.go`.
- [ ] Redis backend stores state, result, traceback, children, expiry, and chord counters.
- [ ] SQL backend stores task results and group results in framework-owned tables.
- [ ] Add migrations for SQL result backend.
- [ ] Add integration tests for both backends.
- [ ] Run `go test ./queue/backends/...`.
- [ ] Commit with message `Add Queue Result Backends`.

## Task 8: Add Worker Runtime

- [ ] Create `queue/worker.go`.
- [ ] Create `queue/pool.go`.
- [ ] Create `queue/autoscale.go`.
- [ ] Support:
  - Concurrency
  - Autoscale minimum and maximum concurrency
  - Prefetch multiplier
  - Graceful shutdown
  - Warm shutdown
  - Cold shutdown
  - Task acknowledgement before or after execution
  - Reject on worker lost
  - Track started
  - Worker hostname
  - Queue subscriptions
  - Max tasks per worker child equivalent
  - Max memory per worker child equivalent where measurable
  - Worker pool strategies for goroutine pool, solo execution, and process-backed execution where supported
  - Structured logs
- [ ] Add tests using fake broker and fake backend for success, failure, ack policy, shutdown, autoscale, max tasks, memory limit behavior, and prefetch.
- [ ] Run `go test ./queue`.
- [ ] Commit with message `Add Queue Worker Runtime`.

## Task 9: Add Retries Timeouts And Revocation

- [ ] Create `queue/retry.go`.
- [ ] Implement retry with max retries, countdown, backoff, jitter, custom retry errors, and retry state storage.
- [ ] Implement soft timeout through context deadline.
- [ ] Implement hard timeout through worker-controlled cancellation boundary.
- [ ] Implement task revocation by task ID and stamped headers.
- [ ] Add tests for retry delay, backoff, jitter, max retries, soft timeout, hard timeout, and revoked tasks.
- [ ] Run `go test ./queue`.
- [ ] Commit with message `Add Queue Retries Timeouts And Revocation`.

## Task 10: Add Routing Rate Limits And Priorities

- [ ] Create `queue/routing.go`.
- [ ] Create `queue/rate_limit.go`.
- [ ] Support:
  - Static task routes
  - Dynamic route functions
  - Queue names
  - Routing keys
  - Priorities
  - Per-task rate limits
  - Worker queue filters
- [ ] Add tests for route selection, default route, dynamic route, priority propagation, and rate limit enforcement.
- [ ] Run `go test ./queue`.
- [ ] Commit with message `Add Queue Routing And Rate Limits`.

## Task 11: Add Scheduler And Beat

- [ ] Create `queue/scheduler.go`.
- [ ] Create `queue/beat.go`.
- [ ] Support schedules:
  - Interval
  - Crontab
  - Solar-style sunrise/sunset hook where location data is configured
  - One-off clocked schedule
- [ ] Support persistent schedule store.
- [ ] Support schedule locking so only one beat instance enqueues a due task.
- [ ] Add tests for next-run computation, missed runs, time zones, locking, and one-off schedules.
- [ ] Run `go test ./queue`.
- [ ] Commit with message `Add Queue Beat Scheduler`.

## Task 12: Add Canvas Workflows

- [ ] Create `queue/canvas/chain.go`.
- [ ] Create `queue/canvas/group.go`.
- [ ] Create `queue/canvas/chord.go`.
- [ ] Create `queue/canvas/map.go`.
- [ ] Create `queue/canvas/chunks.go`.
- [ ] Implement:
  - Chain
  - Group
  - Chord
  - Map
  - Starmap
  - Chunks
  - Callback
  - Errback
  - Link
  - Link error
  - Immutable signatures
- [ ] Add tests for workflow serialization, execution order, group results, chord body execution, errbacks, callbacks, and failure propagation.
- [ ] Run `go test ./queue/canvas ./queue`.
- [ ] Commit with message `Add Queue Canvas Workflows`.

## Task 13: Add Events Inspect And Monitoring

- [ ] Create `queue/events.go`.
- [ ] Create `queue/inspect.go`.
- [ ] Emit events:
  - Worker online
  - Worker heartbeat
  - Worker offline
  - Task sent
  - Task received
  - Task started
  - Task succeeded
  - Task failed
  - Task retried
  - Task revoked
- [ ] Implement inspect commands:
  - Active tasks
  - Scheduled tasks
  - Reserved tasks
  - Registered tasks
  - Queue lengths
  - Worker stats
  - Revoke task
  - Revoke by stamped headers
  - Pool grow
  - Pool shrink
  - Pool restart
  - Enable events
  - Disable events
  - Rate limit
  - Time limit
  - Ping
  - Report
  - Shutdown
- [ ] Add tests for event emission and inspect snapshots.
- [ ] Run `go test ./queue`.
- [ ] Commit with message `Add Queue Events And Inspect`.

## Task 14: Add Queue Message Security

- [ ] Create `queue/security.go`.
- [ ] Support signed task messages with framework secret keys or configured key material.
- [ ] Support accepted content-type allowlists.
- [ ] Support message timestamp validation and replay window checks.
- [ ] Support broker TLS settings validation.
- [ ] Support redaction of sensitive args and kwargs in logs, events, results, and admin.
- [ ] Add tests for valid signature, tampered message, expired message, rejected serializer, TLS settings validation, and redaction.
- [ ] Run `go test ./queue`.
- [ ] Commit with message `Add Queue Message Security`.

## Task 15: Wire Queue CLI

- [ ] Modify `internal/cli/queue.go`.
- [ ] Implement:
  - `gogo worker`
  - `gogo beat`
  - `gogo inspect`
  - `gogo queues`
- [ ] Support flags for app settings, concurrency, autoscale, queues, hostname, log level, broker URL, result backend, beat schedule path, pool strategy, prefetch multiplier, max tasks per child, max memory per child, soft time limit, hard time limit, graceful shutdown timeout, and accepted serializers.
- [ ] Add CLI tests with fake broker/backend.
- [ ] Run `go test ./internal/cli ./queue`.
- [ ] Commit with message `Wire Queue CLI Commands`.

## Task 16: Add Queue Admin Integration

- [ ] Create `queue/admin.go`.
- [ ] Add admin views/models for task results, group results, periodic tasks, interval schedules, crontab schedules, clocked schedules, worker heartbeats, and queue health.
- [ ] Add actions for revoke, retry, purge queue, enable schedule, and disable schedule.
- [ ] Add tests for admin registration metadata and permission checks.
- [ ] Run `go test ./queue ./admin`.
- [ ] Commit with message `Add Queue Admin Integration`.

## Acceptance Checklist

- [ ] Tasks can be registered, enqueued, consumed, acknowledged, retried, timed out, revoked, and inspected.
- [ ] Redis and RabbitMQ brokers are implemented behind the broker interface.
- [ ] Redis and SQL result backends store task and group results.
- [ ] Beat can schedule interval, crontab, solar-style, and one-off tasks.
- [ ] Chain, group, chord, map, starmap, chunks, callbacks, and errbacks work.
- [ ] Worker events and inspect commands expose operational state.
- [ ] Worker autoscale, pool controls, serializer allowlists, compression, message signing, and redaction are implemented.
- [ ] Admin can display and operate queue state.
