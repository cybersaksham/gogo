# Tasks Tutorial

This tutorial covers retries, timeouts, beat scheduling, chains, groups, and chords.

## Register Tasks

```go
app := queue.NewApp(queue.AppOptions{DefaultQueue: "default"})

_, _ = app.RegisterTask("blog.publish_post", func(ctx context.Context, args ...any) (any, error) {
	return "published", nil
}, queue.TaskOptions{
	Queue:             "default",
	MaxRetries:        3,
	RetryBackoff:      true,
	DefaultRetryDelay: time.Minute,
	SoftTimeout:       10 * time.Second,
	HardTimeout:       30 * time.Second,
})
```

`MaxRetries` controls retry attempts. `RetryBackoff` increases retry delay. `SoftTimeout` allows graceful cancellation. `HardTimeout` is the hard execution budget.

## Dispatch Work

```go
signature := queue.NewSignature("blog.publish_post", 42).WithQueue("default")
_ = signature
```

Route signatures through the queue router or publish envelopes through the configured broker.

## Run Workers

```bash
go run manage.go worker --queues default,email
```

Workers consume broker messages, execute registered tasks, store results, emit events, enforce rate limits, and respect ack policy.

## Beat Scheduler

Create periodic schedules and run beat:

```go
store := queue.NewMemoryScheduleStore(queue.MemoryScheduleStoreOptions{})
entry := queue.ScheduleEntry{
	Name:      "publish-drafts",
	Signature: queue.NewSignature("blog.publish_due_drafts"),
	Schedule:  queue.IntervalSchedule{Every: time.Hour},
	Enabled:   true,
}
_ = store.Save(context.Background(), entry)
```

```bash
go run manage.go beat
```

Beat locks due schedule entries, enqueues signatures, and updates last-run metadata.

## Chains, Groups, And Chords

Canvas primitives live in `queue/canvas`.

`Chain` runs tasks sequentially:

```go
workflow := canvas.NewChain(
	canvas.Task(queue.NewSignature("blog.fetch")),
	canvas.Task(queue.NewSignature("blog.render")),
	canvas.Task(queue.NewSignature("blog.publish")),
)
_ = workflow
```

`Group` runs tasks in parallel:

```go
group := canvas.NewGroup(
	canvas.Task(queue.NewSignature("blog.email_author")),
	canvas.Task(queue.NewSignature("blog.email_subscribers")),
)
_ = group
```

`Chord` runs a callback after a group finishes:

```go
group := canvas.NewGroup(
	canvas.Task(queue.NewSignature("blog.email_author")),
	canvas.Task(queue.NewSignature("blog.email_subscribers")),
)
chord := canvas.NewChord(group, canvas.Task(queue.NewSignature("blog.finalize_campaign")))
_ = chord
```

## Retries And Failures

Use retry options on `TaskOptions` for expected transient failures. Store terminal results in the result backend. Inspect failed tasks with:

```bash
go run manage.go inspect
go run manage.go queues
```

## Testing

Use `testing.NewQueueHarness` for eager execution, `Apply` for direct success/failure checks, and `Enqueue` plus `AssertTaskEnqueued` for broker-facing dispatch checks.
