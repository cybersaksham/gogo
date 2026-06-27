package http

import (
	"errors"
	"io"
	"net/http/httptest"
	"testing"
)

func TestRequestHelpersExposeRequestData(t *testing.T) {
	raw := httptest.NewRequest("GET", "https://example.com/posts/?page=2", nil)
	raw.RemoteAddr = "192.0.2.10:12345"

	request := NewRequest(raw).
		WithPathParam("id", "42").
		WithUser("user").
		WithSession("session")

	if got := request.Method(); got != "GET" {
		t.Fatalf("Method() = %q, want GET", got)
	}
	if got := request.Host(); got != "example.com" {
		t.Fatalf("Host() = %q, want example.com", got)
	}
	if got := request.Scheme(); got != "https" {
		t.Fatalf("Scheme() = %q, want https", got)
	}
	if got := request.RemoteIP(); got != "192.0.2.10" {
		t.Fatalf("RemoteIP() = %q, want 192.0.2.10", got)
	}
	if got := request.PathParam("id"); got != "42" {
		t.Fatalf("PathParam(id) = %q, want 42", got)
	}
	if got := request.QueryParam("page"); got != "2" {
		t.Fatalf("QueryParam(page) = %q, want 2", got)
	}
	if got := request.User(); got != "user" {
		t.Fatalf("User() = %#v, want user", got)
	}
	if got := request.Session(); got != "session" {
		t.Fatalf("Session() = %#v, want session", got)
	}
	if request.Context() == nil {
		t.Fatalf("Context() = nil")
	}
}

func TestResponseHelpersWriteHeadersStatusAndBody(t *testing.T) {
	cases := []struct {
		name       string
		response   Response
		wantStatus int
		wantType   string
		wantBody   string
	}{
		{name: "text", response: Text(201, "created"), wantStatus: 201, wantType: "text/plain; charset=utf-8", wantBody: "created"},
		{name: "html", response: HTML(200, "<strong>ok</strong>"), wantStatus: 200, wantType: "text/html; charset=utf-8", wantBody: "<strong>ok</strong>"},
		{name: "json", response: JSON(200, map[string]string{"ok": "yes"}), wantStatus: 200, wantType: "application/json", wantBody: "{\"ok\":\"yes\"}\n"},
		{name: "no content", response: NoContent(), wantStatus: 204, wantType: "", wantBody: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			if err := tc.response.Write(recorder); err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			if recorder.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tc.wantStatus)
			}
			if got := recorder.Header().Get("Content-Type"); got != tc.wantType {
				t.Fatalf("Content-Type = %q, want %q", got, tc.wantType)
			}
			if recorder.Body.String() != tc.wantBody {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), tc.wantBody)
			}
		})
	}
}

func TestJSONResponseReturnsEncodingError(t *testing.T) {
	response := JSON(200, func() {})

	err := response.Write(httptest.NewRecorder())
	if err == nil {
		t.Fatalf("Write() error = nil, want JSON encoding error")
	}
}

func TestStreamResponseReturnsStreamError(t *testing.T) {
	wantErr := errors.New("stream failed")
	response := Stream("text/plain", func(w io.Writer) error {
		_, _ = w.Write([]byte("partial"))
		return wantErr
	})

	recorder := httptest.NewRecorder()
	err := response.Write(recorder)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Write() error = %v, want stream error", err)
	}
	if recorder.Body.String() != "partial" {
		t.Fatalf("body = %q, want partial", recorder.Body.String())
	}
}
