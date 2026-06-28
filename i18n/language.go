package i18n

import (
	"context"
	"strconv"
	"strings"
)

type languageContextKey struct{}
type defaultLanguageContextKey struct{}

// WithLanguage stores the active language in context.
func WithLanguage(ctx context.Context, language string) context.Context {
	return context.WithValue(ctx, languageContextKey{}, language)
}

// LanguageFromContext returns the active language from context.
func LanguageFromContext(ctx context.Context) string {
	language, _ := ctx.Value(languageContextKey{}).(string)
	if language == "" {
		language, _ = ctx.Value(defaultLanguageContextKey{}).(string)
	}
	return language
}

func WithDefaultLanguage(ctx context.Context, language string) context.Context {
	return context.WithValue(ctx, defaultLanguageContextKey{}, language)
}

type Catalog interface {
	Translate(language string, messageID string) (string, bool)
}

type MemoryCatalog struct {
	values map[string]map[string]string
}

func NewMemoryCatalog(values map[string]map[string]string) MemoryCatalog {
	cloned := make(map[string]map[string]string, len(values))
	for language, translations := range values {
		cloned[language] = make(map[string]string, len(translations))
		for key, value := range translations {
			cloned[language][key] = value
		}
	}
	return MemoryCatalog{values: cloned}
}

func (c MemoryCatalog) Translate(language string, messageID string) (string, bool) {
	translations := c.values[language]
	if translations == nil {
		return "", false
	}
	value, ok := translations[messageID]
	return value, ok
}

func Translate(ctx context.Context, catalog Catalog, messageID string) string {
	if catalog == nil {
		return messageID
	}
	if translated, ok := catalog.Translate(LanguageFromContext(ctx), messageID); ok {
		return translated
	}
	return messageID
}

type LazyValue struct {
	messageID string
	catalog   Catalog
}

func Lazy(messageID string, catalog Catalog) LazyValue {
	return LazyValue{messageID: messageID, catalog: catalog}
}

func (v LazyValue) String(ctx context.Context) string {
	return Translate(ctx, v.catalog, v.messageID)
}

// NegotiateLanguage chooses the best supported language from an Accept-Language header.
func NegotiateLanguage(header string, supported []string, fallback string) string {
	if len(supported) == 0 {
		return fallback
	}

	supportedByLower := make(map[string]string, len(supported))
	for _, language := range supported {
		supportedByLower[strings.ToLower(language)] = language
	}

	for _, candidate := range parseAcceptLanguage(header) {
		normalized := strings.ToLower(candidate)
		if language, ok := supportedByLower[normalized]; ok {
			return language
		}
		if index := strings.Index(normalized, "-"); index > -1 {
			if language, ok := supportedByLower[normalized[:index]]; ok {
				return language
			}
		}
	}
	return fallback
}

type languageCandidate struct {
	value string
	q     float64
	order int
}

func parseAcceptLanguage(header string) []string {
	parts := strings.Split(header, ",")
	candidates := make([]languageCandidate, 0, len(parts))
	for index, part := range parts {
		value, q := parseLanguagePart(part)
		if value == "" || q <= 0 {
			continue
		}
		candidates = append(candidates, languageCandidate{value: value, q: q, order: index})
	}

	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].q > candidates[i].q || candidates[j].q == candidates[i].q && candidates[j].order < candidates[i].order {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	values := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		values = append(values, candidate.value)
	}
	return values
}

func parseLanguagePart(part string) (string, float64) {
	segments := strings.Split(part, ";")
	value := strings.TrimSpace(segments[0])
	q := 1.0
	for _, segment := range segments[1:] {
		segment = strings.TrimSpace(segment)
		if !strings.HasPrefix(segment, "q=") {
			continue
		}
		parsed, err := strconv.ParseFloat(strings.TrimPrefix(segment, "q="), 64)
		if err == nil {
			q = parsed
		}
	}
	return value, q
}
