package observability

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestMetricsTracingAndInstrumentationHooks(t *testing.T) {
	meter := NewMemoryMeter()
	tracer := NewMemoryTracer()
	logger := NewMemoryLogger()
	hooks := Hooks{Meter: meter, Tracer: tracer, Logger: logger}
	InstrumentHTTPRequest(context.Background(), hooks, HTTPRequestEvent{Method: "GET", Path: "/", Status: 200, Duration: time.Millisecond})
	InstrumentORMQuery(context.Background(), hooks, ORMQueryEvent{Operation: "select", Table: "users", Duration: time.Millisecond})
	InstrumentMigration(context.Background(), hooks, MigrationEvent{Name: "0001_initial", App: "auth", Duration: time.Millisecond})
	InstrumentAdminAction(context.Background(), hooks, AdminActionEvent{Name: "delete_selected", Model: "auth.User"})
	InstrumentAuthLogin(context.Background(), hooks, AuthLoginEvent{Username: "admin", Success: true})
	InstrumentQueueTask(context.Background(), hooks, QueueTaskEvent{Name: "jobs.run", State: "SUCCESS", Duration: time.Millisecond})
	if meter.CounterValue("http.requests") != 1 || meter.CounterValue("queue.tasks") != 1 || len(tracer.Spans()) != 6 || len(logger.Records()) != 6 {
		t.Fatalf("meter=%#v spans=%#v logs=%#v", meter.Counters(), tracer.Spans(), logger.Records())
	}
}

func TestLoggingConfigFiltersFieldsAndNoopDefaults(t *testing.T) {
	memory := NewMemoryLogger()
	logger := ConfigureLogger(LoggingConfig{
		Level:  slog.LevelInfo,
		Fields: map[string]any{"service": "gogo"},
		Filters: []LogFilter{func(record LogRecord) bool {
			return record.Message != "drop"
		}},
		Handler: memory,
	})
	logger.Log(context.Background(), slog.LevelDebug, "debug")
	logger.Log(context.Background(), slog.LevelInfo, "drop")
	logger.Log(context.Background(), slog.LevelInfo, "keep", "request_id", "req-1")
	records := memory.Records()
	if len(records) != 1 || records[0].Fields["service"] != "gogo" || records[0].Fields["request_id"] != "req-1" {
		t.Fatalf("records = %#v", records)
	}
	NoopHooks().Logger.Log(context.Background(), slog.LevelInfo, "noop")
}
