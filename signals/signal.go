package signals

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type Receiver[T any] func(context.Context, any, T) error

type Options struct {
	RecoverPanics bool
	Async         func(context.Context, func(context.Context) []Response) []Response
}

type Response struct {
	ReceiverID uint64
	Error      error
}

type Signal[T any] struct {
	mu        sync.RWMutex
	name      string
	options   Options
	receivers []receiver[T]
	nextID    uint64
}

type receiver[T any] struct {
	id     uint64
	sender any
	fn     Receiver[T]
	active bool
}

type Handle struct {
	disconnect func() bool
}

func (h Handle) Disconnect() bool {
	if h.disconnect == nil {
		return false
	}
	return h.disconnect()
}

func New[T any](name string, options ...Options) *Signal[T] {
	var option Options
	if len(options) > 0 {
		option = options[0]
	}
	return &Signal[T]{name: name, options: option}
}

func (s *Signal[T]) Name() string {
	return s.name
}

func (s *Signal[T]) Connect(sender any, fn Receiver[T]) Handle {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	id := s.nextID
	s.receivers = append(s.receivers, receiver[T]{id: id, sender: sender, fn: fn, active: true})
	return Handle{disconnect: func() bool { return s.disconnect(id) }}
}

func (s *Signal[T]) Send(ctx context.Context, sender any, payload T) []Response {
	receivers := s.matching(sender)
	responses := make([]Response, 0, len(receivers))
	for _, receiver := range receivers {
		response := Response{ReceiverID: receiver.id}
		response.Error = s.call(ctx, receiver, sender, payload)
		responses = append(responses, response)
	}
	return responses
}

func (s *Signal[T]) SendAsync(ctx context.Context, sender any, payload T) []Response {
	if s.options.Async == nil {
		return s.Send(ctx, sender, payload)
	}
	return s.options.Async(ctx, func(runCtx context.Context) []Response {
		return s.Send(runCtx, sender, payload)
	})
}

func (s *Signal[T]) disconnect(id uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.receivers {
		if s.receivers[i].id == id && s.receivers[i].active {
			s.receivers[i].active = false
			return true
		}
	}
	return false
}

func (s *Signal[T]) matching(sender any) []receiver[T] {
	s.mu.RLock()
	defer s.mu.RUnlock()
	matched := make([]receiver[T], 0, len(s.receivers))
	for _, receiver := range s.receivers {
		if !receiver.active || receiver.fn == nil {
			continue
		}
		if receiver.sender != nil && !reflect.DeepEqual(receiver.sender, sender) {
			continue
		}
		matched = append(matched, receiver)
	}
	return matched
}

func (s *Signal[T]) call(ctx context.Context, receiver receiver[T], sender any, payload T) (err error) {
	if s.options.RecoverPanics {
		defer func() {
			if recovered := recover(); recovered != nil {
				err = fmt.Errorf("signal receiver panic: %v", recovered)
			}
		}()
	}
	return receiver.fn(ctx, sender, payload)
}
