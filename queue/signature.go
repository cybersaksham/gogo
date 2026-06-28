package queue

import "time"

// SignatureOptions stores immutable task dispatch options.
type SignatureOptions struct {
	Queue    string
	Priority int
	ETA      *time.Time
	Expires  *time.Time
}

// Signature describes a task call before it becomes an envelope.
type Signature struct {
	Name    string
	Args    []any
	Kwargs  map[string]any
	Headers map[string]string
	Options SignatureOptions
}

func NewSignature(name string, args ...any) Signature {
	return Signature{Name: name, Args: append([]any(nil), args...), Kwargs: map[string]any{}, Headers: map[string]string{}}
}

func (s Signature) Clone() Signature {
	return Signature{
		Name:    s.Name,
		Args:    append([]any(nil), s.Args...),
		Kwargs:  cloneAnyMap(s.Kwargs),
		Headers: cloneStringMap(s.Headers),
		Options: SignatureOptions{
			Queue:    s.Options.Queue,
			Priority: s.Options.Priority,
			ETA:      cloneTime(s.Options.ETA),
			Expires:  cloneTime(s.Options.Expires),
		},
	}
}

func (s Signature) WithKwarg(name string, value any) Signature {
	clone := s.Clone()
	clone.Kwargs[name] = value
	return clone
}

func (s Signature) WithHeader(name string, value string) Signature {
	clone := s.Clone()
	clone.Headers[name] = value
	return clone
}

func (s Signature) WithQueue(queue string) Signature {
	clone := s.Clone()
	clone.Options.Queue = queue
	return clone
}

func (s Signature) WithPriority(priority int) Signature {
	clone := s.Clone()
	clone.Options.Priority = priority
	return clone
}

func (s Signature) WithCountdown(countdown time.Duration, now ...time.Time) Signature {
	base := time.Now().UTC()
	if len(now) > 0 {
		base = now[0]
	}
	return s.WithETA(base.Add(countdown))
}

func (s Signature) WithETA(eta time.Time) Signature {
	clone := s.Clone()
	clone.Options.ETA = &eta
	return clone
}

func (s Signature) WithExpires(expires time.Time) Signature {
	clone := s.Clone()
	clone.Options.Expires = &expires
	return clone
}
