package queue

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

var (
	ErrWorkerNotConfigured = errors.New("worker not configured")
	ErrWorkerRunning       = errors.New("worker already running")
	ErrWorkerStopped       = errors.New("worker stopped")
	ErrWorkerMemoryLimit   = errors.New("worker memory limit exceeded")
	ErrTaskNotRegistered   = errors.New("task not registered")
)

type ShutdownMode string

const (
	GracefulShutdown ShutdownMode = "graceful"
	WarmShutdown     ShutdownMode = "warm"
	ColdShutdown     ShutdownMode = "cold"
)

type WorkerOptions struct {
	Hostname                string
	Queues                  []string
	Concurrency             int
	PrefetchMultiplier      int
	VisibilityTimeout       time.Duration
	PollInterval            time.Duration
	ShutdownTimeout         time.Duration
	AckPolicy               AckPolicy
	RejectOnWorkerLost      bool
	TrackStarted            bool
	MaxTasksPerWorkerChild  int
	MaxMemoryPerWorkerChild uint64
	Autoscale               AutoscaleConfig
	Pool                    Pool
	Logger                  WorkerLogger
	MemoryUsage             func() uint64
	Revocations             *RevocationRegistry
}

type WorkerLogger interface {
	LogWorkerEvent(context.Context, WorkerLogEntry)
}

type WorkerLogEntry struct {
	Event    string
	TaskID   string
	TaskName string
	Queue    string
	Hostname string
	State    State
	Error    string
	At       time.Time
	Fields   map[string]any
}

type WorkerStats struct {
	Hostname                string
	Queues                  []string
	Concurrency             int
	PrefetchLimit           int
	PoolStrategy            PoolStrategy
	RejectOnWorkerLost      bool
	MaxTasksPerWorkerChild  int
	MaxMemoryPerWorkerChild uint64
	Processed               int
	Succeeded               int
	Failed                  int
	Revoked                 int
	Acked                   int
	Nacked                  int
	Recycled                int
	Running                 int
}

type Worker struct {
	app       *App
	broker    Broker
	backend   ResultBackend
	options   WorkerOptions
	autoscale AutoscaleState

	mu         sync.Mutex
	cancel     context.CancelFunc
	done       chan struct{}
	wg         sync.WaitGroup
	started    bool
	processed  int
	succeeded  int
	failed     int
	revoked    int
	acked      int
	nacked     int
	recycled   int
	running    int
	childTasks int
}

func NewWorker(app *App, broker Broker, backend ResultBackend, options WorkerOptions) *Worker {
	if options.Concurrency < 1 {
		options.Concurrency = 1
	}
	if options.PrefetchMultiplier < 1 {
		options.PrefetchMultiplier = 1
	}
	if options.PollInterval == 0 {
		options.PollInterval = 10 * time.Millisecond
	}
	if options.ShutdownTimeout == 0 {
		options.ShutdownTimeout = 30 * time.Second
	}
	if options.AckPolicy == "" {
		if app != nil && app.options.DefaultAckPolicy != "" {
			options.AckPolicy = app.options.DefaultAckPolicy
		} else {
			options.AckPolicy = AckEarly
		}
	}
	if options.Pool == nil {
		options.Pool = NewGoroutinePool()
	}
	if options.Hostname == "" {
		options.Hostname = defaultHostname()
	}
	if len(options.Queues) == 0 {
		queue := "default"
		if app != nil && app.options.DefaultQueue != "" {
			queue = app.options.DefaultQueue
		}
		options.Queues = []string{queue}
	}
	if options.MemoryUsage == nil {
		options.MemoryUsage = currentMemoryUsage
	}
	if options.Revocations == nil {
		options.Revocations = NewRevocationRegistry()
	}
	return &Worker{
		app:       app,
		broker:    broker,
		backend:   backend,
		options:   options,
		autoscale: ResolveAutoscale(options.Concurrency, options.Autoscale),
	}
}

