package http

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTemporaryAndPermanentRedirects(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		status   int
	}{
		{name: "temporary", response: TemporaryRedirect("/next/"), status: 302},
		{name: "permanent", response: PermanentRedirect("/next/"), status: 301},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			if err := test.response.Write(recorder); err != nil {
				t.Fatalf("Write() error = %v", err)
			}
			if recorder.Code != test.status {
				t.Fatalf("status = %d, want %d", recorder.Code, test.status)
			}
			if got := recorder.Header().Get("Location"); got != "/next/" {
				t.Fatalf("Location = %q, want /next/", got)
			}
		})
	}
}

func TestRedirectToRouteName(t *testing.T) {
	router := NewRouter()
	if err := router.Handle("article-detail", "/articles/<int:id>/", func(context.Context, *Request) Response {
		return NoContent()
	}); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	response, err := RedirectToRoute(router, "article-detail", map[string]any{"id": 42})
	if err != nil {
		t.Fatalf("RedirectToRoute() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	if err := response.Write(recorder); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if got := recorder.Header().Get("Location"); got != "/articles/42/" {
		t.Fatalf("Location = %q, want /articles/42/", got)
	}
}

func TestErrorResponsesUseSafePublicMessages(t *testing.T) {
	private := errors.New("database password leaked")
	response := InternalServerError(private)

	recorder := httptest.NewRecorder()
	if err := response.Write(recorder); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if recorder.Code != 500 {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "database password") {
		t.Fatalf("private detail leaked in body %q", recorder.Body.String())
	}
	if !errors.Is(response.PrivateError(), private) {
		t.Fatalf("PrivateError() = %v, want private error", response.PrivateError())
	}
}

func TestHTTPErrorResponseStatuses(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		status   int
		body     string
		allow    string
	}{
		{name: "bad request", response: BadRequest("Invalid request", nil), status: 400, body: "Invalid request"},
		{name: "forbidden", response: Forbidden("Forbidden", nil), status: 403, body: "Forbidden"},
		{name: "not found", response: NotFound("Missing", nil), status: 404, body: "Missing"},
		{name: "method not allowed", response: MethodNotAllowed([]string{"GET", "POST"}, nil), status: 405, body: "Method Not Allowed", allow: "GET, POST"},
		{name: "conflict", response: Conflict("Conflict", nil), status: 409, body: "Conflict"},
		{name: "internal server error", response: InternalServerError(nil), status: 500, body: "Internal Server Error"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			if err := test.response.Write(recorder); err != nil {
				t.Fatalf("Write() error = %v", err)
			}
			if recorder.Code != test.status {
				t.Fatalf("status = %d, want %d", recorder.Code, test.status)
			}
			if strings.TrimSpace(recorder.Body.String()) != test.body {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), test.body)
			}
			if test.allow != "" && recorder.Header().Get("Allow") != test.allow {
				t.Fatalf("Allow = %q, want %q", recorder.Header().Get("Allow"), test.allow)
			}
		})
	}
}
