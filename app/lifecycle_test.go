package app

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
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

func TestReadyAllowsGeneratedAppResourceRegistration(t *testing.T) {
	registry := NewRegistry()
	mustRegisterApp(t, registry, lifecycleConfig{
		BaseConfig: appConfig(t, "example.blog", "blog"),
		ready: func(context.Context, *Registry) error {
			registry.RegisterModel(ModelResource{AppLabel: "blog", Name: "Post"})
			registry.RegisterAdmin(AdminResource{AppLabel: "blog", ModelName: "Post", Handler: "RegisterAdmin"})
			registry.RegisterRoute(RouteResource{AppLabel: "blog", Name: "blog:index", Path: "/blog/", Handler: "RegisterRoutes"})
			registry.RegisterAPIRoute(APIRouteResource{AppLabel: "blog", Name: "blog-post-list", Path: "/api/blog/posts/", Handler: "RegisterAPI"})
			registry.RegisterForm(FormResource{AppLabel: "blog", Name: "PostForm", Handler: "NewPostForm"})
			registry.RegisterTemplate(TemplateResource{AppLabel: "blog", Path: "templates/blog"})
			registry.RegisterStaticRoot(StaticResource{AppLabel: "blog", Path: "static/blog"})
			registry.RegisterTask(TaskResource{AppLabel: "blog", Name: "blog.example", Handler: "RegisterTasks"})
			registry.RegisterMigration(MigrationResource{AppLabel: "blog", Name: "0001_initial"})
			return nil
		},
	})

	done := make(chan error, 1)
	go func() {
		done <- registry.Ready(context.Background())
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Ready() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Ready() timed out while generated app registered resources")
	}

	if got := registry.Models(); len(got) != 1 || got[0].Name != "Post" {
		t.Fatalf("Models() = %#v, want blog Post", got)
	}
	if got := registry.Tasks(); len(got) != 1 || got[0].Name != "blog.example" {
		t.Fatalf("Tasks() = %#v, want blog.example", got)
	}
}

func TestReadyBlocksNewAppRegistrationDuringReady(t *testing.T) {
	readyStarted := make(chan struct{})
	releaseReady := make(chan struct{})
	registry := NewRegistry()
	mustRegisterApp(t, registry, lifecycleConfig{
		BaseConfig: appConfig(t, "example.blog", "blog"),
		ready: func(context.Context, *Registry) error {
			close(readyStarted)
			<-releaseReady
			return nil
		},
	})

	done := make(chan error, 1)
	go func() {
		done <- registry.Ready(context.Background())
	}()
	<-readyStarted

	err := registry.Register(appConfig(t, "example.accounts", "accounts"))
	if !errors.Is(err, ErrRegistryReady) {
		t.Fatalf("Register() during Ready error = %v, want ErrRegistryReady", err)
	}

	close(releaseReady)
	if err := <-done; err != nil {
		t.Fatalf("Ready() error = %v", err)
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

func TestShutdownRunsEachReadyAppExactlyOnce(t *testing.T) {
	var calls int
	registry := NewRegistry()
	mustRegisterApp(t, registry, lifecycleConfig{
		BaseConfig: appConfig(t, "example.blog", "blog"),
		ready:      func(context.Context, *Registry) error { return nil },
		shutdown:   func(context.Context) error { calls++; return nil },
	})

	if err := registry.Ready(context.Background()); err != nil {
		t.Fatalf("Ready() error = %v", err)
	}
	if err := registry.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if err := registry.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() second call error = %v", err)
	}
	if calls != 1 {
		t.Fatalf("shutdown calls = %d, want 1", calls)
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
