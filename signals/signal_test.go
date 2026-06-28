package signals

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestSignalConnectDisconnectSenderFilteringOrderingAndPanicRecovery(t *testing.T) {
	signal := New[string]("test", Options{RecoverPanics: true})
	var calls []string
	signal.Connect("sender-a", func(context.Context, any, string) error {
		calls = append(calls, "first")
		return nil
	})
	handle := signal.Connect(nil, func(context.Context, any, string) error {
		calls = append(calls, "second")
		panic("boom")
	})
	responses := signal.Send(context.Background(), "sender-a", "payload")
	if !reflect.DeepEqual(calls, []string{"first", "second"}) || len(responses) != 2 || responses[1].Error == nil {
		t.Fatalf("calls=%#v responses=%#v", calls, responses)
	}
	if !handle.Disconnect() || handle.Disconnect() {
		t.Fatalf("disconnect handle did not report correctly")
	}
	calls = nil
	signal.Send(context.Background(), "sender-b", "payload")
	if len(calls) != 0 {
		t.Fatalf("sender filtering failed: %#v", calls)
	}
}

func TestSignalAsyncHookAndBuiltins(t *testing.T) {
	var asyncCalled bool
	signal := New[int]("async", Options{Async: func(ctx context.Context, run func(context.Context) []Response) []Response {
		asyncCalled = true
		return run(ctx)
	}})
	wantErr := errors.New("fail")
	signal.Connect(nil, func(context.Context, any, int) error { return wantErr })
	responses := signal.SendAsync(context.Background(), nil, 1)
	if !asyncCalled || len(responses) != 1 || !errors.Is(responses[0].Error, wantErr) {
		t.Fatalf("async responses=%#v called=%v", responses, asyncCalled)
	}
	if AppReady.Name() != "app_ready" || PostMigrate.Name() != "post_migrate" || CheckRegistered.Name() != "check_registered" {
		t.Fatalf("built-in signal names missing")
	}
}
