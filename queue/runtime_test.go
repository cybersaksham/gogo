package queue_test

import (
	"errors"
	"testing"

	"github.com/cybersaksham/gogo/queue"
	_ "github.com/cybersaksham/gogo/queue/backends"
	_ "github.com/cybersaksham/gogo/queue/brokers"
)

func TestRuntimeFactoriesCreateMemoryImplementations(t *testing.T) {
	broker, err := queue.NewBrokerFromURL(queue.RuntimeConfig{BrokerURL: "memory://"})
	if err != nil {
		t.Fatalf("NewBrokerFromURL(memory) error = %v", err)
	}
	if broker == nil {
		t.Fatal("NewBrokerFromURL(memory) returned nil broker")
	}

	backend, err := queue.NewResultBackendFromURL(queue.RuntimeConfig{ResultBackend: "memory"})
	if err != nil {
		t.Fatalf("NewResultBackendFromURL(memory) error = %v", err)
	}
	if backend == nil {
		t.Fatal("NewResultBackendFromURL(memory) returned nil backend")
	}

	store, err := queue.NewScheduleStoreFromURL(queue.RuntimeConfig{ScheduleStore: "memory://"})
	if err != nil {
		t.Fatalf("NewScheduleStoreFromURL(memory) error = %v", err)
	}
	if store == nil {
		t.Fatal("NewScheduleStoreFromURL(memory) returned nil store")
	}
}

func TestRuntimeFactoriesRejectUnconfiguredProductionURLs(t *testing.T) {
	if _, err := queue.NewBrokerFromURL(queue.RuntimeConfig{BrokerURL: "redis://localhost:6379/0"}); !errors.Is(err, queue.ErrUnsupportedRuntimeURL) {
		t.Fatalf("NewBrokerFromURL(redis) error = %v, want ErrUnsupportedRuntimeURL", err)
	}
	if _, err := queue.NewResultBackendFromURL(queue.RuntimeConfig{ResultBackend: "redis://localhost:6379/1"}); !errors.Is(err, queue.ErrUnsupportedRuntimeURL) {
		t.Fatalf("NewResultBackendFromURL(redis) error = %v, want ErrUnsupportedRuntimeURL", err)
	}
	if _, err := queue.NewScheduleStoreFromURL(queue.RuntimeConfig{ScheduleStore: "redis://localhost:6379/2"}); !errors.Is(err, queue.ErrUnsupportedRuntimeURL) {
		t.Fatalf("NewScheduleStoreFromURL(redis) error = %v, want ErrUnsupportedRuntimeURL", err)
	}
}
