package queue

import "time"

// Result stores task execution state and payloads.
type Result struct {
	TaskID    string
	State     State
	Result    any
	Error     string
	Traceback string
	Children  []string
	ExpiresAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GroupResult stores a group result handle.
type GroupResult struct {
	ID        string
	Children  []string
	CreatedAt time.Time
}

func (r Result) Clone() Result {
	r.Children = append([]string(nil), r.Children...)
	if r.ExpiresAt != nil {
		value := *r.ExpiresAt
		r.ExpiresAt = &value
	}
	return r
}

func (g GroupResult) Clone() GroupResult {
	g.Children = append([]string(nil), g.Children...)
	return g
}
