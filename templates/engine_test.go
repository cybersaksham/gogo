package templates

import (
	"errors"
	"testing"
)

func TestEngineRendersEscapesIncludesAndSafeStrings(t *testing.T) {
	engine := NewEngine(WithTemplates(map[string]string{
		"page":  `Hello {{.name}} {{template "badge" .}} {{.safe}}`,
		"badge": `<span>{{.label}}</span>`,
	}))

	rendered, err := engine.Render("page", Context{
		"name":  `<Saksham>`,
		"label": `<Admin>`,
		"safe":  Safe(`<strong>ok</strong>`),
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	want := `Hello &lt;Saksham&gt; <span>&lt;Admin&gt;</span> <strong>ok</strong>`
	if rendered != want {
		t.Fatalf("Render() = %q, want %q", rendered, want)
	}
}

func TestEngineAppliesContextProcessors(t *testing.T) {
	engine := NewEngine(
		WithTemplates(map[string]string{
			"page": `{{.request}} {{.user}} {{range .messages}}{{.Level}}:{{.Text}}{{end}} {{.csrf_token}} {{.static_url}} {{.media_url}}`,
		}),
		WithContextProcessors(
			RequestProcessor("request-1"),
			UserProcessor("saksham"),
			MessagesProcessor([]Message{{Level: "info", Text: "saved"}}),
			CSRFTokenProcessor("csrf-123"),
			StaticURLProcessor("/static/"),
			MediaURLProcessor("/media/"),
		),
	)

	rendered, err := engine.Render("page", nil)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	want := `request-1 saksham info:saved csrf-123 /static/ /media/`
	if rendered != want {
		t.Fatalf("Render() = %q, want %q", rendered, want)
	}
}

func TestEngineSupportsBlockStyleInheritance(t *testing.T) {
	engine := NewEngine(WithTemplates(map[string]string{
		"base": `<html><body>{{block "content" .}}Default{{end}}</body></html>`,
		"page": `{{define "content"}}Child {{.title}}{{end}}{{template "base" .}}`,
	}))

	rendered, err := engine.Render("page", Context{"title": `<Title>`})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	want := `<html><body>Child &lt;Title&gt;</body></html>`
	if rendered != want {
		t.Fatalf("Render() = %q, want %q", rendered, want)
	}
}

func TestEngineReturnsMissingTemplateError(t *testing.T) {
	engine := NewEngine(WithTemplates(map[string]string{"page": `ok`}))

	_, err := engine.Render("missing", nil)
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Fatalf("Render(missing) error = %v, want ErrTemplateNotFound", err)
	}
}
