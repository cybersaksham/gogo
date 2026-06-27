package hooks

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

var (
	ErrInvalidHook = errors.New("invalid model hook")
	ErrHookFailed  = errors.New("model hook failed")
)

// Event identifies a model lifecycle hook point.
type Event string

const (
	BeforeValidate    Event = "before_validate"
	AfterValidate     Event = "after_validate"
	BeforeSave        Event = "before_save"
	AfterSave         Event = "after_save"
	BeforeDelete      Event = "before_delete"
	AfterDelete       Event = "after_delete"
	ManyToManyChanged Event = "many_to_many_changed"
)

// M2MAction identifies the many-to-many transition.
type M2MAction string

const (
	M2MPreAdd     M2MAction = "pre_add"
	M2MPostAdd    M2MAction = "post_add"
	M2MPreRemove  M2MAction = "pre_remove"
	M2MPostRemove M2MAction = "post_remove"
	M2MPreClear   M2MAction = "pre_clear"
	M2MPostClear  M2MAction = "post_clear"
)

// Payload is the context passed to each hook.
type Payload struct {
	Event      Event
	Target     any
	Relation   string
	Action     M2MAction
	Reverse    bool
	PrimarySet []any
	Using      string
}

// Clone returns a copy safe to pass between hooks.
func (p Payload) Clone() Payload {
	copied := p
	copied.PrimarySet = append([]any(nil), p.PrimarySet...)
	return copied
}

// Func is a context-aware hook function.
type Func func(context.Context, Payload) error

// Hook describes one registered lifecycle hook.
type Hook struct {
	Event Event
	Name  string
	Order int
	Func  Func
}

// Registry stores lifecycle hooks by event.
type Registry struct {
	mu       sync.RWMutex
	handlers map[Event][]Hook
}

// NewRegistry creates an empty hook registry.
func NewRegistry() *Registry {
	return &Registry{handlers: make(map[Event][]Hook)}
}

// Register adds one hook to the registry.
func (r *Registry) Register(hook Hook) error {
	if err := validateHook(hook); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.handlers[hook.Event] {
		if existing.Name == hook.Name {
			return fmt.Errorf("%w: duplicate hook %s for %s", ErrInvalidHook, hook.Name, hook.Event)
		}
	}
	r.handlers[hook.Event] = append(r.handlers[hook.Event], hook)
	return nil
}

// Hooks returns registered hooks for an event in dispatch order.
func (r *Registry) Hooks(event Event) []Hook {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return orderedHooks(r.handlers[event])
}

// Dispatch runs all hooks for a payload event in deterministic order.
func (r *Registry) Dispatch(ctx context.Context, payload Payload) error {
	if !payload.Event.Valid() {
		return fmt.Errorf("%w: unsupported event %q", ErrInvalidHook, payload.Event)
	}

	hooks := r.Hooks(payload.Event)
	for _, hook := range hooks {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := hook.Func(ctx, payload.Clone()); err != nil {
			return fmt.Errorf("%w: %s: %w", ErrHookFailed, hook.Name, err)
		}
	}
	return ctx.Err()
}

// Valid reports whether an event is supported.
func (e Event) Valid() bool {
	switch e {
	case BeforeValidate, AfterValidate, BeforeSave, AfterSave, BeforeDelete, AfterDelete, ManyToManyChanged:
		return true
	default:
		return false
	}
}

// Valid reports whether a many-to-many action is supported.
func (a M2MAction) Valid() bool {
	switch a {
	case "", M2MPreAdd, M2MPostAdd, M2MPreRemove, M2MPostRemove, M2MPreClear, M2MPostClear:
		return true
	default:
		return false
	}
}

func validateHook(hook Hook) error {
	if !hook.Event.Valid() {
		return fmt.Errorf("%w: unsupported event %q", ErrInvalidHook, hook.Event)
	}
	if strings.TrimSpace(hook.Name) == "" {
		return fmt.Errorf("%w: hook name is required", ErrInvalidHook)
	}
	if hook.Func == nil {
		return fmt.Errorf("%w: hook function is required", ErrInvalidHook)
	}
	return nil
}

func orderedHooks(hooks []Hook) []Hook {
	copied := append([]Hook(nil), hooks...)
	sort.SliceStable(copied, func(i, j int) bool {
		if copied[i].Order == copied[j].Order {
			return copied[i].Name < copied[j].Name
		}
		return copied[i].Order < copied[j].Order
	})
	return copied
}
