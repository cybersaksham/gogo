package queue

import "time"

// Envelope is the durable broker message for one task.
type Envelope struct {
	ID            string            `json:"id"`
	RootID        string            `json:"root_id,omitempty"`
	ParentID      string            `json:"parent_id,omitempty"`
	GroupID       string            `json:"group_id,omitempty"`
	ChordID       string            `json:"chord_id,omitempty"`
	Name          string            `json:"name"`
	Args          []any             `json:"args,omitempty"`
	Kwargs        map[string]any    `json:"kwargs,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	Retries       int               `json:"retries"`
	ETA           *time.Time        `json:"eta,omitempty"`
	Expires       *time.Time        `json:"expires,omitempty"`
	Queue         string            `json:"queue,omitempty"`
	Priority      int               `json:"priority,omitempty"`
	ReplyTo       string            `json:"reply_to,omitempty"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
}

type EnvelopeOptions struct {
	ID            string
	RootID        string
	ParentID      string
	GroupID       string
	ChordID       string
	Retries       int
	ReplyTo       string
	CorrelationID string
	CreatedAt     time.Time
}

func NewEnvelope(signature Signature, options EnvelopeOptions) Envelope {
	createdAt := options.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	return Envelope{
		ID:            options.ID,
		RootID:        options.RootID,
		ParentID:      options.ParentID,
		GroupID:       options.GroupID,
		ChordID:       options.ChordID,
		Name:          signature.Name,
		Args:          append([]any(nil), signature.Args...),
		Kwargs:        cloneAnyMap(signature.Kwargs),
		Headers:       cloneStringMap(signature.Headers),
		Retries:       options.Retries,
		ETA:           cloneTime(signature.Options.ETA),
		Expires:       cloneTime(signature.Options.Expires),
		Queue:         signature.Options.Queue,
		Priority:      signature.Options.Priority,
		ReplyTo:       options.ReplyTo,
		CorrelationID: options.CorrelationID,
		CreatedAt:     createdAt,
	}
}

func cloneAnyMap(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
