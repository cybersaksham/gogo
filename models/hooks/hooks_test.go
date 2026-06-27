package hooks

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestDispatchRunsEveryLifecycleEvent(t *testing.T) {
	registry := NewRegistry()
	events := []Event{
		BeforeValidate,
		AfterValidate,
		BeforeSave,
		AfterSave,
		BeforeDelete,
		AfterDelete,
		ManyToManyChanged,
	}
	seen := make([]Event, 0, len(events))
	for _, event := range events {
		err := registry.Register(Hook{
			Event: event,
			Name:  string(event),
			Func: func(_ context.Context, payload Payload) error {
				seen = append(seen, payload.Event)
				return nil
			},
		})
		if err != nil {
			t.Fatalf("Register(%s) error = %v", event, err)
		}
	}

	for _, event := range events {
		if err := registry.Dispatch(context.Background(), Payload{Event: event}); err != nil {
			t.Fatalf("Dispatch(%s) error = %v", event, err)
		}
	}
	if !reflect.DeepEqual(seen, events) {
		t.Fatalf("seen events = %#v, want %#v", seen, events)
	}
}

func TestDispatchUsesDeterministicHookOrder(t *testing.T) {
	registry := NewRegistry()
	var order []string
	for _, hook := range []Hook{
		namedHook(BeforeSave, "z", 10, &order),
		namedHook(BeforeSave, "b", 0, &order),
		namedHook(BeforeSave, "a", 0, &order),
		namedHook(BeforeSave, "first", -10, &order),
	} {
		if err := registry.Register(hook); err != nil {
			t.Fatalf("Register(%s) error = %v", hook.Name, err)
		}
	}

	if err := registry.Dispatch(context.Background(), Payload{Event: BeforeSave}); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	want := []string{"first", "a", "b", "z"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("order = %#v, want %#v", order, want)
	}
}

func TestDispatchStopsOnFailingHook(t *testing.T) {
	registry := NewRegistry()
	failure := errors.New("no save")
	if err := registry.Register(Hook{
		Event: BeforeSave,
		Name:  "fail",
		Func: func(context.Context, Payload) error {
			return failure
		},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if err := registry.Register(namedHook(BeforeSave, "after", 1, nil)); err != nil {
		t.Fatalf("Register(after) error = %v", err)
	}

	err := registry.Dispatch(context.Background(), Payload{Event: BeforeSave})
	if !errors.Is(err, ErrHookFailed) || !errors.Is(err, failure) {
		t.Fatalf("Dispatch() error = %v, want ErrHookFailed wrapping failure", err)
	}
}

func TestDispatchHonorsContextCancellation(t *testing.T) {
	registry := NewRegistry()
	called := false
	if err := registry.Register(Hook{
		Event: AfterDelete,
		Name:  "cancelled",
		Func: func(context.Context, Payload) error {
			called = true
			return nil
		},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := registry.Dispatch(ctx, Payload{Event: AfterDelete}); !errors.Is(err, context.Canceled) {
		t.Fatalf("Dispatch() error = %v, want context.Canceled", err)
	}
	if called {
		t.Fatalf("hook was called after context cancellation")
	}
}

func TestManyToManyPayloadMetadata(t *testing.T) {
	registry := NewRegistry()
	var got Payload
	if err := registry.Register(Hook{
		Event: ManyToManyChanged,
		Name:  "capture",
		Func: func(_ context.Context, payload Payload) error {
			got = payload
			return nil
		},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	payload := Payload{
		Event:      ManyToManyChanged,
		Target:     "article",
		Relation:   "tags",
		Action:     M2MPreAdd,
		Reverse:    true,
		PrimarySet: []any{int64(1), int64(2)},
		Using:      "replica",
	}
	if err := registry.Dispatch(context.Background(), payload); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	if got.Relation != "tags" || got.Action != M2MPreAdd || !got.Reverse || got.Using != "replica" {
		t.Fatalf("payload metadata = %#v", got)
	}
	if !reflect.DeepEqual(got.PrimarySet, []any{int64(1), int64(2)}) {
		t.Fatalf("PrimarySet = %#v", got.PrimarySet)
	}
}

func TestRegisterRejectsInvalidHooks(t *testing.T) {
	registry := NewRegistry()
	cases := []Hook{
		{},
		{Event: BeforeSave, Name: "missing-func"},
		{Event: Event("bad"), Name: "bad", Func: func(context.Context, Payload) error { return nil }},
	}
	for _, hook := range cases {
		if err := registry.Register(hook); !errors.Is(err, ErrInvalidHook) {
			t.Fatalf("Register(%#v) error = %v, want ErrInvalidHook", hook, err)
		}
	}
}

func namedHook(event Event, name string, order int, target *[]string) Hook {
	return Hook{
		Event: event,
		Name:  name,
		Order: order,
		Func: func(context.Context, Payload) error {
			if target != nil {
				*target = append(*target, name)
			}
			return nil
		},
	}
}
