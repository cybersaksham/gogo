package app

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestReadyRunsEachAppOnceInDependencyOrder(t *testing.T) {
	var calls []string
	registry := NewRegistry()
	mustRegisterApp(t, registry, lifecycleConfig{
		BaseConfig: appConfigWithDeps(t, "example.blog", "blog", "example.accounts"),
		ready:      func(context.Context, *Registry) error { calls = append(calls, "blog"); return nil },
	})
	mustRegisterApp(t, registry, lifecycleConfig{
		BaseConfig: appConfig(t, "example.accounts", "accounts"),
		ready:      func(context.Context, *Registry) error { calls = append(calls, "accounts"); return nil },
	})

	if err := registry.Ready(context.Background()); err != nil {
		t.Fatalf("Ready() error = %v", err)
	}
	if err := registry.Ready(context.Background()); err != nil {
		t.Fatalf("Ready() second call error = %v", err)
	}

	want := []string{"accounts", "blog"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("ready calls = %#v, want %#v", calls, want)
	}
}

func TestReadyFailureDoesNotMarkRegistryReady(t *testing.T) {
	readyErr := errors.New("ready failed")
	registry := NewRegistry()
	mustRegisterApp(t, registry, lifecycleConfig{
		BaseConfig: appConfig(t, "example.blog", "blog"),
		ready:      func(context.Context, *Registry) error { return readyErr },
	})

	err := registry.Ready(context.Background())
	if !errors.Is(err, readyErr) {
		t.Fatalf("Ready() error = %v, want readyErr", err)
	}

	err = registry.Register(appConfig(t, "example.accounts", "accounts"))
	if err != nil {
		t.Fatalf("Register() after failed Ready error = %v, want registry to remain mutable", err)
	}
}

func TestShutdownAfterPartialReadyShutsDownStartedApps(t *testing.T) {
	readyErr := errors.New("ready failed")
	var shutdowns []string
	registry := NewRegistry()
	mustRegisterApp(t, registry, lifecycleConfig{
		BaseConfig: appConfig(t, "example.accounts", "accounts"),
		ready:      func(context.Context, *Registry) error { return nil },
		shutdown:   func(context.Context) error { shutdowns = append(shutdowns, "accounts"); return nil },
	})
	mustRegisterApp(t, registry, lifecycleConfig{
		BaseConfig: appConfig(t, "example.blog", "blog"),
		ready:      func(context.Context, *Registry) error { return readyErr },
		shutdown:   func(context.Context) error { shutdowns = append(shutdowns, "blog"); return nil },
	})

	err := registry.Ready(context.Background())
	if !errors.Is(err, readyErr) {
		t.Fatalf("Ready() error = %v, want readyErr", err)
	}

	if err := registry.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	want := []string{"accounts"}
	if !reflect.DeepEqual(shutdowns, want) {
		t.Fatalf("shutdowns = %#v, want %#v", shutdowns, want)
	}
}

func TestReadyWithCanceledContextStopsBeforeHooks(t *testing.T) {
	calls := 0
	registry := NewRegistry()
	mustRegisterApp(t, registry, lifecycleConfig{
		BaseConfig: appConfig(t, "example.blog", "blog"),
		ready:      func(context.Context, *Registry) error { calls++; return nil },
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := registry.Ready(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Ready() error = %v, want context.Canceled", err)
	}
	if calls != 0 {
		t.Fatalf("ready calls = %d, want 0", calls)
	}
}

type lifecycleConfig struct {
	BaseConfig
	ready    func(context.Context, *Registry) error
	shutdown func(context.Context) error
}

func (c lifecycleConfig) Ready(ctx context.Context, registry *Registry) error {
	if c.ready == nil {
		return nil
	}
	return c.ready(ctx, registry)
}

func (c lifecycleConfig) Shutdown(ctx context.Context) error {
	if c.shutdown == nil {
		return nil
	}
	return c.shutdown(ctx)
}