func (w *Worker) Start(ctx context.Context) error {
	if err := w.validate(); err != nil {
		return err
	}
	w.mu.Lock()
	if w.started {
		w.mu.Unlock()
		return ErrWorkerRunning
	}
	runCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	w.done = make(chan struct{})
	w.started = true
	concurrency := w.options.Concurrency
	w.mu.Unlock()

	for i := 0; i < concurrency; i++ {
		w.wg.Add(1)
		go w.loop(runCtx)
	}
	go func() {
		w.wg.Wait()
		_ = w.options.Pool.Close(context.Background())
		w.mu.Lock()
		w.started = false
		w.cancel = nil
		done := w.done
		w.done = nil
		w.mu.Unlock()
		if done != nil {
			close(done)
		}
	}()
	return nil
}

func (w *Worker) Run(ctx context.Context) error {
	if err := w.Start(ctx); err != nil {
		return err
	}
	w.mu.Lock()
	done := w.done
	w.mu.Unlock()
	select {
	case <-ctx.Done():
		_ = w.Shutdown(context.Background(), GracefulShutdown)
		return ctx.Err()
	case <-done:
		return nil
	}
}

func (w *Worker) Shutdown(ctx context.Context, mode ShutdownMode) error {
	w.mu.Lock()
	cancel := w.cancel
	done := w.done
	w.mu.Unlock()
	if cancel == nil {
		return nil
	}
	cancel()
	if mode == ColdShutdown {
		return nil
	}
	if _, ok := ctx.Deadline(); !ok && w.options.ShutdownTimeout > 0 {
		var release context.CancelFunc
		ctx, release = context.WithTimeout(ctx, w.options.ShutdownTimeout)
		defer release()
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *Worker) RunOnce(ctx context.Context) error {
	if err := w.validate(); err != nil {
		return err
	}
	if w.MemoryLimitExceeded() {
		return ErrWorkerMemoryLimit
	}
	for _, queueName := range w.options.Queues {
		message, err := w.broker.Consume(ctx, queueName, BrokerConsumeOptions{VisibilityTimeout: w.options.VisibilityTimeout})
		if err == nil {
			return w.handleMessage(ctx, message)
		}
		if errors.Is(err, ErrQueueEmpty) {
			continue
		}
		return err
	}
	return ErrQueueEmpty
}

func (w *Worker) PrefetchLimit() int {
	return w.options.Concurrency * w.options.PrefetchMultiplier
}

func (w *Worker) TargetConcurrency(readyTasks int) int {
	return w.autoscale.Target(readyTasks)
}

func (w *Worker) MemoryLimitExceeded() bool {
	limit := w.options.MaxMemoryPerWorkerChild
	return limit > 0 && w.options.MemoryUsage() > limit
}

func (w *Worker) Stats() WorkerStats {
	w.mu.Lock()
	defer w.mu.Unlock()
	return WorkerStats{
		Hostname:                w.options.Hostname,
		Queues:                  append([]string(nil), w.options.Queues...),
		Concurrency:             w.options.Concurrency,
		PrefetchLimit:           w.PrefetchLimit(),
		PoolStrategy:            w.options.Pool.Strategy(),
		RejectOnWorkerLost:      w.options.RejectOnWorkerLost,
		MaxTasksPerWorkerChild:  w.options.MaxTasksPerWorkerChild,
		MaxMemoryPerWorkerChild: w.options.MaxMemoryPerWorkerChild,
		Processed:               w.processed,
		Succeeded:               w.succeeded,
		Failed:                  w.failed,
		Revoked:                 w.revoked,
		Acked:                   w.acked,
		Nacked:                  w.nacked,
		Recycled:                w.recycled,
		Running:                 w.running,
	}
}

func (w *Worker) loop(ctx context.Context) {
	defer w.wg.Done()
	for {
		if err := ctx.Err(); err != nil {
			return
		}
		err := w.RunOnce(ctx)
		if err == nil {
			continue
		}
		if errors.Is(err, ErrQueueEmpty) || errors.Is(err, ErrWorkerMemoryLimit) {
			select {
			case <-ctx.Done():
				return
			case <-time.After(w.options.PollInterval):
			}
			continue
		}
		w.log(ctx, WorkerLogEntry{Event: "worker.error", Error: err.Error()})
		select {
		case <-ctx.Done():
			return
		case <-time.After(w.options.PollInterval):
		}
	}
}

func (w *Worker) handleMessage(ctx context.Context, message BrokerMessage) error {
	if w.options.Revocations != nil && w.options.Revocations.IsRevoked(message.Envelope) {
		result := Result{TaskID: message.Envelope.ID, State: StateRevoked, Error: "task revoked"}
		if err := w.backend.StoreResult(ctx, result); err != nil {
			return err
		}
		if err := w.broker.Ack(ctx, message); err != nil {
			return err
		}
		w.recordAck()
		w.recordRevoked()
		w.log(ctx, WorkerLogEntry{Event: "task.revoked", TaskID: message.Envelope.ID, TaskName: message.Envelope.Name, Queue: message.Queue, State: StateRevoked})
		return nil
	}
	task, ok := w.app.Task(message.Envelope.Name)
	if !ok {
		err := fmt.Errorf("%w: %s", ErrTaskNotRegistered, message.Envelope.Name)
		_ = w.backend.StoreResult(ctx, Result{TaskID: message.Envelope.ID, State: StateFailure, Error: err.Error()})
		if nackErr := w.broker.Nack(ctx, message, false); nackErr != nil {
			return nackErr
		}
		w.recordNack()
		return err
	}

	ackPolicy := w.ackPolicy(task)
	if ackPolicy == AckEarly {
		if err := w.broker.Ack(ctx, message); err != nil {
			return err
		}
		w.recordAck()
	}
	if task.Options.TrackStarted || w.options.TrackStarted {
		if err := w.backend.StoreResult(ctx, Result{TaskID: message.Envelope.ID, State: StateStarted}); err != nil {
			return err
		}
		w.log(ctx, WorkerLogEntry{Event: "task.started", TaskID: message.Envelope.ID, TaskName: task.Name, Queue: message.Queue, State: StateStarted})
	}

	taskCtx, cancel, timeoutKind := workerTaskContext(ctx, task.Options)
	defer cancel()
	w.recordRunning(1)
	value, err := w.options.Pool.Run(taskCtx, func(taskCtx context.Context) (any, error) {
		return task.Func(taskCtx, message.Envelope.Args...)
	})
	w.recordRunning(-1)
	err = normalizeTimeoutError(err, timeoutKind)

	if retry, ok := AsRetry(err); ok {
		return w.retryMessage(ctx, message, task, retry, ackPolicy)
	}

	result := Result{TaskID: message.Envelope.ID}
	if task.Options.IgnoreResult {
		result.State = StateIgnored
	} else if err != nil {
		result.State = StateFailure
		result.Error = err.Error()
		result.Traceback = err.Error()
	} else {
		result.State = StateSuccess
		result.Result = value
	}
	if storeErr := w.backend.StoreResult(ctx, result); storeErr != nil {
		return storeErr
	}
	if ackPolicy == AckLate {
		if ackErr := w.broker.Ack(ctx, message); ackErr != nil {
			return ackErr
		}
		w.recordAck()
	}
	if result.State == StateFailure {
		w.recordComplete(false)
		w.log(ctx, WorkerLogEntry{Event: "task.failed", TaskID: message.Envelope.ID, TaskName: task.Name, Queue: message.Queue, State: result.State, Error: result.Error})
		return nil
	}
	w.recordComplete(true)
	w.log(ctx, WorkerLogEntry{Event: "task.succeeded", TaskID: message.Envelope.ID, TaskName: task.Name, Queue: message.Queue, State: result.State})
	return nil
}

func (w *Worker) ackPolicy(task Task) AckPolicy {
	if task.Options.AckPolicy != "" {
		return task.Options.AckPolicy
	}
	return w.options.AckPolicy
}

func (w *Worker) retryMessage(ctx context.Context, message BrokerMessage, task Task, retry *RetryError, ackPolicy AckPolicy) error {
	maxRetries := task.Options.MaxRetries
	if retry.MaxRetries != nil {
		maxRetries = *retry.MaxRetries
	}
	if message.Envelope.Retries >= maxRetries {
		err := retry.Err
		if err == nil {
			err = ErrRetryRequested
		}
		result := Result{TaskID: message.Envelope.ID, State: StateFailure, Error: err.Error(), Traceback: err.Error()}
		if storeErr := w.backend.StoreResult(ctx, result); storeErr != nil {
			return storeErr
		}
		if ackPolicy == AckLate {
			if ackErr := w.broker.Ack(ctx, message); ackErr != nil {
				return ackErr
			}
			w.recordAck()
		}
		w.recordComplete(false)
		w.log(ctx, WorkerLogEntry{Event: "task.retry.exhausted", TaskID: message.Envelope.ID, TaskName: task.Name, Queue: message.Queue, State: StateFailure, Error: result.Error})
		return nil
	}

	delay := retry.Countdown
	if delay == 0 && retry.ETA != nil {
		delay = time.Until(*retry.ETA)
	}
	if delay < 0 {
		delay = 0
	}
	if delay == 0 {
		delay = ComputeRetryDelay(task.Options, message.Envelope.Retries, nil)
	}
	errText := retry.Error()
	if errText == "" {
		errText = ErrRetryRequested.Error()
	}
	if storeErr := w.backend.StoreResult(ctx, Result{TaskID: message.Envelope.ID, State: StateRetry, Error: errText, Traceback: errText}); storeErr != nil {
		return storeErr
	}
	message.Envelope.Retries++
	if err := w.broker.Requeue(ctx, message, delay); err != nil {
		return err
	}
	w.log(ctx, WorkerLogEntry{
		Event:    "task.retry",
		TaskID:   message.Envelope.ID,
		TaskName: task.Name,
		Queue:    message.Queue,
		State:    StateRetry,
		Error:    errText,
		Fields:   map[string]any{"delay": delay.String(), "retries": message.Envelope.Retries},
	})
	return nil
}

func workerTaskContext(ctx context.Context, options TaskOptions) (context.Context, context.CancelFunc, string) {
	if options.HardTimeout > 0 && (options.SoftTimeout == 0 || options.HardTimeout <= options.SoftTimeout) {
		taskCtx, cancel := context.WithTimeout(ctx, options.HardTimeout)
		return taskCtx, cancel, "hard"
	}
	if options.SoftTimeout > 0 {
		taskCtx, cancel := context.WithTimeout(ctx, options.SoftTimeout)
		return taskCtx, cancel, "soft"
	}
	if options.HardTimeout > 0 {
		taskCtx, cancel := context.WithTimeout(ctx, options.HardTimeout)
		return taskCtx, cancel, "hard"
	}
	return ctx, func() {}, ""
}

func normalizeTimeoutError(err error, timeoutKind string) error {
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if timeoutKind == "hard" {
		return ErrHardTimeout
	}
	if timeoutKind == "soft" {
		return ErrSoftTimeout
	}
	return err
}

func (w *Worker) validate() error {
	if w.app == nil || w.broker == nil || w.backend == nil {
		return ErrWorkerNotConfigured
	}
	return nil
}

func (w *Worker) recordRunning(delta int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.running += delta
	if w.running < 0 {
		w.running = 0
	}
}

func (w *Worker) recordAck() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.acked++
}

func (w *Worker) recordNack() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.nacked++
}

func (w *Worker) recordRevoked() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.processed++
	w.revoked++
}

func (w *Worker) recordComplete(success bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.processed++
	w.childTasks++
	if success {
		w.succeeded++
	} else {
		w.failed++
	}
	if w.options.MaxTasksPerWorkerChild > 0 && w.childTasks >= w.options.MaxTasksPerWorkerChild {
		w.recycled++
		w.childTasks = 0
	}
}

func (w *Worker) log(ctx context.Context, entry WorkerLogEntry) {
	if w.options.Logger == nil {
		return
	}
	entry.Hostname = w.options.Hostname
	if entry.At.IsZero() {
		entry.At = time.Now().UTC()
	}
	w.options.Logger.LogWorkerEvent(ctx, entry)
}

func defaultHostname() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return "worker"
	}
	return hostname
}

func currentMemoryUsage() uint64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats.Alloc
}
