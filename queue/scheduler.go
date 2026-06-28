package queue

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ErrScheduleLocked = errors.New("schedule locked")

type Schedule interface {
	NextRun(lastRunAt *time.Time, now time.Time) (time.Time, bool)
}

type NextSchedule interface {
	Next(after time.Time) (time.Time, bool)
}

type IntervalSchedule struct {
	Every   time.Duration
	StartAt time.Time
}

func (s IntervalSchedule) NextRun(lastRunAt *time.Time, now time.Time) (time.Time, bool) {
	if s.Every <= 0 {
		return time.Time{}, false
	}
	if lastRunAt != nil {
		next := lastRunAt.Add(s.Every)
		return next, !next.After(now)
	}
	start := s.StartAt
	if start.IsZero() {
		start = now
	}
	return start, !start.After(now)
}

func (s IntervalSchedule) Next(after time.Time) (time.Time, bool) {
	if s.Every <= 0 {
		return time.Time{}, false
	}
	start := s.StartAt
	if start.IsZero() {
		return after.Add(s.Every), true
	}
	if after.Before(start) {
		return start, true
	}
	elapsed := after.Sub(start)
	steps := elapsed/s.Every + 1
	return start.Add(steps * s.Every), true
}

type CrontabSchedule struct {
	Minute     string
	Hour       string
	DayOfMonth string
	Month      string
	DayOfWeek  string
	Location   *time.Location
}

func (s CrontabSchedule) NextRun(lastRunAt *time.Time, now time.Time) (time.Time, bool) {
	after := now.Add(-time.Minute)
	if lastRunAt != nil {
		after = *lastRunAt
	}
	next, ok := s.Next(after)
	return next, ok && !next.After(now)
}

func (s CrontabSchedule) Next(after time.Time) (time.Time, bool) {
	location := s.Location
	if location == nil {
		location = time.Local
	}
	minutes := parseCronField(s.Minute, 0, 59)
	hours := parseCronField(s.Hour, 0, 23)
	days := parseCronField(s.DayOfMonth, 1, 31)
	months := parseCronField(s.Month, 1, 12)
	weekdays := parseCronField(s.DayOfWeek, 0, 6)
	if len(minutes) == 0 || len(hours) == 0 || len(days) == 0 || len(months) == 0 || len(weekdays) == 0 {
		return time.Time{}, false
	}
	cursor := after.In(location).Truncate(time.Minute).Add(time.Minute)
	deadline := cursor.AddDate(5, 0, 0)
	for cursor.Before(deadline) {
		if containsInt(months, int(cursor.Month())) &&
			containsInt(days, cursor.Day()) &&
			containsInt(weekdays, int(cursor.Weekday())) &&
			containsInt(hours, cursor.Hour()) &&
			containsInt(minutes, cursor.Minute()) {
			return cursor, true
		}
		cursor = cursor.Add(time.Minute)
	}
	return time.Time{}, false
}

type ClockedSchedule struct {
	RunAt time.Time
}

func (s ClockedSchedule) NextRun(lastRunAt *time.Time, now time.Time) (time.Time, bool) {
	if s.RunAt.IsZero() || lastRunAt != nil {
		return time.Time{}, false
	}
	return s.RunAt, !s.RunAt.After(now)
}

func (s ClockedSchedule) Next(after time.Time) (time.Time, bool) {
	if s.RunAt.IsZero() || !s.RunAt.After(after) {
		return time.Time{}, false
	}
	return s.RunAt, true
}

type SolarEvent string

const (
	SolarSunrise SolarEvent = "sunrise"
	SolarSunset  SolarEvent = "sunset"
)

type SolarProvider interface {
	NextSolarEvent(SolarEvent, float64, float64, time.Time, *time.Location) (time.Time, bool)
}

type SolarSchedule struct {
	Event     SolarEvent
	Latitude  float64
	Longitude float64
	Location  *time.Location
	Provider  SolarProvider
}

func (s SolarSchedule) NextRun(lastRunAt *time.Time, now time.Time) (time.Time, bool) {
	after := now.Add(-24 * time.Hour)
	if lastRunAt != nil {
		after = *lastRunAt
	}
	next, ok := s.Next(after)
	return next, ok && !next.After(now)
}

