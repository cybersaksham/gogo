package messages

import (
	"context"
	"net/http"
)

type contextKey struct{}

type StorageFactory func(*http.Request) Storage

func Middleware(factory StorageFactory) func(http.Handler) http.Handler {
	if factory == nil {
		factory = func(*http.Request) Storage { return NewMemoryStorage() }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			storage := factory(r)
			ctx := context.WithValue(r.Context(), contextKey{}, storage)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func FromContext(ctx context.Context) (Storage, bool) {
	storage, ok := ctx.Value(contextKey{}).(Storage)
	return storage, ok
}

func Add(ctx context.Context, level Level, text string, extraTags ...string) {
	if storage, ok := FromContext(ctx); ok {
		storage.Add(level, text, extraTags...)
	}
}

func Messages(ctx context.Context) []Message {
	if storage, ok := FromContext(ctx); ok {
		return storage.Messages()
	}
	return nil
}

func Consume(ctx context.Context) []Message {
	if storage, ok := FromContext(ctx); ok {
		return storage.Consume()
	}
	return nil
}

func TemplateContext(ctx context.Context) map[string]any {
	return map[string]any{"messages": Messages(ctx)}
}
