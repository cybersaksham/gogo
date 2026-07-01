# Queue Reference

The queue package provides Celery-style task registration, signatures, envelopes, brokers, result backends, retries, worker execution, beat schedules, routing, rate limits, events, inspect, security, admin metadata, and canvas primitives.

## Public Types

| Area | Types |
| --- | --- |
| App/tasks | `App`, `AppOptions`, `Task`, `TaskFunc`, `TaskDefinition`, `TaskProvider`, `TaskOptions`, `RateLimit`, `AckPolicy` |
| Signatures/envelopes | `Signature`, `SignatureOptions`, `Envelope`, `EnvelopeOptions` |
| State/results | `State`, `Result`, `GroupResult` |
| Broker | `Broker`, `BrokerPublishOptions`, `BrokerConsumeOptions`, `BrokerQueueOptions`, `BrokerMessage`, `BrokerQueueInfo` |
| Backend | `ResultBackend` |
| Worker | `Worker`, `WorkerOptions`, `WorkerLogger`, `WorkerLogEntry`, `WorkerStats`, `ShutdownMode` |
| Pools | `Pool`, `SoloPool`, `GoroutinePool`, `ProcessPool`, `ProcessPoolOptions`, `PoolStrategy` |
| Routing | `Route`, `Router`, `RouterOptions`, `SendOptions` |
| Beat/scheduler | `Beat`, `BeatOptions`, `Schedule`, `NextSchedule`, `IntervalSchedule`, `CrontabSchedule`, `ClockedSchedule`, `SolarSchedule`, `SolarProvider`, `ScheduleEntry`, `ScheduleStore`, `ScheduleLock`, `MemoryScheduleStore` |
| Retry/revocation | `RetryError`, `RevocationRegistry` |
| Rate limiting | `RateLimiter`, `RateLimiterOptions` |
| Events/inspect | `Event`, `EventSink`, `EventRecorder`, `ActiveTask`, `Inspector`, `InspectOptions`, `InspectReport`, `PingResponse` |
| Serialization/security | `Payload`, `Serializer`, `SerializationRegistry`, `SerializationOptions`, `MessageSigner`, `MessageSignerOptions`, `ContentTypeAllowlist`, `BrokerTLSConfig`, `SensitiveValue`, `Redactor`, `RedactorOptions` |
| Admin | `QueueAdminOptions`, `QueueAdminModel`, `QueueAdminView` |
| Runtime factories | `RuntimeConfig`, `NewBrokerFromURL`, `NewResultBackendFromURL`, `NewScheduleStoreFromURL`, `ErrUnsupportedRuntimeURL` |

## Task Options

`TaskOptions` supports serializer, queue, routing key, priority, max retries, default retry delay, retry backoff, retry jitter, soft timeout, hard timeout, rate limit, ack policy, ignore result, and track started.

Ack policies:

- `AckEarly`
- `AckLate`
- `AckManual`

## Worker Options

`WorkerOptions` covers hostname, queues, concurrency, prefetch multiplier, visibility timeout, poll interval, shutdown timeout, ack policy, reject-on-worker-lost, track started, max tasks per child, max memory per child, autoscale, pool, logger, event sink, memory usage hook, revocations, and rate limiter.

## Broker And Backend Implementations

Runtime factories are URL-driven and fail clearly when a configured production
URL has no registered real implementation. `memory` and `memory://` are local
development and test runtimes only; production deploy checks reject them when
they are configured as broker or result backend URLs.

`redis://` and `rediss://` broker URLs create a Redis-backed broker. The broker
stores task envelopes, queue metadata, ready work, delayed work, and in-flight
deliveries in Redis. A worker that exits before ack leaves the delivery in the
in-flight set until the visibility timeout expires; another worker then reclaims
the task with its task ID, headers, group/chord IDs, callbacks, and attempt
metadata preserved.

`redis://` and `rediss://` result backend URLs create a Redis-backed result
backend. Results, children, group metadata, and chord counters are stored in
Redis. `Wait` polls Redis with context and timeout handling, so independent
processes can wait for terminal results.

RabbitMQ route-planning helpers remain available in `queue/brokers/rabbitmq`,
but `amqp://` and `amqps://` runtime URLs are unsupported until a real AMQP
transport is registered. They never fall back to memory.

Broker packages:

- `queue/brokers` memory broker
- `queue/brokers/redis` real Redis broker for `redis://` and `rediss://`
- `queue/brokers/rabbitmq` route-planning helpers; no runtime factory yet

Result backend packages:

- `queue/backends` memory backend
- `queue/backends/redis` real Redis result backend for `redis://` and `rediss://`
- `queue/backends/sql`

## Canvas

Canvas primitives live under `queue/canvas`:

- `Signature`
- `Chain`
- `Group`
- `Chord`
- `Chunks`
- `Map`
- Serialized workflow types and apply results.

## Beat Schedules

Schedule types:

- `IntervalSchedule`
- `CrontabSchedule`
- `ClockedSchedule`
- `SolarSchedule`

`MemoryScheduleStore` provides deterministic tests and local examples.

## States

Task states:

`PENDING`, `RECEIVED`, `STARTED`, `RETRY`, `SUCCESS`, `FAILURE`, `REVOKED`, and `IGNORED`.

Terminal states are success, failure, revoked, and ignored.

## Errors

`ErrDuplicateTask`, `ErrInvalidTask`, `ErrWorkerNotConfigured`, `ErrWorkerRunning`, `ErrWorkerStopped`, `ErrWorkerMemoryLimit`, `ErrTaskNotRegistered`, `ErrQueueEmpty`, `ErrBrokerClosed`, `ErrScheduleLocked`, and retry/security/serialization errors returned by their packages.

## Example

```go
app := queue.NewApp(queue.AppOptions{})
_, err := app.RegisterTask("blog.publish", func(context.Context, ...any) (any, error) {
	return "ok", nil
}, queue.TaskOptions{Queue: "default"})
signature := queue.NewSignature("blog.publish", 1).WithQueue("default")
_, _ = signature, err
```
