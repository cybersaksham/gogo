package templates

import (
	"html/template"
	"strings"
	"time"
)

// URLResolver resolves named routes for templates.
type URLResolver func(name string, args ...any) (string, error)

// HelperConfig configures template tag and filter helpers.
type HelperConfig struct {
	URLResolver URLResolver
	StaticURL   string
	MediaURL    string
	Now         time.Time
}

func WithTemplateHelpers(config HelperConfig) Option {
	return WithFuncMap(TemplateHelpers(config))
}

func TemplateHelpers(config HelperConfig) template.FuncMap {
	funcs := template.FuncMap{}
	for name, fn := range TemplateTags(config) {
		funcs[name] = fn
	}
	for name, fn := range TemplateFilters() {
		funcs[name] = fn
	}
	return funcs
}

func TemplateTags(config HelperConfig) template.FuncMap {
	return template.FuncMap{
		"url": func(name string, args ...any) (string, error) {
			if config.URLResolver == nil {
				return "", ErrTemplateNotFound
			}
			return config.URLResolver(name, args...)
		},
		"static": func(path string) string {
			return joinURLPath(config.StaticURL, path)
		},
		"media": func(path string) string {
			return joinURLPath(config.MediaURL, path)
		},
	}
}

func joinURLPath(base, path string) string {
	if base == "" {
		return path
	}
	if path == "" {
		return base
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
}
