package messages

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestMemoryStorageAddIterateConsumeAndLevelFiltering(t *testing.T) {
	storage := NewMemoryStorage()
	storage.SetMinimumLevel(LevelInfo)
	storage.Add(LevelDebug, "debug")
	storage.Add(LevelInfo, "info")
	storage.Add(LevelSuccess, "success", "saved")
	if got := messageTexts(storage.Messages()); !reflect.DeepEqual(got, []string{"info", "success"}) {
		t.Fatalf("Messages() = %#v", got)
	}
	if got := storage.Consume(); !reflect.DeepEqual(messageTexts(got), []string{"info", "success"}) {
		t.Fatalf("Consume() = %#v", got)
	}
	if len(storage.Messages()) != 0 {
		t.Fatal("messages should be consumed")
	}
}

func TestSessionStorageAndCookieStorage(t *testing.T) {
	session := map[string]any{}
	sessionStorage := NewSessionStorage(session)
	sessionStorage.Add(LevelWarning, "warn")
	loaded := NewSessionStorage(session)
	if got := messageTexts(loaded.Messages()); !reflect.DeepEqual(got, []string{"warn"}) {
		t.Fatalf("session messages = %#v", got)
	}
	cookie := NewCookieStorage("gogo_messages")
	encoded, err := cookie.Encode([]Message{{Level: LevelError, Text: "bad"}})
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	decoded, err := cookie.Decode(encoded)
	if err != nil || len(decoded) != 1 || decoded[0].Text != "bad" {
		t.Fatalf("Decode() = %#v, %v", decoded, err)
	}
}

func TestMiddlewareAttachesMessagesToContextAndTemplateContext(t *testing.T) {
	storage := NewMemoryStorage()
	middleware := Middleware(func(*http.Request) Storage { return storage })
	var got []Message
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Add(r.Context(), LevelInfo, "hello")
		got = TemplateContext(r.Context())["messages"].([]Message)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if len(got) != 1 || got[0].Text != "hello" {
		t.Fatalf("context messages = %#v", got)
	}
}

func messageTexts(messages []Message) []string {
	texts := make([]string, len(messages))
	for i, message := range messages {
		texts[i] = message.Text
	}
	return texts
}
