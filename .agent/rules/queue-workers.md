# Queue Worker Rules

Use this rule for `queue`, brokers, result backends, workers, beat, canvas, and queue CLI.

## Queue Contracts

- Task names must be stable and namespaced.
- Envelope serialization must remain backward compatible.
- Broker and backend interfaces must support deterministic tests with memory implementations.
- Redis, RabbitMQ, and SQL implementations must degrade safely when integration env vars are absent.

## Workers

- Preserve acknowledgement behavior, retry semantics, timeouts, revocation, rate limits, priorities, events, and result storage.
- Do not introduce goroutine leaks.
- Run race tests for worker, broker, backend, scheduler, and canvas changes.

## Beat And Canvas

- Schedules must be deterministic with injectable clocks.
- Chains, groups, chords, maps, starmaps, and chunks must preserve task IDs, group IDs, chord IDs, callbacks, and errbacks.

## Verification

```bash
go test ./queue/...
go test -race ./queue/...
```

