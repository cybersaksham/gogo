package http

import (
	"encoding/json"
	"io"
	nethttp "net/http"
	"os"
)

// Response is a framework HTTP response.
type Response struct {
	status      int
	headers     nethttp.Header
	body        []byte
	jsonValue   any
	stream      func(io.Writer) error
	filePath    string
	noBody      bool
	contentType string
}

// Text creates a plain text response.
func Text(status int, body string) Response {
	return bodyResponse(status, "text/plain; charset=utf-8", []byte(body))
}

// HTML creates an HTML response.
func HTML(status int, body string) Response {
	return bodyResponse(status, "text/html; charset=utf-8", []byte(body))
}

// JSON creates a JSON response.
func JSON(status int, value any) Response {
	return Response{
		status:      status,
		headers:     make(nethttp.Header),
		jsonValue:   value,
		contentType: "application/json",
	}
}

// NoContent creates a 204 response.
func NoContent() Response {
	return Response{
		status:  nethttp.StatusNoContent,
		headers: make(nethttp.Header),
		noBody:  true,
	}
}

// File creates a file response.
func File(path string) Response {
	return Response{
		status:   nethttp.StatusOK,
		headers:  make(nethttp.Header),
		filePath: path,
	}
}

// Stream creates a streaming response.
func Stream(contentType string, fn func(io.Writer) error) Response {
	return Response{
		status:      nethttp.StatusOK,
		headers:     make(nethttp.Header),
		stream:      fn,
		contentType: contentType,
	}
}

// Header returns mutable response headers.
func (r Response) Header() nethttp.Header {
	return r.headers
}

// Status returns the response status code.
func (r Response) Status() int {
	return r.status
}

// Write writes the response to a standard HTTP response writer.
func (r Response) Write(w nethttp.ResponseWriter) error {
	for key, values := range r.headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	if r.contentType != "" {
		w.Header().Set("Content-Type", r.contentType)
	}

	w.WriteHeader(r.status)
	if r.noBody {
		return nil
	}
	if r.jsonValue != nil {
		return json.NewEncoder(w).Encode(r.jsonValue)
	}
	if r.stream != nil {
		return r.stream(w)
	}
	if r.filePath != "" {
		data, err := os.ReadFile(r.filePath)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}
	if len(r.body) > 0 {
		_, err := w.Write(r.body)
		return err
	}
	return nil
}

func bodyResponse(status int, contentType string, body []byte) Response {
	return Response{
		status:      status,
		headers:     make(nethttp.Header),
		body:        body,
		contentType: contentType,
	}
}
