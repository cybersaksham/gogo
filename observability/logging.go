package observability

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Logger interface {
	Log(context.Context, slog.Level, string, ...any)
}

type LogRecord struct {
	Level   slog.Level
	Message string
	Fields  map[string]any
	At      time.Time
}

type LogFilter func(LogRecord) bool

type LoggingConfig struct {
	Level   slog.Level
	Fields  map[string]any
	Filters []LogFilter
	Handler Logger
}

type ConfiguredLogger struct {
	config LoggingConfig
}

func ConfigureLogger(config LoggingConfig) *ConfiguredLogger {
	if config.Handler == nil {
		config.Handler = NoopLogger{}
	}
	config.Fields = cloneFields(config.Fields)
	return &ConfiguredLogger{config: config}
}

func (l *ConfiguredLogger) Log(ctx context.Context, level slog.Level, message string, args ...any) {
	if level < l.config.Level {
		return
	}
	fields := cloneFields(l.config.Fields)
	for i := 0; i+1 < len(args); i += 2 {
		if key, ok := args[i].(string); ok {
			fields[key] = args[i+1]
		}
	}
	record := LogRecord{Level: level, Message: message, Fields: fields, At: time.Now().UTC()}
	for _, filter := range l.config.Filters {
		if filter != nil && !filter(record) {
			return
		}
	}
	l.config.Handler.Log(ctx, level, message, mapToArgs(fields)...)
}

type NoopLogger struct{}

func (NoopLogger) Log(context.Context, slog.Level, string, ...any) {}

type MemoryLogger struct {
	mu      sync.Mutex
	records []LogRecord
}

func NewMemoryLogger() *MemoryLogger {
	return &MemoryLogger{}
}

func (l *MemoryLogger) Log(_ context.Context, level slog.Level, message string, args ...any) {
	fields := map[string]any{}
	for i := 0; i+1 < len(args); i += 2 {
		if key, ok := args[i].(string); ok {
			fields[key] = args[i+1]
		}
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.records = append(l.records, LogRecord{Level: level, Message: message, Fields: fields, At: time.Now().UTC()})
}

func (l *MemoryLogger) Records() []LogRecord {
	l.mu.Lock()
	defer l.mu.Unlock()
	records := make([]LogRecord, len(l.records))
	for i, record := range l.records {
		record.Fields = cloneFields(record.Fields)
		records[i] = record
	}
	return records
}

type Hooks struct {
	Meter  Meter
	Tracer Tracer
	Logger Logger
}

func NoopHooks() Hooks {
	return Hooks{Meter: NoopMeter{}, Tracer: NoopTracer{}, Logger: NoopLogger{}}
}

func normalizeHooks(hooks Hooks) Hooks {
	if hooks.Meter == nil {
		hooks.Meter = NoopMeter{}
	}
	if hooks.Tracer == nil {
		hooks.Tracer = NoopTracer{}
	}
	if hooks.Logger == nil {
		hooks.Logger = NoopLogger{}
	}
	return hooks
}

type HTTPRequestEvent struct {
	Method   string
	Path     string
	Status   int
	Duration time.Duration
}

type ORMQueryEvent struct {
	Operation string
	Table     string
	Duration  time.Duration
}

type MigrationEvent struct {
	Name     string
	App      string
	Duration time.Duration
}

type AdminActionEvent struct {
	Name  string
	Model string
}

type AuthLoginEvent struct {
	Username string
	Success  bool
}

type QueueTaskEvent struct {
	Name     string
	State    string
	Duration time.Duration
}

func InstrumentHTTPRequest(ctx context.Context, hooks Hooks, event HTTPRequestEvent) {
	hooks = normalizeHooks(hooks)
	fields := map[string]any{"method": event.Method, "path": event.Path, "status": event.Status}
	_, span := hooks.Tracer.Start(ctx, "http.request", fields)
	defer span.End()
	hooks.Meter.AddCounter(ctx, "http.requests", 1, fields)
	hooks.Meter.RecordHistogram(ctx, "http.request.duration", float64(event.Duration.Milliseconds()), fields)
	hooks.Logger.Log(ctx, slog.LevelInfo, "http request", mapToArgs(fields)...)
}

func InstrumentORMQuery(ctx context.Context, hooks Hooks, event ORMQueryEvent) {
	instrument(ctx, hooks, "orm.query", "orm.queries", map[string]any{"operation": event.Operation, "table": event.Table}, event.Duration)
}

func InstrumentMigration(ctx context.Context, hooks Hooks, event MigrationEvent) {
	instrument(ctx, hooks, "migration", "migrations", map[string]any{"name": event.Name, "app": event.App}, event.Duration)
}

func InstrumentAdminAction(ctx context.Context, hooks Hooks, event AdminActionEvent) {
	instrument(ctx, hooks, "admin.action", "admin.actions", map[string]any{"name": event.Name, "model": event.Model}, 0)
}

func InstrumentAuthLogin(ctx context.Context, hooks Hooks, event AuthLoginEvent) {
	instrument(ctx, hooks, "auth.login", "auth.logins", map[string]any{"username": event.Username, "success": event.Success}, 0)
}

func InstrumentQueueTask(ctx context.Context, hooks Hooks, event QueueTaskEvent) {
	instrument(ctx, hooks, "queue.task", "queue.tasks", map[string]any{"name": event.Name, "state": event.State}, event.Duration)
}

func instrument(ctx context.Context, hooks Hooks, spanName string, counter string, fields map[string]any, duration time.Duration) {
	hooks = normalizeHooks(hooks)
	_, span := hooks.Tracer.Start(ctx, spanName, fields)
	defer span.End()
	hooks.Meter.AddCounter(ctx, counter, 1, fields)
	if duration > 0 {
		hooks.Meter.RecordHistogram(ctx, counter+".duration", float64(duration.Milliseconds()), fields)
	}
	hooks.Logger.Log(ctx, slog.LevelInfo, spanName, mapToArgs(fields)...)
}

func cloneFields(fields map[string]any) map[string]any {
	if fields == nil {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(fields))
	for key, value := range fields {
		cloned[key] = value
	}
	return cloned
}

func mapToArgs(fields map[string]any) []any {
	args := make([]any, 0, len(fields)*2)
	for key, value := range fields {
		args = append(args, key, value)
	}
	return args
}
