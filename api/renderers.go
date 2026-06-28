package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Renderer serializes an API response body.
type Renderer interface {
	MediaType() string
	Render(any) ([]byte, string, error)
}

// DefaultRenderers returns built-in renderers.
func DefaultRenderers(enableBrowsable bool) []Renderer {
	renderers := []Renderer{JSONRenderer{}}
	if enableBrowsable {
		renderers = append(renderers, BrowsableAPIRenderer{})
	}
	renderers = append(renderers, PlainTextRenderer{})
	return renderers
}

// NegotiateRenderer selects a renderer from an Accept header.
func NegotiateRenderer(accept string, renderers []Renderer) (Renderer, error) {
	if accept == "" || accept == "*/*" {
		return renderers[0], nil
	}
	parts := strings.Split(accept, ",")
	for _, part := range parts {
		media := strings.TrimSpace(strings.Split(part, ";")[0])
		for _, renderer := range renderers {
			if renderer.MediaType() == media {
				return renderer, nil
			}
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrNotAcceptable, accept)
}

// JSONRenderer renders JSON.
type JSONRenderer struct{}

func (JSONRenderer) MediaType() string { return "application/json" }

func (JSONRenderer) Render(value any) ([]byte, string, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return nil, "", err
	}
	return append(body, '\n'), "application/json", nil
}

// BrowsableAPIRenderer renders minimal HTML metadata.
type BrowsableAPIRenderer struct{}

func (BrowsableAPIRenderer) MediaType() string { return "text/html" }

func (BrowsableAPIRenderer) Render(value any) ([]byte, string, error) {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, "", err
	}
	return []byte("<pre>" + string(body) + "</pre>"), "text/html", nil
}

// PlainTextRenderer renders errors as text.
type PlainTextRenderer struct{}

func (PlainTextRenderer) MediaType() string { return "text/plain" }

func (PlainTextRenderer) Render(value any) ([]byte, string, error) {
	if apiErr, ok := value.(APIError); ok {
		return []byte(apiErr.Code + ": " + apiErr.Message), "text/plain", nil
	}
	return []byte(fmt.Sprint(value)), "text/plain", nil
}
