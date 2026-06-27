package sessions

import "time"

// Session stores request-scoped key/value data and expiry metadata.
type Session struct {
	Key        string
	Data       map[string]string
	ExpireDate time.Time
	Modified   bool
	Accessed   bool
	maxAge     time.Duration
}

// NewSession creates a session with an expiry age.
func NewSession(maxAge time.Duration) *Session {
	session := &Session{
		Data:   make(map[string]string),
		maxAge: maxAge,
	}
	if maxAge > 0 {
		session.ExpireDate = time.Now().Add(maxAge)
	}
	return session
}

// Get returns one session value and marks the session accessed.
func (s *Session) Get(key string) (string, bool) {
	s.Accessed = true
	if s.Data == nil {
		return "", false
	}
	value, ok := s.Data[key]
	return value, ok
}

// GetString returns one value or the empty string.
func (s *Session) GetString(key string) string {
	value, _ := s.Get(key)
	return value
}

// Set stores a value and marks the session modified.
func (s *Session) Set(key, value string) {
	s.ensureData()
	s.Data[key] = value
	s.Modified = true
}

// Delete removes a value and marks the session modified.
func (s *Session) Delete(key string) {
	if s.Data != nil {
		delete(s.Data, key)
	}
	s.Modified = true
}

// ExpiryAge returns the remaining lifetime at a point in time.
func (s *Session) ExpiryAge(now time.Time) time.Duration {
	if s.ExpireDate.IsZero() {
		return 0
	}
	return s.ExpireDate.Sub(now)
}

// ExpiryDateValue returns the absolute expiry time.
func (s *Session) ExpiryDateValue() time.Time {
	return s.ExpireDate
}

// IsExpired reports whether the session expiry has passed.
func (s *Session) IsExpired(now time.Time) bool {
	return !s.ExpireDate.IsZero() && !s.ExpireDate.After(now)
}

func (s *Session) ensureData() {
	if s.Data == nil {
		s.Data = make(map[string]string)
	}
}

func (s *Session) clone() *Session {
	if s == nil {
		return nil
	}
	copied := *s
	copied.Data = cloneData(s.Data)
	return &copied
}

func cloneData(data map[string]string) map[string]string {
	if data == nil {
		return make(map[string]string)
	}
	copied := make(map[string]string, len(data))
	for key, value := range data {
		copied[key] = value
	}
	return copied
}
