package queue

import (
	"context"
	"errors"
	"time"
)

type BeatOptions struct {
	Router  *Router
	Now     func() time.Time
	LockTTL time.Duration
}

type Beat struct {
	app     *App
	broker  Broker
	store   ScheduleStore
	router  *Router
	now     func() time.Time
	lockTTL time.Duration
}

func NewBeat(app *App, broker Broker, store ScheduleStore, options BeatOptions) *Beat {
	now := options.Now
	if now == nil {
		now = time.Now
	}
	lockTTL := options.LockTTL
	if lockTTL == 0 {
		lockTTL = time.Minute
	}
	return &Beat{
		app:     app,
		broker:  broker,
		store:   store,
		router:  options.Router,
		now:     now,
		lockTTL: lockTTL,
	}
}

func (b *Beat) Tick(ctx context.Context) (int, error) {
	if b.app == nil || b.broker == nil || b.store == nil {
		return 0, ErrWorkerNotConfigured
	}
	entries, err := b.store.List(ctx)
	if err != nil {
		return 0, err
	}
	now := b.now()
	enqueued := 0
	for _, entry := range entries {
		if !entry.Enabled || entry.Schedule == nil {
			continue
		}
		dueAt, due := entry.Schedule.NextRun(entry.LastRunAt, now)
		if !due {
			continue
		}
		lock, err := b.store.Lock(ctx, entry.Name, b.lockTTL)
		if errors.Is(err, ErrScheduleLocked) {
			continue
		}
		if err != nil {
			return enqueued, err
		}
		err = b.enqueueEntry(ctx, entry, dueAt)
		releaseErr := lock.Release(ctx)
		if err != nil {
			return enqueued, err
		}
		if releaseErr != nil {
			return enqueued, releaseErr
		}
		enqueued++
	}
	return enqueued, nil
}

func (b *Beat) enqueueEntry(ctx context.Context, entry ScheduleEntry, dueAt time.Time) error {
	send := entry.Send
	if send.Router == nil {
		send.Router = b.router
	}
	if send.CreatedAt.IsZero() {
		send.CreatedAt = b.now().UTC()
	}
	if _, err := b.app.SendTask(ctx, b.broker, entry.Signature, send); err != nil {
		return err
	}
	entry.LastRunAt = &dueAt
	entry.TotalRunCount++
	if entry.OneOff {
		entry.Enabled = false
	}
	return b.store.Save(ctx, entry)
}
