package queue

import (
	"context"
	"sync"
	"time"
)

type EventType string

const (
	EventWorkerOnline    EventType = "worker.online"
	EventWorkerHeartbeat EventType = "worker.heartbeat"
	EventWorkerOffline   EventType = "worker.offline"
	EventTaskSent        EventType = "task.sent"
	EventTaskReceived    EventType = "task.received"
	EventTaskStarted     EventType = "task.started"
	EventTaskSucceeded   EventType = "task.succeeded"
	EventTaskFailed      EventType = "task.failed"
	EventTaskRetried     EventType = "task.retried"
	EventTaskRevoked     EventType = "task.revoked"
)

type Event struct {
	Type     EventType
	Hostname string
	TaskID   string
	TaskName string
	Queue    string
	State    State
	Error    string
	At       time.Time
	Fields   map[string]any
}

type EventSink interface {
	EmitQueueEvent(context.Context, Event)
}

type EventRecorder struct {
	mu      sync.Mutex
	enabled bool
	events  []Event
}

func NewEventRecorder() *EventRecorder {
	return &EventRecorder{enabled: true}
}

func (r *EventRecorder) EmitQueueEvent(_ context.Context, event Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.enabled {
		return
	}
	if event.At.IsZero() {
		event.At = time.Now().UTC()
	}
	event.Fields = cloneEventFields(event.Fields)
	r.events = append(r.events, event)
}

func (r *EventRecorder) Events() []Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	events := make([]Event, len(r.events))
	for i, event := range r.events {
		event.Fields = cloneEventFields(event.Fields)
		events[i] = event
	}
	return events
}

func (r *EventRecorder) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = nil
}

func (r *EventRecorder) Enable() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enabled = true
}

func (r *EventRecorder) Disable() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enabled = false
	r.events = nil
}

func (r *EventRecorder) Enabled() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.enabled
}

func cloneEventFields(fields map[string]any) map[string]any {
	if fields == nil {
		return nil
	}
	cloned := make(map[string]any, len(fields))
	for key, value := range fields {
		cloned[key] = value
	}
	return cloned
}