func (s SolarSchedule) Next(after time.Time) (time.Time, bool) {
	location := s.Location
	if location == nil {
		location = time.Local
	}
	provider := s.Provider
	if provider == nil {
		provider = defaultSolarProvider{}
	}
	return provider.NextSolarEvent(s.Event, s.Latitude, s.Longitude, after, location)
}

type ScheduleEntry struct {
	Name          string
	Signature     Signature
	Schedule      Schedule
	Enabled       bool
	OneOff        bool
	LastRunAt     *time.Time
	TotalRunCount int
	Send          SendOptions
}

func (e ScheduleEntry) Clone() ScheduleEntry {
	e.Signature = e.Signature.Clone()
	e.LastRunAt = cloneTime(e.LastRunAt)
	return e
}

type ScheduleStore interface {
	List(context.Context) ([]ScheduleEntry, error)
	Save(context.Context, ScheduleEntry) error
	Lock(context.Context, string, time.Duration) (ScheduleLock, error)
}

type ScheduleLock interface {
	Release(context.Context) error
}

type MemoryScheduleStoreOptions struct {
	Now func() time.Time
}

type MemoryScheduleStore struct {
	mu      sync.Mutex
	now     func() time.Time
	entries map[string]ScheduleEntry
	locks   map[string]time.Time
}

func NewMemoryScheduleStore(options MemoryScheduleStoreOptions) *MemoryScheduleStore {
	now := options.Now
	if now == nil {
		now = time.Now
	}
	return &MemoryScheduleStore{
		now:     now,
		entries: map[string]ScheduleEntry{},
		locks:   map[string]time.Time{},
	}
}

func (s *MemoryScheduleStore) List(context.Context) ([]ScheduleEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	names := make([]string, 0, len(s.entries))
	for name := range s.entries {
		names = append(names, name)
	}
	sort.Strings(names)
	entries := make([]ScheduleEntry, len(names))
	for i, name := range names {
		entries[i] = s.entries[name].Clone()
	}
	return entries, nil
}

func (s *MemoryScheduleStore) Save(_ context.Context, entry ScheduleEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[entry.Name] = entry.Clone()
	return nil
}

func (s *MemoryScheduleStore) Lock(_ context.Context, name string, ttl time.Duration) (ScheduleLock, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if expiresAt, ok := s.locks[name]; ok && expiresAt.After(now) {
		return nil, ErrScheduleLocked
	}
	if ttl <= 0 {
		ttl = time.Minute
	}
	s.locks[name] = now.Add(ttl)
	return memoryScheduleLock{store: s, name: name}, nil
}

type memoryScheduleLock struct {
	store *MemoryScheduleStore
	name  string
}

func (l memoryScheduleLock) Release(context.Context) error {
	l.store.mu.Lock()
	defer l.store.mu.Unlock()
	delete(l.store.locks, l.name)
	return nil
}

type defaultSolarProvider struct{}

func (defaultSolarProvider) NextSolarEvent(event SolarEvent, _ float64, _ float64, after time.Time, location *time.Location) (time.Time, bool) {
	hour := 6
	if event == SolarSunset {
		hour = 18
	}
	local := after.In(location)
	candidate := time.Date(local.Year(), local.Month(), local.Day(), hour, 0, 0, 0, location)
	if !candidate.After(local) {
		candidate = candidate.AddDate(0, 0, 1)
	}
	return candidate, true
}

func parseCronField(expression string, minimum int, maximum int) []int {
	expression = strings.TrimSpace(expression)
	if expression == "" || expression == "*" {
		return intRange(minimum, maximum, 1)
	}
	if strings.HasPrefix(expression, "*/") {
		step, err := strconv.Atoi(strings.TrimPrefix(expression, "*/"))
		if err != nil || step < 1 {
			return nil
		}
		return intRange(minimum, maximum, step)
	}
	parts := strings.Split(expression, ",")
	values := make([]int, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || value < minimum || value > maximum {
			return nil
		}
		values = append(values, value)
	}
	sort.Ints(values)
	return values
}

func intRange(minimum int, maximum int, step int) []int {
	values := make([]int, 0, (maximum-minimum+1)/step)
	for value := minimum; value <= maximum; value += step {
		values = append(values, value)
	}
	return values
}

func containsInt(values []int, value int) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
