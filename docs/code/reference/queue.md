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
development and test runtimes. A URL such as `redis://` or `amqp://` must be
backed by a registered production factory; otherwise worker and beat startup
returns `ErrUnsupportedRuntimeURL` instead of silently using memory.

Broker packages:

- `queue/brokers` memory broker
- `queue/brokers/redis`
- `queue/brokers/rabbitmq`

Result backend packages:

- `queue/backends` memory backend
- `queue/backends/redis`
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
