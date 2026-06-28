package messages

import (
	"encoding/base64"
	"encoding/json"
	"sync"
)

const sessionKey = "_gogo_messages"

type Storage interface {
	Add(Level, string, ...string)
	Messages() []Message
	Consume() []Message
	SetMinimumLevel(Level)
}

type MemoryStorage struct {
	mu      sync.Mutex
	minimum Level
	items   []Message
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{minimum: LevelDebug}
}

func (s *MemoryStorage) Add(level Level, text string, extraTags ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if level < s.minimum {
		return
	}
	s.items = append(s.items, Message{Level: level, Text: text, Tags: TagsForLevel(level), ExtraTags: append([]string(nil), extraTags...)})
}

func (s *MemoryStorage) Messages() []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneMessages(s.items)
}

func (s *MemoryStorage) Consume() []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := cloneMessages(s.items)
	s.items = nil
	return items
}

func (s *MemoryStorage) SetMinimumLevel(level Level) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.minimum = level
}

type SessionStorage struct {
	session map[string]any
	minimum Level
}

func NewSessionStorage(session map[string]any) *SessionStorage {
	if session == nil {
		session = map[string]any{}
	}
	return &SessionStorage{session: session, minimum: LevelDebug}
}

func (s *SessionStorage) Add(level Level, text string, extraTags ...string) {
	if level < s.minimum {
		return
	}
	items := s.Messages()
	items = append(items, Message{Level: level, Text: text, Tags: TagsForLevel(level), ExtraTags: append([]string(nil), extraTags...)})
	s.session[sessionKey] = cloneMessages(items)
}

func (s *SessionStorage) Messages() []Message {
	value := s.session[sessionKey]
	switch typed := value.(type) {
	case []Message:
		return cloneMessages(typed)
	case []any:
		messages := make([]Message, 0, len(typed))
		for _, item := range typed {
			if message, ok := item.(Message); ok {
				messages = append(messages, message)
			}
		}
		return messages
	default:
		return nil
	}
}

func (s *SessionStorage) Consume() []Message {
	items := s.Messages()
	delete(s.session, sessionKey)
	return items
}

func (s *SessionStorage) SetMinimumLevel(level Level) {
	s.minimum = level
}

type CookieStorage struct {
	Name string
}

func NewCookieStorage(name string) CookieStorage {
	if name == "" {
		name = "gogo_messages"
	}
	return CookieStorage{Name: name}
}

func (s CookieStorage) Encode(messages []Message) (string, error) {
	body, err := json.Marshal(messages)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(body), nil
}

func (s CookieStorage) Decode(value string) ([]Message, error) {
	body, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}
	var messages []Message
	if err := json.Unmarshal(body, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

func cloneMessages(messages []Message) []Message {
	cloned := make([]Message, len(messages))
	for i, message := range messages {
		cloned[i] = Message{
			Level:     message.Level,
			Text:      message.Text,
			Tags:      append([]string(nil), message.Tags...),
			ExtraTags: append([]string(nil), message.ExtraTags...),
		}
	}
	return cloned
}
