package http

import (
	"bytes"
	"context"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"sort"
	"strings"
)

// View handles a framework request and returns a framework response.
type View func(context.Context, *Request) Response

// AsHandler adapts a framework view to a standard HTTP handler.
func AsHandler(view View) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		response := view(r.Context(), NewRequest(r))
		if err := response.Write(w); err != nil {
			nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
		}
	})
}

// FromHandler adapts a standard HTTP handler to a framework view.
func FromHandler(handler nethttp.Handler) View {
	return func(_ context.Context, request *Request) Response {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request.Raw())

		result := recorder.Result()
		defer result.Body.Close()

		body, _ := io.ReadAll(result.Body)
		response := bodyResponse(result.StatusCode, result.Header.Get("Content-Type"), body)
		for key, values := range result.Header {
			for _, value := range values {
				response.Header().Add(key, value)
			}
		}
		return response
	}
}

// Methods dispatches to method-specific views and returns 405 for unsupported methods.
func Methods(views map[string]View) View {
	return func(ctx context.Context, request *Request) Response {
		if view, ok := views[request.Method()]; ok {
			return view(ctx, request)
		}

		allowed := make([]string, 0, len(views))
		for method := range views {
			allowed = append(allowed, method)
		}
		sort.Strings(allowed)

		response := Text(nethttp.StatusMethodNotAllowed, "Method Not Allowed")
		response.Header().Set("Allow", strings.Join(allowed, ", "))
		return response
	}
}

type responseCapture struct {
	status int
	header nethttp.Header
	body   bytes.Buffer
}
