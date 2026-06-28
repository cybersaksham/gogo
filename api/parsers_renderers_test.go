package api

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIParsersParseJSONFormMultipartTextAndRaw(t *testing.T) {
	jsonReq := httptest.NewRequest("POST", "/api/", strings.NewReader(`{"title":"Gogo"}`))
	jsonReq.Header.Set("Content-Type", "application/json")
	parsed, err := DefaultParserRegistry().Parse(jsonReq, 1024)
	if err != nil {
		t.Fatalf("Parse(json) error = %v", err)
	}
	if parsed.(map[string]any)["title"] != "Gogo" {
		t.Fatalf("json parsed = %#v", parsed)
	}

	formReq := httptest.NewRequest("POST", "/api/", strings.NewReader("title=Gogo"))
	formReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	parsed, err = DefaultParserRegistry().Parse(formReq, 1024)
	if err != nil || parsed.(map[string][]string)["title"][0] != "Gogo" {
		t.Fatalf("Parse(form) = %#v, %v", parsed, err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("title", "Gogo")
	part, _ := writer.CreateFormFile("file", "post.txt")
	_, _ = part.Write([]byte("content"))
	_ = writer.Close()
	multipartReq := httptest.NewRequest("POST", "/api/", &body)
	multipartReq.Header.Set("Content-Type", writer.FormDataContentType())
	parsed, err = DefaultParserRegistry().Parse(multipartReq, 1024)
	if err != nil {
		t.Fatalf("Parse(multipart) error = %v", err)
	}
	multipartBody := parsed.(MultipartBody)
	if multipartBody.Values["title"][0] != "Gogo" || multipartBody.Files["file"][0].Filename != "post.txt" {
		t.Fatalf("multipart parsed = %#v", multipartBody)
	}

	textReq := httptest.NewRequest("POST", "/api/", strings.NewReader("hello"))
	textReq.Header.Set("Content-Type", "text/plain")
	parsed, err = DefaultParserRegistry().Parse(textReq, 1024)
	if err != nil || parsed.(string) != "hello" {
		t.Fatalf("Parse(text) = %#v, %v", parsed, err)
	}

	rawReq := httptest.NewRequest("POST", "/api/", strings.NewReader("abc"))
	rawReq.Header.Set("Content-Type", "application/octet-stream")
	parsed, err = DefaultParserRegistry().Parse(rawReq, 1024)
	if err != nil || string(parsed.([]byte)) != "abc" {
		t.Fatalf("Parse(raw) = %#v, %v", parsed, err)
	}
}

func TestAPIParsersRejectUnsupportedInvalidAndOversizedBodies(t *testing.T) {
	unsupported := httptest.NewRequest("POST", "/api/", strings.NewReader("x"))
	unsupported.Header.Set("Content-Type", "application/xml")
	if _, err := DefaultParserRegistry().Parse(unsupported, 1024); !errors.Is(err, ErrUnsupportedMediaType) {
		t.Fatalf("unsupported error = %v, want ErrUnsupportedMediaType", err)
	}

	invalid := httptest.NewRequest("POST", "/api/", strings.NewReader("{bad"))
	invalid.Header.Set("Content-Type", "application/json")
	if _, err := DefaultParserRegistry().Parse(invalid, 1024); !errors.Is(err, ErrParse) {
		t.Fatalf("invalid json error = %v, want ErrParse", err)
	}

	oversized := httptest.NewRequest("POST", "/api/", strings.NewReader("too-large"))
	oversized.Header.Set("Content-Type", "text/plain")
	if _, err := DefaultParserRegistry().Parse(oversized, 3); !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("oversized error = %v, want ErrBodyTooLarge", err)
	}
}

func TestAPIRenderersNegotiateAndRender(t *testing.T) {
	renderer, err := NegotiateRenderer("application/json", DefaultRenderers(true))
	if err != nil || renderer.MediaType() != "application/json" {
		t.Fatalf("NegotiateRenderer(json) = %#v, %v", renderer, err)
	}
	body, contentType, err := renderer.Render(map[string]string{"ok": "yes"})
	if err != nil || string(body) != "{\"ok\":\"yes\"}\n" || contentType != "application/json" {
		t.Fatalf("Render(json) = %q, %q, %v", body, contentType, err)
	}

	browsable, err := NegotiateRenderer("text/html", DefaultRenderers(true))
	if err != nil || browsable.MediaType() != "text/html" {
		t.Fatalf("NegotiateRenderer(html) = %#v, %v", browsable, err)
	}

	plain, err := NegotiateRenderer("text/plain", DefaultRenderers(false))
	if err != nil || plain.MediaType() != "text/plain" {
		t.Fatalf("NegotiateRenderer(text) = %#v, %v", plain, err)
	}
	errorBody, _, err := plain.Render(APIError{Code: "invalid", Message: "Invalid input"})
	if err != nil || string(errorBody) != "invalid: Invalid input" {
		t.Fatalf("Render(text error) = %q, %v", errorBody, err)
	}

	if _, err := NegotiateRenderer("application/xml", DefaultRenderers(false)); !errors.Is(err, ErrNotAcceptable) {
		t.Fatalf("NegotiateRenderer(xml) error = %v, want ErrNotAcceptable", err)
	}
}

func TestParserRegistryIgnoresContentTypeParameters(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/", strings.NewReader(`{"ok":true}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if _, err := DefaultParserRegistry().Parse(req, 1024); err != nil {
		t.Fatalf("Parse(parameterized json) error = %v", err)
	}
}
