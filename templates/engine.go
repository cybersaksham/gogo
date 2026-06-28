package templates

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"sort"
)

var ErrTemplateNotFound = errors.New("template not found")

// SafeString bypasses template escaping for already-trusted HTML.
type SafeString = template.HTML

// Safe marks a value as trusted HTML.
func Safe(value string) SafeString {
	return template.HTML(value)
}

// Engine renders templates with framework conventions.
type Engine struct {
	templates  map[string]string
	funcs      template.FuncMap
	processors []ContextProcessor
}

// Option configures an Engine.
type Option func(*Engine)

func NewEngine(options ...Option) *Engine {
	engine := &Engine{
		templates: map[string]string{},
		funcs: template.FuncMap{
			"safe": Safe,
		},
	}
	for _, option := range options {
		option(engine)
	}
	return engine
}

func WithTemplates(templates map[string]string) Option {
	return func(engine *Engine) {
		engine.templates = make(map[string]string, len(templates))
		for name, source := range templates {
			engine.templates[name] = source
		}
	}
}

func WithFuncMap(funcs template.FuncMap) Option {
	return func(engine *Engine) {
		if engine.funcs == nil {
			engine.funcs = template.FuncMap{}
		}
		for name, fn := range funcs {
			engine.funcs[name] = fn
		}
	}
}

func WithContextProcessors(processors ...ContextProcessor) Option {
	return func(engine *Engine) {
		engine.processors = append(engine.processors, processors...)
	}
}

func (e *Engine) Render(name string, context Context) (string, error) {
	if e == nil {
		return "", fmt.Errorf("%w: engine is nil", ErrTemplateNotFound)
	}
	if _, ok := e.templates[name]; !ok {
		return "", fmt.Errorf("%w: %s", ErrTemplateNotFound, name)
	}
	set, err := e.templateSet(name)
	if err != nil {
		return "", err
	}
	renderContext := e.applyProcessors(context)
	var buffer bytes.Buffer
	if err := set.ExecuteTemplate(&buffer, name, renderContext); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (e *Engine) RenderString(source string, context Context) (string, error) {
	clone := NewEngine(
		WithTemplates(map[string]string{"string": source}),
		WithFuncMap(e.funcs),
		WithContextProcessors(e.processors...),
	)
	return clone.Render("string", context)
}

func (e *Engine) templateSet(target string) (*template.Template, error) {
	root := template.New("").Funcs(e.funcs)
	names := make([]string, 0, len(e.templates))
	for name := range e.templates {
		if name != target {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	names = append(names, target)
	for _, name := range names {
		if _, err := root.New(name).Parse(e.templates[name]); err != nil {
			return nil, err
		}
	}
	return root, nil
}

func (e *Engine) applyProcessors(context Context) Context {
	renderContext := cloneContext(context)
	for _, processor := range e.processors {
		if processor == nil {
			continue
		}
		next := processor(renderContext)
		if next != nil {
			renderContext = next
		}
	}
	return renderContext
}
