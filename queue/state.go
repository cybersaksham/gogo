package queue

type State string

const (
	StatePending  State = "PENDING"
	StateReceived State = "RECEIVED"
	StateStarted  State = "STARTED"
	StateRetry    State = "RETRY"
	StateSuccess  State = "SUCCESS"
	StateFailure  State = "FAILURE"
	StateRevoked  State = "REVOKED"
	StateIgnored  State = "IGNORED"
)

func (s State) String() string {
	return string(s)
}

func (s State) Terminal() bool {
	switch s {
	case StateSuccess, StateFailure, StateRevoked, StateIgnored:
		return true
	default:
		return false
	}
}

func CanTransition(from State, to State) bool {
	if from.Terminal() {
		return false
	}
	switch from {
	case StatePending:
		return to == StateReceived || to == StateStarted || to == StateRetry || to.Terminal()
	case StateReceived:
		return to == StateStarted || to == StateRetry || to.Terminal()
	case StateStarted:
		return to == StateRetry || to.Terminal()
	case StateRetry:
		return to == StateReceived || to == StateStarted || to.Terminal()
	default:
		return false
	}
}
