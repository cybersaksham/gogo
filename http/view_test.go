package http

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestAsHandlerWritesFrameworkViewResponse(t *testing.T) {
	handler := AsHandler(View(func(context.Context, *Request) Response {
		return Text(202, "accepted")
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/", nil))

	if recorder.Code != 202 {
		t.Fatalf("status = %d, want 202", recorder.Code)
	}
	if recorder.Body.String() != "accepted" {
		t.Fatalf("body = %q, want accepted", recorder.Body.String())
	}
}

func TestFromHandlerAdaptsStandardHandler(t *testing.T) {
	view := FromHandler(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.Header().Set("X-Test", "yes")
		w.WriteHeader(203)
		_, _ = w.Write([]byte("standard"))
	}))

	response := view(context.Background(), NewRequest(httptest.NewRequest("GET", "/", nil)))
	recorder := httptest.NewRecorder()
	if err := response.Write(recorder); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if recorder.Code != 203 {
		t.Fatalf("status = %d, want 203", recorder.Code)
	}
	if got := recorder.Header().Get("X-Test"); got != "yes" {
		t.Fatalf("X-Test = %q, want yes", got)
	}
	if recorder.Body.String() != "standard" {
		t.Fatalf("body = %q, want standard", recorder.Body.String())
	}
}

func TestMethodsDispatchesByHTTPMethod(t *testing.T) {
	view := Methods(map[string]View{
		"GET": func(context.Context, *Request) Response {
			return Text(200, "get")
		},
	})

	getResponse := view(context.Background(), NewRequest(httptest.NewRequest("GET", "/", nil)))
	getRecorder := httptest.NewRecorder()
	if err := getResponse.Write(getRecorder); err != nil {
		t.Fatalf("GET Write() error = %v", err)
	}
	if getRecorder.Code != 200 || getRecorder.Body.String() != "get" {
		t.Fatalf("GET response = (%d, %q), want (200, get)", getRecorder.Code, getRecorder.Body.String())
	}

	postResponse := view(context.Background(), NewRequest(httptest.NewRequest("POST", "/", nil)))
	postRecorder := httptest.NewRecorder()
	if err := postResponse.Write(postRecorder); err != nil {
		t.Fatalf("POST Write() error = %v", err)
	}
	if postRecorder.Code != nethttp.StatusMethodNotAllowed {
		t.Fatalf("POST status = %d, want 405", postRecorder.Code)
	}
	if got := postRecorder.Header().Get("Allow"); got != "GET" {
		t.Fatalf("Allow = %q, want GET", got)
	}
}
