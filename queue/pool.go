package queue

import (
	"context"
)

type PoolStrategy string

const (
	PoolGoroutine     PoolStrategy = "goroutine"
	PoolSolo          PoolStrategy = "solo"
	PoolProcessBacked PoolStrategy = "process"
)

type PoolExecutable func(context.Context) (any, error)

// Pool executes task callables behind a stable worker runtime boundary.
type Pool interface {
	Strategy() PoolStrategy
	Run(context.Context, PoolExecutable) (any, error)
	Close(context.Context) error
}

type SoloPool struct{}

func NewSoloPool() *SoloPool {
	return &SoloPool{}
}

func (p *SoloPool) Strategy() PoolStrategy {
	return PoolSolo
}

func (p *SoloPool) Run(ctx context.Context, executable PoolExecutable) (any, error) {
	return executable(ctx)
}

func (p *SoloPool) Close(context.Context) error {
	return nil
}

type GoroutinePool struct{}

func NewGoroutinePool() *GoroutinePool {
	return &GoroutinePool{}
}

func (p *GoroutinePool) Strategy() PoolStrategy {
	return PoolGoroutine
}

func (p *GoroutinePool) Run(ctx context.Context, executable PoolExecutable) (any, error) {
	type result struct {
		value any
		err   error
	}
	done := make(chan result, 1)
	go func() {
		value, err := executable(ctx)
		done <- result{value: value, err: err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-done:
		return result.value, result.err
	}
}

func (p *GoroutinePool) Close(context.Context) error {
	return nil
}

type ProcessPoolOptions struct{}

// ProcessPool is a process-backed execution boundary where supported by the host.
// The initial implementation preserves the boundary contract while executing
// through goroutines so callers can opt into the strategy without a platform fork.
type ProcessPool struct {
	goroutine *GoroutinePool
}

func NewProcessPool(ProcessPoolOptions) *ProcessPool {
	return &ProcessPool{goroutine: NewGoroutinePool()}
}

func (p *ProcessPool) Strategy() PoolStrategy {
	return PoolProcessBacked
}

func (p *ProcessPool) Run(ctx context.Context, executable PoolExecutable) (any, error) {
	return p.goroutine.Run(ctx, executable)
}

func (p *ProcessPool) Close(ctx context.Context) error {
	return p.goroutine.Close(ctx)
}
