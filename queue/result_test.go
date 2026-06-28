package queue

import "testing"

func TestQueueStateTransitionsAndTerminalStates(t *testing.T) {
	valid := []State{StatePending, StateReceived, StateStarted, StateRetry, StateSuccess, StateFailure, StateRevoked, StateIgnored}
	for _, state := range valid {
		if state.String() == "" {
			t.Fatalf("state %q has empty String()", state)
		}
	}
	if !CanTransition(StatePending, StateReceived) || !CanTransition(StateStarted, StateSuccess) || !CanTransition(StateRetry, StateStarted) {
		t.Fatalf("expected transitions rejected")
	}
	if CanTransition(StateSuccess, StateStarted) || CanTransition(StateFailure, StateRetry) {
		t.Fatalf("terminal transitions allowed")
	}
	if !StateSuccess.Terminal() || !StateFailure.Terminal() || !StateRevoked.Terminal() || !StateIgnored.Terminal() {
		t.Fatalf("terminal state check failed")
	}
	if StateStarted.Terminal() {
		t.Fatalf("started should not be terminal")
	}
}
