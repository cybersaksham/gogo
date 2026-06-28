package observability

import (
	"context"
	"sync"
)

type Meter interface {
	AddCounter(context.Context, string, int64, map[string]any)
	RecordHistogram(context.Context, string, float64, map[string]any)
	SetGauge(context.Context, string, float64, map[string]any)
}

type NoopMeter struct{}

func (NoopMeter) AddCounter(context.Context, string, int64, map[string]any)        {}
func (NoopMeter) RecordHistogram(context.Context, string, float64, map[string]any) {}
func (NoopMeter) SetGauge(context.Context, string, float64, map[string]any)        {}

type MemoryMeter struct {
	mu         sync.Mutex
	counters   map[string]int64
	histograms map[string][]float64
	gauges     map[string]float64
}

func NewMemoryMeter() *MemoryMeter {
	return &MemoryMeter{counters: map[string]int64{}, histograms: map[string][]float64{}, gauges: map[string]float64{}}
}

func (m *MemoryMeter) AddCounter(_ context.Context, name string, value int64, _ map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += value
}

func (m *MemoryMeter) RecordHistogram(_ context.Context, name string, value float64, _ map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.histograms[name] = append(m.histograms[name], value)
}

func (m *MemoryMeter) SetGauge(_ context.Context, name string, value float64, _ map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

func (m *MemoryMeter) CounterValue(name string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.counters[name]
}

func (m *MemoryMeter) Counters() map[string]int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	counters := make(map[string]int64, len(m.counters))
	for key, value := range m.counters {
		counters[key] = value
	}
	return counters
}
