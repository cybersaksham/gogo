package app

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestRegistryReturnsAppsInRegistrationOrderBeforeReady(t *testing.T) {
	registry := NewRegistry()
	mustRegisterApp(t, registry, appConfig(t, "example.accounts", "accounts"))
	mustRegisterApp(t, registry, appConfig(t, "example.blog", "blog"))

	got := appNames(registry.Apps())
	want := []string{"example.accounts", "example.blog"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Apps() = %#v, want %#v", got, want)
	}
}

func TestRegistryOrdersAppsByDependenciesAfterReady(t *testing.T) {
	registry := NewRegistry()
	mustRegisterApp(t, registry, appConfigWithDeps(t, "example.blog", "blog", "example.accounts"))
	mustRegisterApp(t, registry, appConfig(t, "example.accounts", "accounts"))

	if err := registry.Ready(context.Background()); err != nil {
		t.Fatalf("Ready() error = %v", err)
	}

	got := appNames(registry.Apps())
	want := []string{"example.accounts", "example.blog"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Apps() = %#v, want dependency order %#v", got, want)
	}
}

func TestRegistryRejectsDuplicateNamesAndLabels(t *testing.T) {
	registry := NewRegistry()
	mustRegisterApp(t, registry, appConfig(t, "example.blog", "blog"))

	err := registry.Register(appConfig(t, "example.blog", "other"))
	if !errors.Is(err, ErrDuplicateApp) {
		t.Fatalf("Register duplicate name error = %v, want ErrDuplicateApp", err)
	}

	err = registry.Register(appConfig(t, "example.other", "blog"))
	if !errors.Is(err, ErrDuplicateApp) {
		t.Fatalf("Register duplicate label error = %v, want ErrDuplicateApp", err)
	}
}

func TestRegistryLookupsByNameOrLabel(t *testing.T) {
	registry := NewRegistry()
	mustRegisterApp(t, registry, appConfig(t, "example.blog", "blog"))

	if config, ok := registry.Get("example.blog"); !ok || config.Label() != "blog" {
		t.Fatalf("Get(name) = (%v, %v), want blog app", config, ok)
	}
	if config, ok := registry.Get("blog"); !ok || config.Name() != "example.blog" {
		t.Fatalf("Get(label) = (%v, %v), want blog app", config, ok)
	}
	if config, ok := registry.Get("missing"); ok || config != nil {
		t.Fatalf("Get(missing) = (%v, %v), want nil false", config, ok)
	}
}

func TestRegistryMustGetPanicsForMissingApp(t *testing.T) {
	registry := NewRegistry()

	defer func() {
		if recover() == nil {
			t.Fatalf("MustGet() did not panic for missing app")
		}
	}()

	_ = registry.MustGet("missing")
}

func TestRegistryLabelsAreDeterministic(t *testing.T) {
	registry := NewRegistry()
	mustRegisterApp(t, registry, appConfig(t, "example.blog", "blog"))
	mustRegisterApp(t, registry, appConfig(t, "example.accounts", "accounts"))

	got := registry.Labels()
	want := []string{"blog", "accounts"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Labels() = %#v, want %#v", got, want)
	}
}

func TestRegistryReadyFailsForMissingDependency(t *testing.T) {
	registry := NewRegistry()
	mustRegisterApp(t, registry, appConfigWithDeps(t, "example.blog", "blog", "example.accounts"))

	err := registry.Ready(context.Background())
	if !errors.Is(err, ErrMissingDependency) {
		t.Fatalf("Ready() error = %v, want ErrMissingDependency", err)
	}
}

func TestRegistryReadyFailsForDependencyCycle(t *testing.T) {
	registry := NewRegistry()
	mustRegisterApp(t, registry, appConfigWithDeps(t, "example.blog", "blog", "example.accounts"))
	mustRegisterApp(t, registry, appConfigWithDeps(t, "example.accounts", "accounts", "example.blog"))

	err := registry.Ready(context.Background())
	if !errors.Is(err, ErrDependencyCycle) {
		t.Fatalf("Ready() error = %v, want ErrDependencyCycle", err)
	}
}

func TestRegistryRejectsRegisterAfterReady(t *testing.T) {
	registry := NewRegistry()
	mustRegisterApp(t, registry, appConfig(t, "example.blog", "blog"))

	if err := registry.Ready(context.Background()); err != nil {
		t.Fatalf("Ready() error = %v", err)
	}

	err := registry.Register(appConfig(t, "example.accounts", "accounts"))
	if !errors.Is(err, ErrRegistryReady) {
		t.Fatalf("Register() after Ready error = %v, want ErrRegistryReady", err)
	}
}

func mustRegisterApp(t *testing.T, registry *Registry, config Config) {
	t.Helper()

	if err := registry.Register(config); err != nil {
		t.Fatalf("Register(%s) error = %v", config.Name(), err)
	}
}

func appNames(configs []Config) []string {
	names := make([]string, 0, len(configs))
	for _, config := range configs {
		names = append(names, config.Name())
	}
	return names
}

func appConfig(t *testing.T, name, label string) BaseConfig {
	t.Helper()

	return BaseConfig{
		AppName:        name,
		AppLabel:       label,
		AppPath:        t.TempDir(),
		AppVerboseName: label,
	}
}

func appConfigWithDeps(t *testing.T, name, label string, deps ...string) BaseConfig {
	t.Helper()

	config := appConfig(t, name, label)
	config.AppDependencies = deps
	return config
}
