package api

import (
	"encoding/json"
	"net/http"
	"os"
)

// APIError is the normalized API error body.
type APIError struct {
	Code      string              `json:"code"`
	Message   string              `json:"message"`
	Fields    map[string][]string `json:"fields,omitempty"`
	RequestID string              `json:"request_id,omitempty"`
}

// Response is an API response writer.
type Response struct {
	status      int
	body        any
	contentType string
	filePath    string
	noBody      bool
}

// JSON creates a JSON response.
func JSON(status int, body any) Response {
	return Response{status: status, body: body, contentType: "application/json"}
}

// Created creates a 201 JSON response.
func Created(body any) Response {
	return JSON(http.StatusCreated, body)
}

// Accepted creates a 202 JSON response.
func Accepted(body any) Response {
	return JSON(http.StatusAccepted, body)
}

// NoContent creates a 204 response.
func NoContent() Response {
	return Response{status: http.StatusNoContent, noBody: true}
}

// Error creates a normalized JSON error response.
func Error(status int, err APIError) Response {
	return JSON(status, map[string]APIError{"error": err})
}

// File creates a file response.
func File(path, contentType string) Response {
	return Response{status: http.StatusOK, filePath: path, contentType: contentType}
}

// Write writes the response to a standard response writer.
func (r Response) Write(w http.ResponseWriter) error {
	if r.contentType != "" {
		w.Header().Set("Content-Type", r.contentType)
	}
	w.WriteHeader(r.status)
	if r.noBody {
		return nil
	}
	if r.filePath != "" {
		body, err := os.ReadFile(r.filePath)
		if err != nil {
			return err
		}
		_, err = w.Write(body)
		return err
	}
	if r.contentType == "application/json" {
		return json.NewEncoder(w).Encode(r.body)
	}
	return nil
}
