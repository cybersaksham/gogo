package observability

import (
	"context"
	"sync"
	"time"
)

type Tracer interface {
	Start(context.Context, string, map[string]any) (context.Context, Span)
}

type Span interface {
	SetAttribute(string, any)
	RecordError(error)
	End()
}

type NoopTracer struct{}
type noopSpan struct{}

func (NoopTracer) Start(ctx context.Context, _ string, _ map[string]any) (context.Context, Span) {
	return ctx, noopSpan{}
}
func (noopSpan) SetAttribute(string, any) {}
func (noopSpan) RecordError(error)        {}
func (noopSpan) End()                     {}

type MemoryTracer struct {
	mu    sync.Mutex
	spans []SpanRecord
}

type SpanRecord struct {
	Name       string
	Fields     map[string]any
	Errors     []string
	StartedAt  time.Time
	FinishedAt time.Time
}

func NewMemoryTracer() *MemoryTracer {
	return &MemoryTracer{}
}

func (t *MemoryTracer) Start(ctx context.Context, name string, fields map[string]any) (context.Context, Span) {
	record := SpanRecord{Name: name, Fields: cloneFields(fields), StartedAt: time.Now().UTC()}
	return ctx, &memorySpan{tracer: t, record: record}
}

func (t *MemoryTracer) Spans() []SpanRecord {
	t.mu.Lock()
	defer t.mu.Unlock()
	spans := make([]SpanRecord, len(t.spans))
	copy(spans, t.spans)
	return spans
}

type memorySpan struct {
	tracer *MemoryTracer
	record SpanRecord
}

func (s *memorySpan) SetAttribute(key string, value any) {
	if s.record.Fields == nil {
		s.record.Fields = map[string]any{}
	}
	s.record.Fields[key] = value
}

func (s *memorySpan) RecordError(err error) {
	if err != nil {
		s.record.Errors = append(s.record.Errors, err.Error())
	}
}

func (s *memorySpan) End() {
	s.record.FinishedAt = time.Now().UTC()
	s.tracer.mu.Lock()
	defer s.tracer.mu.Unlock()
	s.tracer.spans = append(s.tracer.spans, s.record)
}

func InjectTraceContext(ctx context.Context, headers map[string]string) {
	if headers != nil {
		headers["traceparent"] = "00-noop-noop-00"
	}
}

func ExtractTraceContext(ctx context.Context, _ map[string]string) context.Context {
	return ctx
}
