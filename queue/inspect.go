package queue

import (
	"context"
	"time"
)

type ActiveTask struct {
	ID        string
	Name      string
	Queue     string
	Hostname  string
	StartedAt time.Time
}

type InspectOptions struct {
	App         *App
	Broker      Broker
	Store       ScheduleStore
	Workers     []*Worker
	Revocations *RevocationRegistry
	Events      *EventRecorder
}

type Inspector struct {
	app         *App
	broker      Broker
	store       ScheduleStore
	workers     []*Worker
	revocations *RevocationRegistry
	events      *EventRecorder
}

type InspectReport struct {
	Registered []Task
	Active     []ActiveTask
	Scheduled  []ScheduleEntry
	Reserved   []BrokerMessage
	Queues     []BrokerQueueInfo
	Workers    []WorkerStats
}

type PingResponse struct {
	OK       bool
	Hostname string
	At       time.Time
}

func NewInspector(options InspectOptions) *Inspector {
	return &Inspector{
		app:         options.App,
		broker:      options.Broker,
		store:       options.Store,
		workers:     append([]*Worker(nil), options.Workers...),
		revocations: options.Revocations,
		events:      options.Events,
	}
}

func (i *Inspector) RegisteredTasks() []Task {
	if i.app == nil {
		return nil
	}
	return i.app.Tasks()
}

func (i *Inspector) ActiveTasks() []ActiveTask {
	var active []ActiveTask
	for _, worker := range i.workers {
		if worker != nil {
			active = append(active, worker.ActiveTasks()...)
		}
	}
	return active
}

func (i *Inspector) ScheduledTasks(ctx context.Context) ([]ScheduleEntry, error) {
	if i.store == nil {
		return nil, nil
	}
	return i.store.List(ctx)
}

func (i *Inspector) ReservedTasks(context.Context) ([]BrokerMessage, error) {
	return nil, nil
}

func (i *Inspector) QueueLengths(ctx context.Context) ([]BrokerQueueInfo, error) {
	if i.broker == nil {
		return nil, nil
	}
	return i.broker.InspectQueues(ctx)
}

func (i *Inspector) WorkerStats() []WorkerStats {
	stats := make([]WorkerStats, 0, len(i.workers))
	for _, worker := range i.workers {
		if worker != nil {
			stats = append(stats, worker.Stats())
		}
	}
	return stats
}

func (i *Inspector) Report(ctx context.Context) (InspectReport, error) {
	scheduled, err := i.ScheduledTasks(ctx)
	if err != nil {
		return InspectReport{}, err
	}
	reserved, err := i.ReservedTasks(ctx)
	if err != nil {
		return InspectReport{}, err
	}
	queues, err := i.QueueLengths(ctx)
	if err != nil {
		return InspectReport{}, err
	}
	return InspectReport{
		Registered: i.RegisteredTasks(),
		Active:     i.ActiveTasks(),
		Scheduled:  scheduled,
		Reserved:   reserved,
		Queues:     queues,
		Workers:    i.WorkerStats(),
	}, nil
}

func (i *Inspector) RevokeTask(taskID string) {
	if i.revocations != nil {
		i.revocations.RevokeTask(taskID)
	}
}

func (i *Inspector) RevokeByStampedHeaders(name string, value string) {
	if i.revocations != nil {
		i.revocations.RevokeStampedHeader(name, value)
	}
}

func (i *Inspector) PoolGrow(worker *Worker, delta int) {
	if worker != nil {
		worker.Grow(delta)
	}
}

func (i *Inspector) PoolShrink(worker *Worker, delta int) {
	if worker != nil {
		worker.Shrink(delta)
	}
}

func (i *Inspector) PoolRestart(ctx context.Context, worker *Worker) error {
	if worker == nil {
		return nil
	}
	return worker.RestartPool(ctx)
}

func (i *Inspector) EnableEvents() {
	if i.events != nil {
		i.events.Enable()
	}
}

func (i *Inspector) DisableEvents() {
	if i.events != nil {
		i.events.Disable()
	}
}

func (i *Inspector) RateLimit(taskName string, limit RateLimit) error {
	return i.app.SetTaskRateLimit(taskName, limit)
}

func (i *Inspector) TimeLimit(taskName string, soft time.Duration, hard time.Duration) error {
	return i.app.SetTaskTimeLimit(taskName, soft, hard)
}

func (i *Inspector) Ping(context.Context) PingResponse {
	return PingResponse{OK: true, Hostname: defaultHostname(), At: time.Now().UTC()}
}

func (i *Inspector) Shutdown(ctx context.Context, worker *Worker, mode ShutdownMode) error {
	if worker == nil {
		return nil
	}
	return worker.Shutdown(ctx, mode)
}
