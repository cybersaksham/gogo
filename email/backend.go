package email

import (
	"context"
)

type Backend interface {
	SendMessages(context.Context, []Message) (int, error)
	Close() error
}

type FailSilently struct {
	FailSilently bool
}

func handleSendError(sent int, err error, failSilently bool) (int, error) {
	if err == nil {
		return sent, nil
	}
	if failSilently {
		return sent, nil
	}
	return sent, err
}
