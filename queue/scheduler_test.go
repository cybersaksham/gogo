package queue

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestIntervalScheduleNextRunHandlesMissedRuns(t *testing.T) {
	start := time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	last := start
	schedule := IntervalSchedule{Every: time.Minute, StartAt: start}
	dueAt, due := schedule.NextRun(&last, start.Add(5*time.Minute))
	if !due || !dueAt.Equal(start.Add(time.Minute)) {
		t.Fatalf("NextRun() = %s, %v", dueAt, due)
	}
}

func TestCrontabScheduleNextRunWithTimezone(t *testing.T) {
	location := time.FixedZone("IST", 5*60*60+30*60)
	schedule := CrontabSchedule{Minute: "30", Hour: "9", DayOfMonth: "*", Month: "*", DayOfWeek: "1", Location: location}
	next, ok := schedule.Next(time.Date(2026, 6, 29, 9, 29, 0, 0, location))
	if !ok || !next.Equal(time.Date(2026, 6, 29, 9, 30, 0, 0, location)) {
		t.Fatalf("Next() = %s, %v", next, ok)
	}
}

func TestClockedScheduleRunsOnce(t *testing.T) {
	runAt := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	schedule := ClockedSchedule{RunAt: runAt}
	dueAt, due := schedule.NextRun(nil, runAt)
	if !due || !dueAt.Equal(runAt) {
		t.Fatalf("first NextRun() = %s, %v", dueAt, due)
	}
	dueAt, due = schedule.NextRun(&runAt, runAt.Add(time.Hour))
	if due || !dueAt.IsZero() {
		t.Fatalf("second NextRun() = %s, %v", dueAt, due)
	}
}

func TestSolarScheduleUsesConfiguredProvider(t *testing.T) {
	location := time.UTC
	expected := time.Date(2026, 6, 28, 6, 15, 0, 0, location)
	schedule := SolarSchedule{
		Event:     SolarSunrise,
		Latitude:  28.6139,
		Longitude: 77.2090,
		Location:  location,
		Provider: solarProviderFunc(func(SolarEvent, float64, float64, time.Time, *time.Location) (time.Time, bool) {
			return expected, true
		}),
	}
	next, ok := schedule.Next(time.Date(2026, 6, 28, 0, 0, 0, 0, location))
	if !ok || !next.Equal(expected) {
		t.Fatalf("Next() = %s, %v", next, ok)
	}
}

func TestMemoryScheduleStoreLockingAndSave(t *testing.T) {
	now := time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	store := NewMemoryScheduleStore(MemoryScheduleStoreOptions{Now: func() time.Time { return now }})
	entry := ScheduleEntry{Name: "hourly", Signature: NewSignature("jobs.run"), Schedule: IntervalSchedule{Every: time.Hour, StartAt: now}, Enabled: true}
	if err := store.Save(context.Background(), entry); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if _, err := store.Lock(context.Background(), "hourly", time.Minute); err != nil {
		t.Fatalf("Lock() error = %v", err)
	}
	if _, err := store.Lock(context.Background(), "hourly", time.Minute); !errors.Is(err, ErrScheduleLocked) {
		t.Fatalf("second Lock() error = %v, want ErrScheduleLocked", err)
	}
	now = now.Add(2 * time.Minute)
	lock, err := store.Lock(context.Background(), "hourly", time.Minute)
	if err != nil {
		t.Fatalf("expired Lock() error = %v", err)
	}
	if err := lock.Release(context.Background()); err != nil {
		t.Fatalf("Release() error = %v", err)
	}
}

type solarProviderFunc func(SolarEvent, float64, float64, time.Time, *time.Location) (time.Time, bool)

func (f solarProviderFunc) NextSolarEvent(event SolarEvent, latitude float64, longitude float64, after time.Time, location *time.Location) (time.Time, bool) {
	return f(event, latitude, longitude, after, location)
}
