package queue

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

var ErrUnsupportedRuntimeURL = errors.New("unsupported queue runtime URL")

type RuntimeConfig struct {
	BrokerURL     string
	ResultBackend string
	ScheduleStore string
}

type BrokerFactory func(RuntimeConfig) (Broker, error)
type ResultBackendFactory func(RuntimeConfig) (ResultBackend, error)
type ScheduleStoreFactory func(RuntimeConfig) (ScheduleStore, error)

var runtimeFactories = struct {
	sync.RWMutex
	brokers        map[string]BrokerFactory
	resultBackends map[string]ResultBackendFactory
	scheduleStores map[string]ScheduleStoreFactory
}{
	brokers:        map[string]BrokerFactory{},
	resultBackends: map[string]ResultBackendFactory{},
	scheduleStores: map[string]ScheduleStoreFactory{},
}

func init() {
	RegisterScheduleStoreFactory("memory", func(RuntimeConfig) (ScheduleStore, error) {
		return NewMemoryScheduleStore(MemoryScheduleStoreOptions{}), nil
	})
}

func RegisterBrokerFactory(scheme string, factory BrokerFactory) {
	scheme = normalizeRuntimeScheme(scheme)
	if scheme == "" || factory == nil {
		panic("queue: broker factory requires scheme and factory")
	}
	runtimeFactories.Lock()
	defer runtimeFactories.Unlock()
	runtimeFactories.brokers[scheme] = factory
}

func RegisterResultBackendFactory(scheme string, factory ResultBackendFactory) {
	scheme = normalizeRuntimeScheme(scheme)
	if scheme == "" || factory == nil {
		panic("queue: result backend factory requires scheme and factory")
	}
	runtimeFactories.Lock()
	defer runtimeFactories.Unlock()
	runtimeFactories.resultBackends[scheme] = factory
}

func RegisterScheduleStoreFactory(scheme string, factory ScheduleStoreFactory) {
	scheme = normalizeRuntimeScheme(scheme)
	if scheme == "" || factory == nil {
		panic("queue: schedule store factory requires scheme and factory")
	}
	runtimeFactories.Lock()
	defer runtimeFactories.Unlock()
	runtimeFactories.scheduleStores[scheme] = factory
}

func NewBrokerFromURL(config RuntimeConfig) (Broker, error) {
	scheme, err := runtimeScheme(config.BrokerURL, "memory")
	if err != nil {
		return nil, err
	}
	runtimeFactories.RLock()
	factory := runtimeFactories.brokers[scheme]
	runtimeFactories.RUnlock()
	if factory == nil {
		return nil, fmt.Errorf("%w: no broker factory registered for %q URL %q", ErrUnsupportedRuntimeURL, scheme, config.BrokerURL)
	}
	return factory(config)
}

func NewResultBackendFromURL(config RuntimeConfig) (ResultBackend, error) {
	scheme, err := runtimeScheme(config.ResultBackend, "memory")
	if err != nil {
		return nil, err
	}
	runtimeFactories.RLock()
	factory := runtimeFactories.resultBackends[scheme]
	runtimeFactories.RUnlock()
	if factory == nil {
		return nil, fmt.Errorf("%w: no result backend factory registered for %q URL %q", ErrUnsupportedRuntimeURL, scheme, config.ResultBackend)
	}
	return factory(config)
}

func NewScheduleStoreFromURL(config RuntimeConfig) (ScheduleStore, error) {
	scheme, err := runtimeScheme(config.ScheduleStore, "memory")
	if err != nil {
		return nil, err
	}
	runtimeFactories.RLock()
	factory := runtimeFactories.scheduleStores[scheme]
	runtimeFactories.RUnlock()
	if factory == nil {
		return nil, fmt.Errorf("%w: no schedule store factory registered for %q URL %q", ErrUnsupportedRuntimeURL, scheme, config.ScheduleStore)
	}
	return factory(config)
}

func runtimeScheme(raw string, fallback string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback, nil
	}
	if !strings.Contains(value, "://") {
		return normalizeRuntimeScheme(value), nil
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("%w: parse %q: %v", ErrUnsupportedRuntimeURL, raw, err)
	}
	scheme := normalizeRuntimeScheme(parsed.Scheme)
	if scheme == "" {
		return "", fmt.Errorf("%w: missing URL scheme in %q", ErrUnsupportedRuntimeURL, raw)
	}
	return scheme, nil
}

func normalizeRuntimeScheme(scheme string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimSuffix(scheme, "://")))
}
